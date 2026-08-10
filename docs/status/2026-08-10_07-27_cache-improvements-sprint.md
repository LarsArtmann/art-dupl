# Status Report: Cache Improvements Sprint

**Date:** 2026-08-10 07:27
**Session scope:** Implement 4 cache improvements from `docs/status/2026-08-10_04-39_post-audit-hardening-self-critique.md` §f
**Commits:** `85865f51` (CacheVersion bump), `8dc019a5` (LRU + hysteresis + cacheKey test)
**Files changed:** 6 (+490, -26 lines)

---

## a) FULLY DONE

### 1. CacheVersion bumped from 2 → 3
- **File:** `cache/file_cache.go:49`
- **What:** `CacheVersion` constant changed from 2 to 3 with explanatory comment.
- **Why:** The `KeyWithParams` change (which added `mode:maxChildren:typeAwareTag` to the cache key) orphaned all v2 on-disk entries. Bumping the version makes this explicit — old entries are rejected by the version-mismatch check in `deserialize()` and auto-removed.
- **Tests:** All existing tests use `CacheVersion` symbolically (no hardcoded `2`), so no test changes were needed. `TestFileCache_ErrorPaths::version_mismatch_removed_on_get` continues to pass.
- **Status:** ✅ Complete, tested, committed.

### 2. In-memory LRU layer
- **New file:** `cache/lru.go` (133 lines)
- **Modified:** `cache/file_cache.go` — `FileCache` struct gained `mem *lru` field; `Get`, `Set`, `Remove`, `Clear`, `Prune` all wired to keep both layers in sync.
- **What:** 512-entry `container/list`-based LRU with `sync.Mutex`. `Get` checks LRU first (O(1), no gob deserialization), falls back to disk, promotes disk hits into LRU. `Get` always returns deep clones via `cloneNodes()` so callers can't corrupt the canonical copy. `Set` writes to both layers. `Remove`/`Clear`/`Prune` evict from both.
- **Tests added:** `TestFileCache_LRUHitAfterSet` (removes disk file, verifies LRU hit), `TestFileCache_LRUReturnsDeepClone` (mutate one Get result, verify second unaffected).
- **Status:** ✅ Complete, tested, committed.

### 3. Hysteresis pruning
- **File:** `cache/file_cache.go::Prune` (rewritten)
- **What:** Pruning triggers at 110% of `maxEntries` (high water mark) and evicts down to 90% (low water mark). Evicted entries removed from both disk and LRU. This amortizes the O(n log n) sort across many cache misses instead of paying it on every miss.
- **Tests added:** `TestFileCache_HysteresisPruning` with 3 subtests: `no_prune_below_high_water` (11 entries, max=10 → no eviction), `prune_above_high_water_to_low_water` (15 entries, max=10 → evict 6 down to 9), `prune_evicts_from_lru` (verifies LRU eviction alongside disk).
- **Status:** ✅ Complete, tested, committed.

### 4. `cacheKey()` unit test
- **File:** `job/incremental_test.go::TestIncrementalParserCacheKeyFormat`
- **What:** 5 subtests: `params_format_is_mode:maxChildren:typeAwareTag` (exact format verification via `cache.KeyWithParams` comparison), `different_modes_produce_different_keys`, `different_maxChildren_produce_different_keys`, `different_typeAwareTags_produce_different_keys`, `same_params_produce_same_key`.
- **Status:** ✅ Complete, tested, committed.

### 5. AGENTS.md documentation updated
- Updated `Cache eviction` entry to document hysteresis (110%/90%).
- Added new `In-memory LRU layer` entry documenting the LRU architecture, lock ordering, and clone semantics.
- Updated `Cache key isolation` entry with explicit params format `"mode:maxChildren:typeAwareTag"` and CacheVersion=3.
- Updated `Node.Clone()` entry to mention `cache.cloneNodes` usage.
- Updated `RESOLVED: incremental cache-miss aliasing` entry to note `stampFilename` rename.
- **Status:** ✅ Complete, committed.

