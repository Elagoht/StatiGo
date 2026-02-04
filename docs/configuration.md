# Configuration

Statigo uses environment variables and JSON configuration files for flexible application configuration.

## Environment Variables

Create a `.env` file in your project root:

```bash
# Server Configuration
PORT=8080
BASE_URL=http://localhost:8080

# Logging
LOG_LEVEL=INFO

# Cache
CACHE_DIR=./data/cache
DISABLE_CACHE=false

# Rate Limiting
RATE_LIMIT_RPS=10
RATE_LIMIT_BURST=20

# Development
DEV_MODE=false

# Shutdown Timeout
SHUTDOWN_TIMEOUT=30

# Webhook Secret (for cache invalidation)
WEBHOOK_SECRET=your-webhook-secret-key

# Google Tag Manager (optional)
GTM_ID=GTM-XXXXX
```

### Server Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `BASE_URL` | `http://localhost:8080` | Base URL for canonical links |
| `DEV_MODE` | `false` | Enable development mode |

### Logging

| Variable | Default | Description |
|----------|---------|-------------|
| `LOG_LEVEL` | `INFO` | Log level: `DEBUG`, `INFO`, `WARN`, `ERROR` |

### Cache

| Variable | Default | Description |
|----------|---------|-------------|
| `CACHE_DIR` | `./data/cache` | Cache storage directory |
| `DISABLE_CACHE` | `false` | Disable caching (for testing) |

### Rate Limiting

| Variable | Default | Description |
|----------|---------|-------------|
| `RATE_LIMIT_RPS` | `10` | Requests per second |
| `RATE_LIMIT_BURST` | `20` | Burst size |

### Security

| Variable | Default | Description |
|----------|---------|-------------|
| `WEBHOOK_SECRET` | - | Secret for webhook authentication |

## Route Configuration

### routes.json

Define routes in `config/routes.json`:

```json
{
  "routes": [
    {
      "canonical": "/",
      "paths": {
        "en": "/en",
        "tr": "/tr"
      },
      "strategy": "static",
      "template": "index.html",
      "handler": "index",
      "title": "pages.home.title"
    }
  ]
}
```

### Route Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `canonical` | string | Yes | Internal canonical path |
| `paths` | object | Yes | Language-specific URL paths |
| `strategy` | string | Yes | Caching strategy |
| `template` | string | Yes | Template file name |
| `handler` | string | Yes | Handler name for customHandlers map |
| `title` | string | No | Page title or i18n key |

### Caching Strategies

- `immutable` - Never expires (static assets)
- `static` - Long cache, revalidate when stale
- `incremental` - Auto-revalidate after 24 hours
- `dynamic` - Always revalidate when stale

## Redirect Configuration

### redirects.json

Define redirects in `config/redirects.json`:

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

### Redirect Fields

| Field | Type | Description |
|-------|------|-------------|
| `from` | string | Source path (supports `*` wildcard) |
| `to` | string | Destination path (use `*` for matched part) |
| `type` | number | HTTP status code (301 or 302) |
| `pattern` | boolean | Enable wildcard matching |

## Translation Configuration

Translation files are stored in `translations/` directory:

```
translations/
├── en.json    # English
├── tr.json    # Turkish
└── de.json    # German
```

### Translation File Format

```json
{
  "site": {
    "name": "My Site",
    "description": "A description"
  },
  "nav": {
    "home": "Home",
    "about": "About"
  }
}
```

## Accessing Configuration in Go

### Environment Variables

```go
import "statigo/framework/utils"

port := os.Getenv("PORT")
if port == "" {
    port = "8080"
}

// Or use helper
rateLimitRPS := utils.GetEnvInt("RATE_LIMIT_RPS", 10)
rateLimitBurst := utils.GetEnvInt("RATE_LIMIT_BURST", 20)
```

### Loading Routes

```go
import "statigo/framework/router"

routeRegistry := router.NewRegistry([]string{"en", "tr"})

err := router.LoadRoutesFromJSON(
    configFS,
    "routes.json",
    routeRegistry,
    renderer,
    customHandlers,
    logger,
)
```

### Loading Redirects

```go
import "statigo/framework/middleware"

r.Use(middleware.RedirectMiddleware(
    configFS,
    "redirects.json",
    logger,
))
```

## Configuration Best Practices

1. **Never commit `.env`** - Add to `.gitignore`
2. **Use `.env.example`** - Template for required variables
3. **Validate configuration** - Check for required variables on startup
4. **Use sensible defaults** - Provide defaults for optional settings
5. **Document variables** - Explain what each variable does

## Example Configuration

### Production (.env.production)

```bash
PORT=8080
BASE_URL=https://example.com
LOG_LEVEL=WARN
DEV_MODE=false
RATE_LIMIT_RPS=20
RATE_LIMIT_BURST=40
WEBHOOK_SECRET=prod-secret-key
```

### Development (.env.development)

```bash
PORT=3000
BASE_URL=http://localhost:3000
LOG_LEVEL=DEBUG
DEV_MODE=true
RATE_LIMIT_RPS=100
RATE_LIMIT_BURST=200
```

### Testing (.env.test)

```bash
PORT=8081
BASE_URL=http://localhost:8081
LOG_LEVEL=ERROR
DEV_MODE=true
DISABLE_CACHE=true
RATE_LIMIT_RPS=1000
```

## Loading Environment Files

Statigo uses `godotenv` for loading `.env` files:

```go
import "github.com/joho/godotenv"

func main() {
    // Load .env file
    if err := godotenv.Load(); err != nil {
        log.Println("Warning: No .env file found, using defaults")
    }

    // Your application code...
}
```

For environment-specific files:

```go
env := os.Getenv("APP_ENV")
if env == "" {
    env = "development"
}

godotenv.Load(".env." + env)
godotenv.Load() // Load default .env as fallback
```
