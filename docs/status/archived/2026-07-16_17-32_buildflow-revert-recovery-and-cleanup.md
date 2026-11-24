# Session Status: Recovery from BuildFlow Auto-Revert + Commit History Mess

> **Live remainder:** archived 2026-11-14 (cadence wave); every item in this report is struck or routed, and open work is owned by TODO_LIST.md and ROADMAP.md.

**Date:** 2026-07-16 17:32
**Branch:** fork (HEAD: `5cd5b9b9`)
**Trigger:** Follow-up to previous session — BuildFlow's `go-auto-upgrade` tool reverted all `.go` modifications during pre-commit, then auto-committed only the new untracked files, leaving production code changes uncommitted.

---

## What Happened (Timeline)

### Session 1 (17:00-17:16)

- Implemented all 8 TODO items: workers routing fix, progress output, 7 new tests, stdin refactor, BDD expectation fix, docs updates
- All tests passed locally (`go test ./... -count=1`)
- Wrote status report to `docs/status/2026-07-16_17-16_*.md`

### BuildFlow Auto-Commit (17:19)

- BuildFlow pre-commit hook ran on my changes
- `go-auto-upgrade` attempted Go syntax migrations that broke compilation (`slices.Contains` signature changes, undefined `hasNonEmptyFrag`)
- `go-auto-upgrade` auto-recovered via `git restore` on 32 `.go` files — **this reverted ALL my unstaged modifications to existing files**
- BuildFlow then auto-committed only the new untracked files as `9a5fc9d6`
- Result: commit had new files (`cmd/progress.go`, `cmd/exit_codes_process_test.go`, docs) but **zero modifications to existing production code**
- Nix build failed: `cmd/run_analysis.go: undefined: progressFilesChan` — the new file was committed but the wiring was reverted

### Session 2 — Recovery (17:24)

- Discovered the revert via the BuildFlow error log
- Re-applied all 7 file modifications from memory (run_analysis.go, dump_tokens.go, run_crawl_stdin_test.go, cmd_integration_test.go, domain_test.go, bdd_runners.go, error_handling_test.go)
- Hit one syntax error on re-apply: `defer func() { _ = w.Close() }` (missing `()`) — fixed immediately
- All 24 packages passed `go test -count=1`
- Ran `git commit --amend --no-edit` to fold the missing changes into `9a5fc9d6`

### BuildFlow Second Auto-Commit (17:24)

- The amend produced `6e566b4a` with all 13 files
- BuildFlow pre-commit hook ran formatting tools (d2-fmt, nix-fmt, templ-fmt)
- BuildFlow auto-committed formatting fixes as a **new commit** `5cd5b9b9` on top of `9a5fc9d6`
- Result: **two commits with identical messages** — `9a5fc9d6` has 6 files (new + docs), `5cd5b9b9` has 7 files (the re-applied modifications)

---

## a) FULLY DONE

### Production code (across both commits, functionally complete)

1. **Workers routing fix** (`cmd/run_analysis.go:122,158`, `cmd/dump_tokens.go:56`): `> 1` → `!= 1`. `--workers 0` now correctly routes to parallel parsing.
2. **Progress output** (`cmd/progress.go`): File-count reporting to stderr every 5s + final count. Suppressed by `--quiet`, non-text output, `ARTDUPL_NO_PROGRESS=1`.
3. **Progress wiring** (`cmd/run_analysis.go:118,153`): `progressFilesChan` wraps file channel in both incremental and standard tree builders.

### Tests (all pass)

4. `TestWorkers_AutoDetection` — verifies `--workers 0` doesn't hang, produces same results as `--workers 1`
5. `TestNewProcessedCloneGroup` — constructor TokenCount computation, empty clones, Validate pass-through
6. `TestStdinFeed_TimeoutUnblocksScanner` — `context.WithTimeout` unblocks scanner
7. `TestStdinFeed_FiltersGeneratedCode` — real `gogenfilter.Filter` + `FileTypeGo` end-to-end
8. `TestStdinFeed_StderrSuppressedOnCancel` — verifies `ctx.Err()==nil` guard
   ~~9. `TestExitCodes_Process` — subprocess exit codes (0, 2) via real binary build + `exec.Command`~~ done — the env var is documented in AGENTS.md progress bullet

### Refactors

