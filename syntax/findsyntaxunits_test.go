package syntax

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/suffixtree"
)

// makeTestNodes creates a slice of nodes with sequential types and specified ownership.
func makeTestNodes(count int, owns int32) []*Node {
	data := make([]*Node, 0, count)
	for i := range count {
		data = append(data, &Node{Type: int32(i), Owns: owns})
	}

	return data
}

func TestFindSyntaxUnitsOwnershipCheck(t *testing.T) {
	// Test case: Different ownership structures at same positions
	// should result in the last index being removed

	// Create a sequence of nodes with different ownership patterns
	nodes := make([]*Node, 0, 10)
	for i := range 10 {
		nodes = append(nodes, &Node{Type: int32(i), Owns: 0})
	}

	// Set ownership pattern: first node owns 3, others own 0
	nodes[0].Owns = 3

	// Create another sequence with different ownership
	data := make([]*Node, 20)
	copy(data, nodes)

	// Add the same sequence at a different position but with different ownership
	for i := 10; i < 20; i++ {
		data[i] = &Node{Type: int32(i - 10), Owns: 0}
	}

	data[10].Owns = 2 // Different ownership at same relative position

	// Create a found spanning positions with different ownership
	found := suffixtree.Match{
		Ps:  []suffixtree.Pos{0, 10},
		Len: 5,
	}

	result := FindSyntaxUnits(data, found, 3)

	// The function should handle the ownership misfound gracefully
	// and either return an empty found or remove the problematic index
	if len(result.Frags) > 0 && len(result.Frags[0]) > 0 {
		t.Logf("Found syntax units with ownership misfound: %d fragments", len(result.Frags))
	}
}

