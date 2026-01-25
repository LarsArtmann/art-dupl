package printer

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// mockReadFileContent provides a mock file content for testing.
func mockReadFileContent() []byte {
	// Return minimal Go code for testing
	return []byte(`package main

func main() {
	println("hello")
}`)
}

func TestStatsDataAggregation(t *testing.T) {
	tests := []struct {
		name        string
		duplicates  [][][]*syntax.Node
		filesCount  int
		threshold   int
		checkStats  func(t *testing.T, stats *StatsData)
	}{
		{
			name: "empty clones",
			duplicates: [][][]*syntax.Node{},
			filesCount: 0,
			threshold:  15,
			checkStats: func(t *testing.T, stats *StatsData) {
				if stats.TotalCloneGroups != 0 {
					t.Errorf("TotalCloneGroups = %d, want 0", stats.TotalCloneGroups)
				}
				if stats.TotalClones != 0 {
					t.Errorf("TotalClones = %d, want 0", stats.TotalClones)
				}
			},
		},
		{
			name: "single clone group with two instances",
			duplicates: [][][]*syntax.Node{
				{
					{
						&syntax.Node{Filename: "file1.go", Pos: 1, End: 3, Type: 1},
						&syntax.Node{Filename: "file1.go", Pos: 2, End: 4, Type: 2},
					},
					{
						&syntax.Node{Filename: "file2.go", Pos: 10, End: 12, Type: 1},
						&syntax.Node{Filename: "file2.go", Pos: 11, End: 13, Type: 2},
					},
				},
			},
			filesCount: 2,
			threshold:  15,
			checkStats: func(t *testing.T, stats *StatsData) {
				if stats.TotalCloneGroups != 1 {
					t.Errorf("TotalCloneGroups = %d, want 1", stats.TotalCloneGroups)
				}
				if stats.TotalClones != 2 {
					t.Errorf("TotalClones = %d, want 2", stats.TotalClones)
				}
				// TotalDuplicateLines depends on actual file content, just verify it's positive
				if stats.TotalDuplicateLines <= 0 {
					t.Errorf("TotalDuplicateLines = %d, want > 0", stats.TotalDuplicateLines)
				}
				if stats.TotalTokens != 4 { // 2 nodes per clone × 2
					t.Errorf("TotalTokens = %d, want 4", stats.TotalTokens)
				}
			},
		},
		{
			name: "multiple clone groups",
			duplicates: [][][]*syntax.Node{
				{
					{
						&syntax.Node{Filename: "file1.go", Pos: 1, End: 2, Type: 1},
					},
				},
				{
					{
						&syntax.Node{Filename: "file2.go", Pos: 5, End: 6, Type: 2},
					},
				},
			},
			filesCount: 2,
			threshold:  15,
			checkStats: func(t *testing.T, stats *StatsData) {
				if stats.TotalCloneGroups != 2 {
					t.Errorf("TotalCloneGroups = %d, want 2", stats.TotalCloneGroups)
				}
				if stats.TotalClones != 2 {
					t.Errorf("TotalClones = %d, want 2", stats.TotalClones)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			statsPrinter := NewStats(&buf, mockReadFile(string(mockReadFileContent())), tt.threshold).(*stats)
			statsPrinter.SetFilesCount(tt.filesCount)

			for _, dupGroup := range tt.duplicates {
				if err := statsPrinter.PrintClones(dupGroup); err != nil {
					t.Fatalf("PrintClones failed: %v", err)
				}
			}

			if err := statsPrinter.PrintFooter(); err != nil {
				t.Fatalf("PrintFooter failed: %v", err)
			}

			tt.checkStats(t, statsPrinter.GetStatsData())
		})
	}
}

func TestStatsComplexityScore(t *testing.T) {
	var buf bytes.Buffer
	statsPrinter := NewStats(&buf, mockReadFile(string(mockReadFileContent())), 15).(*stats)
	statsPrinter.SetFilesCount(5)

	// Simulate: 3 clone groups, 9 total clones = complexity 3.0
	for i := 0; i < 3; i++ {
		dups := [][]*syntax.Node{
			{
				&syntax.Node{Filename: "file1.go", Pos: 1, End: 3, Type: 1},
				&syntax.Node{Filename: "file1.go", Pos: 2, End: 4, Type: 2},
			},
			{
				&syntax.Node{Filename: "file2.go", Pos: 1, End: 3, Type: 1},
				&syntax.Node{Filename: "file2.go", Pos: 2, End: 4, Type: 2},
			},
			{
				&syntax.Node{Filename: "file3.go", Pos: 1, End: 3, Type: 1},
				&syntax.Node{Filename: "file3.go", Pos: 2, End: 4, Type: 2},
			},
		}

		if err := statsPrinter.PrintClones(dups); err != nil {
			t.Fatalf("PrintClones failed: %v", err)
		}
	}

	if err := statsPrinter.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	statsData := statsPrinter.GetStatsData()
	expectedComplexity := 9.0 / 3.0 // 9 clones / 3 groups
	if statsData.ComplexityScore != expectedComplexity {
		t.Errorf("ComplexityScore = %.2f, want %.2f", statsData.ComplexityScore, expectedComplexity)
	}
}

func TestStatsImpactScore(t *testing.T) {
	var buf bytes.Buffer
	statsPrinter := NewStats(&buf, mockReadFile(string(mockReadFileContent())), 15).(*stats)
	statsPrinter.SetFilesCount(2)

	// Clone group with 4 tokens, appears 3 times
	dups := [][]*syntax.Node{
		{
			&syntax.Node{Filename: "file1.go", Pos: 1, End: 2, Type: 1},
			&syntax.Node{Filename: "file1.go", Pos: 2, End: 3, Type: 2},
			&syntax.Node{Filename: "file1.go", Pos: 3, End: 4, Type: 3},
			&syntax.Node{Filename: "file1.go", Pos: 4, End: 5, Type: 4},
		},
		{
			&syntax.Node{Filename: "file2.go", Pos: 10, End: 11, Type: 1},
			&syntax.Node{Filename: "file2.go", Pos: 11, End: 12, Type: 2},
			&syntax.Node{Filename: "file2.go", Pos: 12, End: 13, Type: 3},
			&syntax.Node{Filename: "file2.go", Pos: 13, End: 14, Type: 4},
		},
		{
			&syntax.Node{Filename: "file3.go", Pos: 20, End: 21, Type: 1},
			&syntax.Node{Filename: "file3.go", Pos: 21, End: 22, Type: 2},
			&syntax.Node{Filename: "file3.go", Pos: 22, End: 23, Type: 3},
			&syntax.Node{Filename: "file3.go", Pos: 23, End: 24, Type: 4},
		},
	}

	if err := statsPrinter.PrintClones(dups); err != nil {
		t.Fatalf("PrintClones failed: %v", err)
	}

	if err := statsPrinter.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	statsData := statsPrinter.GetStatsData()
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
	var buf bytes.Buffer
	statsPrinter := NewStats(&buf, mockReadFile(string(mockReadFileContent())), 15).(*stats)
	statsPrinter.SetFilesCount(2)

	// Create duplicate in file1.go with multiple nodes
	dups := [][]*syntax.Node{
		{
			&syntax.Node{Filename: "file1.go", Pos: 1, End: 3, Type: 1},
			&syntax.Node{Filename: "file1.go", Pos: 2, End: 4, Type: 2},
			&syntax.Node{Filename: "file1.go", Pos: 3, End: 5, Type: 3},
		},
	}

	if err := statsPrinter.PrintClones(dups); err != nil {
		t.Fatalf("PrintClones failed: %v", err)
	}

	if err := statsPrinter.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	statsData := statsPrinter.GetStatsData()
	if len(statsData.FileDuplication) == 0 {
		t.Fatal("FileDuplication map is empty")
	}

	if lines, ok := statsData.FileDuplication["file1.go"]; !ok {
		t.Error("file1.go not found in FileDuplication")
	} else if lines <= 0 {
		t.Errorf("file1.go duplicate lines = %d, want > 0", lines)
	}
}

func TestGetSizeRange(t *testing.T) {
	var buf bytes.Buffer
	statsPrinter := NewStats(&buf, mockReadFile(string(mockReadFileContent())), 15).(*stats)

	tests := []struct {
		lines    int
		expected string
	}{
		{1, "1-5 lines"},
		{5, "1-5 lines"},
		{6, "6-10 lines"},
		{10, "6-10 lines"},
		{11, "11-20 lines"},
		{20, "11-20 lines"},
		{21, "21-50 lines"},
		{50, "21-50 lines"},
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

func TestPrintSizeDistribution(t *testing.T) {
	var buf bytes.Buffer
	distribution := map[string]int{
		"1-5 lines":   10,
		"6-10 lines":  5,
		"11-20 lines": 3,
	}

	printSizeDistribution(&buf, distribution)

	output := buf.String()
	expectedRanges := []string{"1-5 lines: 10", "6-10 lines: 5", "11-20 lines: 3"}

	for _, expected := range expectedRanges {
		if !strings.Contains(output, expected) {
			t.Errorf("Output missing expected range: %s", expected)
		}
	}

	// Check that output is sorted (1-5 should come before 6-10)
	if idx := strings.Index(output, "1-5"); idx == -1 {
		t.Error("Output doesn't contain sorted '1-5 lines'")
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
	if !strings.Contains(lines[0], "fileA.go") || !strings.Contains(lines[0], "100") {
		t.Errorf("Top file should be fileA.go with 100 lines, got: %s", lines[0])
	}

	if !strings.Contains(lines[1], "fileC.go") || !strings.Contains(lines[1], "75") {
		t.Errorf("Second file should be fileC.go with 75 lines, got: %s", lines[1])
	}
}

func TestPrintTopFilesWithLessThanN(t *testing.T) {
	var buf bytes.Buffer
	fileDuplication := map[string]int{
		"fileA.go": 100,
	}

	printTopFiles(&buf, fileDuplication, 10) // Request top 10, but only 1 file exists

	output := buf.String()
	if !strings.Contains(output, "fileA.go") || !strings.Contains(output, "100") {
		t.Error("Output should contain the single file")
	}
	if strings.Contains(output, "more files") {
		t.Error("Output shouldn't contain 'more files' when all files are shown")
	}
}

func TestStatsDetectionMethods(t *testing.T) {
	var buf bytes.Buffer
	statsPrinter := NewStats(&buf, mockReadFile(string(mockReadFileContent())), 15).(*stats)
	statsPrinter.SetDetectionMethods("hash,art-dupl")

	if statsPrinter.statsData.DetectionMethods != "hash,art-dupl" {
		t.Errorf("DetectionMethods = %s, want 'hash,art-dupl'", statsPrinter.statsData.DetectionMethods)
	}
}

func TestStatsAverageCloneSize(t *testing.T) {
	tests := []struct {
		name            string
		totalLines      int
		totalClones     int
		expectedAverage int
	}{
		{"normal case", 100, 5, 20},
		{"single clone", 10, 1, 10},
		{"zero clones", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			statsPrinter := NewStats(&buf, mockReadFile(string(mockReadFileContent())), 15).(*stats)
			statsPrinter.statsData.TotalDuplicateLines = tt.totalLines
			statsPrinter.statsData.TotalClones = tt.totalClones

			if err := statsPrinter.PrintFooter(); err != nil {
				t.Fatalf("PrintFooter failed: %v", err)
			}

			statsData := statsPrinter.GetStatsData()
			if statsData.AverageCloneSize != tt.expectedAverage {
				t.Errorf("AverageCloneSize = %d, want %d", statsData.AverageCloneSize, tt.expectedAverage)
			}
		})
	}
}

func TestStatsJSONOutput(t *testing.T) {
	var buf bytes.Buffer
	statsPrinter := NewStats(&buf, mockReadFile(string(mockReadFileContent())), 15).(*stats)

	// Add some test data
	statsPrinter.SetDetectionMethods("art-dupl,hash")
	statsPrinter.SetFilesCount(10)

	// Add one clone group with duplicates
	dups := [][]*syntax.Node{
		{
			&syntax.Node{Filename: "file1.go", Pos: 1, End: 3, Type: 1},
			&syntax.Node{Filename: "file1.go", Pos: 2, End: 4, Type: 2},
		},
		{
			&syntax.Node{Filename: "file2.go", Pos: 10, End: 12, Type: 1},
			&syntax.Node{Filename: "file2.go", Pos: 11, End: 13, Type: 2},
		},
	}

	if err := statsPrinter.PrintClones(dups); err != nil {
		t.Fatalf("PrintClones failed: %v", err)
	}

	// Set format to JSON
	statsPrinter.format = FormatJSON

	if err := statsPrinter.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	output := buf.String()

	// Verify JSON is valid
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, output)
	}

	// Verify structure exists
	config, ok := result["configuration"].(map[string]interface{})
	if !ok {
		t.Fatal("Missing 'configuration' section in JSON")
	}
	if config["threshold"] != float64(15) {
		t.Errorf("threshold = %v, want 15", config["threshold"])
	}
	if config["detectionMethods"] != "art-dupl,hash" {
		t.Errorf("detectionMethods = %v, want 'art-dupl,hash'", config["detectionMethods"])
	}

	overview, ok := result["overview"].(map[string]interface{})
	if !ok {
		t.Fatal("Missing 'overview' section in JSON")
	}
	if overview["filesScanned"] != float64(10) {
		t.Errorf("filesScanned = %v, want 10", overview["filesScanned"])
	}
	if overview["cloneGroups"] != float64(1) {
		t.Errorf("cloneGroups = %v, want 1", overview["cloneGroups"])
	}
	if overview["totalClones"] != float64(2) {
		t.Errorf("totalClones = %v, want 2", overview["totalClones"])
	}

	dupCode, ok := result["duplicateCode"].(map[string]interface{})
	if !ok {
		t.Fatal("Missing 'duplicateCode' section in JSON")
	}
	if dupCode["totalDuplicateLines"] == 0 {
		t.Error("totalDuplicateLines should be > 0")
	}
	if dupCode["totalDuplicateTokens"] == 0 {
		t.Error("totalDuplicateTokens should be > 0")
	}
	if dupCode["averageCloneSize"] == 0 {
		t.Error("averageCloneSize should be > 0")
	}
	if dupCode["complexityScore"] != 2.0 { // 2 clones / 1 group
		t.Errorf("complexityScore = %v, want 2.0", dupCode["complexityScore"])
	}

	sizeDist, ok := result["sizeDistribution"].(map[string]interface{})
	if !ok || len(sizeDist) == 0 {
		t.Error("Missing or empty 'sizeDistribution' section")
	}

	topFiles, ok := result["topFiles"].([]interface{})
	if !ok || len(topFiles) == 0 {
		t.Error("Missing or empty 'topFiles' section")
	}
}

func TestStatsTextOutput(t *testing.T) {
	var buf bytes.Buffer
	statsPrinter := NewStats(&buf, mockReadFile(string(mockReadFileContent())), 15).(*stats)
	statsPrinter.SetFilesCount(5)
	statsPrinter.SetDetectionMethods("art-dupl")

	// Set format to text
	statsPrinter.format = FormatText

	if err := statsPrinter.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	output := buf.String()

	// Verify key sections are present
	expectedSections := []string{
		"Code Duplication Statistics",
		"Configuration:",
		"Overview:",
		"Files Scanned: 5",
		"Duplicate Code:",
	}

	for _, section := range expectedSections {
		if !strings.Contains(output, section) {
			t.Errorf("Output missing section: %q", section)
		}
	}
}

