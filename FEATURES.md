# art-dupl Feature Documentation

> **Last Updated:** 2026-08-10
> **Version:** v0.6.1 (unreleased features pending v0.7.0)

## Overview

**art-dupl** is a Go tool for finding code clones using suffix tree algorithms and hash-based detection. It analyzes abstract syntax trees (ASTs) to find structural code clones while ignoring literal values. Supports multi-method detection, professional CLI (Fang/Cobra), 7 output formats, and 29 actionability patterns. Tuned on real-world Go projects (6,000+ Go files, 320+ templ files) to minimize false positives at the default threshold.

---

## 🚀 Core Features

### Supported Languages

| Language  | Extension | Status           | Notes                                        |
| --------- | --------- | ---------------- | -------------------------------------------- |
| **Go**    | `.go`     | FULLY_FUNCTIONAL | Full AST analysis, 50 node types             |
| **Templ** | `.templ`  | FULLY_FUNCTIONAL | Pure Go parser, semantic mode, ON by default |

### Detection Methods

| Feature                              | Status           | Description                                                             |
| ------------------------------------ | ---------------- | ----------------------------------------------------------------------- |
| **Suffix Tree Detection (art-dupl)** | FULLY_FUNCTIONAL | Ukkonen's algorithm on serialized ASTs, O(1) map-based transitions      |
| **Hash-Based Detection**             | FULLY_FUNCTIONAL | XXH3 streaming hash (~20x faster than SHA-256), content-addressed dedup |
| **Multi-Detection Mode**             | FULLY_FUNCTIONAL | Run both methods simultaneously via goroutines, results deduplicated    |

### Output Formats

| Feature                | Status           | Description                                                                                     |
| ---------------------- | ---------------- | ----------------------------------------------------------------------------------------------- |
| **Text Output**        | FULLY_FUNCTIONAL | Human-readable clone listing with file paths, line numbers, one-line source preview, diff hints |
| **HTML Output**        | FULLY_FUNCTIONAL | Dark theme, syntax highlighting, VSCode links, diff visualization                               |
| **JSON Output**        | FULLY_FUNCTIONAL | Structured data with version, timestamp, clone_groups, summary                                  |
| **Simple-JSON Output** | FULLY_FUNCTIONAL | Simpler JSON format with score=impact, instances with token_count                               |
| **Plumbing Output**    | FULLY_FUNCTIONAL | Machine-readable `file:startLine-endLine` format for CI/CD                                      |
| **SARIF Output**       | FULLY_FUNCTIONAL | SARIF 2.1.0 for GitHub Advanced Security, clones reported as results                            |
| **CSV Output**         | FULLY_FUNCTIONAL | Stats CSV uses `encoding/csv` for proper escaping and quoting                                   |

### Batch & Report Generation

| Feature                         | Status           | Description                                                               |
| ------------------------------- | ---------------- | ------------------------------------------------------------------------- |
| **All Formats (--all)**         | FULLY_FUNCTIONAL | Generate all output formats for all detection methods at once             |
| **Custom Output Directory**     | FULLY_FUNCTIONAL | `--output-dir` specifies destination for batch generation                 |
| **Output File (--output-file)** | FULLY_FUNCTIONAL | Write stats output to file instead of stdout                              |
| **HTML Diff Visualization**     | FULLY_FUNCTIONAL | Side-by-side or inline diff via LCS algorithm, configurable with `--diff` |
| **Rich Text Output**            | FULLY_FUNCTIONAL | `--rich-text` adds priority/category/actionability badges to text output  |

### Statistics Subcommand

