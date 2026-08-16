# Comprehensive Code Quality Status Report

**Date:** 2026-04-16 22:01\
**Branch:** `fork` (up to date with `origin/fork`)\
**Test Suite:** 28/28 packages PASS\
**Lint:** 0 issues (`just check` clean)\
**Working Tree:** Clean

---

## Executive Summary

Multi-session code quality improvement effort across `art-dupl` and its dependency `gogenfilter`.\
The project is in **excellent shape**: all tests green, zero lint issues, meaningful deduplication\
completed, and all pre-existing BDD test failures fixed at their root causes.

---

## A) FULLY DONE ✅

### 1. BDD Test Failure Fixes (6 tests, 3 root causes)

All 6 pre-existing BDD test failures identified, root-caused, and fixed:

| Test                                          | Root Cause                                                                                                                  | Fix                                                                                            |
| --------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| `cli_commands_test.go:434` (help examples)    | Help output uses "EXAMPLES" uppercase; test only checked "Example"/"example"                                                | Added `ContainSubstring("EXAMPLES")`                                                           |
| `filter_features_test.go` — `buildFilterCmd`  | `binaryPath` passed both as executable AND as first positional arg → CLI treated binary path string as directory to analyze | Removed `binaryPath` from args, only pass `tmpDir`                                             |
| `filter_features_test.go` — SQLC detection    | Test created `sqlc_models.go`/`sqlc_other.go` but detector only matches `models.go`, `querier.go`, etc.                     | Renamed to standard SQLC filenames                                                             |
| `filter_features_test.go` — pattern matching  | Absolute paths from temp dirs didn't match relative patterns like `pkg1/*`                                                  | Fixed in `gogenfilter/pattern.go` — prepend `**/` for relative patterns against absolute paths |
| `filter_features_test.go` — content detection | `fs.ReadFile(os.DirFS("."), "/absolute/path")` fails silently                                                               | Fixed in `gogenfilter/detection.go` — `os.ReadFile` fallback                                   |
| `filter_features_test.go` — SQLC assertion    | Checked `Not(ContainSubstring("sqlc"))` which was imprecise                                                                 | Changed to `Not(ContainSubstring("models.go"))`                                                |

**Commit:** `130dc43` (art-dupl), `32ddc53` (gogenfilter)

### 2. gogenfilter Library Fixes (pushed to master)

- **Pattern matching for absolute paths:** `MatchPattern("pkg1/*", "/tmp/xxx/pkg1/file.go")` now works by prepending `**/` to relative patterns when matching absolute paths
- **Empty segment handling:** `matchPathPattern` strips leading empty segment from absolute path splits (`"/a/b"` → `["a", "b"]`)
- **Content detection fallback:** When `fs.ReadFile(fsys, filePath)` fails (common with `os.DirFS(".")` and absolute paths), falls back to `os.ReadFile(filePath)`

**Commit:** `32ddc53` — pushed to `origin/master`

### 3. Code Deduplication — Refactoring

- **`printer/sorter.go`:** Extracted `compareByNameThenPos` helper used by `SortClonesByHash`, `sortClonesByFilename`, and `byNameAndLine.Less` — eliminated 3 copies of filename+position comparison logic
- **`cmd/stats.go`:** Consolidated 8 repeated `if sp, ok := p.(printer.StatsPrinter); ok { ... }` type assertions into a single block
- **`cmd/run_analysis.go`:** Extracted `withThreshold` helper to eliminate anonymous function duplication in `createPrinter`
- **`printer/sorter.go`:** Used `isEmptyOrLessThanEmpty` in `sortClonesByFilename` (was already used by other sort functions)

**Commits:** `4b228be`, `2201f5d`, `16f7b7f`

### 4. Test Infrastructure Improvements

- **`internal/testutil`:** Centralized assertion patterns (`AssertEqual`, `AssertContains`, etc.)
- **`printer/diff_test.go`:** Added `newTestClone` helper to reduce test clone construction duplication
- **Replaced `syntax.Node` literals** with `testutil.CreateNodeWithPos` across printer tests
- **Replaced local assertion helpers** with `testutil` imports across multiple test files
- **Replaced duplicated test loggers** with `logger.NoOpLogger`

**Commits:** `0f418cb`, `2af5621`, `015e4b7`, `0e203b2`, `57e3ee8`, `470a5fa`

### 5. Lint Cleanup — Zero Issues

