package position

import (
	"testing"

	"testing/quick"
)

func TestLineIndex(t *testing.T) {
	content := []byte("Line 1\nLine 2\nLine 3")
	idx := NewLineIndex(content)

	tests := []struct {
		offset int
		want   int
	}{
		{0, 1},
		{1, 1},
		{6, 1}, // End of Line 1
		{7, 2}, // Start of Line 2
		{13, 2},
		{14, 3},
		{100, 3}, // Out of bounds
	}

	for _, tt := range tests {
		got := idx.Line(tt.offset)
		if got != tt.want {
			t.Errorf("Line(%d) = %d, want %d", tt.offset, got, tt.want)
		}
	}
}

// Property-based tests for position calculation functions

func TestByteRangeToLinesProperty(t *testing.T) {
	t.Parallel()

	// Property 1: Line numbers are always positive
	f1 := func(content string, start, end int) bool {
		if start < 0 || end < 0 || start > end {
			return true // Invalid input, skip
		}
		if end > len(content) {
			return true // End beyond content, skip
		}

		startLine, endLine := ByteRangeToLines([]byte(content), start, end)
		return startLine > 0 && endLine > 0
	}
	if err := quick.Check(f1, nil); err != nil {
		t.Errorf("Positive line numbers property failed: %v", err)
	}

	// Property 2: End line >= start line
	f2 := func(content string, start, end int) bool {
		if start < 0 || end < 0 || start > end {
			return true // Invalid input, skip
		}
		if end > len(content) {
			return true // End beyond content, skip
		}

		startLine, endLine := ByteRangeToLines([]byte(content), start, end)
		return endLine >= startLine
	}
	if err := quick.Check(f2, nil); err != nil {
		t.Errorf("End line >= start line property failed: %v", err)
	}

	// Property 3: Empty content returns 1,1
	f3 := func(start, end int) bool {
		if start != 0 || end != 0 {
			return true // Non-zero offset, skip
		}

		startLine, endLine := ByteRangeToLines([]byte(""), start, end)
		return startLine == 1 && endLine == 1
	}
	if err := quick.Check(f3, nil); err != nil {
		t.Errorf("Empty content property failed: %v", err)
	}
}

func TestSplitLinesProperty(t *testing.T) {
	t.Parallel()

	// Property: Splitting then joining returns original content
	f := func(content string) bool {
		lines := SplitLines([]byte(content))
		joined := JoinLines(lines)
		return joined == content
	}
	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Split-join roundtrip property failed: %v", err)
	}
}

func TestLineIndexProperty(t *testing.T) {
	t.Parallel()

	// Property 1: Line numbers are always positive
	f1 := func(content string, offset int) bool {
		if offset < 0 {
			return true // Negative offset, skip
		}
		if offset > len(content) {
			return true // Offset beyond content, skip
		}

		idx := NewLineIndex([]byte(content))
		lineNum := idx.Line(offset)
		return lineNum >= 0
	}
	if err := quick.Check(f1, nil); err != nil {
		t.Errorf("Positive line numbers property failed: %v", err)
	}

	// Property 2: Monotonicity - larger offsets give >= line numbers
	f2 := func(content string, offset1, offset2 int) bool {
		if offset1 < 0 || offset2 < 0 {
			return true // Negative offsets, skip
		}
		if offset1 > offset2 {
			offset1, offset2 = offset2, offset1 // Ensure offset1 <= offset2
		}
		if offset2 > len(content) {
			return true // Offset beyond content, skip
		}

		idx := NewLineIndex([]byte(content))
		line1 := idx.Line(offset1)
		line2 := idx.Line(offset2)
		return line2 >= line1
	}
	if err := quick.Check(f2, nil); err != nil {
		t.Errorf("Monotonicity property failed: %v", err)
	}
}
