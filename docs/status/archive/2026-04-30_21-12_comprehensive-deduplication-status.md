# Comprehensive Status Report — 2026-04-30 21:12

_Session focused on code deduplication via `art-dupl` self-analysis and `branching-flow` structural analysis._

---

## A) Fully Done

### 1. `art-dupl` Self-Analysis Deduplication

- **`syntax/golang/transform.go`** — Extracted 3 helper methods:
  - `addBodyStatements()` — consolidates body statement iteration
  - `addIdentifierNames()` — consolidates identifier name iteration
  - `addKeyValue()` — consolidates key-value pair transformation
- **`pkg/artdupl/detector_pipeline.go`** — Extracted `processCloneGroups()` to eliminate duplicate loop logic in `runDetection()` and `streamDetectionResults()`
- **Result**: Source code duplicates reduced; all 24 test packages pass.

### 2. `branching-flow dupe` — Duplicate Type Elimination

- **`printer/sarif.go`** — Replaced `SARIFRegion` struct with `LineRangeMixin` embedding (identical fields: `StartLine`, `EndLine`)
- **`printer/stats_visualization.go`** — Replaced local `fileStat` struct with existing `FileStatMixin` embedding
- **`printer/json.go`** — Added JSON tags to `LineRangeMixin` for proper SARIF serialization
- **Result**: **2 → 0 actionable duplicate type groups**

### 3. Lint Fix

- Removed unused `gocognit` from nolint directive in `transform.go` (complexity reduced by refactoring)

### Commits Made (3 total, all pushed to `fork`)

| Commit    | Message                                                             |
| --------- | ------------------------------------------------------------------- |
| `6e0e47f` | `refactor(transform): extract helper methods to reduce duplication` |
| `9bfa62f` | `refactor(printer): unify SARIFRegion with LineRangeMixin`          |
| `6658fa9` | `refactor(printer): replace local fileStat with FileStatMixin`      |

---

## B) Partially Done

Nothing was left partially complete. Each task was either fully done or explicitly deferred/skipped with documented reasoning.

---

## C) Not Started (Deferred)

### High-Effort Structural Refactoring (68+ references)

| Task                                                                  | Scope                             | Reason Deferred                                       |
| --------------------------------------------------------------------- | --------------------------------- | ----------------------------------------------------- |
| Extract `FilterOptions` from `FlagValues` (15 bools)                  | `cmd/config_builder.go` + 68 refs | Massive blast radius; needs dedicated session         |
| Extract `FilterOptions` from `Config` (14 bools)                      | `config/config.go` + many refs    | Must match FlagValues extraction                      |
| Create `CommonFlags` for FlagValues↔Config overlap (20 shared fields) | `cmd/`, `config/`, `cli/`         | Largest structural change; multi-package coordination |
| Wire `CommonFlags` through `BuildConfigFromFlags`                     | `cmd/config_builder.go`           | Depends on above                                      |
| Update all `FlagValues` callers                                       | `cmd/*.go`                        | 45+ call sites                                        |

### Lower-Priority Items

| Task                                                    | Scope              | Reason                                                   |
| ------------------------------------------------------- | ------------------ | -------------------------------------------------------- |
| Extract `OutputOptions` from `RuntimeConfig` (6 bools)  | `cli/runtime.go`   | Below actionable threshold                               |
| Extract `MetadataFlags` from `ReportMetadata` (4 bools) | `printer/html.go`  | Below actionable threshold                               |
| Brand SARIF `RuleID`/`ID`                               | `printer/sarif.go` | Not domain concepts, just SARIF spec strings             |
| Extract LineRange mixin for `JSONClone`                 | `printer/json.go`  | Different JSON tag schemas (`line_start` vs `startLine`) |
| Phantom types (545 issues)                              | Project-wide       | Too many low-value `string` params; separate effort      |
| Panic conditions (532 issues)                           | Project-wide       | Requires manual per-case review                          |

---

## D) Totally Fucked Up — Nothing

No regressions. No broken tests. No failed builds. All changes verified.

---

## E) What We Should Improve

### Architecture

1. **FlagValues/Config convergence** — These share 20 fields. A shared `CommonFlags` or `FilterOptions` struct would eliminate the massive `BuildConfigFromFlags` mapping function and make both structs smaller and more maintainable.
2. **`printer` package has many small struct types** — `CloneWithContentMixin`, `FileInfo`, `JSONClone`, `LineRangeMixin`, `FileStatMixin`, `SARIFRegion` all carry similar `Filename`/`LineStart`/`LineEnd` fields but with different JSON serialization needs. A common core type with JSON-tag overrides could help.
3. **`StatsData` has 23 fields** (branching-flow large-struct warning) — Could be decomposed into `AnalysisMetrics`, `FileStatistics`, `CloneStatistics`.

### Type Safety

4. **Phantom types** — 545 instances of primitive `string`/`int` params that could benefit from branded types. The project already has `domain.CloneGroupID`, `domain.AnalysisID`, `domain.Hash` — this pattern should be extended.
5. **Strong IDs for SARIF** — `SARIFRule.ID` and `SARIFResult.RuleID` are bare `string` but represent rule identifiers.

### Error Handling

