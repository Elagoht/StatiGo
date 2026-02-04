# Statigo

A static-first, SEO-optimized Go web framework extracted from production systems.

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

```
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
