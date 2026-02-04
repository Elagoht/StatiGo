package middleware

import (
	"net/http"
	"path/filepath"
	"strings"
)

// CachingHeaders sets appropriate Cache-Control headers based on file type.
// In development mode (devMode=true), uses shorter cache durations.
func CachingHeaders(devMode bool) func(http.Handler) http.Handler {
	// Static assets that should be cached immutably
	immutableAssets := map[string]bool{
		".css":   true,
		".js":    true,
		".mjs":   true,
		".png":   true,
		".jpg":   true,
		".jpeg":  true,
		".webp":  true,
		".svg":   true,
		".ico":   true,
		".woff":  true,
		".woff2": true,
		".ttf":   true,
	}

	// Config files that should have shorter cache
	configFiles := map[string]bool{
		"robots.txt":       true,
		"manifest.json":    true,
		"sitemap.xml":      true,
		"site.webmanifest": true,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path
			ext := strings.ToLower(filepath.Ext(path))
			filename := filepath.Base(path)

			// Set cache headers based on file type
			if immutableAssets[ext] {
				if devMode {
					// Development mode - use no-cache to always revalidate
					w.Header().Set("Cache-Control", "no-cache")
				} else {
					// Production mode - cache forever with immutable flag
					w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
				}
			} else if configFiles[filename] {
				// Config files with shorter cache
				w.Header().Set("Cache-Control", "public, max-age=3600")
			}

			next.ServeHTTP(w, r)
		})
	}
}
