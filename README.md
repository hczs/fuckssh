# fuckssh

<p align="center">
  <strong>English</strong> | <a href="README_zh.md">中文</a>
</p>

[![CI](https://github.com/hczs/fuckssh/actions/workflows/ci.yml/badge.svg)](https://github.com/hczs/fuckssh/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hczs/fuckssh)](https://goreportcard.com/report/github.com/hczs/fuckssh)
[![Release](https://img.shields.io/github/v/release/hczs/fuckssh?label=release)](https://github.com/hczs/fuckssh/releases)

**Cross-platform CLI for VPS SSH management. Interactive wizard: from IP + password to passwordless `ssh my-vps` in under 3 minutes — no tutorial needed.**

fuckssh reads and writes your standard `~/.ssh/config`. No proprietary formats, no lock-in. Works with `ssh`, VS Code Remote SSH, Tabby, and any OpenSSH-compatible tool out of the box.

---

## Features

- **One-stop setup** — `fuckssh add` walks you through key generation, public key deployment (password mode), and `ssh config` writing. Also supports "I already have a key, just write the config" mode
- **Non-interactive mode** — `fuckssh add -H <ip> -P <pass>` in one line, perfect for scripts and automation. Tests connectivity first; failure means zero side effects
- **Edit & delete** — `fuckssh edit <alias>` to interactively modify a Host entry; `fuckssh delete <alias>` to remove it along with managed keys
- **Standard files only** — no proprietary config format. You edit `~/.ssh/config` and keys under `~/.ssh`; uninstall fuckssh and everything still works
- **List & search** — `list` shows alias, HostName, port, user, and remark in a table; `search` supports multi-keyword OR, `--user`/`--host`/`--port` field filters, and result highlighting
- **Short aliases** — the install script creates a global `fs` alias; subcommands have short forms too: `ls` / `s` / `v` / `a` / `e` / `d`
- **Safe by default** — auto-backup before modifying config; passwords are used only for the initial connection and never persisted in plaintext
- **Cross-platform** — Windows, macOS, Linux; English and Chinese UI (language selectable on first run)
- **Terminal-native** — [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI wizard and table output

## Installation

### One-liner (recommended)

**macOS / Linux** (requires `curl`, installs to `~/.local/bin`):

```bash
curl -fsSL https://raw.githubusercontent.com/hczs/fuckssh/master/scripts/install.sh | sh
```

Pin a version: `curl -fsSL .../install.sh | sh -s -- --version v0.1.0`

**Windows** (PowerShell):

```powershell
irm https://raw.githubusercontent.com/hczs/fuckssh/master/scripts/install.ps1 | iex
```

The script checks your PATH and guides you if `~/.local/bin` (or `%USERPROFILE%\.local\bin` on Windows) isn't included. After installation, a `fs` alias is ready:

```bash
fs ls        # same as fuckssh list
fs a         # same as fuckssh add
fs s prod    # same as fuckssh search prod
```

### Pre-built binaries

Download from [GitHub Releases](https://github.com/hczs/fuckssh/releases) and put `fuckssh` in your `PATH`.

| Platform | Example artifact |
|----------|-----------------|
| Linux x86_64 / arm64 | `fuckssh_linux_x86_64.tar.gz`, `fuckssh_linux_arm64.tar.gz` |
| macOS Intel | `fuckssh_macos_x86_64.tar.gz` |
| macOS Apple Silicon | `fuckssh_macos_arm64.tar.gz` |
| macOS Universal | `fuckssh_macos_all.tar.gz` (Intel + Apple Silicon) |
| Windows | `fuckssh_windows_x86_64.zip` |

### Go install

Requires Go 1.22+:

```bash
go install github.com/fuckssh/fuckssh@latest
```

### Build from source

```bash
git clone https://github.com/hczs/fuckssh.git
cd fuckssh
make build
```

## Quick Start

**Prerequisite:** OpenSSH client must be installed (`ssh` in PATH). The `add` command detects this and provides install guidance.

```bash
# Add a new VPS (interactive wizard)
fuckssh add              # or: fuckssh a

# Non-interactive: one-liner setup
fuckssh add -H 1.2.3.4 -u root -P mypass -a myserver

# List all Hosts in ~/.ssh/config
fuckssh list             # or: fuckssh ls

# Search (multi-keyword OR, field filters, highlighting)
fuckssh search prod      # or: fuckssh s prod
fuckssh search --user root --port 2222 web

# Edit an existing Host (interactive, pre-filled with current values)
fuckssh edit myserver    # or: fuckssh e myserver

# Delete a Host (with confirmation, removes managed keys too)
fuckssh delete myserver  # or: fuckssh d myserver

# Show version
fuckssh version          # or: fuckssh v

# Use a non-default config path
fuckssh --config /path/to/config ls
```

After adding, just run:

```bash
ssh <your-alias>
```

## Command Reference

| Command | Short | Description |
|---------|-------|-------------|
| `fuckssh add` | `a` | Interactive wizard: password mode (generate key + deploy pubkey + write config) or key mode (write config only) |
| `fuckssh add -H <ip> -P <pass>` | — | Non-interactive: one-line setup (`-H` triggers; `-u` user, `-p` port, `-a` alias, `-i` identity file, `-r` remark) |
| `fuckssh list` | `ls` | Parse and display all Hosts in a table; multiple aliases shown comma-separated |
| `fuckssh search <query>` | `s` | Multi-keyword OR search; `--user`/`--host`/`--port` filters; highlighted results in TTY |
| `fuckssh edit <alias>` | `e` | Interactive edit of an existing Host (pre-filled, supports "go back and revise" retry) |
| `fuckssh delete <alias>` | `d` | Delete Host with confirmation; removes managed IdentityFile keys too |
| `fuckssh version` | `v` | Show version, commit, and build date (injected via ldflags in release builds) |

**Global options**

| Option | Description |
|--------|-------------|
| `--config <path>` | Path to ssh config file (default `~/.ssh/config`) |
| `-h`, `--help` | Help (English and Chinese supported) |

**Command options**

| Option | Commands | Description |
|--------|----------|-------------|
| `-H`, `--host` | add | Target host address (triggers non-interactive mode) |
| `-u`, `--user` | add | Login user (default: root) |
| `-p`, `--port` | add | SSH port (default: 22) |
| `-P`, `--password` | add | SSH password (triggers password mode) |
| `-i`, `--identity-file` | add | Private key path (triggers key mode) |
| `-a`, `--alias` | add | Host alias (auto-generated from address if omitted) |
| `-r`, `--remark` | add | Remark / note |
| `--user` | search | Filter by username |
| `--host` | search | Filter by host address |
| `--port` | search | Filter by port |
| `-f`, `--force` | delete | Skip confirmation prompt |

## Demo

### add — Interactive VPS setup

```bash
$ fuckssh add

  Enter VPS details:
  > Address: 1.2.3.4
    Port: 22
    User: root
    Alias: myserver

  ✓ Key pair generated
  ✓ Public key deployed to 1.2.3.4
  ✓ Written to ~/.ssh/config

  ssh myserver
```

### add — Non-interactive mode

```bash
$ fuckssh add -H 1.2.3.4 -u root -P mypass -a myserver

  ✓ Connection test passed
  ✓ Key pair generated
  ✓ Public key deployed to 1.2.3.4
  ✓ Written to ~/.ssh/config

  ssh myserver
  Completed in 3.2s
```

### list — Show all Hosts

```bash
$ fuckssh ls

Reading: /home/user/.ssh/config
3 hosts found

┌──────────────┬───────────────┬──────┬────────┬──────────────┐
│ Alias        │ Address       │ Port │ User   │ Remark       │
├──────────────┼───────────────┼──────┼────────┼──────────────┤
│ myserver     │ 1.2.3.4       │   22 │ root   │ Production   │
│ dev          │ 10.0.0.1      │   22 │ ubuntu │ Dev machine  │
│ staging      │ 172.16.0.5    │ 2222 │ admin  │ Staging      │
└──────────────┴───────────────┴──────┴────────┴──────────────┘
```

### search — Find hosts

```bash
$ fuckssh s prod

Search: prod — 1 match

┌──────────────┬───────────────┬──────┬────────┬──────────────┐
│ Alias        │ Address       │ Port │ User   │ Remark       │
├──────────────┼───────────────┼──────┼────────┼──────────────┤
│ myserver     │ 1.2.3.4       │   22 │ root   │ Production   │
└──────────────┴───────────────┴──────┴────────┴──────────────┘
```

### edit — Modify an existing Host

```bash
$ fuckssh e myserver

  Edit VPS details (Enter to keep current value):
  > Alias: myserver
    Address: 1.2.3.4
    Port: 2222          ← changed
    User: root

  ✓ ~/.ssh/config updated

  ssh myserver
```

### delete — Remove a Host

```bash
$ fuckssh d myserver

  Delete myserver (root@1.2.3.4)? [y/N] y

  ✓ Deleted myserver
  ✓ Removed managed key /home/user/.ssh/id_ed25519_fuckssh_myserver
```

### version

```bash
$ fuckssh v

v0.2.0 (abc1234, 2026-06-03)
```

## Design Principles

1. **Standard-first** — Reads and writes OpenSSH conventions. Uninstall fuckssh and your config still works.
2. **Recoverable** — Backs up config before every write. Low cost of mistakes.
3. **CLI-native** — Built for developers who live in the terminal and VS Code Remote SSH. Complements GUI clients, doesn't replace them.
4. **MVP-focused** — No encrypted cross-device sync yet. See roadmap below.

## Roadmap

| Phase | Planned capabilities |
|-------|---------------------|
| **Current** | `add` (interactive + non-interactive) / `list` / `search` (multi-keyword + field filters + highlighting) / `edit` / `delete` / `version` / short aliases |
| **V2** | Encrypted backup & restore, `ssh config` + private key cross-device sync, alias conflict merging |

## Development

**Prerequisites:**

- Go 1.22+
- [golangci-lint](https://golangci-lint.run/) (for local linting and pre-commit)
- `goimports`: `go install golang.org/x/tools/cmd/goimports@latest`

```bash
make build    # compile to bin/fuckssh
make test     # go test ./...
make lint     # golangci-lint
make fmt      # gofmt + goimports
make run      # go run ./cmd/fuckssh
make hooks    # enable pre-commit (auto fmt + lint on staged .go files)
```

After cloning, run `make hooks` once. Every `git commit` will then auto-format and lint staged `.go` files.

Push a `v*` tag (e.g. `v0.1.0`) to trigger the [Release workflow](.github/workflows/release.yml) — GoReleaser builds and publishes to GitHub Releases. Dry run locally: `make release-dry`.

| Trigger | Workflow | Description |
|---------|----------|-------------|
| push / PR → `main` / `master` | [CI](.github/workflows/ci.yml) | Test, lint, cross-platform build check |

## Documentation

| Doc | Description |
|-----|-------------|
| [PRD](docs/fuckssh-PRD.md) | Product requirements & scenarios |
| [Tech Selection](docs/fuckssh-技术选型.md) | Tech stack & module breakdown |
| [Architecture](docs/fuckssh-架构设计.md) | System architecture ([HTML](docs/fuckssh-架构设计.html)) |
| [AGENTS.md](AGENTS.md) | Contributor & AI collaboration conventions |

## Contributing

Issues and pull requests are welcome. Please ensure `make test` passes before submitting. With `make hooks` enabled, commits are auto-formatted and linted. For larger changes, open an Issue first to discuss direction.

## License

License to be determined. Until then, please do not use this project for commercial redistribution that requires an explicit open-source license. Watch this repo for updates.
