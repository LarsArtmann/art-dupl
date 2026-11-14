~~# Status Report — Superb Clone Detection Engine (Final Sprint)~~ resolved by alternative — GO decision taken: T1 shipped (ADR-0023) + threshold recalibrated (ADR-0009)

**Date:** 2026-06-20 19:56
**Branch:** fork (pushed to origin)
**Head:** `502a63c`
**Plan:** `docs/planning/2026-06-20_16-05_SUPERB-CLONE-DETECTION-ENGINE.md`

---

## a) FULLY DONE (20/24 tasks, committed + pushed)

### Phase 1: Foundation (4/5)

| Task                          | Commit    | What changed                                                                |
| ----------------------------- | --------- | --------------------------------------------------------------------------- |
| ~~**T3** BasicLit Value Hashing~~ done at `1e37fd2` | ~~`1e37fd2`~~ | ~~`return 42` ≠ `return 999` in semantic mode~~ |
| ~~**T4** Non-Commutative Hash~~ done at `d8a90e6` | ~~`d8a90e6`~~ | ~~FNV multiply-pair replaces XOR; `(A,B)≠(B,A)`~~ |
| ~~**T2** Alpha-Normalization~~ done at `91554e5` | ~~`91554e5`~~ | ~~Per-function symbol table canonicalizes locals → **Type 2 clone detection**~~ |
| ~~**T5** Three-Mode System~~ done at `91554e5` | ~~`91554e5`~~ | ~~`--semantic` (default), `--exact`, `--structural`; bool→DetectionMode enum~~ |

### Phase 2: Detection Quality (3/3)

| Task                               | Commit    | What changed                                          |
| ---------------------------------- | --------- | ----------------------------------------------------- |
| ~~**T6** Clone Type Classification~~ done at `359f3d7` | ~~`359f3d7`~~ | ~~type-1/2/3 labels in JSON, SARIF, rich-text~~ |
| ~~**T7** Overlap Elimination~~ done at `6e211ea` | ~~`6e211ea`~~ | ~~Nested clones suppressed (81→~40 groups)~~ |
| ~~**T8** Test-File Noise Suppression~~ done at `09f3b6f` | ~~`09f3b6f`~~ | ~~`--ignore-tests`, `--include-tests`, smarter defaults~~ |

### Phase 3: Refactoring Advisor (3/3)

| Task                          | Commit    | What changed                                                  |
| ----------------------------- | --------- | ------------------------------------------------------------- |
| ~~**T9** Actionability Patterns~~ done at `3cb2d11` | ~~`3cb2d11`~~ | ~~Assertion chains, error-wrapping, Cobra boilerplate detectors~~ |
| ~~**T10** Baseline/CI Mode~~ done at `d620128` | ~~`d620128`~~ | ~~`baseline` + `check` subcommands, exit 1 on new clones~~ |
| ~~**T11** Extractability Score~~ done at `c60e58b` | ~~`c60e58b`~~ | ~~`lines_saved` + `extractable` in JSON output~~ |

### Phase 4: Architecture (1/5)

| Task                     | Commit    | What changed                   |
| ------------------------ | --------- | ------------------------------ |
| ~~**T12** ctx in run_crawl~~ done at `650a4b0` | ~~`650a4b0`~~ | ~~Last goroutine leak eliminated~~ |

### Phase 5: Testing (4/4)

| Task                       | Commit    | What changed                                     |
| -------------------------- | --------- | ------------------------------------------------ |
| ~~**T17** Property Tests~~ done at `2dc1781` | ~~`2dc1781`~~ | ~~6 suffix-tree invariants verified~~ |
| ~~**T18** Detection Coverage~~ done at `9fecc7f` | ~~`9fecc7f`~~ | ~~61.8% → **92.7%**~~ |
| ~~**T19** Domain Coverage~~ done at `692c587` | ~~`692c587`~~ | ~~58.6% → **100%**~~ |
| ~~**T20** Benchmarks~~ done at `ef77469` | ~~`ef77469`~~ | ~~STreeUpdate + FindDuplOver baselines established~~ |

### Phase 6: Ecosystem (3/4)

