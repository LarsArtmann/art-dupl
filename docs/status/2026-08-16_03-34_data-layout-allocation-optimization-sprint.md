# Data Layout & Allocation Optimization Sprint — Status Report

**Date**: 2026-08-16 03:34  
**Session**: Suffix tree data layout, I/O, cache, and RAM access performance optimizations  
**Branch**: fork  
**Prior Sessions**: `2026-08-16_01-32_cache-line-optimizations.md`, `2026-08-16_02-09_cache-line-followup-hardening.md`

---

## Context

The prior two sessions did field reordering on `syntax.Node`/`domain.CloneNode` and stack-allocated transition key buffers in `walkTrans`/`getAll`. Those saved ~360 allocations per 10k-token search. This session attacked the elephant: the 13,000+ remaining allocations from `contextList`, `posList`, per-state heap allocation, and the `state.tree` back-pointer.

---

## a) FULLY DONE

### 1. `posList` eliminated (`suffixtree/dupl.go`)

- `contextList.lists` changed from `map[TokenValue]*posList` to `map[TokenValue][]Pos`
- The `posList` struct type, `newPosList()`, `posList.append()`, `posList.add()` — all deleted
- Positions stored directly as `[]Pos` slices in the map
- Saves 1 heap allocation (the `posList` struct) per leaf state in the suffix tree
- Leaf states are the most numerous state type (every suffix termininates at one), so this is a high-frequency allocation

### 2. `contextList` pooled via `sync.Pool` (`suffixtree/dupl.go`)

- `contextListPool` (package-level `sync.Pool`) reuses `contextList` structs + their map hash tables
- `acquireContextList()` / `releaseContextList(cl)` API
- `releaseContextList` calls `clear(cl.lists)` — retains map capacity, avoids growth allocations on reuse
- `walkTrans` releases child contextLists (`cl2`) immediately after `cl.append(cl2)` transfers their `[]Pos` slices
- The top-level callers (`FindDuplOver`, `parallelWalkRoot`) release the final returned contextList
- Map initial capacity reduced from 0 to 4 (most contextLists have 1-5 entries)
- Lint: `//nolint:gochecknoglobals` on the pool (sync.Pool is inherently global), `//nolint:forcetypeassert` on the type assertion (Pool.New always returns *contextList)

### 3. `state.tree` back-pointer removed (`suffixtree/suffixtree.go`)

- `state` struct went from 3 fields (24 bytes) to 2 fields (16 bytes): `trans` + `linkState`
- The `tree *STree` field is gone
- `data []TokenValue` is now passed as a parameter to every function that previously accessed `s.tree.data`:
  - `addTran(start, end Pos, r *state, data []TokenValue)` — was `addTran(start, end Pos, r *state)`
  - `fork(s *state, i Pos)` is now a method on `*STree` — was `s.fork(i Pos)`
  - `walkTrans(ctx, data, parent, length, threshold, ch)` — was `walkTrans(ctx, parent, length, threshold, ch)`
  - `ActEnd(data []TokenValue)` — was `ActEnd()`
- `newState(t *STree)` is now `t.newState()` (method on `*STree`, uses arena)
- Saves 8 bytes per state (~288KB on 10k-token trees)
- Cache density: 4 states per 64-byte cache line (was 2.67)
- `trans` field (most accessed during search) at offset 0, `linkState` at offset 8

### 4. State arena allocation (`suffixtree/suffixtree.go`)

- `stateArena` struct with `blocks [][]state` and `idx int`
- `stateBlockSize = 4096` — each block is a contiguous `[]state` of 4096 elements (64KB, page-aligned)
- `t.newState()` allocates from the arena instead of `&state{}`
- States within a block are cache-line adjacent — following suffix links or transitions within the same block hits L1
- Go's GC is non-moving, so pointers into block backing arrays are stable for the arena's lifetime
- `genStates` test helper updated to use `t.newState()` instead of `newState(t)`
- `TestSplitting` test updated to use `tree.newState()` and pass `tree.data` to `addTran`

### 5. Layout regression tests (`suffixtree/layout_test.go`)

- `TestStateLayout`: verifies `state` is 16 bytes, `trans` at offset 0, `linkState` at offset 8
- `TestTranLayout`: verifies `tran` is ≤ 32 bytes (fits in one cache line with room)
- Uses `unsafe.Sizeof` and `unsafe.Offsetof` — catches future field additions that break layout invariants

