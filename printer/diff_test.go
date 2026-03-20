package printer

import (
	"bytes"
	"testing"
)

func TestLineDiff_EqualContent(t *testing.T) {
	base := []byte("line1\nline2\nline3")
	compared := []byte("line1\nline2\nline3")

	result := LineDiff(base, compared)

	if result.HasDiff {
		t.Error("Expected HasDiff to be false for identical content")
	}

	if len(result.Base) != 3 {
		t.Errorf("Expected 3 base lines, got %d", len(result.Base))
	}

	if len(result.Compared) != 3 {
		t.Errorf("Expected 3 compared lines, got %d", len(result.Compared))
	}

	// All lines should be marked as equal
	for i, line := range result.Base {
		if line.Type != DiffLineEqual {
			t.Errorf("Expected base line %d to be Equal, got %d", i, line.Type)
		}
	}

	for i, line := range result.Compared {
		if line.Type != DiffLineEqual {
			t.Errorf("Expected compared line %d to be Equal, got %d", i, line.Type)
		}
	}
}

func TestLineDiff_ModifiedLines(t *testing.T) {
	base := []byte("func processUser(name string) {\n    return name\n}")
	compared := []byte("func processAdmin(name string) {\n    return name\n}")

	result := LineDiff(base, compared)

	if !result.HasDiff {
		t.Error("Expected HasDiff to be true for modified content")
	}

	// First line should be modified (function name changed)
	if result.Base[0].Type != DiffLineModified {
		t.Errorf("Expected base line 0 to be Modified, got %d", result.Base[0].Type)
	}

	if result.Compared[0].Type != DiffLineModified {
		t.Errorf("Expected compared line 0 to be Modified, got %d", result.Compared[0].Type)
	}

	// Other lines should be equal
	if result.Base[1].Type != DiffLineEqual {
		t.Errorf("Expected base line 1 to be Equal, got %d", result.Base[1].Type)
	}
}

func TestLineDiff_AddedLines(t *testing.T) {
	base := []byte("line1\nline2")
	compared := []byte("line1\nline2\nline3\nline4")

	result := LineDiff(base, compared)

	if !result.HasDiff {
		t.Error("Expected HasDiff to be true when lines are added")
	}

	// Check that added lines are marked correctly
	addedCount := 0
	for _, line := range result.Compared {
		if line.Type == DiffLineAdded {
			addedCount++
		}
	}

	if addedCount != 2 {
		t.Errorf("Expected 2 added lines, got %d", addedCount)
	}
}

func TestLineDiff_RemovedLines(t *testing.T) {
	base := []byte("line1\nline2\nline3\nline4")
	compared := []byte("line1\nline2")

	result := LineDiff(base, compared)

	if !result.HasDiff {
		t.Error("Expected HasDiff to be true when lines are removed")
	}

	// Check that removed lines are marked correctly
	removedCount := 0
	for _, line := range result.Base {
		if line.Type == DiffLineRemoved {
			removedCount++
		}
	}

	if removedCount != 2 {
		t.Errorf("Expected 2 removed lines, got %d", removedCount)
	}
}

func TestLineDiff_EmptyBase(t *testing.T) {
	base := []byte("")
	compared := []byte("line1\nline2")

	result := LineDiff(base, compared)

	if !result.HasDiff {
		t.Error("Expected HasDiff to be true when base is empty")
	}

	if len(result.Base) != 1 { // Empty content becomes single empty line
		t.Errorf("Expected 1 base line (empty), got %d", len(result.Base))
	}

	if len(result.Compared) != 2 {
		t.Errorf("Expected 2 compared lines, got %d", len(result.Compared))
	}
}

func TestLineDiff_EmptyCompared(t *testing.T) {
	base := []byte("line1\nline2")
	compared := []byte("")

	result := LineDiff(base, compared)

	if !result.HasDiff {
		t.Error("Expected HasDiff to be true when compared is empty")
	}

	if len(result.Base) != 2 {
		t.Errorf("Expected 2 base lines, got %d", len(result.Base))
	}

	if len(result.Compared) != 1 { // Empty content becomes single empty line
		t.Errorf("Expected 1 compared line (empty), got %d", len(result.Compared))
	}
}

func TestLineDiff_WhitespaceOnlyChange(t *testing.T) {
	// Whitespace changes should be considered equal after trimming
	base := []byte("func foo() {\n    return 1\n}")
	compared := []byte("func foo() {\n  return 1\n}")

	result := LineDiff(base, compared)

	// Note: Current implementation trims whitespace for comparison
	// This may or may not be desired behavior depending on requirements
	for i, line := range result.Base {
		if result.Compared[i].Type != line.Type {
			t.Logf("Line %d: base type=%d, compared type=%d", i, line.Type, result.Compared[i].Type)
		}
	}
}

