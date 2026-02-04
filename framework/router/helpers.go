package router

import (
	"fmt"
	"html/template"
)

// SEOHelpers provides template functions for SEO optimization.
type SEOHelpers struct {
	registry  *Registry
	deployURL string // Base URL, e.g., "https://example.com"
}

// NewSEOHelpers creates a new SEO helpers instance.
func NewSEOHelpers(registry *Registry, deployURL string) *SEOHelpers {
	return &SEOHelpers{
		registry:  registry,
		deployURL: deployURL,
	}
}

// GetCanonicalURL returns the full canonical URL for the current page.
func (sh *SEOHelpers) GetCanonicalURL(canonical string, lang string) string {
	// First, try to look up the route in the registry
	if route := sh.registry.GetByCanonical(canonical); route != nil {
		if path, exists := route.Paths[lang]; exists {
			return sh.deployURL + path
		}
	}

	// If not found in registry, the canonical might be a full path (e.g., with slug/id)
	// In this case, just prepend the deploy URL to the canonical path
	if canonical != "" && canonical[0] == '/' {
		return sh.deployURL + canonical
	}

	// Fallback to base deploy URL
	return sh.deployURL
}

// GetAlternateLinks returns HTML for hreflang alternate links.
func (sh *SEOHelpers) GetAlternateLinks(canonical string) template.HTML {
	if route := sh.registry.GetByCanonical(canonical); route != nil {
		var links string

		// Add alternate links for each language
		for lang, path := range route.Paths {
			fullURL := sh.deployURL + path
			links += fmt.Sprintf(`<link rel="alternate" hreflang="%s" href="%s" />`, lang, fullURL)
			links += "\n"
		}

		// Add x-default (typically point to main language)
		if defaultPath, exists := route.Paths["en"]; exists {
			fullURL := sh.deployURL + defaultPath
			links += fmt.Sprintf(`<link rel="alternate" hreflang="x-default" href="%s" />`, fullURL)
		}

		return template.HTML(links)
	}
	return ""
}

// GetAlternateURLs returns a map of language codes to URLs for the current page.
func (sh *SEOHelpers) GetAlternateURLs(canonical string) map[string]string {
	if route := sh.registry.GetByCanonical(canonical); route != nil {
		urls := make(map[string]string)
		for lang, path := range route.Paths {
			urls[lang] = sh.deployURL + path
		}
		return urls
	}
	return nil
}

// GetLocalePath returns the URL path for a canonical path and language.
func (sh *SEOHelpers) GetLocalePath(canonical string, lang string) string {
	if route := sh.registry.GetByCanonical(canonical); route != nil {
		if path, exists := route.Paths[lang]; exists {
			return path
		}
	}
	return "/"
}

// SEOFunctions holds SEO-related template functions.
// This struct is used to pass SEO functions to the template renderer.
type SEOFunctions struct {
	CanonicalURL   func(canonical, lang string) string
	AlternateLinks func(canonical string) template.HTML
	AlternateURLs  func(canonical string) map[string]string
	LocalePath     func(canonical, lang string) string
}

// ToTemplateFunctions converts SEOHelpers to a SEOFunctions struct.
func (sh *SEOHelpers) ToTemplateFunctions() *SEOFunctions {
	return &SEOFunctions{
		CanonicalURL:   sh.GetCanonicalURL,
		AlternateLinks: sh.GetAlternateLinks,
		AlternateURLs:  sh.GetAlternateURLs,
		LocalePath:     sh.GetLocalePath,
	}
}