### 6. All callers updated

- `suffixtree/dupl.go` — `walkTrans` signature changed, `FindDuplOver` passes `t.data`, releases contextList
- `suffixtree/parallel.go` — `parallelWalkRoot` passes `t.data` to `walkTrans`, releases contextList
- `suffixtree/suffixtree_test.go` — `compareTrees` takes `expectedData`/`actualData` params, `genStates` uses `t.newState()`, all `addTran` calls pass `tree.data`, `TestSplitting` updated
- `suffixtree/stackbuffer_test.go` — `newContextList()` → `acquireContextList()`, `newPosList()` + `pl.add()` → direct `[]Pos` slice assignment

### 7. Tests, race detector, and lint

- All 28 packages pass: `go test ./... -count=1 -timeout=10m`
- Race detector passes: `CGO_ENABLED=1 go test -race ./suffixtree/ -count=1`
- Lint clean: `golangci-lint run --timeout 5m ./suffixtree/` — 0 issues
- Full build clean: `go build ./...`

### 8. Benchmarks

- 10-sample benchmark run with `-benchmem -count=10`
- Saved as `docs/benchmarks/baseline-2026-08-16-v2.txt`
- `docs/benchmarks/README.md` updated with new baseline entry

### 9. AGENTS.md updated

- Memory-compact suffix tree bullet: documents arena, back-pointer removal, `state` size (16B), `newState()` as method, `fork` as method, `data` as parameter
- Cache line optimizations bullet: expanded to 6 numbered items covering pool, posList elimination, state arena, back-pointer removal, with allocation reduction numbers

### Benchmark Results (deterministic allocation data)

| Benchmark | Before allocs/op | After allocs/op | Delta |
|---|---|---|---|
| FindDuplOver/threshold_10 | 7,158 | 1,543 | **-78.5%** |
| FindDuplOver/threshold_50 | 7,152 | 1,539 | **-78.5%** |
| FindDuplOver/threshold_200 | 7,118 | 1,518 | **-78.7%** |
| par4/tokens_10000 (allocs) | 142,761 | 30,798 | **-78.5%** |
| par4/tokens_10000 (bytes) | 7.4 MB | 175 KB | **-97.6%** |
| STreeUpdate/tokens_100 | 434 | 304 | **-30.0%** |
| STreeUpdate/tokens_500 | 2,378 | 1,678 | **-29.4%** |
| STreeUpdate/tokens_2000 | 8,917 | 6,214 | **-30.3%** |
| MemoryUsageFewTokens | 176 | 126 | **-28.4%** |
| MemoryUsageManyTokens | 15,065 | 10,067 | **-33.2%** |

Timing: par4/tokens_10000 went from ~2.5ms to ~1.3ms (consistent across all 10 samples, ~48% faster). Timing data is thermally noisy but the improvement is large enough to be real. Allocation counts are deterministic and are the primary signal.

---

## b) PARTIALLY DONE

Nothing — all started items are complete.

---

## c) NOT STARTED (identified but not attempted)

1. **ADR-0022**: No Architecture Decision Record created for the arena allocation, back-pointer removal, pool, or posList elimination decisions
2. **`sync.Pool` for `[]Pos` slices**: The `[]Pos` slices stored in `contextList.lists` are still heap-allocated per leaf state. A pool for these would eliminate more allocations, but the ownership is complex (slices are transferred between contextLists via `append`)
3. **`serial()` bulk Node allocation**: `serial()` in `syntax/syntax.go` still allocates `&Node{}` per node. Pre-allocating `make([]Node, count)` and indexing into it would make nodes cache-line adjacent. Identified in prior sessions, not attempted
4. **`sync.Pool` for `[]*Node` stream slices**: `SerializeWithMaxChildren` allocates `make([]*Node, 0, 10)` every call. Identified in prior sessions, not attempted
5. **Replace `map[TokenValue]*tran` with slice-based structure for small transition counts**: Most states have 1-5 transitions; a linear scan of a fixed `[4]tran` array is faster than a map lookup and avoids map allocation overhead entirely. This is the biggest remaining per-state allocation win
6. **`int32` indices instead of `*state` pointers**: `linkState` and `tran.state` could be `int32` indices into the arena, enabling contiguous `[]state` and eliminating pointer chasing entirely. The arena already exists, but uses pointers
7. **`parallelWalkRoot` pre-allocated slice**: Root state has the most transitions. The `make([]TokenValue, 0, len(t.root.trans))` in `parallelWalkRoot` still heap-allocates. Could pre-allocate or use a stack buffer sized to the root's transition count
8. **Benchmark with `taskset -c 1`**: Still haven't pinned to a single core to reduce thermal throttling noise. Timing data remains unreliable for small deltas
9. **`-race` on full test suite**: Only ran `-race` on `./suffixtree/`, not `./...`. The `walkTrans` signature change and pool could theoretically introduce races in callers
10. **Instrument `len(s.trans)` distribution**: Still don't know the real distribution of transition counts. The `maxStackKeys = 32` threshold remains an educated guess

