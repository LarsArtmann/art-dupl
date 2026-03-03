# Comprehensive Execution Plan: Post-Semantic Optimization

## 1. Reflection: What Was Forgotten & Could Be Better

### What I Forgot:

1. **Memory profiling before/after** - No baseline to quantify memory overhead
2. **ADR (Architecture Decision Record)** - No formal record of why map vs slice
3. **Property-based testing** - No fuzzing for suffix tree correctness
4. **Real-world benchmark comparison** - Only synthetic benchmarks
5. **Cleanup of stale diagnostics** - findtran_simd.go references still in LSP cache

### What Could Be Done Better:

1. **Incremental commits** - Should have committed each file change separately
2. **Benchmark-driven development** - Should have written benchmarks FIRST
3. **Documentation sync** - Should have updated docs WITH code changes
4. **Type safety** - Token.Val() returns int but should be int32 for consistency

### What Could Still Be Improved:

1. **Hybrid slice/map approach** - Memory optimization for small transition counts
2. **Better type models** - Stronger typing for Token values
3. **Concurrent tree building** - For very large codebases
4. **Property-based testing** - Catch edge cases

---

## 2. Multi-Step Execution Plan

### Phase 1: Quick Wins (Low Effort, High Impact) ⭐⭐⭐

#### Step 1.1: Fix Stale LSP Diagnostics

**Work:** 5 minutes  
**Impact:** Clean development environment  
**Action:** Restart LSP or verify file is truly gone  
**Verification:** No more findtran_simd.go errors in diagnostics

#### Step 1.2: Add Real-World End-to-End Benchmark

**Work:** 30 minutes  
**Impact:** Track real performance regression  
**Action:**

```go
// bdd/semantic_performance_bench_test.go
func BenchmarkSemanticDetectionRealWorld(b *testing.B) {
    // Use actual art-dupl codebase as test data
    // Compare semantic vs non-semantic
}
```

**Verification:** Benchmark runs and shows improvement

#### Step 1.3: Create ADR for Map-Based Optimization

**Work:** 20 minutes  
**Impact:** Document architectural decision  
**Action:** Create `docs/adr/0001-map-based-transition-lookup.md`  
**Verification:** ADR follows project template

#### Step 1.4: Update AGENTS.md with Performance Notes

**Work:** 15 minutes  
**Impact:** Team awareness  
**Action:** Add section on suffix tree performance characteristics  
**Verification:** AGENTS.md mentions O(1) lookup

---

### Phase 2: Type Safety & Architecture (Medium Effort, Medium Impact) ⭐⭐

#### Step 2.1: Create TokenValue Type Alias

**Work:** 1 hour  
**Impact:** Better type safety  
**Action:**

```go
// suffixtree/token.go
package suffixtree

type TokenValue int32

type Token interface {
    Val() TokenValue
}
```

**Verification:** All implementations compile, tests pass

#### Step 2.2: Refactor state struct for Better Encapsulation

**Work:** 1.5 hours  
**Impact:** Cleaner API  
**Action:**

```go
type state struct {
    tree      *STree
    trans     map[TokenValue]*tran  // Use typed key
    linkState *state
}

func (s *state) findTran(key TokenValue) *tran {
    return s.trans[key]
}
```

**Verification:** Tests pass, no performance regression

#### Step 2.3: Add Transition Count Statistics

**Work:** 45 minutes  
**Impact:** Debugging and optimization insights  
**Action:**

```go
func (s *state) transitionCount() int { return len(s.trans) }
func (t *STree) Stats() TreeStats { ... }
```

**Verification:** Stats() returns meaningful data

---

### Phase 3: Memory Optimization (Medium Effort, High Impact) ⭐⭐⭐

#### Step 3.1: Implement Hybrid Slice/Map Approach

**Work:** 3 hours  
**Impact:** 20-30% memory reduction  
**Action:**

