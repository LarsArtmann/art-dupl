# Status Report: Actionability False-Positive Fixes — Issues 1, 2, 5, 7

**Date:** 2026-08-07 21:20
**Session start:** ~21:00
**Branch:** master
**Base commit:** `8c0bf45c` (docs(feedback): add clone noise wall analysis)
**Head commit:** `c0e9e374` (feat(printer): add specific refactoring suggestions)
**Commits this session:** 4 (13e3dcac, 71e12c96, e1df0845, c0e9e374)
**Files changed:** 8 files, +151 / -17 lines
**Test status:** 309/309 PASS (`go test ./...`)
**Lint status:** 5 issues found, NOT FIXED (see below)

---

## a) FULLY DONE

### Issue 1: Type alias false positive — FIXED

**Root cause:** `syntax/syntax.go:serial()` creates shallow copies of nodes for the token stream but was missing the `IsAlias` field (added to the `Node` struct after `serial()` was written). Token stream nodes always had `IsAlias=false`, so `isAliasTypeSpec()` in the actionability layer could never identify type aliases. Every type-alias re-export was reported as "actionable | unknown".

**Fix:** Added `IsAlias: n.IsAlias` to the shallow copy at `syntax/syntax.go:220`.

**Verification:** `art-dupl --explain -t 1` on two files with `type ErrorType = errors.ErrorType` now reports 0 clone groups. Against the real file-and-image-renamer codebase, the `ErrorType = domainerrors.ErrorType` clone group is eliminated.

**Regression test:** `TestSerializePreservesIsAlias` in `syntax/syntax_test.go` — constructs a TypeSpec node with `Statement=true` and `IsAlias=true`, serializes it, and asserts the stream node preserves `IsAlias`.

**Commit:** `13e3dcac`

---

### Issue 2: Helper-call-with-different-args false positive — FIXED

**Two root causes found:**

1. **BasicLit values never stored in `Name`** (`syntax/golang/transform.go:43`): The `case *ast.BasicLit` handler encoded the literal value into the `Type` field (via `encodeSemanticType`) but never set `o.Name`. Since `collectStringLiterals()` checks `node.Name != ""`, the parameterizability engine could never find actual literal values. It was silently broken since inception — `collectStringLiterals` always returned empty slices, `sameLiteralCount` always returned true (all empty), and `literalsDifferOnlyInValues` always returned false (all empty). The parameterizability veto NEVER fired for any clone, ever.

2. **Parameterizability veto too aggressive** (`extractability_engine.go:checkParameterizability`): Even after fixing (1), the veto would fire for ANY literal difference — including clones with control flow (if/for/switch) where the structural pattern is genuinely worth extracting. Example: `processUser`/`processProduct` have identical `if name == "" { return fmt.Errorf(...) }` structure with different error messages — these are actionable clones, not "already parameterized".

**Fixes:**

- `transform.go:43`: Added `o.Name = n.Value` for BasicLit nodes.
- `extractability_engine.go:checkParameterizability`: Added `anySeqHasControlFlow()` guard — skip the parameterizability veto when any clone instance contains control-flow statements (IfStmt, SwitchStmt, TypeSwitchStmt, ForStmt, RangeStmt, SelectStmt).

**Verification:** `art-dupl --explain -t 1` on `printHeading("Hash Database Statistics", "=", 50)` vs `printHeading("History Log Statistics", "=", 50)` now reports 0 clone groups. BDD tests for `processUser`/`processProduct` (control-flow clones with different literals) still pass.

**Regression test:** `TestCheckParameterizability_ControlFlowOverridesVeto` in `extractability_integration_test.go` — builds two IfStmt trees with different error-message literals and asserts `Parameterizable=true`.

**Commits:** `71e12c96` (BasicLit Name), `e1df0845` (control-flow guard)

---

### Issue 5: `--recommend-threshold` gives single value — FIXED

**Problem:** Output was a single threshold value with no guidance on audit vs CI usage.

**Fix:** `cmd/recommend_threshold.go:runRecommendThreshold` now outputs two thresholds:

```
Codebase: 225 Go files

CI gate:       art-dupl -t 5 .
Deep audit:    art-dupl -t 1 --explain .

  CI gate filters noise (test boilerplate, guard clauses, etc.)
  Deep audit finds everything; pair with --explain to triage manually.
```

**Commit:** Part of `e1df0845` (same commit as Issue 2 control-flow guard)

---

### Issue 7: Generic fix suggestions — FIXED

**Problem:** All actionable clones in Assignment, Expression, Block, Call, Return, and Defer categories got "Review and extract common logic". Only Function, Struct, Interface, Loop, and Conditional had specific suggestions.

**Fix:** `clone_classify.go:getSuggestion` now returns category-specific suggestions:

- Assignment: "Extract shared assignment or initialization pattern"
- Expression: "Extract repeated expression or builder chain"
- Block: "Extract shared statement block to a helper function"
- Call: "Extract repeated call sequence or parameterize arguments"
- Return: "Consolidate return patterns or extract wrapper"
- Defer: "Extract shared cleanup or resource-management pattern"

**Commit:** `c0e9e374`

---

### Verification Against file-and-image-renamer

**Before this session (from feedback doc):** 4 remaining clone groups at `-t 1`:

- `ErrorType = domainerrors.ErrorType` x2 — FALSE POSITIVE
- `printHeading("X", "=", 50)` with different headings — FALSE POSITIVE
- `basename := filepath.Base(imagePath)` x2 — BORDERLINE
- `d.processed[path] = time.Now(); d.mu.Unlock()` x2 — BORDERLINE

**After this session:** 3 clone groups at `-t 1 --type-aware --explain`:

1. Loop pattern in injector.go + mock.go — actionable, suggestion: "Extract loop body to helper function"
2. Nil-guard conditional in hashdb + history — actionable, suggestion: "Consider strategy pattern or early returns"
3. Config-nil-guard conditional across hashdb + history + logger — actionable, suggestion: "Consider strategy pattern or early returns"

**Both false positives eliminated.** The two borderline clones are gone too (likely suppressed by the now-working parameterizability engine). All suggestions are now category-specific, not generic.

---

## b) PARTIALLY DONE

### Lint compliance — 5 issues found, NOT FIXED

Ran `golangci-lint` on changed packages only (not full project). Found:

| # | File                           | Linter    | Issue                                                                               |
| - | ------------------------------ | --------- | ----------------------------------------------------------------------------------- |
| 1 | `clone_classify.go:166`        | gocyclo   | `getSuggestion` complexity 16 (>15) — needs refactor (likely extract to map lookup) |
| 2 | `clone_classify_test.go:122`   | gofumpt   | File not properly formatted                                                         |
| 3 | `extractability_engine.go:303` | modernize | Loop can use `slices.ContainsFunc`                                                  |
| 4 | `extractability_engine.go:324` | modernize | Loop can use `slices.ContainsFunc`                                                  |
| 5 | `recommend_threshold.go:80`    | wsl_v5    | Missing whitespace above if-statement                                               |

**Impact:** All 5 are in files I changed this session. Issues 3-4 predate my changes (the loops were already there). Issues 1-2 and 5 are directly caused by my edits.

---

### AGENTS.md update — NOT DONE

The following critical discoveries should be recorded in art-dupl's AGENTS.md but were not:

1. **The `serial()` shallow-copy pattern is fragile**: Any new field added to `Node` must be manually added to both `serial()` and `Clone()`. Forgetting one silently breaks downstream consumers (this happened with `IsAlias`). Consider a generated copy or a different serialization strategy.

2. **BasicLit `Name` field is now populated with literal values**: Previously empty, now carries `n.Value` from `go/ast`. This enables `collectStringLiterals` and the parameterizability engine. Any code that previously assumed `Name == ""` for BasicLit nodes needs review.

3. **The parameterizability veto now requires NO control flow**: Clones with if/for/switch/select statements are exempt from the "already parameterized" veto, even when all string literals differ.

---

## c) NOT STARTED

### Issue 3: Threshold cliff between t=2 and t=3 (High effort)

Not addressed. The feedback doc describes a binary cliff: t=2 produces 3 groups, t=3 produces 0. Suggested approaches: fractional thresholds (token-based instead of statement-based), `--explain-threshold` mode, or recommending a range. The `--recommend-threshold` output now shows a range (CI gate vs deep audit), which partially addresses the UX concern, but the underlying cliff remains.

### Issue 4: Structural clone detection / call-sequence matching (High effort)

Not addressed. The feedback doc identifies a 15-step pipeline duplicated across `rename_explain.go` and `rename_process.go` that art-dupl cannot detect because interleaving print/branch code breaks token-level matching. Suggested approaches: CFG matching, call-sequence fingerprinting, data-flow matching. This is the highest-value but highest-effort improvement.

### Issue 6: `--explain-threshold` mode (Medium effort)

Not addressed. No mechanism to show what was suppressed at a given threshold level.

---

## d) TOTALLY FUCKED UP

### Empty commit message on `e1df0845`

The auto-commit daemon created commit `e1df0845` with an empty commit message. This happened because the BasicLit Name fix and the control-flow guard were committed in rapid succession and the daemon may have caught an intermediate state. The commit contains real changes (`extractability_engine.go` + `recommend_threshold.go`) but has no message describing them.

**Impact:** Minor — git history has one unprofessional empty-message commit. The changes are correct and tested.

### Previous session's `isAtomicDeclaration` GenDecl change may be dead code

