.PHONY: build test lint fmt run clean release-dry hooks check

BINARY := fuckssh
MAIN   := ./cmd/fuckssh

check: fmt lint test

build:
	go build -o bin/$(BINARY) $(MAIN)

test:
	go test ./...

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .
	goimports -w .

run:
	go run $(MAIN)

clean:
	rm -rf bin/

# 将 Git hooks 目录设为 .githooks（含 pre-commit：暂存 .go 自动 fmt + lint）
hooks:
	git config core.hooksPath .githooks
	@echo "Git hooks 已启用：core.hooksPath=.githooks"

# 本地模拟发布（不推送到 GitHub），需先安装 GoReleaser。
release-dry:
	goreleaser release --snapshot --clean --skip=publish
