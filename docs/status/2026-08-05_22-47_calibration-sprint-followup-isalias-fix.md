# Status Report: Calibration Sprint Follow-up — IsAlias Fix, Documentation, Tests

**Date:** 2026-08-05 22:47
**Session Goal:** Complete remaining gaps from the calibration sprint (items left partially done or not started)
**Result:** 8 items completed, 1 deliberately skipped, lint has a pre-existing issue

---

## A) FULLY DONE

### 1. Fixed type-alias-block false-negative risk (IsAlias field)
- **Problem:** The `type-alias-block` and `single-declaration` actionability patterns suppressed ALL non-composite TypeSpec nodes, regardless of whether they were true aliases (`type X = pkg.Y`) or named type definitions (`type X pkg.Y`). The `ast.TypeSpec.Assign` token position was discarded during transformation, making the two indistinguishable.
- **Fix:** Added `IsAlias bool` field through the full pipeline:
  - `syntax.Node.IsAlias` — set by transformer from `n.Assign != token.NoPos` (`syntax/golang/transform.go:361`)
  - `Node.Clone()` — copies `IsAlias` in deep copy (`syntax/syntax.go:139`)
  - `domain.CloneNode.IsAlias` — carries the flag to the actionability layer (`domain/clone_node.go`)
  - `syntaxToCloneNode` bridge — maps `syntax.Node.IsAlias` → `domain.CloneNode.IsAlias` (`printer/clone_processor.go:55`)
- **Pattern updates:**
  - `isAtomicDeclaration`: now requires `n.IsAlias` for TypeSpec nodes. Only true aliases (`type X = pkg.Y`) are suppressed. Named type definitions (`type Severity string`) remain actionable. (`actionability_boilerplate.go:268-279`)
  - `isPackageTypeAlias`: now requires `n.IsAlias`. Only true alias blocks are suppressed. (`actionability_boilerplate.go:372-389`)
- **Impact:** Eliminates the false-negative risk where duplicated named type definitions (true positives from calibration) would be wrongly suppressed as "alias blocks."
- **Tests updated:** `pkgAlias` helper now sets `IsAlias: true`. Added negative test case for named type definition (`type Mode pkg.Mode` — should NOT be suppressed by `isSingleDeclaration`). Updated `TestIsTypeAliasBlock` non-alias case to explicitly set `IsAlias: false`.

### 2. Added missing isTerminalStatement ExprStmt(UnaryExpr) test
- **Problem:** The `isTerminalStatement` function was extended to match `ExprStmt` wrapping a single `UnaryExpr` (catches `<-ch` channel receive) but no test verified this behavior.
- **Fix:** Added two test cases in `TestIsSingleSimpleStatement`:
  - `ExprStmt wrapping UnaryExpr (<-ch channel receive)` — expected: true
  - `ExprStmt wrapping CallExpr` — expected: false (verifies the `len(Children) == 1` guard)

### 3. Updated docs/ACTIONABILITY_PATTERNS.md with 2 new patterns
- **Problem:** The patterns doc claimed "23 pattern checks" but the table actually has 25 (interface-assertion and type-alias-block were added in the prior sprint but not documented).
- **Fix:**
  - Updated pattern count: 23 → 25 (both in "How It Works" and "Property-Based" sections)
  - Updated priority order list: added entries 12 (interface-assertion) and 14 (type-alias-block), renumbered 12→15 through 23→25
  - Updated `single-declaration` row: now says "true TypeSpec alias" and mentions `IsAlias`
  - Added `interface-assertion` row with example `var _ Reader = (*MyReader)(nil)`
  - Added `type-alias-block` row with example `type ( ToolCall = proto.ToolCall; ... )`
  - Updated `test-helper-delegate` row: now mentions "matches t/b/tb receivers"

### 4. Added ShowSuppressed support to cmd/diff_report.go
- **Problem:** `diff_report.go` had duplicate suppression gates (actionability + shouldSuppressGroup) that were NOT updated with `&& !suppression.ShowSuppressed`. So `--show-suppressed` had no effect with `--diff-report`.
- **Fix:** Added `&& !suppression.ShowSuppressed` to both `continue` gates in `collectCurrentGroups` (`cmd/diff_report.go:94` and `:106`). Now `--diff-report --show-suppressed` surfaces suppressed groups in diff output.

