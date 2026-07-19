package printer

import (
	"cmp"
	"slices"

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

// makeGroupComparator returns a comparison function (negative = a before b,
// zero = equal, positive = a after b) for slices.SortFunc. All four branches
// are expressed here so call sites never repeat the switch.
func makeGroupComparator[T any](sortBy config.SortCriteria, m GroupMetrics[T]) func(a, b T) int {
	switch sortBy {
	case config.SortByOccurrence:
		return func(a, b T) int { return cmp.Compare(m.Count(b), m.Count(a)) } // descending
	case config.SortByHash:
		return func(a, b T) int {
			ka, kb := m.SortKey(a), m.SortKey(b)
			// Empty keys sort last so they don't destabilise the sort
			if ka == "" && kb == "" {
				return 0
			}

			if ka == "" {
				return 1
			}

			if kb == "" {
				return -1
			}

			return cmp.Compare(ka, kb) // ascending
		}
	case config.SortByTotalTokens:
		return func(a, b T) int {
			return cmp.Compare(m.Size(b)*m.Count(b), m.Size(a)*m.Count(a)) // descending
		}
	default: // SortBySize
		return func(a, b T) int { return cmp.Compare(m.Size(b), m.Size(a)) } // descending
	}
}

// sortGroupsByCriteria sorts a slice of clone groups in-place using the
// generic comparator factory.
func sortGroupsByCriteria[T any](groups []T, sortBy config.SortCriteria, m GroupMetrics[T]) {
	slices.SortFunc(groups, makeGroupComparator(sortBy, m))
}

// SortGroupClones returns a sorted copy of the clones in group: it pulls
// group.Clones, sorts it according to the variadic sortBy, and returns the
// result so callers don't have to repeat the local-variable + sort dance.
func SortGroupClones(group domain.ProcessedCloneGroup, sortBy ...config.SortCriteria) []domain.ProcessedClone {
	clones := group.Clones
	SortProcessedClonesByCriteria(clones, ExtractSortCriteria(sortBy...))

	return clones
}

// SortProcessedClonesByCriteria sorts individual clones within a group.
// Size/TotalTokens sort by TokenCount (descending). Occurrence/Hash are
// group-level properties identical for every clone, so they are no-ops here;
// within-group ordering falls back to filename/line (see byNameAndLineProcessed).
func SortProcessedClonesByCriteria(clones []domain.ProcessedClone, sortBy config.SortCriteria) {
	switch sortBy {
	case config.SortBySize, config.SortByTotalTokens:
		slices.SortFunc(clones, func(a, b domain.ProcessedClone) int {
			return cmp.Compare(b.TokenCount, a.TokenCount) // descending
		})
	case config.SortByOccurrence, config.SortByHash:
		// Intentionally a no-op: occurrence (group size) and hash are group-level
		// attributes that are identical for every clone within a single group, so
		// sorting individual occurrences by them has no defined order.
	}
}
