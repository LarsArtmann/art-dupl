# Denylist Pattern Deprecation Plan

**Date:** 2026-07-27
**Status:** Active — deprecation in progress, removal targeted for v1.1

## Background

The 21-pattern denylist in `printer/actionability/` served as the primary actionability
classifier from v0.1 through v0.5. The property-based extractability engine (ADR-0017)
now handles most of the classification. The denylist patterns remain as a safety net.

## Audit: Pattern → Property Coverage

| Pattern | Subsumed by Property? | Action |
| --- | --- | --- |
| signature-only | N/A (structural, not property) | **Keep** — no property equivalent |
| interface-implementation | N/A | **Keep** — structural pattern |
| interface-method | N/A | **Keep** — structural pattern |
| raii-defer | Property 3 (ROI) | **Deprecate** after coverage proven |
| error-propagation | Property 2 (control-flow) | **Deprecate** after coverage proven |
| guard-clause | Property 2 (control-flow) | **Deprecate** after coverage proven |
| assign-error-check | Property 2 (control-flow) | **Deprecate** after coverage proven |
| bool-guard | Property 2 (control-flow) | **Deprecate** after coverage proven |
| single-call-expression | Property 3 (ROI) | **Deprecate** after coverage proven |
| single-simple-statement | Property 3 (ROI) | **Deprecate** after coverage proven |
| single-declaration | Property 3 (ROI) | **Deprecate** after coverage proven |
| test-helper-delegate | N/A | **Keep** — test-specific |
| error-wrapping | Property 3 (ROI) | **Deprecate** after coverage proven |
| assertion-chain | N/A | **Keep** — test-specific |
| cobra-boilerplate | N/A | **Keep** — framework-specific |
| test-data-pair | N/A | **Keep** — test-specific |
| table-driven-test | N/A | **Keep** — test-specific |
| test-scaffolding | N/A | **Keep** — test-specific |
| data-dominated | N/A | **Keep** — structural pattern |
| describe-table | N/A | **Keep** — test-specific |
| builder-callback | N/A | **Keep** — structural pattern |
| property-* | N/A (IS the property engine) | **Keep** — the new engine itself |

## Timeline

- **v1.0**: All patterns active. Property engine runs as second-pass fallback.
- **v1.0+**: Collect data on which patterns are fully subsumed by properties.
- **v1.1**: Deprecate fully-subsumed patterns (add deprecation warning log).
- **v1.2**: Remove deprecated patterns. Property engine becomes primary classifier.

## Migration Guide for `--disable-pattern` Users

If you use `--disable-pattern <name>`, the flag will continue to work through v1.1.
In v1.2, disabled patterns that have been removed will be silently ignored (no error).

To prepare: transition to `--no-actionability` (disables ALL filtering) or use
`//art-dupl:accept` directives for per-clone control.
