package printer

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// mockSARIFReadFile is a mock file reader for testing.
func mockSARIFReadFile(filename string) ([]byte, error) {
	return []byte("line1\nline2\nline3\nline4\nline5\n"), nil
}

// createTestNodes creates test nodes for SARIF testing.
func createTestSARIFNodes(filename string, startPos, endPos int32) []*syntax.Node {
	return []*syntax.Node{
		{Type: 1, Filename: filename, Pos: startPos, End: startPos + 10},
		{Type: 2, Filename: filename, Pos: startPos + 10, End: endPos},
	}
}

func newTestSARIFPrinter(buf *bytes.Buffer, version string) *sarifPrinter {
	return NewSARIFWithConfig(buf, mockSARIFReadFile, SARIFConfig{
		Threshold: 15,
		Version:   version,
	}).(*sarifPrinter)
}

func TestNewSARIF(t *testing.T) {
	var buf bytes.Buffer

	printer := NewSARIF(&buf, mockSARIFReadFile, 15)

	if printer == nil {
		t.Fatal("NewSARIF returned nil")
	}
}

func TestSARIFPrinter_PrintHeader(t *testing.T) {
	var buf bytes.Buffer

	printer := NewSARIF(&buf, mockSARIFReadFile, 15)

	err := printer.PrintHeader()
	if err != nil {
		t.Errorf("PrintHeader returned error: %v", err)
	}

	// Header should not write anything
	if buf.Len() != 0 {
		t.Errorf("PrintHeader wrote %d bytes, expected 0", buf.Len())
	}
}

func TestSARIFPrinter_PrintClones(t *testing.T) {
	var buf bytes.Buffer

	printer := NewSARIF(&buf, mockSARIFReadFile, 15).(*sarifPrinter)

	// Create test clone group
	printer.SetHash("test-hash")

	sarifDups := [][]*syntax.Node{
		createTestSARIFNodes("test.go", 0, 50),
		createTestSARIFNodes("test2.go", 0, 50),
	}

	err := printer.PrintClones(processTestNodes(mockSARIFReadFile, "test", sarifDups))
	if err != nil {
		t.Errorf("PrintClones returned error: %v", err)
	}

	// Should have 2 results (one per clone instance)
	testutil.AssertCount(t, len(printer.results), 2, "results")
}

func TestSARIFPrinter_PrintClones_Empty(t *testing.T) {
	var buf bytes.Buffer

	printer := NewSARIF(&buf, mockSARIFReadFile, 15).(*sarifPrinter)

	err := printer.PrintClones(processTestNodes(mockSARIFReadFile, "test", [][]*syntax.Node{}))
	if err != nil {
		t.Errorf("PrintClones returned error: %v", err)
	}

	testutil.AssertCount(t, len(printer.results), 0, "results")
}

func TestSARIFPrinter_PrintFooter(t *testing.T) {
	var buf bytes.Buffer

	printer := NewSARIF(&buf, mockSARIFReadFile, 15).(*sarifPrinter)

	// Add some test data
	printer.SetHash("test-hash")

	dups := [][]*syntax.Node{
		createTestSARIFNodes("test.go", 0, 50),
	}
	_ = printer.PrintClones(processTestNodes(mockSARIFReadFile, "test", dups))

	err := printer.PrintFooter()
	if err != nil {
		t.Errorf("PrintFooter returned error: %v", err)
	}

	// Verify output is valid JSON
	var sarifOutput SARIFOutput
	if err := json.Unmarshal(buf.Bytes(), &sarifOutput); err != nil {
		t.Errorf("PrintFooter output is not valid JSON: %v", err)
	}

	// Verify SARIF structure
	testutil.AssertFieldValue(t, sarifOutput.Version, "2.1.0", "Version")

	if len(sarifOutput.Runs) != 1 {
		t.Errorf("Expected 1 run, got %d", len(sarifOutput.Runs))
	}

	if len(sarifOutput.Runs[0].Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(sarifOutput.Runs[0].Results))
	}
}

func TestSARIFPrinter_DetermineLevel(t *testing.T) {
	var buf bytes.Buffer

	printer := NewSARIF(&buf, mockSARIFReadFile, 10).(*sarifPrinter)

	tests := []struct {
		size     int
		expected string
	}{
		{50, "error"},   // 50 >= 10*4
		{30, "warning"}, // 30 >= 10*2
		{15, "note"},    // 15 < 10*2
		{10, "note"},    // 10 < 10*2
	}

	for _, tt := range tests {
		level := printer.determineLevel(tt.size)
		if level != tt.expected {
			t.Errorf("determineLevel(%d) = %s, expected %s", tt.size, level, tt.expected)
		}
	}
}

func TestSARIFPrinter_SetHash(t *testing.T) {
	var buf bytes.Buffer

	printer := NewSARIF(&buf, mockSARIFReadFile, 15).(*sarifPrinter)

	// Test SetHash stores the hash
	printer.SetHash("abc123")

	testutil.AssertFieldValue(t, printer.currentHash, "abc123", "currentHash")

	// Test that empty hash is handled
	printer.SetHash("")

	if printer.currentHash != "" {
		t.Errorf("SetHash() empty = %s, expected %s", printer.currentHash, "")
	}
}

