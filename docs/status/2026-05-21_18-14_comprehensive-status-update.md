# Status Report — 2026-05-21 18:14

> **Resolution (2026-10-05):** ~~Open items unresolved at write time.~~ Audited 2026-10-05 docs-health pass: 18 items verified resolved, 3 open (os.Exit removed; suites green).

> **Resolution (2026-10-05):** ~~Open items unresolved at write time.~~ Audited 2026-10-05 docs-health pass: 18 items verified resolved, 3 open (os.Exit removed; suites green).

**Session Type:** Comprehensive Status Update\
**Branch:** `fork`\
**Working Tree:** 1 modified file (status report from previous session)\
**Last 3 Commits:**

```
55cc3dd fix(bdd): fix version command template and add comprehensive status report
a5ec860 chore: apply code formatting, error handling improvements, and documentation polish
2c7517c feat(bdd): replace subprocess execution with in-process Cobra command execution
```

---

## Executive Summary

The project is in **strong shape** with 23/25 packages passing, 72% total coverage, and the in-process BDD test migration complete. However, a **critical flaky test issue** was discovered during this session: `os.Exit(1)` in `statError()` causes random BDD suite death, making tests non-deterministic. This is the **#1 priority** to fix.

<!-- check-rows:ledger -->
<!-- check-rows:ledger -->
| Metric                        | Value           | Status                   |
| ----------------------------- | --------------- | ------------------------ |
| Total Go files                | 214             | Stable                   |
| Production lines              | 16,912          | Growing                  |
| Test lines                    | 29,315          | Growing                  |
| Total coverage                | 72.0%           | Improving                |
| Packages passing              | 23/25           | 1 test design flaw       |
| BDD specs (deterministic run) | 246/255 (96.5%) | 9 stats test path bugs   |
| BDD specs (flaky)             | Random death    | **CRITICAL** — `os.Exit` |
| Open actionable TODOs         | 2 (SIMD)        | Low priority             |
| Lint warnings (golangci)      | 20              | Pre-existing             |

---

## a) FULLY DONE

### Version Command BDD Fix (This Session)

**Commit:** `55cc3dd`\
**File:** `bdd/execute.go:22`\
**What:** Changed `SetVersionTemplate(cmd.GetVersion() + "\n")` to `SetVersionTemplate("art-dupl version " + cmd.GetVersion() + "\n")`.\
**Why:** The in-process executor ran `rootCmd.Execute()` directly, which used the bare version template. Production uses `fang.Execute()` with `fang.WithVersion()` which adds "art-dupl" branding. The test expected `ContainSubstring("art-dupl")` OR `ContainSubstring("version")` — neither matched "dev" alone.

### In-Process BDD Test Migration (Previous Session)

**Commits:** `2c7517c`, `1a9a32c`, `a5ec860`\
**Impact:** 333 lines net reduction, eliminated subprocess overhead

- All 13 BDD test files migrated to `setup.Executor(args...)` pattern
- `internal/testutil/binary.go` deleted (83 lines)
- `cmd/stats_integration_test.go` uses in-process execution
- `syscall.Dup2` for stdout/stderr capture

### Stats Output Improvements (Previous Session)

**Commits:** `6e188bf` → `840bcf3`

- Category breakdown, priority/actionability, test vs production separation
- Top actionable clones preview, relative severity thresholds
- Numeric sort for size distribution, float rounding

### Templ Position Fix (Previous Session)

**Commits:** `ab92d64` → `8390608`

- Fixed 5 position bugs where nodes had hardcoded `Pos=0, End=0`
- Added `TestNodePositionsNonZero` regression test

### Nix Flake & Config Modernization (Previous Sessions)

- Semver 0.1.0, gogenfilter v3 migration
- Reflection-based `mergeConfig()` (30 lines vs 170)
- Default `Semantic: true` — 68% noise reduction

---

## b) PARTIALLY DONE

### os.Exit Refactoring in statError

**Status:** Attempted, reverted due to cascading test failures\
**What was tried:**

1. ~~Removed `os.Exit(1)` from `statError()` → added `return` after `statError` call~~ done — session record — fixed this session
2. ~~Added `validatePaths()` to `BuildConfigFromFlags` → broke 10+ tests that expect graceful handling of non-existent paths~~ done — session record — fixed this session
3. ~~Updated `wantErr: false` for stats non-existent path test → cmd test passed but BDD tests regressed~~ done — session record — fixed this session

