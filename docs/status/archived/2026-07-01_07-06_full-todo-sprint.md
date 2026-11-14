# Status Report — 2026-07-01 Full TODO Sprint

> **Live remainder:** archived 2026-11-14 (cadence wave); every item in this report is struck or routed, and open work is owned by TODO_LIST.md and ROADMAP.md.

**Date:** 2026-07-01 07:06 CEST
**Branch:** `fork`
**HEAD:** (uncommitted — pending commit of this session's work)
**Previous HEAD:** `7e5a5cb` — feat: decouple printer actionability from syntax.Node via CloneNode DTO

---

## Executive Summary

Executed the **entire feasible TODO list** in a single session: **13 tasks completed**,
**3 assessed (no action needed)**, **2 deferred with rationale**. The codebase
went from 7 open TODO items + 8 Pareto plan items to **zero actionable open items**.

**35 files changed** (+398 / -386 lines), **4 new files**, **1 deleted file**.

The headline achievements: `domain.CloneRef` value object unifies the 7 parallel
Clone types at the embed level; the sort comparator factory eliminates 3 duplicate
switch statements; the hash pipeline shares a single `groupByHash` implementation;
HTML reports gained collapse-all controls; SARIF output is enriched for GitHub
Code Scanning.

All verification gates green: **build clean, 24/24 test packages pass, 0 lint
issues, 0 race detector failures**.

---

## a) FULLY DONE (13 tasks)

### T23 — CloneRef Value Object ✅

**The headline architectural achievement.** Introduced `domain.CloneRef` — a
4-field value object (`Filename`, `LineStart`, `LineEnd`, `Fragment` +
`LineCount()` method) that is **embedded** in `domain.ProcessedClone` and
`pkg/artdupl.Clone`. This eliminates field-name drift across the 7 parallel
Clone types without collapsing the DTO boundary between packages.

| File                                 | Change                                                       |
| ------------------------------------ | ------------------------------------------------------------ |
| ~~`domain/clone_ref.go` (NEW)~~ done at `e581fd92` | ~~`CloneRef` type + `LineCount()` method~~ |
| ~~`domain/processed_clone.go`~~ done at `e581fd92` | ~~Embeds `CloneRef` instead of 4 flat fields~~ |
| ~~`pkg/artdupl/types.go`~~ done at `e581fd92` | ~~Embeds `domain.CloneRef` instead of 4 flat fields~~ |
| ~~`pkg/artdupl/detector_conversion.go`~~ done at `e581fd92` | ~~Struct literal uses nested `CloneRef{}`~~ |
| ~~`printer/clone_processor.go`~~ done at `e581fd92` | ~~Struct literal uses nested `CloneRef{}`~~ |
| ~~8 test files~~ done at `e581fd92` | ~~Struct literals updated to `CloneRef: CloneRef{...}` pattern~~ |

**Result:** Struct literals across the codebase use a consistent
`CloneRef: CloneRef{Filename: ..., LineStart: ..., LineEnd: ...}` pattern.
The 7 types (`ProcessedClone`, `artdupl.Clone`, `JSONClone`, `CloneOccurrenceView`,
`CloneWithContent`, `CloneDiff`, `simpleJSONClone`) remain separate for
Printer/SDK DTO independence — only the common location fields are shared.

### T28 — Sort Comparator Factory ✅

Replaced 3 parallel 4-criteria sort switch implementations with a single
generic comparator factory.

| File                                  | Change                                                                                                                                                                                                                               |
| ------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| ~~`printer/sort_unified.go`~~ done at `e581fd92` | ~~Added `GroupMetrics[T]`, `makeGroupComparator[T]`, `sortGroupsByCriteria[T]`~~ |
| ~~`printer/sorter.go`~~ done at `e581fd92` | ~~`SortCloneGroups` delegates to factory. Deleted 4 dead `[][]*syntax.Node` sort functions (`SortClonesBySize`, `SortClonesByOccurrence`, `SortClonesByHash`, `SortClonesByTotalTokens`), `isEmptyOrLessThanEmpty`, `countNonNilNodes`~~ |
| ~~`printer/text.go`~~ done at `e581fd92` | ~~`OutputText` group sort delegates to factory (was 20-line inline switch)~~ |
| ~~`printer/sorting_integration_test.go`~~ done at `e581fd92` | ~~Removed `TestCommonSortingUtilities` (tested dead functions)~~ |
| ~~`printer/test_helper.go` (DELETED)~~ done at `e581fd92` | ~~`TestCloneSortingWithData` had zero live callers~~ |

**Dead code removed:** ~120 lines of `[][]*syntax.Node` sort functions that had
zero production callers (only test-called).

### T18 — Hash Pipeline Consolidation ✅

Extracted the shared hash+group pipeline from `FindFileDuplicates` (free fn)
and `FindDuplOver` (method) into two shared helpers.

| File                    | Change                                                                                                                                                                                                                  |
| ----------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ~~`hash/file_detector.go`~~ done at `e581fd92` | ~~Added `groupByHash(ctx, files)` + `filterDuplicateGroups(groups, threshold)` + `fileHashToFragment(fh)`. Both `FindFileDuplicates` and `FindDuplOver` now share these. `FindFileDuplicates` now takes `context.Context`~~ |

### T41 — Context Through File Feeders ✅

`filepath.Walk` in `cmd/run_crawl.go` now respects context cancellation.

| File               | Change                                                                                                                                                                                                 |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| ~~`cmd/run_crawl.go`~~ done at `e581fd92` | ~~`handleWalkEntry` checks `ctx.Err()` at each entry, returns the context error to abort the walk. `crawlDirectoryWithOpts` suppresses `context.Canceled`/`context.DeadlineExceeded` from stderr logging~~ |

### HTML Collapsible Clone Groups ✅

| File                        | Change                                                                                                         |
| --------------------------- | -------------------------------------------------------------------------------------------------------------- |
| ~~`printer/html_views.go`~~ done at `e581fd92` | ~~Added "Collapse All" / "Expand All" toolbar buttons~~ |
| ~~`printer/html.go`~~ done at `e581fd92` | ~~Added `collapseAll(collapse)` JS function~~ |
| ~~`printer/html_template.go`~~ done at `e581fd92` | ~~Added `.collapse-controls` / `.collapse-btn` CSS + `▼`/`▶` collapse indicators on clone headers via `::before`~~ |
| ~~`printer/testdata/*.golden`~~ done at `e581fd92` | ~~Updated golden files (2 files)~~ |

### SARIF Rule Metadata Enrichment ✅

| File               | Change                                                                                                                                                                                                       |
| ------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| ~~`printer/sarif.go`~~ done at `e581fd92` | ~~Added `SARIFRuleProperties` struct (`Precision`, `ProblemSeverity`, `Tags`). Single rule now has `precision: "high"`, `problem.severity: "warning"`, tags: `["maintainability", "duplicate-code", "design"]`~~ |

### T36 — Perf Regression CI ✅

| File                                   | Change                                                                                                                                                 |
| -------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| ~~`syntax/perf_regression_test.go` (NEW)~~ done at `e581fd92` | ~~`TestPerfRegressionSerialize` (50ms threshold on 100-depth tree) + `TestPerfRegressionHashSeq` (50ms threshold on 10k nodes). Skipped in `-short` mode~~ |
| ~~`flake.nix`~~ done at `e581fd92` | ~~Added `checks.bench` running `go test -run='^TestPerfRegression' -v ./syntax/...`~~ |

### Cache Version-Mismatch Warning ✅

| File                  | Change                                                                                                                                                        |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ~~`cache/file_cache.go`~~ done at `e581fd92` | ~~`Get()` now logs the deserialize error before removing stale entries (was silently swallowed). `loadMetadata()` warns when `metadata.Version != CacheVersion`~~ |

### JSON Config Migration Shim ✅

| File                             | Change                                                                                                                                                                       |
| -------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| ~~`config/config_migrate.go` (NEW)~~ done at `e581fd92` | ~~`Config.UnmarshalJSON` converts legacy `"semantic": false` → `"detectionMode": "exact"` when `detectionMode` is absent from JSON. Uses type-alias trick to prevent recursion~~ |

### ADR-0008 — Semantic Encoding Layout ✅

| File                                              | Change                                                                                                                                                                                                       |
| ------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| ~~`docs/adr/0008-semantic-encoding-layout.md` (NEW)~~ done at `e581fd92` | ~~Documents the `[24-bit identifier/operator hash][8-bit base AST node type]` packing, consumer contract (`DecodeBaseType`), `hashSeq` alignment, cache format dependency, and 24-bit hash truncation analysis~~ |

### CI Template Validation ✅

| File                                           | Change                                                                                                   |
| ---------------------------------------------- | -------------------------------------------------------------------------------------------------------- |
| ~~`templates/github-actions-duplicate-check.yml`~~ done at `e581fd92` | ~~Fixed `./...` → `.` (Go package pattern → filesystem path). Pre-commit hook already correctly configured~~ |

### AGENTS.md + TODO_LIST.md Updated ✅

| File           | Change                                                                                                                   |
| -------------- | ------------------------------------------------------------------------------------------------------------------------ |
| ~~`AGENTS.md`~~ done at `e581fd92` | ~~Added CloneRef embed pattern, sort comparator factory, config migration shim, `filepath.Walk` context cancellation notes~~ |
| ~~`TODO_LIST.md`~~ done at `e581fd92` | ~~Added "Completed (2026-07-01) — Full TODO Sprint" section with all 13 tasks + 3 assessments + 2 deferred items~~ |

---

## b) PARTIALLY DONE (0 items)

Nothing is half-finished. Every task that was started was completed and verified.

---

## c) NOT STARTED (2 items — deferred with rationale)

### T25 — Split `printer/` into sub-packages

~~**Status:** Assessed — requires interface inversion.~~ DONE 2026-07-26: printer/actionability + printer/stats sub-packages shipped in v0.5.0.

The printer package (~33 source files, ~3500 lines) has clean conceptual boundaries
(`stats/`, `html/`, `analyze/`), but splitting requires:

1. Moving `Printer`, `ReadFile`, `StatsPrinter` interfaces to a separate base
   package (core `printer.go` references `StatsPrinter`, creating circular deps)
2. Extracting shared test helpers (`mockReadFile`, `processTestNodes`,
   `createMockCloneGroup`) from `_test.go` files into a testutil package
3. Handling `.(*stats)` type assertions to unexported types in test files

This is multi-session architectural design work, not a mechanical refactor.

### T24 — Branded `NodeType int32`

**Status:** Assessed — the proposed solution doesn't solve the stated problem.

The TODO described introducing a single `syntax.NodeType` type to prevent
cross-package int32 collision between golang and templ node type constants.
However, a single shared `NodeType` type would still allow golang and templ
constants to share the same `NodeType` type — the collision would still exist.
The proper fix requires **per-package** `NodeType` types (`golang.NodeType`,
`templ.NodeType`), which is even more invasive.

The current 8-bit shared encoding space is intentional (see ADR-0008).
Currently ~70 of 256 values are used, leaving comfortable headroom.

**HIGH RISK:** Touches gob serialization cache format. Needs feature flag +
cache-version migration.

---

## d) TOTALLY FUCKED UP (0 items)

Nothing is broken. Nothing regressed. All 24 test packages pass. All race
tests pass. Zero lint issues.

**One cosmetic issue from the session:** The Python regex-based struct literal
transformation for CloneRef embedding initially produced double/triple-nested
`CloneRef{CloneRef: CloneRef{...}}` patterns in `domain/domain_test.go` and
`pkg/artdupl/detector_types_test.go`. All were caught by the compiler and fixed
before the final verification gate. No broken code was committed.

---

## e) WHAT WE SHOULD IMPROVE

1. ~~**Split printer/ eventually** — T25 is the biggest remaining architectural~~ done — printer split shipped 2026-07-26 (stats, actionability, finding subpackages)
   debt. The package has 33 files and ~3500 lines. T22 (DTO decoupling) + T23
   (CloneRef) removed the blockers, but the circular interface dependency remains.

2. ~~**Test the perf regression thresholds under load** — The 50ms thresholds are~~ **Superseded — perf gate is benchstat-based today (.github/workflows/performance.yml).**
   generous but haven't been validated on a slow CI runner. May need tuning.

3. ~~**Add `Collapse All` state persistence** — The HTML collapse-all feature~~ **Won't implement — dropped — only diff-mode persists (printer/html.go localStorage).**
   doesn't remember its state across page reloads (unlike the diff-mode toggle
   which uses `localStorage`).

