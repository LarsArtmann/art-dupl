# Code Deduplication Session Status Report

**Date:** 2026-04-16 08:51  
**Branch:** fork  
**Commits this session:** 6 (5 by AI + 1 from prior session continuation)

---

## Executive Summary

Systematic code deduplication of the art-dupl codebase. Started from **145 clone groups** (threshold 15), reduced to **139 clone groups**. While the raw count reduction seems modest, the **net lines deleted** is significant: **99 insertions, 171 deletions** (net -72 lines). Many remaining clones are inherent Go patterns (table tests, interface implementations, explicit error handling).

---

## A) FULLY DONE

### Commit `2201f5d` (prior session)

- Extracted `withThreshold` helper in `cmd/run_analysis.go:407-414` — eliminated anonymous function duplication

### Commit `0e203b2` — Replace syntax.Node literals with testutil.CreateNodeWithPos

- **File:** `printer/stats_test.go`
- Replaced all `&syntax.Node{Filename: "file1.go", Pos: 1, End: 3, Type: 1}` inline literals with `testutil.CreateNodeWithPos(1, "file1.go", 1, 3)` calls
- Eliminated 13 clone instances
- Used existing helper discovered in `internal/testutil/node.go:61-70`

### Commit `015e4b7` — Add newTestClone helper in printer/diff_test.go

- **File:** `printer/diff_test.go`
- Added `newTestClone(filename, lineStart, lineEnd, fragment)` helper
- Replaced 6 clone struct literals with helper calls

### Commit `57e3ee8` — Replace local assertion helpers with testutil imports

- **Files:** `printer/stats_test.go`, `printer/diff_test.go`, `internal/testutil/assert.go`
- Deleted local `assertEqual` in stats_test.go, `assertCount`/`assertCloneCount`/`assertStringContains`/`assertStringNotContains` in diff_test.go
- Replaced with `testutil.AssertEqual`, `testutil.AssertCount`, `testutil.AssertStringContains`, `testutil.AssertStringNotContains`
- Added `AssertStringNotContains` to `internal/testutil/assert.go`

### Commit `470a5fa` — Replace duplicated test loggers with logger.NoOpLogger

- **Files:** `pkg/artdupl/detector_validation_test.go`, `pkg/artdupl/basic_test.go`, `pkg/artdupl/detector_uncovered_test.go`, `examples/examples_test.go`
- Deleted 3 identical test logger structs (`testLogger`, `testLoggerBasic`)
- Replaced with existing `logger.NoOpLogger{}` from `pkg/logger/logger.go:106-111`
- Key discovery: `pkg/artdupl/types.go:143` has `type Logger = logger.Logger` (type alias), so `&logger.NoOpLogger{}` works directly

### Commit `16f7b7f` — Use isEmptyOrLessThanEmpty in sortClonesByFilename

- **File:** `printer/sorter.go`
- Replaced inline empty-slice check in `sortClonesByFilename` with existing generic `isEmptyOrLessThanEmpty` helper
- Consistent with `SortClonesByHash` and `SortClonesBySize` patterns

---

## B) PARTIALLY DONE

None — all started work was completed and committed.

---

## C) NOT STARTED

### Intentionally Skipped (low ROI)

- **bdd/sorting_test.go runSortTest extraction**: 5 instances of `RunArtDuplWithFlags(map[string]string{...})` with different params — inherent test pattern, not real duplication
- **pkg/artdupl ErrNoFilesProvided assertion helper**: 5 instances of standard 3-line Go error checks — inherent Go idiom, extracting would add indirection without clarity
- **FileReaderFunc inline closures**: 4 instances with different behaviors each time — not true duplication

### Not Yet Attempted (from remaining 139 clone groups)

See section F for prioritized list.

---

## D) TOTALLY FUCKED UP

### Import Cycle Discovery (recovered)

- **Issue:** Attempted to replace `assertEqual` in `errors/types_test.go` with `testutil.AssertEqual`
- **Root cause:** Import cycle: `errors` → `internal/testutil` → `internal/utils` → `errors`
- **Resolution:** Reverted the change. The `assertEqual` in errors tests must remain local due to package dependency structure.
- **Lesson:** Check import graphs before attempting cross-package helper extraction.

