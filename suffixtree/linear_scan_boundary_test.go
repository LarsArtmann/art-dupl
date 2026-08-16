package suffixtree

import (
	"testing"
)

// This file pins the linearScanMax boundary: states with exactly
// linearScanMax transitions take the linear-scan branch, states with one
// more take the binary-search branch. The pair of benchmarks makes the
// crossover measurable whenever linearScanMax is revisited.

// stateWithExactTransitions builds a tree whose root has EXACTLY n
// transitions with known, well-spread keys. Keys are placed in t.data in
// ASCENDING order so addTran's sorted insertion sees the easy case; key
// spacing (i*7) avoids accidental cache-line aliasing between probes.
func stateWithExactTransitions(n int) (*STree, *state) {
	tree := New()

	for i := 1; i <= n; i++ {
		tree.data = append(tree.data, TokenValue(i*7))
		tree.addTran(tree.root, Pos(i-1), Pos(i-1), nil)
	}

	return tree, tree.root
}

// BenchmarkFindTranBoundary8 is the last linear-scan size (== linearScanMax).
func BenchmarkFindTranBoundary8(b *testing.B) {
	benchmarkBoundaryLookup(b, linearScanMax)
}

// BenchmarkFindTranBoundary9 is the first binary-search size
// (== linearScanMax+1).
func BenchmarkFindTranBoundary9(b *testing.B) {
	benchmarkBoundaryLookup(b, linearScanMax+1)
}

func benchmarkBoundaryLookup(b *testing.B, n int) {
	b.Helper()

	tree, root := stateWithExactTransitions(n)

	if got := len(root.trans); got != n {
		b.Fatalf("setup: root has %d transitions, want exactly %d", got, n)
	}

	// Probe keys: every present key (hit at every position) plus misses
	// between keys. Probing all of them per iteration keeps the benchmark
	// representative of both branches' behavior across the slice.
	probes := make([]TokenValue, 0, n*2)
	for i := 1; i <= n; i++ {
		probes = append(probes, TokenValue(i*7))   // hit
		probes = append(probes, TokenValue(i*7+1)) // miss just above
	}

	b.ReportAllocs()

	sink = 0

	for b.Loop() {
		for _, p := range probes {
			if tr := root.findTran(tree.data, p); tr != nil {
				sink++
			}
		}
	}
}

// TestBoundaryStatesUseExpectedBranch verifies the two benchmark fixtures
// actually straddle the cutoff and answer lookups correctly — otherwise the
// benchmarks measure nothing.
func TestBoundaryStatesUseExpectedBranch(t *testing.T) {
	t.Parallel()

	tree8, root8 := stateWithExactTransitions(linearScanMax)
	if len(root8.trans) != linearScanMax {
		t.Fatalf("8-fixture has %d transitions, want exactly %d", len(root8.trans), linearScanMax)
	}

	tree9, root9 := stateWithExactTransitions(linearScanMax + 1)
	if len(root9.trans) != linearScanMax+1 {
		t.Fatalf("9-fixture has %d transitions, want exactly %d", len(root9.trans), linearScanMax+1)
	}

	// Every present key must hit; a key just above every gap must miss.
	for i := 1; i <= linearScanMax; i++ {
		key := TokenValue(i * 7)

		if root8.findTran(tree8.data, key) == nil {
			t.Fatalf("8-fixture: findTran(%d) missed a present key", key)
		}

		if root9.findTran(tree9.data, key) == nil {
			t.Fatalf("9-fixture: findTran(%d) missed a present key", key)
		}
	}

	if got := root9.findTran(tree9.data, TokenValue((linearScanMax+1)*7)); got == nil {
		t.Fatal("9-fixture: findTran missed the last present key")
	}

	if got := root8.findTran(tree8.data, TokenValue(1000000)); got != nil {
		t.Fatalf("8-fixture: findTran(impossible) = %+v, want nil", got)
	}

	if got := root9.findTran(tree9.data, TokenValue(1000000)); got != nil {
		t.Fatalf("9-fixture: findTran(impossible) = %+v, want nil", got)
	}
}

// sink prevents the compiler from eliminating the benchmarked lookups.
var sink int
