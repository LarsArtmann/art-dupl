# Status Report: Data Model Improvements

> **Date:** 2026-06-15 22:25
> **Branch:** fork
> **Commit:** 3330165 (ahead of origin/fork by 1 commit)

---

## Executive Summary

Completed a focused **data model improvement sprint** targeting the `domain.ProcessedClone` / `ProcessedCloneGroup` DTOs and their consumers. Fixed a **destructive mutation bug** hidden behind a lying field name, removed dead code, consolidated duplicate error sentinels, and added behavioral methods to make invalid states detectable. All 22 packages pass, 264 BDD specs pass, 0 lint issues.

---

## a) FULLY DONE ✅

### Data Model Improvements (This Session)

| #   | Change                                                                                                                           | Files                                                                  | Impact                                                                                                                |
| --- | -------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| 1   | **Removed dead `printer.clone` struct** + `sumFragmentLengths()` + `sortCloneGroupsBySize()` + their tests                       | `printer/common.go`, `printer/sorter.go`, `printer/text_utils_test.go` | Eliminated dead code path that was 100% superseded by `domain.ProcessedClone` versions                                |
| 2   | **Renamed `Size` → `TokenCount`** on `ProcessedClone` and `ProcessedCloneGroup`                                                  | `domain/processed_clone.go` + 10 consumer files                        | Fixed lying name: `Size` meant token count in most places but byte count in `text.go`                                 |
| 3   | **Fixed mutation bug**: `calculateProcessedCloneSizes()` was destructively overwriting `TokenCount` with `len(Fragment)` (bytes) | `printer/text.go`                                                      | Eliminated silent data corruption — downstream code saw different values depending on whether `TextPrinter` ran first |
| 4   | **Consolidated duplicate error sentinel**: removed `ErrInvalidSeverity`, unified on `ErrInvalidCloneSeverity`                    | `domain/analysis_errors.go`, `domain/types_severity.go`                | Two names for the same concept, used inconsistently in Marshal vs Unmarshal of the same type                          |
| 5   | **Added `ProcessedClone.Validate()`** method                                                                                     | `domain/processed_clone.go`                                            | Makes invalid states detectable (empty filename, LineEnd < LineStart, negative TokenCount)                            |
| 6   | **Added `ProcessedClone.LineCount()`** method                                                                                    | `domain/processed_clone.go`, `printer/stats.go`                        | Replaces repeated `LineEnd - LineStart + 1` inline computation                                                        |
| 7   | **Added `ProcessedCloneGroup.TotalTokenCount()`** method                                                                         | `domain/processed_clone.go`, `cmd/run_output.go`                       | Canonical token summation; eliminated local `totalTokenCount()` function                                              |
| 8   | **Added 3 new error sentinels**: `ErrEmptyFilename`, `ErrLineEndBeforeStart`, `ErrNegativeTokenCount`                            | `domain/analysis_errors.go`                                            | Strong typed errors for `Validate()`                                                                                  |
| 9   | **Comprehensive tests** for all new methods                                                                                      | `domain/domain_test.go`                                                | `TestProcessedClone_LineCount`, `TestProcessedClone_Validate`, `TestProcessedCloneGroup_TotalTokenCount`              |
| 10  | **Updated AGENTS.md Known Limitations** — "Four parallel Clone types" → "Three parallel Clone types"                             | `AGENTS.md`                                                            | Reflects `printer.clone` removal                                                                                      |

### Previously Completed (Prior Sessions, Still Uncommitted)

| Change                                                                                    | Files                                                                 |
| ----------------------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| `cnt` → `count` rename in `syntax/syntax.go` (4 functions)                                | `syntax/syntax.go`                                                    |
| `isReturnOrWrappedReturn` dead branch removal + typo fix `(HaveOccurred` → `HaveOccurred` | `printer/actionability.go`                                            |
| `CloneClassification.NodeType` → `NodeTypeName` rename                                    | `printer/clone_classify.go`, `printer/clone_classify_test.go`         |
| Exhaustive switch fix: missing `PriorityLow` case                                         | `printer/stats.go`                                                    |
| `.go-arch-lint.yml` fixes (domain→detection deps, internal/utils mapping)                 | `.go-arch-lint.yml`                                                   |
| FEATURES.md update (stale claims fixed, 11 missing features added)                        | `FEATURES.md`                                                         |
| TODO_LIST.md comprehensive update                                                         | `TODO_LIST.md`                                                        |
| Modularization proposal update                                                            | `docs/modularization/PROPOSAL.md`                                     |
| go.sum update (transitive hash fixes)                                                     | `go.sum`                                                              |
| Architecture review documents generated                                                   | `docs/architecture-understanding/`, `docs/planning/`, `docs/quality/` |

