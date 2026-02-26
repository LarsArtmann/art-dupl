package templ

import (
	templparser "github.com/a-h/templ/parser/v2"

	"github.com/LarsArtmann/art-dupl/syntax"
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
	return t.createNode(int32(nodeType), r.From.Index, r.To.Index)
}

// addChildren processes a slice of nodes and adds them as children to the parent.
func (t *transformer) addChildren(parent *syntax.Node, nodes []templparser.Node) {
	for _, child := range nodes {
		childNode := t.transformNode(child)
		if childNode != nil {
			parent.AddChildren(childNode)
		}
	}
}

// transformTemplateFile converts a templ TemplateFile to a unified syntax.Node.
func (t *transformer) transformTemplateFile(tf *templparser.TemplateFile) *syntax.Node {
	if tf == nil {
		return nil
	}

	// Create root node
	root := syntax.NewNode()
	root.Type = File
	root.Filename = t.filename
	root.Pos = 0
	root.End = int32(len(tf.Nodes)) // #nosec G115 -- File sizes bounded by int32 in practice

	// Process all top-level nodes
	for _, node := range tf.Nodes {
		child := t.transformTemplateFileNode(node)
		if child != nil {
			root.AddChildren(child)
		}
	}

	return root
}

// transformTemplateFileNode converts a TemplateFileNode to a syntax.Node.
func (t *transformer) transformTemplateFileNode(node templparser.TemplateFileNode) *syntax.Node {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *templparser.HTMLTemplate:
		return t.transformHTMLTemplate(n)
	case *templparser.CSSTemplate:
		return t.transformCSSTemplate(n)
	case *templparser.ScriptTemplate:
		return t.transformScriptTemplate(n)
	case *templparser.TemplateFileGoExpression:
		// Skip top-level Go expressions (imports, etc.)
		return nil
	default:
		return nil
	}
}

// transformHTMLTemplate converts an HTMLTemplate to a syntax.Node.
func (t *transformer) transformHTMLTemplate(tmpl *templparser.HTMLTemplate) *syntax.Node {
	if tmpl == nil {
		return nil
	}

	o := t.createNodeFromRange(ComponentDeclaration, tmpl.Range)
	t.addChildren(o, tmpl.Children)

	return o
}

// transformCSSTemplate converts a CSSTemplate to a syntax.Node.
func (t *transformer) transformCSSTemplate(css *templparser.CSSTemplate) *syntax.Node {
	if css == nil {
		return nil
	}

	o := t.createNodeFromRange(CSSDeclaration, css.Range)
	// Process CSS properties as children
	for _, prop := range css.Properties {
		propNode := t.transformCSSProperty(prop)
		if propNode != nil {
			o.AddChildren(propNode)
		}
	}

	return o
}

// transformScriptTemplate converts a ScriptTemplate to a syntax.Node.
func (t *transformer) transformScriptTemplate(script *templparser.ScriptTemplate) *syntax.Node {
	if script == nil {
		return nil
	}

	return t.createNodeFromRange(ScriptDeclaration, script.Range)
}

// transformCSSProperty converts a CSSProperty to a syntax.Node.
func (t *transformer) transformCSSProperty(prop templparser.CSSProperty) *syntax.Node {
	if prop == nil {
		return nil
	}

	o := syntax.NewNode()
	o.Type = Attribute
	o.Filename = t.filename
	o.Pos = 0
	o.End = 0

	switch p := prop.(type) {
	case *templparser.ConstantCSSProperty:
		// Constant properties don't have position info, use 0
		_ = p.Name
	case *templparser.ExpressionCSSProperty:
		if p.Value != nil {
			o.Pos = int32(p.Value.Expression.Range.From.Index) // #nosec G115
			o.End = int32(p.Value.Expression.Range.To.Index)   // #nosec G115
		}
	}

	return o
}
