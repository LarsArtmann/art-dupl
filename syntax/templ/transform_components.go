package templ

import (
	templparser "github.com/a-h/templ/parser/v2"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// transformTemplElementExpression converts a TemplElementExpression to a syntax.Node.
func (t *transformer) transformTemplElementExpression(
	tee *templparser.TemplElementExpression,
) *syntax.Node {
	if tee == nil {
		return nil
	}

	o := t.createNodeFromRange(ComponentRender, tee.Range)
	t.addChildren(o, tee.Children)

	return o
}

// transformCallTemplateExpression converts a CallTemplateExpression to a syntax.Node.
func (t *transformer) transformCallTemplateExpression(
	cte *templparser.CallTemplateExpression,
) *syntax.Node {
	if cte == nil {
		return nil
	}

	return t.createNodeFromRange(ComponentRender, cte.Range)
}

// transformChildrenExpression converts a ChildrenExpression to a syntax.Node.
func (t *transformer) transformChildrenExpression(ce *templparser.ChildrenExpression) *syntax.Node {
	if ce == nil {
		return nil
	}

	return t.createNode(ComponentChildrenExpression, 0, 0)
}

// transformScriptElement converts a ScriptElement to a syntax.Node.
func (t *transformer) transformScriptElement(se *templparser.ScriptElement) *syntax.Node {
	if se == nil {
		return nil
	}

	o := t.createNodeFromRange(ScriptElement, se.Range)
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

	return t.createNodeFromRange(Doctype, dt.Range)
}

// transformGoCode converts a GoCode to a syntax.Node.
func (t *transformer) transformGoCode(gc *templparser.GoCode) *syntax.Node {
	if gc == nil {
		return nil
	}

	return t.createNodeFromRange(RawGoBlock, gc.Expression.Range)
}

// transformStringExpression converts a StringExpression to a syntax.Node.
func (t *transformer) transformStringExpression(se *templparser.StringExpression) *syntax.Node {
	if se == nil {
		return nil
	}

	return t.createNodeFromRange(Expression, se.Expression.Range)
}

// transformRawElement converts a RawElement to a syntax.Node.
func (t *transformer) transformRawElement(re *templparser.RawElement) *syntax.Node {
	if re == nil {
		return nil
	}

	o := t.createNodeFromRange(Element, re.Range)
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

	return t.createNodeFromRange(ComponentSwitchExpressionCase, f.Range)
}
