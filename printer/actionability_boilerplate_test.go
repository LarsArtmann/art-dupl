package printer

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func TestIsAssignWithErrorCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		seqs     [][]*domain.CloneNode
		expected bool
	}{
		{
			name:     "assign + error check if",
			seqs:     [][]*domain.CloneNode{assignWithErrorCheckSeq()},
			expected: true,
		},
		{
			name: "two error check sequences",
			seqs: [][]*domain.CloneNode{
				assignWithErrorCheckSeq(),
				assignWithErrorCheckSeq(),
			},
			expected: true,
		},
		{
			name: "only assign without error check",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.AssignStmt},
				{BaseType: golang.AssignStmt},
			}},
			expected: false,
		},
		{
			name: "single statement (not 2)",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.IfStmt},
			}},
			expected: false,
		},
		{
			name: "three statements (not 2)",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.AssignStmt},
				{BaseType: golang.IfStmt},
				{BaseType: golang.ReturnStmt},
			}},
			expected: false,
		},
		{
			name:     "empty (vacuous true, guarded by caller)",
			seqs:     [][]*domain.CloneNode{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isAssignWithErrorCheck(tt.seqs)
			if result != tt.expected {
				t.Errorf("isAssignWithErrorCheck() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsSingleCallExpression(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		seqs     [][]*domain.CloneNode
		expected bool
	}{
		{
			name: "single CallExpr",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.CallExpr},
			}},
			expected: true,
		},
		{
			name: "ExprStmt wrapping CallExpr (statement-level t.Parallel())",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.ExprStmt, Children: []*domain.CloneNode{
					{BaseType: golang.CallExpr, Children: []*domain.CloneNode{
						{BaseType: golang.Ident, Name: "t"},
					}},
				}},
			}},
			expected: true,
		},
		{
			name: "multiple ExprStmt(CallExpr) clones across files",
			seqs: [][]*domain.CloneNode{
				{{BaseType: golang.ExprStmt, Children: []*domain.CloneNode{
					{BaseType: golang.CallExpr},
				}}},
				{{BaseType: golang.ExprStmt, Children: []*domain.CloneNode{
					{BaseType: golang.CallExpr},
				}}},
			},
			expected: true,
		},
		{
			name: "ExprStmt wrapping non-CallExpr (not lone call)",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.ExprStmt, Children: []*domain.CloneNode{
					{BaseType: golang.BinaryExpr},
				}},
			}},
			expected: false,
		},
		{
			name: "ExprStmt with multiple children (not lone call)",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.ExprStmt, Children: []*domain.CloneNode{
					{BaseType: golang.CallExpr},
					{BaseType: golang.CallExpr},
				}},
			}},
			expected: false,
		},
		{
			name: "two single CallExpr clones",
			seqs: [][]*domain.CloneNode{
				{{BaseType: golang.CallExpr}},
				{{BaseType: golang.CallExpr}},
			},
			expected: true,
		},
		{
			name: "single IfStmt (not CallExpr)",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.IfStmt},
			}},
			expected: false,
		},
		{
			name: "two nodes (not single)",
			seqs: [][]*domain.CloneNode{{
				{BaseType: golang.CallExpr},
				{BaseType: golang.CallExpr},
			}},
			expected: false,
		},
		{
			name:     "empty (vacuous true, guarded by caller)",
			seqs:     [][]*domain.CloneNode{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isSingleCallExpression(tt.seqs)
			if result != tt.expected {
				t.Errorf("isSingleCallExpression() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func assignWithErrorCheckSeq() []*domain.CloneNode {
	return []*domain.CloneNode{
		{BaseType: golang.AssignStmt},
		{BaseType: golang.IfStmt, Children: []*domain.CloneNode{
			{BaseType: golang.BinaryExpr, Children: []*domain.CloneNode{
				{BaseType: golang.Ident, Name: "nil"},
			}},
			{BaseType: golang.BlockStmt, Children: []*domain.CloneNode{
				{BaseType: golang.ReturnStmt},
			}},
		}},
	}
}
