# Status Report — P3 Code Quality, Tests, UX & Documentation Sprint

**Date:** 2026-07-16 06:35
**Session goal:** Execute the full P3 TODO list from the prior session's status report.
**Branch:** `fork` (uncommitted changes — 19 modified, 7 new files)

> **Resolution (2026-07-16, P4 session at 06:57):** All HIGH PRIORITY items from this report were resolved in the very next session: gocyclo on `runCmd` (DONE), recvcheck warnings (ALL 9 DONE), `ProcessedCloneGroup` exhaustruct (DONE via constructor), config validation for `--workers`/`--min-lines`/`--max-cache-entries` (DONE), `--quiet`/`--no-color`/`version --json` BDD tests (6 new), `version --short` flag (DONE), exit codes in `--help` (DONE), ADR-0013 (exit codes), ADR-0014 (SuppressionConfig). See `2026-07-16_06-57_p4-tests-quality-polish.md`.

---

## a) FULLY DONE

### Integration Tests (6 tests, all passing)

| Test                                                                    | What it verifies                                                                                                                  |
| ----------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `TestEvaluateActionabilityWithLabel_ErrorWrapping`                      | `if err != nil { return fmt.Errorf(...) }` → NonActionable (gets `error-propagation` label, NOT `error-wrapping` — see section d) |
| `TestEvaluateActionabilityWithLabel_CobraCommand`                       | `cobra.Command{}` composite lit → NonActionable, `PatternCobraBoilerplate`                                                        |
| `TestEvaluateActionabilityWithLabel_CobraCommand_WrongReceiver`         | `myapp.Command{}` → Actionable (receiver NOT "cobra"/"fang")                                                                      |
| `TestEvaluateActionabilityWithLabel_BuilderCallback`                    | 3 chained calls, 2+ receivers → NonActionable, `PatternBuilderCallback`                                                           |
| `TestEvaluateActionabilityWithLabel_BuilderCallback_BelowThreshold`     | 2 calls, 1 receiver → Actionable (below threshold)                                                                                |
| `TestEvaluateActionabilityWithLabel_TableDrivenTest_NonTestingReceiver` | RangeStmt with `x.Run()` (NOT `t.Run()`) → Actionable                                                                             |

### Exit Code System

- `cmd/exit_codes.go` — `ExitCodeForError()` maps errors to typed exit codes
- `cmd/exit_codes_test.go` — 9 subtests covering nil, context cancel, validation, config, internal, generic
- `cmd/art-dupl/main.go` — wired `ExitCodeForError(err)` replacing hardcoded `os.Exit(1)` and the `exitCodeInterrupt` constant

### CLI Flags

- **`--quiet`/`-q`** — registered in `addSharedFlags`, wired through `applyChangedBoolFlags`, checked in `printBuildingStatus`
- **`--no-color`** — registered in `addSharedFlags`, sets `NO_COLOR=1` env var in `runCmd`
- **`version` subcommand** — `cmd/version_cmd.go` with `--json` flag, registered in `root.go`, `VersionInfo` struct with 7 fields

### Code Quality Refactors

- **`SuppressionConfig` struct** — replaces 3 separate `bool/int/int` params across `printDupls`, `printCloneGroups`, `shouldSuppressGroup`. Updated ALL 6 callers (run_flags, baseline_cmd x2, run_all_modes, stats, cmd_test x2)
- **`parseOutputFormat()` extraction** — reduces `runCmd` complexity by moving the output-format switch into its own function
- **`runStandardAnalysis()` extraction** — separates analysis/printing logic from flag parsing

### Lint Configuration

- `.golangci.yml` — added `exhaustruct` and `gochecknoglobals` to `_test.go` exclusion list
- `.golangci.yml` — added `github.com/spf13/cobra.Command` to `exhaustruct.exclude` list

### Config Validation Hints

- `validateThreshold` — now includes "(use --threshold/-t with a value of 1 or higher; default is 5)" and "(use --min-lines for line-based filtering instead)"
- `validateSortCriteria` — now includes "(valid: size, occurrence, hash, total-tokens; use --sort/-s)"
- `validateDiffMode` — now includes "(valid: side-by-side, inline; use --diff)"

### Documentation