6. **branching-flow context score: 82.9/100** — 6 "high" severity issues in `config/enum_helpers.go` and `domain/helpers.go`. However, these appear to be false positives (flagging function parameters like `isValid` and `validator` as "lost context" — you can't serialize a function into an error message).

### Code Quality

7. **Lint warnings remain** (36 total) — `wsl_v5` (17), `nlreturn` (4), `tparallel` (4), `goconst` (4), `golines` (1), `errchkjson` (1), `prealloc` (1), `revive` (1), `unused` (1), `nolintlint` (1), `exhaustruct` (1). Not introduced by us — pre-existing.

---

## F) Top 25 Things We Should Get Done Next

Sorted by **Impact × Feasibility**:

### Quick Wins (≤30 min each)

| # | Task                                                                   | Impact | Effort |
| - | ---------------------------------------------------------------------- | ------ | ------ |
| 1 | Fix `goconst` warnings — extract repeated test strings to constants    | Medium | Low    |
| 2 | Fix `unused` warning — remove `assertStringEqual` from `html_test.go`  | Low    | Low    |
| 3 | Fix `nolintlint` — remove stale nolint directives                      | Low    | Low    |
| 4 | Fix `errchkjson` — check `json.Marshal` error in `config_enum_test.go` | Low    | Low    |
| 5 | Fix `prealloc` — preallocate `baseLines` in `html_test.go`             | Low    | Low    |
| 6 | Run `gofmt` / `golines` on flagged files                               | Low    | Low    |
| 7 | Add `//nolint:exhaustruct` where needed or fix struct init             | Low    | Low    |

### Medium Effort (1-2 hours each)

| #  | Task                                                              | Impact | Effort |
| -- | ----------------------------------------------------------------- | ------ | ------ |
| 8  | Extract `FilterOptions` struct from `FlagValues` (group 15 bools) | High   | Medium |
| 9  | Extract `FilterOptions` struct from `Config` (reuse same type)    | High   | Medium |
| 10 | Wire `FilterOptions` through `BuildConfigFromFlags`               | High   | Medium |
| 11 | Decompose `StatsData` (23 fields) into focused sub-structs        | Medium | Medium |
| 12 | Add `tparallel` fixes — call `t.Parallel()` in subtests           | Low    | Medium |
| 13 | Fix `wsl_v5` lint warnings across test files                      | Low    | Medium |
| 14 | Fix `nlreturn` lint warnings                                      | Low    | Medium |

### Larger Effort (multi-session)

| #  | Task                                                             | Impact    | Effort |
| -- | ---------------------------------------------------------------- | --------- | ------ |
| 15 | Create `CommonFlags` shared between `FlagValues` and `Config`    | Very High | High   |
| 16 | Refactor `BuildConfigFromFlags` to use `CommonFlags`             | Very High | High   |
| 17 | Update all `FlagValues` callers to use shared types              | High      | High   |
| 18 | Extend phantom types for cache/file params (top 20 by frequency) | Medium    | High   |
| 19 | Review and fix panic conditions (top 50 by severity)             | High      | High   |
| 20 | Decompose `RuntimeConfig` into focused sub-structs               | Medium    | Medium |

### Strategic / Architectural

| #  | Task                                                                            | Impact | Effort    |
| -- | ------------------------------------------------------------------------------- | ------ | --------- |
| 21 | Unify printer clone types around a common `CloneInfo` interface                 | High   | Very High |
| 22 | Create `domain.CloneLocation` type to replace ad-hoc Filename/LineStart/LineEnd | High   | Very High |
| 23 | Evaluate `go-composable-business-types` for branded ID pattern                  | Medium | Medium    |
| 24 | Add `check-pg` (pergola) for structural pattern enforcement                     | Medium | Medium    |
| 25 | Create architecture decision records (ADRs) for major type choices              | Medium | Medium    |

---

## G) Top #1 Question I Cannot Figure Out Myself

**FlagValues ↔ Config convergence strategy:**

`cmd.FlagValues` and `config.Config` share ~20 fields. FlagValues is a CLI flag container (no JSON tags, used by Cobra), while Config is a JSON-serializable configuration object. Three possible approaches:

1. **Embed a shared struct** — Create `config.FilterOptions` with the common bool fields, embed in both. Risk: Cobra flag binding may not work cleanly with embedded struct fields.
2. **Convert FlagValues → Config via method** — Add a `ToConfig() *config.Config` method on FlagValues. Risk: Still duplicates field definitions.
3. **Use Cobra's config binding** — Let Cobra bind flags directly to Config fields. Risk: Requires restructuring how flags are defined.

**Which approach does the project prefer?** This determines whether Phase 5 is a 2-hour or 2-day effort.

---

## Current Project Health

| Metric                               | Value            | Status                   |
| ------------------------------------ | ---------------- | ------------------------ |
| Test packages                        | 24/24 passing    | ✅                       |
| `art-dupl` source duplicates         | 60 (8% of total) | ✅ All legitimate        |
| `art-dupl` test duplicates           | 624 (91%)        | ✅ Expected in BDD       |
| `branching-flow` dupe actionable     | **0** (was 2)    | ✅ Fixed                 |
| `branching-flow` mixin opportunities | 4 (was 6)        | ✅ Improved              |
| `branching-flow` anti-patterns       | 3 warnings       | ⚠️ Large structs          |
| `branching-flow` context score       | 82.9/100         | ⚠️ Mostly false positives |
| `branching-flow` phantom types       | 545              | ⬜ Deferred              |
| `branching-flow` panic conditions    | 532              | ⬜ Deferred              |
| Lint warnings                        | 36               | ⚠️ Pre-existing           |
| Build                                | Clean            | ✅                       |
