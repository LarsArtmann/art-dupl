# ADR-0005: Split-Brain Type Unification

## Date

2026-06-22

## Status

Accepted

## Context

The art-dupl data model suffered from 8 split-brain issues where the same domain concept was modeled by parallel, independently-evolving types across `domain`, `printer`, `pkg/artdupl`, `config`, and `detection`. Each type could drift without compile-time detection.

Key issues documented in `docs/research/SPLIT-BRAIN.html`:

1. **Five parallel Clone instance types** with different field names (`TokenCount` vs `Size`), different types (`[]byte` vs `string` Fragment), and missing fields (byte positions dropped)
2. **Four parallel Clone Group types** with different member field names (`Clones` vs `Files` vs `Instances`)
3. **Triple DetectionMethod definition** with different naming conventions and different levels of type safety (typed enum, typed enum, untyped string)
4. **Parallel error sentinels** with identical messages but different pointer identities (`errors.Is()` fails cross-package)
5. **Fragment type schizophrenia** (`[]byte` in domain vs `string` in printer/SDK)
6. **Logger interface implicit** — no compile-time assertion of conformance
7. **Timeout representation mismatch** — `int` (seconds) in config vs `time.Duration` in SDK

## Decision

Resolve all split-brain issues by establishing `domain` as the single source of truth for shared types, using Go type aliases (`type X = domain.X`) to prevent drift while maintaining backward compatibility.

### Changes made

1. **Fragment type unified to `string`**: `domain.ProcessedClone.Fragment` changed from `[]byte` to `string`. Eliminates conversion tax at every printer boundary.

2. **CloneLocation fields added to domain**: `StartPos`/`EndPos` added to `domain.ProcessedClone`; `LineEnd` added to `printer.CloneOccurrenceView`. All clone types now carry the same location fields.

3. **DetectionMethod unified**: Canonical type defined in `domain/detection_method.go`. `config.DetectionMethod`, `pkg/artdupl.DetectionMethod`, and `detection` constants are now type aliases to `domain.DetectionMethod`. `detection.Config.Methods` upgraded from `[]string` to `[]domain.DetectionMethod`. The manual contract test (`detection/method_contract_test.go`) was deleted — the compiler replaces it.

4. **Error sentinels aligned**: `ErrInvalidThreshold` and `ErrThresholdTooLarge` now defined once in `domain`. `config` and `pkg/artdupl` re-export them, so `errors.Is()` works across all packages.

5. **Logger interface unified**: Explicit `Logger` interface added to `pkg/logger` with compile-time conformance assertions. `pkg/artdupl.Logger` is now a type alias to `logger.Logger`. Duplicate `noOpLogger` removed from SDK.

6. **Timeout type unified to `time.Duration`**: `config.Config.Timeout` changed from `int` (seconds) to `time.Duration`. `utils.ApplyTimeout` signature updated. CLI parses Duration at the flag boundary.

7. **CloneGroup field naming**: `printer.CloneGroup.Files` renamed to `Clones` (JSON tag unchanged for backward compatibility).

## Consequences

- All type aliases (`type DetectionMethod = domain.DetectionMethod`) make the types identical — methods defined on `domain.DetectionMethod` are automatically available everywhere.
- The SDK (`pkg/artdupl`) now imports `domain` and `pkg/logger`, which the `.go-arch-lint.yml` already permitted.
- `config.DetectionMethodHash`/`DetectionMethodArtDupl` remain as constant aliases for backward compatibility with existing config code.
- No external API breakage — JSON output is unchanged (tags preserved).
- The split-brain report (`docs/research/SPLIT-BRAIN.html`) serves as the historical record of what was wrong and why it was fixed.
