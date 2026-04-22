# Changelog

## [Unreleased]

### Fixed

- Pull counts over 1,000 now scrape correctly. GitHub renders these as abbreviated text (e.g. `1.32K` for 1,318) which the old regex could not parse, so packages above the threshold silently reported zero pulls. The scraper now reads the exact integer from the `title` attribute on the count `<h3>`, which GitHub populates regardless of display format.

## v1.1.0

### Added

- **Per-platform image sizes on the `size.json` badge.** The scraper now fetches compressed layer sizes from the GHCR OCI registry API (`ghcr.io/v2/`). Multi-arch images display per-platform breakdowns (e.g. `82.5 MB (amd64) | 81.2 MB (arm64)`). Single-arch images show a plain size.
- Version-aware manifest fetching: uses the scraped version tag first, falls back to `latest` if the tag is not found on the registry.
- Best-effort resilience: a transient failure fetching one platform's manifest does not lose data for the other platforms.
- New test fixtures and 6 new unit tests covering manifest parsing and badge formatting.

### Changed

- `PackageStats.SizeBytes` (int64, never populated) replaced with `PlatformSizes` (map[string]int64) to support per-platform sizes.
- Badge formatting for size: nil/empty map returns `"unknown"`, single platform with known arch shows `"82.5 MB (amd64)"`, multiple platforms show pipe-separated breakdown.
- Updated README, wiki/Home, wiki/Badge-Usage, and wiki/Troubleshooting to reflect new OCI-based size fetching and per-platform output format.

### Fixed

- Size badge no longer shows `"unknown"` for all packages (closes GiteaLN/pkgbadge#1).

## v1.0.0

Initial release. HTML scraping for pull counts, versions, and platform architectures. Shields.io endpoint badge server.
