package printer

import (
	"sort"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// sortByFilenameAndPosition is a generic helper that sorts slices by filename and position
func sortByFilenameAndPosition[T any](slices [][]T, getFilename func([]T) string, getPosition func([]T) int) {
	sort.Slice(slices, func(i, j int) bool {
		if len(slices[i]) == 0 && len(slices[j]) == 0 {
			return false
		}
		if len(slices[i]) == 0 {
			return true
		}
		if len(slices[j]) == 0 {
			return false
		}
		// Use Filename for deterministic sorting
		if getFilename(slices[i]) == getFilename(slices[j]) {
			return getPosition(slices[i]) < getPosition(slices[j])
		}
		return getFilename(slices[i]) < getFilename(slices[j])
	})
}

// sortNodesByFilename sorts []*syntax.Node groups by filename for deterministic output
func sortNodesByFilename(dups [][]*syntax.Node) {
	sortByFilenameAndPosition(dups,
		func(nodes []*syntax.Node) string {
			if len(nodes) > 0 {
				return nodes[0].Filename
			}
			return ""
		},
		func(nodes []*syntax.Node) int {
			if len(nodes) > 0 {
				return nodes[0].Pos
			}
			return 0
		})
}

// sortClonesByFilename sorts []clone groups by filename for deterministic output
func sortClonesByFilename(clones [][]clone) {
	sortByFilenameAndPosition(clones,
		func(cs []clone) string {
			if len(cs) > 0 {
				return cs[0].filename
			}
			return ""
		},
		func(cs []clone) int {
			if len(cs) > 0 {
				return cs[0].lineStart
			}
			return 0
		})
}

// SortClonesBySize sorts clone groups by token count (largest first)
func SortClonesBySize(dups [][]*syntax.Node) [][]*syntax.Node {
	// Sort by token count (largest first)
	sort.Slice(dups, func(i, j int) bool {
		// Handle empty groups
		if len(dups[i]) == 0 && len(dups[j]) == 0 {
			return false
		}
		if len(dups[i]) == 0 {
			return true
		}
		if len(dups[j]) == 0 {
			return false
		}
		// Calculate size of first occurrence (end position - start position)
		sizeI := dups[i][len(dups[i])-1].End - dups[i][0].Pos
		sizeJ := dups[j][len(dups[j])-1].End - dups[j][0].Pos
		return sizeI > sizeJ
	})
	return dups
}

// SortClonesByOccurrence sorts clone groups by number of files (most widespread first)
func SortClonesByOccurrence(dups [][]*syntax.Node) [][]*syntax.Node {
	// Sort by number of files (each dup in a different file)
	sort.Slice(dups, func(i, j int) bool {
		return len(dups[i]) > len(dups[j])
	})
	return dups
}

// SortClonesByHash sorts clone groups by hash (alphabetical)
func SortClonesByHash(dups [][]*syntax.Node) [][]*syntax.Node {
	// Simple lexical sort for deterministic output
	sortNodesByFilename(dups)
	return dups
}

// SortClonesByTotalTokens sorts clone groups by total token count across all files (largest first)
func SortClonesByTotalTokens(dups [][]*syntax.Node) [][]*syntax.Node {
	sort.Slice(dups, func(i, j int) bool {
		// Calculate total tokens for group i
		totalTokensI := 0
		for _, dup := range dups[i] {
			if dup != nil {
				totalTokensI++ // Count each node as a token
			}
		}
		// Calculate total tokens for group j
		totalTokensJ := 0
		for _, dup := range dups[j] {
			if dup != nil {
				totalTokensJ++ // Count each node as a token
			}
		}
		return totalTokensI > totalTokensJ
	})
	return dups
}

// SortCriteria represents different sorting strategies for clone groups
type SortCriteria int

const (
	SortBySize SortCriteria = iota
	SortByOccurrence
	SortByHash
	SortByTotalTokens
)

// SortCloneGroups generic function that sorts based on criteria
func SortCloneGroups[T interface{ GetSize() int }](groups []T, criteria SortCriteria) {
	switch criteria {
	case SortBySize:
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].GetSize() > groups[j].GetSize()
		})
	case SortByOccurrence:
		// This needs to be implemented per type for now
	case SortByHash:
		// This needs to be implemented per type for now
	case SortByTotalTokens:
		// This defaults to size for now
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].GetSize() > groups[j].GetSize()
		})
	}
}