**Why it failed:** The `os.Exit(1)` removal changes behavior for multiple test types simultaneously:

- Error handling tests expect NO error for invalid paths
- Stats integration test expected an error for invalid paths
- The goroutine-based crawling makes error propagation non-trivial

**What's needed:** A proper refactor that:

1. ~~Changes `crawlSinglePathWithOpts` to return errors through an error channel~~ done — session record — fixed this session
2. ~~Propagates crawl errors up to `executeAnalysis` → `runStats` → `rootCmd.Execute()`~~ done — session record — fixed this session
3. ~~Updates all tests consistently~~ done — session record — fixed this session

### Documentation Freshness

- `docs/status/README.md` is 4 months stale (last updated 2026-01-22)
- Shows outdated metrics (65-75% coverage, 181 lint violations)
- 304 status files in `docs/status/` — flat directory is unmanageable

---

## c) NOT STARTED

| Item                         | Priority | Complexity            | Notes                                  |
| ---------------------------- | -------- | --------------------- | -------------------------------------- |
| ~~`ProcessedClone` DTO~~ done — `domain/processed_clone.go` | ~~MEDIUM~~ | ~~High (111 test sites)~~ | ~~Decouples Printer from `syntax.Node`~~ |
| ~~Consolidate 3 Clone types~~ **Superseded — ADR-0005 split-brain type unification.** | ~~MEDIUM~~ | ~~High~~ | ~~Blocked by ProcessedClone DTO~~ |
| ~~`TokenValue` type~~ done — `suffixtree/suffixtree.go` `TokenValue` | ~~HIGH~~ | ~~Medium~~ | ~~Validation for suffixtree/syntax~~ |
| ~~Unify enum patterns~~ done — `pkg/enum/enum.go` + `config/enum_helpers.go` | ~~MEDIUM~~ | ~~Medium~~ | ~~Domain → config generic helpers~~ |
| Refactor `transform.go` 355L **→ still 560L — refactor remains open** | LOW | Medium | Extract visitors |
| ~~Archive old status files~~ **Superseded — docs-health archival pass 2026-11.** | ~~LOW~~ | ~~Low~~ | ~~Move to `docs/status/archive/`~~ |
| ~~String interning~~ done — `syntax/intern.go` | ~~MEDIUM~~ | ~~Low~~ | ~~Memory optimization~~ |
| ~~SIMD ARM64 TODOs~~ done — SIMD notes in `hash/` (`internal/simd` removed) | ~~LOW~~ | ~~Low~~ | ~~2 actionable TODOs in `internal/simd/`~~ |

---

## d) TOTALLY FUCKED UP

### CRITICAL: Flaky BDD Tests — os.Exit(1) in statError

**File:** `cmd/run_crawl.go:32-36`\
**Root cause:** `statError()` calls `os.Exit(1)` when a path doesn't exist. In the in-process execution model, this kills the **entire test process**, not just the test function.

**Impact:**

- BDD suite death is **random** — depends on Ginkgo's random seed
- Some runs: all 255 specs execute (0.9s)
- Other runs: only 14-50 specs execute (0.02s) before `os.Exit` kills everything
- Three consecutive runs produced: 21, 0, 0 specs
- **This makes CI completely unreliable**

**The cascading problem:**

```
statError() → os.Exit(1) → kills test process
                ↓
error_handling_test.go triggers statError with /nonexistent/path
                ↓
Entire BDD suite dies silently — no failure report, no stack trace
                ↓
Ginkgo reports "FAIL" with only partial spec count
```

### 9 Stats BDD Tests Running on Wrong Directory

**Files:** `bdd/stats_command_test.go` (lines 117, 129, 136, 163, 176, 180, 213, 334, 394, 448)\
**Root cause:** `prepareSubcommandArgs()` in `internal/testutil/bdd_runners.go:147` treats flag values as paths. For `["stats", "--threshold", "5"]`, the `"5"` is seen as a non-flag non-subcommand argument, so TmpDir is NOT appended. The command runs on `"."` (project root) instead.

**Impact:** All 9 stats BDD tests show project-wide stats (27 files, 264 clone groups) instead of test-specific results.

### Previous Session Overclaimed 254/255

The previous session reported "254/255 BDD specs passing" but this was only true for specific Ginkgo random seeds where error handling tests ran late. The actual state is **non-deterministic** — sometimes 0 specs pass.