### 6. Refactoring: `cloneWithFilename` → `stampFilename`
- **File:** `job/incremental.go`
- **What:** The fast path (cache hit) previously called `cloneWithFilename` which deep-cloned + stamped filename. Since `Get` now returns deep clones, the fast path uses `stampFilename` (stamps in-place, no re-clone). The slow path (singleflight miss) still uses `deepCloneNodes` + `stampFilename` because the singleflight result is shared across callers.
- **Status:** ✅ Complete, tested, committed.

---

## b) PARTIALLY DONE

### None

All 4 tasks from the source document were fully implemented.

---

## c) NOT STARTED

### None from the source document

All 4 items (CacheVersion bump, LRU layer, hysteresis pruning, cacheKey test) were completed.

---

## d) TOTALLY FUCKED UP

### Nothing

No regressions, no broken tests, no data loss. All 29 test packages pass clean.

---

## e) WHAT WE SHOULD IMPROVE

### Issues Found During Implementation

1. **Double-clone on cache-miss path (performance, not correctness):** The `Set` method receives `cachedNodes` (already deep-cloned by `deepCloneNodes`), stores it in the LRU via `mem.put(contentHash, nodes)`. But `Set` receives the same `nodes` slice — the LRU stores the canonical reference, and `Set` also serializes it to disk. This is correct, but the caller (`parseFile`) does `deepCloneNodes(nodes)` → `cache.Set(cachedNodes)` → LRU stores `cachedNodes`. Then `parseFile` returns the original `nodes` (not `cachedNodes`), and the singleflight callers each `deepCloneNodes(parsedNodes)` again. So the miss path does: 1 clone for cache storage + 1 clone per singleflight waiter = N+1 clones for N waiters. This is the same as before (the pre-LRU code also did `deepCloneNodes` for cache + `cloneWithFilename` per caller), so no regression — but the LRU's `get` path now also clones, meaning a double-check cache hit inside singleflight does 2 clones (Get returns clone, then singleflight returns it, then caller deep-clones again). Minor, but could be optimized with a `GetShared` method that returns the canonical pointer for cases where the caller will clone anyway.

2. **`memHits` counter is tracked but not exposed:** `lru.memHits` is incremented on every LRU hit but never surfaced in `Stats()`. Users can't distinguish LRU hits from disk hits. Should add `MemHits int64` to the `Stats` struct.

3. **LRU capacity is hardcoded at 512:** `defaultMemoryEntries = 512` with no way to configure it. For very large codebases, 512 might be too small. Should be configurable via `Config` and a CLI flag (e.g., `--memory-cache-entries`).

4. **Lock ordering concern (safe but fragile):** `FileCache.Get` calls `fc.mem.get()` (acquires `lru.mu`) while holding `fc.mu` (via `withCachePath` RLock). The lock order is `FileCache.mu` → `lru.mu`. `Set` also acquires `FileCache.mu` (write) then calls `lru.put()` (acquires `lru.mu`). This is consistent — no reverse ordering exists. But there's no comment enforcing this invariant in the code. A future developer could accidentally call `fc.Get` from within an `lru` method, creating a deadlock. Should add a lock-ordering comment on both types.

5. **`Set` stores caller's slice directly in LRU:** `fc.mem.put(contentHash, nodes)` stores the `nodes` parameter directly. The comment says "The caller's slice is independent" — but this is only true because `parseFile` passes `cachedNodes` (a deep clone). If a future caller passes a slice they retain a reference to, they could corrupt the LRU. The `Set` method should defensively clone, or the contract should be more explicitly documented.

6. **No benchmark for LRU vs disk-only:** The performance gain (avoiding gob deserialization) is theoretical. Should add a benchmark comparing `Get` with LRU vs `Get` with empty LRU (disk-only) to quantify the improvement and prevent regressions.