Commit `b6853f43` (from earlier today) extended `isAtomicDeclaration` to handle `GenDecl` wrapping single `TypeSpec` aliases. But the actual root cause was the `IsAlias` field being dropped during serialization. With `IsAlias` now properly propagated, the token stream node for a type alias has `BaseType=TypeSpec` (not GenDecl) because GenDecl's children are marked `Statement=true` and emitted as individual tokens. The GenDecl case in `isAtomicDeclaration` may never be reached in practice. Needs verification and potential cleanup.

### No full-pipeline integration test

The regression tests (`TestSerializePreservesIsAlias`, `TestCheckParameterizability_ControlFlowOverridesVeto`) test individual layers in isolation. Neither test exercises the full pipeline (parse Go source → transform → serialize → suffix-tree match → reconstruct CloneNode → evaluate actionability). The type alias false positive was specifically a cross-layer bug that individual unit tests cannot catch. A `TestTypeAliasReExportNotActionable` that feeds two `.go` files through the full detection pipeline and asserts 0 actionable clones would be the gold-standard regression test.

---

## e) WHAT WE SHOULD IMPROVE

### Process improvements

1. **Always run `golangci-lint`, not just `go vet`** — `go vet` caught nothing; `golangci-lint` found 5 issues including complexity and formatting. This should be automatic before declaring "done".

2. **Check ALL struct fields when a serialization/copy function is the bug** — The `IsAlias` field was missing from `serial()`. After finding that, I should have audited every field in `Node` against both `serial()` and `Clone()` to find any other missing fields. I only fixed `IsAlias` because that was the failing test case.

3. **Full-pipeline integration tests for cross-layer bugs** — The two root causes (IsAlias serialization, BasicLit Name) were both cross-layer bugs that unit tests with hand-constructed data structures cannot catch. The test suite needs end-to-end tests that parse Go source and verify actionability verdicts.

4. **When fixing one consumer of a broken pipeline, check ALL consumers** — The `IsAlias` serialization bug affected every pattern matcher that checks `n.IsAlias` (at minimum `isAliasTypeSpec`, `isTypeAliasBlock`). The BasicLit Name bug affected every consumer of `collectStringLiterals` (at minimum `checkParameterizability`). I only verified the specific failing case, not all downstream consumers.

5. **The `serial()` shallow-copy pattern needs a structural fix** — Every time a field is added to `Node`, someone must remember to add it to both `serial()` and `Clone()`. This is a footgun. Options: generate the copy, use reflection, or restructure so the stream carries the original node pointer instead of a copy.

---

## f) Next 50 Things To Do

#### Immediate (lint + cleanup)

1. Fix `gocyclo` on `getSuggestion` — refactor switch to map lookup
2. Fix `gofumpt` formatting on `clone_classify_test.go`
3. Fix `wsl_v5` whitespace in `recommend_threshold.go`
4. Apply `slices.ContainsFunc` modernization in `extractability_engine.go` (2 sites)
5. Verify whether the `isAtomicDeclaration` GenDecl case (commit b6853f43) is dead code; remove if so
6. Run `golangci-lint run --timeout 5m ./...` on the full project, not just changed packages
7. Run `gofumpt -extra -w .` (the project requires the `-extra` flag per file-and-image-renamer AGENTS.md patterns)

#### Integration tests

8. Write `TestTypeAliasReExportNotActionable` — full pipeline test: parse two .go files with `type X = pkg.Y`, verify 0 actionable clones
9. Write `TestHelperCallDifferentArgsNotActionable` — full pipeline test: parse two .go files with `printHeading("A", "=", 50)` and `printHeading("B", "=", 50)`, verify 0 actionable clones
10. Write `TestControlFlowCloneWithDifferentLiteralsActionable` — full pipeline test: parse `processUser`/`processProduct` pattern, verify clones ARE found (not over-suppressed)
11. Audit all other pattern matchers that use `n.IsAlias` — verify they now work correctly
12. Audit all other consumers of `collectStringLiterals` — verify they now receive actual values
13. Audit ALL `Node` struct fields against `serial()` copy — find any other missing fields
14. Audit ALL `Node` struct fields against `Clone()` copy — find any other missing fields

#### Documentation

15. Update art-dupl AGENTS.md with the `serial()` fragility discovery
16. Update art-dupl AGENTS.md with the BasicLit Name field change
17. Update art-dupl AGENTS.md with the parameterizability control-flow guard rule
18. Mark the feedback doc issues 1, 2, 5, 7 as resolved
19. Update art-dupl CHANGELOG if one exists

#### Issue 3: Threshold cliff (High effort)

