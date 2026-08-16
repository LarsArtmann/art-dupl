# Master Plan Execution — T1 Measured the Real World, Trust Track Underway — Status Report

**Date**: 2026-08-16 07:50
**Session**: Executing `docs/planning/2026-08-16_04-27_measure-first-trust-and-signal-master-plan.md` (T1–T26) from the top
**Branch**: fork
**Prior Sessions**: `2026-08-16_04-23_suffixtree-followup-slice-transitions-race-fixes.md` (verification + suffix-tree overhaul), `2026-08-16_04-27_measure-first-trust-and-signal-master-plan.md` (the plan itself)

---

## Context

The master plan's thesis: _measure reality, then lock in trust, then cut noise, then polish_. This session started Tier A (T1, the measurement that gates everything) and drove through the Tier B trust tasks (T3, T4) plus the start of T5/T6. 4 of 26 tasks fully done, 2 partially done, 20 not started. The headline: **the measurement exists now, and it vindicates and corrects the perf roadmap in specific ways** (see §d).

---

## a) FULLY DONE

### 1. T1 — Real-world end-to-end benchmark with stage-time split (subtasks 1.1–1.5)

**Instrumentation** (new `--timing` flag, off by default):

- `job/timing.go`: `StageTiming` collector carried via context (`WithStageTiming`), nil-safe `RecordStage`. Stages: `crawl`, `parse` (active, summed across workers), `serialize` (active), `tree-build` (active, excludes channel waits). Phases: `ingest`, `search`, `print`, `total` (wall, documented to overlap where the pipeline streams).
- Instrumented: `job/parse.go` (sequential + parallel parse, serializeAST), `job/buildtree.go` (update loop), `job/incremental.go` (both parse paths), `cmd/run_crawl.go` (crawl goroutines incl. stdin), `cmd/run_analysis.go` (ingest phase, search drain), `cmd/run_flags.go` (flag wiring, print phase, report emission).
- `cmd/timing.go`: `runTiming` wrapper + memstats deltas (objects, bytes, GC cycles, pause). Report explains active-vs-wall semantics inline so numbers cannot be misread.
- Tests: `job/timing_test.go` (4 unit tests), `cmd/timing_test.go` (flag-on report contents, flag-off silence, `humanBytes`, row-order). All green.

**Benchmark infrastructure**: `scripts/bench-realworld.sh` (fixture arg, `RUNS`, `PIN_CORES` taskset support, `BINARY`, writes markdown + medians). Fixture: **cli/cli shallow clone pinned at `0eeec0b92edbe70199f9768522f831d3534f41ad`** (916 non-vendored `.go` files; cobra at 36 files was rejected as too small).

**Results** (`docs/benchmarks/realworld-cli.md`, 3 runs/mode, pinned `0-7,16-23` + unpinned comparison):

| component                       | time         | share of total wall (~181ms)       |
| ------------------------------- | ------------ | ---------------------------------- |
| suffix tree build (active)      | 94–102ms     | **52–56%**                         |
| suffix tree search (wall)       | 48–59ms      | **26–33%**                         |
| serialize (active)              | 28–32ms      | **15–18%**                         |
| parse (423–459ms summed active) | ~0% marginal | fully hidden by worker parallelism |
| GC pause                        | 1.1–1.4ms    | <1%                                |

**Verdict memo** (in the results doc, §Verdict):

- **T13/T14: GO** — gate (serialize ≥ 15%) met at the margin (15–18%); reframed as allocation-count wins (4.43M objects/run), not wall-time wins (GC is ~1.3ms).
- **T23 perf stat: GO** — tree dominates; locality evidence worth collecting.
- **ADR-0020 suffix-array gate: MET** (search ≥ 30%); stays parked, criterion now evidenced.
- **Surprise**: tree **build** (sequential Ukkonen) now outweighs search; build+search ≈ 80% of wall. The next wall-time lever is chunked trees/suffix arrays, not search workers.

### 2. T3 — Atomic/mutex-mixing audit, whole repo (subtasks 3.1a–3.4)

Audited every `sync`/`atomic` usage in production code:

| Site                               | Verdict                                                                                                                                                                                                                                                                                                         |
| ---------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `cache` HitCount/MissCount         | already all-atomic (prior session's fix) ✓                                                                                                                                                                                                                                                                      |
| **`printer/html.go::PrintClones`** | **REAL MIX, FIXED**: `dupls` was mutex-guarded while `iota++` and `stats.categoryCounts/priorityCounts/...` mutated unguarded beside it; `OutputHTML` reset `iota` outside the lock too. Fixed: single `mu` now guards all mutable per-run state; `groupIndex` snapshot taken under the lock for view rendering |
| `cmd/filter_stats.go`              | consistent single-mutex + documented immutable-after-construction field ✓                                                                                                                                                                                                                                       |
| `cmd/accept_directive.go`          | correct double-checked locking, `readFile` immutable ✓                                                                                                                                                                                                                                                          |
| `cache/lru.go`                     | consistent single-mutex, lock ordering documented ✓                                                                                                                                                                                                                                                             |
| `syntax/intern.go`                 | correct DCL pattern ✓                                                                                                                                                                                                                                                                                           |
| `job/buildtree.go` sentinelCounter | all-atomic ✓                                                                                                                                                                                                                                                                                                    |
| `suffixtree` build→search handoff  | happens-before via done channel + goroutine spawn; search is read-only ✓                                                                                                                                                                                                                                        |
| `sync.Pool`s, singleflight         | inherently safe ✓                                                                                                                                                                                                                                                                                               |

- **T3.3**: `TESTING.md` gained a "Concurrency: Atomic/Mutex Mixing" section — the cache race story, the 5 rules that prevent the class (one word one discipline; never reset by struct replacement; guard everything a method mutates or nothing; document immutable-after-construction; hand off via channel/spawn), and pointers to the regression tests.
- **T3.4**: full `go test -race ./...` **GREEN** (background job 027, all 25 packages ok) — includes the new tests below and the html fix.

### 3. T4 — Cache `Clear()` concurrent-stats regression tests (4.1a–4.1b)

`cache/clear_race_test.go`:

- `TestClearConcurrentWithGets`: 4 writers + 8 readers + 1 clearer run until 10 Clears complete; readers mutate returned clones to prove LRU-copy independence; asserts non-negative counters. Under `-race` this fails on the old code (struct-replacement reset) and passes on the fix — verified green.
- `TestConcurrentMixedAccess`: Set/Get/Has/Remove/Prune hammered concurrently in 5 goroutines × 200 ops.

---

## b) IN PROGRESS / PARTIALLY DONE

### T5 — `nix flake check` + `-race` cadence (5.1a done, 5.1b pending)

- `nix flake check` **ran** (first time in two sessions) and **failed** on `checks.disabled-linters`: **`tagliatelle` is enabled or configured in `.golangci.yml`** and the guard treats the checkout as read-only (nix sandbox), so its auto-remove `sed` cannot fire. Fix pending: edit `.golangci.yml` directly. Because the failure aborts the check set, `test`/`race`/`lint`/`self-test` checks have not been re-verified through nix this session (they are green via direct go commands).
- The `-race` cadence is **already structurally wired**: `checks.race` in `flake.nix` runs `CGO_ENABLED=1 go test -race ./...`. Remaining: record the cadence decision in AGENTS.md (T5.2a) after the tagliatelle fix.

### T6 — CI allocation-regression gate (6.1a, 6.1b mostly done)

- `scripts/check-alloc-regression.sh` written: runs the guarded suffixtree benchmarks (`STreeUpdate`, `FindDuplOver`, `MemoryUsage*`, `-count=1`, `-benchmem`), parses `allocs/op`, compares against a tab-separated budget file with **±1 tolerance**, fails on exceedance, prints ratchet hints when a benchmark beats its budget.
- Current measurements captured (deterministic across runs): `STreeUpdate/tokens_{100,500,2000}` = 101/547/2046; `FindDuplOver/threshold_{10,50,200}` = 1543/1539/1518; `MemoryUsage{Few,Many}Tokens` = 23/44; plus the parallel-bench matrix.
- **Pending**: write `scripts/alloc-budgets.txt`, wire as a nix check (`alloc-gate`), inject a deliberate regression to prove the gate fails (6.3a), revert and re-verify green (6.3b).

---

## c) NOT STARTED (20 of 26)

T2 (taskset protocol note — partially covered by the bench script's `PIN_CORES`, needs the README protocol + pinned timing re-run), T7 (defer-cleanup pattern — the −12 FP item), T8, T9, T10, T11, T12, T13–T15 (gated, verdict says GO), T16–T18, T19, T20–T22, T23 (verdict says GO), T24, T25, T26. Full detail: the plan file's §2–§3 tables remain the authoritative queue.

---

## d) WHAT I FOUND (worth knowing before continuing)

1. **The suffix tree is 80% of a real run** — but split differently than assumed: build 52–56% (sequential, inherently so), search 26–33%. Three sessions optimized the right component.
2. **Parse CPU is irrelevant to wall time** at default parallelism — the sequential tree builder absorbs it. Any future "make parse faster" work is dead weight unless tree build is parallelized first.
3. **`--search-workers` defaults to sequential** — search is 26–33% of wall and a parallel search already exists behind a flag. Not in plan scope; flagged in the verdict memo as the cheap wall-time lever if users hit slow runs.
4. **The htmlprinter had a latent mixed-access race** of exactly the class the cache race came from — the plan's "where there is one there are siblings" hypothesis was correct.
5. **`.golangci.yml` has drifted**: tagliatelle (banned since ADR-0016-era decisions) is enabled/configured again, which breaks `nix flake check` at `disabled-linters`. Likely reintroduced by tooling/daemon churn; the writable-checkout auto-fix masked it locally.
6. **The daemon committed mid-session** (`c29736a1 feat(perf): add --timing stage report and harden htmlprinter concurrency`) — expected behavior; remaining post-commit edits are in the working tree (TESTING.md, cache/clear_race_test.go, scripts/check-alloc-regression.sh).

---

## e) NEXT UP (resume order)

1. **T5.1b**: remove tagliatelle from `.golangci.yml` (enable list AND settings blocks), re-run `nix flake check` to completion; record the `-race`-per-flake-check cadence in AGENTS.md (T5.2a).
2. **T6.2–6.3**: write `scripts/alloc-budgets.txt` from the captured numbers, add the `alloc-gate` nix check, inject+revert a regression to prove the gate fires.
3. **T7** (defer-cleanup pattern) — the largest evidence-backed noise cut; then T8/T9/T10 as a unit.
4. **T13/T14/T15** (verdict: GO, allocation-count framing) then T23.
5. Everything else per the plan's §3 table order.

---

## f) VERIFICATION STATE AT SESSION END

- `go build ./...` — green
- `go test ./cmd/ ./job/ ./cache/` — green
- `go test -race ./...` — **green (all packages)**
- `nix flake check` — **red**: `disabled-linters` (tagliatelle); fix is next session's first action
- Lint — not yet re-run this session (gopls stdversion warnings are the known false positives; `golangci-lint` run pending after tagliatelle fix)
