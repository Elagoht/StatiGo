# Middleware

Statigo includes a comprehensive middleware pipeline for security, performance, and functionality.

## Available Middleware

### Logging Middleware

Structured request logging with slog:

```go
r.Use(middleware.StructuredLogger(logger))
```

Output:
```
INFO request method=GET path=/en/about status=200 duration=5ms
```

### Compression Middleware

Brotli (preferred) and gzip compression:

```go
r.Use(middleware.Compression())
```

Automatically compresses: HTML, CSS, JS, JSON, XML, SVG

### Rate Limiting Middleware

Token bucket rate limiting:

```go
r.Use(middleware.RateLimiter(middleware.RateLimiterConfig{
    RPS:   10,  // Requests per second
    Burst: 20,  // Burst size
}))
```

Configure via environment:
```bash
RATE_LIMIT_RPS=10
RATE_LIMIT_BURST=20
```

### IP Ban Middleware

Block banned IPs with persistent storage:

```go
ipBanList, _ := security.NewIPBanList("data/banned-ips.json", logger)
r.Use(middleware.IPBanMiddleware(ipBanList, logger))
```

Ban an IP programmatically:
```go
ipBanList.Ban("192.168.1.100", "Abusive behavior", r)
```

### Honeypot Middleware

Trap bots accessing fake admin paths:

```go
honeypotPaths := []string{
    "/admin", "/wp-admin", "/wp-login.php",
    "/.env", "/.git/config",
}
r.Use(middleware.HoneypotMiddleware(ipBanList, honeypotPaths, logger))
```

Bots accessing these paths are automatically banned.

### Security Headers Middleware

Add security headers:

```go
// Simple preset
r.Use(middleware.SecurityHeadersSimple())

// Or customize
r.Use(middleware.SecurityHeaders(middleware.SecurityHeadersConfig{
    CSP:           "default-src 'self'",
    HSTSEnabled:   true,
    HSTSMaxAge:    31536000,
    FrameOptions:  "DENY",
    PermissionsPolicy: "geolocation=(), camera=()",
}))
```

Headers added:
- `Content-Security-Policy`
- `X-Frame-Options`
- `X-Content-Type-Options`
- `Strict-Transport-Security`
- `Permissions-Policy`

### Language Middleware

Detect and set language from URL, cookie, or Accept-Language header:

```go
langConfig := middleware.LanguageConfig{
    SupportedLanguages: []string{"en", "tr"},
    DefaultLanguage:    "en",
    SkipPaths:          []string{"/robots.txt", "/sitemap.xml"},
    SkipPrefixes:       []string{"/static/", "/health/"},
}
r.Use(middleware.Language(i18nInstance, langConfig))
```

Detection priority:
1. URL path prefix (`/en/`, `/tr/`)
2. Cookie (`lang`)
3. `Accept-Language` header
4. Default language

### Caching Headers Middleware

Add browser cache headers:

```go
r.Use(middleware.CachingHeaders(devMode))
```

Cache behavior:
- **Dev mode**: `no-cache`
- **Production**: Based on route strategy

### Layout Data Middleware

Inject shared data into all templates:

```go
layoutData := map[string]interface{}{
    "SiteName": "My Site",
    "Year":     2024,
}
r.Use(middleware.LayoutDataMiddleware(layoutData))
```

Access in templates:
```html
<h1>{{.SiteName}}</h1>
```

### Canonical Path Middleware

Store and validate canonical paths:

```go
r.Use(router.CanonicalPathMiddleware(routeRegistry))
```

### Cache Middleware

Response caching with automatic invalidation:

```go
r.Use(middleware.CacheMiddleware(cacheManager, logger))
```

### Webhook Auth Middleware

Validate webhook requests:

```go
r.Use(middleware.WebhookAuthMiddleware("my-secret-key"))
```

## Middleware Order

The recommended order is important:

```go
r.Use(middleware.StructuredLogger(logger))           // 1. Log everything
r.Use(chiMiddleware.Recoverer)                        // 2. Panic recovery
r.Use(middleware.IPBanMiddleware(ipBanList, logger))  // 3. Block banned IPs
r.Use(middleware.HoneypotMiddleware(ipBanList, paths, logger)) // 4. Trap bots
r.Use(middleware.RateLimiter(config))                 // 5. Rate limit
r.Use(middleware.Compression())                       // 6. Compress responses
r.Use(middleware.SecurityHeadersSimple())             // 7. Security headers
r.Use(middleware.CachingHeaders(devMode))             // 8. Cache headers
r.Use(middleware.Language(i18nInstance, config))      // 9. Language detection
r.Use(router.CanonicalPathMiddleware(routeRegistry))  // 10. Validate paths
r.Use(middleware.CacheMiddleware(cacheManager, logger)) // 11. Response cache
```

## Custom Middleware

Create your own middleware:

```go
func MyMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Before request
        start := time.Now()

        // Call next handler
        next.ServeHTTP(w, r)

        // After request
        duration := time.Since(start)
        log.Println("Request took", duration)
    })
}
```

Use it:
```go
r.Use(MyMiddleware)
```

## Chi Middleware

Statigo is built on chi. You can use any chi middleware:

```go
import chiMiddleware "github.com/go-chi/chi/middleware"

r.Use(chiMiddleware.RequestID)
r.Use(chiMiddleware.RealIP)
r.Use(chiMiddleware.Logger)
r.Use(chiMiddleware.Recoverer)
```

See [chi documentation](https://github.com/go-chi/chi) for more.
