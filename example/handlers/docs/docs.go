// Package docs provides documentation serving with markdown rendering.
package docs

import (
	"io/fs"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"

	"statigo/framework/middleware"
	"statigo/framework/templates"
)

// Handler handles documentation page requests.
type Handler struct {
	renderer     *templates.Renderer
	seoHelpers   interface{} // Can be *router.SEOHelpers or just the LocalePath function
	docFS        fs.FS
	markdown     goldmark.Markdown
	logger       *slog.Logger
	baseURL      string
}

// NewHandler creates a new documentation handler.
func NewHandler(renderer *templates.Renderer, seoHelpers interface{}, docFS fs.FS, logger *slog.Logger, baseURL string) *Handler {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.DefinitionList,
			extension.Footnote,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
		),
	)

	return &Handler{
		renderer:   renderer,
		seoHelpers: seoHelpers,
		docFS:      docFS,
		markdown:   md,
		logger:     logger,
		baseURL:    baseURL,
	}
}

// Doc represents a documentation page.
type Doc struct {
	Title       string
	Content     string
	TOC         []TOCItem
	CurrentSlug string
	Sidebar     []SidebarItem
}

// TOCItem represents a table of contents item.
type TOCItem struct {
	ID    string
	Title string
	Level int
}

// SidebarItem represents a sidebar item.
type SidebarItem struct {
	Title string
	Slug  string
	Level int
}

// ServeHTTP handles documentation page requests.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lang := middleware.GetLanguage(r.Context())

	// Get the doc slug from the route
	slug := chi.URLParam(r, "slug")

	// Default to overview if no slug
	if slug == "" || slug == "/" {
		slug = "overview"
	}

	// Read markdown file
	content, err := fs.ReadFile(h.docFS, slug+".md")
	if err != nil {
		h.logger.Warn("Doc not found", "slug", slug, "error", err)
		h.render404(w, r, lang)
		return
	}

	// Convert markdown to HTML
	var htmlBuf strings.Builder
	h.markdown.Convert(content, &htmlBuf)

	// Parse title from first heading
	title := h.extractTitle(string(content))

	// Generate table of contents
	toc := h.generateTOC(string(content))

	// Generate sidebar
	sidebar := h.generateSidebar()

	// Build canonical path
	canonical := "/docs/" + slug

	doc := Doc{
		Title:       title,
		Content:     htmlBuf.String(),
		TOC:         toc,
		CurrentSlug: slug,
		Sidebar:     sidebar,
	}

	data := map[string]interface{}{
		"Doc":       doc,
		"Title":     title + " - Documentation",
		"BaseURL":   h.baseURL,
		"Lang":      lang,
		"Canonical": canonical,
	}

	h.renderer.Render(w, "docs.html", data)
}

// extractTitle extracts the title from markdown content.
func (h *Handler) extractTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimPrefix(line, "# ")
		}
		if strings.HasPrefix(line, "## ") {
			return strings.TrimPrefix(line, "## ")
		}
	}
	return "Documentation"
}

// generateTOC generates a table of contents from markdown headings.
func (h *Handler) generateTOC(content string) []TOCItem {
	var toc []TOCItem
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "##") {
			level := 0
			if strings.HasPrefix(line, "###") {
				level = 1
			} else if strings.HasPrefix(line, "####") {
				level = 2
			}

			title := strings.TrimLeft(line, "#")
			title = strings.TrimSpace(title)
			id := h.slugify(title)

			// Limit TOC depth
			if level <= 2 {
				toc = append(toc, TOCItem{
					ID:    id,
					Title: title,
					Level: level,
				})
			}
		}
	}

	return toc
}

// generateSidebar generates sidebar navigation from all docs.
func (h *Handler) generateSidebar() []SidebarItem {
	_, _ = fs.ReadDir(h.docFS, ".")
	sidebar := []SidebarItem{
		{Title: "Overview", Slug: "overview", Level: 0},
		{Title: "Getting Started", Slug: "getting-started", Level: 1},
		{Title: "Routing", Slug: "routing", Level: 1},
		{Title: "Middleware", Slug: "middleware", Level: 1},
		{Title: "Caching", Slug: "caching", Level: 1},
		{Title: "i18n", Slug: "i18n", Level: 1},
		{Title: "Templates", Slug: "templates", Level: 1},
		{Title: "Configuration", Slug: "configuration", Level: 1},
		{Title: "CLI", Slug: "cli", Level: 1},
	}

	return sidebar
}

// slugify converts a string to a URL-friendly slug.
func (h *Handler) slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "?", "")
	s = strings.ReplaceAll(s, "!", "")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, ":", "")
	s = strings.ReplaceAll(s, ";", "")
	s = strings.ReplaceAll(s, "/", "")
	s = strings.ReplaceAll(s, "\\", "")
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	s = strings.ReplaceAll(s, "[", "")
	s = strings.ReplaceAll(s, "]", "")
	s = strings.ReplaceAll(s, "{", "")
	s = strings.ReplaceAll(s, "}", "")
	return s
}

// render404 renders a 404 page for documentation.
func (h *Handler) render404(w http.ResponseWriter, r *http.Request, lang string) {
	w.WriteHeader(http.StatusNotFound)

	data := map[string]interface{}{
		"Doc": Doc{
			Title:   "Documentation Not Found",
			Content: "<p>The documentation page you're looking for doesn't exist.</p>",
			Sidebar: h.generateSidebar(),
		},
		"Title":     "Documentation Not Found",
		"Lang":      lang,
		"Canonical": "/docs/404",
	}

	h.renderer.Render(w, "docs.html", data)
}
