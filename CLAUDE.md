# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Build & run (Docker)
make run              # build image + run with .env
make build            # build Docker image only
make stop             # stop running container
make watch            # auto-rebuild on file changes (requires entr)

# Run directly (requires Go 1.25+)
go run ./cmd/server

# Test
go test ./...         # all tests
go test ./internal/content/...  # single package

# Lint & format
golangci-lint run     # requires golangci-lint v2 (v1 won't work with Go 1.25)
go vet ./...
goimports -w .        # format (preferred over gofmt)
gofmt -l .            # check formatting (used in CI)
```

## Architecture

Go web server for williamfindlay.com. Stdlib `net/http` with Go 1.22+ routing patterns, no third-party router.

### Content Pipeline

Content lives in a **separate Git repo** (configured via `CONTENT_REPO_URL`). The server clones it on startup, then syncs every `SYNC_INTERVAL` (default 5m). Content is parsed from Markdown with YAML frontmatter (goldmark + chroma syntax highlighting, Dracula theme). The entire content store is swapped atomically via `atomic.Pointer[ContentStore]` so reads never block during reloads.

Key flow: `sync.CloneOrPull` → `loader.LoadFromDir` → `store.Store()` (atomic swap) → handlers call `store.Load()` per request.

### Embedded Assets

`embed.go` at project root exports `var Embedded embed.FS` containing `static/` and `templates/`. This exists because `cmd/server/main.go` can't use `//go:embed` to reach files outside its package directory. The embedded FS is passed through `server.New()` to both the template renderer and the static file server.

### Request Path

`cmd/server/main.go` → `server.New()` → `routes()` builds mux with middleware stack (logging → securityHeaders → cacheStatic for `/static/`). Handlers are closures returned from methods on `handler.Deps`, which holds the atomic store, renderer, and config.

### Template Rendering

`render.Renderer` clones `templates/base.html` per render, then parses the page-specific template into it. Pages inject content via `{{template "content" .}}`. Template functions: `formatDate`, `formatDateShort`, `formatRFC3339`, `join`, `safeHTML`, `truncate`, `currentYear`.

## CI

Four GitHub Actions workflows run on push/PR to main:
- **test**: `go test ./...`
- **lint**: `golangci-lint-action@v7` + `go vet` (must use v7 for golangci-lint v2)
- **fmt**: `gofmt -l .` (fails if any file is unformatted)
- **docker**: builds image, pushes to `docker.williamfindlay.com` on main

## Configuration

All config via environment variables (see `internal/config/`). Key ones:
- `CONTENT_REPO_URL` (required): Git URL for content repo
- `PORT` (default 8080), `SYNC_INTERVAL` (default 5m), `DEV_MODE` (default false)
- `PARTICLE_*`: Canvas particle system parameters (count, speed, size, colors, etc.)

## Docker

Multi-stage: `golang:1.25-alpine` → `gcr.io/distroless/static:nonroot`. Static assets are embedded in the binary. The `/data/content` directory for Git operations must be pre-created in the build stage since distroless has no shell.