```go
type state struct {
    tree       *STree
    transSmall []*tran           // For <= 8 transitions
    transMap   map[int]*tran     // For > 8 transitions
}

const mapThreshold = 8

func (s *state) addTran(start, end Pos, r *state) {
    key := s.tree.data[start].Val()
    tran := newTran(start, end, r)

    if len(s.transSmall) < mapThreshold {
        s.transSmall = append(s.transSmall, tran)
    } else {
        if s.transMap == nil {
            // Convert slice to map
            s.transMap = make(map[int]*tran, len(s.transSmall)+1)
            for _, t := range s.transSmall {
                s.transMap[s.tree.data[t.start].Val()] = t
            }
            s.transSmall = nil
        }
        s.transMap[key] = tran
    }
}

func (s *state) findTran(key int) *tran {
    if s.transMap != nil {
        return s.transMap[key]
    }
    // Linear search in slice (acceptable for <= 8 elements)
    for _, t := range s.transSmall {
        if s.tree.data[t.start].Val() == key {
            return t
        }
    }
    return nil
}
```

**Verification:**

- Tests pass
- Benchmark shows memory reduction
- No performance regression

#### Step 3.2: Benchmark Memory vs Performance Trade-off

**Work:** 30 minutes  
**Impact:** Data-driven decision  
**Action:** Run benchmarks comparing slice-only vs map-only vs hybrid  
**Verification:** Clear data on trade-offs

---

### Phase 4: Testing & Quality (Medium Effort, High Impact) ⭐⭐⭐

#### Step 4.1: Add Property-Based Tests for Suffix Tree

**Work:** 3 hours  
**Impact:** Catch edge cases  
**Action:**

```go
// suffixtree/prop_test.go
func TestPropertyFindDuplFindsAllDuplicates(t *testing.T) {
    // Generate random token sequences
    // Verify all duplicates are found
    // Verify no false positives
}
```

**Verification:** Tests catch intentional bugs

#### Step 4.2: Add Fuzzing for Tree Construction

**Work:** 1 hour  
**Impact:** Robustness  
**Action:**

```go
func FuzzTreeConstruction(f *testing.F) {
    f.Add([]byte{1, 2, 3, 1, 2, 3})
    f.Fuzz(func(t *testing.T, data []byte) {
        // Build tree with random data
        // Should not panic
    })
}
```

**Verification:** Fuzzing runs without panics

#### Step 4.3: Benchmark Regression Test in CI

**Work:** 1 hour  
**Impact:** Prevent performance regressions  
**Action:** Add benchmark comparison to CI workflow  
**Verification:** CI fails if benchmarks regress > 10%

---

### Phase 5: Advanced Optimizations (High Effort, Medium Impact) ⭐

#### Step 5.1: Research Concurrent Tree Building

**Work:** 4 hours  
**Impact:** 2-4x faster on multi-core  
**Action:**

- Research parallel suffix tree algorithms
- Prototype concurrent construction
- Measure speedup

#### Step 5.2: Memory Pool for tran Objects

**Work:** 2 hours  
**Impact:** Reduce GC pressure  
**Action:** Use `sync.Pool` for `tran` allocation  
**Verification:** Fewer allocations in benchmarks

---

## 3. Sorted by Work vs Impact

### Immediate (Do Today)

| Step | Work | Impact | Task                  |
| ---- | ---- | ------ | --------------------- |
| 1.1  | 5m   | ⭐⭐⭐ | Fix stale diagnostics |
| 1.2  | 30m  | ⭐⭐⭐ | Real-world benchmark  |
| 1.3  | 20m  | ⭐⭐   | Create ADR            |
| 1.4  | 15m  | ⭐⭐   | Update AGENTS.md      |

### This Week

| Step | Work | Impact | Task                  |
| ---- | ---- | ------ | --------------------- |
| 2.1  | 1h   | ⭐⭐   | TokenValue type       |
| 2.2  | 1.5h | ⭐⭐   | Refactor state struct |
| 3.2  | 30m  | ⭐⭐⭐ | Benchmark trade-offs  |
| 4.2  | 1h   | ⭐⭐⭐ | Fuzzing               |

### Next Week

| Step | Work | Impact | Task                 |
| ---- | ---- | ------ | -------------------- |
| 3.1  | 3h   | ⭐⭐⭐ | Hybrid slice/map     |
| 4.1  | 3h   | ⭐⭐⭐ | Property-based tests |
| 4.3  | 1h   | ⭐⭐   | CI benchmark check   |

