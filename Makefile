.PHONY: build test lint fmt run clean release-dry

BINARY := fuckssh
MAIN   := ./cmd/fuckssh

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

# 本地模拟发布（不推送到 GitHub），需先安装 GoReleaser。
release-dry:
	goreleaser release --snapshot --clean --skip=publish