Fixed all lint issues across the codebase:

- `nlreturn` — blank lines before return statements
- `golines` — long lines broken properly
- `gci` — import ordering in `printer/stats.go`
- `nolintlint` — removed unused `gocyclo`/`cyclop` from nolint directives

**Commit:** `6073ba2`

### 6. Stats Refactoring

- **`printer/stats_collector.go`:** Replaced `getTokenRange` switch statement with table-driven approach
- **Test config refactoring:** Centralized test configuration construction in `detection/detection_test.go`

**Commits:** `b2a829a`, `aa29471`

### 7. Config Builder Refactoring

- Extracted `applyBooleanFlags`, `applyPatternFlags`, `applyTimeout`, `applyDiffMode` helper functions from monolithic `BuildConfigFromFlags`

**Commit:** `b14eba0`

---

## B) PARTIALLY DONE ⚠️

### 1. Clone Scan Analysis

A full `art-dupl --semantic -t 15` scan was run. Results: **138 clone groups** detected.\
The analysis categorized them into:

- **Test-only clones** (~110 groups): Table test cases, BDD setup patterns, assertion patterns — accepted
- **Interface implementations** (~7 groups): `PrintClones` across printer types — accepted
- **Small genuine clones** (~21 groups): Minor patterns in production code

The scan was completed and analyzed but **no formal document was committed** cataloging accepted vs. actionable clones.

### 2. Type Model Improvements

- Considered but not yet started: stronger typing for `Pos` (currently `int32` mixed with `int` conversions)
- `clone` struct in `printer/` vs `syntax.Node` — dual representation still exists
- `domain.Clone` vs `syntax.Match` — overlapping concepts could be unified

---

## C) NOT STARTED 📋

### High-Value Items

1. **`pkg/position/lines_test.go` dedup** — 18 clones from repetitive test table entries; could use subtests or shared test data
2. **`bdd/semantic_detection_test.go` dedup** — 9 clones from repetitive BDD setup patterns
3. **`printer/` interface method dedup** — `PrintClones` has similar boilerplate across 7 printers; consider template method or shared base
4. **`hash/detector.go` + `hash/file_detector.go` + `pkg/artdupl/detector_pipeline.go`** — 3 production code clones
5. **`domain/domain_types_test.go`** — 8 clones from repetitive test assertions
6. **`internal/filtertest/` integration tests** — 5 clones from similar test setup patterns
7. **`detection/todos.go`** — 5 clones from repetitive pattern matching
8. **`syntax/golang/transform.go`** — 4 clones from AST node handling patterns
9. **SDK/API public surface audit** — `SDK_DESIGN.md` exists but implementation unclear
10. **Error type unification** — `errors/` package vs `pkg/errors/` coexistence

### Medium-Value Items

11. **`bdd/plumbing_output_test.go`** — 8 clones from output parsing patterns
12. **`pkg/artdupl/detector_*_test.go`** — 11 clones from test construction
13. **`internal/enum/marshal_test.go`** — 6 clones from enum serialization tests
14. **`printer/stats_visualization.go`** — 2 clones from chart rendering
15. **`migration/migration.go`** — 2 clones from version handling
16. **`git/change_detector.go`** — 2 clones from file change detection
17. **Coverage threshold enforcement** — `just check-coverage` exists but no CI gate
18. **Benchmark regression tracking** — benchmarks exist but no baseline comparison
19. **`examples/examples_sdk_demo.go`** — 2 clones from demo construction

### Lower-Value Items

20. **`suffixtree/suffixtree_test.go`** — 6 clones from tree construction tests
21. **`internal/utils/utils_test.go`** — 6 clones from utility tests
22. **`cli/runtime_test.go`** — 4 clones from CLI runtime tests
23. **`config/config_test.go`** — 4 clones from config parsing tests
24. **`printer/diff_test.go`** — 6 clones from diff output tests
25. **`cmd/cmd_test.go` + `cmd/cmd_utils_test.go`** — 7 clones from command tests

---

## D) TOTALLY FUCKED UP 💥

### Nothing is fundamentally broken.

All tests pass. All lint is clean. The codebase compiles and functions correctly. The `gogenfilter` dependency has been fixed and pushed.

**Pre-existing issues that were fixed this session:**

