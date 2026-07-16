# Feedback: Real-world deduplication session on a small Go web project

> ✅ **ADDRESSED** — Default threshold raised to 5 (commit `930b91a`). Semantic mode improvements (literal normalization, alpha-normalization, actionability patterns) now suppress the boilerplate clones reported at t=15. The project achieved 0 clones at t=50, and with default threshold 5, noise is further reduced while maintaining precision.

**Date:** 2026-06-08
**Project:** overview (Go, ~2600 LOC, 17 test files)
**Command:** `art-dupl -t 15 . --semantic --sort total-tokens`
**Result:** 17 clone groups at threshold 15. 0 clone groups at threshold 50.

## Session Summary

Ran art-dupl at the aggressive threshold 15 on a well-maintained Go project (overview — a local project dashboard). The goal was to find and eliminate every meaningful clone.

**What was found:** 17 clone groups totaling 37 clone tokens.
**What was actually deduplicated:** 2 groups (3 changes). The remaining 15 groups are Go idioms and test patterns, not real duplication.

## What Worked Well

### Classification and badges are excellent for triage

The HTML report's category/priority/test badges made triage fast. I could immediately see:

- **Production vs test** split (4 production, 33 test)
- **Priority** (35 low, 2 medium)
- **Category** (method, conditional, assignment, unknown)

This let me prioritize the 2 medium-priority production clones first, then scan the low-priority test clones for anything real.

### Semantic detection at threshold 15 catches structural patterns well

The tool correctly identified structural clones even when formatting/whitespace differed. The `okHandler` pattern (4 occurrences of `func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }`) was detected despite different indentation levels across files.

### The diff view in HTML mode is very useful

Side-by-side comparison with word-level diff highlighting made it easy to see whether two clones were truly identical or just structurally similar with different literals.

## Actionable Deduplications Found

### 1. Extracted `daysAgoGit()` usage (Group #2)

```go
// Before — duplicated in server_test.go:191 and view_test.go:238
Git: &domain.GitInfo{LastCommitDate: now.Add(-time.Duration(s.daysAgo) * 24 * time.Hour)}

// After — used existing helper
Git: daysAgoGit(s.daysAgo, now)
```

The helper already existed in `test_helpers_test.go` but wasn't used consistently. art-dupl caught the gap.

### 2. Extracted repeated test literal (Group #8)

```go
// Before — identical literal at view_test.go:1098 and view_test.go:1136
filter := filterParams{Search: "test", Language: "Go", Activity: ActivityActive, Sort: SortByLOC}

// After — extracted to package-level test variable
var testFullFilter = filterParams{Search: "test", Language: "Go", Activity: ActivityActive, Sort: SortByLOC}
```

## False Positives at Threshold 15

The remaining 15 groups (at 2-4 tokens each) are all Go idioms or test patterns that should NOT be deduplicated:

### Table-driven test data (5 groups)

```go
// Group #7 — two test fixture lines, different data
{Name: "alpha", Path: "/home/user/alpha", Languages: []domain.Language{"Go"}},
{Name: "beta", Path: "/home/user/beta", Languages: []domain.Language{"Python"}},

// Group #11 — XSS test table entries with different attack vectors
{name: "strips javascript URLs (XSS prevention)", markdown: "[click](javascript:alert(1))\n", ...},
{name: "strips event handlers (XSS prevention)", markdown: "<img src=x onerror=alert(1)>\n", ...},
```

Table-driven test entries are structurally identical by design. Extracting them would harm readability.

### Similar but semantically different helpers (3 groups)

```go
// Group #3 — different failure modes (t.Error vs t.Fatal)
func assertStatus(t *testing.T, recorder *httptest.ResponseRecorder, want int)
func assertStatusFatal(t *testing.T, recorder *httptest.ResponseRecorder, want int)

// Group #9 — different types (generic int vs string)
func assertInt[T int | int64](t *testing.T, got, want T, label string)
func assertString(t *testing.T, got, want, label string)
```

These look similar but serve distinct purposes. Merging them would make the API worse.

### Inverse conditions (1 group)

```go
// Group #5 — cache expiry check, used in opposite branches
c.cached == nil || time.Since(c.cachedAt) >= c.ttl    // "cache is stale"
c.cached != nil && time.Since(c.cachedAt) < c.ttl     // "cache is fresh"
```

These are logically inverted (`>=` vs `<`, `== nil` vs `!= nil`). Extracting a shared method would obscure the intent of each call site.

### Standard library patterns (3 groups)

```go
// Group #16 — standard Go test context pattern
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

// Group #17 — writeETag early return pattern (different handler functions)
if writeETag(response, request, data.ETag, nameETagQuery(proj.Name)) {
    return
}
```

These are Go idioms appearing in different functions. "Deduplicating" them would mean extracting a function for a 2-line pattern that's already a standard Go pattern.

### Similar test struct declarations (1 group)

```go
// Group #17 — test table struct types in different test functions
tests := []struct {
    name string
    git  *domain.GitInfo
    want Activity  // vs want int
}{}
```

Different `want` types, different test functions. Not deduplicatable.

## Suggestions

### 1. Suppress test-only `low` priority clones (reiterate previous feedback)

Of the 17 clone groups found, 13 were test-only `low` priority. Every single one was a false positive. This aligns exactly with the 2026-06-04 feedback — test-only low-priority clones have near-zero signal.

At threshold 15, 76% of reported clones were noise. At threshold 50, there were zero clones. The interesting work happens between 15-50, but at 15 the test noise drowns the signal.

### 2. Consider token-count-aware categories

At 2-4 tokens, most clones are structural artifacts (function signatures, import patterns, standard library calls). A category like `idiom` or `structural` for very low token counts (<5 tokens?) would help users skip these immediately.

The current categories (`method`, `conditional`, `assignment`, `unknown`) describe the AST node type but don't convey whether the clone is actionable. A token-count threshold for an `idiom` category would add semantic meaning.

### 3. Inverse-condition detection

Group #5 shows two clones that are logical inverses (`>=` vs `<`, `== nil` vs `!= nil`). These are related by design (both check cache freshness) but in opposite directions. Detecting inverse conditions and grouping them with a note like "inverse of line X" would be more useful than reporting them as clones, since the right response is usually "keep both" rather than "extract shared code."

### 4. Cross-reference with existing helpers

Group #2 was the most actionable because a helper (`daysAgoGit`) already existed but wasn't used consistently. If art-dupl could detect "this clone pattern already has a near-identical function defined in the same package" and suggest "consider using existing function X", it would turn a manual triage step into an automatic suggestion.

This is likely complex to implement (requires function body similarity matching), so it may be a stretch goal.

## Metrics

| Metric                  | Value |
| ----------------------- | ----- |
| Clone groups (t=15)     | 17    |
| Clone groups (t=50)     | 0     |
| Real duplications found | 2     |
| False positives         | 15    |
| False positive rate     | 88%   |
| Production clones       | 4     |
| Test clones             | 33    |
| Actionable (production) | 0     |
| Actionable (test)       | 2     |

## Conclusion

art-dupl at threshold 50 is the sweet spot for this project — zero clones, zero noise, zero false positives. At threshold 15, the signal-to-noise ratio drops to 12% (2 of 17 groups were actionable).

The tool is most valuable for finding gaps where helpers exist but aren't used consistently (Group #2). The classification system and HTML report are genuinely excellent for triage — the badges and diff views saved significant time.

The main improvement opportunity is suppressing test-only low-priority clones, which would have reduced the report from 17 groups to 4 (2 medium production + 2 low production), with a 50% actionable rate instead of 12%.
