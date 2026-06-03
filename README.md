# fuckssh

[![CI](https://github.com/hczs/fuckssh/actions/workflows/ci.yml/badge.svg)](https://github.com/hczs/fuckssh/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hczs/fuckssh)](https://goreportcard.com/report/github.com/hczs/fuckssh)
[![Release](https://img.shields.io/github/v/release/hczs/fuckssh?label=release)](https://github.com/hczs/fuckssh/releases)

**跨平台 CLI：用交互式向导管理 VPS，只读写标准 `~/.ssh/config`。**

新机器从「有 IP 和密码」到 `ssh my-vps` 免密登录，不必再翻教程拼 `ssh-keygen`、公钥上传和 config 字段。列表与搜索直接解析本机 OpenSSH 配置，与系统 `ssh`、VS Code Remote SSH、Tabby 等工具完全互通。

---

## 特性

- **一站式添加** — `fuckssh add` 引导完成密钥生成、公钥部署（密码模式）与 `ssh config` 写入；支持「已有私钥、仅补 config」模式
- **非交互模式** — `fuckssh add -H <ip> -P <pass>` 一行命令完成配置，适合脚本与自动化场景；先测连通性再动文件，失败零副作用
- **编辑与删除** — `fuckssh edit <alias>` 交互式修改已有 Host 条目；`fuckssh delete <alias>` 确认后删除条目及关联密钥
- **只认标准文件** — 不引入私有配置格式；改的是 `~/.ssh/config` 与 `~/.ssh` 下的密钥，未安装本工具时仍可用 `ssh`
- **列表与搜索** — `list` 展示别名、HostName、端口、用户与备注；`search` 支持多关键词 OR 搜索、`--user`/`--host`/`--port` 字段过滤与结果高亮
- **短别名** — 所有命令支持短别名：`ls` / `s` / `v` / `a` / `e` / `d`，打字更快
- **安全习惯** — 修改 config 前自动备份；密码仅用于首次连接，不落盘明文
- **跨平台** — Windows、macOS、Linux；中英文界面（首次运行可选语言）
- **终端友好** — 基于 [Bubble Tea](https://github.com/charmbracelet/bubbletea) 的 TUI 向导与表格输出

## 安装

### 一键安装（推荐）

**macOS / Linux**（需要 `curl`，将二进制安装到 `~/.local/bin`）：

```bash
curl -fsSL https://raw.githubusercontent.com/hczs/fuckssh/master/scripts/install.sh | sh
```

指定版本：`curl -fsSL .../install.sh | sh -s -- --version v0.1.0`

**Windows**（PowerShell）：

```powershell
irm https://raw.githubusercontent.com/hczs/fuckssh/master/scripts/install.ps1 | iex
```

脚本会检测 PATH；若 `~/.local/bin`（或 Windows 下 `%USERPROFILE%\.local\bin`）未在 PATH 中，会提示如何添加。实现见 [`scripts/install.sh`](scripts/install.sh)、[`scripts/install.ps1`](scripts/install.ps1)。

**本地调试：** Windows 默认没有 `sh`，请用 PowerShell 运行 `.\scripts\install.ps1 -BinDir .\bin`；`install.sh` 需在 [Git Bash](https://git-scm.com/) 或 WSL 中执行。

### 预编译二进制（手动）

在 [GitHub Releases](https://github.com/hczs/fuckssh/releases) 下载对应平台的压缩包，解压后将 `fuckssh` 放入 `PATH`。

| 平台 | 常见产物示例 |
|------|----------------|
| Linux x86_64 / arm64 | `fuckssh_linux_x86_64.tar.gz`、`fuckssh_linux_arm64.tar.gz` |
| macOS Intel | `fuckssh_macos_x86_64.tar.gz` |
| macOS Apple Silicon | `fuckssh_macos_arm64.tar.gz` |
| macOS 通用二进制 | `fuckssh_macos_all.tar.gz`（Intel + Apple Silicon） |
| Windows | `fuckssh_windows_x86_64.zip` |

### Go 安装

需要 Go 1.26+（与仓库 `go.mod` 一致）：

```bash
go install github.com/fuckssh/fuckssh@latest
```

指定版本：`go install github.com/fuckssh/fuckssh@v0.1.0`

### 从源码构建

```bash
git clone https://github.com/hczs/fuckssh.git
cd fuckssh
go build -o bin/fuckssh ./cmd/fuckssh
```

或使用 Makefile：`make build`

## 快速开始

**前提：** 本机已安装 OpenSSH 客户端（`ssh` 在 PATH 中）。`add` 子命令会检测并给出安装指引。

```bash
# 添加一台新 VPS（交互式向导）
fuckssh add              # 或 fuckssh a

# 非交互模式：一行命令搞定
fuckssh add -H 1.2.3.4 -u root -P mypass -a myserver

# 列出 ~/.ssh/config 中的 Host
fuckssh list             # 或 fuckssh ls

# 搜索（多关键词 OR、字段过滤、高亮）
fuckssh search prod      # 或 fuckssh s prod
fuckssh search --user root --port 2222 web

# 编辑已有 Host（交互式向导，预填现有值）
fuckssh edit myserver    # 或 fuckssh e myserver

# 删除 Host（确认后删除条目及关联密钥）
fuckssh delete myserver  # 或 fuckssh d myserver

# 查看版本
fuckssh version          # 或 fuckssh v

# 使用非默认 config 路径
fuckssh --config /path/to/config ls
```

添加完成后，按向导提示即可直接执行：

```bash
ssh <你设置的别名>
```

## 命令参考

| 命令 | 短别名 | 说明 |
|------|--------|------|
| `fuckssh add` | `a` | 交互式向导：密码模式（生成密钥 + 部署公钥 + 写 config）或密钥模式（仅写 config） |
| `fuckssh add -H <ip> -P <pass>` | — | 非交互模式：一行命令完成配置（`-H` 触发；`-u` 用户、`-p` 端口、`-a` 别名、`-i` 私钥、`-r` 备注） |
| `fuckssh list` | `ls` | 解析并表格展示所有 Host；多别名以逗号分隔 |
| `fuckssh search <query>` | `s` | 多关键词 OR 搜索；`--user`/`--host`/`--port` 字段过滤；TTY 下结果高亮 |
| `fuckssh edit <alias>` | `e` | 交互式编辑已有 Host 条目（预填现有值，支持"返回修改"重试） |
| `fuckssh delete <alias>` | `d` | 确认后删除 Host 条目；若 IdentityFile 为 fuckssh 管理的密钥则一并删除 |
| `fuckssh version` | `v` | 显示版本、提交与构建时间（Release 构建通过 ldflags 注入） |

**全局选项**

| 选项 | 说明 |
|------|------|
| `--config <path>` | 指定 ssh config 文件（默认 `~/.ssh/config`） |
| `-h`, `--help` | 帮助（支持中英文） |

**命令选项**

| 选项 | 适用命令 | 说明 |
|------|----------|------|
| `-H`, `--host` | add | 目标主机地址（触发非交互模式） |
| `-u`, `--user` | add | 登录用户（默认 root） |
| `-p`, `--port` | add | SSH 端口（默认 22） |
| `-P`, `--password` | add | SSH 密码（触发密码模式） |
| `-i`, `--identity-file` | add | 私钥路径（触发密钥模式） |
| `-a`, `--alias` | add | Host 别名（不填则自动从地址生成） |
| `-r`, `--remark` | add | 备注信息 |
| `--user` | search | 按用户名过滤 |
| `--host` | search | 按主机地址过滤 |
| `--port` | search | 按端口过滤 |
| `-f`, `--force` | delete | 跳过确认提示 |

## Demo 演示

### add — 交互式添加 VPS

```bash
$ fuckssh add

  请输入 VPS 信息：
  > 地址: 1.2.3.4
    端口: 22
    用户: root
    别名: myserver

  ✓ 已生成密钥对
  ✓ 已部署公钥到 1.2.3.4
  ✓ 已写入 ~/.ssh/config

  ssh myserver
```

### add — 非交互模式

```bash
$ fuckssh add -H 1.2.3.4 -u root -P mypass -a myserver

  ✓ 连接测试通过
  ✓ 已生成密钥对
  ✓ 已部署公钥到 1.2.3.4
  ✓ 已写入 ~/.ssh/config

  ssh myserver
  执行耗时 3.2s
```

### list — 列出所有 Host

```bash
$ fuckssh ls

读取: /home/user/.ssh/config
共 3 台主机

┌──────────────┬───────────────┬──────┬────────┬──────────────┐
│ 别名         │ 地址          │ 端口 │ 用户   │ 备注         │
├──────────────┼───────────────┼──────┼────────┼──────────────┤
│ myserver     │ 1.2.3.4       │   22 │ root   │ 生产环境     │
│ dev          │ 10.0.0.1      │   22 │ ubuntu │ 开发机       │
│ staging      │ 172.16.0.5    │ 2222 │ admin  │ 预发布       │
└──────────────┴───────────────┴──────┴────────┴──────────────┘
```

### search — 搜索 Host

```bash
$ fuckssh s prod

搜索: prod，命中 1 台

┌──────────────┬───────────────┬──────┬────────┬──────────────┐
│ 别名         │ 地址          │ 端口 │ 用户   │ 备注         │
├──────────────┼───────────────┼──────┼────────┼──────────────┤
│ myserver     │ 1.2.3.4       │   22 │ root   │ 生产环境     │
└──────────────┴───────────────┴──────┴────────┴──────────────┘
```

```bash
$ fuckssh search --user root --port 2222

搜索: (user=root, port=2222)，命中 1 台

┌──────────────┬───────────────┬──────┬────────┬──────────────┐
│ 别名         │ 地址          │ 端口 │ 用户   │ 备注         │
├──────────────┼───────────────┼──────┼────────┼──────────────┤
│ staging      │ 172.16.0.5    │ 2222 │ admin  │ 预发布       │
└──────────────┴───────────────┴──────┴────────┴──────────────┘
```

### edit — 编辑已有 Host

```bash
$ fuckssh e myserver

  请修改 VPS 信息（回车保留原值）：
  > 别名: myserver
    地址: 1.2.3.4
    端口: 2222          ← 修改端口
    用户: root

  ✓ 已更新 ~/.ssh/config

  ssh myserver
```

### delete — 删除 Host

```bash
$ fuckssh d myserver

  确认删除 myserver (root@1.2.3.4)？[y/N] y

  ✓ 已删除 myserver
  ✓ 已删除关联密钥 /home/user/.ssh/id_ed25519_fuckssh_myserver
```

```bash
# 跳过确认（脚本场景）
$ fuckssh delete myserver --force

  ✓ 已删除 myserver
```

### version — 查看版本

```bash
$ fuckssh v

v0.2.0 (abc1234, 2026-06-03)
```

## 设计原则

1. **标准优先** — 配置来源与落盘均为 OpenSSH 约定路径与语法。
2. **可恢复** — 写入前备份现有 config，降低误操作成本。
3. **CLI 原生** — 面向习惯终端与 Remote SSH 的开发者，不替代 GUI 客户端，与之互补。
4. **MVP 聚焦** — 当前不提供跨机加密同步；见下方路线图。

## 路线图

| 阶段 | 计划能力 |
|------|----------|
| **当前** | `add`（交互 + 非交互）/ `list` / `search`（多关键词 + 字段过滤 + 高亮）/ `edit` / `delete` / `version` / 短别名 |
| **V2** | 加密备份与恢复、`ssh config` + 私钥跨设备同步、别名冲突合并 |

产品细节见 [PRD](docs/fuckssh-PRD.md)。

## 开发

**环境**

- Go 1.26+
- [golangci-lint](https://golangci-lint.run/)（本地 lint 与 pre-commit 需要）
- `goimports`：`go install golang.org/x/tools/cmd/goimports@latest`

```bash
make build    # 构建到 bin/fuckssh
make test     # go test ./...
make lint     # golangci-lint
make fmt      # gofmt + goimports
make run      # go run ./cmd/fuckssh
make hooks    # 一次性启用 pre-commit（提交时自动 fmt + lint）
```

克隆仓库后建议执行一次 `make hooks`，之后每次 `git commit` 会对**已暂存**的 `.go` 文件运行 `gofmt`/`goimports` 并执行 `golangci-lint`。

推送 `v*` 标签（如 `v0.1.0`）会触发 [Release 工作流](.github/workflows/release.yml)，由 GoReleaser 构建并发布到 GitHub Releases。本地试跑：`make release-dry`（需安装 [GoReleaser](https://goreleaser.com/)）。

| 触发 | 工作流 | 说明 |
|------|--------|------|
| push / PR → `main` / `master` | [CI](.github/workflows/ci.yml) | 测试、lint、多平台编译检查 |

## 文档

| 文档 | 说明 |
|------|------|
| [PRD](docs/fuckssh-PRD.md) | 产品需求与场景 |
| [技术选型](docs/fuckssh-技术选型.md) | 技术栈与模块划分 |
| [系统架构](docs/fuckssh-架构设计.md) | 架构设计（[HTML 版](docs/fuckssh-架构设计.html)） |
| [AGENTS.md](AGENTS.md) | 贡献者与 AI 协作约定 |

## 参与贡献

欢迎 Issue 与 Pull Request。提交前请确保 `make test` 通过；若已 `make hooks`，commit 时会自动 fmt/lint。较大改动请先开 Issue 讨论方向。

## 许可证

许可证尚未确定。在明确之前，请勿将本项目用于需要明确开源许可证的商业再分发场景；关注本仓库后续更新。
