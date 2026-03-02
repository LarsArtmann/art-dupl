# ADR 0001: Map-Based Transition Lookup in Suffix Tree

## Status

Accepted

## Context

The `--semantic` flag in art-dupl was causing severe performance degradation (9x slowdown) on large codebases. Investigation revealed the bottleneck was in the suffix tree's `findTran` method, which used O(n) linear search through a slice of transitions.

### Problem Details

- **Without semantic**: Few unique token types (~50), linear search is acceptable
- **With semantic**: Many unique token types (~5000+), linear search becomes O(n²)
- **Impact**: Tree construction time increased from ~900µs to ~978ms (1000x slower)

The semantic feature encodes identifier names into token values, creating many more unique token types than structural-only matching. This caused the suffix tree to have many more transitions per state, making linear search prohibitively expensive.

## Decision

Change `state.trans` from `[]*tran` (slice) to `map[int]*tran` (map) for O(1) transition lookup.

### Implementation

```go
// Before: O(n) linear search
type state struct {
    trans []*tran
}

func (s *state) findTran(c Token) *tran {
    for _, tr := range s.trans {
        if s.tree.data[tr.start].Val() == c.Val() {
            return tr
        }
    }
    return nil
}

// After: O(1) map lookup
type state struct {
    trans map[int]*tran
}

func (s *state) findTran(c Token) *tran {
    return s.trans[c.Val()]
}
```

## Consequences

### Positive

- **Performance**: 546x faster tree construction with many unique tokens
- **Semantic flag**: Now 5.9x faster than before, and 1.2x faster than non-semantic mode
- **Scalability**: Handles large codebases with many unique identifiers efficiently
- **Simplicity**: Code is simpler and easier to understand

### Negative

- **Memory overhead**: Maps use more memory than slices for small collections
  - Few tokens (50 unique): 172 KB → 172 KB (no change)
  - Many tokens (5000 unique): 172 KB → 901 KB (5x increase)
- **Iteration order**: Map iteration is non-deterministic, requiring explicit sorting where order matters
- **Allocation overhead**: More allocations during tree construction

### Neutral

- **API compatibility**: No public API changes, internal implementation only
- **Correctness**: No change in clone detection results

## Alternatives Considered

### 1. Hybrid Slice/Map Approach

Use slice for <= 8 transitions, map for > 8 transitions.

**Pros**: Best of both worlds - memory efficient for small states, fast for large states
**Cons**: More complex code, requires threshold tuning
**Status**: Deferred - may implement if memory becomes a concern

### 2. Sorted Slice with Binary Search

Keep slice sorted and use binary search.

**Pros**: Memory efficient, O(log n) lookup
**Cons**: More complex, insertion is O(n), slower than map for large n
**Status**: Rejected - map is simpler and faster

### 3. SIMD-Optimized Linear Search

Use SIMD instructions for parallel comparison.

**Pros**: No memory overhead, could be fast for medium-sized states
**Cons**: Complex, platform-specific, Go SIMD support is limited
**Status**: Rejected - map is simpler and sufficiently fast

## Performance Results

### Benchmarks

| Scenario | Before | After | Improvement |
|----------|--------|-------|-------------|
| Tree construction (5000 unique tokens) | 978 ms | 1.79 ms | **546x** |
| `--semantic` on art-dupl codebase | 1.077s | 0.182s | **5.9x** |
| vs non-semantic | 3.5x slower | 1.2x faster | **4.2x** |

### Memory Usage

| Token Diversity | Slice | Map | Overhead |
|----------------|-------|-----|----------|
| 50 unique | 172 KB | 172 KB | 0% |
| 5000 unique | 172 KB | 901 KB | 424% |

Note: Memory overhead is acceptable given the dramatic speed improvement and typical codebase characteristics (few files have >1000 unique identifiers).

## Related Decisions

- **Deterministic iteration**: Added explicit sorting in `walkTrans` and test helpers to ensure consistent results
- **Benchmark coverage**: Added memory and performance benchmarks to track future regressions

## References

- Issue: Semantic flag performance degradation
- PR: Map-based transition lookup optimization
- Benchmarks: `suffixtree/memory_bench_test.go`, `suffixtree/semantic_performance_test.go`
- Execution Plan: `docs/EXECUTION_PLAN.md`

## Date

2026-03-02

## Authors

- Lars Artmann (@LarsArtmann)
