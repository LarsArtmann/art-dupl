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

	altSrc := "y := person.Count; if y > 3 { render(y) }"

	a := mustSnippet(t, SnippetStatements, src, "a.templ", 500, DetectionModeSemantic)
	b := mustSnippet(t, SnippetStatements, altSrc, "b.templ", 500, DetectionModeSemantic)

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
	const src = "import \"strings\"\n\nfunc helper(x string) string {\n\treturn strings.ToUpper(x)\n}\n"

	nodes := mustSnippet(t, SnippetGoFile, src, "a.templ", 0, DetectionModeSemantic)
	if len(nodes) != 1 {
		t.Fatalf("expected only the non-import declaration, got %d nodes", len(nodes))
	}
}

func TestParseSnippetGoFileWithImportsPositionsAddressRealFile(t *testing.T) {
	// A templ top-level Go chunk whose imports precede the declarations:
	// snippetTargets skips the import, so the first TARGET node is not at
	// src[0]. The synthetic-to-file shift must anchor on the wrapper
	// geometry, or the whole tree maps onto the import block (garbage ranges
	// that crash go-finding's range validation downstream).
	decl1 := "func helper() int {\n\treturn 7\n}"
	decl2 := "func other() string {\n\treturn \"x\"\n}"
	realFile := "package api\n\nimport \"fmt\"\n\n" + decl1 + "\n\n" + decl2 + "\n"
	chunkOffset := len("package api\n\n")
	src := realFile[chunkOffset:]

	nodes := mustSnippet(t, SnippetGoFile, src, "real.templ", chunkOffset, DetectionModeSemantic)

	if len(nodes) != 2 {
		t.Fatalf("expected both non-import declarations, got %d nodes", len(nodes))
	}

	for i, want := range []string{decl1, decl2} {
		if int(nodes[i].Pos) < chunkOffset || int(nodes[i].End) > len(realFile) {
			t.Fatalf("decl %d range [%d,%d) escapes the real file", i, nodes[i].Pos, nodes[i].End)
		}

		if got := realFile[nodes[i].Pos:nodes[i].End]; got != want {
			t.Errorf("decl %d slices %q, want %q", i, got, want)
		}
	}
}

func TestParseSnippetLeadingWhitespacePositionsAddressRealFile(t *testing.T) {
	// templ GoCode chunks typically start with "\n\t" after the opening
	// fence: fileOffset addresses the raw chunk start while parsing uses the
	// trimmed src. Positions must address the trimmed-away bytes too.
	src := "\n\tx := user.Count\n\ty := user.Total"
	chunkOffset := 500
	firstByte := chunkOffset + len("\n\t")

	nodes := mustSnippet(t, SnippetStatements, src, "ws.templ", chunkOffset, DetectionModeSemantic)

	if len(nodes) != 2 {
		t.Fatalf("expected 2 statement nodes, got %d", len(nodes))
	}

	type rng struct {
		pos, end int32
		text     string
	}

	for i, want := range []rng{
		{pos: int32(firstByte), end: int32(firstByte + len("x := user.Count")), text: "x := user.Count"},
		{pos: int32(firstByte + len("x := user.Count\n\t")), end: int32(firstByte + len("x := user.Count\n\ty := user.Total")), text: "y := user.Total"},
	} {
		if nodes[i].Pos != want.pos || nodes[i].End != want.end {
			t.Errorf("statement %d range = [%d,%d), want [%d,%d)", i, nodes[i].Pos, nodes[i].End, want.pos, want.end)
		}
	}
}
