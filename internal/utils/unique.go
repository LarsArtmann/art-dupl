package utils

import "github.com/LarsArtmann/art-dupl/syntax"

// Unique removes duplicate entries from a group of syntax nodes based on file and position.
func Unique(group [][]*syntax.Node) [][]*syntax.Node {
	fileMap := make(map[string]map[int]struct{})

	var newGroup [][]*syntax.Node
	for _, seq := range group {
		node := seq[0]
		file, ok := fileMap[node.Filename]
		if !ok {
			file = make(map[int]struct{})
			fileMap[node.Filename] = file
		}
		if _, ok := file[int(node.Pos)]; !ok {
			file[int(node.Pos)] = struct{}{}
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
