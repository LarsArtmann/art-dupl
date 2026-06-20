# art-dupl Feature Documentation

> **Last Updated:** 2026-06-20
> **Version:** Analysis of fork branch

## Overview

**art-dupl** is a Go tool for finding code clones using suffix tree algorithms and hash-based detection. It analyzes abstract syntax trees (ASTs) to find structural code clones while ignoring literal values. Supports multi-method detection, professional CLI (Fang/Cobra), and 7 output formats.

---

## 🚀 Core Features

### Supported Languages

| Language  | Extension | Status           | Notes                                        |
| --------- | --------- | ---------------- | -------------------------------------------- |
| **Go**    | `.go`     | FULLY_FUNCTIONAL | Full AST analysis, 45 node types             |
| **Templ** | `.templ`  | FULLY_FUNCTIONAL | Pure Go parser, 28 node types, ON by default |

### Detection Methods

| Feature                              | Status           | Description                                                             |
| ------------------------------------ | ---------------- | ----------------------------------------------------------------------- |
| **Suffix Tree Detection (art-dupl)** | FULLY_FUNCTIONAL | Ukkonen's algorithm on serialized ASTs, O(1) map-based transitions      |
| **Hash-Based Detection**             | FULLY_FUNCTIONAL | XXH3 streaming hash (~20x faster than SHA-256), content-addressed dedup |
| **Multi-Detection Mode**             | FULLY_FUNCTIONAL | Run both methods simultaneously via goroutines, results deduplicated    |

### Output Formats

| Feature                | Status           | Description                                                            |
| ---------------------- | ---------------- | ---------------------------------------------------------------------- |
| **Text Output**        | FULLY_FUNCTIONAL | Human-readable clone listing with file paths, line numbers, diff hints |
| **HTML Output**        | FULLY_FUNCTIONAL | Dark theme, syntax highlighting, VSCode links, diff visualization      |
| **JSON Output**        | FULLY_FUNCTIONAL | Structured data with version, timestamp, clone_groups, summary         |
| **Simple-JSON Output** | FULLY_FUNCTIONAL | Simpler JSON format with score=impact, instances with token_count      |
| **Plumbing Output**    | FULLY_FUNCTIONAL | Machine-readable `file:startLine-endLine` format for CI/CD             |
| **SARIF Output**       | FULLY_FUNCTIONAL | SARIF 2.1.0 for GitHub Advanced Security, clones reported as results   |
| **CSV Output**         | FULLY_FUNCTIONAL | Stats CSV uses `encoding/csv` for proper escaping and quoting          |

### Batch & Report Generation

| Feature                         | Status           | Description                                                               |
| ------------------------------- | ---------------- | ------------------------------------------------------------------------- |
| **All Formats (--all)**         | FULLY_FUNCTIONAL | Generate all output formats for all detection methods at once             |
| **Custom Output Directory**     | FULLY_FUNCTIONAL | `--output-dir` specifies destination for batch generation                 |
| **Output File (--output-file)** | FULLY_FUNCTIONAL | Write stats output to file instead of stdout                              |
| **HTML Diff Visualization**     | FULLY_FUNCTIONAL | Side-by-side or inline diff via LCS algorithm, configurable with `--diff` |
| **Rich Text Output**            | FULLY_FUNCTIONAL | `--rich-text` adds priority/category/actionability badges to text output  |

### Statistics Subcommand