4. ~~**CloneRef migration for remaining types** — `JSONClone`, `CloneOccurrenceView`,~~ **Superseded — printer DTOs deliberately keep flat fields (DTO boundary by design).**
   `CloneWithContent` still have flat fields. They could embed `CloneRef` too,
   but the conversion is mechanical with lower value since they're
   printer-internal types.

5. ~~**Integration test for config migration shim** — The `semantic: false` →~~ done — legacy semantic-false cases in config/config_test.go
   `detectionMode: exact` conversion has no dedicated test yet.

---

## f) TOP 25 THINGS TO DO NEXT

| #  | Task                                                | Impact | Effort | Status                        |
| -- | --------------------------------------------------- | ------ | ------ | ----------------------------- |
| ~~1~~ done — printer split shipped 2026-07-26 | ~~T25 Split printer/ (stats, html, analyze)~~ | ~~HIGH~~ | ~~2-3d~~ | ~~Needs design~~ |
| ~~2~~ **Superseded — tracked in ROADMAP.md (Watch mode).** | ~~T32 Watch mode (`--watch`)~~ | ~~MED~~ | ~~1d~~ | ~~Not started~~ |
| ~~3~~ **Won't implement — dropped — Go + templ only (AGENTS.md).** | ~~T38 TypeScript/JS support~~ | ~~HIGH~~ | ~~1w+~~ | ~~Not started~~ |
| ~~4~~ **Won't implement — dropped — Go + templ only (AGENTS.md).** | ~~T39 Python support~~ | ~~HIGH~~ | ~~1w+~~ | ~~Not started~~ |
| ~~5~~ **NOT-DO — assessed here: shared encoding space is intentional (ADR-0008).** | ~~T24 Branded NodeType (per-package types)~~ | ~~MED~~ | ~~180m~~ | ~~Needs design~~ |
| ~~6~~ **Won't implement — dropped — no backpressure work recorded.** | ~~SDK streaming backpressure handling~~ | ~~LOW~~ | ~~40m~~ | ~~Not started~~ |
| ~~7~~ **Won't implement — dropped — LRU + allocation budgets suffice (ADR-0022).** | ~~Cache eviction policy (beyond LRU)~~ | ~~LOW~~ | ~~40m~~ | ~~Not started~~ |
| ~~8~~ **Superseded — printer DTOs deliberately keep flat fields.** | ~~Extend CloneRef to JSONClone + CloneOccurrenceView~~ | ~~LOW~~ | ~~30m~~ | ~~Not started~~ |
| ~~9~~ done — config/config_test.go legacy semantic cases | ~~Add test for config migration shim~~ | ~~LOW~~ | ~~15m~~ | ~~Not started~~ |
| ~~10~~ **Won't implement — dropped — only diff-mode persists (printer/html.go).** | ~~HTML collapse-all localStorage persistence~~ | ~~LOW~~ | ~~20m~~ | ~~Not started~~ |
| ~~11~~ done — website/src/content/docs/guides/detection-methods.mdx | ~~"How art-dupl detects clones" deep-dive doc~~ | ~~LOW~~ | ~~40m~~ | ~~Not started~~ |
| ~~12~~ **Superseded — benchstat-based performance gate (.github/workflows/performance.yml).** | ~~Perf threshold tuning for slow CI runners~~ | ~~LOW~~ | ~~20m~~ | ~~Needs data~~ |
| ~~13~~ **Won't implement — dropped — Fang stays v1 (go.mod).** | ~~T26 Fang v2 migration~~ | ~~LOW~~ | ~~—~~ | ~~Blocked (Fang v2)~~ |
| ~~14~~ **NOT-DO — direct json/v2 rejected (ADR-0024 + internal/jsonv2gate).** | ~~T12 encoding/json v2 migration~~ | ~~LOW~~ | ~~—~~ | ~~Blocked (Go 1.27)~~ |
| ~~15~~ **Won't implement — dropped — import cycle accepted; arch-lint documents the boundary.** | ~~T40 syntax/golang facade~~ | ~~LOW~~ | ~~—~~ | ~~Blocked (import cycle)~~ |
| ~~16~~ **Won't implement — dropped — same as above; facade never built.** | ~~Hide `syntax/golang` behind facade~~ | ~~LOW~~ | ~~—~~ | ~~Blocked (import cycle)~~ |
| ~~17~~ done — stdin context cancellation session 2026-07-16 07:23 (docs/status/2026-07-16_07-23_stdin-context-cancellation.md) | ~~Thread context through stdin scanner~~ | ~~LOW~~ | ~~—~~ | ~~Blocked (inherently blocking)~~ |
| ~~18~~ **NOT-DO — ADR-0022 chose sorted slices (linear scan <= 8, binary above).** | ~~Hybrid slice/map transition storage~~ | ~~LOW~~ | ~~60m~~ | ~~Deferred (map is O(1))~~ |
| ~~19~~ **Won't implement — dropped — single-rule SARIF wire frozen by golden test.** | ~~SARIF: add more rules (per-clone-type)~~ | ~~LOW~~ | ~~40m~~ | ~~Not started~~ |
| ~~20~~ done — check subcommand shows new clones vs baseline (cmd/baseline_cmd.go, ADR-0029) | ~~Baseline diff mode (show new clones since baseline)~~ | ~~MED~~ | ~~2h~~ | ~~Not started~~ |
| ~~21~~ **Won't implement — dropped — no --since flag; git-aware incremental never adopted.** | ~~`--since <git-ref>` incremental analysis~~ | ~~MED~~ | ~~4h~~ | ~~Not started~~ |
| ~~22~~ done — setup-go module caching in CI (.github/workflows/ci.yml) | ~~Go module proxy caching in CI~~ | ~~LOW~~ | ~~30m~~ | ~~Not started~~ |
| ~~23~~ done — templ alpha-normalization (syntax/templ, Type 2) | ~~Templ semantic mode (identifier/operator encoding)~~ | ~~MED~~ | ~~4h~~ | ~~Not started~~ |
| ~~24~~ done — examples/ directory | ~~SDK examples package (godoc-renderable)~~ | ~~LOW~~ | ~~2h~~ | ~~Not started~~ |
| ~~25~~ done — timing stage report (commit def1bbbb) | ~~Performance profiling dashboard (HTML)~~ | ~~LOW~~ | ~~3h~~ | ~~Not started~~ |

