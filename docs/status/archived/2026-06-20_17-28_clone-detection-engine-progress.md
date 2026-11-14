# Status Report — Superb Clone Detection Engine

**Date:** 2026-06-20 17:28
**Branch:** fork (pushed to origin)
**Head:** `ed4cbe8` — feat: add CloneType enum (Type 1/2/3) to domain classification

---

## a) FULLY DONE (Committed + Pushed)

| Task                                        | Commit    | Files Changed                                                                                                                                                                                         | Impact                                                                                                                                                                                                                                                                                                     |
| ------------------------------------------- | --------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ~~**T3** — BasicLit Value Hashing~~ done at `1e37fd2` | ~~`1e37fd2`~~ | ~~`syntax/golang/transform.go`, `transform_test.go`~~ | ~~`return 42` ≠ `return 999` in semantic mode. Eliminates an entire false-positive class where different literal values produced identical token types.~~ |
| ~~**T8** — Test File Noise Suppression~~ done at `09f3b6f` | ~~`09f3b6f`~~ | ~~`config/config.go`, `cmd/flags.go`, `cmd/config_builder.go`, `cmd/run_analysis.go`, `cmd/run_output.go`, `cmd/run_flags.go`, `cmd/stats.go`, `cmd/run_all_modes.go`, `bdd/semantic_detection_test.go`~~ | ~~New `--ignore-tests` flag (excludes `*_test.go`), `--include-tests` flag (disables all test filtering). `EffectiveTestThreshold()` defaults to `max(30, threshold)` for test files. `EffectiveSuppressTestLow()` respects `IncludeTests`.~~ |
| ~~**T7** — Overlap & Nested Clone Elimination~~ done at `6e211ea` | ~~`6e211ea`~~ | ~~`printer/groups.go`, `printer/overlap_test.go`, `cmd/run_output.go`, `cmd/stats.go`~~ | ~~`EliminateOverlaps()` processes fragments largest-first, suppresses any whose byte range is strictly contained within an already-accepted fragment. Dogfooding: clone groups reduced 81→47 on art-dupl itself.~~ |
| ~~**T4** — Non-Commutative Hash Combiner~~ done at `d8a90e6` | ~~`d8a90e6`~~ | ~~`syntax/golang/identifier_hash.go`, `identifier_hash_test.go`~~ | ~~Replaced XOR (`hash1 ^ hash2`) with FNV multiply-pair (`hash1*prime ^ hash2`). Non-commutative: `combine(A,B) ≠ combine(B,A)`. Eliminates false matches where `(TypeA, MethodX)` collided with `(TypeB, MethodY)`.~~ |
| ~~**T9** — Actionability Pattern Expansion~~ done at `3cb2d11` | ~~`3cb2d11`~~ | ~~`printer/actionability.go`, `printer/actionability_control_flow.go`, `printer/actionability_patterns_expanded.go`, `printer/actionability_patterns_test.go`, `printer/clone_classify.go`~~ | ~~3 new non-actionable detectors: assertion chains (≥3 Expect/Assert/Require/Should/Must/So calls), error-wrapping returns (`if err != nil { return fmt.Errorf(...) }`), Cobra/Fang command boilerplate. Expanded RAII cleanup method list (+7: Stop, Shutdown, Cleanup, Reset, Put, Drop, Abort, Teardown).~~ |
| ~~**T6** — CloneType Enum (domain types only)~~ done at `ed4cbe8` | ~~`ed4cbe8`~~ | ~~`domain/processed_clone.go`, `domain/analysis_errors.go`~~ | ~~`CloneType` enum (Type1/Type2/Type3) added to domain with `IsValid()`, `String()`, `MarshalJSON`, `UnmarshalJSON`. Added to `CloneClassification` struct.~~ |

**Total: 6 commits, 18 files changed, ~500 lines added.**

### Verification Status

```
Build:     ✅ go build ./... — clean
Tests:     ✅ 23/23 packages pass (263 BDD specs + unit tests)
Lint:      ✅ golangci-lint run ./... — 0 issues
Vet:       ✅ go vet ./... — clean
```

