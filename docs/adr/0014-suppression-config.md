# ADR-0014: SuppressionConfig Struct

**Date:** 2026-07-16

## Context

`printDupls`, `printCloneGroups`, and `shouldSuppressGroup` each accepted three separate parameters for suppression settings: `suppressTestLow bool`, `testThreshold int`, and `minLines int`. This led to:

- 6 call sites repeating the same 3-argument pattern
- Risk of argument-order bugs (bool/int/int is easy to swap)
- No type safety for the parameter group

## Decision

Extract a `SuppressionConfig` struct bundling all three fields:

```go
type SuppressionConfig struct {
    SuppressTestLow bool
    TestThreshold   int
    MinLines        int
}
```

All three functions now accept `SuppressionConfig` as a single parameter.

## Rationale

- **Type safety:** Named struct fields eliminate argument-order confusion.
- **Extensibility:** Adding a new suppression criterion is a 1-field change, not a 6-call-site update.
- **Readability:** `SuppressionConfig{MinLines: 5}` is self-documenting.

## Consequences

- 6 call sites updated (run_flags, baseline_cmd x2, run_all_modes, stats, cmd_test x2).
- Future suppression criteria (e.g., `MaxTokens`, `IgnorePatterns`) fit naturally.
- `domain.NewProcessedCloneGroup` constructor complements this by computing `TokenCount` automatically, ensuring the struct passed to `shouldSuppressGroup` is always consistent.
