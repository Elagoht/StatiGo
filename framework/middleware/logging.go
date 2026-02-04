package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"statigo/framework/logger"
)

// responseWriter wraps http.ResponseWriter to capture status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    int64
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.written += int64(n)
	return n, err
}

// StructuredLogger creates a middleware that logs HTTP requests with structured logging.
func StructuredLogger(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Generate and attach request ID
			requestID := logger.GenerateRequestID()
			ctx := logger.WithRequestID(r.Context(), requestID)
			r = r.WithContext(ctx)

			// Wrap response writer to capture status code
			wrapped := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // Default status code
			}

			// Call next handler
			next.ServeHTTP(wrapped, r)

			// Log request details
			duration := time.Since(start)

			log.LogAttrs(
				ctx,
				slog.LevelInfo,
				"HTTP request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("remote_addr", r.RemoteAddr),
				slog.Int("status", wrapped.statusCode),
				slog.Int64("bytes", wrapped.written),
				slog.Duration("duration", duration),
				slog.String("request_id", requestID),
				slog.String("user_agent", r.UserAgent()),
			)
		})
	}
}