| Feature                     | Status           | Description                                                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| --------------------------- | ---------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Text Stats**              | FULLY_FUNCTIONAL | Colored summary via lipgloss (default)                                                                                                                                                                                                                                                                                                                                                                                                                                                                |
| **JSON Stats**              | FULLY_FUNCTIONAL | Structured statistics for CI/CD integration                                                                                                                                                                                                                                                                                                                                                                                                                                                           |
| **CSV Stats**               | FULLY_FUNCTIONAL | Spreadsheet-compatible format using `encoding/csv`                                                                                                                                                                                                                                                                                                                                                                                                                                                    |
| **Health Grade**            | FULLY_FUNCTIONAL | A-F health grade (`domain.HealthScore`) with validation                                                                                                                                                                                                                                                                                                                                                                                                                                               |
| **Clone Metrics**           | FULLY_FUNCTIONAL | Total clones, groups, files affected, duplication %                                                                                                                                                                                                                                                                                                                                                                                                                                                   |
| **Spread Analysis**         | FULLY_FUNCTIONAL | Complexity scores, severity distributions                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| **Actionability Class.**    | FULLY_FUNCTIONAL | AST-based detection of 29 non-actionable patterns (signature-only, interface-implementation, interface-method, RAII defer, error-propagation, error-wrapping, assertion-chain, cobra-boilerplate, testdata-pair, table-driven-test, test-scaffolding, data-dominated, describe-table, builder-callback, assign-error-check, single-call-expression, guard-clause, single-simple-statement, bool-accumulator-initializer, single-declaration, test-helper-delegate, bool-guard, templ-rendering-idiom, interface-assertion, type-alias-block, defer-call, test-framework-call, state-flag-mutation, empty-default) |
| **Clone Classification**    | FULLY_FUNCTIONAL | 17 categories (function, method, test, struct, interface, handler, loop, conditional, test-boilerplate, test-fixture, assignment, expression, block, call, return, defer, unknown), 4 priority levels                                                                                                                                                                                                                                                                                                 |
| **Refactoring Suggestions** | FULLY_FUNCTIONAL | Category + actionability pattern based suggestions (`printer/clone_classify.go::getSuggestion`)                                                                                                                                                                                                                                                                                                                                                                                                       |
| **Stats Recommendations**   | FULLY_FUNCTIONAL | Grade-specific (A-F) actionable next steps in stats output                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| **Stats Visualizations**    | FULLY_FUNCTIONAL | ASCII bar charts for size/token distribution in text stats                                                                                                                                                                                                                                                                                                                                                                                                                                            |
| **Filter Breakdown**        | FULLY_FUNCTIONAL | Reports files filtered by each category (sqlc, templ, etc.) in stats                                                                                                                                                                                                                                                                                                                                                                                                                                  |

### Sorting Options

| Feature                  | Status           | Description                                                  |
| ------------------------ | ---------------- | ------------------------------------------------------------ |
| **Size Sorting**         | FULLY_FUNCTIONAL | Sort clone groups by token count (largest first; default)    |
| **Occurrence Sorting**   | FULLY_FUNCTIONAL | Sort clone groups by number of files (most widespread first) |
| **Hash Sorting**         | FULLY_FUNCTIONAL | Sort clone groups by hash value (alphabetical)               |
| **Total Tokens Sorting** | FULLY_FUNCTIONAL | Sort clone groups by total token count (highest total first) |

### CI Integration

| Feature                           | Status           | Description                                                                                                                                      |
| --------------------------------- | ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------ |
| **`//art-dupl:accept` directive** | FULLY_FUNCTIONAL | Suppress accepted clone groups via inline comments in source code. Optional hash for precision matching. Override with `--no-accept-directives`. |
| **`.gitignore` honoring**         | FULLY_FUNCTIONAL | Files matching `.gitignore` patterns excluded by default. Override with `--include-ignored`.                                                     |
| **Baseline CI gating**            | FULLY_FUNCTIONAL | `art-dupl baseline` + `art-dupl check` for recording and comparing clone states across runs                                                      |
| **Typed exit codes**              | FULLY_FUNCTIONAL | 0=success, 1=error, 2=config/validation, 3=internal, 130=interrupted                                                                             |
| **Type-aware validation**         | FULLY_FUNCTIONAL | `--type-aware` errors with `--structural`/`--exact`, warns with `--incremental`                                                                  |

---

## 🔍 Semantic Detection

