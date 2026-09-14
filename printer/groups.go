package printer

import (
	"sort"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// GetCloneSize returns the token count of the first clone in a group.
// Returns 0 if the group is empty or has no fragments.
func GetCloneSize(group [][]*syntax.Node) int {
	if len(group) > 0 {
		return len(group[0])
	}

	return 0
}

// BuildCloneGroups builds a map of hash to clone groups from matches.
// Each fragment is trimmed of units subsumed by another unit's source range
// (see trimSubsumedUnits).
func BuildCloneGroups(duplChan <-chan syntax.Match) map[string][][]*syntax.Node {
	groups := make(map[string][][]*syntax.Node)
	for dupl := range duplChan {
		for _, frag := range dupl.Frags {
			trimmed := trimSubsumedUnits(frag)
			if len(trimmed) == 0 {
				continue
			}

			groups[dupl.Hash] = append(groups[dupl.Hash], trimmed)
		}
	}

	return groups
}

// trimSubsumedUnits drops fragment units whose byte range is strictly
// contained within another unit's range in the same fragment.
//
// Nested-statement token emission (serialNestedStatements) can produce a
// maximal match that contains both a composite statement token and statement
// tokens from inside its subtree — e.g. an if statement and the return inside
// its body when both files contain the identical guard. The nested tokens
// exist so that divergent interiors can match at statement granularity; they
// add no source range of their own. Size metrics, line ranges, and
// actionability classification want the effective source statements, so the
// redundant inner units are removed here at the single funnel point all
// output paths share.
func trimSubsumedUnits(frag []*syntax.Node) []*syntax.Node {
	if len(frag) < 2 {
		return frag
	}

	kept := make([]*syntax.Node, 0, len(frag))

	for i, unit := range frag {
		if isSubsumed(unit, frag, i) {
			continue
		}

		kept = append(kept, unit)
	}

	return kept
}

// isSubsumed reports whether frag[i]'s [Pos, End) range lies strictly inside
// another unit's range. Equal ranges are kept (they do not occur in Go
// trees, but a degenerate zero-length templ range must not be used to drop a
// real unit either).
func isSubsumed(unit *syntax.Node, frag []*syntax.Node, self int) bool {
	for j, other := range frag {
		if j == self {
			continue
		}

		if other.Pos <= unit.Pos && unit.End <= other.End &&
			(other.Pos < unit.Pos || unit.End < other.End) {
			return true
		}
	}

	return false
}

// cloneRange represents the byte range of a single clone fragment in a file.
type cloneRange struct {
	filename string
	start    int32
	end      int32
}

// EliminateOverlaps removes clone fragments that are fully contained within
// larger fragments from other groups. The suffix tree emits at every internal
// node, producing nested matches — a 50-token clone also yields 15-token
// sub-clones. This function suppresses those redundant sub-clones.
//
// Algorithm: process fragments largest-first, accept each unless its range
// is strictly contained within an already-accepted fragment's range.
func EliminateOverlaps(groups map[string][][]*syntax.Node) map[string][][]*syntax.Node {
	type fragRef struct {
		hash string
		idx  int
		rng  cloneRange
		size int
	}

	var allFrags []fragRef

	for hash, frags := range groups {
		for i, frag := range frags {
			if len(frag) == 0 {
				continue
			}

			first := frag[0]
			last := frag[len(frag)-1]
			allFrags = append(allFrags, fragRef{
				hash: hash,
				idx:  i,
				rng: cloneRange{
					filename: first.Filename,
					start:    first.Pos,
					end:      last.End,
				},
				size: len(frag),
			})
		}
	}

	sort.Slice(allFrags, func(i, j int) bool {
		if allFrags[i].size != allFrags[j].size {
			return allFrags[i].size > allFrags[j].size
		}

		return allFrags[i].rng.start < allFrags[j].rng.start
	})

	accepted := make(map[string][]cloneRange)
	rejected := make(map[string]map[int]bool)

	for _, fr := range allFrags {
		if isContained(fr.rng, accepted[fr.rng.filename]) {
			if rejected[fr.hash] == nil {
				rejected[fr.hash] = make(map[int]bool)
			}

			rejected[fr.hash][fr.idx] = true
		} else {
			accepted[fr.rng.filename] = append(accepted[fr.rng.filename], fr.rng)
		}
	}

	result := make(map[string][][]*syntax.Node)

	for hash, frags := range groups {
		rej := rejected[hash]

		var kept [][]*syntax.Node

		for i, frag := range frags {
			if !rej[i] {
				kept = append(kept, frag)
			}
		}

		if len(kept) > 0 {
			result[hash] = kept
		}
	}

	return result
}

// isContained checks if rng is strictly inside any of the accepted ranges.
// Identical ranges (same start AND end) are NOT considered contained —
// they represent different hash views of the same code and should both be kept.
func isContained(rng cloneRange, accepted []cloneRange) bool {
	for _, acc := range accepted {
		if rng.start >= acc.start && rng.end <= acc.end {
			if rng.start == acc.start && rng.end == acc.end {
				continue
			}

			return true
		}
	}

	return false
}

// ComputeUniqueCounts calculates unique file counts for each clone group.
func ComputeUniqueCounts(groups map[string][][]*syntax.Node) map[string]int {
	uniqueCounts := make(map[string]int)
	for k, v := range groups {
		uniqueCounts[k] = syntax.CountUniqueFiles(v)
	}

	return uniqueCounts
}

// SortCloneGroupKeys sorts clone group hashes based on specified criteria.
func SortCloneGroupKeys(
	keys []string,
	sortBy config.SortCriteria,
	groups map[string][][]*syntax.Node,
	uniqueCounts map[string]int,
) {
	switch sortBy {
	case config.SortByOccurrence:
		sort.Slice(keys, func(i, j int) bool {
			return len(groups[keys[i]]) > len(groups[keys[j]])
		})
	case config.SortByHash:
		sort.Strings(keys)
	case config.SortBySize:
		sort.Slice(keys, func(i, j int) bool {
			return GetCloneSize(groups[keys[i]]) > GetCloneSize(groups[keys[j]])
		})
	case config.SortByTotalTokens:
		sort.Slice(keys, func(i, j int) bool {
			// Count total tokens across all nodes in each group
			tokensI := 0

			for _, nodes := range groups[keys[i]] {
				for range nodes {
					tokensI++
				}
			}

			tokensJ := 0

			for _, nodes := range groups[keys[j]] {
				for range nodes {
					tokensJ++
				}
			}

			return tokensI > tokensJ
		})
	default:
		sort.Strings(keys)
	}
}
