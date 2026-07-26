# Status Report: Post-Fix Sprint — Quality Gate GREEN

**Date:** 2026-07-26 16:18
**Session Goal:** Fix all breakage from the previous Pareto sprint, add missing documentation and tests, verify quality gate is green.

---

## Executive Summary

The previous session committed 17 tasks (T01-T17) but left the quality gate RED with 3 critical bugs (build failure, daemon regression, error swallowing), zero documentation updates, and zero unit tests for new features. This session fixed all 3 bugs, wrote 7 new test files, updated 5 documentation files, and verified the full quality gate passes (`go test -count=1 ./...` + `nix flake check` — 12 checks, all green).

**The quality gate is now GREEN.** But there are still significant gaps (see sections c, e, f).

---

## a) FULLY DONE

### Bug Fixes (3 critical)

1. **`printer/actionability_switch_test.go` build failure** — `isLoggingMethod` was removed (converted to inline `slices.Contains`) but the test still called it directly. Fixed to `slices.Contains(loggingMethodNames, tc.input)`. The root cause was **Go test cache masking the failure** — `go test ./printer/...` returned `(cached)` even after the build broke. Lesson: ALWAYS use `-count=1`.

2. **`.golangci.yml` daemon regression** — `exhaustruct` (enable list + config block) and `tagliatelle` (enable list) re-added by the auto-commit daemon. Removed all 3 occurrences. Had to remove them **TWICE** during this session — the daemon re-added them between my first removal and the final `nix flake check`.

3. **`cmd/diff_report.go` error swallowing** — `collectCurrentGroups` silently swallowed `ProcessClones` errors with `continue` (introduced to satisfy funlen linter in the previous session). Changed function signature to return `([]domain.ProcessedCloneGroup, error)` and propagate errors properly.

### Additional Fixes Discovered During Work

4. **Unused `//nolint:musttag` directive** in `cmd/diff_report.go:121` — caught by `nix flake check` lint pass. Removed.

5. **Self-test duplication** — The `TestAllActionabilityPatterns_ContainsLabels` test used `make(map) + for-range-set-true` which duplicated `containsCallTo` in `printer/actionability_data.go`. The self-test Nix check caught this. Rewrote to use `slices.Contains`.

6. **`wsl_v5` lint violations** — 5 whitespace linter failures across 3 new test files. Fixed all by adding blank lines before `if`/`range` statements.

7. **`mustFuncDeclWithBody` name collision** — My new test helper collided with an existing helper in `actionability_test.go`. Renamed to `mustInterfaceMethodNode`.

