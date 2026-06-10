# fuckssh

跨平台 CLI：围绕标准 `~/.ssh/config` 提供 VPS 一站式 SSH 配置、列表与搜索。核心目标是让初学者不用手动跑 `ssh-keygen`、上传公钥、编辑配置文件，只需一个交互式向导即可完成免密登录设置。

## 注意

- 此仓库是面向初学者的仓库，代码要干净易懂，同时要保证可维护性
- 代码要符合 go 语言最佳实践，可供初学者学习
- 代码要适当的加中文注释，为初学者讲解
- **远程主分支是 `master`，不是 `main`** — 所有链接、脚本引用都用 `master`

## 构建与测试命令

```bash
make build          # 编译到 bin/fuckssh
make test           # 运行 go test ./...（单个包测试：go test ./internal/config/）
make lint           # golangci-lint run ./...
make fmt            # gofmt -w . && goimports -w .
make run            # go run ./cmd/fuckssh
make hooks          # 启用 pre-commit hook（暂存 .go 自动 fmt + lint）
make release-dry    # 本地 GoReleaser 模拟发布
make check          # fmt + lint + test 一站式验证
```

单个测试文件：`go test ./internal/config/ -run TestParseFile`

CI 使用 `FUCKSSH_LANG=en` 跳过交互式语言选择。

## 代码架构

```
cmd/fuckssh/main.go          入口，调用 internal/cmd.Execute()
internal/
  cmd/                        Cobra 命令：root, add, list, search, version
  config/                     SSH config 解析、写入、备份、恢复、搜索、修剪
  i18n/                       国际化（中文 + 英文消息目录）
  keys/                       SSH 密钥生成、加载、写入、拷贝、删除、命名
  platform/                   跨平台路径解析、文件权限、设置、identity 文件
  sshclient/                  SSH 客户端操作：密钥认证、公钥部署、主机检测
  wizard/                     Bubble Tea TUI 向导：表单、字段、校验、主题
testdata/config/              测试用 SSH config fixture 文件
docs/                         PRD、技术选型、架构设计、开发计划
scripts/                      一键安装脚本（install.sh / install.ps1）
```

## 核心设计原则

- **标准优先**：直接读写 `~/.ssh/config`，不引入专有格式，卸载后 SSH 配置依然可用
- **写前备份**：每次修改配置前自动备份（`internal/config/backup.go`），支持恢复
- **两种添加模式**：Password 模式（生成密钥 + 部署公钥 + 写配置）和 Key 模式（已有私钥，只写配置）
- **跨平台路径**：`internal/platform/paths.go` 处理不同系统的 `~/.ssh/config` 路径解析
- **文件锁**：使用 `gofrs/flock` 防止并发写入损坏配置

## 关键依赖

| 库 | 用途 |
|---|---|
| `spf13/cobra` | CLI 框架 |
| `charmbracelet/bubbletea` + `huh` | TUI 交互向导 |
| `golang.org/x/crypto` | SSH 密钥生成与加密操作 |
| `gofrs/flock` | 文件锁 |

## 测试模式

- **依赖注入**：`internal/cmd` 中 `checkSSHFn`、`runWizardFn` 等变量可在测试中替换
- **i18n 测试钩子**：`i18n.ResetForTest()`、`i18n.SetSettingsPathForTest()`、`i18n.SetInteractiveOverrideForTest()`
- **wizard 重试**：`ErrWizardRetryForm` 触发重新填写表单（用户在确认页选择"返回修改"）

## 退出码约定

| 码 | 含义 |
|---|---|
| 0 | 成功 |
| 1 | 输入错误 / 别名已存在 |
| 2 | config 解析失败 |
| 3 | 文件系统错误 |
| 4 | 公钥部署失败 |
| 5 | ssh 命令未找到 |

## 文档

- [PRD](docs/fuckssh-PRD.md) — fuckssh 产品需求文档
- [技术选型方案](docs/fuckssh-技术选型.md) — fuckssh 技术选型与架构设计
- [系统架构设计](docs/fuckssh-架构设计.md) — fuckssh 系统架构设计（含 [HTML 版](docs/fuckssh-架构设计.html)）
- [脚手架计划](docs/plans/fuckssh-scaffold-plan.md) — fuckssh 脚手架搭建记录
- [命令执行时序图](docs/command-flow.md) — list / search / add 三条命令的 Mermaid 时序图，含 add 的三阶段流程与失败回滚
<!-- TRELLIS:START -->
# Trellis Instructions

These instructions are for AI assistants working in this project.

This project is managed by Trellis. The working knowledge you need lives under `.trellis/`:

- `.trellis/workflow.md` — development phases, when to create tasks, skill routing
- `.trellis/spec/` — package- and layer-scoped coding guidelines (read before writing code in a given layer)
- `.trellis/workspace/` — per-developer journals and session traces
- `.trellis/tasks/` — active and archived tasks (PRDs, research, jsonl context)

If a Trellis command is available on your platform (e.g. `/trellis:finish-work`, `/trellis:continue`), prefer it over manual steps. Not every platform exposes every command.

If you're using Codex or another agent-capable tool, additional project-scoped helpers may live in:
- `.agents/skills/` — reusable Trellis skills
- `.codex/agents/` — optional custom subagents

Managed by Trellis. Edits outside this block are preserved; edits inside may be overwritten by a future `trellis update`.

<!-- TRELLIS:END -->
