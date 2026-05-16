package printer

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func TestEvaluateActionability(t *testing.T) {
	tests := []struct {
		name     string
		seqs     [][]*syntax.Node
		expected domain.CloneActionability
	}{
		{
			name:     "empty sequences are actionable",
			seqs:     [][]*syntax.Node{},
			expected: domain.Actionable,
		},
		{
			name: "single FuncDecl is non-actionable (interface signature)",
			seqs: [][]*syntax.Node{
				{{Type: golang.FuncDecl, Owns: 1}},
				{{Type: golang.FuncDecl, Owns: 1}},
			},
			expected: domain.NonActionable,
		},
		{
			name: "bare DeferStmt without children is non-actionable",
			seqs: [][]*syntax.Node{
				{{Type: golang.DeferStmt, Owns: 1}},
				{{Type: golang.DeferStmt, Owns: 1}},
			},
			expected: domain.NonActionable,
		},
		{
			name: "bare IfStmt without children is actionable (cannot verify pattern)",
			seqs: [][]*syntax.Node{
				{{Type: golang.IfStmt, Owns: 1}},
				{{Type: golang.IfStmt, Owns: 1}},
			},
			expected: domain.Actionable,
		},
		{
			name: "FuncDecl with body is actionable",
			seqs: [][]*syntax.Node{
				{{Type: golang.FuncDecl, Owns: 5}, {Type: golang.BlockStmt}, {Type: golang.IfStmt}, {Type: golang.ReturnStmt}},
				{{Type: golang.FuncDecl, Owns: 5}, {Type: golang.BlockStmt}, {Type: golang.IfStmt}, {Type: golang.ReturnStmt}},
			},
			expected: domain.Actionable,
		},
		{
			name: "ForStmt loop is actionable",
			seqs: [][]*syntax.Node{
				{{Type: golang.ForStmt, Owns: 1}},
				{{Type: golang.ForStmt, Owns: 1}},
			},
			expected: domain.Actionable,
		},
		{
			name: "mixed types are actionable",
			seqs: [][]*syntax.Node{
				{{Type: golang.FuncDecl, Owns: 3}, {Type: golang.AssignStmt}},
				{{Type: golang.FuncDecl, Owns: 3}, {Type: golang.AssignStmt}},
			},
			expected: domain.Actionable,
		},
		{
			name: "only one sequence with FuncDecl is still non-actionable",
			seqs: [][]*syntax.Node{
				{{Type: golang.FuncDecl, Owns: 1}},
			},
			expected: domain.NonActionable,
		},
		{
			name: "defer mu.Unlock is non-actionable",
			seqs: [][]*syntax.Node{
				{
					{
						Type: golang.DeferStmt,
						Children: []*syntax.Node{
							{
								Type: golang.CallExpr,
								Children: []*syntax.Node{
									{
										Type: golang.SelectorExpr,
										Children: []*syntax.Node{
											{Type: golang.Ident},
											{Type: golang.Ident},
										},
									},
								},
							},
						},
					},
				},
				{
					{
						Type: golang.DeferStmt,
						Children: []*syntax.Node{
							{
								Type: golang.CallExpr,
								Children: []*syntax.Node{
									{
										Type: golang.SelectorExpr,
										Children: []*syntax.Node{
											{Type: golang.Ident},
											{Type: golang.Ident},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: domain.NonActionable,
		},
		{
			name: "if err != nil { return err } is non-actionable",
			seqs: [][]*syntax.Node{
				{
					{
						Type: golang.IfStmt,
						Children: []*syntax.Node{
							{
								Type: golang.BinaryExpr,
								Children: []*syntax.Node{
									{Type: golang.Ident},
								},
							},
							{
								Type: golang.BlockStmt,
								Children: []*syntax.Node{
									{Type: golang.ReturnStmt},
								},
							},
						},
					},
				},
				{
					{
						Type: golang.IfStmt,
						Children: []*syntax.Node{
							{
								Type: golang.BinaryExpr,
								Children: []*syntax.Node{
									{Type: golang.Ident},
								},
							},
							{
								Type: golang.BlockStmt,
								Children: []*syntax.Node{
									{Type: golang.ReturnStmt},
								},
							},
						},
					},
				},
			},
			expected: domain.NonActionable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := EvaluateActionability(tc.seqs)
			if result != tc.expected {
				t.Errorf("EvaluateActionability() = %q, want %q", result, tc.expected)
			}
		})
	}
}
