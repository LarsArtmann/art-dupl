package printer

import (
	"bytes"
	"encoding/json"
	"strings"
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

	dups := [][]*syntax.Node{
		createTestSARIFNodes("test.go", 0, 50),
		createTestSARIFNodes("test2.go", 0, 50),
	}

	err := printer.PrintClones(dups)
	if err != nil {
		t.Errorf("PrintClones returned error: %v", err)
	}

	// Should have 2 results (one per clone instance)
	testutil.AssertCount(t, len(printer.results), 2, "results")
}

func TestSARIFPrinter_PrintClones_Empty(t *testing.T) {
	var buf bytes.Buffer

	printer := NewSARIF(&buf, mockSARIFReadFile, 15).(*sarifPrinter)

	err := printer.PrintClones([][]*syntax.Node{})
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
	_ = printer.PrintClones(dups)

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
	if sarifOutput.Version != "2.1.0" {
		t.Errorf("Expected version 2.1.0, got %s", sarifOutput.Version)
	}

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

	if printer.currentHash != "abc123" {
		t.Errorf("SetHash() = %s, expected %s", printer.currentHash, "abc123")
	}

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

	err := printer.PrintClones(dups1)
	if err != nil {
		t.Errorf("First PrintClones returned error: %v", err)
	}

	// Second group with same hash (should be filtered)
	printer.SetHash("same-hash")

	dups2 := [][]*syntax.Node{nodes1, nodes2}

	err = printer.PrintClones(dups2)
	if err != nil {
		t.Errorf("Second PrintClones returned error: %v", err)
	}

	// Should still have only 2 results (first group only)
	testutil.AssertCount(t, len(printer.results), 2, "results")
}

func TestSARIFOutput_FingerprintCompliance(t *testing.T) {
	var buf bytes.Buffer

	printer := NewSARIFWithConfig(&buf, mockSARIFReadFile, SARIFConfig{
		Threshold: 15,
		Version:   "1.0.0",
	}).(*sarifPrinter)

	// Add test data with a known hash
	printer.SetHash("abc123def456")

	dups := [][]*syntax.Node{
		createTestSARIFNodes("test.go", 0, 100),
	}
	_ = printer.PrintClones(dups)
	_ = printer.PrintFooter()

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
		t.Errorf("Expected contentFingerprint 'abc123def456', got '%s'", result.Fingerprints.ContentFingerprint)
	}

	// partialFingerprint should be first 8 chars of contentFingerprint
	if result.Fingerprints.PartialFingerprint != "abc123de" {
		t.Errorf("Expected partialFingerprint 'abc123de', got '%s'", result.Fingerprints.PartialFingerprint)
	}
}

func TestSARIFOutput_VersionFromConfig(t *testing.T) {
	var buf bytes.Buffer

	// Test that version is properly set from SARIFConfig
	printer := NewSARIFWithConfig(&buf, mockSARIFReadFile, SARIFConfig{
		Threshold: 15,
		Version:   "2.0.0-test",
	}).(*sarifPrinter)

	if printer.version != "2.0.0-test" {
		t.Errorf("Expected version '2.0.0-test', got '%s'", printer.version)
	}
}

func TestSARIFOutput_Structure(t *testing.T) {
	var buf bytes.Buffer

	printer := NewSARIFWithConfig(&buf, mockSARIFReadFile, SARIFConfig{
		Threshold: 15,
		Version:   "1.0.0",
	}).(*sarifPrinter)

	// Add test data
	printer.SetHash("test-hash-123")

	dups := [][]*syntax.Node{
		createTestSARIFNodes("main.go", 0, 100),
		createTestSARIFNodes("utils.go", 0, 100),
	}
	_ = printer.PrintClones(dups)
	_ = printer.PrintFooter()

	var output SARIFOutput

	err := json.Unmarshal(buf.Bytes(), &output)
	if err != nil {
		t.Fatalf("Failed to unmarshal SARIF output: %v", err)
	}

	// Check schema
	if !strings.Contains(output.Schema, "sarif-schema-2.1.0") {
		t.Errorf("Schema should contain sarif-schema-2.1.0, got: %s", output.Schema)
	}

	// Check tool info
	if len(output.Runs) == 0 {
		t.Fatal("Expected at least one run")
	}

	tool := output.Runs[0].Tool.Driver
	if tool.Name != "art-dupl" {
		t.Errorf("Expected tool name 'art-dupl', got '%s'", tool.Name)
	}

	if tool.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", tool.Version)
	}

	// Check rules
	if len(tool.Rules) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(tool.Rules))
	}

	rule := tool.Rules[0]
	if rule.ID != "art-dupl/duplicate-code" {
		t.Errorf("Expected rule ID 'art-dupl/duplicate-code', got '%s'", rule.ID)
	}

	// Check results
	results := output.Runs[0].Results
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Check result structure
	for _, result := range results {
		if result.RuleID != "art-dupl/duplicate-code" {
			t.Errorf("Expected RuleID 'art-dupl/duplicate-code', got '%s'", result.RuleID)
		}

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
