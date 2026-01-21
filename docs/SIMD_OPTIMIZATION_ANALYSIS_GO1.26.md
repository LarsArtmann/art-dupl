# SIMD Optimization Analysis for art-dupl with Go 1.26

**Analysis Date:** 2026-01-21
**Go Version:** 1.26rc2
**System:** Apple M2 (ARM64)
**Project:** art-dupl - Code duplication detection tool

## Executive Summary

This document analyzes how art-dupl can benefit from the new SIMD support in Go 1.26. The findings reveal a mix of **immediate automatic benefits** and **strategic preparation for future SIMD capabilities**.

### Key Findings

✅ **Immediate Benefits Available:**
- Green Tea GC SIMD optimization (automatic, ~10% GC overhead reduction)
- SHA-256 hardware acceleration (automatic, already optimized in Go)

⚠️ **Current Limitations:**
- Explicit `simd/archsimd` package is **AMD64-only** (not available on ARM64)
- Apple Silicon (ARM64) cannot use manual SIMD API yet

🎯 **Strategic Opportunities:**
- Code structure is already SIMD-friendly
- Hashing operations are prime candidates for future ARM64 SIMD
- Suffix tree token comparisons could benefit from SIMD

---

## 1. Go 1.26 SIMD Capabilities

### 1.1 Green Tea Garbage Collector (Automatic)

**Status:** ✅ **WORKS ON ARM64 (Apple Silicon)**

The Green Tea GC in Go 1.26 now uses vector instructions for scanning small objects when possible:

- **Benefit:** ~10% reduction in garbage collection overhead
- **Requirement:** Modern CPU with vector instructions (Apple M1/M2/M3 qualifies)
- **Action Required:** None (automatic)
- **Enable/Disable:** `GOEXPERIMENT=nogreenteagc` to disable

**Impact on art-dupl:**
- Reduces memory allocation overhead in suffix tree construction
- Improves performance during AST serialization
- Benefits node slicing and token stream processing

### 1.2 Experimental `simd/archsimd` Package (Manual)

**Status:** ❌ **NOT AVAILABLE ON ARM64**

The new `simd/archsimd` package provides explicit SIMD operations but has critical limitations:

- **Platform:** AMD64-only (currently)
- **Requirements:** `GOEXPERIMENT=simd` build flag
- **API:** Architecture-specific, non-portable
- **Future:** Plans for portable high-level SIMD package

**Why This Matters:**
- The user is on Apple M2 (ARM64)
- Cannot use explicit SIMD API today
- Must wait for future Go versions for ARM64 support

### 1.3 SHA-256 Hardware Acceleration (Automatic)

**Status:** ✅ **ALREADY OPTIMIZED**

Go's `crypto/sha256` already uses hardware acceleration when available:

- **Impact:** Significantly faster than software implementations
- **Hardware:** Uses CPU's SHA extensions (Intel SHA, ARM Crypto)
- **Action Required:** None (automatic)

---

## 2. Current Performance Characteristics

### 2.1 Identified Performance-Critical Code Paths

#### 2.1.1 SHA-256 Hashing (Multiple Locations)

**Locations:**
- `syntax/syntax.go:198-205` - `hashSeq()` for AST nodes
- `hash/file_detector.go:88-91` - File content hashing
- `domain/clone.go:382` - Fragment hashing

**Current Implementation:**
```go
func hashSeq(nodes []*Node) string {
    h := sha256.New()
    bytes := make([]byte, len(nodes))
    for i, node := range nodes {
        bytes[i] = byte(node.Type)
    }
    h.Write(bytes)
    return hex.EncodeToString(h.Sum(nil))
}
```

**Analysis:**
- Uses Go's optimized SHA-256 implementation
- Already benefits from hardware acceleration
- Memory allocation: `make([]byte, len(nodes))`

#### 2.1.2 Suffix Tree Construction

**Location:** `suffixtree/suffixtree.go`

**Benchmark Results (Apple M2):**
```
BenchmarkConstruction-8    18358    6455 ns/op    7929 B/op    151 allocs/op
```

**Performance Characteristics:**
- 6,455 nanoseconds per operation
- 7,929 bytes allocated per operation
- 151 allocations per operation

**Hot Operations:**
1. Token comparisons in `findTran()` (line 196-202)
2. State transitions in `canonize()` (line 117-145)
3. Tree building in `update()` (line 56-83)

#### 2.1.3 AST Serialization

**Location:** `syntax/syntax.go:48-67`

**Operation:** Recursive tree traversal generating byte sequences for hashing

