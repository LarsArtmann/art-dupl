package printer

import (
	"fmt"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/pkg/position"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// FileInfo represents processed file information.
type FileInfo struct {
	domain.CloneRef

	Content []byte
	Node    *syntax.Node
}

// ProcessFileContent unified file processing for all printers.
func ProcessFileContent(fread ReadFile, node *syntax.Node) (*FileInfo, error) {
	if node == nil {
		return nil, errors.NewInternalError(fmt.Sprintf("nil node provided (fread=%v)", fread), nil)
	}

	// Read file content
	file, err := fread(node.Filename)
	if err != nil {
		return nil, fmt.Errorf(
			"read file failed (fread=%v, filename=%s): %w",
			fread,
			node.Filename,
			err,
		)
	}

	// Calculate line positions
	lineStart, lineEnd := position.ByteRangeToLines(file, int(node.Pos), int(node.End))

	return &FileInfo{
		CloneRef: domain.CloneRef{
			Filename:  node.Filename,
			LineStart: lineStart,
			LineEnd:   lineEnd,
		},
		Content: file,
		Node:    node,
	}, nil
}

// ProcessNodeRange processes a range of nodes (start to end).
func ProcessNodeRange(fread ReadFile, startNode, endNode *syntax.Node) (*FileInfo, error) {
	if startNode == nil || endNode == nil {
		return nil, errors.NewInternalError(
			fmt.Sprintf(
				"nil node provided (fread=%v, startNode=%v, endNode=%v)",
				fread,
				startNode != nil,
				endNode != nil,
			),
			nil,
		)
	}

	// Use start node for filename, but combine positions
	file, err := fread(startNode.Filename)
	if err != nil {
		return nil, fmt.Errorf(
			"read file failed (fread=%v, filename=%s): %w",
			fread,
			startNode.Filename,
			err,
		)
	}

	lineStart, lineEnd := position.ByteRangeToLines(file, int(startNode.Pos), int(endNode.End))

	return &FileInfo{
		CloneRef: domain.CloneRef{
			Filename:  startNode.Filename,
			LineStart: lineStart,
			LineEnd:   lineEnd,
		},
		Content: file,
		Node:    startNode,
	}, nil
}
