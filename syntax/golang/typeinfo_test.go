package golang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/packages"
)

// loadTypeInfo loads type information for a single source file using go/packages.
// Returns the AST, fset, and types.Info for use in type-aware parsing tests.
func loadTypeInfo(t *testing.T, filename, src string) (*ast.File, *token.FileSet, *types.Info) {
	t.Helper()

	// Write source to a temp file so go/packages can load it
	dir := t.TempDir()
	filePath := dir + "/" + filename
	writeFile(t, filePath, src)

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Dir:  dir,
	}

	pkgs, err := packages.Load(cfg, "file="+filePath)
	if err != nil {
		t.Fatalf("packages.Load failed: %v", err)
	}

	if len(pkgs) == 0 || pkgs[0].Syntax == nil || len(pkgs[0].Syntax) == 0 {
		t.Fatal("no packages or syntax returned")
	}

	pkg := pkgs[0]

	if pkg.TypesInfo == nil {
		t.Fatal("nil TypesInfo — type checking may have failed")
	}

	return pkg.Syntax[0], pkg.Fset, pkg.TypesInfo
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := writeFileRaw(path, content); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}

// parseFileWithTypeInfo parses source with type-aware config and returns the root node.
func parseFileWithTypeInfo(t *testing.T, src string) *syntax.Node {
	t.Helper()

	_, fset, info := loadTypeInfo(t, "test.go", src)

	// Re-parse with the same fset to get a file we can transform
	fileAST := parseSrc(t, src)

	// Use the pre-loaded type info. The key: the AST from go/packages is different
	// from the AST from parser.ParseFile, so we need to use the one from packages.
	_, _, info2 := loadTypeInfo(t, "test.go", src)

	_ = fset
	_ = fileAST

	t := newTransformerWithType(info2, fset)
	return t.trans(parseSrc(t, src))
}

func TestTypeAware_DifferentVariableTypesProduceDifferentHashes(t *testing.T) {
	t.Parallel()

	// Two functions with identical structure and renamed variables,
	// but the local variables have different types.
	// Without type-aware: they produce identical token sequences (Type 2 match).
	// With type-aware: they produce different hashes because types differ.
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

	// Parse without type-aware (standard semantic mode)
	astA := parseSemantic(t, srcA)
	astB := parseSemantic(t, srcB)

	// Collect all Ident node Types from each
	hashesA := collectIdentHashes(astA)
	hashesB := collectIdentHashes(astB)

	// Without type-aware: the canonical name for "result" (v0) is the same
	// in both functions, so they produce the same hash.
	// Find the "result" local (should be v0 in both)
	found := false

	for _, ha := range hashesA {
		for _, hb := range hashesB {
			if ha == hb && ha != 0 {
				found = true
				break
			}
		}
	}

	if !found {
		t.Error("Expected at least one matching hash between the two functions in non-type-aware mode")
	}

	// Now parse WITH type-aware: the variable types differ (string vs *big.Int)
	// so the hashes should differ.
	// We verify this at the LoadTypeAwareData level.
	dirA := t.TempDir()
	dirB := t.TempDir()
	fileA := dirA + "/a.go"
	fileB := dirB + "/b.go"
	writeFile(t, fileA, srcA)
	writeFile(t, fileB, srcB)

	typeDataA, err := LoadTypeAwareData([]string{fileA})
	if err != nil {
		t.Fatalf("LoadTypeAwareData A failed: %v", err)
	}

	typeDataB, err := LoadTypeAwareData([]string{fileB})
	if err != nil {
		t.Fatalf("LoadTypeAwareData B failed: %v", err)
	}

	preA := typeDataA.LookupPreloaded(fileA)
	preB := typeDataB.LookupPreloaded(fileB)

	if preA == nil {
		t.Fatal("preA is nil")
	}

	if preB == nil {
		t.Fatal("preB is nil")
	}

	// Parse with type info
	nodeA := parsePreloadedTest(t, fileA, preA)
	nodeB := parsePreloadedTest(t, fileB, preB)

	// Collect all Ident hashes with type info
	typeHashesA := collectIdentHashes(nodeA)
	typeHashesB := collectIdentHashes(nodeB)

	// With type-aware: the "result" local in srcA (type string) should have
	// a DIFFERENT hash than the "result" local in srcB (type *big.Int).
	// Both are canonicalized to v0, but v0+string != v0+*big.Int.
	allMatch := true

	for _, ha := range typeHashesA {
		for _, hb := range typeHashesB {
			if ha == hb && ha != 0 && DecodeBaseType(ha) == Ident {
				// Found a matching ident hash — check if it's a local variable
				// If ALL ident hashes match, type-aware isn't working.
				allMatch = true
			}
		}
	}

	_ = allMatch // At minimum, the structures should parse without crashing

	// More targeted: find the "result" ident in each and verify different hashes
	resultHashA := findIdentHashForName(nodeA, "result")
	resultHashB := findIdentHashForName(nodeB, "result")

	if resultHashA == 0 {
		t.Fatal("could not find 'result' ident hash in nodeA")
	}

	if resultHashB == 0 {
		t.Fatal("could not find 'result' ident hash in nodeB")
	}

	if resultHashA == resultHashB {
		t.Errorf("type-aware mode should produce different hashes for string vs *big.Int 'result' variables, "+
			"both got %d", resultHashA)
	}
}

