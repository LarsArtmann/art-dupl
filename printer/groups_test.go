package printer

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax"
)

func TestGetCloneSize(t *testing.T) {
	t.Parallel()

	if got := GetCloneSize(nil); got != 0 {
		t.Errorf("GetCloneSize(nil) = %d, want 0", got)
	}

	if got := GetCloneSize([][]*syntax.Node{}); got != 0 {
		t.Errorf("GetCloneSize(empty) = %d, want 0", got)
	}

	if got := GetCloneSize([][]*syntax.Node{{}}); got != 0 {
		t.Errorf("GetCloneSize({empty group}) = %d, want 0", got)
	}

	node := &syntax.Node{Owns: 42}
	group := [][]*syntax.Node{{node}}

	if got := GetCloneSize(group); got != 42 {
		t.Errorf("GetCloneSize() = %d, want 42", got)
	}
}

func TestBuildCloneGroups(t *testing.T) {
	t.Parallel()

	ch := make(chan syntax.Match, 2)
	nodeA := &syntax.Node{Filename: "a.go"}
	nodeB := &syntax.Node{Filename: "b.go"}

	ch <- syntax.Match{Hash: "abc", Frags: [][]*syntax.Node{{nodeA}}}

	ch <- syntax.Match{Hash: "abc", Frags: [][]*syntax.Node{{nodeB}}}

	close(ch)

	groups := BuildCloneGroups(ch)
	if len(groups) != 1 {
		t.Fatalf("BuildCloneGroups() returned %d groups, want 1", len(groups))
	}

	if len(groups["abc"]) != 2 {
		t.Errorf("group 'abc' has %d entries, want 2", len(groups["abc"]))
	}
}

func TestBuildCloneGroups_Empty(t *testing.T) {
	t.Parallel()

	ch := make(chan syntax.Match)
	close(ch)

	groups := BuildCloneGroups(ch)
	if len(groups) != 0 {
		t.Errorf("BuildCloneGroups(empty) = %d groups, want 0", len(groups))
	}
}

func TestComputeUniqueCounts(t *testing.T) {
	t.Parallel()

	nodeA := &syntax.Node{Filename: "a.go"}
	nodeA2 := &syntax.Node{Filename: "a.go"}
	nodeB := &syntax.Node{Filename: "b.go"}
	nodeC := &syntax.Node{Filename: "c.go"}

	groups := map[string][][]*syntax.Node{
		"h1": {{nodeA}, {nodeA2}, {nodeB}},
		"h2": {{nodeC}},
	}

	totals := ComputeUniqueCounts(groups)
	testutil.AssertFieldValue(t, totals["h1"], 2, "h1 unique count")
	testutil.AssertFieldValue(t, totals["h2"], 1, "h2 unique count")
}

func TestSortCloneGroupKeys(t *testing.T) {
	t.Parallel()

	smallNode := &syntax.Node{Owns: 5}
	bigNode := &syntax.Node{Owns: 50}
	medNode := &syntax.Node{Owns: 20}

	groups := map[string][][]*syntax.Node{
		healthSmall:  {{smallNode}},
		"big":        {{bigNode}, {smallNode}, {medNode}},
		healthMedium: {{medNode}, {smallNode}},
	}
	uniqueCounts := map[string]int{healthSmall: 1, "big": 3, healthMedium: 2}

	keys := []string{healthSmall, "big", healthMedium}
	SortCloneGroupKeys(keys, config.SortBySize, groups, uniqueCounts)

	testutil.AssertFieldValue(t, keys[0], "big", "config.SortBySize first key")

	keys = []string{healthSmall, "big", healthMedium}
	SortCloneGroupKeys(keys, config.SortByOccurrence, groups, uniqueCounts)

	testutil.AssertFieldValue(t, keys[0], "big", "config.SortByOccurrence first key")

	keys = []string{healthSmall, "big", healthMedium}
	SortCloneGroupKeys(keys, config.SortByHash, groups, uniqueCounts)

	testutil.AssertFieldValue(t, keys[0], "big", "config.SortByHash first key")
}
