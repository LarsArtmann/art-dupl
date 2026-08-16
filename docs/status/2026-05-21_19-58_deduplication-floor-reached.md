# art-dupl Status Report

**Date:** 2026-05-21 19:58
**Branch:** fork
**Version:** v0.1.0
**Reporter:** Crush (automated)

---

## Executive Summary

art-dupl has reached its **deduplication floor**. Deep analysis of all 61 clone groups at t=15 confirms zero actionable duplication remains. The 3 inherent clones at t=22 and 4 at t=20 are all structural Go patterns (identical function signatures, Error vs Fatal variants, same-type test data). No meaningful extraction is possible below t=22 — everything else is language-level boilerplate. 21/23 packages pass, 20 linter issues, clean working tree.

---

## Deduplication Journey This Session

| Threshold | Clone Groups | Status                                          |
| --------- | ------------ | ----------------------------------------------- |
| t=40      | 0            | All genuine duplication eliminated              |
| t=30      | 0            | All genuine duplication eliminated              |
| t=22      | 3            | **Floor reached** — all inherent                |
| t=20      | 4            | All inherent (one borderline drops below t=22)  |
| t=15      | 61           | Structural noise — language-level patterns only |

### Inherent Clones (Cannot Be Further Deduplicated)

| Location                                         | Pattern                                                          | Why Inherent                                                     |
| ------------------------------------------------ | ---------------------------------------------------------------- | ---------------------------------------------------------------- |
| `cmd/config_builder.go:15/56`                    | `BuildConfigFromFlags` / `buildCLIConfig` — identical signatures | Intentional API: public+private with same param types            |
| `internal/testutil/assert.go:13/20`              | `AssertLen` / `AssertFatalLen` — identical signatures            | Different semantics (Error vs Fatal), already shares `assertLen` |
| `printer/clone_classify_test.go:285-290/295-300` | Map literals with same keys, different values                    | Testing different accessors over same enum                       |
| `syntax/templ/templ_test.go:761-789/904-926`     | `parseTest` slices with similar structure                        | Different test data for different test functions                 |

### Analysis at t=15 (61 groups — All Inherent)

Categories of "clones" found at t=15 that are **not actionable**:

- **Standard Go error wrapping** — `fmt.Errorf("... (threshold: %d, sortBy: %s): %w", ...)` appears in cmd/run_output.go at 3 sites with different format strings and context. A helper would add complexity without reducing real duplication.
- **Test data** — Table-driven test cases, `[]DiffLine` literals, `parseTest` slices, BDD `It` blocks. Each has different values; the structural similarity IS the test pattern.
- **Single-line guard patterns** — `if !cmd.Flags().Changed("flag") { return nil }` in config_builder.go. Standard Cobra pattern, different flag names.
- **Field access** — `t.Run(name, func(t *testing.T)` in test files. Go test boilerplate.
- **Already-extracted helpers** — `diffWriteErr` call sites at lines 177/198. The helper exists; remaining similarity is just two calls with different labels.

---

## a) FULLY DONE

### Core Engine

