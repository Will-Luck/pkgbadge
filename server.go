package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
)

// newMux returns an http.Handler that serves badge endpoints.
func newMux(cache *Cache) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleBadge(w, r, cache)
	})
	return mux
}

func handleBadge(w http.ResponseWriter, r *http.Request, cache *Cache) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		http.Error(w, `{"error":"expected /owner/package/badge.json"}`, http.StatusNotFound)
		return
	}

	owner := parts[0]
	pkg := parts[1]
	badgeFile := parts[2]

	ext := path.Ext(badgeFile)
	if ext != ".json" {
		http.Error(w, `{"error":"expected .json extension"}`, http.StatusNotFound)
		return
	}
	badgeType := strings.TrimSuffix(badgeFile, ext)

	key := strings.ToLower(owner + "/" + pkg)
	stats := cache.Get(key)
	if stats == nil {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "package not configured"})
		return
	}

	badge, ok := buildBadge(badgeType, stats)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": fmt.Sprintf("unknown badge type: %s", badgeType)})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "max-age=3600")
	_ = json.NewEncoder(w).Encode(badge)
}

func buildBadge(badgeType string, stats *PackageStats) (BadgeResponse, bool) {
	switch badgeType {
	case "pulls":
		return BadgeResponse{
			SchemaVersion: 1,
			Label:         "ghcr pulls",
			Message:       formatCount(stats.TotalPulls),
			Color:         "blue",
		}, true

	case "version":
		msg := stats.LatestVersion
		if msg == "" {
			msg = "unknown"
		}
		return BadgeResponse{
			SchemaVersion: 1,
			Label:         "version",
			Message:       msg,
			Color:         "green",
		}, true

	case "size":
		return BadgeResponse{
			SchemaVersion: 1,
			Label:         "image size",
			Message:       formatBytes(stats.SizeBytes),
			Color:         "blue",
		}, true

	case "arch":
		arches := make([]string, len(stats.Architectures))
		for i, a := range stats.Architectures {
			arches[i] = strings.TrimPrefix(a, "linux/")
		}
		msg := strings.Join(arches, " | ")
		if msg == "" {
			msg = "unknown"
		}
		return BadgeResponse{
			SchemaVersion: 1,
			Label:         "platforms",
			Message:       msg,
			Color:         "blue",
		}, true

	default:
		return BadgeResponse{}, false
	}
}

func formatCount(n int) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

func formatBytes(b int64) string {
	switch {
	case b <= 0:
		return "unknown"
	case b >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(b)/(1<<30))
	case b >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(b)/(1<<20))
	case b >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(b)/(1<<10))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
