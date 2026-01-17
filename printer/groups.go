package printer

import (
	"sort"

	"github.com/LarsArtmann/art-dupl/internal/utils"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// GetCloneSize returns the size (token count) of the first clone in a group.
// Returns 0 if the group is empty or has no fragments.
func GetCloneSize(group [][]*syntax.Node) int {
	if len(group) > 0 && len(group[0]) > 0 {
		return group[0][0].Owns
	}
	return 0
}

// BuildCloneGroups builds a map of hash to clone groups from matches.
func BuildCloneGroups(duplChan <-chan syntax.Match) map[string][][]*syntax.Node {
	groups := make(map[string][][]*syntax.Node)
	for dupl := range duplChan {
		groups[dupl.Hash] = append(groups[dupl.Hash], dupl.Frags...)
	}
	return groups
}

// ComputeUniqueCounts calculates unique file counts for each clone group.
func ComputeUniqueCounts(groups map[string][][]*syntax.Node) map[string]int {
	uniqueCounts := make(map[string]int)
	for k, v := range groups {
		uniqueCounts[k] = len(utils.Unique(v))
	}
	return uniqueCounts
}

// SortCloneGroupKeys sorts clone group hashes based on specified criteria.
func SortCloneGroupKeys(keys []string, sortBy SortBy, groups map[string][][]*syntax.Node, uniqueCounts map[string]int) {
	switch sortBy {
	case SortByOccurrence:
		sort.Slice(keys, func(i, j int) bool {
			return uniqueCounts[keys[i]] > uniqueCounts[keys[j]]
		})
	case SortByHash:
		sort.Strings(keys)
	case SortBySize:
		sort.Slice(keys, func(i, j int) bool {
			return GetCloneSize(groups[keys[i]]) > GetCloneSize(groups[keys[j]])
		})
	case SortByTotalTokens:
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
