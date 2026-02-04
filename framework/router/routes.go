// Package router provides multi-language URL routing with canonical path management.
package router

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

// RouteDefinition represents a canonical route with language-specific URLs.
type RouteDefinition struct {
	Canonical string            // Canonical path, e.g., "/features"
	Paths     map[string]string // Language -> URL path: {"en": "/en/features", "tr": "/tr/ozellikler"}
	Handler   http.HandlerFunc  // Handler function for this route
	Template  string            // Template name (e.g., "content.html")
	Title     string            // Translation key for page title (e.g., "main.title")
	Strategy  string            // Caching strategy: "static", "incremental", "dynamic", "immutable"
}

// Registry maintains the mapping between canonical paths and route definitions.
type Registry struct {
	routes       []RouteDefinition
	pathToRoute  map[string]*RouteDefinition // Maps actual paths to route definitions
	canonicalMap map[string]*RouteDefinition // Maps canonical paths to route definitions
	languages    []string                    // Supported languages
}

// NewRegistry creates a new route registry for the given languages.
func NewRegistry(languages []string) *Registry {
	return &Registry{
		routes:       make([]RouteDefinition, 0),
		pathToRoute:  make(map[string]*RouteDefinition),
		canonicalMap: make(map[string]*RouteDefinition),
		languages:    languages,
	}
}

// AddRoute registers a new route definition.
// Returns an error if any language is missing a path definition.
func (r *Registry) AddRoute(def RouteDefinition) error {
	// Validate that all languages have paths
	for _, lang := range r.languages {
		if _, exists := def.Paths[lang]; !exists {
			return fmt.Errorf("missing path for language: %s in route %s", lang, def.Canonical)
		}
	}

	// Store in registry
	r.routes = append(r.routes, def)
	routePtr := &r.routes[len(r.routes)-1]
	r.canonicalMap[def.Canonical] = routePtr

	// Map all language-specific paths to this definition
	for _, path := range def.Paths {
		r.pathToRoute[path] = routePtr
		// Also map the path with trailing slash (unless it's the root path)
		if path != "/" {
			r.pathToRoute[path+"/"] = routePtr
		}
	}

	return nil
}

// GetByPath returns the route definition for a given path.
func (r *Registry) GetByPath(path string) *RouteDefinition {
	return r.pathToRoute[path]
}

// GetByCanonical returns the route definition for a canonical path.
func (r *Registry) GetByCanonical(canonical string) *RouteDefinition {
	return r.canonicalMap[canonical]
}

// GetAlternateURLs returns all language variants for a canonical path.
func (r *Registry) GetAlternateURLs(canonical string) map[string]string {
	if route := r.canonicalMap[canonical]; route != nil {
		return route.Paths
	}
	return nil
}

// GetAll returns all registered routes.
func (r *Registry) GetAll() []RouteDefinition {
	return r.routes
}

// Languages returns the supported languages.
func (r *Registry) Languages() []string {
	return r.languages
}

// RegisterRoutes automatically registers all routes from the registry with a chi router.
// The canonicalMiddleware wraps each handler to inject canonical path context.
func (r *Registry) RegisterRoutes(router chi.Router, canonicalMiddleware func(http.Handler) http.Handler) {
	for _, route := range r.routes {
		// Wrap the handler once per route with canonical middleware
		wrappedHandler := canonicalMiddleware(route.Handler)

		// Convert to HandlerFunc
		handlerFunc := func(w http.ResponseWriter, req *http.Request) {
			wrappedHandler.ServeHTTP(w, req)
		}

		// Register each language-specific path with the same wrapped handler
		for _, path := range route.Paths {
			router.Get(path, handlerFunc)

			// Also register with trailing slash
			if path != "/" {
				router.Get(path+"/", handlerFunc)
			}
		}
	}
}
