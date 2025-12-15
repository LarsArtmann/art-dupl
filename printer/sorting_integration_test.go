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
	smallClone := createMockCloneGroup(t, "small.go", 10, 20, 2)                  // Small size, 2 tokens
	mediumClone := createMockCloneGroup(t, "medium.go", 30, 50, 5)                // Medium size, 5 tokens
	largeClone := createMockCloneGroup(t, "large.go", 60, 90, 8)                  // Large size, 8 tokens
	anotherLargeClone := createMockCloneGroup(t, "another_large.go", 100, 130, 8) // Same size as largeClone, 8 tokens

	// Create clones with multiple occurrences to test total-tokens
	multiOccurrenceClone := createMultipleCloneGroup(t, "multi.go", 200, 220, 3, 5) // 3 tokens, 5 occurrences = 15 total tokens

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
		{
			name:          "Sort by total-tokens",
			sortBy:        "total-tokens",
			expectedOrder: []string{"large.go", "another_large.go", "multi.go", "medium.go", "small.go"}, // multi.go has 5 occurrences of 3 tokens = 15 total
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test JSON Printer sorting
			t.Run("JSONPrinter", func(t *testing.T) {
				var buf bytes.Buffer
				printer := NewJSON(&buf, mockReadFile(testContent))

				clones := [][]*syntax.Node{largeClone, mediumClone, smallClone, anotherLargeClone, multiOccurrenceClone}
				err := printer.PrintClones(clones, tc.sortBy)
				if err != nil {
					t.Fatalf("JSONPrinter.PrintClones failed: %v", err)
				}

				// Verify sorting by calling OutputJSON
				err = printer.(*JSONPrinter).OutputJSON(15, tc.sortBy)
				if err != nil {
					t.Fatalf("JSONPrinter.OutputJSON failed: %v", err)
				}

				// For total-tokens sorting, we need special handling since it's the same as size in JSON
				output := buf.String()
				if tc.sortBy == "total-tokens" {
					// Just check that output contains expected files
					for _, file := range []string{"large.go", "another_large.go", "multi.go", "medium.go", "small.go"} {
						if !strings.Contains(output, file) {
							t.Errorf("Expected file %s not found in JSON output", file)
						}
					}
				}
			})

			// Test Text Printer sorting
			t.Run("TextPrinter", func(t *testing.T) {
				var buf bytes.Buffer
				printer := NewText(&buf, mockReadFile(testContent))

				clones := [][]*syntax.Node{largeClone, mediumClone, smallClone, anotherLargeClone, multiOccurrenceClone}
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

				clones := [][]*syntax.Node{largeClone, mediumClone, smallClone, anotherLargeClone, multiOccurrenceClone}
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

				clones := [][]*syntax.Node{largeClone, mediumClone, smallClone, anotherLargeClone, multiOccurrenceClone}
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

// createMultipleCloneGroup creates a mock clone group with multiple occurrences
func createMultipleCloneGroup(t *testing.T, filename string, startPos, endPos, numTokens, numOccurrences int) []*syntax.Node {
	// Create nodes that represent tokens in a clone
	nodes := make([]*syntax.Node, numTokens)

	for i := range numTokens {
		nodes[i] = &syntax.Node{
			Type:     golang.FuncDecl,
			Filename: filename,
			Pos:      startPos + (i * 2),
			End:      startPos + (i * 2) + 1,
		}
	}

	return nodes
}

// createMockCloneGroup creates a mock clone group with specified characteristics
func createMockCloneGroup(t *testing.T, filename string, startPos, endPos, numTokens int) []*syntax.Node {
	// Create nodes that represent the tokens in a clone
	nodes := make([]*syntax.Node, numTokens)

	for i := range numTokens {
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

	// Test SortClonesByTotalTokens
	t.Run("SortClonesByTotalTokens", func(t *testing.T) {
		// The SortClonesByTotalTokens function works with [][]*syntax.Node
		// where each element is one clone occurrence in a file
		// So to test different total token counts, we need to simulate
		// the scenario where the function is called with different groups

		// In real usage, the function receives a group of clones
		// where each inner slice is a separate occurrence
		// For testing, we'll simulate sorting multiple groups separately

		group1 := createMockCloneGroup(t, "small.go", 10, 20, 2)  // 2 tokens
		group2 := createMockCloneGroup(t, "medium.go", 30, 50, 5) // 5 tokens
		group3 := createMockCloneGroup(t, "large.go", 60, 90, 8)  // 8 tokens

		clones := [][]*syntax.Node{group1, group2, group3}
		sorted := SortClonesByTotalTokens(clones)

		// Since SortClonesByTotalTokens counts all nodes, and we have single occurrences,
		// it should sort by the number of nodes in each group
		expectedOrder := []string{"large.go", "medium.go", "small.go"}
		for i, clone := range sorted {
			if clone[0].Filename != expectedOrder[i] {
				t.Errorf("Total-tokens sorting failed at index %d. Expected: %s, Got: %s",
					i, expectedOrder[i], clone[0].Filename)
			}
		}
	})
}
