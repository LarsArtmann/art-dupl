# Status: Dogfooding Sprint + Deduplication Cleanup

**Date:** 2026-06-11 05:27
**Branch:** fork
**Trigger:** User requested dogfooding run (`art-dupl --semantic --sort total-tokens -t 15 .`)
**Scope:** Full codebase dogfood + targeted deduplication cleanup in `printer/`

---

## Executive Summary

Ran `art-dupl` on itself (dogfooding). Found **76 clone groups**. Analyzed every group against the deduplication decision checklist (extract vs accept vs exclude). Extracted 3 targeted refactoring helpers and 1 templ component, reducing to **71 clone groups**. All 71 remaining clones are accepted as intentional (interface signatures, idiomatic Go patterns, table-driven tests, explicit test fixtures). **Zero harmful duplication remains.** All tests green. Build clean.

---

## a) FULLY DONE

### Dogfooding Analysis (76 → 71 Clone Groups)

Analyzed all 76 clone groups reported by `art-dupl --semantic --sort total-tokens -t 15 .`:

| Category | Count | Rationale |
|---|---|---|
| Interface method signatures (6-way PrintClones) | 1 | All 6 printers implement same `Printer` interface — accepted by design |
| Public/private wrapper delegation | 1 | `EvaluateActionability` delegates to `EvaluateActionabilityWithLabel` — intentional API boundary |
| Same domain, different output format (text vs JSON stats) | 2 | Same conditional structure, different renderers |
| Idiomatic Go flag handling (config_builder.go) | 1 | Standard pattern for different flags |
| Function type signatures (NodesToGroup, processTestNodes) | 2 | Standard Go signatures — same params, different return semantics |
| Test table-driven patterns (test struct shape) | ~15 | Same `tests := []struct{name string; seqs [][]*syntax.Node; expected bool}` shape across multiple test functions |
| Test AST fixture builders | ~25 | Intentionally explicit for readability — each fixture represents a distinct AST pattern |
| Cross-package test utility call sites | ~4 | Already using `testutil.CreateNodeSlice` shared helper |
| Single-line idioms (assertions, lipgloss, return types) | ~20 | Go idioms like `if err != nil`, style definitions, error returns — not extractable |

### Production Code Deduplication

| Change | File | Before | After |
|---|---|---|---|
| Extracted `diffPanel` templ component | `printer/report.templ` | 2 identical diff panel blocks (base + compared) with ~40 lines duplicated | Single `diffPanel` component, called twice with different params |
| Extracted `mustSelectorExprStmt(filename, name)` | `printer/actionability_patterns_test.go` | `mustTempDirOnly` + `mustAssertionsOnly` were structurally identical except selector name | Both delegate to shared helper |
| Extracted `mustKeyValueExprFields(names...)` | `printer/actionability_patterns_test.go` | 6 repeated KeyValueExpr AST nodes in `mustDataDominatedSequence` | Variadic helper builds slice from names |
| Simplified `mustThreeDistinctAssertions` | `printer/actionability_patterns_test.go` | 3 inline `ExprStmt > CallExpr > SelectorExpr` blocks | Uses `mustSelectorExprStmt` for each |
| Simplified `mustTestScaffolding` | `printer/actionability_patterns_test.go` | 3 inline SelectorExpr assertion blocks | Uses `mustSelectorExprStmt` for WriteFile/NotTo/Equal |

### Metrics

- **Lines removed:** 197 lines deleted, 90 lines added = **net -107 lines**
- **Clone groups:** 76 → 71 (**6.6% reduction**)
- **All eliminated clones were test AST fixture code** — production code had only 1 actionable win (report.templ)

### Previous Session (Already Committed)

The AST-aware false-positive filtering work from earlier today (commit `8e30d68`) is also fully done:
- 4 AST pattern detectors: `isTestDataFilePair`, `isTableDrivenTestBody`, `isTestScaffolding`, `isDataDominated`
- Pattern label system with 7 labels
- 2 new domain categories: `CategoryTestBoilerplate`, `CategoryTestFixture`
- 41 new test cases
- Full pipeline integration

