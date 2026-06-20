package printer

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func TestClassifyCloneType(t *testing.T) {
	// Two fragments where every node has identical Name fields → Type 1.
	type1A := []*syntax.Node{
		{Type: golang.FuncDecl, Name: "process", Filename: "a.go"},
		{Type: golang.Ident, Name: "user", Filename: "a.go"},
		{Type: golang.SelectorExpr, Name: "Save", Filename: "a.go"},
	}
	type1B := []*syntax.Node{
		{Type: golang.FuncDecl, Name: "process", Filename: "b.go"},
		{Type: golang.Ident, Name: "user", Filename: "b.go"},
		{Type: golang.SelectorExpr, Name: "Save", Filename: "b.go"},
	}

	// Same structure, but identifier names differ → Type 2.
	type2A := []*syntax.Node{
		{Type: golang.FuncDecl, Name: "calcTotal", Filename: "a.go"},
		{Type: golang.Ident, Name: "items", Filename: "a.go"},
		{Type: golang.SelectorExpr, Name: "Price", Filename: "a.go"},
	}
	type2B := []*syntax.Node{
		{Type: golang.FuncDecl, Name: "sumScores", Filename: "b.go"},
		{Type: golang.Ident, Name: "records", Filename: "b.go"},
		{Type: golang.SelectorExpr, Name: "Score", Filename: "b.go"},
	}

	// Different lengths → Type 3 (defensive fallback).
	shortSeq := []*syntax.Node{
		{Type: golang.FuncDecl, Name: "x", Filename: "a.go"},
	}

	// Nodes with empty Name (e.g. BlockStmt) are equal when both empty → Type 1.
	structA := []*syntax.Node{
		{Type: golang.BlockStmt, Name: "", Filename: "a.go"},
		{Type: golang.Ident, Name: "n", Filename: "a.go"},
	}
	structB := []*syntax.Node{
		{Type: golang.BlockStmt, Name: "", Filename: "b.go"},
		{Type: golang.Ident, Name: "n", Filename: "b.go"},
	}

	cases := []struct {
		name string
		dups [][]*syntax.Node
		want domain.CloneType
	}{
		{name: "nil returns type-1", dups: nil, want: domain.CloneType1},
		{name: "single fragment returns type-1", dups: [][]*syntax.Node{type1A}, want: domain.CloneType1},
		{name: "identical names returns type-1", dups: [][]*syntax.Node{type1A, type1B}, want: domain.CloneType1},
		{name: "renamed identifiers returns type-2", dups: [][]*syntax.Node{type2A, type2B}, want: domain.CloneType2},
		{
			name: "mixed type-1 and type-2 returns type-2",
			dups: [][]*syntax.Node{type1A, type1B, type2B},
			want: domain.CloneType2,
		},
		{name: "unequal lengths returns type-3", dups: [][]*syntax.Node{type1A, shortSeq}, want: domain.CloneType3},
		{name: "empty names match returns type-1", dups: [][]*syntax.Node{structA, structB}, want: domain.CloneType1},
		{
			name: "three identical fragments returns type-1",
			dups: [][]*syntax.Node{type1A, type1B, type1A},
			want: domain.CloneType1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyCloneType(tc.dups)
			if got != tc.want {
				t.Errorf("classifyCloneType() = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestClassifyCloneType_DescendsIntoChildren verifies that clone-type
// classification walks the full subtree of each fragment node, not just the
// top-level roots. The renamed identifiers that distinguish Type 1 from Type 2
// live in descendant nodes.
func TestClassifyCloneType_DescendsIntoChildren(t *testing.T) {
	t.Parallel()

	// Two top-level roots with identical Type, but their child Idents have
	// different Names. Without subtree traversal this would be Type 1.
	rootA := &syntax.Node{
		Type: golang.BlockStmt,
		Children: []*syntax.Node{
			{Type: golang.Ident, Name: "alpha"},
			{Type: golang.Ident, Name: "beta"},
		},
	}
	rootB := &syntax.Node{
		Type: golang.BlockStmt,
		Children: []*syntax.Node{
			{Type: golang.Ident, Name: "gamma"},
			{Type: golang.Ident, Name: "delta"},
		},
	}

	// Children with identical names → Type 1.
	rootC := &syntax.Node{
		Type: golang.BlockStmt,
		Children: []*syntax.Node{
			{Type: golang.Ident, Name: "alpha"},
			{Type: golang.Ident, Name: "beta"},
		},
	}

	got := classifyCloneType([][]*syntax.Node{{rootA}, {rootB}})
	if got != domain.CloneType2 {
		t.Errorf("renamed children: got %q, want type-2", got)
	}

	got = classifyCloneType([][]*syntax.Node{{rootA}, {rootC}})
	if got != domain.CloneType1 {
		t.Errorf("identical children: got %q, want type-1", got)
	}
}
