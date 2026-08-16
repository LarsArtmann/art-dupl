# Status Report: Context Safety, SDK Independence & Detection Correctness

> **Date:** 2026-06-16 10:51
> **Branch:** fork
> **Commit:** (ahead of origin/fork by 1 commit)

---

## Executive Summary

Executed the **next tier of the TODO sprint** from the planning document (`docs/planning/2026-06-15_23-05_SUPERB-FULL-TODO-EXECUTION.md`). Completed **5 of 8 planned tasks** across detection correctness, SDK type independence, context propagation, and data model hardening. The SDK's `pkg/artdupl` package is now fully decoupled from internal `config` and `pkg/logger` types — no more type aliases leaking implementation details. All detection goroutines now respect `context.Context` cancellation, preventing goroutine leaks. All 25 packages pass, 0 lint issues.

---

## a) FULLY DONE ✅

### This Session

| #  | Change                                                                                                          | Files                                                                                                                                             | Impact                                                                                                                                                           |
| -- | --------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1  | **Fixed empty Frags guard bug**: `[][]*Node{{}}` was passing `len(Frags) > 0` check                             | `detection/multidetector.go`, `detection/issue_helpers.go`                                                                                        | Issue matches (TODO, legacy) with empty inner slices no longer leak into the clone pipeline. New `hasNonEmptyFrag()` helper checks inner depth.                  |
| 2  | **Removed `Frags: [][]*Node{{}}` from `createIssueMatch`** — now nil                                            | `detection/issue_helpers.go`, `detection/detection_test.go`                                                                                       | Issues no longer pretend to have fragments. The guard filters them correctly. Tests updated to assert nil Frags.                                                 |
| 3  | **Set default `Actionability` in `ClassifyClone`** — `Actionable` for normal clones, `NonActionable` for idioms | `printer/clone_classify.go`                                                                                                                       | Eliminates zero-value `""` invalid state. `CloneClassification` can no longer be constructed with an invalid `Actionability` through the normal path.            |
| 4  | **Moved `processOrderMethodName` test constant** from production to test file                                   | `printer/actionability.go` → `printer/actionability_test.go`                                                                                      | Test-only constant no longer pollutes production code.                                                                                                           |
| 5  | **Added `context.Context` to all detection goroutines** (M6 complete)                                           | `detection/multidetector.go`, `detection/issue_helpers.go`, `detection/todo_detector.go`, `detection/legacy_detector.go`, `detection/detector.go` | Every `FindDuplOver`, `FindTodos`, `FindLegacy`, `findIssuesInFile`, `findIssuesGeneric` now accepts and checks `ctx`. Prevents goroutine leaks on cancellation. |
| 6  | **Updated all callers** of context-aware detection methods                                                      | `cmd/run_analysis.go`, `pkg/artdupl/detector_pipeline.go`, `pkg/artdupl/detector_uncovered_test.go`                                               | All consumers pass context through. SDK pipeline uses `select` on `ctx.Done()`.                                                                                  |
| 7  | **Broke SDK `DetectionMethod` type alias** — now independent `type DetectionMethod string`                      | `pkg/artdupl/types.go`, `pkg/artdupl/detector_utils.go`                                                                                           | SDK no longer re-exports `config.DetectionMethod`. Has own `String()`, `IsValid()`, and boundary conversion functions.                                           |
| 8  | **Broke SDK `Logger` type alias** — now independent interface with `noOpLogger` default                         | `pkg/artdupl/types.go`                                                                                                                            | SDK defines its own 4-method Logger interface. Structurally compatible with `pkg/logger.Logger` — no adapter needed.                                             |
| 9  | **Refactored `FindDuplOver` complexity** — extracted `runMultiMethodDetection` + `streamMatches` helpers        | `detection/multidetector.go`                                                                                                                      | Reduced cognitive complexity from 57 to under 35 (gocognit lint pass). Each detection method's streaming loop is now DRY via `streamMatches`.                    |
| 10 | **Fixed all test assertions** for type-alias break                                                              | `pkg/artdupl/detector_validation_test.go`                                                                                                         | Tests now compare `string()` values instead of cross-type equality.                                                                                              |

### Previously Completed (Prior Commits, Still on This Branch)

| Change                                                                                 | Commit    |
| -------------------------------------------------------------------------------------- | --------- |
| Data model improvements (Size→TokenCount, mutation fix, dead code removal, Validate()) | `83ccbdd` |
| Architecture docs, planning docs, quality scans                                        | `7ed8183` |
| Comprehensive TODO execution plan (Pareto breakdown)                                   | `4bc7bb3` |
| Tier 1 quick fixes (SDK safety, detection cleanup, ClonePriority.Rank)                 | `29ed8f3` |
| CloneSeverity→ClonePriority rename + JSON serialization                                | `b87d0cb` |