8. **Wrong `CloneType` constants** — Used `domain.CloneTypeRenamed`/`domain.CloneTypeExact` (don't exist) in SARIF test. Fixed to `domain.CloneType2`/`domain.CloneType1`.

9. **Missing `slices` import** — After switching to `slices.Contains`, forgot to add the import. Added.

### Documentation Updates (5 files)

10. **CHANGELOG.md** — Added 18 new entries to `[Unreleased]`: `--diff-report`, `--disable-pattern`, `--list-patterns`, `--recommend-threshold`, `--html-out`, HTML deep-linking, YAML config, `interface-method` pattern (#20), templ expression normalization, SARIF metadata, `version` subcommand, `--quiet`/`--no-color`, self-test Nix check, SARIF validation check, GitHub lint-config-guard, performance guide, cross-package alias tests, BDD diff report tests. Added 5 Changed entries (generatorIncludes refactor, slices.Contains conversions, pattern table extraction, examples threshold fix). Added 3 Fixed entries (isLoggingMethod orphaned test, isTestingVarName note, collectCurrentGroups error swallow). Removed the contradictory "Removed stub flags" section that claimed `--diff-report` etc. were removed (they're now implemented).

11. **TODO_LIST.md** — Removed 10 completed items: YAML config, diff-report mode, HTML improvements, recommend-threshold, SARIF metadata, ExprStmt audit, templ normalization, interface-method suppression, threshold fix, slices.Contains conversions, CI self-test, generatorIncludes refactor. Restructured remaining items. Updated the sentinel audit item to note partial completion.

12. **HOW_TO_USE.md** — Added 5 new sections after "Accepting Clone Groups": Diff Reports, Threshold Recommendation, Controlling Actionability Patterns, YAML Configuration, HTML Report to File. All with concrete code examples.

13. **FEATURES.md** — Updated pattern count 18→20 in overview and statistics table. Added `interface-method` and `single-declaration` to the pattern list. Added 8 new feature rows: Diff Report, CI Self-Test, Lint Config Guard, Disable Specific Pattern, List Patterns, Threshold Recommendation, HTML to File, Version Subcommand, Quiet Mode, Color Control. Updated Pattern in JSON row to mention SARIF. Added new flags to CLI table. Added YAML config row. Added Diff Report + Threshold Recommendation + Pattern Control to Quick Reference.

14. **docs/ACTIONABILITY_PATTERNS.md** — Added `interface-method` pattern row at position #3. Updated count 19→20 in "How It Works" section. Updated priority order list (added #3 Interface method, renumbered all subsequent). Added `--list-patterns`/`--disable-pattern`/`--no-actionability` mention. Added design decision note about the static name list and ROADMAP tracking for the type-aware variant.

### Unit Tests (7 new test files)

15. **`cmd/recommend_threshold_test.go`** — `RecommendThreshold` tested at all 4 boundaries (0, 99→100, 999→1000, 4999→5000) plus extreme values. `isGoFile` tested with 6 cases.

16. **`printer/actionability_interface_method_test.go`** — `isInterfaceMethodBody` tested with 6 cases: matching name + small body (String, Read at limit), non-matching name (ProcessData), body too large (5 > 4), non-FuncDecl, empty seqs.

17. **`config/config_yaml_test.go`** — 4 tests: full YAML config loading, YAML extension detection (.yaml + .yml), malformed YAML negative test, YAML/JSON equivalence.

18. **`syntax/templ/normalize_test.go`** — 6 tests: single var normalization, two vars, var reuse, no-dot unchanged, reserved words not normalized (len/cap/make), nil symbols passthrough, consistent canonicalization across calls, `isReservedWord` lookup.

19. **`printer/sarif_pattern_test.go`** — 2 tests: SARIF result contains `non_actionable_pattern` when set, does NOT contain it when actionable (empty string).

20. **`printer/actionability_patterns_disabled_test.go`** — 4 tests: `AllActionabilityPatterns` returns 20 patterns, all 20 required labels present, `ListActionabilityPatterns` writes correct output, `EvaluateActionabilityWithDisabled` correctly re-enables suppressed patterns when their label is disabled.

### Verification

21. **Full test suite**: `GOEXPERIMENT=jsonv2 go test -count=1 ./...` — all 27 packages pass, zero failures.
22. **Full Nix flake check**: all 12 checks pass (test, lint, fmt, race, bench, self-test, sarif-validate, disabled-linters, treefmt, build, devShell, ci).

---

## b) PARTIALLY DONE

### Templ expression normalization

The `normalizeExprValue` function exists and is wired into `transform_components.go:67`, but:

- The LSP reports the functions as "unused" (stale diagnostic — they ARE used)
- The normalization only covers lowercase-initial identifiers before dots (`user.Name` → `v0.Name`)
- It does NOT cover standalone identifiers in function call arguments (`Component(user, group)` where only `user` and `group` differ)
- No integration test exists verifying that two templ files with renamed variables actually produce the same hash

### SARIF `non_actionable_pattern`

The field is added to the `Properties` map and unit-tested, but:

- No BDD/integration test verifies it end-to-end through the CLI (`art-dupl --sarif` on real code → SARIF file → parse → assert field exists)
- The field is only populated when `cl.Classification.NonActionablePattern != ""`, but there's no test verifying that the classification pipeline actually SETS this field for real boilerplate clones

### `--disable-pattern` flag

The flag works at the evaluation level (`EvaluateActionabilityWithDisabled`), but:

- No integration test verifies the CLI flag end-to-end (run `art-dupl --disable-pattern guard-clause` → verify previously-suppressed clones now appear)
- No validation of label names at startup — a typo like `--disable-pattern guard-clouse` silently does nothing

### `--recommend-threshold` flag

The heuristic function is tested, but:

- The CLI handler (`runRecommendThreshold`) is not tested (it walks the filesystem)
- The output format is not tested

---

## c) NOT STARTED

### From the previous session's status report

1. **Local pre-commit git hook** — The daemon re-added forbidden linters TWICE during this session alone. The GitHub workflow only runs on push/PR. A `.git/hooks/pre-commit` or `.pre-commit-config.yaml` would stop the regression locally. Not implemented.

2. **`--disable-pattern` validation at startup** — Question 1 from the previous report asked whether to fail-fast on typos or silently ignore. I made no decision and the flag silently accepts unknown labels.

3. **SARIF validation tool choice** — Question 2 asked about external tool vs Go library for T14. Not addressed.

### Missing integration tests

4. **`--list-patterns` CLI test** — No test that actually runs the CLI with `--list-patterns` and verifies stdout contains all 20 labels.
5. **`--diff-report` CLI test** — The BDD tests exist (`bdd/diff_report_test.go`) but I didn't verify they actually pass against the fixed `collectCurrentGroups` signature change. (They did pass in `go test ./...` but I didn't examine them individually.)
6. **`--html-out` CLI test** — No test that verifies the file is written and auto-opened.
7. **`--html-out` deep-link test** — No test verifying `id="group-<hash>"` appears in generated HTML.

### Missing edge case tests

8. **YAML config with enum fields** — The YAML→JSON bridge might have issues with custom `UnmarshalJSON` hooks on enum types (`DetectionMode`, `OutputFormat`, `SortCriteria`). Not tested with enum values.
9. **`normalizeExprValue` with multiple dots** — `user.Profile.Name` — does the regex handle chained access? Not tested.
10. **`RecommendThreshold` with negative file count** — `RecommendThreshold(-1)` — not tested (though unlikely in practice).

---

## d) TOTALLY FUCKED UP

### Nothing in this session

The previous session had 3 critical bugs. This session fixed all 3. However:

1. **The daemon is an active enemy that I cannot defeat** — It re-added `exhaustruct`/`tagliatelle` to `.golangci.yml` between my edit and the `nix flake check` run. I removed them TWICE. There is no guarantee they won't be re-added after this session ends. The only durable fix is a local pre-commit hook, which I did not implement.

2. **I created a self-test failure** — My `TestAllActionabilityPatterns_ContainsLabels` used `make(map) + for-range` which duplicated production code (`containsCallTo`). The self-test Nix check caught it. I should have used `slices.Contains` from the start — the project convention is clear in AGENTS.md.

3. **I used wrong domain constant names** — `CloneTypeRenamed` and `CloneTypeExact` don't exist in the domain package. They're `CloneType2` and `CloneType1`. I should have checked the domain types before writing the test.

---

## e) WHAT WE SHOULD IMPROVE

### Process Improvements

1. **Always use `go test -count=1`** — The Go test cache masked a build failure in the previous session. This is now in AGENTS.md but should be a muscle memory rule.

2. **Run `nix flake check` after EVERY batch of changes** — Not just at the end. The daemon can regress between checks. I should have run it after the test files were written, before trying to commit.

3. **Check for name collisions before writing test helpers** — `mustFuncDeclWithBody` already existed. I should have grepped for the name before defining it.

4. **Read domain types before writing tests** — I guessed `CloneTypeRenamed` instead of checking the actual constant names. Wasted a build cycle.

5. **The documentation updates should have been part of the original sprint** — The previous session deferred ALL documentation to "later" and then forgot. Documentation should be part of each task's definition of done.

### Code Quality Improvements

6. **`collectCurrentGroups` still has a funlen risk** — I changed the signature to return error, which added 2 lines. If funlen becomes an issue again, the actionability check block should be extracted to a helper.

7. **The YAML config bridge is fragile** — YAML → JSON → struct relies on the JSON tags being YAML-compatible. If a field uses `json:"output_format"` but the YAML file has `outputFormat`, it won't work (YAML keys are case-sensitive). The test only uses lowercase keys matching the JSON tags.

8. **`normalizeExprValue` regex is narrow** — Only matches `[a-z][a-zA-Z0-9]*\.` — single identifier before dot. Doesn't handle chained access or nested expressions.

9. **No `--disable-pattern` startup validation** — Typos silently accepted. Users get no feedback that their flag did nothing.

10. **Pattern table has 20 patterns but only 19 PatternLabel constants are documented** — The `PatternNone` label (`""`) is the default/actionable state. It's in the constants but not in `AllActionabilityPatterns()` output. This is correct behavior but could confuse someone reading the code.

### Architecture Observations

11. **`printer/` is still ~29 source files / 3500+ lines** — This was flagged in TODO_LIST. The actionability pattern files alone are 8+ files. A `printer/actionability/` sub-package would help.

12. **The `actionabilityPatternTable` is good** — Centralizing patterns in one table was the right call. Adding a pattern is now a single line. But the check functions are spread across 8 files (`actionability.go`, `actionability_control_flow.go`, `actionability_patterns_expanded.go`, `actionability_data.go`, `actionability_interface_method.go`, etc.).

13. **SARIF Properties is `map[string]string`** — This works for simple metadata but can't carry structured data (e.g., nested objects for multi-rule metadata). SARIF spec allows richer properties.

---

## f) Up to 50 Things We Should Get Done Next

### HIGH Priority — Durability & Correctness

1. **Install local pre-commit hook** (`.git/hooks/pre-commit`) that runs `scripts/check-disabled-linters.sh` — stops daemon regression at commit time
2. **Add `--disable-pattern` startup validation** — fail-fast on unknown labels with a helpful error listing valid labels
3. **Write repo-wide `errors.New` sentinel audit** — scan for duplicated sentinel errors across packages (the `ErrInvalidDetectionMode` bug class)
4. **Integration test: `--list-patterns` CLI output** — run the binary, assert all 20 labels in stdout
5. **Integration test: `--diff-report` end-to-end** — create files, record baseline, add clone, run diff report, assert "new" section
6. **Integration test: `--disable-pattern` end-to-end** — run analysis, note suppressed group, add `--disable-pattern <label>`, verify it reappears
7. **Integration test: `--html-out` writes file** — verify file exists, contains `id="group-<hash>"`, is valid HTML
8. **Test YAML config with enum fields** — `detectionMode: semantic`, `outputFormat: json` — verify custom `UnmarshalJSON` hooks survive the YAML→JSON bridge
9. **Test `normalizeExprValue` with chained access** — `user.Profile.Name` edge case
10. **Verify BDD diff_report tests cover the fixed `collectCurrentGroups` error path** — add a test case where a file is unreadable

### MEDIUM Priority — Features & Polish

11. **Add `--recommend-threshold` CLI handler test** — test the filesystem walk and output format
12. **Expand `normalizeExprValue` to handle standalone identifiers** — `Component(user, group)` where only `user`/`group` differ
13. **Add SARIF `rule.tags` metadata** — emit actionability pattern as a SARIF rule tag (not just result property)
14. **Add SARIF rule for each clone type** — separate rules for type-1/2/3 so GitHub Security can filter
15. **Split `printer/` package** — extract `printer/actionability/` sub-package to reduce file count
16. **Add `--config-format` flag** — explicit format override (currently auto-detected by extension only)
17. **Add `--html-out` browser-open test** — verify `openHTMLOutput` calls the right platform command
18. **Add `RecommendThreshold` negative input test** — `RecommendThreshold(-1)` should probably return the small-codebase threshold
19. **Consolidate actionability check functions into fewer files** — 8 files for 20 patterns is too many
20. **Add `version --json` test** — verify structured version output
21. **Add `--quiet` suppression test** — verify `printBuildingStatus` is skipped when `cfg.Quiet` is true
22. **Add `--no-color` test** — verify `NO_COLOR=1` env var is set
23. **Document the YAML config key naming convention** — clarify that keys must match JSON tags (snake_case)
24. **Add a YAML config example to the `examples/` directory**
25. **Add `--diff-report --json` output structure test** — verify JSON diff report schema

### MEDIUM Priority — Code Health

26. **Run `golangci-lint run --fix` on the new test files** — auto-fix any remaining style issues
27. **Add `//art-dupl:accept` directives to any new test boilerplate** that the self-test flags
28. **Audit the `bdd/diff_report_test.go` tests** — verify they exercise the error path, not just the happy path
29. **Check `vendorHash` staleness** — `go-faster/yaml` became a direct dependency; the Nix vendor hash may need updating for non-Nix builds workflows
30. **Update `go.mod`** — ensure `go-faster/yaml` is in the direct dependencies block, not indirect
31. **Add a test for `EvaluateActionabilityWithDisabled` with ALL patterns disabled** — should be equivalent to `--no-actionability`
32. **Add a test for `ListActionabilityPatterns` output ordering** — verify priority order is maintained
33. **Refactor `runRecommendThreshold` to reduce `nilerr` lint warning** — the `_ = filepath.WalkDir` with `return nil` on error is flagged
34. **Add `mnd` nolint or extract constants for the magic numbers in `recommend_threshold.go`** — 5000 and 7 are flagged
35. **Review `cmd/recommend_threshold.go:48`** — the `nilerr` warning where `err != nil` but we `continue` — this is intentional (skip unreadable paths) but should be documented

### LOWER Priority — Documentation & Polish

36. **Update `docs/PERFORMANCE.md` with `--recommend-threshold` mention**
37. **Add `--diff-report` to the GitHub Actions template** in `templates/`
38. **Add YAML config example to HOW_TO_USE.md advanced scenarios** (already added a basic one)
39. **Update `SDK_DESIGN.md` if the SDK needs YAML config support** (currently SDK is JSON-only via `json` tags)
40. **Add a `CHANGELOG.md` entry for the local pre-commit hook** (once implemented)
41. **Create ADR for YAML config support** — document the bridge decision and its limitations
42. **Create ADR for `interface-method` pattern** — document the static-name-list approach and its limitations
43. **Update `docs/DOMAIN_LANGUAGE.md`** with "actionability pattern" and "clone type" definitions if missing
44. **Add `--list-patterns` output to HOW_TO_USE.md** pattern control section (text already mentions it but doesn't show expected output)
45. **Review all new CHANGELOG entries for accuracy** — verify each feature claim matches the implementation

### LOWER Priority — Future Features

46. **Type-aware `interface-method` detection** — use `go/types` to verify a FuncDecl actually implements an interface method, not just name matching
47. **`--diff-report` with SARIF output** — currently only text and JSON; SARIF diff would be useful for GitHub Security
48. **Configurable `maxInterfaceMethodBodyNodes`** — currently hardcoded to 4; some teams may want a higher/lower limit
49. **`--recommend-threshold` with JSON output** — currently text-only; `--json` flag for CI integration
50. **Batch `--disable-pattern` from config file** — `disabledPatterns: ["guard-clause", "raii-defer"]` in YAML/JSON config

---

## g) Questions (3)

### Q1: Should I install a local pre-commit hook RIGHT NOW to stop the daemon regression?

The daemon re-added `exhaustruct`/`tagliatelle` **twice** during this session. The `.github/workflows/lint-config-guard.yml` only runs on push/PR — it does nothing for local commits. A `.git/hooks/pre-commit` script that runs `scripts/check-disabled-linters.sh` would block any commit that re-adds forbidden linters.

**I cannot figure this out myself because:** I don't know if you want `.git/hooks/` files tracked (they're typically gitignored), or if you prefer a `.pre-commit-config.yaml` for the pre-commit framework, or if the daemon should be fixed at the source instead.

### Q2: Should `--disable-pattern` validate labels at startup (fail-fast on typos)?

Currently `--disable-pattern guard-clouse` (typo) silently does nothing — the user gets no feedback that their flag was ineffective. Adding validation would print an error like `unknown pattern "guard-clouse"; valid patterns: signature-only, interface-implementation, ...`.

**I cannot figure this out myself because:** There's a tradeoff. Fail-fast catches typos but breaks forward-compatibility (if someone's CI script uses a label that gets renamed in a future version, their build breaks). Silent-ignore is forward-compatible but user-unfriendly. The AGENTS.md "Error Handling" section says "Every user-facing error must be clear, actionable" which argues for fail-fast, but the project also values CI stability.

### Q3: Should the `printer/` package be split before adding more features?

The package has ~29 source files / 3500+ lines. The actionability patterns alone span 8+ files. Adding more patterns or output formats will make it worse. A `printer/actionability/` sub-package would isolate the pattern logic, but it's blocked by the circular dependency on `Printer`/`ReadFile`/`StatsPrinter` interfaces in `printer.go`.

**I cannot figure this out myself because:** This is a significant architectural decision (moving interfaces to a base package, updating all consumers) with high blast radius. The TODO_LIST has it as HIGH priority but blocked. I don't know if you want to invest in this refactor now or defer it until the pain is worse.