### 5. Updated HOW_TO_USE.md with new CLI flags
- **Problem:** Neither `--include-examples` nor `--show-suppressed` were documented in the user guide.
- **Fix:** Added two new sections after ".gitignore Honoring":
  - "Example/Demo Directory Exclusion" — explains `examples/`, `demo/`, `demos/` exclusion and `--include-examples` override
  - "Showing Suppressed Groups" — explains `--show-suppressed` for calibration and recall measurement

### 6. Added --show-suppressed behavioral tests (BDD)
- **Problem:** No integration test verified that `--show-suppressed` actually surfaces suppressed clone groups.
- **Fix:** Added 3 BDD tests in `bdd/actionability_test.go` using the guard-clause pattern (already proven to be suppressed by actionability):
  - "should surface suppressed guard-clause clones when flag is set" — verifies `guard1.go` and `guard2.go` appear in output with `--show-suppressed`
  - "should suppress guard-clause clones by default (without --show-suppressed)" — verifies 0 clone groups by default
  - "should show at least as many clones as default mode" — verifies `--show-suppressed` >= default count

### 7. Full verification
- `go build ./...` — clean
- `go test ./...` — 27 packages pass, 0 failures (305 BDD specs pass)
- `golangci-lint` — **pre-existing issue** (see D) section below

### 8. Deliberately skipped: --show-suppressed in addSharedFlags
- **Decision:** Kept `--show-suppressed` root-only (not in `addSharedFlags`). The AGENTS.md explicitly says `--explain` and `--no-actionability` are "intentionally root-only (main analysis) — stats/baseline/check have different output semantics where per-clone actionability filtering doesn't apply." `--show-suppressed` follows the same pattern. The `stats` subcommand shows aggregate statistics, not individual clones. The `baseline` subcommand records clones for CI comparison.
- **Risk:** Low. If a user wants to see suppressed groups in stats output, they can use `--no-actionability` on the root command instead.

---

## B) PARTIALLY DONE

### 1. SDK Options for IncludeExamples and ShowSuppressed
- **Finding:** The SDK (`pkg/artdupl/`) takes explicit file lists (no crawling), so `IncludeExamples` is irrelevant — callers filter directories themselves. `ShowSuppressed` is a CLI-only concern because the SDK doesn't run actionability filtering at all (the `detectorConfig` has no suppression fields, and the SDK pipeline has no actionability evaluation step).
- **What was done:** Investigated the SDK architecture. Decided not to add fields that would be dead code.
- **What remains:** If SDK consumers want programmatic calibration (actionability analysis + ShowSuppressed), that requires wiring the full actionability engine into the SDK pipeline — a much larger change beyond this sprint's scope.

---

## C) NOT STARTED

1. **Re-run calibration** on the 6 projects to verify precision improvement (target: 91%+). Cannot do this without access to the calibration projects.
2. **Commit 87 labeled clone groups as regression dataset.** Requires the calibration data.
3. **CI test on regression dataset.** Depends on #2.
4. **Add `test_plugin/` and `test_temp/` to `exampleDirNames`.** The calibration report mentioned these as throwaway dirs, but they're project-specific names that may not generalize. Did not add to avoid over-exclusion.
5. **Refactor `shouldSkipPath` to use a struct** instead of 4 bools. Code smell noted but not fixed — it works, and refactoring it touches many call sites.
6. **Extract shared `shouldSkipGroup` helper** from `run_output.go` + `diff_report.go`. DRY violation noted but not fixed — the two paths have slightly different semantics (diff_report builds a list, run_output prints directly).

---

## D) TOTALLY FUCKED UP

