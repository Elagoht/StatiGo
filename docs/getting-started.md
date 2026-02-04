# Getting Started

## Prerequisites

- Go 1.25 or later
- Make (optional, for convenience commands)

## Installation

### From Source

```bash
git clone https://github.com/statigo/statigo.git
cd statigo
go mod download
```

### Running the Example

```bash
# Development mode with hot reload
make dev

# Or build and run
go build -o statigo .
./statigo
```

Visit `http://localhost:8080` to see the example site.

## Creating Your First Site

### 1. Initialize Project Structure

```
my-site/
├── main.go
├── go.mod
├── go.sum
├── .env
├── embed.go
├── config/
│   └── routes.json
├── templates/
│   └── pages/
│       └── index.html
├── static/
│   └── styles/
│       └── main.css
└── translations/
    └── en.json
```

### 2. Create main.go

```go
package main

import (
    "log/slog"
    "net/http"
    "os"

    "github.com/go-chi/chi/v5"
    chiMiddleware "github.com/go-chi/chi/v5/middleware"
    "github.com/joho/godotenv"

    "statigo/framework/cache"
    "statigo/framework/i18n"
    "statigo/framework/logger"
    "statigo/framework/middleware"
    "statigo/framework/router"
    "statigo/framework/templates"
)

func main() {
    godotenv.Load()
    appLogger := logger.InitLogger("INFO")

    // Get embedded filesystems
    translationsFS := GetTranslationsFS()
    templatesFS := GetTemplatesFS()
    configFS := GetConfigFS()

    // Initialize components
    i18nInstance, _ := i18n.New(translationsFS, "en")
    languages := []string{"en"}
    routeRegistry := router.NewRegistry(languages)

    baseURL := os.Getenv("BASE_URL")
    seoHelpers := router.NewSEOHelpers(routeRegistry, baseURL)
    renderer, _ := templates.NewRenderer(
        templatesFS,
        i18nInstance,
        seoHelpers.ToTemplateFunctions(),
        appLogger,
    )

    // Cache
    cacheDir := os.Getenv("CACHE_DIR")
    cacheManager, _ := cache.NewManager(cacheDir, appLogger)

    // Router setup
    r := chi.NewRouter()
    r.Use(middleware.StructuredLogger(appLogger))
    r.Use(chiMiddleware.Recoverer)
    r.Use(middleware.Compression())
    r.Use(middleware.Language(i18nInstance, middleware.DefaultLanguageConfig()))
    r.Use(router.CanonicalPathMiddleware(routeRegistry))

    // Load routes
    router.LoadRoutesFromJSON(
        configFS,
        "routes.json",
        routeRegistry,
        renderer,
        nil, // custom handlers
        appLogger,
    )

    routeRegistry.RegisterRoutes(r, func(h http.Handler) http.Handler { return h })

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    appLogger.Info("Starting server", "port", port)
    http.ListenAndServe(":"+port, r)
}
```

### 3. Create embed.go

```go
package main

import "embed"

//go:embed templates
var templatesFS embed.FS

//go:embed static
var staticFS embed.FS

//go:embed translations
var translationsFS embed.FS

//go:embed config
var configFS embed.FS

func GetTemplatesFS() embed.FS { return templatesFS }
func GetStaticFS() embed.FS { return staticFS }
func GetTranslationsFS() embed.FS { return translationsFS }
func GetConfigFS() embed.FS { return configFS }
```

### 4. Configure Routes

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
      "handler": "content",
      "title": "pages.home.title"
    }
  ]
}
```

### 5. Create Template

Create `templates/pages/index.html`:

```html
{{define "base"}}
<!DOCTYPE html>
<html lang="{{.Lang}}">
<head>
    <meta charset="UTF-8">
    <title>{{t .Lang "pages.home.title"}}</title>
</head>
<body>
    <h1>{{t .Lang "pages.home.heading"}}</h1>
</body>
</html>
{{end}}
```

### 6. Add Translations

Create `translations/en.json`:

```json
{
  "pages": {
    "home": {
      "title": "My Site",
      "heading": "Welcome!"
    }
  }
}
```

### 7. Environment Variables

Create `.env`:

```bash
PORT=8080
BASE_URL=http://localhost:8080
LOG_LEVEL=INFO
CACHE_DIR=./data/cache
```

## Development

### Hot Reload

Using Air for live reload during development:

```bash
make dev
```

### Running CLI Commands

```bash
# Prerender all routes
./statigo prerender

# Clear cache
./statigo clear-cache
```

## Deployment

### Build for Production

```bash
go build -ldflags="-s -w" -o statigo .
```

### systemd Service

Copy the example service file:

```bash
cp landing-page.service.example /etc/systemd/system/statigo.service
# Edit and customize
sudo systemctl enable statigo
sudo systemctl start statigo
```

### Docker

```dockerfile
FROM golang:1.25 AS build
WORKDIR /app
COPY . .
RUN go build -ldflags="-s -w" -o statigo .

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=build /app/statigo /statigo
EXPOSE 8080
CMD ["/statigo"]
```

## Next Steps

- Learn about [Routing](routing.md)
- Configure [Middleware](middleware.md)
- Set up [Caching](caching.md)
- Add [i18n](i18n.md)
