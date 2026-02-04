# Caching

Statigo features a two-tier caching system for optimal performance: in-memory caching with disk persistence.

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Request   │────▶│ Memory Cache│────▶│Disk Cache   │
└─────────────┘     └─────────────┘     └─────────────┘
                          │
                          ▼
                     ┌─────────────┐
                     │  Brotli     │
                     │ Compressed  │
                     └─────────────┘
```

## Cache Strategies

Each route can have a different caching strategy:

| Strategy | Description | Use Case |
|----------|-------------|----------|
| `immutable` | Never expires | Static assets, versioned files |
| `static` | Long cache, revalidate when marked stale | Pages that change rarely |
| `incremental` | Auto-revalidate after 24 hours | Blog posts, articles |
| `dynamic` | Always revalidate when stale | User-specific content |

Define in `config/routes.json`:

```json
{
  "canonical": "/",
  "paths": {"en": "/en"},
  "strategy": "static",
  "template": "index.html",
  "handler": "index"
}
```

## Initialization

```go
import "statigo/framework/cache"

cacheManager, err := cache.NewManager("data/cache", logger)
if err != nil {
    log.Fatal(err)
}
```

## Cache Middleware

```go
r.Use(middleware.CacheMiddleware(cacheManager, logger))
```

The middleware automatically:
1. Checks memory cache for existing response
2. Checks disk cache if memory miss
3. Executes handler if both miss
4. Stores response in both tiers

## Pre-rendering

Generate all pages upfront:

```bash
# Build the application
go build -o statigo

# Prerender all pages
./statigo prerender
```

Or programmatically:

```go
cacheManager.RebuildAll(r, appLogger)
```

## Cache Invalidation

### Manual Invalidation

Mark a route as stale:

```go
cacheManager.MarkStale("/en/about")
```

### Webhook Invalidation

Configure webhook endpoint:

```go
r.Post("/cache/webhook", middleware.WebhookInvalidate(
    cacheManager,
    os.Getenv("WEBHOOK_SECRET"), // From environment
    logger,
))
```

Send webhook:

```bash
curl -X POST http://localhost:8080/cache/webhook \
  -H "X-Webhook-Secret: your-secret-key" \
  -H "Content-Type: application/json" \
  -d '{"canonical": "/about"}'
```

### Strategy-Based Invalidation

Rebuild by strategy:

```go
// Rebuild all static pages
cacheManager.RebuildByStrategy("static", r, logger)

// Rebuild specific route
cacheManager.RebuildByCanonical("/about", r, logger)
```

## Cache Storage

### Memory Cache

- Stored in `sync.Map` for concurrent access
- Compressed with Brotli
- Automatic ETag generation

### Disk Cache

- Stored in `data/cache/` directory
- Named by SHA256 hash of canonical path
- Survives application restarts

## ETag Support

Statigo automatically generates ETags for cache entries:

```
ETag: "a1b2c3d4e5f6..."
```

Clients with `If-None-Match` header receive `304 Not Modified` responses.

## Configuration

Environment variables:

```bash
# Cache directory
CACHE_DIR=./data/cache

# Disable caching (for testing)
DISABLE_CACHE=false
```

## Monitoring

Check cache health:

```go
stats := cacheManager.GetStats()
fmt.Printf("Memory entries: %d\n", stats.MemoryEntries)
fmt.Printf("Disk entries: %d\n", stats.DiskEntries)
```

## Best Practices

1. **Use `immutable` for truly static content**
   - Assets with version hashes: `/style.v1.css`
   - Documentation pages

2. **Use `static` for pages that change rarely**
   - Home page
   - About pages
   - Feature pages

3. **Use `incremental` for content pages**
   - Blog posts
   - Articles
   - News items

4. **Use `dynamic` for personalized content**
   - User dashboards
   - Admin panels
   - Account settings

5. **Prerender after deployment**
   ```bash
   go build -o app
   ./app prerender
   ./app serve
   ```

6. **Set up webhook invalidation**
   - For CMS integration
   - For content updates
   - For automated deployments
