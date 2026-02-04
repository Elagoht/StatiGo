// Package cache provides a two-tier caching system for the Statigo framework.
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"
)

// Entry represents a cached page with metadata.
type Entry struct {
	Content     []byte    // Brotli-compressed HTML stored in memory
	RenderedAt  time.Time // When this entry was last rendered
	Strategy    string    // Caching strategy: "static", "incremental", "dynamic", "immutable"
	ETag        string    // HTTP ETag for cache validation
	RequestPath string    // Original request path for eager revalidation
	Generation  int64     // Generation number - increments on each update
	stale       atomic.Bool
}

// NewEntry creates a new cache entry with the given content and strategy.
func NewEntry(content []byte, strategy, requestPath string) *Entry {
	now := time.Now()
	entry := &Entry{
		Content:     content,
		RenderedAt:  now,
		Strategy:    strategy,
		ETag:        generateETag(content, 1, now),
		RequestPath: requestPath,
		Generation:  1,
	}
	entry.stale.Store(false)
	return entry
}

// IsStale returns whether this entry has been marked as stale.
func (e *Entry) IsStale() bool {
	return e.stale.Load()
}

// MarkStale marks this entry as stale (needs revalidation).
func (e *Entry) MarkStale() {
	e.stale.Store(true)
}

// MarkFresh marks this entry as fresh (valid cache).
func (e *Entry) MarkFresh() {
	e.stale.Store(false)
}

// Update updates the entry content and marks it as fresh.
func (e *Entry) Update(content []byte, requestPath string) {
	e.Content = content
	e.RenderedAt = time.Now()
	e.Generation++
	e.ETag = generateETag(content, e.Generation, e.RenderedAt)
	if requestPath != "" {
		e.RequestPath = requestPath
	}
	e.MarkFresh()
}

// ShouldRevalidate determines if this entry should be revalidated based on strategy.
func (e *Entry) ShouldRevalidate() bool {
	// Immutable entries never revalidate
	if e.Strategy == "immutable" {
		return false
	}

	// If marked stale, should revalidate
	if e.IsStale() {
		return true
	}

	// Incremental entries revalidate if older than 24 hours
	if e.Strategy == "incremental" {
		return time.Since(e.RenderedAt) > 24*time.Hour
	}

	// Static entries only revalidate when explicitly marked stale
	return false
}

// generateETag generates an ETag from content, generation, and timestamp using SHA-256.
func generateETag(content []byte, generation int64, renderedAt time.Time) string {
	h := sha256.New()
	h.Write(content)
	h.Write([]byte(fmt.Sprintf("%d:%d", generation, renderedAt.Unix())))
	return hex.EncodeToString(h.Sum(nil))
}
