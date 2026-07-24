# ADR-0016: JSON Tag Convention

**Date:** 2026-07-24

## Context

The codebase has two JSON tag conventions:

- `snake_case` in `domain/` and `pkg/artdupl/` (public SDK types)
- `camelCase` in `baseline/`, `cache/`, and `cmd/version` (internal types)

The `tagliatelle` linter was disabled to hide this inconsistency.

## Decision

Keep the split as intentional. Document the convention:

- **Public types** (`domain/`, `pkg/artdupl/`): `snake_case`. These types
  appear in SDK output, JSON API responses, and user-facing documentation.
  Snake_case is the standard for REST APIs and JSON data interchange.
- **Internal types** (`baseline/`, `cache/`, `cmd/version`): `camelCase`. These
  are internal file formats and CLI output. The `baseline` format is a
  persistent file format; changing its tags would break existing baseline files
  in CI pipelines. The `cache` format is ephemeral. The `version` output is
  CLI-specific.

Keep `tagliatelle` disabled. The split is documented and intentional.

## Rationale

- **Backward compatibility**: The `baseline` file format is persistent. Users
  have baseline files on disk. Changing tags would silently break
  `art-dupl check` in CI.
- **Low value**: The internal types are not consumed by external code. The
  inconsistency is invisible to users.
- **Risk vs reward**: Migration touches many tags across multiple packages,
  requires updating golden files and tests, and risks introducing subtle
  serialization bugs. The reward is enabling one linter.

## Consequences

- `tagliatelle` remains disabled globally.
- Contributors should use `snake_case` for new public types in `domain/` and
  `pkg/artdupl/`, and `camelCase` for new internal types.
- The `tagliatelle` exclusion in `.golangci.yml` is intentional, not an orphan.
