package suffixtree

import (
	"testing"
)

// benchmarkMemoryUsage is a helper that benchmarks memory usage with the given number of unique tokens.
func benchmarkMemoryUsage(b *testing.B, uniqueCount int) {
	b.Helper()
	b.ReportAllocs()

	tokens := make([]Token, 0, memoryUsageTokens)
	for i := range memoryUsageTokens {
		tokens = append(tokens, &testToken{val: i % uniqueCount})
	}

	for b.Loop() {
		tree := New()
		mustUpdate(tree, tokens...)
	}
}

// Benchmark sizes for the memory-usage suite: few unique tokens stress
// transition-list reuse; many unique tokens stress state fan-out.
const (
	memoryUsageFewUniqueTokens  = 50
	memoryUsageManyUniqueTokens = 5000
)

// BenchmarkMemoryUsageFewTokens measures memory with few unique tokens.
func BenchmarkMemoryUsageFewTokens(b *testing.B) {
	benchmarkMemoryUsage(b, memoryUsageFewUniqueTokens)
}

// BenchmarkMemoryUsageManyTokens measures memory with many unique tokens.
func BenchmarkMemoryUsageManyTokens(b *testing.B) {
	benchmarkMemoryUsage(b, memoryUsageManyUniqueTokens)
}
