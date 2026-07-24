# art-dupl SDK Design

The public SDK lives in `pkg/artdupl/`. This document records the key design decisions.

## Interface

```go
type Detector interface {
    FindClones(ctx context.Context, files []string) (*Result, error)
    FindClonesStreamResult(ctx context.Context, files []string) (<-chan StreamResult, error)
    Close() error
}
```

- `FindClones` returns complete results (blocks until done).
- `FindClonesStreamResult` emits `StreamResult` values on a channel. A final
  `StreamResult` with `Err != nil` signals pipeline failure.
- `Close` releases resources.

## Types

All types are in `pkg/artdupl/types.go`.

| Type         | Purpose                                                                       |
| ------------ | ----------------------------------------------------------------------------- |
| `Options`    | Configuration: threshold, methods, workers, timeout, type-aware, callbacks    |
| `Result`     | Complete output: clone groups + summary + metadata                            |
| `CloneGroup` | Hash, clones, size, line count, detection method                              |
| `Clone`      | Embeds `domain.CloneRef` (Filename, LineStart, LineEnd, Fragment) + positions |
| `Summary`    | Stats: total files, clones, groups, analysis time                             |
| `Metadata`   | Version, timestamp, config hash, toolchain                                    |

## Usage

```go
detector, err := artdupl.New(artdupl.DefaultOptions())
if err != nil { return err }
defer detector.Close()

result, err := detector.FindClones(ctx, []string{"./src"})
if err != nil { return err }

for _, group := range result.CloneGroups {
    fmt.Printf("Clone group %s: %d occurrences\n", group.Hash, len(group.Clones))
}
```

## Design Decisions

1. **Zero imports of `config/` and `errors/`**: The SDK is independent. It
   aliases `domain` types (`DetectionMethod`, `FileReaderFunc`) and `pkg/logger.Logger`
   rather than importing the full config or errors packages.

2. **`CloneRef` embedding**: `Clone` embeds `domain.CloneRef` so all
   clone-bearing types share the same location fields (`Filename`, `LineStart`,
   `LineEnd`, `Fragment`) without field-name drift.

3. **Streaming via channels**: `FindClonesStreamResult` returns a channel of
   `StreamResult{Group, Err}` values. This allows incremental processing of
   large codebases without buffering all results.

4. **`DefaultOptions()`**: Provides sensible defaults (threshold 15, 4 workers,
   30min timeout). Callers override individual fields.

5. **Type aliases over redefinition**: `DetectionMethod`, `FileReaderFunc`, and
   `Logger` are aliases (`type X = Y`), not new types. This ensures the SDK is
   compatible with domain-level code without requiring conversion functions.

## Architecture Constraints

- The `detection` package uses `[]domain.DetectionMethod` (typed, not `[]string`).
- Arch-lint (`.go-arch-lint.yml`) enforces zero imports of `config/` and `errors/`
  from `pkg/artdupl/`.
- The SDK supports `Options.TypeAware = true` for go/types-based detection.
  This encodes each variable's static type into the hash, eliminating false
  positives where same-name methods on different types (e.g., `time.Time.String`
  vs `*big.Int.String`) match. Falls back to syntax-only if type checking fails.
  10-100x slower than syntax-only mode.
