// Package handlers provides example HTTP handlers demonstrating Statigo framework usage.
package handlers

import (
	"encoding/json"
	"net/http"
	"sync/atomic"

	"statigo/framework/cache"
	"statigo/framework/middleware"
	"statigo/framework/router"
	"statigo/framework/templates"
)

// Global counter state
var counter int64

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

	// Handle counter increment (API endpoint)
	if r.Method == http.MethodPost {
		newCount := atomic.AddInt64(&counter, 1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int64{"count": newCount})
		return
	}

	// Get current counter value
	currentCount := atomic.LoadInt64(&counter)

	// Build page data
	data := map[string]any{
		"Lang":      lang,
		"Canonical": canonical,
		"Title":     "StatiGo - Static Speed With Dynamic Content",
		"Meta": map[string]string{
			"description": "StatiGo - Static Speed With Dynamic Content",
		},
		"Counter": currentCount,
	}

	h.renderer.Render(w, "index.html", data)
}
