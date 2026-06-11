package printer

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func TestEvaluateActionability(t *testing.T) {
	runBoolTests(t, func(seqs [][]*syntax.Node) bool {
		return EvaluateActionability(seqs) == domain.NonActionable
	}, []boolTestCase{
		{name: "empty sequences are actionable", seqs: [][]*syntax.Node{}, expected: false},
		{
			name: "single FuncDecl is non-actionable (interface signature)",
			seqs: [][]*syntax.Node{
				{{Type: golang.FuncDecl, Owns: 1}},
				{{Type: golang.FuncDecl, Owns: 1}},
			},
			expected: true,
		},
		{
			name: "bare DeferStmt without children is actionable (unknown defer)",
			seqs: [][]*syntax.Node{
				{{Type: golang.DeferStmt, Owns: 1}},
				{{Type: golang.DeferStmt, Owns: 1}},
			},
			expected: false,
		},
		{
			name: "bare IfStmt without children is actionable (cannot verify pattern)",
			seqs: [][]*syntax.Node{
				{{Type: golang.IfStmt, Owns: 1}},
				{{Type: golang.IfStmt, Owns: 1}},
			},
			expected: false,
		},
		{
			name: "FuncDecl with body is actionable",
			seqs: [][]*syntax.Node{
				mustFuncDeclWithBody(),
				mustFuncDeclWithBody(),
			},
			expected: false,
		},
		{
			name: "ForStmt loop is actionable",
			seqs: [][]*syntax.Node{
				{{Type: golang.ForStmt, Owns: 1}},
				{{Type: golang.ForStmt, Owns: 1}},
			},
			expected: false,
		},
		{
			name: "mixed types are actionable",
			seqs: [][]*syntax.Node{
				{{Type: golang.FuncDecl, Owns: 3}, {Type: golang.AssignStmt}},
				{{Type: golang.FuncDecl, Owns: 3}, {Type: golang.AssignStmt}},
			},
			expected: false,
		},
		{
			name: "only one sequence with FuncDecl is still non-actionable",
			seqs: [][]*syntax.Node{
				{{Type: golang.FuncDecl, Owns: 1}},
			},
			expected: true,
		},
		{
			name: "defer mu.Unlock is non-actionable",
			seqs: [][]*syntax.Node{
				{mustDeferSelectorCall("mu", cleanupMethodName)},
				{mustDeferSelectorCall("mu", cleanupMethodName)},
			},
			expected: true,
		},
		{
			name: "defer processOrder is actionable (not RAII)",
			seqs: [][]*syntax.Node{
				{mustDeferSelectorCall("svc", processOrderMethodName)},
				{mustDeferSelectorCall("svc", processOrderMethodName)},
			},
			expected: false,
		},
		{
			name: "if err != nil { return err } is non-actionable",
			seqs: [][]*syntax.Node{
				{mustIfErrReturnNil()},
				{mustIfErrReturnNil()},
			},
			expected: true,
		},
	})
}

func mustDeferSelectorCall(receiver, method string) *syntax.Node {
	return &syntax.Node{
		Type: golang.DeferStmt,
		Children: []*syntax.Node{
			{
				Type: golang.CallExpr,
				Children: []*syntax.Node{
					{
						Type: golang.SelectorExpr,
						Name: method,
						Children: []*syntax.Node{
							{Type: golang.Ident, Name: receiver},
							{Type: golang.Ident, Name: method},
						},
					},
				},
			},
		},
	}
}

func mustIfErrReturnNil() *syntax.Node {
	return &syntax.Node{
		Type: golang.IfStmt,
		Children: []*syntax.Node{
			{
				Type: golang.BinaryExpr,
				Children: []*syntax.Node{
					{Type: golang.Ident, Name: "err"},
					{Type: golang.Ident, Name: "nil"},
				},
			},
			{
				Type: golang.BlockStmt,
				Children: []*syntax.Node{
					{Type: golang.ReturnStmt},
				},
			},
		},
	}
}

func mustFuncDeclWithBody() []*syntax.Node {
	return []*syntax.Node{
		{Type: golang.FuncDecl, Owns: 5},
		{Type: golang.BlockStmt},
		{Type: golang.IfStmt},
		{Type: golang.ReturnStmt},
	}
}
