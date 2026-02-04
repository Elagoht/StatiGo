package router

import (
	"context"
	"net/http"

	fwctx "statigo/framework/context"
)

// CanonicalPathMiddleware creates middleware that stores canonical path,
// page title, and cache strategy in the request context.
func CanonicalPathMiddleware(registry *Registry) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// Look up the route definition
			if route := registry.GetByPath(path); route != nil {
				ctx := fwctx.SetCanonicalPath(r.Context(), route.Canonical)
				if route.Title != "" {
					ctx = fwctx.SetPageTitle(ctx, route.Title)
				}
				if route.Strategy != "" {
					ctx = fwctx.SetStrategy(ctx, route.Strategy)
				}
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// No canonical path found, continue without it
			next.ServeHTTP(w, r)
		})
	}
}

// GetCanonicalPath retrieves the canonical path from context.
func GetCanonicalPath(ctx context.Context) string {
	return fwctx.GetCanonicalPath(ctx)
}

// GetPageTitle retrieves the page title translation key from context.
func GetPageTitle(ctx context.Context) string {
	return fwctx.GetPageTitle(ctx)
}

// GetStrategy retrieves the cache strategy from context.
func GetStrategy(ctx context.Context) string {
	return fwctx.GetStrategy(ctx)
}