func TestSARIFPrinter_DuplicateHashFiltering(t *testing.T) {
	var buf bytes.Buffer

	printer := NewSARIF(&buf, mockSARIFReadFile, 15).(*sarifPrinter)

	// Create test clone groups with same hash
	nodes1 := createTestSARIFNodes("test.go", 0, 50)
	nodes2 := createTestSARIFNodes("test2.go", 0, 50)

	// First group with hash "same-hash"
	printer.SetHash("same-hash")

	dups1 := [][]*syntax.Node{nodes1, nodes2}

	err := printer.PrintClones(processTestNodes(mockSARIFReadFile, "test", dups1))
	if err != nil {
		t.Errorf("First PrintClones returned error: %v", err)
	}

	// Second group with same hash (should be filtered)
	printer.SetHash("same-hash")

	dups2 := [][]*syntax.Node{nodes1, nodes2}

	err = printer.PrintClones(processTestNodes(mockSARIFReadFile, "test", dups2))
	if err != nil {
		t.Errorf("Second PrintClones returned error: %v", err)
	}

	// Should still have only 2 results (first group only)
	testutil.AssertCount(t, len(printer.results), 2, "results")
}

func TestSARIFOutput_FingerprintCompliance(t *testing.T) {
	var buf bytes.Buffer

	sarifPrinter := newTestSARIFPrinter(&buf, "1.0.0")

	// Add test data with a known hash
	sarifPrinter.SetHash("abc123def456")

	dups := [][]*syntax.Node{
		createTestSARIFNodes("test.go", 0, 100),
	}
	_ = sarifPrinter.PrintClones(processTestNodes(mockSARIFReadFile, "test", dups))
	_ = sarifPrinter.PrintFooter()

	var output SARIFOutput

	err := json.Unmarshal(buf.Bytes(), &output)
	if err != nil {
		t.Fatalf("Failed to unmarshal SARIF output: %v", err)
	}

	// Verify results exist
	if len(output.Runs) == 0 || len(output.Runs[0].Results) == 0 {
		t.Fatal("Expected at least one result")
	}

	result := output.Runs[0].Results[0]

	// Verify fingerprint fields per SARIF spec
	if result.Fingerprints.ContentFingerprint == "" {
		t.Error("Expected non-empty contentFingerprint")
	}

	if result.Fingerprints.ContentFingerprint != "abc123def456" {
		t.Errorf(
			"Expected contentFingerprint 'abc123def456', got '%s'",
			result.Fingerprints.ContentFingerprint,
		)
	}

	// partialFingerprint should be first 8 chars of contentFingerprint
	if result.Fingerprints.PartialFingerprint != "abc123de" {
		t.Errorf(
			"Expected partialFingerprint 'abc123de', got '%s'",
			result.Fingerprints.PartialFingerprint,
		)
	}
}

func TestSARIFOutput_VersionFromConfig(t *testing.T) {
	var buf bytes.Buffer

	sarifPrinter := newTestSARIFPrinter(&buf, "2.0.0-test")

	testutil.AssertFieldValue(t, sarifPrinter.version, "2.0.0-test", "version")
}

func TestSARIFOutput_Structure(t *testing.T) {
	var buf bytes.Buffer

	sarifInstance := newTestSARIFPrinter(&buf, "1.0.0")

	// Add test data
	sarifInstance.SetHash("test-hash-123")

	dups := [][]*syntax.Node{
		createTestSARIFNodes("main.go", 0, 100),
		createTestSARIFNodes("utils.go", 0, 100),
	}
	_ = sarifInstance.PrintClones(processTestNodes(mockSARIFReadFile, "test", dups))
	_ = sarifInstance.PrintFooter()

	var output SARIFOutput

	err := json.Unmarshal(buf.Bytes(), &output)
	if err != nil {
		t.Fatalf("Failed to unmarshal SARIF output: %v", err)
	}

	// Check schema
	testutil.AssertStringContains(
		t,
		output.Schema,
		"sarif-schema-2.1.0",
		"Schema should contain sarif-schema-2.1.0",
	)

	// Check tool info
	if len(output.Runs) == 0 {
		t.Fatal("Expected at least one run")
	}

	tool := output.Runs[0].Tool.Driver
	testutil.AssertFieldValue(t, tool.Name, "art-dupl", "Name")

	testutil.AssertFieldValue(t, tool.Version, "1.0.0", "Version")

	// Check rules
	if len(tool.Rules) != 3 {
		t.Errorf("Expected 3 rules (duplicate-code, todo, legacy), got %d", len(tool.Rules))
	}

	rule := tool.Rules[0]
	testutil.AssertFieldValue(t, rule.ID, "art-dupl/duplicate-code", "ID")

	// Check results
	results := output.Runs[0].Results
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Check result structure
	for _, result := range results {
		testutil.AssertFieldValue(t, result.RuleID, "art-dupl/duplicate-code", "RuleID")

		if len(result.Locations) != 1 {
			t.Errorf("Expected 1 location, got %d", len(result.Locations))
		}

		loc := result.Locations[0].PhysicalLocation
		if loc.ArtifactLocation.URI == "" {
			t.Error("Expected non-empty URI")
		}

		if loc.Region.StartLine == 0 {
			t.Error("Expected non-zero StartLine")
		}
	}
}
