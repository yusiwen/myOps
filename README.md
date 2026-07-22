# myOps

Unified web interface for self-hosted [Gitea](https://gitea.io) repositories and [Drone](https://drone.io) CI/CD pipelines.

## Features

- **Dashboard** — aggregated overview: repo count, total builds, recent builds across all repos
- **Repository Browser** — search, list, and browse your Gitea repositories
- **Build Monitor** — view build status and logs across all Drone pipelines, with pagination
- **OAuth2 Authentication** — login via Gitea, no local user management
- **Dark Mode** — toggleable, persisted in localStorage
- **Async Page Loading** — skeleton screens with automatic data loading for snappy navigation

## Tech Stack

| Layer        | Choice                    |
|--------------|---------------------------|
| Backend      | Go + `chi` router         |
| Templating   | `html/template`           |
| AJAX         | HTMX                      |
| Interactions | Alpine.js                 |
| Styling      | Tailwind CSS (standalone) |
| Icons        | Lucide (inline SVG)       |
| Syntax Highlighting | Chroma              |

## Prerequisites

- Go 1.26+
- Node.js (for Tailwind CSS CLI)
- Gitea instance with a Personal Access Token (PAT)
- Drone instance with a service token

## Quick Start

### 1. Clone and build

```bash
git clone <repo-url> myops
cd myops
npm install
make build
```

### 2. Configure

Create a `config.yaml` in the project root:

```yaml
listen_addr: ":9090"
base_url: "http://localhost:9090"
session_key: "<openssl rand -hex 32>"

gitea:
  url: "https://git.yusiwen.cn"
  token: "<Gitea Personal Access Token>"
  client_id: "<Gitea OAuth2 Client ID>"
  client_secret: "<Gitea OAuth2 Client Secret>"

drone:
  url: "https://drone.example.com"
  token: "<Drone admin/service token>"
```

### 3. Setup Gitea OAuth2

1. Go to **Settings → Applications → Create OAuth2 Application**
2. Redirect URI: `https://myops-host:9090/auth/callback`
3. Copy the Client ID and Client Secret
4. Create a **Personal Access Token** with `read:repository`, `read:user` scopes

### 4. Run

```bash
./bin/myops
```

Open `http://localhost:9090` and sign in with Gitea.

## Configuration

Configuration is loaded via `config.yaml` (auto-discovered in current directory, `~/.config/myops/`, or `/etc/myops/`) or via environment variables (which override YAML). Explicit path via `--config` / `-c` flag.

| Config Key          | Environment Variable  | Default       | Description                          |
|---------------------|-----------------------|---------------|--------------------------------------|
| `listen_addr`       | `LISTEN_ADDR`         | `:9090`       | Server listen address                |
| `base_url`          | `BASE_URL`            |               | Public-facing URL (OAuth2 redirect)  |
| `log_file`          | `LOG_FILE`            | `myops.log`   | Detailed log file path               |
| `session_key`       | `SESSION_KEY`         |               | Cookie encryption key (32+ chars)    |
| `gitea.url`         | `GITEA_URL`           |               | Gitea server base URL                |
| `gitea.token`       | `GITEA_TOKEN`         |               | Gitea Personal Access Token          |
| `gitea.client_id`   | `GITEA_CLIENT_ID`     |               | OAuth2 client ID                     |
| `gitea.client_secret`| `GITEA_CLIENT_SECRET`|               | OAuth2 client secret                 |
| `drone.url`         | `DRONE_URL`           |               | Drone server base URL                |
| `drone.token`       | `DRONE_TOKEN`         |               | Drone service token                  |

## Project Structure

```
myOps/
├── cmd/myops/main.go             # Entry point
├── internal/
│   ├── config/config.go          # Configuration loader
│   ├── code/                     # CodeRenderer interface + Chroma (Phase A)
│   ├── gitea/client.go           # Gitea SDK wrapper
│   ├── drone/client.go           # Drone SDK wrapper
│   ├── handler/                  # HTTP handlers (auth, home, repo, build, settings)
│   ├── middleware/               # Auth guard, logger
│   └── log/                      # Dual-output logger (file + stderr)
├── web/
│   ├── templates/                # Go HTML templates (7 pages + 1 base layout)
│   └── static/                   # Tailwind CSS output
├── config.example.yaml           # Example configuration
├── PLAN.md                       # Full project plan & checklist
├── Makefile
├── tailwind.config.js
└── go.mod
```

## Makefile Commands

| Command       | Description                     |
|---------------|---------------------------------|
| `make build`  | Build Tailwind CSS + Go binary  |
| `make dev`    | Hot-reload development (air)    |
| `make tailwind` | Build Tailwind CSS only       |
| `make lint`   | Run golangci-lint               |
| `make clean`  | Remove build artifacts          |

## Navigation

| Page         | Route                       | Description                          |
|--------------|-----------------------------|--------------------------------------|
| Login        | `/`                         | Sign in with Gitea OAuth2            |
| Dashboard    | `/dashboard`                | Repo count, build stats, recent builds|
| Repositories | `/repos`                    | Searchable, paginated repo list      |
| Repo Detail  | `/repos/{owner}/{name}`     | Repository metadata, file tree       |
| Builds       | `/builds`                   | Paginated build list across all repos|
| Build Detail | `/builds/{owner}/{name}/{n}`| Build status, logs, auto-refresh     |
| Settings     | `/settings`                 | Connection status, about info        |

## Architecture

- Server-rendered HTML with HTMX partial swaps
- OAuth2 authentication via Gitea (no local user management)
- Service-level Drone token (shared across all myOps users)
- Code browsing via `CodeRenderer` interface: Chroma (Phase A), Tree-sitter (Phase C)
- Async skeleton loading: every page returns a loading shell instantly, data loads asynchronously

## License

MIT
