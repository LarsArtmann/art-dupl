# Contributing to art-dupl

Thank you for your interest in contributing to art-dupl! This guide covers everything you need to get started.

## Quick Start

```bash
# Enter the development environment
nix develop

# Or without Nix:
go install

# Run all checks
just ci
```

## Development Setup

### Prerequisites

- **Go 1.26+**
- **Nix** (recommended) or **Just** (alternative)

### Using Nix (Recommended)

```bash
nix develop          # Enter dev shell with all tools
nix build            # Build the project
nix flake check      # Run all checks (build + tests)
nix run . -- --help  # Run the tool
```

### Using Justfile

```bash
just build           # Build the project
just test            # Run tests
just check           # Run linter
just ci              # Format + lint + test
```

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Functions under 30 lines preferred
- Files under 300 lines preferred
- Explicit error handling — no panics for expected errors
- Context-first function parameters where applicable
- Idiomatic Go: `(T, error)` returns, not Result/Option types
- Strong typing: make impossible states unrepresentable

## Testing

```bash
just test            # All tests with coverage
just test-race       # With race detector
just test-unit       # Unit tests only
just test-integration # Integration tests only
just bench           # Benchmarks
just coverage        # Generate coverage.html
```

### BDD Tests

User-facing features should have BDD tests using Ginkgo/Gomega in the `bdd/` directory:

```bash
go test -v ./bdd     # Run BDD tests
```

## Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat(scope): add new feature
fix(scope): fix bug
refactor(scope): improve code structure
docs(scope): update documentation
test(scope): add or update tests
chore(scope): maintenance tasks
```

## Project Structure

```
cmd/           # CLI command definitions
config/        # Configuration management
detection/     # Multi-method detection coordination
domain/        # Domain models and value objects
errors/        # Typed error hierarchy
hash/          # Hash-based detection
job/           # File parsing and tree building
pkg/artdupl/   # Public SDK
printer/       # Output formatting
suffixtree/    # Suffix tree algorithm
syntax/        # AST handling
```

## Making Changes

1. Create a feature branch from `fork`
2. Make your changes
3. Run `just ci` to verify
4. Write tests for new functionality
5. Keep changes focused and small
6. Submit a pull request

## Questions?

Open an issue on GitHub or check the documentation in `docs/`.