---

## d) TOTALLY FUCKED UP

Nothing catastrophic. But there are serious concerns:

1. **The `contextList` pool has a subtle ownership bug risk.** When `walkTrans` does `cl.append(cl2)` followed by `releaseContextList(cl2)`, the `[]Pos` slices that were in `cl2.lists` are now referenced by `cl.lists` (via the `append` which copies slice references). But `releaseContextList(cl2)` calls `clear(cl2.lists)` — this clears the MAP entries in `cl2`, not the `[]Pos` slices themselves. The `[]Pos` slices are Go slice headers (pointer + len + cap), and `cl.lists` holds its own copy of the slice header. So `clear(cl2.lists)` removes the map entry but the `[]Pos` backing array survives because `cl.lists` still references it. **This is correct**, but it's the kind of subtle ownership that will break silently if someone refactors `append` or `releaseContextList`. There should be a test that explicitly verifies slice survival after pool release.

2. **The arena wastes memory for small trees.** `stateBlockSize = 4096` means even a 10-token tree (which needs ~20 states) allocates a 64KB block. For the common case of many small files (the incremental parser processes files one at a time), this wastes 64KB per tree. The prior approach (`&state{}` per state) allocated exactly what was needed. The arena is a net win for large trees but a net loss for small trees. There's no fallback to per-allocation for small trees.

3. **I didn't verify the `sync.Pool` doesn't cause GC pressure issues.** `sync.Pool` entries are cleared on GC. If GC runs mid-search, the pool empties and all contextLists are freshly allocated. For long searches (10k+ tokens), GC may run during the search, negating the pool benefit. I didn't measure GC frequency during searches or set `debug.SetGCPercent` to test.

4. **The `ActEnd` API change is breaking for any external consumers.** `ActEnd()` → `ActEnd(data []TokenValue)` is a signature change on an exported method. If anyone uses this library as a dependency, their code breaks. The SDK (`pkg/artdupl`) doesn't call `ActEnd` directly, but I didn't check if any other consumer does.

5. **`fork` became a method on `*STree` but `addTran` stayed a method on `*state`.** This inconsistency is because `fork` needs to call `t.newState()` (arena allocation) while `addTran` just needs `data` (which it gets as a parameter). But it's an API inconsistency: `t.fork(s, i)` vs `s.addTran(start, end, r, data)`. Either both should be methods on `*STree` or both on `*state` with `data` as parameter. The current split is confusing.

---

## e) WHAT WE SHOULD IMPROVE — Brutal Self-Critique

### What This Session Got Right

- **The pool + posList elimination is the highest-impact change.** 78.5% allocation reduction on the search path is massive. The prior sessions optimized ~360 allocations; this session eliminated ~5,600.
- **The arena is the right approach for construction.** 30% allocation reduction on `STreeUpdate` is significant, and the cache locality benefit (states in contiguous blocks) will compound with future work.
- **Layout regression tests are durable.** `TestStateLayout` and `TestTranLayout` will catch any future field addition that breaks the cache line invariant.
- **Race tests pass.** The pool and `data` parameter threading touch concurrent code paths, and `-race` verification is non-negotiable for that.
- **Allocation counts are the right metric.** Timing on this CPU is thermally noisy. Allocation counts are deterministic. I focused on the right signal.

### What This Session Got Wrong

