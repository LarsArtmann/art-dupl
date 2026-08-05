# ADR-0017: Property-Based Classification Model

**Date:** 2026-07-27
**Status:** Accepted

## Context

The DiscordSync feedback (82 clone groups at `-t 1`, 97.5% false positives) exposed
a fundamental ceiling in the pattern-denylist approach to actionability classification.
The denylist of 23 patterns cannot converge on zero false positives because every new
codebase reveals new idioms not in the list. Adding patterns is whack-a-mole.

The root cause: type information loaded by `--type-aware` was destroyed in the
transformer (hashed into `Node.Type` and never stored). The actionability layer
operated on pure AST shape with no access to:

- Enclosing function signatures (return arity)
- Variable types (helper detection)
- Error references in call arguments

## Decision

Replace the denylist-first approach with a **property-based classification engine**
that evaluates 4 computable properties defining "harmful duplication":

1. **Mechanically extractable** — always true (wrap in func, capture free vars)
2. **Control-flow extractable** — false when returns/breaks are forced by void signature
3. **ROI positive** — false when extraction costs more tokens than it saves
4. **Parameterizable** — false when clones differ only in domain-value string literals

The engine runs as a **second-pass fallback** after the denylist patterns. Patterns
remain the primary classifier (specific, well-tested). The property engine catches
cases the patterns miss.

### Architectural changes

- `syntax.Node` and `domain.CloneNode` gained `VarType string` and `EnclosingReturnArity int32`
- The transformer stores type strings and tracks enclosing return arity during AST traversal
- The bridge (`syntaxToCloneNode`) carries these fields to the actionability layer
- `ExtractabilityAnalysis` struct with 4 properties + confidence + reason
- `CloneActionability` enum gained `LowConfidence` tier (0.5-0.8 confidence)
- Bridge pattern fixes: raii-defer (bare Ident + FuncLit), error-propagation (any CallExpr),
  bool-ok guard pattern

### Graceful degradation

When type info is unavailable (syntax-only mode):

- `EnclosingReturnArity` is still computed structurally from AST
- `VarType` is empty (only populated with `--type-aware`)
- Property engine runs conservatively (defaults to "harmful" when uncertain)
- Denylist patterns remain fully functional

## Consequences

- **Positive:** Generalizes to any codebase without pattern maintenance
- **Positive:** Type-aware mode now provides actionable data to the classification layer
- **Positive:** Confidence scoring enables nuanced reporting (Actionable / LowConfidence / NonActionable)
- **Negative:** `syntax.Node` struct grew by ~24 bytes (VarType string header + EnclosingReturnArity)
- **Neutral:** Denylist patterns are deprecated, not removed (safety net until property coverage is proven)
- **Risk:** Property engine could suppress true positives (mitigated by running after patterns, defaulting to harmful)

## References

- [DiscordSync feedback](../feedback/new/2026-07-27-discordsync-t1-82-groups-97pct-false-positives.md)
- [Pareto execution plan](../planning/2026-07-27_10-31_ZERO-FP-FN-EXTRACTABILITY-ENGINE.md)
- ADR-0004: Actionability pattern detection (the denylist being supplemented)
- ADR-0015: Type-aware detection (the foundation this builds on)
