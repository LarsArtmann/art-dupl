package printer

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

const overlapTestFile = "test.go"

func TestEliminateOverlaps_NestedCloneSuppressed(t *testing.T) {
	t.Parallel()

	// Group A: large clone (pos 0-100)
	// Group B: small clone nested inside A (pos 10-30)
	// B should be suppressed because it's contained in A
	largeFrag := []*syntax.Node{
		{Type: 1, Pos: 0, End: 50, Filename: overlapTestFile, Owns: 10},
		{Type: 2, Pos: 50, End: 100, Filename: overlapTestFile},
	}
	smallFrag := []*syntax.Node{
		{Type: 1, Pos: 10, End: 20, Filename: overlapTestFile, Owns: 2},
		{Type: 2, Pos: 20, End: 30, Filename: overlapTestFile},
	}

	groups := map[string][][]*syntax.Node{
		"large": {largeFrag},
		"small": {smallFrag},
	}

	result := EliminateOverlaps(groups)

	if _, exists := result["large"]; !exists {
		t.Error("Expected large clone group to be preserved")
	}

	if _, exists := result["small"]; exists {
		t.Error("Expected small (nested) clone group to be eliminated")
	}
}

func TestEliminateOverlaps_NonOverlappingKept(t *testing.T) {
	t.Parallel()

	// art-dupl: accepted: Pos/End values ARE the test; each node literal exercises a distinct overlap scenario. A helper would take more params than duplicated lines.
	// Two clones in the same file at different positions — both should be kept
	fragA := []*syntax.Node{
		{Type: 1, Pos: 0, End: 50, Filename: overlapTestFile, Owns: 5},
	}
	fragB := []*syntax.Node{
		{Type: 1, Pos: 100, End: 150, Filename: overlapTestFile, Owns: 5},
	}

	groups := map[string][][]*syntax.Node{
		"A": {fragA},
		"B": {fragB},
	}

	result := EliminateOverlaps(groups)

	if len(result) != 2 {
		t.Errorf("Expected 2 groups (non-overlapping), got %d", len(result))
	}
}

func TestEliminateOverlaps_IdenticalRangeKept(t *testing.T) {
	t.Parallel()

	// Two clones at the exact same range but different hashes — both kept
	fragA := []*syntax.Node{
		{Type: 1, Pos: 0, End: 50, Filename: overlapTestFile, Owns: 5},
	}
	fragB := []*syntax.Node{
		{Type: 2, Pos: 0, End: 50, Filename: overlapTestFile, Owns: 5},
	}

	groups := map[string][][]*syntax.Node{
		"A": {fragA},
		"B": {fragB},
	}

	result := EliminateOverlaps(groups)

	if len(result) != 2 {
		t.Errorf("Expected 2 groups (identical range, different hash), got %d", len(result))
	}
}

func TestEliminateOverlaps_DifferentFilesKept(t *testing.T) {
	t.Parallel()

	// Same range but different files — both kept
	fragA := []*syntax.Node{
		{Type: 1, Pos: 0, End: 50, Filename: "a.go", Owns: 5},
	}
	fragB := []*syntax.Node{
		{Type: 1, Pos: 0, End: 50, Filename: "b.go", Owns: 5},
	}

	groups := map[string][][]*syntax.Node{
		"A": {fragA},
		"B": {fragB},
	}

	result := EliminateOverlaps(groups)

	if len(result) != 2 {
		t.Errorf("Expected 2 groups (different files), got %d", len(result))
	}
}
