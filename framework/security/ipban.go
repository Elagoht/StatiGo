// Package security provides IP banning and security utilities.
package security

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// BanEntry represents a banned IP with metadata.
type BanEntry struct {
	IP        string    `json:"ip"`
	Reason    string    `json:"reason"`
	BannedAt  time.Time `json:"bannedAt"`
	UserAgent string    `json:"userAgent,omitempty"`
	Path      string    `json:"path,omitempty"`
}

// IPBanList manages a list of banned IP addresses.
type IPBanList struct {
	mu       sync.RWMutex
	banned   map[string]*BanEntry
	filePath string
	logger   *slog.Logger
}

// NewIPBanList creates a new IP ban list manager.
func NewIPBanList(filePath string, logger *slog.Logger) (*IPBanList, error) {
	banList := &IPBanList{
		banned:   make(map[string]*BanEntry),
		filePath: filePath,
		logger:   logger,
	}

	// Load existing ban list from file
	if err := banList.load(); err != nil {
		logger.Warn("Failed to load existing ban list, starting with empty list", "error", err)
	}

	return banList, nil
}

// BanIP adds an IP address to the ban list.
func (bl *IPBanList) BanIP(ip, reason, userAgent, path string) error {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	entry := &BanEntry{
		IP:        ip,
		Reason:    reason,
		BannedAt:  time.Now(),
		UserAgent: userAgent,
		Path:      path,
	}

	bl.banned[ip] = entry
	bl.logger.Warn("IP banned",
		"ip", ip,
		"reason", reason,
		"path", path,
		"user_agent", userAgent,
	)

	// Save to file
	return bl.save()
}

// IsBanned checks if an IP address is banned.
func (bl *IPBanList) IsBanned(ip string) bool {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	_, banned := bl.banned[ip]
	return banned
}

// UnbanIP removes an IP address from the ban list.
func (bl *IPBanList) UnbanIP(ip string) error {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	delete(bl.banned, ip)
	bl.logger.Info("IP unbanned", "ip", ip)

	return bl.save()
}

// Count returns the number of banned IPs.
func (bl *IPBanList) Count() int {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	return len(bl.banned)
}

// GetAll returns all banned IP entries.
func (bl *IPBanList) GetAll() []*BanEntry {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	entries := make([]*BanEntry, 0, len(bl.banned))
	for _, entry := range bl.banned {
		entries = append(entries, entry)
	}
	return entries
}

// save persists the ban list to disk.
func (bl *IPBanList) save() error {
	file, err := os.Create(bl.filePath)
	if err != nil {
		return fmt.Errorf("failed to create ban list file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	entries := make([]*BanEntry, 0, len(bl.banned))
	for _, entry := range bl.banned {
		entries = append(entries, entry)
	}

	if err := encoder.Encode(entries); err != nil {
		return fmt.Errorf("failed to encode ban list: %w", err)
	}

	return nil
}

// load reads the ban list from disk.
func (bl *IPBanList) load() error {
	file, err := os.Open(bl.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, that's fine
		}
		return fmt.Errorf("failed to open ban list file: %w", err)
	}
	defer file.Close()

	var entries []*BanEntry
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&entries); err != nil {
		return fmt.Errorf("failed to decode ban list: %w", err)
	}

	for _, entry := range entries {
		bl.banned[entry.IP] = entry
	}

	bl.logger.Info("Loaded ban list from file", "count", len(entries), "file", bl.filePath)
	return nil
}

// GetClientIP extracts the real client IP from the request.
// It checks X-Forwarded-For, X-Real-IP headers, and falls back to RemoteAddr.
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (common with reverse proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		ip := strings.TrimSpace(xri)
		if net.ParseIP(ip) != nil {
			return ip
		}
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return ip
}
