package syntax

import (
	"errors"
	"testing"
)

// dedupTestFile is the shared source used by member-set dedup tests. Two
// fragments at different columns of the same lines form the line-level
// duplicate case; a wider fragment elsewhere forms the distinct case.
const dedupTestFile = "line one\nline two\nline three\nline four\n"

func dedupReadFile(name string) ([]byte, error) {
	if name == "a.go" {
		return []byte(dedupTestFile), nil
	}

	return nil, errors.New("no such file: " + name)
}

// fragAt builds a one-node fragment in a.go spanning [start, end) bytes.
func fragAt(start, end int32, owns int32) []*Node {
	return []*Node{{Filename: "a.go", Pos: start, End: end, Owns: owns}}
}

func TestDeduplicateMemberSetsNoOpOnSingleGroup(t *testing.T) {
	t.Parallel()

	groups := map[string][][]*Node{
		"hash1": {fragAt(0, 8, 2), fragAt(9, 17, 2)},
	}

	got := DeduplicateMemberSets(dedupReadFile, groups)
	if len(got) != 1 || got["hash1"] == nil {
		t.Fatalf("single group must survive unchanged, got %v", got)
	}
}

func TestDeduplicateMemberSetsIdenticalByteRangesKeepsRichest(t *testing.T) {
	t.Parallel()

	// Two views of the same byte range on line 1: the richer one covers
	// more stream tokens (Owns+1) and must win.
	groups := map[string][][]*Node{
		"thin":  {fragAt(0, 8, 2), fragAt(9, 17, 2)},
		"rich":  {fragAt(0, 8, 6), fragAt(9, 17, 6)},
		"other": {fragAt(18, 26, 2), fragAt(27, 35, 2)},
	}

	got := DeduplicateMemberSets(dedupReadFile, groups)
	if _, ok := got["rich"]; !ok {
		t.Error("richer view must win the duplicate member set")
	}

	if _, ok := got["thin"]; ok {
		t.Error("thinner duplicate view must be dropped")
	}

	if _, ok := got["other"]; !ok {
		t.Error("distinct member set must be kept")
	}
}

// TestDeduplicateMemberSetsLineLevelCollision is the issue #10 shape: two
// groups whose byte ranges are DISJOINT but whose displayed line ranges are
// identical — the param fragment `(days []history.WatchTrendDay)` vs the
// result fragment of the same signature line. A reader sees the same
// file:line-line entries twice; only the richer view may print.
func TestDeduplicateMemberSetsLineLevelCollision(t *testing.T) {
	t.Parallel()

	groups := map[string][][]*Node{
		// "line one" columns 0-3 and columns 4-8, same displayed line.
		"left":  {fragAt(0, 4, 1), fragAt(9, 13, 1)},
		"right": {fragAt(4, 8, 3), fragAt(13, 17, 3)},
	}

	got := DeduplicateMemberSets(dedupReadFile, groups)
	if len(got) != 1 {
		t.Fatalf("line-level duplicates must collapse to one group, got %d: %v", len(got), keysOf(got))
	}

	if _, ok := got["right"]; !ok {
		t.Errorf("richer view must win, got %v", keysOf(got))
	}
}

func TestDeduplicateMemberSetsDifferentLinesKept(t *testing.T) {
	t.Parallel()

	groups := map[string][][]*Node{
		"one": {fragAt(0, 8, 2), fragAt(18, 26, 2)},
		"two": {fragAt(9, 17, 2), fragAt(27, 35, 2)},
	}

	got := DeduplicateMemberSets(dedupReadFile, groups)
	if len(got) != 2 {
		t.Fatalf("groups on different lines must both survive, got %v", keysOf(got))
	}
}

func TestDeduplicateMemberSetsUnreadableFilesAreExempt(t *testing.T) {
	t.Parallel()

	readFile := func(string) ([]byte, error) { return nil, errors.New("gone") }

	// Same byte ranges, but the file cannot be read: the dedup must not
	// drop either group on unknowable line locations.
	groups := map[string][][]*Node{
		"a": {fragAt(0, 8, 2), fragAt(9, 17, 2)},
		"b": {fragAt(0, 8, 6), fragAt(9, 17, 6)},
	}

	got := DeduplicateMemberSets(readFile, groups)
	if len(got) != 2 {
		t.Fatalf("unreadable member files must be exempt from dedup, got %v", keysOf(got))
	}
}

func TestDeduplicateMemberSetsDeterministicTieBreak(t *testing.T) {
	t.Parallel()

	groups := map[string][][]*Node{
		"zzzz": {fragAt(0, 8, 2), fragAt(9, 17, 2)},
		"aaaa": {fragAt(0, 8, 2), fragAt(9, 17, 2)},
	}

	got := DeduplicateMemberSets(dedupReadFile, groups)
	if len(got) != 1 {
		t.Fatalf("equal-richness duplicates must collapse, got %v", keysOf(got))
	}

	if _, ok := got["aaaa"]; !ok {
		t.Errorf("deterministic tie-break must keep the smaller hash, got %v", keysOf(got))
	}
}

func keysOf(groups map[string][][]*Node) []string {
	out := make([]string, 0, len(groups))
	for k := range groups {
		out = append(out, k)
	}

	return out
}