| Task                     | Commit    | What changed                           |
| ------------------------ | --------- | -------------------------------------- |
| ~~**T21** GitHub Actions~~ done at `5356d17` | ~~`5356d17`~~ | ~~`.github/workflows/art-dupl-check.yml`~~ |
| ~~**T22** Pre-Commit Hook~~ done at `5356d17` | ~~`5356d17`~~ | ~~`.pre-commit-hooks.yaml`~~ |
| ~~**T24** Rename Data→View~~ done at `b60aca6` | ~~`b60aca6`~~ | ~~5 `*Data` types renamed to `*View`~~ |

### Documentation & Cleanup

| Commit    | What changed                                                                    |
| --------- | ------------------------------------------------------------------------------- |
| ~~`b86e7de`~~ done at `b86e7de` | ~~Reuse `syntax.Serialize` for clone-type walk; update AGENTS/HOW_TO_USE/FEATURES~~ |
| ~~`f7f73af`~~ done at `f7f73af` | ~~Status report for alpha-normalization + three-mode~~ |
| ~~`83dc554`~~ done at `83dc554` | ~~Comprehensive sprint status report~~ |
| ~~`d27b727`~~ done at `d27b727` | ~~AGENTS.md architecture updated with baseline package~~ |
| ~~`502a63c`~~ done at `502a63c` | ~~HOW_TO_USE/FEATURES/README updated with baseline/check, extractability, modes~~ |

---

## b) PARTIALLY DONE

Nothing. All 20 completed tasks are fully done: implementation, tests, lint, docs.

---

## c) NOT STARTED (4 tasks — deferred with rationale)

| Task                                        | Why deferred                                                                                                                                                                     |
| ------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ~~**T1** Statement-Level Tokenization~~ done — T1 shipped (ADR-0023) | ~~HIGH RISK: replaces `serial()` core algorithm. Cascading change through entire pipeline. Could destabilize detection. Needs dedicated branch. Property tests (T17) now guard it.~~ |
| ~~**T13** Printer Decoupling from syntax.Node~~ done — printer decoupled 2026-07-26, bridge-only imports remain | ~~Medium effort. 13 production files import `syntax.Node`. Pure architecture — no user-facing impact.~~ |
| ~~**T14** Clone Type Consolidation~~ done — CloneRef consolidation (ADR-0005) | ~~3 parallel Clone types (domain/printer/SDK). Risky refactor touching all layers.~~ |
| ~~**T15/T16** Printer Split / Fragment Unify~~ done — printer split (stats, actionability, finding) + Fragment string everywhere | ~~Low impact, mechanical. Fragment `[]byte`↔`string` boundary conversion is minimal.~~ |
| ~~**T23** json/v2 Migration~~ **NOT-DO — NOT-DO — direct json/v2 imports rejected (ADR-0024 + internal/jsonv2gate).** | ~~`encoding/json/v2` exists in Go 1.26.3 but is experimental/unstable. Audited, deferred until stabilization.~~ |

---

## d) TOTALLY FUCKED UP

Nothing. No regressions, no broken builds, no reverted changes.

**Learnings applied:**

1. **T9 threshold change** (from prior session): lowering interface threshold 3→2 caused false negatives. Lesson: threshold changes need immediate dogfooding. Applied throughout.
2. **classifyCloneType initial bug**: compared only top-level fragment nodes, not descendants. Fixed by reusing `syntax.Serialize` for full subtree walk.
3. **Maximality property test**: initial test logic was inverted (`!canExtendLeft` vs `canExtendLeft`). Fixed before commit.

---

## e) WHAT WE SHOULD IMPROVE

1. ~~**`serial()` is still the elephant in the room.** T1 addresses it but was deferred. The 6 completed filter tasks (T3/T4/T7/T8/T9) are post-hoc filters on top of node-level tokenization. Eventually T1 must be done or thresholds will always be inflated.~~ done — T1 shipped: nested emission (ADR-0023), CacheVersion 4

2. ~~**Normalizer is heuristic** — flat per-function symbol table (no nested-scope shadowing resolution). `go/types` would give precise scope resolution but adds type-checking overhead.~~ parked — ROADMAP "Nested-scope shadowing in alpha-normalization"

