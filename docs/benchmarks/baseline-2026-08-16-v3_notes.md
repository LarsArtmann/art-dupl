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
  significant without `taskset` pinning (still open, see TODO_LIST).
- The pool + slice layout survive GC: `GODEBUG=gctrace=1` over a 3 s search
  run shows ~479 GCs while allocs/op varies by ±1 — the pool repopulates
  immediately after each collection.
- Regression protection: `suffixtree/alloc_budget_test.go` asserts
  construction ≤ 240 allocs (200-token tree) and search ≤ 2,800 allocs
  (2k-token tree) via `testing.AllocsPerRun` (`//go:build !race`).
