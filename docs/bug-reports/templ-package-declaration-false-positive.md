# Bug Report: Package Declaration False Positive in `.templ` Files

**Reporter:** Lars Artmann\
**Date:** 2026-05-20\
**Severity:** Low (noise / false positive)\
**Component:** `syntax/templ/` parser & transform pipeline\
**Related:** `syntax/golang/transform.go` (import filtering pattern)\
**Status:** ✅ Fixed

---

## Summary

`art-dupl` reports `package templates` (line 1) as a code duplication clone between multiple `.templ` files. This is a **false positive** — the package declaration position was reported incorrectly due to a bug in the File root node's byte position.

## Reproduction

```bash
cd /path/to/project-with-templ-files
art-dupl -t 10 . --semantic --sort total-tokens --only templ
```

**Observed output:**

```
found 2 clones:
  internal/infrastructure/templates/capital_contribution_agreement.templ:1-1
  internal/infrastructure/templates/personal_netting_agreement.templ:1-1
```

**Files involved:**

- `capital_contribution_agreement.templ` — line 1: `package templates`
- `personal_netting_agreement.templ` — line 1: `package templates`

Both files are structurally different (different template names, completely different body content). The only overlap is the mandatory package declaration.

## Root Cause Analysis

### 1. File Root Node Position Bug (CONFIRMED — Root Cause)

In `syntax/templ/transform.go:50`:

```go
root.End = int32(len(tf.Nodes)) // BUG: node count, NOT byte position
```

The `File` root node's `End` was set to the **number of top-level nodes** (`len(tf.Nodes)`), not a byte offset. If a file has 2 top-level template declarations, `End = 2`. `ByteRangeToLines(content, 0, 2)` maps to **line 1** because bytes 0-2 is just `"pa"` from `"package templates"`.

This means any structural match that includes the `File` root node reported **line 1** as the clone location, even if the actual matching structure starts later in the file.

### 2. Package Declaration Already Correctly Excluded

The bug report's initial hypothesis about `TemplateFileGoExpression` was **incorrect**. Research into the `a-h/templ/parser/v2` library revealed:

- **`tf.Package`** is a **separate field** on `TemplateFile` — the package declaration is NEVER part of `tf.Nodes`
- `tf.Nodes` only contains: `HTMLTemplate`, `CSSTemplate`, `ScriptTemplate`, and `TemplateFileGoExpression` (Go code like imports, vars)
- The existing `return nil` for `TemplateFileGoExpression` correctly filters Go code (imports, variable declarations)

The package declaration was never in the node tree. The false positive was entirely caused by the wrong `End` position on the File root node making ALL file-level matches appear at "line 1-1".

### 3. Comparison: Go Parser Does This Correctly

In `syntax/golang/transform.go:19-22`:

```go
o.Pos, o.End = int32(
    t.fileset.File(st).Offset(st),
), int32(
    t.fileset.File(end).Offset(end),
)
```

The Go parser uses proper byte positions from `go/token.FileSet`. The templ parser's `transformer` had no access to the content length.

## Impact

- **Noise in reports:** Every `.templ` file pair with similar structure reported a false positive at "line 1-1"
- **Wasted time:** Users investigate non-actionable "clones"
- **Threshold bypass:** The clone is only 2 tokens but is reported even with `-t 10` and `-t 15` because the File root node is included in suffix tree matches

## Fix Applied

### Option C: Fix File Root Node Positions (IMPLEMENTED)

Corrected the `File` root node's `End` field to use the actual file byte length:

**`syntax/templ/parser.go`** — Added `contentLen` to transformer:

```go
type transformer struct {
    filename    string
    contentLen  int
}

t := &transformer{
    filename:   filename,
    contentLen: len(content),
}
```

**`syntax/templ/transform.go`** — Use byte length instead of node count:

```go
root.End = int32(t.contentLen) // actual byte length, not len(tf.Nodes)
```

This ensures line number mapping is accurate. When the suffix tree matches include the File root node, the reported line range now covers the actual matching content (e.g., "1-17" instead of "1-1").

### Why Option A (Package Filtering) Was NOT Needed

Research confirmed `a-h/templ/parser/v2` stores the package declaration in `tf.Package` (a separate field), NOT in `tf.Nodes`. The package declaration was never part of the syntax tree. No filtering was needed.

## Tests Added

- `TestFileRootNodeBytePosition` — verifies File root End equals content byte length
- `TestFileRootNodeEndEqualsContentLength` — table-driven test with 3 cases (single component, empty file, two components)

## Verification

```bash
# Before fix: false positive at "1-1"
art-dupl -t 10 /tmp/templ-repro --only templ
# found 2 clones:
#   /tmp/templ-repro/a.templ:1-1
#   /tmp/templ-repro/b.templ:1-1

# After fix: correct line range (or no false positives for structurally different files)
art-dupl -t 10 /tmp/templ-repro --only templ
# found 2 clones:
#   /tmp/templ-repro/a.templ:1-17
#   /tmp/templ-repro/b.templ:1-17
# (These are legitimate clones — the signatureBlock template IS structurally identical)

# Structurally different files: no false positives
art-dupl -t 10 /tmp/templ-repro-diff --only templ
# Found total 0 clone groups.
```

## Files Changed

### Initial Fix (File root node position)

- `syntax/templ/parser.go` — Added `contentLen` field to `transformer`, pass `len(content)` in `ParseBytes`
- `syntax/templ/transform.go` — Changed `root.End` from `int32(len(tf.Nodes))` to `int32(t.contentLen)`
- `syntax/templ/templ_test.go` — Added `TestFileRootNodeBytePosition` + `TestFileRootNodeEndEqualsContentLength`

### Additional Position Fixes (same root cause: hardcoded Pos=0 End=0)

During review, found 4 more node types with the same pattern — hardcoded `Pos=0, End=0` despite the upstream `a-h/templ/parser/v2` providing `Range` data:

- `syntax/templ/transform_node.go` — `ConstantAttribute`: had `Pos=0, End=0` with comment "No Range field" — `Range` field exists on upstream type
- `syntax/templ/transform_node.go` — `BoolConstantAttribute`: same issue, same wrong comment
- `syntax/templ/transform_components.go` — `ChildrenExpression`: used `createNode(type, 0, 0)` despite having `Range`
- `syntax/templ/transform_expressions.go` — `CaseExpression`: used `createNode(type, 0, 0)` despite `Expression.Range` being available

All fixed to use `setNodePosFromRange` / `createNodeFromRange` instead of hardcoded zeros.

### Additional Tests

- `syntax/templ/templ_test.go` — Added `TestNodePositionsNonZero` (walks full tree, flags any node with Pos=0 End=0)
- `bdd/templ_clone_detection_test.go` — Added "package declaration false positives" context with 2 BDD specs

## References

- `syntax/templ/transform.go:50` — File root node creation (fixed)
- `syntax/golang/transform.go:19-22` — Go parser byte position pattern (reference)
- `syntax/templ/parser.go:52-55` — transformer struct (updated)
- `pkg/position/lines.go` — byte-to-line mapping
- `a-h/templ/parser/v2/types.go:117-126` — `TemplateFile` struct (Package is separate from Nodes)
- `a-h/templ/parser/v2/templatefile.go:72-195` — TemplateFileParser (package parsed separately from Nodes)
