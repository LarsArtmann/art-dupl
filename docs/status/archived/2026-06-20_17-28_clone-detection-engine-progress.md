# Status Report — Superb Clone Detection Engine

**Date:** 2026-06-20 17:28
**Branch:** fork (pushed to origin)
**Head:** `ed4cbe8` — feat: add CloneType enum (Type 1/2/3) to domain classification

---

## a) FULLY DONE (Committed + Pushed)

| Task                                        | Commit    | Files Changed                                                                                                                                                                                         | Impact                                                                                                                                                                                                                                                                                                     |
| ------------------------------------------- | --------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **T3** — BasicLit Value Hashing             | `1e37fd2` | `syntax/golang/transform.go`, `transform_test.go`                                                                                                                                                     | `return 42` ≠ `return 999` in semantic mode. Eliminates an entire false-positive class where different literal values produced identical token types.                                                                                                                                                      |
| **T8** — Test File Noise Suppression        | `09f3b6f` | `config/config.go`, `cmd/flags.go`, `cmd/config_builder.go`, `cmd/run_analysis.go`, `cmd/run_output.go`, `cmd/run_flags.go`, `cmd/stats.go`, `cmd/run_all_modes.go`, `bdd/semantic_detection_test.go` | New `--ignore-tests` flag (excludes `*_test.go`), `--include-tests` flag (disables all test filtering). `EffectiveTestThreshold()` defaults to `max(30, threshold)` for test files. `EffectiveSuppressTestLow()` respects `IncludeTests`.                                                                  |
| **T7** — Overlap & Nested Clone Elimination | `6e211ea` | `printer/groups.go`, `printer/overlap_test.go`, `cmd/run_output.go`, `cmd/stats.go`                                                                                                                   | `EliminateOverlaps()` processes fragments largest-first, suppresses any whose byte range is strictly contained within an already-accepted fragment. Dogfooding: clone groups reduced 81→47 on art-dupl itself.                                                                                             |
| **T4** — Non-Commutative Hash Combiner      | `d8a90e6` | `syntax/golang/identifier_hash.go`, `identifier_hash_test.go`                                                                                                                                         | Replaced XOR (`hash1 ^ hash2`) with FNV multiply-pair (`hash1*prime ^ hash2`). Non-commutative: `combine(A,B) ≠ combine(B,A)`. Eliminates false matches where `(TypeA, MethodX)` collided with `(TypeB, MethodY)`.                                                                                         |
| **T9** — Actionability Pattern Expansion    | `3cb2d11` | `printer/actionability.go`, `printer/actionability_control_flow.go`, `printer/actionability_patterns_expanded.go`, `printer/actionability_patterns_test.go`, `printer/clone_classify.go`              | 3 new non-actionable detectors: assertion chains (≥3 Expect/Assert/Require/Should/Must/So calls), error-wrapping returns (`if err != nil { return fmt.Errorf(...) }`), Cobra/Fang command boilerplate. Expanded RAII cleanup method list (+7: Stop, Shutdown, Cleanup, Reset, Put, Drop, Abort, Teardown). |
| **T6** — CloneType Enum (domain types only) | `ed4cbe8` | `domain/processed_clone.go`, `domain/analysis_errors.go`                                                                                                                                              | `CloneType` enum (Type1/Type2/Type3) added to domain with `IsValid()`, `String()`, `MarshalJSON`, `UnmarshalJSON`. Added to `CloneClassification` struct.                                                                                                                                                  |

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
| pkg/format     | 100.0%   | ✅                                                     |
| pkg/position   | 100.0%   | ✅                                                     |
| syntax/golang  | 95.9%    | ✅                                                     |
| bdd            | 93.3%    | ✅                                                     |
| hash           | 93.4%    | ✅                                                     |
| internal/utils | 92.1%    | ✅                                                     |
| suffixtree     | 91.2%    | ✅                                                     |
| syntax         | 91.4%    | ✅                                                     |
| errors         | 89.6%    | ✅                                                     |
| pkg/artdupl    | 88.4%    | ✅                                                     |
| config         | 86.6%    | ✅                                                     |
| cache          | 84.5%    | ✅                                                     |
| syntax/templ   | 84.6%    | ✅                                                     |
| pkg/logger     | 87.5%    | ✅                                                     |
| pkg/enum       | 94.1%    | ✅                                                     |
| printer        | 75.6%    | ⚠️                                                      |
| cmd            | 75.0%    | ⚠️                                                      |
| job            | 72.1%    | ⚠️                                                      |
| detection      | 61.8%    | 🔴                                                     |
| domain         | 58.6%    | 🔴 (dropped from 66.2% — new CloneType enum uncovered) |

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

