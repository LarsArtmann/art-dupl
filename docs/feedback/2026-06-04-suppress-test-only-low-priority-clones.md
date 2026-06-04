# Feedback: Suppress intra-test clone groups that differ only by literal values

**Date:** 2026-06-04
**Project:** cqrs-htmx (Go, ~400 tests)
**Threshold:** 45 (semantic)
**Result:** Production code at zero. 3 test-only clone groups remain.

## Problem

After eliminating all production clones, 3 test-only clone groups remain that are false positives:

### Example 1: Different test scenarios, same structure

```go
// coverage_test.go:692
It("extracts from X-Forwarded-For", func() {
    r := httptest.NewRequest(http.MethodGet, "/", nil)
    r.Header.Set("X-Forwarded-For", "1.2.3.4, 5.6.7.8")       // ← different header + value
    r.RemoteAddr = "10.0.0.1:1234"
    Expect(cqrshtmx.ClientIP(r)).To(Equal("1.2.3.4"))          // ← different expected
})

// coverage_test.go:699 — same file, next test
It("extracts from X-Real-IP when no XFF", func() {
    r := httptest.NewRequest(http.MethodGet, "/", nil)
    r.Header.Set("X-Real-IP", "9.8.7.6")                       // ← different header + value
    r.RemoteAddr = "10.0.0.1:1234"
    Expect(cqrshtmx.ClientIP(r)).To(Equal("9.8.7.6"))          // ← different expected
})
```

### Example 2: Table-driven validation tests

```go
// csrf_test.go:613
It("returns error when TrustedOrigins contains wildcard", func() {
    cfg := cqrshtmx.CSRFConfig{TrustedOrigins: []string{"*"}}   // ← "*"
    err := cfg.Validate()
    Expect(err).To(HaveOccurred())
    Expect(errors.Is(err, cqrshtmx.ErrCSRFConfig)).To(BeTrue())
})

// csrf_test.go:622 — same file, next test
It("returns error when TrustedOrigins contains empty string", func() {
    cfg := cqrshtmx.CSRFConfig{TrustedOrigins: []string{""}}    // ← ""
    err := cfg.Validate()
    Expect(err).To(HaveOccurred())
    Expect(errors.Is(err, cqrshtmx.ErrCSRFConfig)).To(BeTrue())
})
```

### Example 3: Same test across coverage + main test files

```go
// usermgmt/coverage_test.go:862 — tests SessionMaxAge=7200
// usermgmt/handler_test.go:346 — tests SessionMaxAge=3600
// Identical structure, different numeric literal (7200 vs 3600)
```

## Common Pattern

All 3 groups share these properties:

1. **All occurrences are test code** (already detected via `data-test`)
2. **Structurally identical** except for string/numeric literals
3. **Same Describe/Context block or adjacent tests** testing the same function with different inputs
4. This is the standard Go table-driven test pattern — extracting a helper would harm readability

## Proposed Minimal Change

**Lower the priority of test-only clone groups to `info` / `suppressed` (below `low`) when:**

- All occurrences in the group are test files (`isTest == true` for all), AND
- The clone group is already priority `low`

This is the most minimal general change because:

- **No AST changes needed** — no literal-diffing heuristic required
- **No new exclusion flags** — works with existing classification
- **Semantically correct** — test-only `low` priority clones are almost always idiomatic test patterns (Ginkgo `It` blocks, table-driven test cases, assertion sequences)
- **Safe** — medium+ test clones (shared test helpers, copy-pasted setup) would still be reported

### Alternative (slightly more complex): `--suppress-test-low` flag

Add a boolean flag that filters out clone groups where all occurrences are test code AND priority is `low`. This gives users control without changing defaults.

### Alternative (more accurate but more complex): Literal-aware semantic diff

When all occurrences are in test files, do a second pass that ignores string and numeric literals. If clones become identical after literal normalization, suppress them. This would also catch cases where tokens exceed the `low` threshold but the duplication is still just test data.

## Impact

In this project (390+ tests, 96.9% coverage):

| Metric               | Before | After (suppress test-low) |
| -------------------- | ------ | ------------------------- |
| Production clones    | 0      | 0                         |
| Test clones reported | 3      | 0                         |
| Total clone groups   | 3      | 0                         |

The signal-to-noise ratio for the "test clones" category would improve significantly. Today, every run shows 2-3 test clones that are all false positives, training users to ignore the `🧪 test` badge entirely.
