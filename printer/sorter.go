package printer

import (
	"sort"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// SortClonesBySize sorts clone groups by token count (largest first)
func SortClonesBySize(dups [][]*syntax.Node) [][]*syntax.Node {
	// Sort by token count (largest first)
	sort.Slice(dups, func(i, j int) bool {
		return len(dups[i]) > len(dups[j])
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
	return dups
}