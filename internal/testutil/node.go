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

// CreateMatch creates a syntax.Match with a hash and one or more filenames.
// Each filename creates a separate fragment containing a single node.
func CreateMatch(hash string, filenames ...string) syntax.Match {
	frags := make([][]*syntax.Node, len(filenames))
	for i, fname := range filenames {
		frags[i] = []*syntax.Node{{Filename: fname}} //nolint:exhaustruct
	}
	return syntax.Match{Hash: hash, Frags: frags}
}

// CreateMatchWithNodes creates a syntax.Match with explicit node fragments.
func CreateMatchWithNodes(hash string, fragments [][]*syntax.Node) syntax.Match {
	return syntax.Match{Hash: hash, Frags: fragments}
}