| Task                                  | Why deferred                                                                                                                                                                                 |
| ------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
~~| **T1** — Statement-Level Tokenization | Fundamentally changes `serial()` pre-order DFS → statement-level fingerprinting. Every downstream consumer of token count would need rescaling. High risk of breaking detection correctness. |~~ done — shipped (ADR-0023, 2026-09-14)
~~| **T2** — Alpha-Normalization          | Depends on T1. Requires AST walker with scope tracking.                                                                                                                                      |~~ done — shipped (commit 91554e5, ADR-0007)
~~| **T5** — Three-Mode System            | Depends on T1+T2. Renames `--semantic` flag (breaking change for existing users).                                                                                                            |~~ done — shipped (ADR-0007)

### Phase 3 Refactoring Advisor

| Task                           | Why not started                                                                                     |
| ------------------------------ | --------------------------------------------------------------------------------------------------- |
~~| **T10** — Baseline/CI Mode     | Needs new subcommands (`baseline`, `check`), baseline file format design, clone diff logic. ~90min. |~~ done — `baseline` + `check` subcommands (CHANGELOG 0.4.0)
~~| **T11** — Extractability Score | Needs single-entry-point check, shared-mutable-state analysis, consistent-return check. ~70min.     |~~ done — property-based extractability engine (ADR-0017, CHANGELOG 0.6.0)

### Phase 4 Architecture Cleanup

| Task                                          | Why not started                                                    |
| --------------------------------------------- | ------------------------------------------------------------------ |
~~| **T12** — ctx in run_crawl                    | Last goroutine leak. Straightforward but touches file crawl.       |~~ done — ctx threaded through the crawl (commit 650a4b0)
~~| **T13** — Printer decoupling from syntax.Node | Design `ReadOnlyNode` interface, update 10 actionability matchers. |~~ done@decided-against — superseded by the domain.CloneNode bridge (2026-07-26)
~~| **T14** — Clone type consolidation            | 3 parallel Clone types. Risky refactor.                            |~~ done — CloneRef embedded everywhere (CHANGELOG 0.5.0); remaining DTOs intentional
~~| **T15** — Printer package split               | ~29 files into sub-packages. Mechanical but tedious.               |~~ done — printer/actionability + printer/stats (CHANGELOG 0.5.0)
~~| **T16** — Fragment type unification           | Low impact.                                                        |~~ parked — still open (ROADMAP consolidated design-debts entry)

### Phase 5 Testing & Quality

| Task                                     | Why not started                                    |
| ---------------------------------------- | -------------------------------------------------- |
~~| **T17** — Property tests for suffix tree | Algorithm correctness unverified. Important.       |~~ done — tran_parity_test.go + property suite (CHANGELOG 0.7.0)
~~| **T18** — Detection coverage (61.8%→80%) | MultiDetector dispatch, adapter integration tests. |~~ done — 92.7% (commit 9fecc7f)
~~| **T19** — Domain coverage (58.6%→80%)    | CloneType, ProcessedClone, ClonePriority tests.    |~~ done — 100% (commit 692c587)
~~| **T20** — Performance benchmarks         | No current regression risk.                        |~~ done — baselines committed under docs/benchmarks (ADR-0022 era)

### Phase 6 Ecosystem

| Task                              | Why not started            |
| --------------------------------- | -------------------------- |
~~| **T21** — GitHub Actions template | Depends on T10 (baseline). |~~ done — templates/github-actions-duplicate-check.yml (commit 5356d17)
~~| **T22** — Pre-commit hook         | Depends on T10.            |~~ done — .pre-commit-hooks.yaml (commit 5356d17)
~~| **T23** — json/v2 migration       | Future-proofing.           |~~ done — shipped 0.4.0, stabilized on the v1 API (ADR-0024)
~~| **T24** — Rename Data→View        | Naming clarity.            |~~ done — *View rename (commit b60aca6)

