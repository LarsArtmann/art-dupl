package printer

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

func TestCanStripTabs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		block []byte
		start int
		count int
		want  bool
	}{
		{"all tabs returns true", []byte("\t\t\t"), 0, 3, true},
		{"tabs followed by space returns false", []byte("\t\t x"), 0, 3, false},
		{"count exceeds block length capped", []byte("\t"), 0, 10, true},
		{"start beyond block length returns true (empty iteration)", []byte("\t\t"), 5, 3, true},
		{"middle position with tabs returns true", []byte(" \t\t\t "), 1, 3, true},
		{"middle position with non-tab returns false", []byte(" \t x\t"), 1, 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			actual := canStripTabs(tt.block, tt.start, tt.count)
			testutil.ExpectTrue(t, actual == tt.want, "canStripTabs")
		})
	}
}
