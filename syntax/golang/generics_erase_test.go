package golang

import (
	"slices"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// TestEraseHash_DifferentTypesProduceSameHashes verifies the core mechanism of
// --suggest-generics: when EraseHash is true, variables with different types
// but identical canonical names produce the SAME hash (so clones match), while
// VarType is still populated for post-detection classification.
func TestEraseHash_DifferentTypesProduceSameHashes(t *testing.T) {
	t.Parallel()

	srcA := `package testpkg
import "time"
func processTimeData(ts time.Time) string {
	result := ts.String()
	return result
}
`

	srcB := `package testpkg
import "math/big"
func processBigIntData(bi *big.Int) string {
	result := bi.String()
	return result
}
`

	// Load with EraseHash=true (suggest-generics mode)
	pair := newTypeAwarePair(t, srcA, srcB, true)

	if !pair.PreA.EraseHash || !pair.PreB.EraseHash {
		t.Fatal("EraseHash should be true on preloaded data")
	}

	hashA := findIdentHashForName(pair.NodeA, "ts")
	hashB := findIdentHashForName(pair.NodeB, "bi")

	if hashA == 0 {
		t.Fatal("could not find 'ts' ident hash in nodeA")
	}

	if hashB == 0 {
		t.Fatal("could not find 'bi' ident hash in nodeB")
	}

	if hashA != hashB {
		t.Errorf("erase-hash mode should produce SAME hashes for time.Time vs *big.Int receivers "+
			"(both canonicalize to v0 without type encoding), got %d vs %d", hashA, hashB)
	}
}

// TestEraseHash_VarTypeStillPopulated verifies that EraseHash populates VarType
// on nodes even though the type is not encoded in the hash. This is critical
// for post-detection generics classification.
func TestEraseHash_VarTypeStillPopulated(t *testing.T) {
	t.Parallel()

	src := `package testpkg
import "time"
func processTimeData(ts time.Time) string {
	result := ts.String()
	return result
}
`

	dir := t.TempDir()
	file := dir + "/a.go"
	writeFile(t, file, src)

	node, _ := parseTypeAwareFile(t, file, true)

	vt := findVarTypeForName(node, "ts")
	if vt == "" {
		t.Error("VarType should be populated for 'ts' even in erase-hash mode")
	}
}

// TestEraseHash_ComparedWithTypeAware verifies the contrast: type-aware
// (eraseHash=false) produces different hashes for different types, while
// erase-hash (eraseHash=true) produces the same hashes. This is the behavioral
// difference between --type-aware and --suggest-generics.
func TestEraseHash_ComparedWithTypeAware(t *testing.T) {
	t.Parallel()

	srcA := `package testpkg
import "time"
func f(ts time.Time) string { return ts.String() }
`

	srcB := `package testpkg
import "math/big"
func f(bi *big.Int) string { return bi.String() }
`

	// Type-aware mode: different hashes
	pair := newTypeAwarePair(t, srcA, srcB, false)

	taHashA := findIdentHashForName(pair.NodeA, "ts")
	taHashB := findIdentHashForName(pair.NodeB, "bi")

	if taHashA == taHashB {
		t.Error("type-aware mode should produce DIFFERENT hashes for time.Time vs *big.Int")
	}

	// Erase-hash mode: same hashes
	nodeA2, _ := parseTypeAwareFile(t, pair.FileA, true)
	nodeB2, _ := parseTypeAwareFile(t, pair.FileB, true)

	ehHashA := findIdentHashForName(nodeA2, "ts")
	ehHashB := findIdentHashForName(nodeB2, "bi")

	if ehHashA != ehHashB {
		t.Errorf("erase-hash mode should produce SAME hashes for time.Time vs *big.Int, "+
			"got %d vs %d", ehHashA, ehHashB)
	}
}

// findVarTypeForName finds the VarType of the first Ident node with the given Name.
func findVarTypeForName(root *syntax.Node, name string) string {
	var result string

	var walk func(n *syntax.Node) bool

	walk = func(n *syntax.Node) bool {
		if n == nil {
			return false
		}

		if DecodeBaseType(n.Type) == Ident && n.Name == name {
			result = n.VarType

			return true
		}

		return slices.ContainsFunc(n.Children, walk)
	}

	walk(root)

	return result
}
