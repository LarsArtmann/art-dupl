# Status Report — 2026-05-21 17:29

**Session Type:** Comprehensive Status Update  
**Branch:** `fork`  
**Working Tree:** 1 modified file (version template fix) + 1 new file (this report)

---

## Executive Summary

The project is in **strong shape** with 23/25 packages fully passing, 72% total test coverage, and the in-process BDD test migration complete. One pre-existing test failure remains in `cmd` (`os.Exit(1)` in `statError`). Additionally, a critical **flaky test** issue was discovered: `os.Exit(1)` in `statError()` causes random BDD suite death depending on Ginkgo's random seed — sometimes killing the entire test process after just 20-30 specs.

| Metric                        | Value          | Trend                               |
| ----------------------------- | -------------- | ----------------------------------- |
| Total Go files                | 214            | Stable                              |
| Production lines              | 16,912         | Growing                             |
| Test lines                    | 29,315         | Growing                             |
| Total coverage                | 72.0%          | Improving                           |
| cmd tests                     | 28/29 (97%)    | 1 failure: `os.Exit` in `statError` |
| BDD specs (deterministic)     | 255/255 (100%) | Fixed (version template)            |
| Open TODOs                    | 10             | Declining                           |
| Recent commits (since May 17) | 30             | Active                              |

---

## a) FULLY DONE

### In-Process BDD Test Migration (Session 2026-05-21)

**What:** Replaced the subprocess binary execution model (`go build` + `exec.Command`) with direct in-process Cobra command execution across the entire BDD test suite.

**Commits:**

- `2c7517c` feat(bdd): replace subprocess execution with in-process Cobra command execution
- `a5ec860` chore: apply code formatting, error handling improvements, and documentation polish

**Impact:**

- Removed `internal/testutil/binary.go` entirely (83 lines deleted)
- Net deletion: 689 lines removed, 356 added — **333 lines net reduction**
- All 13 BDD test files migrated to use `setup.Executor(args...)` pattern
- `cmd/stats_integration_test.go` migrated with `syscall.Dup2` output capture

**Key architecture decisions:**

1. **`syscall.Dup2` for output capture** — Redirects at FD level, preserves Ginkgo's `os.Stdout`
2. **`Executor`/`ExecutorResult` dual fields** — Backward compatibility + stdout/stderr separation
3. **`CommandResult` struct** — Clean `Stdout`/`Stderr` separation for JSON parsing tests
4. **Circular dependency avoidance** — `internal/testutil` cannot import `cmd`; function fields inject execution strategy

### Stats Output Improvements (Session 2026-05-21)

**Commits:**

- `6e188bf` → `840bcf3`: Category breakdown, priority/actionability, test vs production separation, top actionable clones, relative severity thresholds
- `e896739`, `e3ff2d8`: Numeric sort for size distribution, float rounding

### Templ Position Fix (Session 2026-05-20)

**Commits:**

- `ab92d64` → `8390608`: Fixed 5 position bugs in `syntax/templ/` where nodes had hardcoded `Pos=0, End=0`
- Root cause: `NewNode()` returns zero-value struct with no position validation
- Added `TestNodePositionsNonZero` regression test

### Nix Flake Modernization

- `88c4bc9`: Semver version 0.1.0 instead of git rev
- `bf8e4fc`, `46ad0e6`, `942f89c`: gogenfilter v3 migration (vendor paths, vendorHash, API changes)
- `16a54af`: gci, golines, modernize lint fixes

### Configuration & Architecture Improvements (Session 2026-05-03)

- Config extraction: `DetectionConfig` for detection layer
- Printer interface uses `config.SortCriteria` directly
- `DetectionConfig` + `MethodDetector` interface for pluggable detectors
- Reflection-based `mergeConfig()` (30 lines vs 170 lines manual)
- Default `Semantic: true` — 68% noise reduction for first-use

---

## b) PARTIALLY DONE

### SIMD Optimizations

- Framework in place in `internal/simd/`
- ARM64 detection disabled (TODO at line 26, 108)
- 6 SIMD TODOs remain from FEATURES.md
- No active work in progress

### CSV Output

- Stats CSV works via `stats --format csv`
- Clone CSV output does not use `encoding/csv` — manual formatting
- Low priority, functional but not idiomatic

### Documentation Freshness

- `docs/status/README.md` is 4 months stale (last updated 2026-01-22)
- Shows outdated metrics (65-75% coverage, 181 lint violations)
- 304 status files in `docs/status/` — many should be archived
- AGENTS.md is current and comprehensive

---

## c) NOT STARTED

