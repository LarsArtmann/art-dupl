# Status Report: Semantic Fingerprinting Design

**Date**: 2026-02-15 03:35\
**Topic**: Content-Aware False Positive Filtering Design\
**Status**: Design Complete, Implementation Pending

---

## Problem Statement

art-dupl currently matches code based solely on AST structure (node types), ignoring identifiers and literals. This causes false positives when different code happens to share identical structural patterns.

### Example False Positive

```go
// File 1: String() method tests (line 453)
It("should return 'unknown' for SeverityUnknown", func() {
    Expect(domain.SeverityUnknown.String()).To(Equal("unknown"))
})

// File 2: ToViolationSeverity() method tests (line 585)
It("should map Critical to 0", func() {
    Expect(domain.SeverityCritical.ToViolationSeverity()).To(Equal(0))
})
```

Both serialize to identical token sequences:

```
CallExpr → SelectorExpr → Ident → CallExpr → ...
```

Despite testing **completely different methods** (`String()` vs `ToViolationSeverity()`), they're flagged as duplicates.

---

## Root Cause Analysis

### Current Tokenization (syntax/hash_simd.go)

```go
func hashSeqFallback(nodes []*Node, buf []byte) {
    for i, node := range nodes {
        buf[i] = byte(node.Type)  // Only node TYPE is hashed
    }
}
```

| Preserved         | Ignored          |
| ----------------- | ---------------- |
| Node type (int32) | Identifier names |
| Tree structure    | Literal values   |
| Byte positions    | Method names     |
| Child count       | Call chains      |

### Why This Design?

Original intent: detect refactored code where variable names changed but logic is identical. This is valid for some use cases but produces false positives for:

1. **Test patterns**: Ginkgo `It()` blocks with identical structure
2. **CRUD operations**: `db.Save()` on different entities
3. **Error handling**: Similar `if err != nil { return err }` patterns

---

## Proposed Solutions

### Option 1: Include Identifiers in Token Stream (Recommended First Step)

**Concept**: Hash selector/method names into the token type, not just the AST node type.

**Implementation** (`syntax/golang/transform.go`):

```go
case *ast.SelectorExpr:
    o.Type = SelectorExpr
    if sel, ok := n.Sel.(*ast.Ident); ok {
        // Fold selector name into type hash
        o.Type = SelectorExpr + int32(hashIdent(sel.Name)%1000)
    }
    o.AddChildren(t.trans(n.X), t.trans(n.Sel))

case *ast.CallExpr:
    o.Type = CallExpr
    if fun, ok := n.Fun.(*ast.Ident); ok {
        o.Type = CallExpr + int32(hashIdent(fun.Name)%1000)
    }
    // ... rest unchanged
```

**Pros**:

- ~20 lines of code
- Immediately fixes Ginkgo test false positives
- Minimal performance impact

**Cons**:

- Renamed variables won't match (may miss some clones)
- Slightly larger suffix tree

---

### Option 2: Post-Match Heuristic Filtering (Recommended Second Step)

**Concept**: After structural match, verify semantic similarity.

**Implementation** (new file `syntax/semantic_filter.go`):

```go
type SemanticFilter struct {
    GenericMethods map[string]bool
}

func (sf *SemanticFilter) IsValidDuplicate(fragA, fragB []*Node) bool {
    callsA := extractMethodNames(fragA)
    callsB := extractMethodNames(fragB)

    shared := intersection(callsA, callsB)

    // Require at least one shared non-generic method
    for _, s := range shared {
        if !sf.GenericMethods[s] {
            return true
        }
    }
    return false
}
```

**Default Generic Methods**:

- `Expect`, `To`, `Equal`, `Should` (Ginkgo/Gomega)
- `if`, `return`, `error` (control flow)
- `fmt`, `log` (standard library)

**Pros**:

- Catches structural false positives
- Configurable per-project
- No changes to core algorithm

**Cons**:

- Post-processing overhead
- Requires tuning generic method list

---

### Option 3: Semantic Threshold (Config-Driven)

**Concept**: User-configurable similarity threshold.

```go
type Config struct {
    // ... existing fields ...

    // SemanticThreshold (0.0-1.0):
    // 0.0 = disabled (structural only)
    // 0.8 = require 80% semantic overlap (recommended)
    // 1.0 = exact semantic match
    SemanticThreshold float64 `json:"semanticThreshold,omitempty"`
}
```

**Algorithm**: Jaccard similarity on identifier sets

```
similarity = |identifiers_A ∩ identifiers_B| / |identifiers_A ∪ identifiers_B|
```

**Usage**:

```bash
# Strict filtering
art-dupl --semantic-threshold 0.8 ./src

# Exact matches only
art-dupl --semantic-threshold 1.0 ./src
```

---

### Option 4: Call Chain Fingerprinting

**Concept**: Store secondary hash based on method call chains.

```go
type Node struct {
    Type       int32
    CallChain  uint16  // Hash of "db.Save" or "Expect.To.Equal"
    // ... rest
}
```

Two matches are duplicates only if:

1. Same structural pattern
2. Same call chain fingerprint (or very similar)

**Pros**: Precise differentiation
**Cons**: More complex, larger node size

---

## Recommendation

**Phased Implementation**:

| Phase | Option                        | Effort     | Impact                            |
| ----- | ----------------------------- | ---------- | --------------------------------- |
| 1     | Include identifiers in tokens | ~20 lines  | High - fixes most false positives |
| 2     | Post-match heuristic filter   | ~100 lines | Medium - catches edge cases       |
| 3     | Config-driven threshold       | ~50 lines  | Low - user control                |

**Start with Phase 1** - it's the simplest change with highest impact.

---

## Files to Modify

| File                         | Change                            |
| ---------------------------- | --------------------------------- |
| `syntax/golang/transform.go` | Hash identifiers into node types  |
| `syntax/node.go`             | (Optional) Add SemanticHash field |
| `syntax/semantic_filter.go`  | NEW: Heuristic filtering          |
| `config/config.go`           | Add SemanticThreshold field       |
| `cli/flags.go`               | Add --semantic-threshold flag     |

---

## Test Cases to Add

1. **Ginkgo tests with different methods** → Should NOT match
2. **Copy-pasted code with renamed variables** → Should still match
3. **CRUD operations on different entities** → Configurable behavior
4. **Identical code in different files** → Should match

---

## Metrics to Track

- False positive rate before/after
- Detection recall (true clones found)
- Performance impact (hashing overhead)
- User satisfaction (configurability)

---

## Next Steps

1. Implement Option 1 (identifier hashing)
2. Add test cases for false positive scenarios
3. Run benchmarks on real codebases
4. Gather feedback on default behavior
5. Consider Option 2 if false positives persist

---

## References

- Current hashing: `syntax/hash_simd.go:88-92`
- AST transformation: `syntax/golang/transform.go`
- Suffix tree matching: `suffixtree/suffixtree.go`
- Configuration: `config/config.go`
