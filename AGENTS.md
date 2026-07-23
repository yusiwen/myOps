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

## HTMX Link Convention

Links that include `?partial=1` in `hx-get` MUST have an explicit `hx-push-url` with a clean URL that omits `?partial=1`. This prevents F5 from displaying a raw HTML content fragment (source code view) instead of the full page.

| Rule | Why |
|------|-----|
| **Never** `hx-push-url="true"` on links with `?partial=1` in `hx-get` | F5 re-requests the address bar URL; `?partial=1` makes the handler return a raw fragment |
| **Always** explicit clean `hx-push-url` | Browser history stays clean; F5 always renders the full page |
| `hx-push-url="true"` is safe on base links without `?partial=1` | Sidebar links (`/dashboard`, `/repos`, `/builds`, `/settings`) have no query params |
| `hx-push-url="false"` for search inputs and skeleton triggers | These don't need URL history entries |
| **Never** put `partial=1` directly in any `hx-get` URL | Use `hx-vals='{"partial":"1"}'` instead; HTMX handles `?` vs `&` automatically |

### Correct examples

| Link type | `hx-get` | `hx-push-url` |
|-----------|----------|---------------|
| Sidebar navigation | `/dashboard` | `"true"` |
| Row click to detail | `/builds/o/r/1?partial=1` | `"/builds/o/r/1"` |
| Pagination | `/builds?page=2&partial=1` | `"/builds?page=2"` |
| Back link | `/repos?partial=1` | `"/repos"` |
| Search input | `/repos?partial=1` | `"false"` |
| Skeleton / HX GET with `partial=1` | `{URL}` + `hx-vals='{"partial":"1"}'` | `"false"` / explicit |

## Build & Dev

- `make dev` — run with hot-reload
- `make build` — compile binary
- `make tailwind` — compile Tailwind CSS
- `make lint` — run golangci-lint

## Progress

See PLAN.md for the full checklist (Phase A / B / C).
