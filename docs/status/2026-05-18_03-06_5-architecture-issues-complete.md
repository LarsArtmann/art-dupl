# Status Report: 5 Architecture Debt Issues — FULLY RESOLVED

**Date:** 2026-05-18 03:06  
**Session:** Complete resolution of 5 critical architecture issues  
**Branch:** fork  
**Commits:** 6 new, already pushed to origin/fork  
**Working tree:** clean  

---

## Executive Summary

All 5 "why this is stupid" issues identified in the previous session have been fully resolved. The code compiles, all 22 test packages pass with zero failures, and 6 commits have been pushed. This session represents a fundamental improvement in config reliability, CLI correctness, AST semantics, and clone classification robustness.

---

## A) FULLY DONE ✅ — All 5 Issues Resolved

### #1 — mergeConfig was a manually-maintained 170-line field list

**Status:** ✅ RESOLVED  
**Commit:** `1ea1a7b`

- Replaced 170-line `//nolint:funlen,gocognit,gocyclo,cyclop` manual field list with 30-line reflection-based `mergeConfig`
- Adding a new field to `Config` no longer requires touching merge code
- `isFieldZero()` handles bool/int/string/slice with explicit zero detection
- Net code reduction: -110 lines of fragile code
- All existing tests pass unchanged

### #2 — skipZeroValues on bool fields made false flags silently lose to file config

**Status:** ✅ RESOLVED (full generalization, not just semantic/structural)  
**Commit:** `d437f16`

- Eliminated `FlagValues` struct (28-field manual mirror of `Config`)—adding a Config field no longer requires updating `FlagValues`, `extractFlagValues()`, `applyBooleanFlags()`, and `applyThresholdFlag()`
- Replaced `applyBooleanFlags` (only propagated `true`) with `applyChangedBoolFlags` (uses `cmd.Flags().Changed()` for all 13 bool flags)
- `--vendor=false`, `--include-sqlc=false`, etc. now correctly override file config values
- Error handling preserved: `applyDetectionMethods`, `applyTimeoutFlag`, `applyDiffModeFlag` now return errors up the call chain instead of being silently discarded with `_`

### #3 — transform.go threw away Go identifier names; syntax.Node was semantically blind

**Status:** ✅ RESOLVED  
**Commit:** `66cb17b`

- Added `Name string` to `syntax.Node` struct (56B per node, was 40B)
- Populated `Name` in `syntax/golang/transform.go` for:
  - `*ast.Ident`: `Name = n.Name` (e.g., "err", "nil", "mu")
  - `*ast.SelectorExpr`: `Name = n.Sel.Name` (e.g., "Unlock", "Close")
  - `*ast.FuncDecl`: `Name = funcName` (e.g., "HandleRequest")
  - `*ast.TypeSpec`: `Name = n.Name.Name` (e.g., "Server")
- Gob backward compatibility preserved: new `Name` field gets zero value ("") when decoding old cache entries
- Improved `containsNilIdentifier`: now checks `child.Name == "nil"` instead of just `child.Type == golang.Ident`
- Improved `isPureDeferPattern`:
  - Old: ALL bare `DeferStmt` marked as non-actionable (too broad)
  - New: Only RAII cleanup (Unlock, Close, Done, Cancel, Release, Finish, Disconnect, Free) marked non-actionable
  - Business-logic defer (e.g., `processOrder()`) now correctly detected as actionable
- New test case: "defer processOrder is actionable (not RAII)" proves the improvement

### #4 — ClassifyClone took 4 primitives, so Actionability was silently unpopulated

**Status:** ✅ RESOLVED  
**Commit:** `87902c3`

- Replaced `ClassifyClone(filename string, nodeType int32, tokens, lines int)` with `ClassifyClone(input domain.ClassificationInput)`
- New `domain.ClassificationInput` struct with `Filename`, `NodeType`, `Tokens`, `Lines` fields
- Compiler now catches missing fields when new classification inputs are added
- Updated caller in `clone_processor.go:54` and test in `clone_classify_test.go:202`

### #5 — --structural was the default, producing 68% garbage on first use

**Status:** ✅ RESOLVED  
**Commit:** `1ea1a7b` + `f425695`

