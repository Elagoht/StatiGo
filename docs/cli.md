# CLI Commands

Statigo includes a CLI system for common operations like prerendering and cache management.

## Available Commands

```bash
./statigo <command> [options]
```

### Prerender

Generate static HTML for all routes:

```bash
./statigo prerender
```

**What it does:**

- Fetches all registered routes
- Makes HTTP requests to each route
- Stores responses in cache
- Warms both memory and disk cache

**Use cases:**

- Initial deployment
- After cache clear
- Scheduled cache warming

**Options:**
None currently. Prerenders all configured routes.

### Clear Cache

Remove all cached content:

```bash
./statigo clear-cache
```

**What it does:**

- Deletes all files from cache directory
- Clears memory cache
- Logs number of files deleted

**Use cases:**

- Force refresh all content
- Free disk space
- Debug cache issues

### Help

Show available commands:

```bash
./statigo help
./statigo --help
./statigo -h
```

## Command Implementation

### Creating Custom Commands

```go
package main

import (
    "statigo/framework/cli"
)

func main() {
    // Check if CLI command is provided
    if cli.ShouldRunCommand() {
        runCLI()
        return
    }

    // Start server
    runServer()
}

func runCLI() {
    cliManager := cli.New()

    // Register your command
    cliManager.Register(cli.Command{
        Name:        "custom",
        Description: "My custom command",
        Handler:     myCustomHandler,
    })

    // Execute
    if err := cliManager.Execute(os.Args[1:]); err != nil {
        log.Fatal(err)
    }
}

func myCustomHandler(args []string) error {
    fmt.Println("Running custom command")
    return nil
}
```

### Command Structure

```go
type Command struct {
    Name        string                           // Command name
    Description string                           // Help text
    Handler     func(args []string) error       // Command handler
    Flags       []Flag                          // Command flags
}

type Flag struct {
    Name        string  // Flag name
    Short       string  // Short flag
    Description string  // Help text
    Default     interface{}  // Default value
    Required    bool    // Is required
}
```

## Built-in Commands

### Prerender Command

```go
cli.NewPrerenderCommand(cli.PrerenderCommandConfig{
    ConfigFS:        configFS,
    RoutesFile:      "routes.json",
    Languages:       languages,
    Router:          router,
    ServiceRegistry: services,
    CacheManager:    cacheManager,
    Logger:          logger,
})
```

**Execution flow:**

1. Load routes from `routes.json`
2. For each route and language combination:
   - Construct URL with language prefix
   - Make HTTP request to internal router
   - Store response in cache
3. Log statistics (total, success, failed)

**Output:**

```
Prerendering routes...
  ✓ /en (200)
  ✓ /en/about (200)
  ✓ /tr (200)
  ✓ /tr/hakkinda (200)
Completed: 4 routes prerendered
```

### Clear Cache Command

```go
cli.NewClearCacheCommand(cli.ClearCacheCommandConfig{
    CacheDir: cacheDir,
    Logger:   logger,
})
```

**Execution flow:**

1. Check cache directory exists
2. Count cache files
3. Delete all files
4. Log deletion count

**Output:**

```
Clearing cache: ./data/cache
Deleted 42 cache files
Cache cleared successfully
```

## Integration with main.go

The CLI system integrates with the main application:

```go
func main() {
    // Initialize all components
    // ...

    // Check for CLI commands
    if cli.ShouldRunCommand() {
        runCLICommands(router, cacheManager, cacheDir, configFS, languages, services, logger)
        return
    }

    // Start server
    runServer(router, port, logger)
}

func runCLICommands(router, cacheManager, cacheDir, configFS, languages, services, logger) {
    cliManager := cli.New()

    // Register prerender command
    cliManager.Register(cli.NewPrerenderCommand(cli.PrerenderCommandConfig{
        ConfigFS:        configFS,
        RoutesFile:      "routes.json",
        Languages:       languages,
        Router:          router,
        ServiceRegistry: services,
        CacheManager:    cacheManager,
        Logger:          logger,
    }))

    // Register clear-cache command
    cliManager.Register(cli.NewClearCacheCommand(cli.ClearCacheCommandConfig{
        CacheDir: cacheDir,
        Logger:   logger,
    }))

    // Execute command
    if err := cliManager.Execute(os.Args[1:]); err != nil {
        logger.Error("Command failed", "error", err)
        fmt.Fprintf(os.Stderr, "Error: %v\n\n", err)
        cliManager.PrintHelp()
        os.Exit(1)
    }
}
```

## Running Commands

### Direct Execution

```bash
./statigo prerender
./statigo clear-cache
```

### Via Make

```makefile
.PHONY: prerender clear-cache

prerender:
	./statigo prerender

clear-cache:
	./statigo clear-cache
```

```bash
make prerender
make clear-cache
```

### Via systemd

```bash
# Prerender cache
systemd-run --property=User=statigo /opt/statigo/statigo prerender

# Clear cache
systemd-run --property=User=statigo /opt/statigo/statigo clear-cache
```

### Cron Jobs

```cron
# Prerender daily at 3 AM
0 3 * * * statigo cd /opt/statigo && ./statigo prerender

# Clear cache weekly on Sunday
0 2 * * 0 statigo cd /opt/statigo && ./statigo clear-cache
```

## Automation

### Deployment Script

```bash
#!/bin/bash
set -e

echo "Building statigo..."
go build -o statigo .

echo "Clearing old cache..."
./statigo clear-cache

echo "Prerendering routes..."
./statigo prerender

echo "Starting server..."
./statigo
```

### CI/CD Integration

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: "1.25"

      - name: Build
        run: go build -o statigo .

      - name: Prerender
        run: ./statigo prerender

      - name: Deploy
        run: |
          scp statigo user@server:/opt/statigo/
          ssh user@server "systemctl restart statigo"
```

## Troubleshooting

### Command Not Recognized

1. Check binary is built: `go build`
2. Verify command name: `./statigo help`
3. Check for typos

### Prerender Fails

1. Check routes.json is valid
2. Verify router is initialized
3. Check translations are loaded
4. Ensure templates are embedded

### Clear Cache Fails

1. Check directory permissions
2. Verify CACHE_DIR path
3. Check if files are in use

## Best Practices

1. **Prerender after deployment** - Warm cache before serving traffic
2. **Clear cache before deploy** - Ensure fresh content
3. **Schedule prerender** - Keep cache warm periodically
4. **Monitor command results** - Log command output
5. **Test commands locally** - Verify before running in production
