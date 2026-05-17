# AGENTS.md - art-dupl Repository Guide

This document provides essential information for AI agents working on the **art-dupl** repository - a Go tool for finding code clones using suffix tree algorithms with multi-method detection support.

## Project Overview

**art-dupl** is a modern code duplication detection tool for Go source files. It analyzes abstract syntax trees (ASTs) to find structural code clones while ignoring literal values. The tool supports multiple detection algorithms, professional CLI with Fang framework, and comprehensive output formats.

### Key Features

- **Multi-method detection**: Suffix tree algorithm (art-dupl) and hash-based detection
- **Professional CLI**: Built with Fang framework (Cobra) with auto-completion and version info
- **Multiple output formats**: Text, HTML, JSON, plumbing, and CSV (for stats)
- **Statistics subcommand**: Aggregated duplication metrics and project overview
- **Smart filtering**: SQLC and templ generated code filtering with pattern matching
- **Configuration files**: JSON-based configuration for team consistency
- **Sorting options**: By size, occurrence, or hash
- **BDD tests**: Ginkgo/Gomega behavior-driven development test suite

### Core Architecture

#### Primary Packages

- **cmd/**: CLI command definitions (root, stats, version)
- **config/**: Configuration management and validation
- **cli/**: CLI runtime, validation, and sorting logic
- **detection/**: Multi-method detection coordination (art-dupl, hash, todos, legacy)
- **suffixtree/**: Core suffix tree implementation for AST-based detection
- **syntax/**: AST handling, serialization, and node processing
- **hash/**: Rolling hash-based detection implementation
- **job/**: Orchestrates file parsing and tree building with profiling
- **printer/**: Output formatting (text, HTML, JSON, plumbing, stats)
- **adapter/**: Adapter pattern for printer abstraction

#### Supporting Packages

- **domain/**: Domain types and models (Clone, CloneGroup, StringPool)
- **types/**: Type definitions and shared types
- **errors/**: Error handling with typed error wrappers
- **pkg/**: Utility packages (artdupl, position, logger)
- **internal/**: Internal utilities (testutil, enum, utils, simd)
- **migration/**: Migration utilities for version compatibility
- **lib/**: Legacy utility functions (being phased out)

## Development Commands

### Essential Commands (Prefer Justfile)

**IMPORTANT**: According to project standards, always prefer justfile commands (95% of cases). Only use make commands if justfile is unavailable.

```bash
# Build the project
just build
# Output: dist/art-dupl

# Run all development tasks (clean, check, test, build)
just default
# or just:
just

# Run tests with coverage
just test

# Run linting
just check

# Clean build artifacts
just clean

# Format code
just fmt

# Run all checks (format, lint, test)
just ci

# Install locally
just install-local
```

### Advanced Testing Commands

```bash
# Run tests with race detector
just test-race

# Generate coverage report
just coverage
# Output: coverage.html

# Run benchmarks
just bench

# Check coverage meets 80% threshold
just check-coverage

# Run fuzz tests
just test-fuzz

# Run fuzz tests with longer duration
just test-fuzz-long

# Run unit tests only (exclude integration/bdd)
just test-unit

# Run integration tests only
just test-integration

# Run benchmarks with allocations reporting
just bench-allocs
```

### Alternative Makefile Commands

Makefile uses `GOEXPERIMENT=jsonv2` flag for JSON v2 support:

```bash
make build    # Build with JSONv2 experiment
make test     # Test with JSONv2 experiment
make check    # Lint with JSONv2 experiment
make clean    # Clean build artifacts
```

### Building the CLI Tool

```bash
# Build with justfile (recommended)
just build
# Output: dist/art-dupl

# Install locally with justfile (recommended)
just install-local
# Output: $GOPATH/bin/art-dupl

# Manual build (if justfile unavailable)
go build -ldflags "-s -w" -trimpath -o art-dupl ./cmd/art-dupl

# Run the tool
./art-dupl [flags] [paths]

# Or if installed
art-dupl [flags] [paths]
```

## Code Patterns and Conventions

### Go Standards

- Standard Go formatting with `gofmt`
- Context-first function parameters where applicable
- Error handling with explicit returns, no panics for expected errors
- Package-level constants for configuration values
- Clear separation between public and private APIs
- Dependency injection with `samber/do` for complex dependencies

### Naming Conventions

- Packages use lowercase, single words where possible
- Public functions use PascalCase
- Private functions use camelCase
- Constants use UPPER_SNAKE_CASE
- Error variables follow the `Err` prefix pattern
- Interface names often use -er suffix (e.g., `StatsPrinter`)

### Project-Specific Patterns

#### Multi-Method Detection

- Detection coordinated by `detection.MultiDetector`
- Each detection method runs independently via goroutines
- Results combined and deduplicated through channels
- Detection methods configurable via `config.DetectionMethods`

#### Suffix Tree Implementation

- Uses `Pos` type for positions in sequences
- Node-based tree structure with transition maps
- Stream processing for handling large codebases
- Threshold-based filtering to eliminate noise
- SIMD optimizations for performance (via internal/simd)

#### AST Processing

- Custom Node struct with Type, Filename, Pos, End fields
- Serialization transforms AST nodes to token sequences
- Maximum children limit (10,000) to prevent stack overflow
- SHA1 hashing for identifying identical code structures

#### Concurrency Patterns

- Goroutine-based pipeline processing (parse → serialize → build tree)
- Channel-based communication between stages
- Explicit synchronization with done channels
- Multiple detection methods run in parallel when configured

#### Domain Types

- Domain models in `domain/` package with typed IDs
- StringPool for efficient string deduplication
- Clone and CloneGroup types for type-safe clone representation
- Strong typing prevents impossible states

#### Printer Adapter Pattern

- `adapter/` package provides abstraction over printer implementations
- Interface-based design for multiple output formats
- Format-specific printers (text, HTML, JSON, plumbing, stats)
- Sortable output with configurable sorting options

## Testing Approach

### Test Structure

- Tests follow Go conventions with `_test.go` files
- Use standard `testing` package for unit tests
- **BDD tests** use Ginkgo/Gomega framework in `bdd/` directory
- Table-driven tests for multiple scenarios
- Performance testing with benchmarks (`_bench_test.go`)
- Fuzz testing for robustness (`fuzz/` directory)

### Test Categories

- **BDD tests**: Behavior-driven development tests in `bdd/` package using Ginkgo/Gomega
  - User workflow scenarios
  - CLI command integration tests
  - Configuration file tests
  - Filter feature tests
  - Sorting tests
  - Stats subcommand tests
  - Plumbing and output tests
- **Unit tests**: Core algorithms (suffixtree, syntax, hash packages)
- **Integration tests**: End-to-end workflows in `internal/configtest/` and `internal/filtertest/`
- **Performance tests**: Benchmarks with large synthetic data
- **Fuzz tests**: Property-based testing for edge cases

### Running Tests

```bash
# All tests with coverage
just test
# or
go test -v -cover ./...

# BDD tests specifically
go test -v ./bdd

# Specific package tests
go test -v ./suffixtree
go test -v ./syntax
go test -v ./detection

# Run with race detector
just test-race
# or
go test -race ./...

# Run fuzz tests
just test-fuzz

# Run benchmarks
just bench

# Coverage report
just coverage
# Output: coverage.html

# Check coverage threshold
just check-coverage
```

### BDD Test Utilities

The `internal/testutil/bdd.go` provides comprehensive helpers for BDD tests:

- `BDDTestSetup`: Full test environment setup
- `NewBDDTestSetupForGinkgo()`: Setup for Ginkgo tests
- `CreateTestFiles()`: Create test Go files
- `CreateDuplicateFiles()`: Create files with duplicate code
- `RunArtDupl()`: Execute art-dupl with flags
- `RunArtDuplOnDir()`: Execute on specific directory
- `RunArtDuplWithStdin()`: Execute with stdin input
- `Cleanup()`: Clean up test artifacts

## Build and CI

### Build Process

- Uses Go modules with `go.mod`
- Cross-platform builds supported (Linux, macOS, Windows)
- CGO disabled for static binaries
- Build flags for optimized production binaries (`-ldflags "-s -w" -trimpath`)
- JSONv2 experiment enabled via `GOEXPERIMENT=jsonv2` for Makefile builds
- Justfile builds output to `dist/art-dupl` by default

### CI Pipeline

- GitHub Actions in `.github/workflows/`:
  - `build.yml`: Matrix testing (multiple Go versions and OS)
  - `checks.yml`: Code quality checks
  - `performance.yml`: Performance regression testing
- Golangci-lint for code quality checks (configuration in `.golangci.yml`)
- Dependency management verification
- Tests run on oldstable and stable Go versions
- BDD tests included in CI

### Dependency Management

- **Core runtime dependencies**:
  - `github.com/charmbracelet/fang`: Professional CLI framework
  - `github.com/spf13/cobra`: Command-line interface library
  - `github.com/onsi/ginkgo/v2`: BDD testing framework
  - `github.com/onsi/gomega`: Gomega matchers for Ginkgo
- **Development dependencies**:
  - `github.com/golangci/golangci-lint`: Linter (development only)
- Version pinning through go.mod and go.sum
- Minimal external dependencies preferred

## Important Gotchas

### Performance Considerations

- Large composite literals are serialized to prevent stack overflow
- Maximum children limit prevents excessive memory usage
- Stream processing keeps memory usage bounded
- **Suffix tree uses O(1) map-based transition lookup** (optimized from O(n) linear search)
- SIMD optimizations available for improved performance (via internal/simd)
- Hash-based detection is faster than suffix tree for some scenarios
- Multiple detection methods can be run together (adds overhead)
- `--semantic` flag is now faster than non-semantic mode due to optimization (was 9x slower)

### File Processing

- Only processes `.go` and `.templ` files by default
- Vendor directory excluded by default (use `-vendor` flag)
- Can accept file paths from stdin with `-files` flag
- **Smart filtering** for generated code (filtered by default):
  - SQLC files auto-detected via `sqlc.yaml` in parent directories
  - Templ-generated `*_templ.go` files filtered (source `.templ` files included by default)
  - Protobuf `.pb.go` files filtered by default
  - Mockgen and stringer generated files filtered by default
  - Override with `--include-sqlc`, `--include-templ`, `--include-protobuf`, `--include-mockgen`, `--include-stringer`
  - Custom patterns via `-include-pattern` and `-exclude-pattern`

### Detection Methods

- **art-dupl** (default): Suffix tree algorithm on AST tokens
- **hash**: Rolling hash on file content (faster, different tradeoffs)
- **hash,art-dupl**: Run both for comprehensive analysis
- Methods configured via `-detection-methods` or `-m` flag
- Multiple methods run in parallel via goroutines

### Semantic Detection

- **Default**: Structural-only matching (AST structure, ignoring identifier names). Config default: `Semantic: false`
- With `--semantic`: Clones matched by both structure AND identifier semantics (e.g., `a.String()` ≠ `b.Error()`)
- With `--structural` (explicit): Same as default — matches by AST structure only (e.g., `a.String()` = `b.Error()`)
- Example: `a.String()` WILL match `b.Error()` by default (structural-only); with `--semantic`, they will NOT match
- Useful for reducing false positives from similar-looking but semantically different code
- Implementation: FNV-1a hash of identifiers encoded into AST node types

### Output Formats

- **Default**: Text output with file paths and line numbers
- **HTML**: Includes actual duplicate code fragments with syntax highlighting and diff visualization
- **JSON**: Structured output with statistics and clone groups (JSONv2)
- **Simple-JSON**: Simpler JSON format with impact score and instances
- **Plumbing**: Machine-readable format for script integration
- **SARIF**: SARIF 2.1.0 for GitHub Advanced Security / CodeQL integration
- **Stats**: Multiple formats (text, JSON, CSV) via stats subcommand
- Output directory support via `-all` flag with `--output-dir`

### Sorting Options

- **size**: Shows largest clones first (highest token count) - default
- **occurrence**: Shows most widespread clones first (most files)
- **hash**: Alphabetical order by hash value
- **total-tokens**: Highest total token count across all instances
- Configured via `-sort` flag

### Threshold Configuration

- Default minimum token sequence size: 15
- Adjustable with `-threshold` or `-t` flag
- Higher values reduce false positives but may miss smaller clones
- Apply threshold per detection method

### Configuration Files

- JSON configuration files supported via `-config` or `-c` flag
- Default config file search: `dupl.json` in current directory
- Configuration merges with CLI flags (flags take precedence)
- Supports threshold, output format, paths, filtering, and detection methods

## CLI Usage Patterns

### Common Workflows

```bash
# Scan current directory
art-dupl

# Scan with higher threshold
art-dupl -t 100

# Scan specific paths
art-dupl ./src ./lib

# Generate HTML report
art-dupl -html -t 50 > report.html

# JSON output for CI/CD
art-dupl --json -t 20 ./src | jq

# Generate all formats to directory
art-dupl --all --output-dir ./reports ./src

# Most widespread clones first
art-dupl --plumbing --sort occurrence ./src

# Use hash-based detection (faster)
art-dupl -m hash ./src

# Run both detection methods
art-dupl -m "hash,art-dupl" ./src

# Scan test files only
find . -name '*_test.go' | art-dupl --files

# Include vendor directory
art-dupl --vendor

# Override: include sqlc-generated files
art-dupl --include-sqlc ./src

# Override: include templ-generated files
art-dupl --include-templ ./src

# Custom include patterns
art-dupl --include-pattern "vendor/*" --include-pattern "gen/*" ./src

# Configuration file
art-dupl -c dupl.json
```

### Stats Subcommand

```bash
# Show statistics for current directory
art-dupl stats

# Stats in JSON format
art-dupl stats --format json ./src

# Stats in CSV format for spreadsheets
art-dupl stats --format csv ./src

# Stats with higher threshold
art-dupl stats -t 20 .

# Stats for specific paths
art-dupl stats ./src ./lib

# Extract total clones from JSON
art-dupl stats --format json . | jq '.overview.totalClones'
```

### Debug Mode

- Use `-v` or `--verbose` for verbose logging
- Use `-vv` for extra verbosity
- Shows tree building and clone detection progress
- Helpful for understanding processing flow
- Can enable profiling with `-profile` flag (hidden)

### Shell Completion

```bash
# Bash completion
source <(art-dupl completion bash)

# Zsh completion
source <(art-dupl completion zsh)

# Fish completion
art-dupl completion fish | source

# PowerShell completion
art-dupl completion powershell | Out-String | Invoke-Expression
```

## Module Structure

### Import Paths

All imports use the module path: `github.com/LarsArtmann/art-dupl`

```go
import (
    "github.com/LarsArtmann/art-dupl/cmd"
    "github.com/LarsArtmann/art-dupl/config"
    "github.com/LarsArtmann/art-dupl/detection"
    "github.com/LarsArtmann/art-dupl/suffixtree"
    "github.com/LarsArtmann/art-dupl/syntax"
    "github.com/LarsArtmann/art-dupl/hash"
    "github.com/LarsArtmann/art-dupl/job"
    "github.com/LarsArtmann/art-dupl/printer"
    "github.com/LarsArtmann/art-dupl/domain"
    "github.com/LarsArtmann/art-dupl/adapter"
    "github.com/LarsArtmann/art-dupl/types"
    "github.com/LarsArtmann/art-dupl/errors"
    "github.com/LarsArtmann/art-dupl/pkg/artdupl"
    "github.com/LarsArtmann/art-dupl/pkg/logger"
    "github.com/LarsArtmann/gogenfilter"
    "github.com/LarsArtmann/art-dupl/pkg/position"
)
```

### Package Dependencies

**Runtime dependencies**:

- `github.com/charmbracelet/fang`: Professional CLI framework
- `github.com/spf13/cobra`: Command-line interface

**Testing dependencies**:

- `github.com/onsi/ginkgo/v2`: BDD testing framework
- `github.com/onsi/gomega`: Gomega matchers for Ginkgo

**Development dependencies**:

- `github.com/golangci/golangci-lint`: Linter (development only)

### Package Organization

```
art-dupl/
├── cmd/              # CLI command definitions
├── config/           # Configuration management
├── cli/              # CLI runtime and validation
├── detection/        # Multi-method detection
├── suffixtree/       # Suffix tree algorithm
├── syntax/           # AST processing
├── hash/             # Hash-based detection
├── job/              # Analysis orchestration
├── printer/          # Output formatting
├── adapter/          # Printer adapter pattern
├── domain/           # Domain models
├── types/            # Shared types
├── errors/           # Error handling
├── pkg/              # Utility packages
│   ├── artdupl/
│   ├── logger/
│   └── position/
├── internal/         # Internal utilities
│   ├── testutil/
│   ├── enum/
│   ├── utils/
│   └── simd/
├── migration/        # Migration utilities
├── lib/              # Legacy utilities
├── bdd/              # BDD tests
├── docs/             # Documentation
└── examples/         # Usage examples
```

## Development Guidelines

### Code Style

- Follow Go conventions and idiomatic patterns
- Keep functions focused and small (<30 lines preferred)
- Prefer explicit interfaces over implicit ones
- Use channels for goroutine communication
- Error handling should be explicit and consistent
- Use dependency injection where appropriate (samber/do)
- Make impossible states unrepresentable via strong types

### Testing Guidelines

- **Write BDD tests** for user-facing features using Ginkgo/Gomega
- Write unit tests for all exported functions
- Use table-driven tests for multiple scenarios
- Test edge cases and error conditions
- Consider performance implications of algorithms
- Use race detector for concurrent code (`just test-race`)
- Write fuzz tests for functions with complex inputs
- Maintain high test coverage (>80% preferred)

### Error Handling

- Use typed errors from `errors/` package
- Wrap errors with context using `duplerrors.Wrap*` functions
- Distinguish between validation, config, and analysis errors
- Return errors explicitly, never panic for expected errors
- Provide context in error messages for debugging

### Idiomatic Go Patterns (CRITICAL)

**This project uses idiomatic Go - NOT functional programming patterns:**

✅ **DO use idiomatic Go:**

```go
// Standard (T, error) returns
func ParseFile(filename string) (*Node, error) {
    node, err := parser.Parse(filename)
    if err != nil {
        return nil, fmt.Errorf("failed to parse %s: %w", filename, err)
    }
    return node, nil
}

// Explicit error handling
result, err := ParseFile("test.go")
if err != nil {
    return err
}
```

❌ **DON'T use functional patterns:**

```go
// AVOID: Result[T] types (like samber/mo or Rust)
func ParseFile(filename string) Result[*Node] { ... }

// AVOID: Option[T] types
func FindClone(id string) Option[*Clone] { ... }

// AVOID: Railway-oriented programming
result := ParseFile("test.go").FlatMap(Validate).FlatMap(Process)
```

**Why idiomatic Go?**

- Zero overhead (no wrapper allocations)
- Every Go developer understands it
- Consistent with standard library
- Better stack traces for debugging
- No foreign dependency required

**Domain validation:** Use `IsValid() error` methods, not Result types:

```go
func (c Clone) IsValid() error {
    if c.EndLine < c.StartLine {
        return errors.New("end line must be >= start line")
    }
    return nil
}
```

### Configuration

- Use `config.Config` struct for all configuration
- Support both CLI flags and JSON config files
- Merge config sources with CLI flags taking precedence
- Validate configuration before use with `config.ValidateConfig()`
- Use typed enums for detection methods, output formats, etc.

### Performance

- Measure before optimizing
- Use benchmarks for performance-critical code
- Consider SIMD optimizations for hot paths
- Profile with `-profile` flag when investigating performance
- Use channels efficiently to avoid memory bloat
- Stream large inputs to keep memory bounded

### Contributing

- All changes must pass `just check` (linting)
- All tests must pass with coverage
- Follow the existing code patterns
- Consider performance impact of changes
- Maintain backward compatibility for CLI interface
- Update documentation for new features
- Add BDD tests for new user-facing features
- Run `just ci` before committing (format, lint, test)

## Documentation

### Documentation Structure

The `docs/` directory contains comprehensive documentation:

- **API docs** (`docs/api/`): Generated API documentation
- **Status reports** (`docs/status/`): Daily/weekly development status reports
- **Planning docs** (`docs/planning/`): Implementation plans and strategies
- **Migration guides**: Documentation for breaking changes
- **Feature docs**: Detailed documentation for major features
- **Benchmark results**: Performance analysis and comparisons

### Key Documentation Files

- `README.md`: Quick start and basic usage
- `HOW_TO_USE.md`: Detailed usage examples
- `MIGRATION_GUIDE.md`: Migration guide for version upgrades
- `FEATURES.md`: Comprehensive feature list
- `SDK_DESIGN.md`: SDK design for programmatic access
- `TESTING.md`: Testing guidelines and practices

## Architecture Highlights

### Multi-Method Detection

The detection architecture supports multiple algorithms:

1. **Suffix Tree (art-dupl)**: Original algorithm on AST tokens
2. **Hash-based**: Rolling hash on file content for faster detection
3. **Combinations**: Run both for comprehensive analysis

Each method runs independently via goroutines, results combined and deduplicated.

### Domain Types

Strong typing throughout the codebase:

- `domain.Clone`: Type-safe clone representation
- `domain.CloneGroup`: Grouped clones with metadata
- `domain.StringPool`: Efficient string deduplication
- Typed enums for detection methods, output formats, sorting options

### Printer Adapter Pattern

The `adapter/` package provides abstraction over output formats:

- Interface-based design for extensibility
- Format-specific implementations (text, HTML, JSON, plumbing, stats)
- Sorting and filtering capabilities built-in

### Configuration System

Multi-layered configuration:

1. Default configuration
2. JSON config file (optional)
3. CLI flags (override all)

Configuration merging and validation ensures consistency.

## Important Project Notes

### Memory File Instructions

Check `~/.config/crush/AGENTS.md` for general agent instructions that apply to all projects, including:

- Development standards and quality mandates
- Testing requirements (100% automated)
- Commit workflow standards
- Error handling protocols
- Tool usage preferences

Project-specific AGENTS.md (this file) takes precedence over general instructions.

### Continuous Improvement

This project follows strict quality standards:

- Zero tolerance for technical debt
- Fix issues on sight
- Refactor immediately when code exceeds 300 lines or 30 lines per function
- Extract duplicate code after 3 instances
- Use TODOs liberally, address older TODOs within 1 week

### Build System

**CRITICAL**: Always prefer justfile commands over Makefile commands:

- Justfile: `just build`, `just test`, `just check`, etc.
- Makefile: Only use if justfile unavailable
- Makefile uses `GOEXPERIMENT=jsonv2` for JSON v2 support
- Justfile builds output to `dist/art-dupl` directory

### Nix Flake — Private Dependency Pattern

The `flake.nix` handles the private `gogenfilter` dependency using a **two-phase dummy/replace pattern**:

**Why it's needed:** `buildGoModule` creates two derivations:

1. **goModules** (fixed-output): Runs `go mod vendor` in sandbox (no SSH, no store path refs)
2. **Main build** (regular): Compiles the binary (CAN reference store paths)

**How it works:**

- `gogenfilter` is a flake input fetched via SSH at evaluation time (`flake = false`)
- `builtins.readFile "${gogenfilter}/go.mod"` reads the real go.mod at Nix eval time
- In goModules (`overrideModAttrs`): dummy dir with real go.mod/go.sum + `-replace` to `./dummy`
- In main build (`preBuild`): swap dummy for real gogenfilter from flake input, patch `vendor/modules.txt`

**Key files:** `flake.nix:27-30` (eval-time reads), `flake.nix:54-70` (overrideModAttrs), `flake.nix:72-79` (preBuild)

**When gogenfilter changes:**

1. Update `rev=` in `flake.nix` line 7
2. Set `vendorHash = ""` (empty string, NOT null)
3. Run `nix build` — it will fail with the correct hash
4. Copy that hash into `vendorHash`
5. If gogenfilter adds new **direct** deps, add blank imports to `dummy/dummy.go` in `overrideModAttrs`

**Dummy go.mod is auto-synced:** `builtins.readFile` reads the real go.mod at eval time, so indirect dep changes are automatic. Only new **direct** deps in gogenfilter require updating the dummy.go imports (rare).

### Codebase Architecture — Key Decisions

**printer/html.go split (2026-04-30):** Split into 4 files:

- `html.go` (365L): Core struct, constructors, PrintHeader, PrintClones
- `html_template.go` (523L): CSS/HTML template const
- `html_diff.go` (315L): Diff visualization functions
- `html_summary.go` (302L): Summary, footer with JS, OutputHTML

**Config.Only typed as FileType (2026-04-30):** `config.Config.Only` changed from `string` to `config.FileType`. Uses `FileType.Matches()` instead of hand-rolled `matchesOnlyFilter()`. All callers updated.

**Analysis timestamps as time.Time (2026-04-30):** `domain.Analysis.CreatedAt` changed from `string` to `time.Time`, `CompletedAt` from `*string` to `*time.Time`. Validation uses `IsZero()`.

**Error modernization (2026-04-30):** `errors.As` → `errors.AsType` (Go 1.24+). All `//nolint:err113` replaced with typed errors from the project's error hierarchy.

**Pipeline unification (2026-04-30):** CLI and SDK now share `detection.MultiDetector` dispatch. Fixed critical bug where SDK rebuilt suffix tree from scratch (double CPU/memory). Introduced `pipelineResult` carrying both `data` and `tree`.

**Domain cleanup (2026-04-30):** Removed 6 unused domain types (TokenCount, FileCount, CloneCount, Threshold, BytePosition) and 15 dead error variables. domain/ now contains only production-used types: Filepath, LineNumber, CloneSeverity (consumed by detection/todos.go).

**Config builder safety (2026-04-30):** Replaced `panic(err)` in `cmd/config_builder.go` with proper error returns for `applyTimeoutFlag` and `applyDiffModeFlag`.

**Architecture enforcement (2026-04-30):** Replaced ghost `.go-arch-lint.yml` with project-specific config. Domain must not import syntax, suffixtree must have zero deps. Consolidated `StatsPrinter`'s 8 setters into single `ApplyStatsConfig(StatsConfig)`.

**Config extraction & printer cleanup (2026-05-03):** Deleted `printer/format.go` (moved `ParseFormat` to `config.ParseOutputFormat`) and `printer/sort_type.go` (moved `SortBy` constants to `config.SortCriteria`). All printer and cmd references updated to use `config.SortCriteria` directly. Printer interface now uses `config.SortCriteria` instead of `printer.SortBy`. Extracted `DetectionConfig` from `config.Config` for the detection layer — `MultiDetector` accepts `DetectionConfig` (Methods + Verbose) instead of the full `*config.Config`. Defined `MethodDetector` interface in `detection/detector.go` for pluggable detectors.

**Detector wiring & file splits (2026-05-03):** Wired `TodoDetector` and `LegacyDetector` through `MultiDetector.FindDuplOver()` — all 4 detection methods now accessible via `-m` flag. Consolidated threshold error sentinels into `config/enum_helpers.go` as single source of truth. Split `detection/todos.go` (352L → 3 files), `config/config.go` (344L → 3 files), `cmd/run_analysis.go` (450L → 3 files).

### Architecture — Outstanding Issues

**Printer ↔ syntax.Node coupling:** `Printer.PrintClones(dups [][]*syntax.Node)` forces all 6 implementations to depend on AST internals. Each printer independently calls `ProcessNodeRange()` and `extractContent()`. Fix: introduce ProcessedClone DTO, change Printer interface to accept `[]ProcessedCloneGroup`. This touches 111 test call sites — defer to dedicated PR.

**Three parallel Clone types:**

- `printer.clone` (unexported): has fragment, classification — richest for output
- `pkg/artdupl.Clone`: SDK type with primitives, `IsValid()` validation
- `printer.CloneGroup` vs `pkg/artdupl.CloneGroup`: different JSON shapes for different consumers

Consolidation depends on Printer DTO change above.

**printer/clone_classify.go imports syntax/golang directly:** Language-specific node type constants mapped to categories. Coupling breaks when supporting non-Go languages. Moves naturally with Printer DTO refactor.
