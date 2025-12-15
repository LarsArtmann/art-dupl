# Justfile for art-dupl project

# Default recipe - run all development tasks
default: clean check test build

# Clean build artifacts
clean:
    rm -rf dist/ cover.out

# Run tests with coverage
test: clean
    go test -v -cover ./...

# Run linter
check:
    golangci-lint run

# Build the binary with optimization flags
build:
    #!/usr/bin/env bash
    set -euo pipefail
    mkdir -p dist/
    go build -ldflags "-s -w" -trimpath -o dist/dupl

# Install dependencies
deps:
    go mod download
    go mod tidy

# Run tests with race detector
test-race:
    go test -race -v ./...

# Generate coverage report
coverage:
    go test -coverprofile=cover.out ./...
    go tool cover -html=cover.out -o coverage.html

# Run benchmarks
bench:
    go test -bench=. -benchmem ./...

# Format code
fmt:
    gofmt -s -w .

# Run all checks (format, lint, test)
ci: fmt check test

# Build for different platforms
build-all:
    GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o dist/dupl-linux-amd64
    GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o dist/dupl-darwin-amd64
    GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -trimpath -o dist/dupl-windows-amd64.exe

# Install the binary locally
install-local:
    go install -ldflags "-s -w" -trimpath

# Find code duplicates using golangci-lint with dupl linter
fd:
    golangci-lint run --enable-only dupl -v

# Legacy alias for find-duplicates
find-duplicates: fd

# Show help
help:
    @just --list