package handlers

import (
	"net/http"

	"statigo/framework/middleware"
	"statigo/framework/templates"
)

// NotFoundHandler handles 404 errors.
type NotFoundHandler struct {
	renderer *templates.Renderer
}

// NewNotFoundHandler creates a new 404 handler.
func NewNotFoundHandler(renderer *templates.Renderer) *NotFoundHandler {
	return &NotFoundHandler{
		renderer: renderer,
	}
}

// ServeHTTP handles the 404 page request.
func (h *NotFoundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lang := middleware.GetLanguage(r.Context())
	if lang == "" {
		lang = "en"
	}

	w.WriteHeader(http.StatusNotFound)

	data := map[string]interface{}{
		"Lang":  lang,
		"Title": h.renderer.GetTranslation(lang, "pages.notfound.title"),
		"Content": map[string]string{
			"heading": h.renderer.GetTranslation(lang, "pages.notfound.heading"),
			"message": h.renderer.GetTranslation(lang, "pages.notfound.message"),
			"action":  h.renderer.GetTranslation(lang, "pages.notfound.action"),
		},
	}

	h.renderer.Render(w, "notfound.html", data)
}
