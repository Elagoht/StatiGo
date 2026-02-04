package cache

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/andybalholm/brotli"
)

// Storage handles file I/O operations for cache.
type Storage struct {
	baseDir string
	mu      sync.RWMutex // Protects file operations
}

// NewStorage creates a new storage instance.
func NewStorage(baseDir string) (*Storage, error) {
	// Ensure cache directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Storage{
		baseDir: baseDir,
	}, nil
}

// Write stores cache entry to disk in both formats.
func (s *Storage) Write(cacheKey string, compressedContent, uncompressedContent []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fileName := getCacheFileName(cacheKey)

	// Write brotli-compressed version
	brPath := filepath.Join(s.baseDir, fileName+".br")
	if err := os.WriteFile(brPath, compressedContent, 0644); err != nil {
		return fmt.Errorf("failed to write brotli cache file: %w", err)
	}

	// Write uncompressed version
	htmlPath := filepath.Join(s.baseDir, fileName+".html")
	if err := os.WriteFile(htmlPath, uncompressedContent, 0644); err != nil {
		return fmt.Errorf("failed to write HTML cache file: %w", err)
	}

	return nil
}

// ReadBrotli reads brotli-compressed content from disk.
func (s *Storage) ReadBrotli(cacheKey string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fileName := getCacheFileName(cacheKey)
	brPath := filepath.Join(s.baseDir, fileName+".br")

	content, err := os.ReadFile(brPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read brotli cache file: %w", err)
	}

	return content, nil
}

// ReadHTML reads uncompressed HTML content from disk.
func (s *Storage) ReadHTML(cacheKey string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fileName := getCacheFileName(cacheKey)
	htmlPath := filepath.Join(s.baseDir, fileName+".html")

	content, err := os.ReadFile(htmlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read HTML cache file: %w", err)
	}

	return content, nil
}

// Exists checks if cache files exist for the given key.
func (s *Storage) Exists(cacheKey string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	fileName := getCacheFileName(cacheKey)
	brPath := filepath.Join(s.baseDir, fileName+".br")

	_, err := os.Stat(brPath)
	return err == nil
}

// Delete removes cache files for the given key.
func (s *Storage) Delete(cacheKey string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	fileName := getCacheFileName(cacheKey)

	// Delete both files, ignore errors if files don't exist
	brPath := filepath.Join(s.baseDir, fileName+".br")
	htmlPath := filepath.Join(s.baseDir, fileName+".html")

	_ = os.Remove(brPath)
	_ = os.Remove(htmlPath)

	return nil
}

// CompressBrotli compresses content using brotli.
func CompressBrotli(content []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := brotli.NewWriterLevel(&buf, brotli.DefaultCompression)

	if _, err := writer.Write(content); err != nil {
		writer.Close()
		return nil, fmt.Errorf("failed to compress content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close brotli writer: %w", err)
	}

	return buf.Bytes(), nil
}

// DecompressBrotli decompresses brotli-compressed content.
func DecompressBrotli(compressed []byte) ([]byte, error) {
	reader := brotli.NewReader(bytes.NewReader(compressed))

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(reader); err != nil {
		return nil, fmt.Errorf("failed to decompress content: %w", err)
	}

	return buf.Bytes(), nil
}

// getCacheFileName converts cache key to safe file name.
func getCacheFileName(cacheKey string) string {
	// Replace "/" with "_"
	name := strings.ReplaceAll(cacheKey, "/", "_")

	// Replace ":" with "_"
	name = strings.ReplaceAll(name, ":", "_")

	// Remove leading "_"
	name = strings.TrimPrefix(name, "_")

	return name
}
