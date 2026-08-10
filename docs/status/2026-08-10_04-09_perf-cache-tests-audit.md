# Status Report: Performance Benchmarking, Caching, and Tests Audit

**Date:** 2026-08-10 04:09  
**Session scope:** Investigated performance benchmarking, caching system, and test suite health. Fixed bugs found during audit.

---

## a) FULLY DONE

### 1. Cache Key Isolation Bug — FIXED
- **Bug:** `cache.Key(content)` was SHA-256 of file content only. Switching detection modes (semantic/exact/structural), maxChildren, or type-aware modes incorrectly reused cached ASTs from a prior configuration. This is a **silent correctness bug** — the user gets stale results with no warning.
- **Fix:** Added `cache.KeyWithParams(content, params)` to `cache/file_cache.go`. Added `IncrementalParser.cacheKey()` method to `job/incremental.go` that composes `mode:maxChildren:typeAwareTag` into the key.
- **Type-aware tag detection:** `SetTypeAwareData` inspects the `TypeAwareData` map's first entry for `EraseHash`: `""` = no type info, `"ta"` = type-aware (hash includes types), `"sg"` = suggest-generics (hash erased).
- **Tests added:** `TestKeyWithParams` (4 sub-tests), `TestIncrementalParserCacheKeyIsolationMode`, `TestIncrementalParserCacheKeyIsolationMaxChildren`.
- **Files changed:** `cache/file_cache.go`, `cache/file_cache_test.go`, `job/incremental.go`, `job/incremental_test.go`.

### 2. Cache.Prune Test Coverage — FIXED (0% → ~100%)
- **Bug:** `Prune` had 0% test coverage despite being a critical eviction function called on every cache miss when `maxCacheEntries > 0`.
- **Tests added:** `TestFileCache_Prune` with 4 sub-tests: zero-max no-op, fewer-than-max no-eviction, evicts oldest (with `os.Chtimes` for deterministic mtime ordering), empty-cache no-error.
- **Files changed:** `cache/file_cache_test.go`.

### 3. Crashing Benchmarks — FIXED (2 of 2)
- **BenchmarkFindTranSmall/Medium/Large/VeryLarge/Map:** Crashed with nil pointer dereference. `benchmarkFindTran` indexed `tree.root.trans[0]` treating the `map[TokenValue]*tran` as a slice. When no transition had key `0`, the result was nil. Fixed by iterating the map to get the first available transition.
- **BenchmarkTestAndSplit:** Crashed with index out of range `[1000]`. Called `testAndSplit` with `start=0, end=10` on a tree where `t.end` pointed past `t.data` bounds (set during `Update`). Fixed by resetting `t.end` to the last valid index and using `start > end` to exercise the endpoint-check path.
- **Files changed:** `suffixtree/suffixtree_bench_test.go`.

### 4. Benchmark Baseline — COMMITTED
- Ran full benchmark suite across all performance-critical packages.
- Saved to `docs/benchmarks/baseline-2026-08-10.txt` with comparison instructions in `docs/benchmarks/README.md`.

### 5. Full Test Suite — ALL GREEN
- All 31 packages pass (standard mode).
- All 31 packages pass with `-race` detector (CGO_ENABLED=1).
- `go vet` clean on all changed packages.

### 6. AGENTS.md Updated
- Documented cache key isolation convention.
- Added benchmark baseline pointer to Build & Test section.

---

## b) PARTIALLY DONE

### Benchmark Baseline Quality
- Baseline was a single `count=1` run, not `count=3` with benchstat. The second benchmark run (during the same session) showed 2-4x worse numbers due to thermal throttling on the AMD Ryzen AI MAX+ 395. The baseline file captures a single observation, not a statistically grounded median. A proper baseline needs `count=5` + `benchstat` on an idle machine.

### Test Coverage Holes Identified But Not Filled
- `cache/file_cache.go` `Get` (70%), `Set` (77.8%), `deserialize` (77.8%) — error paths for corrupt/invalid gob files are untested.
- `cache/file_cache.go` `saveMetadata` (75%) — filesystem write failure path untested.
- These were identified during coverage analysis but not addressed (out of session scope — the user asked about health, not to fix every coverage gap).

---

## c) NOT STARTED

