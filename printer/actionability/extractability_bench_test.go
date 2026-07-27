package actionability

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// BenchmarkEvaluateActionability_SyntaxOnly measures the overhead of the
// actionability evaluation in syntax-only mode (no type info).
func BenchmarkEvaluateActionability_SyntaxOnly(b *testing.B) {
	seqs := [][]*domain.CloneNode{
		{
			{
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
			},
		},
		{
			{
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
			},
		},
	}

	b.ResetTimer()

	for range b.N {
		EvaluateActionability(seqs)
	}
}

// BenchmarkEvaluateExtractability measures the property engine overhead.
func BenchmarkEvaluateExtractability(b *testing.B) {
	seqs := [][]*domain.CloneNode{
		{
			{
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
			},
		},
	}

	b.ResetTimer()

	for range b.N {
		EvaluateExtractability(seqs)
	}
}
