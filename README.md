# fuckssh

跨平台 CLI：围绕标准 `~/.ssh/config` 提供 VPS 一站式 SSH 配置、列表与搜索。

## 要求

- Go 1.22+
- （可选）[golangci-lint](https://golangci-lint.run/) 用于本地 lint

## 快速开始

```bash
# 安装依赖并构建
go mod tidy
go build -o bin/fuckssh ./cmd/fuckssh

# 或直接运行
go run ./cmd/fuckssh --help
```

## 命令（MVP 规划中）

| 命令 | 说明 |
|------|------|
| `fuckssh add` | 交互式向导添加 VPS |
| `fuckssh list` | 列出 config 中的 Host |
| `fuckssh search <query>` | 按别名、HostName、IP 搜索 |

当前子命令为占位实现，业务逻辑按 [架构设计](docs/fuckssh-架构设计.md) 分阶段落地。

## 开发

```bash
make build    # 构建到 bin/fuckssh
make test     # 运行测试
make lint     # golangci-lint
make fmt      # gofmt + goimports
make run      # go run ./cmd/fuckssh
```

## 文档

- [PRD](docs/fuckssh-PRD.md)
- [技术选型](docs/fuckssh-技术选型.md)
- [系统架构](docs/fuckssh-架构设计.md)
- [脚手架计划](docs/plans/fuckssh-scaffold-plan.md)

## CI

GitHub Actions 在 push/PR 时执行 `go test`、golangci-lint，并在 Ubuntu / Windows / macOS 上交叉构建。

## 许可证

待定。