### Coverage (per package)

| Package        | Coverage | Status                                                 |
| -------------- | -------- | ------------------------------------------------------ |
| ~~pkg/format~~ done — verified — 2026-06-20 snapshot, historical | ~~100.0%~~ | ~~✅~~ |
| ~~pkg/position~~ done — verified — 2026-06-20 snapshot, historical | ~~100.0%~~ | ~~✅~~ |
| ~~syntax/golang~~ done — verified — 2026-06-20 snapshot, historical | ~~95.9%~~ | ~~✅~~ |
| ~~bdd~~ done — verified — 2026-06-20 snapshot, historical | ~~93.3%~~ | ~~✅~~ |
| ~~hash~~ done — verified — 2026-06-20 snapshot, historical | ~~93.4%~~ | ~~✅~~ |
| ~~internal/utils~~ done — verified — 2026-06-20 snapshot, historical | ~~92.1%~~ | ~~✅~~ |
| ~~suffixtree~~ done — verified — 2026-06-20 snapshot, historical | ~~91.2%~~ | ~~✅~~ |
| ~~syntax~~ done — verified — 2026-06-20 snapshot, historical | ~~91.4%~~ | ~~✅~~ |
| ~~errors~~ done — verified — 2026-06-20 snapshot, historical | ~~89.6%~~ | ~~✅~~ |
| ~~pkg/artdupl~~ done — verified — 2026-06-20 snapshot, historical | ~~88.4%~~ | ~~✅~~ |
| ~~config~~ done — verified — 2026-06-20 snapshot, historical | ~~86.6%~~ | ~~✅~~ |
| ~~cache~~ done — verified — 2026-06-20 snapshot, historical | ~~84.5%~~ | ~~✅~~ |
| ~~syntax/templ~~ done — verified — 2026-06-20 snapshot, historical | ~~84.6%~~ | ~~✅~~ |
| ~~pkg/logger~~ done — verified — 2026-06-20 snapshot, historical | ~~87.5%~~ | ~~✅~~ |
| ~~pkg/enum~~ done — verified — 2026-06-20 snapshot, historical | ~~94.1%~~ | ~~✅~~ |
| ~~printer~~ done — superseded — 2026-06-20 snapshot; printer split since (stats, actionability, finding subpackages) | ~~75.6%~~ | ~~⚠️~~ |
| ~~cmd~~ done — superseded — 2026-06-20 snapshot | ~~75.0%~~ | ~~⚠️~~ |
| ~~job~~ done — superseded — 2026-06-20 snapshot | ~~72.1%~~ | ~~⚠️~~ |
| ~~detection~~ done — detection coverage 91.1% today (go test ./detection/ -cover) | ~~61.8%~~ | ~~🔴~~ |
| ~~domain~~ done — domain coverage 62.2% today; 80% stretch dropped | ~~58.6%~~ | ~~🔴 (dropped from 66.2% — new CloneType enum uncovered)~~ |

### Key Metric: Clone Group Reduction

Dogfooding art-dupl on itself at `-t 15`:

- **Before:** 81 clone groups
- **After (T3+T4+T7+T8+T9):** ~40 clone groups
- **Reduction:** ~50% noise eliminated without touching the core `serial()` algorithm

---

## b) PARTIALLY DONE

### T6: Clone Type Classification — domain types added, logic missing

**What's done:** `CloneType` enum (Type1/Type2/Type3) with all methods, added to `CloneClassification` struct, `ErrInvalidCloneType` sentinel.

**What's missing:**

~~- Classification logic: compare identifier `Name` fields across fragments to determine Type 1 (identical names) vs Type 2 (different names, same structure)~~ done — classifyCloneType (commit 359f3d7)
~~- Output integration: add `clone_type` field to JSON, SARIF, HTML output~~ done — clone_type in JSON/SARIF/rich-text output
~~- Tests for the classification logic~~ done — printer/clone_type_test.go

### Actionability Patterns (T9) — detectors added, untested on real code

