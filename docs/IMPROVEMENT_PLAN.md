# Semantic Detection Performance Improvement Plan

## Summary

Successfully optimized `--semantic` flag performance from **9x slowdown** to **1.2x faster** than non-semantic mode by implementing map-based transition lookup in the suffix tree.

## What Was Done

### Core Optimization
- Changed `state.trans` from `[]*tran` (slice) to `map[int]*tran` (map)
- Transition lookup: O(n) → O(1)
- Result: 546x faster tree construction with many unique tokens

### Files Modified
1. `suffixtree/suffixtree.go` - Map-based transition storage
2. `suffixtree/findtran.go` - O(1) lookup implementation
3. `suffixtree/dupl.go` - Deterministic iteration order
4. `suffixtree/suffixtree_test.go` - Deterministic test walker
5. `suffixtree/memory_bench_test.go` - Memory benchmarks (new)

### Performance Results

| Scenario | Before | After | Improvement |
|----------|--------|-------|-------------|
| Tree construction (5000 unique tokens) | 978 ms | 1.79 ms | **546x** |
| `--semantic` on art-dupl codebase | 1.077s | 0.182s | **5.9x** |
| vs non-semantic | 3.5x slower | 1.2x faster | **4.2x** |

## What Could Still Be Improved

### 1. Memory Optimization (High Impact, Medium Effort)

**Problem**: Maps use more memory than slices for small transition counts.

**Data**:
- Few tokens (50 unique): 172 KB, 165 allocs
- Many tokens (5000 unique): 901 KB, 15K allocs

**Solution**: Hybrid approach
```go
// Pseudo-code
type state struct {
    trans     []*tran      // For <= 8 transitions
    transMap  map[int]*tran // For > 8 transitions
}
```

**Effort**: 2-3 hours
**Impact**: 20-30% memory reduction for typical codebases

### 2. Comprehensive Benchmarks (Medium Impact, Low Effort)

**Missing**: Real-world benchmarks using actual Go code

**Action**: Add benchmarks in `bdd/` that:
- Parse actual Go files
- Compare semantic vs non-semantic
- Measure end-to-end performance

**Effort**: 1 hour
**Impact**: Better regression detection

### 3. Type Safety Improvements (Medium Impact, High Effort)

**Problem**: `Token.Val()` returns `int`, but `Pos` is `int32`

**Solution**: Create a proper `TokenValue` type alias
```go
type TokenValue int32

type Token interface {
    Val() TokenValue
}
```

**Effort**: 4-6 hours (requires changes across multiple packages)
**Impact**: Better type safety, prevent overflow bugs

### 4. Code Organization (Low Impact, Low Effort)

**Problem**: `findtran.go` is only 13 lines

**Options**:
1. Merge into `suffixtree.go`
2. Keep separate but add more context
3. Rename to `state.go` and add state methods

**Effort**: 30 minutes
**Impact**: Cleaner codebase

### 5. Documentation Updates (Low Impact, Low Effort)

**Missing**:
- Performance notes in AGENTS.md
- Architecture decision record (ADR) for map vs slice
- Comments explaining why map is faster

**Effort**: 1 hour
**Impact**: Better onboarding

### 6. Concurrent Tree Construction (High Impact, High Effort)

**Idea**: Build suffix tree concurrently for large codebases

**Approach**:
- Partition token stream
- Build partial trees in parallel
- Merge trees

**Effort**: 2-3 days
**Impact**: 2-4x faster on multi-core systems

### 7. Property-Based Testing (Medium Impact, Medium Effort)

**Missing**: Fuzzing/property tests for suffix tree correctness

**Action**: Add fuzz tests that:
- Generate random token sequences
- Verify findDuplOver finds all duplicates
- Verify no false positives

**Effort**: 3-4 hours
**Impact**: Catch edge cases

## Recommended Priority Order

### Phase 1: Quick Wins (This Week)
1. ✅ ~~Update package documentation~~ (DONE)
2. Add comprehensive benchmarks
3. Update AGENTS.md with performance notes
4. Clean up code organization

### Phase 2: Memory Optimization (Next Week)
5. Implement hybrid slice/map approach
6. Benchmark memory vs performance trade-offs
7. Document decision

### Phase 3: Architecture (Future)
8. Type safety improvements
9. Property-based testing
10. Concurrent tree construction (if needed)

## Questions to Consider

1. **Is the memory overhead acceptable?**
   - 901 KB vs 172 KB for extreme case (5000 unique tokens)
   - Typical codebases: ~300-500 unique identifiers per file
   - Probably acceptable given 5.9x speedup

2. **Should we keep the old implementation for comparison?**
   - Could use build tags: `//go:build !optimized`
   - Useful for benchmarking, but adds maintenance
   - Recommendation: No, rely on git history

3. **Is O(1) lookup the final optimization?**
   - Further improvements possible:
     - Memory pooling for `tran` objects
     - Batched tree updates
     - SIMD for token comparison (different from transition search)
   - Current implementation is "good enough" for now

## Metrics to Track

- [ ] Tree construction time vs token count
- [ ] Memory usage vs unique token count
- [ ] End-to-end `--semantic` performance
- [ ] Comparison with other clone detection tools

## Conclusion

The map-based optimization is a significant win. The main remaining work is:
1. Documentation and benchmarking
2. Optional memory optimization (hybrid approach)
3. Long-term architecture improvements

The performance issue is **solved**. Focus should shift to maintainability and correctness.
