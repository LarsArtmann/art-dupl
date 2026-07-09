# Status Report — 2026-07-09 15:10

## Session Goal

Fix false-positive clone detection noise in semantic mode, as reported in `/home/lars/forks/upd/art-dupl.html` — `errors.New("different strings")` were reported as duplicates.

---

## A) FULLY DONE ✅

| #   | Task                                                                                                                                                                                                                                                                                                                 | Files                                                                                                     | Verified                                                                              |
| --- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------- |
| 1   | **Root cause analysis** — Identified 3 root causes: (a) default threshold=1, (b) ValueSpec not marked as Statement → partial prefix matching, (c) no boilerplate filters for assign+err-check or single-CallExpr                                                                                                     | `docs/feedback/2026-07-09-semantic-noise-declaration-files.md`                                            | ✅                                                                                    |
| 2   | **Default threshold 1→5** — `config.DefaultThreshold` raised from 1 to 5. Updated help text, CLI flag description                                                                                                                                                                                                    | `config/config.go:170`, `cmd/flags.go:16`                                                                 | ✅ Tests passed (before cache invalidation)                                           |
| 3   | **ValueSpec as Statement** — GenDecl spec children now marked `Statement=true`, causing each `var`/`const`/`type` declaration to be fingerprinted as a single composite token. Different string literal values produce different fingerprints → no more partial expression prefix matching in declaration-only files | `syntax/golang/transform.go:187-189`                                                                      | ✅ Verified against `/home/lars/forks/upd/` — Group #1 eliminated even at threshold 1 |
| 4   | **Assign+error-check pattern** — New actionability pattern detecting 2-stmt `err := f(); if err != nil { return ... }` boilerplate. Classified as NonActionable                                                                                                                                                      | `printer/actionability.go:88-90`, `printer/actionability_boilerplate.go` (new)                            | ✅ Unit tests pass                                                                    |
| 5   | **Single-CallExpr pattern** — New actionability pattern detecting lone CallExpr nodes like `errors.New("foo")`. Classified as NonActionable                                                                                                                                                                          | Same file                                                                                                 | ✅ Unit tests pass                                                                    |
| 6   | **Tests updated** — Fixed 4 test files that asserted threshold=1 defaults. Added explicit `-t 1` to stats integration tests that need clones                                                                                                                                                                         | `config/config_test.go`, `config/config_enum_test.go`, `cmd/cmd_test.go`, `cmd/stats_integration_test.go` | ✅                                                                                    |
| 7   | **New unit tests** — 11 test cases for the two new actionability patterns                                                                                                                                                                                                                                            | `printer/actionability_boilerplate_test.go` (new)                                                         | ✅                                                                                    |
| 8   | **AGENTS.md updated** — Added statement-level tokenization convention, default threshold note, actionability patterns list                                                                                                                                                                                           | `AGENTS.md`                                                                                               | ✅                                                                                    |
| 9   | **End-to-end verification** — Ran art-dupl against `/home/lars/forks/upd/`: 0 clone groups at threshold 2 AND default threshold 5. Real duplicates still detected at threshold 3 in our own `printer/` package                                                                                                       | —                                                                                                         | ✅ (before cache invalidation)                                                        |

---

## B) PARTIALLY DONE ⚠️

| #   | Task                                                                                                                                                                                               | Status                                                         | What's Missing                     |
| --- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------- | ---------------------------------- |
| 1   | **Feedback document** — Written but contains recommendations that are now implemented. Should be updated to reflect "FIXED" status                                                                 | `docs/feedback/2026-07-09-semantic-noise-declaration-files.md` | Update status to "resolved"        |
| 2   | **Full test suite verification** — All 24 packages passed earlier in session, but build cache has since been invalidated. Cannot re-verify due to pre-existing json/v2 build break (see section D) | —                                                              | Need json/v2 migration fixed first |

---

## C) NOT STARTED ⏭️

These were identified in the feedback doc as future improvements but not implemented:

1. **Enforce minimum threshold floor of 3** — Currently still `< 1` in `config_validate.go:41`
2. **Fix partial match fragment rendering** — `buildMatch` in `syntax_match.go:37-41` uses `Owns` to extend fragment to full expression even when only a prefix matched
3. **Add `--min-lines` flag** — Complementary line-count filter
4. **Update HOW_TO_USE.md** — Default threshold examples may reference old value of 1

