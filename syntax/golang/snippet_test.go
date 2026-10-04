package golang

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
)

func mustSnippet(t *testing.T, kind SnippetKind, src, filename string, offset int, mode DetectionMode) []*syntax.Node {
	t.Helper()

	nodes, err := ParseSnippet(kind, src, filename, offset, mode)
	if err != nil {
		t.Fatalf("ParseSnippet(%s, %q) error = %v", snippetKindName(kind), src, err)
	}

	return nodes
}

// flattenVals returns the Val() of every node in the subtree, pre-order.
func flattenVals(n *syntax.Node) []suffixtree.TokenValue {
	if n == nil {
		return nil
	}

	vals := []suffixtree.TokenValue{n.Val()}

	for _, child := range n.Children {
		vals = append(vals, flattenVals(child)...)
	}

	return vals
}

func tokenSeqsEqual(a, b []*syntax.Node) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		x := flattenVals(a[i])
		y := flattenVals(b[i])

		if len(x) != len(y) {
			return false
		}

		for j := range x {
			if x[j] != y[j] {
				return false
			}
		}
	}

	return true
}

func TestParseSnippetExpressionNormalizesLocals(t *testing.T) {
	mode := DetectionModeSemantic

	a := mustSnippet(t, SnippetExpression, "user.Name", "a.templ", 100, mode)
	b := mustSnippet(t, SnippetExpression, "person.Name", "b.templ", 100, mode)

	if !tokenSeqsEqual(a, b) {
		t.Errorf("renamed local in expression should match: %v vs %v", flattenVals(a[0]), flattenVals(b[0]))
	}

	c := mustSnippet(t, SnippetExpression, "user.Email", "c.templ", 100, mode)
	if tokenSeqsEqual(a, c) {
		t.Error("different field selectors must not match (API surface)")
	}
}

func TestParseSnippetExpressionKeepsCalleeNames(t *testing.T) {
	a := mustSnippet(t, SnippetExpression, "len(items)", "a.templ", 10, DetectionModeSemantic)
	b := mustSnippet(t, SnippetExpression, "len(entries)", "b.templ", 10, DetectionModeSemantic)

	if !tokenSeqsEqual(a, b) {
		t.Errorf("len() builtin must keep its name: %v vs %v", flattenVals(a[0]), flattenVals(b[0]))
	}

	c := mustSnippet(t, SnippetExpression, "count(items)", "c.templ", 10, DetectionModeSemantic)
	if tokenSeqsEqual(a, c) {
		t.Error("different called functions must not match")
	}
}

func TestParseSnippetExpressionStructuralModeDropsNames(t *testing.T) {
	a := mustSnippet(t, SnippetExpression, "user.Name", "a.templ", 10, DetectionModeStructural)
	b := mustSnippet(t, SnippetExpression, "other.Thing", "b.templ", 10, DetectionModeStructural)

	if !tokenSeqsEqual(a, b) {
		t.Errorf("structural mode must drop identifier names: %v vs %v", flattenVals(a[0]), flattenVals(b[0]))
	}
}

func TestParseSnippetExpressionExactModeKeepsVerbatimNames(t *testing.T) {
	a := mustSnippet(t, SnippetExpression, "user.Name", "a.templ", 10, DetectionModeExact)
	b := mustSnippet(t, SnippetExpression, "person.Name", "b.templ", 10, DetectionModeExact)

	if tokenSeqsEqual(a, b) {
		t.Error("exact mode must keep identifier names verbatim (no normalization)")
	}
}

func TestParseSnippetIfCond(t *testing.T) {
	a := mustSnippet(t, SnippetIfCond, "user.IsAdmin && user.Active", "a.templ", 200, DetectionModeSemantic)
	b := mustSnippet(t, SnippetIfCond, "person.IsAdmin && person.Active", "b.templ", 250, DetectionModeSemantic)

	if !tokenSeqsEqual(a, b) {
		t.Errorf("renamed locals in if-condition should match: %v vs %v", flattenVals(a[0]), flattenVals(b[0]))
	}

	c := mustSnippet(t, SnippetIfCond, "user.IsGuest && user.Active", "c.templ", 200, DetectionModeSemantic)
	if tokenSeqsEqual(a, c) {
		t.Error("different conditions must not match")
	}
}

