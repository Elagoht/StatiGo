package middleware

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
)

// RedirectConfig represents the redirect configuration structure.
// Key: target URL (where to redirect to)
// Value: array of source URLs (that should redirect to the target)
type RedirectConfig map[string][]string

// PatternRedirect represents a pattern-based redirect with regex matching.
type PatternRedirect struct {
	pattern *regexp.Regexp // Compiled regex pattern for matching
	target  string         // Target URL template with {slug} placeholders
	source  string         // Original source pattern for logging
}

// RedirectRegistry maintains an optimized lookup table for redirects.
type RedirectRegistry struct {
	// Static redirects: source URL -> target URL (O(1) lookups)
	staticRedirects map[string]string
	// Pattern-based redirects with dynamic slug matching
	patternRedirects []PatternRedirect
	logger           *slog.Logger
}

// NewRedirectRegistry creates a new redirect registry.
func NewRedirectRegistry(logger *slog.Logger) *RedirectRegistry {
	return &RedirectRegistry{
		staticRedirects:  make(map[string]string),
		patternRedirects: make([]PatternRedirect, 0),
		logger:           logger,
	}
}

// isPatternURL checks if a URL contains dynamic placeholders like {slug}.
func isPatternURL(url string) bool {
	return strings.Contains(url, "{") && strings.Contains(url, "}")
}

// patternToRegex converts a URL pattern with {slug} to a compiled regex.
func patternToRegex(pattern string) (*regexp.Regexp, error) {
	// Escape special regex characters except for our placeholders
	regexPattern := regexp.QuoteMeta(pattern)
	// Replace escaped \{slug\} with a named capture group
	regexPattern = strings.ReplaceAll(regexPattern, `\{slug\}`, `(?P<slug>[^/]+)`)
	// Anchor the pattern to match the entire path
	regexPattern = "^" + regexPattern + "$"
	return regexp.Compile(regexPattern)
}

// LoadRedirectsFromJSON loads redirect configurations from a JSON file.
func LoadRedirectsFromJSON(configFS fs.FS, filePath string, logger *slog.Logger) (*RedirectRegistry, error) {
	registry := NewRedirectRegistry(logger)

	// Read JSON file
	data, err := fs.ReadFile(configFS, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read redirects file: %w", err)
	}

	// Parse JSON
	var config RedirectConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse redirects JSON: %w", err)
	}

	logger.Info("Loading redirects from JSON", "file", filePath)

	// Build optimized lookup tables
	staticCount := 0
	patternCount := 0

	for targetURL, sourceURLs := range config {
		for _, sourceURL := range sourceURLs {
			// Check if this is a pattern-based redirect
			if isPatternURL(sourceURL) {
				// Compile the pattern to regex
				pattern, err := patternToRegex(sourceURL)
				if err != nil {
					logger.Error("Failed to compile redirect pattern, skipping",
						"source", sourceURL,
						"target", targetURL,
						"error", err)
					continue
				}

				// Add to pattern redirects
				registry.patternRedirects = append(registry.patternRedirects, PatternRedirect{
					pattern: pattern,
					target:  targetURL,
					source:  sourceURL,
				})
				patternCount++

				logger.Debug("Registered pattern redirect",
					"source", sourceURL,
					"target", targetURL)
			} else {
				// Static redirect - check for duplicates
				if existingTarget, exists := registry.staticRedirects[sourceURL]; exists {
					logger.Warn("Duplicate redirect source URL found, overwriting",
						"source", sourceURL,
						"old_target", existingTarget,
						"new_target", targetURL)
				}

				registry.staticRedirects[sourceURL] = targetURL
				staticCount++

				logger.Debug("Registered static redirect",
					"source", sourceURL,
					"target", targetURL)
			}
		}
	}

	logger.Info("Successfully loaded redirects",
		"static_redirects", staticCount,
		"pattern_redirects", patternCount,
		"total_redirects", staticCount+patternCount)

	return registry, nil
}

// GetRedirectTarget returns the target URL for a given source URL.
// Returns empty string if no redirect exists.
func (rr *RedirectRegistry) GetRedirectTarget(sourceURL string) string {
	// First, check static redirects (O(1) lookup)
	if target, exists := rr.staticRedirects[sourceURL]; exists {
		return target
	}

	// If no static match, check pattern redirects
	for _, patternRedirect := range rr.patternRedirects {
		if matches := patternRedirect.pattern.FindStringSubmatch(sourceURL); matches != nil {
			// Extract captured groups
			target := patternRedirect.target

			// Replace {slug} in target with captured value
			for i, name := range patternRedirect.pattern.SubexpNames() {
				if i > 0 && i < len(matches) && name == "slug" {
					target = strings.ReplaceAll(target, "{slug}", matches[i])
				}
			}

			return target
		}
	}

	return ""
}

// Count returns the total number of redirects (static + pattern).
func (rr *RedirectRegistry) Count() int {
	return len(rr.staticRedirects) + len(rr.patternRedirects)
}

// RedirectMiddleware handles URL redirects using a 301 Moved Permanently status.
func RedirectMiddleware(registry *RedirectRegistry, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestPath := r.URL.Path

			// Remove trailing slash for lookup (except for root path)
			lookupPath := requestPath
			if len(lookupPath) > 1 && strings.HasSuffix(lookupPath, "/") {
				lookupPath = strings.TrimSuffix(lookupPath, "/")
			}

			// Check if a redirect exists for this path
			if targetURL := registry.GetRedirectTarget(lookupPath); targetURL != "" {
				// Log the redirect
				logger.Info("Redirecting request",
					"source", requestPath,
					"target", targetURL,
					"method", r.Method,
					"remote_addr", r.RemoteAddr)

				// Preserve query string if present
				targetWithQuery := targetURL
				if r.URL.RawQuery != "" {
					targetWithQuery = targetURL + "?" + r.URL.RawQuery
				}

				// 301 Moved Permanently
				http.Redirect(w, r, targetWithQuery, http.StatusMovedPermanently)
				return
			}

			// No redirect found, continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}
