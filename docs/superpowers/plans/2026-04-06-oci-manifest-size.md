# OCI Manifest Size Badge Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Fetch compressed image sizes from the GHCR OCI registry API and display per-platform sizes on the `size.json` badge.

**Architecture:** Add `fetchImageSizes` to the scraper that hits `ghcr.io/v2/` for anonymous OCI manifest data after the HTML scrape. Extract a testable `parseManifestSizes` helper that takes raw JSON + a digest-fetcher callback so all parsing logic can be tested with fixtures. Update the badge formatter to render per-platform breakdowns.

**Tech Stack:** Go stdlib only (net/http, encoding/json). No new dependencies.

**Spec:** `docs/superpowers/specs/2026-04-06-oci-manifest-size-design.md`

---

### Task 1: Update data model in types.go

**Files:**
- Modify: `types.go:6-14`

- [ ] **Step 1: Replace SizeBytes with PlatformSizes**

Replace the `SizeBytes` field in `PackageStats`:

```go
type PackageStats struct {
	Owner         string
	Package       string
	TotalPulls    int
	LatestVersion string
	Architectures []string
	PlatformSizes map[string]int64 // key: "linux/amd64" or "" for unknown platform; value: compressed bytes
	ScrapedAt     int64            // unix timestamp
}
```

- [ ] **Step 2: Run existing tests to confirm nothing breaks**

Run: `cd /home/lns/pkgbadge && go test ./...`
Expected: PASS. The existing `seedCache()` in `server_test.go` does not set `SizeBytes`, so removing the field is safe. The `scraper_test.go` tests only check pulls/version/arch.

- [ ] **Step 3: Commit**

```bash
cd /home/lns/pkgbadge && git add types.go && git commit -m "refactor: replace SizeBytes with PlatformSizes map"
```

---

### Task 2: Add OCI types and parseManifestSizes helper

**Files:**
- Modify: `scraper.go` (add types and helper after existing code)
- Create: `testdata/manifest-index.json`
- Create: `testdata/manifest-single.json`
- Create: `testdata/manifest-amd64.json`
- Create: `testdata/manifest-arm64.json`
- Modify: `scraper_test.go` (add parsing tests)

- [ ] **Step 1: Create test fixtures**

`testdata/manifest-index.json` -- multi-arch OCI index with 2 real platforms + 2 attestation entries:

```json
{
  "mediaType": "application/vnd.oci.image.index.v1+json",
  "schemaVersion": 2,
  "manifests": [
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "digest": "sha256:amd64digest",
      "size": 2197,
      "platform": { "os": "linux", "architecture": "amd64" }
    },
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "digest": "sha256:arm64digest",
      "size": 2197,
      "platform": { "os": "linux", "architecture": "arm64" }
    },
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "digest": "sha256:attestation1",
      "size": 565,
      "platform": { "os": "unknown", "architecture": "unknown" }
    },
    {
      "mediaType": "application/vnd.oci.image.manifest.v1+json",
      "digest": "sha256:attestation2",
      "size": 565,
      "platform": { "os": "unknown", "architecture": "unknown" }
    }
  ]
}
```

`testdata/manifest-amd64.json` -- platform manifest for amd64:

```json
{
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "schemaVersion": 2,
  "layers": [
    { "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip", "size": 3000000 },
    { "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip", "size": 7000000 }
  ]
}
```

`testdata/manifest-arm64.json` -- platform manifest for arm64:

```json
{
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "schemaVersion": 2,
  "layers": [
    { "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip", "size": 2500000 },
    { "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip", "size": 6500000 }
  ]
}
```

`testdata/manifest-single.json` -- single image manifest (no index wrapper):

```json
{
  "mediaType": "application/vnd.oci.image.manifest.v1+json",
  "schemaVersion": 2,
  "layers": [
    { "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip", "size": 4000000 },
    { "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip", "size": 8000000 }
  ]
}
```

- [ ] **Step 2: Write failing tests for parseManifestSizes**

Add to `scraper_test.go`:

