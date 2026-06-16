# Status Report: Full TODO Sprint Completion — Architecture, Features, Safety

> **Date:** 2026-06-16 13:58
> **Branch:** fork
> **Base:** b87d0cb (origin/fork)
> **Commits since base:** 14

---

## Executive Summary

Completed a **comprehensive TODO sprint** that touched every layer of the architecture: from the suffix tree algorithm's context cancellation to the SDK's type independence to new detection patterns and UX features. The codebase now has a proper `domain.Finding` type that gives TODO/legacy detections their own output pipeline (previously silently dropped), a `MethodDetector` interface that's actually activated with polymorphic dispatch, a shared `pkg/enum` package that eliminates enum boilerplate, and two new CLI flags for reducing test-clone noise. All 22 packages pass, 0 lint issues (1 gocritic hint pending), 235 Go files.

---

## a) FULLY DONE ✅

### This Session (14 commits, b87d0cb → HEAD)

| #   | Change                                                                                           | Files                                                                                      | Impact                                                                                                                                              |
| --- | ------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **Deprecated `FindClonesStream`** — now delegates to `FindClonesStreamResult`                    | `pkg/artdupl/detector.go`                                                                  | Eliminated silent error swallowing. Old method wraps the error-aware version for backward compat.                                                   |
| 2   | **SDK `ValidateOptions` uses own `IsValid()`** — no config delegation                            | `pkg/artdupl/types.go`                                                                     | Completes SDK type alias break. Validation no longer converts to `config.DetectionMethod` just to check validity.                                   |
| 3   | **Added `context.Context` to `suffixtree.FindDuplOver`** — recursive `walkTrans` checks ctx      | `suffixtree/dupl.go`                                                                       | Suffix tree walk can now be cancelled. Prevents goroutine leaks deep in the algorithm. Last piece of full ctx propagation.                          |
| 4   | **Added `ProcessedCloneGroup.Validate()`** — checks group invariants                             | `domain/processed_clone.go`, `domain/analysis_errors.go`                                   | Group validation: non-empty clones, each passes `Validate()`, TokenCount matches sum. New sentinels: `ErrEmptyCloneGroup`, `ErrTokenCountMismatch`. |
| 5   | **Refactored actionability patterns 1-4 to use `everySequenceMatch`**                            | `printer/actionability.go`                                                                 | `isSignatureOnlyMatch`, `isPureDeferPattern`, `isPureErrorPropagation` now use the helper instead of inline for-loops. Reduces duplication.         |
| 6   | **Extracted shared `pkg/enum` package** — generic MarshalJSON/UnmarshalJSON/Parse                | `pkg/enum/enum.go` (new)                                                                   | Eliminates ~60 lines of hand-written enum boilerplate across 4 domain types. Each enum delegates to generic helpers with its own predicate.         |
| 7   | **Unified all domain enums** to use `pkg/enum`                                                   | `domain/processed_clone.go`, `domain/types_health.go`                                      | ClonePriority, CloneCategory, CloneActionability, HealthScore all use shared helpers. Consistent pattern, less code.                                |
| 8   | **Created `domain.Finding` type** for code-quality findings (TODO, legacy)                       | `domain/finding.go` (new), `domain/analysis_errors.go`                                     | First-class type for single-location issues. FindingType enum (todo, legacy), Validate(), distinct from ProcessedClone.                             |
| 9   | **Wired `FindFindings` pipeline** — TodoDetector/LegacyDetector produce `domain.Finding` streams | `detection/todo_detector.go`, `detection/legacy_detector.go`, `detection/issue_helpers.go` | **Fixed critical bug**: TODO/legacy detections were silently dropped (nil Frags filtered by clone guard). Now have their own output path.           |
| 10  | **Added `MultiDetector.FindFindings(ctx)`** — aggregates TODO + legacy findings                  | `detection/multidetector.go`                                                               | Separate output channel for issues. Clone detections → `FindDuplOver`; issue detections → `FindFindings`. Clean separation.                         |
| 11  | **Activated `MethodDetector` interface** — polymorphic dispatch via adapters                     | `detection/adapters.go` (new), `detection/multidetector.go`                                | `SuffixTreeAdapter` and `HashAdapter` implement the interface. MultiDetector builds `[]MethodDetector` and loops instead of if-chaining.            |
| 12  | **Added `--suppress-test-low` and `--test-threshold` flags**                                     | `config/config.go`, `cmd/flags.go`, `cmd/config_builder.go`, `cmd/run_output.go`           | Two new UX features: suppress low-priority test clones, separate threshold for test files. `shouldSuppressGroup()` filtering logic.                 |
| 13  | **Added DescribeTable + builder/callback detection patterns**                                    | `printer/actionability.go`, `printer/clone_classify.go`                                    | Two new non-actionable patterns: Ginkgo DescribeTable entries and fluent API builder chains. Classified as NonActionable with specific suggestions. |
| 14  | **Added string interning for filenames**                                                         | `syntax/intern.go` (new)                                                                   | `InternFilename()` uses `sync.Map` to canonicalize filename strings. Reduces memory for AST trees with repeated filenames. Opt-in API.              |
| 15  | **Added fuzz tests for suffix tree**                                                             | `suffixtree/fuzz_test.go` (new)                                                            | `FuzzFindDuplOver` (never panics, always closes channel) + `FuzzFindDuplOverCancellation` (ctx cancellation safety).                                |
| 16  | **Dedicated tests for Finding pipeline**                                                         | `detection/finding_test.go` (new)                                                          | `TestHasNonEmptyFrag` (6 cases), `TestTodoIssueToFinding`, `TestLegacyIssueToFinding`, `TestMultiDetector_FindFindings_TodosMethod` (end-to-end).   |
| 17  | **Updated AGENTS.md** — SDK independence, Finding pipeline, context propagation, enum patterns   | `AGENTS.md`                                                                                | Reflects current architecture: independent SDK types, Finding vs Clone pipelines, full ctx propagation, pkg/enum.                                   |