---

## D) TOTALLY FUCKED UP 💥

### Pre-existing: `encoding/json/v2` Migration is BROKEN

**NOT caused by this session.** These changes were already in the working tree at conversation start (`git status` showed them as modified).

**Scope:** 10+ files were mid-migration from `encoding/json` to `encoding/json/v2`:

| File                         | Issue                                                                                                                                                                                              |
| ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `errors/marshal.go`          | Imports `encoding/json/v2` but references `json.UnsupportedValueError` and `json.UnsupportedTypeError` — **these types don't exist in json/v2**. Compilation fails even with `GOEXPERIMENT=jsonv2` |
| `config/config_migrate.go`   | Imports `encoding/json/v2` + `encoding/json/jsontext`. Requires `GOEXPERIMENT=jsonv2` to compile                                                                                                   |
| `config/enum_helpers.go`     | Same json/v2 import                                                                                                                                                                                |
| `pkg/enum/enum.go`           | Same                                                                                                                                                                                               |
| `pkg/artdupl/types.go`       | Same                                                                                                                                                                                               |
| `baseline/baseline.go`       | Same + uses `jsontext.WithIndentPrefix`/`WithIndent`                                                                                                                                               |
| `printer/json.go`            | Same                                                                                                                                                                                               |
| `printer/sarif.go`           | Same                                                                                                                                                                                               |
| `printer/stats_formatter.go` | Same                                                                                                                                                                                               |

**Impact:** The ENTIRE codebase cannot compile without `GOEXPERIMENT=jsonv2`, and even with it, `errors/marshal.go` fails due to undefined types.

**Build cache note:** Tests passed earlier in this session because the Go build cache still had valid entries from before the json/v2 migration. Once the cache was invalidated, the build broke.

**Flake.nix gap:** `GOEXPERIMENT=jsonv2` is NOT set in `flake.nix` devShell env. It must be exported manually or the migration will never work in CI.

### Other Pre-existing Changes (NOT mine)

| File                               | Nature of Change                                               |
| ---------------------------------- | -------------------------------------------------------------- |
| `.golangci.yml`                    | Massively reduced (125 lines diff) — many linter rules removed |
| `go.mod` / `go.sum`                | Dependencies updated (22 lines changed in go.mod)              |
| `flake.lock`                       | Updated                                                        |
| `printer/clone_classify.go`        | Inlined `isTestFile()` into call site, removed the function    |
| `internal/testutil/assert.go`      | Minor change                                                   |
| `internal/testutil/bdd_runners.go` | Minor change                                                   |

---

## E) WHAT WE SHOULD IMPROVE 🔧

### Code Quality

1. **`errors/marshal.go` json/v2 types** — `UnsupportedValueError`/`UnsupportedTypeError` don't exist in json/v2. Either keep `encoding/json` (v1) import or find the v2 equivalents (likely `*jsonv2.SemanticError` or similar)
2. **`flake.nix` missing `GOEXPERIMENT=jsonv2`** — If the migration is intentional, the env var MUST be set in the devShell
3. **`GOEXPERIMENT` is experimental** — Relying on experimental Go features in production code is risky. Consider whether this migration is worth it, or wait for Go 1.27 where json/v2 may be stabilized

### Architecture

4. **Partial match rendering is misleading** — When suffix tree matches a partial expression prefix (e.g., 4 of 5 tokens of `errors.New("foo")`), `buildMatch` extends the fragment to cover the FULL expression via `Owns`. The user sees two different expressions reported as "the same clone." Fix: only extend to `Owns` boundary when match covers the full subtree
5. **`fileContainsStatements` is O(n) per call** — Called inside `FindSyntaxUnits` which is called per match. For large codebases, this linear scan of ALL data nodes is wasteful. Could pre-compute a `map[string]bool` once
6. **Actionability pattern ordering matters** — `isAssignWithErrorCheck` is checked AFTER `isPureErrorPropagation`. A 2-stmt `err := ...; if err != nil` would be caught by error-propagation first IF the IfStmt is the first statement. The ordering should be documented

### Testing