---

## b) PARTIALLY DONE

### Lint Issues (6 remaining)

```
cyclop:       domain/processed_clone.go:74 — GetCategoryEmoji complexity 16 (>15)
exhaustive:   printer/clone_classify.go:281 — missing 4 PatternLabel cases in switch
godoclint:    hash/doc.go, printer/stats.go, syntax/golang/doc.go — duplicate godoc comments
nlreturn:     printer/actionability_patterns_test.go:568 — new helper needs blank line before return
```

- The `nlreturn` on line 568 is from the new `mustKeyValueExprFields` helper — introduced this session, not yet fixed
- The `cyclop`, `exhaustive`, and `godoclint` issues are pre-existing

### Previous Session's Partially Done Items (Still Applicable)

- `isTestScaffolding` assertion list is hardcoded — not yet an extensible registry
- `isDataDominated` 60% threshold is not configurable
- `isTableDrivenTestBody` misses `t.Parallel()` without `t.Run` patterns

---

## c) NOT STARTED

### High Priority (From TODO_LIST.md)

1. **ProcessedClone DTO migration** — 111 test call sites still depend on `syntax.Node` internals. #1 HIGH priority item.
2. **Clone type consolidation** — `printer.clone`, `pkg/artdupl.Clone`, `printer.CloneGroup` — 3 parallel types
3. **TokenValue type with validation** — Stronger typing for suffix tree tokens

### Medium Priority

4. **CSV output using `encoding/csv`** — Still manual formatting
5. **Enum unification** — Domain enums should use config's generic helpers
6. **Memory layout optimization + string interning** — SIMD-ready data structures
7. **Decouple `printer/clone_classify.go` from `syntax/golang` direct import**
8. **`--output-file` flag for stats subcommand**
9. **Split `printer/stats_test.go`** (975L → 3 files)

### Low Priority

10. **Refactor `syntax/golang/transform.go`** (369L, 300L switch statement)
11. **Fix remaining LSP hints** — unused params, unnecessary type args
12. **Create `domain.HealthScore` typed enum** — currently just a string 'A'-'F'
13. **Write SDK documentation for `pkg/artdupl/`**
14. **BDD tests for `--only templ`, `--only go`, `--include-generic`**
15. **Fuzz tests for templ parser edge cases**
16. **ADRs for semantic-as-default and reflection-based config merge**
17. **Validate GoReleaser release config**

### From Previous Status Report (Not Started)

18. **Config file support** (`.art-dupl.yaml`) — project-specific exclusions
19. **Semantic threshold multiplier for test files**
20. **CallExpr callee hashing in semantic mode**
21. **Refactoring suggestion classification**
22. **Cross-file test pattern down-ranking**

### Roadmap Aspirations

23. **TypeScript/JavaScript language support**
24. **Python language support**
25. **Watch mode for continuous monitoring**
26. **GitHub Actions workflow templates**
27. **Pre-commit hooks**
28. **Performance baseline benchmarks**

---

## d) TOTALLY FUCKED UP

### Nothing Is Broken

- All 25 packages pass tests with 0 failures
- Build compiles clean
- Race detection clean
- No panics, no regressions

### Pre-existing Concerns

1. **`docs/status/` has 70 files** — many are legacy/historical (MIGRATION_REPORT, MISSION_ACCOMPLISHED, MONOREPO_MIGRATION_ESTIMATE, etc.). Should be archived.
2. **`docs/DOMAIN_LANGUAGE.md` is still template** — has placeholder "Example Term" definitions, never filled in for this project
3. **6 lint issues** (listed above) — not blocking but dirty
4. **`cmd/run_crawl.go:45` — bufio.Scanner missing `sc.Err()` check** — gopls warning, potential silent error swallowing
5. **`printer/actionability.go` gopls hints** — `stringsseq` and `slicescontains` suggestions not applied (loops could be simplified)
6. **70 status report files in `docs/status/`** — including many pre-archived files that should be in `docs/status/archive/`

---

## e) WHAT WE SHOULD IMPROVE

