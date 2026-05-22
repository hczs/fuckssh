# Phase 3: Ed25519 密钥生成与落盘

## 当前进度

- [x] Task 1: keys 包 — 生成密钥对（OpenSSH 格式）
- [x] Task 2: keys 包 — 写入文件与权限（platform 协作）
- [x] Task 3: 密钥文件命名规则与冲突处理

> 从第一个 `[ ]` 开始执行。完成后将 `[ ]` 改为 `[x]`。

## 里程碑目标

工具能用 Go 生成默认 Ed25519 密钥对，以 OpenSSH PEM / `.pub` 格式写入 `~/.ssh/`，权限符合安全要求，可供后续 config 引用与密码模式部署。

## 验收标准

- [x] 生成后私钥为 `OPENSSH PRIVATE KEY` 格式，`ssh-keygen -l -f <私钥>` 能读取
- [x] 公钥单行格式可用于 `authorized_keys`
- [x] Unix 上私钥权限为 `0600`
- [x] 目标路径已存在时返回明确错误（不静默覆盖）
- [x] `go test ./internal/keys/...` 全部通过

## 完成条件

- 本阶段所有验收标准全部通过
- 所有 Task 的 checkbox 已全部 `[x]`
- 在临时目录手动生成一次并验证 `ssh-keygen -l`

## 前置依赖

- Phase 2 完成（`platform` 路径与权限封装可复用）

## 任务拆解（TDD 流程）

### Task 1: keys 包 — 生成密钥对（OpenSSH 格式）

**目标**：`GenerateEd25519()` 返回私钥 PEM 字节与 `authorized_keys` 行字符串。

**测试用例**：
- `test_GenerateEd25519_producesParseablePrivateKey`
- `test_GenerateEd25519_publicLineStartsWithSshEd25519`
- `test_GenerateEd25519_eachCallUnique`（随机性）

**实现要点**：
- `internal/keys/generate.go`
- 使用 `crypto/ed25519` + `ssh.MarshalPrivateKey` / `ssh.MarshalAuthorizedKey`
- **禁止** `x509.MarshalPKCS8PrivateKey` 处理 Ed25519

**验收**：测试内用 `ssh.ParsePrivateKey` 解析成功。

### Task 2: keys 包 — 写入文件与权限

**目标**：`WriteKeyPair(dir, baseName, keyPair)` 写入 `baseName` 与 `baseName.pub`。

**测试用例**：
- `test_WriteKeyPair_createsTwoFiles`
- `test_WriteKeyPair_unixMode0600`
- `test_WriteKeyPair_refusesExistingFile`

**实现要点**：
- `internal/keys/write.go`
- `platform.SetPrivateKeyPerm(path)` 封装 Win/Unix 差异

**验收**：`t.TempDir()` 内写入后读权限与内容正确。

### Task 3: 密钥文件命名规则与冲突处理

**目标**：约定命名：`id_ed25519_fuckssh_<alias>`（alias 经 sanitize）；导出 `KeyPaths(alias) (priv, pub string)`。

**测试用例**：
- `test_KeyPaths_sanitizesInvalidChars`
- `test_KeyPaths_defaultWhenAliasEmpty`（如基于 hostname 哈希，实现时定规则并测）

**实现要点**：
- `internal/keys/naming.go`
- 与架构 §2.2.4 对齐，写入注释说明规则

**验收**：给定别名 `my-vps` 得到可预期文件名且无路径穿越字符。

## 涉及文件

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新建 | `internal/keys/generate.go` | 生成 |
| 新建 | `internal/keys/write.go` | 落盘 |
| 新建 | `internal/keys/naming.go` | 命名 |
| 新建 | `internal/keys/*_test.go` | 测试 |
| 修改 | `internal/platform/` | 文件权限辅助 |
| 修改 | `go.mod` | 添加 `golang.org/x/crypto` |
| 删除占位 | `internal/keys/doc.go` | 替换实现 |

## 阶段完成后的系统状态

密钥生成能力就绪，尚未接入向导；Phase 4/5 的 `add` 将直接调用本包。用户可通过测试或后续 `add` 间接验收。

## 已知限制

- MVP 仅默认 Ed25519；RSA 可选留待确认
- 不为用户已有私钥做迁移整理（PRD 待确认项）
- 不执行远端部署