func TestParseSnippetForClause(t *testing.T) {
	a := mustSnippet(t, SnippetForClause, "i, item := range items", "a.templ", 300, DetectionModeSemantic)
	b := mustSnippet(t, SnippetForClause, "j, product := range products", "b.templ", 300, DetectionModeSemantic)

	if !tokenSeqsEqual(a, b) {
		t.Errorf("renamed range loop locals should match: %v vs %v", flattenVals(a[0]), flattenVals(b[0]))
	}
}

func TestParseSnippetCaseList(t *testing.T) {
	a := mustSnippet(t, SnippetCaseList, `"admin", "root"`, "a.templ", 400, DetectionModeSemantic)
	b := mustSnippet(t, SnippetCaseList, `"guest"`, "b.templ", 400, DetectionModeSemantic)

	if tokenSeqsEqual(a, b) {
		t.Error("case lists with different cardinality must not match")
	}

	c := mustSnippet(t, SnippetCaseList, `"one", "two"`, "c.templ", 400, DetectionModeSemantic)
	if !tokenSeqsEqual(a, c) {
		t.Errorf("string literals normalize to kind in semantic mode: %v vs %v", flattenVals(a[0]), flattenVals(c[0]))
	}
}

func TestParseSnippetStatementsGoCode(t *testing.T) {
	src := "x := user.Count; if x > 3 { render(x) }"

	a := mustSnippet(t, SnippetStatements, src, "a.templ", 500, DetectionModeSemantic)
	b := mustSnippet(t, SnippetStatements, "y := person.Count; if y > 3 { render(y) }", "b.templ", 500, DetectionModeSemantic)

	if len(a) != len(b) {
		t.Fatalf("statement snippets should produce matching statement counts: %d vs %d", len(a), len(b))
	}

	for i := range a {
		x := flattenVals(a[i])
		y := flattenVals(b[i])

		if len(x) != len(y) {
			t.Fatalf("statement %d token count mismatch after renaming: %v vs %v", i, x, y)
		}

		for j := range x {
			if x[j] != y[j] {
				t.Errorf("statement %d token %d should match after renaming: %d vs %d", i, j, x[j], y[j])
			}
		}

		if !a[i].Statement {
			t.Errorf("GoCode statement %d must be marked Statement=true for unit selection", i)
		}
	}
}

func TestParseSnippetPositionsAddressRealFile(t *testing.T) {
	nodes := mustSnippet(t, SnippetIfCond, "user.IsAdmin", "real.templ", 1000, DetectionModeSemantic)

	if len(nodes) != 1 {
		t.Fatalf("expected 1 node, got %d", len(nodes))
	}

	if nodes[0].Pos != 1000 {
		t.Errorf("root node Pos should equal real file offset: got %d, want 1000", nodes[0].Pos)
	}

	if nodes[0].End <= nodes[0].Pos {
		t.Errorf("node End (%d) must be greater than Pos (%d)", nodes[0].End, nodes[0].Pos)
	}
}

func TestParseSnippetStatementsPositionsAddressRealFile(t *testing.T) {
	nodes := mustSnippet(t, SnippetStatements, "x := user.Count", "real.templ", 700, DetectionModeSemantic)

	for i, n := range nodes {
		if n.Pos < 700 {
			t.Errorf("statement %d Pos %d must not fall before the real snippet offset 700", i, n.Pos)
		}
	}
}

func TestParseSnippetEmpty(t *testing.T) {
	nodes, err := ParseSnippet(SnippetExpression, "  ", "a.templ", 0, DetectionModeSemantic)
	if err != nil {
		t.Fatalf("empty snippet should not error: %v", err)
	}

	if len(nodes) != 0 {
		t.Errorf("empty snippet should return no nodes, got %d", len(nodes))
	}
}

func TestParseSnippetInvalidGoErrors(t *testing.T) {
	_, err := ParseSnippet(SnippetExpression, "user.(", "a.templ", 0, DetectionModeSemantic)
	if err == nil {
		t.Fatal("invalid Go expression should return an error for the caller's fallback")
	}
}

func TestParseSnippetGoFileSkipsImports(t *testing.T) {
	src := "import \"strings\"\n\nfunc helper(x string) string {\n\treturn strings.ToUpper(x)\n}\n"

	nodes := mustSnippet(t, SnippetGoFile, src, "a.templ", 0, DetectionModeSemantic)
	if len(nodes) != 1 {
		t.Fatalf("expected only the non-import declaration, got %d nodes", len(nodes))
	}
}