### 1. Pre-existing: tagliatelle linter re-added by auto-commit daemon
- **Problem:** Commit `6c4cabc7` (blank message, auto-commit daemon) re-added `tagliatelle` to `.golangci.yml` line 108. The AGENTS.md explicitly says: "tagliatelle is NOT in the `.golangci.yml` enable list (codebase has mixed snake_case/camelCase JSON conventions per ADR-0016)."
- **Impact:** `golangci-lint run` now reports 50 tagliatelle issues (all pre-existing snake_case JSON tags in `printer/stats_data.go` and other files).
- **Root cause:** The auto-commit daemon appears to have reverted the `ebd6a474` commit that removed tagliatelle. This is NOT from my changes — I did not touch `.golangci.yml`.
- **Fix needed:** Remove `- tagliatelle` from `.golangci.yml` enable list, or let the `scripts/check-disabled-linters.sh` guard handle it.

### 2. No mistakes in my own work
- All edits used exact matches via `edit`/`multiedit` tools (no sed/Python hacks).
- All test fixtures were updated correctly for the `IsAlias` field change.
- Build and tests pass cleanly.

---

## E) WHAT WE SHOULD IMPROVE

1. **The `IsAlias` field should be documented in AGENTS.md.** The actionability patterns section mentions `single-declaration` and `type-alias-block` but doesn't mention the `IsAlias` field that now powers them. A future developer reading the code won't know why `IsAlias` exists without tracing through the transformer.

2. **The `isAtomicDeclaration` change may cause regression on single TypeSpec aliases.** Before my change, `isSingleDeclaration` suppressed ALL non-composite TypeSpec nodes (including `type Mode = domain.Mode` AND `type Mode domain.Mode`). Now it only suppresses true aliases (`IsAlias == true`). This means single named-type definitions like `type Severity string` defined in 2+ files will NO LONGER be suppressed by `single-declaration` — they'll be reported as actionable. This is CORRECT per the calibration (duplicated type definitions are true positives), but it may increase the clone count for projects that had these previously suppressed.

3. **The `subtreeHasType` helper is now slightly misleading.** It was originally added for `isTypeAliasBlock` to detect SelectorExpr children. But now that `isPackageTypeAlias` also requires `IsAlias`, the SelectorExpr check is a secondary signal. The primary signal (IsAlias) is sufficient for true aliases. The SelectorExpr check adds a false sense of precision — `type X = pkg.Y` always has IsAlias=true AND a SelectorExpr, but `type X = SomeLocalType` also has IsAlias=true but NO SelectorExpr. The current code would NOT suppress `type X = SomeLocalType` (no SelectorExpr), which may be a false negative.

4. **The `shouldSkipPath` 4-bool parameter smell remains.** This was noted in the prior sprint report and I chose not to fix it. It's a latent argument-swap bug.

5. **The diff_report.go suppression gates are still duplicated** from run_output.go. Both files now have `&& !suppression.ShowSuppressed` on the same two gates, but the logic is copy-pasted, not shared.

---

## F) Up to 50 Things to Get Done Next

### High Priority (precision/safety)
1. Remove `tagliatelle` from `.golangci.yml` (pre-existing regression from auto-commit daemon)
2. Re-run calibration on 6 projects to verify precision improvement after IsAlias fix
3. Commit 87 labeled clone groups as machine-readable regression dataset (JSON/YAML)
4. Write CI test that runs on regression dataset and asserts no precision regression
5. Add `IsAlias` documentation to AGENTS.md actionability patterns section
6. Document the `subtreeHasType` SelectorExpr limitation for local-type aliases
7. Add integration test: run art-dupl on a project with `examples/` dir, verify exclusion
8. Add test for `isPackageTypeAlias` with `Ident` type references (no SelectorExpr — should NOT match)
9. Add test for `pathContainsExampleDir` with nested paths (`pkg/examples/sub/file.go`)
10. Verify the Nix self-test (`art-dupl -t 1 --plumbing .`) still passes after IsAlias change

