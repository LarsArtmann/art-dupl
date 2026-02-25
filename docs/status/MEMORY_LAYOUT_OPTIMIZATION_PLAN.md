# Memory Layout Optimization & CloneID Removal Plan

**Date:** January 26, 2026
**Comprehensive Task Breakdown - Execution Ready**

## Task Prioritization Matrix

| Task                                                      | Importance | Impact | Effort | Customer Value | Priority Score | Time Est |
| --------------------------------------------------------- | ---------- | ------ | ------ | -------------- | -------------- | -------- |
| 1. Remove CloneID validation from CloneGroup.IsValid()    | 5          | 3      | 1      | 3              | 12             | 8 min    |
| 2. Remove CloneID validation from Analysis.IsValid()      | 5          | 3      | 1      | 3              | 12             | 8 min    |
| 3. Update NodeToClone() - remove ID generation            | 5          | 4      | 1      | 4              | 14             | 10 min   |
| 4. Remove CloneID from Clone struct with comment          | 5          | 4      | 1      | 4              | 14             | 5 min    |
| 5. Delete CloneID type from domain_types.go               | 5          | 3      | 1      | 3              | 12             | 10 min   |
| 6. Optimize syntax.Node layout for cache efficiency       | 4          | 5      | 2      | 5              | 16             | 12 min   |
| 7. Optimize domain.Clone layout (reorder fields)          | 4          | 5      | 2      | 5              | 16             | 10 min   |
| 8. Remove CloneID test cases                              | 5          | 2      | 1      | 2              | 10             | 10 min   |
| 9. Implement string interning pool for filenames          | 3          | 4      | 3      | 4              | 14             | 12 min   |
| 10. Optimize suffixtree.state/tran layout                 | 3          | 4      | 2      | 4              | 13             | 10 min   |
| 11. Create Structure-of-Arrays (SoA) for Node.Type fields | 2          | 5      | 4      | 4              | 15             | 12 min   |
| 12. Add SIMD alignment helpers                            | 2          | 4      | 3      | 3              | 12             | 10 min   |
| 13. Benchmark memory layout improvements                  | 3          | 3      | 2      | 3              | 11             | 10 min   |
| 14. Run full test suite verification                      | 5          | 5      | 1      | 5              | 16             | 5 min    |

**Scoring Formula:** (Importance + Impact + Customer Value) - Effort

---

## Phase 1: CloneID Removal (Critical - Foundation Work)

### Task 1: Remove CloneID validation from CloneGroup.IsValid()

**File:** `domain/clone.go:148`  
**Changes:**

- Remove `if cg.ID == ""` check
- Update error message: remove `cg.ID` from `fmt.Errorf()`
  **Risk:** Low - pure validation logic  
  **Impact:** Eliminates unnecessary string comparison

### Task 2: Remove CloneID validation from Analysis.IsValid()

**File:** `domain/clone.go:229`  
**Changes:**

- Remove `if a.ID == ""` check
- Update error message: remove `a.ID` from `fmt.Errorf()`
  **Risk:** Low - pure validation logic

### Task 3: Update NodeToClone() - remove ID generation

**File:** `domain/clone.go:375`  
**Changes:**

- Delete: `cloneID, _ := NewCloneID(fmt.Sprintf(...))`
- Remove `ID: cloneID` from return struct
  **Risk:** Low - isolated function
  **Impact:** Removes atomic operation, string allocation, and fmt.Sprintf

### Task 4: Remove CloneID from Clone struct with comment

**File:** `domain/clone.go:99`  
**Changes:**

```go
type Clone struct {
	// NOTE: Intentionally no ID field. IDs are not needed for deduplication
	// (hash serves that purpose) and only add memory/alloc overhead. If IDs
	// are needed for external systems, generate them at export time.
	Filename   Filepath            `json:"filename"`
	// ... rest of fields
}
```

**Risk:** Medium - JSON API change (removes "id" field)
**Impact:** Saves 40-60 bytes per Clone, eliminates heap allocation

### Task 5: Delete CloneID type from domain_types.go

**File:** `domain/domain_types.go:85-112`  
**Changes:**

- Remove `type CloneID string`
- Remove `NewCloneID()` function
- Remove `String()`, `MarshalJSON()`, `UnmarshalJSON()` methods
- Remove import if now unused
  **Risk:** Low - internal type

