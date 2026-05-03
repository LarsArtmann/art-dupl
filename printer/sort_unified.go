package printer

import (
	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// SortNodesByCriteria applies sorting criteria to node arrays using a unified switch.
func SortNodesByCriteria(dups [][]*syntax.Node, sortBy config.SortCriteria) [][]*syntax.Node {
	sortedDups := make([][]*syntax.Node, len(dups))
	copy(sortedDups, dups)

	switch sortBy {
	case config.SortBySize:
		sortedDups = SortClonesBySize(sortedDups)
	case config.SortByOccurrence:
		sortedDups = SortClonesByOccurrence(sortedDups)
	case config.SortByHash:
		sortedDups = SortClonesByHash(sortedDups)
	case config.SortByTotalTokens:
		sortedDups = SortClonesByTotalTokens(sortedDups)
	default:
		sortedDups = SortClonesBySize(sortedDups) // Default to size
	}

	return sortedDups
}