- `gogenfilter` had a fundamental architectural issue: `os.DirFS(".")` is incompatible with absolute paths passed to `fs.ReadFile` — this is a well-known Go gotcha that affected all filter tests using temp directories
- `buildFilterCmd` had a subtle double-path bug that made all filter pattern tests silently pass with 0 results (the tests were "passing" but not actually testing anything meaningful)

---

## E) WHAT WE SHOULD IMPROVE 🔧

### Architecture

1. **`Pos` type inconsistency** — `syntax.Node.Pos` is `int32`, but `clone.lineStart` is `int`, requiring explicit conversions (`int(dups[i][0].Pos)`). Should pick one and stick with it.
2. **Dual clone representation** — `printer/clone` struct duplicates information already in `syntax.Node`. A unified domain type would eliminate the conversion layer.
3. **`errors/` vs `pkg/errors/`** — Two error packages exist. Unclear boundary between them.
4. **`lib/` package** — Documented as "being phased out" but still exists. Should complete the migration.

### Testing

5. **BDD test helpers** — Many BDD tests repeat identical setup patterns (create dir, write files, run binary, parse output). A richer BDD fixture library would cut clones significantly.
6. **Test coverage gate** — No CI enforcement of coverage threshold despite `just check-coverage` existing.
7. **Integration test dedup** — `internal/filtertest/` and `bdd/` have overlapping test scenarios.

### Code Quality

8. **`printer/` package size** — 14 files, complex hierarchy. The `stats_*.go` split is good but `text.go` is still large.
9. **`cmd/stats.go`** — Still has `//nolint:funlen` despite consolidation. Could extract more helpers.
10. **`detection/todos.go`** — Repetitive pattern matching logic that could use a table-driven approach.

### Dependencies

11. **`samber/do`** — Listed in AGENTS.md as DI framework but usage unclear in current codebase.
12. **`lib/` legacy** — Should audit what's left and complete phase-out.

---

## F) TOP 25 THINGS TO DO NEXT

Ranked by **impact × effort ratio** (highest ROI first):

| #  | Task                                                                                                 | Impact | Effort | Category     |
| -- | ---------------------------------------------------------------------------------------------------- | ------ | ------ | ------------ |
| 1  | Unify `Pos` type: make `compareByNameThenPos` use `int32` consistently or change `Node.Pos` to `int` | High   | Low    | Architecture |
| 2  | Extract BDD test fixtures into shared helpers (create-dir-write-run-parse pattern)                   | High   | Medium | Testing      |
| 3  | Dedup `pkg/position/lines_test.go` — 18 clones, table-driven refactor                                | Medium | Low    | Testing      |
| 4  | Dedup `bdd/semantic_detection_test.go` — 9 clones from repetitive setup                              | Medium | Low    | Testing      |
| 5  | Production clone: `hash/detector.go` + `hash/file_detector.go` + `detector_pipeline.go` (3 clones)   | High   | Medium | Production   |
| 6  | Production clone: `detection/todos.go` — 5 clones from repetitive patterns                           | High   | Medium | Production   |
| 7  | Production clone: `syntax/golang/transform.go` — 4 clones from AST handling                          | High   | Medium | Production   |
| 8  | Production clone: `printer/diff.go` — 3 clones from diff rendering                                   | Medium | Medium | Production   |
| 9  | Eliminate `printer/clone` + `syntax.Node` dual representation                                        | High   | High   | Architecture |
| 10 | Merge or clarify `errors/` vs `pkg/errors/` boundary                                                 | Medium | Medium | Architecture |
| 11 | Complete `lib/` phase-out migration                                                                  | Medium | Medium | Cleanup      |
| 12 | Dedup `domain/domain_types_test.go` — 8 clones from test assertions                                  | Medium | Low    | Testing      |
| 13 | Dedup `internal/filtertest/` — 5 clones from integration test setup                                  | Medium | Low    | Testing      |
| 14 | Dedup `bdd/plumbing_output_test.go` — 8 clones from output parsing                                   | Medium | Low    | Testing      |
| 15 | Dedup `pkg/artdupl/detector_*_test.go` — 11 clones from test construction                            | Medium | Low    | Testing      |
| 16 | Dedup `internal/enum/marshal_test.go` — 6 clones from enum serialization                             | Low    | Low    | Testing      |
| 17 | Add CI coverage gate (`just check-coverage` in GitHub Actions)                                       | High   | Low    | CI/CD        |
| 18 | Add benchmark regression tracking with baseline comparison                                           | Medium | Medium | CI/CD        |
| 19 | Extract shared setup from `bdd/sorting_test.go` — 5 clones from sort verification                    | Low    | Low    | Testing      |
| 20 | Extract shared setup from `internal/utils/file_test.go` — 6 clones from file tests                   | Low    | Low    | Testing      |
| 21 | Refactor `migration/migration.go` — 2 clones from version handling                                   | Low    | Low    | Cleanup      |
| 22 | Refactor `printer/stats_visualization.go` — 2 clones from chart rendering                            | Low    | Low    | Cleanup      |
| 23 | Review `examples/examples_sdk_demo.go` — 2 clones from demo construction                             | Low    | Low    | Docs         |
| 24 | Audit `cmd/cmd_test.go` + `cmd/cmd_utils_test.go` — 7 clones from command tests                      | Low    | Low    | Testing      |
| 25 | Add fuzz tests for `gogenfilter/pattern.go` (the code we just fixed)                                 | Medium | Medium | Robustness   |

