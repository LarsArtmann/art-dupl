# Status: AST-Aware False-Positive Filtering

**Date:** 2026-06-11 04:35
**Branch:** fork
**Trigger:** False-positive pattern report from 25-project deduplication sprint
**Scope:** `printer/actionability.go`, `printer/clone_classify.go`, `printer/clone_processor.go`, `domain/processed_clone.go`

---

## Executive Summary

Implemented **4 AST-aware false-positive pattern detectors** that use the existing `syntax.Node` tree structure to automatically classify and suppress the 8 recurring false-positive patterns identified in `docs/feedback/art-dupl-false-positive-report.md`. The report found that **74% of detected clones across 25 Go projects are structural test patterns** — these are now detected and classified as `non-actionable` with specific pattern labels.

All tests green. All vet clean. Printer coverage: 76.7%.

---

## a) FULLY DONE

### New AST Pattern Detectors (`printer/actionability.go`)

| # | Pattern Detector        | AST Signals Used                                                                                                                  | Report Patterns Covered                                                 |
| - | ----------------------- | --------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------- |
| 1 | `isTestDataFilePair`    | Node `Filename` field, path component analysis for `testdata/`                                                                    | Pattern 3 (golden/input file pairs)                                     |
| 2 | `isTableDrivenTestBody` | Root node `RangeStmt` + child `CallExpr(SelectorExpr{Name:"Run"})` in `_test.go`                                                  | Pattern 1 (table-driven test bodies)                                    |
| 3 | `isTestScaffolding`     | `CallExpr(SelectorExpr{Name:"TempDir"})` + assertion calls (`Expect`, `NotTo`, `Equal`, `HaveLen`, `BeEmpty`, etc.) in `_test.go` | Patterns 2, 4, 6 (Ginkgo/standard test boilerplate)                     |
| 4 | `isDataDominated`       | Ratio of `BasicLit` + `KeyValueExpr` nodes across the full tree; triggers at >=60%                                                | Patterns 5, 7, 8 (struct init arrays, config blocks, builder callbacks) |

### Pattern Label System

- `EvaluateActionabilityWithLabel()` returns `(PatternLabel, CloneActionability)` — identifies _which_ pattern was detected
- 7 labels: `PatternNone`, `PatternTestData`, `PatternTableDrivenTest`, `PatternTestScaffolding`, `PatternDataDominated`, `PatternSignatureOnly`, `PatternRAIIDefer`, `PatternErrorPropagation`
- Backward compatible: existing `EvaluateActionability()` delegates to the labeled version and discards the label

### New Domain Categories

- `CategoryTestBoilerplate` ("test-boilerplate") — emoji: 🧹 — for table-driven test bodies and test scaffolding
- `CategoryTestFixture` ("test-fixture") — emoji: 🎯 — for testdata golden/input file pairs
- New suggestions: `suggestTestDataPair`, `suggestTableDrivenTest`, `suggestTestScaffolding`, `suggestDataDominated`

### Classification Pipeline Update

- `clone_processor.go`: Uses `EvaluateActionabilityWithLabel` instead of `EvaluateActionability`
- `applyPatternLabel()`: Upgrades category/priority/suggestion based on detected pattern
- Non-actionable clones get `PriorityLow` and specific suggestions explaining _why_ they're not actionable

### Test Coverage

- **41 new test cases** in `printer/actionability_patterns_test.go` (681 lines)
- Tests for all 4 new detectors + `EvaluateActionabilityWithLabel` + `applyPatternLabel`
- Edge cases: empty sequences, mixed patterns, production files, partial matches
- All existing tests unchanged and passing

### Files Changed

| File                                     | Delta            | Purpose                               |
| ---------------------------------------- | ---------------- | ------------------------------------- |
| `printer/actionability.go`               | +282 lines       | 4 new detectors + label system        |
| `printer/clone_classify.go`              | +32 lines        | New suggestions + `applyPatternLabel` |
| `printer/clone_processor.go`             | +2 lines         | Wire labeled actionability            |
| `domain/processed_clone.go`              | +8 lines         | New categories + emojis               |
| `printer/actionability_patterns_test.go` | +681 lines (new) | Comprehensive test coverage           |

---

## b) PARTIALLY DONE

### Pattern Detection Precision

- **`isTestScaffolding`** currently detects assertion method names via a hardcoded list (`Expect`, `NotTo`, `Equal`, `HaveLen`, etc.). This covers Ginkgo and testify but could miss custom assertion libraries or DSL-like test frameworks. The pattern is _structurally correct_ but the name list is intentionally conservative.
- **`isDataDominated`** uses a 60% threshold. This is calibrated to the report's examples (struct init arrays with 6+ fields) but hasn't been validated against a large corpus of borderline cases. Could produce false negatives on medium-data clones or false positives on data-heavy production code (e.g., config builders).
- **`isTableDrivenTestBody`** requires the root node to be `RangeStmt` with a `t.Run` call inside. This won't catch table-driven tests using `t.Parallel()` without `t.Run`, or tests using Ginkgo's `DescribeTable` directly (those are caught by `isTestScaffolding` instead, via the assertion signal).

---

