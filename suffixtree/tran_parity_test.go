package suffixtree

import (
	"testing"
)

// This file proves the ADR-0022 sorted-slice transition layout implements the
// same lookup semantics as the pre-refactor map[TokenValue]*tran layout:
// for every state of every tree, findTran must return exactly what a naive
// full scan (the map's semantics) returns, and the slice invariants that the
// binary-search branch depends on (strictly ascending, unique keys) must hold.

// referenceFindTran is the naive reference lookup: a full linear scan over
// ALL transitions, independent of the linear-scan/binary-search cutoff that
// findTran applies. If findTran's binary-search branch had an off-by-one,
// this would disagree on high-fanout states.
func referenceFindTran(s *state, data []TokenValue, c TokenValue) *tran {
	for i := range s.trans {
		if data[s.trans[i].start] == c {
			return &s.trans[i]
		}
	}

	return nil
}

// enumerateStates returns all states reachable from the root via transitions
// (breadth-first). The auxiliary state is excluded: it is a sentinel, not a
// tree state.
func enumerateStates(t *STree) []*state {
	seen := map[*state]bool{t.root: true}
	queue := []*state{t.root}

	for len(queue) > 0 {
		s := queue[0]
		queue = queue[1:]

		for i := range s.trans {
			next := s.trans[i].state
			if next != nil && !seen[next] {
				seen[next] = true
				queue = append(queue, next)
			}
		}
	}

	states := make([]*state, 0, len(seen))
	for s := range seen {
		states = append(states, s)
	}

	return states
}

// checkTranParity verifies the two slice invariants binary search relies on
// (keys strictly ascending and unique) and that findTran agrees with
// referenceFindTran for every present key and every near-miss probe.
func checkTranParity(t *testing.T, tree *STree) {
	t.Helper()

	for _, s := range enumerateStates(tree) {
		prev := TokenValue(0)
		hasPrev := false

		for i := range s.trans {
			key := tree.data[s.trans[i].start]

			if hasPrev && key <= prev {
				t.Fatalf("state transitions not strictly ascending: "+
					"key[%d]=%d <= key[%d]=%d (invariant binary search depends on)",
					i, key, i-1, prev)
			}

			prev, hasPrev = key, true

			// Present key: both lookups must return the SAME transition.
			got := s.findTran(tree.data, key)
			want := referenceFindTran(s, tree.data, key)

			if got != want {
				t.Fatalf("findTran(%d) = %+v, reference = %+v (present key must hit)", key, got, want)
			}

			// Near-miss probes around each boundary catch off-by-one bugs.
			for _, probe := range []TokenValue{key - 1, key + 1} {
				got := s.findTran(tree.data, probe)
				want := referenceFindTran(s, tree.data, probe)

				if got != want {
					t.Fatalf("findTran(%d) = %+v, reference = %+v (probe near key %d)",
						probe, got, want, key)
				}
			}
		}

		// Far-out probes: both must miss.
		for _, probe := range []TokenValue{0, 1, -1} {
			if s.findTran(tree.data, probe) != referenceFindTran(s, tree.data, probe) {
				t.Fatalf("findTran(%d) disagrees with reference on far probe", probe)
			}
		}
	}
}

// TestProperty_TranParityDeterministicStreams builds trees from fixed
// adversarial streams (runs, repeats, alphabets) and checks parity.
func TestProperty_TranParityDeterministicStreams(t *testing.T) {
	t.Parallel()

	sequences := [][]Token{
		// single run: one fanout-1 path
		tokensOf([]TokenValue{1, 1, 1, 1, 1, 1, 1, 1}),
		// all distinct: maximal root fanout
		tokensOf([]TokenValue{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}),
		// descending pairs
		tokensOf([]TokenValue{5, 5, 4, 4, 3, 3, 2, 2, 1, 1}),
		// mixed run lengths
		tokensOf([]TokenValue{7, 7, 7, 3, 3, 9, 9, 9, 9, 1}),
		// pseudo-random 26-letter alphabet
		genTokenSequence(200),
	}

	for i, seq := range sequences {
		tree := New()
		mustUpdate(tree, seq...)

		t.Run(seqName(i), func(t *testing.T) {
			t.Parallel()

			checkTranParity(t, tree)
		})
	}
}

// TestProperty_TranParityRandomStreams runs 1000 seeded pseudo-random
// streams with varying alphabets (small alphabets force high-fanout states,
// crossing the linearScanMax cutoff into the binary-search branch).
func TestProperty_TranParityRandomStreams(t *testing.T) {
	t.Parallel()

	const cases = 1000

	binaryBranchStates := 0

	for i := 1; i <= cases; i++ {
		stream := lcgStream(uint32(i), 30+i%70, 2+i%40)

		tree := New()
		mustUpdate(tree, valueTokens(stream)...)

		checkTranParity(t, tree)

		for _, s := range enumerateStates(tree) {
			if len(s.trans) > linearScanMax {
				binaryBranchStates++
			}
		}
	}

	// Guard that the battery actually exercises the binary-search branch:
	// a green run that never crossed the linearScanMax cutoff proves nothing
	// about that branch.
	if binaryBranchStates == 0 {
		t.Fatal("random battery never produced a state above linearScanMax; binary-search branch untested")
	}
}

// lcgStream generates a deterministic pseudo-random TokenValue stream:
// a fixed LCG (matching genTokenSequence's approach) seeded per case, with
// the given length and alphabet size. Small alphabets (2-4 values) create
// states with many transitions; larger ones spread the tree.
func lcgStream(seed uint32, length, alphabet int) []TokenValue {
	out := make([]TokenValue, 0, length)

	s := seed
	for range length {
		s = s*1103515245 + 12345
		out = append(out, TokenValue(1+(s%uint32(alphabet))))
	}

	return out
}

type valueToken TokenValue

func (v valueToken) Val() TokenValue { return TokenValue(v) }

func valueTokens(vals []TokenValue) []Token {
	out := make([]Token, 0, len(vals))
	for _, v := range vals {
		out = append(out, valueToken(v))
	}

	return out
}

// tokensOf converts plain values to Tokens for table-driven test inputs.
func tokensOf(vals []TokenValue) []Token {
	return valueTokens(vals)
}

func seqName(i int) string {
	switch i {
	case 0:
		return "single_run"
	case 1:
		return "all_distinct"
	case 2:
		return "descending_pairs"
	case 3:
		return "mixed_runs"
	default:
		return "pseudo_random"
	}
}
