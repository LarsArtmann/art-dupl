package syntax

// MinDivergentPositions is the minimum number of distinct structural positions
// that must have differing concrete types across clone instances before a
// clone family qualifies as a generics-extraction candidate. Single-position
// differences are dominated by shallow idiom noise (one differently-typed call
// result, or the same method name on different receiver types) rather than
// genuine same-algorithm-different-types duplication.
//
// This is the canonical home of the threshold: the printer package re-exports
// it for its classification hint, and the combined type-aware +
// suggest-generics mode uses it as the structural gate deciding which
// type-erased families to keep.
const MinDivergentPositions = 2

// CountTypeDivergencePositions reports how many distinct structural positions
// carry differing non-empty VarType values across the given clone fragments.
//
// Fragments are structurally identical (they were matched by the suffix tree),
// so corresponding indexes in a pre-order flattening represent the same AST
// node and VarType comparison at the same index is meaningful. Positions where
// either side has an empty VarType (type info unavailable, or not a local
// variable) are ignored. Fragment length mismatches compare only the shared
// prefix, mirroring printer.ClassifyGenericsCandidate.
//
// The combined type-aware + suggest-generics mode uses this to decide which
// type-erased families to keep: a family with fewer than
// MinDivergentPositions divergent positions is either an exact
// duplicate of a type-aware pass group (zero divergence) or shallow idiom
// noise such as the same method name on different receiver types
// (one divergence) — the false-positive class --type-aware exists to
// eliminate.
func CountTypeDivergencePositions(frags [][]*Node) int {
	if len(frags) < 2 {
		return 0
	}

	base := flattenNodes(frags[0])

	positions := make(map[int]struct{})

	for i := 1; i < len(frags); i++ {
		other := flattenNodes(frags[i])

		maxLen := min(len(base), len(other))

		for j := range maxLen {
			a := base[j].VarType
			b := other[j].VarType

			if a == "" || b == "" || a == b {
				continue
			}

			positions[j] = struct{}{}
		}
	}

	return len(positions)
}

// IsGenericsCandidateStructure reports whether a clone family has enough type
// divergence to be a generics-extraction candidate at the structural level:
// at least MinDivergentPositions distinct positions with differing
// concrete types. Presentation gates (actionability patterns, minimum line
// count) are intentionally NOT part of this check — they apply uniformly to
// every group downstream.
func IsGenericsCandidateStructure(frags [][]*Node) bool {
	return CountTypeDivergencePositions(frags) >= MinDivergentPositions
}

// flattenNodes walks a node slice and its children in pre-order, mirroring the
// traversal used by printer.ClassifyGenericsCandidate so positions align.
func flattenNodes(nodes []*Node) []*Node {
	result := make([]*Node, 0, len(nodes))

	for _, n := range nodes {
		result = append(result, n)
		result = append(result, flattenNodes(n.Children)...)
	}

	return result
}
