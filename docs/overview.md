# Statigo Overview

Statigo is a production-ready Go web framework designed for static-first, SEO-optimized websites. It was extracted from real-world production systems handling high-traffic landing pages.

## Design Philosophy

### Static-First Approach

Statigo is optimized for serving pre-rendered HTML content with intelligent caching:

1. **Prerendering** - Generate static HTML at build time
2. **Two-tier Cache** - Memory cache for hot content, disk for persistence
3. **Cache Warming** - Bootstrap cache on startup from previous runs
4. **Incremental Updates** - Revalidate individual pages via webhooks

### SEO Optimization

Built-in SEO features for modern search engines:

- Canonical URL management
- Hreflang alternate links for multi-language
- Structured data (JSON-LD) support
- Automatic sitemap generation
- Semantic HTML structure

### Developer Experience

- Configuration-driven routing via JSON
- Embedded filesystems for single-binary deployment
- Hot reload during development with Air
- Semantic CLI commands for common tasks

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                         HTTP Request                        │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    Middleware Pipeline                       │
│  ┌──────────┐ ┌──────────┐ ┌──────────┐ ┌──────────┐      │
│  │ Security │ │   Rate   │ │  Cache   │ │ Language │      │
│  │ Headers  │ │  Limit   │ │  Check   │ │ Detect   │      │
│  └──────────┘ └──────────┘ └──────────┘ └──────────┘      │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                      Route Registry                          │
│            Canonical Path → Handler + Template              │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                      Page Handler                            │
│         Render HTML with i18n + SEO metadata                 │
└──────────────────────────┬──────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                      Cache Layer                             │
│              Store compressed response                      │
└─────────────────────────────────────────────────────────────┘
```

## Core Concepts

### Canonical Paths

Every page has a canonical path that serves as its identifier:

```json
{
  "canonical": "/about",
  "paths": {
    "en": "/en/about",
    "tr": "/tr/hakkinda"
  }
}
```

### Cache Strategies

- **static** - Never expires (pages, blog posts)
- **immutable** - Same as static, for truly immutable content
- **incremental** - Time-based revalidation (lists, indexes)
- **dynamic** - Not cached (user-specific, real-time)

### Language Routing

Languages are encoded in the URL path:

```
/en/about    → English version
/tr/hakkinda → Turkish version
```

The framework automatically:

- Detects language from URL
- Redirects root to preferred language
- Sets hreflang links for SEO
- Passes language to templates

## When to Use Statigo

Statigo is ideal for:

- Marketing landing pages
- Product documentation sites
- Company websites
- Blogs and content sites
- Multi-language sites
- High-traffic static content

Not recommended for:

- Highly dynamic applications
- Real-time features
- Complex user authentication
- Database-heavy applications

## Performance Characteristics

| Metric           | Value                      |
| ---------------- | -------------------------- |
| Cold start       | ~50ms                      |
| Cache hit        | <1ms                       |
| Cache miss       | ~10-50ms                   |
| Memory footprint | ~20-50MB                   |
| Binary size      | ~15-25MB (embedded assets) |
