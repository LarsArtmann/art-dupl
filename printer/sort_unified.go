package printer

import (
	"github.com/LarsArtmann/art-dupl/syntax"
)

const (
	sortBySize        = "size"
	sortByOccurrence  = "occurrence"
	sortByHash        = "hash"
	sortByTotalTokens = "total-tokens"
)

// SortNodesByCriteria applies sorting criteria to node arrays using a unified switch.
func SortNodesByCriteria(dups [][]*syntax.Node, sortBy string) [][]*syntax.Node {
	sortedDups := make([][]*syntax.Node, len(dups))
	copy(sortedDups, dups)

	switch sortBy {
	case sortBySize:
		sortedDups = SortClonesBySize(sortedDups)
	case sortByOccurrence:
		sortedDups = SortClonesByOccurrence(sortedDups)
	case sortByHash:
		sortedDups = SortClonesByHash(sortedDups)
	case sortByTotalTokens:
		sortedDups = SortClonesByTotalTokens(sortedDups)
	default:
		sortedDups = SortClonesBySize(sortedDups) // Default to size
	}

	return sortedDups
}