| Feature                       | Status           | Description                                                                                                                                                                                                                                                                                                                                                              |
| ----------------------------- | ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Semantic Mode**             | FULLY_FUNCTIONAL | Alpha-normalizes local identifiers (including closures); detects Type 2 (renamed) clones (default)                                                                                                                                                                                                                                                                       |
| **Exact Mode**                | FULLY_FUNCTIONAL | `--exact` matches identifier names verbatim (copy-paste / Type 1 only)                                                                                                                                                                                                                                                                                                   |
| **Structural-Only Mode**      | FULLY_FUNCTIONAL | `--structural` ignores all names; matches by AST shape only                                                                                                                                                                                                                                                                                                              |
| **DetectionMode Enum**        | FULLY_FUNCTIONAL | Single `Config.DetectionMode` enum replaces former bool flags (ADR-0007)                                                                                                                                                                                                                                                                                                 |
| **Clone Type Classification** | FULLY_FUNCTIONAL | Labels each clone type-1/2/3 in JSON, SARIF, and `--rich-text` output                                                                                                                                                                                                                                                                                                    |
| **Mutual Exclusion**          | FULLY_FUNCTIONAL | `--semantic` / `--exact` / `--structural` are mutually exclusive                                                                                                                                                                                                                                                                                                         |
| **Type-Aware Mode**           | FULLY_FUNCTIONAL | `--type-aware` uses `go/types` to encode variable types into hashes, eliminating same-method-different-receiver-type false positives (e.g. `time.Time.String` vs `*big.Int.String`). Also enables type-aware interface-method detection (same-package interface scanning for the `interface-method` actionability pattern). 10-100x slower. Follow-up items in TODO_LIST |
| **Generics-Extraction Candidates** | PARTIALLY_FUNCTIONAL | `--suggest-generics` finds clones where the algorithm is identical but local variable types differ — the class of duplication Go generics can eliminate. Uses type-erased hashing so structurally-identical functions on different types still match, then classifies by comparing `VarType` at corresponding positions. Outputs `generics:` hint (text) and `generics_candidate`/`generics_hint` fields (JSON). 12.5% precision on real-world validation (3 true / 24 surfaced on DiscordSync); needs precision filtering before general use. Same ~100x type-checking cost as `--type-aware`. Takes precedence when both flags are set. |

**Note:** Default is semantic mode (alpha-normalized). Use `--exact` for verbatim name matching or `--structural` for shape-only analysis.

---

## 🔧 Refactoring Advisor

| Feature                                  | Status           | Description                                                                                                                                                     |
| ---------------------------------------- | ---------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Clone Type Classification**            | FULLY_FUNCTIONAL | Labels each clone type-1/2/3 in JSON, SARIF, and `--rich-text` output                                                                                           |
| **Extractability Score**                 | FULLY_FUNCTIONAL | `lines_saved` + `extractable` fields in JSON for refactoring prioritization                                                                                     |
| **Actionability Verdict**                | FULLY_FUNCTIONAL | Labels clones actionable vs non-actionable (test boilerplate, idioms, etc.)                                                                                     |
| **Actionability Override**               | FULLY_FUNCTIONAL | `--no-actionability` disables filtering, showing ALL clones including boilerplate                                                                               |
| **Disable Specific Pattern**             | FULLY_FUNCTIONAL | `--disable-pattern <label>` selectively re-enables a single boilerplate pattern                                                                                 |
| **List Patterns**                        | FULLY_FUNCTIONAL | `--list-patterns` prints all 29 pattern labels (plus 4 property-engine labels)      |
| **Explain Mode**                         | FULLY_FUNCTIONAL | `--explain` prints why each clone group was reported (type, actionability, category, extractability)                                                            |
| **Property-Based Extractability Engine** | PARTIALLY_DONE   | Second-pass analysis using 4 computable properties (control-flow, ROI, parameterizability, confidence). Patterns run first, engine catches misses. See ADR-0017 |
| **Confidence Tiers**                     | PARTIALLY_DONE   | Three-tier output: actionable / low-confidence / non-actionable with `confidence` field in JSON                                                                 |
| **Pattern in JSON**                      | FULLY_FUNCTIONAL | `non_actionable_pattern` field in JSON/SARIF output identifies which boilerplate pattern matched                                                                |
| **Overlap Elimination**                  | FULLY_FUNCTIONAL | Suppresses nested clone groups; only the largest match is reported                                                                                              |
| **Test Noise Suppression**               | FULLY_FUNCTIONAL | `--ignore-tests` excludes test files; `--include-tests` overrides                                                                                               |