7. **No integration test for ValueSpec-as-Statement** — The fix is verified manually but there's no regression test that asserts "declaration-only files produce zero clones at threshold 1"
8. **No integration test for new actionability patterns** — The unit tests check the pattern functions directly, but no end-to-end test verifies that a real Go file with `err := f(); if err != nil { return nil }` is classified as NonActionable
9. **Stats integration tests are fragile** — They depend on real clones existing in `./printer/` at a specific threshold. Should use fixture files instead

### Documentation

10. **HOW_TO_USE.md still says "default: 1"** — Needs updating to reflect new default of 5
11. **Feedback doc not marked as resolved** — Should update with "IMPLEMENTED" status

---

## F) NEXT 50 THINGS TO DO 📋

### Critical (blocks everything)

| #   | Task                                                                                                                                             | Effort |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------ | ------ |
| 1   | **Fix `errors/marshal.go` json/v2 types** — Replace `json.UnsupportedValueError`/`json.UnsupportedTypeError` with v2 equivalents or revert to v1 | 15 min |
| 2   | **Add `GOEXPERIMENT=jsonv2` to `flake.nix` devShell** — Or decide to abandon json/v2 migration                                                   | 5 min  |
| 3   | **Verify full test suite passes after json/v2 fix**                                                                                              | 5 min  |

### High Impact

| #   | Task                                                                                                  | Effort |
| --- | ----------------------------------------------------------------------------------------------------- | ------ |
| 4   | **Enforce minimum threshold floor of 3** in `config_validate.go`                                      | 5 min  |
| 5   | **Fix `buildMatch` partial-match rendering** — Only extend to `Owns` when match covers full subtree   | 30 min |
| 6   | **Add regression test** — Declaration-only file (`var ErrFoo = errors.New(...)`) produces zero clones | 10 min |
| 7   | **Add regression test** — `err := f(); if err != nil { return }` classified as NonActionable          | 10 min |
| 8   | **Update HOW_TO_USE.md** — Default threshold examples                                                 | 10 min |
| 9   | **Update feedback doc** — Mark as resolved with implementation details                                | 5 min  |
| 10  | **Run `golangci-lint`** — The `.golangci.yml` was changed (pre-existing); verify linting passes       | 10 min |

### Medium Impact

| #   | Task                                                                                                                                                         | Effort |
| --- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------ |
| 11  | **Add `--min-lines` flag** — Complementary line-count threshold filter                                                                                       | 45 min |
| 12  | **Pre-compute `fileContainsStatements` as a map** — Eliminate O(n) scan per match                                                                            | 20 min |
| 13  | **Document actionability pattern precedence** — Add comment explaining why order matters                                                                     | 10 min |
| 14  | **Add BDD test for ValueSpec fingerprinting** — Verify `var` blocks with different values don't match                                                        | 20 min |
| 15  | **Review `.golangci.yml` reduction** — 125 lines of linter rules removed; verify nothing important was lost                                                  | 30 min |
| 16  | **Review `go.mod` dependency changes** — 22 lines changed; verify no unwanted upgrades                                                                       | 15 min |
| 17  | **Review `printer/clone_classify.go` inlining** — `isTestFile()` was removed and inlined; verify no callers remain                                           | 10 min |
| 18  | **Add integration test with fixture files** for stats — Replace fragile `./printer` dependency                                                               | 30 min |
| 19  | **Consider TypeSpec `Statement=true` implications** — We mark ALL GenDecl children as Statement; verify type aliases and generic types fingerprint correctly | 20 min |
| 20  | **Test import-only GenDecl** — `import` GenDecls are already skipped in `transform.go:134`, but verify the new Statement marking doesn't affect them         | 10 min |

### Lower Priority

