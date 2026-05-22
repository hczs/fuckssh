# Phase 5: 密码连接向导（全流程）

## 当前进度

- [x] Task 1: sshclient 包 — 密码登录并部署 authorized_keys
- [ ] Task 2: wizard 包 — 密码模式表单与编排
- [ ] Task 3: add 子命令 — 密码模式端到端（密钥生成 + 部署 + config）

> 从第一个 `[ ]` 开始执行。完成后将 `[ ]` 改为 `[x]`。

## 里程碑目标

用户对新购 VPS（密码可登录）运行 `fuckssh add` 密码模式：生成 Ed25519 → 写密钥与 config → SSH 密码登录部署公钥 → 提示 `ssh <别名>` 免密成功。满足 PRD 核心指标「< 3 分钟」。

## 验收标准

- [ ] 选择密码连接，必填 IP、用户、密码；端口/别名/算法可回车默认
- [ ] 自动生成密钥并写入 `~/.ssh/`，config 含 `IdentityFile`
- [ ] 远端 `~/.ssh/authorized_keys` 追加新公钥（不破坏已有行）
- [ ] 部署成功后，本机 `ssh <别名>` 免密登录（需系统 ssh 已安装）
- [ ] 部署失败时：stderr 说明原因 + 已备份 config 路径；不自动删除已写密钥
- [ ] 密码不出现在 config、日志、备份文件中
- [ ] 可选：网络失败重试 1～2 次

## 完成条件

- 本阶段所有验收标准全部通过
- **场景 A（PRD 主路径）** 可在公司 Win 或家里 Mac 走通

## 前置依赖

- Phase 3（keys 生成）
- Phase 4（config 备份/写入、wizard 框架、add 命令）
- 测试用 VPS：允许密码登录的 root/ubuntu

## 任务拆解（TDD 流程）

### Task 1: sshclient 包 — 密码登录并部署 authorized_keys

**目标**：`DeployPublicKey(ctx, DeployOpts) error` 用密码认证 SSH，创建 `~/.ssh`（如需）、追加公钥。

**测试用例**：
- `test_DeployPublicKey_buildsAuthorizedKeysLine`（纯函数）
- `test_DeployPublicKey_authFailed_returnsTypedError`
- 集成：`//go:build integration` + 文档说明需真机（MVP 可手动）

**实现要点**：
- `internal/sshclient/deploy.go`
- `golang.org/x/crypto/ssh` 客户端；`ssh.InsecureIgnoreHostKey()` 仅 MVP（文档标注待加强 host key 校验）
- 定义 `ErrDeployFailed`，退出码映射 4

**验收**：对测试 VPS 手动部署成功；或 integration 测试在 CI skip。

### Task 2: wizard 包 — 密码模式表单与编排

**目标**：`RunPasswordMode` 收集密码（内存）、调用 keys → config → deploy。

**测试用例**：
- `test_PasswordMode_validateRequiresPassword`
- `test_PasswordMode_defaultAlgorithmEd25519`
- `test_PasswordMode_order_backupBeforeWrite`

**实现要点**：
- `internal/wizard/password_mode.go`
- 编排顺序：检测 ssh（警告）→ 备份 → 生成密钥 → AppendHost → DeployPublicKey
- 函数返回前对密码引用 best-effort 清零

**验收**：表驱动验证调用顺序（可用 mock 接口）。

### Task 3: add 子命令 — 密码模式端到端

**目标**：`fuckssh add` 密码模式完整可用；与密钥模式共存。

**测试用例**：
- 扩展 `test_Add_integration` 覆盖密码模式 mock deploy

**实现要点**：
- 完善 `internal/wizard/run.go` 模式分支
- 重复别名检测：询问覆盖或中止
- 成功输出：`现在可以执行: ssh <别名>`

**验收**：对新 VPS 从启动到 `ssh <别名>` 免密 < 3 分钟（含用户输入）。

## 涉及文件

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新建 | `internal/sshclient/deploy.go` | 远端部署 |
| 新建 | `internal/wizard/password_mode.go` | 密码模式 |
| 修改 | `internal/wizard/run.go` | 完整编排 |
| 修改 | `internal/cmd/add.go` | 错误码映射 |
| 新建 | `internal/sshclient/deploy_test.go` | 测试 |

## 阶段完成后的系统状态

**MVP 核心闭环完成**：新 VPS 密码路径 + 已有密钥路径 + 列表 + 搜索。产品可自用并对外发布最小版本。

## 已知限制

- 不支持：仅密钥登录且禁用密码、需合并他人 authorized_keys 的复杂场景
- Host key 校验 MVP 从简（待确认加强）
- 不自动回滚已写密钥文件（仅提示备份路径）
- V2 备份/恢复、编辑删除不做
