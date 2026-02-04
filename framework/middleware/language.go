// Package middleware provides HTTP middleware for the Statigo framework.
package middleware

import (
	gocontext "context"
	"net/http"
	"strings"

	fwctx "statigo/framework/context"
	"statigo/framework/i18n"
)

// LanguageConfig configures the language detection middleware.
type LanguageConfig struct {
	SupportedLanguages []string // List of supported language codes
	DefaultLanguage    string   // Default/fallback language
	SkipPaths          []string // Paths to skip (exact match)
	SkipPrefixes       []string // Path prefixes to skip
}

// DefaultLanguageConfig returns default configuration.
func DefaultLanguageConfig() LanguageConfig {
	return LanguageConfig{
		SupportedLanguages: []string{"en"},
		DefaultLanguage:    "en",
		SkipPaths:          []string{"/robots.txt", "/sitemap.xml"},
		SkipPrefixes:       []string{"/health/", "/static/", "/webhook/", "/api/"},
	}
}

// Language middleware detects and sets the current language from URL path.
func Language(i18nInstance *i18n.I18n, config LanguageConfig) func(http.Handler) http.Handler {
	supportedLangsMap := make(map[string]bool)
	for _, lang := range config.SupportedLanguages {
		supportedLangsMap[lang] = true
	}

	skipPathsMap := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPathsMap[path] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			path := r.URL.Path

			// Skip language detection for certain paths
			if skipPathsMap[path] {
				next.ServeHTTP(w, r)
				return
			}

			// Skip language detection for certain path prefixes
			for _, prefix := range config.SkipPrefixes {
				if strings.HasPrefix(path, prefix) {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Extract language from path (e.g., /tr/about -> tr)
			parts := strings.Split(strings.Trim(path, "/"), "/")

			var lang string
			var hasLangInPath bool
			var isUnsupportedLang bool

			if len(parts) > 0 && len(parts[0]) == 2 {
				if supportedLangsMap[parts[0]] {
					lang = parts[0]
					hasLangInPath = true
				} else {
					// First part looks like a language code but isn't supported
					isUnsupportedLang = true
				}
			}

			// If no language in path, detect and redirect
			if !hasLangInPath {
				detectedLang := detectLanguage(r, supportedLangsMap, config.DefaultLanguage)

				// Set cookie
				http.SetCookie(w, &http.Cookie{
					Name:     "lang",
					Value:    detectedLang,
					Path:     "/",
					MaxAge:   365 * 24 * 60 * 60, // 1 year
					HttpOnly: true,
					SameSite: http.SameSiteLaxMode,
				})

				// Redirect to path with language prefix
				newPath := "/" + detectedLang
				if path != "/" {
					// If unsupported language detected, strip it from path
					if isUnsupportedLang && len(parts) > 1 {
						// Reconstruct path without the unsupported language prefix
						newPath += "/" + strings.Join(parts[1:], "/")
					} else if !isUnsupportedLang {
						// No language-like prefix, preserve entire path
						newPath += path
					}
					// If isUnsupportedLang but no additional parts, just redirect to /{detectedLang}
				}
				http.Redirect(w, r, newPath, http.StatusFound)
				return
			}

			// Set cookie with detected language
			http.SetCookie(w, &http.Cookie{
				Name:     "lang",
				Value:    lang,
				Path:     "/",
				MaxAge:   365 * 24 * 60 * 60,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})

			// Set language in context
			ctx := fwctx.SetLanguage(r.Context(), lang)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// detectLanguage determines the user's preferred language.
func detectLanguage(r *http.Request, supportedLangs map[string]bool, defaultLang string) string {
	// 1. Check cookie
	if cookie, err := r.Cookie("lang"); err == nil {
		if supportedLangs[cookie.Value] {
			return cookie.Value
		}
	}

	// 2. Parse Accept-Language header
	acceptLang := r.Header.Get("Accept-Language")
	if acceptLang != "" {
		languages := parseAcceptLanguage(acceptLang)
		for _, lang := range languages {
			// Get first 2 characters (e.g., "en-US" -> "en")
			langCode := strings.ToLower(lang)
			if len(langCode) >= 2 {
				langCode = langCode[:2]
			}
			if supportedLangs[langCode] {
				return langCode
			}
		}
	}

	// 3. Default language
	return defaultLang
}

// parseAcceptLanguage parses the Accept-Language header.
// Returns languages in order of preference.
func parseAcceptLanguage(header string) []string {
	var languages []string

	// Split by comma
	parts := strings.Split(header, ",")
	for _, part := range parts {
		// Remove quality value if present (e.g., "en-US;q=0.9" -> "en-US")
		lang := strings.TrimSpace(strings.Split(part, ";")[0])
		if lang != "" {
			languages = append(languages, lang)
		}
	}

	return languages
}

// GetLanguage retrieves the language from context.
func GetLanguage(ctx gocontext.Context) string {
	return fwctx.GetLanguage(ctx)
}
