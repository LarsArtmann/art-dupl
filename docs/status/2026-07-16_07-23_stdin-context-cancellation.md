# Status: Stdin Context Cancellation Implementation

**Date:** 2026-07-16 07:23
**Session Scope:** Thread `context.Context` through the stdin `bufio.Scanner` — the last remaining non-cancellable pipeline entry point.

---

## What Was the Problem?

`bufio.Scanner.Scan()` on `os.Stdin` is a blocking syscall. If the pipe stays open but sends no data, `Scan()` blocks forever. Context cancellation was only checked _between_ lines (the `select` on channel send), never _during_ the blocking read itself. This was documented as a known limitation in AGENTS.md and TODO_LIST.md.

**Root cause:** There is no way to interrupt a blocking `read()` on a file descriptor from another goroutine — except closing the FD.

---

## A) FULLY DONE

### 1. Extracted `feedFromStdin` function (`cmd/run_crawl.go:35-89`)

- **Signature:** `feedFromStdin(ctx context.Context, rc io.ReadCloser, ...) chan string`
- **Mechanism:** A watcher goroutine selects between `ctx.Done()` and a `done` signal. On cancellation, it calls `rc.Close()`, which unblocks the scanner's underlying `read()` syscall. On normal EOF, the scanner goroutine closes `done` to release the watcher.
- **Error suppression:** `sc.Err()` errors are only logged to stderr when `ctx.Err() == nil` — forced-close errors from cancellation are silently swallowed.
- **Design win:** Accepts `io.ReadCloser` instead of hardcoding `os.Stdin`. This eliminated a data race (see section D) and made the function independently testable.

### 2. `filesFeedWithOptions` updated (`cmd/run_crawl.go:92-100`)

- `fromStdin` branch now delegates to `feedFromStdin(ctx, os.Stdin, ...)`.
- No other callers affected — verified via call chain analysis.

### 3. Three race-clean tests (`cmd/run_crawl_stdin_test.go`)

| Test                                  | What it verifies                                                                       |
| ------------------------------------- | -------------------------------------------------------------------------------------- |
| `TestStdinFeed_CancelUnblocksScanner` | Idle pipe + cancel ctx → channel closes within 2s timeout (would hang without the fix) |
| `TestStdinFeed_ReadsPaths`            | Paths piped → correct output + channel close on EOF                                    |
| `TestStdinFeed_PartialReadThenCancel` | Partial data consumed, then cancel → early paths received + prompt goroutine exit      |

- All pass with `-race` (CGO_ENABLED=1).
- Full test suite (`go test -race ./...`) — all 27 packages pass.

### 4. Documentation updated

- **AGENTS.md** (line 75): Replaced "file feeders lack ctx" note with accurate description of `feedFromStdin` cancellation mechanism.
- **TODO_LIST.md** (line 16): Marked as `[x]` complete.

### 5. Lint clean on changed files

- Zero new lint issues in `run_crawl.go` or `run_crawl_stdin_test.go`.
- Pre-existing lint warnings in other files (exhaustruct, funlen, whitespace) untouched.

---

## B) PARTIALLY DONE

### Stderr suppression test — removed, not replaced

I initially wrote a `TestStdinFeed_ErrorSuppressedOnCancel` test to verify that scanner errors from forced close don't pollute stderr. It had a bug in the `os.Stderr` restoration logic (restored to `os.Stdin` instead of `origStderr`). Rather than fixing it, I removed the entire test. The `ctx.Err() == nil` guard on line 83 is **untested**.

### `RunArtDuplWithStdin` test helper still bypasses stdin

`internal/testutil/bdd_runners.go:109-131` — this helper parses "stdin content" as newline-delimited paths and passes them as **positional arguments**. It does NOT exercise the `feedFromStdin` code path. The BDD tests that claim to test stdin (`bdd/bdd_test.go`, `bdd/plumbing_output_test.go`, `bdd/error_handling_test.go`) are testing the file-path-args path, not actual stdin piping. This was pre-existing but I noticed it and did not fix it.

---

## C) NOT STARTED

