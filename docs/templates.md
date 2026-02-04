# Templates

Statigo uses Go's `html/template` package with custom functions and embedded filesystems for fast, secure HTML rendering.

## Template Structure

```
templates/
├── layouts/
│   └── base.html          # Base layout
├── pages/
│   ├── index.html         # Home page
│   ├── about.html         # About page
│   └── docs.html          # Docs page
└── partials/
    ├── header.html        # Header component
    └── footer.html        # Footer component
```

## Base Layout

Create `templates/layouts/base.html`:

```html
<!DOCTYPE html>
<html lang="{{.Lang}}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{if .Title}}{{.Title}} - {{end}}{{t "site.name"}}</title>

    <!-- SEO -->
    <link rel="canonical" href="{{canonicalURL .Canonical .Lang}}">
    {{alternateLinks .Canonical}}

    <!-- Styles -->
    <link rel="stylesheet" href="/static/styles/main.css">
</head>
<body>
    {{template "header" .}}

    <main>
        {{block "main" .}}{{end}}
    </main>

    {{template "footer" .}}
</body>
</html>
```

## Page Templates

Each page extends the base layout:

`templates/pages/index.html`:
```html
{{define "title"}}Home{{end}}
{{template "base" .}}

{{define "main"}}
<div class="hero">
    <h1>{{t "pages.home.title"}}</h1>
    <p>{{t "pages.home.subtitle"}}</p>
</div>
{{end}}
```

## Partials

Reusable components in `templates/partials/`:

`templates/partials/header.html`:
```html
{{define "header"}}
<header class="site-header">
    <nav>
        <a href="{{localePath "/" .Lang}}">{{t "site.name"}}</a>
        <a href="{{localePath "/about" .Lang}}">{{t "nav.about"}}</a>
        <a href="{{localePath "/docs" .Lang}}">{{t "nav.docs"}}</a>
    </nav>

    <div class="lang-switcher">
        <a href="{{localePath .Canonical "en"}}"{{if eq .Lang "en"}} class="active"{{end}}>EN</a>
        <a href="{{localePath .Canonical "tr"}}"{{if eq .Lang "tr"}} class="active"{{end}}>TR</a>
    </div>
</header>
{{end}}
```

`templates/partials/footer.html`:
```html
{{define "footer"}}
<footer class="site-footer">
    <p>&copy; {{t "site.copyright"}}</p>
</footer>
{{end}}
```

## Template Functions

### SEO Functions

```html
<!-- Canonical URL -->
<link rel="canonical" href="{{canonicalURL "/about" .Lang}}">

<!-- Alternate links (hreflang) -->
{{alternateLinks "/about"}}

<!-- Locale-aware path -->
<a href="{{localePath "/about" .Lang}}">About</a>
```

### Translation Function

```html
{{t "pages.home.title"}}
{{t "nav.home"}}
{{t "errors.not_found"}}
```

### Math Functions

```html
{{add 1 2}}        <!-- 3 -->
{{sub 10 3}}       <!-- 7 -->
{{div 100 5}}      <!-- 20 -->
{{mod 10 3}}       <!-- 1 -->
```

### String Functions

```html
{{slugify "Hello World"}}    <!-- hello-world -->
{{safeHTML "<p>Raw HTML</p>"}}  <!-- Renders as HTML -->
{{safeURL "http://example.com"}} <!-- Escapes for URL -->
```

### Date Functions

```html
{{formatDate .Date "2006-01-02"}}
{{formatDateTime .Date "2006-01-02T15:04:05Z"}}
```

### Utility Functions

```html
{{dict "key" "value" "key2" "value2"}}
{{set . "NewField" "value"}}
{{until 5}}  <!-- Returns slice [0,1,2,3,4] -->
```

### Price/Currency Functions

```html
{{formatPrice 19.99 "USD"}}          <!-- $19.99 -->
{{currencySymbol "USD"}}              <!-- $ -->
{{priceWhole 19.99}}                  <!-- 19 -->
{{priceDecimal 19.99}}                <!-- 99 -->
```

## Passing Data to Templates

In your handler:

```go
data := map[string]interface{}{
    "Title":    "Page Title",
    "Lang":     "en",
    "Canonical": "/about",
    "User": map[string]string{
        "Name":  "John",
        "Email": "john@example.com",
    },
    "Posts": []Post{
        {Title: "First Post", Content: "..."},
        {Title: "Second Post", Content: "..."},
    },
}
renderer.Render(w, "about.html", data)
```

In your template:

```html
<h1>{{.Title}}</h1>
<p>Welcome, {{.User.Name}}</p>

{{range .Posts}}
    <article>
        <h2>{{.Title}}</h2>
        <p>{{.Content}}</p>
    </article>
{{end}}
```

## Context Variables

Statigo automatically provides these context variables:

| Variable | Description |
|----------|-------------|
| `{{.Lang}}` | Current language code |
| `{{.Canonical}}` | Canonical path |
| `{{.Title}}` | Page title |
| `{{.Env}}` | Environment variables (e.g., `{{.Env.GTM_ID}}`) |

## Conditional Rendering

```html
{{if .User}}
    <p>Welcome, {{.User.Name}}</p>
{{else}}
    <p>Please log in</p>
{{end}}

{{if eq .Lang "tr"}}
    <p>Türkçe içerik</p>
{{else if eq .Lang "en"}}
    <p>English content</p>
{{end}}
```

## Loops

```html
{{range $index, $item := .Items}}
    <div class="item-{{$index}}">
        {{$item.Name}}
    </div>
{{end}}
```

## Safe HTML

To render raw HTML (use with caution):

```go
data := map[string]interface{}{
    "Content": "<p><strong>Bold text</strong></p>",
}
```

```html
{{safeHTML .Content}}
<!-- Renders: <p><strong>Bold text</strong></p> -->
```

## Template Inheritance

Use `block` for default content that can be overridden:

In `base.html`:
```html
{{block "main" .}}
    <p>Default content</p>
{{end}}
```

In `index.html`:
```html
{{define "main"}}
    <p>Custom content for index page</p>
{{end}}
```

## Environment Variables

Access environment variables in templates:

```html
{{if .Env.GTM_ID}}
    <!-- Google Tag Manager -->
    <script>(function(w,d,s,l,i){...})</script>
{{end}}
```

Set in `.env`:
```
GTM_ID=GTM-XXXXX
```

## Best Practices

1. **Always escape user input** - Use `html/template` auto-escaping
2. **Use `safeHTML` sparingly** - Only for trusted content
3. **Keep layouts simple** - Don't nest too deeply
4. **Use partials for repetition** - Headers, footers, cards
5. **Group translations** - Use dot notation in i18n keys
6. **Validate canonical URLs** - Ensure proper SEO structure
