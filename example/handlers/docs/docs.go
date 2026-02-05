// Package docs provides documentation serving with markdown rendering.
package docs

import (
	"io/fs"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"

	fwi18n "statigo/framework/i18n"
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
	i18n         *fwi18n.I18n
}

// NewHandler creates a new documentation handler.
func NewHandler(renderer *templates.Renderer, seoHelpers interface{}, docFS fs.FS, i18nInstance *fwi18n.I18n, logger *slog.Logger, baseURL string) *Handler {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.DefinitionList,
			extension.Footnote,
			// Use custom AST transformer to generate heading IDs with Turkish support
			&turkishHeadingIDExtension{},
		),
		goldmark.WithParserOptions(
			// Don't use AutoHeadingID - we use our own
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
		i18n:       i18nInstance,
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

	// Extract slug from URL path
	// Path format: /{lang}/docs or /{lang}/docs/{slug}
	path := r.URL.Path

	// Remove language prefix (e.g., /en/ or /tr/)
	langPrefix := "/" + lang + "/"
	path = strings.TrimPrefix(path, langPrefix)

	// Remove /docs or /docs/ prefix to get the slug
	path = strings.TrimPrefix(path, "docs")
	path = strings.TrimPrefix(path, "/")

	slug := path

	// Default to overview if no slug
	if slug == "" || slug == "/" {
		slug = "overview"
	}

	// Try language-specific file first, then fall back to English
	// E.g., tr/overview.md, then en/overview.md
	content, err := fs.ReadFile(h.docFS, lang+"/"+slug+".md")
	if err != nil {
		// Fall back to English version
		content, err = fs.ReadFile(h.docFS, "en/"+slug+".md")
		if err != nil {
			h.logger.Warn("Doc not found", "slug", slug, "error", err)
			h.render404(w, r, lang)
			return
		}
	}

	// Convert markdown to HTML
	var htmlBuf strings.Builder
	h.markdown.Convert(content, &htmlBuf)

	// Post-process HTML to fix Turkish character IDs
	htmlContent := h.fixTurkishIDs(htmlBuf.String())

	// Parse title from first heading
	title := h.extractTitle(string(content))

	// Generate table of contents
	toc := h.generateTOC(string(content))

	// Generate sidebar
	sidebar := h.generateSidebar(lang)

	// Build canonical path
	canonical := "/docs/" + slug

	doc := Doc{
		Title:       title,
		Content:     htmlContent,
		TOC:         toc,
		CurrentSlug: slug,
		Sidebar:     sidebar,
	}

	// Get meta description for this doc
	descKey := "docs." + strings.ReplaceAll(slug, "-", "_") + ".description"
	description := h.i18n.Get(lang, descKey)
	if description == "" {
		// Fallback to general docs description
		description = h.i18n.Get(lang, "docs.description")
	}

	data := map[string]interface{}{
		"Doc":       doc,
		"Title":     title + " - Documentation",
		"BaseURL":   h.baseURL,
		"Lang":      lang,
		"Canonical": canonical,
		"Meta": map[string]string{
			"description": description,
		},
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

	// Parse markdown to get the AST
	parser := h.markdown.Parser()
	document := parser.Parse(text.NewReader([]byte(content)))

	// Walk through the AST to find headings
	err := ast.Walk(document, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if heading, ok := node.(*ast.Heading); ok {
			// Only include h2 and h3 in TOC
			if heading.Level < 2 || heading.Level > 4 {
				return ast.WalkSkipChildren, nil
			}

			// Get the heading text from lines
			var title strings.Builder
			lines := heading.Lines()
			contentBytes := []byte(content)
			for i := 0; i < lines.Len(); i++ {
				segment := lines.At(i)
				title.Write(segment.Value(contentBytes))
			}

			// Get the ID from goldmark (may have stripped Turkish chars)
			headingID := ""
			if id, ok := heading.AttributeString("id"); ok {
				if idStr, ok := id.(string); ok {
					headingID = idStr
				}
			}

			// Fallback to generating from title if no ID
			if headingID == "" {
				headingID = slugifyForTOC(title.String())
			}

			tocLevel := heading.Level - 2 // 0 for h2, 1 for h3, 2 for h4

			toc = append(toc, TOCItem{
				ID:    headingID,
				Title: strings.TrimSpace(title.String()),
				Level: tocLevel,
			})
		}

		return ast.WalkContinue, nil
	})

	if err != nil {
		h.logger.Warn("Error generating TOC", "error", err)
	}

	return toc
}