- Did not run `go build ./...` explicitly (only `go test ./...` which compiles).
- Did not run `nix flake check` (requires Nix evaluation; the user's environment may not have it ready).
- Did not run `templ generate` (no `.templ` files changed — not needed).
- Did not check whether a CHANGELOG.md entry is expected for this type of change.
- Did not add an integration test that runs the full CLI binary with `--stdin` and pipes actual file paths.
- Did not add a test for the filter/include logic path through `feedFromStdin` (only tested with `nil` filter and `""` file type).

---

## D) TOTALLY FUCKED UP (and fixed)

### Data race from `os.Stdin` global mutation — DETECTED AND FIXED

**First attempt** (never committed but ran as a test): The inline closure in `filesFeedWithOptions` called `os.Stdin.Close()` directly on the global. Tests swapped `os.Stdin` to a pipe via `t.Cleanup`. The race detector caught it:

```
Read at 0x...183e40 by goroutine 339 (filesFeedWithOptions.func1.1 — os.Stdin.Close())
Previous write at 0x...183e40 by goroutine 337 (TestStdinFeed_ReadsPaths.func1 — os.Stdin = origStdin)
```

The watcher goroutine was still alive when the test's `t.Cleanup` restored `os.Stdin`, creating a read-after-write race on the global variable.

**Fix:** Extracted `feedFromStdin(ctx, rc io.ReadCloser, ...)` — tests pass pipes directly, zero global state mutation. This was actually a **better design** than the original inline approach, forced by the race.

### Wrong io.Pipe recommendation in analysis phase

I initially recommended an `io.Pipe` intermediary as the "best" solution. On further reflection, it doesn't actually solve the problem — `io.Copy(pw, os.Stdin)` still blocks on `os.Stdin.Read()`. Closing the pipe writer only unblocks goroutines stuck on pipe writes, not pipe reads. The leaked goroutine would just move from the scanner to the copier. This was a conceptual error caught before implementation.

---

## E) WHAT WE SHOULD IMPROVE

### Code-level improvements from this session

1. **`close(done)` is skipped on the send-cancel path** (`run_crawl.go:77`). When the scanner goroutine returns from the `case <-ctx.Done()` in the send select, it exits without calling `close(done)`. This is safe because the watcher has already woken on `ctx.Done()`, but it's asymmetric — the `done` signal is only closed on the EOF/error path. A `defer close(done)` would be cleaner but requires restructuring (done is declared after the watcher starts).

2. **No test for filter integration through `feedFromStdin`** — the existing tests pass `nil` filter and `""` file type. A test with a real `gogenfilter.Filter` and `FileTypeGo` would verify the filtering pipeline works end-to-end through stdin.

3. **`RunArtDuplWithStdin` should actually pipe to stdin** — now that `feedFromStdin` accepts `io.ReadCloser`, the BDD helper could be refactored to inject a pipe reader instead of converting paths to args. This would give real stdin coverage in BDD tests.

4. **No `--stdin` flag documentation update** — the HOW_TO_USE.md or README should mention that Ctrl+C now properly cancels mid-read from stdin.

5. **The watcher goroutine leaks on the send-cancel path** — when ctx is cancelled during a channel send, the scanner goroutine returns (line 77), `defer close(fchan)` runs, but the watcher goroutine has already exited via `ctx.Done()`. No actual leak. BUT: if the consumer is also cancelled and never drains `fchan`, the `defer close(fchan)` blocks until the consumer returns. This is fine for CLI usage (process exits) but worth documenting.

### Pre-existing issues noticed (not caused by this session)

6. **`version_cmd.go` uses go1.27 APIs** (`json.Marshal`, `jsontext.WithIndentPrefix`) while `go.mod` targets go1.26. LSP warns about this. Pre-existing, not caused by this change.

7. **`job/incremental.go:51` exhaustruct warning** — `IncrementalParser` missing field `group`. Pre-existing.

8. **11 exhaustruct warnings across `cmd/`** — all pre-existing, all in code I didn't touch.

9. **`runStats` function is too long** (84 > 80 lines, funlen). Pre-existing.

---

## F) Up to 50 Things We Should Get Done Next

### Direct follow-ups from this session

1. Add test for `feedFromStdin` with active `gogenfilter.Filter` + `FileTypeGo` filtering
2. Add test verifying stderr suppression (`ctx.Err() == nil` guard on line 83)
3. Add test for `feedFromStdin` with `./` prefix stripping (line 64)
4. Refactor `RunArtDuplWithStdin` to actually exercise the stdin pipe path
5. Add `--stdin` cancellation behavior to HOW_TO_USE.md
6. Run `go build ./...` to verify clean compilation (not just tests)
7. Run `nix flake check` for full CI validation
8. Consider adding `defer close(done)` via restructuring to make cleanup symmetric
9. Add an integration test: pipe real Go file paths via stdin to the CLI binary, verify clone detection works
10. Verify `feedFromStdin` handles `\r\n` line endings (bufio.Scanner default split is `ScanLines` which handles this)

### Architecture / code quality

11. Fix the 11 pre-existing `exhaustruct` warnings in `cmd/`
12. Fix `runStats` funlen violation (split into smaller functions)
13. Fix `version_cmd.go` go1.27/go1.26 version mismatch
14. Fix `job/incremental.go:51` exhaustruct (missing `group` field)
15. Fix `version_cmd.go:31` whitespace warning (unnecessary leading newline)
16. Audit all channel sends in the codebase for the check-then-send anti-pattern
17. Add `signal.NotifyContext` for explicit SIGTERM handling (currently delegated to fang)
18. Consider a `WithContext(ctx)` option on the SDK `Detector` interface

### Testing gaps

19. Add stdin integration test in BDD suite (actual pipe, not args)
20. Add test for stdin with empty input (just EOF, no paths)
21. Add test for stdin with whitespace-only lines
22. Add test for stdin with very long paths (exceeding `bufio.MaxScanTokenSize`)
23. Add `-race` to CI pipeline if not already present
24. Add timeout test: `context.WithTimeout` expiry while reading stdin
25. Add test for concurrent stdin reads (should not happen but verify it panics or behaves)

### Documentation

26. Update FEATURES.md if stdin support is listed
27. Add ADR for the close-on-cancel pattern (architecture decision record)
28. Document the `io.ReadCloser` injection pattern in AGENTS.md for future testability
29. Add `feedFromStdin` to any API/SDK documentation if relevant
30. Update docs/DOMAIN_LANGUAGE.md if stdin terminology is domain-relevant

### Performance / robustness

31. Benchmark `feedFromStdin` vs the old inline approach (should be negligible)
32. Consider buffered channel for `fchan` to reduce goroutine wakeups
33. Consider `bufio.Scanner` buffer size increase for long paths (`os.Getwd()` can be long)
34. Add graceful drain: when ctx is cancelled, log how many paths were already processed
35. Consider whether `shouldIncludeFile` errors should be retried or always fail-open

### Broader project health

36. Audit all `os.Stdin`/`os.Stdout`/`os.Stderr` global usage for testability
37. Check if `crawlDirectoryWithOpts` has the same testability gap (it takes `CrawlOptions` — verify it's injectable)
38. Review whether `filesFeedWithOptions` should be split into smaller functions (it's 93+ lines)
39. Consider extracting file filtering into its own `FileFilter` type
40. Check if the `FilterStats` type needs thread-safety verification
41. Review the `generatorIncludes` type — should it be in `config/` instead of `cmd/`?
42. Add fuzzing for `feedFromStdin` with random input
43. Consider whether stdin should support null-delimited input (for `find -print0`)
44. Review if `filepath.Walk` → `filepath.WalkDir` migration would improve perf
45. Check if the suffix tree benefits from `context.Context` for cancellation mid-build
46. Audit `sendCtx[T]` usage — are all sends going through it?
47. Review whether `BuildTree`'s `done` channel pattern could be applied here
48. Consider a `Reader` interface for the entire file-feed pipeline (stdin, Walk, etc.)
49. Add property-based testing for filter + stdin combinations
50. Review if the cache layer needs context cancellation support

---

## G) Top 2 Questions I Cannot Answer Myself

### 1. Should `os.Stdin.Close()` be called in production, or only on test pipes?

In production, `filesFeedWithOptions` passes `os.Stdin` directly to `feedFromStdin`. On cancellation, we now call `os.Stdin.Close()`. This is technically mutating global process state. In a CLI that exits immediately after cancellation (exit code 130), this is harmless. But if art-dupl is ever embedded as a library (via `pkg/artdupl` SDK) where the caller manages the process lifecycle, closing their stdin FD would be hostile.

**Question:** Is there a current or planned use case where art-dupl runs as a long-lived process or library where closing stdin would be problematic? Should `feedFromStdin` skip the close-on-cancel when the reader is `os.Stdin`?

### 2. Should `RunArtDuplWithStdin` be fixed now, or is it a separate task?

The BDD test helper `RunArtDuplWithStdin` claims to test stdin but actually converts paths to CLI arguments. This means the `feedFromStdin` code path has zero integration/E2E test coverage — only unit tests. Fixing it requires either injecting a pipe reader into the `filesFeedWithOptions` call chain (which currently hardcodes `os.Stdin`) or using `os.Stdin` replacement in the test process.

**Question:** Should I refactor `filesFeedWithOptions` to accept an injectable reader (breaking the current signature), or should I add a separate injection point (e.g., a package-level `var stdinReader io.ReadCloser = os.Stdin` that tests can swap)? The former is cleaner but touches more callers; the latter is pragmatic but adds mutable global state.

---

## Session Metrics

| Metric                 | Value                                                                            |
| ---------------------- | -------------------------------------------------------------------------------- |
| Files changed          | 4 (`run_crawl.go`, `run_crawl_stdin_test.go` [new], `AGENTS.md`, `TODO_LIST.md`) |
| Lines added            | ~120 (implementation + tests + docs)                                             |
| Lines removed          | ~25 (old inline stdin reader)                                                    |
| Tests added            | 3                                                                                |
| Tests passing          | 3/3 (+ full suite 27/27 packages)                                                |
| Race detector          | Clean                                                                            |
| Lint                   | Clean on changed files                                                           |
| Time to implement      | ~45 min (including wrong approach + data race fix)                               |
| Wrong approaches tried | 2 (io.Pipe recommendation, inline os.Stdin.Close)                                |
