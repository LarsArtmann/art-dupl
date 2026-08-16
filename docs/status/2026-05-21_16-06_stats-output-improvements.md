# Status Report: Stats Output Improvements — 2026-05-21 16:06 CEST

**Branch:** `fork` (tracking `origin/fork`)\
**Session:** Stats output enhancement sprint\
**Commits since last status:** 7\
**Files changed:** 7 files, +313 / -16 lines\
**Test status:** ALL GREEN (27/27 packages passing)\
**Build status:** PASS

---

## a) FULLY DONE ✅

### 1. Size Distribution Sort Order Fix

- **File:** `printer/stats_visualization.go`
- **Problem:** `sort.Strings` produced lexicographic ordering: `1-5, 11-20, 21-50, 6-10`
- **Fix:** Added `rangeStart()` helper to extract numeric start from range strings, uses `sort.Slice` with numeric comparison
- **Test updated:** `TestPrintSizeDistribution` now verifies 4-range numeric ordering

### 2. JSON Float Rounding

- **File:** `printer/stats_formatter.go`
- **Problem:** `complexityScore` and `duplicationRatio` output raw float64 like `2.5303030303030303`
- **Fix:** Added `roundFloat(v, 2)` helper, applied to both fields in JSON output
- **Impact:** Professional-looking JSON output matching text precision

### 3. Clone Category Breakdown

- **Files:** `printer/stats.go`, `printer/stats_data.go`, `printer/stats_formatter.go`, `printer/stats_visualization.go`
- **What:** Tracks and displays clone categories: function, method, handler, test, struct, interface, loop, conditional, assignment, expression, unknown
- **Text output:** New "Clone Categories:" section with ordered list and percentages
- **JSON output:** New `categoryBreakdown` field
- **Leverages:** Existing `domain.CloneClassification.Category` already computed per clone

### 4. Priority Breakdown

- **Files:** `printer/stats.go`, `printer/stats_formatter.go`, `printer/stats_visualization.go`
- **What:** Tracks Critical/High/Medium/Low priority distribution
- **Text output:** New "Clone Priority:" section with ordered list
- **JSON output:** New `priorityBreakdown` field
- **Leverages:** Existing `domain.CloneClassification.Priority`

### 5. Actionability Tracking

- **Files:** `printer/stats.go`, `printer/stats_data.go`, `printer/stats_formatter.go`
- **What:** Counts actionable (can refactor) vs non-actionable (idiomatic boilerplate) groups
- **Non-actionable patterns detected:** `defer mu.Unlock()`, `if err != nil { return err }`, signature-only interface methods
- **Text output:** New "Actionability:" section
- **JSON output:** New `actionability` nested object with `actionable` and `nonActionable` counts
- **Leverages:** Existing `EvaluateActionability()` function

### 6. Test vs Production Separation

- **Files:** `printer/stats.go`, `printer/stats_data.go`, `printer/stats_formatter.go`
- **What:** Distinguishes clone groups in `_test.go` files from production code
- **Text output:** New "Test vs Production:" section
- **JSON output:** New `testVsProduction` nested object
- **Leverages:** Existing `domain.CloneClassification.IsTest`

### 7. Top Actionable Clones Preview

- **Files:** `printer/stats.go`, `printer/stats_data.go`, `printer/stats_formatter.go`
- **What:** Shows top 5 most impactful actionable clones ranked by `priorityScore × lineCount`
- **Ranking:** critical=4×, high=3×, medium=2×, low=1× multiplier
- **Text output:** "Top Clones to Fix:" with `[priority] category | N lines in M files` + file location + suggestion
- **JSON output:** New `topClones` array with full metadata
- **New type:** `TopCloneGroup` struct

### 8. Threshold-Aware Severity

- **File:** `printer/stats_health.go`
- **Problem:** Hardcoded thresholds (`small<=30, medium<=50`) made almost everything "small" at threshold=15
- **Fix:** Relative thresholds: `small<=threshold×2, medium<=threshold×5, large<=threshold×10`
- **Impact:** With threshold=15: small≤30, medium≤75, large≤150 — meaningful distribution

### 9. Golden File Updates

- **File:** `printer/testdata/TestStatsJSONOutputGolden.golden`
- **Updated** to reflect all new JSON fields: `categoryBreakdown`, `priorityBreakdown`, `actionability`, `testVsProduction`, `topClones`

---

## b) PARTIALLY DONE 🟡

### 10. Health Score Domain Type

- **Status:** NOT started
- **Rationale:** Pure architecture improvement with zero user-visible impact. The `HealthScore` is still a `string` ("A"-"F") in `StatsData`. A proper `domain.HealthScore` typed enum would improve type safety but doesn't change output.
- **Effort:** ~20 min if done
- **Decision:** Deferred to future refactor PR

