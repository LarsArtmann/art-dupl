# Status Report: Output Quality Fixes

**Date:** 2026-08-10 13:45\
**Session Scope:** Fix broken UX in `--suggest-generics` / `--explain` text output\
**Branch:** fork

---

## Context

User ran `art-dupl` on the `erraudit` project and identified multiple output quality issues across three command variants:

1. `--type-aware --suggest-generics --sort total-tokens -t 1 --explain`
2. `--type-aware --suggest-generics --sort total-tokens -t 1`
3. `--type-aware --sort total-tokens -t 1`

The core complaint: removing `--suggest-generics` revealed 10 clone groups that were previously hidden, the generics hint was unreadable, and the explain output had grammar errors.

---

## a) FULLY DONE

### 1. `--suggest-generics` now respects `--show-suppressed`

**File:** `cmd/run_output.go:178-182`\
**Problem:** The generics filter unconditionally skipped non-generics clones with `continue` — no `ShowSuppressed` check, unlike the actionability filter (line 152) and `shouldSuppressGroup` (line 173) which both guard with `if !suppression.ShowSuppressed`.\
**Fix:** Added `if !suppression.ShowSuppressed { continue }` guard, matching the pattern used by the other two suppression branches.\
**Tests:** All existing `cmd/` tests pass. No new test added (should add one — see NOT STARTED).

### 2. "1 tokens" → "1 token" pluralization

**File:** `printer/text.go:258-268` (new helpers), `printer/text.go:334` (explain), `printer/text.go:279` (rich header)\
**Problem:** Both `writeExplanation` and `writeRichGroupHeader` used `fmt.Sprintf("%d tokens, %d lines", ...)`, producing "1 tokens, 3 lines".\
**Fix:** Added `tokenLineSummary(tokens, lines int)` and `plural(word string, n int)` helpers. Both call sites now use `tokenLineSummary`.\
**Tests:** All existing text printer tests pass.

### 3. Generics hint shortens fully-qualified type paths

**File:** `printer/generics_candidate.go:67-87` (new `shortenTypeString`), `printer/generics_candidate.go:103` (call site)\
**Problem:** Hint output was unreadable: `github.com/larsartmann/erraudit/internal/analyzer.MainAnalyzerOption vs github.com/larsartmann/erraudit/internal/ast.FileCacheOption`\
**Fix:** `shortenTypeString` strips Go import path segments (everything matching `[\w.-]+/`) down to the last package segment: `analyzer.MainAnalyzerOption vs ast.FileCacheOption`. Handles prefixes like `[]`, `*`, `map[string]`.\
**Tests:** Added `TestShortenTypeString` (10 cases) and `TestClassifyGenericsCandidate_HintShortensPackagePaths` (integration test with real fully-qualified paths). Both pass.

### 4. Generics candidates show generics-specific fix suggestion

**File:** `printer/clone_processor.go:145-147`\
**Problem:** A generics candidate clone showed "Extract loop body to helper function" (the category-based suggestion from `getSuggestion`), not mentioning generics at all.\
**Fix:** After setting `GenericsCandidate`/`GenericsHint`, if the clone is both a generics candidate AND actionable, override `Suggestion` to "Extract to generic function". Non-actionable clones keep their "why" pattern explanation.\
**Tests:** All existing `clone_processor` and `actionability` tests pass.

### 5. All tests pass

- `go build ./...` — clean
- `go test ./...` — all 30 packages pass
- `golangci-lint run --timeout 5m ./printer/... ./cmd/...` — only pre-existing `nestif` warning on `PrintFooter` (not touched)

---

## b) PARTIALLY DONE

Nothing partially done. All 4 fixes are complete and tested.

---

## c) NOT STARTED

### Tests I should have written but didn't:

1. **No test for `--show-suppressed` + `--suggest-generics` interaction** — The fix in `cmd/run_output.go` has no dedicated test. Existing tests pass but don't specifically verify that `--suggest-generics --show-suppressed` shows non-generics clones. This is a behavioral change that should be regression-guarded.

2. **No test for generics candidate suggestion override** — The `clone_processor.go` change (overriding `Suggestion` to "Extract to generic function") has no test. Should verify that a generics candidate clone's `Classification.Suggestion` equals "Extract to generic function".

3. **No test for `tokenLineSummary` with singular** — The function is indirectly tested by existing tests (which use plural values), but no test specifically checks the `n == 1` path.