// generateSidebar generates sidebar navigation from all docs.
func (h *Handler) generateSidebar(lang string) []SidebarItem {
	_, _ = fs.ReadDir(h.docFS, ".")

	// Translation keys for sidebar items
	items := []struct {
		key   string
		slug  string
		level int
	}{
		{"docs.overview", "overview", 0},
		{"docs.getting_started", "getting-started", 1},
		{"docs.routing", "routing", 1},
		{"docs.middleware", "middleware", 1},
		{"docs.caching", "caching", 1},
		{"docs.i18n", "i18n", 1},
		{"docs.templates", "templates", 1},
		{"docs.configuration", "configuration", 1},
		{"docs.cli", "cli", 1},
	}

	sidebar := make([]SidebarItem, len(items))
	for i, item := range items {
		// Use .title suffix since the key points to an object with title/description
		titleKey := item.key + ".title"
		title := h.i18n.Get(lang, titleKey)
		if title == "" || title == titleKey {
			// Fallback to slug if translation missing
			title = item.slug
		}
		sidebar[i] = SidebarItem{
			Title: title,
			Slug:  item.slug,
			Level: item.level,
		}
	}

	return sidebar
}

// slugify converts a string to a URL-friendly slug.
func (h *Handler) slugify(s string) string {
	return slugifyForTOC(s)
}

// slugifyForTOC converts Turkish characters to ASCII and creates a URL-friendly slug.
func slugifyForTOC(s string) string {
	s = strings.ToLower(s)

	// Turkish character mappings - convert to ASCII
	turkishMap := map[rune]string{
		'ç': "c", 'Ç': "c",
		'ğ': "g", 'Ğ': "g",
		'ı': "i", 'İ': "i",
		'ş': "s", 'Ş': "s",
		'ö': "o", 'Ö': "o",
		'ü': "u", 'Ü': "u",
	}

	var result strings.Builder
	for _, r := range s {
		if replacement, ok := turkishMap[r]; ok {
			result.WriteString(replacement)
		} else if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
		} else if r == ' ' || r == '-' {
			result.WriteRune('-')
		}
		// Skip other characters
	}

	return result.String()
}

// turkishHeadingIDExtension is a goldmark extension that adds proper heading IDs for Turkish.
type turkishHeadingIDExtension struct{}

func (e *turkishHeadingIDExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithASTTransformers(util.Prioritized(&turkishHeadingIDTransformer{}, 100)))
}

// turkishHeadingIDTransformer adds id attributes to headings using Turkish-aware slugify.
type turkishHeadingIDTransformer struct{}

func (t *turkishHeadingIDTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	// We need to keep track of which headings we've already processed
	processed := make(map[ast.Node]bool)

	ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		if heading, ok := n.(*ast.Heading); ok {
			if processed[heading] {
				return ast.WalkContinue, nil
			}
			processed[heading] = true

			// Get heading text
			var title strings.Builder
			lines := heading.Lines()
			for i := 0; i < lines.Len(); i++ {
				segment := lines.At(i)
				title.Write(segment.Value([]byte(reader.Source())))
			}

			// Generate ID using our slugify function
			headingID := slugifyForTOC(title.String())
			heading.SetAttribute([]byte("id"), headingID)
		}

		return ast.WalkContinue, nil
	})
}

// fixTurkishIDs post-processes HTML to fix heading IDs for Turkish characters.
// goldmark's AutoHeadingID strips non-ASCII characters, so "Çeviri" becomes "eviri".
// We convert these back to ASCII equivalents for consistency.
func (h *Handler) fixTurkishIDs(htmlContent string) string {
	// Find all id="xxx" attributes and fix Turkish character IDs
	re := regexp.MustCompile(`id="([^"]+)"`)

	return re.ReplaceAllStringFunc(htmlContent, func(match string) string {
		// Extract the ID value
		id := match[4 : len(match)-1]
		fixedID := slugifyForTOC(id)
		if fixedID != id {
			return `id="` + fixedID + `"`
		}
		return match
	})
}

// render404 renders a 404 page for documentation.
func (h *Handler) render404(w http.ResponseWriter, r *http.Request, lang string) {
	w.WriteHeader(http.StatusNotFound)

	data := map[string]interface{}{
		"Doc": Doc{
			Title:   h.i18n.Get(lang, "docs.not_found"),
			Content: "<p>" + h.i18n.Get(lang, "docs.not_found_message") + "</p>",
			Sidebar: h.generateSidebar(lang),
		},
		"Title":     h.i18n.Get(lang, "docs.not_found"),
		"Lang":      lang,
		"Canonical": "/docs/404",
		"Meta": map[string]string{
			"description": h.i18n.Get(lang, "docs.description"),
		},
	}

	h.renderer.Render(w, "docs.html", data)
}
