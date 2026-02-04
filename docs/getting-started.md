# Getting Started

This guide will help you set up a Statigo project from scratch.

## Prerequisites

- Go 1.25 or later
- Basic knowledge of Go and HTTP

## Installation

### Create a New Project

```bash
mkdir my-project
cd my-project
go mod init my-project
```

### Add Statigo Dependency

If you're using Statigo as a module:

```bash
go get github.com/yourusername/statigo
```

Or copy the framework files directly into your project.

## Basic Setup

### 1. Create Directory Structure

```bash
mkdir -p templates/{layouts,pages,partials}
mkdir -p static/{styles,scripts}
mkdir -p translations
mkdir -p config
```

### 2. Create a Simple Route

Create `config/routes.json`:

```json
{
  "routes": [
    {
      "canonical": "/",
      "paths": {
        "en": "/en"
      },
      "strategy": "static",
      "template": "index.html",
      "handler": "index",
      "title": "Home"
    }
  ]
}
```

### 3. Create Base Template

Create `templates/layouts/base.html`:

```html
<!DOCTYPE html>
<html lang="{{.Lang}}">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>{{if .Title}}{{.Title}}{{else}}Welcome{{end}}</title>
</head>
<body>
    {{block "main" .}}{{end}}
</body>
</html>
```

### 4. Create Page Template

Create `templates/pages/index.html`:

```html
{{define "title"}}Home{{end}}
{{template "base" .}}

{{define "main"}}
<h1>Welcome to Statigo!</h1>
<p>This is your first page.</p>
{{end}}
```

### 5. Create Main Application

Create `main.go`:

```go
package main

import (
    "embed"
    "io/fs"
    "log/slog"
    "net/http"
    "os"

    "statigo/framework/cache"
    "statigo/framework/i18n"
    "statigo/framework/router"
    "statigo/framework/templates"
)

//go:embed templates static translations config
var embedded embed.FS

func main() {
    logger := slog.Default()

    // Get embedded filesystems
    templatesFS, _ := fs.Sub(embedded, "templates")
    translationsFS, _ := fs.Sub(embedded, "translations")
    configFS, _ := fs.Sub(embedded, "config")

    // Initialize i18n
    i18nInstance, _ := i18n.New(translationsFS, "en")

    // Initialize routing
    routeRegistry := router.NewRegistry([]string{"en"})

    // Initialize renderer
    renderer, _ := templates.NewRenderer(
        templatesFS,
        i18nInstance,
        nil,
        logger,
    )

    // Initialize cache
    cacheManager, _ := cache.NewManager("data/cache", logger)

    // Create handler
    handler := &IndexHandler{
        renderer:      renderer,
        cacheManager:  cacheManager,
        routeRegistry: routeRegistry,
    }

    // Register routes
    customHandlers := map[string]http.HandlerFunc{
        "index": handler.ServeHTTP,
    }

    router.LoadRoutesFromJSON(
        configFS,
        "routes.json",
        routeRegistry,
        renderer,
        customHandlers,
        logger,
    )

    // Start server
    r := routeRegistry.CreateRouter()
    http.ListenAndServe(":8080", r)
}

type IndexHandler struct {
    renderer      *templates.Renderer
    cacheManager  *cache.Manager
    routeRegistry *router.Registry
}

func (h *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    data := map[string]interface{}{
        "Title": "Welcome",
        "Lang":  "en",
    }
    h.renderer.Render(w, "index.html", data)
}
```

### 6. Add Translation

Create `translations/en.json`:

```json
{
  "nav": {
    "home": "Home",
    "about": "About"
  }
}
```

## Run Your Application

```bash
go run .
```

Visit http://localhost:8080/en to see your page.

## Environment Variables

Create a `.env` file for configuration:

```bash
# Server
PORT=8080
BASE_URL=http://localhost:8080

# Logging
LOG_LEVEL=INFO

# Cache
CACHE_DIR=./data/cache

# Rate Limiting
RATE_LIMIT_RPS=10
RATE_LIMIT_BURST=20

# Development
DEV_MODE=true
```

## Next Steps

- [Routing](routing) - Learn about multi-language routing
- [Middleware](middleware) - Add security and performance features
- [Templates](templates) - Create reusable templates