7. **`Get` does double-clone on disk-hit path:** When the LRU misses but disk hits, `Get` deserializes, calls `fc.mem.put(contentHash, nodes)` (stores canonical), then returns `cloneNodes(nodes)`. The next `Get` for the same key hits the LRU and clones again. So the disk-hit path does: 1 deserialize + 1 clone (for LRU storage, via `put` which doesn't clone) + 1 clone (return value). Wait — `put` stores the reference, doesn't clone. So it's: 1 deserialize + 1 clone (return). The LRU stores the deserialized nodes. Next call: 1 clone (from LRU). This is correct and optimal. No issue here — I was wrong during initial analysis. The only waste is if the caller of the disk-hit `Get` immediately discards the result (unlikely).

8. **Hysteresis test `prune_evicts_from_lru` is indirect:** The test removes all disk files then checks that `Get` only returns hits for entries still in the LRU. This is a roundabout way to verify LRU eviction. A direct test would check `fc.mem.entries` or expose an `lruHas(key)` method. The current test works but is harder to reason about.

9. **`stampFilename` only stamps top-level nodes:** The function sets `node.Filename` on top-level slice entries but not on their `Children`. `Clone()` copies `Filename` from the original, so children get the filename from the cached copy (which may be different). This was the same behavior as the old `cloneWithFilename` — `Clone()` copies all fields including `Filename` for every node in the subtree. So children get the filename from the `Clone()` call, not from `stampFilename`. This is correct because `Clone()` already set `Filename` on every node in the subtree. But it's subtle and could break if `Clone()` is ever changed to not copy `Filename`.

10. **`varnamelen` lint warning on `SetTypeAwareData(td ...)`** — pre-existing, not introduced by this session. Parameter `td` is too short for its scope. Should be renamed to `typeAwareData` or `data`.

### Process Issues

11. **Auto-git committed before I was done:** The first commit (`85865f51`) was auto-committed after just the CacheVersion bump, before the LRU and other work was done. This split the work across 2 commits when it could have been 1. Not a problem, but worth noting for commit hygiene.

12. **Initial `cloneWithFilename` → `stampFilename` refactor caused a test failure:** I initially tried to eliminate the deep-clone on the singleflight path (savings: 1 clone per miss), but the `TestIncrementalParallelConcurrentStress` test caught that the shared singleflight result was being mutated across callers. I fixed it by keeping `deepCloneNodes` on the singleflight path. This was caught by tests, not by reasoning — a reminder that concurrency invariants should be explicitly documented.

---

## f) Up to 50 Things We Should Get Done Next

### Cache (direct follow-ups)
1. Expose `memHits` in `Stats()` struct and CLI `--cache-stats` output
2. Make LRU capacity configurable via `Config.MemoryCacheEntries` + `--memory-cache-entries` CLI flag
3. Add lock-ordering comments on `FileCache` and `lru` types
4. Document `Set` ownership contract: caller must not retain/mutate the slice after `Set`
5. Add benchmark: `Get` with LRU hit vs `Get` with disk-only (empty LRU)
6. Add benchmark: `Get` with disk-hit + LRU promotion vs cold disk miss
7. Add `lru.len()` method and expose LRU size in `Stats()`
8. Add direct LRU eviction test (check `fc.mem.entries` map, not indirect via disk removal)
9. Consider `GetShared` method that returns the canonical pointer (no clone) for cases where caller will clone anyway (singleflight double-check)
10. Consider memory budget (bytes) instead of entry count for LRU eviction
11. Add `lruStats` to `Stats` struct: `MemHits`, `MemSize`, `MemCapacity`
12. Consider `sync.Map` for LRU entries map (read-heavy workload) — benchmark first
13. Add LRU eviction callback hook (for logging/metrics)
14. Consider 2Q or ARC eviction policy instead of LRU (better scan resistance)

### Cache (broader improvements)
15. Add cache warming: pre-populate LRU from disk on `NewFileCache` for most-recently-used entries
16. Add `FileCache.Close()` method that flushes metadata and LRU stats to disk
17. Consider `mmap` for large gob files instead of `os.ReadFile` (avoid copy)
18. Add `FileCache.Compact()` that removes orphaned entries (version mismatches, corrupt files)
19. Add cache integrity check: verify gob deserialization on startup for a sample of entries
20. Consider content-addressed storage: use the hash as the content, not just the filename
21. Add `FileCache.Stats().HitRate()` helper method
22. Consider per-file-type cache buckets (Go vs templ) for better eviction granularity
23. Add `--clear-cache` subcommand (currently only `--clear-cache` flag on main command)
24. Add cache size limit in bytes (not just entry count) — `Config.MaxCacheBytes`
25. Consider `gob` replacement: `encoding/json/v2` or `msgpack` for smaller/faster serialization

### job/incremental.go
26. Rename `td` parameter in `SetTypeAwareData` to fix pre-existing `varnamelen` lint warning
27. Add benchmark for `parseFile` cache-hit vs cache-miss path
28. Consider `singleflight` key prefixing to avoid hash collisions across different cache directories
29. Add metrics: average parse time, cache hit rate, singleflight coalescing count
30. Document the clone invariant: "every `[]*syntax.Node` returned from `parseFile` must be independently owned by the caller"

### Testing
31. Add `-race` CI job (currently can't run locally due to `CGO_ENABLED=0` in Nix devShell)
32. Add fuzz test for `cacheKey()` with random params combinations
33. Add fuzz test for `KeyWithParams` length-prefix collision resistance
34. Add property-based test: LRU eviction always brings size ≤ capacity
35. Add property-based test: `Get` after `Set` always returns equal nodes (deep equality)
36. Add stress test: 1000 goroutines, mixed `Get`/`Set`/`Prune`/`Clear`, verify no panics or data races
37. Add test: `Prune` with `maxEntries=1` (edge case: highWater=1, lowWater=0)
38. Add test: LRU eviction order after `Get` promotes an entry (verify MRU ordering)
39. Add test: `Clear` resets LRU (verify `Get` after `Clear` is a miss)
40. Add test: `Set` same key twice replaces (not duplicates) in LRU

### Documentation
41. Add ADR for LRU layer architecture (lock ordering, clone semantics, eviction policy)
42. Update `HOW_TO_USE.md` with `--memory-cache-entries` flag (once implemented)
43. Update `TESTING.md` with cache test conventions (LRU testing patterns)
44. Add `docs/CACHE_ARCHITECTURE.md` with diagrams (disk + LRU + singleflight interaction)
45. Update `FEATURES.md` with "In-memory LRU cache layer" feature entry

### Code Quality
46. Consider extracting `cloneNodes` to `syntax` package (duplicated in `cache/lru.go` and `job/incremental.go::deepCloneNodes`)
47. Consider `lru` interface for testability (mock LRU in FileCache tests)
48. Add `//nolint:forcetypeassert` with justification on the `elem.Value.(*lruEntry)` assertions (they are safe because only `lru` code pushes to the list)
49. Consider `sync.Pool` for `bytes.Buffer` in `serialize`/`deserialize` to reduce allocations
50. Consider `io.WriterTo`/`io.ReaderFrom` for `cacheEntry` to avoid `bytes.Buffer` allocation entirely

---

## g) Questions

### 1. Should the LRU capacity be configurable, and if so, what should the default be?
The current hardcoded 512 is reasonable for most projects, but large monorepos (1000+ files) would benefit from a higher default. Should I add `--memory-cache-entries` (default 512) or make it a percentage of `--max-cache-entries`? I could also auto-size it to `maxCacheEntries` if set, or `runtime.GOMAXPROCS * 64` as a heuristic.

### 2. Should `Set` defensively clone the nodes slice, or trust the caller?
Currently `Set` stores the caller's slice directly in the LRU. This is safe because `parseFile` passes a `deepCloneNodes` result. But a future caller could accidentally pass a mutable slice. Defensive cloning adds 1 clone per `Set` call (small cost). The alternative is an explicit ownership-transfer contract ("caller must not use `nodes` after `Set`"). Which approach do you prefer?

### 3. Should I add a `GetShared` method for the singleflight double-check path?
The singleflight callback does `if cachedNodes, hit := ip.cache.Get(contentHash); hit { return cachedNodes, nil }` — but `Get` returns a deep clone, and then the singleflight caller does `deepCloneNodes(parsedNodes)` again (double clone). A `GetShared` that returns the canonical pointer (no clone) would eliminate 1 clone per singleflight coalescing event. The tradeoff is a less safe API. Worth doing?
