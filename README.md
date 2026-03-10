# pkgbadge

[![CI](https://github.com/Will-Luck/pkgbadge/actions/workflows/ci.yml/badge.svg)](https://github.com/Will-Luck/pkgbadge/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/Will-Luck/pkgbadge)](https://github.com/Will-Luck/pkgbadge/releases)
[![Licence](https://img.shields.io/github/license/Will-Luck/pkgbadge)](LICENSE)
[![Docker Pulls](https://img.shields.io/docker/pulls/willluck/pkgbadge)](https://hub.docker.com/r/willluck/pkgbadge)

Self-hosted badge server for GitHub Container Registry. Scrapes GHCR package pages and serves [shields.io endpoint badges](https://shields.io/badges/endpoint-badge) with pull counts, versions, image sizes, and platform info.

## Features

- Pull count, version, image size, and architecture badges for any public GHCR package
- Shields.io [endpoint badge](https://shields.io/badges/endpoint-badge) compatible JSON responses
- Background scraping on a configurable interval (default 6h)
- Supports packages where the repo name differs from the package name
- Zero external dependencies, Go stdlib only
- Distroless container image
- Graceful shutdown (SIGINT/SIGTERM)
- 1-hour client-side cache headers

## Quick Start

### Docker CLI

```bash
docker run -d \
  --name pkgbadge \
  --restart unless-stopped \
  -p 8080:8080 \
  -e PKGBADGE_PACKAGES="owner/repo,owner/repo/package" \
  ghcr.io/will-luck/pkgbadge:latest
```

### Docker Compose

```yaml
services:
  pkgbadge:
    image: ghcr.io/will-luck/pkgbadge:latest
    ports:
      - "8080:8080"
    environment:
      PKGBADGE_PACKAGES: "owner/repo,owner/repo/package"
    restart: unless-stopped
```

Then add badges to your README:

```markdown
![GHCR Pulls](https://img.shields.io/endpoint?url=https://your-host/owner/package/pulls.json)
![Version](https://img.shields.io/endpoint?url=https://your-host/owner/package/version.json)
```

## Badge Types

| Badge | Endpoint | Example |
|-------|----------|---------|
| Pull count | `/owner/package/pulls.json` | `1.5k` |
| Version | `/owner/package/version.json` | `2.11.1` |
| Image size | `/owner/package/size.json` | `12.4 MB` |
| Platforms | `/owner/package/arch.json` | `amd64 \| arm64` |

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PKGBADGE_PACKAGES` | *(required)* | Comma-separated packages to scrape |
| `PKGBADGE_INTERVAL` | `6h` | Scrape interval ([Go duration](https://pkg.go.dev/time#ParseDuration) format) |
| `PKGBADGE_PORT` | `8080` | HTTP listen port |

Package format is `owner/package` when the repo and package names match, or `owner/repo/package` when they differ (e.g. `Will-Luck/Docker-Sentinel/docker-sentinel`).

See the [wiki](https://github.com/Will-Luck/pkgbadge/wiki) for detailed configuration and usage guides.

## Documentation

See the [Wiki](https://github.com/Will-Luck/pkgbadge/wiki) for:

- [Installation options](https://github.com/Will-Luck/pkgbadge/wiki/Installation) (Docker, Compose, from source)
- [Badge types and shields.io usage](https://github.com/Will-Luck/pkgbadge/wiki/Badge-Usage)
- [API reference](https://github.com/Will-Luck/pkgbadge/wiki/API-Reference)
- [Troubleshooting](https://github.com/Will-Luck/pkgbadge/wiki/Troubleshooting)

## Licence

MIT