---

## 🚀 CI Integration

| Feature                | Status           | Description                                                                 |
| ---------------------- | ---------------- | --------------------------------------------------------------------------- |
| **Baseline Recording** | FULLY_FUNCTIONAL | `art-dupl baseline` snapshots accepted clones to `.art-dupl-baseline.json`  |
| **CI Check Mode**      | FULLY_FUNCTIONAL | `art-dupl check` reports only new clones; exits 1 for CI gates              |
| **Diff Report**        | FULLY_FUNCTIONAL | `--diff-report <baseline>` shows new/suppressed/resolved clone groups       |
| **GitHub Actions**     | FULLY_FUNCTIONAL | `templates/github-actions-duplicate-check.yml` template included            |
| **Pre-Commit Hook**    | FULLY_FUNCTIONAL | `templates/pre-commit-hook.yaml` for pre-commit framework integration       |
| **CI Self-Test**       | FULLY_FUNCTIONAL | Nix `self-test` check enforces art-dupl's own zero-duplication invariant    |
| **Lint Config Guard**  | FULLY_FUNCTIONAL | Nix check + GitHub workflow reject `exhaustruct`/`tagliatelle` re-additions |

---

## 🛡️ Smart Filtering

| Feature                       | Status           | Description                                                                                                                                         |
| ----------------------------- | ---------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| **SQLC Detection**            | FULLY_FUNCTIONAL | Auto-detects `sqlc.yaml` in parent dirs, filters generated `.go`                                                                                    |
| **Templ Filtering**           | FULLY_FUNCTIONAL | `.templ` source files included by default; `*_templ.go` filtered by default                                                                         |
| **Lazy Content Reading**      | FULLY_FUNCTIONAL | gogenfilter v3.4.0 `FilterDetailedAndContent` reads file content only when phase-2 detection runs. Phase-1 filename catches skip disk I/O entirely. |
| **Protobuf Filtering**        | FULLY_FUNCTIONAL | Filters `.pb.go`, `_grpc.pb.go` files                                                                                                               |
| **Mockgen Filtering**         | FULLY_FUNCTIONAL | Filters mockgen generated files                                                                                                                     |
| **Stringer Filtering**        | FULLY_FUNCTIONAL | Filters stringer generated files                                                                                                                    |
| **Include Overrides**         | FULLY_FUNCTIONAL | `--include-generated <category>` to override (`sqlc`, `templ`, `protobuf`, `mockgen`, `stringer`, `generic`, `all`)                                 |
| **Custom Include/Exclude**    | FULLY_FUNCTIONAL | `--include-pattern` / `--exclude-pattern` glob patterns                                                                                             |
| **Directory Exclusions**      | FULLY_FUNCTIONAL | `vendor/`, `.git/`, `node_modules/` excluded by default                                                                                             |
| **Node Modules**              | FULLY_FUNCTIONAL | `--include-node-modules` includes for hash detection                                                                                                |
| **File Type Filter (--only)** | FULLY_FUNCTIONAL | Restrict to `go` or `templ` file types                                                                                                              |

---

## ⚡ Performance & Concurrency

