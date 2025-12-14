package printer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// TestSortingIntegration tests the complete sorting functionality across all printers
func TestSortingIntegration(t *testing.T) {
	// Create test file content with multiple clone groups
	testContent := `package main

import "fmt"

// Large function (should be first when sorting by size)
func largeFunction() {
	fmt.Println("line 1")
	fmt.Println("line 2")
	fmt.Println("line 3")
	fmt.Println("line 4")
	fmt.Println("line 5")
}

// Medium function (should be in middle)
func mediumFunction() {
	fmt.Println("line 1")
	fmt.Println("line 2")
	fmt.Println("line 3")
}

// Small function (should be last when sorting by size)
func smallFunction() {
	fmt.Println("line 1")
}

// Another large function (should be second when sorting by size)
func anotherLargeFunction() {
	fmt.Println("line 1")
	fmt.Println("line 2")
	fmt.Println("line 3")
	fmt.Println("line 4")
	fmt.Println("line 5")
}`

	// Create mock clone groups with different sizes and characteristics
	smallClone := createMockCloneGroup(t, "small.go", 10, 20, 2)                  // Small size
	mediumClone := createMockCloneGroup(t, "medium.go", 30, 50, 5)                // Medium size
	largeClone := createMockCloneGroup(t, "large.go", 60, 90, 8)                  // Large size
	anotherLargeClone := createMockCloneGroup(t, "another_large.go", 100, 130, 8) // Same size as largeClone

	testCases := []struct {
		name          string
		sortBy        string
		expectedOrder []string // Expected filenames in order
	}{
		{
			name:          "Sort by size descending",
			sortBy:        "size",
			expectedOrder: []string{"large.go", "another_large.go", "medium.go", "small.go"},
		},
		{
			name:          "Sort by occurrence (file count)",
			sortBy:        "occurrence",
			expectedOrder: []string{"large.go", "another_large.go", "medium.go", "small.go"}, // All have same occurrence, should fallback to size
		},
		{
			name:          "Sort by hash (filename order)",
			sortBy:        "hash",
			expectedOrder: []string{"another_large.go", "large.go", "medium.go", "small.go"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test Text Printer sorting
			t.Run("TextPrinter", func(t *testing.T) {
				var buf bytes.Buffer
				printer := NewText(&buf, mockReadFile(testContent))

				clones := [][]*syntax.Node{largeClone, mediumClone, smallClone, anotherLargeClone}
				err := printer.PrintClones(clones, tc.sortBy)
				if err != nil {
					t.Fatalf("TextPrinter.PrintClones failed: %v", err)
				}

				// Verify sorting by checking the order of files in output
				output := buf.String()
				// Simple check: the first expected file should appear before the last expected file
				firstIndex := strings.Index(output, tc.expectedOrder[0])
				lastIndex := strings.Index(output, tc.expectedOrder[len(tc.expectedOrder)-1])

				if firstIndex == -1 || lastIndex == -1 {
					t.Errorf("Expected files not found in output. First: %s, Last: %s", tc.expectedOrder[0], tc.expectedOrder[len(tc.expectedOrder)-1])
				} else if firstIndex > lastIndex {
					t.Errorf("Sorting failed: %s should appear before %s", tc.expectedOrder[0], tc.expectedOrder[len(tc.expectedOrder)-1])
				}
			})

			// Test HTML Printer sorting
			t.Run("HTMLPrinter", func(t *testing.T) {
				var buf bytes.Buffer
				printer := NewHTML(&buf, mockReadFile(testContent))

				clones := [][]*syntax.Node{largeClone, mediumClone, smallClone, anotherLargeClone}
				err := printer.PrintClones(clones, tc.sortBy)
				if err != nil {
					t.Fatalf("HTMLPrinter.PrintClones failed: %v", err)
				}

				// Verify sorting
				output := buf.String()
				firstIndex := strings.Index(output, tc.expectedOrder[0])
				lastIndex := strings.Index(output, tc.expectedOrder[len(tc.expectedOrder)-1])

				if firstIndex == -1 || lastIndex == -1 {
					t.Errorf("Expected files not found in output. First: %s, Last: %s", tc.expectedOrder[0], tc.expectedOrder[len(tc.expectedOrder)-1])
				} else if firstIndex > lastIndex {
					t.Errorf("HTML sorting failed: %s should appear before %s", tc.expectedOrder[0], tc.expectedOrder[len(tc.expectedOrder)-1])
				}
			})

			// Test Plumbing Printer sorting
			t.Run("PlumbingPrinter", func(t *testing.T) {
				var buf bytes.Buffer
				printer := NewPlumbing(&buf, mockReadFile(testContent))

				clones := [][]*syntax.Node{largeClone, mediumClone, smallClone, anotherLargeClone}
				err := printer.PrintClones(clones, tc.sortBy)
				if err != nil {
					t.Fatalf("PlumbingPrinter.PrintClones failed: %v", err)
				}

				// Verify sorting
				output := buf.String()
				firstIndex := strings.Index(output, tc.expectedOrder[0])
				lastIndex := strings.Index(output, tc.expectedOrder[len(tc.expectedOrder)-1])

				if firstIndex == -1 || lastIndex == -1 {
					t.Errorf("Expected files not found in output. First: %s, Last: %s", tc.expectedOrder[0], tc.expectedOrder[len(tc.expectedOrder)-1])
				} else if firstIndex > lastIndex {
					t.Errorf("Plumbing sorting failed: %s should appear before %s", tc.expectedOrder[0], tc.expectedOrder[len(tc.expectedOrder)-1])
				}
			})
		})
	}
}

