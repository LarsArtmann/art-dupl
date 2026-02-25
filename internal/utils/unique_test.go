package utils

import (
	"testing"
)

func TestUniqueStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name:     "no duplicates",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with duplicates",
			input:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "all same",
			input:    []string{"x", "x", "x"},
			expected: []string{"x"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UniqueStringSlice(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("UniqueStringSlice() returned %d items, expected %d", len(result), len(tt.expected))
			}
			for i, v := range tt.expected {
				if i >= len(result) || result[i] != v {
					t.Errorf("UniqueStringSlice()[%d] = %v, expected %v", i, result[i], v)
				}
			}
		})
	}
}
