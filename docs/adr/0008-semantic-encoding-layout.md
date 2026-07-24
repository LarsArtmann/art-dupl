# ADR-0008: Semantic encoding layout

## Status

Accepted

## Context

The suffix-tree algorithm operates on `syntax.Node` tokens where each node
carries a 32-bit `Type` field. For clone detection to distinguish semantically
different constructs (e.g., `+` vs `==`, `break` vs `continue`, `var` vs
`const`), this `Type` field must encode both the AST node kind AND any
distinguishing operator/token.

A naive encoding, storing the Go AST `token.Token` or `ast.ChanDir` directly
in a separate field, would require suffix-tree consumers to inspect multiple
fields and would complicate the comparison/hashing logic. The suffix tree
needs a single 32-bit integer per node for its transition map.

## Decision

Pack two dimensions into the single `int32 Type` field using a fixed-width
bit layout:

```
┌────────────────────────────┬──────────────────────┐
│  bits 31–8 (24 bits)       │  bits 7–0 (8 bits)   │
│  identifier / operator hash│  base AST node type  │
└────────────────────────────┴──────────────────────┘
```

- **Base type (low 8 bits):** the AST node category (`FuncDecl`, `CallExpr`,
  `AssignStmt`, etc.). These are the constants defined in `syntax/golang/` and
  `syntax/templ/`. Limited to 256 distinct values.

- **Identifier/operator hash (high 24 bits):** a truncated hash of the
  identifier name (`hashIdentifierFast`) or the semantic operator token
  (`encodeSemanticType`). This lets two `BinaryExpr` nodes with different
  operators (`+` vs `==`) occupy different suffix-tree transitions while
  sharing the same base type.

### Encoding rules

| Detection mode     | Identifier hash       | Operator/token encoded? |
| ------------------ | --------------------- | ----------------------- |
| Semantic (default) | Yes, alpha-normalized | Yes                     |
| Exact              | Yes, verbatim name    | Yes                     |
| Structural         | No (all zeroed)       | No                      |

### Consumer contract

Code that compares or decodes `node.Type` must use
`golang.DecodeBaseType(node.Type)` to extract the low 8 bits. **Never compare
raw `node.Type` against `golang.*` constants**, the high bits contain the
identifier hash and will not match.

### `hashSeq` alignment

`hashSeq` uses 4 bytes per node (full `int32 Type`) so the semantic encoding
is preserved in grouping hashes. Changing to a narrower type would lose the
identifier/operator dimension.

## Consequences

- **256 base-type limit:** adding new AST node types must fit within 8 bits.
  Currently ~70 are used across golang + templ, leaving comfortable headroom.

- **Cross-package collision risk:** golang and templ node-type constants share
  the same 8-bit space. See T24 (branded `NodeType`) for the mitigation path.

- **Cache format dependency:** the gob-serialized cache stores `Type` values.
  Changing the encoding layout requires a `CacheVersion` bump to invalidate
  stale entries.

- **24-bit hash truncation:** identifier collisions are possible but
  astronomically unlikely for real codebases (< 16M distinct names needed for
  a birthday-paradox collision at 50% probability).