```go
func TestParseManifestSizes_Index(t *testing.T) {
	indexBody := loadFixture(t, "manifest-index.json")
	amd64Body := loadFixture(t, "manifest-amd64.json")
	arm64Body := loadFixture(t, "manifest-arm64.json")

	fetcher := func(digest string) ([]byte, error) {
		switch digest {
		case "sha256:amd64digest":
			return []byte(amd64Body), nil
		case "sha256:arm64digest":
			return []byte(arm64Body), nil
		default:
			return nil, fmt.Errorf("unexpected digest: %s", digest)
		}
	}

	sizes, err := parseManifestSizes([]byte(indexBody), fetcher)
	if err != nil {
		t.Fatal(err)
	}
	if len(sizes) != 2 {
		t.Fatalf("got %d platforms, want 2", len(sizes))
	}
	if sizes["linux/amd64"] != 10000000 {
		t.Errorf("amd64 = %d, want 10000000", sizes["linux/amd64"])
	}
	if sizes["linux/arm64"] != 9000000 {
		t.Errorf("arm64 = %d, want 9000000", sizes["linux/arm64"])
	}
}

func TestParseManifestSizes_Single(t *testing.T) {
	body := loadFixture(t, "manifest-single.json")

	sizes, err := parseManifestSizes([]byte(body), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(sizes) != 1 {
		t.Fatalf("got %d entries, want 1", len(sizes))
	}
	if sizes[""] != 12000000 {
		t.Errorf("size = %d, want 12000000", sizes[""])
	}
}
```

Add `"fmt"` to the imports if not already present.

Run: `cd /home/lns/pkgbadge && go test -run TestParseManifestSizes -v`
Expected: FAIL -- `parseManifestSizes` undefined.

- [ ] **Step 3: Add OCI types and parseManifestSizes to scraper.go**

Add after the existing `scrapeAll` function at the bottom of `scraper.go`. Add `"encoding/json"` to the imports.

```go
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

// parseManifestSizes parses an OCI manifest (index or single) and returns
// per-platform compressed layer sizes. For an index, fetchManifest is called
// for each real platform digest. For a single manifest, fetchManifest is unused.
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
			return nil, fmt.Errorf("fetch %s: %w", d.Digest, err)
		}
		var pm ociManifest
		if err := json.Unmarshal(raw, &pm); err != nil {
			return nil, fmt.Errorf("unmarshal platform manifest %s: %w", d.Digest, err)
		}
		key := d.Platform.OS + "/" + d.Platform.Architecture
		for _, l := range pm.Layers {
			sizes[key] += l.Size
		}
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
```

- [ ] **Step 4: Run tests to verify they pass**

Run: `cd /home/lns/pkgbadge && go test -run TestParseManifestSizes -v`
Expected: PASS -- both `TestParseManifestSizes_Index` and `TestParseManifestSizes_Single` green.

- [ ] **Step 5: Commit**

```bash
cd /home/lns/pkgbadge && git add scraper.go scraper_test.go testdata/manifest-*.json && git commit -m "feat: add OCI manifest parsing with per-platform sizes"
```

---

### Task 3: Add fetchImageSizes to the scraper

**Files:**
- Modify: `scraper.go` (add `fetchImageSizes`, wire into `scrapeAll`)

- [ ] **Step 1: Add fetchImageSizes**

Add below `parseSingleSizes` in `scraper.go`:

```go
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
```

- [ ] **Step 2: Wire fetchImageSizes into scrapeAll**

Replace the current `scrapeAll` function body:

```go
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
```

- [ ] **Step 3: Verify it compiles**

Run: `cd /home/lns/pkgbadge && go build ./...`
Expected: Build succeeds with no errors.

- [ ] **Step 4: Commit**

```bash
cd /home/lns/pkgbadge && git add scraper.go && git commit -m "feat: fetch image sizes from GHCR OCI registry API"
```

---

### Task 4: Update badge formatter in server.go

**Files:**
- Modify: `server.go:58-90` (buildBadge size case)
- Modify: `server_test.go` (add size badge tests, update seedCache)

- [ ] **Step 1: Write failing tests for size badge formatting**

Add to `server_test.go`. First add `"sort"` to imports.

Update `seedCache` to include `PlatformSizes`:

```go
func seedCache() *Cache {
	c := NewCache()
	c.Set("will-luck/docker-sentinel", &PackageStats{
		Owner:         "Will-Luck",
		Package:       "docker-sentinel",
		TotalPulls:    433,
		LatestVersion: "2.11.1",
		Architectures: []string{"linux/amd64", "linux/arm64"},
		PlatformSizes: map[string]int64{
			"linux/amd64": 86476951,
			"linux/arm64": 83000000,
		},
		ScrapedAt: 1710000000,
	})
	return c
}
```

Add the test functions:

```go
func TestBuildBadge_Size_MultiPlatform(t *testing.T) {
	stats := &PackageStats{
		PlatformSizes: map[string]int64{
			"linux/amd64": 86476951,
			"linux/arm64": 83000000,
		},
	}
	badge, ok := buildBadge("size", stats)
	if !ok {
		t.Fatal("buildBadge returned false")
	}
	want := "82.5 MB (amd64) | 79.2 MB (arm64)"
	if badge.Message != want {
		t.Errorf("message = %q, want %q", badge.Message, want)
	}
}

func TestBuildBadge_Size_SinglePlatformLabelled(t *testing.T) {
	stats := &PackageStats{
		PlatformSizes: map[string]int64{
			"linux/amd64": 86476951,
		},
	}
	badge, ok := buildBadge("size", stats)
	if !ok {
		t.Fatal("buildBadge returned false")
	}
	want := "82.5 MB (amd64)"
	if badge.Message != want {
		t.Errorf("message = %q, want %q", badge.Message, want)
	}
}

func TestBuildBadge_Size_SingleUnknownPlatform(t *testing.T) {
	stats := &PackageStats{
		PlatformSizes: map[string]int64{
			"": 12000000,
		},
	}
	badge, ok := buildBadge("size", stats)
	if !ok {
		t.Fatal("buildBadge returned false")
	}
	want := "11.4 MB"
	if badge.Message != want {
		t.Errorf("message = %q, want %q", badge.Message, want)
	}
}

func TestBuildBadge_Size_NilMap(t *testing.T) {
	stats := &PackageStats{}
	badge, ok := buildBadge("size", stats)
	if !ok {
		t.Fatal("buildBadge returned false")
	}
	if badge.Message != "unknown" {
		t.Errorf("message = %q, want %q", badge.Message, "unknown")
	}
}
```

Run: `cd /home/lns/pkgbadge && go test -run TestBuildBadge_Size -v`
Expected: FAIL -- the current `buildBadge` size case still calls `formatBytes(stats.SizeBytes)` which no longer exists.

- [ ] **Step 2: Update buildBadge size case**

Replace the `case "size":` block in `buildBadge` (server.go). Add `"sort"` and `"strings"` to imports if not already there.

```go
	case "size":
		return BadgeResponse{
			SchemaVersion: 1,
			Label:         "image size",
			Message:       formatSizeMessage(stats.PlatformSizes),
			Color:         "blue",
		}, true
```

Add the `formatSizeMessage` function after `formatBytes`:

```go
func formatSizeMessage(sizes map[string]int64) string {
	if len(sizes) == 0 {
		return "unknown"
	}

	// Single entry with empty key: unknown platform, plain size.
	if len(sizes) == 1 {
		for k, v := range sizes {
			if k == "" {
				return formatBytes(v)
			}
			return formatBytes(v) + " (" + strings.TrimPrefix(k, "linux/") + ")"
		}
	}

	// Multiple platforms: sorted alphabetically by arch.
	keys := make([]string, 0, len(sizes))
	for k := range sizes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = formatBytes(sizes[k]) + " (" + strings.TrimPrefix(k, "linux/") + ")"
	}
	return strings.Join(parts, " | ")
}
```

- [ ] **Step 3: Run tests to verify they pass**

Run: `cd /home/lns/pkgbadge && go test ./... -v`
Expected: ALL PASS -- all existing tests plus the 4 new size tests.

- [ ] **Step 4: Commit**

```bash
cd /home/lns/pkgbadge && git add server.go server_test.go && git commit -m "feat: per-platform size badge formatting"
```

---

### Task 5: Update documentation

**Files:**
- Modify: `README.md`
- Modify: `wiki/Home.md`
- Modify: `wiki/Badge-Usage.md`
- Modify: `wiki/Troubleshooting.md`

- [ ] **Step 1: Update README.md**

Line 8 -- change description from:
```
Self-hosted badge server for GitHub Container Registry. Scrapes GHCR package pages and serves ...
```
to:
```
Self-hosted badge server for GitHub Container Registry. Scrapes GHCR package pages and the OCI registry API to serve ...
```

Line 60 -- change the size row example from `12.4 MB` to `82.5 MB (amd64) | 79.2 MB (arm64)`:

```markdown
| Image size | `/owner/package/size.json` | `82.5 MB (amd64) \| 79.2 MB (arm64)` |
```

- [ ] **Step 2: Update wiki/Home.md**

Line 3 -- change from:
```
Self-hosted badge server for GitHub Container Registry. Scrapes GHCR package pages and serves ...
```
to:
```
Self-hosted badge server for GitHub Container Registry. Scrapes GHCR package pages and the OCI registry API to serve ...
```

- [ ] **Step 3: Update wiki/Badge-Usage.md**

Line 11 -- change the size row from:
```
| Image size | `size.json` | image size | blue | `12.4 MB`, `1.2 GB` |
```
to:
```
| Image size | `size.json` | image size | blue | `82.5 MB (amd64)`, `82.5 MB (amd64) \| 79.2 MB (arm64)` |
```