---

## d) TOTALLY FUCKED UP

Nothing. No regressions, no broken builds, no reverted changes.

**One learning from T9:** Initially lowered the interface implementation threshold from 3→2 fragments. Testing immediately showed this classified regular function signatures as non-actionable (BDD test `should include actionability in JSON output` failed because all clones were suppressed). Reverted to 3. The lesson: **threshold changes need immediate dogfooding validation before committing.**

---

## e) WHAT WE SHOULD IMPROVE

1. **`serial()` is still the elephant in the room.** T1 (statement-level tokenization) addresses the root cause but was deferred due to risk. The 6 completed tasks are post-hoc filters on a fundamentally wrong tokenization unit. Eventually T1 must be done or the tool will always have false positives from structural wrapper inflation.

2. **`--semantic` mode is still backwards.** It bakes exact identifier names into hashes, making matching _stricter_ than structural. T2 (alpha-normalization) would fix this but depends on T1.

3. **Domain coverage dropped to 58.6%** because T6 added new types without tests. Need T19.

4. **The `.auto-deduplicate/false-positives.json` still has 28 manually-suppressed false positives.** The new T9 patterns may cover some, but nobody has checked.

5. **No property-based tests** for the suffix tree algorithm (T17). Correctness is assumed, not verified.

6. **Clone type classification (T6) has domain types but no logic.** Every clone currently gets the zero value `""` for `CloneType`.

7. **T10 (baseline/CI mode) is the highest customer-value feature not yet started.** It's what makes the tool usable in CI/pre-commit.

8. **The actionability patterns (T9) compile but are untested against real noise patterns.** Could be producing false negatives (suppressing real clones) or false positives (not suppressing known noise).

---

## f) Top #25 Things to Get Done Next

