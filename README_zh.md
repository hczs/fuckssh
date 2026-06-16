# fuckssh

<p align="center">
  <a href="README.md">English</a> | <strong>中文</strong>
</p>

[![CI](https://github.com/hczs/fuckssh/actions/workflows/ci.yml/badge.svg)](https://github.com/hczs/fuckssh/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/hczs/fuckssh)](https://goreportcard.com/report/github.com/hczs/fuckssh)
[![Release](https://img.shields.io/badge/release-v0.6.0-blue)](https://github.com/hczs/fuckssh/releases)

> 跨平台 VPS SSH 配置 CLI：IP + 密码几分钟搞定免密登录，只读写标准 `~/.ssh/config`。

## 项目简介

新机器上要免密 SSH，通常得自己跑 `ssh-keygen`、上传公钥、手改 `~/.ssh/config`。fuckssh 用交互式向导一步步带你完成；脚本场景也可以一行命令非交互添加。

只读写标准 OpenSSH 文件，不引入专有格式，没有厂商锁定。卸载后配置照常可用，与系统 `ssh`、VS Code Remote SSH、Tabby 等工具完全互通。

## 目录

- [项目简介](#项目简介)
- [特性](#特性)
- [典型工作流](#典型工作流)
- [前提条件](#前提条件)
- [安装](#安装)
- [快速开始](#快速开始)
- [命令参考](#命令参考)
- [演示](#演示)
  - [add](#add--交互式添加-vps)
  - [list](#list--列出所有-host)
  - [search](#search--搜索-host)
  - [edit](#edit--编辑已有-host)
  - [delete](#delete--删除-host)
  - [export](#export--加密导出)
  - [import](#import--导入到新机器)
  - [version](#version--查看版本)
- [设计原则](#设计原则)
- [路线图](#路线图)
- [开发](#开发)
- [文档](#文档)
- [参与贡献](#参与贡献)
- [许可证](#许可证)

## 特性

- **一站式添加** — `add` 密码模式：生成 Ed25519 密钥、部署公钥、写 config；密钥模式：已有私钥只写 config
- **非交互友好** — `add -H <ip> -P <pass>` 一行完成；先测连通性再写文件，失败零副作用
- **主机全生命周期** — `list` / `search` / `edit` / `delete` 列表、搜索、编辑、删除
- **加密迁移** — `export` / `import` 跨机迁移 config 与私钥（Argon2id + AES-256-GCM）
- **安全习惯** — 每次写 config 前自动备份；密码仅用于首次连接，不落盘明文
- **只认标准文件** — 无专有格式；卸载后仍可用 `ssh`
- **跨平台** — Windows、macOS、Linux；中英文界面（`FUCKSSH_LANG` 或首次运行选择语言）
- **终端友好** — 基于 [Bubble Tea](https://github.com/charmbracelet/bubbletea) 的 TUI 向导与表格输出

## 典型工作流

**新 VPS — 一次会话完成免密登录**

```bash
fuckssh add
ssh myserver
```

**脚本或自动化 — 非交互添加**

```bash
fuckssh add -H 1.2.3.4 -u root -P mypass -a myserver
ssh myserver
```

**换机迁移**

```bash
# 旧机器
fuckssh export ~/Desktop

# 将 fuckssh-backup-*.tar.enc 拷到新机器后：
fuckssh import fuckssh-backup-20260616-102739.tar.enc
```

**快速找一台机器**

```bash
fuckssh search prod
fuckssh search --user root --port 2222 web
```

## 前提条件

- 已安装 OpenSSH 客户端（`ssh` 在 PATH 中）— `add` 与连通性测试需要
- 可写的 `~/.ssh/` 目录
- Go 1.22+ — 仅源码构建时需要

## 安装

### 一键安装（推荐）

**macOS / Linux**（需要 `curl`，将二进制安装到 `~/.local/bin`）：

```bash
curl -fsSL https://raw.githubusercontent.com/hczs/fuckssh/master/scripts/install.sh | sh
```

指定版本：`curl -fsSL .../install.sh | sh -s -- --version v0.6.0`

**Windows**（PowerShell）：

```powershell
irm https://raw.githubusercontent.com/hczs/fuckssh/master/scripts/install.ps1 | iex
```

脚本会检测 PATH；若 `~/.local/bin`（或 Windows 下 `%USERPROFILE%\.local\bin`）未在 PATH 中，会提示如何添加。安装完成后会自动创建 `fs` 短别名：

```bash
fs ls        # 等同于 fuckssh list
fs a         # 等同于 fuckssh add
fs s prod    # 等同于 fuckssh search prod
```

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

需要 Go 1.22+：

```bash
go install github.com/fuckssh/fuckssh@latest
```

### 从源码构建

见 [开发](#开发)。

## 快速开始

```bash
fuckssh add
ssh <你设置的别名>
```

更多场景见 [典型工作流](#典型工作流)。

## 命令参考

| 命令 | 短别名 | 说明 |
|------|--------|------|
| `fuckssh add` | `a` | 交互式向导：密码模式（生成密钥 + 部署公钥 + 写 config）或密钥模式（仅写 config） |
| `fuckssh add -H <ip> -P <pass>` | — | 非交互模式：一行命令完成配置（`-H` 触发；`-u` 用户、`-p` 端口、`-a` 别名、`-i` 私钥、`-r` 备注） |
| `fuckssh list` | `ls` | 解析并表格展示所有 Host；多别名以逗号分隔 |
| `fuckssh search <query>` | `s` | 多关键词 OR 搜索；`--user`/`--host`/`--port` 字段过滤；TTY 下结果高亮 |
| `fuckssh edit <alias>` | `e` | 交互式编辑已有 Host 条目（预填现有值，支持「返回修改」重试） |
| `fuckssh delete <alias>` | `d` | 确认后删除 Host 条目；若 IdentityFile 为 fuckssh 管理的密钥则一并删除 |
| `fuckssh export [目录]` | — | 将 `~/.ssh/config` 与私钥打包为加密 `.tar.enc` 备份 |
| `fuckssh import <备份文件>` | — | 解密并合并备份；别名冲突时可覆盖 / 跳过 / 重命名 |
| `fuckssh version` | `v` | 显示版本、提交与构建时间（Release 构建通过 ldflags 注入） |

**全局选项**

| 选项 | 说明 |
|------|------|
| `-h`, `--help` | 帮助（支持中英文） |
| `--version` | 显示版本信息 |

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

## 演示

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
  ✓ 已删除关联密钥 /home/user/.ssh/keys/id_ed25519_myserver
```

### export — 加密导出

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

### import — 导入到新机器

无冲突：

```bash
$ fuckssh import fuckssh-backup-20260616-102739.tar.enc

请输入主密码: ********
正在解密...
✓ 发现 3 个 Host 配置，无冲突
✓ 导入完成
```

有冲突：

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

### version — 查看版本

```bash
$ fuckssh v

v0.6.0 (abc1234, 2026-06-16)
```

## 设计原则

1. **标准优先** — 配置来源与落盘均为 OpenSSH 约定路径与语法；卸载后仍可用 `ssh`。
2. **可恢复** — 写入前自动备份 config（`~/.ssh/backup/`）；`import` 合并前也会备份现有 config。
3. **CLI 原生** — 面向习惯终端与 Remote SSH 的开发者，与 GUI 客户端互补而非替代。
4. **可迁移** — `export` / `import` 支持换机迁移；备份文件权限为 `0600`，若放在同步盘请自行保管。

## 路线图

| 阶段 | 能力 |
|------|------|
| **已发布** | `add`（交互 + 非交互）/ `list` / `search` / `edit` / `delete` / `export` / `import` / `version` / 短别名 |
| **后续** | `Include` 指令展开、`--json` 输出、自定义 config 路径 |

## 开发

**环境**

- Go 1.22+
- [golangci-lint](https://golangci-lint.run/)（本地 lint 与 pre-commit 需要）
- `goimports`：`go install golang.org/x/tools/cmd/goimports@latest`

```bash
git clone https://github.com/hczs/fuckssh.git
cd fuckssh
make build    # 构建到 bin/fuckssh
make test     # go test ./...
make lint     # golangci-lint
make fmt      # gofmt + goimports
make run      # go run ./cmd/fuckssh
make hooks    # 一次性启用 pre-commit（提交时自动 fmt + lint）
```

克隆仓库后建议执行一次 `make hooks`，之后每次 `git commit` 会对**已暂存**的 `.go` 文件运行 `gofmt`/`goimports` 并执行 `golangci-lint`。

推送 `v*` 标签（如 `v0.6.0`）会触发 [Release 工作流](.github/workflows/release.yml)，由 GoReleaser 构建并发布到 GitHub Releases。本地试跑：`make release-dry`。

| 触发 | 工作流 | 说明 |
|------|--------|------|
| push / PR → `main` / `master` | [CI](.github/workflows/ci.yml) | 测试、lint、多平台编译检查 |

## 文档

| 文档 | 说明 |
|------|------|
| [PRD](docs/fuckssh-PRD.md) | 产品需求与场景 |
| [技术选型](docs/fuckssh-技术选型.md) | 技术栈与模块划分 |
| [系统架构](docs/fuckssh-架构设计.md) | 架构设计（[HTML 版](docs/fuckssh-架构设计.html)） |
| [命令时序](docs/command-flow.md) | add、list、search、export、import 时序图 |
| [AGENTS.md](AGENTS.md) | 贡献者与 AI 协作约定 |

## 参与贡献

欢迎 Issue 与 Pull Request。提交前请确保 `make test` 通过；若已 `make hooks`，commit 时会自动 fmt/lint。较大改动请先开 Issue 讨论方向。

## 许可证

许可证尚未确定。在明确之前，请勿将本项目用于需要明确开源许可证的商业再分发场景；关注本仓库后续更新。
