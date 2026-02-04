package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"golang.org/x/time/rate"
)

// RateLimiterConfig configures the rate limiter middleware.
type RateLimiterConfig struct {
	RPS              int      // Requests per second for dynamic content
	Burst            int      // Maximum burst size
	StaticMultiplier int      // Multiplier for static asset limits (default: 10)
	CrawlerBypass    bool     // Whether to bypass rate limiting for crawlers
	Crawlers         []string // List of crawler user-agent substrings
}

// DefaultRateLimiterConfig returns default configuration.
func DefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		RPS:              10,
		Burst:            20,
		StaticMultiplier: 10,
		CrawlerBypass:    true,
		Crawlers: []string{
			"Googlebot",
			"Google-PageSpeed",
			"Google-InspectionTool",
			"bingbot",
			"Slurp",
			"duckduckbot",
			"baiduspider",
			"yandexbot",
			"facebookexternalhit",
			"twitterbot",
			"linkedinbot",
			"SemrushBot",
			"AhrefsBot",
			"MJ12bot",
			"DotBot",
			"applebot",
			"Alexa",
			"GTmetrix",
			"PTST/",
			"Lighthouse",
			"HeadlessChrome",
		},
	}
}

// RateLimiter creates a middleware that limits requests using a token bucket algorithm.
func RateLimiter(config RateLimiterConfig) func(http.Handler) http.Handler {
	// Limiter for dynamic content (HTML pages, API endpoints)
	dynamicLimiter := rate.NewLimiter(rate.Limit(config.RPS), config.Burst)

	// Limiter for static assets (higher limits)
	staticMultiplier := config.StaticMultiplier
	if staticMultiplier <= 0 {
		staticMultiplier = 10
	}
	staticRPS := config.RPS * staticMultiplier
	staticBurst := config.Burst * staticMultiplier
	staticLimiter := rate.NewLimiter(rate.Limit(staticRPS), staticBurst)

	// Build crawler lookup
	crawlerLower := make([]string, len(config.Crawlers))
	for i, c := range config.Crawlers {
		crawlerLower[i] = strings.ToLower(c)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Bypass rate limiting for internal bootstrap requests
			if r.Header.Get("X-Internal-Bootstrap") == "true" {
				next.ServeHTTP(w, r)
				return
			}

			// Bypass rate limiting for legitimate crawlers
			if config.CrawlerBypass {
				userAgent := strings.ToLower(r.Header.Get("User-Agent"))
				for _, crawler := range crawlerLower {
					if strings.Contains(userAgent, crawler) {
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			var limiter *rate.Limiter
			var limitRPS, limitBurst int

			// Use higher limits for static assets
			if isStaticAsset(r.URL.Path) {
				limiter = staticLimiter
				limitRPS = staticRPS
				limitBurst = staticBurst
			} else {
				limiter = dynamicLimiter
				limitRPS = config.RPS
				limitBurst = config.Burst
			}

			if !limiter.Allow() {
				// Calculate retry-after based on the rate limit
				retryAfter := int(1.0 / float64(limitRPS))
				if retryAfter < 1 {
					retryAfter = 1
				}

				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(limitRPS))
				w.Header().Set("X-RateLimit-Burst", strconv.Itoa(limitBurst))
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// isStaticAsset checks if the request path is for a static asset.
func isStaticAsset(path string) bool {
	staticPrefixes := []string{"/assets/", "/static/", "/favicon.ico", "/robots.txt", "/manifest.json"}
	for _, prefix := range staticPrefixes {
		if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}
