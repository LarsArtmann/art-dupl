# Status Report — Alpha-Normalization & Three-Mode System

**Date:** 2026-06-20 18:23
**Branch:** fork
**Head:** `b86e7de`

---

## Summary

This session completed the foundation tier of the Superb Clone Detection Engine plan: **alpha-normalization (T2)**, **three-mode detection (T5)**, **clone type classification (T6)**, and **coverage hardening (T19, T18)**. The tool now detects **Type 2 (renamed-variable) clones** — the most common real-world duplication — which was previously impossible.

---

## Completed This Session

| Task | Commit | Impact |
|------|--------|--------|
| **T6** — Clone Type Classification | `359f3d7` | `classifyCloneType()` compares Name fields across fragment subtrees; emits `clone_type` (type-1/2/3) in JSON, SARIF, and `--rich-text`. |
| **T19** — Domain Coverage | `692c587` | Domain package: 58.6% → **100%** (all enum methods tested). |
| **T18** — Detection Coverage | `9fecc7f` | Detection package: 61.8% → **92.7%** (adapters, dispatch, cancellation, real duplicate data). |
| **T2** — Alpha-Normalization | `91554e5` | Per-function symbol table canonicalizes locals (params/receiver/body vars) to v0/v1/... before hashing. Selectors/types/field names preserved. |
| **T5** — Three-Mode System | `91554e5` | `--semantic` (default, alpha-normalized), `--exact` (verbatim names), `--structural` (shape only). `bool` → `DetectionMode` enum threaded through job/cmd. |
| **Dedup + Docs** | `b86e7de` | Eliminated `flattenSubtree` (reuse `syntax.Serialize`); updated AGENTS.md/HOW_TO_USE.md/FEATURES.md. |

---

## Key Achievement: Type 2 Clone Detection

Before this session, `--semantic` mode was **backwards**: it baked exact identifier names into the token hash, making matching *stricter* than structural. `processUser` and `processOrder` with identical bodies were *rejected*.

Now, semantic mode **alpha-normalizes** local identifiers before hashing. Two functions with identical structure but completely different variable names produce **identical token streams** and are detected as clones, classified as **type-2**.

### Dogfooding Validation (art-dupl on itself, `-t 20`)

| Mode | Clone Groups | type-1 | type-2 |
|------|-------------|--------|--------|
| `--semantic` (default) | 7 | 3 | **4** |
| `--exact` (old behavior) | 4 | 4 | 0 |
| `--structural` | 7 | — | — |

The 4 type-2 clones are genuine: e.g. `DetectionMethods.Strings()` (receiver `dm`) vs a structurally identical method (receiver `methods`) — previously invisible.

---

## How Clone Type Classification Works

The suffix tree matches on `Node.Type` only; `Name` is never part of the matching criterion. After alpha-normalization, two renamed functions share identical `Type` sequences but different `Name` sequences. `classifyCloneType` walks the full subtree (via `syntax.Serialize`) of each fragment and compares `Name` at every node position:

- **Type 1** (exact): all Names identical across all fragments.
- **Type 2** (parameterized): structure matches but ≥1 Name differs.
- **Type 3** (near-miss): fragments differ in length (defensive fallback).

`o.Name` keeps the original identifier; `o.Type` uses the canonical name — so classification works without mode-specific logic.

---

## Three-Mode System

```
--semantic (default)   alpha-normalize locals → hash → match (Type 1 + Type 2)
--exact                hash verbatim names → match (Type 1 only)
--structural           ignore all names → match by shape (loosest)
```

The old `semantic bool` threaded through `job.Parse` is replaced by `golang.DetectionMode` (an enum). A `detectionMode()` helper in `cmd` resolves config flags to the active mode. The SDK (`pkg/artdupl`) maps its `Semantic bool` via `toDetectionMode()`.

---

## Coverage

| Package | Before | After |
|---------|--------|-------|
| domain | 58.6% | **100%** |
| detection | 61.8% | **92.7%** |
| syntax/golang | 95.9% | **96%+** (normalizer fully tested) |

---

## Remaining Work

| Task | Status | Notes |
|------|--------|-------|
| T10 — Baseline/CI Mode | Next | `baseline` + `check` subcommands |
| T11 — Extractability Score | Pending | Refactoring hints |
| T17 — Property Tests | Pending | Suffix tree correctness |
| T1 — Statement Tokenization | Deferred | High risk; `serial()` still inflates thresholds |
| T12-T16 — Architecture | Pending | Printer decoupling, type consolidation |
| T20-T24 — Ecosystem | Pending | Benchmarks, GitHub Actions, pre-commit |

---

## Known Limitations

1. **Normalizer is heuristic** — flat per-function symbol table (no nested-scope shadowing resolution). Shadowed variables are rare in duplicated code; first declaration wins. `go/types` would give precise resolution but adds import/type-checking overhead.
2. **`serial()` still inflates thresholds** — T1 (statement-level tokenization) remains the root-cause fix for false positives from structural wrapper inflation. Deferred due to cascading risk.
3. **Templ has no semantic mode** — `syntax/templ/` matching is purely structural.
