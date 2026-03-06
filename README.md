<div align="center">

# mcpup

**Manage MCP servers across 13 AI clients from one CLI.**

Define a server once. Enable it where you want. Keep unmanaged client entries intact.

[![CI](https://github.com/mohammedsamin/mcpup/actions/workflows/ci.yml/badge.svg)](https://github.com/mohammedsamin/mcpup/actions/workflows/ci.yml)
[![Release](https://github.com/mohammedsamin/mcpup/releases/latest/badge.svg)](https://github.com/mohammedsamin/mcpup/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/mohammedsamin/mcpup)](https://goreportcard.com/report/github.com/mohammedsamin/mcpup)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

</div>

---

## Why mcpup

MCP servers are useful, but managing them across multiple AI clients is repetitive and error-prone:

- different config locations
- different formats (`JSON`, `JSONC`, `TOML`)
- repeated env/header setup
- easy to overwrite manual client entries
- hard to roll back a bad config write

mcpup gives you one canonical config at `~/.mcpup/config.json` and keeps the client-specific files in sync.

### What you get

- one source of truth for MCP servers
- support for 13 clients
- 97 built-in server templates
- local `stdio` and remote `HTTP/SSE` server definitions
- ownership-aware writes that preserve unmanaged entries
- backups, rollback, drift checks, and `doctor`
- interactive wizard plus full CLI and JSON output

## Quickstart

### Interactive mode

Run `mcpup` with no arguments:

```bash
mcpup
```

Current wizard menu:

```text
? What would you like to do?
  → Quick setup (recommended)
    Add a server
    Remove a server
    Enable / Disable a server
    List servers
    Browse server registry
    Status overview
    Profiles
    Run doctor
    Rollback a client
    Exit
```

### CLI mode

```bash
# Guided onboarding
mcpup setup

# Add a registry server
mcpup add github --env GITHUB_TOKEN=ghp_xxx

# Add a custom local server
mcpup add my-server --command npx --arg -y --arg my-mcp-package

# Add a remote HTTP/SSE server
mcpup add my-remote --url https://api.example.com/mcp --header "Authorization:Bearer sk-xxx"

# Enable it on clients
mcpup enable github --client cursor
mcpup enable github --client claude-code

# Preview changes before writing
mcpup enable github --client codex --dry-run

# Diagnose problems
mcpup doctor

# Roll back a client config
mcpup rollback --client cursor
```

## Install

### Homebrew

```bash
brew tap mohammedsamin/tap
brew install mcpup
```

### Go

```bash
go install github.com/mohammedsamin/mcpup/cmd/mcpup@latest
```

### Binary releases

Download from [Releases](https://github.com/mohammedsamin/mcpup/releases/latest).

## Built-in registry

mcpup ships with **97 curated MCP server templates** so you do not have to chase package names or command syntax by hand.

```bash
mcpup registry
```

Examples:

```bash
mcpup add github --env GITHUB_TOKEN=ghp_xxx
mcpup add notion --env NOTION_TOKEN=ntn_xxx
mcpup add playwright
mcpup add memory
```

The registry includes categories like:

- developer
- search
- productivity
- utility
- database
- automation
- media
- cloud
- ai
- communication
- finance
- devops
- security
- analytics

## Supported clients

mcpup currently manages these clients:

- Claude Code
- Cursor
- Claude Desktop
- Codex
- OpenCode
- Windsurf
- Zed
- Continue
- VS Code
- Cline
- Roo Code
- Amazon Q
- Gemini

For exact config locations and per-client behavior, see [docs/clients.md](docs/clients.md).

## How it works

```text
~/.mcpup/config.json
        |
        +--> mcpup planner + reconciler
                |
                +--> native client config files
```

The core flow is:

1. define a server once in canonical config
2. enable or disable it per client
3. mcpup computes the desired client state
4. it backs up the target config
5. it writes the native client format
6. it validates the result and rolls back on failure

### Safety model

- preserves unmanaged client entries
- creates backups before every write
- supports explicit rollback per client
- warns on destructive managed changes
- validates config shape, executables, env requirements, drift, and ownership via `doctor`

## Common workflows

### Set up a work profile

```bash
mcpup add github --env GITHUB_TOKEN=ghp_xxx
mcpup add slack --env SLACK_BOT_TOKEN=xoxb-xxx
mcpup add notion --env NOTION_TOKEN=ntn_xxx
mcpup add sentry --env SENTRY_AUTH_TOKEN=sntrys_xxx

mcpup profile create work --servers github,slack,notion,sentry
mcpup profile apply work --yes
```

### Update registry-backed definitions

```bash
mcpup update --yes
```

### Export and import server packs

```bash
mcpup export --servers github,notion --output team-pack.json
mcpup import team-pack.json
```

### Script with JSON output

```bash
mcpup list --json | jq '.data.servers[].name'
mcpup status --json | jq '.data.clients'
```

## Command overview

| Command | Description |
|---------|-------------|
| `mcpup` | Launch the interactive wizard |
| `mcpup setup` | Guided onboarding across clients and registry servers |
| `mcpup add <name>` | Add a registry, custom local, or remote server |
| `mcpup update [name...]` | Refresh registry-backed definitions |
| `mcpup enable / disable` | Toggle a server on a client, optionally per tool |
| `mcpup list` | List configured servers |
| `mcpup status` | Show overall status across clients |
| `mcpup export / import` | Share server definitions as JSON |
| `mcpup profile ...` | Create, apply, list, and delete profiles |
| `mcpup registry [query]` | Browse the built-in server catalog |
| `mcpup doctor` | Run diagnostics |
| `mcpup rollback --client <c>` | Restore a backup for one client |
| `mcpup completion <shell>` | Generate shell completions |
| `mcpup clients list` | Show supported clients |

For full command details, see [docs/commands.md](docs/commands.md).

## Documentation

- [Architecture](docs/architecture.md)
- [Commands](docs/commands.md)
- [Clients](docs/clients.md)
- [Config schema](docs/config-schema.md)
- [Examples](docs/examples.md)
- [Troubleshooting](docs/troubleshooting.md)
- [Safety model](docs/safety.md)
- [Contributing](docs/contributing.md)

## Development

```bash
go test ./...
go build ./cmd/mcpup
./mcpup
```

Or use the project helpers:

```bash
make build
make test
make fmt
make lint
```

## License

[MIT](LICENSE)
