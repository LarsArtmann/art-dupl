package printer

import (
	"github.com/LarsArtmann/art-dupl/syntax"
)

// SortNodesByCriteria applies sorting criteria to node arrays using a unified switch.
func SortNodesByCriteria(dups [][]*syntax.Node, sortBy SortBy) [][]*syntax.Node {
	sortedDups := make([][]*syntax.Node, len(dups))
	copy(sortedDups, dups)

	switch sortBy {
	case SortBySize:
		sortedDups = SortClonesBySize(sortedDups)
	case SortByOccurrence:
		sortedDups = SortClonesByOccurrence(sortedDups)
	case SortByHash:
		sortedDups = SortClonesByHash(sortedDups)
	case SortByTotalTokens:
		sortedDups = SortClonesByTotalTokens(sortedDups)
	default:
		sortedDups = SortClonesBySize(sortedDups) // Default to size
	}

	return sortedDups
}