- `DefaultConfig().Semantic` changed from `false` to `true`
- `--semantic` flag default changed from `false` to `true`
- `--structural` help text rewritten: "disable semantic detection; increases false positives"
- Updated 6 docs files (README, HOW_TO_USE, FEATURES, AGENTS.md) that still claimed structural was default
- Dogfooding showed 244→78 clone groups (68% noise reduction) on go-cqrs-lite
- BDD tests updated for new default behavior
- `validateMutualExclusion` now uses `Changed()` tracking (only errors when both flags explicitly set)

---

## B) PARTIALLY DONE 🔧

Nothing partially done. All 5 issues reached completion.

---

## C) NOT STARTED ⏳

Large architectural refactorings that were identified as related but not required for the 5 issues:

1. **Printer ↔ syntax.Node coupling** — `Printer.PrintClones(dups [][]*syntax.Node)` forces 6 implementations to depend on AST internals. Fix requires introducing `ProcessedCloneGroup` DTO to the Printer interface (touches 111 test call sites).
2. **Three parallel Clone types** — `printer.clone`, `pkg/artdupl.Clone`, `domain.ProcessedClone`. Consolidation depends on Printer DTO change.
3. **actionability.go still imports syntax/golang** — Language-specific constants mapped to categories. True language independence requires replacing node type checks with category tags or a registry pattern.

---

## D) TOTALLY FUCKED UP 💥

Nothing.

**Near-miss logged:** When `1ea1a7b` changed `--semantic` default from `false` to `true`, the `validateMutualExclusion` check (based on flag VALUES) would have triggered for every `--structural` invocation because the default `--semantic=true` and `--structural=true` would both be set. Fixed in `1ea1a7b` with `Changed()` tracking.

---

## E) WHAT WE SHOULD IMPROVE

### Immediate (next session)

1. **DetectionMode in `job/` still passed as raw `bool`** — `config.Semantic bool` threaded through 7 functions before becoming `DetectionMode`. Should use `DetectionConfig` (already exists in `domain/`) throughout the pipeline.
2. **Add fuzz/property test for reflection merge** — Verify that adding a NEW field to Config is automatically picked up by `mergeConfig` without code changes. This is the core property we gained from reflection, worth testing.
3. **String pool for `syntax.Node.Name`** — Every Ident clone creates identical `Name` strings (e.g., "err", "nil", "Unlock"). A `sync.Map` pool would reduce allocations significantly for large codebases.

### Medium-term

4. **Printer DTO** — Change `Printer` interface to accept `[]domain.ProcessedCloneGroup` instead of `[][]*syntax.Node`. The `ProcessClones` function already exists — just change the interface contract.
5. **Remove `printer.clone` unexported type** — `printer/clone.go` duplicates `domain.ProcessedClone`. Unify them.
6. **Language-agnostic actionability** — Replace `syntax/golang` imports in `actionability.go` with a node category registry that maps node types → categories at parse time (e.g., a `NodeCategory` field on `syntax.Node`).

---

## F) Top #25 Things to Do Next

### P0 — Critical (do next session)

| # | Task | Impact | Effort |
|---|------|--------|--------|
| 1 | Pass DetectionConfig through job/ pipeline instead of bare bool | High | 2h |
| 2 | Property test: reflection merge automatically picks up new Config fields | Medium | 1h |
| 3 | Add string pool for syntax.Node.Name (reduce allocations) | Medium | 1h |

### P1 — Architecture

| # | Task | Impact | Effort |
|---|------|--------|--------|
| 4 | Printer interface → accept ProcessedCloneGroup[] | High | 4h |
| 5 | Consolidate three Clone types | High | 3h |
| 6 | Remove printer.clone unexported type | Medium | 2h |
| 7 | NodeCategory field on syntax.Node for language-agnostic actionability | Medium | 2h |
| 8 | Remove actionability.go import of syntax/golang | Medium | 1h |
| 9 | Benchmark Name field memory impact on large project | Low | 30min |
| 10 | Add exhaustruct linter check back with Config whitelist | Low | 15min |

### P2 — Quality