10. `RunArtDuplWithStdin` — now exercises real stdin via `os.Stdin` pipe replacement + `--files` flag
11. BDD `error_handling_test.go` expectation updated for real stdin behavior
12. Removed dead `splitLines` helper

### Docs

13. `TODO_LIST.md` — 8 items marked `[x]`
    ~~14. `AGENTS.md` — 3 new convention entries~~ not-do — no PGO run landed
14. `docs/status/2026-07-16_17-16_*.md` — previous session status report

### Verification

- `go build ./...` passes
- `go test ./... -count=1` — all 24 packages pass, zero failures
- `go test -race ./... -count=1` — all 24 packages pass (verified in session 1)
- BuildFlow pre-commit hook: 26/26 tools pass (warnings only, pre-existing)
- Working tree clean

---

## b) PARTIALLY DONE

### Progress output

- Not wired into `executeHashOnlyAnalysis` (`cmd/run_hash.go`)
- No unit tests for `progressFilesChan` itself (suppression, forwarding, count accuracy)
- 5s interval is hardcoded, not configurable

---

## c) NOT STARTED

- Hybrid slice/map transition storage for suffix tree (explicitly deferred)
- Unit tests for `cmd/progress.go`
- Exit code 130 (interrupted) subprocess test
- `dumpTokens` workers test (third code path changed, no direct test)

---

## d) TOTALLY FUCKED UP

### Git History — Two Duplicate Commits

```
5cd5b9b9 test: close 8 testing TODOs and fix --workers 0 routing bug  ← HEAD (7 files)
9a5fc9d6 test: close 8 testing TODOs and fix --workers 0 routing bug  ← parent (6 files)
```

**What went wrong:**

1. The `go-auto-upgrade` BuildFlow tool ran `git restore` on 32 `.go` files as auto-recovery for its own broken migrations. This is a nuclear option — it destroyed my unstaged work to save itself.

2. I should have verified the commit contents BEFORE walking away. Instead, I wrote a status report, and the user caught the Nix build failure from the BuildFlow output.

3. The `git commit --amend` should have squashed everything into one commit. But BuildFlow's pre-commit hook ran formatting tools and auto-committed as a new commit on top, creating a duplicate-message split.

**Impact:** Functionally correct (all 13 files present across both commits, tests pass), but the git history is sloppy. A `git rebase -i HEAD~2` to squash would clean it up, but I was not asked to do that and should not touch git history without explicit instruction.

### Root Cause: No Staging Before BuildFlow

I left all modifications unstaged. The BuildFlow auto-commit only captured new (untracked) files. Had I `git add` all changes first, the auto-commit would have captured everything and the revert would have been visible immediately.

---

## e) WHAT WE SHOULD IMPROVE

### 1. Always stage before BuildFlow runs

The BuildFlow auto-commit captures staged + untracked files but doesn't protect unstaged modifications. `go-auto-upgrade`'s `git restore` nuclear option destroys unstaged work. **Lesson: `git add` everything before committing.**

### 2. Verify commits contain what you think they contain

After `git commit --amend`, I checked the commit hash but didn't verify the file list. I should have run `git diff-tree --name-only -r HEAD` to confirm all 13 files were present.

### 3. The amend was swallowed by BuildFlow

My `git commit --amend --no-edit` produced `6e566b4a` (correct, 13 files). But BuildFlow's pre-commit hook then created `5cd5b9b9` on top, splitting the work. This is a BuildFlow design issue — its auto-commit doesn't respect `--amend`.

### 4. No progress unit tests

I built `cmd/progress.go` with 73 lines of channel-wrapping logic and zero dedicated unit tests. The only coverage is incidental via integration tests that happen to exercise the code path. A proper test suite would verify:

- All paths forwarded faithfully (no drops, no reordering)
- Suppression when `Quiet`, non-text format, `ARTDUPL_NO_PROGRESS=1`
- Correct count in periodic and final output
- Context cancellation propagates

### 5. `os.Stdin` global mutation remains

`RunArtDuplWithStdin` replaces `os.Stdin` process-globally. No parallel test race detected yet, but it's a latent footgun.

---

## f) Up to 50 Things We Should Get Done Next

