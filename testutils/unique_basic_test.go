package testutils

import (
	"strings"
	"testing"
)

// TestUniqueTestHelper_Basic tests unique helper functionality
func TestUniqueTestHelper_Basic(t *testing.T) {
	str := UniqueTestHelper()

	// Check that it starts with "unique_"
	if !strings.HasPrefix(str, "unique_") {
		t.Errorf("String should start with 'unique_', got '%s'", str)
	}
}

// TestGenerateRandomSuffix_Basic tests random suffix generation
func TestGenerateRandomSuffix_Basic(t *testing.T) {
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

// TestUniqueFunction_Basic tests unique function generation
func TestUniqueFunction_Basic(t *testing.T) {
	fn := UniqueFunction()

	// Check that it has proper structure
	if !strings.HasPrefix(fn, "func unique") {
		t.Errorf("Function should start with 'func unique', got '%.20s...'", fn)
	}
}

// TestUniqueness_Basic tests basic uniqueness
func TestUniqueness_Basic(t *testing.T) {
	// Generate strings and check they're unique
	uniqueSet := make(map[string]bool)

	for i := 0; i < 10; i++ {
		str := UniqueTestHelper()

		if uniqueSet[str] {
			t.Errorf("Duplicate found: %s", str)
		}
		uniqueSet[str] = true
	}

	// Should have all unique strings
	if len(uniqueSet) != 10 {
		t.Errorf("Expected 10 unique strings, got %d", len(uniqueSet))
	}
}
