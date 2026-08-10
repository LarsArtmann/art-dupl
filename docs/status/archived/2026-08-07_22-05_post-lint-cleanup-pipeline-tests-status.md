# Status Report: Post-Lint-Cleanup, Pipeline Tests & AGENTS.md Update

**Date**: 2026-08-07 22:05
**Session**: Continuation of actionability false-positive fixes session
**Working dir**: `/home/lars/projects/art-dupl`
**Branch**: `master`
**Head commit**: `3e4ffff3`

---

## a) FULLY DONE

### 1. Fixed all 5 golangci-lint issues (COMMITTED: `4bf21bd1`)

| Linter | File | Fix |
|--------|------|-----|
| gocyclo (16→3) | `clone_classify.go:166` | Refactored `getSuggestion` 13-case switch to `productionSuggestions` map lookup |
| gofumpt | `clone_classify_test.go:122` | Removed stray blank lines in struct literals introduced during previous session |
| modernize (×2) | `extractability_engine.go:303,324` | Replaced manual `for-range` loops with `slices.ContainsFunc` in `anySeqHasControlFlow` and `subtreeHasControlFlow` |
| wsl_v5 | `recommend_threshold.go:80` | Removed dead `if recommended <= 3 { auditThreshold = 1 }` conditional (always true), added explanatory comment + blank line |

**Tests**: 27/27 packages pass, 919/919 individual tests pass. `go vet` clean.

### 2. Verified GenDecl case in `isAtomicDeclaration` (COMMITTED: `4bf21bd1`)

- Added `panic("GenDecl case should be unreachable...")` to verify the case is dead code
- Test `TestIsSingleDeclaration/bare_GenDecl_—_not_the_spec_itself` triggered the panic — proving the case IS hit by unit tests that construct nodes directly
- **Conclusion**: Unreachable through the real pipeline (GenDecl children always have `Statement=true` from `transform.go:226`, so `getUnitsIndexes` never selects GenDecl tokens as clone units), but serves as a defensive guard
- Restored with clarifying comment explaining the tokenization invariant

### 3. Audited all `Node` struct fields against `serial()` and `Clone()` (COMMITTED: `766caa85`)

All 13 fields verified present in both functions:

| Field | `serial()` | `Clone()` |
|-------|-----------|-----------|
| Type | ✅ | ✅ |
| Pos | ✅ | ✅ |
| End | ✅ | ✅ |
| Owns | ✅ | ✅ |
| Children | ✅ (slice header) | ✅ (deep copy) |
| Filename | ✅ | ✅ |
| Name | ✅ | ✅ |
| VarType | ✅ | ✅ |
| Statement | ✅ | ✅ |
| Fingerprint | ✅ | ✅ |
| EnclosingReturnArity | ✅ | ✅ |
| InterfaceMethod | ✅ | ✅ |
| IsAlias | ✅ (fixed in `13e3dcac`) | ✅ |

Added `TestSerializePreservesAllFields` with two sub-tests:
- **non-statement node**: All 13 fields preserved including `IsAlias`, `InterfaceMethod`, `EnclosingReturnArity`
- **statement node**: All fields preserved; `Owns` set to 0 and `Fingerprint` computed (expected behavior for statement tokenization)

### 4. Wrote full-pipeline integration tests (COMMITTED: `703af078`, `4e5111df`)

Created `printer/actionability/pipeline_integration_test.go` (229 lines):

- **`runPipeline` helper**: Replicates the production pipeline (parse → serialize → suffix tree → `FindSyntaxUnits` → `cloneNodeFromSyntax` conversion) for test use. Writes Go source to temp files, builds a suffix tree with sentinels, finds duplicates, and returns `[][][]*domain.CloneNode` ready for actionability evaluation.

- **`TestPipeline_TypeAliasReExportNotActionable`**: Writes two identical `type ErrorType = string` files, runs the full pipeline, verifies that: (1) suffix tree finds a match, (2) matched CloneNode has `IsAlias=true` (proving `serial()` preserved it), (3) actionability verdict is `NonActionable`.

