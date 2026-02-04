# Routing

Statigo's routing system provides multi-language URL mapping with SEO optimization.

## Route Configuration

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

### Route Fields

| Field | Type | Description |
|-------|------|-------------|
| `canonical` | string | Internal canonical path (used for lookups) |
| `paths` | object | Language-specific URL paths |
| `strategy` | string | Caching strategy: `static`, `incremental`, `dynamic`, `immutable` |
| `template` | string | Template file to render |
| `handler` | string | Handler name (registered in `customHandlers` map) |
| `title` | string | Page title (can be i18n key or literal) |

## Multi-Language Routing

### Defining Language-Specific Paths

Each route can have different URLs per language:

```json
{
  "canonical": "/features",
  "paths": {
    "en": "/en/features",
    "tr": "/tr/ozellikler",
    "de": "/de/funktionen"
  }
}
```

### Accessing Current Language

In your handlers:

```go
import "statigo/framework/middleware"

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    lang := middleware.GetLanguage(r.Context())
    // lang is "en", "tr", etc.
}
```

In your templates:

```html
<p>Current language: {{.Lang}}</p>
```

## SEO Features

### Canonical URLs

Statigo automatically generates canonical URLs:

```html
<link rel="canonical" href="{{canonicalURL "/about" .Lang}}">
```

### Alternate Links (Hreflang)

Generate hreflang links for SEO:

```html
{{alternateLinks "/about"}}
```

Output:
```html
<link rel="alternate" hreflang="en" href="https://example.com/en/about">
<link rel="alternate" hreflang="tr" href="https://example.com/tr/hakkinda">
<link rel="alternate" hreflang="x-default" href="https://example.com/en/about">
```

### Locale-Aware Links

Use `localePath` for translated URLs:

```html
<a href="{{localePath "/about" .Lang}}">About</a>
```

This automatically resolves to:
- `/en/about` for English
- `/tr/hakkinda` for Turkish

## Canonical Path Middleware

The `CanonicalPathMiddleware` ensures users are redirected to the correct language-specific URL:

```go
r.Use(router.CanonicalPathMiddleware(routeRegistry))
```

Example redirects:
- `/about` → `/en/about` (for English users)
- `/about` → `/tr/hakkinda` (for Turkish users)

## Programmatic Route Registration

```go
import "statigo/framework/router"

// Create registry
routeRegistry := router.NewRegistry([]string{"en", "tr"})

// Define a route programmatically
routeRegistry.Register(router.RouteDefinition{
    Canonical: "/contact",
    Paths: map[string]string{
        "en": "/en/contact",
        "tr": "/tr/iletisim",
    },
    Strategy: "static",
    Template: "contact.html",
    Handler:  "contact",
    Title:    "pages.contact.title",
})

// Register routes with chi router
routeRegistry.RegisterRoutes(r, nil)
```

## Dynamic Routes

For dynamic routes (e.g., blog posts), use chi's route parameters:

```go
r.Get("/{lang}/blog/{slug}", blogHandler.ServeHTTP)
```

Access parameters in your handler:

```go
slug := chi.URLParam(r, "slug")
lang := middleware.GetLanguage(r.Context())
```

## Redirects

Configure static redirects in `config/redirects.json`:

```json
{
  "redirects": [
    {
      "from": "/old-page",
      "to": "/new-page",
      "type": 301
    },
    {
      "from": "/blog/*",
      "to": "/articles/*",
      "type": 301,
      "pattern": true
    }
  ]
}
```

Apply redirects middleware:

```go
r.Use(middleware.RedirectMiddleware(configFS, "redirects.json", logger))
```
