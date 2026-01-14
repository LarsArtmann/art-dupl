package utils

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

func TestUnique(t *testing.T) {
	tests := []struct {
		name     string
		input    [][]*syntax.Node
		expected int
	}{
		{
			name:     "empty slice",
			input:    [][]*syntax.Node{},
			expected: 0,
		},
		{
			name: "no duplicates",
			input: [][]*syntax.Node{
				{{Filename: "file1.go", Pos: 10}},
				{{Filename: "file1.go", Pos: 20}},
				{{Filename: "file2.go", Pos: 30}},
			},
			expected: 3,
		},
		{
			name: "with duplicates",
			input: [][]*syntax.Node{
				{{Filename: "file1.go", Pos: 10}},
				{{Filename: "file1.go", Pos: 10}},
				{{Filename: "file2.go", Pos: 20}},
			},
			expected: 2,
		},
		{
			name: "same file different positions",
			input: [][]*syntax.Node{
				{{Filename: "file1.go", Pos: 10}},
				{{Filename: "file1.go", Pos: 20}},
				{{Filename: "file1.go", Pos: 10}},
			},
			expected: 2,
		},
		{
			name: "different files same position",
			input: [][]*syntax.Node{
				{{Filename: "file1.go", Pos: 10}},
				{{Filename: "file2.go", Pos: 10}},
				{{Filename: "file1.go", Pos: 10}},
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Unique(tt.input)
			if len(result) != tt.expected {
				t.Errorf("Unique() returned %d items, expected %d", len(result), tt.expected)
			}
		})
	}
}
