# Overview

Statigo is a production-ready Go web framework designed for building high-performance, SEO-optimized, multi-language websites with static-first architecture.

## What is Statigo?

Statigo extracts proven patterns from production landing page systems and provides them as an easy-to-use framework. It's built for developers who need:

- **Fast page loads** with intelligent caching
- **SEO optimization** out of the box
- **Multi-language support** with proper URL routing
- **Security** with comprehensive middleware
- **Simple deployment** as a single binary

## Key Features

### Static-First Architecture
Statigo pre-renders pages and caches them intelligently. This means:
- First request may be slower (generates cache)
- Subsequent requests are extremely fast (serves from cache)
- Cache invalidation happens via webhooks or time-based strategies

### Multi-Language Routing
Built-in support for multiple languages with SEO-friendly URLs:
```json
{
  "canonical": "/about",
  "paths": {
    "en": "/en/about",
    "tr": "/tr/hakkinda"
  }
}
```

### Caching Strategies
Choose the right caching strategy for each route:
- **immutable** - Never expires (e.g., static assets)
- **static** - Long cache, revalidate when marked stale
- **incremental** - Auto-revalidate after 24 hours
- **dynamic** - Always revalidate when stale

### Security Middleware
Comprehensive security protection included:
- Rate limiting with token bucket algorithm
- IP ban list with persistent storage
- Honeypot traps for bot detection
- Security headers (CSP, HSTS, X-Frame-Options)
- Request logging with structured output

## Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────────────┐
│         Middleware Pipeline             │
│  ─────────────────────────────────────  │
│  • Structured Logging                   │
│  • IP Ban List                          │
│  • Honeypot Protection                  │
│  • Rate Limiting                        │
│  • Compression (Brotli/Gzip)           │
│  • Security Headers                     │
│  • Language Detection                   │
│  • Cache Lookup                         │
└──────────────┬──────────────────────────┘
               │
               ▼
┌─────────────────────────────────────────┐
│           Route Handler                 │
│  ─────────────────────────────────────  │
│  • Page Handler (index, about, etc.)    │
│  • Template Rendering                   │
│  • Cache Storage                        │
└─────────────────────────────────────────┘
```

## Project Structure

```
statigo/
├── framework/           # Core framework (exported package)
│   ├── router/         # Multi-language routing
│   ├── middleware/     # HTTP middleware
│   ├── cache/          # Two-tier caching
│   ├── templates/      # HTML rendering
│   ├── i18n/           # Internationalization
│   └── ...
├── example/            # Example application
│   └── handlers/       # Example handlers
├── templates/          # HTML templates
├── static/            # CSS, JS, assets
├── translations/      # i18n JSON files
├── config/            # Routes, redirects
└── docs/              # Documentation (markdown)
```

## Quick Start

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Run the server:**
   ```bash
   go run .
   ```

3. **Visit:**
   - Home: http://localhost:8080/en
   - About: http://localhost:8080/en/about
   - Docs: http://localhost:8080/en/docs

## Next Steps

- [Getting Started](getting-started) - Learn the basics
- [Routing](routing) - Configure multi-language routes
- [Middleware](middleware) - Add middleware to your app
- [Caching](caching) - Understand caching strategies
