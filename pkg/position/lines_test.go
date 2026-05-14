package position

import (
	"testing"
	"testing/quick"
)

func TestByteRangeToLines(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		start     int
		end       int
		wantStart int
		wantEnd   int
	}{
		{"empty content", "", 0, 0, 1, 1},
		{"single line no newline", "hello world", 0, 5, 1, 1},
		{"single line entire content", "hello world", 0, 11, 1, 1},
		{"two lines first line", "line1\nline2", 0, 5, 1, 1},
		{"two lines spanning newline", "line1\nline2", 3, 8, 1, 2},
		{"two lines second line", "line1\nline2", 6, 11, 2, 2},
		{"three lines middle line", "line1\nline2\nline3", 6, 11, 2, 2},
		{"three lines spanning all", "line1\nline2\nline3", 0, 17, 1, 3},
		{"start at newline", "line1\nline2", 5, 11, 2, 2}, // newline at position 5 increments line to 2 before recording
		{"end beyond content uses last line", "line1\nline2", 6, 100, 2, 2},
		{"position not found defaults to 1", "hello", 100, 200, 1, 1},
		{"same position single point on line 2", "line1\nline2\nline3", 10, 10, 2, 2}, // middle of "line2"
		{"same position at newline", "line1\nline2", 5, 5, 2, 2},         // the newline character
		{"same position on first line", "hello world", 3, 3, 1, 1},
		{"multi-line end beyond content returns last line not start line", "line1\nline2\nline3\nline4\nline5", 6, 100, 2, 5},
		{"multi-line spanning all with end at exact content length", "line1\nline2\nline3", 0, 17, 1, 3},
		{"multi-line end one past content length", "line1\nline2\nline3", 0, 18, 1, 3},
		{"multi-line both start and end beyond content", "line1\nline2\nline3", 100, 200, 3, 3},
		{"negative start clamped", "line1\nline2", -1, 5, 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd := ByteRangeToLines([]byte(tt.content), tt.start, tt.end)
			if gotStart != tt.wantStart || gotEnd != tt.wantEnd {
				t.Errorf("ByteRangeToLines() = (%d, %d), want (%d, %d)",
					gotStart, gotEnd, tt.wantStart, tt.wantEnd)
			}
		})
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "empty content",
			content: "",
			want:    []string{},
		},
		{
			name:    "single line",
			content: "hello",
			want:    []string{"hello"},
		},
		{
			name:    "two lines",
			content: "line1\nline2",
			want:    []string{"line1", "line2"},
		},
		{
			name:    "three lines",
			content: "a\nb\nc",
			want:    []string{"a", "b", "c"},
		},
		{
			name:    "trailing newline",
			content: "line1\n",
			want:    []string{"line1", ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SplitLines([]byte(tt.content))
			if len(got) != len(tt.want) {
				t.Errorf("SplitLines() got %d lines, want %d", len(got), len(tt.want))

				return
			}

			for i, line := range got {
				if line != tt.want[i] {
					t.Errorf("SplitLines()[%d] = %q, want %q", i, line, tt.want[i])
				}
			}
		})
	}
}

func TestJoinLines(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  string
	}{
		{
			name:  "empty lines",
			lines: []string{},
			want:  "",
		},
		{
			name:  "single line",
			lines: []string{"hello"},
			want:  "hello",
		},
		{
			name:  "two lines",
			lines: []string{"line1", "line2"},
			want:  "line1\nline2",
		},
		{
			name:  "three lines",
			lines: []string{"a", "b", "c"},
			want:  "a\nb\nc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := JoinLines(tt.lines)
			if got != tt.want {
				t.Errorf("JoinLines() = %q, want %q", got, tt.want)
			}
		})
	}
}

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

	err := quick.Check(f1, nil)
	if err != nil {
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

	err = quick.Check(f2, nil)
	if err != nil {
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

	err = quick.Check(f3, nil)
	if err != nil {
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

	err := quick.Check(f, nil)
	if err != nil {
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

	err := quick.Check(f1, nil)
	if err != nil {
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

	err = quick.Check(f2, nil)
	if err != nil {
		t.Errorf("Monotonicity property failed: %v", err)
	}
}
