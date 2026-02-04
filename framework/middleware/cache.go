package middleware

import (
	"bytes"
	"log/slog"
	"net/http"

	"statigo/framework/cache"
	fwctx "statigo/framework/context"
)

// CacheMiddleware creates middleware that serves cached responses.
func CacheMiddleware(cacheManager *cache.Manager, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only cache GET requests
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			// Get canonical path and language from context
			canonical := fwctx.GetCanonicalPath(r.Context())
			lang := fwctx.GetLanguage(r.Context())

			// Skip if no canonical path
			if canonical == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Generate cache key
			cacheKey := cache.GetCacheKey(canonical, lang, nil)

			// Try to get from cache
			entry, found := cacheManager.Get(cacheKey)
			if found && !entry.IsStale() {
				// Serve from cache
				content, err := cache.GetDecompressedContent(entry)
				if err != nil {
					logger.Warn("Failed to decompress cached content",
						slog.String("key", cacheKey),
						slog.String("error", err.Error()),
					)
					next.ServeHTTP(w, r)
					return
				}

				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Header().Set("X-Cache", "HIT")
				w.Header().Set("ETag", entry.ETag)
				w.Write(content)
				return
			}

			// Cache miss or stale - capture response for caching
			strategy := fwctx.GetStrategy(r.Context())
			if strategy == "" || strategy == "dynamic" {
				// Don't cache dynamic content
				next.ServeHTTP(w, r)
				return
			}

			// Create response recorder
			rec := &responseRecorder{
				ResponseWriter: w,
				body:           &bytes.Buffer{},
				statusCode:     http.StatusOK,
			}

			// Serve the request
			next.ServeHTTP(rec, r)

			// Only cache successful responses
			if rec.statusCode == http.StatusOK {
				content := rec.body.Bytes()

				// Store in cache (async)
				if err := cacheManager.Set(cacheKey, content, strategy, r.URL.Path); err != nil {
					logger.Warn("Failed to cache response",
						slog.String("key", cacheKey),
						slog.String("error", err.Error()),
					)
				} else {
					logger.Debug("Cached response",
						slog.String("key", cacheKey),
						slog.String("strategy", strategy),
					)
				}
			}
		})
	}
}

// responseRecorder captures response data for caching.
type responseRecorder struct {
	http.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

// WriteHeader captures the status code.
func (r *responseRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the response body.
func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}
