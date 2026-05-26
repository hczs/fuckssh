# fuckssh

[![CI](https://github.com/hczs/fuckssh/actions/workflows/ci.yml/badge.svg)](https://github.com/hczs/fuckssh/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hczs/fuckssh)](https://goreportcard.com/report/github.com/hczs/fuckssh)
[![Release](https://img.shields.io/github/v/release/hczs/fuckssh?label=release)](https://github.com/hczs/fuckssh/releases)

**跨平台 CLI：用交互式向导管理 VPS，只读写标准 `~/.ssh/config`。**

新机器从「有 IP 和密码」到 `ssh my-vps` 免密登录，不必再翻教程拼 `ssh-keygen`、公钥上传和 config 字段。列表与搜索直接解析本机 OpenSSH 配置，与系统 `ssh`、VS Code Remote SSH、Tabby 等工具完全互通。

---

## 特性

- **一站式添加** — `fuckssh add` 引导完成密钥生成、公钥部署（密码模式）与 `ssh config` 写入；支持「已有私钥、仅补 config」模式
- **只认标准文件** — 不引入私有配置格式；改的是 `~/.ssh/config` 与 `~/.ssh` 下的密钥，未安装本工具时仍可用 `ssh`
- **列表与搜索** — `list` 展示别名、HostName、端口、用户与备注；`search` 按别名、域名或 IP 模糊匹配
- **安全习惯** — 修改 config 前自动备份；密码仅用于首次连接，不落盘明文
- **跨平台** — Windows、macOS、Linux；中英文界面（首次运行可选语言）
- **终端友好** — 基于 [Bubble Tea](https://github.com/charmbracelet/bubbletea) 的 TUI 向导与表格输出

## 安装

### 预编译二进制

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
fuckssh add

# 列出 ~/.ssh/config 中的 Host
fuckssh list

# 搜索（别名、HostName、IP）
fuckssh search prod

# 查看版本
fuckssh version

# 使用非默认 config 路径
fuckssh --config /path/to/config list
```

添加完成后，按向导提示即可直接执行：

```bash
ssh <你设置的别名>
```

## 命令参考

| 命令 | 说明 |
|------|------|
| `fuckssh add` | 交互式向导：密码模式（生成密钥 + 部署公钥 + 写 config）或密钥模式（仅写 config） |
| `fuckssh list` | 解析并表格展示所有 Host；多别名以逗号分隔 |
| `fuckssh search <query>` | 关键词匹配别名、HostName、IP |
| `fuckssh version` | 显示版本、提交与构建时间（Release 构建通过 ldflags 注入） |

**全局选项**

| 选项 | 说明 |
|------|------|
| `--config <path>` | 指定 ssh config 文件（默认 `~/.ssh/config`） |
| `-h`, `--help` | 帮助（支持中英文） |

## 设计原则

1. **标准优先** — 配置来源与落盘均为 OpenSSH 约定路径与语法。
2. **可恢复** — 写入前备份现有 config，降低误操作成本。
3. **CLI 原生** — 面向习惯终端与 Remote SSH 的开发者，不替代 GUI 客户端，与之互补。
4. **MVP 聚焦** — 当前不提供 config 的编辑/删除与跨机加密同步；见下方路线图。

## 路线图

| 阶段 | 计划能力 |
|------|----------|
| **当前 (MVP)** | `add` / `list` / `search`、密码与密钥两种添加路径 |
| **V2** | 加密备份与恢复、`ssh config` + 私钥跨设备同步、别名冲突合并 |
| **后续** | 列表内编辑/删除 Host 等管理能力 |

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
