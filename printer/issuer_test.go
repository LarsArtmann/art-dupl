package printer

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// assertToFilename checks that issue.To filenames match expected values.
func assertToFilename(t *testing.T, issues []Issue, expected []string) {
	t.Helper()

	for i, issue := range issues {
		if issue.To.Filename() != expected[i] {
			t.Errorf("Issue[%d].To.Filename() = %q, want %q", i, issue.To.Filename(), expected[i])
		}
	}
}

// assertFromFilename checks that all issue.From filenames match the expected value.
func assertFromFilename(t *testing.T, issues []Issue, expected string) {
	t.Helper()

	for i, issue := range issues {
		if issue.From.Filename() != expected {
			t.Errorf("Issue[%d].From.Filename() = %q, want %q", i, issue.From.Filename(), expected)
		}
	}
}

// createIssuerTestNodes creates nodes with proper positions for issuer tests.
func createIssuerTestNodes(filenames ...string) [][]*syntax.Node {
	result := make([][]*syntax.Node, len(filenames))
	for i, filename := range filenames {
		result[i] = []*syntax.Node{
			{Type: 1, Filename: filename, Pos: 0, End: 10},
		}
	}

	return result
}

// createIssuerReadFile creates a ReadFile that returns content matching positions.
func createIssuerReadFile() ReadFile {
	content := "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\n"
	data := []byte(content)

	return func(path string) ([]byte, error) {
		return data, nil
	}
}

func TestMakeIssues_NoBidirectionalPairs(t *testing.T) {
	// Create syntax nodes representing 2 duplicate code blocks
	nodes := createIssuerTestNodes("a.go", "b.go")

	issuer := NewIssuer(createIssuerReadFile())

	issues, err := issuer.MakeIssues(nodes)
	if err != nil {
		t.Fatalf("MakeIssues() error = %v", err)
	}

	// With 2 clones, we should get exactly 1 issue (not 2 bidirectional)
	testutil.AssertCount(t, len(issues), 1, "MakeIssues()")
}

func TestMakeIssues_ThreeClones(t *testing.T) {
	nodes := createIssuerTestNodes("a.go", "b.go", "c.go")

	issuer := NewIssuer(createIssuerReadFile())

	issues, err := issuer.MakeIssues(nodes)
	if err != nil {
		t.Fatalf("MakeIssues() error = %v", err)
	}

	// With 3 clones, we should get exactly 2 issues (a→b, a→c)
	// NOT 3 issues with circular wrapping (a→b, b→c, c→a)
	testutil.AssertCount(t, len(issues), 2, "MakeIssues()")

	// All issues should have the first clone (a.go) as From
	assertFromFilename(t, issues, "a.go")

	// To should be b.go and c.go (sorted)
	assertToFilename(t, issues, []string{"b.go", "c.go"})
}

func TestMakeIssues_SingleClone(t *testing.T) {
	// Single clone should produce no issues (nothing to compare to)
	nodes := createIssuerTestNodes("a.go")

	issuer := NewIssuer(createIssuerReadFile())

	issues, err := issuer.MakeIssues(nodes)
	if err != nil {
		t.Fatalf("MakeIssues() error = %v", err)
	}

	testutil.AssertCount(t, len(issues), 0, "MakeIssues() for single clone")
}

func TestMakeIssues_SortedByFilename(t *testing.T) {
	// Clones should be sorted by filename/line, so "a.go" becomes the reference
	nodes := createIssuerTestNodes("z.go", "a.go", "m.go")

	issuer := NewIssuer(createIssuerReadFile())

	issues, err := issuer.MakeIssues(nodes)
	if err != nil {
		t.Fatalf("MakeIssues() error = %v", err)
	}

	// After sorting, "a.go" should be the reference (From)
	assertFromFilename(t, issues, "a.go")

	// To should be "m.go" and "z.go" (sorted order)
	assertToFilename(t, issues, []string{"m.go", "z.go"})
}

func TestMakeIssues_FourClones(t *testing.T) {
	nodes := createIssuerTestNodes("a.go", "b.go", "c.go", "d.go")

	issuer := NewIssuer(createIssuerReadFile())

	issues, err := issuer.MakeIssues(nodes)
	if err != nil {
		t.Fatalf("MakeIssues() error = %v", err)
	}

	// With 4 clones, we should get exactly 3 issues (a→b, a→c, a→d)
	testutil.AssertCount(t, len(issues), 3, "MakeIssues()")

	// All issues should have the first clone (a.go) as From
	assertFromFilename(t, issues, "a.go")

	// To should be b.go, c.go, and d.go (sorted)
	assertToFilename(t, issues, []string{"b.go", "c.go", "d.go"})
}
