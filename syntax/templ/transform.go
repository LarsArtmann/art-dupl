package templ

import (
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
	templparser "github.com/a-h/templ/parser/v2"
)

// createNode creates a new syntax.Node with the given type and position.
func (t *transformer) createNode(nodeType int32, start, end int64) *syntax.Node {
	n := syntax.NewNode()
	n.Type = nodeType
	n.Filename = t.filename
	n.Pos = int32(start) // #nosec G115 -- File sizes bounded by int32 in practice
	n.End = int32(end)   // #nosec G115 -- File sizes bounded by int32 in practice

	return n
}

// createNodeFromRange creates a new syntax.Node from a templ Range.
func (t *transformer) createNodeFromRange(nodeType int, r templparser.Range) *syntax.Node {
	return t.createNode(
		int32(nodeType),
		r.From.Index,
		r.To.Index,
	)
}

// wrapChildren creates a node for nodeType at the given range, appends the
// transformed children, and returns it. Use this whenever a transformer
// produces nothing but a span + a child slice.
func (t *transformer) wrapChildren(nodeType int, r templparser.Range, children []templparser.Node) *syntax.Node {
	o := t.createNodeFromRange(nodeType, r)
	t.addChildren(o, children)

	return o
}

// newFileNode builds a node with the transformer filename preset, ready for
// the caller to fill in Type/Pos/End/etc. Used by every transformer that
// starts with "create a node and stamp the filename".
func (t *transformer) newFileNode() *syntax.Node {
	o := syntax.NewNode()
	o.Filename = t.filename

	return o
}

// addChildren processes a slice of nodes and adds them as children to the parent.
// Each child is marked as Statement=true so serial() fingerprints the entire
// element subtree into one composite token, matching how Go statements work.
func (t *transformer) addChildren(parent *syntax.Node, nodes []templparser.Node) {
	for _, child := range nodes {
		childNode := t.transformNode(child)
		if childNode != nil {
			childNode.Statement = true
			parent.AddChildren(childNode)
		}
	}
}

// transformTemplateFile converts a templ TemplateFile to a unified syntax.Node.
func (t *transformer) transformTemplateFile(tf *templparser.TemplateFile) *syntax.Node {
	root := t.newFileNode()

	if tf == nil {
		return nil
	}

	root.Type = File
	root.Pos = 0
	root.End = int32(t.contentLen) // #nosec G115 -- File sizes bounded by int32 in practice

	// Process all top-level nodes
	for _, node := range tf.Nodes {
		for _, child := range t.transformTemplateFileNode(node) {
			root.AddChildren(child)
		}
	}

	return root
}

// transformTemplateFileNode converts a TemplateFileNode to syntax.Nodes.
// Top-level Go expressions may contain multiple declarations, so the slice
// can hold more than one node; imports are dropped by the Go bridge.
func (t *transformer) transformTemplateFileNode(node templparser.TemplateFileNode) []*syntax.Node {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *templparser.HTMLTemplate:
		templNode := t.transformHTMLTemplate(n)
		if templNode == nil {
			return nil
		}

		return []*syntax.Node{templNode}
	case *templparser.CSSTemplate:
		cssNode := t.transformCSSTemplate(n)
		if cssNode == nil {
			return nil
		}

		return []*syntax.Node{cssNode}
	case *templparser.ScriptTemplate:
		scriptNode := t.transformScriptTemplate(n)
		if scriptNode == nil {
			return nil
		}

		return []*syntax.Node{scriptNode}
	case *templparser.TemplateFileGoExpression:
		// Top-level Go code (helper functions, vars): parse as declarations so
		// duplicated helpers in templ files are detectable. Imports are
		// boilerplate and dropped by the bridge.
		return t.goExprNodes(golang.SnippetGoFile, n.Expression)
	default:
		return nil
	}
}

// declarationName extracts the declared name from a templ signature
// (`Panel(user User)` -> `Panel`), so the declaration token hashes the name
// instead of the whole signature: renamed parameters across packages still
// match (Type 2).
func declarationName(signature string) string {
	return extractCalleeName(signature)
}

// transformHTMLTemplate converts an HTMLTemplate to a syntax.Node.
func (t *transformer) transformHTMLTemplate(tmpl *templparser.HTMLTemplate) *syntax.Node {
	if tmpl == nil {
		return nil
	}

	o := t.createNodeFromRange(ComponentDeclaration, tmpl.Range)

	// Name keeps the full signature for display and clone-type
	// classification; the hashed identifier is the declared name only.
	o.Name = tmpl.Expression.Value
	if t.mode.HashesIdentifiers() {
		o.Type = syntax.EncodeSemanticType(ComponentDeclaration, declarationName(tmpl.Expression.Value), true)
	}

	t.addChildren(o, tmpl.Children)

	return o
}

// transformCSSTemplate converts a CSSTemplate to a syntax.Node. The template
// is marked as a statement so it forms ONE composite unit; its property
// children (with encoded property names) refine the fingerprint. This also
// fixes duplicate clone-group output: previously the declaration and its
// children were separate non-statement units mapping to the same source
// range, producing identical groups twice.
func (t *transformer) transformCSSTemplate(css *templparser.CSSTemplate) *syntax.Node {
	if css == nil {
		return nil
	}

	o := t.createNodeFromRange(CSSDeclaration, css.Range)
	o.Statement = true
	o.Name = css.Expression.Value
	if t.mode.HashesIdentifiers() {
		o.Type = syntax.EncodeSemanticType(CSSDeclaration, declarationName(css.Expression.Value), true)
	}

	// Process CSS properties as children
	for _, prop := range css.Properties {
		propNode := t.transformCSSProperty(prop, css.Range)
		if propNode != nil {
			o.AddChildren(propNode)
		}
	}

	return o
}

// transformScriptTemplate converts a ScriptTemplate to a syntax.Node. The
// script name is API surface and participates in the declaration token.
func (t *transformer) transformScriptTemplate(script *templparser.ScriptTemplate) *syntax.Node {
	if script == nil {
		return nil
	}

	o := t.createNodeFromRange(ScriptDeclaration, script.Range)
	o.Name = script.Name.Value
	if t.mode.HashesIdentifiers() {
		o.Type = syntax.EncodeSemanticType(ScriptDeclaration, script.Name.Value, true)
	}

	return o
}

// transformCSSProperty converts a CSSProperty to a syntax.Node. The property
// NAME is encoded (API surface, mirroring attribute keys), so `color: red`
// and `background: blue` no longer collapse.
// parentRange is the range of the containing CSSTemplate, used as
// fallback position for ConstantCSSProperty (which has no Range in upstream).
func (t *transformer) transformCSSProperty(
	prop templparser.CSSProperty,
	parentRange templparser.Range,
) *syntax.Node {
	if prop == nil {
		return nil
	}

	o := syntax.NewNode()
	o.Filename = t.filename

	switch p := prop.(type) {
	case *templparser.ConstantCSSProperty:
		t.setAttributeKey(o, p.Name)
		// ConstantCSSProperty has no Range in upstream a-h/templ.
		// Inherit parent CSSTemplate range as best available position.
		t.setNodePosFromRange(o, parentRange)
	case *templparser.ExpressionCSSProperty:
		t.setAttributeKey(o, p.Name)

		if p.Value != nil {
			t.setNodePosFromRange(o, p.Value.Expression.Range)
			o.AddChildren(t.goExprNodes(golang.SnippetExpression, p.Value.Expression)...)
		} else {
			t.setNodePosFromRange(o, parentRange)
		}
	}

	return o
}
