package templ

import (
	"strings"

	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
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
// The callee path is hashed verbatim (callee names are API surface, like
// called functions in the Go pipeline), and the full call expression is
// parsed as Go, so arguments participate in matching: `@Card(user.Name)` and
// `@Card(person.Name)` still match (renamed local), while `@Card(a)` and
// `@Card(a, b)` no longer collapse.
func (t *transformer) transformTemplElementExpression(
	tee *templparser.TemplElementExpression,
) *syntax.Node {
	if tee == nil {
		return nil
	}

	return t.buildComponentRender(tee.Expression, tee.Range, tee.Children)
}

// transformCallTemplateExpression converts a CallTemplateExpression to a syntax.Node.
// Callee and argument handling mirrors transformTemplElementExpression.
func (t *transformer) transformCallTemplateExpression(
	cte *templparser.CallTemplateExpression,
) *syntax.Node {
	if cte == nil {
		return nil
	}

	return t.buildComponentRender(cte.Expression, cte.Range, nil)
}

// buildComponentRender constructs a ComponentRender node from a call
// expression, its source range, and optional children. The callee path is
// extracted verbatim and hashed; the full call expression is parsed as Go and
// attached as argument tokens.
func (t *transformer) buildComponentRender(
	expr templparser.Expression,
	rng templparser.Range,
	children []templparser.Node,
) *syntax.Node {
	name := extractCalleeName(expr.Value)

	node := t.createNodeFromRange(ComponentRender, rng)

	node.Name = name
	if t.mode.HashesIdentifiers() {
		node.Type = syntax.EncodeSemanticType(ComponentRender, name, true)
	}

	node.AddChildren(t.goExprNodes(golang.SnippetExpression, expr)...)

	t.addChildren(node, children)

	return node
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

// transformScriptElement converts a ScriptElement to a syntax.Node, traversing
// its attributes and its GoCode contents. Raw JavaScript text stays skipped
// (JS is not parsed).
func (t *transformer) transformScriptElement(se *templparser.ScriptElement) *syntax.Node {
	if se == nil {
		return nil
	}

	node := t.createNodeWithAttributes(ScriptElement, se.Range, se.Attributes)

	for _, content := range se.Contents {
		if content.GoCode == nil {
			continue
		}

		codeNode := t.transformGoCode(content.GoCode)
		if codeNode != nil {
			node.AddChildren(codeNode)
		}
	}

	return node
}

// transformDocType converts a DocType to a syntax.Node.
func (t *transformer) transformDocType(dt *templparser.DocType) *syntax.Node {
	if dt == nil {
		return nil
	}

	return t.createNodeFromRange(Doctype, dt.Range)
}

// transformGoCode converts a GoCode block (`{{ ... }}`) to a syntax.Node.
// The embedded Go source is parsed as statements and spliced as
// statement-marked children, making `{{ }}` blocks individually detectable.
// Sources go/parser rejects (hybrid templ markup) degrade to one opaque
// token.
func (t *transformer) transformGoCode(goCode *templparser.GoCode) *syntax.Node {
	if goCode == nil {
		return nil
	}

	o := t.createNodeFromRange(RawGoBlock, goCode.Expression.Range)
	o.AddChildren(t.goExprNodes(golang.SnippetStatements, goCode.Expression)...)

	return o
}

// transformStringExpression converts a StringExpression (`{ expr }`) to a
// syntax.Node with the expression parsed as Go, so `{ item.Name }` and
// `{ item.Title }` no longer collapse.
func (t *transformer) transformStringExpression(se *templparser.StringExpression) *syntax.Node {
	if se == nil {
		return nil
	}

	o := t.createNodeFromRange(Expression, se.Expression.Range)
	o.AddChildren(t.goExprNodes(golang.SnippetExpression, se.Expression)...)

	return o
}

// transformRawElement converts a RawElement to a syntax.Node.
func (t *transformer) transformRawElement(re *templparser.RawElement) *syntax.Node {
	if re == nil {
		return nil
	}

	return t.buildElementNode(re.Name, re.Range, re.Attributes, nil)
}
