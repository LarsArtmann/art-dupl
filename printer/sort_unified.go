package printer

import (
	"github.com/LarsArtmann/art-dupl/syntax"
)

// SortNodesByCriteria applies sorting criteria to node arrays using a unified switch
func SortNodesByCriteria(dups [][]*syntax.Node, sortBy string) [][]*syntax.Node {
	sortedDups := make([][]*syntax.Node, len(dups))
	copy(sortedDups, dups)

	switch sortBy {
	case "size":
		sortedDups = SortClonesBySize(sortedDups)
	case "occurrence":
		sortedDups = SortClonesByOccurrence(sortedDups)
	case "hash":
		sortedDups = SortClonesByHash(sortedDups)
	case "total-tokens":
		sortedDups = SortClonesByTotalTokens(sortedDups)
	default:
		sortedDups = SortClonesBySize(sortedDups) // Default to size
	}

	return sortedDups
}