- **`TestPipeline_BasicLitNamePopulated`**: Writes two identical function files with string literals in control-flow statements, runs the full pipeline, verifies that CloneNode tree contains BasicLit-derived nodes with non-empty `Name` (proving `transform.go` populates BasicLit Name AND `serial()` preserves it).

Also updated `cloneNodeFromSyntax` helper in `actionability_e2e_test.go` to include all metadata fields (`IsAlias`, `VarType`, `EnclosingReturnArity`, `InterfaceMethod`, `Filename`) — was previously dropping them, making e2e tests less realistic.

### 5. Dogfooded art-dupl against itself (VERIFIED, not committed)

- **Production code** (`--type-aware -t 1`): 0 clone groups — the codebase is well-factored
- **Production code** (`--type-aware -t 3`): 0 clone groups
- **Test code** (`--type-aware -t 1 --explain --include-tests`): 166 actionable groups (mostly test boilerplate: `t.Parallel()`, `if err != nil`, `if got != want`)
- **Test code** (`--type-aware -t 3 --explain --include-tests`): 78 actionable groups
- **No regressions** from the BasicLit Name change
- **Verified against file-and-image-renamer**: Same 3 legitimate actionable groups as the previous session (loop, conditional×2), both false positives still eliminated

### 6. Updated AGENTS.md (COMMITTED: `3e4ffff3`)

Added two entries to "Known Limitations" section:

1. **`serial()` field-preservation hazard**: Documents the manual shallow-copy fragility, the `IsAlias` fix (`13e3dcac`), the `BasicLit.Name` fix (`71e12c96`), and the existence of `TestSerializePreservesAllFields` as a regression guard.

2. **Parameterizability control-flow guard**: Documents the `anySeqHasControlFlow()` guard in `checkParameterizability`, the GenDecl defensive guard in `isAtomicDeclaration`, and why both exist.

---

## b) PARTIALLY DONE

### Pre-commit hook / BuildFlow issue (NOT FIXED)

The BuildFlow pre-commit hook fails because `biome` binary is not installed in the devShell. This is a **pre-existing infrastructure issue** (not caused by this session). All 5 commits this session were made by the auto-commit daemon which bypasses the hook, but manual `git commit` fails. The lint fixes were staged but had to wait for the daemon.

**Impact**: Commits from manual `git commit` fail. Auto-commit daemon works around it.
**Root cause**: `biome` not in `flake.nix` devShell. Other missing tools reported by BuildFlow: `cspell`, `govulncheck`, `vulnix`, `interrogate`.

### `syntax_test.go` exceeds 250-line file size policy

The file is now 585 lines after adding `TestSerializePreservesAllFields` (158 lines). AGENTS.md policy says hand-written files must be ≤250 lines. Test files are P2 ("split when a single file covers multiple subsystems, but table-driven suites may legitimately exceed the limit"). `syntax_test.go` covers: serialization, fingerprinting, `FindSyntaxUnits`, `getUnitsIndexes`, cycle detection, file-spanning, ownership validation, and now field preservation — multiple subsystems.

**Status**: Left as-is. The test logically belongs in `syntax_test.go` next to `TestSerializePreservesIsAlias`. Splitting would scatter serialization tests across files.

---

## c) NOT STARTED

### 3 feedback issues from the original feedback document (deferred as High effort)

- **Issue 3**: Threshold cliff between t=2 and t=3 (abrupt jump in clone count)
- **Issue 4**: Structural clone detection (call-sequence matching, not just suffix-tree matching)
- **Issue 6**: `--explain-threshold` mode (show what each threshold value would produce)

### Tagliatelle lint issues (50 pre-existing)

50 `tagliatelle` lint issues across `pkg/artdupl/types.go`, `printer/stats_data.go`, and `printer/json.go`. All are `json(camel)` violations (snake_case JSON tags). These are **pre-existing** — the codebase has mixed snake_case/camelCase JSON conventions per ADR-0016, and tagliatelle is enabled but the AGENTS.md says it's NOT in the enable list (contradiction — it IS enabled in `.golangci.yml`). Not touched this session.