### Task 8: Remove CloneID test cases

**File:** `domain/domain_types_test.go`  
**Changes:**

- Remove `TestCloneID_NewCloneID`
- Remove `TestCloneID_String`
- Remove `TestCloneID` suite
- Remove CloneID from test data
  **Risk:** Low - test-only changes

---

## Phase 2: Core Memory Layout Optimization (High Impact)

### Task 6: Optimize syntax.Node layout for cache efficiency

**File:** `syntax/syntax.go:20-26`  
**Current layout (padded):**

```go
type Node struct {
	Type     int       // 8B
	Filename string    // 16B + heap alloc
	Pos      int       // 8B
	End      int       // 8B
	Children []*Node   // 8B + slice alloc
	Owns     int       // 8B
}
// Total: ~56B + 2 heap allocs
```

**Optimized layout:**

```go
type Node struct {
	Type     int32         // 4B (reduced from int)
	Pos      int32         // 4B
	End      int32         // 4B
	Owns     int32         // 4B
	Filename StringID      // 4B (interned string index)
	Children []NodeIndex   // 8B (slices of indices, not pointers)
}
// Total: 28B + 1 heap alloc
```

**Changes:**

- Change int → int32 where safe
- Convert Filename to StringID (interned)
- Convert Children from pointers to indices
- Reorder for alignment: 4B fields first, 8B at end
  **Risk:** Medium - affects tree traversal logic
  **Impact:** 50% size reduction, better cache locality

### Task 7: Optimize domain.Clone layout (reorder fields)

**File:** `domain/clone.go:99-111`  
**Current layout (with padding):**

```go
type Clone struct {
	ID         CloneID         // 16B (REMOVED)
	Filename   Filepath        // 16B
	StartLine  LineNumber      // 8B
	EndLine    LineNumber      // 8B
	StartPos   BytePosition    // 8B
	EndPos     BytePosition    // 8B
	Fragment   string          // 16B
	Hash       Hash            // 16B
	Confidence Confidence      // 8B
	Complexity ComplexityScore // 8B
	Status     FileProcessingState // 8B
}
// Total: ~112B with multiple heap allocs
```

**Optimized layout:**

```go
type Clone struct {
	StartLine  LineNumber          // 8B
	EndLine    LineNumber          // 8B
	StartPos   BytePosition        // 8B
	EndPos     BytePosition        // 8B
	Complexity ComplexityScore     // 8B
	Confidence Confidence          // 8B
	Status     FileProcessingState // 8B
	Filename   Filepath            // 16B
	Fragment   string              // 16B
	Hash       Hash                // 16B
}
// Total: ~80B + 3 heap allocs (down from 112B)
```

**Changes:**

- Group 8B fields first (no padding between them)
- 16B string headers at end
- Reduces padding waste from 24B to 0B
  **Risk:** Medium - field reordering may break some tests
  **Impact:** ~30% size reduction, better sequential scan performance

### Task 10: Optimize suffixtree.state/tran layout

**Files:** `suffixtree/suffixtree.go:170-199`, `suffixtree/findtran_simd.go`  
**Current layout:**

```go
type state struct {
	tree      *STree     // 8B
	trans     []*tran    // 8B + slice alloc
	linkState *state     // 8B
}
type tran struct {
	start, end Pos   // 8B (4B each)
	state      *state // 8B
}
```

**Optimized layout:**

```go
type state struct {
	tree      STreeIndex    // 4B (index not pointer)
	linkState StateIndex    // 4B (index not pointer)
	trans     []TranIndex   // 8B
}
type tran struct {
	start Pos        // 4B
	end   Pos        // 4B
	state StateIndex // 4B
	_     [4]byte   // padding for alignment
}
```

**Changes:**

- Convert pointers to indices (4B vs 8B)
- Add explicit padding for SIMD alignment
- Consider slice of values instead of pointers
  **Risk:** High - pervasive change affects all suffix tree ops
  **Impact:** 33% reduction in state/tran size, better cache locality for SIMD

---

## Phase 3: Advanced Optimizations (SIMD-Ready)

### Task 9: Implement string interning pool for filenames

**Files:** New `domain/stringpool.go`, `domain/clone.go`, `syntax/syntax.go`  
**Implementation:**

