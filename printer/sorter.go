package printer

import (
	"sort"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// SortCloneGroups sorts CloneGroup arrays by specified criteria.
func SortCloneGroups(groups []CloneGroup, sortBy string) {
	switch sortBy {
	case sortBySize:
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].Size > groups[j].Size
		})
	case sortByOccurrence:
		sort.Slice(groups, func(i, j int) bool {
			return len(groups[i].Files) > len(groups[j].Files)
		})
	case sortByHash:
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].Hash < groups[j].Hash
		})
	case sortByTotalTokens:
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].Size > groups[j].Size // Same as size for CloneGroup
		})
	default:
		// Default to size sorting for highest impact
		sort.Slice(groups, func(i, j int) bool {
			return groups[i].Size > groups[j].Size
		})
	}
}

// SortClonesBySize sorts clone groups by token count (largest first).
func SortClonesBySize(dups [][]*syntax.Node) [][]*syntax.Node {
	sort.Slice(dups, func(i, j int) bool {
		if len(dups[i]) == 0 {
			return true
		}
		if len(dups[j]) == 0 {
			return false
		}
		// Calculate size as end position minus start position
		sizeI := dups[i][len(dups[i])-1].End - dups[i][0].Pos
		sizeJ := dups[j][len(dups[j])-1].End - dups[j][0].Pos
		return sizeI > sizeJ
	})
	return dups
}

// SortClonesByOccurrence sorts clone groups by number of files (most widespread first).
func SortClonesByOccurrence(dups [][]*syntax.Node) [][]*syntax.Node {
	// For [][]*syntax.Node, occurrence is always 1 (single group)
	// This function is kept for compatibility but sorts by size instead
	return SortClonesBySize(dups)
}

// SortClonesByHash sorts clone groups by hash (alphabetical).
func SortClonesByHash(dups [][]*syntax.Node) [][]*syntax.Node {
	sort.Slice(dups, func(i, j int) bool {
		if len(dups[i]) == 0 {
			return true
		}
		if len(dups[j]) == 0 {
			return false
		}
		// Use filename for sorting since hash isn't available in Node
		if dups[i][0].Filename == dups[j][0].Filename {
			return dups[i][0].Pos < dups[j][0].Pos
		}
		return dups[i][0].Filename < dups[j][0].Filename
	})
	return dups
}

// SortClonesByTotalTokens sorts clone groups by total token count across all files (largest first).
func SortClonesByTotalTokens(dups [][]*syntax.Node) [][]*syntax.Node {
	sort.Slice(dups, func(i, j int) bool {
		tokensI := 0
		for _, node := range dups[i] {
			if node != nil {
				tokensI++
			}
		}
		tokensJ := 0
		for _, node := range dups[j] {
			if node != nil {
				tokensJ++
			}
		}
		return tokensI > tokensJ
	})
	return dups
}

// Helper functions for text.go compatibility.
func sortCloneGroupsBySize(cloneGroups [][]clone) {
	sort.Slice(cloneGroups, func(i, j int) bool {
		// Sort by total size of all clones in group
		sizeI := 0
		for _, cl := range cloneGroups[i] {
			sizeI += len(cl.fragment)
		}
		sizeJ := 0
		for _, cl := range cloneGroups[j] {
			sizeJ += len(cl.fragment)
		}
		return sizeI > sizeJ
	})
}

func sortClonesByFilename(cloneGroups [][]clone) {
	sort.Slice(cloneGroups, func(i, j int) bool {
		if len(cloneGroups[i]) == 0 && len(cloneGroups[j]) == 0 {
			return false
		}
		if len(cloneGroups[i]) == 0 {
			return true
		}
		if len(cloneGroups[j]) == 0 {
			return false
		}
		// Compare by filename of first clone in each group
		if len(cloneGroups[i]) > 0 && len(cloneGroups[j]) > 0 {
			if cloneGroups[i][0].filename == cloneGroups[j][0].filename {
				return cloneGroups[i][0].lineStart < cloneGroups[j][0].lineStart
			}
			return cloneGroups[i][0].filename < cloneGroups[j][0].filename
		}
		return false
	})
}

func sortNodesByFilename(dups [][]*syntax.Node) {
	sort.Slice(dups, func(i, j int) bool {
		if len(dups[i]) == 0 && len(dups[j]) == 0 {
			return false
		}
		if len(dups[i]) == 0 {
			return true
		}
		if len(dups[j]) == 0 {
			return false
		}
		// Use Filename for deterministic sorting
		if len(dups[i]) > 0 && len(dups[j]) > 0 {
			if dups[i][0].Filename == dups[j][0].Filename {
				return dups[i][0].Pos < dups[j][0].Pos
			}
			return dups[i][0].Filename < dups[j][0].Filename
		}
		return false
	})
}