### Future

| Step | Work | Impact | Task                |
| ---- | ---- | ------ | ------------------- |
| 5.1  | 4h   | ⭐⭐   | Concurrent building |
| 5.2  | 2h   | ⭐⭐   | Memory pooling      |

---

## 4. Existing Code That Fits Requirements

### For Hybrid Approach:

- **Existing pattern:** `internal/simd/simd.go` - conditional optimization
- **Existing benchmark infra:** `suffixtree/suffixtree_bench_test.go`
- **Existing test patterns:** `suffixtree/dupl_test.go`

### For Type Safety:

- **Existing pattern:** `domain/line_number.go` - strong typing with validation
- **Existing pattern:** `domain/threshold.go` - typed int with bounds checking

### For Property Testing:

- **Existing:** `suffixtree/suffixtree_test.go` has basic fuzz test
- **Can extend:** Add more generators and properties

---

## 5. Type Model Improvements

### Current Issues:

```go
// Token.Val() returns int - inconsistent with Pos (int32)
type Token interface {
    Val() int  // Should be int32 or custom type
}

// state.trans uses int key
map[int]*tran  // Should match Token value type
```

### Proposed Type Model:

```go
// suffixtree/types.go
package suffixtree

// TokenValue represents a unique token identifier
type TokenValue int32

// Position in the token sequence
type Position int32

// TransitionKey is the key type for state transitions
type TransitionKey = TokenValue

// Token is the interface for all tokens
type Token interface {
    Val() TokenValue
}

// StateID identifies a unique state (for debugging/analysis)
type StateID uint32
```

### Benefits:

1. Type safety - can't mix up token values with positions
2. Clarity - self-documenting code
3. Consistency - all int32-based for 32-bit systems
4. Validation - can add bounds checking

---

## 6. Well-Established Libraries to Consider

### For Testing:

- **`github.com/stretchr/testify`** - Already used? Check
- **`pgregory/rapid`** - Property-based testing (better than stdlib fuzzing)
- **`google/go-cmp`** - Already in go.mod (indirect)

### For Performance:

- **`github.com/cockroachdb/swiss`** - Swiss table map (faster than Go map)
- **`github.com/dolthub/swiss`** - Alternative Swiss table

### For Concurrency:

- **`github.com/sourcegraph/conc`** - Structured concurrency (by Cody)
- **`golang.org/x/sync/errgroup`** - Already in go.mod

### Analysis:

- **Swiss tables:** Worth benchmarking - could be 10-20% faster for lookups
- **Rapid:** Better than stdlib fuzzing for property tests - worth trying
- **Conc:** Not needed unless we do concurrent tree building

---

## 7. Questions for Consideration

1. **Should we prioritize memory or speed?**
   - Current: Speed optimized (5.9x faster)
   - Option: Hybrid approach for memory
   - Recommendation: Measure first, then decide

2. **Is the type safety refactoring worth the churn?**
   - Pros: Better code quality
   - Cons: Requires changes across multiple packages
   - Recommendation: Do it incrementally

3. **Should we add a feature flag for the optimization?**
   - Pros: Can disable if issues found
   - Cons: More code paths to maintain
   - Recommendation: No, rely on tests

4. **What about concurrent tree building?**
   - Pros: 2-4x faster on multi-core
   - Cons: Complex algorithm, hard to debug
   - Recommendation: Only if profiling shows it's needed

---

## 8. Success Criteria

- [ ] All steps in Phase 1 complete
- [ ] Real-world benchmark shows < 200ms for art-dupl codebase
- [ ] Memory usage documented and acceptable
- [ ] No performance regressions in CI
- [ ] ADR approved/merged
- [ ] Type safety improved (TokenValue type)
- [ ] Property tests catch at least 1 edge case

---

## Next Actions

1. **Immediate:** Pick a step from Phase 1 and execute
2. **Before next optimization:** Run memory profiling
3. **Before type refactor:** Check impact across codebase
4. **Before concurrent work:** Profile to confirm bottleneck
