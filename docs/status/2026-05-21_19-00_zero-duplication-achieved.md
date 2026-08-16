# art-dupl Status Report

**Date:** 2026-05-21 19:00
**Branch:** fork
**Version:** v0.1.0
**Reporter:** Crush (automated)

---

## Executive Summary

art-dupl is in **strong shape** for v0.1.0. Core detection engine works across 2 languages (Go + Templ), 7 output formats, 4 sorting modes, multi-method detection, and a stats subcommand with health grading. The codebase is **zero-duplication** at threshold 40 with semantic mode. Two test suites (`bdd`, `cmd`) have a known **os.Exit flaky test issue** — tests pass their assertions but the process exits non-zero due to Cobra's os.Exit behavior during in-process command execution. 20 linter issues remain (mostly errcheck in test code).

---

## a) FULLY DONE

### Core Engine

- [x] Suffix tree detection (Ukkonen's algorithm, O(1) map transitions)
- [x] Hash-based detection (XXH3 streaming, ~20x faster than SHA-256)
- [x] Multi-detection mode (both methods in parallel via goroutines)
- [x] Semantic-aware matching (FNV-1a, **default ON** since 2026-05-17)
- [x] Structural-only mode (`--structural` flag)
- [x] SIMD-optimized hot paths in internal/simd/

### Languages

- [x] Go AST analysis (45 node types, full coverage)
- [x] Templ template analysis (28 node types, pure Go parser, ON by default)
- [x] Templ position bugs fixed (5 bugs in ConstantAttribute, BoolConstantAttribute, ChildrenExpression, CaseExpression, File root End)

### Output Formats (7)

- [x] Text (human-readable, colored)
- [x] HTML (dark theme, syntax highlighting, VSCode links, diff visualization)
- [x] JSON (structured, JSONv2)
- [x] Simple-JSON (impact score, instances)
- [x] Plumbing (machine-readable for CI/CD)
- [x] SARIF 2.1.0 (GitHub Advanced Security / CodeQL)
- [x] CSV (stats subcommand)

### Statistics Subcommand

- [x] Text stats with lipgloss coloring
- [x] JSON stats
- [x] CSV stats
- [x] Health grade (A-F)
- [x] Clone metrics, spread analysis, severity distribution
- [x] Priority and actionability breakdown
- [x] Test vs production clone separation
- [x] Top actionable clones preview
- [x] Relative severity thresholds (based on configured threshold)

### CLI & Configuration

- [x] Professional CLI via Fang/Cobra framework
- [x] Shell completion (bash, zsh, fish, PowerShell)
- [x] Version command with template
- [x] JSON config file support (`-c`/`--config`)
- [x] Reflection-based config merge (30 lines, auto-handles new fields)
- [x] Smart filtering (SQLC, Templ, Protobuf, Mockgen, Stringer — auto-excluded)
- [x] Custom include/exclude patterns
- [x] `--all` batch generation with `--output-dir`
- [x] 4 sorting modes: size, occurrence, hash, total-tokens

### Build & Infrastructure

- [x] Justfile (preferred) + Makefile (legacy, GOEXPERIMENT=jsonv2)
- [x] Nix flake with private gogenfilter dependency pattern
- [x] GitHub Actions CI (build, checks, performance)
- [x] Golangci-lint configuration (.golangci.yml)
- [x] Comprehensive AGENTS.md documentation

### Code Quality (this session)

- [x] **Zero duplication** at -t 40 --semantic (was 2 clone groups, fixed in this session)
- [x] Extracted `mustDeferSelectorCall` and `mustIfErrReturnNil` helpers in actionability tests
- [x] 214 Go files, 46,155 lines total

### Architecture Decisions

- [x] Config extraction: DetectionConfig, SortCriteria, FileType, ParseOutputFormat
- [x] MethodDetector interface for pluggable detection methods
- [x] Pipeline unification: CLI and SDK share MultiDetector dispatch
- [x] Domain cleanup: removed 6 unused types and 15 dead error variables
- [x] Config builder safety: no more panic(err)
- [x] Architecture enforcement via .go-arch-lint.yml

---

## b) PARTIALLY DONE

### Detection Methods (2 of 4 wired)

- [~] `TodoDetector` — implemented in `detection/todos.go`, wired to MultiDetector, **not exposed as CLI option**
- [~] `LegacyDetector` — implemented in `detection/todos.go`, wired to MultiDetector, **not exposed as CLI option**
- Users can only use `art-dupl` and `hash` via `-m` flag today.

### Test Coverage

- **21 of 23 packages pass** with good coverage
- Average coverage: ~87% across passing packages
- Coverage gaps: `domain` (67.2%), `job` (76.7%), `detection` (78.3%), `printer` (80.1%)

### Linter Cleanliness

- 20 issues remaining (down from higher numbers in previous sessions)
- Categories: err113 (2), errcheck (10), exhaustruct (2), goconst (5), gocyclo (1)

---

## c) NOT STARTED

1. **TokenValue type with validation** — refactor suffixtree/syntax to use it (HIGH priority per TODO_LIST.md)
2. **ProcessedClone DTO** — decouple Printer from syntax.Node (111 test call sites)
3. **Clone type consolidation** — 3 parallel Clone types (printer.clone, pkg/artdupl.Clone, printer.CloneGroup)
4. **Proper CSV output** — use encoding/csv instead of manual formatting
5. **Enum unification** — domain enums should use config's generic helpers
6. **Memory layout optimization** — SIMD-friendly data structures, string interning
7. **SIMD TODOs** — 6 items in syntax/hash_simd.go and internal/simd/
8. **transform.go refactor** — 369L, 300L switch statement
9. **Old status docs archival** — 304 files in docs/status/, keep last 30 days
10. **SDK design** — SDK_DESIGN.md exists but no implementation

---

## d) TOTALLY FUCKED UP

### os.Exit Flaky Test Issue (CRITICAL BLOCKER)

**Packages affected:** `bdd/` and `cmd/`

**Symptom:** All 255 BDD specs show green dots (passing) and all cmd test assertions pass, but the test binary exits with non-zero. Root cause: Cobra calls `os.Exit()` during command execution, and the in-process test runner catches this as a failure.

**Impact:** CI will always show red for these two packages. This masks real test failures.

**History:** Documented in commit `ca6d02e` — "comprehensive status update — os.Exit flaky test discovery"

**Estimated fix complexity:** MEDIUM-HIGH — requires either:
a) Refactor command execution to not call os.Exit in test mode (Cobra provides `SetExitFunc`)
b) Or use subprocess-based testing for commands that exit

