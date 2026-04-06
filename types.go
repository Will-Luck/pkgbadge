package main

import "sync"

type PackageStats struct {
	Owner         string
	Package       string
	TotalPulls    int
	LatestVersion string
	Architectures []string
	PlatformSizes map[string]int64 // key: "linux/amd64" or "" for unknown platform; value: compressed bytes
	ScrapedAt     int64  // unix timestamp
}

type Cache struct {
	mu    sync.RWMutex
	stats map[string]*PackageStats // key: "owner/package"
}

func NewCache() *Cache {
	return &Cache{stats: make(map[string]*PackageStats)}
}

func (c *Cache) Get(key string) *PackageStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.stats[key]
}

func (c *Cache) Set(key string, s *PackageStats) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stats[key] = s
}

// BadgeResponse is the shields.io endpoint-badge schema: https://shields.io/badges/endpoint-badge
type BadgeResponse struct {
	SchemaVersion int    `json:"schemaVersion"`
	Label         string `json:"label"`
	Message       string `json:"message"`
	Color         string `json:"color"`
}
