package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	chiMiddleware "github.com/go-chi/chi/middleware"
	"github.com/joho/godotenv"

	"statigo/example/handlers"
	"statigo/framework/cache"
	"statigo/framework/health"
	"statigo/framework/i18n"
	fwlogger "statigo/framework/logger"
	"statigo/framework/middleware"
	"statigo/framework/router"
	"statigo/framework/security"
	"statigo/framework/templates"
	"statigo/framework/utils"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found, using defaults")
	}

	// Initialize logger
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "INFO"
	}
	appLogger := fwlogger.InitLogger(logLevel)

	// Get embedded filesystems
	translationsFS := GetTranslationsFS()
	templatesFS := GetTemplatesFS()
	configFS := GetConfigFS()
	staticFS := GetStaticFS()

	// Initialize i18n with English as default
	i18nInstance, err := i18n.New(translationsFS, "en")
	if err != nil {
		appLogger.Error("Failed to initialize i18n", "error", err)
		os.Exit(1)
	}

	// Initialize routing system
	languages := []string{"en", "tr"}
	routeRegistry := router.NewRegistry(languages)

	// Initialize SEO helpers
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	seoHelpers := router.NewSEOHelpers(routeRegistry, baseURL)
	routerSEOFuncs := seoHelpers.ToTemplateFunctions()

	// Convert to templates.SEOFunctions (same structure, different package)
	seoFuncs := &templates.SEOFunctions{
		CanonicalURL:   routerSEOFuncs.CanonicalURL,
		AlternateLinks:  routerSEOFuncs.AlternateLinks,
		AlternateURLs:   routerSEOFuncs.AlternateURLs,
		LocalePath:     routerSEOFuncs.LocalePath,
	}

	// Initialize template renderer
	renderer, err := templates.NewRenderer(templatesFS, i18nInstance, seoFuncs, appLogger)
	if err != nil {
		appLogger.Error("Failed to initialize template renderer", "error", err)
		os.Exit(1)
	}

	// Initialize cache manager
	cacheDir := os.Getenv("CACHE_DIR")
	if cacheDir == "" {
		workDir, _ := os.Getwd()
		cacheDir = filepath.Join(workDir, "data", "cache")
	}
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		appLogger.Error("Failed to create cache directory", "error", err)
		os.Exit(1)
	}
	cacheManager, err := cache.NewManager(cacheDir, appLogger)
	if err != nil {
		appLogger.Error("Failed to initialize cache manager", "error", err)
		os.Exit(1)
	}
	appLogger.Info("Cache manager initialized", "dir", cacheDir)

	// Initialize example handlers
	indexHandler := handlers.NewIndexHandler(renderer, cacheManager, routeRegistry)
	notFoundHandler := handlers.NewNotFoundHandler(renderer)

	// Create custom handlers map for route loader
	customHandlers := map[string]http.HandlerFunc{
		"index": indexHandler.ServeHTTP,
	}

	// Load routes from JSON configuration
	if err := router.LoadRoutesFromJSON(
		configFS,
		"routes.json",
		routeRegistry,
		renderer,
		customHandlers,
		appLogger,
	); err != nil {
		appLogger.Error("Failed to load routes", "error", err)
		os.Exit(1)
	}

	// Initialize IP ban list
	banListFile := filepath.Join(filepath.Dir(cacheDir), "banned-ips.json")
	if err := os.MkdirAll(filepath.Dir(banListFile), 0755); err != nil {
		appLogger.Error("Failed to create data directory", "error", err)
		os.Exit(1)
	}
	ipBanList, err := security.NewIPBanList(banListFile, appLogger)
	if err != nil {
		appLogger.Error("Failed to initialize IP ban list", "error", err)
		os.Exit(1)
	}

	// Initialize health check handler
	healthHandler := health.NewHandler(5 * time.Second)

	// Create router
	r := chi.NewRouter()

	// Rate limiting configuration
	rateLimitRPS := utils.GetEnvInt("RATE_LIMIT_RPS", 10)
	rateLimitBurst := utils.GetEnvInt("RATE_LIMIT_BURST", 20)

	// Development mode check
	devMode := os.Getenv("DEV_MODE") == "true"

	// Honeypot paths for bot detection
	honeypotPaths := []string{
		"/admin", "/wp-admin", "/wp-login.php", "/.env", "/.git/config",
		"/phpMyAdmin", "/administrator", "/cpanel",
	}

	// Apply middleware
	r.Use(middleware.StructuredLogger(appLogger))
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.IPBanMiddleware(ipBanList, appLogger))
	r.Use(middleware.HoneypotMiddleware(ipBanList, honeypotPaths, appLogger))
	r.Use(middleware.RateLimiter(middleware.RateLimiterConfig{
		RPS:   rateLimitRPS,
		Burst: rateLimitBurst,
	}))
	r.Use(middleware.Compression())
	r.Use(middleware.SecurityHeadersSimple())
	r.Use(middleware.CachingHeaders(devMode))

	// Static file serving middleware
	minifier := utils.NewMinifier()
	httpFS := http.FS(staticFS)
	r.Use(staticFileMiddleware(staticFS, httpFS, minifier))

	// Language middleware
	langConfig := middleware.LanguageConfig{
		SupportedLanguages: languages,
		DefaultLanguage:    "en",
		SkipPaths:          []string{"/robots.txt", "/sitemap.xml", "/favicon.ico"},
		SkipPrefixes:       []string{"/health/", "/static/", "/styles/", "/scripts/"},
	}
	r.Use(middleware.Language(i18nInstance, langConfig))

	// Canonical path middleware
	r.Use(router.CanonicalPathMiddleware(routeRegistry))

	// Cache middleware
	r.Use(middleware.CacheMiddleware(cacheManager, appLogger))

	// Register routes
	routeRegistry.RegisterRoutes(r, func(h http.Handler) http.Handler { return h })

	// Root redirect
	r.Get("/", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/en", http.StatusFound)
	})

	// 404 handler
	r.NotFound(notFoundHandler.ServeHTTP)

	// Health endpoints
	r.Get("/health/livez", healthHandler.Liveness)
	r.Get("/health/readz", healthHandler.Readiness)

	// Set router on cache manager for revalidation
	cacheManager.SetRouter(r)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := runServer(r, port, appLogger); err != nil {
		appLogger.Error("Server error", "error", err)
		os.Exit(1)
	}
}

