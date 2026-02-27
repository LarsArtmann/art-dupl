package detection

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// TestNewMultiDetector_Working tests MultiDetector constructor.
func TestNewMultiDetector_Working(t *testing.T) {
	cfg := createTestConfig()

	data := []*syntax.Node{
		{Filename: "test.go", Type: 1}, // Use integer type
	}

	tree := suffixtree.New()

	detector := NewMultiDetector(cfg, data, tree, true)

	if detector.config != cfg {
		t.Error("Config not set correctly")
	}

	if len(detector.data) != len(data) {
		t.Error("Data not set correctly")
	}

	if detector.tree != tree {
		t.Error("Tree not set correctly")
	}

	if detector.verbose != true {
		t.Error("Verbose flag not set correctly")
	}
}

// TestTodoDetector_Working tests TODO detector creation.
func TestTodoDetector_Working(t *testing.T) {
	detector := NewTodoDetector()

	if detector.patterns == nil {
		t.Error("Patterns should not be nil")
	}
}

// TestLegacyDetector_Working tests legacy detector creation.
func TestLegacyDetector_Working(t *testing.T) {
	detector := NewLegacyDetector()

	if detector.patterns == nil {
		t.Error("Patterns should not be nil")
	}
}

// TestMultiDetector_logVerbose_Working tests verbose logging.
func TestMultiDetector_logVerbose_Working(t *testing.T) {
	detector := &MultiDetector{verbose: true}

	// This should not panic
	detector.logVerbose("Test message")

	detector = &MultiDetector{verbose: false}

	// This should not output anything (no panic test needed)
	detector.logVerbose("Should not print")
}
