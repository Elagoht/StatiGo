# CLI (Command Line Interface)

Statigo includes a CLI framework for common operations like pre-rendering and cache management.

## Built-in Commands

### prerender

Pre-render all pages to warm the cache:

```bash
./statigo prerender
```

This command:
1. Loads all routes from configuration
2. Makes requests to each route
3. Stores responses in cache
4. Reports success/failure for each page

Useful for:
- Initial deployment
- Cache warming after restart
- Ensuring all pages are cached

### clear-cache

Clear all cached pages:

```bash
./statigo clear-cache
```

This command:
1. Deletes all entries from memory cache
2. Deletes all files from disk cache
3. Reports number of entries cleared

Useful for:
- Forcing cache refresh
- Freeing disk space
- Troubleshooting cache issues

## Using the CLI

### Registering Commands

In your `main.go`:

```go
import "statigo/framework/cli"

func main() {
    // Initialize your components...

    // Create CLI instance
    cliInstance := cli.New()

    // Register built-in commands
    cli.RegisterPrerenderCommand(cliInstance, r, cacheManager, logger)
    cli.RegisterClearCacheCommand(cliInstance, cacheManager, logger)

    // Check if CLI command is provided
    if len(os.Args) > 1 {
        if err := cliInstance.Run(os.Args[1], os.Args[2:]); err == nil {
            // Command executed successfully
            return
        }
    }

    // Start server if no CLI command
    // ... your server code ...
}
```

### Creating Custom Commands

```go
import "statigo/framework/cli"

// Define your command handler
func MyCommandHandler(args []string, appCLI *cli.CLI) error {
    fmt.Println("Running my custom command")
    fmt.Println("Args:", args)
    return nil
}

// Register the command
cliInstance.Register("my-command", cli.Command{
    Handler:    MyCommandHandler,
    Usage:      "my-command [args...]",
    Summary:    "Description of my command",
    Aliases:    []string{"mc", "my-c"},
})
```

### Command with Dependencies

Pass your application components to command handlers:

```go
func ExportDataCommand(args []string, appCLI *cli.CLI) error {
    // Access dependencies from appCLI.Context
    db := appCLI.Context["db"].(*Database)
    cache := appCLI.Context["cache"].(*cache.Manager)

    // Your command logic...
    return nil
}

// Set context before registering
cliInstance.Context["db"] = database
cliInstance.Context["cache"] = cacheManager

cliInstance.Register("export", cli.Command{
    Handler: ExportDataCommand,
    Usage:   "export [format]",
    Summary: "Export data in various formats",
})
```

## Command Aliases

Commands can have multiple names (aliases):

```go
cliInstance.Register("version", cli.Command{
    Handler: VersionHandler,
    Usage:   "version",
    Summary: "Show version information",
    Aliases: []string{"v", "ver", "--version"},
})
```

All of these will work:
```bash
./statigo version
./statigo v
./statigo ver
./statigo --version
```

## CLI Context

Share data between server initialization and CLI commands:

```go
func main() {
    // Initialize components
    cacheManager := cache.NewManager(...)
    database := db.Connect(...)

    // Create CLI
    cliInstance := cli.New()

    // Share components via context
    cliInstance.Context["cache"] = cacheManager
    cliInstance.Context["db"] = database

    // Register commands that use these components
    cli.RegisterPrerenderCommand(cliInstance, r, cacheManager, logger)

    // Check for CLI command
    if len(os.Args) > 1 {
        command := os.Args[1]
        args := os.Args[2:]

        if err := cliInstance.Run(command, args); err == nil {
            return // Command executed
        }
    }

    // Start server...
}
```

## Example: Custom Health Check Command

```go
func HealthCheckCommand(args []string, appCLI *cli.CLI) error {
    logger := appCLI.Context["logger"].(*slog.Logger)

    // Check various components
    checks := map[string]bool{
        "database": checkDatabase(),
        "cache":    checkCache(),
        "api":      checkAPI(),
    }

    allHealthy := true
    for name, healthy := range checks {
        status := "OK"
        if !healthy {
            status = "FAILED"
            allHealthy = false
        }
        fmt.Printf("%s: %s\n", name, status)
    }

    if allHealthy {
        fmt.Println("\nAll systems operational")
        return nil
    }

    return fmt.Errorf("some health checks failed")
}

// Register
cliInstance.Register("health", cli.Command{
    Handler: HealthCheckCommand,
    Usage:   "health",
    Summary: "Run health checks on all components",
})
```

## Help Text

Statigo CLI automatically generates help text:

```bash
./statigo help
```

Output:
```
Available commands:
  prerender    Pre-render all pages to cache
  clear-cache  Clear all cached pages
  health       Run health checks on all components
  help         Show this help message

Use "statigo help <command>" for more information about a command.
```

## Integration with Build Process

Add to your `Makefile`:

```makefile
.PHONY: build prerender run

build:
	go build -o statigo

prerender: build
	./statigo prerender

run: build
	./statigo

deploy: build prerender
	# Deployment commands...
```

Then:
```bash
make prerender  # Build and pre-render
make deploy     # Deploy with warm cache
```