### Pre-existing BDD Test Failures (NOT caused by our changes)

6 BDD tests fail on the `fork` branch — all are **pre-existing** and unrelated to our deduplication work:

1. `bdd/cli_commands_test.go:434` — help documentation examples test
2. `bdd/filter_features_test.go:111` — sqlc filter test
3. `bdd/filter_features_test.go:345` — include patterns test
4. `bdd/filter_features_test.go:406` — multiple include patterns test
5. `bdd/filter_features_test.go:471` — exclude patterns test
6. `bdd/filter_features_test.go:523` — combined include/exclude test

Verified: these fail identically without our 6 commits applied.

---

## E) WHAT WE SHOULD IMPROVE

### Process Improvements

1. **Check import graphs first** — before attempting to extract helpers across packages, verify no import cycles exist. Could use `go list -f '{{.ImportPath}}: {{.Imports}}' ./...` to map dependencies.
2. **Use the tool itself more** — run `art-dupl --semantic --sort total-tokens -t 15` after each commit to track progress, not just at session boundaries.
3. **Accept inherent patterns earlier** — many "clones" are fundamental Go patterns (table tests, interface implementations, explicit error handling). Recognizing these faster saves analysis time.

### Architecture Observations

1. **`testutil` is in `internal/`** — this creates a ceiling on what packages can use it. Packages that `internal/` depends on (like `errors`, `internal/utils`) cannot import `testutil`. Consider a `pkg/testutil` for broader use.
2. **Type aliases underused** — `pkg/artdupl/types.go:143` has `type Logger = logger.Logger`, but test files defined their own identical structs. This pattern could be checked more systematically.
3. **Test helper sprawl** — `internal/testutil/` has many files. Some helpers (like `AssertStringNotContains`) are only used by one test file. Consider if they belong in testutil or should stay local.

---

## F) Top 25 Things We Should Get Done Next

### HIGH IMPACT (real code, not tests)

| #   | What                                                                                      | Files                                                      | Impact | Effort |
| --- | ----------------------------------------------------------------------------------------- | ---------------------------------------------------------- | ------ | ------ |
| 1   | Fix 6 failing BDD tests (pre-existing)                                                    | `bdd/filter_features_test.go`, `bdd/cli_commands_test.go`  | HIGH   | Medium |
| 2   | Extract printer interface initialization pattern (7 clones)                               | `printer/{html,json,plumbing,printer,sarif,stats,text}.go` | HIGH   | Medium |
| 3   | Deduplicate `printer/sorter.go` filename+pos comparison (4 clones at lines 90,94,157,161) | `printer/sorter.go`                                        | Medium | Low    |
| 4   | Extract AST traversal pattern in `syntax/golang/transform.go` (4 clones)                  | `syntax/golang/transform.go`                               | Medium | Medium |
| 5   | Deduplicate hash/detector.go and hash/file_detector.go channel send pattern (3 clones)    | `hash/detector.go`, `hash/file_detector.go`                | Medium | Low    |
| 6   | Extract `cmd/stats.go` repeated stats validation (2 clones at 190-192, 209-211)           | `cmd/stats.go`                                             | Low    | Low    |
| 7   | Extract `printer/diff.go` repeated clone comparison (3 clones at 132, 142, 181)           | `printer/diff.go`                                          | Medium | Low    |

### MEDIUM IMPACT (test code)