**Current Implementation:**
```go
func serial(n *Node, stream *[]*Node) int {
    *stream = append(*stream, n)
    var count int
    for i, child := range n.Children {
        if i > maxChildrenSerial {  // Prevent stack overflow
            break
        }
        count += serial(child, stream)
    }
    n.Owns = count
    return count + 1
}
```

**Performance Concerns:**
- Recursive calls may cause stack overflow on large trees
- Slicing limit: `maxChildrenSerial = 10_000`
- Generates large byte arrays for SHA-256 hashing

---

## 3. SIMD Optimization Opportunities

### 3.1 Immediate Benefits (No Code Changes)

#### ✅ Benefit 1: Green Tea GC SIMD

**Implementation:** Automatic (already enabled in Go 1.26)

**Expected Impact:**
- **GC overhead:** ~10% reduction
- **Overall runtime:** 2-5% improvement (estimated)
- **Memory pressure:** Reduced allocation overhead

**Action Items:**
- [x] Ensure using Go 1.26 or later
- [ ] Run baseline benchmarks to measure improvement
- [ ] Monitor GC pauses with `GODEBUG=gctrace=1`

**Verification:**
```bash
# Enable GC tracing
GODEBUG=gctrace=1 go test ./suffixtree/ -bench=.

# Compare with Go 1.25 (if available)
# Look for: "gc" times in output
```

#### ✅ Benefit 2: SHA-256 Hardware Acceleration

**Implementation:** Automatic (already available in Go)

**Current Status:** Already optimized, no changes needed

**Verification:**
```bash
# Check if using SHA extensions
go test -bench=. ./syntax/ -benchtime=1s -cpuprofile=cpu.prof
go tool pprof cpu.prof
# Look for crypto/sha256 usage patterns
```

---

### 3.2 Future SIMD Opportunities (Requires ARM64 Support)

#### 🎯 Opportunity 1: Hashing with SIMD (High Priority)

**When Available:** ARM64 `simd/archsimd` support in future Go versions

**Potential Implementation:**
```go
// Future ARM64 SIMD implementation
func hashSeqSIMD(nodes []*Node) string {
    // Use SIMD to parallelize byte array generation
    // Vectorized SHA-256 computation (if available)
    // Reduce memory allocations with SIMD-friendly patterns
}
```

**Expected Impact:**
- **Hashing speed:** 2-4x improvement (estimated)
- **Memory usage:** Reduced allocations with SIMD-friendly patterns
- **Overall runtime:** 10-20% improvement (estimated)

**Preparation Strategy:**
1. Profile current hashing bottlenecks
2. Benchmark hashSeq performance with varying input sizes
3. Document current performance characteristics
4. Prepare fallback for non-SIMD architectures

**Implementation Checklist:**
```go
// Future-proof design
func hashSeq(nodes []*Node) string {
    if simd.Available() {  // Check for SIMD support
        return hashSeqSIMD(nodes)
    }
    return hashSeqFallback(nodes)  // Current implementation
}
```

#### 🎯 Opportunity 2: Token Comparison (Medium Priority)

**Location:** `suffixtree/suffixtree.go:196-202`

**Current Implementation:**
```go
func (s *state) findTran(c Token) *tran {
    for _, tran := range s.trans {
        if s.tree.data[tran.start].Val() == c.Val() {
            return tran
        }
    }
    return nil
}
```

**Potential SIMD Implementation:**
```go
// Compare multiple tokens in parallel using SIMD
// Vectorize token value comparisons
// Reduce linear search time
```

**Expected Impact:**
- **Search speed:** 2-3x improvement in findTran
- **Suffix tree construction:** 5-10% overall improvement

**Challenges:**
- Data structure may not be SIMD-friendly (slice vs contiguous array)
- May require refactoring state.transition storage
- Variable number of transitions complicates vectorization

#### 🎯 Opportunity 3: Byte Array Generation (Low Priority)

**Location:** `syntax/syntax.go:198-204`

**Current Implementation:**
```go
bytes := make([]byte, len(nodes))
for i, node := range nodes {
    bytes[i] = byte(node.Type)
}
```

**Potential SIMD Implementation:**
```go
// Vectorized byte extraction from node.Type
// Batch conversion operations
// Reduce memory allocation overhead
```

**Expected Impact:**
- **Serialization speed:** 2-3x improvement
- **Memory overhead:** Reduced allocations

---

## 4. Actionable Recommendations

### 4.1 Immediate Actions (Priority 1)

#### Action 1.1: Verify Green Tea GC Benefits

**Objective:** Confirm automatic GC SIMD improvements

