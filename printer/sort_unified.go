package printer

import (
	"sort"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
)

// GroupMetrics extracts the four comparable dimensions from a clone group of
// any concrete type. The factory uses these accessors so the sort-criteria
// switch lives in exactly one place.
type GroupMetrics[T any] struct {
	Size    func(T) int
	Count   func(T) int
	SortKey func(T) string // hash or filename — the ascending sort key for SortByHash
}

// makeGroupComparator returns a less-than function for the requested sort
// criteria, using the provided metric extractors. All four branches are
// expressed here so call sites never repeat the switch.
func makeGroupComparator[T any](sortBy config.SortCriteria, m GroupMetrics[T]) func(a, b T) bool {
	switch sortBy {
	case config.SortByOccurrence:
		return func(a, b T) bool { return m.Count(a) > m.Count(b) }
	case config.SortByHash:
		return func(a, b T) bool {
			ka, kb := m.SortKey(a), m.SortKey(b)
			if ka == "" || kb == "" {
				return false
			}
			return ka < kb
		}
	case config.SortByTotalTokens:
		return func(a, b T) bool { return m.Size(a)*m.Count(a) > m.Size(b)*m.Count(b) }
	default: // SortBySize
		return func(a, b T) bool { return m.Size(a) > m.Size(b) }
	}
}

// sortGroupsByCriteria sorts a slice of clone groups in-place using the
// generic comparator factory.
func sortGroupsByCriteria[T any](groups []T, sortBy config.SortCriteria, m GroupMetrics[T]) {
	less := makeGroupComparator(sortBy, m)
	sort.Slice(groups, func(i, j int) bool { return less(groups[i], groups[j]) })
}

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
