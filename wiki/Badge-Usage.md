# Badge Usage

pkgbadge serves [shields.io endpoint badges](https://shields.io/badges/endpoint-badge). Each badge is a JSON endpoint that shields.io fetches and renders as an SVG.

## Available Badges

| Badge | Endpoint | Label | Colour | Example Output |
|-------|----------|-------|--------|----------------|
| Pull count | `pulls.json` | ghcr pulls | blue | `433`, `1.5k`, `2.3M` |
| Version | `version.json` | version | green | `2.11.1`, `latest` |
| Image size | `size.json` | image size | blue | `82.5 MB (amd64)`, `82.5 MB (amd64) \| 79.2 MB (arm64)` |
| Platforms | `arch.json` | platforms | blue | `amd64 \| arm64` |

## Using with shields.io

The shields.io [endpoint badge](https://shields.io/badges/endpoint-badge) fetches JSON from a URL you provide and renders it as an SVG image. Point it at your pkgbadge instance:

```
https://img.shields.io/endpoint?url=<your-pkgbadge-url>/<owner>/<package>/<badge>.json
```

### Markdown Examples

```markdown
![GHCR Pulls](https://img.shields.io/endpoint?url=https://badges.example.com/owner/package/pulls.json)
![Version](https://img.shields.io/endpoint?url=https://badges.example.com/owner/package/version.json)
![Image Size](https://img.shields.io/endpoint?url=https://badges.example.com/owner/package/size.json)
![Platforms](https://img.shields.io/endpoint?url=https://badges.example.com/owner/package/arch.json)
```

### With Link to GHCR Package Page

```markdown
[![GHCR Pulls](https://img.shields.io/endpoint?url=https://badges.example.com/owner/package/pulls.json)](https://github.com/owner/repo/pkgs/container/package)
```

### Shields.io Style Overrides

You can append shields.io query parameters to customise the badge appearance:

```markdown
![Pulls](https://img.shields.io/endpoint?url=https://badges.example.com/owner/package/pulls.json&style=flat-square)
![Version](https://img.shields.io/endpoint?url=https://badges.example.com/owner/package/version.json&style=for-the-badge)
![Pulls](https://img.shields.io/endpoint?url=https://badges.example.com/owner/package/pulls.json&color=orange&label=downloads)
```

See the [shields.io endpoint docs](https://shields.io/badges/endpoint-badge) for all available parameters (style, color, label, logo, etc).

## Number Formatting

Pull counts are formatted for readability:

| Raw Value | Displayed |
|-----------|-----------|
| 0–999 | As-is (`433`) |
| 1,000–999,999 | Thousands (`1.5k`) |
| 1,000,000+ | Millions (`2.3M`) |

Image sizes use binary units. Multi-arch images show per-platform sizes (e.g. `82.5 MB (amd64) | 79.2 MB (arm64)`):

| Raw Bytes | Displayed |
|-----------|-----------|
| 0 or nil | `unknown` |
| < 1 KiB | `512 B` |
| < 1 MiB | `45.2 KB` |
| < 1 GiB | `12.4 MB` |
| >= 1 GiB | `1.2 GB` |

## Package Name in Badge URL

The badge URL uses the **package name** (always lowercase), not the repository name. If you configured `Will-Luck/Docker-Sentinel/docker-sentinel`, the badge URL uses `docker-sentinel`:

```
/will-luck/docker-sentinel/pulls.json
```

Both owner and package are case-insensitive in the badge URL (lowercased internally).
