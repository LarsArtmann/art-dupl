# Status Report: Code Hygiene Sprint — Self-Review

**Date:** 2026-07-25 00:30
**Session:** Executing all 11 Code Hygiene items from TODO_LIST.md
**Branch:** fork

> **Resolution (2026-07-25):** The 2 partial items (Task 10: `examples_sdk_demo.go`
> threshold, Task 11: `SourceBreakdown` dead code) were resolved in the gap-closure
> session (`2026-07-25_02-49`). `SourceBreakdown()` is now wired into stats output via
> `SetFilterSourceStats`. All `Threshold: 15` instances replaced with `DefaultThreshold`.
> The remaining open item (`SetFilterSourceStats` unit test) is tracked in TODO_LIST.

---

## A) FULLY DONE (9/11 tasks)

| #   | Task                                                       | What was done                                                                                                                                                                                                                                                                             |
| --- | ---------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | `bdd/type_aware_test.go:14` undefined `CreateBDDTestSetup` | Stale LSP error — package builds clean. Verified with `go build ./bdd/...`. No code change needed.                                                                                                                                                                                        |
| 2   | `cmd/progress_test.go` lint warnings (errcheck, wsl_v5)    | Resolved by root-cause fix: re-removing `exhaustruct`/`tagliatelle` from `.golangci.yml`.                                                                                                                                                                                                 |
| 3   | `cmd/accept_directive.go:97` mnd magic number 64           | Resolved by root-cause fix (same as #2). The `64` was in a named constant `scannerInitBufSize`, never an issue.                                                                                                                                                                           |
| 4   | `cmd/accept_directive_test.go:40` gci formatting           | Resolved by root-cause fix (same as #2).                                                                                                                                                                                                                                                  |
| 5   | Remove 10 dead `//nolint:exhaustruct` directives           | Removed all 10 across 7 files: `pkg/artdupl/types.go`, `pkg/logger/logger.go`, `job/profiler.go`, `job/incremental.go` (2), `internal/utils/file.go`, `internal/testutil/bdd.go` (2), `errors/types.go` (2). Explanatory comments preserved (just removed `//nolint:exhaustruct` prefix). |
| 6   | SDK DefaultOptions threshold split-brain (15 vs 5)         | Added `DefaultThreshold = 5` constant in `pkg/artdupl/types.go` mirroring `config.DefaultThreshold`. Updated `DefaultOptions()` to use it. Updated `TestDefaultOptions_Values` assertion. Updated `SDK_DESIGN.md` documentation.                                                          |
| 7   | SDK TypeAware fallback test asserts nothing                | Replaced `_ = result; _ = err` with real assertions: error must be nil or `ErrNoDuplicatesFound`, result must be non-nil when err is nil.                                                                                                                                                 |
| 8   | Progress test parallelism (`os.Stderr` manipulation)       | Refactored `progressFilesChan` to accept `io.Writer` parameter. Production callers pass `os.Stderr`. Tests pass `io.Discard`. No more global `os.Stderr` swap. Test now calls `t.Parallel()`.                                                                                             |
| 9   | Annotate stale status report as SUPERSEDED                 | Added SUPERSEDED blockquote at top of `docs/status/2026-07-24_23-11_full-todo-execution-sprint.md` explaining M26-M30 stub flags were removed.                                                                                                                                            |

### Root-Cause Fix (unblocked tasks 1-4)

**`.golangci.yml`**: Commit `cbb329a7` (auto-committer) re-added `exhaustruct` and `tagliatelle` to the enable list, plus an `exhaustruct.exclude` settings block. This is exactly what `scripts/check-disabled-linters.sh` was designed to prevent. Removed all three additions. CI guard now passes (`OK: no disabled linters`).

---

## B) PARTIALLY DONE (1/11 tasks)

### Task 11: `FilterResult.Source` field — INFRASTRUCTURE BUILT, OUTPUT NOT WIRED

**What was done:**

- Added `FilterSource` type with two values: `FilterSourceGogenfilter` and `FilterSourceDefenseInDepth` in `cmd/filter_stats.go`
- Added `bySource` map to `FilterStats` struct
- Added `RecordWithSource(result, source)` method
- `Record()` now delegates to `RecordWithSource(result, FilterSourceGogenfilter)` as default
- Defense-in-depth path in `cmd/util.go:shouldIncludeFile` calls `RecordWithSource(result, FilterSourceDefenseInDepth)`
- Added `SourceBreakdown()` method returning `map[string]int`
- 2 unit tests: `TestFilterStatsSourceBreakdown` and `TestShouldIncludeFile_TracksDefenseInDepthSource`

**What was NOT done:**

- `SourceBreakdown()` is **dead code** — nothing calls it
- `applyFilterStats` in `cmd/stats.go:38-63` only passes `totalFiltered` and per-reason `breakdown` to `sp.SetFilterStats()`
- The `printer.StatsPrinter` interface (`SetFilterStats(totalFiltered int, breakdown map[string]int)`) has no parameter for source breakdown
- **The TODO said "Distinguish gogenfilter vs defense-in-depth catches in stats output"** — the tracking infrastructure exists but it does NOT reach the stats output
- To finish: need to extend `SetFilterStats` signature (or add a new method) and wire `SourceBreakdown()` into `applyFilterStats`

---

## C) NOT STARTED

N/A — all 11 items were addressed (10 fully, 1 partially).

---

## D) TOTALLY FUCKED UP

### D1: Left stale `Threshold: 15` in doc.go

`pkg/artdupl/doc.go:31` still shows:

```go
//	opts := &artdupl.Options{
//	    Threshold:         15,                       // Minimum token count for a clone
```

I fixed `SDK_DESIGN.md` but **missed the package-level godoc example**. This is actively misleading — a user reading the godoc sees threshold 15 as an example when the actual default is 5.

### D2: Left stale `Threshold: 15` in 8 test fixtures

`pkg/artdupl/detector_validation_test.go` has 8 occurrences of `Threshold: 15` as arbitrary valid values in test cases. These aren't testing `DefaultOptions` — they're just valid thresholds for validation tests. Not strictly wrong (15 is valid), but inconsistent with the new default of 5. Should use `DefaultThreshold` constant or `5` for consistency.

### D3: Did not update CHANGELOG.md

AGENTS.md says "Completed work is in CHANGELOG.md." I updated `TODO_LIST.md`, `SDK_DESIGN.md`, and `AGENTS.md` but **never touched CHANGELOG.md**.

### D4: Did not run race detector

The progress test parallelism fix (Task 8) specifically addresses a concurrency concern, but I never ran `go test -race`. The Nix devShell requires `CGO_ENABLED=1` for race detection, which I didn't set. `go test -race` failed with `go: -race requires cgo`.

### D5: Did not run `nix flake check`

AGENTS.md says `nix flake check` is the reproducible CI (includes `templ generate` in preBuild). I ran `go build`, `go test`, and `golangci-lint`, but never the Nix check.

---

## E) WHAT WE SHOULD IMPROVE

### Process Improvements

1. **The auto-committer is a repeat offender** — `cbb329a7` re-added disabled linters for the Nth time. The CI guard script exists but didn't prevent the commit. The guard only runs in Nix CI, not as a git hook. Consider a pre-commit hook.

2. **No sync-test between SDK and CLI constants** — `pkg/artdupl.DefaultThreshold` duplicates `config.DefaultThreshold` by architectural necessity. There is no test asserting they're equal. A drift test (`if DefaultThreshold != config.DefaultThreshold { t.Error(...) }`) would catch future drift, but it can't live in `pkg/artdupl` (can't import `config/`). It could live in a top-level integration test package.

3. **`SourceBreakdown()` is dead code** — I added a public method that nothing consumes. This is the definition of YAGNI until it's wired to output. Either finish the wiring or don't merge the infrastructure.

4. **Documentation drift is pervasive** — `doc.go`, `SDK_DESIGN.md`, and test fixtures all had stale threshold 15. The codebase has no single source of truth that documentation references. Every constant that's duplicated (even intentionally) will drift.

5. **Comment preservation after nolint removal** — When removing `//nolint:exhaustruct`, I preserved the explanatory comment (e.g., `// zero-value counters, mutated incrementally`). Some of these comments now state the obvious. Minor, but worth a cleanup pass.

---

## F) Up to 50 Things to Get Done Next

### Immediate Fixes (from this session's gaps)

1. **Fix `pkg/artdupl/doc.go:31`**: Change `Threshold: 15` to `Threshold: 5` (or better: `Threshold: artdupl.DefaultThreshold`)
2. **Update `detector_validation_test.go`**: Replace 8 hardcoded `Threshold: 15` with `DefaultThreshold` constant
3. **Wire `SourceBreakdown()` into stats output**: Extend `SetFilterStats` or add `SetFilterSourceStats`, update `applyFilterStats` in `cmd/stats.go`, update printer stats data struct
4. **Add CHANGELOG.md entry** for all 11 completed Code Hygiene items
5. **Run `CGO_ENABLED=1 go test -race ./...`**: Verify the progress parallelism fix and all other changes under the race detector
6. **Run `nix flake check`**: Full reproducible CI verification

### From TODO_LIST.md (still open)

7. **Split `printer/` into sub-packages**: ~29 files / ~3500+ lines, blocked by circular dep on `StatsPrinter`
8. **Push defense-in-depth into gogenfilter**: Upstream PR against `github.com/LarsArtmann/gogenfilter`
9. **Refactor `generatorIncludes` struct**: 6 boolean fields with shotgun surgery, use `map[string]struct{}`
10. **Unify `allowsContent` and `filterExcludedGenerated`**: Both switch on same content markers
11. **Lazy content reading in `shouldIncludeFile`**: Read only when filename check doesn't match
12. **Use `bytes.Contains` instead of `string(content)`**: Avoid heap allocation
13. **Early-exit optimization**: Skip marker checks if no "Code generated" in content
14. **YAML config file support** (`.artdupl.yml`)
15. **`--diff-report <baseline>` mode**: Show new/suppressed/resolved clones
16. **`--explain` flag**: Explain WHY a clone was reported
17. **HTML report improvements**: File output, TTY auto-detect, stable IDs
18. **`--recommend-threshold`**: Auto-suggest based on codebase size
19. **Templ Phase 3**: Expression normalization for `{ id.String() }` vs `{ groupID.String() }`
20. **Interface-method-aware suppression**: At all thresholds, not just current pattern

### Architecture & Quality

21. **Add pre-commit hook for `check-disabled-linters.sh`**: Prevent auto-committer from re-adding disabled linters
22. **Add drift-detection test for SDK/CLI constant sync**: Assert `DefaultThreshold` values match via integration test
23. **Add `templ generate` to pre-commit or CI**: Ensure generated code is fresh
24. **Audit all godoc examples for stale values**: Threshold, workers, timeout — any hardcoded example value
25. **Consolidate filter marker constants**: `sqlcMarker`, `templMarker`, `protobufMarker` in `cmd/util.go` duplicate gogenfilter's internal markers — single source of truth
26. **Add property-based test for `shouldIncludeFile`**: Verify no generated file slips through regardless of filename suffix
27. **Profile `filterExcludedGenerated`**: `string(content)` on large files is expensive — benchmark and optimize
28. **Add `--filter-stats` flag**: Surface `SourceBreakdown()` in CLI output (finishes Task 11)
29. **Review all `//nolint` directives**: Verify each is still needed after exhaustruct cleanup
30. **Add integration test for `FilterStats` end-to-end**: Verify source tracking through full CLI run, not just unit test
31. **Document the filter pipeline in a diagram**: gogenfilter vs defense-in-depth vs gitignore vs filter patterns — 4 overlapping filter layers
32. **Consolidate BDD test setup boilerplate**: 20+ test files each call `CreateBDDTestSetup()` in BeforeEach — extract shared Describe wrapper
33. **Add `go vet ./...` to CI**: Currently only `golangci-lint` runs, `go vet` catches different issues
34. **Review `progressFilesChan` goroutine lifecycle**: Ensure no goroutine leak on context cancellation (the `io.Writer` change is safe but should be verified under -race)
35. **Add SDK example test**: `pkg/artdupl/example_test.go` with runnable example using `DefaultThreshold`
36. **Update `HOW_TO_USE.md`**: Document `--include-generated generic` defense-in-depth behavior (the BDD test covers it but user docs don't explain it)
37. **Add benchmark for `progressFilesChan`**: Verify the io.Writer abstraction doesn't add overhead
38. **Review `FilterStats` thread safety**: `bySource` map is protected by mutex, but verify no path records without lock
39. **Add fuzz test for `filterExcludedGenerated`**: Edge cases in content markers (partial markers, multiple markers, malformed)
40. **Clean up status report annotations**: Verify all SUPERSEDED reports are properly cross-referenced
41. **Review `DefaultThreshold` naming**: Should it be `DefaultThreshold` or `DefaultCloneThreshold` for clarity in the SDK context?
42. **Add `FilterSource` to JSON stats output**: If/when `SourceBreakdown` is wired, add `filter_source_breakdown` to `StatsData`
43. **Consolidate test helpers**: `newTestFilter` in `cmd/filter_includes_test.go` duplicates setup logic — extract to testutil
44. **Review BDD test for templ defense-in-depth**: The new test uses `parseHeader`-style code — ensure it actually exceeds threshold 5 (statement count)
45. **Add contract test between gogenfilter and art-dupl**: Verify filter behavior matches when gogenfilter is updated
46. **Review `t.Context()` usage**: Go 1.24+ feature, verify all test files use it consistently
47. **Audit `io.Discard` vs `io.MultiWriter`**: Progress tests discard output — consider capturing for debug when test fails
48. **Add `--filter-source-breakdown` to stats subcommand**: Surface the new source distinction in CLI stats output
49. **Review error handling in `shouldIncludeFile`**: `os.ReadFile` error returns `true` (include file) — is this the right default?
50. **Add architectural decision record for FilterSource**: Document why defense-in-depth tracking was added and how it relates to the gogenfilter upstream push

---

## G) Questions I Cannot Answer Myself

### Q1: Should `SourceBreakdown()` be wired into the existing `SetFilterStats(totalFiltered int, breakdown map[string]int)` interface, or should it be a separate method?

`SetFilterStats` is part of the `printer.StatsPrinter` interface, which is implemented by multiple printer types. Changing its signature is a breaking interface change. Options:

- **(A)** Extend the signature: `SetFilterStats(totalFiltered int, breakdown, sourceBreakdown map[string]int)` — breaks all implementations
- **(B)** Add a new method: `SetFilterSourceStats(sourceBreakdown map[string]int)` — optional, implementations can no-op
- **(C)** Merge source into the existing breakdown map with prefixed keys: `"gogenfilter:templ": 5` — hacky but zero interface change

This is a product/API decision I cannot make.

### Q2: Should the SDK `DefaultThreshold` constant have a drift-detection test, and if so, where does it live?

The SDK (`pkg/artdupl`) cannot import `config/` (arch-lint enforced). A test asserting `DefaultThreshold == config.DefaultThreshold` cannot live in `pkg/artdupl/`. It could live in:

- A top-level integration test package (e.g., `test/integration/`)
- The `bdd/` package (imports both)
- Nowhere (accept the duplication risk)

This is an architectural decision about test organization.

### Q3: Should I finish Task 11 (wire `SourceBreakdown` to output) now, or is the infrastructure sufficient until the gogenfilter upstream push happens?

The TODO_LIST has a separate item: "Push defense-in-depth into gogenfilter." If that upstream push happens, the defense-in-depth path in `cmd/util.go` goes away entirely, making `FilterSource` unnecessary. Building the output wiring now might be wasted work. This is a sequencing/prioritization decision.