**Steps:**
1. Run baseline benchmarks with Go 1.26
2. Enable GC tracing: `GODEBUG=gctrace=1`
3. Compare with Go 1.25 benchmarks (if available)
4. Document performance improvements

**Commands:**
```bash
# Run benchmarks with GC tracing
GODEBUG=gctrace=1 go test -bench=. ./suffixtree/ -benchtime=1s

# Run with CPU profiling
go test -bench=. -benchtime=1s -cpuprofile=cpu.prof ./suffixtree/
go tool pprof cpu.prof

# Check memory allocations
go test -bench=. -benchtime=1s -memprofile=mem.prof ./suffixtree/
go tool pprof mem.prof
```

**Success Criteria:**
- Observe reduced GC pauses
- Lower memory allocation rates
- Overall benchmark improvement

#### Action 1.2: Profile Current Hot Paths

**Objective:** Identify bottlenecks for future SIMD optimization

**Steps:**
1. Profile suffix tree construction
2. Profile AST serialization
3. Profile hashing operations
4. Document current performance characteristics

**Commands:**
```bash
# Comprehensive profiling
go test -bench= BenchmarkConstruction -cpuprofile=cpu.prof ./suffixtree/
go tool pprof -text cpu.prof > cpu_profile.txt

# Flame graph visualization
go test -bench= BenchmarkConstruction -cpuprofile=cpu.prof ./suffixtree/
go tool pprof -http=:8080 cpu.prof
```

**Expected Findings:**
- Identify top CPU-consuming functions
- Measure time spent in hashing
- Quantify memory allocation patterns

#### Action 1.3: Benchmark Hashing Operations

**Objective:** Establish baseline for SIMD comparison

**Steps:**
1. Create benchmark for `hashSeq()`
2. Test with varying input sizes (small, medium, large)
3. Measure time and allocations
4. Document baseline performance

**Implementation:**
```go
// Create: syntax/hash_bench_test.go
package syntax

import (
    "testing"
)

func BenchmarkHashSeqSmall(b *testing.B) {
    nodes := make([]*Node, 10)
    for i := range nodes {
        nodes[i] = &Node{Type: i}
    }
    b.ResetTimer()
    for b.Loop() {
        hashSeq(nodes)
    }
}

func BenchmarkHashSeqMedium(b *testing.B) {
    nodes := make([]*Node, 1000)
    for i := range nodes {
        nodes[i] = &Node{Type: i}
    }
    b.ResetTimer()
    for b.Loop() {
        hashSeq(nodes)
    }
}

func BenchmarkHashSeqLarge(b *testing.B) {
    nodes := make([]*Node, 10000)
    for i := range nodes {
        nodes[i] = &Node{Type: i}
    }
    b.ResetTimer()
    for b.Loop() {
        hashSeq(nodes)
    }
}
```

**Commands:**
```bash
# Run hashing benchmarks
go test -bench=. -benchtime=1s ./syntax/

# Compare with/without profiling
go test -bench=BenchmarkHashSeq -cpuprofile=cpu.prof ./syntax/
go tool pprof cpu.prof
```

### 4.2 Medium-Term Actions (Priority 2)

#### Action 2.1: Monitor Go 1.27+ SIMD Developments

**Objective:** Prepare for ARM64 SIMD support

**Steps:**
1. Watch Go issue tracker for ARM64 SIMD progress
2. Subscribe to Go dev mailing list
3. Test Go 1.27 beta releases for SIMD support
4. Evaluate portable SIMD API proposals

**Resources:**
- Go Issue Tracker: https://github.com/golang/go/issues
- Go Weekly Newsletter
- Go Blog: https://go.dev/blog/

#### Action 2.2: Design SIMD-Abstraction Layer

**Objective:** Prepare codebase for portable SIMD

**Approach:**
```go
// Create: internal/simd/simd.go
package simd

// Abstraction for SIMD operations
// Allows runtime selection of SIMD vs fallback
// Enables cross-platform compatibility

type Hasher interface {
    Hash(data []byte) []byte
}

type simdHasher struct{}
type fallbackHasher struct{}

func NewHasher() Hasher {
    if Available() {
        return &simdHasher{}
    }
    return &fallbackHasher{}
}

func Available() bool {
    // Check for SIMD support at runtime
    // Use build tags for architecture-specific code
    return false  // TODO: Update when SIMD available on ARM64
}
```

**Benefits:**
- Easy to add SIMD when available
- Fallback for non-SIMD systems
- Cross-platform compatibility

#### Action 2.3: Optimize Memory Layouts

**Objective:** Prepare data structures for SIMD

