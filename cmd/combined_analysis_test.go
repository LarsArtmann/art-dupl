package cmd

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// mergeTestFile has four lines of 8 bytes each ("line one\n" ...).
const mergeTestFileContent = "line one\nline two\nline three\nline four\n"

func writeMergeFixture(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "fixture.go")

	if err := os.WriteFile(path, []byte(mergeTestFileContent), 0o600); err != nil {
		t.Fatal(err)
	}

	return path
}

// fragOverLines builds a one-node fragment whose byte span covers the given
// 1-based line range of the fixture (each line is 9 bytes).
func fragOverLines(path string, fromLine, toLine int) []*syntax.Node {
	start := int32((fromLine - 1) * 9)
	end := int32(toLine * 9)

	return []*syntax.Node{{Filename: path, Pos: start, End: end}}
}

func TestMergeCombinedMatchesDropsSubMinLinesGenericsFamilies(t *testing.T) {
	t.Parallel()

	path := writeMergeFixture(t)

	// Divergent types at two positions so the structural gate passes; the
	// fragments span only 1-2 lines, below the min-lines gate (4).
	divergent := func(from, to int) []*syntax.Node {
		frag := fragOverLines(path, from, to)
		frag[0].VarType = "int"
		// A child with a differing VarType across frags gives ≥2 divergent
		// positions when the other fragment uses other types.
		frag[0].Children = []*syntax.Node{{VarType: "string"}}
		return frag
	}

	other := func(from, to int) []*syntax.Node {
		frag := fragOverLines(path, from, to)
		frag[0].VarType = "float64"
		frag[0].Children = []*syntax.Node{{VarType: "bool"}}
		return frag
	}

	typeAware := make(chan syntax.Match, 1)
	typeAware <- syntax.Match{Frags: [][]*syntax.Node{fragOverLines(path, 1, 1), fragOverLines(path, 2, 2)}}
	close(typeAware)

	generics := make(chan syntax.Match, 2)
	generics <- syntax.Match{Frags: [][]*syntax.Node{divergent(1, 2), other(3, 4)}}
	generics <- syntax.Match{Frags: [][]*syntax.Node{divergent(1, 4), other(1, 4)}}
	close(generics)

	merged := mergeCombinedMatches(context.Background(), typeAwareChanFor(typeAware), generics, 4)

	var got []syntax.Match
	for match := range merged {
		got = append(got, match)
	}

	// The type-aware match always flows; of the two generics families only
	// the 4-line one survives — the short cross-type family can never carry
	// the generics marker and must not print as a plain clone group.
	if len(got) != 2 {
		t.Fatalf("expected 2 merged matches (1 type-aware + 1 long generics), got %d", len(got))
	}

	seenFourLine := false

	for _, match := range got {
		for _, frag := range match.Frags {
			if frag[0].VarType != "" && frag[0].End-frag[0].Pos >= 4*9 {
				seenFourLine = true
			}
		}
	}

	if !seenFourLine {
		t.Error("the >=minLines divergent family must survive the merge")
	}
}

func TestMergeCombinedMatchesMinLinesZeroKeepsShortFamilies(t *testing.T) {
	t.Parallel()

	path := writeMergeFixture(t)

	fragA := fragOverLines(path, 1, 1)
	fragA[0].VarType = "int"
	fragA[0].Children = []*syntax.Node{{VarType: "string"}}

	fragB := fragOverLines(path, 2, 2)
	fragB[0].VarType = "float64"
	fragB[0].Children = []*syntax.Node{{VarType: "bool"}}

	typeAware := make(chan syntax.Match)
	close(typeAware)

	generics := make(chan syntax.Match, 1)
	generics <- syntax.Match{Frags: [][]*syntax.Node{fragA, fragB}}
	close(generics)

	merged := mergeCombinedMatches(context.Background(), typeAware, generics, 0)

	count := 0
	for range merged {
		count++
	}

	if count != 1 {
		t.Fatalf("minLines=0 disables the line gate: the short divergent family must flow, got %d", count)
	}
}

// typeAwareChanFor is an identity helper keeping the first test's channel
// naming explicit at the call site.
func typeAwareChanFor(ch chan syntax.Match) <-chan syntax.Match { return ch }
