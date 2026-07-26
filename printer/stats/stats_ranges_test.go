package stats

import (
	"bytes"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax"
)

func TestGetSizeRange(t *testing.T) {
	statsPrinter, _ := newTestStatsPrinter()

	tests := []struct {
		lines    int
		expected string
	}{
		{1, sizeRange1to5},
		{5, sizeRange1to5},
		{6, sizeRange6to10},
		{10, sizeRange6to10},
		{11, sizeRange11to20},
		{20, sizeRange11to20},
		{21, sizeRange21to50},
		{50, sizeRange21to50},
		{51, "51-100 lines"},
		{100, "51-100 lines"},
		{101, "100+ lines"},
		{1000, "100+ lines"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := statsPrinter.getSizeRange(tt.lines)
			if result != tt.expected {
				t.Errorf("getSizeRange(%d) = %s, want %s", tt.lines, result, tt.expected)
			}
		})
	}
}

func TestGetTokenRange(t *testing.T) {
	// With threshold=15, ranges are:
	// 1-15 (t), 16-30 (t*2), 31-45 (t*3), 46-75 (t*5), 76-150 (t*10), 151+ (>t*10)
	statsPrinter, _ := newTestStatsPrinter()

	tests := []struct {
		tokens   int
		expected string
	}{
		{1, "1-15 tokens"},
		{15, "1-15 tokens"},
		{16, "16-30 tokens"},
		{30, "16-30 tokens"},
		{31, "31-45 tokens"},
		{45, "31-45 tokens"},
		{46, "46-75 tokens"},
		{75, "46-75 tokens"},
		{76, "76-150 tokens"},
		{150, "76-150 tokens"},
		{151, "151+ tokens"},
		{1000, "151+ tokens"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := statsPrinter.getTokenRange(tt.tokens)
			if result != tt.expected {
				t.Errorf("getTokenRange(%d) = %s, want %s", tt.tokens, result, tt.expected)
			}
		})
	}
}

func TestPrintSizeDistribution(t *testing.T) {
	var buf bytes.Buffer

	distribution := map[string]int{
		sizeRange1to5:   10,
		sizeRange6to10:  5,
		sizeRange11to20: 3,
		sizeRange21to50: 1,
	}

	printSizeDistribution(&buf, distribution)

	output := buf.String()
	// Expected format: "  1-5 lines      :   10 clones [████████████████████] 55.6%"
	expectedRanges := []string{
		sizeRange1to5,
		"10 clones",
		sizeRange6to10,
		"5 clones",
		sizeRange11to20,
		"3 clones",
		sizeRange21to50,
		"1 clone",
	}

	for _, expected := range expectedRanges {
		testutil.AssertStringContains(
			t,
			output,
			expected,
			"Output missing expected range: "+expected,
		)
	}

	// Check that output contains bars
	testutil.AssertStringContains(t, output, "█", "Output doesn't contain ASCII bars")

	// Verify numeric sort order: 1-5 < 6-10 < 11-20 < 21-50 (not lexicographic)
	idx1 := strings.Index(output, sizeRange1to5)
	idx2 := strings.Index(output, sizeRange6to10)
	idx3 := strings.Index(output, sizeRange11to20)
	idx4 := strings.Index(output, sizeRange21to50)

	if idx1 == -1 || idx2 == -1 || idx3 == -1 || idx4 == -1 {
		t.Fatal("One or more ranges not found in output")
	}

	if idx1 >= idx2 || idx2 >= idx3 || idx3 >= idx4 {
		t.Errorf(
			"Output is not properly sorted by range start. Positions: 1-5=%d, 6-10=%d, 11-20=%d, 21-50=%d",
			idx1,
			idx2,
			idx3,
			idx4,
		)
	}
}

func TestPrintTopFiles(t *testing.T) {
	var buf bytes.Buffer

	fileDuplication := map[string]int{
		"fileA.go": 100,
		"fileB.go": 50,
		"fileC.go": 75,
		"fileD.go": 25,
	}

	printTopFiles(&buf, fileDuplication, 2)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) < 2 {
		t.Fatalf("Expected at least 2 lines of output, got %d", len(lines))
	}

	// Check that top 2 files are in correct order (fileA.go: 100, fileC.go: 75)
	assertFileContains(t, lines[0], "fileA.go", "100")

	testutil.AssertStringContains(
		t,
		lines[1],
		"fileC.go",
		"Second file should be fileC.go with 75 lines",
	)
	testutil.AssertStringContains(t, lines[1], "75", "Second file should contain 75")
}

func TestPrintTopFilesWithLessThanN(t *testing.T) {
	var buf bytes.Buffer

	fileDuplication := map[string]int{
		"fileA.go": 100,
	}

	printTopFiles(&buf, fileDuplication, 10) // Request top 10, but only 1 file exists

	output := buf.String()
	assertFileContains(t, output, "fileA.go", "100")

	if strings.Contains(output, "more files") {
		t.Error("Output shouldn't contain 'more files' when all files are shown")
	}
}

