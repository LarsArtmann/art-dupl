# Status Report: Architecture Debt Resolution (5 Issues)

**Date:** 2026-05-17 23:23  
**Session:** 5 critical architecture fixes identified, 2.5 completed  
**Branch:** fork  
**Status:** PARTIAL — 2 issues fully resolved, 1 partially resolved, 2 not started

---

## Executive Summary

User provided 5 deep architecture issues ("why this is stupid" analysis). After thorough READ/UNDERSTAND/RESEARCH/REFLECT cycle, I verified each claim against actual code and executed fixes in dependency order. Two and a half issues are resolved with all tests green. Three remain for next session.

---

## A) FULLY DONE ✅

### #5 — `--structural` was the default, producing 68% garbage on first use

**Status:** ✅ RESOLVED  
**Files changed:** `config/config.go`, `cmd/flags.go`, `config/config_test.go`

**What was wrong:** `DefaultConfig.Semantic = false`. Dogfooding on go-cqrs-lite showed 244 clone groups (structural) vs 78 (semantic) — 68% noise. Every new user's first experience was garbage output.

**What changed:**
- `DefaultConfig().Semantic` changed from `false` to `true`
- `--semantic` flag default changed from `false` to `true`
- `--structural` flag help text rewritten: "disable semantic detection; match by structure only (increases false positives)"
- `--semantic` flag help text: "enable semantic-aware detection (on by default; matches by structure AND identifier names)"
- Test `TestSemanticField` updated: expects `true` as default

**Why this matters:** First-run experience now shows meaningful results. `--structural` becomes the opt-in power-user mode.

---

### #1 — `mergeConfig` was a manually-maintained 170-line field list

**Status:** ✅ RESOLVED  
**Files changed:** `config/config_merge.go` (170 lines → 46 lines)

**What was wrong:** `mergeConfig()` had `//nolint:funlen,gocognit,gocyclo,cyclop` confession. Every new Config field MUST be added here or it silently vanishes. RichText and Workers were already forgotten once. The compiler can't save you.

**What changed:**
- Replaced 170-line manual field-by-field merge with 30-line reflection-based implementation
- `mergeConfig()` now iterates `reflect.ValueOf(cfg).Elem()` fields automatically
- `isFieldZero()` handles bool/int/string/slice with explicit zero-value detection
- Adding a new field to `Config` struct is automatically picked up — zero maintenance
- All existing tests pass unchanged

**Code reduction:** 160 lines deleted, 50 added. Net: -110 lines of fragile code.

---

## B) PARTIALLY DONE 🔧

### #2 — `skipZeroValues` on bool fields makes `false` flags silently lose to file config

**Status:** 🔧 PARTIALLY RESOLVED (semantic/structural only)  
**Files changed:** `cmd/config_builder.go`

**What was wrong:** `mergeCLIConfig` uses `skipZeroValues=true`. For bool fields, `false` is the zero value, so CLI `false` can never override a file config `true`. The `--structural` band-aid at `config_builder.go:73` proved the general case was broken.

**What changed (partial fix):**
- Added `SemanticSet` and `StructuralSet` bool fields to `FlagValues`
- `extractFlagValues` now tracks `cmd.Flags().Changed("semantic")` and `cmd.Flags().Changed("structural")`
- `validateMutualExclusion` now uses `*Set` fields (only errors when BOTH explicitly set)
- `BuildConfigFromFlags` uses `flags.StructuralSet` instead of `flags.Structural` for override
- The `--structural` band-aid now works correctly with `--semantic` as default `true`

**What remains broken:**
- The GENERAL case is still broken. Any future "disable-X" bool flag will silently lose to file config
- `applyBooleanFlags` in config_builder.go:244 still only propagates `true` values
- 13 other bool flags (`--vendor`, `--include-sqlc`, `--profile`, etc.) all have the same potential issue
- Proper fix: extend `Changed()` tracking to ALL bool flags, or use a `map[string]bool` of changed flags

**Risk:** Medium. Currently only semantic/structural are affected because they're the only bool flags where `false` is meaningful AND file config can set `true`. Other bool flags (`--vendor`, `--include-sqlc`) default to `false` and file config would also set `false`, so the bug is latent but not triggered.

---

## C) NOT STARTED ⏳

### #3 — `transform.go` throws away Go identifier names — `syntax.Node` is semantically blind

**Status:** ⏳ NOT STARTED  
**Files to change:** `syntax/syntax.go`, `syntax/golang/transform.go`, `printer/actionability.go`

