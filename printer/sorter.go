package printer

import (
	"sort"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// compareByNameThenPos compares two items by name, then by position for stable ordering.
func compareByNameThenPos(nameI, nameJ string, posI, posJ int) bool {
	if nameI == nameJ {
		return posI < posJ
	}

	return nameI < nameJ
}

// sortCloneGroupsBySizeDescending sorts CloneGroup slice by size (largest first, descending).
func sortCloneGroupsBySizeDescending(groups []CloneGroup) {
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Size > groups[j].Size
	})
}

// SortCloneGroups sorts CloneGroup arrays by specified criteria.
func SortCloneGroups(groups []CloneGroup, sortBy SortBy) {
	switch sortBy {
	case SortBySize:
		sortCloneGroupsBySizeDescending(groups)
	case SortByOccurrence:
		sort.Slice(groups, func(i, j int) bool {
			return len(groups[i].Files) > len(groups[j].Files) // Most files first (descending)
		})
	case SortByHash:
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].Hash < groups[j].Hash // Alphabetical (ascending)
		})
	case SortByTotalTokens:
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].Size*len(groups[i].Files) > groups[j].Size*len(groups[j].Files)
		})
	default:
		// Default to size sorting for highest impact
		sortCloneGroupsBySizeDescending(groups)
	}
}

// isEmptyOrLessThanEmpty returns true if:
// - both slices are empty (equal, return false)
// - i is non-empty and j is empty (i should come first, return true)
// - i is empty and j is non-empty (i should come last, return false).
func isEmptyOrLessThanEmpty[T any](slices [][]T, i, j int) (bool, bool) {
	if len(slices[i]) == 0 && len(slices[j]) == 0 {
		return true, false // Both empty, they're equal
	}

	if len(slices[i]) == 0 {
		return true, false // i is empty, j is not - i should come after j
	}

	if len(slices[j]) == 0 {
		return true, true // j is empty, i is not - i should come before j
	}

	return false, false // Neither is empty, continue with normal comparison
}

// SortClonesBySize sorts clone groups by token count (largest first, descending order).
func SortClonesBySize(dups [][]*syntax.Node) [][]*syntax.Node {
	sort.Slice(dups, func(i, j int) bool {
		if handled, lessThan := isEmptyOrLessThanEmpty(dups, i, j); handled {
			return lessThan
		}
		// Calculate size as end position minus start position
		sizeI := dups[i][len(dups[i])-1].End - dups[i][0].Pos
		sizeJ := dups[j][len(dups[j])-1].End - dups[j][0].Pos

		return sizeI > sizeJ
	})

	return dups
}

// SortClonesByOccurrence sorts clone groups by number of files (most files first, descending order).
func SortClonesByOccurrence(dups [][]*syntax.Node) [][]*syntax.Node {
	sort.Slice(dups, func(i, j int) bool {
		// Sort by number of occurrences (files in each clone group)
		return len(dups[i]) > len(dups[j])
	})

	return dups
}

// SortClonesByHash sorts clone groups by hash (alphabetical, ascending order).
func SortClonesByHash(dups [][]*syntax.Node) [][]*syntax.Node {
	sort.Slice(dups, func(i, j int) bool {
		if handled, lessThan := isEmptyOrLessThanEmpty(dups, i, j); handled {
			return lessThan
		}

		return compareByNameThenPos(
			dups[i][0].Filename, dups[j][0].Filename,
			int(dups[i][0].Pos), int(dups[j][0].Pos),
		)
	})

	return dups
}

// countNonNilNodes counts non-nil nodes in a slice.
func countNonNilNodes(nodes []*syntax.Node) int {
	count := 0

	for _, node := range nodes {
		if node != nil {
			count++
		}
	}

	return count
}

// SortClonesByTotalTokens sorts clone groups by total token count across all files (largest first).
func SortClonesByTotalTokens(dups [][]*syntax.Node) [][]*syntax.Node {
	sort.Slice(dups, func(i, j int) bool {
		tokensI := countNonNilNodes(dups[i])
		tokensJ := countNonNilNodes(dups[j])

		return tokensI > tokensJ
	})

	return dups
}

// sumFragmentLengths sums the lengths of all clone fragments in a group.
func sumFragmentLengths(cloneGroups []clone) int {
	sum := 0
	for _, cl := range cloneGroups {
		sum += len(cl.fragment)
	}

	return sum
}

// Helper functions for text.go compatibility.
func sortCloneGroupsBySize(cloneGroups [][]clone) {
	sort.Slice(cloneGroups, func(i, j int) bool {
		return sumFragmentLengths(cloneGroups[i]) > sumFragmentLengths(cloneGroups[j])
	})
}

// ExtractSortCriteria extracts the sort criteria from variadic sortBy parameter.
// Returns SortBySize as default if no criteria is provided.
func ExtractSortCriteria(sortBy ...SortBy) SortBy {
	if len(sortBy) > 0 {
		return sortBy[0]
	}

	return SortBySize
}

func sortClonesByFilename(cloneGroups [][]clone) {
	sort.Slice(cloneGroups, func(i, j int) bool {
		if handled, lessThan := isEmptyOrLessThanEmpty(cloneGroups, i, j); handled {
			return lessThan
		}

		return compareByNameThenPos(
			cloneGroups[i][0].filename, cloneGroups[j][0].filename,
			cloneGroups[i][0].lineStart, cloneGroups[j][0].lineStart,
		)
	})
}