### Things I noticed but didn't fix:

4. **`shortenTypeString` may shorten too aggressively for same-package types** — If two types are `analyzer.MainAnalyzerOption` and `analyzer.FileCacheOption`, the hint shows `analyzer.MainAnalyzerOption vs analyzer.FileCacheOption` — no shortening needed. But if types are `pkg.Foo` and `pkg.Bar` from the same package, we could show just `Foo vs Bar`. Not a bug, just a potential enhancement.

5. **The `plural()` function is package-private to `printer`** — If other packages need pluralization, it should move to a shared util. Not needed now.

---

## d) TOTALLY FUCKED UP

Nothing. All changes are clean, tested, and build successfully.

---

## e) WHAT WE SHOULD IMPROVE

### Architecture / Design Issues Noticed

1. **`--suggest-generics` is a filter, not an enhancer** — This is the root cause of the user's confusion. `--suggest-generics` HIDES all non-generics clones. The user expected it to HIGHLIGHT generics candidates among all results. The current design means `--suggest-generics` and `--type-aware` produce fundamentally different output sets, not supersets/subsets. Consider:
   - Making `--suggest-generics` an enhancer (show all clones, highlight generics candidates with the hint) rather than a filter
   - Or at minimum, the help text should say "filter to only generics-extraction candidates" not "find generics-extraction candidates"

2. **Three suppression branches with inconsistent `ShowSuppressed` handling** — The actionability branch (line 145-155), `shouldSuppressGroup` (line 170-176), and generics branch (line 178-182, now fixed) all follow the same pattern but the generics one was the only one missing `ShowSuppressed`. This is a code smell — the pattern should be extracted into a helper that always checks `ShowSuppressed`.

3. **Suggestion override happens too late in the pipeline** — `clone_processor.go` sets `Suggestion` via `ClassifyClone` (line 112), then `ApplyPatternLabel` may override it (line 135), then the generics override (line 145). Three layers of override is fragile. A cleaner design would compute the suggestion once, considering all factors (category, pattern, generics).

4. **`shortenTypeString` uses regex where `strings.LastIndex` would suffice** — The regex `[\w.-]+/` works but is harder to reason about than splitting on `/` and taking the last segment. The regex approach was chosen to handle `[]`/`*` prefixes inline, but a split-then-rejoin approach would be clearer.

5. **No integration test for the full `--suggest-generics --explain` pipeline** — The BDD tests in `bdd/` don't cover `--suggest-generics` or `--explain` at all. All testing was unit-level.

6. **The `nestif` lint warning on `PrintFooter`** — Pre-existing (complexity 13), but I touched the file. Should be refactored to extract the suppression-stats formatting into a helper.

### Output Quality Issues Noticed (Not Fixed)

7. **Progress output indentation is inconsistent** — `📖 Parsing...` (4 spaces), `264 files discovered` (4 spaces), `🔍 Type-aware mode...` (0 spaces), `✅` (1 space). These come from different files (`run_analysis.go`, `progress.go`, `type_aware.go`) and have no shared indentation convention.

8. **The `generics:` hint line can be extremely long** — Even with shortened type names, 3 type pairs + the "(N type differences total)" suffix can produce a very long line. No truncation or wrapping is applied.

9. **`--explain` output order: generics hint BEFORE explain** — The output shows `generics:` line, then `explain:` line, then `fix:` line. The generics hint is contextual to the explain, so it might read better as part of the explain line itself.

---

## f) Up to 50 Things We Should Get Done Next

### High Priority (Directly related to this session's work)

1. Add regression test: `--suggest-generics --show-suppressed` shows non-generics clones
2. Add regression test: generics candidate `Classification.Suggestion == "Extract to generic function"`
3. Add test: `tokenLineSummary(1, 1)` returns "1 token, 1 line"
4. Add BDD test: full `--suggest-generics --explain` pipeline output
5. Refactor `PrintFooter` to fix `nestif` complexity warning (extract suppression stats formatting)
6. Consider redesigning `--suggest-generics` as an enhancer rather than a filter (breaking change, needs ADR)
7. Update `--suggest-generics` flag help text to clarify it FILTERS, not enhances
8. Extract the three suppression branches in `printCloneGroups` into a single `shouldSuppressGenerics` helper for consistency

### Medium Priority (Output quality improvements noticed during this session)

