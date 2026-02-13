package utils

import "github.com/LarsArtmann/art-dupl/syntax"

// Unique removes duplicate entries from a group of syntax nodes based on file and range.
// Two clones are considered duplicates if they have the same filename and the same range (start-end).
func Unique(group [][]*syntax.Node) [][]*syntax.Node {
	type rangeKey struct {
		filename string
		start    int
		end      int
	}
	seen := make(map[rangeKey]struct{})

	var newGroup [][]*syntax.Node
	for _, seq := range group {
		if len(seq) == 0 {
			continue
		}
		first := seq[0]
		last := seq[len(seq)-1]
		key := rangeKey{filename: first.Filename, start: int(first.Pos), end: int(last.End)}
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			newGroup = append(newGroup, seq)
		}
	}
	return newGroup
}

// CountUniqueFiles returns the number of unique files in a clone group.
func CountUniqueFiles(group [][]*syntax.Node) int {
	uniqueFiles := make(map[string]bool)
	for _, seq := range group {
		if len(seq) > 0 {
			uniqueFiles[seq[0].Filename] = true
		}
	}
	return len(uniqueFiles)
}
