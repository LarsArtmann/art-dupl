package suffixtree

import (
	"context"
	"slices"
	"sort"
	"testing"
	"unicode/utf8"
)

// TestFindDuplOverExceedsStackThreshold verifies that FindDuplOver produces
// correct results when the suffix tree's root state has more than
// maxStackKeys (32) transitions. This exercises the heap-allocation fallback
// path in walkTrans where the transition map exceeds the stack buffer.
//
// The input uses 40 distinct characters so the root state has 40 transitions
// (>32), forcing walkTrans to use make([]TokenValue, ...) instead of the
// stack-allocated [32]TokenValue buffer. A 5-character prefix is repeated to
// create a detectable duplicate.
func TestFindDuplOverExceedsStackThreshold(t *testing.T) {
	t.Parallel()

	// 40 distinct characters — root state will have 40 transitions (>maxStackKeys).
	chars := "abcdefghijklmnopqrstuvwxyz0123456789!@#$"
	if utf8.RuneCountInString(chars) != 40 {
		t.Fatalf("test setup: expected 40 distinct chars, got %d", utf8.RuneCountInString(chars))
	}

	// Repeat the first 5 chars to create a duplicate of length 5 at positions 0 and 40.
	input := chars + chars[:5] + "$"

	tree := New()
	mustUpdate(tree, str2tok(input)...)

	ch := tree.FindDuplOver(context.Background(), 5)

	var matches []Match
	for m := range ch {
		matches = append(matches, m)
	}

	if len(matches) == 0 {
		t.Fatal("expected at least one match with >32 root transitions, got none")
	}

	// The duplicate "abcde" appears at positions 0 and 40.
	found := false

	for _, m := range matches {
		if m.Len == 5 {
			sortedPs := make([]Pos, len(m.Ps))
			copy(sortedPs, m.Ps)
			slices.Sort(sortedPs)

			if len(sortedPs) == 2 && sortedPs[0] == 0 && sortedPs[1] == 40 {
				found = true

				break
			}
		}
	}

	if !found {
		t.Errorf("expected match at positions {0, 40} with length 5; got %v", matches)
	}
}

// TestContextListGetAllExceedsStackThreshold verifies that contextList.getAll()
// returns all positions correctly when the map has more than maxStackKeys
// entries, exercising the heap-allocation fallback path.
func TestContextListGetAllExceedsStackThreshold(t *testing.T) {
	t.Parallel()

	cl := newContextList()

	// Add maxStackKeys + 8 entries to exceed the stack buffer threshold.
	numEntries := maxStackKeys + 8

	type entry struct {
		key  TokenValue
		poss []Pos
	}

	entries := make([]entry, numEntries)
	for i := range numEntries {
		entries[i] = entry{
			key:  TokenValue(i + 1),
			poss: []Pos{Pos(i * 10), Pos(i*10 + 1)},
		}
		pl := newPosList()
		pl.add(entries[i].poss[0])
		pl.add(entries[i].poss[1])
		cl.lists[entries[i].key] = pl
	}

	if len(cl.lists) != numEntries {
		t.Fatalf("setup: expected %d entries, got %d", numEntries, len(cl.lists))
	}

	// getAll() should return all positions, sorted by key.
	result := cl.getAll()

	expectedCount := numEntries * 2
	if len(result) != expectedCount {
		t.Fatalf("getAll() returned %d positions, expected %d", len(result), expectedCount)
	}

	// Verify positions are grouped by key in sorted order.
	idx := 0

	sortedEntries := make([]entry, numEntries)
	copy(sortedEntries, entries)
	sort.Slice(sortedEntries, func(i, j int) bool { return sortedEntries[i].key < sortedEntries[j].key })

	for _, e := range sortedEntries {
		for _, expectedPos := range e.poss {
			if result[idx] != expectedPos {
				t.Errorf(
					"position at index %d: got %d, expected %d (key=%d)",
					idx, result[idx], expectedPos, e.key,
				)
			}

			idx++
		}
	}
}

// TestWalkTransStackAndHeapPathsProduceSameResults verifies that the stack
// buffer path (<=maxStackKeys transitions) and the heap fallback path
// (>maxStackKeys transitions) produce identical match results for the same
// duplicate pattern. This is the strongest correctness guarantee: regardless
// of which code path executes, the output must be the same.
func TestWalkTransStackAndHeapPathsProduceSameResults(t *testing.T) {
	t.Parallel()

	// A 5-char duplicate pattern "abcde" repeated twice.
	dup := "abcde"

	// Input with < 32 distinct chars (stack path): just the duplicate + sentinel.
	smallInput := dup + dup + "$"

	// Input with > 32 distinct chars (heap path for root): 40 unique chars + dup + sentinel.
	extraChars := "fghijklmnopqrstuvwxyz0123456789!@#$"
	largeInput := extraChars + dup + dup + "$"

	treeSmall := New()
	mustUpdate(treeSmall, str2tok(smallInput)...)

	treeLarge := New()
	mustUpdate(treeLarge, str2tok(largeInput)...)

	chSmall := treeSmall.FindDuplOver(context.Background(), 5)
	chLarge := treeLarge.FindDuplOver(context.Background(), 5)

	// Collect matches from both trees.
	collectMatches := func(ch <-chan Match) []Match {
		var matches []Match

		for m := range ch {
			// Only keep matches with length 5 (the "abcde" duplicate).
			if m.Len == 5 {
				sortedPs := make([]Pos, len(m.Ps))
				copy(sortedPs, m.Ps)
				slices.Sort(sortedPs)
				m.Ps = sortedPs
				matches = append(matches, m)
			}
		}

		return matches
	}

	smallMatches := collectMatches(chSmall)
	largeMatches := collectMatches(chLarge)

	// Both trees should find exactly one match of length 5 with 2 positions.
	if len(smallMatches) != 1 {
		t.Fatalf("small tree: expected 1 match of length 5, got %d: %v", len(smallMatches), smallMatches)
	}

	if len(largeMatches) != 1 {
		t.Fatalf("large tree: expected 1 match of length 5, got %d: %v", len(largeMatches), largeMatches)
	}

	// Both matches should have the same length and number of positions.
	if smallMatches[0].Len != largeMatches[0].Len {
		t.Errorf(
			"match length differs: stack path=%d, heap path=%d",
			smallMatches[0].Len, largeMatches[0].Len,
		)
	}

	if len(smallMatches[0].Ps) != len(largeMatches[0].Ps) {
		t.Errorf(
			"match position count differs: stack path=%d, heap path=%d",
			len(smallMatches[0].Ps), len(largeMatches[0].Ps),
		)
	}
}
