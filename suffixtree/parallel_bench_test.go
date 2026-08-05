package suffixtree

import (
	"context"
	"fmt"
	"runtime"
	"testing"
)

// BenchmarkFindDuplOverParallel compares sequential vs parallel search
// across different tree sizes and worker counts.
func BenchmarkFindDuplOverParallel(b *testing.B) {
	sizes := []int{1000, 5000, 10000}

	for _, size := range sizes {
		// Build a tree with intentional duplicates for meaningful search.
		half := genTokenSequence(size)
		data := append(append(half, half...), char(0))

		tree := New()
		mustUpdate(tree, data...)

		threshold := 10

		// Sequential baseline
		b.Run(fmt.Sprintf("seq/tokens_%d", size), func(b *testing.B) {
			b.ResetTimer()
			for range b.N {
				for range tree.FindDuplOver(context.Background(), threshold) {
				}
			}
		})

		// Parallel with different worker counts
		for _, workers := range []int{2, 4, runtime.NumCPU()} {
			b.Run(fmt.Sprintf("par%d/tokens_%d", workers, size), func(b *testing.B) {
				b.ResetTimer()
				for range b.N {
					for range tree.FindDuplOverParallel(context.Background(), threshold, workers) {
					}
				}
			})
		}
	}
}

// BenchmarkTokenValueMemory measures memory allocation reduction from storing
// TokenValue (4 bytes) instead of Token interface (16 bytes) in the tree.
// The tree's data slice should be ~4x smaller than if it stored interfaces.
func BenchmarkTokenValueMemory(b *testing.B) {
	b.ReportAllocs()

	for range b.N {
		tree := New()
		tokens := generateRandomTokens(10000)
		mustUpdate(tree, tokens...)
	}
}