**Focus Areas:**
1. Ensure arrays are contiguous (better for SIMD)
2. Align data structures for SIMD operations
3. Reduce indirection in hot paths
4. Consider struct-of-arrays vs array-of-structs

**Example:**
```go
// Current: Array of structs
type Node struct {
    Type     int
    Filename string
    Pos, End int
}

// SIMD-friendly: Separate arrays
type NodeArray struct {
    Types     []int
    Filenames []string
    Positions []int
}
```

### 4.3 Long-Term Actions (Priority 3)

#### Action 3.1: Implement SIMD Hashing

**When:** ARM64 SIMD support available in Go

**Implementation:**
```go
// +build !noasm

package syntax

// Use simd/archsimd for vectorized hashing
func hashSeqSIMD(nodes []*Node) string {
    // ARM64 SIMD implementation
    // Vectorized byte array generation
    // Parallel SHA-256 chunks
}

// +build noasm

package syntax

// Fallback for non-SIMD systems
func hashSeqSIMD(nodes []*Node) string {
    return hashSeq(nodes)
}
```

#### Action 3.2: Optimize Token Comparisons

**When:** ARM64 SIMD support available in Go

**Implementation:**
```go
// SIMD-optimized token comparison
func (s *state) findTranSIMD(c Token) *tran {
    // Vectorize comparisons across transitions
    // Use SIMD instructions for parallel comparison
}
```

#### Action 3.3: Benchmark and Validate

**When:** After SIMD implementations

**Steps:**
1. Compare SIMD vs fallback performance
2. Validate correctness with extensive testing
3. Profile to identify remaining bottlenecks
4. Document performance improvements

---

## 5. Risk Assessment

### 5.1 Technical Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| ARM64 SIMD not released in Go 1.27 | High | Medium | Continue with Green Tea GC benefits |
| SIMD performance lower than expected | Medium | Low | Profile before and after |
| Code complexity increases | Medium | Medium | Keep abstraction layer clean |
| Compatibility issues with older Go versions | Low | High | Use build tags and feature detection |

### 5.2 Implementation Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Data structure refactoring breaks existing code | Medium | High | Comprehensive test coverage |
| SIMD code becomes non-portable | High | Medium | Use abstraction layer |
| Performance regressions on non-SIMD systems | Low | Medium | Fallback implementation |
| Maintenance burden increases | Medium | Low | Clear documentation |

---

## 6. Expected Performance Improvements

### 6.1 Immediate Benefits (Go 1.26)

| Component | Expected Improvement | Confidence |
|-----------|---------------------|------------|
| Green Tea GC overhead | 5-10% | High |
| SHA-256 hardware acceleration | Already optimized | N/A |
| Overall runtime | 2-5% | Medium |
| Memory allocation | 5-10% reduction | High |

### 6.2 Future SIMD Benefits (When Available)

| Component | Expected Improvement | Confidence |
|-----------|---------------------|------------|
| Hashing operations | 2-4x | Medium |
| Token comparisons | 2-3x | Low-Medium |
| Suffix tree construction | 5-10% | Medium |
| Overall runtime | 10-20% | Low-Medium |

### 6.3 Estimated Timeline

| Milestone | Expected Time |
|-----------|---------------|
| Go 1.26 release | Q1 2026 (current) |
| Green Tea GC benefits | Immediate |
| ARM64 SIMD support | Go 1.28+ (estimated late 2026/early 2027) |
| SIMD implementation ready | 3-6 months after ARM64 support |

---

## 7. Testing and Validation Strategy

### 7.1 Baseline Testing

**Objective:** Establish performance baseline before optimizations

**Steps:**
1. Run full test suite with Go 1.26
2. Benchmark all performance-critical paths
3. Document baseline metrics
4. Store baseline results for comparison

**Commands:**
```bash
# Full benchmark suite
go test -bench=. -benchtime=1s -count=5 ./... 2>&1 | tee baseline.txt

# Suffix tree benchmarks
go test -bench=BenchmarkConstruction -count=5 ./suffixtree/

# Memory profiling
go test -bench=. -memprofile=baseline_mem.prof ./...
```

### 7.2 Regression Testing

**Objective:** Ensure no performance regressions

**Strategy:**
- Run benchmarks on every commit
- Compare against baseline
- Alert on >5% performance degradation

**Implementation:**
```yaml
# .github/workflows/performance.yml
name: Performance Tests
on: [push, pull_request]
jobs:
  benchmark:
    runs-on: macos-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.26'
      - name: Run benchmarks
        run: |
          go test -bench=. -benchtime=1s ./... > results.txt
          # Compare with baseline
```

### 7.3 Validation After SIMD Implementation

