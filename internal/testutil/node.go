package testutil

import (
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// CreateMockNode creates a simple mock AST node for testing.
func CreateMockNode(nodeType int, filename string, pos, end int) *syntax.Node {
	return &syntax.Node{
		Type:     int32(nodeType), // #nosec G115 -- Test helper with controlled values
		Filename: filename,
		Pos:      int32(pos), // #nosec G115 -- Test helper with controlled values
		End:      int32(end), // #nosec G115 -- Test helper with controlled values
	}
}

// CreateMockNodes creates multiple mock nodes for testing.
func CreateMockNodes(count int, filename string) []*syntax.Node {
	nodes := make([]*syntax.Node, count)
	for i := range nodes {
		nodes[i] = CreateMockNode(golang.FuncDecl, filename, i*10, (i+1)*10)
	}
	return nodes
}

// CreateMockCloneGroup creates a mock clone group with given files.
func CreateMockCloneGroup(hash string, size int, filenames []string) []*syntax.Node {
	nodes := make([]*syntax.Node, len(filenames))
	for i, filename := range filenames {
		nodes[i] = CreateMockNode(golang.FuncDecl, filename, i*20, (i+1)*20)
	}
	return nodes
}
