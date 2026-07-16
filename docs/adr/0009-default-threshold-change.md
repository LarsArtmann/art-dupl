# ADR-0009: Default Threshold Change (1 to 5)

**Date:** 2026-07-16
**Status:** Accepted

## Context

The original default threshold was 1, which reported any duplicated AST node
sequence of length 1 or more. This produced excessive noise: single-statement
clones, one-liner boilerplate, and trivial patterns flooded the output.

Real-world testing across 15 projects (6,222 Go files, 320 templ files) showed
that meaningful duplication starts at threshold 5. Below that, the vast majority
of reported clones are non-actionable boilerplate (error checks, defer patterns,
test scaffolding).

## Decision

Change the default threshold from 1 to 5.

With statement-level tokenization, each Go statement produces one composite
token. A threshold of 5 means "report clone groups where at least 5 duplicated
statements appear." This filters:

- Single-call expressions (`errors.New("foo")`)
- Error-check boilerplate (`if err != nil { return err }`)
- Defer cleanup patterns (`defer m.Unlock()`)
- Single assignments and declarations

while catching meaningful cloning:

- Duplicated function bodies
- Copy-pasted business logic
- Repeated validation chains

## Consequences

- Users who want more sensitivity can lower to 3 (`-t 3`)
- Users who want less noise can raise to 10+ (`-t 10`)
- The actionability pattern system provides additional filtering on top of
  the threshold, so even at threshold 5, non-actionable clones are suppressed
- The `--test-threshold` flag provides separate filtering for test files
  (default: `max(30, threshold)` when not explicitly set)
- Validated at 100% precision across 15 real-world projects
