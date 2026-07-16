package templ

import (
	"strings"

	"github.com/LarsArtmann/art-dupl/syntax"
	templparser "github.com/a-h/templ/parser/v2"
)

// createNodeWithAttributes creates a syntax.Node for the given nodeType at the
// supplied range and attaches its transformed attributes. Callers that need to
// add children should append them after this returns.
func (t *transformer) createNodeWithAttributes(
	nodeType int,
	rng templparser.Range,
	attrs []templparser.Attribute,
) *syntax.Node {
	o := t.createNodeFromRange(nodeType, rng)
	t.addAttributesToNode(attrs, o)

	return o
}

// transformTemplElementExpression converts a TemplElementExpression to a syntax.Node.
// In semantic mode, the callee name is encoded into the node Type so that
// @demoSection(...) and @display.DataTable(...) produce distinct tokens.
// Only the callee name is encoded (not arguments), so calls with the same
// callee but different arguments still match (Type-2 clone detection).
func (t *transformer) transformTemplElementExpression(
	tee *templparser.TemplElementExpression,
) *syntax.Node {
	if tee == nil {
		return nil
	}

	name := extractCalleeName(tee.Expression.Value)
	o := t.createNodeFromRange(ComponentRender, tee.Range)
	o.Name = name
	if t.semantic {
		o.Type = syntax.EncodeSemanticType(ComponentRender, name, true)
	}
	t.addChildren(o, tee.Children)

	return o
}

// transformCallTemplateExpression converts a CallTemplateExpression to a syntax.Node.
// In semantic mode, the callee name is encoded into the node Type so that
// @demoSection(...) and @display.DataTable(...) produce distinct tokens.
// Only the callee name is encoded (not arguments), so calls with the same
// callee but different arguments still match (Type-2 clone detection).
func (t *transformer) transformCallTemplateExpression(
	cte *templparser.CallTemplateExpression,
) *syntax.Node {
	if cte == nil {
		return nil
	}

	name := extractCalleeName(cte.Expression.Value)
	o := t.createNodeFromRange(ComponentRender, cte.Range)
	o.Name = name
	if t.semantic {
		o.Type = syntax.EncodeSemanticType(ComponentRender, name, true)
	}

	return o
}

// extractCalleeName extracts the callee name from a templ call expression value.
//   - "demoSection(\"foo\", \"bar\")" → "demoSection"
//   - "display.DataTable(display.DataTableProps{...})" → "display.DataTable"
//   - "display.Card" (no parens) → "display.Card"
func extractCalleeName(expr string) string {
	if idx := strings.IndexByte(expr, '('); idx > 0 {
		return expr[:idx]
	}

	return expr
}

// transformChildrenExpression converts a ChildrenExpression to a syntax.Node.
func (t *transformer) transformChildrenExpression(ce *templparser.ChildrenExpression) *syntax.Node {
	if ce == nil {
		return nil
	}

	return t.createNodeFromRange(ComponentChildrenExpression, ce.Range)
}

// transformScriptElement converts a ScriptElement to a syntax.Node.
func (t *transformer) transformScriptElement(se *templparser.ScriptElement) *syntax.Node {
	if se == nil {
		return nil
	}

	return t.createNodeWithAttributes(ScriptElement, se.Range, se.Attributes)
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

	o := t.createNodeWithAttributes(Element, re.Range, re.Attributes)
	o.Name = re.Name
	if t.semantic {
		o.Type = syntax.EncodeSemanticType(Element, re.Name, true)
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
