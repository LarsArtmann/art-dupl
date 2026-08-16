# Status Report — Wave 2: Flake Green, Perf Gate Extended, Cache Stats Shipped

**Date**: 2026-08-16 12:27 · **Branch**: `fork` · **Task scope**: `docs/planning/2026-08-16_04-27_measure-first-trust-and-signal-master-plan.md` (T1–T26)

---

## a) Fully Done (this session)

### T5.1b — `nix flake check` GREEN (the blocker, resolved)

- Staged all previously-untracked files (`actionability_preamble*.go`, `scripts/alloc-budgets.txt`) plus every modified file.
- First check run **failed on a real bug** (not the untracked-file issue): the `self-test` check found duplication in art-dupl's own source at threshold 1 — `cmd/timing.go:26-27` vs `job/profiler.go:79-80`, the `var x runtime.MemStats; runtime.ReadMemStats(&x)` idiom duplicated by the `--timing` feature commit `c29736a1`. The canonical gate had been skipped long enough for a real regression to ship unnoticed — exactly what T5 exists to stop.
- **Fix**: extracted `job.ReadMemStats()` (single home for the declare-and-read idiom), used at all three snapshot sites (`Profile`, `PrintProfileResult`, `startRunTiming`). Self-scan now emits 0 lines at threshold 1.
- **Final run: `all checks passed!`** — 13 checks including `self-test`, `alloc-gate` (first green CI run with the syntax budgets), `race`, `lint`, `bench`.

### T13 — `serial()` bulk Node allocation (−99.98% serialize allocs)

- Replaced per-node `&Node{}` copies with a pre-counted arena: `countSerializedNodes()` pre-pass → `make([]Node, count)` → `make([]*Node, 0, count)` exactly-sized stream; `nodeSerializer` struct writes shallow copies by arena index (index ≡ stream length).
- **Measured**: `BenchmarkSerialize/Size10000`: **10015 → 2 allocs/op**; bytes 1.47MB → 1.12MB (−24%); 100-node tree 105 → 2. All syntax/job/cache tests green, `TestSerializePreservesAllFields` green, self-scan output unchanged.

### T14 — `sync.Pool` for `[]*Node`: **NO-GO (measured, not laziness)**

- After T13, serialization costs 2 allocs/file (arena + stream header). Pooling the stream header saves ~916 allocs over a 916-file corpus = **~0.03% of the 4.43M objects/run**, at the price of cross-package lifetime plumbing (acquire in `syntax`, release in `job/buildtree` + `cmd/dump_tokens`, ownership contract, race surface). Same anti-verschlimmbesserung discipline as T7's defer verdict.

### T15 — Alloc budgets + gate extended to `syntax/`

- New `syntax/alloc_budget_test.go` (2 `AllocsPerRun` budget tests, `//go:build !race`, budget = 2 exact).
- `scripts/check-alloc-regression.sh` now runs `syntax/` benches too; 8 syntax budgets added to `scripts/alloc-budgets.txt`.
- Gate verified green; ratchet notes (MemoryUsage 4783 < 4802, par32 −1..−2) intentionally left — the deltas are map-iteration-order variance spanning the current budgets.

### T11 — Cache stats in `stats` subcommand (complete, e2e-verified)

- **Plumbing**: `job.RunCacheStats` (run-scoped!) on `ParseStatsMixin` → `printer.CacheMetrics` on `StatsConfig`/`StatsView` → all three output formats (text `Cache:` section, JSON `cache` object, CSV rows). Section omitted entirely without `--incremental`.
- **Critical correction mid-task**: first wiring used the cache's _persisted lifetime_ `HitCount/MissCount` — a fully-cached 1-file run displayed "33.3%" (2 misses from the double-check Get pattern + persisted counters). Reworked to run-scoped per-file counts (`IncrementalStats.CacheHits/CacheMisses`). Cold run now 0%, warm run 100%, honestly labeled.
- **Enabler**: the stats subcommand had no `--incremental` flag group at all — extracted `addIncrementalFlags()` shared by root and stats.
- Tests: 4 printer-format tests + 1 conversion test, all green. HOW_TO_USE section + help text updated.

## b) Partially Done

### T12 — `--exclude-pattern` zero-match warning (~70%)

- Done: `FilterStats.TrackExcludePatterns` / `recordPatternCandidate` (uses `gogenfilter.MatchPattern` — identical semantics to the real filter) / `WarnUnmatchedExcludePatterns`; wired into **both** file sources (directory walk + stdin feed), warning fires after crawl completes.
- Not done: the `TrackExcludePatterns(cfg.ExcludePatterns)` registration call (site identified: `cmd/run_analysis.go:404`, `NewFilterStats` creation), tests, glob-vs-regex docs (12.2), build+test of the current edits — **the T12 edits are unbuilt and unverified**.

### T2.2 — Pinned timing annotation (blocked, then deferred)

