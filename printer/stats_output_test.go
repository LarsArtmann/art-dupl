package printer

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax"
)

func TestStatsJSONOutput(t *testing.T) {
	statsPrinter, buf := newTestStatsPrinter()

	// Add some test data
	statsPrinter.SetDetectionMethods("art-dupl,hash")
	statsPrinter.SetFilesCount(10)

	// Add one clone group with duplicates
	dups := [][]*syntax.Node{
		{
			testutil.CreateNodeWithPos(1, "file1.go", 1, 3),
			testutil.CreateNodeWithPos(2, "file1.go", 2, 4),
		},
		{
			testutil.CreateNodeWithPos(1, "file2.go", 10, 12),
			testutil.CreateNodeWithPos(2, "file2.go", 11, 13),
		},
	}

	err := printTestClones(statsPrinter, mockReadFile(string(mockReadFileContent())), dups)
	if err != nil {
		t.Fatalf("PrintClones failed: %v", err)
	}

	// Set format to JSON
	statsPrinter.format = config.OutputFormatJSON

	err = statsPrinter.PrintFooter()
	if err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	output := buf.String()

	// Verify JSON is valid
	var result map[string]any

	err = json.Unmarshal([]byte(output), &result)
	if err != nil {
		t.Fatalf("Output is not valid JSON: %v\nOutput: %s", err, output)
	}

	// Verify structure exists
	cfgSection, ok := result["configuration"].(map[string]any)
	if !ok {
		t.Fatal("Missing 'configuration' section in JSON")
	}

	if cfgSection["threshold"] != float64(15) {
		t.Errorf("threshold = %v, want 15", cfgSection["threshold"])
	}

	if cfgSection["detectionMethods"] != "art-dupl,hash" {
		t.Errorf("detectionMethods = %v, want 'art-dupl,hash'", cfgSection["detectionMethods"])
	}

	overview, ok := result["overview"].(map[string]any)
	if !ok {
		t.Fatal("Missing 'overview' section in JSON")
	}

	assertMapFloat64Equal(t, overview, "filesScanned", 10)
	assertMapFloat64Equal(t, overview, "cloneGroups", 1)
	assertMapFloat64Equal(t, overview, "totalClones", 2)

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
	statsPrinter, buf := newTestStatsPrinter()
	statsPrinter.SetFilesCount(5)
	statsPrinter.SetDetectionMethods("art-dupl")

	// Set format to text
	statsPrinter.format = config.OutputFormatText

	err := statsPrinter.PrintFooter()
	if err != nil {
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
		testutil.AssertStringContains(t, output, section, "Output missing section: "+section)
	}
}

func TestStatsCSVOutput(t *testing.T) {
	var buf bytes.Buffer

	sp := NewStats(&buf, mockReadFile("package main\nfunc foo(){return 0}\nfunc main(){}"), 1).(*stats)
	sp.SetFormat(config.OutputFormatCSV)
	sp.SetFilesCount(3)
	sp.SetDetectionMethods("art-dupl")

	// Create some clones
	dups1 := createTestCloneGroups()

	err := sp.PrintClones(
		processTestNodes(mockReadFile(string(mockReadFileContent())), "test", dups1),
	)
	if err != nil {
		t.Fatalf("PrintClones failed: %v", err)
	}

	err = sp.PrintFooter()
	if err != nil {
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
		testutil.AssertStringContains(t, output, header, "CSV output missing header: "+header)
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
	testutil.AssertStringContains(
		t,
		output,
		"Health Score",
		"CSV output should contain Health Score",
	)
}
