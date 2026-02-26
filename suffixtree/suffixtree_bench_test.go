package suffixtree

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

// generateRandomTokens creates a slice of random token values.
func generateRandomTokens(count int) []Token {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	tokens := make([]Token, count)
	for i := range tokens {
		tokens[i] = &testToken{val: r.Intn(1000)}
	}

	return tokens
}

// testToken implements Token interface for benchmarking.
type testToken struct {
	val int
}

func (t *testToken) Val() int {
	return t.val
}

// generateTreeWithRandomTokens creates a tree with the specified number of random tokens.
func generateTreeWithRandomTokens(tokenCount int) *STree {
	tree := New()
	tokens := generateRandomTokens(tokenCount)
	tree.Update(tokens...)

	return tree
}

// generateTreeWithTransitions creates a tree with states having
// a specified number of transitions for benchmarking.
func generateTreeWithTransitions(stateCount, transPerState int) *STree {
	tree := New()

	// Add tokens to build tree structure
	tokens := generateRandomTokens(stateCount * transPerState * 10)
	tree.Update(tokens...)

	return tree
}

// benchmarkFindTranMethod is a helper for benchmarking s.findTran with different parameters
func benchmarkFindTranMethod(b *testing.B, stateCount, transPerState int) {
	benchmarkFindTran(b, stateCount, transPerState, nil, func(s *state, t Token) *tran {
		return s.findTran(t)
	})
}

// benchmarkFindTran is a helper for benchmarking findTran functions.
// If setup is nil, uses default tree generation with stateCount and transPerState.
func benchmarkFindTran(
	b *testing.B,
	stateCount, transPerState int,
	setup func() *STree,
	fn func(*state, Token) *tran,
) {
	b.Helper()

	var tree *STree
	if setup != nil {
		tree = setup()
	} else {
		tree = generateTreeWithTransitions(stateCount, transPerState)
	}

	if len(tree.root.trans) == 0 {
		b.Skip("No transitions to benchmark")
	}

	target := tree.root.trans[0]
	token := tree.data[target.start]

	for b.Loop() {
		fn(tree.root, token)
	}
}

// BenchmarkFindTranSmall benchmarks findTran with a small number of transitions.
func BenchmarkFindTranSmall(b *testing.B) {
	benchmarkFindTran(b, 0, 0, func() *STree {
		tree := New()
		tokens := generateRandomTokens(100)
		tree.Update(tokens...)

		return tree
	}, func(s *state, t Token) *tran {
		return s.findTran(t)
	})
}

// BenchmarkFindTranMedium benchmarks findTran with a medium number of transitions.
func BenchmarkFindTranMedium(b *testing.B) {
	benchmarkFindTranMethod(b, 50, 10)
}

// BenchmarkFindTranLarge benchmarks findTran with a large number of transitions.
func BenchmarkFindTranLarge(b *testing.B) {
	benchmarkFindTranMethod(b, 100, 50)
}

// BenchmarkFindTranVeryLarge benchmarks findTran with a very large number of transitions.
func BenchmarkFindTranVeryLarge(b *testing.B) {
	benchmarkFindTranMethod(b, 200, 100)
}

// BenchmarkFindTranFallback benchmarks the fallback (linear) implementation.
func BenchmarkFindTranFallback(b *testing.B) {
	benchmarkFindTran(b, 100, 50, nil, func(s *state, t Token) *tran {
		return s.findTranFallback(t)
	})
}

// benchmarkTreeOperation benchmarks tree-level operations with standard setup.
func benchmarkTreeOperation(b *testing.B, tokenCount int, operation func(*STree)) {
	b.Helper()

	tree := generateTreeWithRandomTokens(tokenCount)

	for b.Loop() {
		operation(tree)
	}
}

// BenchmarkConstruction benchmarks full suffix tree construction.
func BenchmarkConstruction(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size%d", size), func(b *testing.B) {
			tokens := generateRandomTokens(size)

			b.ResetTimer()

			for b.Loop() {
				tree := New()
				tree.Update(tokens...)
			}
		})
	}
}

// BenchmarkConstructionParallel benchmarks parallel tree construction.
func BenchmarkConstructionParallel(b *testing.B) {
	sizes := []int{10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size%d", size), func(b *testing.B) {
			tokens := generateRandomTokens(size)

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					tree := New()
					tree.Update(tokens...)
				}
			})
		})
	}
}

// BenchmarkCanonize benchmarks the canonize operation.
func BenchmarkCanonize(b *testing.B) {
	benchmarkTreeOperation(b, 1000, func(tree *STree) {
		_, _, _ = tree.canonize(tree.root, 0, 100) // benchmark
	})
}

// BenchmarkUpdate benchmarks incremental tree updates.
func BenchmarkUpdate(b *testing.B) {
	batchSizes := []int{1, 10, 100}

	for _, batchSize := range batchSizes {
		b.Run(fmt.Sprintf("Batch%d", batchSize), func(b *testing.B) {
			tree := New()
			tokens := generateRandomTokens(batchSize)

			b.ResetTimer()

			for b.Loop() {
				tree.Update(tokens...)
			}
		})
	}
}

// BenchmarkAt benchmarks random access to tree data.
func BenchmarkAt(b *testing.B) {
	tree := New()
	tokens := generateRandomTokens(10000)
	tree.Update(tokens...)

	positions := make([]int, 100)

	r := rand.New(rand.NewSource(42))
	for i := range positions {
		positions[i] = r.Intn(len(tokens))
	}

	for b.Loop() {
		for _, pos := range positions {
			tree.At(Pos(pos))
		}
	}
}

// BenchmarkMemoryUsage measures memory usage during tree construction.
func BenchmarkMemoryUsage(b *testing.B) {
	b.ReportAllocs()

	tokens := generateRandomTokens(10000)

	for b.Loop() {
		tree := New()
		tree.Update(tokens...)
	}
}

// BenchmarkTestAndSplit benchmarks the testAndSplit operation.
func BenchmarkTestAndSplit(b *testing.B) {
	benchmarkTreeOperation(b, 1000, func(tree *STree) {
		// Use valid parameters within bounds
		if len(tree.data) > 10 {
			tree.testAndSplit(tree.root, 0, 10)
		}
	})
}

// BenchmarkSearch benchmarks searching for specific patterns in the tree.
func BenchmarkSearch(b *testing.B) {
	tree := New()
	tokens := generateRandomTokens(1000)
	tree.Update(tokens...)

	if len(tokens) < 10 {
		b.Skip("Not enough tokens for search benchmark")
	}

	// Search for the first 10 tokens
	searchTokens := tokens[:10]

	for b.Loop() {
		// Simulate a search by traversing transitions
		s := tree.root
		for _, token := range searchTokens {
			tr := s.findTran(token)
			if tr == nil {
				break
			}

			s = tr.state
		}
	}
}
