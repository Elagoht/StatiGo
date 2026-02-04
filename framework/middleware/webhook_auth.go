package middleware

import (
	"log/slog"
	"net/http"
)

// WebhookAuth validates webhook requests using X-Webhook-Secret header.
func WebhookAuth(webhookSecret string, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the secret from header
			providedSecret := r.Header.Get("X-Webhook-Secret")

			// Validate secret
			if providedSecret == "" {
				logger.Warn("webhook auth failed - missing X-Webhook-Secret header",
					slog.String("remote_addr", r.RemoteAddr),
					slog.String("path", r.URL.Path),
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"success":false,"message":"Missing X-Webhook-Secret header"}`))
				return
			}

			if providedSecret != webhookSecret {
				logger.Warn("webhook auth failed - invalid X-Webhook-Secret",
					slog.String("remote_addr", r.RemoteAddr),
					slog.String("path", r.URL.Path),
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"success":false,"message":"Invalid X-Webhook-Secret"}`))
				return
			}

			// Secret is valid, proceed
			logger.Debug("webhook authenticated",
				slog.String("remote_addr", r.RemoteAddr),
				slog.String("path", r.URL.Path),
			)
			next.ServeHTTP(w, r)
		})
	}
}
