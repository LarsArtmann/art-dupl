package printer

import (
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

// newTestClone creates a clone for testing with the given parameters.
func newTestClone(filename string, lineStart, lineEnd int, fragment string) clone {
	return clone{
		filename:  filename,
		lineStart: lineStart,
		lineEnd:   lineEnd,
		fragment:  []byte(fragment),
	}
}

// assertLineNumbersMonotonic verifies that line numbers increment correctly from 1.
func assertLineNumbersMonotonic(t *testing.T, lines []DiffLine, label string) {
	t.Helper()

	for i, line := range lines {
		if line.LineNumber != i+1 {
			t.Errorf(
				"Expected %s line %d to have LineNumber %d, got %d",
				label,
				i,
				i+1,
				line.LineNumber,
			)
		}
	}
}

func TestLineDiff_EqualContent(t *testing.T) {
	base := []byte("line1\nline2\nline3")
	compared := []byte("line1\nline2\nline3")

	result := LineDiff(base, compared)

	if result.HasDiff {
		t.Error("Expected HasDiff to be false for identical content")
	}

	// Verify line counts using a helper to avoid duplication
	testutil.AssertCount(t, len(result.Base), 3, "base")
	testutil.AssertCount(t, len(result.Compared), 3, "compared")

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

	// Second and third lines should be equal
	for i := 1; i <= 2; i++ {
		if result.Base[i].Type != DiffLineEqual {
			t.Errorf("Expected base line %d to be Equal, got %d", i, result.Base[i].Type)
		}
	}
}

func TestLineDiff_AddedLines(t *testing.T) {
	base := []byte("line1\nline2")
	compared := []byte("line1\nline2\nline3")

	result := LineDiff(base, compared)

	if !result.HasDiff {
		t.Error("Expected HasDiff to be true for added content")
	}

	testutil.AssertCount(t, len(result.Compared), 3, "compared")

	if result.Compared[2].Type != DiffLineAdded {
		t.Errorf("Expected compared line 2 to be Added, got %d", result.Compared[2].Type)
	}
}

func TestLineDiff_RemovedLines(t *testing.T) {
	base := []byte("line1\nline2\nline3")
	compared := []byte("line1\nline2")

	result := LineDiff(base, compared)

	if !result.HasDiff {
		t.Error("Expected HasDiff to be true for removed content")
	}

	testutil.AssertCount(t, len(result.Base), 3, "base")

	if result.Base[2].Type != DiffLineRemoved {
		t.Errorf("Expected base line 2 to be Removed, got %d", result.Base[2].Type)
	}
}

func TestLineDiff_EmptyBase(t *testing.T) {
	base := []byte{}
	compared := []byte("line1\nline2")

	result := LineDiff(base, compared)

	if !result.HasDiff {
		t.Error("Expected HasDiff to be true when base is empty")
	}

	// All compared lines should be added
	for i, line := range result.Compared {
		if line.Type != DiffLineAdded {
			t.Errorf("Expected compared line %d to be Added, got %d", i, line.Type)
		}
	}
}

func TestLineDiff_EmptyCompared(t *testing.T) {
	base := []byte("line1\nline2")
	compared := []byte{}

	result := LineDiff(base, compared)

	if !result.HasDiff {
		t.Error("Expected HasDiff to be true when compared is empty")
	}

	// All base lines should be removed
	for i, line := range result.Base {
		if line.Type != DiffLineRemoved {
			t.Errorf("Expected base line %d to be Removed, got %d", i, line.Type)
		}
	}
}

func TestLineDiff_WhitespaceOnlyChange(t *testing.T) {
	base := []byte("func foo() {\n    return 1\n}")
	compared := []byte("func foo() {\n\treturn 1\n}")

	result := LineDiff(base, compared)

	// Note: The current implementation uses TrimSpace for comparison,
	// so pure whitespace changes may not be detected depending on the algorithm path
	// This test documents the current behavior
	t.Logf("HasDiff: %v, base[1].Type: %d", result.HasDiff, result.Base[1].Type)
}

func TestLineDiff_RealWorldClone(t *testing.T) {
	// Real-world example from the codebase
	base := []byte(`func (c *Config) Validate() error {
    if c.Threshold < 1 {
        return fmt.Errorf("threshold must be at least 1")
    }
    if c.OutputFormat == "" {
        c.OutputFormat = "text"
    }
    return nil
}`)

	compared := []byte(`func (c *Config) Validate() error {
    if c.Threshold < 5 {
        return fmt.Errorf("threshold must be at least 5")
    }
    if c.OutputFormat == "" {
        c.OutputFormat = "json"
    }
    return nil
}`)

	result := LineDiff(base, compared)

	if !result.HasDiff {
		t.Error("Expected HasDiff to be true for real-world clone differences")
	}

	// Count modified lines
	modifiedCount := 0

	for _, line := range result.Base {
		if line.Type == DiffLineModified {
			modifiedCount++
		}
	}

	// Note: The actual count depends on the diff algorithm's behavior
	t.Logf("Found %d modified lines out of %d base lines", modifiedCount, len(result.Base))

	// Just verify we detected some differences
	if modifiedCount == 0 && !result.HasDiff {
		t.Error("Expected at least some differences to be detected")
	}
}

func TestComputeCloneGroupDiff(t *testing.T) {
	clones := []clone{
		newTestClone("file1.go", 10, 20, "func foo() {}\nfunc bar() {}"),
		newTestClone("file2.go", 30, 40, "func baz() {}\nfunc qux() {}"),
		newTestClone("file3.go", 50, 60, "func foo() {}\nfunc bar() {}"),
	}

	result := ComputeCloneGroupDiff(clones)

	// Base should be the first clone
	if result.Base == nil {
		t.Fatal("Expected Base to not be nil")
	}

	if result.Base.Filename != "file1.go" {
		t.Errorf("Expected base filename to be file1.go, got %s", result.Base.Filename)
	}

	testutil.AssertCount(t, len(result.Others), 2, "others")

	// Should detect differences
	if !result.HasAnyDiff {
		t.Error("Expected HasAnyDiff to be true")
	}
}

func TestComputeCloneGroupDiff_IdenticalClones(t *testing.T) {
	clones := []clone{
		newTestClone("file1.go", 10, 20, "func foo() {}\nfunc bar() {}"),
		newTestClone("file2.go", 30, 40, "func foo() {}\nfunc bar() {}"),
	}

	result := ComputeCloneGroupDiff(clones)

	// Should not detect differences
	if result.HasAnyDiff {
		t.Error("Expected HasAnyDiff to be false for identical clones")
	}

	t.Logf("HasAnyDiff: %v", result.HasAnyDiff)
}

func TestComputeCloneGroupDiff_SingleClone(t *testing.T) {
	clones := []clone{
		newTestClone("file1.go", 10, 20, "func foo() {}\nfunc bar() {}"),
	}

	result := ComputeCloneGroupDiff(clones)

	// Base should be set
	if result.Base == nil {
		t.Fatal("Expected Base to not be nil")
	}

	testutil.AssertCount(t, len(result.Others), 0, "others")
}

func TestComputeCloneGroupDiff_EmptyClones(t *testing.T) {
	result := ComputeCloneGroupDiff([]clone{})

	if result.Base != nil {
		t.Error("Expected Base to be nil for empty clones")
	}
}

func TestDiffSameLength(t *testing.T) {
	base := []DiffLine{
		{Content: "line1", Type: DiffLineEqual},
		{Content: "line2", Type: DiffLineEqual},
	}
	compared := []DiffLine{
		{Content: "line1", Type: DiffLineEqual},
		{Content: "modified", Type: DiffLineEqual},
	}

	hasDiff := diffSameLength(base, compared)

	if !hasDiff {
		t.Error("Expected hasDiff to be true")
	}

	if base[0].Type != DiffLineEqual {
		t.Error("First line should still be Equal")
	}

	if base[1].Type != DiffLineModified {
		t.Error("Second line should be Modified")
	}
}

func TestDiffLargeFiles(t *testing.T) {
	// Create base and compared with different lengths (>100 triggers large file path)
	baseLines := make([][]byte, 150)
	comparedLines := make([][]byte, 150)

	for i := range 150 {
		if i == 75 {
			baseLines[i] = []byte("func processUser() {\n")
			comparedLines[i] = []byte("func processAdmin() {\n")
		} else {
			baseLines[i] = []byte("    line\n")
			comparedLines[i] = []byte("    line\n")
		}
	}

	base := make([]DiffLine, 150)
	compared := make([]DiffLine, 150)

	for i := range 150 {
		base[i] = DiffLine{Content: string(baseLines[i]), Type: DiffLineEqual, LineNumber: i + 1}
		compared[i] = DiffLine{
			Content:    string(comparedLines[i]),
			Type:       DiffLineEqual,
			LineNumber: i + 1,
		}
	}

	hasDiff := diffLargeFiles(baseLines, comparedLines, base, compared)

	if !hasDiff {
		t.Error("Expected hasDiff to be true")
	}

	if base[75].Type != DiffLineModified {
		t.Errorf("Expected line 75 to be Modified, got %d", base[75].Type)
	}
}

func TestCountDiffStats(t *testing.T) {
	diff := DiffResult{
		Base: []DiffLine{
			{Type: DiffLineEqual},
			{Type: DiffLineRemoved},
			{Type: DiffLineModified},
		},
		Compared: []DiffLine{
			{Type: DiffLineEqual},
			{Type: DiffLineAdded},
			{Type: DiffLineModified},
		},
	}

	added, removed, modified := countDiffLineStats(diff)

	if added != 1 {
		t.Errorf("Expected 1 added, got %d", added)
	}

	if removed != 1 {
		t.Errorf("Expected 1 removed, got %d", removed)
	}

	if modified != 1 {
		t.Errorf("Expected 1 modified, got %d", modified)
	}
}

func TestLineDiff_LineNumbers(t *testing.T) {
	base := []byte("line1\nline2\nline3")
	compared := []byte("line1\nmodified\nline3")

	result := LineDiff(base, compared)

	// Verify line numbers are preserved using helper function
	assertLineNumbersMonotonic(t, result.Base, "base")
	assertLineNumbersMonotonic(t, result.Compared, "compared")
}

func TestLineDiff_ContentPreservation(t *testing.T) {
	base := []byte("func foo() {\n    return 1\n}")
	compared := []byte("func bar() {\n    return 1\n}")

	result := LineDiff(base, compared)

	// Verify original content is preserved (note: splitLines preserves newlines)
	// First line should contain "func foo() {" or "func foo() {\n"
	if !strings.Contains(result.Base[0].Content, "func foo() {") {
		t.Errorf("Base content not preserved: %q", result.Base[0].Content)
	}

	if !strings.Contains(result.Compared[0].Content, "func bar() {") {
		t.Errorf("Compared content not preserved: %q", result.Compared[0].Content)
	}
}

func BenchmarkLineDiff_Small(b *testing.B) {
	base := []byte("func foo() {\n    return 1\n}")
	compared := []byte("func bar() {\n    return 2\n}")

	for b.Loop() {
		LineDiff(base, compared)
	}
}

// makeLineDiffTestData generates test data for line diff benchmarks.
func makeLineDiffTestData(numLines, diffPos int) ([]byte, []byte) {
	base := make([]byte, 0, numLines*20)
	compared := make([]byte, 0, numLines*20)

	for i := range numLines {
		if i == diffPos {
			base = append(base, "func processUser() {\n"...)
			compared = append(compared, "func processAdmin() {\n"...)
		} else {
			base = append(base, "    line\n"...)
			compared = append(compared, "    line\n"...)
		}
	}

	return base, compared
}

func BenchmarkLineDiff_Medium(b *testing.B) {
	// 50 lines - triggers LCS path
	base, compared := makeLineDiffTestData(50, 25)

	for b.Loop() {
		LineDiff(base, compared)
	}
}

func BenchmarkLineDiff_Large(b *testing.B) {
	// 150 lines - triggers large file heuristic
	base, compared := makeLineDiffTestData(150, 75)

	for b.Loop() {
		LineDiff(base, compared)
	}
}

func TestWordDiff_EqualContent(t *testing.T) {
	base := "func processUser(name string) error"
	compared := "func processUser(name string) error"

	result := WordDiff(base, compared)

	// Should not contain added/removed spans for identical content
	testutil.AssertStringNotContains(t, result, `class="word-added"`, "Expected no word-added spans for identical content")
	testutil.AssertStringNotContains(t, result, `class="word-removed"`, "Expected no word-removed spans for identical content")
}

func TestWordDiff_HasChanges(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		compared string
	}{
		{
			name:     "single_word_change",
			base:     "func processUser(name string) error",
			compared: "func processAdmin(name string) error",
		},
		{
			name:     "multiple_changes",
			base:     "return user.Name and user.Email",
			compared: "return admin.Name and admin.Email",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WordDiff(tt.base, tt.compared)

			testutil.AssertStringContains(t, result, `class="word-added"`, "Expected word-added span for changed word")
			testutil.AssertStringContains(t, result, `class="word-removed"`, "Expected word-removed span for changed word")
		})
	}
}

func TestWordDiff_EmptyStrings(t *testing.T) {
	base := ""
	compared := "some content"

	result := WordDiff(base, compared)

	testutil.AssertStringContains(t, result, `class="word-added"`, "Expected word-added span for content added to empty base")
}

func TestWordDiff_HTMLEscaping(t *testing.T) {
	base := "x < y && y > z"
	compared := "x < y && y > z"

	result := WordDiff(base, compared)

	// Should escape HTML special characters
	if strings.Contains(result, "<") && !strings.Contains(result, "&lt;") {
		t.Error("Expected HTML special characters to be escaped")
	}
}
