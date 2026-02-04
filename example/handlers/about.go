package handlers

import (
	"net/http"

	"statigo/framework/cache"
	"statigo/framework/middleware"
	"statigo/framework/router"
	"statigo/framework/templates"
)

// AboutHandler handles the about page.
type AboutHandler struct {
	renderer     *templates.Renderer
	cacheManager *cache.Manager
}

// NewAboutHandler creates a new about handler.
func NewAboutHandler(renderer *templates.Renderer, cacheManager *cache.Manager) *AboutHandler {
	return &AboutHandler{
		renderer:     renderer,
		cacheManager: cacheManager,
	}
}

// ServeHTTP handles the about page request.
func (h *AboutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
		"Title":     h.renderer.GetTranslation(lang, "pages.about.title"),
		"Meta": map[string]string{
			"description": h.renderer.GetTranslation(lang, "pages.about.description"),
		},
		"Content": map[string]string{
			"heading": h.renderer.GetTranslation(lang, "pages.about.heading"),
			"body":    h.renderer.GetTranslation(lang, "pages.about.body"),
		},
	}

	h.renderer.Render(w, "about.html", data)
}
