package actionability

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// TestEvaluateExtractability_FormatSpecifierDifference is an integration test
// that exercises EvaluateExtractability end-to-end with CloneNode trees whose
// string literals differ only in format specifiers. The unit tests call
// isFormatSpecifierDifference directly, but this test verifies the full
// pipeline: tree walk → literal collection → format comparison → verdict.
func TestEvaluateExtractability_FormatSpecifierDifference(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		clone1Literal string
		clone2Literal string
		wantParam     bool // expected Parameterizable verdict
	}{
		{
			name:          "different format verbs (%d vs %s) → parameterizable (genuinely different behavior)",
			clone1Literal: "user %d has %d items",
			clone2Literal: "user %s has %d items",
			wantParam:     true,
		},
		{
			name:          "same specifiers, different surrounding text → not parameterizable (just data)",
			clone1Literal: "error: %s in module %s",
			clone2Literal: "warning: %s in package %s",
			wantParam:     false,
		},
		{
			name:          "different arg count (%d vs %d %d) → parameterizable",
			clone1Literal: "count: %d",
			clone2Literal: "count: %d total: %d",
			wantParam:     true,
		},
		{
			name:          "no format specifiers, different values → not parameterizable",
			clone1Literal: "hello world",
			clone2Literal: "goodbye world",
			wantParam:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Build two CloneNode sequences representing fmt.Sprintf calls
			// with string literals that differ as described.
			seq1 := []*domain.CloneNode{
				buildSprintfCall(tt.clone1Literal),
			}
			seq2 := []*domain.CloneNode{
				buildSprintfCall(tt.clone2Literal),
			}

			nodeSeqs := [][]*domain.CloneNode{seq1, seq2}

			analysis := EvaluateExtractability(nodeSeqs)

			if analysis.Parameterizable != tt.wantParam {
				t.Errorf("Parameterizable = %v, want %v (reason: %s, confidence: %.2f)",
					analysis.Parameterizable, tt.wantParam, analysis.Reason, analysis.Confidence)
			}
		})
	}
}

// buildSprintfCall creates a CloneNode tree for: fmt.Sprintf(literal, args...)
func buildSprintfCall(literal string) *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.CallExpr,
		Name:     "fmt.Sprintf",
		Children: []*domain.CloneNode{
			{BaseType: golang.BasicLit, Name: literal},
			{BaseType: golang.Ident, Name: "v0"},
			{BaseType: golang.Ident, Name: "v1"},
		},
	}
}
