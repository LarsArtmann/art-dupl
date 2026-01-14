package position

import "testing"

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
