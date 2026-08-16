//go:build !race

package suffixtree

// Allocation budgets are measured without the race detector: its
// instrumentation inflates allocation counts and would make the budgets
// meaningless.

import (
	"context"
	"math/rand"
	"testing"
)

// allocBudgetTokens generates a deterministic pseudo-random token stream, the
// same distribution class as the benchmark generators.
func allocBudgetTokens(count, values int, seed int64) []Token {
	r := rand.New(rand.NewSource(seed))
	tokens := make([]Token, 0, count)

	for range count {
		tokens = append(tokens, &testToken{val: r.Intn(values)})
	}

	return tokens
}

// TestSTreeUpdateAllocationBudget guards the construction-path allocation
// invariants: leaf states must not allocate (nil transition slice), internal
// states allocate one lazily grown slice instead of a map hash table plus one
// *tran per edge, and states come from the arena. Reintroducing a per-state
// map allocation (~200 extra allocs for this tree) or per-tran pointers blows
// the budget. Measured baseline: 167 allocs for 200 tokens; budget has ~40%
// headroom for data-slice growth variance.
func TestSTreeUpdateAllocationBudget(t *testing.T) {
	tokens := allocBudgetTokens(200, 40, 7)

	allocs := testing.AllocsPerRun(30, func() {
		tree := New()
		mustUpdate(tree, tokens...)
	})

	if allocs > 240 {
		t.Errorf(
			"construction of a 200-token tree allocated %.0f times, budget is 240 — per-state allocations likely reintroduced",
			allocs,
		)
	}
}

// TestFindDuplOverAllocationBudget guards the search-path allocation
// invariants: contextList structs come from the sync.Pool and only []Pos
// position slices and Match results are heap allocated. Measured baseline:
// 2011 allocs for a 2000-token tree; budget has ~40% headroom.
func TestFindDuplOverAllocationBudget(t *testing.T) {
	tree := New()
	mustUpdate(tree, allocBudgetTokens(2000, 500, 9)...)

	allocs := testing.AllocsPerRun(30, func() {
		for range tree.FindDuplOver(context.Background(), 10) {
		}
	})

	if allocs > 2800 {
		t.Errorf(
			"search of a 2000-token tree allocated %.0f times, budget is 2800 — the contextList pool is likely broken",
			allocs,
		)
	}
}
