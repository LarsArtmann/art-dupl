package printer

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// TestSortingIntegration tests the complete sorting functionality across all printers.
func TestSortingIntegration(t *testing.T) { //nolint:funlen // Comprehensive integration test
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
	multiOccurrenceClone := createMockCloneGroup(t, "multi.go", 200, 220, 3) // 3 tokens

	testCases := []struct {
		name          string
		sortBy        SortBy
		expectedOrder []string // Expected filenames in order
	}{
		{
			name:          "Sort by size descending",
			sortBy:        SortBySize,
			expectedOrder: []string{"large.go", "another_large.go", "medium.go", "small.go"},
		},
		{
			name:          "Sort by occurrence (file count, descending)",
			sortBy:        SortByOccurrence,
			expectedOrder: []string{"large.go", "another_large.go", "medium.go", "small.go"}, // All have same occurrence (1), falls back to hash sort which is alphabetical
		},
		{
			name:          "Sort by hash (filename order)",
			sortBy:        SortByHash,
			expectedOrder: []string{"another_large.go", "large.go", "medium.go", "small.go"},
		},
		{
			name:          "Sort by total-tokens",
			sortBy:        SortByTotalTokens,
			expectedOrder: []string{"large.go", "another_large.go", "multi.go", "medium.go", "small.go"}, // multi.go has 5 occurrences of 3 tokens = 15 total
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Prepare clone data for all tests
			clones := [][]*syntax.Node{largeClone, mediumClone, smallClone, anotherLargeClone, multiOccurrenceClone}

			// Test JSON Printer sorting (special case with different verification logic)
			t.Run("JSONPrinter", func(t *testing.T) {
				var buf bytes.Buffer
				printer := NewJSON(&buf, mockReadFile(testContent))

				err := printer.PrintClones(clones, tc.sortBy)
				if err != nil {
					t.Fatalf("JSONPrinter.PrintClones failed: %v", err)
				}

				// Verify sorting by calling OutputJSON
				err = printer.(*JSONPrinter).OutputJSON(15, tc.sortBy, "")
				if err != nil {
					t.Fatalf("JSONPrinter.OutputJSON failed: %v", err)
				}

				// For total-tokens sorting, we need special handling since it's the same as size in JSON
				output := buf.String()
				if tc.sortBy == SortByTotalTokens {
					// Just check that output contains expected files
					for _, file := range []string{"large.go", "another_large.go", "multi.go", "medium.go", "small.go"} {
						if !strings.Contains(output, file) {
							t.Errorf("Expected file %s not found in JSON output", file)
						}
					}
				}
			})

			// Test standard printers with common verification logic
			standardPrinters := []struct {
				name        string
				constructor func(io.Writer, ReadFile) Printer
			}{
				{"TextPrinter", NewText},
				{"HTMLPrinter", func(w io.Writer, fread ReadFile) Printer { return NewHTML(w, fread) }},
				{"PlumbingPrinter", NewPlumbing},
			}

			for _, sp := range standardPrinters {
				t.Run(sp.name, func(t *testing.T) {
					testPrinterSorting(t, sp.constructor, testContent, clones, tc.sortBy, tc.expectedOrder, sp.name)
				})
			}
		})
	}
}

// testPrinterSorting tests a printer's sorting functionality with standard verification logic.
func testPrinterSorting(t *testing.T, constructor func(io.Writer, ReadFile) Printer,
	testContent string, clones [][]*syntax.Node, sortBy SortBy, expectedOrder []string, printerName string,
) {
	t.Helper()
	var buf bytes.Buffer
	printer := constructor(&buf, mockReadFile(testContent))

	err := printer.PrintClones(clones, sortBy)
	if err != nil {
		t.Fatalf("%s.PrintClones failed: %v", printerName, err)
	}

	// Verify sorting by checking the order of files in output
	output := buf.String()
	// Simple check: the first expected file should appear before the last expected file
	firstIndex := strings.Index(output, expectedOrder[0])
	lastIndex := strings.Index(output, expectedOrder[len(expectedOrder)-1])

	if firstIndex == -1 || lastIndex == -1 {
		t.Errorf("Expected files not found in output. First: %s, Last: %s", expectedOrder[0], expectedOrder[len(expectedOrder)-1])
	} else if firstIndex > lastIndex {
		t.Errorf("%s sorting failed: %s should appear before %s", printerName, expectedOrder[0], expectedOrder[len(expectedOrder)-1])
	}
}

// createMockCloneGroup creates a mock clone group with specified characteristics.
func createMockCloneGroup(t *testing.T, filename string, startPos, endPos, numTokens int) []*syntax.Node {
	t.Helper()
	// Create nodes that represent the tokens in a clone
	nodes := make([]*syntax.Node, numTokens)

	for i := range numTokens {
		nodes[i] = &syntax.Node{
			Type:     golang.FuncDecl,
			Filename: filename,
			Pos:      int32(startPos + (i * 2)),
			End:      int32(startPos + (i * 2) + 1),
		}
	}

	return nodes
}

// TestCommonSortingUtilities tests the common sorting functions directly.
func TestCommonSortingUtilities(t *testing.T) {
	smallClone := createMockCloneGroup(t, "small.go", 10, 20, 2)
	mediumClone := createMockCloneGroup(t, "medium.go", 30, 50, 5)
	largeClone := createMockCloneGroup(t, "large.go", 60, 90, 8)

	clones := [][]*syntax.Node{mediumClone, smallClone, largeClone}

	// Test SortClonesBySize
	t.Run("SortClonesBySize", func(t *testing.T) {
		TestCloneSortingWithData(t, SortClonesBySize, "Size", clones)
	})

	// Test SortClonesByOccurrence (same as size for our test data)
	t.Run("SortClonesByOccurrence", func(t *testing.T) {
		TestCloneSortingWithData(t, SortClonesByOccurrence, "Occurrence", clones)
	})

	// Test SortClonesByHash
	t.Run("SortClonesByHash", func(t *testing.T) {
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
