# Status: `--suggest-generics` Feature Implementation

**Date:** 2026-08-10 02:39
**Session:** Implementing generics-extraction candidate detection to close the `--type-aware` false-negative gap identified by DiscordSync feedback.

---

## Context

The feedback at `docs/feedback/new/2026-08-10_discordsync_type-aware-false-negatives-generics-extraction-candidates.md` identified that `--type-aware` systematically hides 3 of 31 real clone groups (10% false-negative rate). The root cause: `--type-aware` bakes the type string into the identifier hash at the token level, so "same algorithm, different types" clones never form as groups in the suffix tree. The feedback proposed an opt-in `--suggest-generics` flag.

---

## a) FULLY DONE

### Core mechanism (type-erased hashing)
- **`PreloadedAST.EraseHash`** field added (`syntax/golang/typeinfo.go`). When true, the transformer populates `Node.VarType` but does NOT encode the type into the identifier hash.
- **`LoadTypeAwareData(files, eraseHash)`** signature updated. The `eraseHash` flag flows to each `PreloadedAST`.
- **`transformer.typeEraseHash`** field wired in `parsePreloaded()` (`syntax/golang/parse.go`).
- **`encodeTypeIfAware()`** extracted method on transformer (`syntax/golang/transform.go`). Separates VarType population from hash encoding. Replaces the inline nested-if block (also fixes a nestif lint warning).

### Pipeline wiring
- **CLI pipeline**: `loadTypeAwareData` takes `eraseHash bool` param (`cmd/type_aware.go`). Both `buildSuffixTreeStandard` and `buildSuffixTreeIncremental` check `cfg.TypeAware || cfg.SuggestGenerics` and pass `cfg.SuggestGenerics` as `eraseHash` (`cmd/run_analysis.go`).
- **SDK pipeline**: `loadTypeAwareDataIfEnabled` checks `cfg.TypeAware || cfg.SuggestGenerics` and passes `cfg.SuggestGenerics` as `eraseHash` (`pkg/artdupl/detector_pipeline.go`).

### Generics classification
- **`printer/generics_candidate.go`** (NEW): `ClassifyGenericsCandidate()` walks clone instance node trees in pre-order parallel, comparing `VarType` at corresponding positions. Any divergence → `GenericsCandidate=true` + human-readable `GenericsHint`.
- **`domain.CloneClassification`** extended with `GenericsCandidate bool` and `GenericsHint string` fields (`domain/processed_clone.go`).
- **`ProcessClones`** calls `ClassifyGenericsCandidate` and sets the fields on every clone in the group (`printer/clone_processor.go`).

### Output filtering + display
- **`SuppressionConfig.SuggestGenerics`** field added (`cmd/run_output.go`). When true, `printCloneGroups` filters to ONLY generics-extraction candidates.
- **Text printer** shows a `generics:` hint line with the type differences (`printer/text.go`).
- **JSON printer** includes `generics_candidate` and `generics_hint` fields (`printer/json.go`).

### Config + CLI flag
- **`config.Config.SuggestGenerics`** field added (`config/config.go`).
- **`--suggest-generics`** flag registered as root-only (`cmd/flags.go`).
- **Flag mapping** in `config_builder.go` (`"suggest-generics": &cfg.SuggestGenerics`).

### SDK support
- **`Options.SuggestGenerics`** field (`pkg/artdupl/types.go`).
- **`detectorConfig.SuggestGenerics`** + `convertOptionsToConfig` wiring (`pkg/artdupl/detector_utils.go`).

### Tests (13 new, all passing)
- `syntax/golang/generics_erase_test.go` (3 tests): EraseHash produces same hashes for different types; VarType still populated; type-aware vs erase-hash behavioral comparison.
- `printer/generics_candidate_test.go` (7 tests): Identical types → not candidate; different types → candidate; single sequence → not candidate; no VarType → not candidate; empty VarType skipped; children traversed; hint format correct.
- `printer/generics_integration_test.go` (3 tests): ProcessClones sets GenericsCandidate when types differ; doesn't when types match; doesn't when no VarType.

### Documentation
- `AGENTS.md` updated with `--suggest-generics` convention entry and `SuppressionConfig` field list updated.

