package suffixtree

import (
	"context"
	"strings"
	"testing"
)

// BenchmarkSTreeUpdate measures suffix tree construction from token sequences
// of increasing size. Run with: go test -bench=. -benchmem ./suffixtree/.
func BenchmarkSTreeUpdate(b *testing.B) {
	sizes := []int{100, 500, 2000}
	for _, size := range sizes {
		b.Run("tokens_"+itoa(size), func(b *testing.B) {
			data := genTokenSequence(size)

			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				tree := New()
				tree.Update(data...)
			}
		})
	}
}

// BenchmarkFindDuplOver measures duplicate finding at various thresholds.
func BenchmarkFindDuplOver(b *testing.B) {
	// Create a sequence with intentional duplicates: half + half + sentinel.
	half := genTokenSequence(500)
	data := append(append(half, half...), char(0))

	tree := New()
	tree.Update(data...)

	thresholds := []int{10, 50, 200}
	for _, t := range thresholds {
		b.Run("threshold_"+itoa(t), func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()

			for range b.N {
				for range tree.FindDuplOver(context.Background(), t) {
				}
			}
		})
	}
}

// genTokenSequence generates a pseudo-random sequence of tokens for benchmarking.
// Uses a simple LCG to avoid importing math/rand in benchmarks.
func genTokenSequence(n int) []Token {
	data := make([]Token, 0, n)

	seed := uint32(42)
	for range n {
		seed = seed*1103515245 + 12345
		data = append(data, char(rune('a'+seed%26)))
	}

	return data
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	var sb strings.Builder
	for n > 0 {
		sb.WriteByte(byte('0' + n%10))
		n /= 10
	}

	s := sb.String()
	// reverse
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}

	return string(runes)
}
