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

- [Nix](https://nixos.org/download/) with flakes enabled — provides the entire toolchain
- [direnv](https://direnv.net/) — optional, but loads the shell automatically
- Gitea instance with a Personal Access Token (PAT)
- Drone instance with a service token

The toolchain is declared in `flake.nix` and locked in `flake.lock`. It provides
Go 1.26, `gopls`, `golangci-lint`, `air`, and the Tailwind CSS v4 standalone CLI.
Node.js and npm are deliberately **not** used.

## Quick Start

### 1. Clone and build

```bash
git clone <repo-url> myops
cd myops
direnv allow      # load the dev shell (or run `nix develop` manually)
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

## Development Environment

The toolchain is pinned with a Nix flake, so every checkout builds with exactly
the same Go and Tailwind versions regardless of what is installed system-wide.

| File | Purpose |
|------|---------|
| `flake.nix` | Dev shell: Go, `gopls`, `golangci-lint`, `air`, Tailwind CSS v4 CLI, `make` |
| `flake.lock` | Pins the exact nixpkgs revision and content hash |
| `.envrc` | `use flake` — direnv loads the shell automatically on `cd` |
| `.air.toml` | Hot-reload rules for `make dev` |

```bash
direnv allow        # once after cloning; afterwards the shell loads on cd
nix develop         # alternative for people who do not use direnv
nix flake update    # bump nixpkgs, then commit flake.lock
nix flake check     # evaluate the flake
```

`GOTOOLCHAIN=local` is set inside the shell so Go always uses the Nix-provided
toolchain instead of downloading one of its own.

`air` rebuilds Tailwind CSS and the Go binary whenever a `.go` or `.html` file
changes. myOps resolves `web/` relative to the working directory, so run the
server from the project root.

### macOS: `syntax error near unexpected token ';&'`

On macOS, direnv may print this on every shell load:

```
/bin/bash: eval:3349: syntax error near unexpected token `;'
```

This is an upstream issue, not a problem with this flake. nixpkgs' `stdenv`
helper functions use the bash 4 `;&` case fallthrough, macOS ships bash 3.2 as
`/bin/bash`, and direnv evaluates `.envrc` with it. The toolchain still loads —
only five internal variables defined after that point (`NIX_BUILD_TOP`, `TMP`,
`TMPDIR`, `TEMP`, `TEMPDIR`) are skipped. See
[direnv#1256](https://github.com/direnv/direnv/issues/1256) and
[nixpkgs#296936](https://github.com/NixOS/nixpkgs/issues/296936).

To silence it, give direnv a bash 5:

```bash
nix profile install nixpkgs#bashInteractive
export DIRENV_BASH="$(command -v bash)"   # add to your shell rc, then restart it
```

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
├── flake.nix                     # Nix dev shell (pinned toolchain)
├── flake.lock                    # Locked nixpkgs revision
├── .envrc                        # direnv entry point (`use flake`)
├── .air.toml                     # air hot-reload rules
├── Makefile
├── tailwind.config.js
└── go.mod
```

## Makefile Commands

| Command         | Description                     |
|-----------------|---------------------------------|
| `make build`    | Build Tailwind CSS + Go binary  |
| `make dev`      | Hot-reload development (air)    |
| `make tailwind` | Build Tailwind CSS only         |
| `make lint`     | Run golangci-lint               |
| `make fmt`      | Format `*.nix` files (`nix fmt`)|
| `make clean`    | Remove build artifacts          |

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
