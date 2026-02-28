package job

import (
	"context"
	"testing"
	"time"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// waitForCompletion waits for the done channel or times out after 5 seconds.
// It uses the provided error message if a timeout occurs.
func waitForCompletion(t *testing.T, done chan bool, errorMessage string) {
	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Error(errorMessage)

		return
	}
}

func TestBuildTree(t *testing.T) {
	ctx := t.Context()
	// Create a simple sequence of nodes
	nodes := make([]*syntax.Node, 3)
	for i := range nodes {
		nodes[i] = &syntax.Node{Type: int32(i)}
	}

	// Create channel with test data
	schan := make(chan []*syntax.Node, 1)
	schan <- nodes

	close(schan)

	// Test BuildTree
	tree, data, done := BuildTree(ctx, schan)

	// Wait for processing to complete
	waitForCompletion(t, done, "BuildTree timed out")

	// Verify tree is not nil
	if tree == nil {
		t.Error("Expected non-nil tree")
	}

	// Verify data contains our nodes
	if len(*data) != len(nodes) {
		t.Errorf("Expected %d nodes in data, got %d", len(nodes), len(*data))
	}
}

func TestBuildTreeEmptyInput(t *testing.T) {
	ctx := t.Context()
	// Test with empty channel
	schan := make(chan []*syntax.Node)
	close(schan)

	tree, data, done := BuildTree(ctx, schan)

	// Wait for processing
	waitForCompletion(t, done, "BuildTree with empty input should not time out")

	// Tree should still be created (but empty)
	if tree == nil {
		t.Error("Expected tree even with empty input")
	}

	// Data should be empty
	if len(*data) != 0 {
		t.Errorf("Expected empty data, got %d nodes", len(*data))
	}
}

func TestBuildTreeMultipleSequences(t *testing.T) {
	ctx := t.Context()
	// Create multiple sequences
	sequence1 := make([]*syntax.Node, 2)
	sequence1[0] = &syntax.Node{Type: 1}
	sequence1[1] = &syntax.Node{Type: 2}

	sequence2 := make([]*syntax.Node, 2)
	sequence2[0] = &syntax.Node{Type: 3}
	sequence2[1] = &syntax.Node{Type: 4}

	// Send both sequences
	schan := make(chan []*syntax.Node, 2)
	schan <- sequence1

	schan <- sequence2

	close(schan)

	tree, data, done := BuildTree(ctx, schan)

	// Wait for processing
	waitForCompletion(t, done, "BuildTree with multiple sequences timed out")

	// Should contain all nodes from both sequences
	if len(*data) != 4 {
		t.Errorf("Expected 4 nodes total, got %d", len(*data))
	}

	// We don't need to use tree variable, but verify it exists
	if tree == nil {
		t.Error("Expected non-nil tree")
	}
}
