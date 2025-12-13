# AGENTS.md - dupl Repository Guide

This document provides essential information for AI agents working on the **dupl** repository - a Go tool for finding code clones using suffix tree algorithms.

## Project Overview

**dupl** is a code duplication detection tool specifically for Go source files. It analyzes abstract syntax trees (ASTs) to find structural code clones while ignoring literal values. The tool uses suffix tree algorithms to efficiently identify duplicate code patterns.

### Core Architecture
- **Main Package**: Entry point and CLI interface in `main.go`
- **suffixtree**: Core suffix tree implementation for clone detection
- **syntax**: AST handling, serialization, and node processing
- **job**: Orchestrates file parsing and tree building
- **printer**: Output formatting (text, HTML, plumbing formats)
- **lib**: Utility functions and test helpers

## Development Commands

### Essential Commands
```bash
# Build the project
make build
# or
go build -ldflags "-s -w" -trimpath

# Run tests
make test
# or
go test -v -cover ./...

# Run linting
make check
# or
golangci-lint run

# Clean build artifacts
make clean

# Run all development tasks
make  # Runs clean, check, test, build in sequence
```

### Building the CLI Tool
```bash
# Build for current platform
go build

# Build with production flags
go build -ldflags "-s -w" -trimpath

# Run the tool
./dupl [flags] [paths]
```

## Code Patterns and Conventions

### Go Standards
- Standard Go formatting with `gofmt`
- Context-first function parameters where applicable
- Error handling with explicit returns, no panics for expected errors
- Package-level constants for configuration values
- Clear separation between public and private APIs

### Naming Conventions
- Packages use lowercase, single words where possible
- Public functions use PascalCase
- Private functions use camelCase
- Constants use UPPER_SNAKE_CASE
- Error variables follow the `Err` prefix pattern

### Project-Specific Patterns

#### Suffix Tree Implementation
- Uses `Pos` type for positions in sequences
- Node-based tree structure with transition maps
- Stream processing for handling large codebases
- Threshold-based filtering to eliminate noise

#### AST Processing
- Custom Node struct with Type, Filename, Pos, End fields
- Serialization transforms AST nodes to token sequences
- Maximum children limit (10,000) to prevent stack overflow
- SHA1 hashing for identifying identical code structures

#### Concurrency Patterns
- Goroutine-based pipeline processing (parse → serialize → build tree)
- Channel-based communication between stages
- Explicit synchronization with done channels

## Testing Approach

### Test Structure
- Tests follow Go conventions with `_test.go` files
- Use standard `testing` package
- Table-driven tests for multiple scenarios
- Performance testing for large inputs

### Test Categories
- Unit tests for core algorithms (suffixtree package)
- Integration tests for end-to-end workflows
- Performance tests with large synthetic data
- Template-based test data generation

### Running Tests
```bash
# All tests with coverage
go test -v -cover ./...

# Specific package tests
go test -v ./suffixtree
go test -v ./syntax

# Run with race detector
go test -race ./...
```

## Build and CI

### Build Process
- Uses Go modules with `go.mod`
- Cross-platform builds supported (Linux, macOS, Windows)
- CGO disabled for static binaries
- Build flags for optimized production binaries

### CI Pipeline
- GitHub Actions with matrix testing (multiple Go versions and OS)
- Golangci-lint for code quality checks
- Dependency management verification
- Tests run on oldstable and stable Go versions

### Dependency Management
- Minimal dependencies (only standard library and Go tools)
- Golangci-lint is the only external development dependency
- Version pinning through go.mod

## Important Gotchas

### Performance Considerations
- Large composite literals are serialized to prevent stack overflow
- Maximum children limit prevents excessive memory usage
- Stream processing keeps memory usage bounded

### File Processing
- Only processes `.go` files by default
- Vendor directory excluded by default (use `-vendor` flag)
- Can accept file paths from stdin with `-files` flag

### Output Formats
- Default: Text output with file paths and line numbers
- HTML: Includes actual duplicate code fragments
- Plumbing: Machine-readable format for script integration
- Mutually exclusive output formats (can't combine HTML and plumbing)

### Threshold Configuration
- Default minimum token sequence size: 15
- Adjustable with `-threshold` or `-t` flag
- Higher values reduce false positives but may miss smaller clones

## CLI Usage Patterns

### Common Workflows
```bash
# Scan current directory
./dupl

# Scan with higher threshold
./dupl -t 100

# Scan specific paths
./dupl ./src ./lib

# Generate HTML report
./dupl -html -t 50 > report.html

# Scan test files only
find . -name '*_test.go' | ./dupl -files

# Include vendor directory
./dupl -vendor
```

### Debug Mode
- Use `-v` or `-verbose` for verbose logging
- Shows tree building and clone detection progress
- Helpful for understanding processing flow

## Module Structure

### Import Paths
All imports use the module path: `github.com/LarsArtmann/art-dupl`
```go
import (
    "github.com/LarsArtmann/art-dupl/suffixtree"
    "github.com/LarsArtmann/art-dupl/syntax"
    "github.com/LarsArtmann/art-dupl/job"
    "github.com/LarsArtmann/art-dupl/printer"
)
```

### Package Dependencies
- No external runtime dependencies
- Only Go standard library for core functionality
- Golangci-lint for development tooling only

## Development Guidelines

### Code Style
- Follow Go conventions and idiomatic patterns
- Keep functions focused and small
- Prefer explicit interfaces over implicit ones
- Use channels for goroutine communication
- Error handling should be explicit and consistent

### Testing Guidelines
- Write tests for all exported functions
- Use table-driven tests for multiple scenarios
- Test edge cases and error conditions
- Consider performance implications of algorithms
- Use race detector for concurrent code

### Contributing
- All changes must pass `make check` (linting)
- All tests must pass with coverage
- Follow the existing code patterns
- Consider performance impact of changes
- Maintain backward compatibility for CLI interface