# art-dupl

[![Go Report Card](https://goreportcard.com/badge/github.com/LarsArtmann/art-dupl)](https://goreportcard.com/report/github.com/LarsArtmann/art-dupl)
[![CI](https://github.com/LarsArtmann/art-dupl/actions/workflows/ci.yml/badge.svg)](https://github.com/LarsArtmann/art-dupl/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-e8a020.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/LarsArtmann/art-dupl.svg)](https://pkg.go.dev/github.com/LarsArtmann/art-dupl)
[![Website](https://img.shields.io/badge/website-art--dupl.web.app-e8a020.svg)](https://art-dupl.web.app)

**Professional code clone detection for Go.** Analyzes source code at the AST level to find structural and semantic clones — ignoring literal values so it focuses on patterns, not string contents.

A fork of [mibk/dupl](https://github.com/mibk/dupl) (via [golangci/dupl](https://github.com/golangci/dupl)) with major enhancements: multi-method detection, three matching modes (semantic / exact / structural), 7 output formats, `.templ` support, generated-code filtering, incremental analysis, baseline CI gating, and a professional CLI.

---

## Quick Start

```bash
go install github.com/LarsArtmann/art-dupl/cmd/art-dupl@latest
```

```bash
art-dupl                          # Scan current directory (threshold: 5 statements)
art-dupl -t 15                    # Larger clones only
art-dupl --html > report.html     # HTML report with syntax highlighting
art-dupl --json -t 10             # JSON for CI/CD
art-dupl stats                    # Project health statistics
```

Full documentation: **[art-dupl.web.app](https://art-dupl.web.app)**

---

## Why art-dupl?

| Capability               | Original dupl    | art-dupl                                                                     |
| ------------------------ | ---------------- | ---------------------------------------------------------------------------- |
| Detection algorithm      | Suffix tree only | Suffix tree + hash-based + multi-method                                      |
| Semantic matching        | No               | Yes — alpha-normalization finds renamed clones (Type 2)                      |
| Matching modes           | 1                | 3 — semantic (default), exact, structural                                    |
| Output formats           | Text, HTML       | Text, Rich-text, HTML, JSON, Simple-JSON, SARIF, plumbing                    |
| Stats subcommand         | No               | Yes — health grades (A–F), CSV, JSON, recommendations                        |
| Templ support            | No               | Full `.templ` AST analysis                                                   |
| Generated code filtering | No               | Auto-detects sqlc, protobuf, mockgen, stringer, templ, generic               |
| Incremental analysis     | No               | AST caching with SHA-256 content hashing                                     |
| CI baseline gating       | No               | `baseline` + `check` subcommands                                             |
| Diff visualization       | No               | Side-by-side and inline diffs in HTML                                        |
| Clone classification     | No               | Type 1 / 2 / 3 labels + extractability scores                                |
| Actionability filtering  | No               | Suppresses 12+ boilerplate patterns (test scaffolding, error wrapping, etc.) |
| Parallel parsing         | No               | Worker pool with auto-detect                                                 |
| Programmatic SDK         | No               | `pkg/artdupl` — Detector interface with streaming                            |

---

## Detection Methods

Controlled with `-m` / `--detection-methods`:

| Method      | Flag                    | Description                                                      |
| ----------- | ----------------------- | ---------------------------------------------------------------- |
| Suffix tree | `-m art-dupl` (default) | Ukkonen's algorithm on serialized ASTs — finds structural clones |
| Hash-based  | `-m hash`               | Rolling hash on file content — faster, different tradeoffs       |
| Combined    | `-m "hash,art-dupl"`    | Runs both in parallel, deduplicates results                      |

### Three Matching Modes

| Mode                   | Flag           | What it detects                                                          | Example                 |
| ---------------------- | -------------- | ------------------------------------------------------------------------ | ----------------------- |
| **Semantic** (default) | `--semantic`   | Structure AND identifier names. `a.String()` does NOT match `b.Error()`. | Type 2 renamed clones   |
| Exact                  | `--exact`      | Verbatim name matching. Copy-paste only.                                 | Type 1 exact clones     |
| Structural             | `--structural` | AST shape only. `a.String()` matches `b.Error()`.                        | Raw structural patterns |

---

## Output Formats

| Format      | Flag                | Use case                                              |
| ----------- | ------------------- | ----------------------------------------------------- |
| Text        | _(default)_         | Quick terminal review                                 |
| Rich text   | `--rich-text`       | Priority badges and refactoring suggestions           |
| HTML        | `--html`            | Detailed report with syntax highlighting and diffs    |
| JSON        | `--json` / `-j`     | CI/CD integration with metadata and summary           |
| Simple JSON | `--simple-json`     | Lightweight JSON with impact scores                   |
| SARIF       | `--sarif`           | GitHub Advanced Security / CodeQL integration         |
| Plumbing    | `--plumbing` / `-p` | Machine-readable `file:startLine-endLine` for scripts |

Generate all formats at once: `art-dupl --all --output-dir ./reports`

Sort results: `--sort size` (default), `occurrence`, `hash`, `total-tokens`

---

## CI/CD Integration

```bash
# Record accepted clones (commit the baseline file)
art-dupl baseline . -t 5

# CI gate: exits 1 if NEW clones are found
art-dupl check . -t 5
```

**GitHub Actions template** included at [`templates/github-actions-duplicate-check.yml`](templates/github-actions-duplicate-check.yml).

**Pre-commit hook** template at [`templates/pre-commit-hook.yaml`](templates/pre-commit-hook.yaml).

**SARIF upload** to GitHub Security tab:

```bash
art-dupl --sarif -t 10 > results.sarif
# Upload via github/codeql-action/upload-sarif
```

---

## Smart Filtering

Generated code is filtered by default. Override with `--include-generated`:

```bash
art-dupl                                    # Default: filters all generated code
art-dupl --include-generated sqlc           # Include sqlc-generated files
art-dupl --include-generated templ,protobuf # Include multiple categories
art-dupl --include-generated all            # Include everything
```

Detected categories: `sqlc`, `templ`, `protobuf`, `mockgen`, `stringer`, `generic` (any "Code generated by" comment).

Additional filtering:

```bash
art-dupl --only go                          # Analyze .go files only
art-dupl --only templ                       # Analyze .templ files only
art-dupl --exclude-pattern "*_test.go"      # Exclude test files
art-dupl --include-pattern "internal/*"     # Include specific paths
```

---

## Performance

```bash
art-dupl --workers 8          # Parallel parsing (0 = auto-detect CPU)
art-dupl --incremental        # AST caching for faster re-runs
art-dupl --timeout 10m        # Execution timeout
```

---

## Configuration

Create `dupl.json`:

```json
{
  "threshold": 10,
  "outputFormat": "json",
  "paths": ["./src", "./lib"]
}
```

```bash
art-dupl -c dupl.json         # CLI flags override config; config overrides defaults
```

---

## Programmatic SDK

```go
import "github.com/LarsArtmann/art-dupl/pkg/artdupl"

detector, err := artdupl.New(artdupl.Options{
    Threshold: 10,
    Paths:     []string{"./src"},
})
// ...
result, err := detector.FindClones(ctx)
```

The SDK has zero imports of `config/` or `errors/` — fully independent types. See [SDK Design](SDK_DESIGN.md) and [pkg.go.dev](https://pkg.go.dev/github.com/LarsArtmann/art-dupl/pkg/artdupl).

---

## Supported Languages

| Language | Extension | Notes                                                      |
| -------- | --------- | ---------------------------------------------------------- |
| Go       | `.go`     | Full AST analysis, 45 node types                           |
| Templ    | `.templ`  | Full AST analysis via official templ parser, 28 node types |

`.templ` source files included by default. Templ-generated `*_templ.go` files filtered by default.

---

## Architecture

```
Source Files (.go, .templ)
    ↓
Parsing (go/parser + templ parser)     ← parallel worker pool
    ↓
AST Serialization → unified syntax.Node[]
    ↓
Suffix Tree (Ukkonen's algorithm)      ← O(1) map-based transitions
    ↓
Duplicate Search → syntax.Match
    ↓
Clone Classification → domain.ProcessedCloneGroup
    ↓
Output Formatting (text, HTML, JSON, SARIF, plumbing, rich-text)
```

| Package        | Purpose                                                 |
| -------------- | ------------------------------------------------------- |
| `suffixtree/`  | Ukkonen's suffix tree with O(1) map-based transitions   |
| `syntax/`      | AST handling, Go and Templ parsers, alpha-normalization |
| `detection/`   | Multi-method detection coordination via goroutines      |
| `hash/`        | XXH3 rolling hash-based detection                       |
| `job/`         | File parsing pipeline with parallel workers             |
| `printer/`     | 7 output formats with sorting and classification        |
| `domain/`      | Domain types: `ProcessedClone`, enums, `Extractability` |
| `config/`      | Multi-source configuration with typed enums             |
| `baseline/`    | Baseline recording + CI check file format               |
| `cache/`       | File-based AST caching with SHA-256 content hashing     |
| `pkg/artdupl/` | Public SDK with `Detector` interface                    |

---

## Development

```bash
go build ./...                              # Build all packages
go test ./...                               # Run all tests
go test -race ./...                         # Run with race detector
golangci-lint run --timeout 5m ./...        # Lint
nix build                                   # Reproducible build (Go 1.26)
nix flake check                             # Full CI check
```

Tests use standard `testing` + Ginkgo/Gomega BDD. See [CONTRIBUTING.md](CONTRIBUTING.md) and [TESTING.md](TESTING.md).

---

## License

[MIT](LICENSE) — Based on [mibk/dupl](https://github.com/mibk/dupl) (via [golangci/dupl](https://github.com/golangci/dupl)).
