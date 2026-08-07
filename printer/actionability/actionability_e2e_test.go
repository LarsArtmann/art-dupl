package actionability

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
	templpkg "github.com/LarsArtmann/art-dupl/syntax/templ"
)

// cloneNodeFromSyntax mirrors printer.syntaxToCloneNode. It cannot live in the
// printer package because that would create a circular import (printer imports
// actionability). The conversion is identical: decode the base type and recurse.
func cloneNodeFromSyntax(n *syntax.Node) *domain.CloneNode {
	cn := &domain.CloneNode{
		BaseType:             golang.DecodeBaseType(n.Type),
		Name:                 n.Name,
		Filename:             n.Filename,
		VarType:              n.VarType,
		EnclosingReturnArity: n.EnclosingReturnArity,
		InterfaceMethod:      n.InterfaceMethod,
		IsAlias:              n.IsAlias,
	}
	if len(n.Children) > 0 {
		cn.Children = make([]*domain.CloneNode, len(n.Children))
		for i, c := range n.Children {
			cn.Children[i] = cloneNodeFromSyntax(c)
		}
	}

	return cn
}

// findNodesByBaseType walks a syntax tree and collects all nodes whose decoded
// base type matches the target. Used to extract real IfStmt/RangeStmt nodes
// from parsed source.
func findNodesByBaseType(root *syntax.Node, target int32) []*syntax.Node {
	var result []*syntax.Node

	var walk func(n *syntax.Node)

	walk = func(n *syntax.Node) {
		if golang.DecodeBaseType(n.Type) == target {
			result = append(result, n)
		}

		for _, c := range n.Children {
			walk(c)
		}
	}
	walk(root)

	return result
}

// writeTempGo writes Go source to a temp .go file and returns its path.
func writeTempGo(t *testing.T, name, src string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, name)

	err := os.WriteFile(path, []byte(src), 0o600)
	if err != nil {
		t.Fatalf("write temp go file: %v", err)
	}

	return path
}

// TestIsTemplRenderingIdiom_RealGoSource is the definitive end-to-end test: it
// parses ACTUAL Go source containing the templ empty-state rendering idiom,
// converts the resulting syntax tree into domain.CloneNodes, and verifies that
// isTemplRenderingIdiom returns true on the real (non-hand-built) node tree.
//
// This proves the pattern is NOT dead code — it fires on real Go source that
// follows the `if len(x) == 0 { ... } else { for ... }` convention.
func TestIsTemplRenderingIdiom_RealGoSource(t *testing.T) {
	t.Parallel()

	src := `package fixture

func renderUsers(users []string) string {
	if len(users) == 0 {
		return "no users"
	} else {
		for _, u := range users {
			_ = u
		}
		return "ok"
	}
}

func renderItems(items []string) string {
	if len(items) == 0 {
		return "no items"
	} else {
		for _, i := range items {
			_ = i
		}
		return "ok"
	}
}
`

	path := writeTempGo(t, "fixture.go", src)

	root, err := golang.Parse(path)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	ifStmts := findNodesByBaseType(root, golang.IfStmt)
	if len(ifStmts) < 2 {
		t.Fatalf("expected at least 2 IfStmt nodes, got %d", len(ifStmts))
	}

	// Filter to IfStmts that look like the templ idiom (have len in condition).
	var idiomNodes []*domain.CloneNode

	for _, n := range ifStmts {
		cn := cloneNodeFromSyntax(n)
		if isTemplEmptyStateIf(cn) {
			idiomNodes = append(idiomNodes, cn)
		}
	}

	if len(idiomNodes) < 2 {
		t.Fatalf("expected at least 2 templ-idiom IfStmts, got %d", len(idiomNodes))
	}

	// Build sequences (each clone occurrence is a single-node sequence).
	seqs := make([][]*domain.CloneNode, len(idiomNodes))
	for i, cn := range idiomNodes {
		seqs[i] = []*domain.CloneNode{cn}
	}

	label, action := EvaluateActionabilityWithLabel(seqs)
	if action != domain.NonActionable {
		t.Errorf("action = %q, want %q", action, domain.NonActionable)
	}

	if label != PatternTemplRenderingIdiom {
		t.Errorf("label = %q, want %q", label, PatternTemplRenderingIdiom)
	}
}

// TestIsTemplRenderingIdiom_TemplSourceDoesNotMatch documents that .templ source
// files use DIFFERENT node types (templ.ComponentIfStatement=11, not
// golang.IfStmt=31). The pattern therefore does NOT fire on clones detected in
// .templ source. This is by design: the templ transform drops the condition
// expression entirely, so the len() check cannot be verified. The pattern is
// only effective on .go files (including generated _templ.go when included).
//
// If this test ever FAILS (templ nodes start matching), it means either the
// templ transform changed to include conditions, or the node type constants
// changed — both warrant a review of the pattern.
func TestIsTemplRenderingIdiom_TemplSourceDoesNotMatch(t *testing.T) {
	t.Parallel()

	templSrc := `package views

templ userList(users []string) {
	if len(users) == 0 {
		<p>No users found</p>
	} else {
		for _, user := range users {
			<li>{ user }</li>
		}
	}
}

templ itemList(items []string) {
	if len(items) == 0 {
		<p>No items found</p>
	} else {
		for _, item := range items {
			<li>{ item }</li>
		}
	}
}
`

	root, _, err := templpkg.ParseBytes("test.templ", []byte(templSrc))
	if err != nil {
		t.Fatalf("parse templ: %v", err)
	}

	// Verify templ uses ComponentIfStatement, NOT golang.IfStmt.
	componentIfs := findNodesByBaseType(root, int32(templpkg.ComponentIfStatement))
	if len(componentIfs) < 2 {
		t.Fatalf("expected at least 2 ComponentIfStatement nodes, got %d", len(componentIfs))
	}

	// Convert to CloneNodes and check that isTemplRenderingIdiom returns false.
	seqs := make([][]*domain.CloneNode, len(componentIfs))
	for i, n := range componentIfs {
		seqs[i] = []*domain.CloneNode{cloneNodeFromSyntax(n)}
	}

	if isTemplRenderingIdiom(seqs) {
		t.Error("isTemplRenderingIdiom returned true for templ ComponentIfStatement nodes; " +
			"the templ transform does not encode conditions, so this match is unexpected — " +
			"review whether the pattern needs templ-specific handling")
	}

	// Also confirm via the full pipeline: should NOT be suppressed as templ-rendering-idiom.
	label, _ := EvaluateActionabilityWithLabel(seqs)
	if label == PatternTemplRenderingIdiom {
		t.Errorf("label = %q for templ source; templ-rendering-idiom should not fire on .templ files", label)
	}
}
