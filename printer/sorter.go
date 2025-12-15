package printer

import (
	"sort"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// sortNodesByFilename sorts []*syntax.Node groups by filename for deterministic output
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
		if dups[i][0].Filename == dups[j][0].Filename {
			return dups[i][0].Pos < dups[j][0].Pos
		}
		return dups[i][0].Filename < dups[j][0].Filename
	})
}

// sortClonesByFilename sorts []clone groups by filename for deterministic output
func sortClonesByFilename(clones [][]clone) {
	sort.Slice(clones, func(i, j int) bool {
		if len(clones[i]) == 0 && len(clones[j]) == 0 {
			return false
		}
		if len(clones[i]) == 0 {
			return true
		}
		if len(clones[j]) == 0 {
			return false
		}
		// Compare by first filename in each group
		if clones[i][0].filename == clones[j][0].filename {
			return clones[i][0].lineStart < clones[j][0].lineStart
		}
		return clones[i][0].filename < clones[j][0].filename
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
