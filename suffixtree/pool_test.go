package suffixtree

import (
	"slices"
	"testing"
)

// TestContextListPoolSliceSurvival verifies the pool ownership invariant used
// throughout walkTrans: cl.append(cl2) transfers cl2's []Pos slice headers
// into cl's map, after which releaseContextList(cl2) must NOT invalidate
// them. releaseContextList clears only cl2's map entries — the backing
// arrays survive because cl still holds slice headers referencing them.
//
// If releaseContextList or contextList.append is ever refactored to recycle
// the []Pos backing arrays themselves, this test fails.
func TestContextListPoolSliceSurvival(t *testing.T) {
	t.Parallel()

	cl := acquireContextList()
	defer releaseContextList(cl)

	cl2 := acquireContextList()

	cl.lists[10] = []Pos{1, 2}
	cl2.lists[10] = []Pos{7, 8}
	cl2.lists[20] = []Pos{9}

	cl.append(cl2)
	releaseContextList(cl2)

	// NOTE: cl2 must not be touched after release — another goroutine may
	// acquire it from the global pool at any moment. That contract is exactly
	// what the survival check below relies on: only cl's copied slice headers
	// keep the []Pos backing arrays alive.

	// getAll groups positions by sorted key: key 10 then key 20.
	got := cl.getAll()
	want := []Pos{1, 2, 7, 8, 9}

	if !slices.Equal(got, want) {
		t.Fatalf(
			"positions lost after releasing the appended contextList: got %v, want %v",
			got, want,
		)
	}
}

// TestContextListPoolAppendOverwrite verifies that appending a contextList
// whose key already exists concatenates the positions under that key rather
// than replacing them.
func TestContextListPoolAppendOverwrite(t *testing.T) {
	t.Parallel()

	cl := acquireContextList()
	defer releaseContextList(cl)

	cl2 := acquireContextList()
	defer releaseContextList(cl2)

	cl.lists[5] = []Pos{1, 2, 3}
	cl2.lists[5] = []Pos{4, 5}

	cl.append(cl2)

	if got := cl.lists[5]; !slices.Equal(got, []Pos{1, 2, 3, 4, 5}) {
		t.Errorf("append did not concatenate existing key: got %v, want [1 2 3 4 5]", got)
	}
}

// TestContextListPoolReuse verifies a released contextList is handed back in
// a reusable state: no stale map entries, still functional.
func TestContextListPoolReuse(t *testing.T) {
	t.Parallel()

	first := acquireContextList()
	first.lists[1] = []Pos{100}
	releaseContextList(first)

	second := acquireContextList()
	defer releaseContextList(second)

	if len(second.lists) != 0 {
		t.Errorf("reused contextList has stale entries: %v", second.lists)
	}

	second.lists[2] = []Pos{200}

	if got := second.getAll(); !slices.Equal(got, []Pos{200}) {
		t.Errorf("reused contextList not functional: got %v, want [200]", got)
	}
}