| Document                               | What was added                                                                                                                               |
| -------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `CHANGELOG.md`                         | 6 Added entries (quiet, no-color, exit codes, SuppressionConfig, tests, min-lines) + 5 Fixed entries                                         |
| `HOW_TO_USE.md`                        | 4 new sections: Line-Count Filtering, Debugging Token Output, Quiet and Color Control, Exit Codes table                                      |
| `TODO_LIST.md`                         | New "Completed (2026-07-16) — P3 Code Quality, Tests & UX" section with all items checked                                                    |
| `docs/adr/0011-min-lines-minimum.md`   | ADR for --min-lines design decision (minimum across all clones)                                                                              |
| `docs/adr/0012-dumptokens-iowriter.md` | ADR for dumpTokensOutput io.Writer injection                                                                                                 |
| `docs/ACTIONABILITY_PATTERNS.md`       | Comprehensive 15-pattern reference table with descriptions, examples, priority order                                                         |
| `AGENTS.md`                            | 7 new conventions: exit codes, SuppressionConfig, parseOutputFormat, quiet/no-color, version subcommand, actionability ordering, lint config |

### CLI Verification (end-to-end)

- `--min-lines 5` correctly suppressed 4-line clone groups
- `--min-lines 3` correctly showed 4-line clone groups
- `--dump-tokens` produced tab-separated token stream
- `version` subcommand printed text format
- `version --json` printed structured JSON with 7 fields

---

## b) PARTIALLY DONE

### Shell completion

**Status:** Already provided by Fang framework. Verified it exists via BDD tests. No custom implementation needed. But: I did NOT verify it works end-to-end via CLI (`art-dupl completion bash`). Marked as "done" but could be more thorough.

### Config file support (`.artdupl.yml`)

**Status:** Already exists as `--config/-c` flag accepting JSON files. I did NOT add YAML support. The original TODO mentioned `.artdupl.yml` but the existing system uses JSON (`config.LoadOptionalConfig`). I left it as-is and documented the JSON format. YAML is a future enhancement.

### `--help` text audit

**Status:** NOT explicitly performed. I added new flags with descriptions but did not audit existing flag descriptions for accuracy or completeness. The `--dump-tokens` flag description says "skip clone detection" which is accurate but could be more descriptive about the output format.

### `recvcheck` linter

**Status:** Already enabled in `.golangci.yml` (line 95). I confirmed it's there but did NOT investigate the 4 warnings in `domain/processed_clone.go` about mixed receiver types. These are pre-existing and may be intentional (string enums legitimately mix pointer/value receivers).

---

## c) NOT STARTED

1. **Progress output for long runs** — No spinner or percentage output added. `--quiet` suppresses status but no progress indicator was built for non-quiet mode.
2. **`.artdupl.yml` YAML config file** — Only JSON is supported. No YAML parser added.
3. **`--help` text comprehensive audit** — Did not review every flag description for accuracy.
4. **Fix 4 `recvcheck` warnings** in `domain/processed_clone.go` — Pre-existing, not investigated.
5. **Fix `domain.ProcessedCloneGroup` exhaustruct warning** — `cmd/run_output.go:131` still triggers it (missing `TokenCount` field in struct literal). This is a pre-existing pattern where TokenCount is set on the next line.
6. **`--files-from-stdin` documentation** — The flag exists (`-f`/`--files`) but I did not add explicit docs for stdin mode beyond what was already in HOW_TO_USE.md.

---

## d) TOTALLY FUCKED UP

### 1. `run_flags.go` rewrite — FIRST ATTEMPT BROKE THE FILE

I attempted to use `multiedit` to extract `runAnalysis` from `runCmd`, but the edits created a malformed function with a `placeholder` variable, a broken `return nil}` that merged with the comment of `setJSONPrinterFilesCount`, and a missing `context` import. Had to rewrite the entire file from scratch with `write`.

**Root cause:** I tried to do too much in a single `multiedit` — extracting functions, changing signatures, and moving code blocks simultaneously. The `old_string` matches were ambiguous and the new code was incomplete.

**Fix:** Rewrote `run_flags.go` completely with `write`, being careful about imports, function boundaries, and code flow.

### 2. Error wrapping test expected wrong label

`TestEvaluateActionabilityWithLabel_ErrorWrapping` initially asserted `PatternErrorWrapping` but the pipeline returns `PatternErrorPropagation` because `isPureErrorPropagation` is checked BEFORE `isErrorWrappingReturn` in the priority order. Both classify as NonActionable, but the label differs.

**Root cause:** I didn't trace the full pipeline order before writing the assertion. I looked at `isErrorWrappingBody` and assumed it would fire, but `isErrorOnlyIf` matches the same tree structure first.

**Fix:** Changed expected label to `PatternErrorPropagation` and documented why in the test comment.

### 3. `SuppressionConfig` refactor broke 6 callers

When I changed `printDupls` and `printCloneGroups` signatures, I forgot about callers in `baseline_cmd.go` (2 calls), `run_all_modes.go` (1 call), `stats.go` (1 call), and `cmd_test.go` (2 calls). The build failed with 4 "too many arguments" errors.

**Root cause:** I only thought about the `run_flags.go` call site and the function definitions. Didn't use `find_references` or grep for all callers before making the change.

