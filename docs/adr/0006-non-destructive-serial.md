# ADR-0006: Non-destructive serial via shallow copies

## Status

Accepted

## Context

`syntax.Serialize()` called `serial()` which destructively mutated `n.Type`
(fingerprinting statement subtrees) and `n.Owns` (child counts) on the original
AST tree nodes. This caused two problems:

1. **Non-idempotent**: Calling Serialize twice on the same tree produced different
   results because the second call read already-fingerprinted Type values.
2. **Cache aliasing**: Cached node trees shared across goroutines on cache hits
   had their Type values corrupted by concurrent serial calls.

The original plan proposed a side-table (`map[*Node]int32`) approach (Option B).
During implementation, a simpler alternative was chosen.

## Decision

Use **shallow copies** in `serial()`: each node is copied before mutation, so the
serialized stream contains copies with the correct Type/Owns values while the
original tree is never modified.

```go
func serial(n *Node, stream *[]*Node, maxChildren int) int {
    node := &Node{
        Type: n.Type, Pos: n.Pos, End: n.End, Owns: n.Owns,
        Children: n.Children, Filename: n.Filename,
        Name: n.Name, Statement: n.Statement,
    }
    *stream = append(*stream, node)
    // ... mutate node (the copy), not n (the original)
}
```

## Rationale

- **Simpler than side table**: No new type, no threading through consumers.
- **Correct**: Original tree untouched; Serialize is idempotent.
- **Memory cost**: ~40 bytes per node extra (one Node struct copy). Acceptable
  for the safety and correctness gains.
- **Consumer-transparent**: The suffix tree, hashSeq, and all downstream code
  operate on the copies in the stream without changes.

## Consequences

- `fingerprintSubtree()` reads original Type values from the source node, not
  the copy — this is correct by design.
- Cache-miss path additionally deep-clones before storing, providing full
  isolation between cached and returned slices.
- `classifyCloneType` in `printer/` was independently fixed (T1) to avoid
  Serialize entirely, walking Children directly for Name comparison.
