package printer

import (
	"sort"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
)

func SortProcessedClonesByCriteria(clones []domain.ProcessedClone, sortBy config.SortCriteria) {
	switch sortBy {
	case config.SortBySize, config.SortByTotalTokens:
		sort.Slice(clones, func(i, j int) bool {
			return clones[i].TokenCount > clones[j].TokenCount
		})
	case config.SortByOccurrence, config.SortByHash:
		// Intentionally a no-op: occurrence (group size) and hash are group-level
		// attributes that are identical for every clone within a single group, so
		// sorting individual occurrences by them has no defined order. Within-group
		// ordering falls back to filename/line (see byNameAndLineProcessed).
	}
}