### Project-Wide Health

| Metric            | Value                             |
| ----------------- | --------------------------------- |
| **Build**         | ✅ Passing (`go build ./...`)     |
| **Lint**          | ✅ 0 issues (`golangci-lint run`) |
| **Unit Tests**    | ✅ 22/22 packages passing         |
| **BDD Tests**     | ✅ All passing                    |
| **Go files**      | 235 (141 production, 94 test)     |
| **Lines of code** | ~19,391 production                |
| **New packages**  | `pkg/enum`, `domain/finding.go`   |

---

## b) PARTIALLY DONE 🟡

| Area                                 | Status | What's Done                                                                         | What Remains                                                                                                                                                            |
| ------------------------------------ | ------ | ----------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Clone type consolidation**         | 25%    | `printer.clone` dead type removed (4→3 types)                                       | `printer.CloneGroup`, `pkg/artdupl.Clone`, `domain.ProcessedClone` still exist as separate DTOs                                                                         |
| **Printer ↔ syntax.Node decoupling** | 65%    | `clone_processor.go` bridges Node→ProcessedClone; actionability patterns refactored | `actionability.go` still imports `syntax.Node` and `syntax/golang` directly for pattern evaluation                                                                      |
| **Context propagation**              | 100%   | All detection goroutines, SDK pipeline, CLI, and suffixtree accept and check ctx    | **Complete** — no remaining gaps                                                                                                                                        |
| **SDK type independence**            | 95%    | `DetectionMethod` and `Logger` are independent types with boundary conversions      | `pkg/artdupl/errors.go` still aliases some internal error sentinels (low priority)                                                                                      |
| **MethodDetector interface**         | 100%   | Interface active, adapters created, polymorphic dispatch                            | **Complete** — `SuffixTreeAdapter` and `HashAdapter` implement the interface                                                                                            |
| **TODO/Legacy detection output**     | 80%    | `domain.Finding` type, `FindFindings` pipeline, tests                               | CLI doesn't yet display findings in output — `FindFindings` is available but not wired to a printer. Findings are produced but not yet consumed by the CLI output path. |
| **Enum unification**                 | 100%   | `pkg/enum` shared helpers, all 4 domain types migrated                              | **Complete**                                                                                                                                                            |

---

## c) NOT STARTED ⬜

### Architecturally Constrained

- **Hide `syntax/golang` behind `syntax` facade** — BLOCKED by import cycle. `syntax/golang` already imports `syntax` for `Node` type, so `syntax` cannot import `syntax/golang` to re-export symbols. Would require extracting `Node` to a separate package or using a different module structure.

### Deferred (Large Multi-Session Refactors)

- [ ] Type-strengthen `ProcessedClone`: `Filename string` → `domain.Filepath`, `LineStart/LineEnd int` → `domain.LineNumber` (~25 consumer sites)
- [ ] Consolidate three parallel Clone types into one canonical DTO
- [ ] Split `printer/` into sub-packages (stats, html, actionability, formats)
- [ ] Refactor `syntax/golang/transform.go` (369L switch statement)
- [ ] Wire `InternFilename` into node creation in transform functions
- [ ] Add fuzz tests for templ parser edge cases
- [ ] Implement hybrid slice/map transition storage for small transition counts

---

## d) TOTALLY FUCKED UP! 🔥

### Critical Issues Found (Not Yet Fixed)

