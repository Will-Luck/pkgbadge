# Installation

## Docker CLI

```bash
docker run -d \
  --name pkgbadge \
  --restart unless-stopped \
  -p 8080:8080 \
  -e PKGBADGE_PACKAGES="owner/repo,owner/repo/package" \
  ghcr.io/will-luck/pkgbadge:latest
```

## Docker Compose

```yaml
services:
  pkgbadge:
    image: ghcr.io/will-luck/pkgbadge:latest
    ports:
      - "8080:8080"
    environment:
      PKGBADGE_PACKAGES: "owner/repo,owner/repo/package"
      # PKGBADGE_INTERVAL: "6h"
      # PKGBADGE_PORT: "8080"
    restart: unless-stopped
```

```bash
docker compose up -d
```

## Docker Swarm

```yaml
services:
  pkgbadge:
    image: ghcr.io/will-luck/pkgbadge:latest
    environment:
      PKGBADGE_PACKAGES: "owner/repo"
      PKGBADGE_INTERVAL: "1h"
    ports:
      - "8080:8080"
    deploy:
      replicas: 1
      restart_policy:
        condition: on-failure
```

Each replica maintains its own in-memory cache and scrapes independently. There is no shared state, so running multiple replicas behind a load balancer works without any extra configuration.

## Building from Source

Requires Go 1.24 or later.

```bash
git clone https://github.com/Will-Luck/pkgbadge.git
cd pkgbadge
go build -o pkgbadge .
```

Run it:

```bash
export PKGBADGE_PACKAGES="owner/repo"
./pkgbadge
```

## Building the Docker Image

```bash
git clone https://github.com/Will-Luck/pkgbadge.git
cd pkgbadge
docker build -t pkgbadge .
```

The Dockerfile uses a multi-stage build: compiles with `golang:1.24-alpine`, then copies the binary into `gcr.io/distroless/static-debian12` for a minimal runtime image with no shell or package manager.

## Reverse Proxy

pkgbadge serves plain HTTP. Put it behind a reverse proxy (Nginx, Caddy, Traefik) for TLS.

Example Nginx location block:

```nginx
location / {
    proxy_pass http://127.0.0.1:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}
```

The shields.io endpoint URL in your README badges should point to the public HTTPS address, not the internal HTTP port.
