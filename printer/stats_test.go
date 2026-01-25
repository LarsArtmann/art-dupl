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

// createNodeSlice creates a slice of syntax.Node with sequential positions and types.
// startPos is the starting position (inclusive), endPos is the ending position (inclusive)
func createNodeSlice(filename string, startPos, endPos int) []*syntax.Node {
	var nodes []*syntax.Node
	for i := 0; i <= endPos-startPos; i++ {
		pos := startPos + i
		end := pos + 1
		typ := i + 1
		nodes = append(nodes, &syntax.Node{
			Filename: filename,
			Pos:      pos,
			End:      end,
			Type:     typ,
		})
	}
	return nodes
}

// printFooterAndGetData is a helper function to call PrintFooter and return stats data
func printFooterAndGetData(t *testing.T, statsPrinter *stats) *StatsData {
	if err := statsPrinter.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}
	data := statsPrinter.GetStatsData()
	return data.(*StatsData)
}

func TestStatsDataAggregation(t *testing.T) {
	tests := []struct {
		name       string
		duplicates [][][]*syntax.Node
		filesCount int
		threshold  int
		checkStats func(t *testing.T, stats *StatsData)
	}{
		{
			name:       "empty clones",
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

			tt.checkStats(t, statsPrinter.GetStatsData().(*StatsData))
		})
	}
}

// createCloneNodeGroup creates a group of clone nodes with specified files
func createCloneNodeGroup(filenames []string) [][]*syntax.Node {
	var dups [][]*syntax.Node
	for _, filename := range filenames {
		nodes := []*syntax.Node{
			{Filename: filename, Pos: 1, End: 3, Type: 1},
			{Filename: filename, Pos: 2, End: 4, Type: 2},
		}
		dups = append(dups, nodes)
	}
	return dups
}

func TestStatsComplexityScore(t *testing.T) {
	var buf bytes.Buffer
	statsPrinter := NewStats(&buf, mockReadFile(string(mockReadFileContent())), 15).(*stats)
	statsPrinter.SetFilesCount(5)

	// Simulate: 3 clone groups, 9 total clones = complexity 3.0
	for range 3 {
		dups := createCloneNodeGroup([]string{"file1.go", "file2.go", "file3.go"})

		if err := statsPrinter.PrintClones(dups); err != nil {
			t.Fatalf("PrintClones failed: %v", err)
		}
	}

	if err := statsPrinter.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	statsData := statsPrinter.GetStatsData().(*StatsData)
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
		createNodeSlice("file1.go", 1, 4),
		createNodeSlice("file2.go", 10, 13),
		createNodeSlice("file3.go", 20, 23),
	}

	if err := statsPrinter.PrintClones(dups); err != nil {
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

	statsData := statsPrinter.GetStatsData().(*StatsData)
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
	// Expected format: "  1-5 lines      :   10 clones [████████████████████] 55.6%"
	expectedRanges := []string{"1-5 lines", "10 clones", "6-10 lines", "5 clones", "11-20 lines", "3 clones"}

	for _, expected := range expectedRanges {
		if !strings.Contains(output, expected) {
			t.Errorf("Output missing expected range: %s", expected)
		}
	}

	// Check that output contains bars
	if !strings.Contains(output, "█") {
		t.Error("Output doesn't contain ASCII bars")
	}

	// Check that output is sorted (1-5 should come before 6-10)
	idx1 := strings.Index(output, "1-5 lines")
	idx2 := strings.Index(output, "6-10 lines")
	if idx1 == -1 || idx2 == -1 || idx1 > idx2 {
		t.Error("Output is not properly sorted")
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

			statsData := statsPrinter.GetStatsData().(*StatsData)
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
	var result map[string]any
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, output)
	}

	// Verify structure exists
	config, ok := result["configuration"].(map[string]any)
	if !ok {
		t.Fatal("Missing 'configuration' section in JSON")
	}
	if config["threshold"] != float64(15) {
		t.Errorf("threshold = %v, want 15", config["threshold"])
	}
	if config["detectionMethods"] != "art-dupl,hash" {
		t.Errorf("detectionMethods = %v, want 'art-dupl,hash'", config["detectionMethods"])
	}

	overview, ok := result["overview"].(map[string]any)
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

	dupCode, ok := result["duplicateCode"].(map[string]any)
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

	sizeDist, ok := result["sizeDistribution"].(map[string]any)
	if !ok || len(sizeDist) == 0 {
		t.Error("Missing or empty 'sizeDistribution' section")
	}

	topFiles, ok := result["topFiles"].([]any)
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

