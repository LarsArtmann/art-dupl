# Cache Line Optimization Sprint — Status Report

**Date**: 2026-08-16 01:32  
**Session**: Cache line optimizations for hot-path data structures  
**Branch**: fork  

---

## What Was Done

### a) FULLY DONE

1. **`syntax.Node` field reordering** (`syntax/syntax.go`)
   - Reordered from `[Type, Pos, End, Owns, Children, Filename, Name, VarType, Statement, Fingerprint, EnclosingReturnArity, InterfaceMethod, IsAlias]` to `[Children, Filename, Name, VarType, Type, Pos, End, Owns, Fingerprint, EnclosingReturnArity, Statement, InterfaceMethod, IsAlias]`
   - Pointer/string fields (72B) first, scalar fields (32B) last
   - `Val()`'s hot reads (`Type` @72, `Fingerprint` @88, `Statement` @96) now in a single 64-byte cache line instead of straddling two lines (previously `Type` @0, `Fingerprint` @84, `Statement` @80)
   - Updated `serial()` and `Clone()` struct literals to match new order
   - Size unchanged: 104 bytes

2. **`domain.CloneNode` field reordering** (`domain/clone_node.go`)
   - Same pattern: `Children`/strings first, scalars last
   - Actionability pattern matching reads (`BaseType` @72, `InterfaceMethod` @80, `IsAlias` @81) now in one cache line
   - Updated `syntaxToCloneNode` in `printer/clone_processor.go`
   - Size unchanged: 88 bytes

3. **Stack-allocated transition key buffers** (`suffixtree/dupl.go`)
   - `walkTrans`: replaced `make([]TokenValue, 0, len(s.trans))` with stack-allocated `[32]TokenValue` for maps with <=32 entries
   - `contextList.getAll()`: same stack buffer optimization
   - Fallback to heap allocation for maps >32 entries (rare, mostly root)
   - Verified via `go build -gcflags="-m"` that `stackBuf` does not escape to heap

4. **All tests pass** (28 packages, `go test ./... -count=1`)
5. **Build clean** (`go build ./...`)
6. **Lint clean** (no new issues; 46 pre-existing tagliatelle warnings in untouched files)
7. **AGENTS.md updated** with cache line optimization documentation

### b) PARTIALLY DONE

Nothing — all started items are complete.

### c) NOT STARTED (identified but not attempted this session)

- `suffixtree.state` field reordering (3 pointers = 24B, fits one cache line; `tree` pointer could be removed to save 8B per state but too invasive)
- `suffixtree.tran` field reordering (16B, already fits one cache line; nothing to optimize)
- Replacing `map[TokenValue]*tran` with a slice-based open-addressing hash table (would eliminate map overhead per state)
- `contextList` and `posList` pool/sync.Pool reuse (they escape to heap; each `walkTrans` call allocates new ones)
- `fingerprintSubtree` iterative (non-recursive) implementation for deep ASTs
- `serial()` bulk Node allocation (pre-allocate `[]Node` instead of per-node `&Node{}`)
- `TestSerializePreservesAllFields` does not test the `CloneNode` bridge (`syntaxToCloneNode`)

### d) TOTALLY FUCKED UP

Nothing catastrophic, but see **e)** for serious concerns.

---

## e) WHAT WE SHOULD IMPROVE — Brutal Self-Critique

### Critical Failures

1. **Benchmark methodology was garbage.** I ran only 3 samples (`-count=3`) which benchstat explicitly says is insufficient ("need >= 6 samples for confidence interval at level 0.95" and "need >= 4 samples to detect a difference"). The timing results show wild variance (FindDuplOver/threshold_10 went from 276us to 820us to 1542us) that makes the before/after comparison **statistically meaningless** for timing. The allocation reduction is the only thing that's actually measurable and trustworthy from these runs. I should have run `-count=10` minimum and used `benchstat` with proper sample sizes.

2. **No A/B isolation.** The before and after benchmarks ran at different times on a laptop CPU (AMD RYZEN AI MAX+ 395) with thermal throttling and frequency scaling. The par32/tokens_10000 benchmark went from 6.1ms to 2.2ms in the "after" run — that's not a real improvement, it's noise. I cannot claim any timing improvement from this data.

3. **The stack buffer threshold of 32 is unjustified.** I picked 32 because it felt right. I did not measure the actual distribution of `len(s.trans)` across suffix tree states in real-world usage. If most states have 1-3 transitions, a buffer of 8 would suffice and waste less stack space. If some states have 50+, the fallback path fires more than expected. I should have instrumented the distribution first.

4. **No regression test for the stack buffer optimization.** There is no test that verifies `walkTrans` produces identical results whether using the stack buffer or heap path. The existing tests cover correctness indirectly, but a dedicated test with a state that has >32 transitions would verify the fallback path.