| Feature                   | Status           | Description                                                                                                                                         |
| ------------------------- | ---------------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Parallel Parsing**      | FULLY_FUNCTIONAL | Worker pool via `--workers` flag (0=auto, NumCPU)                                                                                                   |
| **Parallel Search**       | FULLY_FUNCTIONAL | `--search-workers` flag parallelizes suffix tree DFS across root subtrees (0/1=sequential, >1=N workers). 1.5-3.4x speedup.                         |
| **Memory-Compact Tree**   | FULLY_FUNCTIONAL | Suffix tree data stored as `[]TokenValue` (int32, 4 bytes) instead of `[]Token` (interface, 16 bytes). 75% pointer-array memory reduction.          |
| **Incremental Analysis**  | FULLY_FUNCTIONAL | SHA-256 content-hash AST caching, `--incremental` flag, CacheVersion 3                                                                              |
| **Git-Aware Incremental** | REMOVED          | `--since` flag removed (was a dead stub, never read). Only content-hash caching via `--incremental` works. Git-diff file selection not implemented. |
| **Cache Management**      | FULLY_FUNCTIONAL | `--cache-dir`, `--clear-cache`, `--max-cache-entries`, file-based gob serialization, hysteresis pruning (110%/90%)                                   |
| **In-Memory LRU Cache**   | FULLY_FUNCTIONAL | 512-entry `container/list`-based LRU on top of `FileCache`. O(1) hit, no gob deserialization. Deep-clone on read. (`cache/lru.go`)                 |
| **Parallel Incremental**  | FULLY_FUNCTIONAL | `ParseIncrementalParallel` worker pool + `singleflight.Group` dedup for byte-identical files (commit `94b5205`)                                     |
| **Execution Timeout**     | FULLY_FUNCTIONAL | `--timeout` with context cancellation (default 30m)                                                                                                 |
| **Performance Profiling** | EXPERIMENTAL     | Hidden `--profile` flag; pprof CPU/mem profile capture                                                                                              |
| **SIMD Optimizations**    | N/A              | `syntax/hash_seq.go` (renamed from `hash_simd.go`) uses sync.Pool + xxh3, no hand-written SIMD                                                      |

---

## 🖥️ Professional CLI (via Fang)

| Feature                      | Status           | Description                                                                         |
| ---------------------------- | ---------------- | ----------------------------------------------------------------------------------- |
| **Styled Help Output**       | FULLY_FUNCTIONAL | Rich, themed help text via Fang framework                                           |
| **Shell Completion**         | FULLY_FUNCTIONAL | bash, zsh, fish, PowerShell with `--no-descriptions` option                         |
| **Man Page Generation**      | FULLY_FUNCTIONAL | `art-dupl man` generates manual pages                                               |
| **Version Information**      | FULLY_FUNCTIONAL | Version, commit, build date; `version --json`, `version --short`                    |
| **Configurable Verbosity**   | FULLY_FUNCTIONAL | `-v` (verbose), `-vv` (extra verbose), `--quiet`/`-q` (suppress status)             |
| **Color Control**            | FULLY_FUNCTIONAL | `--no-color` flag, `NO_COLOR` env var (lipgloss native)                             |
| **Typed Exit Codes**         | FULLY_FUNCTIONAL | 0=success, 1=general, 2=config/validation, 3=internal, 130=interrupted (ADR-0013)   |
| **Line-Count Filtering**     | FULLY_FUNCTIONAL | `--min-lines` suppresses clone groups with fewer lines (minimum across all clones)  |
| **Token-Count Filtering**    | FULLY_FUNCTIONAL | `--min-tokens` suppresses clone groups where any clone has fewer than N tokens       |
| **Test Threshold**           | FULLY_FUNCTIONAL | `--test-threshold` sets a separate (higher) threshold for test files                |
| **Test Suppression**         | FULLY_FUNCTIONAL | `--suppress-test-low` suppresses low-priority clones in test files                  |
| **Token Dump**               | FULLY_FUNCTIONAL | `--dump-tokens` outputs serialized token stream for debugging false positives       |
| **Rich Text Output**         | FULLY_FUNCTIONAL | `--rich-text` adds priority/category/actionability badges to text output            |
| **Explain Mode**             | FULLY_FUNCTIONAL | `--explain` prints a per-group rationale (type, actionability, category, savings)   |
| **Actionability Toggle**     | FULLY_FUNCTIONAL | `--no-actionability` shows all clones, including non-actionable boilerplate         |
| **Disable Pattern**          | FULLY_FUNCTIONAL | `--disable-pattern <label>` re-enables a specific boilerplate pattern               |
| **List Patterns**            | FULLY_FUNCTIONAL | `--list-patterns` prints all 29 pattern labels                                      |
| **Threshold Recommendation** | FULLY_FUNCTIONAL | `--recommend-threshold` suggests CI gate and deep-audit thresholds                  |
| **Token-Count Filter**       | FULLY_FUNCTIONAL | `--min-tokens N` suppresses clone groups where any clone has < N tokens             |
| **Diff Report**              | FULLY_FUNCTIONAL | `--diff-report <baseline>` shows new/suppressed/resolved clones                     |
| **HTML to File**             | FULLY_FUNCTIONAL | `--html-out <file>` writes HTML report to file with auto-open                       |
| **Quiet Mode**               | FULLY_FUNCTIONAL | `--quiet`/`-q` suppresses progress and status output                                |
| **Color Control**            | FULLY_FUNCTIONAL | `--no-color` disables colored output                                                |
| **Type-Aware Mode**          | FULLY_FUNCTIONAL | `--type-aware` encodes variable types into hashes via `go/types` (see Semantic)     |
| **Parallel Search**          | FULLY_FUNCTIONAL | `--search-workers N` parallelizes suffix tree search (0/1=sequential, >1=N workers) |
| **Version Subcommand**       | FULLY_FUNCTIONAL | `art-dupl version [--json                                                           | --short]` prints structured version info |