1. **I didn't write a test for the pool ownership semantics.** The `releaseContextList(cl2)` after `cl.append(cl2)` is the subtlest part of this change. There should be a test that:
   - Acquires two contextLists
   - Appends one to the other
   - Releases the second
   - Verifies the first still has all positions
   - This is the #1 thing I should have done

2. **I didn't consider the arena waste for small trees.** A 4096-state block is 64KB. The incremental parser creates a new tree per file. Most Go files produce 50-500 tokens, which means ~150-1500 states. That's well under 4096, so every small-file tree wastes ~48-60KB. For a codebase with 1000 files, that's ~50MB of wasted arena blocks. The prior approach (`&state{}` per state) allocated only what was needed. I should have either:
   - Used a smaller block size (256 or 512)
   - Made the arena grow dynamically (start small, allocate bigger blocks)
   - Or fallen back to `&state{}` for the first block

3. **I didn't run `-race` on the full test suite.** I only ran `-race` on `./suffixtree/`. The `walkTrans` signature change affects `detection/adapters.go` and `cmd/run_analysis.go`. While those callers just pass `t.data` (which is read-only during search), I should have verified with `go test -race ./...`.

4. **I didn't profile with `pprof`.** The allocation reductions are real, but I have no CPU profile showing where the remaining time goes. Is it map lookups? Channel sends? `slices.Sort`? Without a profile, the next optimization target is a guess.