## c) NOT STARTED

### From the False-Positive Report (Not Yet Implemented)

1. **Config file support** (`.art-dupl.yaml`) — The report's suggestion #2: project-specific exclusions. Currently all filtering is CLI flags only.
2. **Semantic threshold multiplier for test files** — Automatically raising effective threshold when both clones are in `_test.go` files. Would catch patterns that slip through at the detection level rather than post-hoc classification.
3. **CallExpr callee hashing in semantic mode** — `CallExpr` currently doesn't encode the function being called in semantic mode. Adding callee name hashing would make `rule.Check()` vs `os.WriteFile()` distinct even with identical AST structure.
4. **Refactoring suggestion classification** — The report's suggestion #4: classify recommended fix type (extract helper, table-driven test, acceptable similarity, data duplication).
5. **Cross-file test pattern down-ranking** — Pattern 6 specifically mentions cross-file test matches. Currently handled by `isTestScaffolding` but could benefit from a separate weight/ranking for cross-file vs within-file test matches.

### From Existing TODO_LIST.md (Not Started)

6. **ProcessedClone DTO migration** — 111 test call sites still depend on `syntax.Node` internals through `[][]*syntax.Node` in printers. This is the #1 HIGH priority item.
7. **Clone type consolidation** — Three parallel types: `printer.clone`, `pkg/artdupl.Clone`, `printer.CloneGroup`
8. **CSV output using encoding/csv** — Currently manual formatting
9. **TokenValue type with validation**
10. **Enum unification** — domain enums should use config's generic helpers

---

## d) TOTALLY FUCKED UP / RISKS

### Nothing Is Broken

All tests pass. All vet clean. Build green. No regressions.

### Potential Concerns

