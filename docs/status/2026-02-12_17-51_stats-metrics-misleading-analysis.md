# Stats Command Misleading Metrics Report

**Date:** 2026-02-12 17:51
**Type:** Critical Analysis
**Status:** Issues Identified - Action Required

---

## Executive Summary

A comprehensive review of the `art-dupl stats` command revealed **multiple misleading metrics** that could cause users to misinterpret their code duplication status. The most critical issue is the **Duplication Ratio** which combines two flawed inputs.

---

## Critical Issues Found

### 1. Estimated Total Lines (CRITICAL)

**Location:** `cmd/stats.go:214`

```go
estimatedLines := filesCount * 100
sp.SetTotalEstimatedLines(estimatedLines)
```

**Problem:** Assumes every file has exactly 100 lines. This is a crude placeholder that makes the duplication ratio meaningless.

**Impact:**

- A 500-line file with 10 clones reports better duplication than reality
- A 50-line file with 5 clones reports worse than reality
- Users cannot trust the duplication percentage at all

**Severity:** 🔴 CRITICAL - Fundamental metric is unreliable

---

### 2. Total Duplicate Lines - Double Counting (CRITICAL)

**Location:** `printer/stats.go:259`

```go
for _, dup := range dups {
    // ...
    p.statsData.TotalDuplicateLines += lineCount
}
```

**Problem:** Counts ALL instances, not unique code patterns. If the same 20-line code appears in 3 files, reports **60 lines**, not 20.

**Impact:**

- Users think there's 60 lines of duplicated code
- Actually only 20 unique lines are duplicated 3x
- Inflates perceived duplication by 2-5x in typical projects

**Severity:** 🔴 CRITICAL - Misleading by design

---

### 3. Duplication Ratio - Compound Error (CRITICAL)

**Location:** `printer/stats.go:290`

```go
p.statsData.DuplicationRatio = float64(p.statsData.TotalDuplicateLines) / float64(p.statsData.TotalEstimatedLines) * 100
```

**Problem:** Combines two flawed inputs:

1. Double-counted duplicate lines (numerator)
2. Arbitrary estimated lines (denominator)

**Impact:**

- Reported "15%" could actually be 5% or 45%
- Impossible for users to know true duplication level
- Makes trend tracking unreliable

**Severity:** 🔴 CRITICAL - Primary metric is unreliable

---

## Moderate Issues

### 4. File Duplication Overcounting

**Location:** `printer/stats.go:263`

```go
p.statsData.FileDuplication[nstart.Filename] += lineCount
```

**Problem:** If a file has overlapping clones, the same lines get counted multiple times in the file's total.

---

### 5. Impact Score - Arbitrary Formula

**Location:** `printer/stats.go:271`

```go
p.statsData.ImpactScore += tokensInGroup * len(dups)
```

**Problem:**

- Formula grows quadratically: `tokens × instances`
- Arbitrary 10,000 max cap (`printer/stats.go:321`)
- No clear meaning to users

---

### 6. Complexity Score - Misleading Name

**Location:** `printer/stats.go:285`

```go
p.statsData.ComplexityScore = float64(p.statsData.TotalClones) / float64(p.statsData.TotalCloneGroups)
```

**Problem:** Just "average instances per group" — nothing to do with code complexity. Confusingly named.

**Suggestion:** Rename to `AverageInstancesPerGroup` or `SpreadScore`

---

### 7. Health Score - Arbitrary Thresholds

**Location:** `printer/stats.go:302-342`

**Arbitrary Magic Numbers:**

- Max complexity: `10.0`
- Max impact: `10000.0`
- Weights: `0.6` (duplication), `0.25` (complexity), `0.15` (impact)
- Grade thresholds: `<3` → A, `<6` → B, `<10` → C, `<15` → D, else F

**Problem:** No scientific basis for these values. Different project types would need different thresholds.

---

### 8. Size Distribution Counting

**Location:** `printer/stats.go:267`

```go
p.statsData.SizeDistribution[sizeRange]++
```

**Problem:** Counts by individual clone instances, not by unique code patterns. Same code appearing 10 times shows as "10 clones" not "1 pattern with 10 instances".

---

## Documentation Gaps

| Gap                               | Current State                  | User Impact                    |
| --------------------------------- | ------------------------------ | ------------------------------ |
| No unique vs instance distinction | "Total Clones" = all instances | Users expect unique patterns   |
| No error margins                  | Displays single numbers        | Users trust as accurate        |
| No Health Score explanation       | Magic formula hidden           | Users don't understand grades  |
| No Estimated Lines disclaimer     | Crude `files × 100`            | Users assume actual line count |

---

## Current Compiler Errors

The `cmd/stats.go` file has compilation errors related to incomplete filter stats integration:

```
cmd/stats.go:224:18: undefined: filter
cmd/stats.go:225:23: reason.String undefined
cmd/stats.go:228:7: sp.SetFilterStats undefined
```

This suggests partially implemented filter statistics feature.

---

## Recommendations

### Immediate Actions (High Priority)

| Action                                  | Impact                                | Effort |
| --------------------------------------- | ------------------------------------- | ------ |
| 1. Fix Estimated Lines                  | Actually count lines in scanned files | Medium |
| 2. Add Unique Duplicate Lines           | Count each pattern once               | Medium |
| 3. Add disclaimer for estimated metrics | Prevents false confidence             | Low    |
| 4. Document Health Score formula        | User understanding                    | Low    |

### Future Improvements (Medium Priority)

| Action                           | Description                             |
| -------------------------------- | --------------------------------------- |
| Rename Complexity Score          | → `AverageInstancesPerGroup` or similar |
| Add Unique Clone Patterns metric | Distinct from Total Clones              |
| Make Health Score configurable   | Allow project-specific thresholds       |
| Add confidence intervals         | Show error ranges for estimates         |

---

## Files Analyzed

- `cmd/stats.go` - Stats command implementation
- `printer/stats.go` - Statistics calculation (727 lines, needs refactoring)
- `printer/stats_data.go` - StatsData struct definition
- `printer/printer.go` - StatsPrinter interface

---

## Technical Debt Noted

From `printer/stats.go` comments:

> TODO: ARCHITECTURE ISSUES - This file exceeds 350 lines (currently 727 lines) and handles multiple concerns, violating Single Responsibility Principle

Suggested refactoring into:

- `stats_collector.go` - Statistics collection
- `stats_health.go` - Health score calculation
- `stats_formatter.go` - Format-specific output
- `stats_visualization.go` - Visualization helpers
- `stats_recommendations.go` - Recommendation logic
- `stats_styles.go` - Style management

---

## Conclusion

The `art-dupl stats` command has fundamental issues with its core metrics. The **Duplication Ratio** is the most deceptive because it combines two flawed inputs. Users relying on this metric for decision-making may take incorrect actions.

**Recommended Priority:** Fix Estimated Lines and add Unique Duplicate Lines metric before promoting stats feature for production use.

---

_Report generated by Crush AI Assistant_
