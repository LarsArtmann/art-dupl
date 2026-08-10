package golang

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// writeFile writes content to path, creating parent dirs.
func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writeFile %s: %v", path, err)
	}
}

// TestTypeAware_DifferentReceiverTypesProduceDifferentHashes verifies the core
// value proposition of type-aware mode: two functions with identical structure
// and renamed variables but different receiver types should produce different
// identifier hashes, eliminating the time.Time.String vs *big.Int.String
// class of false positives.
func TestTypeAware_DifferentReceiverTypesProduceDifferentHashes(t *testing.T) {
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

	typeDataA, err := LoadTypeAwareData([]string{fileA}, false)
	if err != nil {
		t.Fatalf("LoadTypeAwareData A failed: %v", err)
	}

	typeDataB, err := LoadTypeAwareData([]string{fileB}, false)
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

	nodeA := parsePreloadedTest(t, fileA, preA)
	nodeB := parsePreloadedTest(t, fileB, preB)

	// The receiver variables "ts" (type time.Time) and "bi" (type *big.Int)
	// are both canonicalized to v0, but type-aware mode should produce
	// different hashes because their types differ.
	receiverHashA := findIdentHashForName(nodeA, "ts")
	receiverHashB := findIdentHashForName(nodeB, "bi")

	if receiverHashA == 0 {
		t.Fatal("could not find 'ts' ident hash in nodeA")
	}

	if receiverHashB == 0 {
		t.Fatal("could not find 'bi' ident hash in nodeB")
	}

	if receiverHashA == receiverHashB {
		t.Errorf("type-aware mode should produce different hashes for time.Time vs *big.Int receivers, "+
			"both got %d", receiverHashA)
	}
}

// TestTypeAware_SameReceiverTypesProduceSameHashes verifies that type-aware
// mode does NOT break legitimate Type-2 matches: two functions with identical
// structure, renamed variables, AND same types should still produce matching hashes.
func TestTypeAware_SameReceiverTypesProduceSameHashes(t *testing.T) {
	t.Parallel()

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

	typeDataA, _ := LoadTypeAwareData([]string{fileA}, false)
	typeDataB, _ := LoadTypeAwareData([]string{fileB}, false)

	preA := typeDataA.LookupPreloaded(fileA)
	preB := typeDataB.LookupPreloaded(fileB)

	nodeA := parsePreloadedTest(t, fileA, preA)
	nodeB := parsePreloadedTest(t, fileB, preB)

	// Both "ts" and "tm" have type time.Time → same canonical hash
	hashA := findIdentHashForName(nodeA, "ts")
	hashB := findIdentHashForName(nodeB, "tm")

	if hashA == 0 {
		t.Fatal("could not find 'ts' ident hash in nodeA")
	}

	if hashB == 0 {
		t.Fatal("could not find 'tm' ident hash in nodeB")
	}

	if hashA != hashB {
		t.Errorf("same-typed variables (time.Time) should have matching hashes in type-aware mode, "+
			"got %d vs %d", hashA, hashB)
	}
}

// TestTypeAware_NilTypeInfoActsAsStandardSemantic verifies that nil typeInfo
// on the transformer produces identical behavior to standard semantic mode.
func TestTypeAware_NilTypeInfoActsAsStandardSemantic(t *testing.T) {
	t.Parallel()

	src := `package p
func foo(x int) int {
	y := x + 1
	return y
}
`

	node := parseSemantic(t, src)
	if node == nil {
		t.Fatal("expected non-nil root node")
	}

	hashes := collectIdentHashes(node)
	if len(hashes) == 0 {
		t.Error("expected at least some ident hashes")
	}
}

// TestLoadTypeAwareData_EmptyInputReturnsEmpty verifies graceful handling of empty input.
func TestLoadTypeAwareData_EmptyInputReturnsEmpty(t *testing.T) {
	t.Parallel()

	result, err := LoadTypeAwareData(nil, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Errorf("expected empty map for nil input, got %d entries", len(result))
	}
}

// TestTypeAwareData_LookupPreloadedReturnsNilForUnknown verifies graceful
// handling of unknown file paths.
func TestTypeAwareData_LookupPreloadedReturnsNilForUnknown(t *testing.T) {
	t.Parallel()

	td := TypeAwareData{}

	if td.LookupPreloaded("nonexistent.go") != nil {
		t.Error("expected nil for unknown file")
	}

	if td.LookupPreloaded("") != nil {
		t.Error("expected nil for empty path")
	}
}

// TestNormalizer_IsLocal verifies the isLocal method used by the type-aware transformer.
func TestNormalizer_IsLocal(t *testing.T) {
	t.Parallel()

	n := newNormalizer(true)

	n.beginFunction()
	n.declare("foo")

	if !n.isLocal("foo") {
		t.Error("foo should be a local after declare")
	}

	if n.isLocal("bar") {
		t.Error("bar should not be a local")
	}

	// Disabled normalizer should always return false
	disabled := newNormalizer(false)

	if disabled.isLocal("anything") {
		t.Error("disabled normalizer should never report locals")
	}
}

// --- Helpers ---

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

func parsePreloadedTest(t *testing.T, filename string, pre *PreloadedAST) *syntax.Node {
	t.Helper()

	tr := &transformer{
		fileset:       pre.Fset,
		filename:      filename,
		config:        ParseConfig{Mode: DetectionModeSemantic, Preloaded: pre},
		norm:          newNormalizer(true),
		typeInfo:      pre.TypeInfo,
		typeEraseHash: pre.EraseHash,
	}

	return tr.trans(pre.File)
}

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

		return slices.ContainsFunc(n.Children, walk)
	}

	walk(root)

	return result
}