5. **Gob serialization compatibility was NOT explicitly tested with old cache files.** I verified that gob uses field names (not positions) by running the cache tests, but I did not test that a cache file written with the OLD field order can be read with the NEW field order. If any existing user has a cache directory, their cache entries may be invalid. Gob encodes the concrete type, and field reordering should be safe since gob uses field names, but this was not explicitly verified with a real old-cache-file test.

6. **`NewSyntheticFileNode` struct literal was not updated.** The field order in struct literals throughout the codebase (e.g., `syntax/syntax.go:185`) still works because Go uses named field initialization, but I only updated `serial()` and `Clone()`. The `NewSyntheticFileNode` function may have a different field order in its literal that I missed.

7. **The `parallelWalkRoot` function was skipped without documenting why.** I said "it's a one-time call" but it has the same `make([]TokenValue, 0, len(t.root.trans))` pattern. The root state CAN have many transitions (one per unique first token), so this might actually be a case where the heap path is always needed. But I should have measured.

8. **No `unsafe.Sizeof` regression test added.** I verified struct sizes manually with a throwaway program but did not add a permanent test that fails if the struct layout changes in a way that breaks the cache line assumptions. A test like `TestNodeScalarFieldsInOneCacheLine` that checks `unsafe.Offsetof(n.Type) >= 64 && unsafe.Offsetof(n.IsAlias) < 128` would prevent future regressions.

9. **The `state` struct has a `tree *STree` pointer (8 bytes per state) that is only used for `data` access.** I identified this but didn't act on it. Passing `data` as a parameter instead of storing a back-pointer would save 8 bytes per state. For a 10k-token tree with ~36k states, that's ~288KB. I dismissed it as "too invasive" without actually assessing the blast radius.

10. **`contextList` and `posList` heap allocations are the elephant in the room.** Every `walkTrans` call allocates a `contextList` (map header + map) and potentially multiple `posList` slices. The stack buffer optimization saves ~360 allocs out of ~14,000 total. The 13,000+ remaining allocations are from `newContextList()`, `newPosList()`, and `contextList.append()` map operations. A `sync.Pool` for `contextList` or a pre-allocated arena would have far more impact than the stack buffer trick.

### Minor Issues

11. **AGENTS.md documentation is too verbose.** The cache line optimization entry is a wall of text. A concise one-liner with a link to an ADR would be better.
12. **No ADR was created.** Architectural decisions (field ordering, stack buffer threshold) should be documented in `docs/adr/`.
13. **The `maxStackTransKeys` constant is duplicated** (defined as `maxStackTransKeys` in `walkTrans` and `maxStackKeys` in `getAll`). Should be a shared package constant.

---

## f) Up to 50 Things to Get Done Next

### High Priority — Validate and Harden Current Work

1. Run proper benchmarks with `-count=10` and `benchstat` to get statistically significant results
2. Run benchmarks on an isolated CPU core (`taskset -c 1`) to reduce noise
3. Instrument `len(s.trans)` distribution across real-world codebases to validate the 32-entry threshold
4. Add a `TestNodeScalarFieldsInOneCacheLine` regression test using `unsafe.Offsetof`
5. Add a `TestCloneNodeScalarFieldsInOneCacheLine` regression test
6. Add a test for `walkTrans` with a state that has >32 transitions (verify fallback path)
7. Test gob backward compatibility: write cache file with old field order, read with new
8. Update `NewSyntheticFileNode` struct literal field order for consistency
9. Extract `maxStackTransKeys` / `maxStackKeys` into a shared `const maxStackKeys = 32`
10. Create ADR-0022 documenting the field reordering decision and tradeoffs

### Medium Priority — Deeper Cache Line Work

11. Remove `state.tree` back-pointer; pass `data []TokenValue` as parameter to `testAndSplit`/`canonize`/`addTran`/`fork` — saves 8B per state
12. Reorder `state` fields: `trans` first (most accessed), then `linkState`, then `tree` (or remove `tree`)
13. Consider `int32` instead of `*state` for `linkState` (state index into a pre-allocated `[]state` arena) — eliminates pointer chasing and enables cache-line-aligned bulk allocation
14. Consider `int32` indices for `tran.state` too — all states in one contiguous `[]state` slice
15. Profile with `pprof` to identify actual cache miss hotspots (not just theoretical)
16. Run `go test -bench=. -cpuprofile=cpu.prof` and analyze with `go tool pprof`

### Medium Priority — Allocation Reduction (bigger wins than stack buffers)