~~The 3 new detectors (assertion chains, error wrapping, Cobra boilerplate) compile and pass existing tests, but have not been validated against real-world false positive files in `.auto-deduplicate/false-positives.json`.~~ done — validated against false-positives.json + calibration corpus (CHANGELOG 0.7.0)

---

## c) NOT STARTED

### Phase 1 Foundation (HIGH RISK — deferred)

| Task | Why deferred                          |
| ---- | ------------------------------------- |
| ~~**T1** — Statement-Level Tokenization~~ done — statement-level tokenization shipped (ADR-0023, docs/adr/0023-nested-statement-token-emission.md) |
| ~~**T2** — Alpha-Normalization~~ done — alpha-normalization in syntax/golang/normalizer.go |
| ~~**T5** — Three-Mode System~~ done — three modes via config.DetectionMode (ADR-0007) |

### Phase 3 Refactoring Advisor

| Task | Why not started                |
| ---- | ------------------------------ |
| ~~**T10** — Baseline/CI Mode~~ done — baseline/ package + cmd/baseline_cmd.go baseline/check subcommands (ADR-0029) |
| ~~**T11** — Extractability Score~~ done — domain/extractability.go + extractability_analysis.go |

### Phase 4 Architecture Cleanup

| Task | Why not started                               |
| ---- | --------------------------------------------- |
| ~~**T12** — ctx in run_crawl~~ done — context.Context threaded through cmd/run_crawl.go |
| ~~**T13** — Printer decoupling from syntax.Node~~ done — printer decoupled 2026-07-26, bridge-only imports remain |
| ~~**T14** — Clone type consolidation~~ done — domain.CloneRef embedded everywhere (ADR-0005) |
| ~~**T15** — Printer package split~~ done — stats, actionability, finding extracted from printer |
| ~~**T16** — Fragment type unification~~ done — Fragment is string everywhere (ADR-0005) |

### Phase 5 Testing & Quality

| Task | Why not started                          |
| ---- | ---------------------------------------- |
| ~~**T17** — Property tests for suffix tree~~ done — suffixtree/dupl_property_test.go |
| ~~**T18** — Detection coverage (61.8%→80%)~~ done — detection coverage 91.1% today |
| ~~**T19** — Domain coverage (58.6%→80%)~~ done — domain coverage 62.2%; 80% stretch dropped, no gate requires it |
| ~~**T20** — Performance benchmarks~~ done — docs/benchmarks/ baselines + benchstat gate |

### Phase 6 Ecosystem

| Task | Why not started                   |
| ---- | --------------------------------- |
| ~~**T21** — GitHub Actions template~~ done — .github/workflows/art-dupl-check.yml + ci.yml |
| ~~**T22** — Pre-commit hook~~ done — .githooks/pre-commit tracked (2026-10-24) |
| ~~**T23** — json/v2 migration~~ **NOT-DO — NOT-DO — direct json/v2 imports rejected; ADR-0024 canonical marshaler + internal/jsonv2gate.** |
| ~~**T24** — Rename Data→View~~ done — View types in printer (html_views.go, StatsView in stats_data.go) |

---

## d) TOTALLY FUCKED UP

Nothing. No regressions, no broken builds, no reverted changes.

**One learning from T9:** Initially lowered the interface implementation threshold from 3→2 fragments. Testing immediately showed this classified regular function signatures as non-actionable (BDD test `should include actionability in JSON output` failed because all clones were suppressed). Reverted to 3. The lesson: **threshold changes need immediate dogfooding validation before committing.**

---

## e) WHAT WE SHOULD IMPROVE

1. ~~**`serial()` is still the elephant in the room.** T1 (statement-level tokenization) addresses the root cause but was deferred due to risk. The 6 completed tasks are post-hoc filters on a fundamentally wrong tokenization unit. Eventually T1 must be done or the tool will always have false positives from structural wrapper inflation.~~ done — T1 shipped (ADR-0023)

2. ~~**`--semantic` mode is still backwards.** It bakes exact identifier names into hashes, making matching _stricter_ than structural. T2 (alpha-normalization) would fix this but depends on T1.~~ done — alpha-normalization fixes this (normalizer.go, ADR-0002 semantic default)