---

## 🔧 Configuration

| Feature                      | Status           | Description                                                         |
| ---------------------------- | ---------------- | ------------------------------------------------------------------- |
| **Command-Line Flags**       | FULLY_FUNCTIONAL | 45+ flags for full control                                          |
| **JSON Configuration Files** | FULLY_FUNCTIONAL | `--config` / `-c` flag, JSON-tagged Config struct                   |
| **Configuration Merging**    | FULLY_FUNCTIONAL | CLI flags override file config, file overrides defaults             |
| **Threshold Control**        | FULLY_FUNCTIONAL | Adjustable minimum duplicated statement count (default: 5)          |
| **Threshold Recommendation** | FULLY_FUNCTIONAL | `--recommend-threshold` suggests CI gate and deep-audit thresholds based on codebase |
| **Vendor Directory Control** | FULLY_FUNCTIONAL | `--vendor` to include vendor directory                              |
| **File Input from Stdin**    | FULLY_FUNCTIONAL | `--files` / `-f` reads file paths from stdin                        |
| **YAML Config Files**        | FULLY_FUNCTIONAL | `--config` auto-detects `.yaml`/`.yml` alongside JSON               |

---

## 📦 SDK / Programmatic API

| Feature                | Status           | Description                                                       |
| ---------------------- | ---------------- | ----------------------------------------------------------------- |
| **Detector Interface** | FULLY_FUNCTIONAL | `FindClones()`, `FindClonesStreamResult()`, `Close()`             |
| **Options Builder**    | FULLY_FUNCTIONAL | Full configuration via `Options` struct                           |
| **Progress Callbacks** | FULLY_FUNCTIONAL | Stage, completed, total, percentage, current file                 |
| **Result Type**        | FULLY_FUNCTIONAL | CloneGroups, Summary, Metadata                                    |
| **Validation**         | FULLY_FUNCTIONAL | 12 domain sentinel errors, `ValidateOptions()`, `Clone.IsValid()` |
| **Custom FileReader**  | FULLY_FUNCTIONAL | Injectable file reader for testing/custom sources                 |
| **TypeAware Mode**     | FULLY_FUNCTIONAL | `Options.TypeAware` — go/types-based false-positive elimination    |
| **SuggestGenerics**    | FULLY_FUNCTIONAL | `Options.SuggestGenerics` — finds clones extractable via Go generics |

---

## 🧪 Testing

| Feature               | Status           | Description                                      |
| --------------------- | ---------------- | ------------------------------------------------ |
| **Unit Tests**        | FULLY_FUNCTIONAL | Standard `testing` package, table-driven tests   |
| **BDD Tests**         | FULLY_FUNCTIONAL | Ginkgo/Gomega in `bdd/` directory, user-focused  |
| **Integration Tests** | FULLY_FUNCTIONAL | CLI integration tests in `cmd/`                  |
| **Golden File Tests** | FULLY_FUNCTIONAL | HTML output golden testing                       |
| **Fuzz Tests**        | FULLY_FUNCTIONAL | Fuzz tests for templ parser and suffix tree      |
| **Benchmarks**        | FULLY_FUNCTIONAL | `_bench_test.go` files with allocation reporting |
| **Race Detection**    | FULLY_FUNCTIONAL | All tests pass with `-race` flag                 |