17. `sync.Pool` for `contextList` objects (each `walkTrans` allocates one)
18. `sync.Pool` for `posList` objects
19. Pre-allocated arena for `state` and `tran` objects (eliminates per-node heap allocation during construction)
20. Bulk Node allocation in `serial()`: pre-allocate `make([]Node, count)` and index into it instead of `&Node{}` per node
21. Reuse the `stream` slice across `SerializeWithMaxChildren` calls via a pool
22. Consider a `Reset()` method on `contextList` instead of `newContextList()` for pool reuse

### Medium Priority — Algorithm-Level Improvements

23. Replace `map[TokenValue]*tran` with a slice-based hash table for small transition counts (most states have <5 transitions; a linear scan of a `[4]tran` is faster than a map lookup)
24. Consider a compact `[]tran` per state instead of a map — linear scan for small N, switch to map only when N > threshold
25. The `slices.Sort(transKeys)` in `walkTrans` is for deterministic iteration order — consider whether this is actually necessary for correctness or just for test stability
26. If sort is only for test stability, make it optional and rely on the natural map iteration order in production (with a test-only sort wrapper)

### Low Priority — Code Quality

27. Simplify the AGENTS.md cache line entry to a one-liner + ADR link
28. Add `//go:noinline` to `walkTrans` to prevent the compiler from inlining it into `parallelWalkRoot` (which would duplicate the stack buffer)
29. Check if `fingerprintSubtree` could use an iterative stack-based traversal instead of recursion (avoids goroutine stack growth on deep ASTs)
30. Consider `FNV-1a` vs `xxHash` for `fingerprintSubtree` (xxHash is faster but adds a dependency)
31. Benchmark `slices.Sort` vs manual insertion sort for the small `transKeys` slices (insertion sort is O(n) for nearly-sorted data, and Go's `slices.Sort` has overhead for very small slices)

### Low Priority — Measurement Infrastructure

32. Add a benchmark with a real-world-sized Go project (not synthetic tokens) to measure end-to-end impact
33. Add memory profiling (`-memprofile`) to benchmark runs
34. Add `runtime.MemStats` before/after comparison to benchmarks
35. Set up CI benchmark regression detection (compare against committed baselines in `docs/benchmarks/`)
36. Consider adding `testing.AllocsPerRun` assertions to tests for hot-path functions

### Low Priority — Exploration

37. Investigate whether `go/types` checker output can be cached more aggressively (it's the 10-100x slowdown for `--type-aware`)
38. Consider `ragel` or table-driven state machine for the suffix tree construction (potential for better branch prediction)
39. Explore whether the `STree.data []TokenValue` could be `[]int32` directly (alias type may add overhead)
40. Consider SIMD-accelerated `slices.Sort` for `TokenValue` (int32 sorting can use SIMD on modern CPUs)
41. Explore ` synced.Pool` for `[]TokenValue` buffers used in `parallelWalkRoot`
42. Consider lock-free `contextList.append` using atomic CAS instead of mutex (for the parallel search path)
43. Investigate `runtime.GOMAXPROCS` pinning for benchmark runs to reduce scheduling noise
44. Consider a `BenchmarkWalkTransAllocs` that directly measures allocations per `walkTrans` call
45. Add a `BenchmarkStackBufferFallback` that specifically tests the >32 transitions path
46. Explore whether `map[TokenValue]*tran` could use a custom hasher with better cache behavior than Go's built-in map
47. Consider `swiss.Map` or `swiss.Table` (Go's new Swiss table map implementation) for transition lookup
48. Profile the `serial()` function's `&Node{}` allocation pattern — if nodes are allocated sequentially, they may already be cache-line aligned by the allocator
49. Investigate whether `[]*Node` (slice of pointers) in `serial()` causes pointer chasing that could be eliminated by storing `[]Node` (slice of values) instead
50. Run `go test -race ./...` to verify the field reordering doesn't introduce any data races in concurrent access patterns

---

## g) Questions I Cannot Answer Myself

1. **Do users have existing on-disk cache directories that need backward compatibility?** I verified gob uses field names (not positions), but if any user has a stale cache from before this change, I cannot verify without knowing the deployment context. Should I bump `CacheVersion` from 3 to 4 to force a cache invalidation?

2. **Is the `slices.Sort(transKeys)` in `walkTrans` required for correctness or only for deterministic test output?** If it's only for tests, removing it in production would eliminate the allocation entirely (including the stack buffer optimization, making it moot). I can't tell from the code alone whether downstream consumers depend on sorted match order.

3. **Should the stack buffer threshold be tuned per-workload?** The threshold of 32 is a guess. For small files (most Go files have <100 unique AST node types in the first-token position), 8 would suffice. For large monorepo analysis, the root state might have hundreds of transitions. Should this be a runtime-configurable parameter, or is a compile-time constant fine?
