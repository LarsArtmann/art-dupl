# ADR-0010: encoding/json/v2 Migration

**Date:** 2026-07-16
**Status:** Accepted

## Context

The project used `encoding/json` (v1) for all JSON serialization. The Go team
released `encoding/json/v2` as an experimental package (via `GOEXPERIMENT=jsonv2`)
with significant improvements:

- Better performance (2-5x faster marshaling/unmarshaling)
- Improved type handling with `omitzero` (omits zero-value fields without
  needing `omitempty` on every field)
- Cleaner API for custom marshalers
- Stricter unmarshaling (no silent type coercion bugs)

## Decision

Migrate to `encoding/json/v2` behind `GOEXPERIMENT=jsonv2`.

### Changes Required

- **go.mod**: No changes needed (uses stdlib experimental package)
- **flake.nix**: `GOEXPERIMENT=jsonv2` set in devShell `env`
- **CI**: `GOEXPERIMENT=jsonv2` exported in GitHub Actions
- **Build**: `export GOEXPERIMENT=jsonv2` required for non-Nix builds
- **Convention**: Use `omitzero` (not `omitempty`) on custom `MarshalJSON` types,
  and `format:nano` for `time.Duration` fields

## Consequences

- Non-Nix developers must export `GOEXPERIMENT=jsonv2` manually
- The Nix devShell and CI handle this automatically
- `omitzero` is preferred over `omitempty` for custom marshalers
- This is a forward-compatible migration: when json/v2 becomes non-experimental,
  the `GOEXPERIMENT` flag can be removed with no code changes