**Original claim:** "Partially inaccurate" — `encodeSemanticType` DOES hash names into the Type field in semantic mode. But `Node` has no `Name` field, so downstream analysis like `containsNilIdentifier` can't distinguish `nil` from `err`.

**What needs to happen:**
1. Add `Name string` field to `syntax.Node` (memory cost: ~16 bytes per node)
2. Populate `Name` in `transform.go` for `*ast.Ident` nodes (`n.Name`)
3. Use `Name` in `actionability.go` for real identifier matching
4. Consider string pool (`domain.StringPool`) if memory is tight
5. Update all Node construction in `syntax/templ/transform.go`

**Estimated effort:** 2-3 hours. The change is mechanical but touches the core data structure.

---

### #4 — `ClassifyClone` takes 4 primitives, so `Actionability` was added to the struct but never populated

**Status:** ⏳ NOT STARTED  
**Files to change:** `printer/clone_classify.go`, `printer/clone_processor.go`, `domain/processed_clone.go`

**Original claim:** "True, but mitigated" — `EvaluateActionability` IS called in `clone_processor.go:65`. The two-step design is correct. But `ClassifyClone`'s primitive signature is fragile.

**What needs to happen:**
1. Define `ClassificationInput` struct in `domain/` with: `Filename string`, `NodeType int32`, `Tokens int`, `Lines int`
2. Change `ClassifyClone(input ClassificationInput) CloneClassification`
3. Update `clone_processor.go:54` to construct `ClassificationInput` from node data
4. The compiler then catches missing fields when new fields are added

**Estimated effort:** 30 minutes. Small refactoring with clear scope.

---

## D) TOTALLY FUCKED UP 💥

Nothing totally fucked up. All changes compile and pass tests. However:

**Near-miss:** The `--semantic=true` default change almost broke all BDD tests because `validateMutualExclusion` was checking the flag VALUE (now `true`) instead of whether the flag was CHANGED. This is exactly issue #2 manifesting — the band-aid fix for structural was needed before the semantic default could flip. Caught and fixed in the same session.

**Technical debt note:** The `DiffMode` type uses `IsEnabled()` as its zero-value check in `isFieldZero`. The reflection-based merge handles it via the `default` case (`v.IsZero()`), which works because `DiffModeDisabled` (`""`) is the string zero value. This is correct but relies on implicit behavior. A struct tag approach (`merge:"skip"` or `merge:"always"`) would be more explicit.

---

## E) WHAT WE SHOULD IMPROVE

### Immediate (this session missed)

1. **Generalize `Changed()` tracking** — Extend the `SemanticSet`/`StructuralSet` pattern to ALL bool flags. Every `--include-*`, `--vendor`, `--profile`, `--incremental`, `--clear-cache` flag should track `Changed()` so `false` can override file config `true`.

2. **Add a round-trip merge test** — Test that adds a NEW field to Config and verifies it's picked up by the reflection-based merge without any code changes. This proves the reflection fix is future-proof.

3. **`config_builder.go` still has two code paths for bools** — `applyBooleanFlags` does the `if bf.flagValue { *bf.configPtr = true }` pattern AND `cfg.IncludeTempl = flags.IncludeTempl` unconditionally. This is confusing and should be unified with the `Changed()` approach.

### Structural (across sessions)

4. **Printer ↔ syntax.Node coupling** — `Printer.PrintClones(dups [][]*syntax.Node)` forces all 6 implementations to depend on AST internals. The `ProcessedClone` DTO exists but `ProcessClones` still takes `[][]*syntax.Node`. Fix: change the Printer interface to accept `[]ProcessedCloneGroup`.

5. **Three parallel Clone types** — `printer.clone` (unexported), `pkg/artdupl.Clone` (SDK), `domain.ProcessedClone`. Consolidation depends on Printer DTO change.

6. **`actionability.go` imports `syntax/golang` directly** — Language-specific constants mapped to categories. Breaks for non-Go languages. Should use the `Name` field from issue #3.

---

## F) Top #25 Things We Should Get Done Next

### P0 — Complete this session's work

