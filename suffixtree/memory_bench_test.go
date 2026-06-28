package suffixtree

import (
	"testing"
)

// benchmarkMemoryUsage is a helper that benchmarks memory usage with the given number of unique tokens.
func benchmarkMemoryUsage(b *testing.B, uniqueCount int) {
	b.Helper()
	b.ReportAllocs()

	tokens := make([]Token, 0, 10000)
	for i := range 10000 {
		tokens = append(tokens, &testToken{val: i % uniqueCount})
	}

	b.ResetTimer()

	for range b.N {
		tree := New()
		tree.Update(tokens...)
	}
}

// BenchmarkMemoryUsageFewTokens measures memory with few unique tokens.
func BenchmarkMemoryUsageFewTokens(b *testing.B) {
	benchmarkMemoryUsage(b, 50) // 50 unique
}

// BenchmarkMemoryUsageManyTokens measures memory with many unique tokens.
func BenchmarkMemoryUsageManyTokens(b *testing.B) {
	benchmarkMemoryUsage(b, 5000) // 5000 unique
}
