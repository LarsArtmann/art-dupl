# ADR 0002: Semantic Matching as Default

## Status

Accepted

## Context

art-dupl detects code clones using suffix tree matching on AST tokens. Two modes existed:

- **Semantic** (`--semantic`): Matches AST structure AND identifier names (e.g., `a.String()` ≠ `b.Error()`)
- **Structural** (`--structural`): Matches structure only, ignoring identifiers (e.g., `a.String()` = `b.Error()`)

The default was structural matching, which produced many false positives — similar-looking but semantically different code was flagged as clones. Users reported that 68% of initial results were noise from structural matches on common patterns like error checks, getters, and assignments.

## Decision

Changed `DefaultConfig.Semantic` from `false` to `true`. Semantic matching is now the default. Users can disable it with `--structural`.

Added `Changed()` tracking for semantic/structural flags to handle mutual exclusion with the new `true` default — previously `--semantic` was a simple boolean flag, but with `true` as default, `--structural` needs to override it.

## Consequences

### Positive

- First-use noise reduced by 68% — users see meaningful clones immediately
- Default behavior matches user expectations ("find actual duplicate code")
- Structural mode still available for broader pattern matching

### Negative

- May miss some structural clones that are valid refactoring targets
- Users who relied on structural default need to add `--structural` flag
- `Changed()` tracking adds complexity to flag parsing

### Mitigation

The `--structural` flag is clearly documented as "structure-only matching" to guide users who need broader results.
