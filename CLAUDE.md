# pkgbadge - Agent Instructions

Shields-style package-version badge service for GHCR, Docker Hub, and arbitrary Go modules. Single Go binary, no DB.

## Build & Run

- `go build ./...` — compile
- `go test ./...` — run tests
- `go run .` — local dev server (default port 8080)
- Dockerfile builds a static Go binary on `alpine`. No external deps.

## Release workflow

Gitea is the squash authority. All merges happen via Gitea PRs. Never squash-merge on GitHub -- squashing the same branch twice (once per remote) produces two different commits with identical content, and the next PR conflicts on shared files (`CHANGELOG.md` is a reliable offender).

Remotes here: `origin` = Gitea, `github` = GitHub. Default branch: `main`.

1. Open the PR on Gitea (`origin`), get CI green, merge (squash).
2. Smoke-test: `go test ./...` + `curl http://localhost:8080/badge/ghcr.io/will-luck/iplayer-arr` returns a valid SVG.
3. Fast-forward GitHub: `git push github origin/main:main`. No GitHub PR needed.
4. Cut the release tag: move `[Unreleased]` -> `[X.Y.Z] - YYYY-MM-DD` in `CHANGELOG.md`, commit on `main`, `git tag -a vX.Y.Z -m "..."`, push tag to both remotes.

External GitHub contributor? Pull their branch, push to Gitea, open a Gitea PR, squash there, fast-forward GitHub. Their GitHub PR auto-closes as "merged".
