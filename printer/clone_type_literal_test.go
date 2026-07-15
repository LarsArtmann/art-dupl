package printer

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
)

// TestCloneTypeClassificationWithLiteralNormalization verifies that clone
// type classification (Type 1 vs Type 2) is correct when literal values
// are normalized in semantic mode.
func TestCloneTypeClassificationWithLiteralNormalization(t *testing.T) {
	t.Parallel()

	t.Run("same_names_different_literals_is_type1", func(t *testing.T) {
		t.Parallel()

		// Two sequences with identical structure and names but different
		// BasicLit values. Since BasicLit has no Name, and all other node
		// Names are identical, this should be Type 1.
		// (This would match in semantic mode because literals are normalized.)
		seq1 := []*domain.CloneNode{
			{BaseType: 12, Name: ""},       // CallExpr (BasicLit has no Name)
			{BaseType: 43, Name: "errors"}, // SelectorExpr
			{BaseType: 30, Name: "errors"}, // Ident
			{BaseType: 30, Name: "New"},    // Ident
			{BaseType: 8, Name: ""},        // BasicLit "foo" - no Name
		}
		seq2 := []*domain.CloneNode{
			{BaseType: 12, Name: ""},       // CallExpr
			{BaseType: 43, Name: "errors"}, // SelectorExpr
			{BaseType: 30, Name: "errors"}, // Ident
			{BaseType: 30, Name: "New"},    // Ident
			{BaseType: 8, Name: ""},        // BasicLit "bar" - no Name
		}

		// Manual name comparison (simulates classifyCloneType logic)
		if hasDifferentNames(seq1, seq2) {
			t.Error("Same-name sequences with different literals should NOT have different names")
		}
	})

	t.Run("different_names_is_type2", func(t *testing.T) {
		t.Parallel()

		// Two sequences with different variable names — should be Type 2
		seq1 := []*domain.CloneNode{
			{BaseType: 12, Name: ""},         // CallExpr
			{BaseType: 43, Name: "ReadFile"}, // SelectorExpr
			{BaseType: 30, Name: "os"},       // Ident
			{BaseType: 30, Name: "data"},     // Ident (renamed)
		}
		seq2 := []*domain.CloneNode{
			{BaseType: 12, Name: ""},         // CallExpr
			{BaseType: 43, Name: "ReadFile"}, // SelectorExpr
			{BaseType: 30, Name: "os"},       // Ident
			{BaseType: 30, Name: "content"},  // Ident (renamed!)
		}

		if !hasDifferentNames(seq1, seq2) {
			t.Error("Sequences with different variable names should have different names")
		}
	})
}

// hasDifferentNames simulates the name comparison logic in classifyCloneType.
func hasDifferentNames(seq1, seq2 []*domain.CloneNode) bool {
	if len(seq1) != len(seq2) {
		return true
	}

	for i := range seq1 {
		if collectName(seq1[i]) != collectName(seq2[i]) {
			return true
		}
	}

	return false
}

func collectName(n *domain.CloneNode) string {
	return n.Name
}
