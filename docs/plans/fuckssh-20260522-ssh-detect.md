# Phase 2: 系统 ssh 检测与安装指引

## 当前进度

- [x] Task 1: platform 包 — 分平台安装指引常量
- [x] Task 2: sshclient 包 — LookPath 检测与结构化错误
- [x] Task 3: 集成到 add 前置检查（list/search 可选警告）

> 从第一个 `[ ]` 开始执行。完成后将 `[ ]` 改为 `[x]`。

## 里程碑目标

工具能在启动关键流程前检测 `ssh` 是否在 PATH 中；缺失时向 stderr 输出 Windows/macOS/Linux 各自的安装指引，不崩溃。

## 验收标准

- [x] PATH 中有 `ssh` 时，检测通过且无警告（或仅 verbose 提示路径）
- [x] 模拟无 `ssh`（改 PATH 或 mock）时，stderr 出现当前 OS 对应的安装说明（Windows 需提到「可选功能」）
- [x] `fuckssh add` 在检测失败时：给出指引后**仍可继续**向导（密码部署不依赖系统 ssh；见架构 §9.2）
- [x] `go test ./internal/sshclient/...` 全部通过

## 完成条件

- 本阶段所有验收标准全部通过
- 所有 Task 的 checkbox 已全部 `[x]`

## 前置依赖

- Phase 1 完成（`platform` 包已有路径能力）

## 任务拆解（TDD 流程）

### Task 1: platform 包 — 分平台安装指引常量

**目标**：集中维护各平台 OpenSSH 客户端安装文案（中文、可复制步骤）。

**测试用例**：
- `test_InstallGuide_windows_containsOptionalFeature`
- `test_InstallGuide_darwin_containsBuiltin`
- `test_InstallGuide_linux_mentionsOpensshClient`

**实现要点**：
- `internal/platform/ssh_guide.go`
- `InstallOpenSSHGuide() string` 按 `runtime.GOOS` 返回

**验收**：表驱动断言各 OS 文案含关键词。

### Task 2: sshclient 包 — LookPath 检测与结构化错误

**目标**：`CheckSSH()` 返回 `(path string, err error)`；定义 `ErrSSHNotFound` 供上层判断。

**测试用例**：
- `test_CheckSSH_found`（可注入 `exec.LookPath` 或仅测 err 类型分支）
- `test_CheckSSH_notFound_wrapsGuide`

**实现要点**：
- `internal/sshclient/detect.go`
- 错误信息拼接 `platform.InstallOpenSSHGuide()`

**验收**：无 ssh 时 `errors.Is(err, ErrSSHNotFound)` 为真。

### Task 3: 集成到 add 前置检查

**目标**：`fuckssh add` 开始时调用检测；失败打印指引，不阻止继续（与 PRD/架构一致）。

**测试用例**：
- `test_AddCmd_warnsWhenSSHMissing`（mock sshclient）

**实现要点**：
- 修改 `internal/cmd/add.go`，调用 `sshclient.CheckSSH`
- 将具体向导逻辑留 Phase 4/5，本 Task 仅接线检测

**验收**：本机有 ssh 时 add 无多余噪音；无 ssh 时可见警告文案。

## 涉及文件

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新建 | `internal/platform/ssh_guide.go` | 安装指引 |
| 新建 | `internal/sshclient/detect.go` | 检测逻辑 |
| 新建 | `internal/sshclient/detect_test.go` | 测试 |
| 修改 | `internal/cmd/add.go` | 前置检测 |
| 删除占位 | `internal/sshclient/doc.go` | 替换实现 |

## 阶段完成后的系统状态

用户在缺少 OpenSSH 客户端的环境（尤其 Windows）能得到明确下一步，而不是神秘的连接失败。`list`/`search` 仍不依赖 ssh。

## 已知限制

- 不自动安装 OpenSSH（仅指引）
- 不包装 `ssh` 命令执行日常连接
- 远端公钥部署在 Phase 5
