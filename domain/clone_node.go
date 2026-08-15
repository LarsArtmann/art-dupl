package domain

// CloneNode is a minimal, immutable recursive tree representation used by the
// printer's actionability evaluation layer. It decouples pattern detection from
// syntax.Node internals (which carry mutable serialization state, positions,
// and ownership tracking that the actionability layer does not need).
//
// The bridge in printer/clone_processor.go converts []*syntax.Node into
// []*CloneNode before passing them to EvaluateActionability.
// Field ordering groups pointer/string fields first, scalar fields last,
// so scalar-only access patterns (actionability pattern matching that reads
// BaseType, InterfaceMethod, IsAlias) touch a single cache line.
type CloneNode struct {
	// Children are the direct sub-nodes in tree order.
	Children []*CloneNode

	// Name is the identifier or method name associated with this node.
	Name string

	// Filename is the source file this node originated from.
	Filename string

	// VarType is the go/types type string for Ident nodes when type-aware
	// mode is active. Empty string means no type info available. The
	// actionability layer uses this for type-aware false-positive detection.
	VarType string

	// BaseType is the 8-bit base AST node type, already decoded from the raw
	// syntax.Node.Type (which packs identifier/operator hashes into the upper
	// 24 bits). Compare against syntax/golang constants (golang.FuncDecl, etc.).
	BaseType int32

	// EnclosingReturnArity is the number of return values in the enclosing
	// function (0 for void, >0 for functions returning values). Used by the
	// property engine to detect control-flow traps where bare returns are
	// forced by the function signature (e.g., http.HandlerFunc).
	EnclosingReturnArity int32

	// InterfaceMethod is set on FuncDecl nodes when type-aware mode confirms
	// via go/types that the method satisfies an interface declared in the same
	// package. The actionability layer uses this to suppress interface-contract
	// boilerplate beyond the static stdlib name list. Always false when
	// type-aware mode is not active.
	InterfaceMethod bool

	// IsAlias is set on TypeSpec nodes when the source uses `type X = Y`
	// (alias syntax) rather than `type X Y` (named type definition). The
	// actionability layer uses this to distinguish re-export shims (aliases,
	// non-actionable) from named type definitions (potentially actionable).
	IsAlias bool
}

// HasChildren returns true if the node has at least one child.
func (n *CloneNode) HasChildren() bool {
	return len(n.Children) > 0
}
