package domain

import (
	"testing"
)

// BenchmarkCloneMemory measures the memory usage of a Clone struct
func BenchmarkCloneMemory(b *testing.B) {
	// Create a sample clone
	clone := Clone{}
	clone.SetFilename("test.go")
	clone.SetFragment("func main() {}")
	clone.SetHash("abc123")
	
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Create a copy to measure allocation
		_ = clone
	}
}

// BenchmarkCloneCreation measures the cost of creating a new Clone
func BenchmarkCloneCreation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		clone := Clone{}
		clone.SetFilename("test.go")
		clone.SetFragment("func main() {}")
		clone.SetHash("abc123")
		_ = clone
	}
}

// BenchmarkCloneSliceMemory measures memory for a slice of clones
func BenchmarkCloneSliceMemory(b *testing.B) {
	// Pre-create clones
	clones := make([]Clone, 1000)
	for i := range clones {
		clones[i].SetFilename("test.go")
		clones[i].SetFragment("func main() {}")
		clones[i].SetHash("abc123")
	}
	
	b.ReportAllocs()
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// Create a copy of the slice
		_ = clones
	}
}