// staticFileMiddleware serves static files from embedded filesystem
func staticFileMiddleware(staticFS fs.FS, httpFS http.FileSystem, minifier *utils.Minifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			urlPath := req.URL.Path

			// Strip language prefix if present
			parts := strings.SplitN(strings.TrimPrefix(urlPath, "/"), "/", 2)
			if len(parts) >= 2 && len(parts[0]) == 2 {
				urlPath = "/" + parts[1]
			}

			// Strip /static/ prefix if present
			if strings.HasPrefix(urlPath, "/static/") {
				urlPath = strings.TrimPrefix(urlPath, "/static")
			}

			filePath := strings.TrimPrefix(urlPath, "/")

			// Try to serve static file
			if info, err := fs.Stat(staticFS, filePath); err == nil && !info.IsDir() {
				ext := strings.ToLower(path.Ext(filePath))

				// Minify CSS/JS files
				if ext == ".css" || ext == ".js" {
					data, err := fs.ReadFile(staticFS, filePath)
					if err != nil {
						next.ServeHTTP(w, req)
						return
					}

					var mimeType string
					switch ext {
					case ".js":
						mimeType = "application/javascript"
					case ".css":
						mimeType = "text/css"
					default:
						mimeType = "text/plain"
					}

					minified, err := minifier.MinifyBytes(mimeType, data)
					if err != nil {
						minified = data
					}

					w.Header().Set("Content-Type", mimeType+"; charset=utf-8")
					w.Write(minified)
					return
				}

				// Serve other static files
				file, err := httpFS.Open(filePath)
				if err != nil {
					next.ServeHTTP(w, req)
					return
				}
				defer file.Close()

				http.ServeContent(w, req, path.Base(filePath), info.ModTime(), file)
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}

// runServer starts the HTTP server with graceful shutdown
func runServer(handler http.Handler, port string, log *slog.Logger) error {
	shutdownTimeout := utils.GetEnvInt("SHUTDOWN_TIMEOUT", 30)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Info("Starting server", "port", port, "url", fmt.Sprintf("http://localhost:%s", port))
		serverErrors <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case sig := <-shutdown:
		log.Info("Shutdown signal received", "signal", sig.String())
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(shutdownTimeout)*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error("Graceful shutdown failed", "error", err)
			srv.Close()
			return fmt.Errorf("shutdown error: %w", err)
		}
		log.Info("Server stopped gracefully")
	}

	return nil
}
