# ADR-0013: Typed Exit Codes

**Date:** 2026-07-16

## Context

art-dupl previously returned exit code 1 for all errors (general, config, timeout) and 130 only for SIGINT. CI/CD pipelines and shell scripts could not distinguish between "no clones found" (exit 0), "bad configuration" (should warn the user, not page anyone), and "internal error" (potentially a bug requiring investigation).

## Decision

Implement a 5-level exit code system via `cmd.ExitCodeForError(err)`:

| Code | Meaning           | Error Type                                     |
| ---- | ----------------- | ---------------------------------------------- |
| 0    | Success           | nil                                            |
| 1    | General error     | default                                        |
| 2    | Config/validation | `duplerrors.ValidationError`, `ConfigError`    |
| 3    | Internal error    | `duplerrors.InternalError`                     |
| 130  | Interrupted       | `context.Canceled`, `context.DeadlineExceeded` |

Uses `errors.Is` for wrapped error chains, so wrapped cancellation and nested wrapping are correctly detected.

## Rationale

- **Code 2** lets CI distinguish "user mistake" (fix the config) from "tool failure" (file a bug).
- **Code 3** is separated from general errors so monitoring can alert differently.
- **Code 130** follows the Unix convention (128 + SIGINT=2).
- `errors.Is` chain unwrapping handles `fmt.Errorf("...: %w", err)` wrapping patterns.

## Consequences

- `main.go` calls `ExitCodeForError(err)` instead of hardcoded `os.Exit(1)`.
- Exit codes are documented in `--help` output and `HOW_TO_USE.md`.
- CI scripts can use `if [ $? -eq 2 ]` to fail-fast on config errors.