**Fix:** Found all callers via grep and updated them one by one.

### 4. Binary name collision when testing CLI end-to-end

Named the binary `/tmp/art-dupl` which art-dupl tried to parse as a Go source file (`could not parse /tmp/art-dupl: '>' must be followed by a word`). The tool walks the current directory and the binary path contained bytes that looked like invalid Go.

**Fix:** Used `go run ./cmd/art-dupl/` instead of building a standalone binary.

### 5. Missing `mustCallExprWithSelector` helper

The integration test referenced a helper function that didn't exist in any test file. I had assumed it was defined in the prior session's tree test file, but it wasn't.

**Fix:** Added the helper function to the integration test file.

---

## e) WHAT WE SHOULD IMPROVE

### Process improvements

1. **Use `find_references` before signature changes** — The `SuppressionConfig` refactor broke 6 callers. LSP `find_references` would have found them all instantly. I used grep AFTER the build failed instead of BEFORE the change.

2. **Don't use `multiedit` for large structural refactors** — The `run_flags.go` rewrite broke because `multiedit` is designed for surgical find-and-replace, not function extraction with moving code blocks. Use `write` for full-file rewrites.

3. **Trace the full pipeline before writing test assertions** — The error wrapping test assumed `PatternErrorWrapping` but the pipeline returns `PatternErrorPropagation` due to priority ordering. Should have traced `evaluateActionabilityDetailed` before writing assertions.

4. **Test new flags end-to-end via CLI** — I verified `--min-lines` and `--dump-tokens` via `go run` but did NOT verify `--quiet`, `--no-color`, `version --json` via CLI. Only verified `version` via `go run`. Should test ALL new flags end-to-end.

5. **The `gocyclo` warning on `runCmd` is NOT fully resolved** — The LSP still reports gocyclo 16. The `parseOutputFormat` extraction reduced the actual complexity, but the function still has many branches (sortBy validation, config merge, output format set, validate, timeout, allFlag, dumpTokens, standard analysis). May need further extraction or a `//nolint:gocyclo` directive.

### Code improvements

6. **`run_flags.go` still imports `context` at package level** — The `runStandardAnalysis` function takes `context.Context` but `runCmd` gets it from `cmd.Context()`. This is correct but the import was added at the last minute and could be cleaner.

7. **`--no-color` uses `os.Setenv` which is a side effect** — A cleaner approach would be passing a `noColor bool` through the config and having the printer check it. But lipgloss checks `NO_COLOR` env var, so setting it is the pragmatic choice. Still, it's a global side effect from a flag.

8. **`VersionInfo` struct could use `omitzero` tags** — The json/v2 convention in this project is `omitzero` for custom types, but `VersionInfo` uses plain `json:"field"` tags. Minor inconsistency.

9. **`exit_codes.go` imports `context` and `errors`** — The `errors.Is` check for wrapped cancellation is correct but the import list could be minimized if we used a simpler check.

---

## f) Up to 50 Next Steps

### High priority (should do next)

1. **Commit all changes** — 19 modified + 7 new files are uncommitted. This is the most important next step.
2. **Verify `--quiet` suppresses status output end-to-end** — Run CLI with `--quiet` and confirm no progress messages.
3. **Verify `--no-color` disables colored output end-to-end** — Run CLI with `--no-color` and confirm no ANSI codes.
4. **Write tests for `--quiet` flag behavior** — Unit test that `printBuildingStatus` returns early when `cfg.Quiet` is true.
5. **Write tests for `version` subcommand** — Test text and JSON output, verify all 7 fields present.
6. **Fix `gocyclo` on `runCmd`** — Either extract more logic or add a targeted `//nolint` with justification.
7. **Write BDD test for exit codes** — `bdd/cli_commands_test.go` should verify exit code 2 for invalid config.
8. **Write BDD test for `version` subcommand** — Verify `art-dupl version --json` produces valid JSON.

### Medium priority (code quality)

9. **Fix 4 `recvcheck` warnings** in `domain/processed_clone.go` — Investigate mixed pointer/value receivers.
10. **Fix `ProcessedCloneGroup` exhaustruct warning** — `TokenCount` is set post-construction; consider a constructor function.
11. **Add `cobra.Command` to exhaustruct exclude** — Already done in this session, verify it's working in lint output.
12. **Progress output for long runs** — Add a simple spinner or file-count progress to stderr when not `--quiet`.
13. **YAML config file support** — `.artdupl.yml` parser alongside existing JSON support.
14. **`--help` text audit** — Verify every flag description is accurate and includes valid values.
15. **Config validation for `--workers`** — No validation that workers >= 0. Add check.
16. **Config validation for `--min-lines`** — No validation that min-lines >= 0. Add check.
17. **Config validation for `--max-cache-entries`** — No validation. Add check.
18. **Test `parseOutputFormat`** — Extracted function has no direct unit test.
19. **Test `runStandardAnalysis`** — Extracted function has no direct unit test (covered indirectly by integration tests).
20. **Test `ExitCodeForError` with wrapped internal errors** — Verify error wrapping chains work.

