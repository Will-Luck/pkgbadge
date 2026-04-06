# OCI Manifest Size Badge

**Issue:** GiteaLN/pkgbadge#1
**Date:** 2026-04-06
**Status:** Approved

## Problem

The `size.json` badge returns `"unknown"` for all GHCR packages. `SizeBytes` is declared in `PackageStats` but never populated because the GitHub packages HTML page does not contain image size information. Size data is only available via the OCI registry API at `ghcr.io/v2/`.

## Solution

Add OCI registry manifest fetching to the scraper. After scraping the HTML page for pulls/version/arch, make a second pass against `ghcr.io` to fetch compressed layer sizes per platform.

## Data Model

Replace `SizeBytes int64` with:

```go
PlatformSizes map[string]int64 // key: "linux/amd64", value: total compressed layer bytes
```

This supports both single-arch and multi-arch images. A nil or empty map means size is unknown.

## New Function: `fetchImageSizes`

Location: `scraper.go`

```go
func fetchImageSizes(ctx context.Context, owner, pkg string) (map[string]int64, error)
```

### Flow

1. **Token:** `GET https://ghcr.io/token?scope=repository:{owner}/{pkg}:pull` -- anonymous auth for public packages. Parse JSON response for `token` field.

2. **Manifest:** `GET https://ghcr.io/v2/{owner}/{pkg}/manifests/latest` with headers:
   - `Authorization: Bearer {token}`
   - `Accept: application/vnd.oci.image.index.v1+json, application/vnd.docker.distribution.manifest.list.v2+json, application/vnd.oci.image.manifest.v1+json, application/vnd.docker.distribution.manifest.v2+json`

3. **Parse response by `mediaType`:**

   **OCI image index** (`application/vnd.oci.image.index.v1+json` or `application/vnd.docker.distribution.manifest.list.v2+json`):
   - Filter `manifests[]` to entries where `platform.os` is non-empty and not `"unknown"` (skips attestation manifests)
   - For each platform entry, fetch its manifest by digest:
     `GET https://ghcr.io/v2/{owner}/{pkg}/manifests/{digest}`
   - Sum `layers[].size` for each platform manifest
   - Store as `"linux/amd64" -> 86476951`

   **Single image manifest** (`application/vnd.oci.image.manifest.v1+json` or `application/vnd.docker.distribution.manifest.v2+json`):
   - Sum `layers[].size` directly
   - Store under `"linux/amd64"` as default key

4. Return the map.

### JSON Structures (minimal, for parsing)

```go
type ociTokenResponse struct {
    Token string `json:"token"`
}

type ociManifest struct {
    MediaType string          `json:"mediaType"`
    Manifests []ociDescriptor `json:"manifests,omitempty"` // present in index
    Layers    []ociDescriptor `json:"layers,omitempty"`    // present in image manifest
}

type ociDescriptor struct {
    MediaType string      `json:"mediaType"`
    Digest    string      `json:"digest"`
    Size      int64       `json:"size"`
    Platform  *ociPlatform `json:"platform,omitempty"`
}

type ociPlatform struct {
    OS           string `json:"os"`
    Architecture string `json:"architecture"`
}
```

These are private types in `scraper.go`, not exported.

## Error Handling

Size fetching is best-effort. If any step fails (token fetch, manifest fetch, JSON parse, network timeout), log a warning and leave `PlatformSizes` nil. The badge falls back to `"unknown"`. A size fetch failure never causes the entire scrape cycle to fail or skip other fields.

The existing `httpClient` (30s timeout) is reused for registry requests.

## Badge Formatting

The `"size"` case in `buildBadge` changes to:

- `PlatformSizes` nil or empty: message = `"unknown"`
- One platform: message = `"82.5 MB"` (just the formatted size, no platform label)
- Multiple platforms: message = `"82.5 MB (amd64) | 79.1 MB (arm64)"` -- sorted alphabetically by architecture name, `linux/` prefix stripped

## Integration

In `scrapeAll`, after `parsePackagePage` succeeds:

```go
sizes, err := fetchImageSizes(ctx, ref.Owner, ref.Package)
if err != nil {
    log.Warn("size fetch failed", "package", ref.Key(), "error", err)
} else {
    stats.PlatformSizes = sizes
}
```

## Testing

### Unit tests for manifest parsing

Extract the JSON-to-map logic into a testable helper:

```go
func parseManifestSizes(indexBody []byte, fetchManifest func(digest string) ([]byte, error)) (map[string]int64, error)
```

Test with JSON fixtures:
- `testdata/manifest-index.json` -- multi-arch OCI index with 2 real platforms + 2 attestation entries
- `testdata/manifest-single.json` -- single image manifest with layers
- `testdata/manifest-amd64.json` -- platform manifest fetched by digest (used by index test)
- `testdata/manifest-arm64.json` -- second platform manifest

### Unit tests for badge formatting

In `server_test.go`, test the size badge case:
- Nil `PlatformSizes` returns `"unknown"`
- Single platform returns plain size (e.g. `"82.5 MB"`)
- Two platforms returns breakdown (e.g. `"82.5 MB (amd64) | 79.1 MB (arm64)"`)

### Existing tests

Unchanged. HTML parsing tests are unaffected since size comes from a separate code path.

## Files Changed

| File | Change |
|------|--------|
| `types.go` | Replace `SizeBytes int64` with `PlatformSizes map[string]int64` |
| `scraper.go` | Add OCI types, `fetchImageSizes`, `parseManifestSizes`, call from `scrapeAll` |
| `server.go` | Update `buildBadge` size case to format per-platform sizes |
| `server_test.go` | Add size badge formatting tests |
| `scraper_test.go` | Add manifest parsing tests |
| `testdata/manifest-index.json` | New fixture: multi-arch OCI index |
| `testdata/manifest-single.json` | New fixture: single image manifest |
| `testdata/manifest-amd64.json` | New fixture: amd64 platform manifest |
| `testdata/manifest-arm64.json` | New fixture: arm64 platform manifest |