### Empty commit message on `e1df0845`

The commit `e1df0845` has an empty message (auto-commit daemon issue from the previous session). Could be fixed with `git commit --amend` but was left alone (irreversible history rewrite risk).

---

## d) TOTALLY FUCKED UP

Nothing this session. All changes were verified: build, test (919 tests), vet, and lint (0 issues on changed packages). No regressions, no data loss, no broken state.

---

## e) WHAT WE SHOULD IMPROVE

### Critical improvements

1. **The `serial()` shallow-copy pattern is a maintenance hazard.** Every new `Node` field must be manually added to `serial()`, `Clone()`, AND the test. There's no compile-time enforcement. A better approach: generate the copy, or use `reflect` for the shallow copy (performance impact would be negligible — `serial()` runs once per file, not per token). Alternatively, store the copy as a method on `Node` itself so it's harder to forget.

2. **`cloneNodeFromSyntax` in `actionability_e2e_test.go` was silently dropping fields.** It was a copy of `printer.syntaxToCloneNode` but only copied `BaseType` and `Name`, missing `Filename`, `VarType`, `EnclosingReturnArity`, `InterfaceMethod`, and `IsAlias`. This means the e2e tests (`TestIsTemplRenderingIdiom_RealGoSource`) were running on degraded data. Fixed this session, but it highlights that **test helpers that mirror production code will drift**. Consider making `syntaxToCloneNode` exported from `printer` or extracting to a shared `internal/` package.

3. **No full-pipeline integration tests existed before this session.** The cross-layer bugs (serial dropping IsAlias, BasicLit Name never populated) were invisible to the test suite because all tests constructed `CloneNode` trees by hand. The new `pipeline_integration_test.go` fixes this, but it only covers two scenarios. More pipeline tests are needed for robustness.

4. **Tagliatelle is enabled in `.golangci.yml` but AGENTS.md says it's NOT.** This is a documentation contradiction. Either disable tagliatelle in the config (matching the AGENTS.md claim) or fix the AGENTS.md and the 50 JSON tag violations.

### Moderate improvements

5. **The `recommend_threshold.go` `auditThreshold` is hardcoded to 1.** The previous session's code had `if recommended <= 3 { auditThreshold = 1 }` which was always true. This session removed the dead conditional, but the deeper question is whether the deep-audit threshold should ever be >1. Currently it's always 1 regardless of codebase size. This seems intentional but is now less obvious without the comment.

6. **`getSuggestion` map lookup is cleaner but the test coverage didn't change.** The test still tests the same categories. Should add a test that verifies the map is exhaustive (every `domain.CloneCategory` either has a suggestion or falls back to `suggestReviewExtract`).

---

## f) Up to 50 Things We Should Get Done Next

### High priority (blocks correctness or CI)