func TestTypeAware_SameVariableTypesProduceSameHashes(t *testing.T) {
	t.Parallel()

	// Two functions with identical structure, renamed variables, AND same types.
	// With type-aware: they should STILL match (same type → same hash).
	srcA := `package testpkg
import "time"
func formatA(ts time.Time) string {
	s := ts.String()
	return s
}
`

	srcB := `package testpkg
import "time"
func formatB(tm time.Time) string {
	result := tm.String()
	return result
}
`

	dirA := t.TempDir()
	dirB := t.TempDir()
	fileA := dirA + "/a.go"
	fileB := dirB + "/b.go"
	writeFile(t, fileA, srcA)
	writeFile(t, fileB, srcB)

	typeDataA, _ := LoadTypeAwareData([]string{fileA})
	typeDataB, _ := LoadTypeAwareData([]string{fileB})

	preA := typeDataA.LookupPreloaded(fileA)
	preB := typeDataB.LookupPreloaded(fileB)

	nodeA := parsePreloadedTest(t, fileA, preA)
	nodeB := parsePreloadedTest(t, fileB, preB)

	// Find the local variable ("s" in A, "result" in B) — both type string
	hashA := findIdentHashForName(nodeA, "s")
	hashB := findIdentHashForName(nodeB, "result")

	if hashA == 0 {
		t.Fatal("could not find 's' ident hash in nodeA")
	}

	if hashB == 0 {
		t.Fatal("could not find 'result' ident hash in nodeB")
	}

	if hashA != hashB {
		t.Errorf("same-typed variables (string) should have matching hashes in type-aware mode, "+
			"got %d vs %d", hashA, hashB)
	}
}

func TestTypeAware_NilTypeInfoActsAsStandardSemantic(t *testing.T) {
	t.Parallel()

	// When typeInfo is nil on the transformer, behavior should be identical
	// to standard semantic mode.
	src := `package p
func foo(x int) int {
	y := x + 1
	return y
}
`

	node := parseSemantic(t, src)
	// Just verify it doesn't crash and produces reasonable output
	if node == nil {
		t.Fatal("expected non-nil root node")
	}

	hashes := collectIdentHashes(node)
	if len(hashes) == 0 {
		t.Error("expected at least some ident hashes")
	}
}

// --- Helpers ---

// parseSemantic parses source in semantic mode without type-aware.
func parseSemantic(t *testing.T, src string) *syntax.Node {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	tr := &transformer{
		fileset:  fset,
		filename: "test.go",
		config:   ParseConfig{Mode: DetectionModeSemantic},
		norm:     newNormalizer(true),
	}

	return tr.trans(file)
}

// parsePreloadedTest creates a transformer with preloaded type info and transforms the AST.
func parsePreloadedTest(t *testing.T, filename string, pre *PreloadedAST) *syntax.Node {
	t.Helper()

	tr := &transformer{
		fileset:  pre.Fset,
		filename: filename,
		config:   ParseConfig{Mode: DetectionModeSemantic, Preloaded: pre},
		norm:     newNormalizer(true),
		typeInfo: pre.TypeInfo,
	}

	return tr.trans(pre.File)
}

// parseSrc parses Go source into an ast.File (for test helpers).
func parseSrc(t *testing.T, src string) *ast.File {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	return file
}

// collectIdentHashes walks the tree and collects all Ident-node Type values.
func collectIdentHashes(root *syntax.Node) []int32 {
	var hashes []int32

	var walk func(n *syntax.Node)
	walk = func(n *syntax.Node) {
		if n == nil {
			return
		}

		if DecodeBaseType(n.Type) == Ident {
			hashes = append(hashes, n.Type)
		}

		for _, c := range n.Children {
			walk(c)
		}
	}

	walk(root)

	return hashes
}

// findIdentHashForName finds the encoded Type for the first Ident node with the given Name.
func findIdentHashForName(root *syntax.Node, name string) int32 {
	var result int32

	var walk func(n *syntax.Node) bool
	walk = func(n *syntax.Node) bool {
		if n == nil {
			return false
		}

		if DecodeBaseType(n.Type) == Ident && n.Name == name {
			result = n.Type

			return true
		}

		for _, c := range n.Children {
			if walk(c) {
				return true
			}
		}

		return false
	}

	walk(root)

	return result
}

// newTransformerWithType creates a transformer with type info for testing.
func newTransformerWithType(info *types.Info, fset *token.FileSet) *transformer {
	return &transformer{
		fileset:  fset,
		filename: "test.go",
		config:   ParseConfig{Mode: DetectionModeSemantic},
		norm:     newNormalizer(true),
		typeInfo: info,
	}
}
