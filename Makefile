.PHONY: build test lint fmt run clean

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
