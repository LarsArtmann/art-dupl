# Baseline v3 Notes — slice-based suffix tree transitions (ADR-0022)

Comparison of `baseline-2026-08-16-v2.txt` (map transitions, 4096-state arena
blocks) against `baseline-2026-08-16-v3.txt` (sorted `[]tran` value slices,
lazy allocation, 512-state blocks). Same machine, count=10, deterministic
allocation columns; timing is thermally noisy on this CPU and treated as
secondary evidence only.

## Construction (STreeUpdate)

| Benchmark   | v2 (map)              | v3 (slice)            | Delta                   |
| ----------- | --------------------- | --------------------- | ----------------------- |
| tokens_100  | 304 allocs / 80 KB    | 101 allocs / 24 KB    | allocs −67%, bytes −70% |
| tokens_500  | 1,678 allocs / 154 KB | 545 allocs / 162 KB   | allocs −68%             |
| tokens_2000 | 6,214 allocs / 378 KB | 2,046 allocs / 232 KB | allocs −67%, bytes −39% |

Where the wins come from:

1. **Leaves never allocate.** 80–90% of states are leaves; their transition
   storage is a nil slice (was: an empty map).
2. **No `*tran` per edge.** Edges are values inside the state's slice.
3. **No map hash table per internal state** — one lazily grown slice instead
   (growth chain explains why allocs are not exactly 1 per internal state).
4. **512-state blocks** (16 KB) instead of 4096-state (128 KB with the larger
   `state`) cap small-tree waste: tokens_100 total bytes dropped 3.3x.

`tokens_500` bytes are slightly higher than v2 — the block boundary straddles
the state count, costing one extra partially-filled block. Negligible.

## Memory usage (10k tokens per tree)

| Benchmark                | v2                     | v3                 | Delta                     |
| ------------------------ | ---------------------- | ------------------ | ------------------------- |
| FewTokens (50 unique)    | 126 allocs / 237 KB    | 23 allocs / 187 KB | allocs −82%, bytes −21%   |
| ManyTokens (5000 unique) | 10,067 allocs / 914 KB | 44 allocs / 591 KB | allocs −99.6%, bytes −35% |

## Search (FindDuplOver / parallel)

Search allocs are unchanged by design (1,543 allocs/op threshold_10; 30,798
par4/10k): the remaining allocations are the `[]Pos` position slices and
`Match` results — the algorithm's output, not overhead. The win is CPU:

- CPU profile before: map iteration/init/clear/assign ≈ 32% of search samples.
- CPU profile after: ≈ 14% (remaining maps are the contextList position maps).
- Timing: threshold_10 ~160µs → ~143µs (median-of-10), par4/tokens_10000
  ~1.28 ms → ~1.0–1.2 ms. Direction consistent, magnitude thermally noisy.

## Interpretation

- Allocation columns are the trustworthy signal (deterministic, ±1 across
  samples). Timing deltas below ~15% on this CPU should not be treated as
  significant without `taskset` pinning — but see the 2026-09-22 pinning A/B
  below: on the slice layout pinning no longer buys speed, only tighter CIs.
- The pool + slice layout survive GC: `GODEBUG=gctrace=1` over a 3 s search
  run shows ~479 GCs while allocs/op varies by ±1 — the pool repopulates
  immediately after each collection.
- Regression protection: `suffixtree/alloc_budget_test.go` asserts
  construction ≤ 240 allocs (200-token tree) and search ≤ 2,800 allocs
  (2k-token tree) via `testing.AllocsPerRun` (`//go:build !race`).

## Pinned/unpinned A/B (2026-09-22, T2.2)

Interleaved A/B on compiled test binaries (`go test -c`, go1.27.1,
`GOEXPERIMENT=jsonv2`, fork@`dd7d211c` clean tree): 6 alternations per arm,
`-count=5` per invocation (30 samples/arm), 8 s cooldown between invocations.
Raw output: `pinned-unpinned-2026-09-22.txt`. The `-16`/`-32` GOMAXPROCS
suffixes were normalized so benchstat groups the arms.

**Ambient-load disclosure:** the entry criterion (load sustained < 4, no
foreign jobs) never opened — 12+ min of polling showed 1-min load oscillating
8–27 with bursts of `go`/`golangci-lint` from other sessions. The runs below
took place at 1-min load 8–19 (recorded per round in the raw file). The
interleaved design carries the comparison; note the residual bias runs
AGAINST the unpinned arm (it shares all 32 CPUs with foreign load while the
pinned arm claims 16 of them), so conclusions that favor unpinned are robust
in direction.

| Benchmark (30 samples/arm)      | Unpinned (32 CPUs) | Pinned (CCX 0-7,16-23) | Delta            |
| ------------------------------- | ------------------ | ---------------------- | ---------------- |
| FindDuplOver/threshold_10       | 117.5µs ± 19%      | 114.2µs ± 4%           | ~ (p=0.146)      |
| Parallel par4/tokens_10000      | 987.0µs ± 7%       | 921.1µs ± 11%          | −6.7% (p=0.080)  |
| Parallel parNumCPU/tokens_10000 | 463.1µs ± 2% (n=32) | 481.3µs ± 4% (n=16)  | **+3.9% (p=0.011)** |

