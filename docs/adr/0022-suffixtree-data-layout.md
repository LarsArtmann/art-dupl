# ADR-0022: Suffix tree data layout — arena, pooled search state, slice-based transitions

## Status

Accepted (2026-08-16)

Supersedes the layout aspects of ADR-0019; ADR-0019's parallel-search scope
is unchanged.

## Context

Three optimization sessions (2026-08-16) attacked allocation pressure and
cache behavior in the suffix tree, the hottest non-parsing stage of the
pipeline. The starting point allocated per: state struct (`&state{}`), state
transition map (`make(map[TokenValue]*tran, 4)`), edge object (`&tran{}`),
position-list wrapper (`posList` struct), and search context
(`contextList` + its map). A 10k-token search allocated ~142k times
(parallel, 4 workers); construction of a 2k-token tree ~6.2k times.

CPU profiling of the search path showed map machinery dominating:
`maps.(*Iter).Next` 17.7%, `(*Iter).Init` 7.6%, `(*Map).Clear` 6.5%,
`mapassign_fast32` 7.0% — ~40% of search CPU spent hashing, iterating, and
clearing maps, before any tree logic runs.

Instrumentation (10k-token trees, uniform / Zipf-like / repetitive token
streams) established the state fanout distribution:

- **80–90% of states are leaves** (0 transitions; every suffix terminates at one)
- Internal states hold mostly **2–10 transitions**
- Only **1–34 states per 10k tokens exceed 32 transitions** (essentially only
  the root, which has one transition per distinct token)

Production builds **one tree per analysis run** (`job.BuildTree` feeds every
file's tokens into a single `STree`) — arena waste is bounded per run, not
per file.

## Decisions

### 1. States are arena-allocated in 16 KB blocks

`stateArena` hands out states from contiguous `[]state` blocks of
`stateBlockSize = 512` (512 × 32 B = 16 KB). Go's non-moving GC keeps
pointers into blocks stable. Block size was benchmarked at 256/512/1024/4096:
smaller blocks cap worst-case waste (one partially-filled block) for small
trees at no measurable cost for large trees; 512 is the documented sweet
spot.

### 2. No `tree *STree` back-pointer; `data` is a parameter

The back-pointer (8 B/state) was removed. Functions that previously read
`s.tree.data` take `data []TokenValue` as a parameter (`t.addTran`,
`t.fork`, `walkTrans`, `(*tran).actEnd`). This is verbose but keeps `state`
minimal and makes the read-only contract of search explicit. For
consistency, both mutation helpers are methods on `*STree`
(`t.addTran(s, ...)` and `t.fork(s, i)`); `ActEnd` was unexported to
`actEnd` because only tests use it and a `data`-parameter export is a bad
API.

### 3. Transitions are a sorted `[]tran` of values, not a map

`state.trans` is `[]tran` sorted by key, where a **transition's key is
derivable as `data[tr.start]`** — no key field is stored. Consequences:

- **Leaves never allocate** (nil slice) — this alone removed ~50% of state
  allocations, since 80–90% of states are leaves.
- Internal states allocate one lazily grown slice instead of a map hash
  table plus one `*tran` per edge; edges are stored by value.
- `findTran(data, c)` linearly scans ≤ `linearScanMax` (8) entries and
  binary-searches above that (root-scale fanout).
- `addTran` binary-searches the insert position and `slices.Insert`s.
- `walkTrans` iterates the slice directly in deterministic key order —
  no key extraction, no sort, no map re-lookup. `parallelWalkRoot` walks
  the root slice directly (its `rootKeys` allocation disappeared).

**Pointer-stability contract:** `findTran` returns `&s.trans[i]`, valid
only until the next `addTran` on the same state (insertion may grow the
backing array). All call sites (`testAndSplit`, `canonize`, search) hold
the pointer only across reads or across an `addTran` on a _different_
state, so the contract holds; `findTran`'s doc comment states it.

`state` is now 32 B (slice header + `linkState`), `tran` 16 B. Layout
invariants are enforced by `TestStateLayout` / `TestTranLayout`.

### 4. Search contexts are pooled; positions are plain slices

`contextList` (the per-`walkTrans` position collector) is recycled through
a `sync.Pool`; `releaseContextList` clears the map (retaining its hash
table for reuse) and returns the struct. The former `posList` wrapper was
eliminated: positions are stored directly as `[]Pos` map values.

**Ownership contract** (documented on `releaseContextList`, guarded by
`TestContextListPoolSliceSurvival`): `contextList.append` copies slice
_headers_ into the destination map; releasing the source clears only its
map entries — the `[]Pos` backing arrays survive via the destination's
headers. After release, the caller must not touch the contextList at all
(another goroutine may acquire it immediately).

### 5. Allocation budgets are regression-tested

`alloc_budget_test.go` asserts via `testing.AllocsPerRun` that constructing
a 200-token tree stays ≤ 240 allocs and searching a 2k-token tree ≤ 2800
(measured: 167 / 2011, ~40% headroom). Reintroducing per-state maps or
per-edge pointers blows the budget. The test is `//go:build !race` because
race instrumentation inflates allocation counts.

## Results (deterministic allocation data, 10-sample baselines)

| Benchmark                          | ADR-0019 era           | This ADR              | Delta                                |
| ---------------------------------- | ---------------------- | --------------------- | ------------------------------------ |
| STreeUpdate 100 tokens             | 304 allocs / 80 KB     | 101 allocs / 24 KB    | −67% / −70%                          |
| STreeUpdate 2000 tokens            | 6,214 allocs / 378 KB  | 2,046 allocs / 232 KB | −67% / −39%                          |
| MemoryUsage 10k tokens (5k unique) | 10,067 allocs / 914 KB | 44 allocs / 591 KB    | −99.6% / −35%                        |
| FindDuplOver (search allocs)       | 1,543                  | 1,543                 | unchanged (inherent `[]Pos`/`Match`) |
| Parallel search par4/10k           | ~2.5 ms                | ~1.0 ms               | ~2.5× faster                         |

CPU profile after: map machinery fell from ~32% to ~14% of search samples;
what remains is the `contextList` maps (positions keyed by preceding token).

## GC behavior

`GODEBUG=gctrace=1` on a 3 s search benchmark: ~479 GC cycles, each
emptying the pool twice, yet allocs/op varies by ±1 across samples — the
pool repopulates within a few operations after each GC. GC accounts for
2–5% of wall clock. No `SetGCPercent` tuning is warranted.

## Consequences

- `-race` on the full test suite surfaced (and we fixed) an unrelated
  pre-existing race in `cache`: metadata hit/miss counters are now read
  atomically in `saveMetadata` and reset atomically in `Clear`.
- `sync.Pool` for the `[]Pos` slices themselves was **rejected for now**:
  slices transfer between contextLists via `append` (which may reallocate),
  so lifetime tracking would out complexity the savings; the remaining
  search allocations are match output, which is the algorithm's product.
- `int32` arena indices instead of `*state` pointers were **deferred**: the
  slice-based layout already removed most pointer chasing; index arithmetic
  everywhere is a large, risky diff for an uncertain marginal win.
- The stack-buffer threshold `maxStackKeys` now applies only to
  `contextList.getAll` key extraction (contextList maps hold 1–5 entries);
  the transition side no longer needs it.
