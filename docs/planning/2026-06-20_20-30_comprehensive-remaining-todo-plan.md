# Comprehensive Remaining TODO Plan — All Tasks, ≤12min Each

**Date:** 2026-06-20 20:30
**Status:** PLANNING — awaiting execution
**Sources:** Sprint status report Top-25, deferred tasks (T1/T13/T14/T15/T16/T23), TODO_LIST.md, verified build/lint state
**Verified:** `go build ./...` ✅ | `golangci-lint run ./...` 0 issues ✅

---

## Sorting Methodology

Each parent task scored on four axes, then ranked:

| Axis        | Values                                          |
| ----------- | ----------------------------------------------- |
| **Impact**  | Critical → High → Medium → Low                  |
| **Effort**  | Total minutes (sum of sub-tasks)                |
| **Risk**    | Low → Medium → High                             |
| **CustVal** | High (user-facing/CI) → Medium → Low (internal) |

**Priority tiers (highest value first):**

| Tier  | Meaning                      | Tasks                                |
| ----- | ---------------------------- | ------------------------------------ |
| **S** | Quick wins proving value     | BDD Type 2, dogfood, baseline, docs  |
| **A** | High value, low effort       | SARIF, HTML badges, --mode, coverage |
| **B** | Medium features              | --diff, --update, SDK mode, init     |
| **C** | Architecture cleanup         | Printer decouple, clone unify, split |
| **D** | Risky / advanced             | T1 tokenization, go/types, templ sem |
| **E** | Deferred (external blockers) | json/v2, facade, hybrid transitions  |

---

## ⛔ BLOCKER: T1 GO/NO-GO Decision Required

**T1 (Statement-Level Tokenization)** is ranked #1 by impact but is **HIGH RISK**. It replaces the core `serial()` algorithm — a cascading change through the entire pipeline. The user has been asked for a GO/NO-GO decision and has **NOT yet answered**.

**Do not start T1 sub-tasks without explicit approval.**

All other tasks can proceed independently.

---

## Complete Task Table — 28 Parent Tasks / 100 Sub-Tasks

### TIER S — Immediate Wins (prove value, fix stale docs)

| #    | Sub-Task                                                          | Impact | Effort | Risk | Deps | CustVal |
| ---- | ----------------------------------------------------------------- | ------ | ------ | ---- | ---- | ------- |
| S1.1 | Create test fixture: two .go files, renamed funcs, identical body | High   | 10min  | Low  | —    | High    |
| S1.2 | BDD scenario: "semantic mode detects Type 2 clones" in bdd/       | High   | 12min  | Low  | S1.1 | High    |
| S1.3 | Assert clone_type=Type2; verify exact mode does NOT match         | High   | 8min   | Low  | S1.2 | High    |
| S2.1 | Fix stale TODO_LIST.md: mark T12, T24 done; update remaining list | Medium | 12min  | Low  | —    | Medium  |
| S3.1 | Run art-dupl on itself at -t15/-t30; catalog remaining false pos  | High   | 10min  | Low  | —    | High    |
| S3.2 | If gaps found, add actionability pattern detectors; re-verify     | High   | 10min  | Low  | S3.1 | High    |
| S4.1 | Run `art-dupl baseline` on art-dupl source; review + commit file  | Medium | 10min  | Low  | —    | High    |

**Tier S total: 72min (7 sub-tasks)**

### TIER A — High Value, Low Effort

| #     | Sub-Task                                                         | Impact | Effort | Risk | Deps  | CustVal |
| ----- | ---------------------------------------------------------------- | ------ | ------ | ---- | ----- | ------- |
| A5.1  | Add `lines_saved` + `extractable` to SARIF Properties map        | Low    | 10min  | Low  | —     | Medium  |
| A6.1  | Add clone-type badge CSS + render in HTML template               | Medium | 12min  | Low  | —     | Medium  |
| A6.2  | Verify HTML shows type-1/type-2/type-3 badges                    | Medium | 8min   | Low  | A6.1  | Medium  |
| A7.1  | Add extractability column to HTML clone table                    | Medium | 12min  | Low  | —     | Medium  |
| A7.2  | Verify column renders with real output                           | Medium | 8min   | Low  | A7.1  | Medium  |
| A8.1  | Add CI/coverage/Go-version badges to README.md header            | Low    | 10min  | Low  | —     | Medium  |
| A9.1  | Add `--mode {exact,semantic,structural}` flag + mutual-exclusion | Medium | 12min  | Low  | —     | High    |
| A10.1 | Add baseline edge-case tests: empty/corrupt/missing/dup hashes   | Medium | 12min  | Low  | —     | Low     |
| A10.2 | Add Add/Has tests with many entries; verify ≥95% coverage        | Low    | 8min   | Low  | A10.1 | Low     |

