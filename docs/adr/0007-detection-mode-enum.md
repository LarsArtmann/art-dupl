# ADR-0007: DetectionMode enum replacing Semantic/Exact bools

## Status

Accepted

## Context

The detection mode was represented as two boolean fields in Config:
`Semantic bool` (default true) and `Exact bool`. The `--structural` CLI flag
had no corresponding config field, it was lossy, folded into `Semantic = false`.
This created a class of "silent-drop" bugs where invalid combinations could
slip through, and the config layer couldn't distinguish "user didn't set
semantic" from "user passed --structural".

## Decision

Replace the two booleans with a single `DetectionMode` enum in `config/`:

```go
type DetectionMode string

const (
    DetectionModeSemantic   DetectionMode = "semantic"
    DetectionModeExact      DetectionMode = "exact"
    DetectionModeStructural DetectionMode = "structural"
)
```

The enum is set from CLI flags in `config_builder.go`:

- `--structural` → `DetectionModeStructural`
- `--exact` → `DetectionModeExact`
- `--semantic` (or default) → `DetectionModeSemantic`

## Rationale

- **Eliminates invalid states**: Three modes are three values, not a 2-bit space
  with impossible combinations.
- **Lossless mapping**: `--structural` is now a first-class config value.
- **Self-documenting**: `DetectionMode: "structural"` in JSON is clearer than
  `Semantic: false`.
- **Helper method**: `DetectionMode.IsSemantic()` provides backward-compatible
  boolean semantics for display/metadata code.

## Consequences

- JSON config files using `"semantic": true` must migrate to
  `"detectionMode": "semantic"`. The old format is no longer supported.
- `config.DetectionMode` uses `pkg/enum` helpers for MarshalJSON/UnmarshalJSON.
- All consumers updated to use `cfg.DetectionMode.IsSemantic()` where boolean
  semantics were needed (HTML metadata, stats display, etc.).