| # | Task | Impact | Effort |
|---|------|--------|--------|
| 11 | BDD test: --structural=false overrides config.json semantic: true | Medium | 30min |
| 12 | BDD test: --vendor=false overrides file vendor: true | Low | 30min |
| 13 | Fuzz test for MergeConfigs with all field combinations | Medium | 1h |
| 14 | Add --no-semantic alias for --structural (clearer UX) | Low | 15min |
| 15 | Migrate from gob cache to custom binary format (faster, smaller) | Medium | 3h |

### P3 — Polish

| # | Task | Impact | Effort |
|---|------|--------|--------|
| 16 | Update CHANGELOG.md with semantic default breaking change | Medium | 30min |
| 17 | MIGRATION_GUIDE.md section: v1→v2 semantic default change | Medium | 30min |
| 18 | README screenshot/asciinema with new default behavior | Low | 30min |
| 19 | Deprecation warning for --structural (suggest it's power-user mode) | Low | 15min |
| 20 | Config file template: semantic: true for new users | Low | 15min |

### P4 — Performance

| # | Task | Impact | Effort |
|---|------|--------|--------|
| 21 | Profile: where is Name string allocation heaviest? | Medium | 30min |
| 22 | Consider arena allocation for syntax.Node in transform | Low | 2h |
| 23 | SIMD-optimized string comparison for Name matching | Low | 2h |
| 24 | Parallel transform.go (per-file goroutines) | Medium | 2h |
| 25 | Memory-mapped AST cache for incremental mode | Low | 4h |

---

## G) Top #1 Question I Cannot Figure Out Myself

**How do we cleanly migrate `job/` from `bool` to `DetectionConfig` without breaking the incremental parser cache format?**

The current chain is:
```
config.Config.Semantic (bool)
  → job.Parse() → job.ParseParallel() → job.ParseFileByExtensionWithConfig() → bool
    → job/file_parser.go:35 — bool → DetectionMode conversion
```

The `IncrementalParser` caches `[]*syntax.Node` via gob. The gob-encoded data doesn't know about `DetectionConfig`. But the pipeline from config → parser still uses `bool`, meaning:

1. If we change `job.Parse(semantic bool)` to `job.Parse(cfg DetectionConfig)`, we break every caller and every cached signature.
2. If we only change the conversion point (`file_parser.go:35`), we solve 80% of the problem with 20% of the churn.
3. OR — we accept that `bool` is the right abstraction at the job level (it only cares about semantic vs structural), and the typed `DetectionConfig` belongs at the `detection/` package boundary.

I lean toward option 2 (change only `file_parser.go:35` to accept `DetectionConfig` and convert there), but I want your call: is the bare `bool` parameter through `job/` acceptable as a "primitive at the boundary" pattern, or should we push the typed config all the way through?

---

## Test Results

All 22 test packages pass:

```
ok  github.com/LarsArtmann/art-dupl/bdd           1.585s
ok  github.com/LarsArtmann/art-dupl/cmd            1.413s
ok  github.com/LarsArtmann/art-dupl/config         0.005s
ok  github.com/LarsArtmann/art-dupl/detection      0.012s
ok  github.com/LarsArtmann/art-dupl/domain         0.004s
ok  github.com/LarsArtmann/art-dupl/errors         0.005s
ok  github.com/LarsArtmann/art-dupl/hash           0.011s
ok  github.com/LarsArtmann/art-dupl/job            0.288s
ok  github.com/LarsArtmann/art-dupl/printer        0.009s
ok  github.com/LarsArtmann/art-dupl/syntax/golang  0.010s
(all others cached/passing)
```

---

## Commit History

| Hash | Files | Description |
|------|-------|-------------|
| `66cb17b` | 4 | feat(syntax): Name field + actionability improvement |
| `87902c3` | 4 | refactor(printer): ClassificationInput struct |
| `d437f16` | 1 | refactor(cmd): Eliminate FlagValues split brain |
| `89e62b2` | 1 | fix(config): Complete DefaultConfig |
| `f425695` | 4 | docs: Update docs for semantic default |
| `1ea1a7b` | 6 | fix(config): Semantic default + reflection merge |
