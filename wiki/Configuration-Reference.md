# Configuration Reference

All configuration is via environment variables. No config file is needed.

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PKGBADGE_PACKAGES` | Yes | *(required)* | Comma-separated list of GHCR packages to scrape |
| `PKGBADGE_INTERVAL` | No | `6h` | How often to re-scrape. Accepts any [Go duration](https://pkg.go.dev/time#ParseDuration): `30m`, `1h`, `12h`, etc. |
| `PKGBADGE_PORT` | No | `8080` | HTTP listen port |

## Package Format

Each entry in `PKGBADGE_PACKAGES` is either:

- **`owner/package`**: the GitHub repo name matches the GHCR package name
- **`owner/repo/package`**: they differ

### Why Two Formats?

GitHub Container Registry package names are always lowercase, but repository names can use mixed case. For example, the repo `Will-Luck/Docker-Sentinel` publishes a package called `docker-sentinel`. To scrape it, pkgbadge needs both names:

```
Will-Luck/Docker-Sentinel/docker-sentinel
```

The 2-part format `Will-Luck/docker-sentinel` assumes the repo name equals the package name, which would look for a repo called `docker-sentinel` (wrong).

### Examples

```bash
# Single package (repo name matches package name)
PKGBADGE_PACKAGES="willfarrell/autoheal"

# Single package (repo name differs from package name)
PKGBADGE_PACKAGES="Will-Luck/Docker-Sentinel/docker-sentinel"

# Multiple packages
PKGBADGE_PACKAGES="Will-Luck/Docker-Sentinel/docker-sentinel,Will-Luck/Docker-Guardian/docker-guardian"
```

Whitespace around commas and entries is trimmed automatically.

## Scrape Interval

The scraper runs once at startup (blocking, before the HTTP server starts), then repeats on the configured interval. GitHub package pages are public HTML, so no API token is needed, but aggressive scraping may trigger rate limits.

Recommended intervals:

| Use case | Interval |
|----------|----------|
| Development/testing | `5m` |
| General use | `1h` – `6h` |
| Low-traffic packages | `12h` – `24h` |
