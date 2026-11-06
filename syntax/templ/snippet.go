package templ

import (
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
	templparser "github.com/a-h/templ/parser/v2"
)

// goExprNodes parses an embedded Go expression and returns its syntax nodes.
// Positions address the real templ file. On a parse error (templ tolerates
// some hybrid sources go/parser rejects, e.g. templ markup inside GoCode) the
// expression degrades to a single opaque token hashed from the raw source, so
// detection stays conservative: identical source matches, divergent source
// does not.
func (t *transformer) goExprNodes(kind golang.SnippetKind, expr templparser.Expression) []*syntax.Node {
	if expr.Value == "" {
		return nil
	}

	nodes, err := golang.ParseSnippet(
		kind,
		expr.Value,
		t.filename,
		int(expr.Range.From.Index), // #nosec G115 -- templ file sizes bounded by int32 in practice
		t.mode,
	)
	if err != nil {
		return []*syntax.Node{t.opaqueExprNode(expr)}
	}

	// A successful parse may legitimately produce nothing (import-only
	// top-level Go code); those are dropped rather than made opaque.
	markGoSnippet(nodes)

	return nodes
}

// markGoSnippet tags every node parsed from Go source embedded in a templ file
// (component signatures and bodies, script code, embedded expressions,
// attribute expressions). Classification needs the distinction because the
// golang and templ node-type enums overlap numerically: file kind alone would
// read a Go Field (18) as a templ ConditionalAttributeIfStatement (18) and
// attach a markup fix suggestion to Go code (issue #11).
func markGoSnippet(nodes []*syntax.Node) {
	for _, n := range nodes {
		markGoSnippetNode(n)
	}
}

func markGoSnippetNode(n *syntax.Node) {
	if n == nil {
		return
	}

	n.GoSnippet = true

	for _, child := range n.Children {
		markGoSnippetNode(child)
	}
}

// opaqueExprNode builds a single token node hashed from the raw expression
// source. Identifier names stay out of the tree so clone-type classification
// is unaffected by unparseable fragments. It is Go source that failed to
// parse, so it carries the Go-snippet marker as well.
func (t *transformer) opaqueExprNode(expr templparser.Expression) *syntax.Node {
	o := t.newFileNode()
	o.Type = syntax.EncodeSemanticType(Expression, expr.Value, t.mode.HashesIdentifiers())
	o.GoSnippet = true
	t.setNodePosFromRange(o, expr.Range)

	return o
}