### Items Identified During Audit But Not Actioned
1. **CacheVersion bump** — The cache key scheme changed, making ALL pre-existing on-disk cache entries orphaned (they'll never be looked up again because the key format is different). `CacheVersion` was NOT bumped. This is not a correctness bug (orphans are inert), but users with large caches will accumulate dead `.gob` files until manual `--clear-cache`. Should bump to `CacheVersion = 3` or document.
2. **BDD semantic performance benchmarks** (`bdd/semantic_performance_bench_test.go`) — Not run. These execute the full `art-dupl` binary on the project root and are slow.
3. **No in-memory cache layer** — Every `Get` reads a file + gob-deserializes. For large codebases with many cache hits, this is I/O-bound. An LRU in-process layer on top of the disk cache would eliminate redundant deserialization. Identified, not implemented.
4. **Prune is O(n log n) per miss** — When near capacity, every cache miss triggers a full directory scan + sort. Should batch pruning (e.g., prune when entries hit 110% of max, evict down to 90%).

---

## d) TOTALLY FUCKED UP

### Nothing destroyed, but:

1. **I didn't bump CacheVersion** — This is the closest thing to a fuckup. Every existing cache on disk is now silently orphaned. Users won't notice until disk space grows. Not a correctness issue, but a cleanliness miss.

2. **The baseline benchmark numbers are noisy** — The first run produced clean numbers (e.g., `BenchmarkSTreeUpdate/tokens_100 = 14,580 ns/op`), but the second run showed 2-4x degradation (`31,638 ns/op`) due to thermal throttling. The committed baseline captures a single potentially-throttled run. The numbers in the first 3x run (before the crash) are more representative but are in `/tmp/bench_baseline.txt`, not the committed file.

3. **65 gopls warnings unexamined** — The project has 65 gopls warnings about `json.Unmarshal requires go1.27 or later (file is go1.26)`. The go.mod says `go 1.26.5` but gopls thinks the `encoding/json/v2` Unmarshal requires 1.27. These are pre-existing (not caused by my changes), but I didn't investigate whether they're real version-compatibility issues or false positives from the `GOEXPERIMENT=jsonv2` setup.

---

## e) WHAT WE SHOULD IMPROVE

### Architecture
1. **In-memory LRU cache layer** on top of `FileCache` — avoids redundant gob deserialization on repeated cache hits within a single run.
2. **Batch pruning** — prune in batches (e.g., at 110% capacity, evict to 90%) instead of O(n log n) per miss.
3. **Cache key should include CacheVersion** — so a version bump automatically invalidates old entries without orphaning them on disk. Currently the version is checked at deserialization time, not at key-lookup time.

### Benchmarking
4. **CI benchmark regression gates** — The `TestPerfRegression*` tests have 50ms thresholds (very generous). Should tighten or add more targeted regression tests.
5. **benchstat in CI** — Run benchmarks in CI with `benchstat` comparison against committed baseline, fail on >15% regression.
6. **Proper baseline procedure** — `count=5` on an idle machine, committed as `.txt` with CPU/timestamp metadata.

### Testing
7. **Cache error-path tests** — Corrupt gob file, permission-denied directory, disk-full write failure.
8. **Cache concurrent Prune test** — No test for concurrent `Set` + `Prune` race.
9. **Suffix tree internal API benchmarks** — `BenchmarkTestAndSplit` now works but only exercises the `start > end` path (endpoint check). The `start <= end` path (transition split) is not benchmarked.

### Code Quality
10. **`typeAwareTag` relies on map iteration order** — `SetTypeAwareData` iterates `TypeAwareData` (a map) and breaks on first entry to detect `EraseHash`. This assumes ALL entries have the same `EraseHash`, which is true when constructed by `LoadTypeAwareData` but not enforced by the type system. Should validate or document the invariant.

---

## f) Up to 50 Things We Should Get Done Next

| # | Priority | Task |
|---|----------|------|
| 1 | HIGH | Bump `CacheVersion` to 3 (or document that old caches are orphaned) |
| 2 | HIGH | Re-run benchmark baseline with `count=5` on idle machine, save with benchstat |
| 3 | HIGH | Investigate 65 gopls go1.27 warnings — real issue or false positive from GOEXPERIMENT? |
| 4 | HIGH | Add cache error-path tests (corrupt gob, permission denied, write failure) |
| 5 | MED | Add concurrent Prune + Set race test |
| 6 | MED | Add in-memory LRU layer to FileCache (eliminate redundant gob deserialization) |
| 7 | MED | Implement batch pruning (110%/90% hysteresis) |
| 8 | MED | Encode `CacheVersion` into the key itself (auto-invalidation on version bump) |
| 9 | MED | Tighten perf regression test thresholds (50ms → realistic numbers) |
| 10 | MED | Add benchstat comparison step to CI (fail on >15% regression) |
| 11 | MED | Benchmark `testAndSplit` with `start <= end` path (transition split) |
| 12 | LOW | Add `CacheStats` method to `IncrementalParser` for programmatic hit/miss/size query |
| 13 | LOW | Document `TypeAwareData` invariant: all entries share the same `EraseHash` |
| 14 | LOW | Run BDD semantic performance benchmarks and compare with baseline |
| 15 | LOW | Add cache size metric to `--profile` output (disk bytes used by cache) |
| 16 | LOW | Consider `sync.Pool` for `[]byte` buffers in gob serialization |
| 17 | LOW | Add `cache.KeyWithParams` to the SDK boundary (`pkg/artdupl`) if incremental caching is ever exposed there |
| 18 | LOW | Consider content-hash-based cache invalidation for `CacheVersion` changes (check version at key-lookup time) |

---

## g) Questions I CANNOT Answer Myself

1. **Should I bump `CacheVersion` to 3?** The cache key scheme changed, orphaning all existing entries. Bumping the version would cause `Get` to reject old entries (version mismatch → delete + miss), cleaning them up proactively. But it also invalidates ALL caches for ALL users on their next run (minor UX impact). Alternatively, I could add a migration/cleanup step. What's your preference?

2. **Are the 65 gopls `go1.27` warnings real?** The go.mod says `go 1.26.5`, but gopls claims `json.Unmarshal` requires go1.27. This could be a gopls false positive from `GOEXPERIMENT=jsonv2`, or it could indicate the project actually needs go1.27 and the go.mod is wrong. I can't determine if this is blocking or cosmetic without knowing your intended Go version policy.

3. **Should the benchmark baseline be committed or generated in CI?** Committed baselines are machine-specific (this one is AMD Ryzen AI MAX+ 395). CI-generated baselines are reproducible but require a dedicated runner. What's the intended workflow — local comparison against committed baseline, or CI-gated regression detection?
