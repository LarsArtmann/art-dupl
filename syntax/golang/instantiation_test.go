package golang

import (
	"path/filepath"
	"slices"
	"testing"

	"golang.org/x/tools/go/packages"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// findIndexExprNode returns the first node whose base type is IndexExpr.
func findIndexExprNode(root *syntax.Node) *syntax.Node {
	var found *syntax.Node

	var walk func(n *syntax.Node) bool

	walk = func(n *syntax.Node) bool {
		if n == nil {
			return false
		}

		if DecodeBaseType(n.Type) == IndexExpr {
			found = n

			return true
		}

		return slices.ContainsFunc(n.Children, walk)
	}

	walk(root)

	return found
}

// TestTypeAware_InstantiationAndIndexingHashDifferently pins the T7 fix: with
// go/types data loaded (--type-aware / --suggest-generics), a generic
// instantiation `Box[int]` and a slice index `xs[i]` produce DIFFERENT node
// type hashes (types.Info.Instances disambiguates what the parser cannot).
// This kills the instantiation-vs-indexing false-pair class in opt-in modes.
func TestTypeAware_InstantiationAndIndexingHashDifferently(t *testing.T) {
	t.Parallel()

	srcA := `package testpkg

type Box[T any] struct{ Value T }

func makeIntBox(v int) Box[int] {
	b := Box[int]{Value: v}
	return b
}
`

	srcB := `package testpkg

func first(xs []int, i int) int {
	h := xs[i]
	return h
}
`

	pair := newTypeAwarePair(t, srcA, srcB, false)

	instNode := findIndexExprNode(pair.NodeA)
	indexNode := findIndexExprNode(pair.NodeB)

	if instNode == nil || indexNode == nil {
		t.Fatalf("IndexExpr nodes not found: inst=%v index=%v", instNode, indexNode)
	}

	if instNode.Type == indexNode.Type {
		t.Errorf("instantiation and indexing must hash differently in type-aware mode, both %d", instNode.Type)
	}

	if instNode.VarType == "" {
		t.Error("instantiation node must carry the instantiated type on VarType")
	}
}

// TestSemanticDefault_InstantiationAndIndexingStayEquivalent pins the Q2
// decision: the default syntax-only pipeline cannot distinguish instantiation
// from indexing (parser-level ambiguity), so both stay plain IndexExpr nodes.
// The limitation is documented in FEATURES.md and pinned by
// bdd/instantiation_collision_test.go.
func TestSemanticDefault_InstantiationAndIndexingStayEquivalent(t *testing.T) {
	t.Parallel()

	srcA := `package testpkg

type Box[T any] struct{ Value T }

func makeIntBox(v int) Box[int] {
	b := Box[int]{Value: v}
	return b
}
`

	srcB := `package testpkg

func first(xs []int, i int) int {
	h := xs[i]
	return h
}
`

	nodeA := parseSemantic(t, srcA)
	nodeB := parseSemantic(t, srcB)

	instNode := findIndexExprNode(nodeA)
	indexNode := findIndexExprNode(nodeB)

	if instNode == nil || indexNode == nil {
		t.Fatal("IndexExpr nodes not found")
	}

	if instNode.Type != indexNode.Type {
		t.Errorf("default semantic mode must stay ambiguity-blind, got %d vs %d", instNode.Type, indexNode.Type)
	}
}

// TestTypeAware_SelectorFormInstantiationFlagged covers the
// `pkg.Box[int]` selector form: go/types keys Instances by the identifier
// inside the selector, so the IndexExpr operand must resolve through
// SelectorExpr.Sel. Loads through a module-scoped go/packages call (the
// file= pattern cannot resolve module-internal imports without a Dir).
func TestTypeAware_SelectorFormInstantiationFlagged(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), "module t\n\ngo 1.27\n")
	writeFile(t, filepath.Join(dir, "box", "box.go"), "package box\n\ntype Box[T any] struct{ Value T }\n")

	mainSrc := `package main

import "t/box"

func makeBox(v int) box.Box[int] {
	b := box.Box[int]{Value: v}
	return b
}
`
	mainFile := filepath.Join(dir, "main.go")
	writeFile(t, mainFile, mainSrc)

	pkgs, err := packages.Load(
		&packages.Config{
			Mode: packages.NeedName | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
			Dir:  dir,
		},
		"file="+mainFile,
	)
	if err != nil || len(pkgs) == 0 || pkgs[0].TypesInfo == nil {
		t.Fatalf("module load failed: err=%v pkgs=%d", err, len(pkgs))
	}

	pkg := pkgs[0]
	for _, e := range pkg.Errors {
		t.Fatalf("package error: %v", e)
	}

	pre := &PreloadedAST{File: pkg.Syntax[0], TypeInfo: pkg.TypesInfo, Fset: pkg.Fset}
	node := parsePreloadedTest(t, mainFile, pre)

	instNode := findIndexExprNode(node)
	if instNode == nil {
		t.Fatal("IndexExpr node not found in selector-form instantiation")
	}

	if DecodeBaseType(instNode.Type) != IndexExpr {
		t.Fatalf("base type = %d, want IndexExpr", DecodeBaseType(instNode.Type))
	}

	if instNode.Type == IndexExpr {
		t.Error("selector-form instantiation must carry the instantiation marker hash")
	}

	if instNode.VarType == "" {
		t.Error("selector-form instantiation must carry the instantiated type on VarType")
	}
}

// TestTypeAware_PlainIndexingNodeUnchanged pins that non-generic indexing is
// byte-stable: the node type equals the plain IndexExpr constant and VarType
// stays empty (no behavioral change for indexing code in type-aware runs).
func TestTypeAware_PlainIndexingNodeUnchanged(t *testing.T) {
	t.Parallel()

	src := `package testpkg

func first(xs []int, i int) int {
	h := xs[i]
	return h
}
`

	pair := newTypeAwarePair(t, src, src, false)

	indexNode := findIndexExprNode(pair.NodeA)
	if indexNode == nil {
		t.Fatal("IndexExpr node not found")
	}

	if indexNode.Type != IndexExpr {
		t.Errorf("plain indexing must keep the unencoded node type, got %d", indexNode.Type)
	}

	if indexNode.VarType != "" {
		t.Errorf("plain indexing must not set VarType, got %q", indexNode.VarType)
	}
}
