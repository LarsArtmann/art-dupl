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

	dirA := t.TempDir()
	dirB := t.TempDir()

	fileA := dirA + "/a.go"
	fileB := dirB + "/b.go"

	writeFile(t, fileA, srcA)
	writeFile(t, fileB, srcB)

	// Load with EraseHash=true (suggest-generics mode)
	typeDataA, err := LoadTypeAwareData([]string{fileA}, true)
	if err != nil {
		t.Fatalf("LoadTypeAwareData A failed: %v", err)
	}

	typeDataB, err := LoadTypeAwareData([]string{fileB}, true)
	if err != nil {
		t.Fatalf("LoadTypeAwareData B failed: %v", err)
	}

	preA := typeDataA.LookupPreloaded(fileA)
	preB := typeDataB.LookupPreloaded(fileB)

	if preA == nil || preB == nil {
		t.Fatal("preloaded AST is nil")
	}

	if !preA.EraseHash || !preB.EraseHash {
		t.Fatal("EraseHash should be true on preloaded data")
	}

	nodeA := parsePreloadedTest(t, fileA, preA)
	nodeB := parsePreloadedTest(t, fileB, preB)

	hashA := findIdentHashForName(nodeA, "ts")
	hashB := findIdentHashForName(nodeB, "bi")

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

	typeData, err := LoadTypeAwareData([]string{file}, true)
	if err != nil {
		t.Fatalf("LoadTypeAwareData failed: %v", err)
	}

	pre := typeData.LookupPreloaded(file)
	if pre == nil {
		t.Fatal("preloaded AST is nil")
	}

	node := parsePreloadedTest(t, file, pre)

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

	dirA := t.TempDir()
	dirB := t.TempDir()

	fileA := dirA + "/a.go"
	fileB := dirB + "/b.go"

	writeFile(t, fileA, srcA)
	writeFile(t, fileB, srcB)

	// Type-aware mode: different hashes
	typeDataA, _ := LoadTypeAwareData([]string{fileA}, false)
	typeDataB, _ := LoadTypeAwareData([]string{fileB}, false)

	preA := typeDataA.LookupPreloaded(fileA)
	preB := typeDataB.LookupPreloaded(fileB)

	nodeA := parsePreloadedTest(t, fileA, preA)
	nodeB := parsePreloadedTest(t, fileB, preB)

	taHashA := findIdentHashForName(nodeA, "ts")
	taHashB := findIdentHashForName(nodeB, "bi")

	if taHashA == taHashB {
		t.Error("type-aware mode should produce DIFFERENT hashes for time.Time vs *big.Int")
	}

	// Erase-hash mode: same hashes
	typeDataA2, _ := LoadTypeAwareData([]string{fileA}, true)
	typeDataB2, _ := LoadTypeAwareData([]string{fileB}, true)

	preA2 := typeDataA2.LookupPreloaded(fileA)
	preB2 := typeDataB2.LookupPreloaded(fileB)

	nodeA2 := parsePreloadedTest(t, fileA, preA2)
	nodeB2 := parsePreloadedTest(t, fileB, preB2)

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
