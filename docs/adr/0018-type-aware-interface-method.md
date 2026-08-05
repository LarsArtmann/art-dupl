# ADR-0018: Type-Aware Interface-Method Detection

**Date:** 2026-08-05

## Context

The `interface-method` actionability pattern (ADR-0004) suppresses clone groups
where every clone is a FuncDecl implementing an interface contract with a small
body (<=4 statements). Before this ADR, detection relied exclusively on a static
name list (`commonInterfaceMethodNames`) covering ~25 standard library interface
methods (`String`, `Read`, `Error`, `MarshalJSON`, etc.).

This approach has two gaps:

1. **Custom interfaces**: A project defining `type Validator interface { Validate() error }`
   with multiple implementations (`User.Validate()`, `Product.Validate()`) would not
   be suppressed because `Validate` is not in the stdlib name list.
2. **Incomplete static coverage**: The list must be manually maintained and can never
   cover all interfaces across the Go ecosystem.

Additionally, the transformer did not set `Node.Name` on FuncDecl nodes (despite
the struct docstring claiming it should), making the static name list check
ineffective in practice for real clone groups.

## Decision

Add a type-aware detection path that uses `go/types` to verify whether a FuncDecl
satisfies an interface declared in the same package. The static name list remains
as a complementary fallback for cross-package and stdlib interfaces.

### Detection Logic

`golang.IsInterfaceMethod(info *types.Info, fn *ast.FuncDecl)` performs:

1. Resolve the method object from `info.Defs[fn.Name]` as a `*types.Func`.
2. Extract the receiver type from the method signature.
3. Scan the package scope for interface types (`*types.Interface` underlying).
4. For each interface that has a method with the same name as `fn.Name.Name`,
   check `types.Implements` for both value and pointer-to receiver types.
5. Return true if any interface matches.

Embedded interface methods are handled automatically because `go/types` flattens
embedded methods in `Interface.NumMethods()`.

### Data Flow

```
Transformer (FuncDecl case, typeInfo != nil)
  → golang.IsInterfaceMethod(typeInfo, fn)
  → o.InterfaceMethod = true/false (on FuncDecl node)
  → t.enclosingInterfaceMethod = o.InterfaceMethod (save for body)
  → body statement nodes inherit flag via trans() stamp
  → syntax.Node.InterfaceMethod (on FuncDecl AND body statements)
  → serial() shallow copy
  → syntaxToCloneNode() → domain.CloneNode.InterfaceMethod
  → actionability: isInterfaceMethodBody() path 1 (FuncDecl root) OR path 2 (statement root)
```

### Statement-Level Propagation

FuncDecl nodes are never clone roots in real Go files because the structural
filter in `FindSyntaxUnits` rejects non-Statement clone roots (when the file
contains any statements, which all real Go files do). This makes the
FuncDecl-root path unreachable for standard Go code.

To fix this, the transformer propagates `InterfaceMethod` from the FuncDecl
to its body statement nodes using the same save/restore pattern as
`EnclosingReturnArity`: the FuncDecl case sets
`t.enclosingInterfaceMethod = o.InterfaceMethod` before processing the body,
and `trans()` stamps `o.InterfaceMethod = t.enclosingInterfaceMethod` on every
node. FuncLit resets the flag to `false` (closures are not interface methods).

This means the `interface-method` pattern has two paths:

- **Path 1 (FuncDecl root)**: Edge case for files without statements. Uses the
  static name list OR the `InterfaceMethod` flag. Body size checked via the
  BlockStmt child.
- **Path 2 (statement root)**: Normal Go files. Uses the propagated
  `InterfaceMethod` flag. Body size is `len(seq)` (number of clone units).
  Requires `--type-aware` mode (the flag is only set when type info is
  available).

### Bug Fix: FuncDecl Name Field

The transformer now sets `o.Name = funcName` on FuncDecl nodes. This was
documented in the `syntax.Node` struct comment but never implemented, making
the static name list check in the actionability pattern ineffective. The `Name`
field is metadata (not part of the hash or fingerprint), so this fix has no
effect on clone detection, only on metadata available to the actionability layer.

## Tradeoffs

**Same-package only**: Only interfaces declared in the same package as the FuncDecl
are checked. Stdlib interfaces (`fmt.Stringer`, `io.Reader`, etc.) and cross-package
custom interfaces are not detected by the type-aware path. The static name list
covers the most common stdlib cases. Full cross-package scanning is tracked in
ROADMAP.

**Performance**: `IsInterfaceMethod` scans all package scope names for each FuncDecl.
This is O(scope_names * methods_per_interface) per method. For typical packages
(~50 scope entries, ~5 interfaces), this is negligible compared to the go/packages
load time that dominates type-aware mode. Caching the interface-to-method-name map
per package would eliminate repeated work but is deferred until profiling
demonstrates a need.

**Precision over recall**: By requiring `types.Implements` (full interface
satisfaction), we avoid false positives where a method name coincidentally matches
an interface method but the type doesn't implement the full contract. This is more
precise than checking method name + signature alone.

## Status

Implemented. Same-package interface detection is active with `--type-aware`.
The `InterfaceMethod` flag propagates from FuncDecl to body statement nodes so
the `interface-method` pattern fires on real detected clones (statement-level
clones, not FuncDecl-level clones which are structurally unreachable in Go
files). Cross-package interface scanning remains future work (ROADMAP).
