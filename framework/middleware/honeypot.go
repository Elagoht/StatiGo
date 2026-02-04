package middleware

import (
	"log/slog"
	"net/http"

	"statigo/framework/security"
)

// HoneypotMiddleware creates a middleware that intercepts honeypot paths and bans IPs.
func HoneypotMiddleware(banList *security.IPBanList, honeypotPaths []string, logger *slog.Logger) func(http.Handler) http.Handler {
	// Create a map for faster lookup
	pathMap := make(map[string]bool)
	for _, path := range honeypotPaths {
		pathMap[path] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if the current path is a honeypot
			if pathMap[r.URL.Path] {
				clientIP := GetClientIP(r)
				userAgent := r.UserAgent()
				path := r.URL.Path

				// Ban the IP
				if err := banList.BanIP(clientIP, "Honeypot trigger", userAgent, path); err != nil {
					logger.Error("Failed to ban IP",
						"ip", clientIP,
						"error", err,
					)
				}

				logger.Warn("Honeypot triggered",
					"ip", clientIP,
					"path", path,
					"method", r.Method,
					"user_agent", userAgent,
					"referer", r.Referer(),
				)

				// Return 404 to make it look like the endpoint doesn't exist
				http.NotFound(w, r)
				return
			}

			// Not a honeypot path, continue to next middleware
			next.ServeHTTP(w, r)
		})
	}
}

// DefaultHoneypotPaths returns common paths that attackers probe.
func DefaultHoneypotPaths() []string {
	return []string{
		"/admin/login",
		"/administrator",
		"/cpanel",
		"/phpMyAdmin",
		"/.env",
		"/.git/config",
		"/backup.sql",
		"/wp-login.php",
		"/xmlrpc.php",
		"/server-status",
		"/config.php",
		"/dashboard",
		"/api/v1/admin",
		"/robots.txt.bak",
		"/wp-admin",
	}
}
