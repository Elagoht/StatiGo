package router

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"

	"statigo/framework/middleware"
	"statigo/framework/templates"
)

// RouteConfig represents a single route configuration from JSON.
type RouteConfig struct {
	Canonical string            `json:"canonical"`
	Paths     map[string]string `json:"paths"`
	Template  string            `json:"template"`
	Handler   string            `json:"handler"`  // Handler name (e.g., "index", "content")
	Title     string            `json:"title"`    // Translation key for page title
	Strategy  string            `json:"strategy"` // Caching strategy: "static", "incremental", "dynamic", "immutable"
}

// RoutesConfig represents the complete routes configuration file.
type RoutesConfig struct {
	Routes []RouteConfig `json:"routes"`
}

// LoadRoutesFromJSON loads route configurations from a JSON file.
func LoadRoutesFromJSON(
	configFS fs.FS,
	filePath string,
	registry *Registry,
	renderer *templates.Renderer,
	customHandlers map[string]http.HandlerFunc,
	logger *slog.Logger,
) error {
	// Read JSON file
	data, err := fs.ReadFile(configFS, filePath)
	if err != nil {
		return fmt.Errorf("failed to read routes file: %w", err)
	}

	// Parse JSON
	var config RoutesConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse routes JSON: %w", err)
	}

	logger.Info("Loading routes from JSON", "file", filePath, "count", len(config.Routes))

	// Register each route
	for _, routeConfig := range config.Routes {
		var handler http.HandlerFunc

		// Determine which handler to use
		switch routeConfig.Handler {
		case "content":
			// Create inline content handler
			templateName := routeConfig.Template
			handler = func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				lang := middleware.GetLanguage(ctx)
				layoutData := middleware.GetLayoutData(ctx)
				canonical := GetCanonicalPath(ctx)

				renderer.Render(w, templateName, map[string]interface{}{
					"Lang":      lang,
					"Data":      map[string]interface{}{},
					"Layout":    layoutData,
					"Canonical": canonical,
					"WebAppURL": os.Getenv("WEBAPP_URL"),
				})
			}
		default:
			// Check if custom handler exists
			if customHandlers != nil {
				if customHandler, exists := customHandlers[routeConfig.Handler]; exists {
					handler = customHandler
				} else {
					// Fall back to content handler
					logger.Warn("Custom handler not found, using content handler",
						"handler", routeConfig.Handler,
						"canonical", routeConfig.Canonical)
					templateName := routeConfig.Template
					handler = func(w http.ResponseWriter, r *http.Request) {
						ctx := r.Context()
						lang := middleware.GetLanguage(ctx)
						layoutData := middleware.GetLayoutData(ctx)
						canonical := GetCanonicalPath(ctx)

						renderer.Render(w, templateName, map[string]interface{}{
							"Lang":      lang,
							"Data":      map[string]interface{}{},
							"Layout":    layoutData,
							"Canonical": canonical,
							"WebAppURL": os.Getenv("WEBAPP_URL"),
						})
					}
				}
			} else {
				// No custom handlers provided, use content handler
				templateName := routeConfig.Template
				handler = func(w http.ResponseWriter, r *http.Request) {
					ctx := r.Context()
					lang := middleware.GetLanguage(ctx)
					layoutData := middleware.GetLayoutData(ctx)
					canonical := GetCanonicalPath(ctx)

					renderer.Render(w, templateName, map[string]interface{}{
						"Lang":      lang,
						"Data":      map[string]interface{}{},
						"Layout":    layoutData,
						"Canonical": canonical,
						"WebAppURL": os.Getenv("WEBAPP_URL"),
					})
				}
			}
		}

		// Add route to registry
		if err := registry.AddRoute(RouteDefinition{
			Canonical: routeConfig.Canonical,
			Paths:     routeConfig.Paths,
			Handler:   handler,
			Template:  routeConfig.Template,
			Title:     routeConfig.Title,
			Strategy:  routeConfig.Strategy,
		}); err != nil {
			return fmt.Errorf("failed to add route %s: %w", routeConfig.Canonical, err)
		}

		logger.Debug("Registered route",
			"canonical", routeConfig.Canonical,
			"handler", routeConfig.Handler,
			"template", routeConfig.Template)
	}

	logger.Info("Successfully loaded all routes", "count", len(config.Routes))
	return nil
}