```go
type StringID uint32

var filenamePool = &StringInternPool{}

type StringInternPool struct {
	mu    sync.RWMutex
	pool  map[string]StringID
	strings []string
	nextID StringID
}

func (p *StringInternPool) Get(s string) StringID {
	// Return existing or allocate new ID
}
```

**Changes:**

- Replace bare strings with StringID indices
- Single copy of each unique filename in memory
- Use in Clone.Filename and Node.Filename
  **Risk:** Medium - requires thread-safe implementation
  **Impact:** Eliminates duplicate filename allocations (often 60-80% duplicates)
  **Customer Value:** High - reduces memory 20-40% on typical projects

### Task 11: Create Structure-of-Arrays (SoA) for Node.Type fields

**Files:** New `syntax/node_soa.go`  
**Implementation:**

```go
type NodeBatch struct {
	Types []int32      // contiguous for SIMD
	Poss []int32
	Ends []int32
	Owns []int32
	FilenameIDs []StringID
	Children [][]NodeIndex
}
```

**Changes:**

- For batch operations (hashing, comparison), use SoA
- Extract Type fields into contiguous slice
- SIMD can load 4-8 Types per instruction
  **Risk:** High - new data structure pattern
  **Impact:** 3-5x speedup for hashing when SIMD enabled

### Task 12: Add SIMD alignment helpers

**Files:** `internal/simd/simd.go` (extend)  
**Implementation:**

```go
func AlignedAlloc(size int, alignment int) []byte {
	// Use mmap or special allocation with alignment guarantees
}

func PadToVectorSize(n int) int {
	vectorSize := VectorSize()
	return (n + vectorSize - 1) / vectorSize * vectorSize
}
```

**Changes:**

- Ensure data allocated on SIMD boundaries
- Pad arrays to multiple of vector size
- Prevents cache line splits during SIMD loads
  **Risk:** Low - additive utility functions
  **Impact:** 10-20% SIMD performance improvement

---

## Phase 4: Verification & Testing

### Task 13: Benchmark memory layout improvements

**Files:** New benchmarks for:

- Node allocation/traversal
- Clone memory footprint
- Suffix tree build performance
- Cache miss rates (using pprof)

**Metrics to track:**

- Memory per Clone (target: 80B → 48B)
- Cache misses per 1000 clones
- Time to process 10K nodes
- Suffix tree construction time

### Task 14: Run full test suite verification

**Command:** `make test` (all packages)  
**Coverage checks:**

- domain package tests
- syntax package tests
- suffixtree package tests
- JSON serialization/deserialization

**Risk assessment:**

- JSON API removes "id" field → breaking change for external consumers
- Field reordering → may break reflection-based tests
- String interning → must maintain thread safety

---

## Expected Outcomes

### Memory Savings

| Struct | Before | After | Savings | % Reduction |
| ------ | ------ | ----- | ------- | ----------- |
| Clone  | 112B   | 48B   | 64B     | 57%         |
| Node   | 56B    | 28B   | 28B     | 50%         |
| state  | 24B    | 16B   | 8B      | 33%         |
| tran   | 24B    | 16B   | 8B      | 33%         |

**Total per 10,000 clones:** ~900KB saved

### Performance Improvements

| Operation          | Before             | After       | Improvement |
| ------------------ | ------------------ | ----------- | ----------- |
| Clone allocation   | 2 allocs           | 0 allocs    | ∞ (no heap) |
| Suffix tree search | Baseline           | +SIMD       | 2-3x faster |
| Hash computation   | 96% alloc overhead | 5% overhead | 20x less GC |
| Cache miss rate    | 15-20%             | 5-8%        | 3x better   |

### Customer Value

- ✅ Process larger codebases without OOM
- ✅ 50-70% faster analysis on large projects
- ✅ Lower memory usage in CI/CD pipelines
- ✅ Better SIMD utilization when Go 1.28 arrive

---

## Implementation Order (Strict)

1. **Foundation First**: Remove CloneID (Tasks 1-5, 8)
2. **Layout Optimization**: Reorder fields (Tasks 6, 7, 10)
3. **Advanced Features**: SIMD prep (Tasks 9, 11, 12)
4. **Verification**: Benchmarks and tests (Tasks 13, 14)

**Total Estimated Time:** ~2-3 hours  
**Risk Level:** Medium (breaking JSON API change)  
**Rollback Plan:** Git revert + re-run tests
