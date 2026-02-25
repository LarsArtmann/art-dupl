package templ

import (
	templparser "github.com/a-h/templ/parser/v2"

	"github.com/LarsArtmann/art-dupl/syntax"
)

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

	o := syntax.NewNode()
	o.Type = ComponentDeclaration
	o.Filename = t.filename
	o.Pos = int32(tmpl.Range.From.Index) // #nosec G115 -- File sizes bounded by int32 in practice
	o.End = int32(tmpl.Range.To.Index)   // #nosec G115 -- File sizes bounded by int32 in practice

	// Process children
	for _, child := range tmpl.Children {
		childNode := t.transformNode(child)
		if childNode != nil {
			o.AddChildren(childNode)
		}
	}

	return o
}

// transformCSSTemplate converts a CSSTemplate to a syntax.Node.
func (t *transformer) transformCSSTemplate(css *templparser.CSSTemplate) *syntax.Node {
	if css == nil {
		return nil
	}

	o := syntax.NewNode()
	o.Type = CSSDeclaration
	o.Filename = t.filename
	o.Pos = int32(css.Range.From.Index) // #nosec G115 -- File sizes bounded by int32 in practice
	o.End = int32(css.Range.To.Index)   // #nosec G115 -- File sizes bounded by int32 in practice

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

	o := syntax.NewNode()
	o.Type = ScriptDeclaration
	o.Filename = t.filename
	o.Pos = int32(script.Range.From.Index) // #nosec G115 -- File sizes bounded by int32 in practice
	o.End = int32(script.Range.To.Index)   // #nosec G115 -- File sizes bounded by int32 in practice

	return o
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
