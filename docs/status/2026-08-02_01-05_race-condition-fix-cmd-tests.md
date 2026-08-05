# Status Report: 2026-08-02 Race Condition Fix in `cmd` Tests

> **Post-session annotation (2026-08-05):** Race fix is stable — `go test -race ./...` passes across all packages. The **root architectural cause remains open** (section C/E1: inject output writers into production code). Harvested into TODO_LIST.md as HIGH priority. The `nix flake check` loop was closed in later sessions (CI is green except for the recurring tagliatelle issue). Section F items are brainstorms — key actionable ones harvested into TODO_LIST/ROADMAP.

> **Session**: Single-session race-condition investigation and fix
> **Date**: 2026-08-02 01:05 CEST
> **Trigger**: `nix flake check` → `test-race` failure (4 named tests + ~20 collateral FAILs)
> **Outcome**: Races eliminated. `go test -race ./...` passes (26/26 packages, exit 0, stable x3).

---

## A) FULLY DONE

### Race condition root-caused and fixed

**Root cause**: `testutil.CaptureStdoutStderr` mutates the process-global `os.Stdout`/`os.Stderr` pointers. Two `t.Parallel()` integration tests used this helper while parallel sibling tests read those globals directly. The `captureMu` mutex only serialized captures against each other, NOT against direct readers. The race was on the pointer read/write itself.

**Data race pairs identified**:

| Writer (capture test)                      | Reader (parallel sibling)                 | Read site                                        |
| ------------------------------------------ | ----------------------------------------- | ------------------------------------------------ |
| `TestStatsHonorsAcceptDirectives`          | `TestVersionCommand_Text`                 | `fmt.Printf` → `os.Stdout` (`version.go:44`)     |
| `TestStatsHonorsAcceptDirectives`          | `TestPrintBuildingStatus_NonQuietVerbose` | `fmt.Fprintln(os.Stderr)` (`run_analysis.go:84`) |
| `TestBaselineRecordHonorsAcceptDirectives` | (same readers, different timing)          | (same)                                           |

### Files changed (4 files, +53/-6 lines)

| File                                       | Change                                             | Purpose                                                                                      |
| ------------------------------------------ | -------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| `cmd/accept_directive_integration_test.go` | Removed `t.Parallel()` from 2 tests                | Eliminate the race                                                                           |
| `internal/testutil/capture.go`             | Corrected misleading docs                          | The old comment said "safe for concurrent use" — it's not, in the presence of direct readers |
| `cmd/stats_integration_test.go`            | Added serial-contract note to `executeTestCommand` | Prevent future regressions via the helper                                                    |
| `TESTING.md`                               | Added "Global state and `t.Parallel()`" section    | Codify the rule for all future tests                                                         |

### Verification performed

- `go test -race ./cmd/` — PASS (single run)
- `go test -race -count=3 ./cmd/` — PASS (3 consecutive runs, no flakiness)
- `go test -race ./...` — PASS (all 26 packages)
- `go build ./...` — clean
- `go vet ./cmd/ ./internal/testutil/` — clean
- `golangci-lint run ./cmd/... ./internal/testutil/...` — only pre-existing `godox` finding in `suppression_config_test.go` (not touched)

### Inventory: all global stdout/stderr mutators in `cmd/` tests

| Test                                       | File                                   | Mutates                               | Parallel?        | Status        |
| ------------------------------------------ | -------------------------------------- | ------------------------------------- | ---------------- | ------------- |
| `TestStatsHonorsAcceptDirectives`          | `accept_directive_integration_test.go` | `os.Stdout`+`os.Stderr` (via capture) | ~~Yes~~ → **No** | **FIXED**     |
| `TestBaselineRecordHonorsAcceptDirectives` | `accept_directive_integration_test.go` | `os.Stdout`+`os.Stderr` (via capture) | ~~Yes~~ → **No** | **FIXED**     |
| `TestWarnTypeAwareIncremental`             | `config_validation_test.go:126`        | `os.Stderr` (direct)                  | No               | Safe (serial) |
| `TestStdinFeed_StderrSuppressedOnCancel`   | `run_crawl_stdin_test.go:259`          | `os.Stderr` (direct)                  | No               | Safe (serial) |

