// Package templates provides HTML template rendering for the Statigo framework.
package templates

import (
	"bytes"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path"

	"statigo/framework/i18n"
	"statigo/framework/utils"
)

// Renderer handles HTML template rendering.
type Renderer struct {
	templates     *template.Template            // Base templates (layouts + partials)
	pageTemplates map[string]*template.Template // Per-page template instances
	i18n          *i18n.I18n
	minifier      *utils.Minifier
	logger        *slog.Logger
}

// SEOFunctions holds SEO-related template functions.
type SEOFunctions struct {
	CanonicalURL   func(canonical, lang string) string
	AlternateLinks func(canonical string) template.HTML
	AlternateURLs  func(canonical string) map[string]string
	LocalePath     func(canonical, lang string) string
}

// NewRenderer creates a new template renderer.
func NewRenderer(templatesFS fs.FS, i18nInstance *i18n.I18n, seoFuncs *SEOFunctions, logger *slog.Logger) (*Renderer, error) {
	minifier := utils.NewMinifier()
	funcMap := template.FuncMap{
		"prettyJson":     PrettyJson,
		"safeHTML":       SafeHTML,
		"safeURL":        SafeURL,
		"add":            Add,
		"sub":            Sub,
		"subFloat":       SubFloat,
		"div":            Div,
		"mod":            Mod,
		"until":          Until,
		"slugify":        Slugify,
		"formatDate":     FormatDate,
		"formatDateTime": FormatDateTime,
		"youtubeID":      YouTubeID,
		"currencySymbol": CurrencySymbol,
		"formatPrice":    FormatPrice,
		"priceWhole":     PriceWhole,
		"priceDecimal":   PriceDecimal,
		"dict":           Dict,
		"set":            Set,
		"hasDiscount":    HasDiscount,
		"t":              i18nInstance.GetRaw,
	}

	// Add SEO functions if provided
	if seoFuncs != nil {
		funcMap["canonicalURL"] = seoFuncs.CanonicalURL
		funcMap["alternateLinks"] = seoFuncs.AlternateLinks
		funcMap["alternateURLs"] = seoFuncs.AlternateURLs
		funcMap["localePath"] = seoFuncs.LocalePath
	} else {
		// Provide default no-op implementations
		funcMap["canonicalURL"] = func(canonical, lang string) string { return "" }
		funcMap["alternateLinks"] = func(canonical string) template.HTML { return "" }
		funcMap["alternateURLs"] = func(canonical string) map[string]string { return nil }
		funcMap["localePath"] = func(canonical, lang string) string { return "" }
	}

	templates := template.New("base").Funcs(funcMap)

	// Load base templates (optional - skip if no files match)
	baseMatches, err := fs.Glob(templatesFS, "*.html")
	if err != nil {
		return nil, err
	}
	if len(baseMatches) > 0 {
		if templates, err = templates.ParseFS(templatesFS, "*.html"); err != nil {
			return nil, err
		}
	}

	// Load layouts
	if err := loadTemplatesRecursivelyFromFS(templates, templatesFS, "layouts"); err != nil {
		return nil, err
	}

	// Load partials recursively from subdirectories
	if err := loadTemplatesRecursivelyFromFS(templates, templatesFS, "partials"); err != nil {
		return nil, err
	}

	// Load pages - each page gets its own template instance to avoid block conflicts
	pageFiles, err := fs.Glob(templatesFS, "pages/*.html")
	if err != nil {
		return nil, err
	}

	pageTemplates := make(map[string]*template.Template)
	for _, pageFile := range pageFiles {
		// Clone the base templates (layouts + partials)
		pageTemplate, err := templates.Clone()
		if err != nil {
			return nil, err
		}

		// Parse this specific page file into the cloned template
		if _, err := pageTemplate.ParseFS(templatesFS, pageFile); err != nil {
			return nil, err
		}

		// Store by filename (e.g., "index.html", "blog.html")
		pageName := path.Base(pageFile)
		pageTemplates[pageName] = pageTemplate
	}

	return &Renderer{
		templates:     templates,
		pageTemplates: pageTemplates,
		i18n:          i18nInstance,
		minifier:      minifier,
		logger:        logger,
	}, nil
}

// GetTranslation returns a translation for the given language and key.
func (r *Renderer) GetTranslation(lang, key string) string {
	if value := r.i18n.GetRaw(lang, key); value != nil {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return key // Fallback to key if translation not found
}

// enrichDataWithEnv adds environment variables to template data.
func (r *Renderer) enrichDataWithEnv(data interface{}) interface{} {
	// Convert data to map if it's already a map
	if dataMap, ok := data.(map[string]interface{}); ok {
		// Add Env if it doesn't already exist
		if _, exists := dataMap["Env"]; !exists {
			dataMap["Env"] = map[string]string{
				"GTM_ID": os.Getenv("GTM_ID"),
			}
		}

		return dataMap
	}

	// If data is not a map, wrap it
	return map[string]interface{}{
		"Data": data,
		"Env": map[string]string{
			"GTM_ID": os.Getenv("GTM_ID"),
		},
	}
}

// Render renders a template with the given data.
func (r *Renderer) Render(w http.ResponseWriter, templateName string, data interface{}) {
	var buf bytes.Buffer

	// Inject environment variables into template data
	enrichedData := r.enrichDataWithEnv(data)

	// Try to use page-specific template first
	var err error
	if pageTemplate, ok := r.pageTemplates[templateName]; ok {
		err = pageTemplate.ExecuteTemplate(&buf, templateName, enrichedData)
	} else {
		// Fallback to base templates for partials and other templates
		err = r.templates.ExecuteTemplate(&buf, templateName, enrichedData)
	}

	if err != nil {
		r.logger.Error("Error rendering template", "template", templateName, "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	minifiedHTML, err := r.minifier.MinifyString("text/html", buf.String())
	if err != nil {
		r.logger.Error("Error minifying template", "template", templateName, "error", err)
		// Fall back to unminified HTML
		w.Header().Set("Content-Type", "text/html")
		buf.WriteTo(w)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(minifiedHTML))
}

// loadTemplatesRecursivelyFromFS walks a directory in an fs.FS and loads all .html files as templates.
func loadTemplatesRecursivelyFromFS(tmpl *template.Template, fsys fs.FS, dir string) error {
	return fs.WalkDir(fsys, dir, func(filePath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if d.IsDir() {
			return nil
		}

		// Only process .html files
		if path.Ext(filePath) == ".html" {
			data, err := fs.ReadFile(fsys, filePath)
			if err != nil {
				return err
			}
			if _, err := tmpl.New(path.Base(filePath)).Parse(string(data)); err != nil {
				return err
			}
		}

		return nil
	})
}