| Feature                     | Status           | Description                                                                                                                                                                                                    |
| --------------------------- | ---------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Text Stats**              | FULLY_FUNCTIONAL | Colored summary via lipgloss (default)                                                                                                                                                                         |
| **JSON Stats**              | FULLY_FUNCTIONAL | Structured statistics for CI/CD integration                                                                                                                                                                    |
| **CSV Stats**               | FULLY_FUNCTIONAL | Spreadsheet-compatible format using `encoding/csv`                                                                                                                                                             |
| **Health Grade**            | FULLY_FUNCTIONAL | A-F health grade (`domain.HealthScore`) with validation                                                                                                                                                        |
| **Clone Metrics**           | FULLY_FUNCTIONAL | Total clones, groups, files affected, duplication %                                                                                                                                                            |
| **Spread Analysis**         | FULLY_FUNCTIONAL | Complexity scores, severity distributions                                                                                                                                                                      |
| **Actionability Class.**    | FULLY_FUNCTIONAL | AST-based detection of 10 non-actionable patterns (signature-only, interface-impl, RAII defer, error-propagation, test-data, table-driven, test-scaffolding, data-dominated, describe-table, builder-callback) |
| **Clone Classification**    | FULLY_FUNCTIONAL | 14 categories (function, method, test, struct, etc.), 4 priority levels                                                                                                                                        |
| **Refactoring Suggestions** | FULLY_FUNCTIONAL | 20 suggestion constants mapped from category + actionability pattern (`printer/clone_classify.go`)                                                                                                             |
| **Stats Recommendations**   | FULLY_FUNCTIONAL | Grade-specific (A-F) actionable next steps in stats output                                                                                                                                                     |
| **Stats Visualizations**    | FULLY_FUNCTIONAL | ASCII bar charts for size/token distribution in text stats                                                                                                                                                     |
| **Filter Breakdown**        | FULLY_FUNCTIONAL | Reports files filtered by each category (sqlc, templ, etc.) in stats                                                                                                                                           |

### Sorting Options

| Feature                  | Status           | Description                                                  |
| ------------------------ | ---------------- | ------------------------------------------------------------ |
| **Size Sorting**         | FULLY_FUNCTIONAL | Sort clone groups by token count (largest first) — default   |
| **Occurrence Sorting**   | FULLY_FUNCTIONAL | Sort clone groups by number of files (most widespread first) |
| **Hash Sorting**         | FULLY_FUNCTIONAL | Sort clone groups by hash value (alphabetical)               |
| **Total Tokens Sorting** | FULLY_FUNCTIONAL | Sort clone groups by total token count (highest total first) |

---

## 🔍 Semantic Detection

| Feature                  | Status           | Description                                                                          |
| ------------------------ | ---------------- | ------------------------------------------------------------------------------------ |
| **Semantic Mode**        | FULLY_FUNCTIONAL | Alpha-normalizes local identifiers; detects Type 2 (renamed) clones (default)        |
| **Exact Mode**           | FULLY_FUNCTIONAL | `--exact` matches identifier names verbatim (copy-paste / Type 1 only)               |
| **Structural-Only Mode** | FULLY_FUNCTIONAL | `--structural` ignores all names; matches by AST shape only                          |
| **Clone Type Classification** | FULLY_FUNCTIONAL | Labels each clone type-1/2/3 in JSON, SARIF, and `--rich-text` output          |
| **Mutual Exclusion**     | FULLY_FUNCTIONAL | `--semantic` / `--exact` / `--structural` are mutually exclusive                     |

**Note:** Default is semantic mode (alpha-normalized). Use `--exact` for verbatim name matching or `--structural` for shape-only analysis.

---

## 🛡️ Smart Filtering

| Feature                       | Status           | Description                                                                 |
| ----------------------------- | ---------------- | --------------------------------------------------------------------------- |
| **SQLC Detection**            | FULLY_FUNCTIONAL | Auto-detects `sqlc.yaml` in parent dirs, filters generated `.go`            |
| **Templ Filtering**           | FULLY_FUNCTIONAL | `.templ` source files included by default; `*_templ.go` filtered by default |
| **Protobuf Filtering**        | FULLY_FUNCTIONAL | Filters `.pb.go`, `_grpc.pb.go` files                                       |
| **Mockgen Filtering**         | FULLY_FUNCTIONAL | Filters mockgen generated files                                             |
| **Stringer Filtering**        | FULLY_FUNCTIONAL | Filters stringer generated files                                            |
| **Include Overrides**         | FULLY_FUNCTIONAL | `--include-sqlc/templ/protobuf/mockgen/stringer/generic` to override        |
| **Custom Include/Exclude**    | FULLY_FUNCTIONAL | `--include-pattern` / `--exclude-pattern` glob patterns                     |
| **Directory Exclusions**      | FULLY_FUNCTIONAL | `vendor/`, `.git/`, `node_modules/` excluded by default                     |
| **Node Modules**              | FULLY_FUNCTIONAL | `--include-node-modules` includes for hash detection                        |
| **File Type Filter (--only)** | FULLY_FUNCTIONAL | Restrict to `go` or `templ` file types                                      |

