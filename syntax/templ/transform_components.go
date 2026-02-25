package templ

import (
	templparser "github.com/a-h/templ/parser/v2"

	"github.com/LarsArtmann/art-dupl/syntax"
)

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
