// Package handlers provides example HTTP handlers demonstrating Statigo framework usage.
package handlers

import (
	"net/http"

	"statigo/framework/cache"
	"statigo/framework/middleware"
	"statigo/framework/router"
	"statigo/framework/templates"
)

// IndexHandler handles the home page.
type IndexHandler struct {
	renderer     *templates.Renderer
	cacheManager *cache.Manager
	registry     *router.Registry
}

// NewIndexHandler creates a new index handler.
func NewIndexHandler(renderer *templates.Renderer, cacheManager *cache.Manager, registry *router.Registry) *IndexHandler {
	return &IndexHandler{
		renderer:     renderer,
		cacheManager: cacheManager,
		registry:     registry,
	}
}

// ServeHTTP handles the home page request.
func (h *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lang := middleware.GetLanguage(r.Context())
	canonical := router.GetCanonicalPath(r.Context())

	// Try to serve from cache
	cacheKey := cache.GetCacheKey(canonical, lang, nil)
	if entry, found := h.cacheManager.Get(cacheKey); found && !entry.IsStale() {
		content, err := cache.GetDecompressedContent(entry)
		if err == nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("X-Cache", "HIT")
			w.Header().Set("ETag", entry.ETag)
			w.Write(content)
			return
		}
	}

	// Build page data
	data := map[string]interface{}{
		"Lang":      lang,
		"Canonical": canonical,
		"Title":     h.renderer.GetTranslation(lang, "pages.home.title"),
		"Meta": map[string]string{
			"description": h.renderer.GetTranslation(lang, "pages.home.description"),
		},
		"Content": map[string]string{
			"heading":    h.renderer.GetTranslation(lang, "pages.home.heading"),
			"subheading": h.renderer.GetTranslation(lang, "pages.home.subheading"),
		},
	}

	h.renderer.Render(w, "index.html", data)
}
