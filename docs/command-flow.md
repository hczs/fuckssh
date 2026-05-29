# 命令执行时序图

三条核心命令的执行流程，用 Mermaid 时序图描述。

---

## 1. `fuckssh list` — 列出所有 Host

```mermaid
sequenceDiagram
    participant User as 用户
    participant Cmd as cmd/list.go
    participant Cfg as config.ParseFile
    participant Report as WriteHostsReport

    User->>Cmd: fuckssh list
    Cmd->>Cmd: ConfigFilePath()
    Cmd->>Cfg: ParseFile(configPath)
    Cfg-->>Cmd: []HostEntry
    Cmd->>Report: WriteHostsReport(entries, query="")
    Report-->>User: 展示 Host 列表
```

**特点**：只读操作，不修改任何文件。

---

## 2. `fuckssh search <query>` — 搜索 Host

```mermaid
sequenceDiagram
    participant User as 用户
    participant Cmd as cmd/search.go
    participant Cfg as config.ParseFile
    participant Filter as config.FilterHosts
    participant Report as WriteHostsReport

    User->>Cmd: fuckssh search "prod"
    Cmd->>Cmd: ConfigFilePath()
    Cmd->>Cfg: ParseFile(configPath)
    Cfg-->>Cmd: []HostEntry
    Cmd->>Filter: FilterHosts(entries, "prod")
    Filter-->>Cmd: []HostEntry (匹配项)
    Cmd->>Report: WriteHostsReport(matched, query="prod")
    Report-->>User: 展示搜索结果
```

**特点**：只读操作，在 list 基础上增加过滤。

---

## 3. `fuckssh add` — 添加 VPS Host

`add` 分为三个阶段：

### 阶段一：填写表单

```mermaid
sequenceDiagram
    participant User as 用户
    participant Run as wizard/run.go
    participant Form as collectAddInput

    User->>Run: fuckssh add
    Run->>Form: collectAddInput(configPath, draft)
    Form->>Form: TUI 表单（主机/别名/用户/端口/认证/备注）
    Form-->>User: 交互式输入
    User-->>Form: 填写完成
    Form-->>Run: AddInput

    Note over Run: 此阶段 config 文件未被读写，未备份
```

### 阶段二：确认执行

```mermaid
sequenceDiagram
    participant Run as wizard/run.go
    participant Confirm as wizard_confirm.go
    participant User as 用户

    alt 密码模式
        Run->>Confirm: confirmPasswordRun(input, configPath)
    else 密钥模式
        Run->>Confirm: confirmKeyRun(input, configPath)
    end
    Confirm->>Confirm: 构建操作摘要
    Confirm-->>User: 展示摘要 + 确认按钮

    alt 用户选「确认执行」
        User-->>Confirm: 确认
        Confirm-->>Run: nil (继续)
    else 用户选「返回修改」
        User-->>Confirm: 返回
        Confirm-->>Run: ErrWizardRetryForm
        Run->>Run: 重新进入阶段一（带 draft）
    end

    Note over Run: 此阶段 config 文件未被读写，未备份
```

### 阶段三：执行写盘

#### 密钥模式（Key Mode）

```mermaid
sequenceDiagram
    participant Cmd as cmd/add.go
    participant Flow as key_flow.go
    participant Backup as config.Backup
    participant Stage as stageKeyForConfig
    participant Append as config.AppendHost
    participant FS as ~/.ssh/config

    Cmd->>Flow: RunKeyFlow(configPath, result)
    Flow->>Backup: backup(configPath)
    Backup->>FS: 读取并复制到 .bak 时间戳文件
    Backup-->>Flow: bakPath

    Flow->>Stage: stageKey(alias, identityFile)
    Stage->>Stage: 复制密钥到 ~/.ssh/keys/
    Stage-->>Flow: destPriv

    Flow->>Append: appendHost(configPath, entry)
    Append->>FS: 追加 Host 配置块
    Append-->>Flow: nil

    Flow-->>Cmd: WizardResult (含 BackupPath)

    Note over Flow: 任一步失败 → RollbackAfterAddFailure + RemoveKeyPair
```

#### 密码模式（Password Mode）

```mermaid
sequenceDiagram
    participant Run as wizard/run.go
    participant Setup as setupPasswordFlow
    participant Deploy as deployPublicKeyWithRetry
    participant Backup as config.Backup
    participant KeyGen as writeKeys
    participant Append as config.AppendHost
    participant SSH as ssh 远程主机
    participant FS as ~/.ssh/config

    Run->>Setup: setupPasswordFlow(ctx, input, configPath)
    Setup->>Backup: backup(configPath)
    Backup->>FS: 读取并复制到 .bak 时间戳文件
    Backup-->>Setup: bakPath

    Setup->>KeyGen: writeKeys(sshDir, alias)
    KeyGen-->>Setup: privPath, pubLine

    Setup->>Append: appendHost(configPath, entry)
    Append->>FS: 追加 Host 配置块
    Append-->>Setup: nil

    Setup-->>Run: setupState

    Run->>Deploy: deployPublicKeyWithRetry(ctx, input, pubLine)
    Deploy->>SSH: ssh-copy-id (密码认证 + 上传公钥)
    SSH-->>Deploy: 成功/失败

    alt 部署失败
        Deploy-->>Run: error
        Run->>Run: rollbackPasswordChanges(configPath, state)
        Note over Run: 恢复 config 备份 + 删除生成的密钥
    else 部署成功
        Deploy-->>Run: nil
        Run-->>Run: 返回 WizardResult
    end
```

---

## 设计要点

| 要点 | 说明 |
|---|---|
| **阶段隔离** | 表单收集、确认、执行三个阶段严格分离；前两个阶段是纯内存操作 |
| **写前备份** | config 文件只在阶段三被修改，备份也在阶段三才发生 |
| **失败回滚** | 任一步失败都回滚：恢复 config 备份 + 删除已生成的密钥 |
| **依赖注入** | backup/stageKey/appendHost 等均可注入，便于单测验证调用顺序 |