func TestStatsDetectionMethods(t *testing.T) {
	statsPrinter, _ := newTestStatsPrinter()
	statsPrinter.SetDetectionMethods("hash,art-dupl")

	if statsPrinter.statsData.DetectionMethods != "hash,art-dupl" {
		t.Errorf(
			"DetectionMethods = %s, want 'hash,art-dupl'",
			statsPrinter.statsData.DetectionMethods,
		)
	}
}

func TestStatsAverageCloneSize(t *testing.T) {
	tests := []struct {
		name             string
		totalLines       int
		totalClones      int
		totalCloneGroups int
		expectedAverage  int
	}{
		{"normal case", 100, 5, 5, 20},
		{"single clone", 10, 1, 1, 10},
		{"zero clones", 0, 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statsPrinter, _ := newTestStatsPrinter()
			statsPrinter.statsData.TotalDuplicateLines = tt.totalLines
			statsPrinter.statsData.TotalClones = tt.totalClones
			statsPrinter.statsData.TotalCloneGroups = tt.totalCloneGroups

			err := statsPrinter.PrintFooter()
			if err != nil {
				t.Fatalf("PrintFooter failed: %v", err)
			}

			statsData := statsPrinter.GetStatsView()
			if statsData.AverageCloneSize != tt.expectedAverage {
				t.Errorf(
					"AverageCloneSize = %d, want %d",
					statsData.AverageCloneSize,
					tt.expectedAverage,
				)
			}
		})
	}
}

// createCloneNodeGroup creates a group of clone nodes with specified files.
func createCloneNodeGroup(filenames []string) [][]*syntax.Node {
	dups := make([][]*syntax.Node, 0, len(filenames))

	for _, filename := range filenames {
		nodes := []*syntax.Node{
			testutil.CreateNodeWithPos(1, filename, 1, 3),
			testutil.CreateNodeWithPos(2, filename, 2, 4),
		}
		dups = append(dups, nodes)
	}

	return dups
}

func TestStatsComplexityScore(t *testing.T) {
	statsPrinter, _ := newTestStatsPrinter()
	statsPrinter.SetFilesCount(5)

	// Simulate: 3 clone groups, 9 total clones = complexity 3.0
	for range 3 {
		dups := createCloneNodeGroup([]string{"file1.go", "file2.go", "file3.go"})

		err := printTestClones(statsPrinter, mockReadFile(string(mockReadFileContent())), dups)
		if err != nil {
			t.Fatalf("PrintClones failed: %v", err)
		}
	}

	err := statsPrinter.PrintFooter()
	if err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	statsData := statsPrinter.GetStatsView()

	expectedComplexity := 9.0 / 3.0 // 9 clones / 3 groups
	if statsData.ComplexityScore != expectedComplexity {
		t.Errorf("ComplexityScore = %.2f, want %.2f", statsData.ComplexityScore, expectedComplexity)
	}
}

func TestStatsImpactScore(t *testing.T) {
	statsPrinter, _ := newTestStatsPrinter()
	statsPrinter.SetFilesCount(2)

	// Clone group with 4 tokens, appears 3 times
	dups := [][]*syntax.Node{
		createNodeSlice("file1.go", 1, 4),
		createNodeSlice("file2.go", 10, 13),
		createNodeSlice("file3.go", 20, 23),
	}

	err := printTestClones(statsPrinter, mockReadFile(string(mockReadFileContent())), dups)
	if err != nil {
		t.Fatalf("PrintClones failed: %v", err)
	}

	statsData := printFooterAndGetData(t, statsPrinter)
	// Impact score formula: sum(node_count_of_each_clone) × number_of_clones
	// Each clone has 4 nodes, so tokensInGroup = 4 + 4 + 4 = 12
	// len(dups) = 3 (number of clones)
	// ImpactScore = 12 × 3 = 36
	expectedImpact := 36
	if statsData.ImpactScore != expectedImpact {
		t.Errorf("ImpactScore = %d, want %d", statsData.ImpactScore, expectedImpact)
	}
}

func TestStatsFileDuplicationTracking(t *testing.T) {
	statsPrinter, _ := newTestStatsPrinter()
	statsPrinter.SetFilesCount(2)

	// Create duplicate in file1.go with multiple nodes
	dups := [][]*syntax.Node{
		{
			testutil.CreateNodeWithPos(1, "file1.go", 1, 3),
			testutil.CreateNodeWithPos(2, "file1.go", 2, 4),
			testutil.CreateNodeWithPos(3, "file1.go", 3, 5),
		},
	}

	err := printTestClones(statsPrinter, mockReadFile(string(mockReadFileContent())), dups)
	if err != nil {
		t.Fatalf("PrintClones failed: %v", err)
	}

	err = statsPrinter.PrintFooter()
	if err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	statsData := statsPrinter.GetStatsView()
	if len(statsData.FileDuplication) == 0 {
		t.Fatal("FileDuplication map is empty")
	}

	if lines, ok := statsData.FileDuplication["file1.go"]; !ok {
		t.Error("file1.go not found in FileDuplication")
	} else if lines <= 0 {
		t.Errorf("file1.go duplicate lines = %d, want > 0", lines)
	}
}
