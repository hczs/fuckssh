# 新增 edit 命令：支持编辑已有的 SSH Host 条目

## Goal

为 fuckssh 新增 `edit` 命令，允许用户通过 TUI 向导交互式编辑已有的 SSH Host 条目。向导预填现有值，用户修改后确认即可更新配置文件。

## Requirements

### 核心功能

- `fuckssh edit <alias>` 打开 TUI 向导，预填现有 Host 条目的值
- 可编辑字段：Alias（Host 行第一个词）、HostName、User、Port、IdentityFile、Remark
- 别名可修改，需检测新别名是否与已有条目冲突
- 采用行级编辑策略：只替换被修改字段对应的行，保留文件中的未知配置项（如 ProxyJump、ForwardAgent 等）
- 用户确认修改后、写文件前自动备份（复用现有 backup 机制），避免用户放弃编辑时产生无意义备份
- 修改 Remark 时：有旧注释则替换，无旧注释则在 Host 行上方插入

### 交互模式

- 仅交互模式，不支持非交互 flag
- 复用现有 wizard 组件体系（baseField、aliasField、hostField 等）
- 向导流程：预填表单 → 用户修改 → 确认 → 执行更新 → 输出成功摘要

### 边界情况

- 别名不存在时返回 `config: host alias not found` 错误
- 别名改为已存在的别名时提示冲突
- 只修改了 Remark（其他字段未变）时也需要正确处理
- Port 留空时默认为 22

## Acceptance Criteria

- [ ] `fuckssh edit myserver` 打开预填表单，显示当前值
- [ ] 修改 HostName 后确认，config 文件中对应行被更新
- [ ] 修改别名后确认，旧 Host 块被替换为新别名
- [ ] 别名改为已存在的别名时，提示冲突并阻止提交
- [ ] 修改后文件中的未知配置项（ProxyJump 等）被保留
- [ ] 修改 Remark 后，注释行被正确更新/插入
- [ ] 编辑前自动创建备份
- [ ] 单元测试覆盖：行级编辑、别名冲突、Remark 处理

## Definition of Done

- 测试通过（`make test`）
- Lint 通过（`make lint`）
- 新增 i18n 消息（中英文）
- 新增单元测试

## Technical Approach

### 新增文件

- `internal/cmd/edit.go` — edit 命令定义、runEdit 入口
- `internal/config/edit.go` — 行级编辑核心逻辑（ReplaceHostLines）
- `internal/config/edit_test.go` — 行级编辑测试
- `internal/wizard/edit_wizard.go` — 编辑向导（预填表单）

### 行级编辑策略

`ReplaceHostLines(path, alias, newEntry)` 的工作方式：

1. 读取文件所有行
2. 找到 `Host <alias>` 所在行（LineStart）
3. 从 Host 行开始，向下扫描到下一个 Host 行或文件末尾，确定块范围
4. 在块范围内，按行匹配已知指令（HostName、User、Port、IdentityFile），替换为新值
5. 处理 Remark：找到 Host 行上方的连续 # 注释行，替换或插入
6. 如果别名被修改，替换 Host 行本身
7. 写回文件

### 编辑向导设计

复用现有 field 组件，创建 `EditWizard`：

- 预填模式：初始化时将现有 HostEntry 的值填入各字段
- Alias 字段：使用 aliasField，但需要特殊处理（允许与自身别名相同）
- 其他字段：使用现有的 hostField、userField 等
- 确认页：显示变更对比（哪些字段被修改了）

### 与 delete 命令的一致性

- 命令签名：`edit <alias>`（与 delete 一致）
- 错误处理：别名不存在时返回 ErrHostNotFound
- 输出风格：成功后输出摘要信息

## Out of Scope

- 非交互模式（flag 参数）—— 后续版本考虑
- 批量编辑多个 Host
- 编辑后自动重新部署公钥（密码模式）
- 支持 SSH config 的 Include 指令展开
- 编辑未知配置项（只保留，不编辑）

## Technical Notes

- `internal/config/types.go` 定义了 HostEntry 结构
- `internal/config/write.go` 有 formatHostBlock 和 AppendHost，但 edit 需要行级编辑而非整块重写
- `internal/config/parse.go` 的 applyOption 只处理 hostname/user/port/identityfile
- `internal/cmd/delete.go` 是 edit 命令的参考模板（命令结构、错误处理、确认流程）
- `internal/wizard/` 下有完整的 TUI 组件体系，edit 向导应复用这些组件
- `internal/config/backup.go` 提供备份功能
