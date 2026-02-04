# Templates

Statigo uses Go's `html/template` package with embedded filesystems and custom template functions.

## Template Structure

```
templates/
├── layouts/
│   └── base.html          # Base layout
├── pages/
│   ├── index.html         # Home page
│   ├── about.html         # About page
│   └── notfound.html      # 404 page
└── partials/
    ├── header.html        # Site header
    └── footer.html        # Site footer
```

## Base Layout

The base layout defines the HTML structure:

```html
{{define "base"}}
<!DOCTYPE html>
<html lang="{{.Lang}}">
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <title>{{block "title" .}}{{.Title}}{{end}}</title>

    {{if .Canonical}}
    <link rel="canonical" href="{{canonicalURL .Canonical .Lang}}" />
    {{alternateLinks .Canonical}}
    {{end}}
</head>
<body>
    {{template "header.html" .}}
    <main>{{block "main" .}}{{end}}</main>
    {{template "footer.html" .}}
</body>
</html>
{{end}}
```

## Page Templates

Pages extend the base layout:

```html
{{define "title"}}{{t .Lang "pages.about.title"}}{{end}}
{{template "base" .}}

{{define "main"}}
<section class="page-content">
    <h1>{{t .Lang "pages.about.heading"}}</h1>
    <p>{{t .Lang "pages.about.body"}}</p>
</section>
{{end}}
```

## Template Blocks

### Defining Blocks

```html
{{block "title" .}}Default Title{{end}}
```

### Overriding Blocks

```html
{{define "title"}}My Custom Title{{end}}
```

## Partials

Include partial templates:

```html
{{template "header.html" .}}
{{template "footer.html" .}}
```

The `.` passes all data to the partial.

## Template Functions

### SEO Functions

```html
{{canonicalURL .Canonical .Lang}}
{{alternateLinks .Canonical}}
{{alternateURLs .Canonical}}
{{pathForLanguage .Canonical "tr"}}
```

### Translation Function

```html
{{t .Lang "pages.home.title"}}
{{t .Lang "pages.home.description" "Default description"}}
```

### Utility Functions

```html
{{add 1 2}}           → 3
{{sub 5 2}}           → 3
{{div 10 2}}          → 5
{{mod 10 3}}          → 1
{{until 5}}           → [0, 1, 2, 3, 4]

{{slugify "Hello World"}}  → "hello-world"
{{formatDate .Date}}       → "2024-01-15"
{{formatDateTime .Time}}    → "2024-01-15 10:30"

{{safeHTML "<b>Bold</b>"}}     → <b>Bold</b> (unescaped)
{{safeURL "http://example.com"}} → http://example.com
{{prettyJson .Data}}
```

## Passing Data to Templates

From handlers:

```go
data := map[string]interface{}{
    "Lang":      "en",
    "Canonical": "/about",
    "Title":     "About Us",
    "Meta": map[string]string{
        "description": "Learn about our company",
    },
    "Content": map[string]string{
        "heading": "About Us",
        "body":    "Our story...",
    },
}

renderer.Render(w, "about.html", data)
```

In templates:

```html
<h1>{{.Content.heading}}</h1>
<p>{{.Content.body}}</p>
```

## Context Variables

Available in all templates:

| Variable | Type | Description |
|----------|------|-------------|
| `.Lang` | string | Current language code |
| `.Canonical` | string | Canonical path |
| `.Title` | string | Page title |
| `.Meta` | map | Meta information |
| `.Layout` | interface{} | Layout data |
| `.Data` | interface{} | Custom data |

## Conditional Rendering

### If/Else

```html
{{if .User}}
    <p>Welcome, {{.User.Name}}</p>
{{else}}
    <p>Please log in</p>
{{end}}
```

### With

```html
{{with .User}}
    <p>{{.Name}}</p>  <!-- . refers to .User -->
{{end}}
```

### Range

```html
{{range .Items}}
    <p>{{.Title}}</p>
{{end}}
```

With index:

```html
{{range $index, $item := .Items}}
    <p>{{$index}}: {{$item.Title}}</p>
{{end}}
```

## Loops

### Simple Range

```html
<ul>
{{range .Items}}
    <li>{{.}}</li>
{{end}}
</ul>
```

### With Key-Value

```html
{{range $key, $value := .Map}}
    <p>{{$key}}: {{$value}}</p>
{{end}}
```

## Escaping

Go templates auto-escape HTML by default.

### To Escape

```html
{{.Content}}  <!-- Escaped -->
```

### To Not Escape (use carefully!)

```html
{{safeHTML .Content}}  <!-- Not escaped -->
```

## Custom Template Functions

Add custom functions in `templates/functions.go`:

```go
func CustomFunction(value string) string {
    return strings.ToUpper(value)
}

// Register in template.FuncMap
funcMap := template.FuncMap{
    "custom": CustomFunction,
    // ...
}
```

Use in templates:

```html
{{custom "hello"}}  → HELLO
```

## Layout Inheritance

### Extend Base

```html
{{template "base" .}}
```

### Override Blocks

```html
{{define "title"}}My Title{{end}}
{{define "main"}}My Content{{end}}
{{template "base" .}}
```

## Example: Complete Page Template

```html
{{define "title"}}{{t .Lang "pages.services.title"}} | {{t .Lang "branding.name"}}{{end}}
{{template "base" .}}

{{define "main"}}
<section class="services-page">
    <div class="container">
        <h1>{{t .Lang "pages.services.heading"}}</h1>
        <p>{{t .Lang "pages.services.subtitle"}}</p>

        <div class="services-grid">
            {{range .Services}}
            <article class="service-card">
                <img src="{{.Icon}}" alt="{{.Name}}" />
                <h3>{{.Name}}</h3>
                <p>{{.Description}}</p>
                <a href="{{.URL}}">{{t $.Lang "common.learn_more"}}</a>
            </article>
            {{end}}
        </div>
    </div>
</section>
{{end}}
```

## Debugging Templates

### Missing Variables

Check for missing keys:

```html
{{if .Title}}
    <title>{{.Title}}</title>
{{else}}
    <title>Default Title</title>
{{end}}
```

### Pretty Print JSON

```html
<pre>{{prettyJson .}}</pre>
```

### Comments

```html
{{/* This is a comment */}}
```

## Best Practices

1. **Always escape** user input
2. **Use partials** for reusable components
3. **Keep templates simple** - complex logic in handlers
4. **Use translation keys** instead of hardcoded text
5. **Define blocks** in base layout for flexibility
6. **Pass data consistently** - use same keys across templates

## Troubleshooting

### Template Not Found

1. Check file exists in `templates/` directory
2. Verify correct path in handler
3. Ensure templates FS is embedded

### Variables Not Showing

1. Check variable name matches (case-sensitive)
2. Verify data is passed to Render()
3. Use `prettyJson .` to debug

### Translation Not Working

1. Check translation key exists
2. Verify language code is correct
3. Ensure i18n is passed to renderer