### Lower priority (polish)

21. **Add `--verbose` to `version` subcommand** — Show Go GC stats, build flags.
22. **Add `art-dupl version --short`** — Just the version string, no build info.
23. **Progress bar for `--all` mode** — Show which format/method is being generated.
24. **Colored output for `--rich-text`** — Verify color codes work with `--no-color`.
25. **Config file validation** — Validate JSON/YAML config file structure on load.
26. **`--config` with YAML** — Accept `.yml`/`.yaml` extensions.
27. **Deprecation warning for `--semantic`** — It's the default now; flag is redundant but not deprecated.
28. **Exit code documentation in `--help`** — Add exit code table to root command Long description.
29. **Shell completion testing** — Verify `art-dupl completion bash` produces valid bash script.
30. **Integration test: exit code 2 for bad threshold** — Run CLI with `-t -1` and check exit code.
31. **Integration test: exit code 2 for bad sort** — Run CLI with `--sort invalid` and check exit code.
32. **Integration test: exit code 0 for no clones** — Run CLI on empty dir and check exit code.
33. **`docs/ACTIONABILITY_PATTERNS.md` cross-link** — Link from HOW_TO_USE.md and AGENTS.md.
34. **ADR for exit codes** — Document the exit code design decision formally.
35. **ADR for SuppressionConfig** — Document the struct extraction rationale.
36. **Profile-guided optimization** — Run with `--profile` and check for hot spots.
37. **Memory usage for large repos** — Test on a large codebase (10000+ files).
38. **`--workers` auto-detection test** — Verify 0 defaults to CPU count.
39. **Cache invalidation on version change** — Verify `CacheVersion` bump works.
40. **Templ actionability patterns** — No actionability checks for templ files currently.
41. **Cross-file `goconst` management** — Document the cross-file goconst behavior for future contributors.
42. **`flake.nix` check for new flags** — Verify Nix flake CI passes with new flags.
43. **Pre-commit hook update** — Verify the 26-check BuildFlow passes with all changes.
44. **SDK exposure of exit codes** — Consider exposing `ExitCodeForError` in `pkg/artdupl`.
45. **SDK exposure of version info** — Consider exposing `VersionInfo` in `pkg/artdupl`.
46. **Benchmark SuppressionConfig vs bare params** — Verify no performance regression from struct passing.
47. **Document `NO_COLOR` env var in HOW_TO_USE.md** — Already mentioned but could be more prominent.
48. **Test `--no-color` with `--rich-text`** — Verify badges still show but without color.
49. **Test `--quiet` with `--json`** — Verify JSON output is unaffected by quiet mode.
50. **Update `FEATURES.md`** — Add exit codes, quiet/no-color flags, version subcommand to feature inventory.

---

## g) Top 2 Questions

### Q1: Should `gocyclo` on `runCmd` be fixed with further extraction or a `//nolint` directive?

The function is now at complexity 16 (threshold 15) per LSP. I extracted `parseOutputFormat` and `runStandardAnalysis`, but the remaining branches (sortBy validation, config merge, validate, timeout, allFlag, dumpTokens dispatch) are inherent to a command-dispatch function. Extracting more might make the flow harder to follow. Options:

- **A:** Add `//nolint:gocyclo // command dispatch inherently branches` and move on
- **B:** Extract a `dispatchAnalysis(ctx, cmd, cfg, sortBy)` that handles the allFlag/dumpTokens/standard dispatch
- **C:** Raise the gocyclo threshold from 15 to 20 in `.golangci.yml`

**My recommendation:** Option B — extract the dispatch logic. But I want your input before adding more indirection.

### Q2: Should `--no-color` set `NO_COLOR` env var (current) or pass a typed flag through config to the printer?

Current approach: `runCmd` calls `os.Setenv("NO_COLOR", "1")` when `--no-color` is passed. This is a global side effect from a flag — it affects not just art-dupl but any subprocess it spawns. Alternative: add `NoColor bool` to `Config`, pass it to the printer, and have the printer's color logic check both `cfg.NoColor` and `os.LookupEnv("NO_COLOR")`. This is cleaner but requires changes to every printer that uses color (text, HTML, rich-text).

**My recommendation:** Keep the current `os.Setenv` approach — lipgloss checks `NO_COLOR` natively, it's the standard Go ecosystem convention, and the side effect is intentional and documented. But this is a design decision worth confirming.
