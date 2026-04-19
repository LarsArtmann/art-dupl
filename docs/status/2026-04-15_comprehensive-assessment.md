# Comprehensive Architectural Assessment & Execution Plan

**Date**: 2026-04-15
**Session**: Hardening sprint on branch `fork`
**Status**: Build passing, all tests green, 10 commits pushed

---

## A. Session Summary

### What was done (10 commits)

| Commit | Description |
|--------|-------------|
| `bfce6fd` | Fix gogenfilter API migration (functional options) |
| `b2d72fe` | Delete dead code: `pkg/errors/`, `testutils/`, `detection/simple_detector.go` |
| `42d0e88` | Fix lint warnings in `config_builder.go`, fix justfile issues |
| `48f83de` | Extract `addSharedFlags()` to deduplicate 18 shared flags |
| `d245eee` | Add configuration builder module |
| `520f219` | Add missing `go.sum` entries from `go mod tidy` |
| `6c7755d` | Fix `TestContextTimeoutExpired` timing race (5ms/10ms → 1ms/50ms) |
| `3098157` | Fix gci import grouping in `config_builder.go` |
| `b895849` | Add test coverage for syntax, cli, printer packages |

### Key metrics

| Package | Before | After | Delta |
|---------|--------|-------|-------|
| syntax | 67.6% | 81.1% | +13.5% |
| cli | 62.5% | 75.0% | +12.5% |
| printer | 63.5% | 66.5% | +3.0% |
| All tests | 2 failures | 0 failures | Fixed |

---

## B. Ghost Systems (Dead Code)

### Deleted this session
- `pkg/errors/` — zero importers, unused error package
- `testutils/` — zero importers, unused test utility
- `detection/simple_detector.go` — zero importers, unused interface implementation

### Still present (needs user decision)

| Package | Lines | Status | Risk to delete |
|---------|-------|--------|----------------|
| `internal/enum/` | ~260 | Zero external imports, dead | Low — only its own tests |
| `git/` | ~368 | Zero production imports, dead | Low — only its own tests |
| `migration/` | ~342 | Only imported by `adapter/` (also dead) | Low — dead dependency chain |
| `adapter/` | ~100 | Only imported by `migration/` (also dead) | Low — dead dependency chain |

**Dead dependency chain**: `adapter/` ← `migration/` ← nothing. All four packages are unreachable from production code.

---

## C. Split Brain Analysis

### SB-1: 4 Clone types + 3 CloneGroup types

| Type | Package | Key difference |
|------|---------|----------------|
| `domain.Clone` | `domain` | Strong types (LineNumber, StringID, ComplexityScore) |
| `artdupl.Clone` | `pkg/artdupl` | Plain primitives, has `Size` field |
| `printer.clone` | `printer` | Unexported, `fragment []byte`, `fileSize`, `classification` |
| `printer.JSONClone` | `printer` | Minimal fields for JSON serialization |

**Conversion functions**: No unified conversion. `adapter/` package tried to bridge but is dead code. Each consumer does its own manual conversion.

**Impact**: Any field addition requires changes in 4+ places. No compile-time safety across boundaries.

### SB-2: Dual analysis pipelines

- `cmd/run_analysis.go` — main CLI pipeline (~447 lines)
- `pkg/artdupl/detector_pipeline.go` — SDK pipeline (~100 lines)

Both parse files, run detection, and produce output. Divergent logic, different error handling patterns.

### SB-3: Enum system duplication

- `internal/enum/` — DEAD, never imported
- `config/detectionmethod.go` — Active, ~426 lines, handles DetectionMethod, OutputFormat, SortBy, DiffMode, FileType
- `domain/types_enums.go` — Active, handles CloneSeverity, FileProcessingState

---

## D. Error Handling Inconsistency

| Pattern | Used in | Count |
|---------|---------|-------|
| `duplerrors.Wrap*` (structured) | `cmd/`, `printer/` | ~15 sites |
| `fmt.Errorf("...: %w", err)` | `pkg/artdupl/`, `adapter/`, `job/`, `printer/` | ~13 sites |
| `errors.New` (sentinels) | `config/`, `domain/`, `pkg/artdupl/` | ~10 sites |

**The `cmd/` package consistently uses `duplerrors.Wrap*`, but most other packages use bare `fmt.Errorf`.** This means callers outside `cmd/` cannot distinguish error types programmatically.

---

## E. Largest Files

| File | Lines | Issue |
|------|-------|-------|
| `printer/html.go` | ~1484 | Mixes Go with embedded JS/CSS/HTML, 11 functions at 0% coverage |
| `config/detectionmethod.go` | ~426 | God file: DetectionMethod, OutputFormat, SortBy, DiffMode, FileType all in one |
| `printer/diff.go` | ~398 | Complex diff algorithm |
| `cmd/run_analysis.go` | ~447 | Main pipeline, high cyclomatic complexity |

---

## F. Self-Assessment: What We Could Have Done Better

1. **Should have run `go mod tidy` immediately** after the gogenfilter migration — this would have prevented 240 BDD test failures from the missing `doublestar` entry.
2. **Should have tested BDD suite specifically** after changes that affect binary building.
3. **Wasted time fighting broken Go 1.26.1 Nix installations** — should have verified PATH Go 1.26.0 first.
4. **Should have scoped coverage improvements more aggressively** — the `printer/` package still has 11 functions at 0% (mostly HTML diff view functions).

---

## G. Remaining Work (Execution Plan)

See next section for the full prioritized plan.
