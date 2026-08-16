//go:build !race

package syntax

// Allocation budgets are measured without the race detector: its
// instrumentation inflates allocation counts and would make the budgets
// meaningless.

import "testing"

// TestSerializeAllocationBudget guards the bulk-arena serialization layout:
// serial() must emit nodes into one pre-counted []Node arena plus one
// exactly-sized []*Node stream — two allocations total, independent of tree
// size. Reintroducing per-node &Node{} copies (~100 allocs for this tree) or
// append-grown streams blows the budget immediately.
func TestSerializeAllocationBudget(t *testing.T) {
	root := genDeepTree(100)

	allocs := testing.AllocsPerRun(100, func() {
		_ = Serialize(root)
	})

	if allocs > 2 {
		t.Errorf(
			"Serialize(100-node tree) allocated %.0f times, budget is 2 — per-node allocations likely reintroduced",
			allocs,
		)
	}
}

// TestSerializeWithMaxChildrenAllocationBudget covers the caller-supplied
// maxChildren path and the statement (fingerprint) path: statement nodes
// emit exactly one token each, so the arena count and the flat layout hold
// for statement-rooted trees as well.
func TestSerializeWithMaxChildrenAllocationBudget(t *testing.T) {
	root := &Node{
		Type: 1,
		Children: []*Node{
			{Type: 2, Statement: true, Children: []*Node{{Type: 3, Name: "x"}}},
			{Type: 4, Statement: true, Children: []*Node{{Type: 5, Name: "y"}}},
			{Type: 6, Statement: true, Children: []*Node{{Type: 7, Name: "z"}}},
		},
	}

	allocs := testing.AllocsPerRun(100, func() {
		_ = SerializeWithMaxChildren(root, 4)
	})

	if allocs > 2 {
		t.Errorf(
			"SerializeWithMaxChildren(statement tree) allocated %.0f times, budget is 2",
			allocs,
		)
	}
}
