# Justfile for art-dupl project

# Default recipe - run all development tasks
default: clean check test build

# Clean build artifacts
clean:
    rm -rf dist/ cover.out
    find . -name "*.test" -type f -delete 2>/dev/null || true

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
    go build -ldflags "-s -w" -trimpath -o dist/art-dupl ./cmd/art-dupl

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

# Run benchmarks with memory tracking
bench:
    go test -bench=. -benchmem ./...

# Run tests with coverage and show total percentage
test-coverage:
    go test -coverprofile=cover.out ./...
    go tool cover -func=cover.out | grep total

# Check coverage meets 80% threshold
check-coverage:
    @go test -coverprofile=cover.out ./...
    @COVERAGE=$$(go tool cover -func=cover.out | grep total | awk '{print $$3}' | tr -d '%'); \
    if [ $$(echo "$$COVERAGE < 80" | bc) -eq 1 ]; then \
        echo "Coverage is $$COVERAGE%, below 80% threshold"; \
        exit 1; \
    fi; \
    echo "Coverage is $$COVERAGE%, meets 80% threshold"

# Run fuzz tests
test-fuzz:
    go test -fuzz=. -fuzztime=30s ./...

# Run fuzz tests with longer duration
test-fuzz-long:
    go test -fuzz=. -fuzztime=60s ./...

# List all tests
list-tests:
    go test -list=. ./...

# Run unit tests only (exclude integration/bdd)
test-unit:
    go test -run=^Test -v ./...

# Run integration tests only
test-integration:
    go test -run=Integration -v ./...

# Run benchmarks with allocs reporting
bench-allocs:
    go test -bench=. -benchmem -run=^$ ./...

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
    go install -ldflags "-s -w" -trimpath ./cmd/art-dupl

# Find code duplicates using art-dupl
fd:
    art-dupl -v

# Legacy alias for find-duplicates
find-duplicates: fd

# Show help
help:
    @just --list