package printer

import (
	"bytes"
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

// TestTextCloneOutputGolden locks the exact byte format of the default text
// output, including the one-line code preview introduced in the Tier 1
// feedback sprint.
//
// Regenerate after intentional format changes via:
//
//	go test -run TestTextCloneOutputGolden -args -update ./printer/
func TestTextCloneOutputGolden(t *testing.T) {
	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(""))

	group := domain.ProcessedCloneGroup{
		Hash: "abc123",
		Clones: []domain.ProcessedClone{
			newTestProcessedClone("alpha.go", 10, 15, "result := x + y\nreturn result"),
			newTestProcessedClone("beta.go", 20, 25, "sum := a + b\nreturn sum"),
		},
	}

	if err := p.PrintHeader(); err != nil {
		t.Fatalf("PrintHeader() error: %v", err)
	}

	if err := p.PrintClones(group); err != nil {
		t.Fatalf("PrintClones() error: %v", err)
	}

	if err := p.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter() error: %v", err)
	}

	testutil.RequireGolden(t, buf.Bytes())
}

// TestTextCloneOutputGolden_Truncated locks the truncation behavior when the
// first source line exceeds maxPreviewRunes (60). The preview ends with "…"
// and the golden file captures the exact truncation point.
func TestTextCloneOutputGolden_Truncated(t *testing.T) {
	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(""))

	longLine := "result := computeSomethingVeryLong(this, is, a, really, long, expression, that, exceeds, max, preview, runes"
	group := domain.ProcessedCloneGroup{
		Hash: "long1",
		Clones: []domain.ProcessedClone{
			newTestProcessedClone("wide.go", 1, 2, longLine),
		},
	}

	if err := p.PrintHeader(); err != nil {
		t.Fatalf("PrintHeader() error: %v", err)
	}

	if err := p.PrintClones(group); err != nil {
		t.Fatalf("PrintClones() error: %v", err)
	}

	if err := p.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter() error: %v", err)
	}

	testutil.RequireGolden(t, buf.Bytes())
}

// TestTextCloneOutputGolden_FileFallback locks the ReadFile fallback path used
// when a clone's Fragment is empty. The preview is read from the line at
// LineStart. Guards against drift in the fallback formatting.
func TestTextCloneOutputGolden_FileFallback(t *testing.T) {
	var buf bytes.Buffer

	// 1-indexed LineStart=3 returns "third line" (file below has empty line 1,
	// "second" on line 2, "third line" on line 3, "fourth" on line 4).
	fileContent := "\nsecond\nthird line\nfourth\n"

	p := NewText(&buf, mockReadFile(fileContent))

	group := domain.ProcessedCloneGroup{
		Hash: "fallback",
		Clones: []domain.ProcessedClone{
			newTestProcessedClone("fallback.go", 3, 4, ""),
		},
	}

	if err := p.PrintHeader(); err != nil {
		t.Fatalf("PrintHeader() error: %v", err)
	}

	if err := p.PrintClones(group); err != nil {
		t.Fatalf("PrintClones() error: %v", err)
	}

	if err := p.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter() error: %v", err)
	}

	testutil.RequireGolden(t, buf.Bytes())
}
