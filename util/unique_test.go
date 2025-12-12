package util

import (
	"testing"

	"github.com/golangci/dupl/syntax"
)

func TestUnique(t *testing.T) {
	// Create test nodes
	node1 := &syntax.Node{Filename: "file1.go", Pos: 10}
	node2 := &syntax.Node{Filename: "file1.go", Pos: 20}
	node3 := &syntax.Node{Filename: "file2.go", Pos: 10}
	node4 := &syntax.Node{Filename: "file1.go", Pos: 10} // duplicate of node1

	// Create sequences
	seq1 := []*syntax.Node{node1}
	seq2 := []*syntax.Node{node2}
	seq3 := []*syntax.Node{node3}
	seq4 := []*syntax.Node{node4} // duplicate location

	group := [][]*syntax.Node{seq1, seq2, seq3, seq4}

	// Test unique function
	result := Unique(group)

	// Should have 3 unique sequences (seq4 is duplicate of seq1)
	if len(result) != 3 {
		t.Errorf("Expected 3 unique sequences, got %d", len(result))
	}

	// Verify no duplicates in result
	positions := make(map[string]map[int]struct{})
	for _, seq := range result {
		node := seq[0]
		if _, exists := positions[node.Filename]; !exists {
			positions[node.Filename] = make(map[int]struct{})
		}
		if _, exists := positions[node.Filename][node.Pos]; exists {
			t.Errorf("Found duplicate at %s:%d", node.Filename, node.Pos)
		}
		positions[node.Filename][node.Pos] = struct{}{}
	}
}

func TestUniqueEmpty(t *testing.T) {
	group := [][]*syntax.Node{}

	result := Unique(group)

	if len(result) != 0 {
		t.Errorf("Expected empty result for empty input, got %d", len(result))
	}
}

func TestUniqueSingle(t *testing.T) {
	node := &syntax.Node{Filename: "file1.go", Pos: 10}
	seq := []*syntax.Node{node}

	group := [][]*syntax.Node{seq}

	result := Unique(group)

	if len(result) != 1 {
		t.Errorf("Expected 1 result for single input, got %d", len(result))
	}

	if result[0][0] != node {
		t.Error("Input node should be returned unchanged")
	}
}
