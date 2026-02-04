# Middleware

Statigo provides a comprehensive set of middleware for security, performance, and functionality.

## Middleware Stack

The standard middleware stack (in order):

```go
r.Use(middleware.StructuredLogger(logger))
r.Use(chiMiddleware.Recoverer)
r.Use(middleware.IPBanMiddleware(ipBanList, logger))
r.Use(middleware.HoneypotMiddleware(ipBanList, honeypotPaths, logger))
r.Use(middleware.RateLimiter(config))
r.Use(middleware.Compression())
r.Use(middleware.SecurityHeaders(config))
r.Use(middleware.CachingHeaders(devMode))
r.Use(staticFileMiddleware(...))
r.Use(middleware.Language(i18nInstance, config))
r.Use(router.CanonicalPathMiddleware(registry))
r.Use(middleware.CacheMiddleware(cacheManager, logger))
```

## Available Middleware

### Logging Middleware

Structured logging for all requests:

```go
r.Use(middleware.StructuredLogger(logger))
```

Output format:
```
[2024-01-15 10:30:45][INFO][request handled][method=GET][path=/en/about][status=200][duration=5ms]
```

### Compression Middleware

Brotli and gzip compression:

```go
r.Use(middleware.Compression())
```

Automatically:
- Compresses responses based on Accept-Encoding
- Prefers Brotli over gzip
- Minifies CSS/JS/HTML in static files

### Rate Limiter

Token bucket rate limiting:

```go
config := middleware.RateLimiterConfig{
    RPS:              10,   // Requests per second
    Burst:            20,   // Burst size
    StaticMultiplier: 10,   // Higher limit for static assets
    CrawlerBypass:    true, // Bypass for known crawlers
}

r.Use(middleware.RateLimiter(config))
```

Headers on rate limit:
```
HTTP/1.1 429 Too Many Requests
Retry-After: 1
X-RateLimit-Limit: 10
X-RateLimit-Burst: 20
```

### Security Headers

```go
config := middleware.SecurityHeadersConfig{
    HSTSMaxAge:         31536000,
    FrameOptions:       "DENY",
    ContentTypeOptions: "nosniff",
    ReferrerPolicy:     "strict-origin-when-cross-origin",
}

r.Use(middleware.SecurityHeaders(config))
```

Headers added:
```
Strict-Transport-Security: max-age=31536000; includeSubDomains
X-Frame-Options: DENY
X-Content-Type-Options: nosniff
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=()
```

### IP Ban Middleware

Block requests from banned IPs:

```go
ipBanList, _ := security.NewIPBanList("data/banned-ips.json", logger)
r.Use(middleware.IPBanMiddleware(ipBanList, logger))
```

Ban list format:
```json
{
  "banned": [
    {
      "ip": "1.2.3.4",
      "reason": "abuse",
      "banned_at": "2024-01-15T10:00:00Z"
    }
  ]
}
```

### Honeypot Middleware

Trap and ban malicious bots:

```go
honeypotPaths := []string{
    "/admin", "/wp-admin", "/wp-login.php",
    "/.env", "/.git/config", "/phpMyAdmin",
}

r.Use(middleware.HoneypotMiddleware(ipBanList, honeypotPaths, logger))
```

Accessing honeypot paths results in:
- IP added to ban list
- HTTP 403 response
- Logged entry

### Language Middleware

Detect and set language from URL:

```go
config := middleware.LanguageConfig{
    SupportedLanguages: []string{"en", "tr"},
    DefaultLanguage:    "en",
    SkipPaths:          []string{"/robots.txt", "/sitemap.xml"},
    SkipPrefixes:       []string{"/health/", "/static/"},
}

r.Use(middleware.Language(i18nInstance, config))
```

Behavior:
- Extracts language from URL path (e.g., `/tr/about`)
- Redirects root to detected language
- Sets language cookie
- Passes language via context

### Cache Middleware

Serve cached responses:

```go
r.Use(middleware.CacheMiddleware(cacheManager, logger))
```

Cache flow:
1. Check cache for canonical path + language
2. If hit, serve cached response
3. If miss, generate and cache response
4. Respect cache strategy (static/incremental/dynamic)

### Caching Headers

Set cache-related headers:

```go
r.Use(middleware.CachingHeaders(devMode))
```

Headers for static assets:
```
Cache-Control: public, max-age=31536000, immutable
```

Headers for HTML in dev mode:
```
Cache-Control: no-cache
```

### Canonical Path Middleware

Store route metadata in context:

```go
r.Use(router.CanonicalPathMiddleware(routeRegistry))
```

Sets context values:
- `CanonicalPathKey` - Canonical path
- `PageTitleKey` - Page title translation key
- `StrategyKey` - Cache strategy

### Layout Data Middleware

Inject shared data for templates:

```go
r.Use(middleware.LayoutDataMiddleware(logger))
```

Available in templates:
```go
{{.Layout.SiteURL}}
```

### Redirect Middleware

Handle static redirects:

```go
redirectRegistry, _ := middleware.LoadRedirectsFromJSON(
    configFS,
    "redirects.json",
    logger,
)

r.Use(middleware.RedirectMiddleware(redirectRegistry, logger))
```

Redirects format:
```json
{
  "/old-path": ["/new-path"],
  "/another-old": ["/another-new"]
}
```

### Webhook Authentication

Authenticate webhook requests:

```go
webhookAuth := middleware.WebhookSecretAuth("your-secret", logger)
r.With(webhookAuth).Patch("/webhook/revalidate", handler)
```

Uses HMAC signature verification.

## Custom Middleware

Create custom middleware:

```go
func MyMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Before handler
        ctx := r.Context()

        // Do something
        customValue := "value"

        // Add to context
        ctx = context.WithValue(ctx, "customKey", customValue)

        // Call next
        next.ServeHTTP(w, r.WithContext(ctx))

        // After handler
    })
}
```

## Middleware Order

Order matters! General guidelines:

1. **First**: Logging, recovery
2. **Security**: IP ban, rate limit, honeypot
3. **Performance**: Compression, cache
4. **Routing**: Language, canonical path
5. **Last**: Application handlers

## Context Keys

Access middleware values from context:

```go
// Language
lang := middleware.GetLanguage(r.Context())

// Canonical path
canonical := router.GetCanonicalPath(r.Context())

// Cache strategy
strategy := router.GetStrategy(r.Context())

// Layout data
layoutData := middleware.GetLayoutData(r.Context())
```

## Conditional Middleware

Skip middleware for certain paths:

```go
func ConditionalMiddleware(skipPaths []string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            for _, path := range skipPaths {
                if r.URL.Path == path {
                    next.ServeHTTP(w, r)
                    return
                }
            }
            // Apply middleware
        })
    }
}
```
