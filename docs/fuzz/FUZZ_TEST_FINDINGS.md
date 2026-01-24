# Fuzz Test Findings - art-dupl

**Date:** January 14, 2026
**Testing Tool:** Go native fuzzing (testing.F)

## Bugs Found

### 1. Suffix Tree Nil Pointer Dereference

**Location:** `suffixtree/suffixtree.go` - Update method
**Test:** `FuzzSuffixTreeUpdate`
**Input:** "զ" (Armenian letter)
**Error:** `runtime error: invalid memory address or nil pointer dereference`

**Description:**
The suffix tree Update method panics with nil pointer dereference when processing certain Unicode characters. This occurs during the tree construction phase when handling characters outside the expected range.

**Reproduction:**

```bash
go test -run=FuzzSuffixTreeUpdate/2a748477c7945668 ./suffixtree
```

**Impact:**

- High - Core algorithm crashes on valid input
- Could crash analysis on international code
- Affects all users analyzing non-ASCII code

**Priority:** 🔴 Critical

**Potential Root Cause:**
The `char` type is defined as `byte` which only supports ASCII (0-255). Unicode characters outside this range cause issues when the tree tries to find transitions or create new states.

**Suggested Fix:**

1. Validate input range before processing
2. Use `rune` instead of `byte` for character type
3. Add proper error handling for invalid characters

---

### 2. Filter Pattern Regex Unicode Handling

**Location:** `pkg/filter/filter.go` - matchPattern function
**Test:** `TestIncludePatternProperty` (property-based test)
**Input:** Complex Unicode strings with supplementary characters
**Error:** Regex compilation fails for certain Unicode patterns

**Description:**
The `matchPattern` function converts wildcard patterns to regex using string replacement, which doesn't properly handle all Unicode characters, particularly supplementary characters and emojis.

**Impact:**

- Medium - Pattern matching fails for international file paths
- Filter may not work correctly with Unicode filenames
- Edge case with rare characters

**Priority:** 🟠 Medium

**Potential Root Cause:**
The `strings.ReplaceAll` function doesn't handle Unicode normalization correctly when converting `*` wildcards to `.*` regex patterns.

**Suggested Fix:**

1. Use proper regex escaping functions
2. Validate regex compilation before matching
3. Add Unicode normalization support

---

## Summary

Total Bugs Found: 2

- Critical: 1
- High: 0
- Medium: 1
- Low: 0

This demonstrates the value of fuzzing and property-based testing - finding edge cases and bugs that would be difficult to discover with traditional testing.