### Test results
- **All 30 test suites pass** (27 `ok` + 3 `[no test files]`), 0 failures.
- **Build clean**: `go build ./...` succeeds.
- **gofmt clean**: all modified files formatted.
- **Lint**: no new issues introduced (only pre-existing tagliatelle/gopls warnings remain).

### Stats
- 19 modified files, 4 new files (138 insertions, 43 deletions in `.go` files).
- 3 new test files with 13 tests.

---

## b) PARTIALLY DONE

### End-to-end integration test on real Go code
The unit tests verify the mechanism in isolation (transformer EraseHash, classification logic, ProcessClones wiring). However, there is **no full-pipeline integration test** that:
1. Writes 2 Go files with structurally-identical functions operating on different types.
2. Runs the full pipeline: parse with EraseHash → suffix tree → FindSyntaxUnits → classify.
3. Asserts the clone group is detected and flagged as GenericsCandidate.

The `runPipeline` helper in `printer/actionability/pipeline_integration_test.go` exists and could be adapted, but it doesn't support type-aware data loading (it uses `golang.Parse`, not `parsePreloaded`). Extending it would require plumbing `LoadTypeAwareData` into the test fixture.

### SDK test for SuggestGenerics
The SDK field is wired but there is **no SDK unit test** that exercises `Options.SuggestGenerics: true` through the detector. The existing `pkg/artdupl/detector_type_aware_test.go` covers `TypeAware: true` but not the new flag.

---

## c) NOT STARTED

### FEATURES.md / HOW_TO_USE.md / TODO_LIST.md updates
The feature is not documented in user-facing docs yet:
- `HOW_TO_USE.md` needs a `--suggest-generics` section with usage examples.
- `FEATURES.md` needs the feature added under the detection modes section.
- `TODO_LIST.md` should have the feedback item marked as done.

### ADR
No Architecture Decision Record created for the EraseHash design decision. The choice to use `EraseHash` on `PreloadedAST` (rather than a separate `DetectionMode` or a transformer config flag) is an architectural decision that should be recorded.

### `--explain` integration
The text output shows a `generics:` hint line unconditionally for generics candidates, but `--explain` doesn't mention the generics classification in its structured explanation output. The `writeExplanation` function in `printer/text.go` could add a `generics-extraction candidate` note.

### SARIF output
SARIF output (`printer/sarif.go`) does not include `generics_candidate` or `generics_hint` fields. SARIF consumers (GitHub Security tab) would not see the generics classification.

