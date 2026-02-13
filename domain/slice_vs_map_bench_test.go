package domain_test

import (
	"fmt"
	"testing"
)

// BenchmarkSliceLookup measures slice index access.
func BenchmarkSliceLookup(b *testing.B) {
	pool := make([]string, 1000)
	for i := range pool {
		pool[i] = fmt.Sprintf("file_%d.go", i)
	}

	b.ResetTimer()
	var i int
	for b.Loop() {
		_ = pool[i%1000] // Slice access
		i++
	}
}

// BenchmarkMapLookup measures map access with uint32 key.
func BenchmarkMapLookup(b *testing.B) {
	pool := make(map[uint32]string, 1000)
	for i := range uint32(1000) {
		pool[i] = fmt.Sprintf("file_%d.go", i)
	}

	b.ResetTimer()
	var i int
	for b.Loop() {
		_ = pool[uint32(i%1000)] // Map access
		i++
	}
}

// BenchmarkSliceMemory measures memory overhead.
func BenchmarkSliceMemory(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = make([]string, 1000) // 1000 string slots
	}
}

// BenchmarkMapMemory measures memory overhead.
func BenchmarkMapMemory(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = make(map[uint32]string, 1000) // 1000 map slots
	}
}

// TestSliceMapEquivalence tests both approaches work.
func TestSliceMapEquivalence(t *testing.T) {
	// Slice + Map approach
	slice := make([]string, 10)
	index := make(map[string]uint32)
	for i := range 10 {
		s := fmt.Sprintf("string_%d", i)
		slice[i] = s
		index[s] = uint32(i)
	}

	// Test round-trip
	testString := "string_5"
	id := index[testString] // string → ID
	retrieved := slice[id]  // ID → string

	if retrieved != testString {
		t.Errorf("Round-trip failed: expected %s, got %s", testString, retrieved)
	}

	// Test all strings
	for i := range 10 {
		s := fmt.Sprintf("string_%d", i)
		id := index[s]
		if slice[id] != s {
			t.Errorf("String %s round-trip failed", s)
		}
	}
}

// TestMapMapEquivalence tests double map approach.
func TestMapMapEquivalence(t *testing.T) {
	// Double map approach
	strToID := make(map[string]uint32)
	idToStr := make(map[uint32]string)

	for i := range 10 {
		s := fmt.Sprintf("string_%d", i)
		id := uint32(i)
		strToID[s] = id
		idToStr[id] = s
	}

	// Test round-trip
	testString := "string_5"
	id := strToID[testString] // string → ID
	retrieved := idToStr[id]  // ID → string

	if retrieved != testString {
		t.Errorf("Round-trip failed: expected %s, got %s", testString, retrieved)
	}

	// Test all strings
	for i := range 10 {
		s := fmt.Sprintf("string_%d", i)
		id := strToID[s]
		if idToStr[id] != s {
			t.Errorf("String %s round-trip failed", s)
		}
	}
}

// Benchmark round-trip performance.
func BenchmarkRoundTripSliceMap(b *testing.B) {
	slice := make([]string, 1000)
	index := make(map[string]uint32)
	for i := range 1000 {
		s := fmt.Sprintf("file_%d.go", i)
		slice[i] = s
		index[s] = uint32(i)
	}

	b.ResetTimer()
	var i int
	for b.Loop() {
		s := fmt.Sprintf("file_%d.go", i%1000)
		id := index[s] // string → ID
		_ = slice[id]  // ID → string
		i++
	}
}

func BenchmarkRoundTripMapMap(b *testing.B) {
	strToID := make(map[string]uint32, 1000)
	idToStr := make(map[uint32]string, 1000)

	for i := range 1000 {
		s := fmt.Sprintf("file_%d.go", i)
		id := uint32(i)
		strToID[s] = id
		idToStr[id] = s
	}

	b.ResetTimer()
	var i int
	for b.Loop() {
		s := fmt.Sprintf("file_%d.go", i%1000)
		id := strToID[s] // string → ID
		_ = idToStr[id]  // ID → string
		i++
	}
}
