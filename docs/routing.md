# Routing

Statigo uses a configuration-driven routing system with built-in multi-language support and SEO features.

## Route Definition

Routes are defined in `config/routes.json`:

```json
{
  "routes": [
    {
      "canonical": "/about",
      "paths": {
        "en": "/en/about",
        "tr": "/tr/hakkinda"
      },
      "strategy": "static",
      "template": "about.html",
      "handler": "about",
      "title": "pages.about.title"
    }
  ]
}
```

### Route Properties

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `canonical` | string | Yes | Canonical path identifier |
| `paths` | object | Yes | Language-specific URL paths |
| `strategy` | string | No | Cache strategy (static/incremental/dynamic/immutable) |
| `template` | string | Handler-dependent | Template filename |
| `handler` | string | No | Handler name to use |
| `title` | string | No | Translation key for page title |

## Canonical Paths

The canonical path is the internal identifier for a route. It's used for:

- Cache key generation
- SEO canonical URLs
- Translation lookups

```go
// Get canonical path from request context
canonical := router.GetCanonicalPath(r.Context())
// canonical = "/about"
```

## Language Paths

Each route defines paths for supported languages:

```json
{
  "canonical": "/products/{id}",
  "paths": {
    "en": "/en/products/{id}",
    "tr": "/tr/urunler/{id}"
  }
}
```

Path parameters (like `{id}`) are preserved and passed to handlers.

## Handlers

### Built-in Handlers

#### `content` Handler

Renders a template with default data:

```json
{
  "canonical": "/",
  "paths": { "en": "/en" },
  "template": "index.html",
  "handler": "content"
}
```

The template receives:
```go
{
    "Lang": "en",
    "Canonical": "/",
    "Data": {},
    "Layout": {...},
    "WebAppURL": "https://..."
}
```

### Custom Handlers

Register custom handlers in `main.go`:

```go
customHandlers := map[string]http.HandlerFunc{
    "index": myIndexHandler.ServeHTTP,
    "about": myAboutHandler.ServeHTTP,
}

router.LoadRoutesFromJSON(
    configFS,
    "routes.json",
    registry,
    renderer,
    customHandlers,
    logger,
)
```

Example handler:

```go
type IndexHandler struct {
    renderer     *templates.Renderer
    cacheManager *cache.Manager
}

func (h *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    lang := middleware.GetLanguage(r.Context())
    canonical := router.GetCanonicalPath(r.Context())

    data := map[string]interface{}{
        "Lang":      lang,
        "Canonical": canonical,
        "Title":     h.renderer.GetTranslation(lang, "pages.home.title"),
    }

    h.renderer.Render(w, "index.html", data)
}
```

## Route Registration

Routes are automatically registered with the router:

```go
routeRegistry.RegisterRoutes(r, func(h http.Handler) http.Handler {
    // Optional wrapper middleware
    return h
})
```

## Dynamic Routes

Use path parameters for dynamic routes:

```json
{
  "canonical": "/blog/{slug}",
  "paths": {
    "en": "/en/blog/{slug}",
    "tr": "/tr/blog/{slug}"
  },
  "strategy": "incremental",
  "template": "blog.html",
  "handler": "blog"
}
```

Access parameters in handlers:

```go
// Using chi URL params
slug := chi.URLParam(r, "slug")
```

## Cache Strategies

| Strategy | Description | Use Case |
|----------|-------------|----------|
| `static` | Cached indefinitely, no expiration | Pages, evergreen content |
| `immutable` | Same as static, for truly immutable content | Assets, archived content |
| `incremental` | Time-based revalidation | Lists, indexes, feeds |
| `dynamic` | Not cached | User-specific, real-time data |

## SEO Features

### Canonical URLs

Automatically generated from canonical path:

```html
<link rel="canonical" href="https://example.com/about" />
```

### Hreflang Links

Alternate language links for SEO:

```html
<link rel="alternate" hreflang="en" href="https://example.com/en/about" />
<link rel="alternate" hreflang="tr" href="https://example.com/tr/hakkinda" />
<link rel="alternate" hreflang="x-default" href="https://example.com/en/about" />
```

### Template Functions

```html
{{canonicalURL .Canonical .Lang}}
{{alternateLinks .Canonical}}
{{alternateURLs .Canonical}}
```

## Route Lookup

### By Path

```go
route := registry.GetByPath("/en/about")
// Returns RouteDefinition for /about canonical
```

### By Canonical

```go
route := registry.GetByCanonical("/about")
// Returns RouteDefinition
```

### Get Path for Language

```go
seoHelpers := router.NewSEOHelpers(registry, baseURL)
path := seoHelpers.GetPathForLanguage("/about", "tr")
// Returns "/tr/hakkinda"
```

## Advanced Patterns

### Route Groups

Organize related routes by canonical prefix:

```json
{
  "canonical": "/blog",
  "paths": {"en": "/en/blog"},
  "strategy": "incremental"
},
{
  "canonical": "/blog/{slug}",
  "paths": {"en": "/en/blog/{slug}"},
  "strategy": "static"
}
```

### Nested Routes

Use canonical hierarchy for organization:

```json
{
  "canonical": "/products/category/{cat}",
  "paths": {"en": "/en/products/category/{cat}"}
},
{
  "canonical": "/products/category/{cat}/{id}",
  "paths": {"en": "/en/products/{cat}/{id}"}
}
```