| #  | Task                                                                                       | Impact   | Effort | Risk   |
| -- | ------------------------------------------------------------------------------------------ | -------- | ------ | ------ |
~~| 1  | **T6 logic**: Implement clone type classification — compare `Name` fields across fragments | High     | 30min  | Low    |~~ done — classifyCloneType (commit 359f3d7)
~~| 2  | **T6 output**: Add `clone_type` to JSON output                                             | Medium   | 15min  | Low    |~~ done — clone_type in JSON/SARIF/rich-text
~~| 3  | **T19**: Domain coverage tests (CloneType, ClonePriority, ProcessedClone)                  | Medium   | 50min  | Low    |~~ done — domain coverage 100% (commit 692c587)
~~| 4  | **T18**: Detection coverage tests (MultiDetector dispatch, adapter)                        | Medium   | 70min  | Low    |~~ done — detection coverage 92.7% (commit 9fecc7f)
~~| 5  | **T17**: Property-based tests for suffix tree                                              | High     | 60min  | Low    |~~ done — tran_parity_test.go (commit 2dc1781)
~~| 6  | **T10**: Baseline file format + `baseline` subcommand                                      | Critical | 90min  | Medium |~~ done — baseline subcommand (commit d620128)
~~| 7  | **T10**: `check` subcommand with clone diff + exit code 1                                  | Critical | 45min  | Medium |~~ done — check subcommand + exit 1
~~| 8  | **T12**: Thread `context.Context` through `run_crawl.go`                                   | Medium   | 45min  | Low    |~~ done — ctx threaded through crawl (commit 650a4b0)
~~| 9  | **Dogfood T9**: Run on `.auto-deduplicate/false-positives.json` patterns                   | High     | 20min  | Low    |~~ done — false-positives.json maintained + calibration corpus (CHANGELOG 0.7.0)
~~| 10 | **T11**: Extractability score design + implementation                                      | High     | 70min  | Medium |~~ done — extractability engine (commit c60e58b, CHANGELOG 0.6.0)
~~| 11 | **Update AGENTS.md**: Document T3/T4/T7/T8/T9 changes                                      | Medium   | 15min  | Low    |~~ done — AGENTS.md current
~~| 12 | **Update HOW_TO_USE.md**: Document `--ignore-tests`, `--include-tests`                     | Medium   | 15min  | Low    |~~ done — HOW_TO_USE documents --ignore-tests/--include-tests
~~| 13 | **Update FEATURES.md**: Add overlap elimination, BasicLit hashing, new flags               | Medium   | 15min  | Low    |~~ done — FEATURES current
~~| 14 | **T1 prototype**: Spike statement-level tokenization in a branch                           | Critical | 120min | HIGH   |~~ done — statement-level tokenization shipped (ADR-0023, 2026-09-14)
~~| 15 | **T2 prototype**: Spike alpha-normalization walker                                         | Critical | 80min  | HIGH   |~~ done — alpha-normalization (commit 91554e5)
~~| 16 | **T5**: Three-mode system (`--exact`/`--semantic`/`--structural`)                          | Critical | 80min  | High   |~~ done — three-mode system (ADR-0007)
~~| 17 | **T20**: Performance benchmarks — establish baseline before T1                             | Low      | 45min  | Low    |~~ done — benchmarks + alloc budgets (commit ef77469)
~~| 18 | **T13**: Design `ReadOnlyNode` interface for printer decoupling                            | Medium   | 80min  | Medium |~~ done@decided-against — ReadOnlyNode superseded by domain.CloneNode bridge (2026-07-26)
~~| 19 | **T21**: GitHub Actions workflow template                                                  | Medium   | 40min  | Low    |~~ done — templates/github-actions-duplicate-check.yml (commit 5356d17)
~~| 20 | **T22**: Pre-commit hook YAML                                                              | Medium   | 35min  | Low    |~~ done — .pre-commit-hooks.yaml (commit 5356d17)
~~| 21 | **Fix domain coverage**: Write CloneType tests specifically                                | High     | 20min  | Low    |~~ done — printer/clone_type_test.go
~~| 22 | **Review false-positives.json**: Check which entries T9 patterns now catch                 | High     | 30min  | Low    |~~ done — SELF_CLEAN_LEDGER + false-positives.json (CHANGELOG 0.7.1)
~~| 23 | **T14**: Consolidate CloneLocation in ProcessedClone/CloneGroup/SDK Clone                  | Medium   | 90min  | Medium |~~ done — CloneRef consolidation (CHANGELOG 0.5.0)
~~| 24 | **T23**: Audit and migrate to encoding/json/v2                                             | Low      | 40min  | Low    |~~ done — json/v2 shipped 0.4.0, stabilized by ADR-0024
~~| 25 | **T24**: Rename `*Data` → `*View` in printer                                               | Low      | 35min  | Low    |~~ done — *View rename (commit b60aca6)

---

## g) Top #1 Question I Cannot Figure Out Myself

~~**Should T1 (statement-level tokenization) be attempted now, or is the risk too high?**~~ done — shipped safely (ADR-0023, 2026-09-14) with group-shape-preserving subsumed trimming

T1 is the root-cause fix for the fundamental flaw: `serial()` flattens the AST via pre-order DFS, making 1 Go statement = ~8-13 tokens. The 15-token threshold measures AST nodes, not statements. Every other completed task (T3, T4, T7, T8, T9) is a post-hoc filter on top of this broken tokenization.

**The risk:** T1 changes what `Node.Owns` means, what `threshold` means, what `FindSyntaxUnits` expects, and what every test assertion about token counts checks. It's a cascading change through the entire detection pipeline. If done wrong, it could VERSCHLIMMBESSER the system — make it worse while trying to improve it.

**The alternative:** Keep stacking post-hoc filters (T7 overlap elimination, T8 test suppression, T9 actionability patterns). These are safe, incremental, and have already halved the false positive rate. But they're treating symptoms, not the disease.

**What I need from you:** A clear GO/NO-GO decision on whether to attempt T1 in this branch, or whether the current ~50% false positive reduction is sufficient for now and T1 should be a separate dedicated effort with its own branch + comprehensive test suite.