| #   | Task                                                                                                                         | Effort   |
| --- | ---------------------------------------------------------------------------------------------------------------------------- | -------- |
| 21  | **Run `nix flake check`** — Verify reproducible build                                                                        | 10 min   |
| 22  | **Run `nix build`** — Verify Nix packaging works with new changes                                                            | 5 min    |
| 23  | **Commit all changes** — Create a proper commit with the threshold + ValueSpec + actionability fixes                         | 10 min   |
| 24  | **Consider `--threshold` minimum in CLI** — Currently accepts 1; could warn below 3                                          | 15 min   |
| 25  | **Add `--aggressive` / `--sensitive` preset flags** — Convenience presets for threshold + mode                               | 30 min   |
| 26  | **Audit all actionability patterns for ordering** — Some patterns may shadow others                                          | 20 min   |
| 27  | **Add `PatternAssignErrorCheck` to suggestion text** — New patterns should have user-facing explanations                     | 10 min   |
| 28  | **Add `PatternSingleCallExpr` to suggestion text** — Same                                                                    | 10 min   |
| 29  | **Consider normalizing import paths** — Multiple import styles might cause false negatives                                   | 30 min   |
| 30  | **Test templ files with ValueSpec equivalent** — Verify templ files aren't affected by GenDecl change                        | 15 min   |
| 31  | **Profile performance impact** of ValueSpec fingerprinting — More fingerprinting = more CPU per node                         | 15 min   |
| 32  | **Consider hash collisions** in fingerprintSubtree — int32 FNV hash; with more fingerprints, collision probability increases | 20 min   |
| 33  | **Add SARIF rule IDs for new patterns** — AssignErrorCheck, SingleCallExpr should have SARIF metadata                        | 15 min   |
| 34  | **Update `domain/analysis_errors.go`** — `ErrInvalidThreshold` message says ">= 1"; should say ">= 3" if floor changes       | 5 min    |
| 35  | **Consider `--no-boilerplate-filter` flag** — Power users may want to see ALL matches including boilerplate                  | 20 min   |
| 36  | **Add `--explain` flag** — Show WHY a group is classified as actionable/non-actionable                                       | 45 min   |
| 37  | **Review all `isTestFile` usages** — The function was inlined in `clone_classify.go`; check if it exists elsewhere           | 10 min   |
| 38  | **Consider semantic mode for templ** — Currently structural-only; could add identifier hashing                               | 2 hours+ |
| 39  | **Add clone group deduplication** — Same clone may appear in multiple groups via different match paths                       | 30 min   |
| 40  | **Document the suffix tree partial-match behavior** — Users should understand that partial expression prefixes can match     | 15 min   |
| 41  | **Consider `--max-occurrences` flag** — Filter groups with too many occurrences (often generated code)                       | 20 min   |
| 42  | **Add `--category` filter flag** — Show only specific categories (function, method, etc.)                                    | 30 min   |
| 43  | **Test with real-world repos** — Run against Kubernetes, Go stdlib, etc. to validate noise reduction                         | 1 hour   |
| 44  | **Benchmark detection speed** — Verify threshold 5 is faster than threshold 1 (fewer comparisons)                            | 15 min   |
| 45  | **Consider `--threshold-mode` flag** — `statements` (current) vs `tokens` (legacy) vs `lines`                                | 1 hour+  |
| 46  | **Add `--diff-threshold` flag** — Only report clones whose diff is below N% of total size                                    | 45 min   |
| 47  | **Review `isCyclic` filter** — May produce false negatives for legitimately repeated patterns                                | 20 min   |
| 48  | **Add `--ignore-pattern` for actionability** — Let users mark specific patterns as non-actionable                            | 30 min   |
| 49  | **Consider machine-learning-based actionability** — Train on labeled clone groups                                            | 4 hours+ |
| 50  | **Write ADR for threshold change** — Document the decision to raise default from 1 to 5                                      | 20 min   |

---

## G) TOP 2 QUESTIONS ❓

### 1. Should I fix the broken `encoding/json/v2` migration?

The migration spans 10+ files and was NOT authored by me. `errors/marshal.go` references types (`json.UnsupportedValueError`, `json.UnsupportedTypeError`) that don't exist in `encoding/json/v2`. `GOEXPERIMENT=jsonv2` is not set in `flake.nix`. The entire codebase cannot compile.

**I cannot determine:** Was this migration intentional and in-progress (should I complete it)? Or was it an accidental mass-replace (should I revert it)? The AGENTS.md says "Check `flake.nix` first" but the flake doesn't configure json/v2.

### 2. Should the minimum threshold floor be enforced via validation error or warning?

Raising the default from 1→5 is done, but users can still pass `--threshold 1`. Options:

- **Hard floor (error):** Reject `< 3` in `validateThreshold` — prevents noise but breaks scripts/tests that use `-t 1`
- **Soft floor (warning):** Print a warning when `< 3` — non-breaking, but users may ignore it
- **No floor:** Trust the user — but then noise complaints will recur

The BDD tests pass `threshold: "1"` explicitly, so a hard floor would break them. I cannot determine your preference here.