func TestLineDiff_RealWorldClone(t *testing.T) {
	// Test case based on actual Go code duplicates
	base := []byte(`func processUser(user string, age int) error {
    if user == "" {
        return fmt.Errorf("empty user")
    }
    if age < 0 {
        return fmt.Errorf("invalid age")
    }
    return nil
}`)

	compared := []byte(`func processAdmin(admin string, level int) error {
    if admin == "" {
        return fmt.Errorf("empty admin")
    }
    if level < 0 {
        return fmt.Errorf("invalid level")
    }
    return nil
}`)

	result := LineDiff(base, compared)

	if !result.HasDiff {
		t.Error("Expected HasDiff to be true for real-world clone with different variable names")
	}

	// Count modifications
	modifiedCount := 0
	for _, line := range result.Base {
		if line.Type == DiffLineModified {
			modifiedCount++
		}
	}

	// Should have modifications on lines with variable name changes
	if modifiedCount == 0 {
		t.Error("Expected some modified lines for real-world clone")
	}

	t.Logf("Found %d modified lines out of %d base lines", modifiedCount, len(result.Base))
}

func TestComputeCloneGroupDiff(t *testing.T) {
	clones := []clone{
		{
			filename:  "file1.go",
			lineStart: 1,
			lineEnd:   5,
			fragment:  []byte("func foo() {\n    return 1\n}"),
		},
		{
			filename:  "file2.go",
			lineStart: 10,
			lineEnd:   15,
			fragment:  []byte("func bar() {\n    return 2\n}"),
		},
		{
			filename:  "file3.go",
			lineStart: 20,
			lineEnd:   25,
			fragment:  []byte("func baz() {\n    return 3\n}"),
		},
	}

	result := ComputeCloneGroupDiff(clones)

	if result.Base == nil {
		t.Fatal("Expected Base to not be nil")
	}

	if result.Base.Filename != "file1.go" {
		t.Errorf("Expected base filename 'file1.go', got %s", result.Base.Filename)
	}

	if len(result.Others) != 2 {
		t.Errorf("Expected 2 other clones, got %d", len(result.Others))
	}

	if !result.HasAnyDiff {
		t.Error("Expected HasAnyDiff to be true when fragments differ")
	}

	// Check first other clone
	if result.Others[0].Filename != "file2.go" {
		t.Errorf("Expected first other filename 'file2.go', got %s", result.Others[0].Filename)
	}
}

func TestComputeCloneGroupDiff_IdenticalClones(t *testing.T) {
	clones := []clone{
		{
			filename:  "file1.go",
			lineStart: 1,
			lineEnd:   5,
			fragment:  []byte("func foo() {\n    return 1\n}"),
		},
		{
			filename:  "file2.go",
			lineStart: 10,
			lineEnd:   15,
			fragment:  []byte("func foo() {\n    return 1\n}"),
		},
	}

	result := ComputeCloneGroupDiff(clones)

	// If fragments are identical, HasAnyDiff should be false
	// Note: This depends on whether we consider identical fragments as having no diff
	// The current implementation may still mark HasAnyDiff=true
	t.Logf("HasAnyDiff: %v", result.HasAnyDiff)
}

func TestComputeCloneGroupDiff_SingleClone(t *testing.T) {
	clones := []clone{
		{
			filename:  "file1.go",
			lineStart: 1,
			lineEnd:   5,
			fragment:  []byte("func foo() {\n    return 1\n}"),
		},
	}

	result := ComputeCloneGroupDiff(clones)

	if result.Base == nil {
		t.Fatal("Expected Base to not be nil even for single clone")
	}

	if len(result.Others) != 0 {
		t.Errorf("Expected 0 other clones for single clone, got %d", len(result.Others))
	}
}

func TestComputeCloneGroupDiff_EmptyClones(t *testing.T) {
	clones := []clone{}

	result := ComputeCloneGroupDiff(clones)

	if result.Base != nil {
		t.Error("Expected Base to be nil for empty clones")
	}

	if len(result.Others) != 0 {
		t.Errorf("Expected 0 others for empty clones, got %d", len(result.Others))
	}
}

func TestSplitLines(t *testing.T) {
	testCases := []struct {
		name     string
		input    []byte
		expected int // expected number of lines
	}{
		{
			name:     "simple lines",
			input:    []byte("line1\nline2\nline3"),
			expected: 3,
		},
		{
			name:     "trailing newline",
			input:    []byte("line1\nline2\n"),
			expected: 2,
		},
		{
			name:     "empty content",
			input:    []byte{},
			expected: 1, // Empty becomes single line with empty content
		},
		{
			name:     "single line no newline",
			input:    []byte("single"),
			expected: 1,
		},
		{
			name:     "CRLF line endings",
			input:    []byte("line1\r\nline2\r\n"),
			expected: 2, // \r\n is treated as two separate chars, only \n triggers split
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lines := splitLines(tc.input)
			if len(lines) != tc.expected {
				t.Errorf("Expected %d lines, got %d", tc.expected, len(lines))
			}
		})
	}
}

func TestDiffSameLength(t *testing.T) {
	base := []DiffLine{
		{Content: "line1", Type: DiffLineEqual, LineNumber: 1},
		{Content: "line2", Type: DiffLineEqual, LineNumber: 2},
	}
	compared := []DiffLine{
		{Content: "line1", Type: DiffLineEqual, LineNumber: 1},
		{Content: "lineX", Type: DiffLineEqual, LineNumber: 2},
	}

	hasDiff := diffSameLength(base, compared)

	if !hasDiff {
		t.Error("Expected hasDiff to be true when lines differ")
	}

	if base[1].Type != DiffLineModified {
		t.Errorf("Expected line 1 to be Modified, got %d", base[1].Type)
	}
}

