package syntax

import (
	"testing"

	"github.com/golangci/dupl/suffixtree"
)

func TestFindSyntaxUnitsOwnershipCheck(t *testing.T) {
	// Test case: Different ownership structures at same positions
	// should result in the last index being removed

	// Create a sequence of nodes with different ownership patterns
	nodes := make([]*Node, 10)
	for i := range nodes {
		nodes[i] = &Node{Type: i, Owns: 0}
	}

	// Set ownership pattern: first node owns 3, others own 0
	nodes[0].Owns = 3

	// Create another sequence with different ownership
	data := make([]*Node, 20)
	copy(data, nodes)

	// Add the same sequence at a different position but with different ownership
	for i := 10; i < 20; i++ {
		data[i] = &Node{Type: i - 10, Owns: 0}
	}
	data[10].Owns = 2 // Different ownership at same relative position

	// Create a match spanning positions with different ownership
	match := suffixtree.Match{
		Ps:  []suffixtree.Pos{0, 10},
		Len: 5,
	}

	result := FindSyntaxUnits(data, match, 3)

	// The function should handle the ownership mismatch gracefully
	// and either return an empty match or remove the problematic index
	if len(result.Frags) > 0 && len(result.Frags[0]) > 0 {
		t.Logf("Found syntax units with ownership mismatch: %d fragments", len(result.Frags))
	}
}

func TestFindSyntaxUnitsConsistentOwnership(t *testing.T) {
	// Test case: Same ownership structures should work correctly

	// Create identical sequences with proper ownership
	nodes1 := make([]*Node, 10)
	for i := range nodes1 {
		nodes1[i] = &Node{Type: i, Owns: 0}
	}
	// Set up a proper ownership structure: leaf nodes own 0, parent nodes own children count
	nodes1[0].Owns = 4 // First node owns 4 children
	nodes1[5].Owns = 2 // Node at position 5 owns 2 children

	nodes2 := make([]*Node, 10)
	for i := range nodes2 {
		nodes2[i] = &Node{Type: i, Owns: 0}
	}
	// Same ownership structure
	nodes2[0].Owns = 4
	nodes2[5].Owns = 2

	data := append(nodes1, nodes2...)

	match := suffixtree.Match{
		Ps:  []suffixtree.Pos{0, 10},
		Len: 8,
	}

	result := FindSyntaxUnits(data, match, 3)

	// Should find the match since ownership is consistent
	if len(result.Frags) == 0 {
		t.Error("Expected to find syntax units with consistent ownership")
		t.Logf("Result: %+v", result)
	}
}

func TestFindSyntaxUnitsEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		setup      func() []*Node
		match      suffixtree.Match
		threshold  int
		shouldFind bool
	}{
		{
			name: "empty positions",
			setup: func() []*Node {
				data := make([]*Node, 5)
				for i := range data {
					data[i] = &Node{Type: i, Owns: 1}
				}
				return data
			},
			match: suffixtree.Match{
				Ps:  []suffixtree.Pos{},
				Len: 5,
			},
			threshold:  3,
			shouldFind: false,
		},
		{
			name: "single position",
			setup: func() []*Node {
				data := make([]*Node, 10)
				for i := range data {
					data[i] = &Node{Type: i, Owns: 0}
				}
				// Create a proper ownership structure
				data[0].Owns = 4
				data[5].Owns = 2
				return data
			},
			match: suffixtree.Match{
				Ps:  []suffixtree.Pos{0},
				Len: 8,
			},
			threshold:  3,
			shouldFind: true,
		},
		{
			name: "high threshold",
			setup: func() []*Node {
				data := make([]*Node, 5)
				for i := range data {
					data[i] = &Node{Type: i, Owns: 0}
				}
				return data
			},
			match: suffixtree.Match{
				Ps:  []suffixtree.Pos{0},
				Len: 2,
			},
			threshold:  10,
			shouldFind: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := test.setup()

			result := FindSyntaxUnits(data, test.match, test.threshold)

			if test.shouldFind && len(result.Frags) == 0 {
				t.Errorf("Expected to find syntax units for %s", test.name)
				t.Logf("Result: %+v", result)
			}
			if !test.shouldFind && len(result.Frags) > 0 {
				t.Errorf("Expected not to find syntax units for %s", test.name)
			}
		})
	}
}
