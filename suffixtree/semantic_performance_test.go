package suffixtree

import (
	"testing"
)

// Benchmark with few unique tokens (like without semantic)
func BenchmarkTreeConstructionFewUniqueTokens(b *testing.B) {
	// Simulate AST without semantic - few unique types
	tokens := make([]Token, 10000)
	for i := range tokens {
		// Only 50 unique token types
		tokens[i] = &testToken{val: i % 50}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree := New()
		tree.Update(tokens...)
	}
}

// Benchmark with many unique tokens (like with semantic)
func BenchmarkTreeConstructionManyUniqueTokens(b *testing.B) {
	// Simulate AST with semantic - many unique types
	tokens := make([]Token, 10000)
	for i := range tokens {
		// 5000 unique token types (identifiers)
		tokens[i] = &testToken{val: i % 5000}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree := New()
		tree.Update(tokens...)
	}
}