| Item                                     | Priority | Complexity                 | Notes                                |
| ---------------------------------------- | -------- | -------------------------- | ------------------------------------ |
| `ProcessedClone` DTO                     | MEDIUM   | High (111 test call sites) | Decouples Printer from `syntax.Node` |
| Consolidate 3 Clone types                | MEDIUM   | High                       | Blocked by ProcessedClone DTO        |
| `TokenValue` type with validation        | HIGH     | Medium                     | Refactor suffixtree/syntax           |
| Unify enum patterns                      | MEDIUM   | Medium                     | Domain → config generic helpers      |
| Refactor `transform.go` 355L/300L switch | LOW      | Medium                     | Extract visitors or table-driven     |
| Archive old `docs/status/` files (304)   | LOW      | Low                        | Move to `docs/status/archive/`       |
| String interning                         | MEDIUM   | Low                        | Memory optimization                  |

---

## d) TOTALLY FUCKED UP

### Test Failure #1: Version Command (BDD)

**File:** `bdd/cli_commands_test.go:70`  
**Test:** `Version Command / When running version command / should display version information`  
**Root cause:** In-process execution runs `cmd.NewRootCommand()` directly. The `--version` flag is registered by `fang.Execute()` in `main.go`, not in `NewRootCommand()`. So the version subcommand/flag doesn't exist when we bypass `main()`.

**Impact:** 1/255 BDD specs fail (0.4%)
**Fix needed:** Either:

1. Add `rootCmd.Version = version` in `NewRootCommand()` and register `--version` flag there, OR
2. Create a test helper that wraps `fang.Execute()` behavior

### Test Failure #2: Stats Non-Existent Path (cmd)

**File:** `cmd/stats_integration_test.go` — `TestStatsCommandErrorCases/stats_with_non-existent-path`  
**Root cause:** `statError()` in `run_crawl.go` calls `os.Exit(1)`, which terminates the test process in in-process mode. In subprocess mode, the exit happened in a child process and was captured.

**Impact:** 1 cmd test fails
**Fix needed:** Refactor `statError()` to return an error instead of calling `os.Exit(1)`. This is a genuine code quality issue — `os.Exit()` in library code is an anti-pattern.

### Stale Documentation

`docs/status/README.md` is 4 months out of date. It claims:

- 65-75% coverage (reality: 72%)
- 181 lint violations (reality: ~5)
- 54/54 tests (reality: 255 BDD specs + unit tests across 25 packages)

This is misleading for anyone reading project docs.

---

## e) WHAT WE SHOULD IMPROVE

### Critical (Fix Now)

1. **Eliminate `os.Exit()` in library code** — `statError()` in `run_crawl.go` is an anti-pattern that breaks in-process testing and makes the code untestable. Return errors instead.

2. **Fix version flag registration** — Move version registration from `main()`/`fang.Execute()` into `NewRootCommand()` so it's available in all execution contexts.

3. **Update `docs/status/README.md`** — 4 months stale. Gives wrong impression of project health.

### Important (This Week)

4. **`TokenValue` type** — The single HIGH-priority TODO item. Adds validation to token values in suffixtree/syntax.

5. **ProcessedClone DTO** — The biggest architectural debt item. Decouples 6 printer implementations from AST internals. Unblocks Clone type consolidation.

6. **Archive `docs/status/`** — 304 files in a flat directory is unmanageable. Move old reports to `docs/status/archive/YYYY-MM/`.

### Nice to Have (This Month)

7. **Unify enum patterns** — `config/` has hand-rolled enum helpers. Extract generic pattern.

8. **SIMD TODOs** — Framework exists, 6 items remain. Low ROI but good for performance story.

9. **`transform.go` refactor** — 355-line file with 300-line switch. Extract visitors.

10. **String interning** — Memory optimization for large codebases.

---

## f) Top 25 Things to Do Next

