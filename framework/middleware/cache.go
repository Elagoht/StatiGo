package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"strings"

	"statigo/framework/cache"
	fwctx "statigo/framework/context"
)

// CacheMiddleware creates middleware that serves cached responses.
// Supports ETag-based cache validation, returning 304 Not Modified
// when the client's cached version matches.
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
				etag := `W/"` + entry.ETag + `"`

				// Check If-None-Match for 304 Not Modified
				if etagMatch(r.Header.Get("If-None-Match"), etag) {
					w.Header().Set("ETag", etag)
					w.Header().Set("Cache-Control", "no-cache")
					w.WriteHeader(http.StatusNotModified)
					return
				}

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
				w.Header().Set("ETag", etag)
				w.Header().Set("Cache-Control", "no-cache")
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

			// Create response recorder that buffers the response
			rec := &responseRecorder{
				ResponseWriter: w,
				body:           &bytes.Buffer{},
				statusCode:     http.StatusOK,
			}

			// Serve the request (response is buffered in the recorder)
			next.ServeHTTP(rec, r)

			// Only cache successful responses
			if rec.statusCode == http.StatusOK {
				content := rec.body.Bytes()

				// Store in cache
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

					// Set ETag from the newly cached entry
					if cachedEntry, ok := cacheManager.Get(cacheKey); ok {
						w.Header().Set("ETag", `W/"`+cachedEntry.ETag+`"`)
						w.Header().Set("Cache-Control", "no-cache")
					}
				}
			}

			// Write the buffered response to the underlying writer
			w.WriteHeader(rec.statusCode)
			w.Write(rec.body.Bytes())
		})
	}
}

// etagMatch checks if the If-None-Match header value matches the given ETag.
func etagMatch(ifNoneMatch, etag string) bool {
	if ifNoneMatch == "" {
		return false
	}
	if ifNoneMatch == "*" {
		return true
	}
	for _, candidate := range strings.Split(ifNoneMatch, ",") {
		if strings.TrimSpace(candidate) == etag {
			return true
		}
	}
	return false
}

// responseRecorder captures response data for caching without writing through.
// This allows setting headers (like ETag) before the response is sent.
type responseRecorder struct {
	http.ResponseWriter
	body        *bytes.Buffer
	statusCode  int
	wroteHeader bool
}

// WriteHeader captures the status code without writing to the underlying writer.
func (r *responseRecorder) WriteHeader(statusCode int) {
	if !r.wroteHeader {
		r.statusCode = statusCode
		r.wroteHeader = true
	}
}

// Write captures the response body without writing to the underlying writer.
func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.wroteHeader {
		r.wroteHeader = true
	}
	return r.body.Write(b)
}