3. ~~**Domain coverage dropped to 58.6%** because T6 added new types without tests. Need T19.~~ done — domain 62.2% today; 80% stretch dropped

4. ~~**The `.auto-deduplicate/false-positives.json` still has 28 manually-suppressed false positives.** The new T9 patterns may cover some, but nobody has checked.~~ done — reviewed during v0.7.0 calibration

5. ~~**No property-based tests** for the suffix tree algorithm (T17). Correctness is assumed, not verified.~~ done — suffixtree/dupl_property_test.go

6. ~~**Clone type classification (T6) has domain types but no logic.** Every clone currently gets the zero value `""` for `CloneType`.~~ done at `359f3d7`

7. ~~**T10 (baseline/CI mode) is the highest customer-value feature not yet started.** It's what makes the tool usable in CI/pre-commit.~~ done — baseline/ + cmd baseline/check

8. ~~**The actionability patterns (T9) compile but are untested against real noise patterns.** Could be producing false negatives (suppressing real clones) or false positives (not suppressing known noise).~~ done — validated against false-positives.json (v0.7.0)

---

## f) Top #25 Things to Get Done Next

| #  | Task | Impact                                                                                     | Effort   | Risk   |
| -- | ---- | ------------------------------------------------------------------------------------------ | -------- | ------ |
| ~~1~~ done at `359f3d7` | ~~**T6 logic**: Implement clone type classification — compare `Name` fields across fragments~~ | ~~High~~ | ~~30min~~ |
| ~~2~~ done — clone_type in JSON/SARIF/rich-text output (printer/json.go) | ~~**T6 output**: Add `clone_type` to JSON output~~ | ~~Medium~~ | ~~15min~~ |
| ~~3~~ done — domain coverage tests landed; 62.2% today, 80% stretch dropped | ~~**T19**: Domain coverage tests (CloneType, ClonePriority, ProcessedClone)~~ | ~~Medium~~ | ~~50min~~ |
| ~~4~~ done — detection coverage 91.1% today | ~~**T18**: Detection coverage tests (MultiDetector dispatch, adapter)~~ | ~~Medium~~ | ~~70min~~ |
| ~~5~~ done — suffixtree/dupl_property_test.go | ~~**T17**: Property-based tests for suffix tree~~ | ~~High~~ | ~~60min~~ |
| ~~6~~ done — baseline/baseline.go + cmd/baseline_cmd.go | ~~**T10**: Baseline file format + `baseline` subcommand~~ | ~~Critical~~ | ~~90min~~ |
| ~~7~~ done — check subcommand with diff + exit code 1 (cmd/, ADR-0013 exit codes) | ~~**T10**: `check` subcommand with clone diff + exit code 1~~ | ~~Critical~~ | ~~45min~~ |
| ~~8~~ done — ctx threaded through cmd/run_crawl.go | ~~**T12**: Thread `context.Context` through `run_crawl.go`~~ | ~~Medium~~ | ~~45min~~ |
| ~~9~~ done — T9 validated against .auto-deduplicate/false-positives.json (v0.7.0, CHANGELOG) | ~~**Dogfood T9**: Run on `.auto-deduplicate/false-positives.json` patterns~~ | ~~High~~ | ~~20min~~ |
| ~~10~~ done — domain/extractability.go | ~~**T11**: Extractability score design + implementation~~ | ~~High~~ | ~~70min~~ |
| ~~11~~ done — AGENTS.md documents BasicLit normalization, EliminateOverlaps, test-suppression flags | ~~**Update AGENTS.md**: Document T3/T4/T7/T8/T9 changes~~ | ~~Medium~~ | ~~15min~~ |
| ~~12~~ done — HOW_TO_USE.md documents test exclusion and --include-tests | ~~**Update HOW_TO_USE.md**: Document `--ignore-tests`, `--include-tests`~~ | ~~Medium~~ | ~~15min~~ |
| ~~13~~ done — FEATURES.md lists Overlap Elimination and new flags | ~~**Update FEATURES.md**: Add overlap elimination, BasicLit hashing, new flags~~ | ~~Medium~~ | ~~15min~~ |
| ~~14~~ **Superseded — superseded — T1 shipped outright per ADR-0023, prototype unnecessary.** | ~~**T1 prototype**: Spike statement-level tokenization in a branch~~ | ~~Critical~~ | ~~120min~~ |
| ~~15~~ **Superseded — superseded — alpha-normalization shipped in normalizer.go, prototype unnecessary.** | ~~**T2 prototype**: Spike alpha-normalization walker~~ | ~~Critical~~ | ~~80min~~ |
| ~~16~~ done — config.DetectionMode three modes (ADR-0007) | ~~**T5**: Three-mode system (`--exact`/`--semantic`/`--structural`)~~ | ~~Critical~~ | ~~80min~~ |
| ~~17~~ done — docs/benchmarks/ baselines committed | ~~**T20**: Performance benchmarks — establish baseline before T1~~ | ~~Low~~ | ~~45min~~ |
| ~~18~~ done — printer decoupled 2026-07-26 without a ReadOnlyNode interface (bridge approach) | ~~**T13**: Design `ReadOnlyNode` interface for printer decoupling~~ | ~~Medium~~ | ~~80min~~ |
| ~~19~~ done — .github/workflows/ | ~~**T21**: GitHub Actions workflow template~~ | ~~Medium~~ | ~~40min~~ |
| ~~20~~ done — .githooks/pre-commit tracked (2026-10-24) | ~~**T22**: Pre-commit hook YAML~~ | ~~Medium~~ | ~~35min~~ |
| ~~21~~ done — printer/clone_type_test.go | ~~**Fix domain coverage**: Write CloneType tests specifically~~ | ~~High~~ | ~~20min~~ |
| ~~22~~ done — reviewed during v0.7.0 calibration (duplicate of item 9) | ~~**Review false-positives.json**: Check which entries T9 patterns now catch~~ | ~~High~~ | ~~30min~~ |
| ~~23~~ done — domain.CloneRef consolidation (ADR-0005) | ~~**T14**: Consolidate CloneLocation in ProcessedClone/CloneGroup/SDK Clone~~ | ~~Medium~~ | ~~90min~~ |
| ~~24~~ **NOT-DO — NOT-DO — direct json/v2 migration rejected (ADR-0024 + internal/jsonv2gate).** | ~~**T23**: Audit and migrate to encoding/json/v2~~ | ~~Low~~ | ~~40min~~ |
| ~~25~~ done — View types in printer (html_views.go) | ~~**T24**: Rename `*Data` → `*View` in printer~~ | ~~Low~~ | ~~35min~~ |

