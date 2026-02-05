<div align="center">

<img src="statigo.png" alt="Statigo Logo" width="200" />

# Statigo

![Go Version](https://img.shields.io/badge/Go-1.25-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/License-MIT-blue.svg)

</div>

A lightweight Go web framework for building static-first, SEO-optimized websites.

## What is Statigo?

Statigo is a framework designed for content-driven websites that prioritize performance and search engine optimization. Unlike dynamic-heavy frameworks, Statigo takes a static-first approach:

- **Prerender pages at build time** for instant page loads and perfect SEO
- **Serve static HTML** by default, with optional dynamic features
- **Cache everything** with a two-tier (memory + disk) caching system
- **Deploy as a single binary** with all assets embedded

It's ideal for documentation sites, blogs, marketing pages, and any content-focused web application where speed and SEO matter more than real-time dynamic features.

## Features

- **Performance** - Two-tier caching (memory + disk) with Brotli compression
- **SEO** - Canonical URLs, hreflang links, sitemaps, and structured data
- **i18n** - Multi-language support with URL-based language routing
- **Security** - Rate limiting, IP banning, honeypot, security headers
- **Deployment** - Single binary with embedded assets
- **CLI** - Prerender and cache management commands

## Quick Start

```bash
# Clone the repository
git clone https://github.com/statigo/statigo.git
cd statigo

# Run the development server
make dev

# Or build and run
make build
./statigo
```

Visit `http://localhost:8080` to see the example site.

## Documentation

- [Overview](docs/overview.md)
- [Getting Started](docs/getting-started.md)
- [Routing](docs/routing.md)
- [Middleware](docs/middleware.md)
- [Caching](docs/caching.md)
- [i18n](docs/i18n.md)
- [Templates](docs/templates.md)
- [Configuration](docs/configuration.md)
- [CLI](docs/cli.md)

## Project Structure

```text
statigo/
├── framework/          # Framework packages
│   ├── router/        # Multi-language routing
│   ├── middleware/    # HTTP middleware
│   ├── cache/         # Two-tier caching
│   ├── templates/     # HTML rendering
│   ├── i18n/          # Internationalization
│   ├── security/      # IP ban list
│   ├── health/        # Health checks
│   ├── logger/        # Structured logging
│   ├── client/        # HTTP client
│   ├── cli/           # CLI commands
│   └── ...
├── example/           # Example handlers
├── templates/         # HTML templates
├── static/           # Static assets
├── translations/     # Translation files
├── config/           # Configuration files
└── main.go           # Application entry point
```

## Configuration

Copy `.env.example` to `.env` and configure:

```bash
PORT=8080
BASE_URL=http://localhost:8080
LOG_LEVEL=INFO
CACHE_DIR=./data/cache
RATE_LIMIT_RPS=10
RATE_LIMIT_BURST=20
```

## License

MIT License - see LICENSE for details.