### Linter: 20 Issues

| Category    | Count | Severity |
| ----------- | ----- | -------- |
| errcheck    | 10    | MEDIUM   |
| goconst     | 5     | LOW      |
| err113      | 2     | LOW      |
| exhaustruct | 2     | LOW      |
| gocyclo     | 1     | MEDIUM   |

**Worst offender:** `printer/stats_formatter.go:361` — `buildJSONData()` has cyclomatic complexity 16 (threshold: 15).

### Architecture: Printer ↔ syntax.Node Coupling

`Printer.PrintClones(dups [][]*syntax.Node)` forces all 6 printer implementations to depend on AST internals. Each printer independently calls `ProcessNodeRange()` and `extractContent()`. Fix requires touching 111 test call sites.

### Architecture: Three Parallel Clone Types

- `printer.clone` (unexported): fragment, classification — richest for output
- `pkg/artdupl.Clone`: SDK type with primitives, `IsValid()` validation
- `printer.CloneGroup` vs `pkg/artdupl.CloneGroup`: different JSON shapes

### Upstream: ConstantCSSProperty Position Bug

`a-h/templ`'s `ConstantCSSProperty` has no `Range` field — only `Name` and `Value` strings. Mitigated by inheriting parent `CSSTemplate` range, but could still cause inaccurate line reporting.

---

## e) WHAT WE SHOULD IMPROVE

### Immediate (This Week)

1. **Fix os.Exit test flakiness** — this is the #1 quality blocker. SetExitFunc or subprocess pattern.
2. **Fix 10 errcheck issues** — unchecked syscall.Dup2/Close errors in test code.
3. **Reduce buildJSONData complexity** — split the 16-complexity function into smaller helpers.

### Short-term (Next 2 Weeks)

4. **Wire TODO/Legacy detectors to CLI** — they exist but users can't access them.
5. **Introduce ProcessedClone DTO** — start decoupling printer from AST internals.
6. **Proper CSV output** — use encoding/csv, not manual string formatting.
7. **Consolidate Clone types** — 3 types is 2 too many.

### Medium-term (Next Month)

8. **TokenValue type** — strong typing for token sequences in suffixtree/syntax.
9. **Enum unification** — use config's generic helpers for domain enums.
10. **Memory layout optimization** — SIMD-friendly, string interning for large codebases.

### Ongoing

