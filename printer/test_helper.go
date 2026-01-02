package printer

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// TestCloneSorting is a helper function for testing clone sorting algorithms that takes pre-created clones.
func TestCloneSortingWithData(t *testing.T, sortFunc func([][]*syntax.Node) [][]*syntax.Node, sortName string, clones [][]*syntax.Node) {
	sorted := sortFunc(clones)

	if len(sorted[0]) < len(sorted[1]) || len(sorted[1]) < len(sorted[2]) {
		t.Errorf("%s sorting failed. Expected: 8, 5, 2. Got: %d, %d, %d",
			sortName, len(sorted[0]), len(sorted[1]), len(sorted[2]))
	}
}