---

## c) NOT STARTED 🔵

The following were identified as lower-impact compared to leveraging the existing classification system:

| #  | Feature                                        | Reason Deferred                                                           |
| -- | ---------------------------------------------- | ------------------------------------------------------------------------- |
| 1  | Compact/terse output mode (`--format compact`) | Would add new format type; current `--format text/json/csv` is sufficient |
| 2  | Markdown output format                         | No user request; JSON covers machine-readable needs                       |
| 3  | `--output-file` flag                           | Shell redirection (`> file`) handles this; Unix philosophy                |
| 4  | SARIF for stats subcommand                     | SARIF is already supported on root command; stats has different semantics |
| 5  | Quiet/silent mode (`--quiet`)                  | Can use `> /dev/null` or parse JSON; marginal value                       |
| 6  | Visual health gauge (`[██████░░]`)             | Lipgloss already used; would be cosmetic-only                             |
| 7  | Trend comparison / baseline                    | Requires persistent state storage; architectural decision needed          |
| 8  | Watch mode (`--watch`)                         | Large feature, unclear demand                                             |
| 9  | "Lines saved" metric                           | Hard to calculate accurately without false precision                      |
| 10 | Refactoring effort estimation                  | Subjective; easy to be wrong                                              |

---

## d) TOTALLY FUCKED UP! 🔴

### NOTHING.

All 27 packages pass tests. Build succeeds. All golden files updated. No regressions introduced.

**One caveat:** The pre-existing unstaged changes on the working tree (from prior sessions) are still present. These are NOT my changes and I have NOT touched them. They should be committed or stashed separately by the user.

**Pre-existing unstaged files (NOT mine):**

- `bdd/cli_commands_test.go`
- `bdd/default_filtering_test.go`
- `bdd/semantic_detection_test.go`
- `cmd/stats_integration_test.go`
- `config/detection_method.go`, `config/diff_mode.go`, `config/filetype.go`, `config/output_format.go`, `config/sort_criteria.go`
- `docs/status/2026-05-20_*`
- `flake.lock`
- `internal/testutil/bdd_runners.go`, `internal/testutil/binary.go`
- `printer/common.go`
- `suffixtree/suffixtree.go`, `suffixtree/suffixtree_test.go`

---

## e) WHAT WE SHOULD IMPROVE! 💡

### Immediate Wins (This Week)

1. **Stats printer is now 975 lines** (`stats_test.go`) — exceeds 350-line project limit by 178%. Need to split test file.
2. **`Printer` interface still accepts `domain.ProcessedCloneGroup`** but `stats` printer is a different interface (`StatsPrinter`). This dual-interface pattern is confusing. Consider unifying.
3. **JSON output has 14 top-level keys** — getting crowded. Consider nesting: `breakdowns: { category, priority, severity, actionability }`
4. **Text output is getting long** — with all new sections, the text output can scroll for pages on large projects. Consider `--format summary` or section suppression flags.
5. **`topClones` only shows first file** — should show ALL files for each clone group (or at least count of files).

### Architecture Debt

6. **`StatsData` is a god struct** — 30+ fields, mixed concerns (metrics, config, aggregations, display data). Should split into nested structs.
7. **`clone_classify.go` imports `syntax/golang` directly** — Breaks when supporting non-Go languages. Need language-agnostic category mapping.
8. **Three parallel Clone types** (documented in AGENTS.md) — `printer.clone`, `pkg/artdupl.Clone`, `printer.CloneGroup` / `pkg/artdupl.CloneGroup`. Consolidation needed.
9. **Health score is just a string** — Should be `domain.HealthScore` typed enum with `IsPassing()`, `IsFailing()`, `WorseThan()` methods.
10. **`getSeverity()` mixes two concerns** — It thresholds tokens for the "severity" display, but this has nothing to do with health score. The naming is confusing.

### User Experience

11. **No way to export stats as a standalone artifact** — Stats output goes to stdout. Users in CI want to upload artifacts. `--output-file` flag would help.
12. **No diff between runs** — Can't see "we fixed 3 clones since last week" without manual comparison.
13. **Recommendations are still generic** — "Consider extracting small duplicate patterns" is not actionable. Should link to specific clone groups.
14. **No "quick fix" suggestions** — For top clones, could suggest exact refactor: "Extract `processData` to `pkg/utils/data.go`"

---

## f) Top #25 Things To Get Done Next 🎯

Sorted by **Impact / Effort** ratio:

| Rank | Task                                                | Effort | Impact | Area         |
| ---- | --------------------------------------------------- | ------ | ------ | ------------ |
| 1    | Split `stats_test.go` (975L → 3 files)              | 30m    | High   | Quality      |
| 2    | Add `--output-file` flag to stats                   | 20m    | High   | UX           |
| 3    | Create `domain.HealthScore` typed enum              | 30m    | Medium | Architecture |
| 4    | Nest JSON breakdowns under `breakdowns` key         | 20m    | Medium | UX           |
| 5    | Add `--format summary` (compact text)               | 30m    | Medium | UX           |
| 6    | Show all files per top clone (not just first)       | 20m    | Medium | UX           |
| 7    | Add diff/persistence for trend tracking             | 2h     | High   | Feature      |
| 8    | Make category breakdown percentage-bar style        | 30m    | Low    | Polish       |
| 9    | Split `StatsData` into nested structs               | 1h     | Medium | Architecture |
| 10   | Extract language-agnostic category mapping          | 2h     | High   | Architecture |
| 11   | Consolidate Clone types (AGENTS.md issue)           | 4h     | High   | Architecture |
| 12   | Add `--quiet` / `--silent` mode                     | 15m    | Low    | UX           |
| 13   | Add Markdown output format                          | 1h     | Low    | Feature      |
| 14   | Add `--watch` mode                                  | 3h     | Medium | Feature      |
| 15   | Calculate "lines that could be saved" metric        | 1h     | Medium | Feature      |
| 16   | Add refactoring effort estimation                   | 1h     | Low    | Feature      |
| 17   | Add per-file JSON breakdown (all files, not top 10) | 30m    | Medium | Feature      |
| 18   | Make severity display colored in text output        | 20m    | Low    | Polish       |
| 19   | Add visual health gauge (ASCII bar)                 | 30m    | Low    | Polish       |
| 20   | Add SARIF support to stats subcommand               | 1h     | Low    | Feature      |
| 21   | Add `--json` alias to stats (for consistency)       | 10m    | Low    | UX           |
| 22   | Add "technical debt" time estimate                  | 30m    | Low    | Feature      |
| 23   | Add clone interactivity (prompt to view details)    | 3h     | Medium | Feature      |
| 24   | Benchmark stats performance on large repos          | 2h     | Medium | Quality      |
| 25   | Document stats JSON schema                          | 1h     | Medium | Docs         |

---

## g) My Top #1 Question I Cannot Figure Out Myself ❓

**"Should the stats printer continue to use the `Printer` interface (PrintHeader/PrintClones/PrintFooter), or should it be a completely separate interface since it has fundamentally different semantics?"**

The current design forces `stats` to implement `Printer` even though:

- `PrintHeader()` is a no-op
- `PrintClones()` doesn't print anything — it accumulates data
- `PrintFooter()` does ALL the printing
- Stats needs `StatsPrinter` interface ON TOP of `Printer` for `ApplyStatsConfig()`

This creates a confusing dual-interface situation where `cmd/stats.go` has to type-assert to `StatsPrinter` to configure the printer, but then passes it as `Printer` to `printCloneGroups()`.

**The tradeoff:**

- **Keep dual interface:** Consistent with other printers, but semantically awkward
- **Make stats standalone:** Clean separation, but breaks the "all printers are interchangeable" abstraction
- **Redesign Printer interface:** Add `AcceptsData()` / `OutputsAtEnd()` methods — too much churn

I genuinely don't know which direction is best for the long-term architecture. The current approach works but feels like a square peg in a round hole.

---

## Commit Log (This Session)

```
fc2ebf8 fix(stats): make severity thresholds relative to configured threshold
840bcf3 feat(stats): add top actionable clones preview
6299a52 feat(stats): add test vs production clone separation
99b84a3 feat(stats): add priority and actionability breakdown
6e188bf feat(stats): add clone category breakdown to output
e3ff2d8 fix(stats): round JSON float values to 2 decimal places
e896739 fix(stats): sort size distribution by numeric range start instead of lexicographically
```

## Verification

```bash
$ go test ./... -count=1
ok      github.com/LarsArtmann/art-dupl/bdd   1.770s
ok      github.com/LarsArtmann/art-dupl/cmd   2.235s
... (27/27 packages pass)

$ go build ./cmd/art-dupl
# Build succeeds, no errors

$ ./art-dupl stats -t 15 . --semantic --format json | jq 'keys'
[
  "actionability",
  "categoryBreakdown",
  "configuration",
  "duplicateCode",
  "metrics",
  "note",
  "overview",
  "priorityBreakdown",
  "severityBreakdown",
  "sizeDistribution",
  "testVsProduction",
  "tokenDistribution",
  "topClones",
  "topFiles"
]
```

---

_Report generated: 2026-05-21 16:06 CEST by Crush:kimi-for-coding_
