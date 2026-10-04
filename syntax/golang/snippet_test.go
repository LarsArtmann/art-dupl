package golang

import (
	"testing"

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

func TestParseSnippetExpressionNormalizesLocals(t *testing.T) {
	mode := DetectionModeSemantic

	a := mustSnippet(t, SnippetExpression, "user.Name", "a.templ", 100, mode)
	b := mustSnippet(t, SnippetExpression, "person.Name", "b.templ", 100, mode)

	if a[0].Val() != b[0].Val() {
		t.Errorf("renamed local in expression should match: %d vs %d", a[0].Val(), b[0].Val())
	}

	c := mustSnippet(t, SnippetExpression, "user.Email", "c.templ", 100, mode)
	if a[0].Val() == c[0].Val() {
		t.Error("different field selectors must not match (API surface)")
	}
}

func TestParseSnippetExpressionKeepsCalleeNames(t *testing.T) {
	a := mustSnippet(t, SnippetExpression, "len(items)", "a.templ", 10, DetectionModeSemantic)
	b := mustSnippet(t, SnippetExpression, "len(entries)", "b.templ", 10, DetectionModeSemantic)

	if a[0].Val() != b[0].Val() {
		t.Errorf("len() builtin must keep its name: %d vs %d", a[0].Val(), b[0].Val())
	}

	c := mustSnippet(t, SnippetExpression, "count(items)", "c.templ", 10, DetectionModeSemantic)
	if a[0].Val() == c[0].Val() {
		t.Error("different called functions must not match")
	}
}

func TestParseSnippetExpressionStructuralModeDropsNames(t *testing.T) {
	a := mustSnippet(t, SnippetExpression, "user.Name", "a.templ", 10, DetectionModeStructural)
	b := mustSnippet(t, SnippetExpression, "other.Thing", "b.templ", 10, DetectionModeStructural)

	if a[0].Val() != b[0].Val() {
		t.Errorf("structural mode must drop identifier names: %d vs %d", a[0].Val(), b[0].Val())
	}
}

func TestParseSnippetExpressionExactModeKeepsVerbatimNames(t *testing.T) {
	a := mustSnippet(t, SnippetExpression, "user.Name", "a.templ", 10, DetectionModeExact)
	b := mustSnippet(t, SnippetExpression, "person.Name", "b.templ", 10, DetectionModeExact)

	if a[0].Val() == b[0].Val() {
		t.Error("exact mode must keep identifier names verbatim (no normalization)")
	}
}

func TestParseSnippetIfCond(t *testing.T) {
	a := mustSnippet(t, SnippetIfCond, "user.IsAdmin && user.Active", "a.templ", 200, DetectionModeSemantic)
	b := mustSnippet(t, SnippetIfCond, "person.IsAdmin && person.Active", "b.templ", 250, DetectionModeSemantic)

	if a[0].Val() != b[0].Val() {
		t.Errorf("renamed locals in if-condition should match: %d vs %d", a[0].Val(), b[0].Val())
	}

	c := mustSnippet(t, SnippetIfCond, "user.IsGuest && user.Active", "c.templ", 200, DetectionModeSemantic)
	if a[0].Val() == c[0].Val() {
		t.Error("different conditions must not match")
	}
}

func TestParseSnippetForClause(t *testing.T) {
	a := mustSnippet(t, SnippetForClause, "i, item := range items", "a.templ", 300, DetectionModeSemantic)
	b := mustSnippet(t, SnippetForClause, "j, product := range products", "b.templ", 300, DetectionModeSemantic)

	if a[0].Val() != b[0].Val() {
		t.Errorf("renamed range loop locals should match: %d vs %d", a[0].Val(), b[0].Val())
	}
}

func TestParseSnippetCaseList(t *testing.T) {
	a := mustSnippet(t, SnippetCaseList, `"admin", "root"`, "a.templ", 400, DetectionModeSemantic)
	b := mustSnippet(t, SnippetCaseList, `"guest"`, "b.templ", 400, DetectionModeSemantic)

	if a[0].Val() != b[0].Val() {
		t.Errorf("string literals normalize to kind in semantic mode: %d vs %d", a[0].Val(), b[0].Val())
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
		if a[i].Val() != b[i].Val() {
			t.Errorf("statement %d should match after renaming: %d vs %d", i, a[i].Val(), b[i].Val())
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

	end := mustSnippet(t, SnippetIfCond, "user.IsAdmin", "real.templ", 1000, DetectionModeSemantic)[0]
	if end.End <= end.Pos {
		t.Errorf("node End (%d) must be greater than Pos (%d)", end.End, end.Pos)
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