1. **Pattern ordering matters** — `isSignatureOnlyMatch` runs before `isTestDataFilePair`. If a testdata file has a single-node FuncDecl clone, it gets labeled as `signature-only` instead of `testdata-pair`. This is technically correct (signature-only IS the reason it's non-actionable) but may not match user expectations who see testdata files and expect that label.
2. **`isTestScaffolding` assertion list** — The `walkForTestScaffoldingSignals` switch statement has `Should` as a standalone case to avoid duplicate switch cases. This is fragile — if more Gomega/assertion methods are added, the list needs manual maintenance.
3. **No end-to-end validation** — The new patterns have unit tests with manually constructed AST trees, but haven't been validated against the actual output of running `art-dupl --semantic` on a real Go project. The detectors may fire differently on real serialized ASTs where semantic encoding is active (node types have identifier hashes in upper bits).

---

## e) WHAT WE SHOULD IMPROVE

### Architecture

1. **Thread more data into `ClassifyClone`** — Currently `ClassificationInput` only has `Filename`, `NodeType`, `Tokens`, `Lines`. The per-clone classification is blind to AST structure. The group-level `EvaluateActionability` has full AST access, but the per-clone category/suggestion is set before that runs. Consider adding a `PatternLabel` field to `ClassificationInput` so classification can be pattern-aware from the start.
2. **Extract assertion name detection into a registry** — The hardcoded switch in `walkForTestScaffoldingSignals` should be a `map[string]bool` or similar, making it extensible without modifying the detector.
3. **Make data dominance threshold configurable** — The 60% constant is opinionated. Users with data-heavy codebases (e.g., configuration management tools) may want to tune this.
4. **Add pattern-specific BDD tests** — Current tests use manually constructed AST trees. End-to-end BDD tests that parse real Go files and verify the full pipeline (parse → detect → classify → output) would catch semantic encoding interactions.

### Semantic Mode Gaps

5. **`CallExpr` callee hashing** — The single highest-ROI semantic improvement. Currently `CallExpr` is type-only. Hashing `CallExpr.Fun` (when it's a `SelectorExpr` or `Ident`) would separate `rule.Check()` from `os.WriteFile()` at the detection level, eliminating many false positives before classification even runs.
6. **BasicLit sub-categorization** — `BasicLit` is currently opaque regardless of literal kind (string, int, char). Adding semantic sub-types for `STRING_LIT` vs `INT_LIT` vs `CHAR_LIT` would allow more precise data-dominance detection.

---

## f) Top 25 Things To Do Next

### Tier 1: High Impact, Directly Related (False-Positive Elimination)

| # | Task                                                                                                                                                   | Impact   | Effort |
| - | ------------------------------------------------------------------------------------------------------------------------------------------------------ | -------- | ------ |
| 1 | **Validate new patterns against real projects** — Run `art-dupl -t 45 --semantic --rich-text` on 5+ projects, compare before/after false-positive rate | Critical | 1hr    |
| 2 | **Add BDD end-to-end tests** — Parse real Go files with table-driven tests, Ginkgo patterns, testdata fixtures. Verify full pipeline                   | High     | 2hr    |
| 3 | **CallExpr callee semantic hashing** — Hash `CallExpr.Fun` name in semantic mode. Eliminates cross-function test pattern matches                       | High     | 3hr    |
| 4 | **Extract assertion name registry** — Replace switch in `walkForTestScaffoldingSignals` with extensible map                                            | Medium   | 30min  |
| 5 | **Add `--exclude-testdata` and `--exclude-tests` CLI flags** — Pre-detection filtering for users who want to skip test files entirely                  | Medium   | 1hr    |
| 6 | **Config file support** (`.art-dupl.yaml`) — Project-specific exclusions, thresholds, pattern overrides                                                | High     | 4hr    |

### Tier 2: Architecture Improvements

| #  | Task                                                                                                                          | Impact    | Effort |
| -- | ----------------------------------------------------------------------------------------------------------------------------- | --------- | ------ |
| 7  | **ProcessedClone DTO migration** — Decouple printers from `syntax.Node`. 111 test call sites. #1 HIGH in TODO_LIST.md         | Very High | 8hr    |
| 8  | **Clone type consolidation** — Merge `printer.clone`, `pkg/artdupl.Clone`, `printer.CloneGroup` into unified type             | High      | 4hr    |
| 9  | **Thread PatternLabel into ClassifyClone** — Make per-clone classification pattern-aware, not just group-level                | Medium    | 2hr    |
| 10 | **Test threshold multiplier** — Auto-raise threshold for test-to-test clones at detection time                                | Medium    | 1hr    |
| 11 | **Refactoring suggestion classification** — Classify recommended fix type: extract helper, table-driven, acceptable, data-dup | Medium    | 2hr    |

### Tier 3: Quality & Observability

| #  | Task                                                                                                            | Impact | Effort |
| -- | --------------------------------------------------------------------------------------------------------------- | ------ | ------ |
| 12 | **Pattern detection metrics** — Add stats counters for how many clones each pattern catches                     | Medium | 1hr    |
| 13 | **Printer coverage >85%** — Currently 76.7%. Add tests for new classification paths in HTML/JSON/SARIF printers | Medium | 2hr    |
| 14 | **Fuzz tests for actionability detectors** — Random AST trees, ensure no panics on edge cases                   | Medium | 2hr    |
| 15 | **Update FEATURES.md** — Add actionability pattern detection as a feature                                       | Low    | 30min  |
| 16 | **Update TODO_LIST.md** — Reflect completed work and new items from this session                                | Low    | 30min  |

### Tier 4: Semantic Mode Improvements

| #  | Task                                                                                                                                                     | Impact | Effort |
| -- | -------------------------------------------------------------------------------------------------------------------------------------------------------- | ------ | ------ |
| 17 | **BasicLit sub-categorization** — Separate string/int/char literals for better data-dominance detection                                                  | Medium | 3hr    |
| 18 | **FuncLit semantic hashing** — Currently `FuncLit` is type-only. Hashing function literal signatures would reduce false positives in callback-heavy code | Medium | 2hr    |
| 19 | **Interface method semantic dedup** — Methods declared in interfaces are currently hashed with `~interface~`. Consider per-method-name hashing           | Low    | 2hr    |
| 20 | **CompositeLit type hashing** — Hash the type name in composite literals to separate `types.Module{}` from `types.Dep{}`                                 | Medium | 1hr    |

### Tier 5: Infrastructure

| #  | Task                                                                    | Impact | Effort |
| -- | ----------------------------------------------------------------------- | ------ | ------ |
| 21 | **CSV output using encoding/csv** — Replace manual formatting           | Low    | 1hr    |
| 22 | **Enum unification** — Domain enums use config's generic helpers        | Low    | 2hr    |
| 23 | **TokenValue type validation** — Stronger typing for suffix tree tokens | Low    | 3hr    |
| 24 | **Split printer/stats_test.go** — 975L → 3 files                        | Low    | 1hr    |
| 25 | **Write SDK documentation for pkg/artdupl/**                            | Low    | 2hr    |

---

## g) Top #1 Question I Cannot Answer Myself

**How do the new pattern detectors interact with semantic encoding in practice?**

**VALIDATED:** Ran on `go-structure-linter` and `library-policy` at thresholds 5, 15, 30, 45.

Findings:

- **Semantic mode (`--semantic`, default):** The detectors rarely fire because semantic encoding prevents most test scaffolding clones from being detected in the first place. Identifiers like `err`, `issues`, `tmpDir` have different semantic hashes across test functions, so the suffix tree never produces a match. This is correct — semantic mode is already excellent at filtering test noise.
- **Structural mode (`--structural`):** The detectors fire correctly and catch significant numbers of false positives. On `go-structure-linter`: 10+ groups caught as `test-boilerplate`. On `library-policy`: 5 groups caught as `test-boilerplate`, 4 as `data-dominated`. All correctly labeled `non-actionable`.
- **Both modes:** The `isTestDataFilePair` and `isTableDrivenTestBody` detectors are mode-independent (they check filenames and AST structure, not node types), so they work equally well in both modes.

The key insight: **semantic mode is the first line of defense** (prevents detection of test clones), and **the new patterns are the second line** (classifies remaining clones as non-actionable). They're complementary.

---

## Test Results

```
go test ./...          → ALL GREEN (25 packages)
go vet ./...           → CLEAN
printer coverage       → 76.7%
new test cases         → 41
total files changed    → 5 (4 modified + 1 new)
total lines added      → ~333 (net)
```
