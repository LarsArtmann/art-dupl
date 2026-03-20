package printer

import (
	"github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/pkg/position"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// FileInfo represents processed file information.
type FileInfo struct {
	Filename  string
	LineStart int
	LineEnd   int
	Content   []byte
	Node      *syntax.Node
}

// ProcessFileContent unified file processing for all printers.
func ProcessFileContent(fread ReadFile, node *syntax.Node) (*FileInfo, error) {
	if node == nil {
		return nil, errors.NewInternalError("nil node provided", nil)
	}

	// Read file content
	file, err := fread(node.Filename)
	if err != nil {
		return nil, errors.WrapIO(err, node.Filename, "read")
	}

	// Calculate line positions
	lineStart, lineEnd := position.ByteRangeToLines(file, int(node.Pos), int(node.End))

	return &FileInfo{
		Filename:  node.Filename,
		LineStart: lineStart,
		LineEnd:   lineEnd,
		Content:   file,
		Node:      node,
	}, nil
}

// ProcessNodeRange processes a range of nodes (start to end).
func ProcessNodeRange(fread ReadFile, startNode, endNode *syntax.Node) (*FileInfo, error) {
	if startNode == nil || endNode == nil {
		return nil, errors.NewInternalError("nil start or end node provided", nil)
	}

	// Use start node for filename, but combine positions
	file, err := fread(startNode.Filename)
	if err != nil {
		return nil, errors.WrapIO(err, startNode.Filename, "read")
	}

	lineStart, lineEnd := position.ByteRangeToLines(file, int(startNode.Pos), int(endNode.End))

	return &FileInfo{
		Filename:  startNode.Filename,
		LineStart: lineStart,
		LineEnd:   lineEnd,
		Content:   file,
		Node:      startNode,
	}, nil
}