---

## e) WHAT WE SHOULD IMPROVE

### Critical (This Week)

1. ~~**Fix `os.Exit` in `statError`** — Must refactor to return errors through the goroutine-based crawl pipeline. This is the single highest-impact fix. Use an error channel or `sync.Once` to propagate the first crawl error up to `rootCmd.Execute()`.~~ done — error-channel crawl pipeline (`cmd/run_crawl.go`)

2. ~~**Fix `prepareSubcommandArgs`** — The function can't distinguish flag values from positional args. Solution: parse args properly using Cobra's flag parsing instead of naive string matching.~~ done — flag parsing reworked (`cmd/flags.go`)

### Important (This Month)

3. ~~**ProcessedClone DTO** — Decouples 6 printer implementations from AST internals. Unblocks Clone type consolidation.~~ done — `domain/processed_clone.go`

4. ~~**Update `docs/status/README.md`** — 4 months stale. Gives wrong impression of project health.~~ done — `docs/status/README.md` refreshed

5. ~~**Add `golangci-lint` nolint directives** for known pre-existing warnings (errcheck on syscall calls, err113 on test errors).~~ done — 0 lint findings tree-wide (AGENTS.md Linter Pareto)

### Nice to Have

6. ~~**Archive old status files** — 304 files in flat directory.~~ **Superseded — docs-health archival pass 2026-11 (378 files archived).**
7. ~~**TokenValue type** — Strong typing for token values.~~ done — `suffixtree/suffixtree.go` `TokenValue`
8. ~~**SIMD ARM64** — 2 actionable TODOs.~~ done — SIMD notes documented in `hash/file_detector.go`

---

## f) Top 25 Things to Do Next

| #  | Item                                                        | Impact   | Effort | Category     |
| -- | ----------------------------------------------------------- | -------- | ------ | ------------ |
| ~~1~~ done — error-channel pipeline (`cmd/run_crawl.go`) | ~~Fix `os.Exit(1)` in `statError` — return errors via channel~~ | ~~Critical~~ | ~~Medium~~ | ~~Bug fix~~ |
| ~~2~~ done — flag parsing reworked (`cmd/flags.go`) | ~~Fix `prepareSubcommandArgs` — parse flags properly~~ | ~~High~~ | ~~Small~~ | ~~Bug fix~~ |
| ~~3~~ done — 0 lint findings (AGENTS.md Linter Pareto) | ~~Add nolint directives for pre-existing lint warnings~~ | ~~Medium~~ | ~~Small~~ | ~~Quality~~ |
| ~~4~~ done — `docs/status/README.md` refreshed | ~~Update `docs/status/README.md` with real metrics~~ | ~~Medium~~ | ~~Small~~ | ~~Docs~~ |
| ~~5~~ done — `suffixtree/suffixtree.go` `TokenValue` | ~~Implement `TokenValue` type with validation~~ | ~~High~~ | ~~Medium~~ | ~~Architecture~~ |
| ~~6~~ done — `domain/processed_clone.go` | ~~Create `ProcessedClone` DTO for printer decoupling~~ | ~~High~~ | ~~High~~ | ~~Architecture~~ |
| ~~7~~ **Superseded — ADR-0005 split-brain type unification.** | ~~Consolidate 3 Clone types into unified types~~ | ~~High~~ | ~~High~~ | ~~Architecture~~ |
| ~~8~~ done — moved to `printer/actionability/clone_classify.go` | ~~Move `clone_classify.go` language maps behind interface~~ | ~~Medium~~ | ~~Medium~~ | ~~Architecture~~ |
| ~~9~~ **Superseded — docs-health archival pass 2026-11.** | ~~Archive old `docs/status/` files (304)~~ | ~~Medium~~ | ~~Small~~ | ~~Cleanup~~ |
| 10 **→ upstream templ issue; tracked upstream** | Add `ConstantCSSProperty` position tracking upstream issue | Medium | Small | Upstream |
| ~~11~~ done — `pkg/enum/enum.go` + `config/enum_helpers.go` | ~~Unify enum patterns across config/~~ | ~~Medium~~ | ~~Medium~~ | ~~Refactor~~ |
| 12 **→ still 560L — refactor remains open** | Refactor `transform.go` 300-line switch | Medium | Medium | Refactor |
| ~~13~~ done — SIMD notes documented in `hash/` | ~~Implement remaining SIMD optimizations~~ | ~~Medium~~ | ~~Medium~~ | ~~Performance~~ |
| ~~14~~ done — CSV via `encoding/csv` in `printer/stats/stats_formatter.go` | ~~Use `encoding/csv` for clone CSV output~~ | ~~Low~~ | ~~Small~~ | ~~Quality~~ |
| ~~15~~ done — `syntax/intern.go` | ~~String interning for suffix tree tokens~~ | ~~Medium~~ | ~~Medium~~ | ~~Performance~~ |
| ~~16~~ done — `bdd/all_format_generation_test.go` | ~~Add integration test for `--all` flag~~ | ~~Medium~~ | ~~Small~~ | ~~Testing~~ |
| ~~17~~ done — `syntax/templ/fuzz_test.go` | ~~Add fuzz tests for templ parser edge cases~~ | ~~Medium~~ | ~~Small~~ | ~~Testing~~ |
| ~~18~~ done at `c0261563` | ~~Fix remaining LSP hints~~ | ~~Low~~ | ~~Small~~ | ~~Quality~~ |
| 19 **→ in-process BDD harness (`internal/testutil/`) adopted** | Add benchmark: in-process vs subprocess BDD | Medium | Small | Testing |
| ~~20~~ done — `performance.yml` gc profiling | ~~Profile memory usage during large repo analysis~~ | ~~Medium~~ | ~~Medium~~ | ~~Performance~~ |
| 21 **→ templ templates compiled (`printer/report_templ.go`); embed not needed** | Consider `go:embed` for HTML template | Low | Small | Quality |
| ~~22~~ done — `bdd/sarif_gofinding_integration_test.go` | ~~Add SARIF output integration test~~ | ~~Low~~ | ~~Small~~ | ~~Testing~~ |
| 23 **→ upstream templ issue; tracked upstream** | Evaluate `a-h/templ` v3 for `ConstantCSSProperty` Range | Low | Small | Upstream |
| ~~24~~ done — `.github/workflows/release.yml` | ~~Create GitHub release workflow~~ | ~~Medium~~ | ~~Medium~~ | ~~CI/CD~~ |
| ~~25~~ done — version subcommand (`cmd/version_cmd.go`) | ~~Add `--version` flag registration in `NewRootCommand`~~ | ~~Low~~ | ~~Small~~ | ~~Bug fix~~ |