// createMockCloneGroup creates a mock clone group with specified characteristics
func createMockCloneGroup(t *testing.T, filename string, startPos, endPos, numTokens int) []*syntax.Node {
	// Create nodes that represent the tokens in a clone
	nodes := make([]*syntax.Node, numTokens)

	for i := 0; i < numTokens; i++ {
		nodes[i] = &syntax.Node{
			Type:     golang.FuncDecl,
			Filename: filename,
			Pos:      startPos + (i * 2),
			End:      startPos + (i * 2) + 1,
		}
	}

	return nodes
}

// TestCommonSortingUtilities tests the common sorting functions directly
func TestCommonSortingUtilities(t *testing.T) {
	smallClone := createMockCloneGroup(t, "small.go", 10, 20, 2)
	mediumClone := createMockCloneGroup(t, "medium.go", 30, 50, 5)
	largeClone := createMockCloneGroup(t, "large.go", 60, 90, 8)

	// Test SortClonesBySize
	t.Run("SortClonesBySize", func(t *testing.T) {
		clones := [][]*syntax.Node{mediumClone, smallClone, largeClone}
		sorted := SortClonesBySize(clones)

		if len(sorted[0]) < len(sorted[1]) || len(sorted[1]) < len(sorted[2]) {
			t.Errorf("Size sorting failed. Expected: 8, 5, 2. Got: %d, %d, %d",
				len(sorted[0]), len(sorted[1]), len(sorted[2]))
		}
	})

	// Test SortClonesByOccurrence (same as size for our test data)
	t.Run("SortClonesByOccurrence", func(t *testing.T) {
		clones := [][]*syntax.Node{mediumClone, smallClone, largeClone}
		sorted := SortClonesByOccurrence(clones)

		if len(sorted[0]) < len(sorted[1]) || len(sorted[1]) < len(sorted[2]) {
			t.Errorf("Occurrence sorting failed. Expected: 8, 5, 2. Got: %d, %d, %d",
				len(sorted[0]), len(sorted[1]), len(sorted[2]))
		}
	})

	// Test SortClonesByHash
	t.Run("SortClonesByHash", func(t *testing.T) {
		clones := [][]*syntax.Node{mediumClone, smallClone, largeClone}
		sorted := SortClonesByHash(clones)

		// Should be sorted alphabetically by filename
		expectedOrder := []string{"large.go", "medium.go", "small.go"}
		for i, clone := range sorted {
			if clone[0].Filename != expectedOrder[i] {
				t.Errorf("Hash sorting failed at index %d. Expected: %s, Got: %s",
					i, expectedOrder[i], clone[0].Filename)
			}
		}
	})
}