### Project-Wide Health

| Metric            | Value                             |
| ----------------- | --------------------------------- |
| **Build**         | ✅ Passing (`go build ./...`)     |
| **Lint**          | ✅ 0 issues (`golangci-lint run`) |
| **Unit Tests**    | ✅ 25/25 packages passing         |
| **BDD Tests**     | ✅ All passing (cached)           |
| **Go files**      | 229 (137 production, 92 test)     |
| **Lines of code** | ~18,915 production, ~29,817 test  |

---

## b) PARTIALLY DONE 🟡

| Area                                 | Status | What's Done                                                                               | What Remains                                                                                                                                              |
| ------------------------------------ | ------ | ----------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Clone type consolidation**         | 25%    | `printer.clone` dead type removed (4→3 types)                                             | `printer.CloneGroup`, `pkg/artdupl.Clone`, `domain.ProcessedClone` still exist as separate DTOs                                                           |
| **Printer ↔ syntax.Node decoupling** | 60%    | `clone_processor.go` bridges Node→ProcessedClone; `common.go` no longer has `clone` type  | `actionability.go` still imports `syntax.Node` directly for pattern evaluation                                                                            |
| **Domain validation**                | 40%    | `ProcessedClone.Validate()` added; enums have JSON validation; Actionability defaults set | `CloneCategory`, `CloneActionability` lack `MarshalJSON`/`UnmarshalJSON` (decided YAGNI for now — they're always `string()`-converted at output boundary) |
| **Typed fields in ProcessedClone**   | 0%     | Not started — `Filename` is still `string`, `LineStart`/`LineEnd` still `int`             | Could use `domain.Filepath` and `domain.LineNumber` — deferred due to ~25 consumer sites requiring `.String()`/`.Int()` conversions                       |
| **SDK type independence**            | 85%    | `DetectionMethod` and `Logger` aliases broken; conversion functions at boundary           | `pkg/artdupl/errors.go` still uses internal `errors` package directly (low priority — errors are standard `error` interface)                              |
| **Context propagation**              | 90%    | All detection goroutines, SDK pipeline, and CLI use `ctx`                                 | `suffixtree.STree.FindDuplOver` does not accept `ctx` (deep in algorithm — would require significant rework)                                              |

---

## c) NOT STARTED ⬜

### From Planning Doc (Tier 2-3, Not Yet Begun)

- [ ] **M9: Printer code quality** — Extract `everySequenceMatch` helper, `validateLocation` helper, ADR for actionability
- [ ] **M10: Activate `MethodDetector` interface** — Create adapters for all 4 detection methods, replace if-chain with loop dispatch
- [ ] **M11: UX Features** — `--suppress-test-low` flag, separate test/production threshold
- [ ] **M12: Syntax facade** — Hide `syntax/golang` and `syntax/templ` behind `syntax` package
- [ ] **M13: Enum unification** — Move domain enum helpers to use config's generic patterns

### From TODO_LIST.md (Deferred)

- [ ] Split `printer/` into sub-packages (50 files)
- [ ] Consolidate three parallel Clone types
- [ ] Implement string interning (performance)
- [ ] Refactor `transform.go` (369L switch)
- [ ] Add fuzz tests for templ parser + suffix tree
- [ ] Add property-based tests for suffix tree

---

## d) TOTALLY FUCKED UP! 🔥

### Critical Bugs Found (Not Yet Fixed)

| # | Bug                                          | Severity | Location                     | Impact                                                                                                                                                                                                                                                                                                                                                   |
| - | -------------------------------------------- | -------- | ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1 | **`suffixtree.FindDuplOver` ignores ctx**    | Low      | `suffixtree/suffixtree.go`   | The core suffix tree algorithm doesn't accept context. If a user cancels mid-scan, the goroutine continues until the algorithm completes. Workaround: the caller's `select` on `ctx.Done()` prevents the result from being consumed, but the goroutine still runs. Fixing requires threading ctx through the suffix tree traversal — significant rework. |
| 2 | **TODO/Legacy detection produces no output** | Medium   | `detection/multidetector.go` | Issue-type detectors (TODO, legacy) produce matches with nil Frags. The `hasNonEmptyFrag` guard correctly filters them from the clone pipeline, but this means **TODO and legacy detection currently produce zero results in practice**. The MethodDetector pipeline (M10) is needed to route issue matches to a separate output path.                   |

### Architecture Smells (Unchanged)

| # | Smell                                                 | Impact                                                                                                                                                             |
| - | ----------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1 | **`printer/` is a 50-file mega-package**              | No module boundaries; everything is friends with everything; untestable in isolation                                                                               |
| 2 | **Three parallel Clone types**                        | `domain.ProcessedClone`, `printer.CloneGroup` (JSON DTO), `pkg/artdupl.Clone` (SDK DTO) — same data, three shapes, manual conversion at every boundary             |
| 3 | **`actionability.go` imports `syntax.Node` directly** | Breaks the decoupling contract — `domain.ProcessedCloneGroup` was supposed to free printers from AST knowledge, but actionability evaluation still walks raw nodes |

---

## e) WHAT WE SHOULD IMPROVE! 💡

### Architecture Specific

1. **Activate `MethodDetector` interface (M10)**: The interface exists (`detection/detector.go`) but is unused. `MultiDetector.FindDuplOver` still uses an if-chain. Creating adapters for each detector and dispatching via a loop would eliminate the if-chain and make the system extensible.

2. **Route issue-type detections separately**: TODO and legacy detectors produce code-quality findings (single-line issues), not code clones. They need their own output channel type, not the `syntax.Match` clone channel. The MethodDetector pipeline should distinguish between "clone matches" and "issue matches."

3. **Split `printer/` package**: 50 files is too many. Natural seams: `printer/stats/`, `printer/html/`, `printer/json/`, `printer/text/`, `printer/sarif/`, `printer/diff/`.

4. **Type-strengthen `ProcessedClone` fields**: `Filename string` → `domain.Filepath`, `LineStart/LineEnd int` → `domain.LineNumber`. The domain types exist for exactly this — they validate at construction.

5. **Add `TokenCount` typed integer**: `type TokenCount int` — prevents mixing token counts with byte counts at the type level. The mutation bug we fixed earlier would have been impossible with this.

### Code Quality

6. **Extract `everySequenceMatch` helper in actionability.go**: Multiple `is*` functions repeat the "check every sequence" pattern. DRY opportunity.

7. **Create ADR for actionability system**: The pattern detection logic in `actionability.go` (558 lines) is complex and undocumented. An ADR would preserve institutional knowledge.

8. **Unify enum patterns**: Domain enums (`ClonePriority`, `CloneCategory`, etc.) each have their own `IsValid()`/`String()` implementations. Config package has generic helpers. Consolidate.

---

## f) Top #25 Things We Should Get Done Next! 🎯

### 🔴 Critical (Do First)

| # | Task                                                                                | Impact                              | Effort |
| - | ----------------------------------------------------------------------------------- | ----------------------------------- | ------ |
| 1 | **Route issue detections separately** — TODO/legacy need their own output path      | Correctness — currently zero output | M      |
| 2 | **Activate `MethodDetector` interface** (M10) — replace if-chain with loop dispatch | Extensibility, enables #1           | M      |
| 3 | **Type-strengthen `ProcessedClone`**: `Filepath` + `LineNumber` fields              | Compile-time safety                 | M      |
| 4 | **Add `TokenCount` typed integer** to prevent unit mixing                           | Prevents mutation bug class         | S      |
| 5 | **Add `ProcessedCloneGroup.Validate()`** — validate group invariants                | Data integrity                      | S      |

### 🟠 High Impact

| #  | Task                                                                    | Impact                         | Effort |
| -- | ----------------------------------------------------------------------- | ------------------------------ | ------ |
| 6  | **Extract `everySequenceMatch` helper** in actionability.go (M9)        | Code quality — DRY             | S      |
| 7  | **Move actionability behind interface** (no `syntax.Node` in printer)   | Decoupling                     | M      |
| 8  | **Split `printer/` into sub-packages**                                  | Module boundaries, testability | L      |
| 9  | **Consolidate Clone DTOs** — single canonical type or generated mappers | Eliminates manual conversion   | L      |
| 10 | **Hide `syntax/golang` + `syntax/templ` behind `syntax` facade** (M12)  | Clean module boundary          | M      |
| 11 | **Add `--suppress-test-low` flag** (M11)                                | UX for noisy test clones       | S      |
| 12 | **Separate test/production threshold** (M11)                            | UX precision                   | S      |
| 13 | **Unify enum patterns** across domain and config (M13)                  | Consistency                    | M      |

### 🟡 Medium Impact

| #  | Task                                                                | Impact                    | Effort |
| -- | ------------------------------------------------------------------- | ------------------------- | ------ |
| 14 | **Add `context.Context` to `suffixtree.FindDuplOver`**              | Goroutine leak prevention | M      |
| 15 | **Create ADR for actionability pattern detection** (M9)             | Knowledge preservation    | S      |
| 16 | **Extract `validateLocation` helper** in todo/legacy detectors (M9) | Code quality — DRY        | S      |
| 17 | **Break SDK error sentinels** from internal `errors` package        | SDK independence          | S      |
| 18 | **Add fuzz tests** for templ parser + suffix tree invariants        | Robustness                | M      |
| 19 | **Refactor `syntax/golang/transform.go`** (369L, 300L switch)       | Maintainability           | M      |
| 20 | **Refactor `printer/actionability.go`** (558L)                      | Maintainability           | M      |

### 🟢 Polish

| #  | Task                                                                          | Impact                         | Effort |
| -- | ----------------------------------------------------------------------------- | ------------------------------ | ------ |
| 21 | **Add `DescribeTable` detection pattern**                                     | Feature enhancement            | M      |
| 22 | **Add builder/callback detection pattern**                                    | Feature enhancement            | M      |
| 23 | **Implement hybrid slice/map transition storage** for small transition counts | Performance micro-optimization | M      |
| 24 | **Implement string interning** for node filenames                             | Performance                    | M      |
| 25 | **Add property-based tests for suffix tree**                                  | Testing robustness             | M      |

---

## g) Top #1 Question I Cannot Figure Out Myself 🤔

> **Should TODO/Legacy issue detections flow through the same `syntax.Match` channel as clone detections, or should they have a completely separate pipeline?**
>
> Currently, TODO and legacy detectors produce `syntax.Match` values with `nil` Frags — they represent single-line code-quality findings, not code clones with AST fragments. The `hasNonEmptyFrag` guard correctly filters them from the clone pipeline, but this means **TODO and legacy detection methods currently produce zero visible output** when used.
>
> **Two approaches:**
>
> 1. **Separate output channel**: Add a `FindIssues(ctx) <-chan Issue` method to `MethodDetector`. CLI handles issues and clones differently. Clean separation, but doubles the pipeline surface.
> 2. **Synthetic fragments**: Give issue matches a single-node fragment pointing to the line. They flow through the existing pipeline and get printed as 1-line "clones." Hacky but requires zero pipeline changes.
>
> **I cannot determine**: Which approach aligns with the product vision. Are TODO/legacy detections meant to be first-class output (separate report section), or should they be intermixed with clone results? This requires UX/product knowledge I don't have.

---

## Verification

```
Build:  ✅ go build ./... — clean
Lint:   ✅ golangci-lint run — 0 issues
Tests:  ✅ 25/25 packages passing
BDD:    ✅ All passing (cached)
```

## Files Changed (15 files, +240/-106 lines)

**Detection layer (context + correctness):**

- `detection/multidetector.go` — Added ctx param, extracted `runMultiMethodDetection` + `streamMatches`, `hasNonEmptyFrag` guard
- `detection/detector.go` — Updated `MethodDetector` interface with ctx
- `detection/issue_helpers.go` — Added ctx to `findIssuesInFile`/`findIssuesGeneric`, fixed `createIssueMatch` Frags
- `detection/todo_detector.go` — Added ctx to `FindTodos`
- `detection/legacy_detector.go` — Added ctx to `FindLegacy`
- `detection/detection_test.go` — Updated all test callers with ctx, fixed assertions

**SDK layer (type independence):**

- `pkg/artdupl/types.go` — Broke `DetectionMethod` and `Logger` aliases, added conversion functions, `noOpLogger`
- `pkg/artdupl/detector_utils.go` — Uses `toConfigDetectionMethods` for boundary conversion
- `pkg/artdupl/detector_pipeline.go` — Passes ctx to `FindDuplOver`, fixed frag filtering
- `pkg/artdupl/detector_uncovered_test.go` — Updated test caller
- `pkg/artdupl/detector_validation_test.go` — Fixed cross-type equality assertions

**Printer layer (data integrity):**

- `printer/clone_classify.go` — Set default `Actionability` in both classify paths
- `printer/actionability.go` — Moved `processOrderMethodName` to test file
- `printer/actionability_test.go` — Received `processOrderMethodName` const

**CLI layer:**

- `cmd/run_analysis.go` — Passes ctx to `FindDuplOver`
