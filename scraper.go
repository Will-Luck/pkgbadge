package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	// Matches: <h3 title="433">433</h3> after "Total downloads"
	reDownloads = regexp.MustCompile(`Total downloads</span>\s*<h3[^>]*>([0-9,]+)</h3>`)

	// Matches: <span class="text-normal h2 mr-1 color-fg-muted" >2.11.1</span>
	reVersion = regexp.MustCompile(`class="text-normal h2[^"]*color-fg-muted"[^>]*>([^<]+)</span>`)

	// Matches: <small>linux/amd64</small>
	reArch = regexp.MustCompile(`<small>(linux/[a-z0-9]+)</small>`)
)

// parsePackagePage extracts stats from the GitHub packages HTML page.
func parsePackagePage(html, owner, pkg string) (*PackageStats, error) {
	stats := &PackageStats{
		Owner:   owner,
		Package: pkg,
	}

	if m := reDownloads.FindStringSubmatch(html); len(m) > 1 {
		n, _ := strconv.Atoi(strings.ReplaceAll(m[1], ",", ""))
		stats.TotalPulls = n
	}

	if m := reVersion.FindStringSubmatch(html); len(m) > 1 {
		stats.LatestVersion = strings.TrimSpace(m[1])
	}

	seen := make(map[string]bool)
	for _, m := range reArch.FindAllStringSubmatch(html, -1) {
		arch := m[1]
		if !seen[arch] {
			stats.Architectures = append(stats.Architectures, arch)
			seen[arch] = true
		}
	}

	return stats, nil
}

// PackageRef identifies a configured package to scrape.
type PackageRef struct {
	Owner   string
	Repo    string // may differ from Package (e.g. Docker-Sentinel vs docker-sentinel)
	Package string
}

// Key returns the cache key for this package (always lowercase).
func (r PackageRef) Key() string {
	return strings.ToLower(r.Owner + "/" + r.Package)
}

// fetchPackagePage downloads the GitHub packages HTML page.
func fetchPackagePage(ctx context.Context, ref PackageRef) (string, error) {
	url := fmt.Sprintf("https://github.com/%s/%s/pkgs/container/%s",
		ref.Owner, ref.Repo, ref.Package)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// scrapeAll fetches and parses stats for every configured package.
func scrapeAll(ctx context.Context, packages []PackageRef, cache *Cache, log *slog.Logger) {
	for _, ref := range packages {
		html, err := fetchPackagePage(ctx, ref)
		if err != nil {
			log.Warn("scrape failed, keeping stale data", "package", ref.Key(), "error", err)
			continue
		}
		stats, err := parsePackagePage(html, ref.Owner, ref.Package)
		if err != nil {
			log.Warn("parse failed", "package", ref.Key(), "error", err)
			continue
		}
		stats.ScrapedAt = time.Now().Unix()
		cache.Set(ref.Key(), stats)
		log.Info("scraped", "package", ref.Key(), "pulls", stats.TotalPulls, "version", stats.LatestVersion)
	}
}
