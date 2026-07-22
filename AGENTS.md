# myOps

Unified web interface for self-hosted Gitea repositories and Drone CI/CD pipelines.

## Tech Stack

| Layer        | Choice                    |
|--------------|---------------------------|
| Backend      | Go + `chi` router         |
| Templating   | `html/template`           |
| AJAX         | HTMX                      |
| Interactions | Alpine.js                 |
| Styling      | Tailwind CSS (standalone) |
| Icons        | Lucide (inline SVG)       |

## SDKs

- `code.gitea.io/sdk/gitea` — Gitea API client
- `github.com/drone/drone-go` — Drone API client

## Architecture

- Server-rendered HTML with HTMX partial swaps
- OAuth2 authentication via Gitea (no local user management)
- Service-level Drone token (shared across all myOps users)
- Code browsing via `CodeRenderer` interface: Chroma (Phase A), Tree-sitter (Phase C)

## Project Structure

```
cmd/myops/main.go             # Entry point
internal/
  config/config.go            # Config (env / yaml)
  code/                       # CodeRenderer interface + implementations
  gitea/client.go             # Gitea SDK wrapper
  drone/client.go             # Drone SDK wrapper
  handler/                    # HTTP handlers (home, repo, build, auth)
  middleware/                 # Auth guard, logger
web/
  templates/                  # Go HTML templates
  static/                     # Tailwind CSS, JS
```

## Conventions

- All documentation and code comments in English
- Handler code depends on `CodeRenderer` interface, not concrete implementation
- No npm — Tailwind standalone CLI, JS libs via CDN
- Single-binary deployment

## Build & Dev

- `make dev` — run with hot-reload
- `make build` — compile binary
- `make tailwind` — compile Tailwind CSS
- `make lint` — run golangci-lint

## Progress

See PLAN.md for the full checklist (Phase A / B / C).
