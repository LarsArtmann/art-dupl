# Feedback: `-t 1` reports independent boolean flags as a clone group

**Date:** 2026-08-05  
**Project:** `go-humanize-linter`, an AST-based Go linter detecting hand-rolled reimplementations of `dustin/go-humanize`  
**Command:** `art-dupl --type-aware --sort total-tokens -t 1`  
**Goal:** Verify whether any harmful duplication exists in the linter's AST-pattern helpers.

> **Verdict:** The report found **1 clone group / 2 occurrences**, but it is a structural false positive. The matched code is two independent pairs of local boolean flags initialized to `false` before unrelated `ast.Inspect` walks. There is zero shared logic and no sensible extraction.

---

## Results

| Group              | Locations                                                    | Report category | Actual category               | Decision |
| ------------------ | ------------------------------------------------------------ | --------------- | ----------------------------- | -------- |
| Boolean flag pair  | `pattern_commaf.go:37-38`, `pattern_helpers.go:399-400`      | `unknown`, low  | Go AST-traversal accumulator idiom | Accept   |

No source files were modified.

---

## Finding: Two-line `name := false` initializer pair

### What the report matched

```go
// pattern_commaf.go:37-38 (hasCommafPattern)
hasFloatFormat := false
hasSeparatorLoop := false
```

```go
// pattern_helpers.go:399-400 (isSuppressedRule)
hasAll := false
hasLinter := false
```

Both snippets are two consecutive `bool` variable declarations with `:= false` initialization.

### Why they are not duplicated logic

The contexts are unrelated:

- `hasCommafPattern` is an AST-pattern detector for rule **H009** (manual `humanize.Commaf` reimplementation). It walks a function declaration looking for a `fmt.Sprintf("%.Nf", ...)` or `strconv.FormatFloat` call **and** a comma-separator insertion. The two flags independently record whether each prerequisite was observed.
- `isSuppressedRule` is a `//nolint` suppression parser. It walks a list of `//nolint:` tokens looking for the `all` marker, the linter name (`gohumanize`), and any scoped rule IDs. The two flags independently record which suppression categories were observed.

The only commonality is the shape of the implementation mechanism: `ast.Inspect` (or a plain loop) plus two local accumulator booleans. That is a standard Go idiom, not duplicated domain logic.

### Why extraction would be worse

A helper to "deduplicate" the initialization would look like:

```go
func newBoolPair() (bool, bool) { return false, false }
```

or some generic accumulator struct. Either would:

1. Obscure the meaning of each flag by separating declaration from use.
2. Couple two unrelated detectors (`H009` pattern logic and `//nolint` parsing).
3. Add an abstraction that saves zero lines of meaningful code.
4. Violate the project's convention of keeping each rule's detection self-contained.

### Tool feedback

The clone appears only because the token threshold is set to `1`, the minimum possible. At this threshold the detector matches trivial syntactic fragments such as:

```go
foo := false
bar := false
```

Suggested classification:

- **Pattern:** `bool-accumulator-initializer`
- **Priority:** informational or suppressed at `-t 1`
- **Suggestion:** "Local boolean flags used as AST-traversal accumulators; no shared logic."

At a practical threshold (the project uses `-t 3` in its duplication gate), this finding disappears. The report should either suppress `:= false` pairs below a higher default threshold or explicitly label them as low-confidence structural matches rather than actionable duplication.

---

## What works well

### Type-aware detection

Even with `-t 1`, the tool correctly identified the two locations as textually similar. Type awareness did not produce a false positive from type differences because the variables happen to share the same `bool` type. The issue is purely threshold sensitivity.

### `--sort total-tokens`

With only one group this was not decisive, but sorting by total tokens is the right triage default for larger reports.

---

## Summary of suggestions

| Priority | Suggestion                                                                        | Effort |
| -------- | --------------------------------------------------------------------------------- | ------ |
| MEDIUM   | Suppress or down-rank `name := false` / `name := true` initializer pairs at `-t 1` | Small  |
| LOW      | Add an `idiom` category for accumulator-flag declarations in Go                     | Small  |
| LOW      | Default minimum threshold recommendation should be higher than `1` for Go           | Small  |

The `go-humanize-linter` codebase has no actionable duplication from this scan. The single reported group should be treated as a known low-threshold artifact, not as a refactoring opportunity.
