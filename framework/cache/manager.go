package cache

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"
)

// Manager handles cache operations with memory and file storage.
type Manager struct {
	entries sync.Map // Thread-safe map of cache entries (key: cacheKey, value: *Entry)
	storage *Storage
	logger  *slog.Logger
	router  http.Handler
	mu      sync.RWMutex
}

// NewManager creates a new cache manager.
func NewManager(cacheDir string, logger *slog.Logger) (*Manager, error) {
	storage, err := NewStorage(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	return &Manager{
		storage: storage,
		logger:  logger,
	}, nil
}

// Get retrieves a cache entry from memory or disk.
func (m *Manager) Get(cacheKey string) (*Entry, bool) {
	// Try memory cache first
	if entry, ok := m.entries.Load(cacheKey); ok {
		return entry.(*Entry), true
	}

	// Try loading from disk
	if m.storage.Exists(cacheKey) {
		entry, err := m.loadFromDisk(cacheKey)
		if err != nil {
			m.logger.Warn("failed to load cache from disk",
				slog.String("key", cacheKey),
				slog.String("error", err.Error()),
			)
			return nil, false
		}

		// Store in memory for faster subsequent access
		m.entries.Store(cacheKey, entry)
		return entry, true
	}

	return nil, false
}

// Set stores a cache entry in memory and disk.
func (m *Manager) Set(cacheKey string, uncompressedContent []byte, strategy, requestPath string) error {
	return m.set(cacheKey, uncompressedContent, strategy, requestPath, false)
}

// SetSync stores a cache entry in memory and disk synchronously.
func (m *Manager) SetSync(cacheKey string, uncompressedContent []byte, strategy, requestPath string) error {
	return m.set(cacheKey, uncompressedContent, strategy, requestPath, true)
}

// set is the internal method that handles cache storage.
func (m *Manager) set(cacheKey string, uncompressedContent []byte, strategy, requestPath string, sync bool) error {
	// Compress content for memory storage
	compressedContent, err := CompressBrotli(uncompressedContent)
	if err != nil {
		m.logger.Error("failed to compress cache content",
			slog.String("key", cacheKey),
			slog.String("error", err.Error()),
		)
		// Fall back to uncompressed storage
		compressedContent = uncompressedContent
	}

	// Check if entry exists and update it, or create new one
	if existingValue, exists := m.entries.Load(cacheKey); exists {
		// Update existing entry
		existingEntry := existingValue.(*Entry)
		existingEntry.Update(compressedContent, requestPath)

		m.logger.Debug("cache updated",
			slog.String("key", cacheKey),
			slog.String("strategy", strategy),
			slog.String("request_path", requestPath),
			slog.Int64("generation", existingEntry.Generation),
		)
	} else {
		// Create new cache entry
		entry := NewEntry(compressedContent, strategy, requestPath)
		m.entries.Store(cacheKey, entry)

		m.logger.Debug("cache created",
			slog.String("key", cacheKey),
			slog.String("strategy", strategy),
			slog.String("request_path", requestPath),
		)
	}

	// Write to disk
	writeFunc := func() {
		if err := m.storage.Write(cacheKey, compressedContent, uncompressedContent); err != nil {
			m.logger.Error("failed to write cache to disk",
				slog.String("key", cacheKey),
				slog.String("error", err.Error()),
			)
		}
	}

	if sync {
		writeFunc()
	} else {
		go writeFunc()
	}

	return nil
}

// Delete removes a cache entry from memory and disk.
func (m *Manager) Delete(cacheKey string) error {
	m.entries.Delete(cacheKey)

	if err := m.storage.Delete(cacheKey); err != nil {
		return fmt.Errorf("failed to delete cache from disk: %w", err)
	}

	return nil
}

// MarkStale marks cache entries matching the strategy as stale.
func (m *Manager) MarkStale(strategy string, eager bool) int {
	count := 0
	var staleEntries []*Entry

	m.entries.Range(func(key, value interface{}) bool {
		entry := value.(*Entry)

		if entry.Strategy == "immutable" {
			return true
		}

		if entry.Strategy == strategy {
			entry.MarkStale()
			count++

			if eager {
				staleEntries = append(staleEntries, entry)
			}

			m.logger.Debug("marked cache as stale",
				slog.String("key", key.(string)),
				slog.String("strategy", strategy),
			)
		}

		return true
	})

	m.logger.Info("marked caches as stale",
		slog.String("strategy", strategy),
		slog.Int("count", count),
		slog.Bool("eager", eager),
	)

	if eager && len(staleEntries) > 0 {
		go m.eagerRevalidate(staleEntries)
	}

	return count
}

// MarkAllStale marks all cache entries as stale (except immutable).
func (m *Manager) MarkAllStale(eager bool) int {
	count := 0
	var staleEntries []*Entry

	m.entries.Range(func(key, value interface{}) bool {
		entry := value.(*Entry)

		if entry.Strategy == "immutable" {
			return true
		}

		entry.MarkStale()
		count++

		if eager {
			staleEntries = append(staleEntries, entry)
		}

		return true
	})

	m.logger.Info("marked all caches as stale",
		slog.Int("count", count),
		slog.Bool("eager", eager),
	)

	if eager && len(staleEntries) > 0 {
		go m.eagerRevalidate(staleEntries)
	}

	return count
}

// GetCacheKey generates a cache key from canonical path, language, and path params.
func GetCacheKey(canonical, lang string, pathParams map[string]string) string {
	key := canonical

	// Replace {param} placeholders with actual values
	for param, value := range pathParams {
		key = strings.ReplaceAll(key, "{"+param+"}", value)
	}

	return key + ":" + lang
}

// SetRouter sets the HTTP router for eager revalidation.
func (m *Manager) SetRouter(router http.Handler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.router = router
}

// GetDecompressedContent decompresses and returns the cached HTML content.
func GetDecompressedContent(entry *Entry) ([]byte, error) {
	return DecompressBrotli(entry.Content)
}

// loadFromDisk loads a cache entry from disk.
func (m *Manager) loadFromDisk(cacheKey string) (*Entry, error) {
	compressedContent, err := m.storage.ReadBrotli(cacheKey)
	if err != nil {
		return nil, fmt.Errorf("failed to read brotli cache: %w", err)
	}

	renderedAt := time.Now()
	entry := &Entry{
		Content:    compressedContent,
		RenderedAt: renderedAt,
		Strategy:   "static",
		ETag:       generateETag(compressedContent, 1, renderedAt),
		Generation: 1,
	}
	entry.stale.Store(false)

	m.logger.Debug("loaded cache from disk",
		slog.String("key", cacheKey),
	)

	return entry, nil
}

// eagerRevalidate re-renders all stale entries in the background.
func (m *Manager) eagerRevalidate(entries []*Entry) {
	m.mu.RLock()
	router := m.router
	m.mu.RUnlock()

	if router == nil {
		m.logger.Warn("eager revalidation skipped - router not set")
		return
	}

	m.logger.Info("starting eager revalidation",
		slog.Int("count", len(entries)),
	)

	start := time.Now()
	successCount := 0
	errorCount := 0

	// Process entries with limited concurrency
	semaphore := make(chan struct{}, 10)
	var wg sync.WaitGroup

	for _, entry := range entries {
		if entry.RequestPath == "" {
			m.logger.Warn("skipping entry with empty request path")
			continue
		}

		wg.Add(1)
		go func(reqPath string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			req := httptest.NewRequest(http.MethodGet, reqPath, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			if rec.Code == http.StatusOK {
				successCount++
			} else {
				errorCount++
			}
		}(entry.RequestPath)
	}

	wg.Wait()

	m.logger.Info("eager revalidation completed",
		slog.Int("total", len(entries)),
		slog.Int("success", successCount),
		slog.Int("errors", errorCount),
		slog.Duration("duration", time.Since(start)),
	)
}
