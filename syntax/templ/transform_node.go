package templ

import (
	templparser "github.com/a-h/templ/parser/v2"

	"github.com/LarsArtmann/art-dupl/syntax"
)

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
	o := t.createNodeFromRange(Element, el.Range)
	for _, attr := range el.Attributes {
		attrNode := t.transformAttribute(attr)
		if attrNode != nil {
			o.AddChildren(attrNode)
		}
	}
	t.addChildren(o, el.Children)
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
