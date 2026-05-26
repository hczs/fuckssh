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

## CI / Release

| 触发条件 | 工作流 | 作用 |
|----------|--------|------|
| push / PR 到 `main` / `master` | [CI](.github/workflows/ci.yml) | `go test -race`、golangci-lint、三平台编译检查 |
| 推送 tag `v*`（如 `v0.1.0`） | [Release](.github/workflows/release.yml) | GoReleaser 构建并发布到 GitHub Releases |

发布新版本：

```bash
git tag v0.1.0
git push origin v0.1.0
```

本地试跑发布（不上传）：`make release-dry`（需安装 [GoReleaser](https://goreleaser.com/)）。

安装已发布版本：`go install github.com/fuckssh/fuckssh@v0.1.0`

**macOS 产物**（GitHub Releases 附件）：

| 文件（示例） | 适用 |
|--------------|------|
| `fuckssh_macos_x86_64.tar.gz` | Intel Mac |
| `fuckssh_macos_arm64.tar.gz` | Apple Silicon（M1/M2/M3 等） |
| `fuckssh_macos_all.tar.gz` | 通用二进制（Intel + Apple Silicon 合一，任选其一即可） |

## 许可证

待定。
