package actionability

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// processOrderMethodName is a non-RAII method name for testing.
const processOrderMethodName = "processOrder"

func TestEvaluateActionability(t *testing.T) {
	runBoolTests(t, func(seqs [][]*domain.CloneNode) bool {
		return EvaluateActionability(seqs) == domain.NonActionable
	}, []boolTestCase{
		{name: "empty sequences are actionable", seqs: [][]*domain.CloneNode{}, expected: false},
		{
			name: "single FuncDecl is non-actionable (interface signature)",
			seqs: [][]*domain.CloneNode{
				{{BaseType: golang.FuncDecl}},
				{{BaseType: golang.FuncDecl}},
			},
			expected: true,
		},
		{
			name: "bare DeferStmt without children is actionable (unknown defer)",
			seqs: [][]*domain.CloneNode{
				{{BaseType: golang.DeferStmt}},
				{{BaseType: golang.DeferStmt}},
			},
			expected: false,
		},
		{
			name: "bare IfStmt without children is actionable (cannot verify pattern)",
			seqs: [][]*domain.CloneNode{
				{{BaseType: golang.IfStmt}},
				{{BaseType: golang.IfStmt}},
			},
			expected: false,
		},
		{
			name: "FuncDecl with body is actionable",
			seqs: [][]*domain.CloneNode{
				mustFuncDeclWithBody(),
				mustFuncDeclWithBody(),
			},
			expected: false,
		},
		{
			name: "ForStmt loop is actionable",
			seqs: [][]*domain.CloneNode{
				{{BaseType: golang.ForStmt}},
				{{BaseType: golang.ForStmt}},
			},
			expected: false,
		},
		{
			name: "mixed types are actionable",
			seqs: [][]*domain.CloneNode{
				{{BaseType: golang.FuncDecl}, {BaseType: golang.AssignStmt}},
				{{BaseType: golang.FuncDecl}, {BaseType: golang.AssignStmt}},
			},
			expected: false,
		},
		{
			name: "only one sequence with FuncDecl is still non-actionable",
			seqs: [][]*domain.CloneNode{
				{{BaseType: golang.FuncDecl}},
			},
			expected: true,
		},
		{
			name: "defer mu.Unlock is non-actionable",
			seqs: [][]*domain.CloneNode{
				{mustDeferSelectorCall("mu", cleanupMethodName)},
				{mustDeferSelectorCall("mu", cleanupMethodName)},
			},
			expected: true,
		},
		{
			name: "defer processOrder is actionable (not RAII)",
			seqs: [][]*domain.CloneNode{
				{mustDeferSelectorCall("svc", processOrderMethodName)},
				{mustDeferSelectorCall("svc", processOrderMethodName)},
			},
			expected: false,
		},
		{
			name: "if err != nil { return err } is non-actionable",
			seqs: [][]*domain.CloneNode{
				{mustIfErrReturnNil()},
				{mustIfErrReturnNil()},
			},
			expected: true,
		},
		{
			name: "m.Lock(); defer m.Unlock() is non-actionable",
			seqs: [][]*domain.CloneNode{
				mustLockDeferUnlock(),
				mustLockDeferUnlock(),
			},
			expected: true,
		},
	})
}

func mustDeferSelectorCall(receiver, method string) *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.DeferStmt,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.CallExpr,
				Children: []*domain.CloneNode{
					{
						BaseType: golang.SelectorExpr,
						Name:     method,
						Children: []*domain.CloneNode{
							{BaseType: golang.Ident, Name: receiver},
							{BaseType: golang.Ident, Name: method},
						},
					},
				},
			},
		},
	}
}

func mustIfErrReturnNil() *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.IfStmt,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.BinaryExpr,
				Children: []*domain.CloneNode{
					{BaseType: golang.Ident, Name: "err"},
					{BaseType: golang.Ident, Name: "nil"},
				},
			},
			{
				BaseType: golang.BlockStmt,
				Children: []*domain.CloneNode{
					{BaseType: golang.ReturnStmt},
				},
			},
		},
	}
}

func mustFuncDeclWithBody() []*domain.CloneNode {
	return []*domain.CloneNode{
		{BaseType: golang.FuncDecl},
		{BaseType: golang.BlockStmt},
		{BaseType: golang.IfStmt},
		{BaseType: golang.ReturnStmt},
	}
}

func mustLockDeferUnlock() []*domain.CloneNode {
	return []*domain.CloneNode{
		{
			BaseType: golang.ExprStmt,
			Children: []*domain.CloneNode{
				{
					BaseType: golang.CallExpr,
					Children: []*domain.CloneNode{
						{
							BaseType: golang.SelectorExpr,
							Name:     "Lock",
							Children: []*domain.CloneNode{
								{BaseType: golang.Ident, Name: "m"},
								{BaseType: golang.Ident, Name: "Lock"},
							},
						},
					},
				},
			},
		},
		mustDeferSelectorCall("m", cleanupMethodName),
	}
}
