package domain

// CloneNode is a minimal, immutable recursive tree representation used by the
// printer's actionability evaluation layer. It decouples pattern detection from
// syntax.Node internals (which carry mutable serialization state, positions,
// and ownership tracking that the actionability layer does not need).
//
// The bridge in printer/clone_processor.go converts []*syntax.Node into
// []*CloneNode before passing them to EvaluateActionability.
type CloneNode struct {
	// BaseType is the 8-bit base AST node type, already decoded from the raw
	// syntax.Node.Type (which packs identifier/operator hashes into the upper
	// 24 bits). Compare against syntax/golang constants (golang.FuncDecl, etc.).
	BaseType int32

	// Name is the identifier or method name associated with this node.
	Name string

	// Filename is the source file this node originated from.
	Filename string

	// Children are the direct sub-nodes in tree order.
	Children []*CloneNode
}

// HasChildren returns true if the node has at least one child.
func (n *CloneNode) HasChildren() bool {
	return len(n.Children) > 0
}
