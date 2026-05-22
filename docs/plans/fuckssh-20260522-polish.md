# Phase 6: 文档、双机冒烟与发布准备

## 当前进度

- [ ] Task 1: README 与子命令用法示例
- [ ] Task 2: 解析范围与用户文档（Include/限制说明）
- [ ] Task 3: 双机冒烟清单与 Makefile 发布目标
- [ ] Task 4: GoReleaser 与模块路径（可选）

> 从第一个 `[ ]` 开始执行。完成后将 `[ ]` 改为 `[x]`。

## 里程碑目标

MVP 可在 GitHub 对外说明用法；开发者与公司 Win、家里 Mac 各完成一次主路径冒烟；CI 稳定，具备发布二进制的基础。

## 验收标准

- [ ] README 含：安装、`list`/`search`/`add` 示例、Windows OpenSSH 指引链接
- [ ] `docs/` 或 README 标明 config **解析支持范围**（与 Phase 1 实现一致）
- [ ] 公司 Windows：`add` 密码模式 + `list` + `search` 通过
- [ ] 家里 macOS：同上通过
- [ ] `make test && make lint` 本地绿；GitHub Actions 绿
- [ ] （可选）打 `v0.1.0` tag 后 GoReleaser 产出三平台二进制

## 完成条件

- 本阶段所有验收标准全部通过（Task 4 可选不计入阻塞）
- PRD §6.2 MVP 里程碑验收项可满足

## 前置依赖

- Phase 5 完成（全功能可用）

## 任务拆解（TDD 流程）

### Task 1: README 与子命令用法示例

**目标**：新用户仅凭 README 能完成首次 `add` 与 `list`。

**测试用例**：
- 无自动化；人工检查示例命令可复制运行

**实现要点**：
- 更新根目录 `README.md`
- 每子命令一段示例；常见错误（无 ssh、解析失败）链到指引

**验收**：同事/未来的你按 README 从零操作无歧义。

### Task 2: 解析范围与用户文档

**目标**：避免用户因 `Include`/通配符导致「列表不全」的困惑。

**测试用例**：
- `test_ReadmeDocumentsParseLimits`（可选：轻量检查 README 含关键词）

**实现要点**：
- `docs/ssh-config-support.md` 或在 README 一节
- 列出：支持字段、不支持语法、遇到 Include 时的行为

**验收**：文档与 `internal/config` 行为一致。

### Task 3: 双机冒烟清单与 Makefile

**目标**：可重复的发布前检查单。

**测试用例**：
- 无；维护 `docs/smoke-checklist.md`

**实现要点**：
- 清单项：密码 add、密钥 add、list、search、无 ssh 警告、备份文件存在
- `Makefile` 增加 `build-all` 或 `release-dry`（`GOOS` 交叉编译）

**验收**：Win + Mac 各勾选全部项。

### Task 4: GoReleaser 与模块路径（可选）

**目标**：`go install` 与 GitHub Releases 可分发。

**测试用例**：
- CI tag 触发 dry-run（可选）

**实现要点**：
- 将 `go.mod` 模块路径改为真实 GitHub 仓库
- 添加 `.goreleaser.yaml`、文档说明 tag 发布流程

**验收**：`v0.1.0` 产物含 win/mac/linux 二进制。

## 涉及文件

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 修改 | `README.md` | 用户文档 |
| 新建 | `docs/ssh-config-support.md` | 解析范围 |
| 新建 | `docs/smoke-checklist.md` | 冒烟清单 |
| 修改 | `Makefile` | 交叉编译 |
| 可选 | `.goreleaser.yaml` | 发布 |
| 修改 | `go.mod` | 模块路径 |

## 阶段完成后的系统状态

fuckssh **MVP 可发布**：文档齐全、双机验证通过、CI 可靠。后续可启动 V2（加密备份/恢复）计划。

## 已知限制

- V2 加密备份、编辑删除、JumpHost 仍不在范围
- 自动化集成测试（真机 SSH）非 MVP 必需
- VS Code Remote 偶发密码问题不专项排查
