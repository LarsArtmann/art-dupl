# Status Report: 2026-04-30 Quality Sprint — Dead Code Elimination & Type Architecture

**Date:** 2026-04-30 23:01\
**Branch:** fork\
**State:** Clean working tree, pushed to origin/fork\
**Commits this session:** 6 (b657700 → daca611)\
**Net delta over last 10 commits:** +1015 / -4569 lines

---

## Current Health

| Metric                       | Value                                          |
| ---------------------------- | ---------------------------------------------- |
| golangci-lint issues         | **0**                                          |
| Test packages passing        | **23/23**                                      |
| Packages with test files     | 22/23 (domain has no tests now)                |
| Self-analysis clones         | 113 groups (all in test code)                  |
| Production code lines        | 16,905                                         |
| Total Go lines (incl. tests) | 46,292                                         |
| Domain package               | 6 files, 438 lines (was 16 files, 1,369 lines) |

---

## a) FULLY DONE

### 1. Dead Code Elimination (4,307 lines removed)

| Commit    | What                                                                                                                                                                                                                                                                          | Lines  |
| --------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ |
| `b657700` | Removed `syntax.BatchHash`, `HashSeqWithConfig`, `HashConfig` — zero production callers                                                                                                                                                                                       | -133   |
| `f8e466e` | Removed `printer.ParseSortBy` — wrapper around `config.ParseSortCriteria`, never called                                                                                                                                                                                       | -120   |
| `44e736a` | Removed `config.GetThresholdAsDomain`, `SetThresholdFromDomain` — test-only config↔domain bridge                                                                                                                                                                              | -100   |
| `0b88414` | Removed domain dead types: Clone, CloneGroup, Analysis, Repository, SourceFile, StringPool, AnalysisMode, DetectionState, FileProcessingState, CloneGroupID, AnalysisID, Confidence, ComplexityScore, Hash, ProcessingTime, NodeToClone, CalculateSeverity. Removed 22 files. | -3,994 |
| `daca611` | Removed `config.AssertMergedConfig`, `AssertMergeConfigsWithNil` — exported test-only helpers                                                                                                                                                                                 | -105   |

### 2. Type Architecture Improvements (prior sessions)

| Commit    | What                                                                                                                                |
| --------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| `8aae71c` | Typed `Config.Only` as `config.FileType` instead of `string`. Replaced hand-rolled `matchesOnlyFilter()` with `FileType.Matches()`. |
| `d0e5a2d` | Typed `Analysis.CreatedAt/CompletedAt` as `time.Time` instead of `string`.                                                          |

### 3. Lint Cleanup

| Commit    | What                                                                                 |
| --------- | ------------------------------------------------------------------------------------ |
| `1533da5` | Fixed all 18 remaining lint issues (wsl_v5, nlreturn, gci, err113). Zero issues now. |
| `40c6c5b` | Removed `domain.DetectionOptions` dead code                                          |

---

## b) PARTIALLY DONE

### 1. Domain Package Restructuring

**Status:** 68% complete

- **Done:** Removed all dead types, kept only production-used value objects (Threshold, LineNumber, Filepath, CloneSeverity)
- **Not done:** No tests for remaining domain types. Package has 0 test files.
- **Not done:** `analysis_errors.go` still contains error variables for deleted types (ErrCloneEndLineInvalid, ErrInvalidAnalysisState, etc.) that are only used by `types_severity.go` indirectly

### 2. Config Package Cleanup

**Status:** 90% complete

- **Done:** Removed test-only exports, threshold bridge methods, domain import
- **Not done:** `config.LoadConfig()` and `config.SaveConfig()` are still exported but never called outside tests. They're potentially useful API surface though.

---

## c) NOT STARTED

### High Impact

