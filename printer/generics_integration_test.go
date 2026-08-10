package printer

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// TestProcessClones_GenericsCandidateSet verifies that ProcessClones correctly
// sets GenericsCandidate and GenericsHint on clone groups whose instances have
// differing VarType values.
func TestProcessClones_GenericsCandidateSet(t *testing.T) {
	t.Parallel()

	// Two structurally-identical clones but with different VarType on the
	// range variable — the hallmark of a generics-extraction candidate.
	dupA := []*syntax.Node{
		{
			Type:     golang.Ident,
			Name:     "k",
			VarType:  "github.com/example/db.AuthorKindActivity",
			Filename: "test_a.go",
			Pos:      1,
			End:      2,
		},
	}

	dupB := []*syntax.Node{
		{
			Type:     golang.Ident,
			Name:     "g",
			VarType:  "github.com/example/db.MemberGrowthPoint",
			Filename: "test_b.go",
			Pos:      1,
			End:      2,
		},
	}

	dups := [][]*syntax.Node{dupA, dupB}

	clones, err := ProcessClones(noopReadFile, dups)
	if err != nil {
		t.Fatalf("ProcessClones failed: %v", err)
	}

	if len(clones) != 2 {
		t.Fatalf("expected 2 clones, got %d", len(clones))
	}

	for i, c := range clones {
		if !c.Classification.GenericsCandidate {
			t.Errorf("clone %d: expected GenericsCandidate=true", i)
		}

		if c.Classification.GenericsHint == "" {
			t.Errorf("clone %d: expected non-empty GenericsHint", i)
		}
	}
}

// TestProcessClones_SameTypesNotGenericsCandidate verifies that ProcessClones
// does NOT mark clones as generics candidates when all VarTypes are identical.
func TestProcessClones_SameTypesNotGenericsCandidate(t *testing.T) {
	t.Parallel()

	dupA := []*syntax.Node{
		{
			Type:     golang.Ident,
			Name:     "k",
			VarType:  "int64",
			Filename: "test_a.go",
			Pos:      1,
			End:      2,
		},
	}

	dupB := []*syntax.Node{
		{
			Type:     golang.Ident,
			Name:     "g",
			VarType:  "int64",
			Filename: "test_b.go",
			Pos:      1,
			End:      2,
		},
	}

	dups := [][]*syntax.Node{dupA, dupB}

	clones, err := ProcessClones(noopReadFile, dups)
	if err != nil {
		t.Fatalf("ProcessClones failed: %v", err)
	}

	for i, c := range clones {
		if c.Classification.GenericsCandidate {
			t.Errorf("clone %d: expected GenericsCandidate=false (same types)", i)
		}
	}
}

// TestProcessClones_NoVarTypeNotGenericsCandidate verifies that clones without
// any VarType (standard non-type-aware mode) are never marked as generics candidates.
func TestProcessClones_NoVarTypeNotGenericsCandidate(t *testing.T) {
	t.Parallel()

	dupA := []*syntax.Node{
		{
			Type:     golang.Ident,
			Name:     "k",
			Filename: "test_a.go",
			Pos:      1,
			End:      2,
		},
	}

	dupB := []*syntax.Node{
		{
			Type:     golang.Ident,
			Name:     "g",
			Filename: "test_b.go",
			Pos:      1,
			End:      2,
		},
	}

	dups := [][]*syntax.Node{dupA, dupB}

	clones, err := ProcessClones(noopReadFile, dups)
	if err != nil {
		t.Fatalf("ProcessClones failed: %v", err)
	}

	for i, c := range clones {
		if c.Classification.GenericsCandidate {
			t.Errorf("clone %d: expected GenericsCandidate=false (no VarType data)", i)
		}
	}
}

func noopReadFile(filename string) ([]byte, error) {
	return []byte("package p\n"), nil
}
