package templ

import (
	"strings"

	"github.com/LarsArtmann/art-dupl/syntax/golang"
	templparser "github.com/a-h/templ/parser/v2"
)

// transformIfExpression converts an IfExpression to a syntax.Node.
// The condition is parsed as Go and emitted as expression child tokens, so
// `if user.IsAdmin` and `if user.IsGuest` produce different fingerprints.
// The else-if chain is emitted as nested statement links and the else branch
// is wrapped in a statement container, mirroring the Go IfStmt layout, so
// `if X { A } else { B }` no longer matches `if X { A; B }`.
func (t *transformer) transformIfExpression(ie *templparser.IfExpression) *syntax.Node {
	if ie == nil {
		return nil
	}

	o := t.createNodeFromRange(ComponentIfStatement, ie.Range)
	o.AddChildren(t.goExprNodes(golang.SnippetIfCond, ie.Expression)...)
	t.addChildren(o, ie.Then)

	for _, elseIf := range ie.ElseIfs {
		link := t.transformElseIf(elseIf)
		if link != nil {
			o.AddChildren(link)
		}
	}

	if len(ie.Else) > 0 {
		o.AddChildren(t.elseContainer(ie.Range, ie.Else))
	}

	return o
}

// transformElseIf converts an ElseIfExpression into a statement-marked
// ComponentIfStatement link (the golang else-if chain pattern): the serializer
// emits it as its own statement token so divergent else-if interiors stay
// detectable (ADR-0023 loop-skeleton class).
func (t *transformer) transformElseIf(e templparser.ElseIfExpression) *syntax.Node {
	o := t.createNodeFromRange(ComponentIfStatement, e.Range)
	o.Statement = true
	o.AddChildren(t.goExprNodes(golang.SnippetIfCond, e.Expression)...)
	t.addChildren(o, e.Then)

	return o
}

// elseContainer wraps an else branch in an unmarked statement container so
// its statement children are emitted through the IsStatementContainer descent
// (same semantics as the Go BlockStmt else).
func (t *transformer) elseContainer(rng templparser.Range, children []templparser.Node) *syntax.Node {
	o := t.createNodeFromRange(ComponentElseStatement, rng)
	t.addChildren(o, children)

	return o
}

// transformForExpression converts a ForExpression to a syntax.Node. The
// clause (`i, item := range items`) is parsed as Go and emitted as a child
// token before the body, so loop variables normalize (Type 2) while the
// ranged expression's API surface (fields, package names) distinguishes.
func (t *transformer) transformForExpression(fe *templparser.ForExpression) *syntax.Node {
	if fe == nil {
		return nil
	}

	o := t.createNodeFromRange(ComponentForStatement, fe.Range)
	o.AddChildren(t.goExprNodes(golang.SnippetForClause, fe.Expression)...)
	t.addChildren(o, fe.Children)

	return o
}

// transformSwitchExpression converts a SwitchExpression to a syntax.Node with
// its tag parsed as Go, default cases typed distinctly, and fallthrough
// separated from case bodies.
func (t *transformer) transformSwitchExpression(se *templparser.SwitchExpression) *syntax.Node {
	if se == nil {
		return nil
	}

	o := t.createNodeFromRange(ComponentSwitchStatement, se.Range)
	o.AddChildren(t.goExprNodes(golang.SnippetSwitchTag, se.Expression)...)

	for _, c := range se.Cases {
		caseNode := t.transformCaseExpression(c)
		if caseNode != nil {
			o.AddChildren(caseNode)
		}
	}

	return o
}

// transformCaseExpression converts a CaseExpression to a syntax.Node. Case
// expressions are parsed as Go (`case "admin":` strips the keyword and
// colon); `default:` cases get their own node type so they no longer collapse
// into ordinary cases.
func (t *transformer) transformCaseExpression(ce templparser.CaseExpression) *syntax.Node {
	value := ce.Expression.Value

	o := t.createNodeFromRange(caseNodeType(value), ce.Expression.Range)

	if !isDefaultCase(value) {
		o.AddChildren(t.goExprNodes(golang.SnippetCaseList, stripCaseSyntax(value))...)
	}

	t.addChildren(o, ce.Children)

	return o
}

// isDefaultCase reports whether a case expression value is a `default:`
// label (the upstream parser keeps the keyword and trailing colon).
func isDefaultCase(value string) bool {
	return strings.TrimSpace(strings.TrimSuffix(value, ":")) == "default"
}

// caseNodeType picks the case node type: default cases are distinct from
// ordinary cases (ComponentSwitchDefaultCase existed but was never used).
func caseNodeType(value string) int {
	if isDefaultCase(value) {
		return ComponentSwitchDefaultCase
	}

	return ComponentSwitchExpressionCase
}

// stripCaseSyntax removes the `case` keyword and trailing colon the upstream
// parser keeps in CaseExpression.Expression.Value.
func stripCaseSyntax(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.TrimPrefix(trimmed, "case ")
	trimmed = strings.TrimSpace(trimmed)

	return strings.TrimSuffix(trimmed, ":")
}

// transformFallthroughStatement converts a Fallthrough to its own node type;
// previously it was typed as a case body, making it indistinguishable from
// the surrounding case.
func (t *transformer) transformFallthroughStatement(f *templparser.Fallthrough) *syntax.Node {
	if f == nil {
		return nil
	}

	return t.createNodeFromRange(ComponentFallthroughStatement, f.Range)
}
