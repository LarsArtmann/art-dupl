package printer

import (
	"testing"
)

// TestSortByParse tests ParseSortBy function with various inputs.
func TestSortByParse(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    SortBy
		shouldError bool
	}{
		{"valid size", "size", SortBySize, false},
		{"valid occurrence", "occurrence", SortByOccurrence, false},
		{"valid hash", "hash", SortByHash, false},
		{"valid total-tokens", "total-tokens", SortByTotalTokens, false},
		{"invalid option", "invalid", "", true},
		{"empty string", "", "", true},
		{"case sensitive SIZE", "SIZE", SortBySize, false}, // Should normalize to lowercase
		{"case sensitive OCCURRENCE", "OCCURRENCE", SortByOccurrence, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseSortBy(tt.input)
			if tt.shouldError {
				if err == nil {
					t.Errorf("Expected error for input %q, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error for input %q: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestSortByString tests String method for all SortBy values.
func TestSortByString(t *testing.T) {
	tests := []struct {
		sortBy   SortBy
		expected string
	}{
		{SortBySize, "size"},
		{SortByOccurrence, "occurrence"},
		{SortByHash, "hash"},
		{SortByTotalTokens, "total-tokens"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.sortBy.String() != tt.expected {
				t.Errorf("Expected String() to return %q, got %q", tt.expected, tt.sortBy.String())
			}
		})
	}
}

// TestSortByIsValid tests IsValid method for all SortBy values.
func TestSortByIsValid(t *testing.T) {
	tests := []struct {
		sortBy  SortBy
		isValid bool
	}{
		{SortBySize, true},
		{SortByOccurrence, true},
		{SortByHash, true},
		{SortByTotalTokens, true},
		{SortBy("invalid"), false},
		{SortBy(""), false},
		{SortBy("random"), false},
	}

	for _, tt := range tests {
		t.Run(tt.sortBy.String(), func(t *testing.T) {
			if tt.sortBy.IsValid() != tt.isValid {
				t.Errorf("Expected IsValid() to return %v for %q, got %v", tt.isValid, tt.sortBy, tt.sortBy.IsValid())
			}
		})
	}
}

// TestSortByConstants tests that all SortBy constants are valid.
func TestSortByConstants(t *testing.T) {
	// Verify all defined constants are valid
	constants := []SortBy{
		SortBySize,
		SortByOccurrence,
		SortByHash,
		SortByTotalTokens,
	}

	for _, sortBy := range constants {
		if !sortBy.IsValid() {
			t.Errorf("Constant %q is not valid", sortBy)
		}
	}
}