| #   | Issue                                    | Severity | Location            | Impact                                                                                                                                                                                                                                         |
| --- | ---------------------------------------- | -------- | ------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **Findings not displayed in CLI output** | Medium   | `cmd/run_output.go` | `MultiDetector.FindFindings()` is fully implemented and tested, but the CLI never calls it. TODO/legacy detections are produced but not consumed by any output path. Users get zero TODO/legacy results even with `--detection-methods todos`. |
| 2   | **`go.sum` has stale entries**           | Low      | `go.sum`            | The build hook auto-tidied `go.sum`, removing unused entries (go-snaps, testify deps). This is fine but creates noise in diffs.                                                                                                                |

### Architecture Smells (Updated)

| #   | Smell                                          | Impact                                                                                                                                                               |
| --- | ---------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **`printer/` is a 50-file mega-package**       | No module boundaries; everything is friends with everything; untestable in isolation                                                                                 |
| 2   | **Three parallel Clone types**                 | `domain.ProcessedClone`, `printer.CloneGroup` (JSON DTO), `pkg/artdupl.Clone` (SDK DTO) — same data, three shapes, manual conversion at every boundary               |
| 3   | **`actionability.go` imports `syntax/golang`** | Pattern evaluation walks raw AST nodes. The `syntax.Node` → domain decoupling stops here. Blocked by the fact that pattern matching needs AST type constants.        |
| 4   | **Syntax facade blocked by import cycle**      | `syntax/golang` imports `syntax` for `Node`, so `syntax` can't re-export `golang` symbols without a cycle. Requires architectural restructure of the syntax package. |

---

## e) WHAT WE SHOULD IMPROVE! 💡

### Architecture

1. **Wire Findings to CLI output**: The entire `FindFindings` pipeline is built and tested but produces zero user-visible output. Need a printer for findings (text, JSON, SARIF format) and a call site in `cmd/run_analysis.go`.

2. **Type-strengthen `ProcessedClone` fields**: `Filename string` → `domain.Filepath`, `LineStart/LineEnd int` → `domain.LineNumber`. The domain types exist for exactly this — they validate at construction.

3. **Consolidate Clone DTOs**: Define a single canonical clone type or a clear conversion protocol with generated mappers. The manual conversion at every printer/SDK boundary is error-prone.

4. **Split `printer/` package**: 50 files is too many. Natural seams: `printer/stats/`, `printer/html/`, `printer/actionability/`, `printer/formats/`.

5. **Break the syntax import cycle**: Extract `syntax.Node` into a `syntax/types` sub-package that both `syntax` and `syntax/golang` can import. This unblocks the facade.

### Type Safety

6. **Add `TokenCount` typed integer**: `type TokenCount int` — prevents mixing token counts with byte counts or line counts at the type level. The mutation bug we fixed earlier would have been impossible.

7. **Add `domain.Issue` interface**: Both `ProcessedClone` and `Finding` could implement a common interface for "things found in code". Would allow unified reporting.

---

## f) Top #25 Things We Should Get Done Next! 🎯

### 🔴 Critical (Do First)

| #   | Task                                                                   | Impact                           | Effort |
| --- | ---------------------------------------------------------------------- | -------------------------------- | ------ |
| 1   | **Wire Findings to CLI output** — call `FindFindings`, print results   | Correctness — features invisible | M      |
| 2   | **Add Finding printer** — text + JSON format for findings              | UX — TODO/legacy output          | M      |
| 3   | **Type-strengthen `ProcessedClone`**: `Filepath` + `LineNumber` fields | Compile-time safety              | M      |
| 4   | **Add `TokenCount` typed integer** to prevent unit mixing              | Prevents mutation bug class      | S      |
| 5   | **Break syntax import cycle** — extract `syntax/types` sub-package     | Unblocks facade + decoupling     | M      |

### 🟠 High Impact

| #   | Task                                                                    | Impact                         | Effort |
| --- | ----------------------------------------------------------------------- | ------------------------------ | ------ |
| 6   | **Move actionability behind interface** (no `syntax.Node` in printer)   | Decoupling                     | M      |
| 7   | **Split `printer/` into sub-packages**                                  | Module boundaries, testability | L      |
| 8   | **Consolidate Clone DTOs** — single canonical type or generated mappers | Eliminates manual conversion   | L      |
| 9   | **Wire `InternFilename` into transform functions**                      | Memory reduction               | S      |
| 10  | **Unify enum patterns** — config enums should also use `pkg/enum`       | Consistency                    | S      |
| 11  | **Add `domain.Issue` interface** for unified reporting                  | Extensibility                  | S      |
| 12  | **Add SARIF output for Findings**                                       | Security tool integration      | M      |

