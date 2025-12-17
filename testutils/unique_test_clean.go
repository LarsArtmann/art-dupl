package testutils

import (
	"strings"
	"sync"
	"testing"
)

// TestUniqueTestHelper_Clean tests unique helper functionality
func TestUniqueTestHelper_Clean(t *testing.T) {
	// Generate multiple unique strings
	for i := 0; i < 10; i++ {
		str := UniqueTestHelper()

		// Check that it starts with "unique_"
		if !strings.HasPrefix(str, "unique_") {
			t.Errorf("String should start with 'unique_', got '%s'", str)
		}
	}
}

// TestGenerateRandomSuffix_Clean tests random suffix generation
func TestGenerateRandomSuffix_Clean(t *testing.T) {
	// Generate multiple suffixes
	for i := 0; i < 100; i++ {
		suffix := generateRandomSuffix()

		// Should be single letter
		if len(suffix) != 1 {
			t.Errorf("Suffix should be single character, got '%s' (len %d)", suffix, len(suffix))
		}

		// Should be lowercase letter a-z
		char := suffix[0]
		if char < 'a' || char > 'z' {
			t.Errorf("Suffix should be lowercase letter a-z, got '%c'", char)
		}
	}
}

// TestUniqueFunction_Clean tests unique function generation
func TestUniqueFunction_Clean(t *testing.T) {
	// Generate multiple unique functions
	for i := 0; i < 5; i++ {
		fn := UniqueFunction()

		// Check that it has proper structure
		if !strings.HasPrefix(fn, "func unique") {
			t.Errorf("Function should start with 'func unique', got '%.20s...'", fn)
		}
	}
}

// TestUniqueness_Clean tests basic uniqueness
func TestUniqueness_Clean(t *testing.T) {
	// Generate strings and check they're unique
	uniqueSet := make(map[string]bool)

	for i := 0; i < 50; i++ {
		str := UniqueTestHelper()

		if uniqueSet[str] {
			t.Errorf("Duplicate found: %s", str)
		}
		uniqueSet[str] = true
	}

	// Should have all unique strings
	if len(uniqueSet) != 50 {
		t.Errorf("Expected 50 unique strings, got %d", len(uniqueSet))
	}
}

// TestUniqueness_Concurrent_Clean tests concurrency
func TestUniqueness_Concurrent_Clean(t *testing.T) {
	numGoroutines := 3
	numPerGoroutine := 5

	var wg sync.WaitGroup
	var mu sync.Mutex
	allStrings := make([]string, 0, numGoroutines*numPerGoroutine)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()

			localStrings := make([]string, numPerGoroutine)
			for j := range localStrings {
				localStrings[j] = UniqueTestHelper()
			}

			mu.Lock()
			allStrings = append(allStrings, localStrings...)
			mu.Unlock()
		}()
	}

	wg.Wait()

	// Check that all strings are unique
	uniqueSet := make(map[string]bool)
	for _, s := range allStrings {
		if uniqueSet[s] {
			t.Errorf("Duplicate found: %s", s)
		}
		uniqueSet[s] = true
	}

	expected := numGoroutines * numPerGoroutine
	if len(uniqueSet) != expected {
		t.Errorf("Expected %d unique strings, got %d", expected, len(uniqueSet))
	}
}

// BenchmarkUniqueTestHelper_Clean benchmarks unique helper
func BenchmarkUniqueTestHelper_Clean(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UniqueTestHelper()
	}
}
