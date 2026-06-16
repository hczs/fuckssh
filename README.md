# fuckssh

<p align="center">
  <strong>English</strong> | <a href="README_zh.md">中文</a>
</p>

[![CI](https://github.com/hczs/fuckssh/actions/workflows/ci.yml/badge.svg)](https://github.com/hczs/fuckssh/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hczs/fuckssh)](https://goreportcard.com/report/github.com/hczs/fuckssh)
[![Release](https://img.shields.io/badge/release-v0.5.0-blue)](https://github.com/hczs/fuckssh/releases)

> Cross-platform CLI for VPS SSH — from IP + password to `ssh my-vps` in minutes, using only standard `~/.ssh/config`.

## About

Setting up passwordless SSH on a new VPS usually means juggling `ssh-keygen`, uploading public keys, and hand-editing `~/.ssh/config`. fuckssh walks you through that in an interactive wizard — or in one non-interactive command for scripts.

It reads and writes standard OpenSSH files only. No proprietary formats, no lock-in. Uninstall fuckssh and everything still works with `ssh`, VS Code Remote SSH, Tabby, and any OpenSSH-compatible tool.

## Table of contents

- [About](#about)
- [Features](#features)
- [Use cases](#use-cases)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Quick start](#quick-start)
- [Command reference](#command-reference)
- [Examples](#examples)
  - [add](#add--interactive-vps-setup)
  - [list](#list--show-all-hosts)
  - [search](#search--find-hosts)
  - [edit](#edit--modify-an-existing-host)
  - [delete](#delete--remove-a-host)
  - [export](#export--encrypted-backup)
  - [import](#import--restore-on-another-machine)
  - [version](#version)
- [Design principles](#design-principles)
- [Roadmap](#roadmap)
- [Development](#development)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [License](#license)

## Features

- **One-stop setup** — `add` password mode: generate Ed25519 key, deploy public key, write config; key mode: write config for an existing private key
- **Script-friendly** — `add -H <ip> -P <pass>` for non-interactive setup; tests connectivity before writing any files
- **Manage hosts** — `list` / `search` / `edit` / `delete` for the full host lifecycle
- **Encrypted migration** — `export` / `import` to move config and private keys across machines (Argon2id + AES-256-GCM)
- **Safe by default** — auto-backup before every config write; passwords used only for the initial connection, never stored in plaintext
- **Standard files only** — no proprietary config format; uninstall and your setup still works
- **Cross-platform** — Windows, macOS, Linux; English and Chinese UI (`FUCKSSH_LANG` or language selection on first run)
- **Terminal-native** — [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI wizard and table output

## Use cases

**New VPS — passwordless login in one session**

```bash
fuckssh add
ssh myserver
```

**Script or automation — non-interactive add**

```bash
fuckssh add -H 1.2.3.4 -u root -P mypass -a myserver
ssh myserver
```

**Migrate to a new laptop**

```bash
# On the old machine
fuckssh export ~/Desktop

# Copy fuckssh-backup-*.tar.enc to the new machine, then:
fuckssh import fuckssh-backup-20260616-102739.tar.enc
```

**Find a host quickly**

```bash
fuckssh search prod
fuckssh search --user root --port 2222 web
```

## Prerequisites

- OpenSSH client (`ssh` in PATH) — required for `add` and connectivity tests
- Writable `~/.ssh/` directory
- Go 1.22+ — only if building from source

## Installation

### One-liner (recommended)

**macOS / Linux** (requires `curl`, installs to `~/.local/bin`):

```bash
curl -fsSL https://raw.githubusercontent.com/hczs/fuckssh/master/scripts/install.sh | sh
```

Pin a version: `curl -fsSL .../install.sh | sh -s -- --version v0.5.0`

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

See [Development](#development).

## Quick start

```bash
fuckssh add
ssh <your-alias>
```

For more scenarios, see [Use cases](#use-cases).

## Command reference

| Command | Short | Description |
|---------|-------|-------------|
| `fuckssh add` | `a` | Interactive wizard: password mode (generate key + deploy pubkey + write config) or key mode (write config only) |
| `fuckssh add -H <ip> -P <pass>` | — | Non-interactive: one-line setup (`-H` triggers; `-u` user, `-p` port, `-a` alias, `-i` identity file, `-r` remark) |
| `fuckssh list` | `ls` | Parse and display all Hosts in a table; multiple aliases shown comma-separated |
| `fuckssh search <query>` | `s` | Multi-keyword OR search; `--user`/`--host`/`--port` filters; highlighted results in TTY |
| `fuckssh edit <alias>` | `e` | Interactive edit of an existing Host (pre-filled, supports "go back and revise" retry) |
| `fuckssh delete <alias>` | `d` | Delete Host with confirmation; removes managed IdentityFile keys too |
| `fuckssh export [dir]` | — | Pack `~/.ssh/config` and private keys into an encrypted `.tar.enc` backup |
| `fuckssh import <file>` | — | Decrypt and merge backup; interactive conflict resolution (overwrite / skip / rename) |
| `fuckssh version` | `v` | Show version, commit, and build date (injected via ldflags in release builds) |

**Global options**

| Option | Description |
|--------|-------------|
| `-h`, `--help` | Help (English and Chinese supported) |
| `--version` | Print version information |

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

## Examples

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
  ✓ Removed managed key /home/user/.ssh/keys/id_ed25519_myserver
```

### export — Encrypted backup

```bash
$ fuckssh export ~/Desktop

请设置主密码（至少 6 位，不能纯数字）：
主密码: ********
确认密码: ********
正在导出...
✓ 导出成功
  文件: /home/user/Desktop/fuckssh-backup-20260616-102739.tar.enc
  大小: 4096 字节
  包含: 3 个 Host 配置, 2 个私钥

请将此文件拷贝到目标机器，然后运行:
  fuckssh import fuckssh-backup-20260616-102739.tar.enc
```

### import — Restore on another machine

No conflicts:

```bash
$ fuckssh import fuckssh-backup-20260616-102739.tar.enc

请输入主密码: ********
正在解密...
✓ 发现 3 个 Host 配置，无冲突
✓ 导入完成
```

With conflicts:

```bash
$ fuckssh import fuckssh-backup-20260616-102739.tar.enc

请输入主密码: ********
正在解密...
✓ 发现 3 个 Host 配置，其中 1 个与现有配置冲突：

  [myserver] root@1.2.3.4:22
      [1] 覆盖  [2] 跳过  [3] 重命名
> 3
  新别名: myserver-new

✓ 导入完成（新增 2，覆盖 0，重命名 1，跳过 0）
```

### version

```bash
$ fuckssh v

v0.5.0 (abc1234, 2026-06-16)
```

## Design principles

1. **Standard-first** — Reads and writes OpenSSH conventions. Uninstall fuckssh and your config still works.
2. **Recoverable** — Backs up config before every write (`~/.ssh/backup/`); import also backs up existing config before merging.
3. **CLI-native** — Built for developers who live in the terminal and VS Code Remote SSH. Complements GUI clients, doesn't replace them.
4. **Portable** — `export` / `import` for moving hosts and keys to a new machine. Backup files are written with mode `0600`; keep them safe if you store them on a sync drive.

## Roadmap

| Phase | Capabilities |
|-------|-------------|
| **Shipped** | `add` (interactive + non-interactive) / `list` / `search` / `edit` / `delete` / `export` / `import` / `version` / short aliases |
| **Next** | `Include` directive expansion, `--json` output, custom config path |

## Development

**Prerequisites:**

- Go 1.22+
- [golangci-lint](https://golangci-lint.run/) (for local linting and pre-commit)
- `goimports`: `go install golang.org/x/tools/cmd/goimports@latest`

```bash
git clone https://github.com/hczs/fuckssh.git
cd fuckssh
make build    # compile to bin/fuckssh
make test     # go test ./...
make lint     # golangci-lint
make fmt      # gofmt + goimports
make run      # go run ./cmd/fuckssh
make hooks    # enable pre-commit (auto fmt + lint on staged .go files)
```

After cloning, run `make hooks` once. Every `git commit` will then auto-format and lint staged `.go` files.

Push a `v*` tag (e.g. `v0.5.0`) to trigger the [Release workflow](.github/workflows/release.yml) — GoReleaser builds and publishes to GitHub Releases. Dry run locally: `make release-dry`.

| Trigger | Workflow | Description |
|---------|----------|-------------|
| push / PR → `main` / `master` | [CI](.github/workflows/ci.yml) | Test, lint, cross-platform build check |

## Documentation

| Doc | Description |
|-----|-------------|
| [PRD](docs/fuckssh-PRD.md) | Product requirements & scenarios |
| [Tech Selection](docs/fuckssh-技术选型.md) | Tech stack & module breakdown |
| [Architecture](docs/fuckssh-架构设计.md) | System architecture ([HTML](docs/fuckssh-架构设计.html)) |
| [Command flow](docs/command-flow.md) | Sequence diagrams for add, list, search, export, import |
| [AGENTS.md](AGENTS.md) | Contributor & AI collaboration conventions |

## Contributing

Issues and pull requests are welcome. Please ensure `make test` passes before submitting. With `make hooks` enabled, commits are auto-formatted and linted. For larger changes, open an Issue first to discuss direction.

## License

License to be determined. Until then, please do not use this project for commercial redistribution that requires an explicit open-source license. Watch this repo for updates.