3. ~~**3 parallel Clone types** — `domain.ProcessedClone`, `printer.CloneGroup`, `pkg/artdupl.Clone`. Consolidation (T14) is pending.~~ resolved by alternative — CloneRef embedded across types; format DTOs deliberate

4. ~~**No BDD test for Type 2 detection** — the headline feature. The unit tests in `normalizer_test.go` and `clone_type_test.go` verify the logic, but there's no end-to-end BDD scenario.~~ done — bdd/semantic_detection_test.go end-to-end semantic scenarios

5. ~~**printer/ imports syntax.Node in 13 production files** — coupling that makes printer untestable in isolation.~~ done — printer decoupled 2026-07-26

---

## f) Top #25 Things to Get Done Next

<!-- check-rows:ledger — item 25 (init subcommand) is genuinely open; all other rows carry verdicts -->
| #  | Task                                                            | Impact                                                              | Effort   | Risk  |
| -- | --------------------------------------------------------------- | ------------------------------------------------------------------- | -------- | ----- |
| ~~1~~ done — T1 shipped (ADR-0023) | ~~**T1 Statement Tokenization** (feature branch)~~ | ~~Critical~~ | ~~90min~~ |
| ~~2~~ done — bdd/semantic_detection_test.go + bdd/generic_method_test.go renamed-clone scenario | ~~**BDD test: Type 2 clone detection**~~ | ~~High~~ | ~~30min~~ |
| ~~3~~ done — decoupled 2026-07-26 without a ReadOnlyNode interface (bridge approach) | ~~**T13 Printer Decoupling** (ReadOnlyNode interface)~~ | ~~Medium~~ | ~~80min~~ |
| ~~4~~ done — domain.CloneRef consolidation (ADR-0005) | ~~**T14 Clone Consolidation** (CloneLocation shared type)~~ | ~~Medium~~ | ~~90min~~ |
| ~~5~~ **Superseded — superseded — nested-scope shadowing tracked in ROADMAP.md; go/types upgrade not adopted.** | ~~**go/types normalizer upgrade**~~ | ~~Medium~~ | ~~60min~~ |
| ~~6~~ done — stats, actionability, finding extracted | ~~**T15 Printer Split** (stats/html/analyze sub-packages)~~ | ~~Low-Med~~ | ~~80min~~ |
| ~~7~~ done — Fragment is string everywhere | ~~**T16 Fragment Unify** ([]byte vs string)~~ | ~~Low~~ | ~~40min~~ |
| ~~8~~ **NOT-DO — NOT-DO — ADR-0024 + internal/jsonv2gate.** | ~~**T23 json/v2** (when stable)~~ | ~~Low~~ | ~~40min~~ |
| ~~9~~ done — validated during v0.7.0 calibration | ~~**Dogfood T9** against real noise patterns~~ | ~~High~~ | ~~20min~~ |
| ~~10~~ done — CI creates and checks it (.github/workflows/art-dupl-check.yml) | ~~**`.art-dupl-baseline.json` for art-dupl itself**~~ | ~~Medium~~ | ~~10min~~ |
| ~~11~~ **Superseded — superseded — coverage snapshot targets drift; printer 77.9% today, no 80% gate.** | ~~**Coverage: printer 75.8%→80%**~~ | ~~Medium~~ | ~~50min~~ |
| ~~12~~ **Superseded — superseded — coverage snapshot targets drift; baseline 71.6% today, no 95% gate.** | ~~**Coverage: baseline 88.2%→95%**~~ | ~~Low~~ | ~~20min~~ | ~~Low~~ |
| ~~13~~ done — docs/benchmarks/ baselines + benchstat CI gate | ~~**End-to-end benchmark** (parse→detect→report)~~ | ~~Medium~~ | ~~30min~~ |
| ~~14~~ done — diff visualization flag (cmd/flags.go --diff) + cmd/diff_report.go | ~~**`check --diff` flag** (show what changed since baseline)~~ | ~~Medium~~ | ~~40min~~ |
| ~~15~~ **Superseded — superseded — HTML carries classification badges; clone_type ships in JSON/SARIF/rich-text (printer/json.go).** | ~~**HTML report: clone type badge**~~ | ~~Low~~ | ~~20min~~ |
| ~~16~~ **Superseded — superseded — extractability surfaces in JSON extractable/extractability fields, no HTML column.** | ~~**HTML report: extractability column**~~ | ~~Low~~ | ~~20min~~ | ~~Low~~ |
| ~~17~~ **Superseded — superseded — SDK keeps the Semantic bool, mapped in pkg/artdupl/detector_utils.go.** | ~~**SDK: expose DetectionMode** (not just Semantic bool)~~ | ~~Medium~~ | ~~30min~~ | ~~Low~~ |
| ~~18~~ **NOT-DO — NOT-DO — single --mode alias not adopted; --semantic/--exact/--structural are the flags (cmd/config_builder.go).** | ~~**Config: `--mode` flag** (alias for semantic/exact/structural)~~ | ~~Low~~ | ~~15min~~ | ~~Low~~ |
| ~~19~~ **Superseded — superseded — re-running art-dupl baseline updates the file (cmd/baseline_cmd.go).** | ~~**Baseline: `--update` flag** (add new clones to existing baseline)~~ | ~~Medium~~ | ~~30min~~ |
| ~~20~~ **Superseded — superseded — surfaces as go-finding Confidence via printer/finding, not a SARIF property.** | ~~**SARIF: add extractability to properties**~~ | ~~Low~~ | ~~10min~~ | ~~Low~~ |
| ~~21~~ done — templ alpha-normalization for Type 2 (syntax/templ) | ~~**Templ semantic mode** (alpha-normalization for .templ)~~ | ~~Medium~~ | ~~60min~~ |
| ~~22~~ **Superseded — superseded — no diff subcommand; diff reporting ships via root flag path (cmd/run_flags.go, cmd/diff_report.go).** | ~~**`art-dupl diff` subcommand** (compare two baselines)~~ | ~~Low~~ | ~~30min~~ |
| ~~23~~ done — bdd/semantic_performance_bench_test.go | ~~**Performance: profile normalizer overhead**~~ | ~~Medium~~ | ~~20min~~ | ~~Low~~ |
| ~~24~~ done — README badges (CI, License, Go Reference, Website) | ~~**README: add badges (CI, coverage, Go version)**~~ | ~~Low~~ | ~~10min~~ |
| 25 | **`art-dupl init` subcommand** (create config + baseline)       | Low                                                                 | 20min    | Low   |

