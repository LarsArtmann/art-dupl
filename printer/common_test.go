package printer

import "testing"

func TestCanStripTabs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		block []byte
		start int
		count int
		want  bool
	}{
		{
			name:  "all tabs returns true",
			block: []byte("\t\t\t"),
			start: 0,
			count: 3,
			want:  true,
		},
		{
			name:  "tabs followed by space returns false",
			block: []byte("\t\t x"),
			start: 0,
			count: 3,
			want:  false,
		},
		{
			name:  "count exceeds block length capped",
			block: []byte("\t"),
			start: 0,
			count: 10,
			want:  true,
		},
		{
			name:  "start beyond block length returns true (empty iteration)",
			block: []byte("\t\t"),
			start: 5,
			count: 3,
			want:  true,
		},
		{
			name:  "middle position with tabs returns true",
			block: []byte(" \t\t\t "),
			start: 1,
			count: 3,
			want:  true,
		},
		{
			name:  "middle position with non-tab returns false",
			block: []byte(" \t x\t"),
			start: 1,
			count: 3,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := canStripTabs(tt.block, tt.start, tt.count)
			if got != tt.want {
				t.Errorf("canStripTabs() = %v, want %v", got, tt.want)
			}
		})
	}
}