---

## g) THE #1 QUESTION

**Is art-dupl meant to be a polished Go-only tool, or a multi-language platform?**

This question has been open since the prior session and remains unanswered.
The entire feasible TODO list is now done — the remaining work is either
multi-day architectural refactoring (T25 printer split), blocked on external
dependencies (Fang v2, Go 1.27), or strategic product decisions (T32 watch
mode, T38/T39 new languages).

The codebase is now **correct, deterministic, safe (race-tested), well-documented
(8 ADRs), lint-clean, and architecturally decoupled** (printer actionability
off `syntax.Node`, CloneRef shared value object, sort factory, hash pipeline
consolidated). The foundation is solid for either path.

**If Go-only:** invest in T25 (printer split), T24 (per-package NodeType), baseline
diff mode, and templ semantic mode.

**If multi-language:** invest in T38/T39 (TypeScript/Python frontends) — the
`syntax/` package architecture is designed for this (`syntax/golang/` and
`syntax/templ/` are existing frontends; new languages follow the same pattern).

---

## Verification Gate

| Check                               | Result                 |
| ----------------------------------- | ---------------------- |
| ~~`go build ./...`~~ done — verified — session gate, historical | ~~✅ clean~~ |
| ~~`go test ./...`~~ done — verified — session gate, historical | ~~✅ 24/24 packages pass~~ |
| ~~`golangci-lint run`~~ done — verified — session gate, historical | ~~✅ 0 issues~~ |
| ~~`CGO_ENABLED=1 go test -race ./...`~~ done — verified — session gate, historical | ~~✅ 24/24, 0 failures~~ |

---

## Files Changed (35 total)

### New files (4)

- `domain/clone_ref.go` — CloneRef value object
- `config/config_migrate.go` — JSON config migration shim
- `docs/adr/0008-semantic-encoding-layout.md` — Semantic encoding ADR
- `syntax/perf_regression_test.go` — Performance regression tests

### Deleted files (1)

- `printer/test_helper.go` — Dead code (TestCloneSortingWithData had zero callers)

### Modified files (30)

See `git diff --stat HEAD` for the complete list.