---

## ⚡ Performance & Concurrency

| Feature                   | Status           | Description                                                                                                                                         |
| ------------------------- | ---------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Parallel Parsing**      | FULLY_FUNCTIONAL | Worker pool via `--workers` flag (0=auto, NumCPU)                                                                                                   |
| **Incremental Analysis**  | FULLY_FUNCTIONAL | SHA1 content-hash AST caching, `--incremental` flag                                                                                                 |
| **Git-Aware Incremental** | REMOVED          | `--since` flag removed (was a dead stub, never read). Only content-hash caching via `--incremental` works. Git-diff file selection not implemented. |
| **Cache Management**      | FULLY_FUNCTIONAL | `--cache-dir`, `--clear-cache`, file-based gob serialization                                                                                        |
| **Execution Timeout**     | FULLY_FUNCTIONAL | `--timeout` with context cancellation (default 30m)                                                                                                 |
| **Performance Profiling** | EXPERIMENTAL     | Hidden `--profile` flag; pprof CPU/mem profile capture                                                                                              |
| **SIMD Optimizations**    | N/A              | `syntax/hash_seq.go` (renamed from `hash_simd.go`) uses sync.Pool + xxh3 — no hand-written SIMD                                                     |

---

## 🖥️ Professional CLI (via Fang)

| Feature                    | Status           | Description                                                 |
| -------------------------- | ---------------- | ----------------------------------------------------------- |
| **Styled Help Output**     | FULLY_FUNCTIONAL | Rich, themed help text via Fang framework                   |
| **Shell Completion**       | FULLY_FUNCTIONAL | bash, zsh, fish, PowerShell with `--no-descriptions` option |
| **Man Page Generation**    | FULLY_FUNCTIONAL | `art-dupl man` generates manual pages                       |
| **Version Information**    | FULLY_FUNCTIONAL | Version, commit, build date                                 |
| **Configurable Verbosity** | FULLY_FUNCTIONAL | `-v` (verbose), `-vv` (extra verbose)                       |

---

## 🔧 Configuration

| Feature                      | Status           | Description                                             |
| ---------------------------- | ---------------- | ------------------------------------------------------- |
| **Command-Line Flags**       | FULLY_FUNCTIONAL | 35+ flags for full control                              |
| **JSON Configuration Files** | FULLY_FUNCTIONAL | `--config` / `-c` flag, JSON-tagged Config struct       |
| **Configuration Merging**    | FULLY_FUNCTIONAL | CLI flags override file config, file overrides defaults |
| **Threshold Control**        | FULLY_FUNCTIONAL | Adjustable minimum token sequence size (default: 15)    |
| **Vendor Directory Control** | FULLY_FUNCTIONAL | `--vendor` to include vendor directory                  |
| **File Input from Stdin**    | FULLY_FUNCTIONAL | `--files` / `-f` reads file paths from stdin            |

---

## 📦 SDK / Programmatic API

| Feature                | Status           | Description                                                 |
| ---------------------- | ---------------- | ----------------------------------------------------------- |
| **Detector Interface** | FULLY_FUNCTIONAL | `FindClones()`, `FindClonesStreamResult()`, `Close()`       |
| **Options Builder**    | FULLY_FUNCTIONAL | Full configuration via `Options` struct                     |
| **Progress Callbacks** | FULLY_FUNCTIONAL | Stage, completed, total, percentage, current file           |
| **Result Type**        | FULLY_FUNCTIONAL | CloneGroups, Summary, Metadata                              |
| **Validation**         | FULLY_FUNCTIONAL | 20+ sentinel errors, `ValidateOptions()`, `Clone.IsValid()` |
| **Custom FileReader**  | FULLY_FUNCTIONAL | Injectable file reader for testing/custom sources           |

---

## 🧪 Testing