(parNumCPU = the sub-benchmark whose worker count follows `runtime.NumCPU()`
under the affinity mask: par32 unpinned vs par16 pinned — the configuration
behind `CPU_TOPOLOGY.md`'s original 25-30% claim.)

**Verdict: on the slice layout, CCX pinning no longer wins.**

- Fixed par4: pinned is ~7% faster but not significant at p<0.05.
- NumCPU-matched workers: the full machine is significantly FASTER (+3.9%)
  than one CCX — the claim's direction reversed.
- Sequential search: no difference in central tendency; pinned runs have
  ~5× tighter CIs (±4% vs ±19%), which is why pinned runs remain the right
  protocol for regression detection.
- Scale: par32 improved from ~2.17-2.31 ms (2026-08-16, map layout) to
  463 µs — the layout change made the search scale across workers, which is
  what actually obsoleted the pinning advice.

`CPU_TOPOLOGY.md`'s measurement was taken on the map-based layout (its own
text says "map traversal over a read-only tree"): pointer-heavy transition
maps made the search L3-latency bound, so duplicating the tree across both
L3s was the dominant cost. ADR-0022's contiguous `[]tran` layout localizes
the working set well enough that cross-CCX duplication stopped mattering.

## Cache-counter A/B vs 23fa1b4f (2026-09-22, T23)

`perf stat -e cache-references,cache-misses` on compiled binaries at
`23fa1b4f` (last committed map-layout state, pre contextList-pool) vs
fork@`dd7d211c`, both arms pinned to CCX 0-7,16-23, interleaved ×4; each
invocation runs STreeUpdate/tokens_2000 + FindDuplOver/threshold_10 +
par4/tokens_10000 at `-count=5` (~equal wall time per arm, so process-total
counters are directly comparable).

**Counter mapping note (this platform):** perf maps the generic events to
AMD core-PMU L2-request/L2-miss counters (select 0x64 — raw config 0x964).
The Intel-style `LLC-load-misses` has no openable counterpart here: the
kernel registers no `amd_nb`/`l3` uncore PMU (`perf list` shows `l3_cache/*`
events, but they fail to open). `cache-misses` (L2 misses = LLC-bound
traffic) stands in as the closest equivalent, measured identically on both
arms.

| Counter (totals over 4 rounds) | 23fa1b4f (map) | HEAD (slice) | Delta  |
| ------------------------------ | -------------- | ------------ | ------ |
| cache-references               | 35.81e9        | 11.55e9      | −67.8% |
| cache-misses                   | 6.07e9         | 2.48e9       | −59.2% |
| miss rate                      | 16.9%          | 21.4%        | —      |

Round-to-round variation < 3% — event counters are immune to the ambient
load that muddies wall-clock. Absolute LLC-bound traffic fell 59% while the
miss RATE rose (16.9%→21.4%): the slice layout touches far fewer lines
overall, and the surviving references (output `[]Pos`/`Match` allocations,
GC) miss proportionally more often.

Same-window timing medians (both pinned, interleaved): construction
tokens_2000 240µs→141µs (−41%), sequential search threshold_10 236µs→112µs
(−52%), par4/tokens_10000 2.36ms→0.87ms (−63%).

**Provenance honesty:** `23fa1b4f` predates the search-path pooling, so its
search allocs (7,158/op threshold_10) exceed the v2 baseline file's 1,543 —
that file was captured mid-overhaul in an uncommitted intermediate state.
The counter A/B therefore measures the full `081e347f` overhaul (pool +
slices + arena) — i.e., exactly what ADR-0022 shipped.

## linearScanMax boundary (2026-09-22)

See `linear-scan-boundary-2026-09-22.txt` (pinned, 20 samples/size, the
quietest window of the session: load ~3-8). Per-probe cost = ns/op ÷ 2n
(each iteration probes every present key plus a miss above each):

| Size (transitions) | Branch used   | ns/op (2n probes) | ns/probe |
| ------------------ | ------------- | ----------------- | -------- |
| 4                  | linear scan   | 22.97n ± 5%       | 2.87     |
| 8 (= linearScanMax)| linear scan   | 67.66n ± 4%       | 4.23     |
| 9                  | binary search | 62.24n ± 1%       | 3.46     |
| 16                 | binary search | 123.2n ± 2%       | 3.85     |

The linear/binary crossover sits between 4 and 8: linear at 4 is the
cheapest lookup measured (~2.9 ns), linear at 8 is already ~18%/probe
slower than binary at 9, and binary degrades only logarithmically after.
`linearScanMax = 8` therefore gives up a little on the 6-8-transition
states but keeps the early-exit scan for the 2-5-transition majority;
the cutoff's doc comment in `suffixtree/findtran.go` carries these numbers.
