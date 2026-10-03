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
	trees := [][][]*syntax.Node{
		{
			{Name: "v0", VarType: "int", Children: []*syntax.Node{
				{Name: "v1", VarType: "string"},
			}},
			{Name: "v0", VarType: "int", Children: []*syntax.Node{
				{Name: "v1", VarType: "string"},
			}},
		},
		{
			{Name: "v0", VarType: "int", Children: []*syntax.Node{
				{Name: "v1", VarType: "string"},
			}},
			{Name: "v0", VarType: "int64", Children: []*syntax.Node{
				{Name: "v1", VarType: "string"},
			}},
		},
		{
			{Name: "v0", VarType: "int", Children: []*syntax.Node{
				{Name: "v1", VarType: "string"},
			}},
			{Name: "v0", VarType: "int64", Children: []*syntax.Node{
				{Name: "v1", VarType: "[]byte"},
			}},
		},
		{
			{Name: "v0", VarType: "main.First", Children: []*syntax.Node{
				{Name: "v1", VarType: "main.First"},
				{Name: "v2", VarType: "int"},
			}},
			{Name: "v0", VarType: "main.Second", Children: []*syntax.Node{
				{Name: "v1", VarType: "main.Second"},
				{Name: "v2", VarType: "int"},
			}},
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
