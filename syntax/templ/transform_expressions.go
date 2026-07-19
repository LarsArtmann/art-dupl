package templ

import (
	"github.com/LarsArtmann/art-dupl/syntax"
	templparser "github.com/a-h/templ/parser/v2"
)

// transformIfExpression converts an IfExpression to a syntax.Node.
func (t *transformer) transformIfExpression(ie *templparser.IfExpression) *syntax.Node {
	if ie == nil {
		return nil
	}

	o := t.createNodeFromRange(ComponentIfStatement, ie.Range)
	t.addChildren(o, ie.Then)

	for _, elseIf := range ie.ElseIfs {
		t.addChildren(o, elseIf.Then)
	}

	t.addChildren(o, ie.Else)

	return o
}

// transformForExpression converts a ForExpression to a syntax.Node.
func (t *transformer) transformForExpression(fe *templparser.ForExpression) *syntax.Node {
	if fe == nil {
		return nil
	}

	return t.wrapChildren(ComponentForStatement, fe.Range, fe.Children)
}

// transformSwitchExpression converts a SwitchExpression to a syntax.Node.
func (t *transformer) transformSwitchExpression(se *templparser.SwitchExpression) *syntax.Node {
	if se == nil {
		return nil
	}

	o := t.createNodeFromRange(ComponentSwitchStatement, se.Range)
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
	return t.wrapChildren(ComponentSwitchExpressionCase, ce.Expression.Range, ce.Children)
}
