package printer

import (
	"fmt"
	"strings"

	"github.com/LarsArtmann/art-dupl/domain"
)

// TypeDivergence describes a single position where corresponding nodes across
// clone instances have different VarType values. This is the signal that a clone
// group is a generics-extraction candidate.
type TypeDivergence struct {
	Position int
	TypeA    string
	TypeB    string
}

// ClassifyGenericsCandidate examines the VarType fields across corresponding
// CloneNode positions in a clone group. When any pair of instances has
// different non-empty VarType values at the same structural position, the group
// is a generics-extraction candidate: the algorithm is identical but the
// concrete types differ, which is exactly what Go generics can eliminate.
//
// Returns (true, hint) when the group is a generics-extraction candidate.
// The hint summarizes the type differences for user-facing output.
func ClassifyGenericsCandidate(seqs [][]*domain.CloneNode) (bool, string) {
	if len(seqs) < 2 {
		return false, ""
	}

	base := flattenCloneNodes(seqs[0])

	var divergences []TypeDivergence

	for i := 1; i < len(seqs); i++ {
		other := flattenCloneNodes(seqs[i])

		maxLen := min(len(base), len(other))

		for j := range maxLen {
			a := base[j].VarType
			b := other[j].VarType

			if a == "" || b == "" {
				continue
			}

			if a != b {
				divergences = append(divergences, TypeDivergence{
					Position: j,
					TypeA:    a,
					TypeB:    b,
				})
			}
		}
	}

	if len(divergences) == 0 {
		return false, ""
	}

	return true, formatGenericsHint(divergences)
}

// formatGenericsHint produces a concise summary of the type differences.
// Shows up to 3 unique type-pair divergences to keep the hint readable.
func formatGenericsHint(divs []TypeDivergence) string {
	seen := make(map[string]bool)

	var parts []string

	for _, d := range divs {
		key := d.TypeA + " | " + d.TypeB
		if seen[key] {
			continue
		}

		seen[key] = true

		parts = append(parts, fmt.Sprintf("%s vs %s", d.TypeA, d.TypeB))

		if len(parts) >= 3 {
			break
		}
	}

	hint := "same algorithm, different types: " + strings.Join(parts, "; ")
	if len(divs) > 3 {
		hint += fmt.Sprintf(" (%d type differences total)", len(divs))
	}

	return hint
}

// flattenCloneNodes walks a CloneNode subtree in pre-order and returns a flat
// slice. Corresponding positions in flattened slices from structurally-identical
// clones represent the same AST node — so VarType comparison at the same index
// is meaningful.
func flattenCloneNodes(nodes []*domain.CloneNode) []*domain.CloneNode {
	result := make([]*domain.CloneNode, 0, len(nodes))

	for _, n := range nodes {
		result = append(result, n)
		result = append(result, flattenChildren(n)...)
	}

	return result
}

func flattenChildren(n *domain.CloneNode) []*domain.CloneNode {
	result := make([]*domain.CloneNode, 0, len(n.Children))

	for _, c := range n.Children {
		result = append(result, c)
		result = append(result, flattenChildren(c)...)
	}

	return result
}
