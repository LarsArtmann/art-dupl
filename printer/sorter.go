package printer

import (
	"github.com/LarsArtmann/art-dupl/config"
)

// compareByNameThenPos compares two items by name, then by position for stable ordering.
func compareByNameThenPos(nameI, nameJ string, posI, posJ int) bool {
	if nameI == nameJ {
		return posI < posJ
	}

	return nameI < nameJ
}

var cloneGroupMetrics = GroupMetrics[CloneGroup]{
	Size:    func(g CloneGroup) int { return g.Size },
	Count:   func(g CloneGroup) int { return len(g.Clones) },
	SortKey: func(g CloneGroup) string { return g.Hash },
}

// SortCloneGroups sorts CloneGroup arrays by specified criteria.
func SortCloneGroups(groups []CloneGroup, sortBy config.SortCriteria) {
	sortGroupsByCriteria(groups, sortBy, cloneGroupMetrics)
}

// ExtractSortCriteria extracts the sort criteria from variadic sortBy parameter.
// Returns SortBySize as default if no criteria is provided.
func ExtractSortCriteria(sortBy ...config.SortCriteria) config.SortCriteria {
	if len(sortBy) > 0 {
		return sortBy[0]
	}

	return config.SortBySize
}
