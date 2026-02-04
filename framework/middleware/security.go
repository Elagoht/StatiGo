package middleware

import (
	"log/slog"
	"net/http"
)

// SecurityHeadersConfig configures the security headers middleware.
type SecurityHeadersConfig struct {
	HSTSMaxAge            int      // HSTS max-age in seconds (default: 31536000)
	FrameOptions          string   // X-Frame-Options value (default: "DENY")
	ContentTypeOptions    string   // X-Content-Type-Options value (default: "nosniff")
	ReferrerPolicy        string   // Referrer-Policy value (default: "strict-origin-when-cross-origin")
	ContentSecurityPolicy string   // Custom CSP (optional)
	AllowedImageSources   []string // Additional image sources for CSP
}

// DefaultSecurityHeadersConfig returns default configuration.
func DefaultSecurityHeadersConfig() SecurityHeadersConfig {
	return SecurityHeadersConfig{
		HSTSMaxAge:         31536000,
		FrameOptions:       "DENY",
		ContentTypeOptions: "nosniff",
		ReferrerPolicy:     "strict-origin-when-cross-origin",
	}
}

// SecurityHeaders middleware adds security headers to responses.
func SecurityHeaders(config SecurityHeadersConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// HSTS
			if config.HSTSMaxAge > 0 {
				w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}

			// X-Frame-Options
			if config.FrameOptions != "" {
				w.Header().Set("X-Frame-Options", config.FrameOptions)
			}

			// X-Content-Type-Options
			if config.ContentTypeOptions != "" {
				w.Header().Set("X-Content-Type-Options", config.ContentTypeOptions)
			}

			// Referrer-Policy
			if config.ReferrerPolicy != "" {
				w.Header().Set("Referrer-Policy", config.ReferrerPolicy)
			}

			// Content-Security-Policy
			if config.ContentSecurityPolicy != "" {
				w.Header().Set("Content-Security-Policy", config.ContentSecurityPolicy)
			}

			// Permissions-Policy
			w.Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

			next.ServeHTTP(w, r)
		})
	}
}

// SecurityHeadersSimple is a simplified version with sensible defaults.
func SecurityHeadersSimple() func(http.Handler) http.Handler {
	return SecurityHeaders(DefaultSecurityHeadersConfig())
}

// IPBanMiddleware creates a middleware that blocks requests from banned IPs.
func IPBanMiddleware(banList IPBanList, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := GetClientIP(r)

			if banList.IsBanned(clientIP) {
				logger.Info("Blocked request from banned IP",
					"ip", clientIP,
					"path", r.URL.Path,
					"user_agent", r.UserAgent(),
				)

				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// IPBanList interface for IP ban checking.
type IPBanList interface {
	IsBanned(ip string) bool
}

// GetClientIP extracts the real client IP from the request.
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (common with reverse proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		for i := 0; i < len(xff); i++ {
			if xff[i] == ',' {
				return xff[:i]
			}
		}
		return xff
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	addr := r.RemoteAddr
	for i := len(addr) - 1; i >= 0; i-- {
		if addr[i] == ':' {
			return addr[:i]
		}
	}

	return addr
}