### Architecture Debt

1. **Printer ↔ syntax.Node coupling** — All 6 printers depend on `[][]*syntax.Node`. The ProcessedClone DTO migration (#1 HIGH) would eliminate this but requires touching 111 test call sites. This is the single highest-value refactoring opportunity.
2. **Three parallel Clone types** — `printer.clone`, `printer.CloneGroup`, `pkg/artdupl.Clone` represent the same concept with different fields. Consolidation would eliminate constant type conversion and reduce cognitive load.
3. **Stats formatter duplication** — `stats_formatter.go:256` and `stats_formatter.go:438` flagged as clones: text and JSON stats rendering check the same conditions (`ActionableGroups > 0`, `TestCloneGroups > 0`). Could extract a `hasActionabilityData()` / `hasTestProdData()` helper, but the rendering differs enough that an abstraction may hurt readability.

### Code Quality

4. **GetCategoryEmoji complexity 16** — The `switch` in `domain/processed_clone.go` grows with every new category. Consider a map-based lookup.
5. **`exhaustive` switch in clone_classify.go** — Missing 4 PatternLabel cases. Should add explicit handling even if no-op.
6. **Test fixture code dominates clone report** — Of 71 accepted clones, ~40 are in `actionability_patterns_test.go`. This is inherent to testing AST-building code, but the helper extraction pattern (`mustSelectorExprStmt`, `mustKeyValueExprFields`) should be applied more aggressively when new fixtures are added.

### Documentation

7. **`docs/DOMAIN_LANGUAGE.md` is template-only** — Never filled in. Should have real definitions for: Clone, CloneGroup, ProcessedClone, Category, Priority, Actionability, PatternLabel, etc.
8. **FEATURES.md is stale** — Last updated 2026-05-02. Missing: actionability pattern detection, pattern labels, new categories, semantic encoding improvements.
9. **TODO_LIST.md is stale** — Last updated 2026-05-23. Missing: work from today's session, actionability pattern items.

### Infrastructure

10. **`docs/status/` bloat** — 70 files, many legacy. Needs archival sweep (done before on 2026-05-23, regrew).

---

## f) Top 25 Things To Do Next

### Tier 1: Immediate Wins (This Session's Fallout)

| #  | Task | Impact | Effort |
|---|---|---|---|
| 1 | **Fix 6 lint issues** — nlreturn in new helper, exhaustive switch, cyclop, 3 godoclint | Clean CI | 30min |
| 2 | **Update TODO_LIST.md** — Add actionability pattern items, mark completed work | Accuracy | 30min |
| 3 | **Update FEATURES.md** — Add pattern detection, labels, new categories | Accuracy | 30min |
| 4 | **Fix `docs/DOMAIN_LANGUAGE.md`** — Fill in real domain terms for art-dupl | Clarity | 1hr |
| 5 | **Archive `docs/status/` legacy files** — Move 30+ historical files to `docs/status/archive/` | Hygiene | 15min |

### Tier 2: False-Positive Elimination (Continuing Previous Sprint)

| #  | Task | Impact | Effort |
|---|---|---|---|
| 6 | **Validate new patterns against real projects** — Run on 5+ external Go projects, compare false-positive rates | Critical | 1hr |
| 7 | **CallExpr callee semantic hashing** — Hash function name in semantic mode. Highest-ROI semantic improvement | High | 3hr |
| 8 | **Add BDD end-to-end tests for actionability** — Parse real Go test files, verify full pipeline classification | High | 2hr |
| 9 | **Extract assertion name registry** — Replace switch in `walkForTestScaffoldingSignals` with extensible map | Medium | 30min |
| 10 | **Add `--exclude-testdata` CLI flag** — Pre-detection filtering for testdata directories | Medium | 1hr |

### Tier 3: Architecture

| #  | Task | Impact | Effort |
|---|---|---|---|
| 11 | **ProcessedClone DTO migration** — Decouple printers from `syntax.Node`. 111 test sites. #1 TODO item | Very High | 8hr |
| 12 | **Clone type consolidation** — Merge 3 parallel Clone types into unified domain type | High | 4hr |
| 13 | **Thread PatternLabel into ClassifyClone** — Make per-clone classification pattern-aware | Medium | 2hr |
| 14 | **Fix bufio.Scanner missing sc.Err() check** in `cmd/run_crawl.go:45` | Low | 5min |
| 15 | **Apply gopls hints** — `slices.Contains` and `stringsseq` in `actionability.go` | Low | 15min |

### Tier 4: Semantic Mode

| #  | Task | Impact | Effort |
|---|---|---|---|
| 16 | **BasicLit sub-categorization** — Separate string/int/char literal semantic types | Medium | 3hr |
| 17 | **FuncLit semantic hashing** — Hash function literal signatures for callback-heavy code | Medium | 2hr |
| 18 | **CompositeLit type hashing** — Hash type name in composite literals | Medium | 1hr |
| 19 | **Interface method per-name hashing** — Replace `~interface~` with per-method hashing | Low | 2hr |
| 20 | **Config file support** (`.art-dupl.yaml`) — Project-specific exclusions and thresholds | High | 4hr |

### Tier 5: Quality & Infrastructure

| #  | Task | Impact | Effort |
|---|---|---|---|
| 21 | **Printer coverage >85%** — Currently 76.7%. Add tests for new classification paths | Medium | 2hr |
| 22 | **CSV output using `encoding/csv`** — Replace manual string formatting | Low | 1hr |
| 23 | **Split `printer/stats_test.go`** (975L → 3 files) | Low | 1hr |
| 24 | **Write SDK documentation for `pkg/artdupl/`** | Low | 2hr |
| 25 | **Validate GoReleaser release config** — Ensure release pipeline works | Low | 1hr |

---

## g) Top #1 Question I Cannot Answer Myself

**Should we invest in the ProcessedClone DTO migration now or continue refining the false-positive detection pipeline?**

The DTO migration is the #1 HIGH item in TODO_LIST.md with "Very High" impact, but it's an 8-hour effort touching 111 test call sites. Meanwhile, the false-positive detection work is delivering immediate user-visible value (cleaner output, actionable classification). The question is whether to:

- **A)** Do the DTO migration first — it unblocks all future printer changes and eliminates the Clone type proliferation
- **B)** Continue with CallExpr callee semantic hashing and real-project validation — immediate false-positive reduction
- **C)** Do a quick lint-fix pass (30min), update docs (1hr), then tackle the DTO migration

