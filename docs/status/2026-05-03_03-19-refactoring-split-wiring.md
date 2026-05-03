# Status Report — 2026-05-03 03:19

## Summary

Config extraction, dead code elimination, detector wiring, and file splitting. Seven commits across 50+ files.

---

## ✅ COMPLETED (this session)

| # | Task | Commit | Impact |
|---|------|--------|--------|
| 1 | **Commit staged changes** — previous session's config/printer/detection refactoring | `508236e` | 41 files, -62 net lines |
| 2 | **Wire TODO/Legacy detectors** — enabled as valid methods, wired through MultiDetector | `f6466ba` | Users can now use `-m todos`, `-m legacy` |
| 3 | **Consolidate threshold errors** — moved ErrInvalidThreshold/ErrThresholdTooLarge to config/ as single source of truth | `a834f19` | Eliminated duplicate error definitions |
| 4 | **Split detection/todos.go** — 352L → 3 focused files (issue_helpers, todo_detector, legacy_detector) | `9049d75` | Clear separation of concerns |
| 5 | **Split config/config.go** — 344L → 3 files (config, config_io, config_validate) | `fff434b` | Struct/IO/validation separation |
| 6 | **Split cmd/run_analysis.go** — 450L → 3 files (run_analysis, run_hash, run_printer) | `25a6602` | Core/hash/printer separation |

### Session Stats

- **Commits**: 7 (including prior session commit)
- **Build**: Clean
- **Tests**: 22/22 packages pass
- **Lint**: 0 issues

---

## Architecture Changes

### Detection Methods Now Fully Wired

All four detection methods are now available through the CLI:

```
art-dupl -m art-dupl    # suffix tree (default)
art-dupl -m hash         # rolling hash
art-dupl -m todos        # TODO/FIXME/HACK comments
art-dupl -m legacy       # deprecated function detection
art-dupl -m "hash,art-dupl,todos,legacy"  # all methods
```

### Threshold Validation — Single Source of Truth

Before: `config/validateThreshold` used `errors.NewValidationError`, `pkg/artdupl` had its own `ErrInvalidThreshold` and `ErrThresholdTooLarge`.

After: Both sentinels live in `config/enum_helpers.go`, `pkg/artdupl/errors.go` aliases them.

### File Organization

```
detection/
  issue_helpers.go      # shared types + generic helpers
  todo_detector.go      # TodoDetector
  legacy_detector.go    # LegacyDetector
  multidetector.go      # MultiDetector dispatch
  detector.go           # MethodDetector interface

config/
  config.go             # Config struct + DetectionMethods
  config_io.go          # Load/Save config
  config_validate.go    # ValidateConfig + validators
  enum_helpers.go       # sentinel errors + generic marshaling

cmd/
  run_analysis.go       # core pipeline (275L)
  run_hash.go           # hash-only path (137L)
  run_printer.go        # printer factory (47L)
```

---

## ❌ REMAINING (from TODO_LIST.md)

| Priority | Task |
|----------|------|
| HIGH | Implement TokenValue type with validation |
| MEDIUM | Introduce ProcessedClone DTO to decouple Printer from syntax.Node |
| MEDIUM | Consolidate three parallel Clone types |
| LOW | Refactor syntax/golang/transform.go (355L switch) |
| LOW | Implement SIMD TODOs |
| LOW | Archive old docs/status/ files |
