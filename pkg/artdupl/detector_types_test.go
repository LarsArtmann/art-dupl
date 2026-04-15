package artdupl

import (
	"fmt"
	"testing"
	"time"
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
func assertCloneCount(t *testing.T, got, expected int) {
	t.Helper()

	if got != expected {
		t.Errorf("clone count: expected %d, got %d", expected, got)
	}
}

// assertMethodsCount asserts the number of detection methods matches expected.
func assertMethodsCount(t *testing.T, got, expected int) {
	t.Helper()

	if got != expected {
		t.Errorf("methods count: expected %d, got %d", expected, got)
	}
}

// assertSummaryFields asserts Summary field values.
func assertSummaryFields(
	t *testing.T,
	summary *Summary,
	files, clones, groups int,
	methods int,
) {
	t.Helper()

	if summary.TotalFiles != files {
		t.Errorf("TotalFiles: expected %d, got %d", files, summary.TotalFiles)
	}

	if summary.TotalClones != clones {
		t.Errorf("TotalClones: expected %d, got %d", clones, summary.TotalClones)
	}

	if summary.TotalGroups != groups {
		t.Errorf("TotalGroups: expected %d, got %d", groups, summary.TotalGroups)
	}

	if len(summary.MethodsUsed) != methods {
		t.Errorf("MethodsUsed: expected %d, got %d", methods, len(summary.MethodsUsed))
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

	if group.Size != 100 {
		t.Errorf("Size should be 100, got %d", group.Size)
	}

	if group.LineCount != 11 {
		t.Errorf("LineCount should be 11, got %d", group.LineCount)
	}

	if group.Method != MethodArtDupl {
		t.Errorf("Method should be MethodArtDupl, got %v", group.Method)
	}
}

// TestClone_Fields tests Clone field assignments.
func TestClone_Fields(t *testing.T) {
	clone := Clone{
		Filename:  "test.go",
		StartLine: 10,
		EndLine:   20,
		StartPos:  100,
		EndPos:    200,
		Fragment:  "code here",
		Size:      50,
	}

	if clone.Filename != "test.go" {
		t.Errorf("Filename should be 'test.go', got %s", clone.Filename)
	}

	if clone.StartLine != 10 {
		t.Errorf("StartLine should be 10, got %d", clone.StartLine)
	}

	if clone.EndLine != 20 {
		t.Errorf("EndLine should be 20, got %d", clone.EndLine)
	}

	if clone.StartPos != 100 {
		t.Errorf("StartPos should be 100, got %d", clone.StartPos)
	}

	if clone.EndPos != 200 {
		t.Errorf("EndPos should be 200, got %d", clone.EndPos)
	}

	if clone.Fragment != "code here" {
		t.Errorf("Fragment should be 'code here', got %s", clone.Fragment)
	}

	if clone.Size != 50 {
		t.Errorf("Size should be 50, got %d", clone.Size)
	}
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
			Version: "1.0.0",
		},
	}

	if len(result.CloneGroups) != 2 {
		t.Errorf("Should have 2 clone groups, got %d", len(result.CloneGroups))
	}

	if result.Summary.TotalFiles != 10 {
		t.Errorf("TotalFiles should be 10, got %d", result.Summary.TotalFiles)
	}

	if result.Metadata.Version != "1.0.0" {
		t.Errorf("Version should be '1.0.0', got %s", result.Metadata.Version)
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

	if summary.TotalFiles != 100 {
		t.Errorf("TotalFiles should be 100, got %d", summary.TotalFiles)
	}

	if summary.TotalClones != 50 {
		t.Errorf("TotalClones should be 50, got %d", summary.TotalClones)
	}

	if summary.TotalGroups != 25 {
		t.Errorf("TotalGroups should be 25, got %d", summary.TotalGroups)
	}

	if summary.AnalysisTime != 5*time.Second {
		t.Errorf("AnalysisTime should be 5s, got %v", summary.AnalysisTime)
	}

	assertMethodsCount(t, len(summary.MethodsUsed), 2)

	if summary.LinesAnalyzed != 10000 {
		t.Errorf("LinesAnalyzed should be 10000, got %d", summary.LinesAnalyzed)
	}
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

	if metadata.Version != "2.0.0" {
		t.Errorf("Version should be '2.0.0', got %s", metadata.Version)
	}

	if !metadata.Timestamp.Equal(now) {
		t.Errorf("Timestamp should be %v, got %v", now, metadata.Timestamp)
	}

	if metadata.ConfigHash != "config-hash-123" {
		t.Errorf("ConfigHash should be 'config-hash-123', got %s", metadata.ConfigHash)
	}

	if metadata.Toolchain != "go1.21" {
		t.Errorf("Toolchain should be 'go1.21', got %s", metadata.Toolchain)
	}
}

// TestProgress_Fields tests Progress field assignments.
func TestProgress_Fields(t *testing.T) {
	progress := newTestProgress("parsing", 75, 100, 75.5, "Processing files", "main.go")

	if progress.Stage != "parsing" {
		t.Errorf("Stage should be 'parsing', got %s", progress.Stage)
	}

	if progress.Completed != 75 {
		t.Errorf("Completed should be 75, got %d", progress.Completed)
	}

	if progress.Total != 100 {
		t.Errorf("Total should be 100, got %d", progress.Total)
	}

	if progress.Percentage != 75.5 {
		t.Errorf("Percentage should be 75.5, got %f", progress.Percentage)
	}

	if progress.Message != "Processing files" {
		t.Errorf("Message should be 'Processing files', got %s", progress.Message)
	}

	if progress.CurrentFile != "main.go" {
		t.Errorf("CurrentFile should be 'main.go', got %s", progress.CurrentFile)
	}
}

// TestClone_Empty tests empty Clone struct.
func TestClone_Empty(t *testing.T) {
	var clone Clone

	if clone.Filename != "" {
		t.Errorf("Empty Clone Filename should be empty string, got %s", clone.Filename)
	}

	if clone.StartLine != 0 {
		t.Errorf("Empty Clone StartLine should be 0, got %d", clone.StartLine)
	}

	if clone.Size != 0 {
		t.Errorf("Empty Clone Size should be 0, got %d", clone.Size)
	}
}

// TestCloneGroup_Empty tests empty CloneGroup struct.
func TestCloneGroup_Empty(t *testing.T) {
	var group CloneGroup

	if group.Hash != "" {
		t.Errorf("Empty CloneGroup Hash should be empty string, got %s", group.Hash)
	}

	if group.Clones != nil {
		t.Errorf("Empty CloneGroup Clones should be nil, got %v", group.Clones)
	}

	if group.Size != 0 {
		t.Errorf("Empty CloneGroup Size should be 0, got %d", group.Size)
	}
}

// TestResult_Empty tests empty Result struct.
func TestResult_Empty(t *testing.T) {
	var result Result

	if result.CloneGroups != nil {
		t.Errorf("Empty Result CloneGroups should be nil, got %v", result.CloneGroups)
	}

	if result.Summary != nil {
		t.Errorf("Empty Result Summary should be nil, got %v", result.Summary)
	}

	if result.Metadata != nil {
		t.Errorf("Empty Result Metadata should be nil, got %v", result.Metadata)
	}
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

	if clone.Fragment != "" {
		t.Errorf("Clone without fragment should have empty Fragment, got %s", clone.Fragment)
	}
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

	assertMethodsCount(t, 4, len(summary.MethodsUsed))
}

// TestProgress_Zero tests Progress with zero values.
func TestProgress_Zero(t *testing.T) {
	progress := Progress{}

	if progress.Stage != "" {
		t.Errorf("Zero Progress Stage should be empty, got %s", progress.Stage)
	}

	if progress.Percentage != 0 {
		t.Errorf("Zero Progress Percentage should be 0, got %f", progress.Percentage)
	}
}

// TestProgress_Full tests Progress at 100%.
func TestProgress_Full(t *testing.T) {
	progress := newTestProgress("complete", 100, 100, 100.0, "Done", "")

	if progress.Percentage != 100.0 {
		t.Errorf("Progress at 100%% should be 100.0, got %f", progress.Percentage)
	}

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

	if group.Clones != nil {
		t.Error("Clones should be nil")
	}
}
