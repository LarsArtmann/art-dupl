# ADR-0015: Type-Aware Detection

**Date:** 2026-07-24

## Context

Semantic mode alpha-normalizes local identifiers to `v0`, `v1`, etc. before
hashing. This detects Type-2 clones (renamed variables) but creates a class of
false positives: two variables with the same canonical name but different static
types produce identical hashes. For example, `a.String()` where `a` is
`time.Time` would match `a.String()` where `a` is `*big.Int`, because both
canonicalize to `v0.String()`.

This is the most common false positive reported across feedback sessions
(httputil, cyberdom, go-auto-upgrade).

## Decision

Add a `--type-aware` flag that uses `go/types` to encode each local variable's
static type into the identifier hash. When type-aware mode is active, the
transformer's `Ident` case appends the type string to the canonical name before
hashing: `v0\x00time.Time` vs `v0\x00*big.Int`.

### Pipeline

1. `cmd/type_aware.go::loadTypeAwareData` drains the file channel, collects
   `.go` file paths, and loads type information via `golang.org/x/tools/go/packages`.
2. The resulting `TypeAwareData` map (filename to `*types.Info`) is passed
   through `ParseConfig.Preloaded` to the transformer.
3. In the transformer's `Ident` case, when `typeInfo != nil` and the identifier
   is a local variable, the type string is appended to the canonical name.
4. Files are replayed from the collected list for the subsequent parsing phase.

### Encoding

The type string is appended with a null byte separator to the canonical name:
`canonicalName + "\x00" + typeString`. This ensures the type information is
part of the hash without changing the identifier encoding layout (ADR-0008).

## Tradeoffs

- **10-100x slower** than syntax-only analysis: `go/packages` performs full type
  checking, including loading all imports and resolving types.
- **Not compatible with `--incremental`**: the incremental parser does not
  support preloaded type data. A warning is printed when both flags are set.
- **Not effective with `--structural` or `--exact`**: structural mode ignores
  identifiers entirely; exact mode hashes identifiers verbatim without
  canonicalization. Both combinations are rejected with a validation error.
- **Full file set required**: type checking requires all files to be known
  upfront, so the file channel is drained before analysis begins.

## Fallback

If type checking fails (missing dependencies, compilation errors, etc.), the
system falls back gracefully to syntax-only detection. The `typeData` is set to
`nil` and the transformer skips type encoding for all identifiers.

## Consequences

- Adds `golang.org/x/tools/go/packages` as a dependency.
- New `Config.TypeAware` field, `--type-aware` CLI flag.
- New `TypeAwareData` type in `syntax/golang/`.
- New `PreloadedAST` struct carries the go/packages AST + `*types.Info`.
- Validation errors for incompatible flag combinations (M04).
- Warning for `--type-aware` + `--incremental` combination (M05).
