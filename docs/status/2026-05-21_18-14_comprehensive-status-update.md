# Status Report — 2026-05-21 18:14

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

1. Removed `os.Exit(1)` from `statError()` → added `return` after `statError` call
2. Added `validatePaths()` to `BuildConfigFromFlags` → broke 10+ tests that expect graceful handling of non-existent paths
3. Updated `wantErr: false` for stats non-existent path test → cmd test passed but BDD tests regressed

**Why it failed:** The `os.Exit(1)` removal changes behavior for multiple test types simultaneously:

- Error handling tests expect NO error for invalid paths
- Stats integration test expected an error for invalid paths
- The goroutine-based crawling makes error propagation non-trivial

**What's needed:** A proper refactor that:

1. Changes `crawlSinglePathWithOpts` to return errors through an error channel
2. Propagates crawl errors up to `executeAnalysis` → `runStats` → `rootCmd.Execute()`
3. Updates all tests consistently

### Documentation Freshness

- `docs/status/README.md` is 4 months stale (last updated 2026-01-22)
- Shows outdated metrics (65-75% coverage, 181 lint violations)
- 304 status files in `docs/status/` — flat directory is unmanageable

---

## c) NOT STARTED

| Item                         | Priority | Complexity            | Notes                                  |
| ---------------------------- | -------- | --------------------- | -------------------------------------- |
| `ProcessedClone` DTO         | MEDIUM   | High (111 test sites) | Decouples Printer from `syntax.Node`   |
| Consolidate 3 Clone types    | MEDIUM   | High                  | Blocked by ProcessedClone DTO          |
| `TokenValue` type            | HIGH     | Medium                | Validation for suffixtree/syntax       |
| Unify enum patterns          | MEDIUM   | Medium                | Domain → config generic helpers        |
| Refactor `transform.go` 355L | LOW      | Medium                | Extract visitors                       |
| Archive old status files     | LOW      | Low                   | Move to `docs/status/archive/`         |
| String interning             | MEDIUM   | Low                   | Memory optimization                    |
| SIMD ARM64 TODOs             | LOW      | Low                   | 2 actionable TODOs in `internal/simd/` |

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

1. **Fix `os.Exit` in `statError`** — Must refactor to return errors through the goroutine-based crawl pipeline. This is the single highest-impact fix. Use an error channel or `sync.Once` to propagate the first crawl error up to `rootCmd.Execute()`.

2. **Fix `prepareSubcommandArgs`** — The function can't distinguish flag values from positional args. Solution: parse args properly using Cobra's flag parsing instead of naive string matching.

### Important (This Month)

3. **ProcessedClone DTO** — Decouples 6 printer implementations from AST internals. Unblocks Clone type consolidation.

4. **Update `docs/status/README.md`** — 4 months stale. Gives wrong impression of project health.

5. **Add `golangci-lint` nolint directives** for known pre-existing warnings (errcheck on syscall calls, err113 on test errors).

### Nice to Have

6. **Archive old status files** — 304 files in flat directory.
7. **TokenValue type** — Strong typing for token values.
8. **SIMD ARM64** — 2 actionable TODOs.

---

## f) Top 25 Things to Do Next

| #  | Item                                                        | Impact   | Effort | Category     |
| -- | ----------------------------------------------------------- | -------- | ------ | ------------ |
| 1  | Fix `os.Exit(1)` in `statError` — return errors via channel | Critical | Medium | Bug fix      |
| 2  | Fix `prepareSubcommandArgs` — parse flags properly          | High     | Small  | Bug fix      |
| 3  | Add nolint directives for pre-existing lint warnings        | Medium   | Small  | Quality      |
| 4  | Update `docs/status/README.md` with real metrics            | Medium   | Small  | Docs         |
| 5  | Implement `TokenValue` type with validation                 | High     | Medium | Architecture |
| 6  | Create `ProcessedClone` DTO for printer decoupling          | High     | High   | Architecture |
| 7  | Consolidate 3 Clone types into unified types                | High     | High   | Architecture |
| 8  | Move `clone_classify.go` language maps behind interface     | Medium   | Medium | Architecture |
| 9  | Archive old `docs/status/` files (304)                      | Medium   | Small  | Cleanup      |
| 10 | Add `ConstantCSSProperty` position tracking upstream issue  | Medium   | Small  | Upstream     |
| 11 | Unify enum patterns across config/                          | Medium   | Medium | Refactor     |
| 12 | Refactor `transform.go` 300-line switch                     | Medium   | Medium | Refactor     |
| 13 | Implement remaining SIMD optimizations                      | Medium   | Medium | Performance  |
| 14 | Use `encoding/csv` for clone CSV output                     | Low      | Small  | Quality      |
| 15 | String interning for suffix tree tokens                     | Medium   | Medium | Performance  |
| 16 | Add integration test for `--all` flag                       | Medium   | Small  | Testing      |
| 17 | Add fuzz tests for templ parser edge cases                  | Medium   | Small  | Testing      |
| 18 | Fix remaining LSP hints                                     | Low      | Small  | Quality      |
| 19 | Add benchmark: in-process vs subprocess BDD                 | Medium   | Small  | Testing      |
| 20 | Profile memory usage during large repo analysis             | Medium   | Medium | Performance  |
| 21 | Consider `go:embed` for HTML template                       | Low      | Small  | Quality      |
| 22 | Add SARIF output integration test                           | Low      | Small  | Testing      |
| 23 | Evaluate `a-h/templ` v3 for `ConstantCSSProperty` Range     | Low      | Small  | Upstream     |
| 24 | Create GitHub release workflow                              | Medium   | Medium | CI/CD        |
| 25 | Add `--version` flag registration in `NewRootCommand`       | Low      | Small  | Bug fix      |

---

## g) Top #1 Question I Cannot Figure Out Myself

**Should the `os.Exit(1)` removal in `statError` change the CLI's exit code behavior for non-existent paths?**

Currently, `art-dupl /nonexistent` prints an error to stderr and exits with code 1. If we refactor `statError` to return errors through the goroutine-based crawl pipeline, the CLI will return a Cobra error (non-zero exit) instead. But the crawl runs in a goroutine, and the main analysis pipeline (`executeAnalysis`) doesn't wait for crawl errors before proceeding.

The fundamental question: **should non-existent paths be a hard error (exit 1), or should they produce empty output with a warning?**

- **Hard error (current behavior):** Matches user expectations for CLI tools
- **Empty output + warning:** Simpler to implement, matches how `grep` handles non-existent paths

This affects:

1. The error channel design for the goroutine-based crawler
2. Whether `executeAnalysis` should block on crawl completion before reporting
3. All error handling BDD tests that currently expect graceful handling

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
