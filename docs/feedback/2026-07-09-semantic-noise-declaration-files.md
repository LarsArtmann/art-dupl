# Feedback: Semantic Mode Reports Noise in Declaration-Only Files

**Date:** 2026-07-09
**Source:** `/home/lars/forks/upd/` — run with `--semantic` (threshold 2)
**Report:** `/home/lars/forks/upd/art-dupl.html`

## Problem

With `--semantic` mode and low thresholds, art-dupl reports 3 clone groups that are all noise:

| Group | Occurrences | Example                                           | Issue                                                      |
| ----- | ----------- | ------------------------------------------------- | ---------------------------------------------------------- |
| #1    | 13          | `errors.New("different strings")`                 | Matched on shared `errors.New(` prefix, strings all differ |
| #2    | 2           | `json.Unmarshal` + `if err != nil { return nil }` | Common boilerplate, only 2 statements                      |
| #3    | 2           | Same as #2 in different files                     | Same boilerplate pattern                                   |

The user expects NONE of these to be reported, especially Group #1 where string values are all different.

## Root Cause Analysis

### 1. Clone Group #1: Partial Expression Prefix Matching

**What happens:** `errors.go` and `config.go` contain ONLY `var` blocks with error definitions:

```go
var (
    ErrFileNotFound = errors.New("package configuration file not found")
    ErrInvalidJSON  = errors.New("invalid JSON in package configuration file")
    ...
)
```

These files have **zero function bodies**, so **zero `Statement = true` nodes** exist.

**The cascade:**

1. `serial()` (`syntax.go:178`) fingerprints statement nodes into single composite tokens. But these files have no statements, so **legacy node-level tokenization** is used — each AST node becomes a separate token.

2. The token sequence for `errors.New("help requested")` is:

   ```
   [CallExpr] [SelectorExpr:"New"] [Ident:"errors"] [Ident:"New"] [BasicLit:"\"help requested\""]
   ```

   In semantic mode, the BasicLit **does** hash its value (`transform.go:42`). So the 5th token differs across calls.

3. The suffix tree finds the **common prefix** of 4 tokens: `[CallExpr, SelectorExpr:"New", Ident:"errors", Ident:"New"]`. With threshold=2, this easily matches.

4. `FindSyntaxUnits` (`syntax.go:239`) has a safeguard at line 260-263 that rejects non-statement matches in files that **contain** statements. But since these files have **no statements**, `fileContainsStatements()` returns `false`, and the safeguard is **bypassed**.

5. `buildMatch` (`syntax_match.go:37-41`) uses the CallExpr's `Owns=4` to compute `endIdx=4`, so the reported fragment covers the **entire CallExpr source range** — including the differing BasicLit. The user sees `errors.New("help requested")` vs `errors.New("version requested")` reported as "the same clone," even though only the prefix matched.

### 2. Default Threshold of 1

`config.go:170`: `const DefaultThreshold = 1`

The comment says "report any duplicated statement." This is far too aggressive. Most clone detection tools use minimum thresholds of 5-15 statements or 30-100 tokens.

### 3. No Hard Minimum Threshold

`config_validate.go:41`: `if threshold < 1` — the floor is 1. There's no "sensible minimum" enforcement. A threshold of 1-2 produces massive noise.

### 4. Boilerplate Patterns Not Filtered

Groups #2 and #3 are `json.Unmarshal` + error-check boilerplate. The actionability system (`actionability.go`) recognizes error-propagation and error-wrapping patterns, but not the generic `if err != nil { return nil }` after a function call.

## Recommendations (Prioritized by Impact)

### Recommendation 1: Raise Default Threshold to 5 (HIGH IMPACT, LOW RISK)

**File:** `config/config.go:170`

```go
// Before:
const DefaultThreshold = 1

// After:
const DefaultThreshold = 5
```

A threshold of 5 statements filters most trivial matches while still catching real duplication. The comment should explain the tradeoff.

**Also update:** `cmd/flags.go:16` — update the help text.

### Recommendation 2: Enforce Minimum Threshold Floor of 3 (HIGH IMPACT, LOW RISK)

**File:** `config/config_validate.go:40-43`

```go
// Before:
func validateThreshold(threshold int) error {
    if threshold < 1 {
        return fmt.Errorf("%w: %d", ErrInvalidThreshold, threshold)
    }

// After:
func validateThreshold(threshold int) error {
    if threshold < 3 {
        return fmt.Errorf("%w: %d (minimum is 3 to avoid noise)", ErrInvalidThreshold, threshold)
    }
```

This prevents users from accidentally generating noise. Also update `domain.ErrInvalidThreshold` message if needed.

### Recommendation 3: Mark Top-Level `ValueSpec` as Statement-Equivalent (HIGH IMPACT, MEDIUM RISK)

