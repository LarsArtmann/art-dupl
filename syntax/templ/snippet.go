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
	return nodes
}

// opaqueExprNode builds a single token node hashed from the raw expression
// source. Identifier names stay out of the tree so clone-type classification
// is unaffected by unparseable fragments.
func (t *transformer) opaqueExprNode(expr templparser.Expression) *syntax.Node {
	o := t.newFileNode()
	o.Type = syntax.EncodeSemanticType(Expression, expr.Value, t.mode.HashesIdentifiers())
	t.setNodePosFromRange(o, expr.Range)

	return o
}
