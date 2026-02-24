// Package templ provides AST parsing for templ files using the official templ parser.
// It transforms templ template syntax into unified syntax.Node for clone detection.
//
// We focus on structural elements: declarations, HTML elements, flow control, attributes.
// We skip: text content, attribute values, identifiers.
package templ

import (
	"os"
	"strings"

	templparser "github.com/a-h/templ/parser/v2"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// Node type constants for templ syntax.
// Meaningful types capture structure without content.
const (
	BadNode = iota

	// Core declarations.
	ComponentDeclaration
	CSSDeclaration
	ScriptDeclaration

	// HTML structure.
	Element
	TagStart
	TagEnd
	SelfClosingTag
	Doctype

	// Style/Script elements (structural, not content).
	StyleElement
	ScriptElement

	// Flow control.
	ComponentIfStatement
	ComponentForStatement
	ComponentSwitchStatement
	ComponentSwitchExpressionCase
	ComponentSwitchDefaultCase

	// Attributes (structural).
	Attribute
	SpreadAttributes
	ConditionalAttributeIfStatement

	// Expressions and blocks.
	Expression
	ComponentBlock
	ComponentRender
	ComponentChildrenExpression
	RawGoBlock

	// Imports.
	ComponentImport

	// File root.
	File
)

// Parse parses the given templ file and returns the unified syntax tree.
func Parse(filename string) (*syntax.Node, error) {
	node, _, err := ParseWithLineCount(filename)
	return node, err
}

// ParseWithLineCount parses the given templ file and returns the syntax tree along with the line count.
func ParseWithLineCount(filename string) (*syntax.Node, int, error) {
	content, err := os.ReadFile(filename) // #nosec G304 -- filename comes from controlled file system walk
	if err != nil {
		return nil, 0, err //nolint:wrapcheck // Pass through os.ReadFile error
	}

	return ParseBytes(filename, content)
}

// ParseBytes parses templ content and returns the syntax tree along with the line count.
func ParseBytes(filename string, content []byte) (*syntax.Node, int, error) {
	// Parse using the official templ parser
	tf, err := templparser.ParseString(string(content))
	if err != nil {
		return nil, 0, err //nolint:wrapcheck // Pass through parser error
	}

	// Transform to unified syntax tree
	t := &transformer{
		filename: filename,
	}

	node := t.transformTemplateFile(tf)

	// Count lines from content
	lineCount := strings.Count(string(content), "\n") + 1

	return node, lineCount, nil
}

type transformer struct {
	filename string
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

// transformNode converts a Node to a syntax.Node.
func (t *transformer) transformNode(node templparser.Node) *syntax.Node {
	if node == nil {
		return nil
	}

	switch n := node.(type) {
	case *templparser.Element:
		return t.transformElement(n)
	case *templparser.IfExpression:
		return t.transformIfExpression(n)
	case *templparser.ForExpression:
		return t.transformForExpression(n)
	case *templparser.SwitchExpression:
		return t.transformSwitchExpression(n)
	case *templparser.TemplElementExpression:
		return t.transformTemplElementExpression(n)
	case *templparser.CallTemplateExpression:
		return t.transformCallTemplateExpression(n)
	case *templparser.ChildrenExpression:
		return t.transformChildrenExpression(n)
	case *templparser.ScriptElement:
		return t.transformScriptElement(n)
	case *templparser.DocType:
		return t.transformDocType(n)
	case *templparser.GoCode:
		return t.transformGoCode(n)
	case *templparser.StringExpression:
		return t.transformStringExpression(n)
	case *templparser.Text:
		// Skip text content - focus on structure
		return nil
	case *templparser.Whitespace:
		// Skip whitespace
		return nil
	case *templparser.GoComment:
		// Skip comments
		return nil
	case *templparser.HTMLComment:
		// Skip HTML comments
		return nil
	case *templparser.RawElement:
		return t.transformRawElement(n)
	case *templparser.Fallthrough:
		return t.transformFallthrough(n)
	default:
		return nil
	}
}

// transformElement converts an Element to a syntax.Node.
func (t *transformer) transformElement(el *templparser.Element) *syntax.Node {
	if el == nil {
		return nil
	}

	o := syntax.NewNode()
	o.Type = Element
	o.Filename = t.filename
	o.Pos = int32(el.Range.From.Index) // #nosec G115 -- File sizes bounded by int32 in practice
	o.End = int32(el.Range.To.Index)   // #nosec G115 -- File sizes bounded by int32 in practice

	// Process attributes
	for _, attr := range el.Attributes {
		attrNode := t.transformAttribute(attr)
		if attrNode != nil {
			o.AddChildren(attrNode)
		}
	}

	// Process children
	for _, child := range el.Children {
		childNode := t.transformNode(child)
		if childNode != nil {
			o.AddChildren(childNode)
		}
	}

	return o
}

// transformAttribute converts an Attribute to a syntax.Node.
func (t *transformer) transformAttribute(attr templparser.Attribute) *syntax.Node {
	if attr == nil {
		return nil
	}

	o := syntax.NewNode()
	o.Filename = t.filename

	switch a := attr.(type) {
	case *templparser.ConstantAttribute:
		o.Type = Attribute
		o.Pos = 0 // No Range field, use 0
		o.End = 0
	case *templparser.ExpressionAttribute:
		o.Type = Attribute
		o.Pos = int32(a.Expression.Range.From.Index) // #nosec G115
		o.End = int32(a.Expression.Range.To.Index)   // #nosec G115
	case *templparser.BoolConstantAttribute:
		o.Type = Attribute
		o.Pos = 0 // No Range field, use 0
		o.End = 0
	case *templparser.BoolExpressionAttribute:
		o.Type = Attribute
		o.Pos = int32(a.Expression.Range.From.Index) // #nosec G115
		o.End = int32(a.Expression.Range.To.Index)   // #nosec G115
	case *templparser.SpreadAttributes:
		o.Type = SpreadAttributes
		o.Pos = int32(a.Expression.Range.From.Index) // #nosec G115
		o.End = int32(a.Expression.Range.To.Index)   // #nosec G115
	case *templparser.ConditionalAttribute:
		o.Type = ConditionalAttributeIfStatement
		o.Pos = int32(a.Expression.Range.From.Index) // #nosec G115
		o.End = int32(a.Expression.Range.To.Index)   // #nosec G115
	default:
		return nil
	}

	return o
}

// transformIfExpression converts an IfExpression to a syntax.Node.
func (t *transformer) transformIfExpression(ie *templparser.IfExpression) *syntax.Node {
	if ie == nil {
		return nil
	}

	o := syntax.NewNode()
	o.Type = ComponentIfStatement
	o.Filename = t.filename
	o.Pos = int32(ie.Range.From.Index) // #nosec G115 -- File sizes bounded by int32 in practice
	o.End = int32(ie.Range.To.Index)   // #nosec G115 -- File sizes bounded by int32 in practice

	// Process 'then' branch
	for _, child := range ie.Then {
		childNode := t.transformNode(child)
		if childNode != nil {
			o.AddChildren(childNode)
		}
	}

	// Process 'else if' branches
	for _, elseIf := range ie.ElseIfs {
		for _, child := range elseIf.Then {
			childNode := t.transformNode(child)
			if childNode != nil {
				o.AddChildren(childNode)
			}
		}
	}

	// Process 'else' branch
	for _, child := range ie.Else {
		childNode := t.transformNode(child)
		if childNode != nil {
			o.AddChildren(childNode)
		}
	}

	return o
}

// transformForExpression converts a ForExpression to a syntax.Node.
func (t *transformer) transformForExpression(fe *templparser.ForExpression) *syntax.Node {
	if fe == nil {
		return nil
	}

	o := syntax.NewNode()
	o.Type = ComponentForStatement
	o.Filename = t.filename
	o.Pos = int32(fe.Range.From.Index) // #nosec G115 -- File sizes bounded by int32 in practice
	o.End = int32(fe.Range.To.Index)   // #nosec G115 -- File sizes bounded by int32 in practice

	// Process children
	for _, child := range fe.Children {
		childNode := t.transformNode(child)
		if childNode != nil {
			o.AddChildren(childNode)
		}
	}

	return o
}

// transformSwitchExpression converts a SwitchExpression to a syntax.Node.
func (t *transformer) transformSwitchExpression(se *templparser.SwitchExpression) *syntax.Node {
	if se == nil {
		return nil
	}

	o := syntax.NewNode()
	o.Type = ComponentSwitchStatement
	o.Filename = t.filename
	o.Pos = int32(se.Range.From.Index) // #nosec G115 -- File sizes bounded by int32 in practice
	o.End = int32(se.Range.To.Index)   // #nosec G115 -- File sizes bounded by int32 in practice

	// Process cases
	for _, c := range se.Cases {
		caseNode := t.transformCaseExpression(c)
		if caseNode != nil {
			o.AddChildren(caseNode)
		}
	}

	return o
}

// transformCaseExpression converts a CaseExpression to a syntax.Node.
func (t *transformer) transformCaseExpression(ce templparser.CaseExpression) *syntax.Node {
	o := syntax.NewNode()
	o.Type = ComponentSwitchExpressionCase
	o.Filename = t.filename
	o.Pos = 0
	o.End = 0

	// Process case children
	for _, child := range ce.Children {
		childNode := t.transformNode(child)
		if childNode != nil {
			o.AddChildren(childNode)
		}
	}

	return o
}

// transformTemplElementExpression converts a TemplElementExpression to a syntax.Node.
func (t *transformer) transformTemplElementExpression(tee *templparser.TemplElementExpression) *syntax.Node {
	if tee == nil {
		return nil
	}

	o := syntax.NewNode()
	o.Type = ComponentRender
	o.Filename = t.filename
	o.Pos = int32(tee.Range.From.Index) // #nosec G115 -- File sizes bounded by int32 in practice
	o.End = int32(tee.Range.To.Index)   // #nosec G115 -- File sizes bounded by int32 in practice

	// Process children (for block-style component calls)
	for _, child := range tee.Children {
		childNode := t.transformNode(child)
		if childNode != nil {
			o.AddChildren(childNode)
		}
	}

	return o
}

// transformCallTemplateExpression converts a CallTemplateExpression to a syntax.Node.
func (t *transformer) transformCallTemplateExpression(cte *templparser.CallTemplateExpression) *syntax.Node {
	if cte == nil {
		return nil
	}

	o := syntax.NewNode()
	o.Type = ComponentRender
	o.Filename = t.filename
	o.Pos = int32(cte.Range.From.Index) // #nosec G115 -- File sizes bounded by int32 in practice
	o.End = int32(cte.Range.To.Index)   // #nosec G115 -- File sizes bounded by int32 in practice

	return o
}

// transformChildrenExpression converts a ChildrenExpression to a syntax.Node.
func (t *transformer) transformChildrenExpression(ce *templparser.ChildrenExpression) *syntax.Node {
	if ce == nil {
		return nil
	}

	o := syntax.NewNode()
	o.Type = ComponentChildrenExpression
	o.Filename = t.filename
	o.Pos = 0 // No Range field available
	o.End = 0

	return o
}

// transformScriptElement converts a ScriptElement to a syntax.Node.
func (t *transformer) transformScriptElement(se *templparser.ScriptElement) *syntax.Node {
	if se == nil {
		return nil
	}

	o := syntax.NewNode()
	o.Type = ScriptElement
	o.Filename = t.filename
	o.Pos = int32(se.Range.From.Index) // #nosec G115 -- File sizes bounded by int32 in practice
	o.End = int32(se.Range.To.Index)   // #nosec G115 -- File sizes bounded by int32 in practice

	// Process attributes
	for _, attr := range se.Attributes {
		attrNode := t.transformAttribute(attr)
		if attrNode != nil {
			o.AddChildren(attrNode)
		}
	}

	return o
}

// transformDocType converts a DocType to a syntax.Node.
func (t *transformer) transformDocType(dt *templparser.DocType) *syntax.Node {
	if dt == nil {
		return nil
	}

	o := syntax.NewNode()
	o.Type = Doctype
	o.Filename = t.filename
	o.Pos = int32(dt.Range.From.Index) // #nosec G115 -- File sizes bounded by int32 in practice
	o.End = int32(dt.Range.To.Index)   // #nosec G115 -- File sizes bounded by int32 in practice

	return o
}

// transformGoCode converts a GoCode to a syntax.Node.
func (t *transformer) transformGoCode(gc *templparser.GoCode) *syntax.Node {
	if gc == nil {
		return nil
	}

	o := syntax.NewNode()
	o.Type = RawGoBlock
	o.Filename = t.filename
	o.Pos = int32(gc.Expression.Range.From.Index) // #nosec G115 -- File sizes bounded by int32 in practice
	o.End = int32(gc.Expression.Range.To.Index)   // #nosec G115 -- File sizes bounded by int32 in practice

	return o
}

// transformStringExpression converts a StringExpression to a syntax.Node.
func (t *transformer) transformStringExpression(se *templparser.StringExpression) *syntax.Node {
	if se == nil {
		return nil
	}

	o := syntax.NewNode()
	o.Type = Expression
	o.Filename = t.filename
	o.Pos = int32(se.Expression.Range.From.Index) // #nosec G115 -- File sizes bounded by int32 in practice
	o.End = int32(se.Expression.Range.To.Index)   // #nosec G115 -- File sizes bounded by int32 in practice

	return o
}

// transformRawElement converts a RawElement to a syntax.Node.
func (t *transformer) transformRawElement(re *templparser.RawElement) *syntax.Node {
	if re == nil {
		return nil
	}

	o := syntax.NewNode()
	o.Type = Element
	o.Filename = t.filename
	o.Pos = int32(re.Range.From.Index) // #nosec G115 -- File sizes bounded by int32 in practice
	o.End = int32(re.Range.To.Index)   // #nosec G115 -- File sizes bounded by int32 in practice

	// Process attributes
	for _, attr := range re.Attributes {
		attrNode := t.transformAttribute(attr)
		if attrNode != nil {
			o.AddChildren(attrNode)
		}
	}

	return o
}

// transformFallthrough converts a Fallthrough to a syntax.Node.
func (t *transformer) transformFallthrough(f *templparser.Fallthrough) *syntax.Node {
	if f == nil {
		return nil
	}

	o := syntax.NewNode()
	o.Type = ComponentSwitchExpressionCase
	o.Filename = t.filename
	o.Pos = int32(f.Range.From.Index) // #nosec G115 -- File sizes bounded by int32 in practice
	o.End = int32(f.Range.To.Index)   // #nosec G115 -- File sizes bounded by int32 in practice

	return o
}
