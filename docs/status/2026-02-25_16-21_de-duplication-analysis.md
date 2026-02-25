# De-duplication Analysis Report

**Date:** 2026-02-25 16:21
**Author:** Crush (AI Assistant)
**Status:** Planning Complete - Ready for Execution

---

## Executive Summary

Ran `art-dupl --semantic -t 15 --sort total-tokens` to analyze the codebase for code duplication. Found **251 clone groups** with actionable de-duplication opportunities concentrated in three main areas:

| Area | Impact | Effort | Priority |
|------|--------|--------|----------|
| Printer helpers | HIGH | LOW | 🔴 Critical |
| Templ transforms | MEDIUM | MEDIUM | 🟡 Medium |
| BDD tests | MEDIUM | MEDIUM | 🟢 Low |

---

## Current State

### Command Executed
```bash
art-dupl --semantic -t 15 --sort total-tokens
art-dupl --semantic -t 15 --sort total-tokens --html
```

### Results Summary
- **Total clone groups:** 251
- **Largest clone group:** 35 clones (printer/stats files)
- **Most widespread:** 24 clones across BDD tests
- **Primary duplication patterns:** 3 major categories identified

---

## Key Duplication Patterns

### 1. Printer Output Formatting (CRITICAL)

**Location:** `printer/stats_recommendations.go`, `printer/stats_formatter.go`

**Pattern:**
```go
_, _ = fmt.Fprintf(p.w, "  %s\n", p.base.Render("..."))
```

**Occurrences:** 35+ duplicate lines

**Files affected:**
- `printer/stats_recommendations.go:10-57`
- `printer/stats_formatter.go:96-173`

**Impact:** HIGH - Core output formatting, affects all stats output
**Effort:** LOW - Simple helper extraction
**Lines saved:** ~80+

**Recommended solution:**
```go
// Add to printer/stats.go
func (p *stats) printLine(format string, args ...any) {
    _, _ = fmt.Fprintf(p.w, "  %s\n", p.base.Render(fmt.Sprintf(format, args...)))
}

func (p *stats) printSection(title string) {
    _, _ = fmt.Fprintf(p.w, "%s\n", p.section.Render(title))
}

func (p *stats) printMetric(label, value string) {
    _, _ = fmt.Fprintf(p.w, "  %s %s\n", p.metric.Render(label+":"), p.base.Render(value))
}
```

---

### 2. Templ Transform Node Creation (MEDIUM)

**Location:** `syntax/templ/transform*.go` (6 files)

**Pattern:**
```go
if tmpl == nil {
    return nil
}
o := syntax.NewNode()
o.Type = <TYPE>
o.Filename = t.filename
o.Pos = int32(...Range.From.Index)
o.End = int32(...Range.To.Index)
```

**Occurrences:** 11+ duplicate blocks

**Files affected:**
- `syntax/templ/transform.go:67-72`
- `syntax/templ/transform_components.go:22-27, 75-80, 143-148`
- `syntax/templ/transform_expressions.go:22-27, 31-36, 40-45, 63-68, 105-110`
- `syntax/templ/transform_node.go:72-77, 80-85, 106-123`

**Impact:** MEDIUM - Templ parsing support
**Effort:** MEDIUM - Careful refactoring needed
**Lines saved:** ~40+

**Recommended solution:**
```go
// Add to syntax/templ/transform.go
func (t *transformer) createNode(nodeType syntax.NodeType, start, end int) *syntax.Node {
    return &syntax.Node{
        Type:     nodeType,
        Filename: t.filename,
        Pos:      int32(start),
        End:      int32(end),
    }
}
```

---

### 3. BDD Test Setup Patterns (LOW)

**Location:** `bdd/*.go` (15+ files)

**Pattern:** Ginkgo `BeforeEach` and `It` blocks with similar setup

**Occurrences:** 21+ similar blocks

**Files affected:**
- `bdd/bdd_test.go:382-390`
- `bdd/default_filtering_test.go:79-87, 372-380`
- `bdd/detection_methods_test.go:16-24`
- `bdd/error_handling_test.go:48-56`
- `bdd/filter_features_test.go:28-36`
- `bdd/incremental_detection_test.go:22-30, 324-332`
- `bdd/plumbing_and_paths_test.go:32-40, 163-171, 302-310`
- `bdd/plumbing_output_test.go:30-38, 318-326, 520-528`
- `bdd/semantic_detection_test.go:27-35`
- `bdd/sorting_test.go:27-35`
- `bdd/stats_command_test.go:108-116, 421-429`
- `bdd/stats_subcommand_test.go:29-37, 353-361`
- `bdd/templ_clone_detection_test.go:21-29`

**Impact:** MEDIUM - Test maintainability
**Effort:** MEDIUM - Requires test refactoring
**Lines saved:** ~60+

**Note:** Test duplication is less critical than production code. Consider if refactoring provides enough value.

---

## Other Notable Duplications

### Test Helper Patterns
- `t.Parallel()` + temp dir setup: 30+ occurrences
- `GinkgoT()` helper calls: 33+ occurrences
- Table test patterns: Multiple similar structures

### Utility Duplications
- Error handling patterns in `detection/detection_test.go`
- File cache test patterns in `cache/file_cache_test.go`
- Config test patterns in `config/config_test.go`

---

## Comprehensive De-duplication Plan

### Phase 1: Critical (Printer Helpers)

