package artdupl

import (
	"fmt"
	"testing"
	"time"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

// assertCloneGroupBasic asserts basic CloneGroup properties (Hash and Clone count).
func assertCloneGroupBasic(
	t *testing.T,
	group *CloneGroup,
	expectedHash string,
	expectedCloneCount int,
) {
	t.Helper()

	if group.Hash != expectedHash {
		t.Errorf("Expected Hash=%q, got %s", expectedHash, group.Hash)
	}

	if len(group.Clones) != expectedCloneCount {
		t.Errorf("Expected %d clones, got %d", expectedCloneCount, len(group.Clones))
	}
}

// assertCloneCount asserts the number of clones matches expected.
func assertCloneCount(t *testing.T, actual, expected int) {
	t.Helper()

	if actual != expected {
		t.Errorf("clone count: expected %d, actual %d", expected, actual)
	}
}

// assertMethodsCount asserts the number of detection methods matches expected.
func assertMethodsCount(t *testing.T, got, expected int) {
	t.Helper()

	if got != expected {
		t.Errorf("methods count: expected %d, got %d", expected, got)
	}
}

// TestCloneGroup_Fields tests CloneGroup field assignments.
func TestCloneGroup_Fields(t *testing.T) {
	group := CloneGroup{
		Hash: "abc123",
		Clones: []*Clone{
			{Filename: "file1.go", StartLine: 10, EndLine: 20},
			{Filename: "file2.go", StartLine: 15, EndLine: 25},
		},
		Size:      100,
		LineCount: 11,
		Method:    MethodArtDupl,
	}

	assertCloneGroupBasic(t, &group, "abc123", 2)

	testutil.AssertFieldValue(t, group.Size, 100, "Size")
	testutil.AssertFieldValue(t, group.LineCount, 11, "LineCount")
	testutil.AssertFieldValue(t, group.Method, MethodArtDupl, "Method")
}

// TestClone_Fields tests Clone field assignments.
func TestClone_Fields(t *testing.T) {
	clone := Clone{
		Filename:  testFilename,
		StartLine: 10,
		EndLine:   20,
		StartPos:  100,
		EndPos:    200,
		Fragment:  "code here",
		Size:      50,
	}

	testutil.AssertFieldValue(t, clone.Filename, testFilename, "Filename")
	testutil.AssertFieldValue(t, clone.StartLine, 10, "StartLine")
	testutil.AssertFieldValue(t, clone.EndLine, 20, "EndLine")
	testutil.AssertFieldValue(t, clone.StartPos, 100, "StartPos")
	testutil.AssertFieldValue(t, clone.EndPos, 200, "EndPos")
	testutil.AssertFieldValue(t, clone.Fragment, "code here", "Fragment")
	testutil.AssertFieldValue(t, clone.Size, 50, "Size")
}

// TestResult_Fields tests Result field assignments.
func TestResult_Fields(t *testing.T) {
	result := Result{
		CloneGroups: []*CloneGroup{
			{Hash: "hash1"},
			{Hash: "hash2"},
		},
		Summary: &Summary{
			TotalFiles:  10,
			TotalClones: 5,
			TotalGroups: 2,
		},
		Metadata: &Metadata{
			Version: "dev",
		},
	}

	if len(result.CloneGroups) != 2 {
		t.Errorf("Should have 2 clone groups, got %d", len(result.CloneGroups))
	}

	if result.Summary.TotalFiles != 10 {
		t.Errorf("TotalFiles should be 10, got %d", result.Summary.TotalFiles)
	}

	if result.Metadata.Version == "" {
		t.Errorf("Version should not be empty")
	}
}

// TestSummary_Fields tests Summary field assignments.
func TestSummary_Fields(t *testing.T) {
	summary := Summary{
		TotalFiles:    100,
		TotalClones:   50,
		TotalGroups:   25,
		AnalysisTime:  5 * time.Second,
		MethodsUsed:   []DetectionMethod{MethodArtDupl, MethodHash},
		LinesAnalyzed: 10000,
	}

	testutil.AssertFieldValue(t, summary.TotalFiles, 100, "TotalFiles")
	testutil.AssertFieldValue(t, summary.TotalClones, 50, "TotalClones")
	testutil.AssertFieldValue(t, summary.TotalGroups, 25, "TotalGroups")

	testutil.AssertFieldValue(t, summary.AnalysisTime, 5*time.Second, "AnalysisTime")

	assertMethodsCount(t, len(summary.MethodsUsed), 2)

	testutil.AssertFieldValue(t, summary.LinesAnalyzed, 10000, "LinesAnalyzed")
}

// TestMetadata_Fields tests Metadata field assignments.
func TestMetadata_Fields(t *testing.T) {
	now := time.Now()
	metadata := Metadata{
		Version:    "2.0.0",
		Timestamp:  now,
		ConfigHash: "config-hash-123",
		Toolchain:  "go1.21",
	}

	testutil.AssertFieldValue(t, metadata.Version, "2.0.0", "Version")
	testutil.AssertFieldValue(t, metadata.Timestamp, now, "Timestamp")
	testutil.AssertFieldValue(t, metadata.ConfigHash, "config-hash-123", "ConfigHash")
	testutil.AssertFieldValue(t, metadata.Toolchain, "go1.21", "Toolchain")
}

// TestProgress_Fields tests Progress field assignments.
func TestProgress_Fields(t *testing.T) {
	progress := newTestProgress("parsing", 75, 100, 75.5, "Processing files", "main.go")

	testutil.AssertFieldValue(t, progress.Stage, "parsing", "Stage")

	testutil.AssertFieldValue(t, progress.Completed, 75, "Completed")

	testutil.AssertFieldValue(t, progress.Total, 100, "Total")

	testutil.AssertFieldValue(t, progress.Percentage, 75.5, "Percentage")

	testutil.AssertFieldValue(t, progress.Message, "Processing files", "Message")

	testutil.AssertFieldValue(t, progress.CurrentFile, "main.go", "CurrentFile")
}

// TestClone_Empty tests empty Clone struct.
func TestClone_Empty(t *testing.T) {
	var clone Clone

	testutil.AssertFieldValue(t, clone.Filename, "", "Filename")

	testutil.AssertFieldValue(t, clone.StartLine, 0, "StartLine")

	testutil.AssertFieldValue(t, clone.Size, 0, "Size")
}

// TestCloneGroup_Empty tests empty CloneGroup struct.
func TestCloneGroup_Empty(t *testing.T) {
	var group CloneGroup

	testutil.AssertFieldValue(t, group.Hash, "", "Hash")

	if group.Clones != nil {
		t.Errorf("Expected nil Clones, got %v", group.Clones)
	}

	testutil.AssertFieldValue(t, group.Size, 0, "Size")
}

// TestResult_Empty tests empty Result struct.
func TestResult_Empty(t *testing.T) {
	var result Result

	if result.CloneGroups != nil {
		t.Errorf("Expected nil CloneGroups, got %v", result.CloneGroups)
	}

	testutil.AssertFieldValue(t, result.Summary, nil, "Summary")

	testutil.AssertFieldValue(t, result.Metadata, nil, "Metadata")
}

// TestClone_WithFragment tests Clone with fragment.
func TestClone_WithFragment(t *testing.T) {
	clone := createTestClone(t, "package main\n\nfunc main() {}\n")

	if clone.Fragment == "" {
		t.Error("Clone with fragment should have non-empty Fragment")
	}
}

// TestClone_WithoutFragment tests Clone without fragment content.
func TestClone_WithoutFragment(t *testing.T) {
	clone := createTestClone(t, "")

	testutil.AssertFieldValue(t, clone.Fragment, "", "Fragment")
}

// TestClone_LargeFragment tests Clone with large fragment.
func TestClone_LargeFragment(t *testing.T) {
	largeFragment := make([]byte, 10000)
	for i := range largeFragment {
		largeFragment[i] = 'x'
	}

	clone := Clone{
		Filename:  "large.go",
		StartLine: 1,
		EndLine:   1000,
		Fragment:  string(largeFragment),
		Size:      len(largeFragment),
	}

	testutil.AssertFieldValue(t, clone.Filename, "large.go", "Filename")
	testutil.AssertFieldValue(t, clone.StartLine, 1, "StartLine")
	testutil.AssertFieldValue(t, clone.EndLine, 1000, "EndLine")
	testutil.AssertFieldValue(t, clone.Size, len(largeFragment), "Size")

	if len(clone.Fragment) != 10000 {
		t.Errorf("Fragment should be 10000 bytes, got %d", len(clone.Fragment))
	}
}

// TestCloneGroup_ManyClones tests CloneGroup with many clones.
func TestCloneGroup_ManyClones(t *testing.T) {
	clones := make([]*Clone, 100)
	for i := range clones {
		clones[i] = &Clone{
			Filename:  fmt.Sprintf("file%d.go", i),
			StartLine: i * 10,
			EndLine:   i*10 + 5,
			Size:      50,
		}
	}

	group := CloneGroup{
		Hash:   "many-clones",
		Clones: clones,
		Size:   5000,
		Method: MethodHash,
	}

	testutil.AssertFieldValue(t, group.Hash, "many-clones", "Hash")
	testutil.AssertFieldValue(t, group.Size, 5000, "Size")
	testutil.AssertFieldValue(t, group.Method, MethodHash, "Method")
	assertCloneCount(t, 100, len(group.Clones))
}

// TestSummary_AllMethodsUsed tests Summary with all detection methods.
func TestSummary_AllMethodsUsed(t *testing.T) {
	summary := Summary{
		TotalFiles:  10,
		TotalClones: 20,
		TotalGroups: 5,
		MethodsUsed: []DetectionMethod{
			MethodArtDupl,
			MethodHash,
			MethodTodos,
			MethodLegacy,
		},
	}

	testutil.AssertFieldValue(t, summary.TotalFiles, 10, "TotalFiles")
	testutil.AssertFieldValue(t, summary.TotalClones, 20, "TotalClones")
	testutil.AssertFieldValue(t, summary.TotalGroups, 5, "TotalGroups")
	assertMethodsCount(t, 4, len(summary.MethodsUsed))
}

// TestProgress_Zero tests Progress with zero values.
func TestProgress_Zero(t *testing.T) {
	progress := Progress{}

	testutil.AssertFieldValue(t, progress.Stage, "", "Stage")

	testutil.AssertFieldValue(t, progress.Percentage, 0, "Percentage")
}

// TestProgress_Full tests Progress at 100%.
func TestProgress_Full(t *testing.T) {
	progress := newTestProgress("complete", 100, 100, 100.0, "Done", "")

	testutil.AssertFieldValue(t, progress.Percentage, 100.0, "Percentage")

	if progress.Completed != progress.Total {
		t.Error("At 100%, Completed should equal Total")
	}
}

// TestClone_EdgeCases tests edge cases for Clone struct.
func TestClone_EdgeCases(t *testing.T) {
	clone := Clone{
		Filename:  "single.go",
		StartLine: 10,
		EndLine:   10,
		Size:      1,
	}

	testutil.AssertFieldValue(t, clone.Filename, "single.go", "Filename")
	testutil.AssertFieldValue(t, clone.Size, 1, "Size")

	if clone.StartLine != clone.EndLine {
		t.Error("Single-line clone should have same start and end line")
	}
}

// TestCloneGroup_EmptyClones tests CloneGroup with empty clones slice.
func TestCloneGroup_EmptyClones(t *testing.T) {
	group := CloneGroup{
		Hash:   "empty",
		Clones: []*Clone{},
		Size:   0,
		Method: MethodArtDupl,
	}

	testutil.AssertFieldValue(t, group.Hash, "empty", "Hash")
	testutil.AssertFieldValue(t, group.Size, 0, "Size")
	testutil.AssertFieldValue(t, group.Method, MethodArtDupl, "Method")
	assertCloneCount(t, 0, len(group.Clones))
}

// TestCloneGroup_NilClones tests CloneGroup with nil clones.
func TestCloneGroup_NilClones(t *testing.T) {
	group := CloneGroup{
		Hash:   "nil-clones",
		Clones: nil,
		Size:   0,
		Method: MethodArtDupl,
	}

	testutil.AssertFieldValue(t, group.Hash, "nil-clones", "Hash")
	testutil.AssertFieldValue(t, group.Size, 0, "Size")
	testutil.AssertFieldValue(t, group.Method, MethodArtDupl, "Method")

	if group.Clones != nil {
		t.Error("Clones should be nil")
	}
}
