package printer

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
)

const lowTokenHighLineFilename = "widget.templ"

// TestProcessClones_LowTokenHighLine_NotIdiom verifies that clones spanning
// many source lines but composed of few syntax nodes are not dismissed as
// idioms. This is the common case for templ files, where one Element node can
// cover many lines of HTML attributes.
func TestProcessClones_LowTokenHighLine_NotIdiom(t *testing.T) {
	t.Parallel()

	content := []byte(`line01
line02
line03
line04
line05
line06
line07
line08
line09
line10
line11
line12
line13`)

	fread := mockReadFile(string(content))

	// Four nodes spanning bytes 0..89, which maps to lines 1..13.
	nodes := []*syntax.Node{
		{Type: 1, Filename: lowTokenHighLineFilename, Pos: 0, End: 1},
		{Type: 2, Filename: lowTokenHighLineFilename, Pos: 1, End: 2},
		{Type: 3, Filename: lowTokenHighLineFilename, Pos: 2, End: 3},
		{Type: 4, Filename: lowTokenHighLineFilename, Pos: 3, End: 89},
	}

	clones, err := ProcessClones(fread, [][]*syntax.Node{nodes})
	if err != nil {
		t.Fatalf("ProcessClones() error: %v", err)
	}

	if len(clones) != 1 {
		t.Fatalf("ProcessClones() returned %d clones, want 1", len(clones))
	}

	clone := clones[0]

	if clone.LineCount() != 13 {
		t.Errorf("LineCount() = %d, want 13", clone.LineCount())
	}

	if clone.TokenCount != 4 {
		t.Errorf("TokenCount = %d, want 4", clone.TokenCount)
	}

	if clone.Classification.Category == domain.CategoryIdiom {
		t.Errorf(
			"Category = %v, want non-idiom for a %d-token/%d-line clone",
			clone.Classification.Category,
			clone.TokenCount,
			clone.LineCount(),
		)
	}
}