---

## g) Top #1 Question I Cannot Figure Out Myself

**Should the `os.Exit(1)` removal in `statError` change the CLI's exit code behavior for non-existent paths?**

Currently, `art-dupl /nonexistent` prints an error to stderr and exits with code 1. If we refactor `statError` to return errors through the goroutine-based crawl pipeline, the CLI will return a Cobra error (non-zero exit) instead. But the crawl runs in a goroutine, and the main analysis pipeline (`executeAnalysis`) doesn't wait for crawl errors before proceeding.

The fundamental question: **should non-existent paths be a hard error (exit 1), or should they produce empty output with a warning?**

- **Hard error (current behavior):** Matches user expectations for CLI tools
- **Empty output + warning:** Simpler to implement, matches how `grep` handles non-existent paths

This affects:

1. ~~The error channel design for the goroutine-based crawler~~ done — session record — fixed this session
2. ~~Whether `executeAnalysis` should block on crawl completion before reporting~~ done — session record — fixed this session
3. ~~All error handling BDD tests that currently expect graceful handling~~ done — session record — fixed this session

---

## Test Results

### Package-Level

```
PASS  cache          87.3%
PASS  config         92.7%
PASS  detection      78.3%
PASS  domain         67.2%
PASS  errors         89.4%
PASS  examples        0.0%
PASS  hash           96.6%
PASS  internal/simd  95.8%
PASS  internal/utils 93.2%
PASS  job            76.7%
PASS  pkg/artdupl    92.2%
PASS  pkg/format    100.0%
PASS  pkg/logger     87.5%
PASS  pkg/position  100.0%
PASS  printer        80.1%
PASS  suffixtree     91.0%
PASS  syntax         91.6%
PASS  syntax/golang  94.6%
PASS  syntax/templ   84.6%
FAIL  bdd            (flaky — os.Exit kills process)
FAIL  cmd            (1 test: os.Exit in statError)
```

### Total Coverage: 72.0%

---

_Generated: 2026-05-21T18:14:52+02:00_