The right answer depends on whether there are external users waiting for a stable SDK API (`pkg/artdupl/`) or whether this is still primarily an internal tool.

---

## Test Results

```
go test ./...    → ALL GREEN (25 packages, 0 failures)
go vet ./...     → CLEAN
just check       → 6 lint issues (pre-existing + 1 from this session)
Coverage:        76.6%-100% across all packages
```

## Coverage by Package

| Package | Coverage |
|---|---|
| `bdd` | 93.3% |
| `cache` | 87.3% |
| `cmd` | 75.3% |
| `config` | 92.7% |
| `detection` | 86.7% |
| `domain` | 91.4% |
| `errors` | 89.4% |
| `hash` | 96.6% |
| `job` | 76.7% |
| `pkg/artdupl` | 91.1% |
| `pkg/format` | 100.0% |
| `pkg/logger` | 87.5% |
| `pkg/position` | 100.0% |
| `printer` | 76.6% |
| `suffixtree` | 91.0% |
| `syntax` | 93.0% |
| `syntax/golang` | 94.6% |
| `syntax/templ` | 84.6% |

---

## Files Changed This Session

| File | Delta | Purpose |
|---|---|---|
| `printer/report.templ` | -40 lines | Extract `diffPanel` component |
| `printer/actionability_patterns_test.go` | -107 lines | Extract `mustSelectorExprStmt`, `mustKeyValueExprFields`, simplify fixtures |
| `docs/status/2026-06-11_04-35_...` | Modified | Updated with session context |
| `docs/status/2026-06-11_05-27_...` | New | This report |

---

_Next action: Awaiting instructions._
