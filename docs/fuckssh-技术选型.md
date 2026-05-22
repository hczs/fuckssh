# fuckssh 技术选型方案

## 文档信息

- 版本：v1.0
- 日期：2026-05-22
- 项目周期：无硬性 deadline，边学 Go 边迭代
- 团队规模：1 人（开发者兼维护）
- 预算范围：$0 托管（本机 CLI，无服务端）

## 1. 约束条件总结

### 1.1 团队约束

| 项 | 内容 |
|----|------|
| 技术储备 | Java 为主；**有意用本项目学习 Go** |
| 团队规模 | 单人开发 |
| 技术水平 | 后端经验丰富，Go 为学习阶段 |
| 开发与自测环境 | **公司 Windows、家里 macOS**（与 PRD 双机场景一致） |

### 1.2 项目约束

| 项 | 内容 |
|----|------|
| 产品形态 | 跨平台 CLI，仅操作标准 `~/.ssh/config` 与密钥文件 |
| 项目周期 | **无硬性上线时间**，以 MVP 可自用、可发布为里程碑 |
| 迭代计划 | 持续迭代；MVP → V2 加密备份/恢复 |

### 1.3 资源约束

| 项 | 内容 |
|----|------|
| 预算 | 无云服务开销；可选 GitHub Actions 免费额度做 CI |
| 部署环境 | 用户本机；分发靠 GitHub Releases / `go install` |
| 数据托管 | 不托管用户数据；V2 依赖用户自备同步盘 |

### 1.4 业务约束

| 项 | 内容 |
|----|------|
| 用户量级 | 个人工具，单机使用 |
| 并发 / QPS | 无 |
| 数据量级 | 每人约 3～60 条 Host；config 解析与搜索需 **< 1 秒** |
| 性能目标 | 新 VPS 向导全流程 **< 3 分钟**（含用户输入与网络） |

### 1.5 已确认的产品技术决策（访谈结论）

1. **认可「检测 `ssh` + 友好提示」**，不假设系统一定已安装 OpenSSH 客户端。
2. **MVP 用 Go 生成密钥**，降低对 `ssh-keygen` 的依赖；日常连接仍走系统 `ssh`，与 VS Code / Tabby 互通。
3. **主语言选 Go**，兼顾 PRD 与「以项目学 Go」的目标。

---

## 2. 技术选型决策

### 2.1 编程语言：Go

**推荐方案：Go 1.22+（建议跟随当前 stable）**

| 维度 | 评估 |
|------|------|
| 学习曲线 | 中（有 Java 基础，语法与错误处理需适应） |
| 开发效率 | 高（CLI、文件 IO、交叉编译成熟） |
| 社区生态 | 活跃（CLI、SSH、DevOps 工具链丰富） |
| 性能表现 | 优（本机工具绰绰有余） |
| 适合场景 | 跨平台单二进制 CLI、无服务端 |

**为什么选 Go 而不是 Rust / Python / Java？**

| 对比 | 说明 |
|------|------|
| **vs Rust** | Rust 更适合极致性能与安全，但学习曲线陡，单人无 deadline 下会拉长 MVP；与「学 Go」目标不一致。 |
| **vs Python** | 开发快，但 Windows 分发、单文件发布、类型与打包体验不如 Go；与标准 `ssh` 生态对齐需额外胶水。 |
| **vs Java** | 你熟悉 Java，但 CLI 分发需 JRE 或 GraalVM 原生编译，体积与启动、路径处理不如 Go 常见做法直接。 |
| **Go 优势** | 交叉编译 Win/macOS/Linux；与 kubectl、gh 等 CLI 同属一类技术栈；`golang.org/x/crypto/ssh` 可生成 OpenSSH 格式密钥并做密码模式部署。 |

**参考资料：**