**Tier A total: 92min (9 sub-tasks)**

### TIER B — Medium Features

| #     | Sub-Task                                                       | Impact | Effort | Risk | Deps  | CustVal |
| ----- | -------------------------------------------------------------- | ------ | ------ | ---- | ----- | ------- |
| B11.1 | Implement `check --diff`: show added/removed clones vs base    | Medium | 12min  | Low  | —     | High    |
| B11.2 | Format diff: green=removed, red=new; tabular output            | Medium | 12min  | Low  | B11.1 | High    |
| B11.3 | Tests: add clone→check shows new; remove→shows fixed           | Medium | 12min  | Low  | B11.2 | Medium  |
| B12.1 | Implement `baseline --update`: merge new clones into file      | Medium | 12min  | Low  | —     | High    |
| B12.2 | Tests: record→add clone→update→verify in baseline              | Medium | 12min  | Low  | B12.1 | Medium  |
| B13.1 | Add DetectionMode to SDK Options (replace Semantic bool)       | Medium | 12min  | Low  | —     | High    |
| B13.2 | Update ValidateOptions + SDK docs/examples for mode            | Medium | 12min  | Low  | B13.1 | Medium  |
| B14.1 | Implement `art-dupl init`: create baseline + default config    | Low    | 12min  | Low  | —     | Medium  |
| B14.2 | Tests + document init in HOW_TO_USE.md                         | Low    | 8min   | Low  | B14.1 | Medium  |
| B15.1 | Implement `art-dupl diff <base1> <base2>`: compare two files   | Low    | 12min  | Low  | —     | Medium  |
| B15.2 | Format: only-in-1, only-in-2, common; tabular output           | Low    | 12min  | Low  | B15.1 | Medium  |
| B16.1 | Run `go test -bench -cpuprofile` semantic vs structural        | Medium | 10min  | Low  | —     | Low     |
| B16.2 | Analyze profile; document normalizer overhead; find hotspots   | Medium | 10min  | Low  | B16.1 | Low     |
| B17.1 | Create benchmark corpus + end-to-end parse→detect→report bench | Medium | 12min  | Low  | —     | Low     |
| B17.2 | Record baseline numbers in docs/baselines/                     | Medium | 12min  | Low  | B17.1 | Low     |
| B18.1 | Identify untested printer funcs via coverage profile           | Medium | 10min  | Low  | —     | Low     |
| B18.2 | Add clone_processor tests (classifyCloneType edge cases)       | Medium | 12min  | Low  | B18.1 | Low     |
| B18.3 | Add JSON/SARIF output edge-case tests (empty, nil fields)      | Medium | 12min  | Low  | B18.1 | Low     |
| B18.4 | Add text/HTML rendering path tests; verify ≥80% coverage       | Medium | 12min  | Low  | B18.2 | Low     |

**Tier B total: 198min (19 sub-tasks)**

### TIER C — Architecture Cleanup

