# ADR-0012: dumpTokensOutput accepts io.Writer for testability

## Status

Accepted

## Date

2026-07-16

## Context

The `dumpTokensOutput` function (triggered by `--dump-tokens`) hardcoded `os.Stdout` as its output target. This made the function impossible to unit test without capturing stdout via process execution or redirecting file descriptors.

The original signature was:

```go
func dumpTokensOutput(ctx context.Context, cfg *config.Config) error
```

Inside, it did `w := os.Stdout` and wrote to `w`.

## Decision

Change the signature to accept `io.Writer`:

```go
func dumpTokensOutput(ctx context.Context, cfg *config.Config, w io.Writer) error
```

The caller in `run_flags.go` passes `os.Stdout`. Tests pass a `*bytes.Buffer`.

## Rationale

- **Dependency injection**: Standard Go pattern for testability. The function depends on an interface (`io.Writer`), not a concrete global (`os.Stdout`).
- **Minimal change**: Only the signature and one call site changed. No behavioral difference for users.
- **Enables unit tests**: Tests can verify the tab-separated output format without process spawning.

## Consequences

- The function signature changed, so any direct callers must be updated (only `run_flags.go` in practice).
- The `io` import replaces `os` in `dump_tokens.go`.
