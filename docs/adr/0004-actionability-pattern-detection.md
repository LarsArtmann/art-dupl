# ADR-0004: Actionability Pattern Detection System

## Date

2026-06-15

## Status

Accepted

## Context

art-dupl's suffix tree and hash detection algorithms find raw structural clones — any repeated token sequence above the threshold. Many of these are not actionable: they represent standard Go idioms (interface implementations, RAII defer patterns, table-driven tests, error propagation) that cannot or should not be deduplicated.

Without filtering, the signal-to-noise ratio is too low for the tool to be useful in practice. Users reported that 60-80% of reported clones were false positives in the sense that no developer would act on them.

## Decision

We implemented a multi-pattern actionability detection system in `printer/actionability.go` that classifies each clone group as `Actionable` or `NonActionable` based on 8 AST-structural patterns:

1. **testdata-pair** — Clones from testdata directory file pairs
2. **table-driven-test** — `for _, tt := range tests` + `t.Run` bodies
3. **test-scaffolding** — Temp file I/O + assertions in test files
4. **data-dominated** — ≥60% of nodes are BasicLit/KeyValueExpr
5. **signature-only** — Interface method declarations with empty bodies
6. **raii-defer** — Defer of cleanup methods (Close, Unlock, Done)
7. **error-propagation** — Standard `if err != nil { return err }`
8. **interface-implementation** — 3+ files implementing same interface method

A clone group is `NonActionable` only when **every** clone in the group matches the same pattern. If any clone differs, the group remains `Actionable`.

## Consequences

### Positive

- Dramatically improved signal-to-noise ratio
- Users can focus on clones that represent real refactoring opportunities
- Binary classification is simple to understand and act on
- Each pattern detector is independently testable

### Negative

- Binary (actionable/non-actionable) classification lacks nuance — some clones are "partially actionable"
- Adding new patterns requires modifying the `evaluateActionabilityDetailed` function
- Pattern detection adds ~5ms per clone group to analysis time
- The `everySequenceMatch` requirement means a group with 9 non-actionable + 1 actionable clone is reported as actionable

### Future Considerations

- Pattern-aware weighting system (multiply priority by pattern weight) was considered but deferred
- A `--suppress-test-low` flag for blanket suppression of test-only low-priority clones is planned
- Separate test/production thresholds would reduce test noise further
