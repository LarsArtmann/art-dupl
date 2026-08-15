package printer

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// multiLineContent provides 6 source lines so node byte ranges can span 4+
// lines, exercising the generics line gate.
const multiLineContent = "l0\nl1\nl2\nl3\nl4\nl5\n"

func multiLineReadFile(filename string) ([]byte, error) {
	return []byte(multiLineContent), nil
}

func noopReadFile(filename string) ([]byte, error) {
	return []byte("package p\n"), nil
}

// divergentSeqNodes builds a two-instance duplicate group whose flattened node
// positions carry two VarType divergences (the classification minimum), with a
// byte range spanning lines 2..5 (4 lines) of multiLineContent.
func divergentSeqNodes() [][]*syntax.Node {
	makeNodes := func(filename, typeA, typeB string) []*syntax.Node {
		return []*syntax.Node{
			{
				Type:     golang.Ident,
				Name:     "a",
				VarType:  typeA,
				Filename: filename,
				Pos:      3, // byte offset of "l1" — line 2
				End:      4,
			},
			{
				Type:     golang.Ident,
				Name:     "b",
				VarType:  typeB,
				Filename: filename,
				Pos:      6,
				End:      13, // byte offset inside "l4" — line 5
			},
		}
	}

	return [][]*syntax.Node{
		makeNodes("test_a.go", "github.com/example/db.AuthorKindActivity", "db.AuthorLabel"),
		makeNodes("test_b.go", "github.com/example/db.MemberGrowthPoint", "db.MemberAlias"),
	}
}

// TestProcessClones_GenericsCandidateSet verifies that ProcessClones correctly
// sets GenericsCandidate and GenericsHint on sufficiently long clone groups
// whose instances have differing VarType values at multiple positions.
func TestProcessClones_GenericsCandidateSet(t *testing.T) {
	t.Parallel()

	clones, err := ProcessClones(multiLineReadFile, divergentSeqNodes())
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

// TestProcessClones_GenericsLineGate_DefaultRejectsShortClones verifies that
// the default min-lines gate (DefaultGenericsMinLines) suppresses generics
// candidacy for 1-line clones even when types diverge at two positions.
func TestProcessClones_GenericsLineGate_DefaultRejectsShortClones(t *testing.T) {
	t.Parallel()

	clones, err := ProcessClones(noopReadFile, divergentSeqNodes())
	if err != nil {
		t.Fatalf("ProcessClones failed: %v", err)
	}

	for i, c := range clones {
		if c.Classification.GenericsCandidate {
			t.Errorf("clone %d: expected GenericsCandidate=false (clone spans 1 line, below default gate)", i)
		}
	}
}

// TestProcessClones_GenericsLineGate_DisabledByZeroOption verifies that
// WithGenericsMinLines(0) turns the line gate off, restoring ungated
// classification for calibration.
func TestProcessClones_GenericsLineGate_DisabledByZeroOption(t *testing.T) {
	t.Parallel()

	clones, err := ProcessClones(noopReadFile, divergentSeqNodes(), WithGenericsMinLines(0))
	if err != nil {
		t.Fatalf("ProcessClones failed: %v", err)
	}

	for i, c := range clones {
		if !c.Classification.GenericsCandidate {
			t.Errorf("clone %d: expected GenericsCandidate=true (gate disabled)", i)
		}
	}
}

// TestProcessClones_GenericsLineGate_CustomThreshold verifies that a custom
// gate value lower than the clone span admits the candidate.
func TestProcessClones_GenericsLineGate_CustomThreshold(t *testing.T) {
	t.Parallel()

	clones, err := ProcessClones(noopReadFile, divergentSeqNodes(), WithGenericsMinLines(1))
	if err != nil {
		t.Fatalf("ProcessClones failed: %v", err)
	}

	for i, c := range clones {
		if !c.Classification.GenericsCandidate {
			t.Errorf("clone %d: expected GenericsCandidate=true (gate 1 <= 1-line clone)", i)
		}
	}
}

// TestProcessClones_GenericsPatternDisqualifies verifies that groups matching
// an actionability boilerplate pattern are never generics candidates: each
// clone here is a lone CallExpr (single-call-expression pattern) whose
// arguments have divergent types — the exact noise class the cross-reference
// eliminates.
func TestProcessClones_GenericsPatternDisqualifies(t *testing.T) {
	t.Parallel()

	makeCall := func(filename, typeA, typeB string) []*syntax.Node {
		return []*syntax.Node{
			{
				Type:     golang.CallExpr,
				Name:     "convert",
				Filename: filename,
				Pos:      3,
				End:      13,
				Children: []*syntax.Node{
					{Type: golang.Ident, Name: "arg0", VarType: typeA, Filename: filename},
					{Type: golang.Ident, Name: "arg1", VarType: typeB, Filename: filename},
				},
			},
		}
	}

	dups := [][]*syntax.Node{
		makeCall("test_a.go", "db.Author", "db.AuthorLabel"),
		makeCall("test_b.go", "db.Member", "db.MemberAlias"),
	}

	clones, err := ProcessClones(multiLineReadFile, dups, WithGenericsMinLines(0))
	if err != nil {
		t.Fatalf("ProcessClones failed: %v", err)
	}

	for i, c := range clones {
		if c.Classification.GenericsCandidate {
			t.Errorf("clone %d: expected GenericsCandidate=false (single-call-expression boilerplate)", i)
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
		t.Fatal(err)
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
		t.Fatal(err)
	}

	for i, c := range clones {
		if c.Classification.GenericsCandidate {
			t.Errorf("clone %d: expected GenericsCandidate=false (no VarType data)", i)
		}
	}
}