| # | Task | Time | Status |
|---|------|------|--------|
| 1.1 | Create `printLine()` helper in `printer/stats.go` | 5min | Pending |
| 1.2 | Create `printSection()` helper | 3min | Pending |
| 1.3 | Create `printMetric()` helper | 3min | Pending |
| 1.4 | Refactor `stats_recommendations.go` | 10min | Pending |
| 1.5 | Refactor `stats_formatter.go` | 12min | Pending |
| 1.6 | Run tests to verify | 5min | Pending |

**Total time:** ~40min
**Lines saved:** ~80+
**Risk:** LOW

### Phase 2: Medium Priority (Templ Transforms)

| # | Task | Time | Status |
|---|------|------|--------|
| 2.1 | Create `createNode()` helper | 5min | Pending |
| 2.2 | Refactor `transform.go` | 5min | Pending |
| 2.3 | Refactor `transform_components.go` | 8min | Pending |
| 2.4 | Refactor `transform_expressions.go` | 10min | Pending |
| 2.5 | Refactor `transform_node.go` | 8min | Pending |
| 2.6 | Run tests to verify | 5min | Pending |

**Total time:** ~40min
**Lines saved:** ~40+
**Risk:** MEDIUM (templ parsing is sensitive)

### Phase 3: Low Priority (BDD Tests)

| # | Task | Time | Status |
|---|------|------|--------|
| 3.1 | Evaluate if BDD refactoring is worthwhile | 5min | Pending |
| 3.2 | Create BDD helper functions (if worth it) | 15min | Pending |
| 3.3 | Refactor common BDD patterns | 20min | Pending |
| 3.4 | Run tests to verify | 5min | Pending |

**Total time:** ~45min
**Lines saved:** ~60+
**Risk:** LOW (tests only)
**Note:** May not be worth the effort - test duplication is less critical

---

## Priority Matrix

```
                    HIGH IMPACT
                         │
    ┌────────────────────┼────────────────────┐
    │                    │                    │
    │   Phase 1          │                    │
    │   Printer Helpers  │                    │
    │   (DO FIRST)       │                    │
    │                    │                    │
 LOW├────────────────────┼────────────────────┤HIGH
    │                    │                    │ EFFORT
    │   Phase 3          │   Phase 2          │
    │   BDD Tests        │   Templ Transforms │
    │   (OPTIONAL)       │   (DO SECOND)      │
    │                    │                    │
    └────────────────────┼────────────────────┘
                         │
                    LOW IMPACT
```

---

## Expected Outcomes

### After Phase 1 (Printer Helpers)
- Clone groups: 251 → ~220 (-12%)
- Printer package: Significantly cleaner
- Maintainability: Improved for all stats output

### After Phase 2 (Templ Transforms)
- Clone groups: ~220 → ~200 (-9%)
- Templ package: More maintainable
- Consistency: Better node creation patterns

### After Phase 3 (BDD Tests - Optional)
- Clone groups: ~200 → ~180 (-10%)
- BDD tests: More DRY
- Trade-off: May reduce test readability

---

## Recommendations

### Must Do
1. ✅ **Phase 1: Printer Helpers** - High impact, low effort, immediate value

### Should Do
2. ✅ **Phase 2: Templ Transforms** - Medium impact, improves code quality

### Consider Carefully
3. ⚠️ **Phase 3: BDD Tests** - Lower value, may reduce test clarity

### Don't Do
- Refactor test helper duplications (intentional for clarity)
- Consolidate similar error messages (different contexts)
- Merge similar test assertions (test isolation is valuable)

---

## Metrics

| Metric | Before | After Phase 1 | After Phase 2 |
|--------|--------|---------------|---------------|
| Clone groups | 251 | ~220 | ~200 |
| Largest group | 35 | ~20 | ~15 |
| Printer duplicates | 35+ | ~5 | ~5 |
| Templ duplicates | 11+ | 11+ | ~3 |

---

## Next Steps

1. **User approval** - Confirm plan execution
2. **Execute Phase 1** - Printer helpers refactoring
3. **Run full test suite** - Verify no regressions
4. **Execute Phase 2** - Templ transforms (if approved)
5. **Final verification** - Run `art-dupl` again to measure improvement

---

## Appendix: Full Clone List (Top 50)

### Largest Clone Groups (by occurrence count)

1. **35 clones** - `printer/stats_*.go` single-line prints
2. **24 clones** - BDD test setup patterns
3. **22 clones** - BDD test variations
4. **22 clones** - Test helper patterns
5. **18 clones** - Stats formatting patterns
6. **13 clones** - Single-line test patterns
7. **11 clones** - Templ transform patterns
8. **10 clones** - Stats recommendation patterns
9. **9 clones** - Test assertion patterns
10. **9 clones** - Stats test patterns

---

## Conclusion

The codebase has moderate duplication concentrated in:
- **Printer output formatting** (highest priority)
- **Templ AST transforms** (medium priority)
- **BDD test patterns** (low priority)

Executing **Phase 1** alone will reduce clone groups by ~12% with minimal effort and risk. **Phase 2** adds another ~9% reduction. **Phase 3** is optional and should be evaluated based on team preferences for test DRYness vs. test clarity.

**Recommended action:** Execute Phases 1 and 2, skip Phase 3.

---

_Generated by Crush (AI Assistant) on 2026-02-25 16:21_
