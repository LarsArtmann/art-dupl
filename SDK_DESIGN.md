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

4. **`DefaultOptions()`**: Provides sensible defaults (threshold 5, matching
   `config.DefaultThreshold`; 4 workers, 30min timeout). Callers override
   individual fields.

5. **Type aliases over redefinition**: `DetectionMethod`, `FileReaderFunc`, and
   `Logger` are aliases (`type X = Y`), not new types. This ensures the SDK is
   compatible with domain-level code without requiring conversion functions.

6. **Classification-free boundary (ADR-0025)**: The SDK emits positions,
   fragments, and group identity (`CloneGroup.Hash`) — never actionability
   verdicts, categories, or boilerplate-pattern labels. Classification is an
   OUTPUT-layer concern owned by `printer/actionability` and applied in the
   CLI pipeline (gated by `--no-actionability`). The go-finding adapter
   (`printer/finding`) adds classification metadata under the `art-dupl/`
   metadata namespace; the BuildFlow provider deliberately ships without it.
   This keeps the SDK contract stable while the heuristic taxonomy evolves.

## Downstream Consumers

```
                ┌─────────────────────────────┐
                │        pkg/artdupl          │
                │   Detector (this contract)  │
                └──────────┬──────────────────┘
                           │
          ┌────────────────┼───────────────────┐
          │                │                   │
┌─────────▼─────────┐ ┌────▼─────┐   ┌─────────▼──────────┐
│   cmd/ pipeline   │ │ pkg/     │   │ external SDK users │
│ (classification   │ │ provider │   └────────────────────┘
│  via actionability│ │ (toolsdk)│
│  + printer/       │ └────┬─────┘
│  finding adapter) │      │ blank import in BuildFlow
└───────────────────┘      │ toolsdk.Register at init
                           ▼
                    BuildFlow core lane
```

**`printer/finding` (CLI output adapter)**: converts
`domain.ProcessedCloneGroup`s into go-finding `Finding`s via the validated
`Builder` + `Template`. Carries the
`GroupID` contract (group content hash, 16-char lowercase hex — the same id
as JSON `clone_groups[].hash`, the SARIF `go-finding/groupId` property, and
the `--lsp` diagnostic `data.group_id`), semantic tags (`duplication` +
`type-1/2/3`), byte-derived start/end columns, and the classification
metadata under `art-dupl/*` metadata keys. Accept-directive groups can be
surfaced as `Suppression{InSource}` findings (opt-in via `Options.EmitSuppressedAccepted`).
CLI SARIF and `Report.ToSARIF()` are pinned equivalent by a cross-check test.
See `docs/research/2026-10-05_go-finding-deep-dive.html` §05 for the full
ledger (verdicts, SARIF divergence table, re-scored capabilities).

**`pkg/provider` (BuildFlow toolsdk provider)**: self-registers via
package-level `toolsdk.Register`; BuildFlow wires it with a single blank
import. Detect runs the PUBLIC SDK (semantic mode) on the working dir from
context and returns one finding per clone occurrence with
positions/snippets/GroupID/severity/columns but no classification metadata
(ADR-0025). Declared options: `threshold` (int, 1–1000, default 5, applied
via `toolsdk.WithOptions`) and `emit-suppressed-accepted` (bool, default
false). Crawl mirrors CLI defaults (`.go` +
`.templ`, vendor/generated/examples excluded, `.gitignore` honored via
`gitignore.LoadTree`) EXCEPT `_test.go` files are included — CLI
default-ignores them, so pipeline group counts run higher than a default
CLI run on the same tree. Severity is advisory-capped at warning (the
pre-cap value survives in an `original-severity-` tag) so detector findings
can never fail BuildFlow gates keyed on error-or-above.
`artdupl.ErrNoDuplicatesFound` maps to an empty finding list — a clean repo
is success, not an error.

## Architecture Constraints

- The `detection` package uses `[]domain.DetectionMethod` (typed, not `[]string`).
- Arch-lint (`.go-arch-lint.yml`) enforces zero imports of `config/` and `errors/`
  from `pkg/artdupl/`.
- The SDK supports `Options.TypeAware = true` for go/types-based detection.
  This encodes each variable's static type into the hash, eliminating false
  positives where same-name methods on different types (e.g., `time.Time.String`
  vs `*big.Int.String`) match. Falls back to syntax-only if type checking fails.
  10-100x slower than syntax-only mode.
- The SDK supports `Options.SuggestGenerics = true` to find clones where the
  algorithm is identical but local variable types differ — candidates for Go
  generics extraction. Uses type-erased hashing, then classifies by comparing
  types at corresponding positions. Same ~100x cost as `TypeAware`.
- Setting BOTH `TypeAware` and `SuggestGenerics` runs the combined two-pass
  mode (ADR-0026): one shared type-check, a type-aware pass plus a type-erased
  pass; the erased pass contributes only families with ≥2 divergent type
  positions. Works for `FindClones` and `FindClonesStreamResult`
  (`pkg/artdupl/combined.go`).