---

## g) Top #1 Question I Cannot Figure Out Myself

**Should T1 (statement-level tokenization) be attempted now, or is the risk too high?**

T1 is the root-cause fix for the fundamental flaw: `serial()` flattens the AST via pre-order DFS, making 1 Go statement = ~8-13 tokens. The threshold of 15 measures AST nodes, not statements. Every completed filter task (T3/T4/T7/T8/T9/T2) is a post-hoc mitigation on top of this node-level tokenization.

**The risk:** T1 changes what `Node.Owns` means, what `threshold` means, what `FindSyntaxUnits` expects, and what every test assertion about token counts checks. It's a cascading change. Property tests (T17) now guard correctness, and benchmarks (T20) establish the performance baseline — but the change could still VERSCHLIMMBESSER the system.

**The alternative:** The alpha-normalization (T2) already delivers the most valuable improvement (Type 2 detection). The remaining false positives from wrapper inflation are filtered by T7 (overlap elimination) and T8 (test suppression). The system is usable now.

**What I need:** A GO/NO-GO decision on whether T1 should be attempted in a feature branch, or whether the current ~50% false positive reduction + Type 2 detection is sufficient.

---

## Verification

```
Build:     ✅ go build ./... — clean (24 packages)
Tests:     ✅ 24/24 packages pass
Lint:      ✅ golangci-lint run ./... — 0 issues
Coverage:  domain 100%, detection 92.7%, suffixtree 91.2%, syntax/golang 95.7%
Remote:    ✅ origin/fork up to date (head 502a63c)
```
