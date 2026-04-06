package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var httpClient = &http.Client{Timeout: 30 * time.Second}

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

	if stats.TotalPulls == 0 && stats.LatestVersion == "" && len(stats.Architectures) == 0 {
		return stats, fmt.Errorf("no data extracted from page, HTML structure may have changed")
	}

	return stats, nil
}

type PackageRef struct {
	Owner   string
	Repo    string // may differ from Package (e.g. Docker-Sentinel vs docker-sentinel)
	Package string
}

// Key returns the cache key for this package (always lowercase).
func (r PackageRef) Key() string {
	return strings.ToLower(r.Owner + "/" + r.Package)
}

func fetchPackagePage(ctx context.Context, ref PackageRef) (string, error) {
	url := fmt.Sprintf("https://github.com/%s/%s/pkgs/container/%s",
		ref.Owner, ref.Repo, ref.Package)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "pkgbadge/1.0 (github.com/Will-Luck/pkgbadge)")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d from %s", resp.StatusCode, url)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20)) // 2 MiB cap
	if err != nil {
		return "", err
	}
	return string(body), nil
}

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

		sizes, err := fetchImageSizes(ctx, ref.Owner, ref.Package, stats.LatestVersion)
		if err != nil {
			log.Warn("size fetch failed", "package", ref.Key(), "error", err)
		} else {
			stats.PlatformSizes = sizes
		}

		stats.ScrapedAt = time.Now().Unix()
		cache.Set(ref.Key(), stats)
		log.Info("scraped", "package", ref.Key(), "pulls", stats.TotalPulls, "version", stats.LatestVersion)
	}
}

// OCI registry types (private, for JSON parsing only).

type ociTokenResponse struct {
	Token string `json:"token"`
}

type ociManifest struct {
	MediaType string          `json:"mediaType"`
	Manifests []ociDescriptor `json:"manifests,omitempty"`
	Layers    []ociDescriptor `json:"layers,omitempty"`
}

type ociDescriptor struct {
	MediaType string       `json:"mediaType"`
	Digest    string       `json:"digest"`
	Size      int64        `json:"size"`
	Platform  *ociPlatform `json:"platform,omitempty"`
}

type ociPlatform struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
}

func parseManifestSizes(body []byte, fetchManifest func(digest string) ([]byte, error)) (map[string]int64, error) {
	var m ociManifest
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, fmt.Errorf("unmarshal manifest: %w", err)
	}

	if len(m.Manifests) > 0 {
		return parseIndexSizes(m.Manifests, fetchManifest)
	}
	if len(m.Layers) > 0 {
		return parseSingleSizes(m.Layers), nil
	}
	return nil, fmt.Errorf("manifest has no manifests[] or layers[]")
}

func parseIndexSizes(descriptors []ociDescriptor, fetchManifest func(string) ([]byte, error)) (map[string]int64, error) {
	sizes := make(map[string]int64)
	for _, d := range descriptors {
		if d.Platform == nil || d.Platform.OS == "" || d.Platform.OS == "unknown" {
			continue
		}
		raw, err := fetchManifest(d.Digest)
		if err != nil {
			continue
		}
		var pm ociManifest
		if err := json.Unmarshal(raw, &pm); err != nil {
			continue
		}
		key := d.Platform.OS + "/" + d.Platform.Architecture
		for _, l := range pm.Layers {
			sizes[key] += l.Size
		}
	}
	if len(sizes) == 0 {
		return nil, fmt.Errorf("no platform manifests resolved")
	}
	return sizes, nil
}

func parseSingleSizes(layers []ociDescriptor) map[string]int64 {
	var total int64
	for _, l := range layers {
		total += l.Size
	}
	return map[string]int64{"": total}
}

const (
	ghcrTokenURL    = "https://ghcr.io/token?scope=repository:%s/%s:pull"
	ghcrManifestURL = "https://ghcr.io/v2/%s/%s/manifests/%s"
	ociAccept       = "application/vnd.oci.image.index.v1+json, " +
		"application/vnd.docker.distribution.manifest.list.v2+json, " +
		"application/vnd.oci.image.manifest.v1+json, " +
		"application/vnd.docker.distribution.manifest.v2+json"
)

func fetchImageSizes(ctx context.Context, owner, pkg, version string) (map[string]int64, error) {
	token, err := fetchGHCRToken(ctx, owner, pkg)
	if err != nil {
		return nil, fmt.Errorf("token: %w", err)
	}

	tag := version
	if tag == "" {
		tag = "latest"
	}

	body, err := fetchManifestByTag(ctx, owner, pkg, tag, token)
	if err != nil && tag != "latest" {
		body, err = fetchManifestByTag(ctx, owner, pkg, "latest", token)
	}
	if err != nil {
		return nil, err
	}

	return parseManifestSizes(body, func(digest string) ([]byte, error) {
		return fetchManifestByTag(ctx, owner, pkg, digest, token)
	})
}

func fetchGHCRToken(ctx context.Context, owner, pkg string) (string, error) {
	url := fmt.Sprintf(ghcrTokenURL, strings.ToLower(owner), strings.ToLower(pkg))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint HTTP %d", resp.StatusCode)
	}

	var tr ociTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", fmt.Errorf("decode token: %w", err)
	}
	return tr.Token, nil
}

func fetchManifestByTag(ctx context.Context, owner, pkg, ref, token string) ([]byte, error) {
	url := fmt.Sprintf(ghcrManifestURL, strings.ToLower(owner), strings.ToLower(pkg), ref)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", ociAccept)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("manifest HTTP %d for %s", resp.StatusCode, ref)
	}

	return io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1 MiB cap
}