1. Squash the two duplicate commits (`git rebase -i HEAD~2`)
2. Write unit tests for `cmd/progress.go` (forwarding, suppression, counts)
3. Wire progress into `executeHashOnlyAnalysis` (`cmd/run_hash.go`)
4. Add `--progress-interval` flag (or keep hardcoded 5s)
5. Test exit code 130 via SIGINT subprocess
6. Test `dumpTokens --workers 0`
7. Add stdin test with `FileTypeTempl`
8. Replace `os.Stdin` mutation with reader injection through call chain
   ~~9. Add `ARTDUPL_NO_PROGRESS` to `HOW_TO_USE.md`~~ done — the env var is documented in AGENTS.md progress bullet
9. Add progress to `FEATURES.md`
10. Update `CHANGELOG.md` with workers routing fix
11. Add `--no-progress` CLI flag as alternative to env var
12. Fix the 7 `exhaustruct` warnings in `cmd/run_analysis.go`
    ~~14. Profile-guided optimization run~~ not-do — no PGO run landed
13. Memory usage test on 10000+ files
14. `--help` text audit
15. YAML config file support
16. Shell completion e2e test
17. SDK: expose `ExitCodeForError` in `pkg/artdupl`
18. Deprecation warning for `--semantic`
    ~~21. Cross-link `docs/ACTIONABILITY_PATTERNS.md`~~ done — ACTIONABILITY_PATTERNS.md is referenced from AGENTS.md
19. Consider per-file progress in verbose mode
20. Consider lines-count in progress (currently file count only)
    ~~24. Test progress output goes to stderr not stdout~~ done — progress writes to the injectable stderr writer (AGENTS.md output writer injection)
21. Test `ARTDUPL_NO_PROGRESS=1` in subprocess
    ~~26. Benchmark `--workers 0` vs `--workers 1` on large file sets~~ not-do — the workers benchmark never landed; alloc-gate owns perf
22. Config file loading e2e test with `--config`
    ~~28. Baseline `check` subcommand CI scenario test~~ done — the check subcommand ships with pinned exit codes (AGENTS.md)
23. SARIF output schema validation
24. Consider `--progress-interval 0` to disable progress at runtime
25. Move `shouldShowProgress` logic into `config.Config`
26. Extract progress reporting into an interface for testability
27. Document testing conventions (subprocess vs in-process)
28. Document the stdin test matrix in `TESTING.md`
29. Consider a `Spinner` abstraction for richer UX
30. Unify the three parallel/sequential gates into single dispatch
31. Evaluate `dump_tokens.go` sharing progress wrapper
32. Consider streaming file count to `ParseStats`
33. Add context cancellation to hash-only file collection
34. Audit all `fmt.Fprintf(os.Stderr, ...)` calls
35. Document workers routing convention in an ADR
36. ~~Add workers bug to CHANGELOG as behavioral fix~~ done — CHANGELOG --workers 0 routing bug entry verified by title (2026-11-23 docs-health pass)
37. Create testing conventions doc
38. Test stdin with mixed `.go` and `.templ` paths
39. Test `--workers -1` (negative auto-detect)
40. Consider `--workers auto` as explicit alias for `0`
    ~~47. Test progress suppression with JSON output format~~ done — progress is suppressed for non-text output by design (AGENTS.md)
    ~~48. Add progress to stats subcommand~~ not-do — stats has no progress output
41. Consider progress for baseline recording/checking
42. Document BuildFlow's `git restore` behavior in AGENTS.md as a hazard

---

## g) Top 2 Questions I Cannot Answer Myself

### 1. Should I squash the two duplicate commits?

The git history has `9a5fc9d6` and `5cd5b9b9` with identical messages. A `git rebase -i HEAD~2` + squash would produce a clean single commit. But this rewrites history on the `fork` branch, which may have implications I can't assess (CI, remote tracking, other worktrees). **Should I clean this up, or leave it as-is since the content is functionally correct?**

### 2. Is the `go-auto-upgrade` `git restore` behavior a known BuildFlow risk I should document?

The `go-auto-upgrade` tool broke compilation with bad migrations (changing `slices.Contains` signatures, removing undefined functions), then auto-recovered by running `git restore` on 32 files — destroying my unstaged work. This is extremely dangerous behavior for a pre-commit hook. **Should this be documented as a known hazard in AGENTS.md, or is this a BuildFlow bug that should be reported upstream?**
