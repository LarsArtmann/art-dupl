package detection

import (
	"context"
	"testing"

	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func createTestNode(filename string, pos, end int32) *syntax.Node {
	return &syntax.Node{
		Filename: filename,
		Type:     int32(golang.File),
		Pos:      pos,
		End:      end,
	}
}

func TestMultiDetector_logVerbose(t *testing.T) {
	t.Run("verbose enabled", func(t *testing.T) {
		detector := &MultiDetector{cfg: Config{Verbose: true}}
		detector.logVerbose("Test message")
	})

	t.Run("verbose disabled", func(t *testing.T) {
		detector := &MultiDetector{cfg: Config{Verbose: false}}
		detector.logVerbose("Should not print")
	})
}

func TestMultiDetector_FindDuplOver(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		methods []string
		verbose bool
		empty   bool
	}{
		{"default method", []string{MethodArtDupl}, false, false},
		{"hash method", []string{MethodHash}, false, false},
		{
			"both methods",
			[]string{MethodArtDupl, MethodHash},
			false,
			false,
		},
		{"verbose", []string{MethodHash}, true, false},
		{"empty data", []string{MethodArtDupl}, false, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tree := suffixtree.New()

			var data []*syntax.Node
			if !tc.empty {
				data = []*syntax.Node{createTestNode("test.go", 1, 10)}
			}

			detector := NewMultiDetector(
				Config{Methods: tc.methods, Verbose: tc.verbose},
				data,
				tree,
			)

			matches := detector.FindDuplOver(context.Background(), 15)
			if matches == nil {
				t.Error("FindDuplOver() returned nil channel")
			}

			for range matches {
			}
		})
	}
}
