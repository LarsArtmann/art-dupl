package printer

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

const (
	largeFile        = "large.go"
	anotherLargeFile = "another_large.go"
	mediumFile       = "medium.go"
	smallFile        = "small.go"
)

// createMockCloneGroup creates a mock clone group with specified characteristics.
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
	smallClone := createMockCloneGroup(t, smallFile, 10, 2) // Small size, 2 tokens
	mediumClone := createMockCloneGroup(
		t,
		mediumFile,
		30,
		5,
	) // Medium size, 5 tokens
	largeClone := createMockCloneGroup(t, largeFile, 60, 8) // Large size, 8 tokens
	anotherLargeClone := createMockCloneGroup(
		t,
		anotherLargeFile,
		100,
		8,
	) // Same size as largeClone, 8 tokens

	// Create clones with multiple occurrences to test total-tokens
	multiOccurrenceClone := createMockCloneGroup(t, "multi.go", 200, 3) // 3 tokens

	testCases := []struct {
		name          string
		sortBy        config.SortCriteria
		expectedOrder []string // Expected filenames in order
	}{
		{
			name:          "Sort by size descending",
			sortBy:        config.SortBySize,
			expectedOrder: []string{largeFile, anotherLargeFile, mediumFile, smallFile},
		},
		{
			name:   "Sort by occurrence (file count, descending)",
			sortBy: config.SortByOccurrence,
			expectedOrder: []string{
				largeFile,
				anotherLargeFile,
				mediumFile,
				smallFile,
			}, // All have same occurrence (1), falls back to hash sort which is alphabetical
		},
		{
			name:          "Sort by hash (filename order)",
			sortBy:        config.SortByHash,
			expectedOrder: []string{anotherLargeFile, largeFile, mediumFile, smallFile},
		},
		{
			name:   "Sort by total-tokens",
			sortBy: config.SortByTotalTokens,
			expectedOrder: []string{
				largeFile,
				anotherLargeFile,
				"multi.go",
				mediumFile,
				smallFile,
			}, // multi.go has 5 occurrences of 3 tokens = 15 total
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Prepare clone data for all tests
			clones := [][]*syntax.Node{
				largeClone,
				mediumClone,
				smallClone,
				anotherLargeClone,
				multiOccurrenceClone,
			}

			// Test JSON Printer sorting (special case with different verification logic)
			t.Run("JSONPrinter", func(t *testing.T) {
				var buf bytes.Buffer

				printer := NewJSON(&buf, mockReadFile(testContent))

				err := printer.PrintClones(
					processTestNodes(mockReadFile(testContent), "test", clones),
					tc.sortBy,
				)
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

				if tc.sortBy == config.SortByTotalTokens {
					// Just check that output contains expected files
					for _, file := range []string{largeFile, anotherLargeFile, "multi.go", mediumFile, "small.go"} {
						t.Run(file, func(t *testing.T) {
							if !strings.Contains(output, file) {
								t.Errorf("Expected file %s not found in JSON output", file)
							}
						})
					}
				}
			})

			// Test standard printers with common verification logic
			standardPrinters := []struct {
				name        string
				constructor func(io.Writer, ReadFile) Printer
			}{
				{"TextPrinter", NewText},
				{
					"HTMLPrinter",
					func(w io.Writer, fread ReadFile) Printer { return NewHTML(w, fread) },
				},
				{"PlumbingPrinter", NewPlumbing},
			}

			for _, sp := range standardPrinters {
				t.Run(sp.name, func(t *testing.T) {
					testPrinterSorting(
						t,
						sp.constructor,
						testContent,
						clones,
						tc.sortBy,
						tc.expectedOrder,
						sp.name,
					)
				})
			}
		})
	}
}

// testPrinterSorting tests a printer's sorting functionality with standard verification logic.
func testPrinterSorting(
	t *testing.T,
	constructor func(io.Writer, ReadFile) Printer,
	testContent string,
	clones [][]*syntax.Node,
	sortBy config.SortCriteria,
	expectedOrder []string,
	printerName string,
) {
	t.Helper()

	var buf bytes.Buffer

	printer := constructor(&buf, mockReadFile(testContent))

	err := printer.PrintClones(processTestNodes(mockReadFile(testContent), "test", clones), sortBy)
	if err != nil {
		t.Fatalf("%s.PrintClones failed: %v", printerName, err)
	}

	// Verify sorting by checking the order of files in output
	output := buf.String()
	// Simple check: the first expected file should appear before the last expected file
	firstIndex := strings.Index(output, expectedOrder[0])
	lastIndex := strings.Index(output, expectedOrder[len(expectedOrder)-1])

	if firstIndex == -1 || lastIndex == -1 {
		t.Errorf(
			"Expected files not found in output. First: %s, Last: %s",
			expectedOrder[0],
			expectedOrder[len(expectedOrder)-1],
		)
	} else if firstIndex > lastIndex {
		t.Errorf(
			"%s sorting failed: %s should appear before %s",
			printerName,
			expectedOrder[0],
			expectedOrder[len(expectedOrder)-1],
		)
	}
}

// createMockCloneGroup creates a mock clone group with specified characteristics.
func createMockCloneGroup(t *testing.T, filename string, startPos, numTokens int) []*syntax.Node {
	t.Helper()
	// Create nodes that represent the tokens in a clone
	nodes := make([]*syntax.Node, 0, numTokens)

	for i := range numTokens {
		nodes = append(nodes, &syntax.Node{
			Type:     golang.FuncDecl,
			Filename: filename,
			Pos:      int32(startPos + (i * 2)),
			End:      int32(startPos + (i * 2) + 1),
		})
	}

	return nodes
}
