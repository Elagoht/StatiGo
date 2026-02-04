// Package models provides shared data models for the Statigo framework.
package models

import "time"

// SitemapEntry represents a single entry in a sitemap.
type SitemapEntry struct {
	Title     string    `json:"title"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Cover     string    `json:"cover"`
}