### Project-Wide Health

| Metric            | Value                              |
| ----------------- | ---------------------------------- |
| **Build**         | ✅ Passing (`go build ./...`)      |
| **Lint**          | ✅ 0 issues (`golangci-lint run`)  |
| **Unit Tests**    | ✅ 22/22 packages passing          |
| **BDD Tests**     | ✅ 264 passed, 0 failed, 4 pending |
| **Go files**      | 229 (137 production, 92 test)      |
| **Lines of code** | ~18,668 production, ~29,805 test   |

---

## b) PARTIALLY DONE 🟡

| Area                                 | Status | What's Done                                                                              | What Remains                                                                                                                                                               |
| ------------------------------------ | ------ | ---------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Clone type consolidation**         | 25%    | `printer.clone` dead type removed (4→3 types)                                            | `printer.CloneGroup`, `pkg/artdupl.Clone`, `domain.ProcessedClone` still exist as separate DTOs                                                                            |
| **Printer ↔ syntax.Node decoupling** | 60%    | `clone_processor.go` bridges Node→ProcessedClone; `common.go` no longer has `clone` type | `actionability.go` still imports `syntax.Node` directly for pattern evaluation                                                                                             |
| **Domain validation**                | 40%    | `ProcessedClone.Validate()` added; `CloneSeverity`/`HealthScore` have JSON validation    | `CloneCategory`, `ClonePriority`, `CloneActionability` lack `MarshalJSON`/`UnmarshalJSON` (decided YAGNI for now — they're always `string()`-converted at output boundary) |
| **Typed fields in ProcessedClone**   | 0%     | Not started — `Filename` is still `string`, `LineStart`/`LineEnd` still `int`            | Could use `domain.Filepath` and `domain.LineNumber` — deferred due to ~25 consumer sites requiring `.String()`/`.Int()` conversions                                        |
| **Error sentinel cleanup**           | 50%    | `ErrInvalidSeverity` consolidated into `ErrInvalidCloneSeverity`                         | `CloneSeverity` vs `ClonePriority` still represent overlapping concepts (same 4 levels, two types)                                                                         |

---

## c) NOT STARTED ⬜

### From TODO_LIST.md HIGH Priority

- [ ] Activate `MethodDetector` interface for polymorphic dispatch
- [ ] Split `printer/` into sub-packages (50 files is too many)
- [ ] Fix SDK `FindClonesStream` error handling (silently swallowed)
- [ ] Deep-copy `Options` in `NewDetector` (shared pointer mutation risk)
- [ ] Fix `legacy_detector.go:36` fragile string matching
- [ ] Wire `ErrNoDuplicatesFound` sentinel
- [ ] Fix `Clone.IsValid()` skipping length check at `StartPos == 0`
- [ ] Collapse `CloneSeverity`/`ClonePriority` into one type
- [ ] Wire `Actionability` field in `CloneClassification` (always zero value)
- [ ] Break SDK type aliases (`DetectionMethod = config.DetectionMethod`)

### From TODO_LIST.md MEDIUM Priority

- [ ] Add `context.Context` to MultiDetector goroutines
- [ ] Document detector thread-safety contract
- [ ] Remove dead `Patterns`/`Imports` fields in `LegacyPattern`
- [ ] Fix HealthScore legend vs formula mismatch
- [ ] Add `ClonePriority.Rank()` to domain
- [ ] Hide `syntax/golang` and `syntax/templ` behind `syntax` facade
- [ ] Unify enum patterns across domain and config packages

---

## d) TOTALLY FUCKED UP! 🔥

### Critical Bugs Found (Not Yet Fixed)

| # | Bug | Severity | Location | Impact |
| --- | -------------------------------------- | ----------------------- | ---------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --- | ------------------------------------------------------------------------------------------ |
| 1 | **Mutation bug WAS here** — now FIXED | ~~Critical~~ → ✅ Fixed | ~~`printer/text.go:271`~~ | `calculateProcessedCloneSizes()` was silently overwriting `TokenCount` (AST node count) with `len(Fragment)` (byte count). Every printer that ran after `TextPrinter` saw corrupted data. **This is now fixed by replacing with non-mutating `totalFragmentSize()`.** |
| 2 | **SDK `Options` shared pointer** | High | `pkg/artdupl/detector.go` | `NewDetector` stores `*Options` directly — callers can mutate config after construction, causing races. Should deep-copy. |
| 3 | **`FindClonesStream` swallows errors** | High | `pkg/artdupl/detector_pipeline.go` | Pipeline errors are logged but never returned to caller. Stream appears to succeed even when parsing fails. |
| 4 | **`Clone.IsValid()` skips zero-check** | Medium | `pkg/artdupl/types.go:77` | `if c.StartPos > 0                                                                                                                                                                                                                                                    |     | c.EndPos > 0` — when both are 0, length check is skipped entirely. Byte offset 0 is valid. |
| 5 | **`Actionability` field always zero** | Medium | `domain/processed_clone.go` | `CloneClassification.Actionability` is set correctly in `clone_processor.go` but `CloneClassification` struct's zero value is `""` which is NOT a valid `CloneActionability`. Invalid state is representable. |

### Architecture Smells

| #   | Smell                                                 | Impact                                                                                                                                                             |
| --- | ----------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| 1   | **`printer/` is a 50-file mega-package**              | No module boundaries; everything is friends with everything; untestable in isolation                                                                               |
| 2   | **Three parallel Clone types**                        | `domain.ProcessedClone`, `printer.CloneGroup` (JSON DTO), `pkg/artdupl.Clone` (SDK DTO) — same data, three shapes, manual conversion at every boundary             |
| 3   | **`actionability.go` imports `syntax.Node` directly** | Breaks the decoupling contract — `domain.ProcessedCloneGroup` was supposed to free printers from AST knowledge, but actionability evaluation still walks raw nodes |

---

## e) WHAT WE SHOULD IMPROVE! 💡

### Data Model Specific

1. **Type-strengthen `ProcessedClone` fields**: Change `Filename string` → `Filename domain.Filepath`, `LineStart/LineEnd int` → `domain.LineNumber`. The domain types exist precisely for this — they validate at construction and marshal correctly. Deferred due to ~25 consumer sites, but high value for type safety.

2. **Make `CloneActionability` zero-value invalid**: Currently `""` is the zero value, which passes no validation. Use a constructor or make the zero value explicitly invalid with a `IsValid()` that rejects empty strings (already done — but `CloneClassification` can still be constructed with zero-value `Actionability`).

3. **Add `ProcessedCloneGroup.Validate()`**: Should validate that all clones share the same hash, that TokenCount equals the sum of clone TokenCounts, and that no clone is invalid.

4. **Consider `TokenCount` as a typed integer**: `type TokenCount int` — prevents mixing token counts with byte counts or line counts at the type level. The mutation bug we just fixed would have been impossible.

### Architecture Specific

5. **Split `printer/` package**: 50 files is too many. Natural seams: `printer/stats/`, `printer/html/`, `printer/json/`, `printer/text/`, `printer/sarif/`, `printer/diff/`.

6. **Move actionability evaluation behind an interface**: Currently `EvaluateActionability(nodeSeqs [][]*syntax.Node)` takes raw AST nodes. Should take `[]ProcessedClone` or a domain-level representation, breaking the `syntax.Node` dependency in `printer/`.

7. **Consolidate Clone DTOs**: Either make `domain.ProcessedClone` the single canonical type that all printers use directly, or define a clear conversion protocol with generated mappers.

---

## f) Top #25 Things We Should Get Done Next! 🎯

### 🔴 Critical (Do First)

| #   | Task                                                                           | Impact                   | Effort |
| --- | ------------------------------------------------------------------------------ | ------------------------ | ------ |
| 1   | **Deep-copy `Options` in `NewDetector`** — prevents post-construction mutation | Prevents race conditions | S      |
| 2   | **Fix `FindClonesStream` error handling** — return errors, don't just log      | SDK correctness          | M      |
| 3   | **Fix `Clone.IsValid()` zero-position skip**                                   | SDK correctness bug      | S      |
| 4   | **Wire `ErrNoDuplicatesFound` sentinel**                                       | SDK contract honesty     | S      |
| 5   | **Wire `Actionability` field properly** in `CloneClassification`               | Data integrity           | S      |

### 🟠 High Impact

| #   | Task                                                                   | Impact                                 | Effort |
| --- | ---------------------------------------------------------------------- | -------------------------------------- | ------ |
| 6   | **Collapse `CloneSeverity`/`ClonePriority`** into one type             | Eliminates split brain                 | M      |
| 7   | **Type-strengthen `ProcessedClone`**: `Filepath` + `LineNumber` fields | Compile-time safety                    | M      |
| 8   | **Add `TokenCount` typed integer** to prevent unit mixing              | Prevents mutation bug class            | S      |
| 9   | **Split `printer/` into sub-packages**                                 | Module boundaries, testability         | L      |
| 10  | **Move actionability behind interface** (no `syntax.Node` in printer)  | Decoupling                             | M      |
| 11  | **Activate `MethodDetector` interface** for polymorphic dispatch       | Extensibility                          | M      |
| 12  | **Add `context.Context` to MultiDetector goroutines**                  | Prevents goroutine leaks               | S      |
| 13  | **Add `ClonePriority.Rank()`** to domain                               | Deduplicates ordinal logic in 2 places | S      |

### 🟡 Medium Impact

| #   | Task                                                                               | Impact                   | Effort |
| --- | ---------------------------------------------------------------------------------- | ------------------------ | ------ |
| 14  | **Fix HealthScore legend vs formula mismatch**                                     | User trust               | S      |
| 15  | **Hide `syntax/golang` + `syntax/templ` behind `syntax` facade**                   | Clean module boundary    | M      |
| 16  | **Break SDK type aliases** (`DetectionMethod`, `Logger`)                           | SDK independence         | M      |
| 17  | **Remove dead `Patterns`/`Imports` fields in `LegacyPattern`**                     | Dead code                | S      |
| 18  | **Fix `issue_helpers.go:82`** — `Frags: [][]*syntax.Node{{}}` passes length filter | Correctness              | S      |
| 19  | **Add `--suppress-test-low` flag**                                                 | UX for noisy test clones | S      |
| 20  | **Unify enum patterns** (domain should use config's generic helpers)               | Consistency              | M      |

### 🟢 Polish

| #   | Task                                                                          | Impact                         | Effort |
| --- | ----------------------------------------------------------------------------- | ------------------------------ | ------ |
| 21  | **Refactor `syntax/golang/transform.go`** (369L, 300L switch)                 | Maintainability                | M      |
| 22  | **Refactor `printer/actionability.go`** (558L)                                | Maintainability                | M      |
| 23  | **Add fuzz tests** for templ parser + suffix tree invariants                  | Robustness                     | M      |
| 24  | **Create ADR for actionability pattern detection**                            | Knowledge preservation         | S      |
| 25  | **Implement hybrid slice/map transition storage** for small transition counts | Performance micro-optimization | M      |

---

## g) Top #1 Question I Cannot Figure Out Myself 🤔

> **Should `CloneSeverity` and `ClonePriority` be collapsed into a single type?**
>
> They have identical constant values (`low`, `medium`, `high`, `critical`) and identical semantics (importance level). The DOMAIN_LANGUAGE.md defines them as separate concepts:
>
> - **Priority**: "How important it is to address"
> - **Severity**: "Impact level of a clone"
>
> But in practice, they are always the same 4 values with the same meaning. `ClonePriority` is used in the classification pipeline (clone_classify.go). `CloneSeverity` is used in the legacy/todo detectors. They never interact.
>
> **I cannot determine**: Is there a domain reason to keep them separate (e.g., future plans where severity and priority diverge), or is this an accidental duplication from separate development cycles? This requires business domain knowledge I don't have.

---

## Verification

```
Build:  ✅ go build ./... — clean
Lint:   ✅ golangci-lint run — 0 issues
Tests:  ✅ 22/22 packages passing
BDD:    ✅ 264 passed, 0 failed, 4 pending
```

## Files Changed (25 files, +367/-184 lines)

**Data model core:**

- `domain/processed_clone.go` — Renamed Size→TokenCount, added Validate(), LineCount(), TotalTokenCount()
- `domain/analysis_errors.go` — Removed ErrInvalidSeverity, added 3 new validation errors
- `domain/types_severity.go` — Fixed to use consolidated ErrInvalidCloneSeverity
- `domain/domain_test.go` — Updated tests + 3 new test functions

**Printer consumers:**

- `printer/clone_processor.go` — Updated to TokenCount
- `printer/common.go` — Removed dead `clone` struct
- `printer/sorter.go` — Removed dead `sumFragmentLengths`, `sortCloneGroupsBySize`
- `printer/text.go` — Replaced mutating `calculateProcessedCloneSizes` with non-mutating `totalFragmentSize`
- `printer/text_test.go` — Updated test name + function reference
- `printer/text_utils_test.go` — Removed dead tests
- `printer/json.go`, `sarif.go`, `stats.go`, `sort_unified.go` — Updated Size→TokenCount

**SDK consumers:**

- `cmd/run_output.go` — Updated to TokenCount, use TotalTokenCount(), removed local function

**Pre-existing changes (prior sessions):**

- `syntax/syntax.go`, `printer/actionability.go`, `printer/clone_classify.go`, `.go-arch-lint.yml`, `FEATURES.md`, `TODO_LIST.md`, `AGENTS.md`, `docs/modularization/PROPOSAL.md`, `go.sum`
