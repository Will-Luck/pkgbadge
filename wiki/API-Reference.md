# API Reference

pkgbadge exposes a single HTTP endpoint that serves badge data in the [shields.io endpoint schema](https://shields.io/badges/endpoint-badge).

## Endpoint

```
GET /{owner}/{package}/{badge}.json
```

### Path Parameters

| Parameter | Description |
|-----------|-------------|
| `owner` | GitHub user or organisation (case-insensitive) |
| `package` | GHCR package name (case-insensitive) |
| `badge` | Badge type: `pulls`, `version`, `size`, or `arch` |

The `.json` extension is required.

### Success Response (200)

```json
{
  "schemaVersion": 1,
  "label": "ghcr pulls",
  "message": "1.5k",
  "color": "blue"
}
```

Headers:
- `Content-Type: application/json`
- `Cache-Control: max-age=3600`

### Error Responses (404)

**Invalid path format** (not exactly 3 path segments):

```json
{"error": "expected /owner/package/badge.json"}
```

**Wrong file extension:**

```json
{"error": "expected .json extension"}
```

**Package not in configured list:**

```json
{"error": "package not configured"}
```

**Unknown badge type:**

```json
{"error": "unknown badge type: foo"}
```

## Caching Behaviour

Badge responses include a `Cache-Control: max-age=3600` header, telling clients (and shields.io) to cache the response for 1 hour. The scraper interval controls how often pkgbadge fetches fresh data from GitHub; the cache header controls how often downstream consumers re-fetch from pkgbadge.

For shields.io specifically: shields.io has its own caching layer, so updates may take a few minutes to appear even after pkgbadge has fresh data. You can append `&cacheSeconds=300` to the shields.io URL to reduce this.

## Health Check

There is no dedicated health endpoint. To check if the service is running, request any valid badge URL and check for a 200 response, or request an invalid path and check for a 404 (which still indicates the server is up).
