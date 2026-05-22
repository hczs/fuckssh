# Phase 1: 列出与搜索 VPS

## 当前进度

- [x] Task 1: platform 包 — SSH 目录路径解析
- [x] Task 2: config 包 — 受限解析器与 HostEntry 模型
- [x] Task 3: list 子命令 — 表格输出
- [x] Task 4: search 子命令 — 模糊匹配

> 从第一个 `[ ]` 开始执行。完成后将 `[ ]` 改为 `[x]`。

## 里程碑目标

用户在本机执行 `fuckssh list` 与 `fuckssh search <词>`，能从真实 `~/.ssh/config` 看到结构化 Host 列表并按关键词筛选。

## 验收标准

- [x] 在含至少 2 个 `Host` 块的 config 上运行 `fuckssh list`，stdout 显示别名、HostName、Port、User 列
- [x] `fuckssh search` 某台机器的 IP 片段，只输出匹配行
- [x] `fuckssh search` 某别名子串，大小写不敏感也能命中
- [x] config 含语法错误时，stderr 有文件名、行号、问题行片段，退出码非 0
- [x] 空 config 或无可解析 Host 时，友好提示而非 panic
- [x] `go test ./internal/config/...` 全部通过

## 完成条件

- 本阶段所有验收标准全部通过
- 所有 Task 的 checkbox 已全部 `[x]`
- 核心流程能手动走通一遍

## 前置依赖

- 脚手架已就绪（`go build`、`go test ./...` 已通过）

## 任务拆解（TDD 流程）

### Task 1: platform 包 — SSH 目录路径解析

**目标**：跨平台解析 `~/.ssh` 与默认 `config` 路径（Windows `USERPROFILE`，Unix `HOME`）。

**测试用例**：
- `test_SSHDir_windows_usesUserProfile`
- `test_SSHDir_unix_usesHome`
- `test_DefaultConfigPath_joinsConfig`

**实现要点**：
- 新建 `internal/platform/paths.go`
- 提供 `SSHDir()`, `DefaultConfigPath()`, `ExpandPath(string)`
- 支持 `--config` flag 覆盖（在 `internal/cmd/root.go` 挂 PersistentFlag）

**验收**：单元测试在 Windows/macOS 逻辑分支下通过（可用 `runtime.GOOS` 表驱动）。

### Task 2: config 包 — 受限解析器与 HostEntry 模型

**目标**：将 config 解析为 `[]HostEntry`（Alias、HostName、User、Port、IdentityFile、LineStart）。

**测试用例**：
- `test_Parse_singleHost_minimal`
- `test_Parse_multipleHosts`
- `test_Parse_defaultPort22_whenOmitted`
- `test_Parse_invalidLine_returnsErrorWithLineNumber`
- `test_Parse_ignoresCommentAndBlankLines`
- `test_Parse_hostWithMultipleAliases`（待确认：取第一个或全部，实现时定一种并写测试）

**实现要点**：
- `internal/config/parse.go`、`types.go`
- `testdata/config/` 放样例：单 Host、多 Host、缺 Port、故意坏行
- MVP **不展开** `Include`；遇到 `Include` 可记录或 stderr 提示「未加载」（不崩溃）

**验收**：`go test ./internal/config/...` 绿；对 `testdata` 快照断言条数与字段。

### Task 3: list 子命令 — 表格输出

**目标**：`fuckssh list` 调用 parser，用 `text/tabwriter` 输出列对齐表格。

**测试用例**：
- `test_ListCmd_printsTableFromFixture`（可测 RunE 返回的 buffer 或抽 `formatHosts`）
- `test_ListCmd_respectsConfigFlag`

**实现要点**：
- 修改 `internal/cmd/list.go`，依赖 `config.ParseFile`
- 无 Host 时 stdout 提示「未找到 Host 条目」

**验收**：对本机或 `testdata` 拷贝的 config 运行 `go run ./cmd/fuckssh list`，列对齐可读。

### Task 4: search 子命令 — 模糊匹配

**目标**：`fuckssh search <query>` 对 Alias、HostName、IP 做大小写不敏感子串匹配。

**测试用例**：
- `test_Search_matchesAlias`
- `test_Search_matchesHostName`
- `test_Search_matchesIP`
- `test_Search_noMatch_returnsEmptyWithHint`
- `test_Search_emptyQuery_returnsUsageError`

**实现要点**：
- `internal/config/search.go` 或 `filter.go`
- 修改 `internal/cmd/search.go`

**验收**：`fuckssh search 192` 能筛出含该 IP 的 Host；无匹配时有明确提示。

## 涉及文件

| 操作 | 文件路径 | 说明 |
|------|----------|------|
| 新建 | `internal/platform/paths.go` | 路径解析 |
| 新建 | `internal/config/types.go` | HostEntry |
| 新建 | `internal/config/parse.go` | 解析器 |
| 新建 | `internal/config/search.go` | 搜索过滤 |
| 新建 | `internal/config/*_test.go` | 表驱动测试 |
| 新建 | `testdata/config/*.conf` | 夹具 |
| 修改 | `internal/cmd/list.go` | 实现 list |
| 修改 | `internal/cmd/search.go` | 实现 search |
| 修改 | `internal/cmd/root.go` | `--config` flag |
| 删除占位 | `internal/config/doc.go` | 替换为真实实现 |

## 阶段完成后的系统状态

用户已拥有一台「只读 VPS 目录」：`list` 看全量，`search` 快速定位。尚未能添加新 VPS，但可验证 config 解析是否正确，为后续写入打底。

## 已知限制（本阶段及近期都不处理的）

- 不解析/展开 `Include`、通配符 `Host *`、`Match` 块
- 不提供编辑、删除 Host
- 不检测系统 `ssh`（Phase 2）
- 不生成密钥、不连远端