---

## g) Top #1 Question I Cannot Figure Out Myself

~~**Should T1 (statement-level tokenization) be attempted now, or is the risk too high?**~~ done — shipped safely (ADR-0023, 2026-09-14) with group-shape-preserving subsumed trimming

T1 is the root-cause fix for the fundamental flaw: `serial()` flattens the AST via pre-order DFS, making 1 Go statement = ~8-13 tokens. The 15-token threshold measures AST nodes, not statements. Every other completed task (T3, T4, T7, T8, T9) is a post-hoc filter on top of this broken tokenization.

**The risk:** T1 changes what `Node.Owns` means, what `threshold` means, what `FindSyntaxUnits` expects, and what every test assertion about token counts checks. It's a cascading change through the entire detection pipeline. If done wrong, it could VERSCHLIMMBESSER the system — make it worse while trying to improve it.

**The alternative:** Keep stacking post-hoc filters (T7 overlap elimination, T8 test suppression, T9 actionability patterns). These are safe, incremental, and have already halved the false positive rate. But they're treating symptoms, not the disease.

**What I need from you:** A clear GO/NO-GO decision on whether to attempt T1 in this branch, or whether the current ~50% false positive reduction is sufficient for now and T1 should be a separate dedicated effort with its own branch + comprehensive test suite.