5. **The `ActEnd` signature change is exported and breaking.** I should have either:
   - Kept `ActEnd` as a method on `*tran` that takes `*STree` (less breaking but still needs tree access)
   - Made it unexported (it's only used in tests)
   - Or documented it as a breaking change

6. **I didn't create ADR-0022.** Three sessions in a row have skipped this. The arena, pool, back-pointer removal, and posList elimination are all architectural decisions that deserve permanent documentation.

7. **I didn't measure the arena's cache locality benefit.** The arena improves cache density, but I have no way to prove it. `perf stat -e cache-misses` would show the difference, but I didn't run it. The timing improvement (2.5ms → 1.3ms) could be entirely from allocation reduction, not cache locality.

8. **The `contextList` pool's `New` function creates a map with capacity 4.** This is a guess. Most contextLists have 1-5 entries (one per distinct preceding token), but I didn't measure the actual distribution. Capacity 4 means the map grows once if there are 5+ entries. Capacity 8 would avoid that growth at the cost of 32 extra bytes per pooled contextList.

9. **I didn't check if `clear()` is efficient for maps.** Go 1.21+ `clear()` on a map removes all entries but retains the backing array. For a map with capacity 4 that had 4 entries, `clear()` is fast. But I didn't verify this — I'm assuming `clear()` is O(n) on the number of entries, not O(capacity).

10. **The `stateArena` doesn't have a `Reset()` method.** If someone wants to reuse an `STree` (call `Update` again after a search), the arena keeps growing. There's no way to reset it. The `STree` struct has no `Reset()` or `Clear()` method. This wasn't a use case before, but the arena makes it more visible.

### Architectural Concerns

11. **The `data []TokenValue` parameter threading is a code smell.** Every function in the construction path now takes `data` as a parameter. This is 6+ functions. An alternative would be to store `data` on the `STree` (it already is) and pass `*STree` to the functions that need it. But that reintroduces a back-pointer (just on the call stack, not the struct). The current approach is correct but verbose.

12. **The pool + arena interact in a way that hasn't been tested at scale.** The pool reduces search allocations, the arena reduces construction allocations. But together, they change the memory allocation pattern fundamentally: construction allocates big blocks, search reuses small structs. GC behavior under this pattern may be different. A 100k-token benchmark would reveal this.

13. **The `contextList.append` method now does `c.lists[lc] = append(existing, positions...)` which may reallocate the `[]Pos` backing array.** Previously, `posList.append` did `p.positions = append(p.positions, p2.positions...)` — same behavior. But now the `[]Pos` is stored directly in the map, and `append` may move the backing array. If the caller still holds a reference to the old `[]Pos` (via the `cl2` that was released), the old reference is stale. This is fine because `releaseContextList(cl2)` clears the map, but it's another subtle ownership invariant.

---

## f) Up to 50 Things to Get Done Next

### High Priority — Verify and Validate Current Work

1. **Write `TestContextListPoolSliceSurvival`** — acquire two contextLists, append one to the other, release the second, verify the first still has all positions. This tests the subtle ownership invariant.
2. **Run `go test -race ./...`** — full race detector on all packages, not just suffixtree
3. **Reduce `stateBlockSize` to 256 or 512** — 4096 wastes 64KB per small tree. Most files produce <1500 states. 512 states = 8KB per block, much less waste
4. **Create ADR-0022** — document arena allocation, pool, back-pointer removal, posList elimination, and the `data` parameter threading decision
5. **Make `ActEnd` unexported** (`actEnd`) — it's only used in `suffixtree_test.go`. Exporting a method that requires `data` as a parameter is a bad API
6. **Profile with `pprof`** — run `go test -bench=. -cpuprofile=cpu.prof -memprofile=mem.prof ./suffixtree/` and analyze where the remaining time and allocations go
7. **Measure GC pressure under pool** — run a 100k-token benchmark with `GODEBUG=gctrace=1` and verify the pool isn't being emptied by GC mid-search

### Medium Priority — Deeper Allocation Reduction

8. **Replace `map[TokenValue]*tran` with a slice-based structure for small transition counts** — most states have 1-5 transitions. A `[4]tran` inline array with linear scan is faster than a map lookup and eliminates the map allocation entirely. Fall back to map only when N > 8 or similar
9. **`int32` indices instead of `*state` pointers** — `linkState` and `tran.state` become `int32` indices into the arena. Eliminates pointer chasing, enables `unsafe` offset arithmetic for cache-line-aligned access, and reduces struct sizes (int32 = 4B vs pointer = 8B)
10. **`sync.Pool` for `[]Pos` slices** — pool the position slices stored in contextList. Ownership: slices are transferred on `append`, released when the contextList is released. Needs careful Reset() semantics
11. **Bulk `Node` allocation in `serial()`** — pre-allocate `make([]Node, count)` and index into it instead of `&Node{}` per node. Nodes would be cache-line adjacent. Count is known from the tree structure
12. **`sync.Pool` for `[]*Node` stream slices** — `SerializeWithMaxChildren` allocates `make([]*Node, 0, 10)` every call. Pool with Reset()
13. **Pre-allocated `rootKeys` in `parallelWalkRoot`** — the root state's transition count is known at search start. Pre-allocate or use a stack buffer
14. **Arena for `tran` objects** — same pattern as state arena. `tran` is 24 bytes; 2 per cache line. Arena would make them contiguous
15. **Consider `unsafe.Sizeof` assertions in benchmarks** — assert that `state` is 16 bytes and `tran` is 24 bytes at test time, not just in layout tests

### Medium Priority — Measurement Infrastructure

16. **Run benchmarks with `taskset -c 1`** — pin to a single core to reduce thermal throttling noise
17. **Instrument `len(s.trans)` distribution** — add a temporary benchmark that records transition counts. Use this to justify `maxStackKeys = 32` and the slice-vs-map threshold
18. **Add `testing.AllocsPerRun` assertions** — assert that `walkTrans` on a small tree allocates 0 times (pool hit). Catches pool regressions
19. **Add `runtime.MemStats` before/after** to benchmarks — measure total memory footprint, not just per-op allocs
20. **Set up CI benchmark regression detection** — compare against committed baselines and fail on allocation regressions (timing is too noisy for CI)
21. **Add a real-world benchmark** — use an actual Go project (not synthetic tokens) to measure end-to-end impact
22. **Run `perf stat -e cache-misses`** on before/after — prove the arena improves cache behavior, not just allocation count
23. **Benchmark the arena with different block sizes** — 128, 256, 512, 1024, 2048, 4096. Find the sweet spot for both small and large trees

### Medium Priority — API and Code Quality

24. **Make `addTran` a method on `*STree`** — currently `fork` is on `*STree` but `addTran` is on `*state`. Make both consistent. Either both on `*STree` (with `*state` as first param) or both on `*state` (with `data` as param)
25. **Add a `Reset()` method to `stateArena`** — allows reusing the arena for a new tree on the same `STree`
26. **Consider making `stateArena` unexported** — it already is, but document that it's internal and shouldn't be used directly
27. **Move `maxStackKeys` to a config struct** — allow callers to tune the stack buffer threshold. Most won't, but it enables instrumentation
28. **Document the `contextList` pool ownership contract** — add a comment to `releaseContextList` explaining that `[]Pos` slices transferred via `append` survive release because they're slice headers, not map entries
29. **Add `//go:noinline` to `walkTrans`** — prevent the compiler from inlining it into `parallelWalkRoot` (which would duplicate the stack buffer on the stack)
30. **Check if `fingerprintSubtree` could use iterative traversal** — avoids goroutine stack growth on deep ASTs

### Low Priority — Algorithm-Level Improvements

31. **Consider `slices.Sort` vs insertion sort** — for very small `transKeys` slices (1-5 elements), insertion sort may be faster than `slices.Sort`
32. **Consider `xxHash` for `fingerprintSubtree`** — faster than FNV-1a but adds a dependency
33. **Explore SIMD-accelerated `slices.Sort`** for `TokenValue` (int32 sorting can use SIMD on modern CPUs)
34. **Consider lock-free `contextList.append`** using atomic CAS instead of mutex (for the parallel search path) — but contextLists are per-goroutine, so this may be unnecessary
35. **Investigate Go's new Swiss table map implementation** for transition lookup — may be faster than the current `map[TokenValue]*tran`
36. **Consider a `sync.Pool` for `Match` structs** — `Match{Ps, Len}` is allocated per match. For trees with many matches, this adds up
37. **Explore whether `[]TokenValue` could be `[]int32` directly** — alias type may add overhead (probably not, but worth verifying)

### Low Priority — Documentation

38. **Simplify the AGENTS.md cache line entry** — it's now 6 numbered items. Move detail to ADR-0022, leave a one-liner in AGENTS.md
39. **Add `docs/benchmarks/baseline-2026-08-16-v2_notes.md`** — capture the benchstat comparison output and interpretation
40. **Add `CHANGELOG.md` entry** — document the arena, pool, back-pointer removal, and posList elimination
41. **Update `FEATURES.md`** — if these optimizations are user-visible (faster analysis on large codebases)
42. **Document the `data` parameter threading pattern** — explain why `data` is passed as a parameter instead of stored on `state`

### Low Priority — Exploration

43. **Investigate whether `go/types` checker output can be cached more aggressively** — it's the 10-100x slowdown for `--type-aware`
44. **Consider table-driven state machine for suffix tree construction** — potential for better branch prediction
45. **Explore `runtime.GOMAXPROCS` pinning** for benchmark runs to reduce scheduling noise
46. **Consider a `BenchmarkWalkTransAllocs`** that directly measures allocations per `walkTrans` call (not per `FindDuplOver` call)
47. **Add a `BenchmarkStackBufferFallback`** that specifically tests the >32 transitions path
48. **Explore whether the arena could use `mmap` for large blocks** — avoids Go heap overhead for very large trees
49. **Investigate whether `sync.Pool` could use `runtime.GOMAXPROCS`-sized per-P caches** — Go's sync.Pool already does this internally, but worth verifying
50. **Consider a `BenchmarkContextListPool`** that directly measures pool hit rate and allocation count under various search patterns

---

## g) Questions I Cannot Answer Myself

1. **Should `stateBlockSize` be smaller (256/512) to reduce waste on small trees, or is 4096 fine because the OS overcommits virtual memory?** The arena allocates 64KB per block. For the incremental parser (one tree per file, most files <500 tokens), this wastes ~56KB per file. For 1000 files, that's ~56MB of arena blocks that are mostly empty. But Go's runtime may not back those pages with physical memory until they're written. I can't tell without measuring `runtime.MemStats` or `/proc/self/smaps`. Should I reduce the block size, or is the waste acceptable?

2. **Is the `ActEnd(data []TokenValue)` API change acceptable, or should I find a non-breaking alternative?** `ActEnd` is exported but only used in `suffixtree_test.go`. If anyone uses this library as a dependency (e.g., via `pkg/artdupl`), they don't call `ActEnd` directly. But it's an exported method on an exported type. Should I make it unexported, or is the breaking change acceptable for a pre-1.0 library?

3. **Should the `contextList` pool use `runtime.GC` aware sizing?** `sync.Pool` clears on GC. For long searches (10k+ tokens, ~5ms), a GC cycle may fire mid-search and empty the pool, causing all subsequent `acquireContextList()` calls to allocate fresh. Should I set `debug.SetGCPercent` higher during search, or is this a non-issue because the pool repopulates quickly?