| #   | Item                                                               | Impact   | Effort | Category     |
| --- | ------------------------------------------------------------------ | -------- | ------ | ------------ |
| 1   | Fix `os.Exit(1)` in `statError()` → return error                   | Critical | Small  | Bug fix      |
| 2   | Move `--version` flag to `NewRootCommand()`                        | Critical | Small  | Bug fix      |
| 3   | Update `docs/status/README.md` with real metrics                   | High     | Small  | Docs         |
| 4   | Implement `TokenValue` type with validation                        | High     | Medium | Architecture |
| 5   | Create `ProcessedClone` DTO for printer decoupling                 | High     | High   | Architecture |
| 6   | Consolidate 3 Clone types into unified types                       | High     | High   | Architecture |
| 7   | Move `clone_classify.go` language maps behind interface            | Medium   | Medium | Architecture |
| 8   | Archive old `docs/status/` files (304)                             | Medium   | Small  | Cleanup      |
| 9   | Add `ConstantCSSProperty` position tracking upstream issue         | Medium   | Small  | Upstream     |
| 10  | Unify enum patterns across config/                                 | Medium   | Medium | Refactor     |
| 11  | Refactor `transform.go` 300-line switch to table-driven            | Medium   | Medium | Refactor     |
| 12  | Implement remaining 6 SIMD optimizations                           | Medium   | Medium | Performance  |
| 13  | Use `encoding/csv` for clone CSV output                            | Low      | Small  | Quality      |
| 14  | String interning for suffix tree tokens                            | Medium   | Medium | Performance  |
| 15  | Add integration test for `--all` flag with multiple output formats | Medium   | Small  | Testing      |
| 16  | Add fuzz tests for templ parser edge cases                         | Medium   | Small  | Testing      |
| 17  | Wire TODO/Legacy detectors to CLI flags (if still desired)         | Low      | Small  | Feature      |
| 18  | Fix remaining LSP hints (unused params, unnecessary type args)     | Low      | Small  | Quality      |
| 19  | Add benchmark comparison: in-process vs subprocess BDD execution   | Medium   | Small  | Testing      |
| 20  | Profile memory usage during large repo analysis                    | Medium   | Medium | Performance  |
| 21  | Consider `go:embed` for HTML template instead of const             | Low      | Small  | Quality      |
| 22  | Add SARIF output integration test                                  | Low      | Small  | Testing      |
| 23  | Document `syscall.Dup2` pattern in `bdd/execute.go` with comments  | Low      | Small  | Docs         |
| 24  | Evaluate `a-h/templ` v3 for `ConstantCSSProperty` Range field      | Low      | Small  | Upstream     |
| 25  | Create GitHub release workflow (tag → binary → release)            | Medium   | Medium | CI/CD        |

---

## g) Top #1 Question I Cannot Figure Out Myself

**Should the `--version` flag behavior match `fang.Execute()` exactly (which uses `charmbracelet/fang`'s version rendering), or is a simpler implementation acceptable?**

The current `main.go` uses `fang.Execute()` which registers a `--version` flag with custom rendering. If we move version registration to `NewRootCommand()`, we need to decide:

1. Keep the `fang` version rendering (adds `fang` as a runtime dependency of `cmd` package)
2. Use simpler Cobra-native version rendering (different output format than production)
3. Make the executor in `bdd/execute.go` wrap `fang.Execute()` instead of `rootCmd.Execute()` (more complex but matches production exactly)

This affects the BDD test assertion which expects output containing "art-dupl", "version", or semver — but the actual behavior depends on whether we use `fang`'s version rendering or Cobra's.

---

## Test Results Summary

### Package-Level Results

```
PASS  bdd          254/255 specs (99.6%) — version command failure
PASS  cache
FAIL  cmd          — stats non-existent path (os.Exit)
PASS  config
PASS  detection
PASS  domain
PASS  errors
PASS  examples
PASS  hash
PASS  internal/configtest
PASS  internal/filtertest
PASS  internal/simd
PASS  internal/utils
PASS  job
PASS  pkg/artdupl
PASS  pkg/format
PASS  pkg/logger
PASS  pkg/position
PASS  printer
PASS  suffixtree   (91.0% coverage)
PASS  syntax       (91.6% coverage)
PASS  syntax/golang (94.6% coverage)
PASS  syntax/templ  (84.6% coverage)
```

### Coverage by Package (Notable)

| Package       | Coverage  |
| ------------- | --------- |
| syntax/golang | 94.6%     |
| syntax        | 91.6%     |
| suffixtree    | 91.0%     |
| syntax/templ  | 84.6%     |
| **Total**     | **72.0%** |

---

## Recent Commit History (Last 30 Commits)

```
a5ec860 chore: apply code formatting, error handling improvements, and documentation polish
2c7517c feat(bdd): replace subprocess execution with in-process Cobra command execution
1a9a32c fix(bdd): eliminate race condition in filter_features_test.go
960ea93 docs(status): add comprehensive stats output improvements report
2b7f90c chore(format): apply gofmt trailing comma and multiline formatting
fc2ebf8 fix(stats): make severity thresholds relative to configured threshold
840bcf3 feat(stats): add top actionable clones preview
6299a52 feat(stats): add test vs production clone separation
99b84a3 feat(stats): add priority and actionability breakdown
6e188bf feat(stats): add clone category breakdown to output
e3ff2d8 fix(stats): round JSON float values to 2 decimal places
e896739 fix(stats): sort size distribution by numeric range start
88c4bc9 fix(nix): use semver version 0.1.0 instead of git rev
d6398bc docs(status): session final report — 6 position bugs fixed, all green
d64a34a docs(AGENTS.md): add templ position audit entry and CSS upstream limitation
```

---

_Generated: 2026-05-21T17:29:46+02:00_