**Problem:** Declaration-only files (error definitions, constants, config structs) fall into legacy matching because no nodes are marked `Statement = true`. This causes partial expression prefixes like `errors.New(` to match.

**Fix:** In `syntax/golang/transform.go`, mark `ValueSpec` nodes inside `GenDecl` as statement-equivalent, so they get fingerprinted into single composite tokens:

```go
// In the GenDecl case or a new post-processing pass:
// Mark each ValueSpec child of a GenDecl as Statement = true
case *ast.GenDecl:
    o.Type = encodeSemanticType(GenDecl, n.Tok.String(), t.config.Mode.hashesIdentifiers())
    for _, spec := range n.Specs {
        child := t.trans(spec)
        child.Statement = true  // ← ADD THIS
        o.AddChildren(child)
    }
```

This ensures `ErrFoo = errors.New("foo")` and `ErrBar = errors.New("bar")` become **single composite tokens** with **different fingerprints** (since the BasicLit values hash differently). The suffix tree won't match them because the entire ValueSpec is one token, and that token differs.

**Files to change:**

- `syntax/golang/transform.go` — mark ValueSpec as Statement
- `syntax/golang/transform_test.go` — add test verifying ValueSpec fingerprinting
- Existing tests that may break from changed tokenization

**Risk:** Changes tokenization for ALL var/const blocks. Must verify no regressions in clone detection quality.

### Recommendation 4: Add "Single Call Expression" Noise Filter (MEDIUM IMPACT, LOW RISK)

**Problem:** Clones that are a single `CallExpr` node (like `errors.New(...)`) are almost never actionable duplicates — they're function calls with different arguments.

**Fix:** Add a new actionability pattern in `printer/actionability.go`:

```go
// isSingleCallExpression reports whether every clone is exactly one CallExpr
// node. Single function calls with different arguments are not actionable
// duplication — they're just using the same API.
func isSingleCallExpression(nodeSeqs [][]*domain.CloneNode) bool {
    return everySequenceMatch(nodeSeqs, func(seq []*domain.CloneNode) bool {
        if len(seq) != 1 {
            return false
        }
        return seq[0].BaseType == golang.CallExpr
    })
}
```

Add to `evaluateActionabilityDetailed()` before the final `return PatternNone, domain.Actionable`.

### Recommendation 5: Improve Fragment Rendering for Partial Matches (MEDIUM IMPACT, MEDIUM RISK)

**Problem:** When the suffix tree matches a partial expression prefix (e.g., 4 of 5 tokens), `buildMatch` uses `Owns` to extend the fragment to the **full** expression — including tokens that **didn't** match. The user sees two different expressions reported as "the same clone."

**Fix:** In `buildMatch` (`syntax/syntax_match.go:37-41`), only extend to `Owns` boundary when the match actually covers the full subtree:

```go
// Before:
endIdx := lastIndex + int(lastNode.Owns)

// After: only extend if the match covers the full owned subtree
endIdx := lastIndex + int(lastNode.Owns)
if lastIndex+1+int(lastNode.Owns) > len(firstSeq) {
    // Match doesn't cover the full subtree — use match length, not Owns
    endIdx = len(firstSeq) - 1  // or lastIndex + 1
}
```

This way, if only 4 of 5 tokens matched, the fragment shows only the matched prefix, not the full expression.

### Recommendation 6: Add `--min-lines` Flag as Complementary Filter (LOW PRIORITY)

Most clone detection tools have both a token/statement threshold AND a minimum line count. A clone must satisfy BOTH to be reported. This filters single-statement clones that happen to be long (like a deeply nested struct literal).

```go
// In config:
MinLines int `json:"minLines,omitempty"`

// In printer, filter groups where the average line count < MinLines
```

### Recommendation 7: Semantic Mode Should Normalize String Literals (DISCUSSION)

**Current behavior:** Semantic mode hashes string literal VALUES verbatim (`transform.go:42`). This means `errors.New("foo")` and `errors.New("bar")` produce different tokens.

**The argument for normalizing:** True Type-2 clones (renamed variables AND renamed literals) would be detected. This is what some academic clone detection tools do.

**The argument against:** The user explicitly says these should NOT be reported. Normalizing string literals would make `errors.New("foo")` and `errors.New("bar")` match as clones — the opposite of what the user wants.

**Recommendation:** Do NOT normalize string literals. The current behavior is correct for the user's use case. The noise comes from partial prefix matching (Recommendation 3), not from literal handling.

## Verification Steps

After implementing Recommendations 1-3:

1. Run `art-dupl --semantic /home/lars/forks/upd/` — Group #1 should disappear entirely
2. Run with `--threshold 5` — Groups #2 and #3 should also disappear (only 2-4 tokens each)
3. Run full test suite: `go test ./...`
4. Run BDD tests: `go test ./bdd/`
5. Verify no regressions on real codebases with intentional duplication