- [golang.org/x/crypto/ssh — OpenSSH 私钥序列化](https://github.com/golang/crypto/blob/master/ssh/keys.go)（`MarshalPrivateKey` / `MarshalAuthorizedKey`）
- [Go issue #37132 — OpenSSH 格式 marshal 支持](https://github.com/golang/go/issues/37132)

---

### 2.2 CLI 框架：Cobra

**推荐方案：[spf13/cobra](https://github.com/spf13/cobra)**

| 维度 | 评估 |
|------|------|
| 学习曲线 | 中（子命令、PersistentFlags 需读文档） |
| 开发效率 | 高（子命令结构清晰，适合 `add` / `list` / `search`） |
| 社区生态 | 极活跃（kubectl、gh、Hugo 等） |
| 适合场景 | 多子命令、需 `--help` 与 Shell 补全的 CLI |

**为什么选 Cobra 而不是 urfave/cli？**

- 子命令树与 PRD 功能（向导、列表、搜索）天然匹配。
- 生态与示例多，利于 Java 背景开发者对照「命令 → 处理函数」心智模型。
- urfave/cli 更轻，但近年有维护与边界行为争议；fuckssh 功能会持续扩展（V2 备份恢复），Cobra 扩展成本更低。

**参考资料：**

- [Cobra 官方站](https://cobra.dev/)
- [CLI 框架对比（含 Cobra / urfave/cli）](https://github.com/Oursin/Go-CLI-Comparison)

---

### 2.3 交互式向导

**推荐方案：MVP 使用 [charmbracelet/huh](https://github.com/charmbracelet/huh) 或轻量 [AlecAivazis/survey](https://github.com/AlecAivazis/survey)**

| 维度 | 评估 |
|------|------|
| 学习曲线 | 低～中 |
| 开发效率 | 高（表单、默认值、回车确认） |
| 适合场景 | 密码/密钥模式、端口、别名、算法选择 |

**建议：**

- 优先 **huh**（维护活跃、API 现代）；若希望依赖更少，MVP 可对少量字段用 `fmt` + 简单校验。
- **V1 不引入** 全屏 TUI（如 `bubbletea`），避免与学习 Go 核心逻辑抢精力；列表/搜索用表格输出（`text/tabwriter` 或后续 [charmbracelet/lipgloss](https://github.com/charmbracelet/lipgloss)）即可。

---

### 2.4 SSH 与密钥

**推荐方案：分层能力**

| 能力 | 实现 | 依赖 |
|------|------|------|
| 密钥生成（Ed25519 默认） | `crypto/ed25519` + `golang.org/x/crypto/ssh` 的 `MarshalPrivateKey` / `MarshalAuthorizedKey` | 无 `ssh-keygen` |
| 密码模式：登录并写 `authorized_keys` | `golang.org/x/crypto/ssh` 客户端 | 无系统 `ssh` 亦可完成部署 |
| 用户验收与日常连接 | 调用系统 **`ssh`** | 需检测并提示安装 |
| 密钥连接模式 | 仅写 `ssh config`，引用已有私钥 | 无额外依赖 |
| 列出 / 搜索 Host | 解析 `~/.ssh/config`（见下） | 无 |

**密钥格式要点：**

- 默认 **Ed25519**；私钥使用 OpenSSH PEM（`OPENSSH PRIVATE KEY`），勿用 `x509.MarshalPKCS8PrivateKey` 处理 Ed25519（与 OpenSSH 不兼容）。
- 公钥写入 `~/.ssh/*.pub`，私钥 `0600`（Unix）；Windows 注意 ACL / 文档说明。

**OpenSSH 客户端检测（必须）：**

| 平台 | 说明 |
|------|------|
| macOS | 通常已自带 `/usr/bin/ssh` |
| Linux | 多数桌面已装 `openssh-client` |
| **Windows** | **可选功能，非所有环境默认安装**；需在向导/文档中引导「可选功能 → OpenSSH 客户端」或 `Add-WindowsCapability` |

**参考资料：**

- [Microsoft Learn — OpenSSH 可选功能](https://learn.microsoft.com/en-us/troubleshoot/windows-server/system-management-components/cant-install-openssh-features)
- [Stack Overflow — Go 生成 OpenSSH 兼容 Ed25519](https://stackoverflow.com/questions/71850135/generate-ed25519-key-pair-compatible-with-openssh)

---

### 2.5 `ssh config` 解析

**推荐方案：MVP 自研「受限解析器」+ 测试夹具；必要时引用 [kevinburke/ssh_config](https://github.com/kevinburke/ssh_config)**

| 维度 | 评估 |
|------|------|
| 原因 | PRD 已提示 Include、通配符 Host 的复杂度；MVP 应**明确支持范围**（例如：逐 `Host` 块、HostName/User/Port/IdentityFile） |
| 风险缓解 | 修改前备份 `config`；解析失败给出行号与片段 |

**搜索：** 对别名、HostName、IP 做大小写不敏感子串匹配即可；Host 数量 ≤ 60 时无需 Elasticsearch 类方案。

---

### 2.6 不需要的组件（明确排除）

| 组件 | 结论 |
|------|------|
| 前端框架 | 不需要 |
| 后端 HTTP 服务 | 不需要 |
| 数据库 | 不需要 |
| Redis / 消息队列 | 不需要 |
| 对象存储 | 不需要（V2 用户自备同步盘） |

---

### 2.7 测试、CI 与发布

| 项 | 推荐 | 时机 |
|----|------|------|
| 单元测试 | 标准库 `testing` + 表驱动；可选 `testify/assert` | MVP 起，解析与密钥逻辑必测 |
| 集成测试 | 标记 `//go:build integration`，真机 SSH 手动测为主 | 有空再做 |
| CI | GitHub Actions：`go test ./...`、lint | 准备对外发布时 |
| 交叉编译发布 | [GoReleaser](https://goreleaser.com/) 产出 Win/macOS/Linux 二进制 | MVP 稳定后 |
| 本地安装 | `go install github.com/you/fuckssh@latest` | 自用阶段即可 |

---

## 3. 架构设计

### 3.1 系统架构（逻辑）

```
用户
  │
  ▼
fuckssh CLI (Go 单二进制)
  │
  ├─► 解析 / 写入 ~/.ssh/config
  ├─► Go 生成密钥 → ~/.ssh/id_* / *.pub
  ├─► (密码模式) x/crypto/ssh → 远端 authorized_keys
  └─► 检测 PATH 中 ssh → 提示安装指引
        │
        ▼
系统 OpenSSH 客户端 (ssh)  ← 用户日常 ssh <别名> / VS Code / Tabby
```

### 3.2 技术栈全景

| 层次 | 技术选型 | 说明 |
|------|----------|------|
| 语言 | Go 1.22+ | 主开发语言，交叉编译 |
| CLI 框架 | Cobra | 子命令与帮助 |
| 交互 | huh / survey | 向导表单 |
| 密钥与 SSH 协议 | `crypto/ed25519`、`golang.org/x/crypto/ssh` | 生成密钥、密码模式部署 |
| 配置解析 | 自研受限解析（+ 可选 kevinburke/ssh_config） | list / search / 写入 |
| 系统依赖 | OpenSSH 客户端 `ssh` | 检测 + 提示；非密钥生成必需 |
| 测试 | `go test` | 表驱动为主 |
| 发布 | GoReleaser + GitHub Releases | 无 deadline，可后置 |

### 3.3 目录结构建议

```
fuckssh/
├── cmd/
│   └── fuckssh/
│       └── main.go              # 入口，调用 root.Execute()
├── internal/
│   ├── cmd/                     # Cobra 命令定义（add, list, search）
│   ├── config/                  # ssh config 读/写/解析/备份
│   ├── keys/                    # Ed25519/RSA 生成与落盘
│   ├── sshclient/               # 检测 ssh、密码模式部署公钥
│   ├── wizard/                  # 交互向导编排
│   └── platform/                # Win/macOS 路径、权限差异
├── testdata/                    # 样例 config、密钥夹具
├── docs/
├── go.mod
├── go.sum
├── Makefile                     # 可选：test、build、lint
└── .goreleaser.yaml             # 发布阶段再加
```

### 3.4 MVP 实现顺序（建议）

1. **`list` / `search`**：只读解析 + 测试夹具（风险低，先熟悉 Go 与 Cobra）。
2. **`ssh` 检测** + 统一错误提示文案（Win/macOS 分支）。
3. **`keys` 包**：Ed25519 生成 + OpenSSH 格式写文件。
4. **`add` 向导**：密钥连接模式（只写 config）→ 密码连接模式（SSH 部署公钥）。
5. **修改 config 前自动备份**（PRD 安全要求）。
6. **发布与 CI**（有余力再做）。

---

## 4. 风险评估

### 4.1 技术风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| Windows 无 `ssh` | 高（无法立即验收 `ssh 别名`） | 中 | 启动/向导前检测；分平台安装指引 |
| `ssh config` 手写复杂语法 | 中（列表不全） | 中 | MVP 文档写明支持范围；逐步增强 |
| Go 私钥格式与 OpenSSH 不兼容 | 高 | 低 | 仅用 `ssh.MarshalPrivateKey`；用系统 `ssh -i` 做冒烟测试 |
| Windows 文件权限与 Unix 不一致 | 中 | 中 | `internal/platform` 封装；文档说明 |
| 公钥部署失败（网络/防火墙） | 中 | 中 | 清晰错误、回滚指引、保留 config 备份 |

### 4.2 学习成本（Go / 生态）

| 技术 | 学习曲线 | 预估时间 | 建议 |
|------|----------|----------|------|
| Go 基础语法与模块 | 低～中 | 1～2 周业余 | 官方 Tour + 写 `list` 命令 |
| Cobra | 中 | 2～3 天 | 先做一个子命令跑通 |
| x/crypto/ssh | 中 | 3～5 天 | 先密钥生成，再密码登录 |
| huh / survey | 低 | 1 天 | 照着示例改字段 |
| 跨平台路径/权限 | 中 | 持续 | 公司 Win、家里 Mac 各测一遍 |

### 4.3 替代方案

| 场景 | 主要方案 | 备选方案 | 切换成本 |
|------|----------|----------|----------|
| 密钥生成 | Go `x/crypto/ssh` | 调用 `ssh-keygen` | 低 |
| 密码部署 | Go SSH 客户端 | `ssh` + `ssh-copy-id` 脚本 | 中 |
| 交互 UI | huh | 纯 stdin 问答 | 低 |
| CLI 框架 | Cobra | urfave/cli | 中 |

---

## 5. 实施建议

### 5.1 学习路径（无 deadline 节奏）

1. **第 1 阶段**：`go mod init`、Cobra 骨架、`list` 读 config + 单元测试。
2. **第 2 阶段**：`keys` 生成 Ed25519，在 Mac 上用 `ssh -i` 验证可读。
3. **第 3 阶段**：Windows 路径与 `ssh` 检测文案。
4. **第 4 阶段**：`add` 向导（密钥模式 → 密码模式）。
5. **第 5 阶段**：`search`、README、GoReleaser。

### 5.2 开发规范

| 项 | 建议 |
|----|------|
| 代码规范 | `gofmt` / `goimports`；可选 `golangci-lint` |
| Git | `main` 稳定；功能分支 `feat/list` 等 |
| 提交信息 | 简短英文或中文均可，保持一致 |
| 文档 | 每个子命令在 README 有一行示例 |

### 5.3 监控告警

本机 CLI **不需要** 应用监控栈；可选：

- 用户报错靠 GitHub Issues
- 发布前手动冒烟：Win + macOS 各走一遍「密码连接」主路径

---

## 6. 成本估算

### 6.1 开发成本

单人业余项目，不按人天计价；主要成本为**时间**与**双机测试**（公司 Win + 家里 Mac）。

### 6.2 运营成本（月）

| 项目 | 规格 | 成本 |
|------|------|------|
| 托管 | 无 | $0 |
| CI | GitHub Actions 免费档 | $0 |
| 域名 / 官网 | 可选 | $0～按需 |

---

## 7. 总结

### 7.1 推荐技术栈（一句话）

**Go + Cobra + huh/survey + `golang.org/x/crypto/ssh`（密钥生成与密码部署）+ 自研 ssh config 解析；系统 `ssh` 仅检测与引导安装，不依赖 `ssh-keygen`。**

### 7.2 关键决策点

1. **Go**：跨平台 CLI、单二进制、符合学 Go 与 PRD。
2. **Go 生成密钥**：减少 `ssh-keygen` 依赖；格式必须用 OpenSSH marshal API。
3. **系统 `ssh` 非 garantee**：检测 + 分平台提示；Windows 优先写好指引。
4. **无服务端、无数据库**：边界清晰，专注 config 与 OpenSSH 文件。
5. **实现顺序**：只读命令先行，降低学习与调试难度。

### 7.3 下一步行动

1. 使用 **system-architect** skill 做系统架构设计（模块接口、数据流、错误模型）。
2. 架构确认后，使用 **scaffold** skill 搭建 Go 项目脚手架。
3. 将 PRD §5.2「技术约束」与本文件交叉引用（架构文档定稿时同步）。

---

## 8. 参考资料汇总

- [fuckssh PRD](./fuckssh-PRD.md)
- [golang.org/x/crypto/ssh keys.go](https://github.com/golang/crypto/blob/master/ssh/keys.go)
- [Cobra](https://cobra.dev/)
- [Microsoft — Windows OpenSSH 可选功能](https://learn.microsoft.com/en-us/troubleshoot/windows-server/system-management-components/cant-install-openssh-features)
- [GoReleaser](https://goreleaser.com/)