func TestStatsCSVOutput(t *testing.T) {
	var buf bytes.Buffer

	sp := NewStats(&buf, mockReadFile("package main\nfunc foo(){return 0}\nfunc main(){}"), 1).(*stats)
	sp.SetFormat(FormatCSV)
	sp.SetFilesCount(3)
	sp.SetDetectionMethods("art-dupl")
	
	// Create some clones
	dups1 := [][]*syntax.Node{
		{{Filename: "file1.go", Pos: 2, End: 3}, {Filename: "file1.go", Pos: 2, End: 3}},
		{{Filename: "file2.go", Pos: 2, End: 3}, {Filename: "file2.go", Pos: 2, End: 3}},
	}
	
	if err := sp.PrintClones(dups1); err != nil {
		t.Fatalf("PrintClones failed: %v", err)
	}

	if err := sp.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	output := buf.String()

	// Verify CSV structure
	expectedHeaders := []string{
		"Metric,Value",
		"Threshold,",
		"Detection Methods,",
		"Files Scanned,",
		"Clone Groups,",
		"Total Clones,",
		"Total Duplicate Lines,",
		"Health Score,",
	}

	for _, header := range expectedHeaders {
		if !strings.Contains(output, header) {
			t.Errorf("CSV output missing header: %q", header)
		}
	}

	// Verify CSV has proper structure (lines separated by commas)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 10 {
		t.Errorf("CSV output should have at least 10 lines, got %d", len(lines))
	}

	// Check that each non-empty line has at least one comma
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if i > 0 && trimmed != "" && !strings.Contains(line, ",") {
			t.Errorf("CSV line %d doesn't contain comma: %q", i, line)
		}
	}

	// Verify Health Score is present
	if !strings.Contains(output, "Health Score,A") && !strings.Contains(output, "Health Score,") {
		t.Error("CSV output should contain Health Score")
	}
}

func TestHealthScoreCalculation(t *testing.T) {
	tests := []struct {
		name             string
		duplicationRatio float64
		complexityScore  float64
		impactScore      int
		expectedGrade    string
	}{
		{"Perfect health", 0.0, 0.0, 0, "A"},
		{"Excellent health", 2.0, 1.0, 500, "B"}, // Weighted score: 4.45% = B
		{"Good health", 5.0, 2.0, 1000, "C"},     // Weighted score: 10.75% = C
		{"Moderate health", 8.0, 3.0, 2000, "F"}, // Weighted score: 15.3% = F (threshold is < 15 for D)
		{"Poor health", 12.0, 4.0, 3000, "F"},    // Weighted score: 21.7% = F
		{"Critical health", 20.0, 5.0, 5000, "F"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			sp := NewStats(&buf, mockReadFile("package main\nfunc main(){}"), 15).(*stats)
			
			// Set up stats data
			sp.statsData.TotalDuplicateLines = int(tt.duplicationRatio * 10) // Simulating
			sp.statsData.TotalEstimatedLines = 1000
			sp.statsData.DuplicationRatio = tt.duplicationRatio // Set directly for health calculation
			sp.statsData.ComplexityScore = tt.complexityScore
			sp.statsData.ImpactScore = tt.impactScore
			
			grade := sp.calculateHealthScore()
			
			if grade != tt.expectedGrade {
				t.Errorf("calculateHealthScore() = %q, want %q for inputs (ratio=%.1f%%, complexity=%.2f, impact=%d)",
					grade, tt.expectedGrade, tt.duplicationRatio, tt.complexityScore, tt.impactScore)
			}
		})
	}
}

func TestPrintRecommendations(t *testing.T) {
	tests := []struct {
		name             string
		healthScore      string
		totalCloneGroups int
		averageCloneSize int
		complexityScore  float64
		shouldContain    []string
		shouldNotContain []string
	}{
		{
			name:             "Grade A recommendations",
			healthScore:      "A",
			totalCloneGroups: 1,
			averageCloneSize: 5,
			complexityScore:  1.0,
			shouldContain:    []string{"Excellent", "Keep up the good work"},
			shouldNotContain: []string{"action needed", "Critical"},
		},
		{
			name:             "Grade C recommendations with metrics",
			healthScore:      "C",
			totalCloneGroups: 15,
			averageCloneSize: 60,
			complexityScore:  3.5,
			shouldContain:    []string{"Moderate", "50+ lines", "15 clone groups"},
			shouldNotContain: []string{"Excellent", "Critical"},
		},
		{
			name:             "Grade F critical recommendations",
			healthScore:      "F",
			totalCloneGroups: 30,
			averageCloneSize: 80,
			complexityScore:  6.0,
			shouldContain:    []string{"Critical", "immediate action", "Halt new feature"},
			shouldNotContain: []string{"Excellent", "minor"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			sp := NewStats(&buf, mockReadFile("package main\nfunc main(){}"), 15).(*stats)
			
			// Set up stats data
			sp.statsData.HealthScore = tt.healthScore
			sp.statsData.TotalCloneGroups = tt.totalCloneGroups
			sp.statsData.AverageCloneSize = tt.averageCloneSize
			sp.statsData.ComplexityScore = tt.complexityScore
			sp.statsData.TotalFilesScanned = 10
			sp.statsData.TotalDuplicateLines = 500
			sp.statsData.TotalEstimatedLines = 1000
			
			// Call printRecommendations
			sp.printRecommendations()
			
			output := buf.String()
			
			// Check that all expected strings are present
			for _, expected := range tt.shouldContain {
				if !strings.Contains(output, expected) {
					t.Errorf("Recommendations output missing expected text: %q\nGot: %s", expected, output)
				}
			}
			
			// Check that unexpected strings are NOT present
			for _, notExpected := range tt.shouldNotContain {
				if strings.Contains(output, notExpected) {
					t.Errorf("Recommendations output should not contain: %q\nGot: %s", notExpected, output)
				}
			}
			
			// Verify next steps section is present
			if !strings.Contains(output, "Next Steps:") {
				t.Error("Recommendations should include 'Next Steps:' section")
			}
		})
	}
}