| #     | Sub-Task                                                       | Impact | Effort | Risk   | Deps  | CustVal |
| ----- | -------------------------------------------------------------- | ------ | ------ | ------ | ----- | ------- |
| C19.1 | Design `ReadOnlyNode` interface in printer/                    | Medium | 12min  | Low    | —     | Low     |
| C19.2 | Update actionability\*.go matchers → accept ReadOnlyNode       | Medium | 12min  | Medium | C19.1 | Low     |
| C19.3 | Update clone_processor.go bridge: \*syntax.Node → ReadOnlyNode | Medium | 12min  | Medium | C19.2 | Low     |
| C19.4 | Remove direct syntax.Node imports from printer/actionability\* | Medium | 12min  | Medium | C19.3 | Low     |
| C19.5 | Update go-arch-lint: printer/ may not depend on syntax/        | Low    | 10min  | Low    | C19.4 | Low     |
| C19.6 | Verify all printer tests pass; grep syntax.Node (target: 0)    | Medium | 10min  | Low    | C19.4 | Low     |
| C20.1 | Decide canonical Fragment type ([]byte vs string); document    | Low    | 10min  | Low    | —     | Low     |
| C20.2 | Update domain Fragment to canonical type                       | Low    | 10min  | Low    | C20.1 | Low     |
| C20.3 | Update SDK Fragment to canonical type                          | Low    | 10min  | Low    | C20.1 | Low     |
| C20.4 | Remove boundary conversion code; verify tests                  | Low    | 10min  | Low    | C20.2 | Low     |
| C21.1 | Design CloneLocation value type (Filename, LineStart, etc.)    | Medium | 12min  | Low    | —     | Low     |
| C21.2 | Embed CloneLocation in domain.ProcessedClone                   | Medium | 12min  | Medium | C21.1 | Low     |
| C21.3 | Embed CloneLocation in printer.CloneGroup                      | Medium | 12min  | Medium | C21.1 | Low     |
| C21.4 | Embed CloneLocation in pkg/artdupl.Clone                       | Medium | 12min  | Medium | C21.1 | Low     |
| C21.5 | Update all JSON serialization paths for embedded type          | Medium | 12min  | Medium | C21.2 | Low     |
| C21.6 | Remove duplicate fields; verify SDK independence (0 imports)   | Medium | 12min  | Medium | C21.5 | Low     |
| C21.7 | Full test pass + lint clean                                    | Medium | 10min  | Low    | C21.6 | Low     |
| C22.1 | Create printer/stats/ sub-package                              | Low    | 10min  | Low    | C19   | Low     |
| C22.2 | Move stats\*.go to printer/stats/                              | Low    | 12min  | Low    | C22.1 | Low     |
| C22.3 | Create printer/html/ sub-package                               | Low    | 10min  | Low    | C19   | Low     |
| C22.4 | Move html*.go, diff*.go to printer/html/                       | Low    | 12min  | Low    | C22.3 | Low     |
| C22.5 | Create printer/analyze/ sub-package                            | Low    | 10min  | Low    | C19   | Low     |
| C22.6 | Move actionability*.go, clone\_*.go to printer/analyze/        | Low    | 12min  | Low    | C22.5 | Low     |
| C22.7 | Update all imports + go-arch-lint; verify build                | Low    | 12min  | Low    | C22.2 | Low     |

**Tier C total: 262min (24 sub-tasks)**

### TIER D — Risky / Advanced (requires care)

| #     | Sub-Task                                                        | Impact       | Effort | Risk   | Deps     | CustVal |
| ----- | --------------------------------------------------------------- | ------------ | ------ | ------ | -------- | ------- |
| D23.1 | ⛔ Design StatementFingerprint type: int32 token per statement  | **Critical** | 12min  | High   | GO/NO-GO | High    |
| D23.2 | ⛔ Implement statement boundary detection (walk BlockStmt.List) | **Critical** | 12min  | High   | D23.1    | High    |
| D23.3 | ⛔ Replace serial() with statement-level serialization          | **Critical** | 12min  | High   | D23.2    | High    |
| D23.4 | ⛔ Update FindSyntaxUnits for statement-level tokens            | **Critical** | 12min  | High   | D23.3    | High    |
| D23.5 | ⛔ Update threshold semantics + document new defaults           | **Critical** | 10min  | High   | D23.4    | High    |
| D23.6 | ⛔ Update all token-count assertions in existing tests          | **Critical** | 12min  | High   | D23.5    | Medium  |
| D23.7 | ⛔ Dogfood on art-dupl; verify no VERSCHLIMMBESSERUNG           | **Critical** | 10min  | High   | D23.6    | High    |
| D24.1 | Research go/types scope API; prototype on sample function       | Medium       | 12min  | Medium | —        | Low     |
| D24.2 | Replace flat symbol table with go/types scope walker            | Medium       | 12min  | Medium | D24.1    | Low     |
| D24.3 | Handle nested-scope variable shadowing                          | Medium       | 12min  | Medium | D24.2    | Low     |
| D24.4 | Update normalizer tests for shadowing edge cases                | Medium       | 12min  | Medium | D24.3    | Low     |
| D24.5 | Benchmark go/types overhead vs current heuristic                | Medium       | 12min  | Medium | D24.4    | Low     |
| D25.1 | Port alpha-normalization to syntax/templ/ parser                | Medium       | 12min  | Medium | —        | Medium  |
| D25.2 | Walk templ AST: canonicalize locals in templ blocks             | Medium       | 12min  | Medium | D25.1    | Medium  |
| D25.3 | Integrate into templ transform pipeline                         | Medium       | 12min  | Medium | D25.2    | Medium  |
| D25.4 | Tests: renamed templ variables detected as clones               | Medium       | 12min  | Medium | D25.3    | Medium  |
| D25.5 | Dogfood on .templ files                                         | Medium       | 12min  | Medium | D25.4    | Medium  |

