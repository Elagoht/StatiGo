package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// RouteConfig represents a route configuration for bootstrap caching.
type RouteConfig struct {
	Canonical string            `json:"canonical"`
	Paths     map[string]string `json:"paths"`
	Strategy  string            `json:"strategy"`
}

// RebuildConfig contains configuration for cache rebuilding operations.
type RebuildConfig struct {
	ConfigFS     fs.FS
	RoutesFile   string
	Languages    []string
	Router       http.Handler
	Logger       *slog.Logger
	ForceRebuild bool // If true, rebuild even if cache exists
}

// RebuildAll rebuilds all caches from routes configuration.
func (m *Manager) RebuildAll(ctx context.Context, config RebuildConfig) (int, error) {
	config.ForceRebuild = true
	return m.rebuildCaches(ctx, config, "")
}

// RebuildByStrategy rebuilds caches filtered by strategy.
func (m *Manager) RebuildByStrategy(ctx context.Context, config RebuildConfig, strategy string) (int, error) {
	config.ForceRebuild = true
	return m.rebuildCaches(ctx, config, strategy)
}

// Bootstrap pre-caches all cacheable pages on startup.
func (m *Manager) Bootstrap(ctx context.Context, config RebuildConfig) error {
	config.Logger.Info("Starting bootstrap cache warming...")

	// Load routes configuration
	data, err := fs.ReadFile(config.ConfigFS, config.RoutesFile)
	if err != nil {
		return fmt.Errorf("failed to read routes file: %w", err)
	}

	var routesConfig struct {
		Routes []RouteConfig `json:"routes"`
	}

	if err := json.Unmarshal(data, &routesConfig); err != nil {
		return fmt.Errorf("failed to parse routes JSON: %w", err)
	}

	var totalCached atomic.Int32
	startTime := time.Now()

	// Use worker pool for parallel processing
	maxWorkers := 10
	routeChan := make(chan RouteConfig, len(routesConfig.Routes))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for route := range routeChan {
				// Skip dynamic routes
				if route.Strategy == "dynamic" {
					config.Logger.Debug("Skipping dynamic route",
						slog.String("canonical", route.Canonical),
					)
					continue
				}

				// Check if route has parameters
				hasParams := strings.Contains(route.Canonical, "{")

				var count int
				var err error

				if !hasParams {
					config.Logger.Debug("Processing static route",
						slog.String("canonical", route.Canonical),
						slog.String("strategy", route.Strategy),
					)
					count, err = m.cacheStaticRoute(ctx, route, config)
					if err != nil {
						config.Logger.Error("Failed to cache static route",
							slog.String("canonical", route.Canonical),
							slog.String("error", err.Error()),
						)
						continue
					}
				}

				totalCached.Add(int32(count))
			}
		}()
	}

	// Send routes to workers
	for _, route := range routesConfig.Routes {
		routeChan <- route
	}
	close(routeChan)

	// Wait for all workers to finish
	wg.Wait()

	duration := time.Since(startTime)
	config.Logger.Info("Bootstrap cache warming completed",
		slog.Int("total_pages", int(totalCached.Load())),
		slog.Duration("duration", duration),
	)

	return nil
}

// rebuildCaches is the internal method that rebuilds caches.
func (m *Manager) rebuildCaches(ctx context.Context, config RebuildConfig, strategyFilter string) (int, error) {
	config.Logger.Info("Starting cache rebuild",
		slog.String("strategy", strategyFilter),
	)

	data, err := fs.ReadFile(config.ConfigFS, config.RoutesFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read routes file: %w", err)
	}

	var routesConfig struct {
		Routes []RouteConfig `json:"routes"`
	}

	if err := json.Unmarshal(data, &routesConfig); err != nil {
		return 0, fmt.Errorf("failed to parse routes JSON: %w", err)
	}

	var totalCached atomic.Int32
	startTime := time.Now()

	maxWorkers := 10
	routeChan := make(chan RouteConfig, len(routesConfig.Routes))
	var wg sync.WaitGroup

	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for route := range routeChan {
				if strategyFilter != "" && route.Strategy != strategyFilter {
					continue
				}

				if route.Strategy == "dynamic" {
					continue
				}

				hasParams := strings.Contains(route.Canonical, "{")

				var count int
				var err error

				if !hasParams {
					count, err = m.cacheStaticRoute(ctx, route, config)
					if err != nil {
						config.Logger.Error("Failed to cache static route",
							slog.String("canonical", route.Canonical),
							slog.String("error", err.Error()),
						)
						continue
					}
				}

				totalCached.Add(int32(count))
			}
		}()
	}

	for _, route := range routesConfig.Routes {
		routeChan <- route
	}
	close(routeChan)

	wg.Wait()

	duration := time.Since(startTime)
	config.Logger.Info("Cache rebuild completed",
		slog.String("strategy", strategyFilter),
		slog.Int("total_cached", int(totalCached.Load())),
		slog.Duration("duration", duration),
	)

	return int(totalCached.Load()), nil
}

// cacheStaticRoute caches a static route for all languages.
func (m *Manager) cacheStaticRoute(ctx context.Context, route RouteConfig, config RebuildConfig) (int, error) {
	var count atomic.Int32
	var wg sync.WaitGroup

	for _, lang := range config.Languages {
		wg.Add(1)
		go func(lang string) {
			defer wg.Done()

			cacheKey := GetCacheKey(route.Canonical, lang, nil)

			// Skip if already cached (unless force rebuild)
			if !config.ForceRebuild {
				if _, found := m.Get(cacheKey); found {
					return
				}
			}

			// Get the path for this language
			path := route.Paths[lang]
			if path == "" {
				config.Logger.Warn("No path found for language",
					slog.String("canonical", route.Canonical),
					slog.String("lang", lang),
				)
				return
			}

			// Make HTTP request to render the page
			content, err := m.makeCacheRequest(ctx, config.Router, path)
			if err != nil {
				config.Logger.Error("Failed to render page",
					slog.String("canonical", route.Canonical),
					slog.String("lang", lang),
					slog.String("path", path),
					slog.String("error", err.Error()),
				)
				return
			}

			// Store in cache (synchronous during rebuild)
			if err := m.SetSync(cacheKey, content, route.Strategy, path); err != nil {
				config.Logger.Error("Failed to store in cache",
					slog.String("key", cacheKey),
					slog.String("error", err.Error()),
				)
				return
			}

			count.Add(1)
		}(lang)
	}

	wg.Wait()
	return int(count.Load()), nil
}

// makeCacheRequest makes an HTTP request to the router and returns the response body.
func (m *Manager) makeCacheRequest(ctx context.Context, router http.Handler, path string) ([]byte, error) {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req = req.WithContext(ctx)

	// Add header to bypass rate limiting for bootstrap requests
	req.Header.Set("X-Internal-Bootstrap", "true")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		return nil, fmt.Errorf("request returned non-OK status: %d", rec.Code)
	}

	return rec.Body.Bytes(), nil
}