- The handoff's `/tmp/bench_pinned_t2.txt` (captured 10:13) is **uniformly slower than the unpinned v3 baseline (04:19) across all benchmarks including `seq`** — which pinning cannot affect. Verdict: machine-state difference 6h apart, not a pinning effect. A pinned-vs-unpinned claim from these two files would violate the project's own compare-like-with-like rule.
- Correct procedure defined: interleaved A/B runs (pinned, unpinned, pinned, unpinned) once the machine is free. Deferred behind the 40-min flake check (now done).

## c) Not Started

T16 (coverage baseline), T17 (TTY HTML auto-write), T18 (stable display IDs), T19 (cleanup trio), T20–T22 (property test, fuzz, boundary bench), T23 (perf stat A/B vs `23fa1b4f`), T24 (suffixtree doc polish), T25 (TODO_LIST + CHANGELOG sync), T26 (parked-tier triggers). Also: AGENTS.md updates for T13/T11, realworld-cli.md verdict outcome note.

## d) Totally Fucked Up (honest)

1. **Wasted a build cycle on a non-binary**: `go build -o /tmp/artdupl-selftest ./cmd` produced an `ar` archive (`./cmd` is a library; main is `./cmd/art-dupl`). I then chased the _shell's_ parse error (`>` must be followed by a word) as if it were tool output — two rounds — before running `file` on it. Lesson: verify the artifact type before debugging its behavior.
2. **Nearly shipped a lying metric**: the first T11 hit-rate wiring (lifetime counters) would have shown users a 33% hit rate on a 100%-cached run. Caught only because I ran the cold/warm E2E and the numbers looked wrong. Lesson: for every new metric, verify against a scenario where the ground truth is known.
3. **First flake check ran against a stale snapshot**: started it before finishing the session's code edits, so even a green run would not have covered the newest work. It failed on a real bug anyway (lucky), but the sequencing was wrong: run the canonical gate _after_ the tree settles, not in parallel with editing.

## e) What To Improve (in me)

- **Verify artifacts before interpreting errors** (`file`, `ls -la`) — 5 seconds that save 5 minutes of misdirected debugging.
- **E2E-verify metrics against known ground truth** — a metric without a controlled-scenario test is a liability, not a feature.
- **Serialize gate runs vs. editing** — the flake check is 40 minutes; starting it mid-edit wastes it. Stage → check → then continue.
- **Read the daemon's commit log before resuming** — expected auto-commits this time (none arrived mid-session, but the expectation was correct).

## f) Next Steps (in order)

1. Finish T12: add `TrackExcludePatterns` call at filter-stats creation; build; tests (matched/unmatched/nil); glob-vs-regex docs in HOW_TO_USE + flag help; verify warning fires on a real misconfig E2E.
2. Run `golangci-lint` on all touched packages (`cmd/`, `job/`, `printer/...`, `syntax/`).
3. T19 cleanup trio: delete dead `benchmarkFindTranMethod`, `b.N`→`b.Loop()` ×3 in `parallel_bench_test.go`, AGENTS "Workers routing" bullet fix.
4. T2.2: interleaved pinned/unpinned A/B (machine now free), annotate `baseline-2026-08-16-v3_notes.md` with honest numbers.
5. T23: `perf stat` cache-miss A/B vs `23fa1b4f` worktree → ADR-0022 evidence.
6. T16: coverage baseline (`go test -coverprofile`), commit under `docs/`.
7. T17: TTY-aware HTML auto-write. 8. T18: stable display IDs.
8. T20: slice-vs-reference-map property test. 10. T21: fuzz `findTran`/`addTran` high-fanout. 11. T22: `linearScanMax` 8-vs-9 boundary bench.
9. T24: suffixtree doc polish trio.
10. AGENTS.md updates: T13 arena serialization (incl. the `serial()` field-preservation hazard now living in `nodeSerializer.serial`), T11 run-scoped cache stats, T12 pattern warning, alloc-gate syntax coverage.
11. T25: TODO_LIST checkboxes (T5–T15), CHANGELOG entries (incl. the ReadMemStats self-test fix, T13 numbers).
12. T26: parked-tier review triggers.
13. Final verification: build, full tests, `-race`, lint, `nix flake check` (with everything staged _before_ starting it).
14. Commit hygiene: let the daemon pick up the staged work or commit at task boundaries with proper messages.

## g) Questions (max 3)

1. **Corpus drift** (carried over): go-cqrs-lite counts drifted 2665→2670 across runs — investigate with hash-seed pinning, or accept the concurrent-editing explanation?
2. **Standing rule** (carried over): should "stage new files immediately" be my standing rule, or do you prefer the auto-commit daemon as sole index-writer?
3. **Budget ratcheting**: the gate reports `MemoryUsage 4783 < budget 4802` (map-iteration-order variance, ±19 observed). Tighten to observed-min and accept occasional CI noise, or leave headroom as now?

---

_Wave 2 result: the trust layer is fully green (flake check passes with the new alloc-gate), serialization allocations are functionally eliminated and regression-guarded, and cache effectiveness is finally visible to users — honestly scoped._
