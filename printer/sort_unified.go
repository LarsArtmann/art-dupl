package printer

import (
	"sort"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
)

func SortProcessedClonesByCriteria(clones []domain.ProcessedClone, sortBy config.SortCriteria) {
	switch sortBy {
	case config.SortBySize:
		sort.Slice(clones, func(i, j int) bool {
			return clones[i].Size > clones[j].Size
		})
	case config.SortByOccurrence, config.SortByHash:
		sort.Slice(clones, func(i, j int) bool {
			return clones[i].Filename < clones[j].Filename
		})
	case config.SortByTotalTokens:
		sort.Slice(clones, func(i, j int) bool {
			return clones[i].Size > clones[j].Size
		})
	}
}

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
