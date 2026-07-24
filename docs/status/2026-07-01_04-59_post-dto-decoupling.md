# Status Report — 2026-07-01 Post-DTO-Decoupling

**Date:** 2026-07-01 04:59 CEST
**Branch:** `fork`
**HEAD:** (uncommitted — pending commit of this session's work)
**Previous reports:** `2026-07-01_01-34_post-execution-sprint.md`, `2026-06-30_23-52_post-dedup-comprehensive-status.html`

---

## Executive Summary

> **Resolution (2026-07-01, later session):** All "Top 25" items listed as "not started" below (T23 CloneRef, T28 sort factory, T18 hash pipeline, T41 context, T36 perf regression, ADR-0008, JSON config migration, cache warning, SARIF enrichment, HTML collapse-all) were completed ~2 hours later in the `2026-07-01_07-06_full-todo-sprint.md` session. T25 (printer split) and T24 (branded NodeType) remain deferred — see TODO_LIST.md "Deferred" section.

Executed Tier 2 (D2, D3, G1) + Path A (T22 ProcessedClone DTO) from the
Pareto plan. **17 files changed**, **+600 / -200 lines (net)**. The printer
package's actionability layer is now **fully decoupled from `syntax.Node`** —
zero `syntax` imports in the 5 actionability files. All verification gates green:
build, 24/24 test packages, 0 lint issues, 0 race detector failures.

---

## a) FULLY DONE (7 tasks)

### D2 — Refresh TODO_LIST.md ✅

Updated to 2026-07-01 reality: marked 8 items done (cache deep-copy, Name()
method, debug.Stack removal, SortCriteria/OutputFormat relocation, FuncLit
normalization, MaxChildrenSerial, destructive serial, ProcessedClone DTO),
added the sprint completion section + lint cleanup section.

### D3 — Refresh FEATURES.md ✅

Updated to 2026-07-01: SHA-256 cache (was SHA-1), CacheVersion 2, cache
eviction (`--max-cache-entries`), parallel incremental parsing +
singleflight, DetectionMode enum (ADR-0007), CI template paths, FuncLit
normalization, 9 dedicated race tests.

### G1 — Wire `-race` + lint into CI ✅

Added `race` check to `flake.nix` (`CGO_ENABLED=1 go test -race ./...`).
Discovered and fixed a **pre-existing data race** in
`detection/coverage_test.go:167` — parallel subtests shared a mutable
`MultiDetector` via `md.cfg = ...`. Fixed by constructing a fresh
`MultiDetector` per subtest.

### T22 — ProcessedClone DTO (Path A architecture) ✅

**The headline achievement.** Introduced `domain.CloneNode` — a minimal
recursive tree type with 4 fields (`BaseType int32`, `Name string`,
`Filename string`, `Children []*CloneNode`). The actionability evaluation
layer now operates exclusively on `[][]*domain.CloneNode`, not
`[][]*syntax.Node`.

| File                                         | Change                                                                                                           |
| -------------------------------------------- | ---------------------------------------------------------------------------------------------------------------- |
| `domain/clone_node.go` (NEW)                 | `CloneNode` type with `HasChildren()` method                                                                     |
| `printer/clone_processor.go`                 | Added `ToCloneNodeSeqs` (exported) + `toCloneNodeSeqs` + `syntaxToCloneNode` bridge; removed `baseTypeOf` helper |
| `printer/actionability.go`                   | All signatures: `*syntax.Node` → `*domain.CloneNode`; removed `syntax` import                                    |
| `printer/actionability_control_flow.go`      | Same signature migration                                                                                         |
| `printer/actionability_data.go`              | Same signature migration                                                                                         |
| `printer/actionability_patterns_expanded.go` | Same signature migration                                                                                         |
| `printer/actionability_test_patterns.go`     | Same signature migration                                                                                         |
| `printer/actionability_test.go`              | Test data: `syntax.Node{Type:}` → `domain.CloneNode{BaseType:}`                                                  |
| `printer/actionability_patterns_test.go`     | Same test data migration                                                                                         |
| `cmd/run_output.go`                          | Updated caller: `printer.ToCloneNodeSeqs(uniq)`                                                                  |

**Result:** The 5 actionability files import only `domain` and
`syntax/golang` (for AST node-type constants). Zero `syntax` package
imports. The printer is now testable without `syntax.Node` trees.

### L1-L2 — Production + test lint cleanup (carried from prior session) ✅

Fixed exhaustive switch (`cmd/util.go`), wrapcheck (`types.go`),
27 test errcheck via `mustUpdate` helper, stale AGENTS.md reconciliation.

### D1 — Stale doc fix (carried from prior session) ✅

Updated all 5 historical reports (2026-06-30 + 2026-07-01) with resolution
banners marking the 4 CRITICALs as RESOLVED.

---

## b) PARTIALLY DONE (1 item)

### T25 — Split printer/ into sub-packages

**Status:** Unblocked (was blocked by T22, now done). The printer package
(~29 files, ~3500 lines) can now be split into `stats/`, `html/`,
`analyze/` sub-packages. Not started this session.

---

## c) NOT STARTED (remaining Pareto plan items)

| Task                             | Priority | Effort | Notes                                        |
| -------------------------------- | -------- | ------ | -------------------------------------------- |
| T25 Split printer/               | P1       | 90m    | Unblocked — next architectural step          |
| T36 Perf regression CI           | P1       | 40m    | Benchmarks exist, need threshold gating      |
| T23 CloneRef unification         | P2       | 80m    | 7 parallel Clone types → shared value object |
| T28 Sort comparator factory      | P3       | 12m    | 4 implementations over 3 representations     |
| T18 Hash pipeline consolidation  | P3       | 12m    | Free fn vs method                            |
| T41 Context through file feeders | P2       | 60m    | Latent gap (process exits on cancel)         |
| T24 Branded NodeType             | P2       | 180m   | HIGH RISK — touches gob cache format         |
| T32 Watch mode                   | Future   | 1d     | `--watch` continuous monitoring              |
| T38/T39 TypeScript/Python        | Future   | 1w+    | New language frontends                       |

---

## d) TOTALLY FUCKED UP (0 items)

Nothing is broken. Nothing regressed. All 24 test packages pass. All race
tests pass. Zero lint issues.

---

## e) WHAT WE SHOULD IMPROVE

1. **Split printer/ now** — T22 removed the blocker. The package has 29 files
   and ~3500 lines. Clean boundaries exist: `stats/`, `html/`, `analyze/`.
2. **Perf regression CI** — benchmarks exist but aren't gated. A 10x regression
   would go unnoticed.
3. **CloneRef unification** — 7 parallel Clone types is still messy. T22 proved
   the DTO approach works; CloneRef would apply it at the type level.
4. **Test the CI templates** — `templates/github-actions-duplicate-check.yml`
   and `templates/pre-commit-hook.yaml` exist but haven't been verified in a
   real CI run.

---

## f) TOP 25 THINGS TO DO NEXT

| #   | Task                                                | Impact | Effort                 |
| --- | --------------------------------------------------- | ------ | ---------------------- |
| 1   | T25 Split printer/ (stats, html, analyze)           | HIGH   | 90m                    |
| 2   | T36 Perf regression CI (threshold assertions)       | HIGH   | 40m                    |
| 3   | T23 CloneRef value object (unify 7 Clone types)     | MED    | 80m                    |
| 4   | T28 Sort comparator factory                         | LOW    | 12m                    |
| 5   | T18 Hash pipeline consolidation                     | LOW    | 12m                    |
| 6   | T41 Context through file feeders                    | LOW    | 60m                    |
| 7   | T24 Branded NodeType (HIGH RISK: cache format)      | MED    | 180m                   |
| 8   | Test GitHub Actions template in real CI             | MED    | 30m                    |
| 9   | Test pre-commit hook template                       | MED    | 30m                    |
| 10  | Apply sendCtx to remaining 10 send sites            | LOW    | 12m                    |
| 11  | Cache version-mismatch warning log                  | MED    | 8m                     |
| 12  | JSON config migration shim (semantic→detectionMode) | MED    | 10m                    |
| 13  | ADR for semantic encoding layout                    | LOW    | 30m                    |
| 14  | "How art-dupl detects clones" deep-dive             | LOW    | 40m                    |
| 15  | go.mod dependency audit in CI                       | LOW    | 20m                    |
| 16  | HTML report: collapsible clone groups               | LOW    | 40m                    |
| 17  | T32 Watch mode (`--watch`)                          | MED    | 1d                     |
| 18  | T38 TypeScript/JS support                           | HIGH   | 1w+                    |
| 19  | T39 Python support                                  | HIGH   | 1w+                    |
| 20  | SDK streaming backpressure handling                 | LOW    | 40m                    |
| 21  | SARIF rule metadata enrichment                      | LOW    | 20m                    |
| 22  | Cache eviction policy (beyond LRU)                  | LOW    | 40m                    |
| 23  | T26 Fang v2 migration                               | LOW    | blocked                |
| 24  | T12 encoding/json v2 migration                      | LOW    | blocked (Go 1.27)      |
| 25  | T40 syntax/golang facade                            | LOW    | blocked (import cycle) |

---

## g) THE #1 QUESTION

**The strategic direction question remains open:** Is art-dupl meant to be a
polished Go-only tool, or a multi-language platform? The answer determines
whether to invest in T23/T24/T25 (architectural purity) or T32/T38/T39
(user-facing features). The codebase is now correct, deterministic, safe, and
the printer is decoupled — the foundation for either path is solid.

---

## Verification Gate

| Check                               | Result                 |
| ----------------------------------- | ---------------------- |
| `go build ./...`                    | ✅ clean               |
| `go test ./...`                     | ✅ 24/24 packages pass |
| `golangci-lint run`                 | ✅ 0 issues            |
| `CGO_ENABLED=1 go test -race ./...` | ✅ 0 failures          |
