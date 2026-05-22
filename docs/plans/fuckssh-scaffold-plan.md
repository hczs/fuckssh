# fuckssh 脚手架计划

## 文档信息

- 版本：v1.0
- 日期：2026-05-22
- 基于：[技术选型](../fuckssh-技术选型.md) | [系统架构](../fuckssh-架构设计.md)

## 1. 项目概况

- 项目名称：fuckssh
- 项目路径：`e:\tmp\fuckssh`
- 技术栈：Go 1.22+、Cobra、huh（后续）、`golang.org/x/crypto/ssh`（后续）

## 2. 搭建方案

### 2.1 目录结构

```
fuckssh/
├── cmd/fuckssh/main.go          # 薄入口
├── internal/
│   ├── cmd/                     # Cobra 根命令与子命令
│   ├── config/                  # ssh config（占位）
│   ├── keys/                    # 密钥生成（占位）
│   ├── sshclient/               # ssh 检测与部署（占位）
│   ├── wizard/                  # 向导编排（占位）
│   └── platform/                # 跨平台路径（占位）
├── testdata/                    # 测试夹具目录
├── docs/
├── go.mod / go.sum
├── Makefile
├── .golangci.yml
├── .github/workflows/ci.yml
├── .gitignore
└── README.md
```

### 2.2 初始化方式

**手动搭建**（非 `cobra-cli init`）：技术选型与架构已约定 `cmd/fuckssh` + `internal/cmd` 分层，`cobra-cli` 默认生成顶层 `cmd/root.go`，与架构不一致。手动创建可一次对齐模块边界。

### 2.3 技术选型摘要

| 层次 | 技术 | 版本 |
|------|------|------|
| 语言 | Go | 1.22+（本机 1.26.2） |
| CLI | Cobra | latest stable |
| Linter | golangci-lint | v2 |
| CI | GitHub Actions | ubuntu + windows |

### 2.4 开发工具链

- Linter：golangci-lint（`gofmt`、`goimports`、`govet`、`staticcheck` 等默认启用）
- Formatter：`gofmt` / `goimports`（通过 golangci-lint formatter 或 `make fmt`）
- Git hooks：MVP 仅 CI；本地用 `make lint` / `make test`

### 2.5 CI/CD

- 平台：GitHub Actions
- 流水线：`go test ./...`、`golangci-lint`、多平台 `go build`

## 3. 预期生成文件清单

| 文件 | 说明 |
|------|------|
| `go.mod` | 模块 `github.com/fuckssh/fuckssh` |
| `cmd/fuckssh/main.go` | 入口 |
| `internal/cmd/*.go` | root + add/list/search 占位 |
| `internal/*/doc.go` | 各包占位 |
| `Makefile` | build、test、lint、fmt |
| `.golangci.yml` | lint 配置 |
| `.github/workflows/ci.yml` | CI |
| `.gitignore` | Go 忽略项 |
| `README.md` | 启动与开发说明 |

## 4. 验证计划

搭建完成后将逐项验证：

- [x] 依赖安装（`go mod tidy`，Cobra v1.10.2）
- [x] `go build` 成功（`bin/fuckssh.exe`）
- [x] Lint 检查通过（golangci-lint v2.12.2，0 issues）
- [x] `go test ./...` 通过（`internal/cmd` 含 help 冒烟测试）
- [x] CI 配置有效（`.github/workflows/ci.yml` 语法与步骤已对齐 Makefile 能力）

## 5. 开发指引（搭建完成后补充）

### 5.1 启动命令

```bash
go run ./cmd/fuckssh
go run ./cmd/fuckssh list
```

### 5.2 常用命令

```bash
make build
make test
make lint
make fmt
```

### 5.3 项目结构说明

- `internal/cmd`：仅 CLI 路由，业务逻辑放入 `config`、`keys`、`wizard` 等包
- `testdata/`：样例 `config` 与密钥夹具，供表驱动测试使用

## 6. 已知限制

- 子命令仅为占位实现，不含业务逻辑
- 未安装 huh、`x/crypto/ssh` 等业务依赖（按 MVP 顺序在开发阶段引入）
- GoReleaser 配置留待发布阶段
- 模块路径 `github.com/fuckssh/fuckssh` 为占位，发布前请改为实际 GitHub 仓库路径