11. **Archive old docs/status/** — 304 files is excessive.
12. **Refactor transform.go** — 369L, 300L switch needs table-driven or visitor pattern.

---

## f) Top #25 Things We Should Get Done Next

| #  | Task                                                                  | Priority | Effort | Impact           |
| -- | --------------------------------------------------------------------- | -------- | ------ | ---------------- |
| 1  | Fix os.Exit test flakiness (bdd + cmd packages)                       | CRITICAL | M      | CI trust         |
| 2  | Fix 10 errcheck issues in test code                                   | HIGH     | S      | Linter clean     |
| 3  | Extract buildJSONData into smaller functions (gocyclo 16→<10)         | HIGH     | S      | Maintainability  |
| 4  | Extract `--threshold` and `art-dupl` string constants in tests        | LOW      | S      | Linter clean     |
| 5  | Wire TodoDetector and LegacyDetector to CLI `-m` flag                 | HIGH     | M      | Feature complete |
| 6  | Fix 2 exhaustruct issues in testutil                                  | LOW      | S      | Linter clean     |
| 7  | Fix 2 err113 issues in testutil                                       | LOW      | S      | Linter clean     |
| 8  | Introduce ProcessedClone DTO to decouple Printer from syntax.Node     | HIGH     | L      | Architecture     |
| 9  | Consolidate 3 parallel Clone types                                    | MEDIUM   | L      | Simplicity       |
| 10 | Implement proper CSV output using encoding/csv                        | MEDIUM   | S      | Correctness      |
| 11 | Implement TokenValue type with validation                             | HIGH     | M      | Type safety      |
| 12 | Unify enum patterns (domain → config generic helpers)                 | MEDIUM   | M      | Consistency      |
| 13 | Refactor transform.go (369L, 300L switch)                             | MEDIUM   | M      | Maintainability  |
| 14 | Optimize memory layouts for SIMD-friendly structures                  | MEDIUM   | L      | Performance      |
| 15 | Implement string interning for large codebases                        | LOW      | M      | Memory           |
| 16 | Wire remaining SIMD TODOs (6 items)                                   | LOW      | M      | Performance      |
| 17 | Fix ConstantCSSProperty position (upstream PR or workaround)          | LOW      | S      | Correctness      |
| 18 | Archive old docs/status/ files (keep last 30 days)                    | LOW      | S      | Cleanliness      |
| 19 | Fix remaining LSP hints: unused params, unnecessary type args         | LOW      | S      | Cleanliness      |
| 20 | Implement SDK from SDK_DESIGN.md                                      | LOW      | XL     | Extensibility    |
| 21 | Add more fuzz tests for edge cases                                    | LOW      | M      | Robustness       |
| 22 | Coverage improvement: domain (67%), job (77%), detection (78%)        | MEDIUM   | M      | Quality          |
| 23 | Add benchmark regression CI (performance.yml exists but needs tuning) | MEDIUM   | S      | Performance      |
| 24 | Consider modularization (go-modularize skill exists)                  | LOW      | XL     | Architecture     |
| 25 | Migrate justfile → nix flake (per AGENTS.md preference)               | LOW      | M      | Tooling          |

---

## g) Top #1 Question I Cannot Figure Out Myself

**What is the intended relationship between `bdd/` and `cmd/` integration tests?**

Both packages test similar end-to-end scenarios (run command → check output). The `bdd/` package uses Ginkgo/Gomega with in-process Cobra execution. The `cmd/` package uses `testing` with `syscall.Dup2` for stdout capture. They overlap significantly.

**The question:** Should we:

- (a) Merge them — make `bdd/` the single source of truth for integration tests and delete redundant `cmd/` integration tests?
- (b) Keep separate — `cmd/` for unit-ish integration, `bdd/` for user-facing BDD scenarios?
- (c) Refactor both to use a shared test harness that solves the os.Exit problem once?

This decision affects the os.Exit fix strategy and long-term test architecture.

---

## Metrics Summary

| Metric                        | Value                           |
| ----------------------------- | ------------------------------- |
| Go files                      | 214                             |
| Total lines                   | 46,155                          |
| Non-test lines                | ~16,912                         |
| Test lines                    | ~29,243                         |
| Packages                      | 23                              |
| Passing packages              | 21/23 (91%)                     |
| Failing packages              | 2 (bdd, cmd — os.Exit flaky)    |
| Average coverage              | ~87% (passing packages)         |
| Linter issues                 | 20                              |
| Clone groups (t=40, semantic) | **0**                           |
| Largest non-test file         | printer/html_template.go (523L) |
| Version                       | v0.1.0                          |
| Languages supported           | 2 (Go, Templ)                   |
| Output formats                | 7                               |
| Detection methods             | 2 accessible + 2 unwired        |

---

## Test Results Detail

### Passing Packages (21/23)

| Package             | Coverage |
| ------------------- | -------- |
| pkg/format          | 100%     |
| pkg/position        | 100%     |
| hash                | 96.6%    |
| internal/simd       | 95.8%    |
| syntax/golang       | 94.6%    |
| config              | 92.7%    |
| pkg/artdupl         | 92.2%    |
| suffixtree          | 91.0%    |
| syntax              | 91.6%    |
| errors              | 89.4%    |
| cache               | 87.3%    |
| pkg/logger          | 87.5%    |
| syntax/templ        | 84.6%    |
| printer             | 80.1%    |
| detection           | 78.3%    |
| job                 | 76.7%    |
| internal/utils      | 93.2%    |
| domain              | 67.2%    |
| internal/filtertest | 50.0%    |
| examples            | 0.0%     |
| internal/configtest | n/a      |

### Failing Packages (2/23)

| Package | Status                                    | Root Cause       |
| ------- | ----------------------------------------- | ---------------- |
| bdd     | FAIL (all 255 specs pass, exit non-zero)  | os.Exit in Cobra |
| cmd     | FAIL (all assertions pass, exit non-zero) | os.Exit in Cobra |

---

## Session Changes (This Report)

- **printer/actionability_test.go**: Extracted `mustDeferSelectorCall(receiver, method)` and `mustIfErrReturnNil()` helpers, eliminating 120 lines of duplicated AST tree construction across 2 clone groups (4 instances). Zero duplication at t=40 --semantic.

---

_Generated by Crush on 2026-05-21 at 19:00_
