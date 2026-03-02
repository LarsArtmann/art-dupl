package suffixtree

import (
	"testing"
)

// BenchmarkMemoryUsageFewTokens measures memory with few unique tokens
func BenchmarkMemoryUsageFewTokens(b *testing.B) {
	b.ReportAllocs()
	tokens := make([]Token, 10000)
	for i := range tokens {
		tokens[i] = &testToken{val: i % 50} // 50 unique
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree := New()
		tree.Update(tokens...)
	}
}

// BenchmarkMemoryUsageManyTokens measures memory with many unique tokens
func BenchmarkMemoryUsageManyTokens(b *testing.B) {
	b.ReportAllocs()
	tokens := make([]Token, 10000)
	for i := range tokens {
		tokens[i] = &testToken{val: i % 5000} // 5000 unique
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree := New()
		tree.Update(tokens...)
	}
}
