package main

import "sync"

// PackageStats holds scraped stats for a single GHCR package.
type PackageStats struct {
	Owner         string
	Package       string
	TotalPulls    int
	LatestVersion string
	Architectures []string
	SizeBytes     int64  // from OCI manifest, 0 if unknown
	ScrapedAt     int64  // unix timestamp
}

// Cache is a concurrency-safe store for scraped package stats.
type Cache struct {
	mu    sync.RWMutex
	stats map[string]*PackageStats // key: "owner/package"
}

// NewCache returns an initialised Cache.
func NewCache() *Cache {
	return &Cache{stats: make(map[string]*PackageStats)}
}

// Get returns the stats for a package, or nil if not found.
func (c *Cache) Get(key string) *PackageStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats[key]
}

// Set stores stats for a package.
func (c *Cache) Set(key string, s *PackageStats) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stats[key] = s
}

// BadgeResponse is the shields.io endpoint badge schema.
// See: https://shields.io/badges/endpoint-badge
type BadgeResponse struct {
	SchemaVersion int    `json:"schemaVersion"`
	Label         string `json:"label"`
	Message       string `json:"message"`
	Color         string `json:"color"`
}
