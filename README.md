# art-dupl

[![Go Report Card](https://goreportcard.com/badge/github.com/LarsArtmann/art-dupl)](https://goreportcard.com/report/github.com/LarsArtmann/art-dupl) [![codecov](https://codecov.io/gh/LarsArtmann/art-dupl/graph/badge.svg?token=art-dupl)](https://codecov.io/gh/LarsArtmann/art-dupl)

**art-dupl** is a code duplication detection tool for Go projects. It analyzes source code at the AST level to find structural clones — ignoring literal values so it focuses on patterns, not string contents.

A fork of [mibk/dupl](https://github.com/mibk/dupl) (via [golangci/dupl](https://github.com/golangci/dupl)) with major enhancements: multi-method detection, semantic-aware matching, 7 output formats, templ support, incremental analysis, and a professional CLI.

## Why art-dupl?

| Capability               | Original dupl    | art-dupl                                                  |
| ------------------------ | ---------------- | --------------------------------------------------------- |
| Detection algorithm      | Suffix tree only | Suffix tree + hash-based + multi-method                   |
| Semantic matching        | No               | Yes — distinguishes identifiers by name                   |
| Output formats           | Text, HTML       | Text, HTML, JSON, Simple-JSON, SARIF, plumbing, rich-text |
| Stats subcommand         | No               | Yes — with health grades, CSV, JSON                       |
| Templ support            | No               | Full `.templ` AST analysis                                |
| Generated code filtering | No               | Auto-detects sqlc, protobuf, mockgen, stringer            |
| Incremental analysis     | No               | AST caching with git-aware mode                           |
| Diff visualization       | No               | Side-by-side and inline diffs in HTML                     |
| Parallel parsing         | No               | Worker pool with auto-detect                              |

## Installation

```bash
go install github.com/LarsArtmann/art-dupl/cmd/art-dupl@latest
```

Or build from source:

```bash
git clone https://github.com/LarsArtmann/art-dupl.git
cd art-dupl
nix build     # Builds with Go 1.26, static binary
nix run .#art-dupl -- ./src
```

## Quick Start

```bash
# Scan current directory (default: threshold 15 tokens)
art-dupl

# Larger clones only (50+ tokens)
art-dupl -t 50

# HTML report with diff visualization
art-dupl --html --diff side-by-side > report.html

# JSON for CI/CD pipelines
art-dupl --json -t 20

# Structural-only matching (shows more results, including noise)
art-dupl --structural ./src

# Hash-based detection (faster)
art-dupl -m hash ./src

# Both detection methods combined
art-dupl -m "hash,art-dupl" ./src

# All output formats to a directory
art-dupl --all --output-dir ./reports ./src

# Project statistics
art-dupl stats
art-dupl stats --format json
```

## Supported Languages

| Language | Extension | Notes                                       |
| -------- | --------- | ------------------------------------------- |
| Go       | `.go`     | Full AST analysis                           |
| Templ    | `.templ`  | Full AST analysis via official templ parser |

`.templ` source files are included by default. Templ-generated `*_templ.go` files are filtered by default (use `--include-templ` to include them). Use `--only go` or `--only templ` to restrict analysis to a single file type.

## Detection Methods

Controlled with `-m` / `--detection-methods`:

| Method      | Flag                    | Description                                                      |
| ----------- | ----------------------- | ---------------------------------------------------------------- |
| Suffix tree | `-m art-dupl` (default) | Ukkonen's algorithm on serialized ASTs — finds structural clones |
| Hash-based  | `-m hash`               | Rolling hash on file content — faster, different tradeoffs       |
| Combined    | `-m "hash,art-dupl"`    | Runs both in parallel, deduplicates results                      |

### Semantic vs. Structural Matching

- **Default** (semantic): Matches by structure AND identifier names. `a.String()` does NOT match `b.Error()`. Distinguishes methods by receiver type.
- **`--structural`**: Matches by AST shape only. `a.String()` matches `b.Error()`. Use for raw structural analysis.

## Output Formats

| Format      | Flag                | Use case                                                        |
| ----------- | ------------------- | --------------------------------------------------------------- |
| Text        | _(default)_         | Quick terminal review                                           |
| Rich text   | `--rich-text`       | Enhanced text with priority badges and suggestions              |
| HTML        | `--html`            | Detailed report with syntax highlighting and diff visualization |
| JSON        | `--json` / `-j`     | CI/CD integration with metadata and summary                     |
| Simple JSON | `--simple-json`     | Lightweight JSON with impact scores                             |
| SARIF       | `--sarif`           | GitHub Advanced Security / CodeQL integration                   |
| Plumbing    | `--plumbing` / `-p` | Machine-readable `file:startLine-endLine` for scripts           |

Sort results with `-s` / `--sort`: `size` (default), `occurrence`, `hash`, `total-tokens`.

## CLI Reference

```bash
art-dupl [flags] [paths...]
art-dupl stats [flags] [paths...]
art-dupl completion [bash|zsh|fish|powershell]
```

### Core Flags

| Flag                  | Short | Default    | Description                                       |
| --------------------- | ----- | ---------- | ------------------------------------------------- |
| `--threshold`         | `-t`  | `15`       | Minimum token sequence size                       |
| `--detection-methods` | `-m`  | `art-dupl` | Detection method(s)                               |
| `--sort`              | `-s`  | `size`     | Sort: size, occurrence, hash, total-tokens        |
| `--semantic`          |       | `true`     | Match by structure AND identifier names (default) |
| `--config`            | `-c`  |            | Path to JSON config file                          |
| `--verbose`           | `-v`  |            | Verbose logging (repeat for more)                 |
| `--version`           |       |            | Show version                                      |

### Output Flags

| Flag            | Short | Description                                             |
| --------------- | ----- | ------------------------------------------------------- |
| `--html`        |       | HTML report with syntax highlighting                    |
| `--json`        | `-j`  | Structured JSON with statistics                         |
| `--simple-json` |       | Lightweight JSON with impact scores                     |
| `--sarif`       |       | SARIF 2.1.0 format                                      |
| `--plumbing`    | `-p`  | Machine-readable format                                 |
| `--rich-text`   |       | Enhanced text with badges                               |
| `--all`         | `-a`  | Generate all formats                                    |
| `--output-dir`  | `-o`  | Output directory (with `--all`)                         |
| `--diff`        |       | Diff visualization: `side-by-side`, `inline`, or `true` |

### Filtering Flags

| Flag                     | Default | Description                                                |
| ------------------------ | ------- | ---------------------------------------------------------- |
| `--vendor`               | `false` | Include vendor directory                                   |
| `--include-node-modules` | `false` | Include node_modules (hash detection)                      |
| `--include-sqlc`         | `false` | Include sqlc-generated files (filtered by default)         |
| `--include-templ`        | `false` | Include `*_templ.go` generated files (filtered by default) |
| `--include-protobuf`     | `false` | Include `.pb.go` files (filtered by default)               |
| `--include-mockgen`      | `false` | Include mockgen-generated files (filtered by default)      |
| `--include-stringer`     | `false` | Include stringer-generated files (filtered by default)     |
| `--include-pattern`      | `[]`    | Glob patterns to include (overrides filters)               |
| `--exclude-pattern`      | `[]`    | Glob patterns to exclude                                   |
| `--only`                 | `""`    | Restrict to `go` or `templ` file type                      |
| `--files`                | `-f`    | Read file paths from stdin                                 |

### Performance Flags

| Flag            | Default | Description                                           |
| --------------- | ------- | ----------------------------------------------------- |
| `--workers`     | `0`     | Concurrent workers (`0` = auto-detect CPU cores)      |
| `--incremental` | `false` | Enable AST caching for incremental analysis           |
| `--cache-dir`   | `""`    | Cache directory (default: `.cache/art-dupl`)          |
| `--clear-cache` | `false` | Clear cache before running                            |

### Stats Subcommand

```bash
art-dupl stats [flags] [paths...]
```

Supports all shared flags plus:

| Flag       | Default | Description                          |
| ---------- | ------- | ------------------------------------ |
| `--format` | `text`  | Output format: `text`, `json`, `csv` |

Statistics include files scanned, clone groups, duplicate lines/tokens, complexity score, impact score, clone size distribution, health grade (A–F), and top files by duplication.

## Configuration

Create `dupl.json`:

```json
{
  "threshold": 30,
  "outputFormat": "json",
  "paths": ["./src", "./lib"]
}
```

```bash
art-dupl -c dupl.json
```

CLI flags override config file values. Config file overrides defaults.

## Examples

```bash
# CI/CD: fail if too many clones
TOTAL=$(art-dupl --json . | jq '.summary.total_clones')
[ "$TOTAL" -gt 100 ] && echo "Too many clones: $TOTAL" && exit 1

# Find test file duplicates
find . -name '*_test.go' | art-dupl -files

# Progressive refactoring: baseline → refactor → measure
art-dupl --json -t 15 > baseline.json
# ... refactor ...
art-dupl --json -t 15 > after.json

# Most widespread clones first
art-dupl --plumbing --sort occurrence ./src

# Filter all generated code (default behavior)
art-dupl ./src

# Override: include sqlc-generated files
art-dupl --include-sqlc ./src
```

## Architecture

```
Source Files (.go, .templ)
    ↓
Parsing (go/parser + templ parser)
    ↓
AST Serialization → unified syntax.Node[]
    ↓
Suffix Tree (Ukkonen's algorithm, O(1) transitions)
    ↓
Duplicate Search → syntax.Match
    ↓
Clone Classification → domain.ProcessedCloneGroup
    ↓
Output Formatting (text, HTML, JSON, SARIF, plumbing)
```

### Key Packages

| Package        | Purpose                                                                   |
| -------------- | ------------------------------------------------------------------------- |
| `suffixtree/`  | Ukkonen's suffix tree with O(1) map-based transitions                     |
| `syntax/`      | AST handling, Go and Templ parsers, serialization                         |
| `detection/`   | Multi-method detection coordination via goroutines                        |
| `hash/`        | XXH3 rolling hash-based detection                                         |
| `job/`         | File parsing pipeline with parallel workers                               |
| `printer/`     | 7 output formats with sorting and classification                          |
| `domain/`      | Domain types: `Filepath`, `LineNumber`, `CloneSeverity`, `ProcessedClone` |
| `config/`      | Multi-source configuration with typed enums                               |
| `errors/`      | 11 typed error categories with stack traces                               |
| `pkg/artdupl/` | Public SDK with `Detector` interface for programmatic use                 |

## Development

```bash
just build          # Build binary
just test           # Run all tests
just check          # Run linting
just ci             # Format + lint + test
just coverage       # Generate coverage report
just bench          # Run benchmarks
just test-race      # Run with race detector
just check-coverage # Verify ≥80% coverage
```

Tests use standard `testing` + Ginkgo/Gomega BDD. See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[MIT](LICENSE) — Based on [mibk/dupl](https://github.com/mibk/dupl) (via [golangci/dupl](https://github.com/golangci/dupl)).