### 🟡 Medium Impact

| #   | Task                                                               | Impact           | Effort |
| --- | ------------------------------------------------------------------ | ---------------- | ------ |
| 13  | **Add fuzz tests for templ parser**                                | Robustness       | M      |
| 14  | **Refactor `syntax/golang/transform.go`** (369L switch)            | Maintainability  | M      |
| 15  | **Refactor `printer/actionability.go`** (650L — reduce further)    | Maintainability  | M      |
| 16  | **Break SDK error sentinels** from internal `errors` package       | SDK independence | S      |
| 17  | **Add BDD tests for Finding pipeline** — TODO detection end-to-end | Test coverage    | M      |
| 18  | **Add BDD tests for `--suppress-test-low`**                        | Test coverage    | S      |
| 19  | **Add BDD tests for `--test-threshold`**                           | Test coverage    | S      |
| 20  | **Clean up `go.sum`** — stale entries from auto-tidy               | Hygiene          | S      |

### 🟢 Polish

| #   | Task                                                              | Impact                         | Effort |
| --- | ----------------------------------------------------------------- | ------------------------------ | ------ |
| 21  | **Fix gocritic unlambda hint** in actionability.go                | Code quality                   | S      |
| 22  | **Add property-based tests for suffix tree**                      | Testing robustness             | M      |
| 23  | **Implement hybrid slice/map transition storage**                 | Performance micro-optimization | M      |
| 24  | **Add `ProcessedCloneGroup.Validate()` calls in production code** | Runtime invariant checking     | S      |
| 25  | **Document Finding pipeline in SDK_DESIGN.md**                    | Knowledge preservation         | S      |

---

## g) Top #1 Question I Cannot Figure Out Myself 🤔

> **How should Findings be displayed in the CLI output?**
>
> The entire `domain.Finding` pipeline is built: `MultiDetector.FindFindings(ctx)` returns `<-chan domain.Finding`, TodoDetector and LegacyDetector produce findings correctly (verified by `TestMultiDetector_FindFindings_TodosMethod`). But the CLI never calls `FindFindings` — it only calls `FindDuplOver` for clones.
>
> **The question is about output design, not implementation:**
>
> 1. **Same output stream as clones?** Findings intermixed with clone groups, sorted by line number? This is simplest but may confuse — a TODO comment at line 42 and a clone spanning lines 50-70 would appear adjacent despite being fundamentally different result types.
> 2. **Separate output section?** After clone output, a "--- Findings ---" header listing all TODOs/legacy issues? This is cleaner separation but doubles the output complexity.
> 3. **Separate output format flag?** `--findings-format text|json|sarif` independent from clone output format? Maximum flexibility but maximum complexity.
>
> **I cannot determine**: Should findings appear automatically whenever `--detection-methods todos,legacy` is used, or should there be a `--show-findings` flag? And should the existing output formats (text, JSON, SARIF, HTML) each get a findings section, or only some?

---

## Verification

```
Build:  ✅ go build ./... — clean
Lint:   ✅ golangci-lint run — 0 issues (1 gocritic info hint)
Tests:  ✅ 22/22 packages passing
BDD:    ✅ All passing
```

## Files Changed This Session

**New files:**

- `pkg/enum/enum.go` — Shared generic enum helpers
- `domain/finding.go` — Finding type for code-quality issues
- `detection/adapters.go` — SuffixTreeAdapter, HashAdapter for MethodDetector
- `detection/finding_test.go` — Tests for Finding pipeline
- `syntax/intern.go` — String interning for filenames
- `suffixtree/fuzz_test.go` — Fuzz tests for suffix tree

**Significantly modified:**

- `detection/multidetector.go` — Polymorphic dispatch, FindFindings, adapter integration
- `detection/todo_detector.go`, `detection/legacy_detector.go` — FindFindings methods
- `detection/issue_helpers.go` — findFindingsInFile generic helper
- `domain/processed_clone.go` — ProcessedCloneGroup.Validate(), enum unification
- `domain/types_health.go` — enum unification
- `printer/actionability.go` — everySequenceMatch refactor, new detection patterns
- `printer/clone_classify.go` — New pattern labels and suggestions
- `suffixtree/dupl.go` — context.Context propagation
- `pkg/artdupl/detector.go` — FindClonesStream deprecation
- `pkg/artdupl/types.go` — ValidateOptions independence
- `config/config.go` — SuppressTestLow, TestThreshold fields
- `cmd/flags.go`, `cmd/config_builder.go`, `cmd/run_output.go` — New CLI flags, suppression logic
- `AGENTS.md` — Updated architecture docs
- `TODO_LIST.md` — Comprehensive completion status
