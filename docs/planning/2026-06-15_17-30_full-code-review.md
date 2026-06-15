# Full Code Review — Pareto Execution Plan

**Date:** 2026-06-15
**Scope:** 229 Go files (48k lines), 25 packages

## Pareto Breakdown

### The 1% that delivers 51% of the result

| #   | Task                                                          | Impact        | Effort |
| --- | ------------------------------------------------------------- | ------------- | ------ |
| 1   | ✅ Fix assertion matcher typo `(HaveOccurred`                 | Correctness   | Done   |
| 2   | ✅ Fix `Actionability` field never set in CloneClassification | Correctness   | Done   |
| 3   | Wire `ErrNoDuplicatesFound` sentinel in `FindClones`          | API honesty   | 15min  |
| 4   | Fix `Clone.IsValid()` StartPos==0 bypass                      | Type safety   | 5min   |
| 5   | ✅ Fix emoji collision (handler/test-fixture)                 | UX            | Done   |
| 6   | ✅ Fix `isReturnOrWrappedReturn` dead code                    | Clean code    | Done   |
| 7   | ✅ Fix data dominance comment/code mismatch                   | Documentation | Done   |
| 8   | ✅ Rename `CloneClassification.NodeType` → `NodeTypeName`     | Clarity       | Done   |
| 9   | ✅ Fix exhaustive switch `PriorityLow`                        | Lint          | Done   |

### The 4% that delivers 64% of the result

| #   | Task                                                                            | Impact        | Effort |
| --- | ------------------------------------------------------------------------------- | ------------- | ------ |
| 10  | Activate `MethodDetector` interface (real polymorphic dispatch)                 | Composability | 2h     |
| 11  | Fix SDK streaming error handling (errors swallowed in goroutine)                | Correctness   | 1h     |
| 12  | Deep-copy `Options` in `NewDetector` (prevent post-construction mutation panic) | Safety        | 30min  |
| 13  | Fix `legacy_detector.go:36` fragile string matching (stringified AST)           | Correctness   | 1h     |
| 14  | Remove dead `Patterns`/`Imports` fields in `LegacyPattern`                      | Clean code    | 15min  |
| 15  | Fix `issue_helpers.go:82` broken `Frags` length filter                          | Correctness   | 15min  |
| 16  | Fix HealthScore legend vs formula mismatch                                      | UX            | 30min  |
| 17  | Add context.Context to MultiDetector goroutines                                 | Safety        | 1h     |

### The 20% that delivers 80% of the result

| #   | Task                                                                    | Impact        | Effort |
| --- | ----------------------------------------------------------------------- | ------------- | ------ |
| 18  | Collapse `CloneSeverity`/`ClonePriority` into one type                  | Type safety   | 2h     |
| 19  | Add JSON validation to CloneCategory/Priority/Actionability             | Type safety   | 1h     |
| 20  | Add `ParseClone*` constructors for enum types                           | Type safety   | 1h     |
| 21  | Break SDK type aliases (`DetectionMethod`, `Logger`)                    | API isolation | 1h     |
| 22  | Split `printer/` into sub-packages (stats, html, analyze)               | Modularity    | 4h     |
| 23  | Hide `syntax/golang` behind `syntax` facade                             | Modularity    | 3h     |
| 24  | Fix streaming path nil CloneGroup inconsistency                         | Correctness   | 30min  |
| 25  | Add `ClonePriority.Rank()` method to domain (deduplicate ordinal logic) | DRY           | 30min  |
| 26  | Extract `everySequenceMatch` helper in actionability.go                 | DRY           | 30min  |
| 27  | Move test-only constants out of production code                         | Clean code    | 15min  |
| 28  | Document detector thread-safety contract                                | Documentation | 15min  |

## Execution Graph (D2)

```
           ┌─────────────────────────────┐
           │  Tier 1 (1% → 51%)          │
           │  Bugs & Quick Fixes          │
           └──────────┬──────────────────┘
                      │
           ┌──────────▼──────────────────┐
           │  Tier 2 (4% → 64%)          │
           │  Architecture & Safety       │
           └──────────┬──────────────────┘
                      │
           ┌──────────▼──────────────────┐
           │  Tier 3 (20% → 80%)         │
           │  Type Safety & Modularity    │
           └─────────────────────────────┘
```

## Files Changed This Session

| File                        | Change                                          | Type     |
| --------------------------- | ----------------------------------------------- | -------- |
| `printer/stats.go`          | Fix exhaustive switch (add `PriorityLow` case)  | Bug fix  |
| `printer/actionability.go`  | Fix assertion typo, dead code, comment mismatch | Bug fix  |
| `printer/clone_classify.go` | Rename NodeType → NodeTypeName                  | Refactor |
| `domain/processed_clone.go` | Fix emoji collision, rename field               | Bug fix  |
| `syntax/syntax.go`          | Rename `cnt` → `count` in 4 functions           | Naming   |
| `.go-arch-lint.yml`         | Fix 2 dependency rule violations                | Config   |
| `FEATURES.md`               | Update stale claims, add 11 missing features    | Docs     |
| `TODO_LIST.md`              | Remove stale SIMD item                          | Docs     |
| `AGENTS.md`                 | Fix architecture section, Clone count           | Docs     |
| `go.sum`                    | `go mod tidy` transitive hashes                 | Deps     |

## Remaining Issues (Tracked for Future Work)

All issues #3 through #28 above are documented here for planning. The most critical unaddressed items:

- **#11**: SDK streaming errors silently swallowed
- **#13**: `legacy_detector.go` fragile string matching → false positives
- **#17**: MultiDetector goroutine leak with no context cancellation
- **#18**: CloneSeverity/ClonePriority duplicate taxonomy