- [x] Suffix tree detection (Ukkonen's, O(1) map transitions)
- [x] Hash-based detection (XXH3, ~20x faster than SHA-256)
- [x] Multi-detection mode (parallel goroutines)
- [x] Semantic-aware matching (FNV-1a, default ON)
- [x] SIMD-optimized hot paths

### Languages

- [x] Go AST (45 node types) + Templ (28 node types, 5 position bugs fixed)

### Output Formats (7)

- [x] Text, HTML, JSON, Simple-JSON, Plumbing, SARIF 2.1.0, CSV

### Statistics Subcommand

- [x] Text/JSON/CSV stats, Health grade (A-F), all metrics

### CLI & Configuration

- [x] Fang/Cobra CLI, shell completion, JSON config, reflection-based merge
- [x] Smart filtering: SQLC, Templ, Protobuf, Mockgen, Stringer, Generic
- [x] 4 sorting modes, `--all` batch generation

### Code Quality

- [x] **Deduplication floor reached** — 0 actionable clones at t=30, 3 inherent at t=22
- [x] 8 helpers extracted this session across 6 files
- [x] 214 Go files, ~46,155 lines

---

## b) PARTIALLY DONE

### Detection Methods (2 of 4 wired to CLI)

- [~] `TodoDetector` — implemented, not accessible via `-m` flag
- [~] `LegacyDetector` — implemented, not accessible via `-m` flag

### Linter Cleanliness

- 20 issues: err113 (2), errcheck (10), exhaustruct (2), goconst (5), gocyclo (1)

---

## c) NOT STARTED

1. **TokenValue type** — strong typing for token sequences (HIGH)
2. **ProcessedClone DTO** — decouple Printer from syntax.Node (111 test call sites)
3. **Clone type consolidation** — 3 parallel Clone types
4. **Proper CSV output** — use encoding/csv
5. **Enum unification** — domain enums → config generic helpers
6. **Memory layout optimization** — SIMD-friendly, string interning
7. **SIMD TODOs** — 6 items
8. **transform.go refactor** — 369L, 300L switch
9. **Old docs archival** — 300+ files in docs/status/
10. **SDK** — SDK_DESIGN.md exists, no implementation

---

## d) TOTALLY FUCKED UP

### 1. os.Exit Flaky Test (CRITICAL — UNCHANGED ALL SESSION)

**bdd/** and **cmd/** — all assertions pass, process exits non-zero due to Cobra's os.Exit.

### 2. Linter: 20 Issues (UNCHANGED ALL SESSION)

| Category    | Count | Worst Offender                                                      |
| ----------- | ----- | ------------------------------------------------------------------- |
| errcheck    | 10    | syscall.Dup2/Close in bdd/execute.go, cmd/stats_integration_test.go |
| goconst     | 5     | `--threshold` (4x), `art-dupl` (9x)                                 |
| err113      | 2     | dynamic errors in testutil/bdd_runners                              |
| exhaustruct | 2     | BDDTestSetup partial initialization                                 |
| gocyclo     | 1     | printer/stats_formatter.go:361 (16)                                 |

### 3. Printer ↔ syntax.Node Coupling (UNCHANGED)

6 printer implementations depend on AST internals. 111 test call sites.

### 4. Three Parallel Clone Types (UNCHANGED)

`printer.clone`, `pkg/artdupl.Clone`, `printer.CloneGroup`.

---

## e) WHAT WE SHOULD IMPROVE

### Immediate

1. **Fix os.Exit test flakiness** — has been #1 blocker ALL session, untouched
2. **Fix 10 errcheck issues** — unchecked syscall errors
3. **Reduce buildJSONData complexity** — gocyclo 16 → <10

### This Week

4. **Wire TODO/Legacy detectors to CLI** — implemented but inaccessible
5. **ProcessedClone DTO** — start decoupling printer from AST
6. **Proper CSV output**

### Next 2 Weeks

7. **Consolidate Clone types**
8. **TokenValue type**
9. **Enum unification**

---

## f) Top #25 Things We Should Get Done Next

| #  | Task                                                           | Priority |
| -- | -------------------------------------------------------------- | -------- |
| 1  | Fix os.Exit test flakiness (bdd + cmd)                         | CRITICAL |
| 2  | Fix 10 errcheck issues in test code                            | HIGH     |
| 3  | Extract buildJSONData into smaller functions (gocyclo 16→<10)  | HIGH     |
| 4  | Extract `--threshold` / `art-dupl` string constants in tests   | LOW      |
| 5  | Wire TodoDetector and LegacyDetector to CLI `-m` flag          | HIGH     |
| 6  | Fix 2 exhaustruct issues in testutil                           | LOW      |
| 7  | Fix 2 err113 issues in testutil                                | LOW      |
| 8  | Introduce ProcessedClone DTO                                   | HIGH     |
| 9  | Consolidate 3 parallel Clone types                             | MEDIUM   |
| 10 | Implement proper CSV output using encoding/csv                 | MEDIUM   |
| 11 | Implement TokenValue type with validation                      | HIGH     |
| 12 | Unify enum patterns                                            | MEDIUM   |
| 13 | Refactor transform.go (369L, 300L switch)                      | MEDIUM   |
| 14 | Optimize memory layouts for SIMD-friendly structures           | MEDIUM   |
| 15 | Implement string interning                                     | LOW      |
| 16 | Wire remaining SIMD TODOs (6 items)                            | LOW      |
| 17 | Fix ConstantCSSProperty position (upstream)                    | LOW      |
| 18 | Archive old docs/status/ files                                 | LOW      |
| 19 | Fix remaining LSP hints                                        | LOW      |
| 20 | Implement SDK from SDK_DESIGN.md                               | LOW      |
| 21 | Add more fuzz tests                                            | LOW      |
| 22 | Coverage improvement: domain (67%), job (77%), detection (78%) | MEDIUM   |
| 23 | Add benchmark regression CI                                    | MEDIUM   |
| 24 | Consider modularization                                        | LOW      |
| 25 | Migrate justfile → nix flake                                   | LOW      |

---

## g) Top #1 Question I Cannot Figure Out Myself

**What is the intended relationship between `bdd/` and `cmd/` integration tests?**

This has been the #1 question across ALL status reports this session. The overlap is real — both test end-to-end command execution. This session extracted `testutil.CopyToBuffer` and `testutil.appendFlagsToArgs` as shared helpers between them, but the fundamental duplication of test infrastructure remains. The answer to this question determines the os.Exit fix strategy.

---

## Session Summary (2026-05-21, Full Day)

| Time   | Milestone                                                |
| ------ | -------------------------------------------------------- |
| ~18:00 | Status report, 2 clone groups at t=40                    |
| ~18:10 | Extracted mustDeferSelectorCall, mustIfErrReturnNil      |
| ~18:15 | Zero at t=40. Committed 3 commits.                       |
| ~19:05 | 1 clone at t=30 in templ_test.go                         |
| ~19:10 | Extracted parseTest type. Zero at t=30.                  |
| ~19:15 | 9 clone groups at t=20                                   |
| ~19:40 | Eliminated 5 of 9. 4 inherent remain at t=20.            |
| ~19:55 | Analyzed all 61 groups at t=15 — zero actionable. Floor. |

### Helpers Extracted This Session (8 total)

| Helper                  | Location                         | Eliminated Groups |
| ----------------------- | -------------------------------- | ----------------- |
| `mustDeferSelectorCall` | printer/actionability_test.go    | 1 (t=40)          |
| `mustIfErrReturnNil`    | printer/actionability_test.go    | 1 (t=40)          |
| `parseTest` (type)      | syntax/templ/templ_test.go       | 1 (t=30)          |
| `mustWriteFile`         | config/config_enum_test.go       | 1 (t=20)          |
| `diffWriteErr`          | printer/html_diff.go             | 1 (t=20)          |
| `testutil.CopyToBuffer` | internal/testutil/bdd_helpers.go | 1 (t=20)          |
| `mustFuncDeclWithBody`  | printer/actionability_test.go    | 1 (t=20)          |
| `appendFlagsToArgs`     | internal/testutil/bdd.go         | 1 (t=20)          |

### Metrics (Unchanged Since Last Report)

| Metric                        | Value                        |
| ----------------------------- | ---------------------------- |
| Go files                      | 214                          |
| Total lines                   | 46,155                       |
| Passing packages              | 21/23 (91%)                  |
| Failing packages              | 2 (bdd, cmd — os.Exit flaky) |
| Linter issues                 | 20                           |
| Clone groups (t=22, semantic) | **3** (inherent floor)       |
| Clone groups (t=30, semantic) | **0**                        |

---

_Generated by Crush on 2026-05-21 at 19:58_
