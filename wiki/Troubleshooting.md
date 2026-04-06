# Troubleshooting

## Badge Shows "unknown"

The version and architecture badges return `unknown` when the scraper could not extract that field from the GitHub packages page. The size badge returns `unknown` when the OCI registry API could not be reached or the manifest could not be parsed. Possible causes:

- **Package is new with no published versions yet.** The packages page won't have version or architecture data until at least one tagged version is pushed.
- **GitHub changed their HTML structure.** The scraper uses regex patterns to extract version and architecture data. If GitHub redesigns the packages page, the patterns may stop matching. Check the logs for `parse failed` warnings.
- **OCI registry unreachable or rate-limited.** The size badge fetches manifests from `ghcr.io/v2/`. If the registry is down or rate-limiting, size will show `unknown`. Check the logs for `size fetch failed` warnings.
- **Private package.** The OCI token fetch uses anonymous auth. Private packages will always show `unknown` for size.
- **Scrape hasn't completed yet.** On startup, pkgbadge scrapes all packages before starting the HTTP server. If you see `unknown` immediately after a restart, wait for the initial scrape to finish.

## Badge Returns 404 "package not configured"

The requested `owner/package` combination is not in your `PKGBADGE_PACKAGES` list. Check:

1. The badge URL uses the **package name**, not the repository name. For `Will-Luck/Docker-Sentinel/docker-sentinel`, the badge URL is `/will-luck/docker-sentinel/pulls.json`.
2. The package is in your `PKGBADGE_PACKAGES` environment variable.
3. Spelling and case. The lookup is case-insensitive, but the package must match exactly what you configured (minus case).

## Scrape Failures in Logs

```
level=WARN msg="scrape failed, keeping stale data" package=owner/package error="HTTP 429 from ..."
```

GitHub may rate-limit requests if you scrape too frequently or have many packages. Increase `PKGBADGE_INTERVAL` or reduce the number of configured packages.

When a scrape fails, pkgbadge keeps serving the last successfully scraped data. Badges will not go blank. If no data has ever been scraped for a package (first boot + immediate failure), that package will return 404 until a successful scrape.

## Parse Failures in Logs

```
level=WARN msg="parse failed" package=owner/package error="no data extracted from page, HTML structure may have changed"
```

This means the page was fetched, but none of the regex patterns matched. Most likely GitHub changed their HTML. Open an issue with the current HTML structure and the patterns can be updated.

## Startup Fails with "invalid PKGBADGE_PACKAGES"

The package format is `owner/package` or `owner/repo/package`. Common mistakes:

- Missing owner: `docker-sentinel` (needs `owner/docker-sentinel`)
- Too many segments: `github.com/owner/repo/package` (only 2 or 3 segments allowed)
- Trailing comma with empty entry (trimmed automatically, but double commas with spaces may parse oddly)

## Startup Fails with "invalid PKGBADGE_INTERVAL"

The interval must be a valid [Go duration string](https://pkg.go.dev/time#ParseDuration): `30m`, `1h`, `6h`, `24h`. Common mistakes:

- Using `6` instead of `6h` (number alone is not valid)
- Using `1d` (Go durations don't support days, use `24h`)
- Using `6 hours` (no spaces or words, just the short format)

## Shields.io Shows Stale Data

Shields.io caches badge responses on their end. Even after pkgbadge has fresh data, shields.io may serve its cached version for several minutes. Options:

- Append `&cacheSeconds=300` to the shields.io URL to lower their cache time
- Wait a few minutes for the shields.io cache to expire
- The `Cache-Control: max-age=3600` header from pkgbadge tells shields.io how long to cache. If you need faster updates, this value is set in the source code (`server.go`).

## Container Won't Start

Check the logs:

```bash
docker logs pkgbadge
```

The most common startup failure is a missing or invalid `PKGBADGE_PACKAGES`. The service exits immediately with an error message if this variable is missing or malformed.
