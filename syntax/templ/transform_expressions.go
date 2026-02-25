package templ

import (
	templparser "github.com/a-h/templ/parser/v2"

	"github.com/LarsArtmann/art-dupl/syntax"
)

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
