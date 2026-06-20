package golang

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func parseFuncBody(t *testing.T, src string) *ast.FuncDecl {
	t.Helper()

	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, "test.go", src, 0)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}

	for _, decl := range file.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok {
			return fd
		}
	}

	t.Fatal("no FuncDecl found in source")

	return nil
}

func TestNormalizer_DisabledIsPassthrough(t *testing.T) {
	t.Parallel()

	n := newNormalizer(false)

	n.beginFunction()
	n.declare("foo")

	if got := n.resolve("foo"); got != "foo" {
		t.Errorf("disabled normalizer should return original name, got %q", got)
	}
}

func TestNormalizer_CanonicalizesRenamedVariables(t *testing.T) {
	t.Parallel()

	srcA := `package p
func calcTotal(items []int) int {
	total := 0
	for _, it := range items {
		total += it
	}
	return total
}`

	srcB := `package p
func sumScores(records []int) int {
	sum := 0
	for _, rec := range records {
		sum += rec
	}
	return sum
}`

	fdA := parseFuncBody(t, srcA)
	fdB := parseFuncBody(t, srcB)

	nA := newNormalizer(true)
	nA.beginFunction()
	nA.collectFunctionLocals(fdA)

	nB := newNormalizer(true)
	nB.beginFunction()
	nB.collectFunctionLocals(fdB)

	// Corresponding variables must map to the same canonical name.
	pairs := []struct{ aName, bName string }{
		{"items", "records"}, // params
		{"total", "sum"},     // locals
		{"it", "rec"},        // range vars
	}

	for _, p := range pairs {
		aCanonical := nA.resolve(p.aName)
		bCanonical := nB.resolve(p.bName)

		if aCanonical != bCanonical {
			t.Errorf("%q→%q vs %q→%q: canonical names differ (%s ≠ %s)",
				p.aName, aCanonical, p.bName, bCanonical, aCanonical, bCanonical)
		}
	}

	// Non-local names (e.g. a type or field) must be unchanged.
	if got := nA.resolve("NotALocal"); got != "NotALocal" {
		t.Errorf("non-local should be unchanged, got %q", got)
	}
}

func TestNormalizer_PreservesDeclarationOrder(t *testing.T) {
	t.Parallel()

	src := `package p
func f(a int, b string) {
	x := 1
	y := 2
	_ = x + y
	_ = a + len(b)
}`

	fd := parseFuncBody(t, src)

	n := newNormalizer(true)
	n.beginFunction()
	n.collectFunctionLocals(fd)

	// First-declared param = v0, second = v1, first local = v2, etc.
	expected := []struct {
		name      string
		canonical string
	}{
		{"a", "v0"},
		{"b", "v1"},
		{"x", "v2"},
		{"y", "v3"},
	}

	for _, e := range expected {
		if got := n.resolve(e.name); got != e.canonical {
			t.Errorf("%q should canonicalize to %s, got %s", e.name, e.canonical, got)
		}
	}
}

func TestNormalizer_IgnoresBlankIdentifier(t *testing.T) {
	t.Parallel()

	src := `package p
func f() {
	_, keep := twoValues()
	_ = keep
}`

	fd := parseFuncBody(t, src)

	n := newNormalizer(true)
	n.beginFunction()
	n.collectFunctionLocals(fd)

	// Blank identifier must not be declared.
	if got := n.resolve("_"); got != "_" {
		t.Errorf("blank identifier should stay %q, got %q", "_", got)
	}

	// "keep" is the first real local → v0.
	if got := n.resolve("keep"); got != "v0" {
		t.Errorf("keep should be v0, got %s", got)
	}
}

func TestNormalizer_RedoesNotDoubleDeclare(t *testing.T) {
	t.Parallel()

	n := newNormalizer(true)
	n.beginFunction()
	n.declare("x")
	n.declare("x") // re-declaration ignored

	if got := n.resolve("x"); got != "v0" {
		t.Errorf("re-declaration should keep first canonical, got %s", got)
	}

	if n.counter != 1 {
		t.Errorf("counter should be 1 after one unique declaration, got %d", n.counter)
	}
}

func TestNormalizer_HandlesTypeSwitchGuard(t *testing.T) {
	t.Parallel()

	src := `package p
func f(v interface{}) {
	switch t := v.(type) {
	case int:
		_ = t
	}
}`

	fd := parseFuncBody(t, src)

	n := newNormalizer(true)
	n.beginFunction()
	n.collectFunctionLocals(fd)

	// Both the param and the type-switch guard should be declared.
	if got := n.resolve("v"); got == "v" {
		t.Error("param v should be canonicalized")
	}

	if got := n.resolve("t"); got == "t" {
		t.Error("type-switch guard t should be canonicalized")
	}
}

func TestNormalizer_DoesNotDescendIntoFuncLit(t *testing.T) {
	t.Parallel()

	src := `package p
func f() {
	cb := func(closureVar int) {
		_ = closureVar
	}
	_ = cb
}`

	fd := parseFuncBody(t, src)

	n := newNormalizer(true)
	n.beginFunction()
	n.collectFunctionLocals(fd)

	// closureVar is inside a FuncLit — it should NOT be declared in the outer
	// function's table (nested functions have their own scope).
	if got := n.resolve("closureVar"); got != "closureVar" {
		t.Errorf("func-lit local should not be canonicalized by outer scope, got %q", got)
	}

	// cb is a local of the outer function — it should be canonicalized.
	if got := n.resolve("cb"); got == "cb" {
		t.Error("outer local cb should be canonicalized")
	}
}