9. Standardize progress output indentation across `run_analysis.go`, `progress.go`, `type_aware.go`
10. Add line-length truncation for the `generics:` hint line
11. Consider merging `generics:` hint into the `explain:` line when both are shown
12. Add `--suggest-generics` coverage to BDD tests
13. Add `--explain` coverage to BDD tests
14. Audit all user-facing strings for pluralization issues (search for `%d.*s` patterns)
15. Add golden file test for `--suggest-generics --explain` output combination

### Lower Priority (General improvements noticed while reading code)

16. `shortenTypeString` — consider `strings.LastIndex` approach instead of regex for clarity
17. Consolidate suggestion computation in `clone_processor.go` — single pass instead of 3 overrides
18. Add `--no-generics-hint` flag to suppress the `generics:` line (for machine consumers)
19. JSON output: verify `generics_candidate` and `generics_hint` fields are present and correct
20. SARIF output: verify generics candidate info is included
21. HTML output: verify generics candidate rendering
22. SDK (`pkg/artdupl`): expose `GenericsCandidate` and `GenericsHint` on `Clone` type
23. Review `formatGenericsHint` deduplication — `seen` map uses `TypeA | TypeB` key, but `TypeA vs TypeB` and `TypeB vs TypeA` would be different keys (order matters). Should normalize.
24. Consider showing type differences count without the full type names when there are many (e.g., "4 type differences" instead of listing all)
25. Add `--max-generics-hint-types` flag to control how many type pairs are shown
26. Progress output: consider `\r` carriage return for inline progress instead of multiple lines
27. The `✅` status message on line 72 of `run_analysis.go` has a leading space — inconsistent with 4-space indentation of other messages
28. `printBuildingStatus` and `printSearchStatus` should share an indentation constant
29. Consider a `--compact` flag that shows just `file:line` without previews/hints/explanations
30. The `writeExplanation` function builds parts as `[]string` then joins with `|` — consider structured formatting
31. `cls.Suggestion` is shown as `fix:` for actionable and `why:` for non-actionable — consider `hint:` for generics candidates
32. Add `--output-width` flag to control preview truncation and hint wrapping
33. Golden file tests should cover singular and plural token/line counts
34. `previewFirstLine` truncates at 60 runes — consider making this configurable
35. The `generics:` hint always uses `same algorithm, different types:` prefix — redundant with the `generics:` label
36. Consider adding the generics hint to the JSON output as `generics_hint` field (verify it's there)
37. Add integration test: `--type-aware` alone vs `--suggest-generics` — verify they produce different clone sets
38. Document the `--suggest-generics` filtering behavior in HOW_TO_USE.md
39. Add `--suggest-generics` to the SDK `Options` struct (verify it's there)
40. Consider `--highlight-generics` as an alternative flag name that enhances rather than filters
41. The `formatGenericsHint` function limits to 3 unique pairs — consider making this configurable
42. `TypeDivergence.Position` field is never shown to the user — consider including it for debugging
43. Add `--debug-generics` flag that shows all type divergences, not just 3
44. Consider grouping divergences by type pair count (most common first)
45. The `flattenCloneNodes` function is O(n) but called for every clone instance — consider caching
46. `ClassifyGenericsCandidate` compares against `seqs[0]` as base — consider pairwise comparison for better coverage
47. Add test: generics hint with identical types from different packages (should still be a candidate)
48. Add test: generics hint with empty VarType on some nodes (should skip, not crash)
49. Review all `fmt.Fprintf` calls in `text.go` for consistent error wrapping patterns
50. Consider extracting output formatting into a separate `formatter` package for testability

---

## g) Questions (Cannot Determine Without User Input)

1. **Should `--suggest-generics` be redesigned as an enhancer (show all clones, highlight generics candidates) rather than a filter?** This is a breaking behavioral change. The current filter design may be intentional ("I only want to see generics opportunities"), but the user's confusion ("removing the flag found MORE things") suggests the mental model doesn't match. This requires a product decision.

2. **Should the `generics:` hint line be merged into the `explain:` line when `--explain` is used, or kept as a separate line?** Currently they're separate lines, which can be redundant since both describe the same clone group. Merging would make `--explain --suggest-generics` output more compact but less scannable.

3. **Is there a maximum line length convention for terminal output?** The generics hint can produce very long lines (even after shortening type paths). Should we truncate/wrap at 80 chars, 120 chars, or leave it unbounded?
