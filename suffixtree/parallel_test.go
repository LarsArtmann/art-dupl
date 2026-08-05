package suffixtree

import (
	"context"
	"sort"
	"testing"
	"time"
)

type matchKey struct {
	len Pos
	ps  string
}

func matchSetKey(m Match) matchKey {
	ps := make([]int, len(m.Ps))

	for i, p := range m.Ps {
		ps[i] = int(p)
	}

	sort.Ints(ps)

	return matchKey{m.Len, sortedString(ps)}
}

func sortedString(vals []int) string {
	sb := make([]byte, 0, 2*len(vals))

	for _, v := range vals {
		sb = append(sb, byte(v), ',')
	}

	return string(sb)
}

func TestParallelFindsSameMatchesAsSequential(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		seq       string
		threshold int
		workers   int
	}{
		{"abab threshold 2", "abab$", 2, 4},
		{"abcbcabc threshold 2", "abcbcabc$", 2, 4},
		{"abcbcabc threshold 3", "abcbcabc$", 3, 2},
		{"long repeat", "abcabcabcabcabcabc$", 3, 8},
		{"many duplicates", "aabbccaabbccaabbcc$", 2, 4},
		{"single worker", "abcbcabc$", 2, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			data := str2tok(tc.seq)
			tree := New()
			mustUpdate(tree, data...)

			seqMatches := collectMatches(tree.FindDuplOver(context.Background(), tc.threshold))
			parMatches := collectMatches(tree.FindDuplOverParallel(context.Background(), tc.threshold, tc.workers))

			seqSet := make(map[matchKey]bool)
			for _, m := range seqMatches {
				seqSet[matchSetKey(m)] = true
			}

			parSet := make(map[matchKey]bool)
			for _, m := range parMatches {
				parSet[matchSetKey(m)] = true
			}

			for key := range seqSet {
				if !parSet[key] {
					t.Errorf("parallel search missing match present in sequential: %+v", key)
				}
			}

			for key := range parSet {
				if !seqSet[key] {
					t.Errorf("parallel search has extra match not in sequential: %+v", key)
				}
			}
		})
	}
}

func TestParallelEmptyTree(t *testing.T) {
	t.Parallel()

	tree := New()
	ch := tree.FindDuplOverParallel(context.Background(), 2, 4)

	count := 0
	for range ch {
		count++
	}

	if count != 0 {
		t.Errorf("empty tree produced %d matches (expected 0)", count)
	}
}

func TestParallelContextCancellation(t *testing.T) {
	t.Parallel()

	tree := New()
	data := str2tok("abcdefghijabcdefghijabcdefghij$")
	mustUpdate(tree, data...)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	ch := tree.FindDuplOverParallel(ctx, 2, 4)
	for range ch {
	}
}

func TestParallelThresholdZero(t *testing.T) {
	t.Parallel()

	tree := New()
	data := str2tok("abab$")
	mustUpdate(tree, data...)

	ch := tree.FindDuplOverParallel(context.Background(), 0, 2)

	count := 0
	for range ch {
		count++
		if count > 1000 {
			t.Fatal("threshold 0 produced too many matches (possible infinite loop)")
		}
	}
}

func TestParallelWorkersAuto(t *testing.T) {
	t.Parallel()

	tree := New()
	data := str2tok("abababab$")
	mustUpdate(tree, data...)

	ch := tree.FindDuplOverParallel(context.Background(), 2, 0)

	count := 0
	for range ch {
		count++
	}

	if count == 0 {
		t.Error("auto workers produced no matches")
	}
}

func TestParallelLargeTree(t *testing.T) {
	t.Parallel()

	tree := New()
	tokens := make([]Token, 0, 2000)

	for i := range 100 {
		c := char(rune('A' + i%26))
		tokens = append(tokens, c, char('x'), char('y'), char('z'))
	}

	mustUpdate(tree, tokens...)

	seqMatches := collectMatches(tree.FindDuplOver(context.Background(), 3))
	parMatches := collectMatches(tree.FindDuplOverParallel(context.Background(), 3, 4))

	if len(seqMatches) != len(parMatches) {
		t.Errorf("match count mismatch: sequential=%d parallel=%d", len(seqMatches), len(parMatches))
	}
}

func TestParallelChannelCloses(t *testing.T) {
	t.Parallel()

	tree := New()
	data := str2tok("abcdefghijabcdefghijabcdefghijabcdefghij$")
	mustUpdate(tree, data...)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	ch := tree.FindDuplOverParallel(ctx, 2, 4)
	for range ch {
	}
}