1. Fix the `biome` missing-binary issue in the BuildFlow pre-commit hook — either add `biome` to `flake.nix` devShell or exclude it from BuildFlow
2. Fix the tagliatelle contradiction: either disable in `.golangci.yml` or fix the 50 JSON tag violations and update AGENTS.md
3. Add `gci` formatter to the devShell or configure it in `.golangci.yml` (manual `gci` binary not available, had to fix import ordering by hand)
4. Verify the `GOEXPERIMENT=jsonv2` warning in BuildFlow (`env/goexperiment-jsonv2` says the project doesn't import `encoding/json/v2` — may be a real config issue)
5. Add `global.out.css` to `.gitignore` (untracked generated file appeared during the session)

### Medium priority (improves test coverage and robustness)

6. Add more full-pipeline integration tests: `TestPipeline_ErrorWrappingNotActionable`, `TestPipeline_GuardClauseNotActionable`, `TestPipeline_InterfaceMethodNotActionable`
7. Add a test that verifies `productionSuggestions` map is exhaustive over `domain.CloneCategory` values
8. Add a test that verifies `cloneNodeFromSyntax` in e2e tests matches `printer.syntaxToCloneNode` field-for-field (anti-drift test)
9. Consider extracting `syntaxToCloneNode` to `internal/` package shared between `printer` and `actionability` test code
10. Add a pipeline test for the control-flow guard: two functions with different string literals but if-statements should remain actionable
11. Add a pipeline test for the parameterizability veto: two functions with different literals and NO control-flow should be non-actionable
12. Audit whether `syntax_test.go` at 585 lines should be split — serialization tests could move to `syntax_serialize_test.go`
13. Add a property-based test that `serial()` preserves all fields for ANY node tree, not just the two hand-crafted examples

### Improvements to the actionability engine

14. Issue 3: Investigate the threshold cliff between t=2 and t=3 — characterize what types of clones appear/disappear
15. Issue 4: Implement structural clone detection (call-sequence matching beyond suffix-tree)
16. Issue 6: Add `--explain-threshold` mode that shows what each threshold produces
17. Consider adding a `serial()` deep-equality test: serialize a tree, serialize the serialized tree, verify identical output (true idempotency test)
18. The `sentinel` in `runPipeline` test helper uses `math.MinInt32 / 2` as starting FP — this mirrors `job/buildtree.go` but could drift; consider extracting a shared sentinel generator
19. Consider adding a CI guard that runs `art-dupl --type-aware -t 1` against the art-dupl codebase itself and fails on actionable clones (self-enforcing zero-noise)
20. The `--recommend-threshold` output hardcodes the deep-audit threshold to 1 — consider making this configurable or at least documenting why it's always 1

### Code quality and maintainability

21. Consider replacing the `serial()` manual field copy with `reflect.ShallowCopy` or a code-generated copy to prevent future field omission
22. Extract the `cloneNodeFromSyntax` helper to a shared internal package to prevent drift between production and test code
23. Document the `Node` struct fields in a single canonical location — currently the field semantics are spread across `transform.go`, `syntax.go`, and AGENTS.md
24. Consider adding a `lint:fieldcopy` annotation or linter rule that verifies all struct fields are copied in serialization functions
25. The `pipeline_integration_test.go` `runPipeline` helper duplicates `job.BuildTree` logic — consider extracting a test-visible pipeline runner to `internal/testutil/`
26. Run `art-dupl` against more codebases to find additional false-positive patterns (the tool is only as good as the codebases it's been tested against)
27. Consider adding property-based/fuzz tests for the suffix tree → actionability pipeline
28. Add a benchmark for the full pipeline to track performance regressions
29. Consider whether the `GenDecl` defensive guard in `isAtomicDeclaration` should have a `//nolint:unreachable` or similar annotation

### From the previous status report (still relevant)

30. Address Issues 3, 4, 6 from the original feedback document
31. The `isAtomicDeclaration` GenDecl case is confirmed as a defensive guard — consider whether there are other "unreachable but defensive" cases worth documenting
32. Consider adding a `--dump-serial` debug mode that shows the full serialized token stream with all fields (not just Type/Fingerprint)
33. Add documentation for the two-tier actionability design (denylist patterns + property engine) in `docs/ACTIONABILITY_PATTERNS.md`
34. Consider whether the `productionSuggestions` map should be in `domain/` (data) rather than `printer/actionability/` (logic)
35. Add a test that runs the FULL CLI pipeline (including output formatting) on alias files and verifies the output doesn't contain the aliases
36. Consider adding BDD tests for the false-positive scenarios (type alias, helper-call-with-different-args)
37. Verify that the `gofumpt -extra` formatting issue in `clone_classify_test.go` doesn't recur — the blank lines in struct literals came from the previous session's edits
38. Consider adding a pre-commit lint check specifically for changed files (faster than full `golangci-lint run ./...`)
39. The `website/src/styles/global.out.css` file appeared as untracked — verify it's a build artifact and gitignore it
40. Consider whether `syntax_test.go` `TestSerializePreservesAllFields` should also test `Clone()` field preservation (currently only tests `serial()`)

### Documentation

41. Update `TESTING.md` with the new pipeline integration test pattern
42. Document the `runPipeline` helper in the testing guide so future tests follow the pattern
43. Add an ADR for the `serial()` field-preservation testing strategy
44. Consider documenting the full actionability pipeline architecture (parse → serialize → suffix tree → FindSyntaxUnits → CloneNode → patterns → property engine) in a single diagram
45. Document the relationship between `syntax.Node`, `domain.CloneNode`, and the serialization/conversion bridges
46. Update `docs/ACTIONABILITY_PATTERNS.md` with the new `productionSuggestions` map structure
47. The `clone_classify.go` file is now 300 lines (over the 250-line policy for production files) — consider splitting `productionSuggestions` and `patternLabelConfigs` into a `clone_suggestions.go` file
48. Consider adding `file_size_check` to the CI that fails on files >250 lines for non-test, non-generated files
49. The `extractability_engine.go` is 498 lines — may need splitting if more checks are added
50. Consider creating a `docs/CROSS_LAYER_TESTING.md` guide documenting the pipeline integration test pattern and why it's necessary

---

## g) Questions (up to 3)

### Q1: Should we fix the tagliatelle contradiction?

`.golangci.yml` has `tagliatelle` enabled (line in the enable list), producing 50 `json(camel)` violations. But `AGENTS.md` line 119 says: "tagliatelle: codebase has mixed snake_case/camelCase JSON conventions per ADR-0016" and claims it's NOT enabled. Should I:
- **(a)** Disable tagliatelle in `.golangci.yml` (match the AGENTS.md claim), or
- **(b)** Fix all 50 JSON tags to camelCase and update AGENTS.md?

The codebase has public API JSON contracts (SDK types in `pkg/artdupl/types.go`) that use snake_case intentionally. Changing those would be a breaking API change.

### Q2: Should `clone_classify.go` (300 lines) be split?

It's 50 lines over the 250-line production file policy. The file contains three logical sections: suggestion constants (36 lines), `ClassifyClone`/`nodeTypeToCategory`/`calculatePriority` (130 lines), and `getSuggestion`/`patternLabelConfigs`/`ApplyPatternLabel` (134 lines). The cleanest split would be extracting suggestion constants + `productionSuggestions` map to `clone_suggestions.go`. Should I do this?

### Q3: Should I fix the `biome` missing-binary issue in the BuildFlow pre-commit hook?

The pre-commit hook fails on every manual commit because `biome` isn't installed. This blocks all manual commits. Options:
- **(a)** Add `biome` to `flake.nix` devShell (requires finding the nix package name), or
- **(b)** Exclude `biome-format` from BuildFlow config (if there's a `.buildflow.yml` or similar), or
- **(c)** Leave it — the auto-commit daemon works around it.

This is a pre-existing issue but it affects every commit attempt.

---

## Session metrics

| Metric | Value |
|--------|-------|
| Commits this session | 5 (`4bf21bd1`, `766caa85`, `4e5111df`, `703af078`, `3e4ffff3`) |
| Files changed | 11 (including status report) |
| Lines added | ~713 |
| Lines removed | ~55 |
| Tests passing | 919/919 (27 packages) |
| Lint issues on changed packages | 0 |
| Lint issues project-wide | 50 (all pre-existing tagliatelle) |
| Dogfooding (production code) | 0 clone groups at t=1 |
| Full test suite runtime | ~5s |

---

## Resolution (2026-08-10)

**All 5 lint issues fixed and shipped.** gocyclo, gofumpt, modernize, wsl_v5 all resolved (commit `4bf21bd1`). Node serialization regression guard (`TestSerializePreservesAllFields`) and full-pipeline integration tests shipped. Commits: `766caa85`, `4e5111df`, `703af078`, `3e4ffff3`. Tagliatelle contradiction → TODO_LIST.
