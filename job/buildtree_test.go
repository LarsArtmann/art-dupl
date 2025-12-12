package job

import (
	"testing"
	"time"

	"github.com/golangci/dupl/syntax"
)

func TestBuildTree(t *testing.T) {
	// Create a simple sequence of nodes
	nodes := make([]*syntax.Node, 3)
	for i := range nodes {
		nodes[i] = &syntax.Node{Type: i}
	}

	// Create channel with test data
	schan := make(chan []*syntax.Node, 1)
	schan <- nodes
	close(schan)

	// Test BuildTree
	tree, data, done := BuildTree(schan)

	// Wait for processing to complete
	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Error("BuildTree timed out")
		return
	}

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
	// Test with empty channel
	schan := make(chan []*syntax.Node)
	close(schan)

	tree, data, done := BuildTree(schan)

	// Wait for processing
	select {
	case <-done:
		// Should complete successfully
	case <-time.After(5 * time.Second):
		t.Error("BuildTree with empty input should not time out")
		return
	}

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

	tree, data, done := BuildTree(schan)

	// Wait for processing
	select {
	case <-done:
		// Success
	case <-time.After(5 * time.Second):
		t.Error("BuildTree with multiple sequences timed out")
		return
	}

	// Should contain all nodes from both sequences
	if len(*data) != 4 {
		t.Errorf("Expected 4 nodes total, got %d", len(*data))
	}

	// We don't need to use tree variable, but verify it exists
	if tree == nil {
		t.Error("Expected non-nil tree")
	}
}
