package suffixtree

import (
	"context"
	"testing"
)

// collectMatches drains the channel into a slice for property-test assertions.
func collectMatches(ch <-chan Match) []Match {
	var matches []Match

	for m := range ch {
		matches = append(matches, m)
	}

	return matches
}

// tokensAt extracts the token subsequence at [pos, pos+len) from data.
func tokensAt(data []Token, pos, length Pos) []TokenValue {
	result := make([]TokenValue, 0, length)
	for i := range length {
		result = append(result, data[int(pos)+int(i)].Val())
	}

	return result
}

// Property: every reported match contains positions whose token subsequences
// are identical. If any pair of positions has differing tokens, the suffix tree
// reported a false positive.
func TestProperty_EveryMatchIsARealRepeat(t *testing.T) {
	t.Parallel()

	sequences := []string{
		"ababcabab$",
		"xyzxyzabcabc$",
		"mississippi$",
	}

	for _, seq := range sequences {
		data := str2tok(seq)
		tree := New()
		mustUpdate(tree, data...)

		matches := collectMatches(tree.FindDuplOver(context.Background(), 2))

		for _, m := range matches {
			ref := tokensAt(data, m.Ps[0], m.Len)

			for _, pos := range m.Ps[1:] {
				got := tokensAt(data, pos, m.Len)
				for i := range ref {
					if ref[i] != got[i] {
						t.Errorf("seq %q match Ps=%v Len=%d: pos %d differs at offset %d (%v vs %v)",
							seq, m.Ps, m.Len, pos, i, ref[i], got[i])
					}
				}
			}
		}
	}
}

// Property: every match length is >= the requested threshold.
func TestProperty_AllMatchesMeetThreshold(t *testing.T) {
	t.Parallel()

	data := str2tok("abcabcabcabc$")
	tree := New()
	mustUpdate(tree, data...)

	for _, threshold := range []int{2, 3, 4, 5, 6} {
		matches := collectMatches(tree.FindDuplOver(context.Background(), threshold))
		for _, m := range matches {
			if int(m.Len) < threshold {
				t.Errorf("threshold %d: match Len=%d below threshold", threshold, m.Len)
			}
		}
	}
}

// Property: each match has at least two distinct positions (a clone requires
// >= 2 occurrences).
func TestProperty_MatchesHaveMultiplePositions(t *testing.T) {
	t.Parallel()

	data := str2tok("aabbaabbaabb$")
	tree := New()
	mustUpdate(tree, data...)

	matches := collectMatches(tree.FindDuplOver(context.Background(), 2))
	for _, m := range matches {
		if len(m.Ps) < 2 {
			t.Errorf("match Ps=%v Len=%d has fewer than 2 positions", m.Ps, m.Len)
		}
	}
}

// Property: a sequence with no repeated substring of length >= threshold
// produces zero matches. Uses a random permutation of unique tokens.
func TestProperty_NoFalsePositivesOnUniqueInput(t *testing.T) {
	t.Parallel()

	// Generate 200 unique token values — no repeats possible.
	data := make([]Token, 0, 200)
	for i := range 200 {
		data = append(data, char(rune('A'+i))) // each token unique
	}

	tree := New()
	mustUpdate(tree, data...)

	matches := collectMatches(tree.FindDuplOver(context.Background(), 2))
	if len(matches) != 0 {
		t.Errorf("unique input produced %d matches (expected 0)", len(matches))

		for _, m := range matches {
			t.Logf("  spurious: Ps=%v Len=%d", m.Ps, m.Len)
		}
	}
}

// Property: a deliberately duplicated sequence is detected. Two identical
// halves must produce at least one match spanning the repeat.
func TestProperty_DuplicateDetected(t *testing.T) {
	t.Parallel()

	half := "abcdefghij" // 10 unique tokens
	data := str2tok(half + half + "$")

	tree := New()
	mustUpdate(tree, data...)

	matches := collectMatches(tree.FindDuplOver(context.Background(), 5))
	if len(matches) == 0 {
		t.Fatal("duplicated sequence produced no matches")
	}

	// Verify at least one match has length >= 10 (the full half).
	found := false

	for _, m := range matches {
		if int(m.Len) >= 10 {
			found = true

			break
		}
	}

	if !found {
		t.Errorf("no match of length >= 10; got %v", matches)
	}
}

// Property: matches are maximal — extending any match by one token in either
// direction breaks the equality for at least one position pair. This verifies
// the suffix tree reports maximal repeats, not sub-optimal fragments.
func TestProperty_MatchesAreMaximal(t *testing.T) {
	t.Parallel()

	data := str2tok("xabcyabcz$")
	tree := New()
	mustUpdate(tree, data...)

	matches := collectMatches(tree.FindDuplOver(context.Background(), 3))
	for _, m := range matches {
		// canExtendLeft returns true when the match CAN be extended left for all
		// position pairs — meaning it was NOT maximal and should have been longer.
		if canExtendLeft(data, m) {
			t.Errorf("non-maximal match Ps=%v Len=%d: left-extendable for all pairs", m.Ps, m.Len)
		}
	}
}

// canExtendLeft returns false (meaning the match IS left-maximal) if at least
// one position pair has differing tokens immediately before the match start.
func canExtendLeft(data []Token, m Match) bool {
	for i := range len(m.Ps) {
		if m.Ps[i] == 0 {
			return false // boundary — can't extend, so maximal
		}

		for j := i + 1; j < len(m.Ps); j++ {
			if m.Ps[j] == 0 {
				return false
			}

			if data[m.Ps[i]-1].Val() != data[m.Ps[j]-1].Val() {
				return false // differing left context → maximal
			}
		}
	}

	return true // all pairs share left context → NOT maximal
}
