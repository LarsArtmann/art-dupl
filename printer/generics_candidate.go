package printer

import (
	"fmt"
	"regexp"
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

// MinDivergentPositions is the minimum number of distinct structural positions
// that must have differing concrete types across clone instances before a group
// qualifies as a generics-extraction candidate. Single-position differences
// are dominated by shallow idiom noise (one differently-typed call result)
// rather than genuine same-algorithm-different-types duplication.
const MinDivergentPositions = 2

// ClassifyGenericsCandidate examines the VarType fields across corresponding
// CloneNode positions in a clone group. When instances have different
// non-empty VarType values at the same structural positions, the group is a
// generics-extraction candidate: the algorithm is identical but the concrete
// types differ, which is exactly what Go generics can eliminate.
//
// At least MinDivergentPositions distinct positions must diverge; fewer is
// treated as noise.
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

	if len(distinctPositions(divergences)) < MinDivergentPositions {
		return false, ""
	}

	return true, formatGenericsHint(divergences)
}

// distinctPositions returns the set of structural positions that appear in
// the given divergences.
func distinctPositions(divs []TypeDivergence) map[int]bool {
	positions := make(map[int]bool, len(divs))
	for _, d := range divs {
		positions[d.Position] = true
	}

	return positions
}

// pkgPathSegmentRe matches a Go import path segment followed by a slash,
// e.g. "github.com/larsartmann/" in "github.com/larsartmann/erraudit/pkg.Type".
// Used to strip package paths to just the last segment for readability.
var pkgPathSegmentRe = regexp.MustCompile(`[\w.-]+/`)

// shortenTypeString strips Go import path prefixes from type strings,
// keeping only the last path segment. For example:
//   - github.com/pkg.Type → pkg.Type
//   - []github.com/pkg.Type → []pkg.Type
//   - *github.com/pkg.Type → *pkg.Type
func shortenTypeString(s string) string {
	for {
		loc := pkgPathSegmentRe.FindStringIndex(s)
		if loc == nil {
			return s
		}

		s = s[:loc[0]] + s[loc[1]:]
	}
}

// formatGenericsHint produces a concise summary of the type differences.
// Shows up to 3 unique type-pair divergences to keep the hint readable.
// Pair ordering is canonicalized (A vs B == B vs A) so reversed pairs in
// different instance comparisons do not duplicate.
func formatGenericsHint(divs []TypeDivergence) string {
	seen := make(map[string]bool)

	var parts []string

	for _, d := range divs {
		key := divergenceKey(d)
		if seen[key] {
			continue
		}

		seen[key] = true

		parts = append(parts, fmt.Sprintf("%s vs %s", shortenTypeString(d.TypeA), shortenTypeString(d.TypeB)))

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

// divergenceKey builds a canonical dedup key for a type divergence so that
// "A vs B" and "B vs A" (observed from different instance pairings) collapse
// to a single entry.
func divergenceKey(d TypeDivergence) string {
	if d.TypeB < d.TypeA {
		return d.TypeB + " | " + d.TypeA
	}

	return d.TypeA + " | " + d.TypeB
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