20. Research token-based thresholds (`--min-tokens` flag) as alternative to statement-count thresholds
21. Implement `--explain-threshold` mode that shows what's suppressed at each threshold level
22. Consider percentile-based threshold recommendation (compute the distribution of clone sizes and recommend the 90th percentile)
23. Add threshold sweep command (`art-dupl --threshold-sweep 1-10`) that shows clone counts at each level
24. Consider a "noise ratio" metric: at threshold N, what percentage of clones are actionability-suppressed?

#### Issue 4: Structural clone detection (High effort)

25. Research call-sequence fingerprinting: hash the ordered list of qualified function calls in each function body
26. Prototype CFG-based matching: compare call sequences ignoring interleaving print/branch statements
27. Research data-flow matching: track variable dependencies between calls
28. Consider gap-tolerant suffix tree matching (allow N non-matching tokens between matching windows)
29. Evaluate whether PDG (Program Dependence Graph) matching is feasible for Go
30. Benchmark the approach against the `rename_explain.go` / `rename_process.go` test case

#### Issue 6: `--explain-threshold` mode (Medium effort)

31. Design the output format: "At t=5: N groups suppressed (M actionability-filtered, K threshold-filtered)"
32. Implement threshold-level breakdown in the `--explain` output
33. Add `--show-suppressed` enhancement to include threshold-suppressed clones (currently only shows actionability-suppressed)

#### Actionability engine improvements

34. Add `helper-call-different-args` as a first-class denylist pattern (currently handled via parameterizability engine, which is indirect)
35. Consider whether the parameterizability control-flow guard should also check for BlockStmt (nested control flow inside blocks)
36. Verify `isTypeAliasBlock` pattern now works correctly with the IsAlias fix (it was also silently broken)
37. Add more denylist patterns for common Go idioms: `http.HandlerFunc` wrappers, `select{}` stubs, `panic("not implemented")`
38. Consider whether the `anySeqHasControlFlow` function should also count `LabeledStmt` (break/continue labels imply loop control flow)

#### `--recommend-threshold` improvements

39. Add JSON output format for `--recommend-threshold` (currently text-only)
40. Consider file-type-aware recommendations (Go vs templ vs mixed)
41. Add `--threshold-sweep` as a subcommand or flag

#### Code quality

42. The `getSuggestion` function has 13 categories in a switch — consider a `map[CloneCategory]string` lookup table
43. The `patternLabelConfigs` map in `clone_classify.go` has 20+ entries — consider generating from a config file
44. Consider adding property-based testing (rapid/gopter) for the actionability engine
45. The `extractability_engine.go` is 490+ lines — consider splitting into `parameterizability.go`, `controlflow.go`, `roi.go`

#### CI / Release

46. Run full `nix build` and `nix run .#test` (haven't tested nix integration this session)
47. Consider cutting a v0.6.2 or v0.7.0 release with these fixes
48. Update the baseline snapshots if any exist (the `baseline/` package suggests snapshot testing)
49. Run `art-dupl` against itself (dogfooding) to check for new false positives from the BasicLit Name change
50. Run `art-dupl` against 2-3 other codebases to validate the fixes generalize

---

## g) Questions (3)

### Q1: Should I fix the 5 lint issues right now, or batch them with the next sprint?

The `gocyclo` on `getSuggestion` requires a small refactor (switch → map lookup). The `gofumpt` and `wsl_v5` issues are trivial formatting fixes. The `modernize` issues are pre-existing. I can fix all 5 in under 5 minutes if you want, or defer them.

### Q2: Should I clean up the `isAtomicDeclaration` GenDecl case from commit b6853f43?

The GenDecl handling was added in a previous session as a speculative fix for Issue 1. The actual root cause was the `IsAlias` serialization bug, so the GenDecl case may be dead code. I need to verify whether any token stream node for a type alias has `BaseType=GenDecl` (vs `BaseType=TypeSpec`). If dead, should I remove it or keep it as a belt-and-suspenders guard?

### Q3: Is Issue 4 (structural clone detection) worth prioritizing next?

The feedback doc calls it "the single largest maintenance burden" that art-dupl misses. But it's also the highest-effort item (CFG/call-sequence/data-flow matching). Should I research and prototype call-sequence fingerprinting as the next sprint, or focus on lower-hanging fruit (Issue 3 threshold cliff, Issue 6 explain-threshold, more denylist patterns)?

---

## Resolution (2026-08-10)

**All 4 issues (1, 2, 5, 7) shipped.** IsAlias serialization fix, BasicLit Name fix, control-flow parameterizability guard, --recommend-threshold range, and category-specific suggestions are in CHANGELOG `[Unreleased]` → Fixed/Changed. Commits: `13e3dcac`, `71e12c96`, `e1df0845`, `c0e9e374`. Issues 3 (threshold cliff), 4 (structural clone detection), 6 (--explain-threshold) → ROADMAP (high-effort, long-term).