---

## B) PARTIALLY DONE

### Documentation hardening

The `capture.go` and `TESTING.md` docs now state the serial contract, but the note is only on `executeTestCommand` and the two specific tests. The other 5 test functions that call `executeTestCommand` (in `stats_integration_test.go`) don't have the note on themselves — they rely on the helper's doc. Acceptable, but not maximally defensive.

### Root-cause analysis scope

I identified the architectural root cause (production code writes to `os.Stdout`/`os.Stderr` directly instead of accepting injectable writers) but did NOT fix it. I only patched the symptom (test parallelism).

---

## C) NOT STARTED

### Architectural fix: injectable output writers

The real fix for this entire class of bugs is to stop having production code write to global `os.Stdout`/`os.Stderr`. Instead:

- `PrintVersion()` should accept an `io.Writer` or use cobra's `cmd.OutOrStdout()`
- `printBuildingStatus()` should accept an `io.Writer`
- All `fmt.Fprintf(os.Stderr, ...)` calls in `run_analysis.go`, `run_crawl.go`, `run_hash.go`, `run_all_modes.go`, `baseline_cmd.go`, `type_aware.go`, `stats.go`, `gitignore.go` should route through an injectable writer or cobra's `cmd.ErrOrStderr()`

This would eliminate the need for `CaptureStdoutStderr` entirely and make all tests safely parallelizable.

### Other direct `os.Stderr` mutators not consolidated

`TestWarnTypeAwareIncremental` and `TestStdinFeed_StderrSuppressedOnCancel` still hand-roll `os.Stderr = w` / `os.Stderr = oldStderr` instead of using `CaptureStdoutStderr`. They work and are serial, but they're inconsistent and fragile — anyone adding `t.Parallel()` to them later reintroduces the race.

### `bdd/execute.go` capture safety not deeply verified

`bdd/execute.go:22` uses `CaptureStdoutStderr`. The BDD package passed `-race`, but I didn't verify whether any BDD sibling test or parallel Ginkgo node reads `os.Stdout`/`os.Stderr` directly. Ginkgo runs specs sequentially by default, so this is likely safe, but it wasn't explicitly confirmed.

### Nix flake check not run

The original failure was from `nix flake check` → `buildflow test-race`. I verified via `go test -race` directly but did NOT re-run `nix flake check` to confirm the exact CI path passes. The Go command is what Nix invokes under the hood, so it should be equivalent, but I didn't close the loop on the exact failing command.

---

## D) TOTALLY FUCKED UP

Nothing. No regressions introduced, no files damaged, no false fixes. The fix is correct, minimal, and well-documented. Tests pass.

---

## E) WHAT WE SHOULD IMPROVE

### E1. Symptom-fix vs root-cause-fix

**This is the biggest issue.** I patched `t.Parallel()` removal instead of fixing the architectural smell that CAUSES the race. The production code has **23+ direct writes to `os.Stderr`** and **2 direct writes to `os.Stdout`** across `cmd/`. Every one of these is a potential race source for any future capture-based parallel test. Removing `t.Parallel()` from 2 tests is a band-aid; the disease is global mutable state in production code.

**The right fix**: Make output destinations injectable. Cobra already provides `cmd.OutOrStdout()` / `cmd.ErrOrStderr()`. Some tests already use `cmd.SetOut(buf)` / `cmd.SetErr(buf)` (see `cmd_test.go`, `cmd_integration_test.go` — 6 call sites). The pattern exists but isn't applied consistently. `PrintVersion()` and `printBuildingStatus()` bypass it entirely with raw `fmt.Printf` / `fmt.Fprintln(os.Stderr, ...)`.

