# ADR-0019: Memory-compact suffix tree + parallel search

## Status

Accepted

## Context

The ROADMAP listed two performance goals for large codebases (10000+ files):

1. **Parallel suffix tree construction** — Ukkonen's algorithm is single-threaded
2. **Streaming suffix tree** — reduce memory by not buffering all serialized nodes

Investigation revealed:

- **Ukkonen's is inherently sequential.** The active point `(s, start, end)` propagates
  through suffix links across every token. There is no way to parallelize the algorithm
  itself without replacing it with a fundamentally different data structure (e.g., parallel
  suffix array construction via DC3/skew).
- **The pipeline already streams.** Files arrive from parallel parsers via
  `chan []*syntax.Node`, and `BuildTree` feeds tokens to `STree.Update()` one at a time.
  The suffix tree IS built incrementally. The memory concern was the **duplicate storage**:
  `STree.data` stored `[]Token` (16 bytes per interface element) alongside `BuildTree.data`
  which stored `[]*syntax.Node` (8 bytes per pointer) — both pointing at the same objects.
  Total: 24N bytes for N tokens.
- **The DFS search phase (`FindDuplOver`) is parallelizable.** Root-level subtrees are
  disjoint (each suffix starts with exactly one first token), so they can be walked
  concurrently without affecting the match set.

## Decision

### 1. Store `[]TokenValue` instead of `[]Token`

Changed `STree.data` from `[]Token` (interface, 16 bytes/element) to
`[]TokenValue` (int32, 4 bytes/element). The tree only needs the comparison value
for its internal operations (`testAndSplit`, `canonize`, `findTran`, `walkTrans`).
Original `Token` objects are not retained.

Memory impact: tree storage drops from 16N to 4N bytes. Combined with the 8N
`[]*syntax.Node` in `BuildTree` (still needed for `FindSyntaxUnits` position-indexed
lookups), total goes from 24N to 12N — a **50% reduction**.

`Update(data ...Token)` still accepts `Token` for API compatibility; it extracts
`tok.Val()` immediately and discards the original object.

### 2. Parallel DFS search via `FindDuplOverParallel`

Added `FindDuplOverParallel(ctx, threshold, workers)` to `STree`. It dispatches
each root-level transition's subtree to a goroutine, limited by a semaphore to
`workers` concurrent walks. Each goroutine calls the same `walkTrans` function;
the returned `contextList` is discarded at the root level (root never emits a match
since `length=0 < threshold`).

The match set is identical to sequential `FindDuplOver`. Only the output order
differs (goroutine scheduling), which is safe because downstream code groups
matches by hash.

Wired through:

- `detection.Config.SearchWorkers` → `suffixTreeAdapter.searchWorkers`
- CLI: `--search-workers N` flag (root-only)
- SDK: `Options.SearchWorkers` field
- Convention: 0 or 1 = sequential (default), >1 = parallel with N workers

### 3. Ukkonen's construction remains sequential

Documented that true parallel construction is impossible without replacing Ukkonen's
algorithm entirely. The parallel search addresses the practical concern (throughput
on large codebases) since the search phase can dominate on trees with many nodes.

## Consequences

- `STree.At()` now returns `TokenValue` instead of `Token` (only used in benchmarks).
- `findTran` takes `TokenValue` instead of `Token` (internal API change, no external callers).
- Test helpers updated: `str2vals()` helper for `[]TokenValue`, `genStates` uses it.
- Benchmark results (AMD Ryzen AI MAX+ 395, 32 threads):
  - 1K tokens: 708μs seq → 456μs par4 (1.55x)
  - 5K tokens: 3501μs seq → 1535μs par4 (2.28x)
  - 10K tokens: 7934μs seq → 2355μs par32 (3.37x)
- Memory overhead of parallel search is negligible (~0.03% more allocs for sync primitives).
