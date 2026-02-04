package cli

import (
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"statigo/framework/cache"
)

// PrerenderCommandConfig contains configuration for the prerender command.
type PrerenderCommandConfig struct {
	ConfigFS     fs.FS
	RoutesFile   string
	Languages    []string
	Router       http.Handler
	CacheManager *cache.Manager
	Logger       *slog.Logger
}

// NewPrerenderCommand creates a new prerender command.
func NewPrerenderCommand(config PrerenderCommandConfig) *Command {
	return &Command{
		Name:    "prerender",
		Aliases: []string{"pre-render", "bake", "warm", "prepare", "cache-all"},
		Desc:    "Pre-render and cache all cacheable pages",
		Run: func() error {
			config.Logger.Info("Starting cache pre-rendering...")

			if err := config.CacheManager.Bootstrap(context.Background(), cache.RebuildConfig{
				ConfigFS:   config.ConfigFS,
				RoutesFile: config.RoutesFile,
				Languages:  config.Languages,
				Router:     config.Router,
				Logger:     config.Logger,
			}); err != nil {
				return fmt.Errorf("pre-rendering failed: %w", err)
			}

			config.Logger.Info("Cache pre-rendering completed successfully")
			return nil
		},
	}
}

// ClearCacheCommandConfig contains configuration for the clear-cache command.
type ClearCacheCommandConfig struct {
	CacheDir string
	Logger   *slog.Logger
}

// NewClearCacheCommand creates a new clear-cache command.
func NewClearCacheCommand(config ClearCacheCommandConfig) *Command {
	return &Command{
		Name:    "clear-cache",
		Aliases: []string{"invalidate"},
		Desc:    "Clear all cached files",
		Run: func() error {
			config.Logger.Info("Clearing cache...", slog.String("dir", config.CacheDir))

			// Check if cache directory exists
			if _, err := os.Stat(config.CacheDir); os.IsNotExist(err) {
				config.Logger.Info("Cache directory does not exist, nothing to clear")
				return nil
			}

			// Count files before deletion
			count := 0
			err := filepath.Walk(config.CacheDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !info.IsDir() {
					count++
				}
				return nil
			})

			if err != nil {
				return fmt.Errorf("failed to count cache files: %w", err)
			}

			// Remove all files in cache directory
			err = filepath.Walk(config.CacheDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				// Skip the cache directory itself
				if path == config.CacheDir {
					return nil
				}

				// Remove file or directory
				if err := os.RemoveAll(path); err != nil {
					config.Logger.Warn("Failed to remove cache file",
						slog.String("path", path),
						slog.String("error", err.Error()),
					)
				}

				return nil
			})

			if err != nil {
				return fmt.Errorf("failed to clear cache: %w", err)
			}

			config.Logger.Info("Cache cleared successfully", slog.Int("files_removed", count))
			return nil
		},
	}
}