### Medium Priority (completeness)
11. Wire actionability engine into SDK pipeline (enables `ShowSuppressed` + `IncludeExamples` in SDK)
12. Add `ShowSuppressed` to SDK `Options` struct (requires #11)
13. Refactor `shouldSkipPath` to use a `PathFilter` struct instead of 4 bools
14. Extract shared `shouldSkipGroup` helper from `run_output.go` + `diff_report.go`
15. Make `exampleDirNames` configurable via `--exclude-dirs` flag
16. Add `--show-suppressed` to stats subcommand if users request it (currently root-only)
17. Add BDD test for `--include-examples` flag
18. Add BDD test for `interface-assertion` pattern suppression
19. Add BDD test for `type-alias-block` pattern suppression
20. Update the stale comment in `run_crawl.go:103` (says `crawlSinglePathWithOpts` but the function is `filesFeedWithOptions`)
21. Run calibration with `--show-suppressed` to measure recall for the first time
22. Investigate the 3 "coincidental report WriteString" FPs — could a conservative pattern help?
23. Consider excluding `testdata/` directories by default (common in Go projects)
24. Add a `--list-excluded-dirs` flag to show which directories are excluded by default
25. Add test for the `interface-assertion` pattern when ValueSpec has multiple names (`var _, x I = ...`)

### Architecture / Refactoring
26. Consider making actionability patterns extensible via plugins (the pattern table is now 25 entries)
27. Move `exampleDirNames` from a global var to a config field
28. Consider a `SkipConfig` or `CrawlFilter` struct that bundles all skip-related fields
29. Extract `pathContainsExampleDir` and vendor/node_modules checks into a unified `pathFilter` type
30. Consider whether the 4 bools in `CrawlOptions` (IncludeVendor, IncludeNodeMods, IncludeExamples, and potentially more) should be a nested `IncludeConfig` struct
31. Add `Assign` field to `syntax.Node` serialization (for cache round-trip) if needed
32. Consider whether `IsAlias` should be encoded into the semantic Type hash (so alias and non-alias TypeSpecs produce different tokens and don't match as clones at all)

### Testing
33. Add fuzz test for actionability pattern evaluation (never panics on arbitrary CloneNode trees)
34. Add test verifying `--show-suppressed` surfaces min-lines suppressed groups
35. Add test verifying `--show-suppressed` surfaces accept-directive suppressed groups
36. Add test verifying `--show-suppressed` has no effect without `--semantic`
37. Add test for `isTerminalStatement` with `*ptr` dereference (UnaryExpr)
38. Add test for `isTerminalStatement` with `!flag` negation (UnaryExpr)
39. Add test for `syntaxToCloneNode` preserving `IsAlias` field
40. Add test for `Node.Clone()` preserving `IsAlias` field
41. Add test verifying `IsAlias` is set correctly by transformer for `type X = Y` vs `type X Y`
42. Add race test for `IsAlias` field in concurrent cache-hit scenarios

### Documentation
43. Update `SDK_DESIGN.md` with note about actionability being CLI-only (and why)
44. Add ADR for the `IsAlias` field addition and its impact on alias vs named type detection
45. Update the calibration report with post-fix precision numbers
46. Add `--include-examples` and `--show-suppressed` to the README flag reference
47. Write an ADR for the demo-directory exclusion policy
48. Update `TODO_LIST.md` to mark calibration items as done
49. Update `CHANGELOG.md` with the IsAlias fix and diff_report ShowSuppressed support
50. Add `docs/adr/` entry for the property-based classification engine's interaction with IsAlias

---

## G) Questions

1. **Should `type X = SomeLocalType` (alias to a local type, no SelectorExpr) be suppressed by `type-alias-block`?** Currently it's NOT suppressed because `isPackageTypeAlias` requires a SelectorExpr child. But it IS a true alias (`IsAlias == true`). Should we drop the SelectorExpr requirement and suppress ALL true alias blocks, or keep the SelectorExpr check to be conservative?

2. **Should I remove `tagliatelle` from `.golangci.yml` now, or let the `scripts/check-disabled-linters.sh` guard handle it on the next Nix check?** The auto-commit daemon re-added it (commit `6c4cabc7`), and I'm not sure if that was intentional or a fluke. I don't want to revert a change that was made intentionally.

3. **Should the `isAtomicDeclaration` change (only suppress true aliases, not named type definitions) be considered a breaking change for users who relied on the old behavior?** Projects that had `type Severity string` defined in 2+ files previously had those suppressed as "single-declaration." Now they'll be reported as actionable clones. This is correct per the calibration, but it increases noise for those projects. Should we add a `--suppress-type-definitions` flag for users who want the old behavior?