### E2. Inconsistent stderr capture patterns

Three different patterns exist in the test suite for capturing stderr:

1. `testutil.CaptureStdoutStderr` (centralized, mutex-serialized, documented)
2. `cmd.executeTestCommand` → `testutil.CaptureCombinedOutput` (wrapper around #1)
3. Hand-rolled `os.Stderr = w` / `os.Stderr = oldStderr` (in `config_validation_test.go`, `run_crawl_stdin_test.go`)

Pattern #3 is the most dangerous — it doesn't use the mutex at all. Should be consolidated to pattern #1 or #2.

### E3. Doc claim was actively misleading

The old `CaptureStdoutStderr` doc said "It is safe for concurrent use (captures are serialized via a mutex)." This is technically true (captures don't clobber each other) but practically false (the race is between capture and direct readers, not between two captures). The misleading doc may have encouraged someone to add `t.Parallel()` in the first place. I fixed the doc, but the damage was already done.

### E4. Didn't measure test-suite slowdown

I removed `t.Parallel()` from 2 tests without measuring the wall-clock impact. The tests are fast (0.01s each), so the impact is negligible, but I asserted this without measurement. A before/after timing comparison would have been more rigorous.

### E5. Didn't check git history for WHEN the `t.Parallel()` was added

A `git blame` on `accept_directive_integration_test.go:22` would show when `t.Parallel()` was introduced and whether it was present from the file's creation or added later. This context would help understand whether the race was introduced recently or has been latent since the tests were written. I skipped this.

### E6. The 4 originally-named failing tests were collateral, not independently broken

The CI output listed `TestValidateMutualExclusionTypeAware`, `TestRecommendThreshold`, `TestStatsHonorsAcceptDirectives`, `TestBaselineRecordHonorsAcceptDirectives` as the 4 failures. I correctly identified that only the latter 2 were the actual racers and the first 2 were collateral (flagged because they were in the same parallel batch). But I should have stated this more explicitly in my initial response — the user might have thought all 4 needed independent fixes.

### E7. 60+ parallel tests in `cmd/` package — latent race surface

The `cmd/` package has **107 test functions**, of which **60+ call `t.Parallel()`**. Many of these touch code paths that read `os.Stdout`/`os.Stderr`. The race only fired because the capture tests happened to run at the exact same time as a reader. This is a large latent surface — any future capture-based test added with `t.Parallel()` will reintroduce the bug. The `TESTING.md` rule helps, but enforcement is manual.

### E8. `executeTestCommand` is used by non-parallel tests that don't document why

`TestStatsCommandIntegration`, `TestStatsCommandErrorCases`, `TestStatsOutputFormat`, `TestStatsOutputFile` all call `executeTestCommand` but none call `t.Parallel()`. This is correct, but the absence of `t.Parallel()` is accidental (they were written without it) rather than intentional (documented as serial). If someone "helpfully" adds `t.Parallel()` to speed them up, the race returns.

---

## F) Up to 50 Things to Get Done Next

### Architecture / Root Cause (high impact)

1. **Inject output writers into `PrintVersion()`** — accept `io.Writer` or use `cmd.OutOrStdout()` instead of `fmt.Printf`
2. **Inject output writers into `printBuildingStatus()`** — accept `io.Writer` instead of `fmt.Fprintln(os.Stderr, ...)`
3. **Audit all 23+ `fmt.Fprintf(os.Stderr, ...)` sites in `cmd/`** and route through `cmd.ErrOrStderr()` or an injected writer
4. **Consider a centralized `OutputConfig` struct** — `{ Stdout, Stderr io.Writer }` threaded through command handlers, defaulting to `os.Stdout`/`os.Stderr`
5. **Eliminate `CaptureStdoutStderr` entirely** — once writers are injectable, tests pass buffers directly (like `cmd_test.go` already does with `cmd.SetOut(buf)`)
6. **Migrate hand-rolled `os.Stderr = w` captures** in `config_validation_test.go` and `run_crawl_stdin_test.go` to the centralized helper (or to injected writers)
7. **Add a lint rule or test guard** that fails if any `*_test.go` file in `cmd/` both calls `t.Parallel()` and touches `os.Stdout`/`os.Stderr`

### Test Hardening (medium impact)

8. **Add `// Serial: see TESTING.md` comment convention** to all capture-based tests, so the serial requirement is visible at the test site, not just in the helper
9. **Run `git blame` on `accept_directive_integration_test.go:22`** to understand when the race was introduced
10. **Add a static analysis test** that scans for `t.Parallel()` + `CaptureStdoutStderr`/`os.Stdout=` in the same test function and fails
11. **Consider `t.Setenv`-like pattern for stdout/stderr** — there's no Go stdlib equivalent, but a `testutil.WithStdout(t, buf)` helper that auto-restores via `t.Cleanup` would be safer than manual save/restore
12. **Verify BDD `bdd/execute.go` capture safety** under Ginkgo's parallel node execution (`ginkgo -p`)
13. **Run `nix flake check` to confirm the exact CI path passes** (not just `go test -race`)
14. **Add a `go test -race -count=10 ./cmd/` to CI** to catch intermittent races that survive single runs
15. **Consider `-race` in pre-commit hooks** for `cmd/` package specifically

### Test Quality (medium impact)

16. **Add timing comparison** — measure cmd test suite before/after `t.Parallel()` removal to quantify the cost
17. **Review whether any of the 60+ parallel tests in `cmd/` have hidden global-state dependencies** beyond stdout/stderr
18. **Add integration test that runs `art-dupl version` via `executeTestCommand`** to cover the `fmt.Printf` path that was racing
19. **Document the `cobra.Command.SetOut/SetErr` pattern** in TESTING.md as the preferred alternative to capture
20. **Consolidate `executeTestCommand` and `newVersionCmdWithFlags`** — they solve the same problem (run a command, capture output) with different approaches

### Code Quality (lower priority but good hygiene)

21. **Check if `.golangci.yml` modification in working tree is related** — it shows as `M .golangci.yml` in git status, unexamined
22. **Check if `go.mod`/`go.sum`/`flake.lock`/`flake.nix` changes are related** — they're in the working tree, unexamined
23. **Review `internal/testutil/capture.go` `CopyToBuffer` helper** — ensure it handles large outputs without deadlock (pipe buffer fills before fn returns)
24. **Consider buffered pipe or larger pipe buffer** for capture — current `os.Pipe()` has a 64KB limit on Linux; large output can deadlock if the reader goroutine is slow
25. **Add `CaptureStdout` and `CaptureStderr` single-stream variants** — most tests only need one, and capturing both is unnecessary overhead

### Documentation

26. **Add an ADR for the output-writer injection architecture** — records the decision and the race condition that motivated it
27. **Update AGENTS.md** with a note about the stdout/stderr global-state testing constraint
28. **Add a "Testing Anti-Patterns" section to TESTING.md** — `t.Parallel()` + global mutation as example #1
29. **Cross-reference `capture.go` docs with TESTING.md** — bidirectional links so developers find the rule from either entry point

### Broader Test Suite Health

30. **Run `go test -race ./...` with `-count=5`** across ALL packages to find other latent races
31. **Check `job/` package for similar patterns** — it has parallel goroutines and channels; verify no global-state races
32. **Check `detection/` package** — MultiDetector dispatches to workers; verify thread safety
33. **Check `cache/` package** — file-based caching with LRU eviction; verify no race on cache map
34. **Review `suffixtree/` for goroutine safety** — it's the core algorithm; if it has shared state, it needs verification
35. **Add `-race` to the Nix `test-race` check with higher count** — `count=3` minimum to catch intermittent races

### Pre-existing Issues Noticed (not caused by this session)

36. **godox finding in `suppression_config_test.go:10`** — pre-existing TODO/BUG comment flagged by linter, not my change
37. **gopls `stdversion` warnings** — 9 warnings about `json.Marshal`/`json.Unmarshal`/`jsontext` requiring go1.27 while files are go1.26. This is the `GOEXPERIMENT=jsonv2` migration — the warnings may be expected or may indicate a version mismatch
38. **`go.mod`/`go.sum`/`flake.nix`/`flake.lock` uncommitted changes** — pre-existing in working tree, purpose unknown
39. **`docs/feedback/new/2026-07-29_*.md` modified** — pre-existing, unexamined
40. **`.golangci.yml` modified** — pre-existing, unexamined

### Future-Proofing

41. **Consider Go 1.24+ `testing.T` enhancements** — any new stdlib testing features that help with global state isolation?
42. **Evaluate whether `os.Stdout`/`os.Stderr` should be wrapped in a package-level var** in `cmd/` (e.g., `var stdout io.Writer = os.Stdout`) to make injection trivial
43. **Consider a `cmd.Output` interface** — `{ PrintVersion(io.Writer), PrintStatus(io.Writer, ...), ... }` — fully mockable
44. **Review Fang/Cobra integration** — does Fang set `cmd.SetOut/SetErr` automatically? If so, leverage it
45. **Add a CI check that runs `go test -race` on the exact Nix build** (not just local `go test`) to catch environment-specific races
46. **Consider test-tags** — `//go:build integration` for tests that need capture/serial execution, separating them from fast parallel unit tests
47. **Profile the test suite** — identify the slowest 5 tests; if capture-based serial tests are among them, prioritize writer injection
48. **Review whether `executeTestCommand` should use `cobra.Command.SetArgs` + `SetOut/SetErr` instead of global capture** — `cmd_integration_test.go` already does this pattern
49. **Add a CONTRIBUTING.md note** about the serial-test rule for new contributors
50. **Consider renaming `CaptureStdoutStderr` to `CaptureStdoutStderr_Serial`** or `UnsafeCaptureGlobalStdoutStderr` to make the danger obvious at call sites

---

## G) Questions (3)

### Q1. Should I do the architectural fix now (inject output writers into production code)?

The band-aid (removing `t.Parallel()`) eliminates the race today. The architectural fix (injectable writers) eliminates the entire CLASS of races permanently and would allow re-enabling `t.Parallel()` on all capture-based tests. It's a larger change (~25 call sites in `cmd/`) but high-value. Should I proceed with it, or is the band-aid sufficient for now?

### Q2. Are the uncommitted `go.mod`/`go.sum`/`flake.nix`/`flake.lock`/`.golangci.yml` changes related to this race issue?

The working tree has modifications to these files from before this session. I didn't examine them. If they're part of a dependency upgrade or lint config change, they might be relevant to the testing setup. Should I investigate, or are they unrelated work-in-progress?

### Q3. Should I run `nix flake check` to confirm the exact CI path?

I verified via `go test -race ./...` (which is what Nix invokes), but I didn't run the actual `nix flake check` command that originally failed. It takes longer and requires a Nix build. Should I close the loop on the exact failing command, or is the `go test` verification sufficient?

---

## Summary

| Metric                | Value                                                                   |
| --------------------- | ----------------------------------------------------------------------- |
| Races fixed           | 2 (4 named failures + ~20 collateral)                                   |
| Files changed         | 4 (+53/-6 lines)                                                        |
| Tests passing         | 26/26 packages with `-race`                                             |
| Root cause addressed  | No — symptom patched, architectural debt remains                        |
| Risk of regression    | Medium — any future `t.Parallel()` + capture test reintroduces the race |
| Recommended next step | Inject output writers into production code (items #1-7)                 |
