# Caching

Statigo implements a sophisticated two-tier caching system designed for high-performance static content delivery.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      HTTP Request                           │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
                    ┌─────────────┐
                    │ Memory Cache│  ← First tier (fastest)
                    └──────┬──────┘
                           │ Miss
                           ▼
                    ┌─────────────┐
                    │  Disk Cache │  ← Second tier (persistent)
                    └──────┬──────┘
                           │ Miss
                           ▼
                    ┌─────────────┐
                    │   Handler   │  ← Generate response
                    └──────┬──────┘
                           │
                           ▼
                    ┌─────────────┐
                    │   Store     │  ← Cache for next time
                    │  (Memory+   │
                    │   Disk)     │
                    └─────────────┘
```

## Cache Manager

Initialize the cache manager:

```go
cacheManager, err := cache.NewManager("./data/cache", logger)
```

### Memory Cache

- In-memory LRU cache
- Sub-millisecond access times
- Configurable size
- Lost on restart (mitigated by disk cache)

### Disk Cache

- Persistent storage on filesystem
- Brotli compression for all content
- Survives restarts
- Used for cache warming on startup

## Cache Strategies

### Static

Cached indefinitely, no expiration:

```json
{
  "canonical": "/about",
  "strategy": "static"
}
```

Use for:
- Pages
- Evergreen content
- Blog posts

### Immutable

Same as static, for truly immutable content:

```json
{
  "canonical": "/terms",
  "strategy": "immutable"
}
```

Use for:
- Legal pages
- Archived content
- Versioned assets

### Incremental

Time-based revalidation:

```json
{
  "canonical": "/blog",
  "strategy": "incremental"
}
```

Revalidated daily at configured hour (default: 3 AM):

```bash
CACHE_REVALIDATION_HOUR=3
```

Use for:
- List pages
- Indexes
- Feeds

### Dynamic

Not cached:

```json
{
  "canonical": "/user/profile",
  "strategy": "dynamic"
}
```

Use for:
- User-specific content
- Real-time data
- Authentication pages

## Cache Keys

Cache keys are generated from:

```
{canonical}:{lang}:{params...}
```

Examples:
- `/about:en` → English about page
- `/blog:en` → English blog list
- `/blog/{slug}:en:post-1` → Blog post with parameter

## Cache Warming

On startup, the cache manager loads previously cached entries:

```go
// Automatically done in NewManager
cacheManager, _ := cache.NewManager(cacheDir, logger)
// Disk cache entries loaded into memory
```

Benefits:
- First requests are fast (from memory)
- No cold-start penalty
- Leverage previous cache runs

## Cache Invalidation

### Manual Invalidation

```go
cacheManager Invalidate("/about:en")
```

### Via Webhook

Configure webhook endpoint:

```go
webhookAuth := middleware.WebhookAuth(secret, logger)
r.With(webhookAuth).Patch("/webhook/revalidate/static", revalidateHandler)
```

Trigger revalidation:

```bash
curl -X PATCH http://localhost:8080/webhook/revalidate/static \
  -H "X-Webhook-Secret: your-secret" \
  -H "Content-Type: application/json" \
  -d '{"paths": ["/about", "/blog"]}'
```

### CLI Commands

```bash
# Prerender all routes
./statigo prerender

# Clear all cache
./statigo clear-cache
```

## Cache Entry Structure

```go
type Entry struct {
    Key         string
    Content     []byte      // Compressed content
    Strategy    string      // static/incremental/dynamic
    RequestPath string
    ETag        string
    CreatedAt   time.Time
    ExpiresAt   time.Time   // For incremental strategy
}
```

## Cache Middleware

The cache middleware automatically:

1. Checks memory cache
2. Checks disk cache (on memory miss)
3. Calls handler (on disk miss)
4. Stores response in both caches
5. Sets appropriate headers

### Response Headers

```
X-Cache: HIT
ETag: "abc123"
Content-Encoding: br
```

## Compression

All cached content is compressed using Brotli:

- Better compression than gzip
- Faster decompression
- Browser support: 95%+

Compression happens automatically:
- Before storing on disk
- Before storing in memory
- Decompressed on cache hit

## Monitoring

### Cache Hit Rate

```go
// Access cache statistics (you'd need to add this)
hitRate := cacheManager.HitRate()
logger.Info("cache stats", "hit_rate", hitRate)
```

### Cache Size

```bash
# Check disk cache size
du -sh ./data/cache

# List cached entries
ls ./data/cache
```

## Best Practices

1. **Use static strategy** for most content
2. **Prerender critical pages** on deployment
3. **Set up webhook revalidation** for CMS updates
4. **Monitor cache hit rates** in production
5. **Use incremental** for frequently updated lists
6. **Avoid dynamic** unless absolutely necessary

## Configuration

Environment variables:

```bash
# Cache directory
CACHE_DIR=./data/cache

# Revalidation hour (0-23)
CACHE_REVALIDATION_HOUR=3
```

## Example: Full Setup

```go
// Initialize cache
cacheDir := os.Getenv("CACHE_DIR")
cacheManager, err := cache.NewManager(cacheDir, logger)

// Add cache middleware
r.Use(middleware.CacheMiddleware(cacheManager, logger))

// Set router for eager revalidation
cacheManager.SetRouter(r)

// Prerender on startup (optional)
cli.PrerenderCommand{
    Router:       r,
    CacheManager: cacheManager,
    Logger:       logger,
}.Execute()
```

## Troubleshooting

### Cache Not Working

1. Check cache directory permissions
2. Verify middleware order (cache after canonical path)
3. Check for `X-Cache: HIT/MISS` headers
4. Ensure route has cache strategy set

### High Memory Usage

Reduce memory cache size by limiting entries:

```go
// You'd need to add this option
cacheManager.SetMaxEntries(1000)
```

### Stale Content

1. Check revalidation hour setting
2. Manually clear cache: `./statigo clear-cache`
3. Verify webhook is firing on content updates
