package cmd

import (
	"testing"
)

func TestRecommendThreshold(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		fileCount int
		expected  int
	}{
		{"zero files", 0, smallCodebaseThreshold},
		{"small codebase (50)", 50, smallCodebaseThreshold},
		{"boundary: just under 100", 99, smallCodebaseThreshold},
		{"boundary: exactly 100", 100, mediumCodebaseThreshold},
		{"medium codebase (500)", 500, mediumCodebaseThreshold},
		{"boundary: just under 1000", 999, mediumCodebaseThreshold},
		{"boundary: exactly 1000", 1000, largeCodebaseThreshold},
		{"large codebase (2500)", 2500, largeCodebaseThreshold},
		{"boundary: just under 5000", 4999, largeCodebaseThreshold},
		{"boundary: exactly 5000", 5000, hugeCodebaseThreshold},
		{"huge codebase (10000)", 10000, hugeCodebaseThreshold},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := RecommendThreshold(tc.fileCount)
			if result != tc.expected {
				t.Errorf("RecommendThreshold(%d) = %d, want %d", tc.fileCount, result, tc.expected)
			}
		})
	}
}

func TestIsGoFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path     string
		expected bool
	}{
		{"main.go", true},
		{"foo/bar_test.go", true},
		{"file.txt", false},
		{"README.md", false},
		{"noext", false},
		{"", false},
	}

	for _, tc := range tests {
		t.Run(tc.path, func(t *testing.T) {
			t.Parallel()

			result := isGoFile(tc.path)
			if result != tc.expected {
				t.Errorf("isGoFile(%q) = %v, want %v", tc.path, result, tc.expected)
			}
		})
	}
}
