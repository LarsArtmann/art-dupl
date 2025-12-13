// Package util provides shared utility functions for dupl
package util

import "github.com/LarsArtmann/art-dupl/syntax"

// Unique removes duplicate entries from a group of syntax nodes based on file and position
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
		if _, ok := file[node.Pos]; !ok {
			file[node.Pos] = struct{}{}
			newGroup = append(newGroup, seq)
		}
	}
	return newGroup
}
