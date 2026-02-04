package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	fwctx "statigo/framework/context"
)

// LayoutData contains shared data for page layouts.
type LayoutData struct {
	SiteURL     string
	CurrentYear int
}

// LayoutDataMiddleware injects shared layout data into the request context.
func LayoutDataMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get site URL from environment
			siteURL := os.Getenv("BASE_URL")
			if siteURL == "" {
				siteURL = "http://localhost:8080"
			}

			layoutData := LayoutData{
				SiteURL: siteURL,
			}

			ctx := fwctx.SetLayoutData(r.Context(), layoutData)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetLayoutData retrieves layout data from context.
func GetLayoutData(ctx context.Context) LayoutData {
	if data := fwctx.GetLayoutData(ctx); data != nil {
		if layoutData, ok := data.(LayoutData); ok {
			return layoutData
		}
	}
	return LayoutData{}
}