**Tier D total: 220min (17 sub-tasks) — T1 blocked on GO/NO-GO**

### TIER E — Deferred (External Blockers)

| #     | Sub-Task                                             | Impact | Effort | Risk | Deps           | CustVal |
| ----- | ---------------------------------------------------- | ------ | ------ | ---- | -------------- | ------- |
| E26.1 | Audit remaining files on encoding/json v1            | Low    | 10min  | Low  | Go 1.26 stable | Low     |
| E26.2 | Migrate MarshalJSON/UnmarshalJSON to v2 API          | Low    | 12min  | Low  | E26.1          | Low     |
| E26.3 | Verify output unchanged via golden-file comparison   | Low    | 12min  | Low  | E26.2          | Low     |
| E27.1 | Break import cycle: extract Node to separate package | Medium | 12min  | High | —              | Low     |
| E27.2 | Create syntax/golang facade API                      | Medium | 12min  | High | E27.1          | Low     |
| E27.3 | Update all callers; verify no cycle                  | Medium | 12min  | High | E27.2          | Low     |
| E28.1 | Implement hybrid slice/map: slice for ≤K, map for >K | Low    | 12min  | Low  | —              | Low     |
| E28.2 | Benchmark: find crossover point K                    | Low    | 10min  | Low  | E28.1          | Low     |
| E28.3 | Tests: verify correctness at transition boundary     | Low    | 10min  | Low  | E28.1          | Low     |

**Tier E total: 102min (9 sub-tasks)**

---

## Summary Statistics

| Tier    | Sub-Tasks | Total Effort       | Unblock T1?       |
| ------- | --------- | ------------------ | ----------------- |
| **S**   | 7         | 72min              | No                |
| **A**   | 9         | 92min              | No                |
| **B**   | 19        | 198min             | No                |
| **C**   | 24        | 262min             | No                |
| **D**   | 17        | 220min             | T1 needs GO/NO-GO |
| **E**   | 9         | 102min             | External blockers |
| **ALL** | **85**    | **~946min (~16h)** |                   |

### Dependency Chain (critical path)

```
S1 (BDD Type 2) ──────────────────────────────────────► proves headline feature
S4 (baseline file) ──► B11 (--diff) ──► B15 (diff cmd)
                       B12 (--update)
C19 (decouple) ──► C22 (split packages)
C21 (clone unify) ──► C20 (fragment unify)
D23 (T1 tokenization) ── BLOCKED on GO/NO-GO ──► enables accurate thresholds
```

### Recommended Execution Order

1. **Start with Tier S** (72min) — proves the headline feature works, fixes stale docs
2. **Then Tier A** (92min) — completes output formats, adds UX polish
3. **Then Tier B** (198min) — feature completeness for CI workflows
4. **Tier C** (262min) when architecture debt becomes painful
5. **Tier D** only after T1 GO/NO-GO decision
6. **Tier E** when external blockers clear (json/v2 stabilizes, import cycle solvable)

### What's NOT Included (already done)

T2-T12, T17-T22, T24 are **all complete and committed** (20/24 sprint tasks). See `docs/status/2026-06-20_19-56_final-sprint-complete.md` for details.