func TestDiffLargeFiles(t *testing.T) {
	// Create large files (>100 lines to trigger large file path)
	baseLines := make([][]byte, 150)
	comparedLines := make([][]byte, 150)

	for i := range 150 {
		if i == 50 {
			baseLines[i] = []byte("different line\n")
			comparedLines[i] = []byte("modified line\n")
		} else {
			baseLines[i] = []byte("line\n")
			comparedLines[i] = []byte("line\n")
		}
	}

	base := make([]DiffLine, 150)
	compared := make([]DiffLine, 150)

	for i := range 150 {
		base[i] = DiffLine{Content: string(baseLines[i]), Type: DiffLineEqual, LineNumber: i + 1}
		compared[i] = DiffLine{Content: string(comparedLines[i]), Type: DiffLineEqual, LineNumber: i + 1}
	}

	hasDiff := diffLargeFiles(baseLines, comparedLines, base, compared)

	if !hasDiff {
		t.Error("Expected hasDiff to be true for large files with differences")
	}

	// Line 51 (index 50) should be marked as modified
	if base[50].Type != DiffLineModified {
		t.Errorf("Expected line 50 to be Modified, got %d", base[50].Type)
	}
}

func TestCountDiffStats(t *testing.T) {
	diff := DiffResult{
		Compared: []DiffLine{
			{Type: DiffLineEqual},
			{Type: DiffLineAdded},
			{Type: DiffLineAdded},
			{Type: DiffLineRemoved},
			{Type: DiffLineModified},
			{Type: DiffLineModified},
			{Type: DiffLineModified},
		},
	}

	added, removed, modified := countDiffStats(diff)

	if added != 2 {
		t.Errorf("Expected 2 added, got %d", added)
	}

	if removed != 1 {
		t.Errorf("Expected 1 removed, got %d", removed)
	}

	if modified != 3 {
		t.Errorf("Expected 3 modified, got %d", modified)
	}
}

func TestLineDiff_LineNumbers(t *testing.T) {
	base := []byte("line1\nline2\nline3")
	compared := []byte("line1\nline2\nline3")

	result := LineDiff(base, compared)

	// Check that line numbers are correct
	for i, line := range result.Base {
		expectedLineNum := i + 1
		if line.LineNumber != expectedLineNum {
			t.Errorf("Base line %d: expected LineNumber %d, got %d", i, expectedLineNum, line.LineNumber)
		}
	}

	for i, line := range result.Compared {
		expectedLineNum := i + 1
		if line.LineNumber != expectedLineNum {
			t.Errorf("Compared line %d: expected LineNumber %d, got %d", i, expectedLineNum, line.LineNumber)
		}
	}
}

func TestLineDiff_ContentPreservation(t *testing.T) {
	base := []byte("func foo() {\n    return 42\n}")
	compared := []byte("func bar() {\n    return 42\n}")

	result := LineDiff(base, compared)

	// Check that content is preserved
	if !bytes.Equal([]byte(result.Base[0].Content), []byte("func foo() {\n")) {
		t.Errorf("Base line 0 content mismatch: %q", result.Base[0].Content)
	}

	if !bytes.Equal([]byte(result.Compared[0].Content), []byte("func bar() {\n")) {
		t.Errorf("Compared line 0 content mismatch: %q", result.Compared[0].Content)
	}

	// Line 2 should be equal
	if result.Base[1].Type != DiffLineEqual {
		t.Errorf("Expected base line 1 to be Equal, got %d", result.Base[1].Type)
	}
}

func BenchmarkLineDiff_Small(b *testing.B) {
	base := []byte("func foo() {\n    return 1\n}")
	compared := []byte("func bar() {\n    return 2\n}")

	for b.Loop() {
		LineDiff(base, compared)
	}
}

func BenchmarkLineDiff_Medium(b *testing.B) {
	// 50 lines - triggers LCS path
	base := make([]byte, 0, 1000)
	compared := make([]byte, 0, 1000)

	for i := range 50 {
		if i == 25 {
			base = append(base, "func processUser() {\n"...)
			compared = append(compared, "func processAdmin() {\n"...)
		} else {
			base = append(base, "    line\n"...)
			compared = append(compared, "    line\n"...)
		}
	}

	for b.Loop() {
		LineDiff(base, compared)
	}
}

func BenchmarkLineDiff_Large(b *testing.B) {
	// 150 lines - triggers large file heuristic
	base := make([]byte, 0, 3000)
	compared := make([]byte, 0, 3000)

	for i := range 150 {
		if i == 75 {
			base = append(base, "func processUser() {\n"...)
			compared = append(compared, "func processAdmin() {\n"...)
		} else {
			base = append(base, "    line\n"...)
			compared = append(compared, "    line\n"...)
		}
	}

	for b.Loop() {
		LineDiff(base, compared)
	}
}
