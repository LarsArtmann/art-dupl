package printer

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// TestGenericsCandidacyParityWithSyntaxHelper pins the invariant that the
// printer's ClassifyGenericsCandidate structural verdict agrees with the
// syntax-level gate used by the combined type-aware + suggest-generics mode.
// The two walk the same pre-order traversal; if either changes, combined-mode
// filtering and printed hints would disagree about which families are
// generics candidates.
func TestGenericsCandidacyParityWithSyntaxHelper(t *testing.T) {
	node := func(name, varType string, children ...*syntax.Node) *syntax.Node {
		return &syntax.Node{Name: name, VarType: varType, Children: children}
	}

	trees := [][][]*syntax.Node{
		{
			[]*syntax.Node{node("v0", "int", node("v1", "string"))},
			[]*syntax.Node{node("v0", "int", node("v1", "string"))},
		},
		{
			[]*syntax.Node{node("v0", "int", node("v1", "string"))},
			[]*syntax.Node{node("v0", "int64", node("v1", "string"))},
		},
		{
			[]*syntax.Node{node("v0", "int", node("v1", "string"))},
			[]*syntax.Node{node("v0", "int64", node("v1", "[]byte"))},
		},
		{
			[]*syntax.Node{node("v0", "main.First", node("v1", "main.First"), node("v2", "int"))},
			[]*syntax.Node{node("v0", "main.Second", node("v1", "main.Second"), node("v2", "int"))},
		},
	}

	for i, frags := range trees {
		structural := syntax.IsGenericsCandidateStructure(frags)

		candidate, _ := ClassifyGenericsCandidate(ToCloneNodeSeqs(frags))

		if structural != candidate {
			t.Errorf(
				"case %d: syntax.IsGenericsCandidateStructure() = %t but printer.ClassifyGenericsCandidate() = %t",
				i, structural, candidate,
			)
		}
	}
}