| # | Task | Impact | Effort |
|---|------|--------|--------|
| 1 | Generalize `Changed()` tracking for all bool flags in `FlagValues` | High | 1h |
| 2 | Add `ClassificationInput` struct, change `ClassifyClone` signature (#4) | Medium | 30min |
| 3 | Add `Name string` to `syntax.Node`, populate in transform.go (#3) | High | 2h |
| 4 | Use `Name` in `actionability.go` for real identifier matching | High | 1h |
| 5 | Add round-trip merge test for reflection-based `mergeConfig` | Medium | 30min |

### P1 — Architecture cleanup

| # | Task | Impact | Effort |
|---|------|--------|--------|
| 6 | Change Printer interface to accept `[]ProcessedCloneGroup` instead of `[][]*syntax.Node` | High | 4h |
| 7 | Consolidate three parallel Clone types into unified domain types | High | 3h |
| 8 | Remove `printer.clone` unexported type, use `domain.ProcessedClone` everywhere | Medium | 2h |
| 9 | Extract `actionability.go` patterns to language-agnostic strategy | Medium | 2h |
| 10 | Remove `printer/clone_classify.go` direct import of `syntax/golang` | Medium | 1h |

### P2 — Config system hardening

| # | Task | Impact | Effort |
|---|------|--------|--------|
| 11 | Add struct tags for merge behavior (`merge:"skip"` / `merge:"always"`) | Low | 2h |
| 12 | Add exhaustive config merge property test (all fields, all combinations) | Medium | 1h |
| 13 | Add `--no-semantic` as alias for `--structural` (clearer UX) | Low | 15min |
| 14 | Deprecation warning when `--structural` is used (suggest it's the power-user mode) | Low | 15min |
| 15 | Config migration guide: document that `semantic` default changed from `false` to `true` | Medium | 30min |

### P3 — Testing and documentation

| # | Task | Impact | Effort |
|---|------|--------|--------|
| 16 | BDD test: `--structural=false` overrides `config.json: semantic: true` | Medium | 30min |
| 17 | BDD test: new Config field is automatically merged by reflection | Medium | 30min |
| 18 | Update README.md: `--semantic` is now default | Medium | 15min |
| 19 | Update HOW_TO_USE.md with new default behavior | Low | 30min |
| 20 | Update AGENTS.md with new default and reflection merge info | Medium | 15min |

### P4 — Performance and quality

| # | Task | Impact | Effort |
|---|------|--------|--------|
| 21 | Benchmark reflection-based merge vs old manual merge | Low | 30min |
| 22 | Consider `sync.Pool` for syntax.Node to amortize `Name` field allocation | Low | 1h |
| 23 | String pool integration for `syntax.Node.Name` field | Low | 1h |
| 24 | Profile memory impact of `Name string` on large codebases | Medium | 1h |
| 25 | Add `golangci-lint` `exhaustruct` check for `CloneClassification` | Low | 15min |

---

## G) Top #1 Question I Cannot Figure Out Myself

**Should the reflection-based `isFieldZero` use the same zero-value heuristics for ALL types, or should it use struct tags?**

Specifically: `DiffMode` currently checks `IsEnabled()` (which checks `!= "disabled" && IsValid()`). The reflection code uses `v.IsZero()` for the `default` case, which works because `DiffModeDisabled = ""` is the string zero value. But if someone changes `DiffModeDisabled` to `"off"`, the reflection code silently breaks.

Options:
1. **Keep as-is** — implicit correctness via string zero values. Works today, fragile tomorrow.
2. **Add `merge:"custom"` struct tag** — fields with this tag use a custom `IsSet() bool` method. Explicit but more code.
3. **Make all custom types implement `IsZero() bool`** — convention-based, Go-idiomatic. Requires adding methods to `DiffMode`, `OutputFormat`, `SortCriteria`, etc.

I lean toward option 3 because it's Go-idiomatic and catches edge cases at the type level, but I want your call before adding interface methods to 6 types.

---

## Test Results

All 23 packages pass, 0 failures:

```
ok  github.com/LarsArtmann/art-dupl/bdd          1.838s
ok  github.com/LarsArtmann/art-dupl/cmd           1.681s
ok  github.com/LarsArtmann/art-dupl/config        0.005s
ok  github.com/LarsArtmann/art-dupl/printer       0.011s
(all others cached/passing)
```

---

## Files Changed This Session

| File | Lines Changed | Description |
|------|--------------|-------------|
| `config/config.go` | +5/-5 | Semantic default true, updated docs |
| `config/config_merge.go` | +46/-174 | Reflection-based merge replacing 170-line manual field list |
| `config/config_test.go` | +2/-2 | Updated semantic default assertion |
| `cmd/flags.go` | +2/-2 | Semantic default true, structural help text |
| `cmd/config_builder.go` | +12/-6 | Changed() tracking for semantic/structural flags |
