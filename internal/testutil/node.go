package testutil

import (
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// CreateMockNode creates a simple mock AST node for testing.
func CreateMockNode(nodeType int, filename string, pos, end int) *syntax.Node {
	return &syntax.Node{ //nolint:exhaustruct
		Type:     int32(nodeType), // #nosec G115 -- Test helper with controlled values
		Filename: filename,
		Pos:      int32(pos), // #nosec G115 -- Test helper with controlled values
		End:      int32(end), // #nosec G115 -- Test helper with controlled values
	}
}

// CreateMockNodes creates multiple mock nodes for testing.
func CreateMockNodes(count int, filename string) []*syntax.Node {
	nodes := make([]*syntax.Node, count)
	for i := 0; i < count; i++ {
		offset := i * 10
		nodes[i] = CreateMockNode(golang.FuncDecl, filename, offset, offset+10)
	}

	return nodes
}

// CreateMockCloneGroup creates a mock clone group with given files.
func CreateMockCloneGroup(hash string, size int, filenames []string) []*syntax.Node {
	nodes := make([]*syntax.Node, len(filenames))
	for idx, fname := range filenames {
		step := 25
		start := idx * step
		nodes[idx] = CreateMockNode(golang.FuncDecl, fname, start, start+step)
	}

	return nodes
}
