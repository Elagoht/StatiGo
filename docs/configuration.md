# Configuration

Statigo uses environment variables and JSON files for configuration.

## Environment Variables

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

### Server Configuration

```bash
# Server port (default: 8080)
PORT=8080

# Base URL for canonical URLs and sitemaps
BASE_URL=http://localhost:8080
```

### Logging

```bash
# Log level: DEBUG, INFO, WARN, ERROR (default: INFO)
LOG_LEVEL=INFO

# Log format: BRACKET, JSON (default: BRACKET)
LOG_FORMAT=BRACKET
```

### Cache Configuration

```bash
# Cache directory (default: ./data/cache)
CACHE_DIR=./data/cache

# Hour for daily cache revalidation (0-23, default: 3)
CACHE_REVALIDATION_HOUR=3
```

### Rate Limiting

```bash
# Requests per second (default: 10)
RATE_LIMIT_RPS=10

# Maximum burst size (default: 20)
RATE_LIMIT_BURST=20
```

### HTTP Client

```bash
# Overall request timeout in seconds (default: 30)
HTTP_TIMEOUT=30

# Connection timeout in seconds (default: 10)
HTTP_CONNECT_TIMEOUT=10

# TLS handshake timeout in seconds (default: 10)
HTTP_TLS_TIMEOUT=10

# Idle connection timeout in seconds (default: 90)
HTTP_IDLE_TIMEOUT=90

# Maximum retry attempts (default: 3)
HTTP_MAX_RETRIES=3

# Base delay for exponential backoff in ms (default: 500)
HTTP_RETRY_BASE_DELAY=500
```

### Graceful Shutdown

```bash
# Graceful shutdown timeout in seconds (default: 30)
SHUTDOWN_TIMEOUT=30
```

### Webhook (Optional)

```bash
# Secret for webhook authentication
WEBHOOK_SECRET=your-webhook-secret-here
```

## JSON Configuration Files

### routes.json

Define all application routes:

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

### redirects.json

Static URL redirects:

```json
{
  "/old-path": ["/new-path"],
  "/another-old": ["/another-new"]
}
```

## File Structure

```
.
├── .env                  # Environment variables (not in git)
├── .env.example          # Environment template
├── config/
│   ├── routes.json      # Route definitions
│   └── redirects.json   # Redirect mappings
├── data/
│   ├── cache/           # Cache storage
│   └── banned-ips.json  # IP ban list (auto-created)
└── translations/
    ├── en.json          # English translations
    └── tr.json          # Turkish translations
```

## Loading Configuration

### Environment Variables

Using godotenv:

```go
import "github.com/joho/godotenv"

func main() {
    // Load .env file
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file, using defaults")
    }

    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
}
```

### Routes

```go
import "statigo/framework/router"

routeRegistry := router.NewRegistry([]string{"en", "tr"})

router.LoadRoutesFromJSON(
    configFS,
    "routes.json",
    routeRegistry,
    renderer,
    customHandlers,
    logger,
)
```

### Redirects

```go
import "statigo/framework/middleware"

redirectRegistry, err := middleware.LoadRedirectsFromJSON(
    configFS,
    "redirects.json",
    logger,
)
```

## Default Values

| Variable                  | Default               | Description       |
| ------------------------- | --------------------- | ----------------- |
| `PORT`                    | 8080                  | Server port       |
| `BASE_URL`                | http://localhost:8080 | Base URL          |
| `LOG_LEVEL`               | INFO                  | Logging level     |
| `LOG_FORMAT`              | BRACKET               | Log format        |
| `CACHE_DIR`               | ./data/cache          | Cache directory   |
| `CACHE_REVALIDATION_HOUR` | 3                     | Revalidation hour |
| `RATE_LIMIT_RPS`          | 10                    | Rate limit        |
| `RATE_LIMIT_BURST`        | 20                    | Burst size        |
| `HTTP_TIMEOUT`            | 30                    | Request timeout   |
| `SHUTDOWN_TIMEOUT`        | 30                    | Shutdown timeout  |

## Configuration Helpers

### Get Environment Variable with Default

```go
import "statigo/framework/utils"

port := utils.GetEnvString("PORT", "8080")
debug := utils.GetEnvBool("DEBUG", false)
timeout := utils.GetEnvInt("TIMEOUT", 30)
```

### Required Environment Variables

```go
func checkRequiredEnv(vars []string) {
    for _, v := range vars {
        if os.Getenv(v) == "" {
            log.Fatalf("Required env var %s is missing", v)
        }
    }
}

checkRequiredEnv([]string{"BASE_URL", "CACHE_DIR"})
```

## Production Configuration

### .env for Production

```bash
# Server
PORT=8080
BASE_URL=https://example.com

# Logging
LOG_LEVEL=WARN
LOG_FORMAT=JSON

# Cache
CACHE_DIR=/var/cache/statigo
CACHE_REVALIDATION_HOUR=3

# Rate limiting
RATE_LIMIT_RPS=100
RATE_LIMIT_BURST=200

# HTTP client
HTTP_TIMEOUT=10
HTTP_CONNECT_TIMEOUT=5
HTTP_TLS_TIMEOUT=5

# Security
WEBHOOK_SECRET=${WEBHOOK_SECRET}
```

### systemd Service

```ini
[Unit]
Description=Statigo Web Application
After=network.target

[Service]
Type=simple
User=statigo
WorkingDirectory=/opt/statigo
Environment="PORT=8080"
Environment="BASE_URL=https://example.com"
Environment="LOG_LEVEL=WARN"
Environment="CACHE_DIR=/var/cache/statigo"
ExecStart=/opt/statigo/statigo
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

## Development Configuration

### .env for Development

```bash
# Server
PORT=8080
BASE_URL=http://localhost:8080

# Logging
LOG_LEVEL=DEBUG
LOG_FORMAT=BRACKET

# Cache (optional, disable for faster development)
# CACHE_DIR=/tmp/statigo-cache

# Rate limiting (relaxed for dev)
RATE_LIMIT_RPS=100
RATE_LIMIT_BURST=200

# Development mode (disables static caching)
DEV_MODE=true
```

## Secrets Management

### Never Commit .env

The `.gitignore` already excludes `.env`:

```gitignore
# environment variables
.env
```

### Using Secret Managers

For production, use a secret manager:

```go
import (
    os"
    vault "github.com/hashicorp/vault/api"
)

func getSecret(key string) string {
    // Try environment first
    if val := os.Getenv(key); val != "" {
        return val
    }

    // Fall back to vault
    client := vault.DefaultClient()
    secret, _ := client.Logical().Read("secret/statigo/" + key)
    return secret.Data["value"].(string)
}
```

## Configuration Validation

Validate configuration on startup:

```go
func validateConfig() error {
    // Check required variables
    if os.Getenv("BASE_URL") == "" {
        return errors.New("BASE_URL is required")
    }

    // Validate port
    port := utils.GetEnvInt("PORT", 8080)
    if port < 1 || port > 65535 {
        return errors.New("PORT must be between 1 and 65535")
    }

    // Validate cache directory
    cacheDir := os.Getenv("CACHE_DIR")
    if cacheDir != "" {
        if err := os.MkdirAll(cacheDir, 0755); err != nil {
            return fmt.Errorf("cannot create cache directory: %w", err)
        }
    }

    return nil
}
```

## Hot Reload

Configuration changes require restart:

```bash
# Using Air for development
make dev

# Air watches for file changes and restarts
```

For production, use graceful restart:

```bash
# Send SIGTERM to trigger graceful shutdown
kill -TERM $(pidof statigo)

# Start new instance (systemd handles this)
systemctl restart statigo
```