---

## 🏗️ Architecture Components

| Component        | Status           | Description                                                                                              |
| ---------------- | ---------------- | -------------------------------------------------------------------------------------------------------- |
| **suffixtree/**  | FULLY_FUNCTIONAL | Core Ukkonen's suffix tree, O(1) map transitions, memory-compact `[]TokenValue` storage, parallel search |
| **syntax/**      | FULLY_FUNCTIONAL | AST handling, Go + Templ parsers, serialization                                                          |
| **job/**         | FULLY_FUNCTIONAL | Parsing pipeline, parallel workers, incremental                                                          |
| **printer/**     | FULLY_FUNCTIONAL | 7 output formats, sorting, classification, templ-based HTML                                              |
| **hash/**        | FULLY_FUNCTIONAL | XXH3 streaming hash detection                                                                            |
| **config/**      | FULLY_FUNCTIONAL | Multi-source config with validation                                                                      |
| **detection/**   | FULLY_FUNCTIONAL | Multi-detector coordination via goroutines                                                               |
| **cache/**       | FULLY_FUNCTIONAL | File-based AST caching with SHA-256 content hashing, in-memory LRU layer, hysteresis pruning (110%/90%)  |
| **domain/**      | FULLY_FUNCTIONAL | Value objects: ProcessedClone, enums, validation sentinels                                               |
| **errors/**      | FULLY_FUNCTIONAL | 7 error categories (ErrorType), single DuplError struct, typed wrapping                                  |
| **pkg/artdupl/** | FULLY_FUNCTIONAL | Public SDK with Detector interface, comprehensive godoc                                                  |

---

## 🔮 Experimental / In Progress

| Feature                   | Status       | Description                                  |
| ------------------------- | ------------ | -------------------------------------------- |
| **Performance Profiling** | EXPERIMENTAL | `--profile` flag exists, pprof capture works |

---

## 🚫 Known Limitations

| Limitation                  | Impact | Description                                                                                    |
| --------------------------- | ------ | ---------------------------------------------------------------------------------------------- |
| **Go and Templ Only**       | High   | Only `.go` and `.templ` files supported                                                        |
| **No Git-Diff Incremental** | Low    | Only content-hash caching (`--incremental`); git-diff file selection not implemented           |
| **No GitHub Releases**      | Low    | v0.1.0 and v0.4.0 have GitHub Releases. v0.2.0 and v0.3.0 have git tags but no release assets. |

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
art-dupl --html-out report.html  # HTML report to file (auto-opens)
art-dupl --json                # JSON with statistics
art-dupl --plumbing            # Machine-readable
art-dupl --sarif               # SARIF 2.1.0
art-dupl --all -o ./reports    # All formats to directory
```

### Diff Reports

```bash
# Record baseline, then see what changed
art-dupl baseline . -t 15
art-dupl --diff-report .art-dupl-baseline.json . -t 15
art-dupl --diff-report .art-dupl-baseline.json --json . -t 15
```

### Type-Aware & Generics Detection

```bash
# Type-aware: encode variable types into hashes (eliminates false positives)
art-dupl --type-aware -t 5 ./src

# Generics-extraction candidates (same algorithm, different types)
art-dupl --suggest-generics -t 1 ./src

# Filter noise by token count
art-dupl --min-tokens 15 -t 1 ./src
```

### Threshold Recommendation

```bash
# Get CI gate and deep-audit thresholds based on codebase size
art-dupl --recommend-threshold ./src
```

### Pattern Control

```bash
# List all 29 actionability pattern labels (plus 4 property-engine labels)
art-dupl --list-patterns

# Re-enable a specific boilerplate pattern
art-dupl --disable-pattern guard-clause ./src
```

### Filtering

```bash
art-dupl --include-generated sqlc  # Include sqlc-generated files (filtered by default)
art-dupl --only go             # Go files only
art-dupl --include-pattern "gen/*" --exclude-pattern "mock_*"
```

### Performance

```bash
art-dupl --workers 8           # 8 parallel parsing workers
art-dupl --search-workers 4    # 4 parallel search workers
art-dupl --incremental         # AST caching
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