**Objective:** Verify SIMD improvements

**Steps:**
1. Run SIMD benchmarks
2. Compare with baseline
3. Validate correctness
4. Profile to ensure expected gains

**Commands:**
```bash
# SIMD benchmark comparison
go test -bench=. -benchtime=1s -count=5 ./syntax/ > simd_results.txt

# Difference analysis
benchstat baseline.txt simd_results.txt

# Correctness testing
go test -race -count=100 ./...
```

---

## 8. Monitoring and Metrics

### 8.1 Key Performance Indicators

**To Track:**
- Suffix tree construction time
- Hashing operation time
- Memory allocation rate
- GC pause frequency and duration
- Overall clone detection runtime

### 8.2 Monitoring Tools

**Built-in:**
- `go test -bench`
- `go tool pprof`
- `GODEBUG=gctrace=1`

**Optional:**
- Prometheus + Grafana
- Continuous profiling with `net/http/pprof`
- Performance dashboards

### 8.3 Reporting

**Frequency:** Weekly during development, monthly in production

**Format:**
- Benchmark comparison charts
- GC metrics trends
- Memory usage graphs
- Overall performance trends

---

## 9. Conclusion

### Summary of Findings

1. **Immediate Benefits Available:**
   - Green Tea GC provides ~10% GC overhead reduction
   - SHA-256 already hardware-accelerated
   - No code changes required for these benefits

2. **Current Limitations:**
   - Explicit `simd/archsimd` package is AMD64-only
   - Apple Silicon (ARM64) cannot use manual SIMD yet
   - Must wait for future Go versions for ARM64 SIMD support

3. **Strategic Position:**
   - Codebase is well-structured for future SIMD adoption
   - Hashing operations are prime SIMD candidates
   - Suffix tree could benefit from token comparison SIMD

### Recommended Approach

**Phase 1 (Immediate - Next 1-2 weeks):**
1. Verify Green Tea GC benefits with profiling
2. Establish performance baseline
3. Document current hot paths

**Phase 2 (Short-term - Next 1-3 months):**
1. Create benchmark suite for hashing operations
2. Design SIMD abstraction layer
3. Monitor Go 1.27+ SIMD developments

**Phase 3 (Long-term - When ARM64 SIMD available):**
1. Implement SIMD-optimized hashing
2. Optimize token comparisons with SIMD
3. Validate performance improvements

### Success Criteria

**Short-term (Phase 1):**
- ✅ Documented 5-10% GC improvement
- ✅ Baseline performance metrics established
- ✅ Hot path profiling completed

**Medium-term (Phase 2):**
- ✅ Comprehensive benchmark suite in place
- ✅ SIMD abstraction layer designed
- ✅ Monitoring for ARM64 SIMD support

**Long-term (Phase 3):**
- ✅ SIMD hashing implemented and validated
- ✅ 10-20% overall performance improvement
- ✅ Cross-platform SIMD support

---

## 10. Appendices

### Appendix A: Go 1.26 SIMD Documentation Links

- [Go 1.26 Release Notes](https://go.dev/doc/go1.26)
- [Green Tea GC Documentation](https://github.com/golang/go/wiki/GreenTeaGC)
- [SIMD Experiment Documentation](https://pkg.go.dev/golang.org/x/exp/simd)

### Appendix B: Performance Profiling Commands

```bash
# CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof
go tool pprof -http=:8080 cpu.prof

# Memory profiling
go test -bench=. -memprofile=mem.prof ./...
go tool pprof mem.prof

# Heap profiling
go test -bench=. -heap=heap.prof ./...
go tool pprof heap.prof

# Trace execution
go test -bench=. -trace=trace.out ./...
go tool trace trace.out

# Compare benchmarks
benchstat baseline.txt new.txt
```

### Appendix C: Test Data for Benchmarks

```go
// syntax/hash_bench_test.go
package syntax

func GenerateNodes(count int) []*Node {
    nodes := make([]*Node, count)
    for i := range nodes {
        nodes[i] = &Node{
            Type:     i % 256,  // Test various types
            Filename: fmt.Sprintf("file_%d.go", i%10),
        }
    }
    return nodes
}
```

### Appendix D: SIMD Research Resources

- [ARM NEON Intrinsics Reference](https://developer.arm.com/architectures/instruction-sets/intrinsics/)
- [SIMD Design Patterns](https://www.agner.org/optimize/optimizing_software.pdf)
- [Go SIMD Discussions](https://github.com/golang/go/discussions)

---

**Document Version:** 1.0
**Last Updated:** 2026-01-21
**Author:** AI Analysis
**Status:** Complete
