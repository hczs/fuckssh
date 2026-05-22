# Phase 4: 密钥连接向导（写 config）

## 当前进度

- [ ] Task 1: config 包 — 修改前备份与追加 Host 块
- [ ] Task 2: wizard 包 — huh 表单与密钥模式流程
- [ ] Task 3: add 子命令 — 串联向导与成功提示

> 从第一个 `[ ]` 开始执行。完成后将 `[ ]` 改为 `[x]`。

## 里程碑目标

用户运行 `fuckssh add`，选择「密钥连接」，提供已有私钥路径与连接信息，工具备份 config 后写入新 `Host` 块，并提示 `ssh <别名>`（前提：远端已接受该密钥）。

## 验收标准

- [ ] 交互式选择密钥连接模式，输入 IP、用户、私钥路径、端口（回车默认 22）、别名（回车自动生成）
- [ ] 执行前生成 `~/.ssh/config.fuckssh.bak.<timestamp>`
- [ ] config 末尾出现正确 `Host` / `HostName` / `User` / `Port` / `IdentityFile`
- [ ] 私钥路径不存在时，向导内报错且不写 config
- [ ] 同名 `Host` 已存在时，提示覆盖或中止（实现一种并文档化）
- [ ] 完成后 stdout 提示：`ssh <别名>` 进行连接
- [ ] 在已能用该钥登录的 VPS 上，手动 `ssh <别名>` 成功

## 完成条件

- 本阶段所有验收标准全部通过
- 场景 B（PRD）可完整走通

## 前置依赖

- Phase 1（config 解析/路径）
- Phase 3（命名规则可参考；本模式不写新密钥）
- Phase 2（add 前置 ssh 检测，可选警告）
- 引入依赖：`github.com/charmbracelet/huh`

## 任务拆解（TDD 流程）

### Task 1: config 包 — 修改前备份与追加 Host 块

**目标**：`Backup(path) (bakPath, error)`；`AppendHost(path, HostEntry) error`。

**测试用例**：
- `test_Backup_createsTimestampedCopy`
- `test_AppendHost_appendsBlockToFile`
- `test_AppendHost_preservesExistingContent`
- `test_AppendHost_formatsIdentityFileWithQuotesWhenNeeded`

**实现要点**：
- `internal/config/backup.go`、`write.go`
- Host 块模板：标准 OpenSSH 指令顺序
- 写入策略：**追加到文件末尾**（与架构一致，简单可测）

**验收**：对 `t.TempDir()` 内样例 config 追加后，再 `Parse` 能读出新 Host。

### Task 2: wizard 包 — huh 表单与密钥模式流程

**目标**：`RunKeyMode(ctx, deps) (*WizardResult, error)` 收集字段并校验。

**测试用例**：
- `test_KeyMode_validateRequiresHostUserIdentity`
- `test_KeyMode_defaultPort22`
- `test_KeyMode_generatesAliasWhenEmpty`
- `test_KeyMode_rejectsMissingPrivateKey`（文件不存在）

**实现要点**：
- `internal/wizard/key_mode.go`
- 模式选择：密钥 / 密码（密码模式 Phase 5 占位 stub）
- 密码不落盘、不进日志

**验收**：可用 huh 的测试模式或抽纯函数 `collectKeyModeInput` 单测。

### Task 3: add 子命令 — 串联向导与成功提示

**目标**：`fuckssh add` → ssh 检测 → 向导 → 备份 → AppendHost → 打印成功文案。

**测试用例**：
- `test_Add_keyMode_integrationWithTempConfig`（表驱动 + 临时文件）

**实现要点**：
- `internal/cmd/add.go` 调用 `wizard.Run`
- `internal/wizard/run.go` 编排
- 退出码：输入无效 1，IO 失败 3（对齐架构 §4.4）

**验收**：对测试 VPS 或本机已有密钥场景完整跑通 `fuckssh add`。

## 涉及文件

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新建 | `internal/config/backup.go` | 备份 |
| 新建 | `internal/config/write.go` | 追加 Host |
| 新建 | `internal/wizard/key_mode.go` | 密钥模式 |
| 新建 | `internal/wizard/run.go` | 编排入口 |
| 修改 | `internal/cmd/add.go` | 实现 add |
| 修改 | `go.mod` | 添加 huh |
| 删除占位 | `internal/wizard/doc.go` | 替换实现 |

## 阶段完成后的系统状态

**场景 B 可用**：已有私钥、仅缺 config 的用户可在 3 分钟内完成本地配置。密码连接、公钥部署留 Phase 5。

## 已知限制

- 不生成新密钥、不连远端部署公钥
- 不把用户私钥复制到 `~/.ssh`（仅引用路径；整理密钥目录为待确认）
- 密码模式向导未实现