func TestFindSyntaxUnitsConsistentOwnership(t *testing.T) {
	// Test case: Same ownership structures should work correctly

	// Create identical sequences with proper ownership
	nodes1 := make([]*Node, 0, 10)
	for i := range 10 {
		nodes1 = append(nodes1, &Node{Type: int32(i), Owns: 0})
	}
	// Set up a proper ownership structure: leaf nodes own 0, parent nodes own children count
	nodes1[0].Owns = 4 // First node owns 4 children
	nodes1[5].Owns = 2 // Node at position 5 owns 2 children

	nodes2 := make([]*Node, 0, 10)
	for i := range 10 {
		nodes2 = append(nodes2, &Node{Type: int32(i), Owns: 0})
	}
	// Same ownership structure
	nodes2[0].Owns = 4
	nodes2[5].Owns = 2

	data := make([]*Node, 0, len(nodes1)+len(nodes2))
	data = append(data, nodes1...)
	data = append(data, nodes2...)
	_ = data

	searchResult := suffixtree.Match{
		Ps:  []suffixtree.Pos{0, 10},
		Len: 8,
	}

	result := FindSyntaxUnits(data, searchResult, 3)

	// Should find the searchResult since ownership is consistent
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
				return makeTestNodes(5, 1)
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
				data := make([]*Node, 0, 10)
				for i := range 10 {
					data = append(data, &Node{Type: int32(i), Owns: 0})
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
				return makeTestNodes(5, 0)
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

func TestFindSyntaxUnits_MixedCorpusKeepsNonStatementMatches(t *testing.T) {
	// Regression test for: --only templ showed clones that default analysis hid.
	// Statement-level tokenization is Go-only. When Go and templ files are mixed,
	// the non-statement guard must be scoped to each file. A templ file without
	// statement tokens should still produce matches even when other Go files in
	// the corpus contain statement tokens.
	goFile := "main.go"
	templFile := "view.templ"

	// Go file: one statement node, then two identical structural nodes that form
	// a legacy match. The statement token ensures the file uses statement-level
	// tokenization.
	goStmt := &Node{Type: 1000, Filename: goFile, Statement: true, Owns: 0}
	goA := &Node{Type: 2000, Filename: goFile, Statement: false, Owns: 0}
	goB := &Node{Type: 2000, Filename: goFile, Statement: false, Owns: 0}

	// Templ file: two identical non-statement nodes forming a legacy match.
	templA := &Node{Type: 3000, Filename: templFile, Statement: false, Owns: 0}
	templB := &Node{Type: 3000, Filename: templFile, Statement: false, Owns: 0}

	data := []*Node{goStmt, goA, goB, templA, templB}

	// Match the two templ nodes.
	match := suffixtree.Match{
		Ps:  []suffixtree.Pos{3, 4},
		Len: 1,
	}

	result := FindSyntaxUnits(data, match, 1)

	if len(result.Frags) == 0 {
		t.Fatalf("Expected templ match to be kept in mixed corpus, got empty result")
	}

	if len(result.Frags) != 2 {
		t.Fatalf("Expected 2 fragments, got %d", len(result.Frags))
	}

	for i, frag := range result.Frags {
		if len(frag) != 1 {
			t.Fatalf("Expected fragment %d to have 1 node, got %d", i, len(frag))
		}

		if frag[0].Filename != templFile {
			t.Errorf("Expected fragment %d to be from %s, got %s", i, templFile, frag[0].Filename)
		}
	}
}

func TestFindSyntaxUnits_GoNonStatementMatchStillSkipped(t *testing.T) {
	// Verify the per-file guard still skips non-statement structural matches in
	// files that use statement-level tokenization.
	goFile := "main.go"

	goStmt := &Node{Type: 1000, Filename: goFile, Statement: true, Owns: 0}
	goA := &Node{Type: 2000, Filename: goFile, Statement: false, Owns: 0}
	goB := &Node{Type: 2000, Filename: goFile, Statement: false, Owns: 0}

	data := []*Node{goStmt, goA, goB}

	match := suffixtree.Match{
		Ps:  []suffixtree.Pos{1, 2},
		Len: 1,
	}

	result := FindSyntaxUnits(data, match, 1)

	if len(result.Frags) > 0 {
		t.Fatalf("Expected Go non-statement match to be skipped, got %d fragments", len(result.Frags))
	}
}

// mixedParserStream builds a token stream containing:
//   - a templ file fragment of non-statement expression units (owning enough
//     children to clear the legacy threshold), followed by a statement token
//     so the templ file also carries statements,
//   - a Go file with a statement token and a structurally identical
//     non-statement param fragment (the issue #11 shape: one-line func
//     params), and
//   - a second templ fragment identical to the first.
//
// Layout (indices):
//
//	 0..4  templ expression subtree (root Owns=4, then children)
//	 5     templ statement token
//	 6     go statement token
//	 7..11 go param subtree (root Owns=4 — same Types as the templ subtree)
//	12..16 second templ expression subtree (identical to the first)
func mixedParserStream() []*Node {
	templExpr := func() []*Node {
		root := &Node{Filename: "page.templ", Type: 2, Owns: 4, Pos: 100, End: 120}
		for i := int32(1); i <= 4; i++ {
			root.Children = append(root.Children, &Node{Filename: "page.templ", Type: 10 + i, Pos: 100, End: 120})
		}

		// Flatten pre-order like a serialized stream.
		out := make([]*Node, 0, 1+len(root.Children))
		out = append(out, root)
		out = append(out, root.Children...)

		return out
	}

	data := append([]*Node{}, templExpr()...)
	data = append(data, &Node{Filename: "page.templ", Type: 999, Statement: true, Pos: 130, End: 140})
	data = append(data, &Node{Filename: "main.go", Type: 998, Statement: true, Pos: 200, End: 210})

	goParam := templExpr()[0]
	goParam.Filename = "main.go"
	goParam.Pos, goParam.End = 300, 320

	data = append(data, goParam)
	for _, child := range goParam.Children {
		child.Filename = "main.go"
		child.Pos, child.End = 300, 320
		data = append(data, child)
	}

	data = append(data, templExpr()...)

	return data
}

// TestFindSyntaxUnitsMixedParserGoParamsRejected (issue #11): when a match's
// fragments mix parsers, a Go fragment whose unit roots are non-statement
// nodes (func params) inside a statement-bearing Go file invalidates the
// whole match — templ participation must never lower the effective unit bar
// for Go members. The pre-fix code only checked the FIRST fragment, so a
// templ-first ordering smuggled .go params through.
func TestFindSyntaxUnitsMixedParserGoParamsRejected(t *testing.T) {
	data := mixedParserStream()

	// Match the two identical templ expression subtrees (positions 0 and
	// 12) plus the structurally identical Go param subtree (position 7).
	found := suffixtree.Match{
		Ps:  []suffixtree.Pos{0, 7, 12},
		Len: 5,
	}

	result := FindSyntaxUnits(data, found, 3)
	if len(result.Frags) != 0 {
		t.Fatalf(
			"mixed-parser match with non-statement Go param fragment must be rejected, got %d frags",
			len(result.Frags),
		)
	}
}

// TestFindSyntaxUnitsPureTemplExpressionMatchKept: the same expression
// match WITHOUT the Go fragment stays a valid group — templ files may use
// non-statement component/expression units.
func TestFindSyntaxUnitsPureTemplExpressionMatchKept(t *testing.T) {
	data := mixedParserStream()

	found := suffixtree.Match{
		Ps:  []suffixtree.Pos{0, 12},
		Len: 5,
	}

	result := FindSyntaxUnits(data, found, 3)
	if len(result.Frags) == 0 {
		t.Fatal("pure templ expression match must survive the mixed-parser guard")
	}
}

// TestFindSyntaxUnitsGoFirstOrderingAlsoRejected: the Go fragment leads the
// position list — the historical single-fragment guard already caught this;
// pins that the new guard did not relax it.
func TestFindSyntaxUnitsGoFirstOrderingAlsoRejected(t *testing.T) {
	data := mixedParserStream()

	found := suffixtree.Match{
		Ps:  []suffixtree.Pos{7, 0, 12},
		Len: 5,
	}

	result := FindSyntaxUnits(data, found, 3)
	if len(result.Frags) != 0 {
		t.Fatalf("Go-first mixed match with param fragment must be rejected, got %d frags", len(result.Frags))
	}
}
