# Status Report — P4 Code Quality, Tests, CLI Polish & Hard Self-Critique

**Date:** 2026-07-16 06:57
**Session goal:** Execute the remaining P3 TODO list items (the 50 next steps from the prior session's status report).
**Branch:** `fork`, head `9307ace4` (uncommitted — 17 modified, 5 new files)

---

## a) FULLY DONE

### 1. gocyclo on `runCmd` — RESOLVED

**What:** Extracted `dispatchAnalysis()` from `runCmd` to handle the `allFlag`/`dumpTokens`/`standardAnalysis` routing. `runCmd` now delegates all post-config dispatch to this function.

**Before:** gocyclo 16 (threshold 15)
**After:** 0 gocyclo warnings in `cmd/run_flags.go`

**Files:** `cmd/run_flags.go:101-117` (new `dispatchAnalysis` function)

### 2. recvcheck warnings — ALL 9 RESOLVED

**What:** All 9 domain string-enum types (CloneCategory, ClonePriority, CloneActionability, CloneType, DetectionMethod, DiffMode, OutputFormat, SortCriteria, HealthScore) mixed value/pointer receivers. This is the standard Go JSON convention (`MarshalJSON` on value receiver, `UnmarshalJSON` on pointer receiver). Added `//nolint:recvcheck` with explanatory comment to each.

**Files:** `domain/processed_clone.go`, `domain/detection_method.go`, `domain/diff_mode.go`, `domain/output_format.go`, `domain/sort_criteria.go`, `domain/types_health.go`

**Verified:** `golangci-lint run ./domain/... 2>&1 | grep recvcheck` returns 0 results.

### 3. ProcessedCloneGroup exhaustruct — RESOLVED via constructor

**What:** Added `NewProcessedCloneGroup(hash, clones)` constructor to `domain/processed_clone.go` that computes `TokenCount` from clones automatically. Updated 3 production call sites:

- `cmd/run_output.go:131` — `group := domain.NewProcessedCloneGroup(k, clones)`
- `printer/clone_processor.go:215` — `return domain.NewProcessedCloneGroup(hash, clones), nil`
- `printer/html.go:300` — `domain.NewProcessedCloneGroup("", []domain.ProcessedClone{cl})`

**Before:** exhaustruct warning on `ProcessedCloneGroup{Hash: k, Clones: clones}` (missing `TokenCount`)
**After:** No exhaustruct warning (constructor handles TokenCount internally)

### 4. Config validation for Workers, MinLines, MaxCacheEntries

**What:** Added `validateNonNegative(name, value)` to `config/config_validate.go`. Wired into `ValidateConfig` for all three previously-unvalidated fields.

**Files:** `config/config_validate.go:28-30` (3 new validation entries), `config/config_validate.go:152-162` (helper function)

**Tests:** 2 integration tests (`negative_workers`, `negative_min-lines`) verify exit code 2 for negative values.

### 5. `printSearchStatus` quiet bug — FIXED

**What:** `printSearchStatus` was NOT checking `cfg.Quiet`, so even with `--quiet`, the search-status emoji ("✅") was still printed to stderr. Added the same early-return guard that `printBuildingStatus` already had.

**Verified end-to-end:** `go run ./cmd/art-dupl/ --quiet /tmp/...` produces 0 bytes on stderr.

### 6. Unit Tests (15 new, all passing)

| Test File               | Tests | What                                                                                              |
| ----------------------- | ----- | ------------------------------------------------------------------------------------------------- |
| `cmd/run_flags_test.go` | 4     | `printBuildingStatus` quiet/non-quiet × verbose/non-verbose                                       |
| `cmd/run_flags_test.go` | 3     | `version` subcommand: text, JSON (all 7 fields), short                                            |
| `cmd/run_flags_test.go` | 6     | `parseOutputFormat` for all 6 output formats                                                      |
| `cmd/run_flags_test.go` | 3     | `ExitCodeForError` with wrapped internal, wrapped validation (errors.Join), nested wrapped cancel |

### 7. Integration Tests (5 new, all passing)

| Test                                          | Verifies                        |
| --------------------------------------------- | ------------------------------- |
| `bad_threshold_maps_to_config_exit_code`      | Threshold=-1 → exit 2           |
| `bad_sort_maps_to_config_exit_code`           | SortCriteria="invalid" → exit 2 |
| `negative_workers_maps_to_config_exit_code`   | Workers=-3 → exit 2             |
| `negative_min-lines_maps_to_config_exit_code` | MinLines=-10 → exit 2           |
| `nil_error_maps_to_success_exit_code`         | nil error → exit 0              |

### 8. BDD Tests (6 new scenarios, all passing)

| Scenario                | Verifies                                 |
| ----------------------- | ---------------------------------------- |
| version without flags   | stdout contains "art-dupl" and "version" |
| version --json          | stdout is valid JSON with all 7 fields   |
| version --short         | stdout is just the version string        |
| --threshold -1          | returns error                            |
| --sort invalid          | returns error                            |
| --help shows exit codes | contains "Exit codes" and "130"          |

### 9. `version --short`/`-s` flag

**What:** Added `--short`/`-s` flag to version subcommand. Prints just the version string (e.g., `dev\n`).

**Files:** `cmd/version_cmd.go:30-34` (short flag check), `cmd/version_cmd.go:60` (flag registration)

### 10. Exit codes in `--help` output

**What:** Root command Long description now includes the exit code table (0/1/2/3/130).

**Files:** `cmd/root.go:12-18`

### 11. Version cmd uses `cmd.OutOrStdout()`

**What:** Changed `fmt.Println(...)` to `fmt.Fprintln(cmd.OutOrStdout(), ...)` in version subcommand. This makes the command testable — output goes to the cobra command's output buffer during tests instead of raw stdout.

**Why:** The first test attempt (`TestVersionCommand_Short`) failed because `fmt.Println` bypassed cobra's output redirection.

### 12. ADRs

| ADR                                   | Topic                                                                                  |
| ------------------------------------- | -------------------------------------------------------------------------------------- |
| `docs/adr/0013-exit-codes.md`         | Typed exit code system (0/1/2/3/130), `ExitCodeForError`, `errors.Is` chain unwrapping |
| `docs/adr/0014-suppression-config.md` | `SuppressionConfig` struct extraction, type safety, extensibility                      |

### 13. Documentation Updates

| Document       | Changes                                                                                                                                        |
| -------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `FEATURES.md`  | Updated CLI section: added Typed Exit Codes, Color Control, Line-Count Filtering; expanded Version Information and Verbosity entries           |
| `CHANGELOG.md` | 10 new entries under Added (version --short, exit codes in help, config validation, NewProcessedCloneGroup, dispatchAnalysis, 20+ tests, ADRs) |
| `TODO_LIST.md` | New "Completed (2026-07-16) — P4 Tests, Code Quality & CLI Polish" section with all items                                                      |

### 14. End-to-End Verification

| Flag/Feature                     | Verified                              |
| -------------------------------- | ------------------------------------- |
| `--quiet`                        | 0 bytes on stderr (wc -c confirmed)   |
| `--no-color`                     | No ANSI escape codes in output        |
| `version --short`                | Prints `dev\n`                        |
| `version --json`                 | Valid JSON with all 7 fields          |
| Exit codes in `--help`           | Table visible in help output          |
| `go build ./...`                 | PASS                                  |
| `go test ./... -count=1`         | 25/25 packages PASS                   |
| `go test -race`                  | PASS (cmd/printer/job/pkg)            |
| `golangci-lint` on changed files | 0 recvcheck/gocyclo/errcheck warnings |

---

## b) PARTIALLY DONE

### 1. Progress output for long runs

**Status:** NOT STARTED. The prior session flagged this as item #12 in the 50 next steps. No spinner or file-count progress was added. The `--quiet` flag suppresses status, but in non-quiet mode, there's still no progress indicator for long-running analyses. This is a feature gap, not a bug.

### 2. `--help` text audit

**Status:** PARTIALLY DONE. I added exit codes to the root command Long description and verified the new flag descriptions are accurate. But I did NOT audit every existing flag's description for accuracy or completeness. The prior session noted that `--dump-tokens` says "skip clone detection" which is accurate but could be more descriptive.

### 3. `--files-from-stdin` documentation

**Status:** NOT STARTED. The flag exists (`-f`/`--files`) but explicit docs for stdin mode beyond what was already in HOW_TO_USE.md were not added.

### 4. Cross-link ACTIONABILITY_PATTERNS.md

**Status:** NOT STARTED. `docs/ACTIONABILITY_PATTERNS.md` exists but is not linked from `HOW_TO_USE.md` or `AGENTS.md`. It's only discoverable by browsing the `docs/` directory.

### 5. `ProcessedCloneGroup` exhaustruct in test files

**Status:** NOT FULLY DONE. I fixed the production call sites but left the test-file struct literals (`domain.ProcessedCloneGroup{Clones: ...}`) as-is since `exhaustruct` is excluded for `_test.go` files. This is correct — test files are allowed to use partial initialization.

---

## c) NOT STARTED

1. **Progress output for long runs** — No spinner or percentage output. `--quiet` suppresses but no progress indicator exists for non-quiet mode.
2. **`.artdupl.yml` YAML config file** — Only JSON is supported. No YAML parser added.
3. **Shell completion end-to-end test** — `art-dupl completion bash` was NOT verified to produce a valid bash script.
4. **`--workers` auto-detection test** — Not verified that 0 defaults to NumCPU.
5. **Cache invalidation on version change** — Not tested.
6. **SDK exposure of exit codes / version info** — `ExitCodeForError` and `VersionInfo` are not exposed in `pkg/artdupl`.
7. **Templ actionability patterns** — No actionability checks for templ files (all actionability patterns are Go-specific).
8. **`--verbose` to version subcommand** — Not added.
9. **Progress bar for `--all` mode** — Not added.
10. **Config file structure validation** — Not added.
11. **Profile-guided optimization** — Not run.
12. **Memory usage for large repos** — Not tested.
13. **Deprecation warning for `--semantic`** — Not added (it's the default now, flag is redundant).
14. **Benchmark SuppressionConfig vs bare params** — Not run.

---

## d) TOTALLY FUCKED UP

### 1. `TestVersionCommand_Short` — First attempt wrote to stdout, not test buffer

**What happened:** I initially wrote the version subcommand to use `fmt.Println(Version)` which writes directly to `os.Stdout`. The test set `cmd.SetOut(&bytes.Buffer{})`, but `fmt.Println` ignores cobra's output writer. The test captured an empty buffer.

**Root cause:** I didn't understand that `cobra.Command.OutOrStdout()` is the testable output path. `fmt.Println` is the raw-stdout path.

**Fix:** Changed `fmt.Println(...)` to `fmt.Fprintln(cmd.OutOrStdout(), ...)` in version_cmd.go. Added `//nolint:forbidigo` since forbidigo flags `fmt.Print*` calls.

**Lesson:** Always use `cmd.OutOrStdout()` / `cmd.OutOrStderr()` in cobra RunE functions. Never `fmt.Println`.

### 2. `TestExitCodes_ConfigValidation` — threshold=-1 returned exit code 1, not 2

**What happened:** I called `config.ValidateConfig(cfg)` directly and passed the error to `ExitCodeForError`. But `validateThreshold` returns a plain `fmt.Errorf("%w: ...", ErrInvalidThreshold)` — it wraps a sentinel error, not a `duplerrors.ValidationError`. So `ExitCodeForError` fell through to `ExitGeneralError` (1).

**Root cause:** I didn't trace the error wrapping chain. In production, `runCmd` wraps the entire validation error in `duplerrors.WrapValidation(...)`, which is what `ExitCodeForError` checks for. My test bypassed that wrapping.

**Fix:** Changed the test to wrap the error the same way `runCmd` does: `duplerrors.WrapValidation(err, "config validation failed")`. This correctly produces exit code 2.

**Lesson:** Test through the same wrapping path as production code. Don't call internal functions directly and expect the same error classification.

### 3. BDD `version` test treated `version` as a file path

**What happened:** `setup.RunArtDupl("version")` calls `RunArtDuplOnDir(s.TmpDir, "version")`, which prepends the tmp dir as the first argument. So cobra received `["<tmpDir>", "version"]` — treating `version` as a file path to analyze, not as a subcommand.

**Root cause:** I didn't read the `RunArtDupl` helper closely. It's designed for `art-dupl [flags] [paths...]`, not subcommands. For subcommands, you must use `setup.ExecutorResult("version")` directly.

**Fix:** Changed all BDD version subcommand tests to use `setup.ExecutorResult(...)` instead of `setup.RunArtDupl(...)`.

### 4. `printSearchStatus` quiet bug — MISSED in prior session

**What happened:** The prior session added `cfg.Quiet` check to `printBuildingStatus` but NOT to `printSearchStatus`. So `--quiet` still printed the search-status emoji ("✅") to stderr. I only discovered this during end-to-end verification when stderr had content despite `--quiet`.

**Root cause:** The prior session didn't verify `--quiet` end-to-end. I caught it because I ran the CLI and checked stderr byte count.

**Fix:** Added `if cfg.Quiet { return }` to `printSearchStatus`.

**Lesson:** End-to-end verification catches bugs that unit tests miss. Always run the actual binary.

### 5. `exit_codes_integration_test.go` — first draft was overcomplicated

**What happened:** I initially wrote a `Config_for_test` struct and `validateTestConfig` function that duplicated the real validation logic. This was unnecessary — I should have used `config.Config` and `config.ValidateConfig` directly.

**Root cause:** I overthought the test structure instead of using the existing API.

**Fix:** Rewrote the entire test file to use `config.Config`, `config.ValidateConfig`, and `ExitCodeForError` directly.

---

## e) WHAT WE SHOULD IMPROVE

### Process improvements

1. **Read helper function signatures before using them** — The BDD `version` test failed because I didn't read `RunArtDupl` closely. It prepends the tmp dir. For subcommands, `ExecutorResult` is the right helper. Should have read the helper before writing the test.

2. **Trace error wrapping chains before writing exit-code assertions** — The threshold test failed because I didn't trace how `runCmd` wraps `config.ValidateConfig` errors. The wrapping layer matters for `ExitCodeForError`. Should have traced `runCmd → duplerrors.WrapValidation → ExitCodeForError`.

3. **Use `cmd.OutOrStdout()` in all cobra RunE functions** — The version test failed because `fmt.Println` bypasses cobra's output buffer. This is a cobra testing fundamental. Should have known from the start.

4. **Verify ALL new flags end-to-end via CLI, not just unit tests** — The `printSearchStatus` quiet bug was caught only because I ran the CLI and checked stderr byte count. Unit tests for `printBuildingStatus` passed because they tested that function in isolation. The bug was in a DIFFERENT function (`printSearchStatus`) that the unit tests didn't cover.

5. **Check for sibling functions when adding guards** — When I added `cfg.Quiet` to `printBuildingStatus` in the prior session, I should have searched for ALL status-output functions and added the guard to all of them at once. Instead, `printSearchStatus` was missed.

### Code improvements

6. **`domain.ProcessedCloneGroup` constructor doesn't validate** — `NewProcessedCloneGroup` computes `TokenCount` but doesn't call `Validate()`. Callers can still create groups with empty clones or invalid individual clones. A validating constructor would be stronger.

7. **`validateNonNegative` is too generic** — The error message is `"%s must be non-negative: %d"`. It doesn't include the flag name (e.g., `--workers`). The other validators include actionable hints like `(use --threshold/-t with a value of 1 or higher)`. `validateNonNegative` should too.

8. **`dispatchAnalysis` still accesses flags directly** — The function reads `cmd.Flags().GetBool("all")` and `cmd.Flags().GetBool("dump-tokens")`. These could be passed as parameters for testability. But the function is simple enough that this is acceptable.

9. **Test for `dispatchAnalysis` was started then deleted** — I initially wrote a `TestDispatchAnalysis` test but it required mocking too many internals (executeAnalysis, createPrinter, etc.). Removed it because it wasn't testing anything meaningful. The function is covered indirectly by BDD tests.

10. **`exit_codes_integration_test.go` tests `ExitCodeForError` not actual process exit codes** — The tests verify the function returns the right code, but don't actually run the binary and check `$?`. The BDD tests are closer to this (they check `err != nil`), but neither checks the actual OS exit code. A true integration test would `exec.Command("go", "run", "./cmd/art-dupl/", "--threshold", "-1")` and check `cmd.ProcessState.ExitCode()`.

11. **No test for `NewProcessedCloneGroup`** — The new constructor has no direct unit test. It's covered indirectly by the clone_processor tests, but a focused test (`TestNewProcessedCloneGroup_ComputesTokenCount`) would be more explicit.

12. **`version_cmd.go` still has stale LSP warnings** — gci (file not properly formatted) and forbidigo (fmt.Println pattern). These are LSP artifacts from the `fmt.Fprintln` calls — the `//nolint:forbidigo` directives are present but the LSP hasn't refreshed. Verified clean via `golangci-lint run` directly.

---

## f) Up to 50 Next Steps

### High priority (should do next)

1. **Commit all changes** — 17 modified + 5 new files are uncommitted.
2. **Add `NewProcessedCloneGroup` unit test** — Direct test for the new constructor.
3. **Improve `validateNonNegative` error messages** — Include flag names (e.g., "--workers must be non-negative" not "workers must be non-negative").
4. **Cross-link `docs/ACTIONABILITY_PATTERNS.md`** — Link from HOW_TO_USE.md and AGENTS.md.
5. **Add progress output for long runs** — Simple file count or spinner to stderr when not `--quiet`.
6. **True exit code integration test** — Run `exec.Command` and check actual process exit code, not just `ExitCodeForError` return value.
7. **Audit all flag descriptions in `--help`** — Verify every flag's `Short` and usage string is accurate and helpful.
8. **Test `--workers 0` auto-detection** — Verify 0 defaults to NumCPU.
9. **Document `NO_COLOR` env var more prominently** — Already mentioned in HOW_TO_USE.md but could be more visible.
10. **SDK: expose `ExitCodeForError`** — Consider adding to `pkg/artdupl` for library consumers.

### Medium priority (code quality)

11. **Add `Validate()` call in `NewProcessedCloneGroup`** — Or at least guard against empty clones.
12. **Refactor `dispatchAnalysis` to accept parameters** — Instead of reading flags directly, for testability.
13. **Test `--quiet` with `--json`** — Verify JSON output to stdout is unaffected by quiet mode.
14. **Test `--no-color` with `--rich-text`** — Verify badges still show but without color codes.
15. **Add `--verbose` to version subcommand** — Show GC stats, build flags.
16. **YAML config file support** — `.artdupl.yml` parser alongside existing JSON support.
17. **Config file structure validation** — Validate JSON/YAML config on load.
18. **Progress bar for `--all` mode** — Show which format/method is being generated.
19. **Deprecation warning for `--semantic`** — It's the default now; flag is redundant.
20. **Shell completion end-to-end test** — Verify `art-dupl completion bash` produces valid bash.
21. **Profile-guided optimization** — Run with `--profile` and check for hot spots.
22. **Memory usage for large repos** — Test on 10000+ files.
23. **Cache invalidation on version change** — Verify `CacheVersion` bump works.
24. **Templ actionability patterns** — No actionability checks for templ files currently.
25. **Benchmark `NewProcessedCloneGroup` vs inline** — Verify no perf regression.
26. **Add `cmd.OutOrStderr()` to `printBuildingStatus`/`printSearchStatus`** — Currently uses raw `os.Stderr`. Should use cobra's writer for testability.
27. **Consider `NoColor bool` in Config** — Current `os.Setenv("NO_COLOR", "1")` is a global side effect. A typed config flag through to the printer would be cleaner.
28. **Test wrapped internal errors through full pipeline** — Not just `ExitCodeForError` but actual `runCmd` → `main.go` → `os.Exit`.
29. **Add goconst documentation** — Document the cross-file goconst behavior for contributors.
30. **Verify Nix flake CI passes** — `nix flake check` with new flags.

### Lower priority (polish)

31. **`--version --verbose`** — Show Go version, compiler, platform in text mode too.
32. **Colored `--help` output** — Fang already does this, verify it's working.
33. **`art-dupl completion zsh`** — Verify zsh completion works.
34. **`art-dupl completion fish`** — Verify fish completion works.
35. **`art-dupl completion powershell`** — Verify PowerShell completion works.
36. **SDK: expose `VersionInfo`** — For library consumers who want version metadata.
37. **Add `--dry-run` flag** — Show what would be analyzed without running detection.
38. **Add `--progress` flag** — Explicit progress control (always/never/auto).
39. **HTML report: add exit code info** — Show exit code in HTML summary.
40. **SARIF: add exit code to rule metadata** — For CI tooling integration.
41. **Config: add `Config.NoColor` field** — Complement `os.Setenv` approach.
42. **Config: validate `DetectionMode` mutually exclusive flags** — `--semantic`/`--exact`/`--structural`.
43. **Test: `config.ValidateConfig` with all fields set** — Comprehensive validation test.
44. **Doc: add `docs/CLI_FLAGS.md`** — Comprehensive flag reference with examples.
45. **Doc: add exit codes to README.md** — Currently only in HOW_TO_USE.md.
46. **Refactor: extract `StatusPrinter` interface** — Bundle `printBuildingStatus` + `printSearchStatus` into a struct.
47. **Test: concurrent `ExitCodeForError` calls** — Verify thread safety.
48. **Lint: add `testifylint`** — Enforce testify usage patterns.
49. **Lint: verify `wrapcheck` covers new files** — Check error wrapping in `dispatchAnalysis`.
50. **ADR: document `NewProcessedCloneGroup` constructor pattern** — When to use constructor vs struct literal.

---

## g) Top 2 Questions

### Q1: Should `validateNonNegative` include the flag name in the error message?

Current: `"workers must be non-negative: -3"`
Could be: `"--workers must be non-negative: -3 (use 0 for auto-detection)"`

The other validators include flag names and actionable hints. `validateNonNegative` is generic and doesn't know which flag it's validating. Options:

- **A:** Pass the flag name as a third parameter: `validateNonNegative("workers", "--workers", cfg.Workers)`
- **B:** Create three separate validators: `validateWorkers`, `validateMinLines`, `validateMaxCacheEntries` — each with flag-specific hints
- **C:** Keep generic, accept the current message quality

**My recommendation:** Option B — separate validators with specific hints, matching the existing `validateThreshold` pattern. But I want to avoid premature complexity for 3 simple non-negative checks.

### Q2: Should `NewProcessedCloneGroup` call `Validate()` internally?

Current: Constructor computes `TokenCount` but doesn't validate clones.
Options:

- **A:** Call `g.Validate()` inside the constructor and panic/error on invalid input
- **B:** Keep constructor simple, let callers validate separately
- **C:** Add a `MustNewProcessedCloneGroup` variant that panics on invalid input

**My recommendation:** Option B — keep the constructor simple. It's a value constructor, not a validation gate. The `Validate()` method exists for when validation is needed (e.g., at API boundaries). Adding validation to the constructor would couple construction with validation, making it harder to use in intermediate states.

---

## Files & Changes (uncommitted)

### Modified files (17):

| File                         | Changes                                                                                                                   |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| `CHANGELOG.md`               | 10 new Added entries (version --short, exit codes in help, config validation, constructor, dispatchAnalysis, tests, ADRs) |
| `FEATURES.md`                | Updated CLI section: exit codes, quiet/no-color, version --short, line-count filtering; updated date to 2026-07-16        |
| `TODO_LIST.md`               | New "Completed P4" section (code quality, tests, CLI features, docs, E2E verification)                                    |
| `cmd/root.go`                | Exit code table added to Long description                                                                                 |
| `cmd/run_analysis.go`        | `printSearchStatus` now checks `cfg.Quiet`                                                                                |
| `cmd/run_flags.go`           | Extracted `dispatchAnalysis()` from `runCmd`                                                                              |
| `cmd/run_output.go`          | Uses `NewProcessedCloneGroup` constructor                                                                                 |
| `cmd/version_cmd.go`         | Added `--short`/`-s` flag; uses `cmd.OutOrStdout()` instead of `fmt.Println`; errcheck-safe                               |
| `config/config_validate.go`  | Added `validateNonNegative`; 3 new validation entries                                                                     |
| `domain/detection_method.go` | `//nolint:recvcheck` on type declaration                                                                                  |
| `domain/diff_mode.go`        | `//nolint:recvcheck` on type declaration                                                                                  |
| `domain/output_format.go`    | `//nolint:recvcheck` on type declaration                                                                                  |
| `domain/processed_clone.go`  | `//nolint:recvcheck` on 4 type declarations; `NewProcessedCloneGroup` constructor                                         |
| `domain/sort_criteria.go`    | `//nolint:recvcheck` on type declaration                                                                                  |
| `domain/types_health.go`     | `//nolint:recvcheck` on type declaration                                                                                  |
| `printer/clone_processor.go` | Uses `NewProcessedCloneGroup` constructor (removed manual TokenCount loop)                                                |
| `printer/html.go`            | Uses `NewProcessedCloneGroup` constructor                                                                                 |

### New files (5):

| File                                  | Lines | Description                                                                                            |
| ------------------------------------- | ----- | ------------------------------------------------------------------------------------------------------ |
| `cmd/run_flags_test.go`               | 182   | 15 unit tests: printBuildingStatus (4), version cmd (3), parseOutputFormat (6), wrapped exit codes (3) |
| `cmd/exit_codes_integration_test.go`  | 107   | 5 integration tests: config validation → exit code mapping                                             |
| `bdd/exit_codes_test.go`              | 88    | 6 BDD scenarios: version subcommand, exit code errors, exit codes in help                              |
| `docs/adr/0013-exit-codes.md`         | 38    | ADR for typed exit code system                                                                         |
| `docs/adr/0014-suppression-config.md` | 36    | ADR for SuppressionConfig struct extraction                                                            |

---

## Verification Summary

| Check                                                                | Status                   |
| -------------------------------------------------------------------- | ------------------------ |
| `go build ./...`                                                     | PASS                     |
| `go test ./... -count=1`                                             | 25/25 PASS               |
| `go test -race -count=1 ./cmd/... ./printer/... ./job/... ./pkg/...` | PASS                     |
| `golangci-lint` recvcheck on domain/                                 | 0 warnings               |
| `golangci-lint` gocyclo on cmd/                                      | 0 warnings               |
| `golangci-lint` errcheck on changed files                            | 0 warnings               |
| `--quiet` E2E (stderr byte count)                                    | 0 bytes                  |
| `--no-color` E2E (no ANSI codes)                                     | Verified                 |
| `version --short` E2E                                                | Prints `dev\n`           |
| `version --json` E2E                                                 | Valid JSON, all 7 fields |
| Exit codes in `--help`                                               | Table visible            |
