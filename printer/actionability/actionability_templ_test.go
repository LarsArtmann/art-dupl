package actionability

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
	templpkg "github.com/LarsArtmann/art-dupl/syntax/templ"
)

// templChildrenSrc contains two wrapper components whose only shared markup
// is the `{ children... }` slot-injection expression (issue #7).
const templChildrenSrc = `package views

templ shellA() {
	<div class="wrap">{ children... }</div>
}

templ shellB() {
	<section class="wrap">{ children... }</section>
}
`

// TestIsTemplChildrenIdiom_RealTemplSource is the definitive end-to-end test:
// it parses ACTUAL .templ source, converts the children-expression nodes into
// domain.CloneNodes, and verifies the pattern fires on the real (non-hand-
// built) node tree through the full evaluation pipeline.
func TestIsTemplChildrenIdiom_RealTemplSource(t *testing.T) {
	t.Parallel()

	root, _, err := templpkg.ParseBytes("test.templ", []byte(templChildrenSrc))
	if err != nil {
		t.Fatalf("parse templ: %v", err)
	}

	children := findNodesByBaseType(root, templpkg.ComponentChildrenExpression)
	if len(children) != 2 {
		t.Fatalf("expected 2 ComponentChildrenExpression nodes, got %d", len(children))
	}

	seqs := make([][]*domain.CloneNode, len(children))
	for i, n := range children {
		seqs[i] = []*domain.CloneNode{cloneNodeFromSyntax(n)}
	}

	if !isTemplChildrenIdiom(seqs) {
		t.Error("isTemplChildrenIdiom = false for real { children... } nodes")
	}

	label, action := EvaluateActionabilityWithLabel(seqs)
	if action != domain.NonActionable {
		t.Errorf("action = %q, want %q", action, domain.NonActionable)
	}

	if label != PatternTemplChildrenIdiom {
		t.Errorf("label = %q, want %q", label, PatternTemplChildrenIdiom)
	}
}

// TestIsTemplChildrenIdiom_Disabled guards --disable-pattern semantics: with
// the label disabled, the group is not suppressed through this pattern.
func TestIsTemplChildrenIdiom_Disabled(t *testing.T) {
	t.Parallel()

	root, _, err := templpkg.ParseBytes("test.templ", []byte(templChildrenSrc))
	if err != nil {
		t.Fatalf("parse templ: %v", err)
	}

	children := findNodesByBaseType(root, templpkg.ComponentChildrenExpression)
	seqs := make([][]*domain.CloneNode, len(children))
	for i, n := range children {
		seqs[i] = []*domain.CloneNode{cloneNodeFromSyntax(n)}
	}

	label, action := EvaluateActionabilityWithLabelAndDisabled(
		seqs, map[PatternLabel]bool{PatternTemplChildrenIdiom: true},
	)
	if action == domain.NonActionable && label == PatternTemplChildrenIdiom {
		t.Error("pattern still suppressed with templ-children-idiom disabled")
	}
}

// TestIsTemplChildrenIdiom_WrapperElementsDoNotMatch pins the scope: only the
// BARE children expression matches. The surrounding wrapper elements (div vs
// section) are real markup and stay actionable — identical wrappers are
// genuinely extractable shared components.
func TestIsTemplChildrenIdiom_WrapperElementsDoNotMatch(t *testing.T) {
	t.Parallel()

	root, _, err := templpkg.ParseBytes("test.templ", []byte(templChildrenSrc))
	if err != nil {
		t.Fatalf("parse templ: %v", err)
	}

	wrappers := findNodesByBaseType(root, templpkg.Element)
	if len(wrappers) != 2 {
		t.Fatalf("expected 2 Element nodes, got %d", len(wrappers))
	}

	seqs := make([][]*domain.CloneNode, len(wrappers))
	for i, n := range wrappers {
		seqs[i] = []*domain.CloneNode{cloneNodeFromSyntax(n)}
	}

	if isTemplChildrenIdiom(seqs) {
		t.Error("isTemplChildrenIdiom matched wrapper Element nodes; only the bare children expression matches")
	}
}

// TestIsTemplChildrenIdiom_GoFuncLitCollisionDoesNotMatch pins the parallel
// iota collision: golang.FuncLit is numerically 22 == templ
// ComponentChildrenExpression. A Go FuncLit clone must never be suppressed by
// this pattern (the .templ filename scoping is load-bearing).
func TestIsTemplChildrenIdiom_GoFuncLitCollisionDoesNotMatch(t *testing.T) {
	t.Parallel()

	src := `package fixture

func a() func() int { return func() int { return 1 } }

func b() func() int { return func() int { return 2 } }
`
	path := writeTempGo(t, "funclit.go", src)

	root, err := golang.Parse(path)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	lits := findNodesByBaseType(root, golang.FuncLit)
	if len(lits) != 2 {
		t.Fatalf("expected 2 FuncLit nodes, got %d", len(lits))
	}

	seqs := make([][]*domain.CloneNode, len(lits))
	for i, n := range lits {
		seqs[i] = []*domain.CloneNode{cloneNodeFromSyntax(n)}
	}

	if isTemplChildrenIdiom(seqs) {
		t.Error("isTemplChildrenIdiom matched golang.FuncLit (22) nodes; " +
			"the .templ filename scoping is missing — numeric collision with ComponentChildrenExpression")
	}
}