### Config validation
There is **no explicit validation** preventing `--type-aware --suggest-generics` simultaneously. The code handles it (suggest-generics takes precedence because it's passed as `eraseHash=true` which wins when both are checked), but the user gets no warning. `config.ValidateConfig` should either error or warn.

### Incremental cache interaction
The `--suggest-generics` flag works with `--incremental` (the type-aware data is loaded and passed to the incremental parser), but the **cache key does not distinguish between `type-aware` and `suggest-generics` modes**. Running `art-dupl --suggest-generics --incremental` after `art-dupl --type-aware --incremental` would use cached ASTs that have types encoded in the hash, not erased. This is a latent correctness bug.

---

## d) TOTALLY FUCKED UP

### Nothing is totally fucked up.
The implementation is sound, all tests pass, and the build is clean. The gaps are omissions, not errors.

---

## e) WHAT WE SHOULD IMPROVE

1. **The `contains` helper functions in `generics_candidate_test.go` were originally hand-written** instead of using `strings.Contains`. Fixed during the session, but this shows I should default to stdlib functions immediately.

2. **The `TestEraseHash_ComparedWithTypeAware` test has a stale LSP diagnostic** (`typecheck: expected ';', found TA` at line 145) that was from a typo I introduced and fixed. The LSP cache may not have refreshed. The file compiles and tests pass. This is cosmetic.

3. **The `encodeTypeIfAware` method was extracted mid-session** to fix a nestif lint warning. The original implementation nested 4 levels deep inside the Ident case. I should have structured it as a separate method from the start, anticipating the complexity.

4. **I didn't run the BDD test suite with a `--suggest-generics` scenario.** The BDD tests in `bdd/` exercise the CLI end-to-end and would be the most convincing proof that the feature works on real Go code. I should have added a Ginkgo `Describe` block for suggest-generics.

5. **The feedback doc itself was not annotated** to show it was addressed. The convention in this project is to annotate feedback docs when work is done.

---

## f) Next Steps (up to 50)

### Critical (correctness)
1. **Fix incremental cache key** to distinguish `type-aware` from `suggest-generics` mode. Cache entries built with `eraseHash=true` are incompatible with `eraseHash=false`.
2. **Add config validation** for `--type-aware --suggest-generics` conflict (warn or error).
3. **Write full-pipeline integration test** that exercises parse → suffix tree → classification with EraseHash on real Go source files containing generics-extraction candidates.

### High value (completeness)
4. **Update `HOW_TO_USE.md`** with `--suggest-generics` section, usage examples, and output format.
5. **Update `FEATURES.md`** to list `--suggest-generics` as DONE under detection modes.
6. **Update `TODO_LIST.md`** to mark the feedback item as done.
7. **Add `--explain` integration**: include `generics-extraction candidate` in the structured explanation output.
8. **Add SARIF output support**: include `generics_candidate` and `generics_hint` in SARIF properties.
9. **Add SDK unit test** for `Options.SuggestGenerics: true`.
10. **Write an ADR** (e.g., ADR-0020) documenting the EraseHash design decision.

### BDD / E2E tests
11. **Add BDD scenario**: "suggest-generics detects same-algorithm-different-type clones".
12. **Add BDD scenario**: "suggest-generics does not report same-type clones as generics candidates".
13. **Add BDD scenario**: "suggest-generics + --explain shows the generics hint".
14. **Add BDD scenario**: "suggest-generics + --json includes generics_candidate field".
15. **Add test fixture files** that replicate the DiscordSync feedback examples (maxAuthorKindCount, totalAuthorKindCount, kind-derivation logic twin).

### Classification refinement
16. **Add minimum clone size filter** for generics candidates — 1-2 statement clones are too noisy (e.g., single `return x.Field` accessors on different types). Consider requiring >= 3 statements.
17. **Consider grouping related generics candidates** — Finding 1 (max loops) and Finding 2 (sum loops) from the feedback are the same algorithm family. A "generics family" grouping could reduce output noise.
18. **Explore field-access normalization** as suggested in the feedback — normalizing `x.Field` to `SELECTOR(int64)` so that even non-type-aware mode can detect structural equivalence. This was the feedback's primary proposal but is more complex.
19. **Add "duck-typed field-access equivalence" check** for Finding 3 (cross-package logic twin) — when accessed fields have the same name and same type across different container types, the clone is a generics candidate even without full type-aware data.

### Performance
20. **Profile suggest-generics on a large codebase** — the type-checking cost is the same as `--type-aware` (10-100x), but the classification pass (`ClassifyGenericsCandidate`) adds a tree-walk per clone group. Measure the overhead.
21. **Optimize `flattenCloneNodes`** — the current implementation recursively flattens the entire subtree. For large clones, this could be expensive. Consider an iterative approach or early termination on first divergence.

### Output polish
22. **Add color/rich-text support** for the generics hint line in text output.
23. **Add `--suggest-generics --no-actionability` combination** test — this should show ALL generics candidates including boilerplate.
24. **Add generics candidate count to summary statistics** (e.g., "Found 5 clone groups (3 generics-extraction candidates)").
25. **Consider a `--suggest-generics-only` shorthand** that implies `--no-actionability` since generics candidates are the user's explicit goal.
26. **Add HTML output support** — highlight generics-extraction candidates with a distinct badge/icon in HTML output.

### Architecture / cleanup
27. **Extract type-aware data loading** to a shared helper** — the feedback doc from 2026-07-24 noted that `loadTypeAwareData` (CLI) and `loadTypeAwareDataIfEnabled` (SDK) duplicate the same drain/filter/load/error-handle logic. This session added more divergence (eraseHash). Extract to a shared function.
28. **Consider making `EraseHash` a `DetectionMode`** rather than a boolean flag on `PreloadedAST`. A `DetectionModeGenericsCandidate` mode would be more discoverable and consistent with the existing exact/semantic/structural taxonomy. However, it would require threading through more code.
29. **Document the VarType field lifecycle** — VarType is set in the transformer, copied through `Clone()`, `serial()`, `syntaxToCloneNode`, and consumed by the generics classifier. This chain is fragile (as the `serial()` field-preservation hazard showed). Add a test that asserts VarType survives the full pipeline.
30. **Add `generics_candidate` to `simpleJSONClone`** (currently intentionally minimal, but the `generics_candidate` flag is small and high-signal).

### Docs / changelog
31. **Update `CHANGELOG.md`** with the `--suggest-generics` feature entry.
32. **Update `SDK_DESIGN.md`** to document `Options.SuggestGenerics`.
33. **Annotate the feedback doc** (`docs/feedback/new/2026-08-10_...`) to show it was addressed.
34. **Add a `docs/adr/0020-suggest-generics.md`** ADR.
35. **Move the feedback doc** from `docs/feedback/new/` to `docs/feedback/done/` (or annotate inline).

### Edge cases
36. **Test: suggest-generics with no .go files** (only .templ) — should gracefully produce no results, not crash.
37. **Test: suggest-generics on a single file** — no clones possible, should produce empty output cleanly.
38. **Test: suggest-generics + --sort total-tokens** — verify sorting works correctly with the filtered output.
39. **Test: suggest-generics with --threshold 1** — the DiscordSync use case used `-t 1` for maximum recall.
40. **Test: suggest-generics fallback** — when type checking fails, it should fall back to syntax-only and produce no generics candidates (since VarType would be empty).
41. **Test: suggest-generics + --show-suppressed** — should this show non-generics clones too? Currently it doesn't (the SuggestGenerics filter runs before ShowSuppressed). Clarify the interaction.
42. **Test: suggest-generics + accept directives** — accepted clones should still be suppressed even in suggest-generics mode.
43. **Test: suggest-generics + --min-lines** — min-lines filter should still apply to generics candidates.

### Validation hardening
44. **Validate: suggest-generics requires semantic mode** — the feature only works with alpha-normalization (the Ident case's `NormalizesLocals()` gate). If someone passes `--exact --suggest-generics`, it silently does nothing. Add a warning or error.
45. **Validate: suggest-generics is incompatible with --dump-tokens** — dump-tokens doesn't run detection.
46. **Validate: suggest-generics is incompatible with --hash method** — hash method doesn't use AST.

### Misc
47. **Consider adding generics candidate count to the `stats` subcommand** output.
48. **Consider a `--suggest-generics --count-only` mode** that just reports the number of generics candidates without full output.
49. **Add a `--generics-min-instances N` flag** — only show generics candidates with >= N instances (some families have 4-6 variants, others have just 2).
50. **Consider fuzzing `ClassifyGenericsCandidate`** with random CloneNode trees to ensure no panics on malformed input.

---

## g) Questions I cannot answer myself

1. **Should `--suggest-generics` be mutually exclusive with `--type-aware` at the config validation level, or should it silently take precedence?** The current implementation silently takes precedence (suggest-generics wins). A validation error would be more explicit but adds friction for users who alias flags in scripts. I need your preference.

2. **Should the incremental cache distinguish between type-aware and suggest-generics modes?** The cache key currently doesn't include the `eraseHash` flag, so cached ASTs from a `--type-aware` run would be reused by a `--suggest-generics` run (and vice versa), producing incorrect results. I can fix this by adding the mode to the cache key, but it doubles the cache size for users who switch between modes. Is this acceptable, or should I invalidate the cache on mode switch?

3. **Should generics candidates bypass actionability filtering by default?** Currently they don't — a guard-clause pattern with different types would be suppressed by actionability even in suggest-generics mode. The feedback's Finding 1 (max loops) would likely survive actionability (it's a full function body), but shorter clones might not. Should `--suggest-generics` imply `--no-actionability`, or should the user opt into `--suggest-generics --no-actionability` explicitly?

---

## Resolution (2026-08-10)

**Core feature shipped.** `--suggest-generics` with EraseHash mode, ClassifyGenericsCandidate, text/JSON output, CLI flag, and SDK support all in CHANGELOG `[Unreleased]` → Added. Section c items: cache key fix → done (`956c1125`); config validation → TODO_LIST; SARIF support → TODO_LIST; full-pipeline integration test → TODO_LIST; BDD test → TODO_LIST; ADR-0020 → TODO_LIST. Section f brainstorm items routed to TODO_LIST (bounded) and ROADMAP (long-term).