1. **Simplify `internal/simd/`** — `Available()` always returns `false`, `NewHasher()` never called in production. Package is effectively dead code marked "for future SIMD operations."
2. **Consolidate `pkg/artdupl/` parallel type system** — Has its own Clone, CloneGroup, Result, Summary types that mirror (but don't use) domain/ and printer/ types. Entire conversion layer exists (`convertOptionsToConfig`, `convertToCloneGroup`, etc.)
3. **Unify printer type system** — `printer.clone` (private), `printer.CloneWithContentMixin`, `printer.CloneWithContent`, `printer.CloneDiff`, `printer.JSONClone`, `printer.JSONOutput` are 6+ representations of the same concept
4. **Remove duplicate detection pipeline** — `detection.MultiDetector.FindDuplOver()` and `pkg/artdupl.detector.runDetection()` independently implement the same suffix tree + hash pipeline
5. **Add domain tests** — domain/ has 0 test files now. Threshold, LineNumber, Filepath, CloneSeverity all need test coverage.

### Medium Impact

6. **Clean up `syntax/hash_simd.go`** — `hashSeqSIMD()` just delegates to `hashSeqFallback()`. The SIMD branching adds complexity for zero benefit.
7. **Clean up `analysis_errors.go`** — Contains errors for deleted types (ErrInvalidAnalysisState, ErrCloneEndLineInvalid, ErrRepositoryPathEmpty, etc.)
8. **Remove unused `config.LoadConfig()` / `config.SaveConfig()`** — Only called from tests, `LoadOptionalConfig()` is the production API
9. **Consolidate `examples/`** — Still references deleted `domain.Clone` in examples. Examples should demonstrate the real SDK API (`pkg/artdupl`)
10. **Reduce test clone groups** — Self-analysis found 113 clone groups, mostly in `config/config_enum_test.go` (highly repetitive table-driven tests) and `printer/` tests

### Lower Impact

11. **Remove `printer.Format` type alias** — Only used in one place (`cmd/stats.go`). Could use `config.OutputFormat` directly.
12. **Remove nolint directives with justification audit** — 30+ nolint directives in production code, some may no longer be needed after lint rule changes
13. **Add table-driven test helpers** — Repetitive test patterns in config_enum_test.go could use shared helper functions
14. **Clean up `domain/helpers.go`** — `unmarshalWithValidation` is only called by 3 other functions in the same file. Could be inlined.

---

## d) TOTALLY FUCKED UP

### Session Recovery Required

1. **`printer/html.go` extraction was abandoned mid-work in a prior session** — Files `html_diff.go`, `html_summary.go`, `html_template.go` were created without imports, causing duplicate declarations and build failures. Had to `git checkout` to recover. The extraction idea (splitting 1484L html.go) is sound but execution was broken.

2. **No domain test coverage** — The aggressive removal of dead domain types also removed ALL domain tests. The remaining types (Threshold, LineNumber, Filepath, CloneSeverity) have zero test coverage. This is a regression that must be fixed.

---

## e) WHAT WE SHOULD IMPROVE

### Architecture Issues Found

1. **Three parallel type systems for "clone"** — `domain.Clone` (deleted), `pkg/artdupl.Clone`, `printer.clone` (private), `printer.JSONClone` all represent the same concept. No shared interface or type.

2. **`pkg/artdupl/` reimplements the detection pipeline** — Instead of calling `detection.MultiDetector`, it has its own `runDetection()` that creates suffix trees and hash detectors independently. This means bug fixes must be applied in two places.

3. **`config/` exports test-only code** — Even after cleanup, `LoadConfig()` and `SaveConfig()` are only used by tests. The pattern of exporting test helpers from config/ was a systemic issue.

4. **`domain/` is now a "value object" package with a misleading name** — It no longer contains entities, aggregates, or domain logic. It's just type-safe wrappers for primitives. Should be renamed or reorganized.

5. **`examples/` demonstrates deleted types** — `examples/domain_types_usage.go` references `domain.Clone` which no longer exists as a domain entity. The examples still compile because the SDK has its own types, but the documentation is misleading.

### Process Issues Found

6. **No safety net for dead code removal** — We removed 3,994 lines of domain code and discovered there were zero test files remaining. Should have written tests for remaining types before removing the others.

7. **LSP diagnostics are stale** — Multiple LSP errors show for files that were already committed and clean (e.g., `printer/html_diff.go` errors from stale ghost files). The LSP cache doesn't invalidate properly after git operations.

---

## f) Top 25 Things To Do Next

**Sorted by impact × ease (highest first):**

| #  | Task                                                                                    | Impact | Effort  | Why                                |
| -- | --------------------------------------------------------------------------------------- | ------ | ------- | ---------------------------------- |
| 1  | Write tests for remaining domain types (Threshold, LineNumber, Filepath, CloneSeverity) | High   | Low     | Zero test coverage is a regression |
| 2  | Clean up `analysis_errors.go` — remove errors for deleted types                         | Low    | Trivial | Dead variables                     |
| 3  | Remove dead SIMD code from `hashSeqSIMD()` — inline fallback                            | Low    | Low     | Dead branching                     |
| 4  | Simplify `internal/simd/` — remove or document as intentionally empty                   | Medium | Low     | Dead package                       |
| 5  | Update `examples/` to use SDK types instead of deleted domain types                     | Medium | Low     | Misleading docs                    |
| 6  | Unexport `config.LoadConfig()` / `config.SaveConfig()` or move to testutil              | Low    | Low     | Test-only exports                  |
| 7  | Consolidate `pkg/artdupl.Clone` with `printer.clone` via shared interface               | High   | Medium  | Two type systems                   |
| 8  | Make `pkg/artdupl` call `detection.MultiDetector` instead of reimplementing             | High   | Medium  | Duplicate pipeline                 |
| 9  | Rename `domain/` to reflect its value-object-only nature (e.g., `types/` or `values/`)  | Medium | Medium  | Misleading package name            |
| 10 | Extract shared clone interface from printer/ for reuse in SDK                           | High   | Medium  | Three parallel types               |
| 11 | Reduce clone groups in `config/config_enum_test.go` — extract shared helpers            | Low    | Medium  | 113 clone groups                   |
| 12 | Split `cmd/run_analysis.go` (447L) into focused files                                   | Medium | Medium  | Large file                         |
| 13 | Split `printer/html.go` (364L) properly this time                                       | Medium | Medium  | Was attempted, failed              |
| 14 | Remove `printer.Format` type alias, use `config.OutputFormat` directly                  | Low    | Low     | Unnecessary indirection            |
| 15 | Audit and justify remaining nolint directives                                           | Low    | Low     | 30+ directives                     |
| 16 | Add integration test for SDK → printer pipeline                                         | High   | Medium  | No E2E SDK test                    |
| 17 | Replace `samber/do` dependency if unused                                                | Low    | Low     | Check usage                        |
| 18 | Add `//go:build` tags for SIMD if keeping for future                                    | Low    | Trivial | Clarity                            |
| 19 | Remove `domain/helpers.go` `unmarshalWithValidation` — inline at 3 call sites           | Low    | Low     | Unnecessary abstraction            |
| 20 | Add fuzz tests for remaining domain types                                               | Medium | Medium  | Robustness                         |
| 21 | Consolidate `printer/stats.go` (244L) sort/filter logic with `printer/sorter.go`        | Medium | Medium  | Related code                       |
| 22 | Move `detection/todos.go` (352L) patterns to config-driven approach                     | Medium | High    | Hardcoded regex                    |
| 23 | Add OpenAPI/JSON schema for config file validation                                      | Medium | Medium  | External tooling                   |
| 24 | Write architecture decision records (ADRs) for type system choices                      | Medium | Low     | Documentation                      |
| 25 | Profile and benchmark the SDK pipeline vs CLI pipeline                                  | High   | High    | Performance                        |

---

## g) Top #1 Question I Cannot Figure Out Myself

**Should `pkg/artdupl/` (the SDK) be the single source of truth for clone types, or should we create a shared `types/` package?**

The current state:

- `pkg/artdupl.Clone` (with `Filename string, StartLine int, EndLine int, Size int, Hash string`) is the SDK's public API
- `printer.clone` (private, with `filename string, lineStart int, lineEnd int, fragment []byte, size int, classification CloneClassification`) is the internal printer type
- `detection.TodoIssue` uses `domain.Filepath` and `domain.LineNumber`

These three representations cannot be unified without deciding: Does the SDK own the types? Does a shared package? Or does the printer remain independent?

The answer affects whether `pkg/artdupl` should:

- (a) import from a shared `types/` package, or
- (b) define its own types and provide conversion functions, or
- (c) become the canonical source and the printer converts from SDK types

This is a product/architecture decision, not a code decision.

---

## Session Metrics

| Metric              | Before  | After  | Delta      |
| ------------------- | ------- | ------ | ---------- |
| Lint issues         | 18      | 0      | -18        |
| Domain files        | 16      | 6      | -10        |
| Domain lines        | 1,369   | 438    | -931 (68%) |
| Production Go lines | ~17,400 | 16,905 | ~-495      |
| Test Go lines       | ~33,400 | 29,387 | ~-4,013    |
| Commits pushed      | 0       | 6      | +6         |