| Feature               | Status           | Description                                      |
| --------------------- | ---------------- | ------------------------------------------------ |
| **Unit Tests**        | FULLY_FUNCTIONAL | Standard `testing` package, table-driven tests   |
| **BDD Tests**         | FULLY_FUNCTIONAL | Ginkgo/Gomega in `bdd/` directory, user-focused  |
| **Integration Tests** | FULLY_FUNCTIONAL | CLI integration tests in `cmd/`                  |
| **Golden File Tests** | FULLY_FUNCTIONAL | HTML output golden testing                       |
| **Fuzz Tests**        | FULLY_FUNCTIONAL | Property-based testing in `fuzz/` directory      |
| **Benchmarks**        | FULLY_FUNCTIONAL | `_bench_test.go` files with allocation reporting |
| **Race Detection**    | FULLY_FUNCTIONAL | All tests pass with `-race` flag                 |

---

## 🏗️ Architecture Components

| Component        | Status           | Description                                                 |
| ---------------- | ---------------- | ----------------------------------------------------------- |
| **suffixtree/**  | FULLY_FUNCTIONAL | Core Ukkonen's suffix tree, O(1) map transitions            |
| **syntax/**      | FULLY_FUNCTIONAL | AST handling, Go + Templ parsers, serialization             |
| **job/**         | FULLY_FUNCTIONAL | Parsing pipeline, parallel workers, incremental             |
| **printer/**     | FULLY_FUNCTIONAL | 7 output formats, sorting, classification, templ-based HTML |
| **hash/**        | FULLY_FUNCTIONAL | XXH3 streaming hash detection                               |
| **config/**      | FULLY_FUNCTIONAL | Multi-source config with validation                         |
| **detection/**   | FULLY_FUNCTIONAL | Multi-detector coordination via goroutines                  |
| **cache/**       | FULLY_FUNCTIONAL | File-based AST caching with SHA1 keys                       |
| **domain/**      | FULLY_FUNCTIONAL | Value objects: ProcessedClone, enums, validation sentinels  |
| **errors/**      | FULLY_FUNCTIONAL | 6 error types, typed wrapping, stack traces                 |
| **pkg/artdupl/** | FULLY_FUNCTIONAL | Public SDK with Detector interface, comprehensive godoc     |

---

## 🔮 Experimental / In Progress

| Feature                   | Status       | Description                                  |
| ------------------------- | ------------ | -------------------------------------------- |
| **Performance Profiling** | EXPERIMENTAL | `--profile` flag exists, pprof capture works |

---

## 🚫 Known Limitations

| Limitation                  | Impact   | Description                                                                                           |
| --------------------------- | -------- | ----------------------------------------------------------------------------------------------------- |
| **Go & Templ Only**         | High     | Only `.go` and `.templ` files supported                                                               |
| **No Git-Diff Incremental** | Low      | Only content-hash caching (`--incremental`); git-diff file selection not implemented                  |
| **SDK Stream Errors**       | Resolved | `FindClonesStream` removed; `FindClonesStreamResult` propagates errors via `StreamResult{Group, Err}` |

---

## 📋 Quick Reference

### Detection Methods

```bash
art-dupl                       # Default: suffix tree
art-dupl -m hash               # Hash-based (faster)
art-dupl -m "hash,art-dupl"    # Both methods
```

### Output Formats

```bash
art-dupl                       # Text (default)
art-dupl --html                # HTML report
art-dupl --json                # JSON with statistics
art-dupl --plumbing            # Machine-readable
art-dupl --sarif               # SARIF 2.1.0
art-dupl --all -o ./reports    # All formats to directory
```

### Filtering

```bash
art-dupl --include-sqlc       # Include sqlc-generated files (filtered by default)
art-dupl --only go             # Go files only
art-dupl --include-pattern "gen/*" --exclude-pattern "mock_*"
```

### Performance

```bash
art-dupl --workers 8           # 8 parallel workers
art-dupl --incremental         # AST caching
art-dupl --since HEAD~10       # Git-aware incremental
art-dupl --cache-dir /tmp/cache
art-dupl --clear-cache
```

### Stats

```bash
art-dupl stats                 # Text statistics
art-dupl stats --format json   # JSON statistics
art-dupl stats --format csv    # CSV for spreadsheets
art-dupl stats --output-file report.txt  # Write to file
```