| #   | What                                                                               | Files                                     | Impact | Effort |
| --- | ---------------------------------------------------------------------------------- | ----------------------------------------- | ------ | ------ |
| 8   | Extract `assertErrNoFiles` helper in detector_validation_test.go (5 clones)        | `pkg/artdupl/detector_validation_test.go` | Medium | Low    |
| 9   | Deduplicate `pkg/position/lines_test.go` table test patterns (18 clones)           | `pkg/position/lines_test.go`              | Medium | Medium |
| 10  | Extract `internal/utils/file_test.go` assertion pattern (6 clones)                 | `internal/utils/file_test.go`             | Medium | Low    |
| 11  | Extract `internal/utils/utils_test.go` assertion pattern (6 clones)                | `internal/utils/utils_test.go`            | Medium | Low    |
| 12  | Deduplicate `domain/domain_types_test.go` test struct patterns (8 clones)          | `domain/domain_types_test.go`             | Medium | Medium |
| 13  | Extract `git/change_detector_test.go` temp dir setup pattern (3 clones)            | `git/change_detector_test.go`             | Low    | Low    |
| 14  | Deduplicate `bdd/semantic_detection_test.go` assertion patterns (6 clones)         | `bdd/semantic_detection_test.go`          | Medium | Low    |
| 15  | Extract `bdd/detection_methods_test.go` file creation pattern (8 clones)           | `bdd/detection_methods_test.go`           | Medium | Medium |
| 16  | Deduplicate `pkg/artdupl/detector_types_test.go` assertion patterns (11 clones)    | `pkg/artdupl/detector_types_test.go`      | Medium | Low    |
| 17  | Extract `bdd/plumbing_output_test.go` verification patterns (4+3 clones)           | `bdd/plumbing_output_test.go`             | Medium | Low    |
| 18  | Deduplicate `internal/enum/marshal_test.go` test body patterns (5+3 clones)        | `internal/enum/marshal_test.go`           | Medium | Medium |
| 19  | Extract `errors/types_test.go` + `errors/marshal_test.go` test patterns (6 clones) | `errors/`                                 | Low    | Low    |
| 20  | Deduplicate `cache/file_cache_test.go` assertion patterns (4 clones)               | `cache/file_cache_test.go`                | Low    | Low    |

### INFRASTRUCTURE / PROCESS

| #   | What                                                                     | Files              | Impact               | Effort |
| --- | ------------------------------------------------------------------------ | ------------------ | -------------------- | ------ | --- |
| 21  | Move `internal/testutil` to `pkg/testutil` to break import cycle barrier | All test files     | HIGH                 | HIGH   |
| 22  | Add CI gate for clone count (`art-dupl -t 15                             | grep -c "^found"`) | `.github/workflows/` | Medium | Low |
| 23  | Fix `gopls unusedparams` diagnostics across codebase                     | Multiple files     | Low                  | Low    |
| 24  | Run `just ci` to full green (currently blocked by BDD failures)          | All                | HIGH                 | Medium |
| 25  | Create `FEATURES.md` or `CHANGELOG.md` documenting recent improvements   | Root               | Low                  | Low    |

---

## G) Top #1 Question I Cannot Figure Out Myself

**What is the intended behavior for the 6 failing BDD tests?**

Specifically:

- `filter_features_test.go:111` — Should `--filter-generated` auto-detect and exclude sqlc files? What changed?
- `filter_features_test.go:345/406/471/523` — Should include/exclude pattern flags work? Are they implemented or planned?
- `cli_commands_test.go:434` — What examples should appear in help output?

These tests may be testing features that are **not yet implemented** (spec tests) or they may have **regressed** due to other changes. I cannot determine intent from the test code alone. Should these tests be fixed, skipped, or are they tracking unimplemented features?

---

## Scan Results Summary

**Before this session:** 145 clone groups (threshold 15)  
**After this session:** 139 clone groups (threshold 15)  
**Net reduction:** 6 clone groups  
**Net lines:** +99 / -171 (net -72 lines deleted)

### Clone Distribution (remaining 139 groups)

| Category                  | Count | Example                                                |
| ------------------------- | ----- | ------------------------------------------------------ |
| Table test boilerplate    | ~35   | `pkg/position/lines_test.go` (18 clones alone)         |
| Test assertions/helpers   | ~30   | Various `*_test.go` files                              |
| BDD test setup patterns   | ~20   | `bdd/sorting_test.go`, `bdd/detection_methods_test.go` |
| Interface implementations | ~10   | Logger, printer initializers                           |
| AST/pipeline patterns     | ~8    | `syntax/golang/transform.go`, `hash/detector.go`       |
| Sorter comparisons        | ~6    | `printer/sorter.go`                                    |
| Other                     | ~30   | Various small patterns across codebase                 |

### Accepted Inherent Clones

These patterns are fundamental to Go and should NOT be deduplicated:

- **Table test cases** with similar structure but different test data
- **Interface method implementations** (Logger, Printer, etc.)
- **Explicit error handling** (`if err != nil { return err }`)
- **AST node traversal** in syntax/golang/transform.go
- **Standard library patterns** (sort.Slice closures, test cleanup)
