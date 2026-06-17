package detection

import (
	"context"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
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

// createTestConfig creates a test configuration for MultiDetector tests.
func createTestConfig() *config.Config {
	return &config.Config{
		Threshold:        15,
		DetectionMethods: config.DetectionMethods{config.DetectionMethodArtDupl},
	}
}

func TestMultiDetector_logVerbose(t *testing.T) {
	t.Run("verbose enabled", func(t *testing.T) {
		detector := &MultiDetector{detCfg: config.DetectionConfig{Verbose: true}}
		detector.logVerbose("Test message")
	})

	t.Run("verbose disabled", func(t *testing.T) {
		detector := &MultiDetector{detCfg: config.DetectionConfig{Verbose: false}}
		detector.logVerbose("Should not print")
	})
}

func TestMultiDetector_FindDuplOver(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		methods config.DetectionMethods
		verbose bool
		empty   bool
	}{
		{"default method", config.DetectionMethods{config.DetectionMethodArtDupl}, false, false},
		{"hash method", config.DetectionMethods{config.DetectionMethodHash}, false, false},
		{
			"both methods",
			config.DetectionMethods{config.DetectionMethodArtDupl, config.DetectionMethodHash},
			false,
			false,
		},
		{"verbose", config.DetectionMethods{config.DetectionMethodHash}, true, false},
		{"empty data", config.DetectionMethods{config.DetectionMethodArtDupl}, false, true},
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
				config.DetectionConfig{Methods: tc.methods, Verbose: tc.verbose},
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