Lines 59-67 -- update the "Image sizes" table to note multi-platform output:

```markdown
Image sizes use binary units. Multi-arch images show per-platform sizes (e.g. `82.5 MB (amd64) | 79.2 MB (arm64)`):

| Raw Bytes | Displayed |
|-----------|-----------|
| 0 or nil | `unknown` |
| < 1 KiB | `512 B` |
| < 1 MiB | `45.2 KB` |
| < 1 GiB | `12.4 MB` |
| >= 1 GiB | `1.2 GB` |
```

- [ ] **Step 4: Update wiki/Troubleshooting.md**

Replace the "Badge Shows unknown" section (lines 3-9) with:

```markdown
## Badge Shows "unknown"

The version and architecture badges return `unknown` when the scraper could not extract that field from the GitHub packages page. The size badge returns `unknown` when the OCI registry API could not be reached or the manifest could not be parsed. Possible causes:

- **Package is new with no published versions yet.** The packages page won't have version or architecture data until at least one tagged version is pushed.
- **GitHub changed their HTML structure.** The scraper uses regex patterns to extract version and architecture data. If GitHub redesigns the packages page, the patterns may stop matching. Check the logs for `parse failed` warnings.
- **OCI registry unreachable or rate-limited.** The size badge fetches manifests from `ghcr.io/v2/`. If the registry is down or rate-limiting, size will show `unknown`. Check the logs for `size fetch failed` warnings.
- **Private package.** The OCI token fetch uses anonymous auth. Private packages will always show `unknown` for size.
- **Scrape hasn't completed yet.** On startup, pkgbadge scrapes all packages before starting the HTTP server. If you see `unknown` immediately after a restart, wait for the initial scrape to finish.
```

- [ ] **Step 5: Commit**

```bash
cd /home/lns/pkgbadge && git add README.md wiki/ && git commit -m "docs: update size badge examples and troubleshooting for OCI manifest fetching"
```

---

### Task 6: Smoke test against live GHCR

**Files:** None (verification only)

- [ ] **Step 1: Smoke test with a version-tagged package**

Run the full test suite first:

```bash
cd /home/lns/pkgbadge && go test ./... -v
```
Expected: ALL PASS.

Then do a live smoke test. Build and run locally against docker-sentinel (has version tag `2.12.2`):

```bash
cd /home/lns/pkgbadge && go build -o pkgbadge . && \
  PKGBADGE_PACKAGES="Will-Luck/Docker-Sentinel/docker-sentinel" \
  PKGBADGE_PORT=19876 \
  ./pkgbadge &
sleep 5
curl -s http://localhost:19876/will-luck/docker-sentinel/size.json | python3 -m json.tool
kill %1
```

Expected: JSON response with `"message"` containing size(s) like `"XX.X MB (amd64) | XX.X MB (arm64)"`, not `"unknown"`.

- [ ] **Step 2: Smoke test with iplayer-arr (the original issue reporter)**

```bash
cd /home/lns/pkgbadge && \
  PKGBADGE_PACKAGES="Will-Luck/iplayer-arr/iplayer-arr" \
  PKGBADGE_PORT=19876 \
  ./pkgbadge &
sleep 5
curl -s http://localhost:19876/will-luck/iplayer-arr/size.json | python3 -m json.tool
kill %1
```

Expected: JSON with per-platform sizes (iplayer-arr is multi-arch amd64+arm64).

- [ ] **Step 3: Verify badge message length is reasonable for shields.io**

Check that the longest expected badge message fits. Shields.io has no hard limit but messages over ~40 characters start to look cramped. A two-platform message like `82.5 MB (amd64) | 79.2 MB (arm64)` is 35 characters -- fine.

Visually verify by opening:
```
https://img.shields.io/badge/image%20size-82.5%20MB%20(amd64)%20%7C%2079.2%20MB%20(arm64)-blue
```

Expected: Badge renders cleanly without truncation.

- [ ] **Step 4: Clean up test binary**

```bash
cd /home/lns/pkgbadge && rm -f pkgbadge
```

---

### Task 7: Final commit and push

**Files:** None

- [ ] **Step 1: Run full test suite one final time**

```bash
cd /home/lns/pkgbadge && go test ./... -v
```
Expected: ALL PASS.

- [ ] **Step 2: Review git log**

```bash
cd /home/lns/pkgbadge && git log --oneline -10
```

Expected: Clean commit history with one commit per task.

- [ ] **Step 3: Push to Gitea**

```bash
cd /home/lns/pkgbadge && git push
```