---

## G) TOP #1 QUESTION I CANNOT FIGURE OUT MYSELF ❓

**Should `syntax.Node.Pos` remain `int32` or change to `int`?**

The current codebase has a tension:

- `syntax.Node.Pos` and `syntax.Node.End` are `int32` (likely for memory efficiency in the suffix tree)
- `clone.lineStart` and `clone.lineEnd` are `int`
- Every call to `compareByNameThenPos` requires `int(dups[i][0].Pos)` conversions
- The suffix tree internally works with `Pos` as `int32` for its positions

Changing `Pos` to `int` would eliminate all conversions but may increase memory usage for large codebases (the suffix tree stores millions of positions). Keeping `int32` requires conversions everywhere but is more memory-efficient.

**Question for you:** Is there a specific memory or performance constraint that drove the `int32` choice, or would switching to `int` be acceptable? This decision affects whether we clean up the conversions or create a proper typed `Position` abstraction.

---

## Session Commits Summary

### art-dupl (branch: `fork`, pushed to `origin/fork`)

| Commit    | Description                                                                       |
| --------- | --------------------------------------------------------------------------------- |
| `130dc43` | `fix(bdd): fix help examples test and filter pattern tests`                       |
| `4b228be` | `refactor: consolidate repeated type assertions and extract compareByNameThenPos` |
| `6073ba2` | `style: fix lint issues (nlreturn, golines, gci, nolintlint)`                     |

### gogenfilter (branch: `master`, pushed to `origin/master`)

| Commit    | Description                                                            |
| --------- | ---------------------------------------------------------------------- |
| `32ddc53` | `fix: handle absolute paths in pattern matching and content detection` |

### Earlier session commits already pushed (for context):

| Commit    | Description                                                                       |
| --------- | --------------------------------------------------------------------------------- |
| `16f7b7f` | `refactor(printer): use isEmptyOrLessThanEmpty in sortClonesByFilename`           |
| `470a5fa` | `refactor(tests): replace duplicated test loggers with logger.NoOpLogger`         |
| `57e3ee8` | `refactor(tests): replace local assertion helpers with testutil imports`          |
| `015e4b7` | `refactor(printer): add newTestClone helper to reduce diff_test duplication`      |
| `0e203b2` | `refactor(printer): replace syntax.Node literals with testutil.CreateNodeWithPos` |
| `2201f5d` | `refactor(cmd/run_analysis): extract withThreshold helper`                        |
| `683a27f` | `test: add unit tests for errors, detector, and printer`                          |
| `0f418cb` | `refactor(tests): centralize assertion patterns across test files`                |
| `2af5621` | `refactor(tests): centralize config field assertions and improve test utilities`  |
| `b2a829a` | `refactor(stats): replace switch statement with table-driven approach`            |
| `b14eba0` | `refactor(config): extract boolean, pattern, timeout, and diff mode flag helpers` |

---

## Metrics

| Metric                            | Value                          |
| --------------------------------- | ------------------------------ |
| Go files                          | 236                            |
| Total Go lines                    | 49,378                         |
| Test packages                     | 28/28 passing                  |
| Lint issues                       | 0                              |
| Clone groups detected (t=15)      | 138                            |
| Commits this multi-session effort | 20+                            |
| Commits this session              | 3 (art-dupl) + 1 (gogenfilter) |
| BDD test failures fixed           | 6 → 0                          |
| Production code clones addressed  | 3 patterns                     |
| Test infrastructure improvements  | 6+ refactorings                |
