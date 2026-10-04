package golang

import (
	"fmt"
	"go/ast"
	"go/token"
)

// normalizer canonicalizes local identifier names within a single function so
// that structurally identical functions with different variable names produce
// identical token sequences. This is the foundation of Type 2 (parameterized)
// clone detection.
//
// Only LOCAL bindings are renamed: function parameters, the receiver, and
// variables declared inside the body. Exported identifiers that are part of
// the API surface — field selectors, type names, package qualifiers, called
// functions — keep their original names so that genuinely different code does
// not collapse into the same token stream.
//
// Canonical names are assigned in source order (first-declared = v0), which is
// deterministic: two functions with the same declaration structure but renamed
// variables receive the same canonical mapping.
//
// Limitation: the symbol table is flat per function (no nested-scope shadowing
// resolution). Shadowed variables are rare in duplicated code; the first
// declaration wins, which is correct for the common case.
type normalizer struct {
	enabled bool
	locals  map[string]string
	counter int
}

func newNormalizer(enabled bool) *normalizer {
	return &normalizer{enabled: enabled, locals: make(map[string]string)}
}

// beginFunction resets the symbol table for a new function scope.
func (n *normalizer) beginFunction() {
	if !n.enabled {
		return
	}

	n.locals = make(map[string]string)
	n.counter = 0
}

// declare registers a local binding (parameter, receiver, or body variable)
// under the next canonical name. Blank identifiers and re-declarations are
// ignored. Calling this when disabled is a no-op.
func (n *normalizer) declare(realName string) {
	if !n.enabled || realName == "" || realName == "_" {
		return
	}

	if _, exists := n.locals[realName]; exists {
		return
	}

	n.locals[realName] = fmt.Sprintf("v%d", n.counter)
	n.counter++
}

// resolve returns the canonical name for a declared local, or the original
// name unchanged if it is not a local (field, type, package, or called
// function). When disabled, always returns the original name.
func (n *normalizer) resolve(realName string) string {
	if !n.enabled {
		return realName
	}

	if canonical, ok := n.locals[realName]; ok {
		return canonical
	}

	return realName
}

// isLocal reports whether the given name is a declared local binding in the
// current function scope. Used by the type-aware transformer to decide whether
// to look up and encode the identifier's type.
func (n *normalizer) isLocal(name string) bool {
	if !n.enabled {
		return false
	}

	_, ok := n.locals[name]

	return ok
}

// collectFunctionLocals walks a FuncDecl in source order and declares every
// local binding: receiver, type parameters (generics), parameters, results,
// and body-local variables. This pre-pass populates the symbol table before
// the main transform runs, so every Ident in the body resolves to the correct
// canonical name.
func (n *normalizer) collectFunctionLocals(fd *ast.FuncDecl) {
	if !n.enabled || fd == nil {
		return
	}

	// Receiver binding (value or pointer receiver variable name).
	if fd.Recv != nil {
		declareFieldListNames(n, fd.Recv)
	}

	// Type parameter bindings (Go generics: func Foo[T any](...)).
	// Type params are declared BEFORE regular params so they get canonical
	// names in source order (T before receiver/params alphabetically).
	if fd.Type.TypeParams != nil {
		declareFieldListNames(n, fd.Type.TypeParams)
	}

	// Parameter and result bindings.
	declareFieldListNames(n, fd.Type.Params)

	if fd.Type.Results != nil {
		declareFieldListNames(n, fd.Type.Results)
	}

	// Body-local bindings, walked in source order for deterministic numbering.
	if fd.Body != nil {
		n.declareBodyLocals(fd.Body)
	}
}

// declareBodyLocals walks a function body declaring every locally-introduced
// binding. Closure (FuncLit) parameters and body locals are also declared
// using the same flat symbol table — this is a known limitation (no scope
// shadowing) but correct for the common case of non-shadowed names.
func (n *normalizer) declareBodyLocals(body *ast.BlockStmt) {
	ast.Inspect(body, func(node ast.Node) bool {
		switch d := node.(type) {
		case *ast.AssignStmt:
			n.declareDefineLHS(d)
		case *ast.ValueSpec:
			for _, name := range d.Names {
				n.declare(name.Name)
			}
		case *ast.RangeStmt:
			n.declareIdent(d.Key)
			n.declareIdent(d.Value)
		case *ast.TypeSwitchStmt:
			n.declareTypeSwitchGuard(d)
		case *ast.FuncLit:
			// Declare closure parameters and results, then continue
			// descending into the body to collect closure-local variables.
			declareFieldListNames(n, d.Type.Params)

			if d.Type.Results != nil {
				declareFieldListNames(n, d.Type.Results)
			}
		}

		return true
	})
}

func (n *normalizer) declareDefineLHS(d *ast.AssignStmt) {
	if d.Tok != token.DEFINE {
		return
	}

	for _, lhs := range d.Lhs {
		n.declareIdent(lhs)
	}
}

func (n *normalizer) declareTypeSwitchGuard(d *ast.TypeSwitchStmt) {
	if d.Assign == nil {
		return
	}

	assign, ok := d.Assign.(*ast.AssignStmt)
	if !ok || assign.Tok != token.DEFINE {
		return
	}

	for _, lhs := range assign.Lhs {
		n.declareIdent(lhs)
	}
}

func (n *normalizer) declareIdent(expr ast.Expr) {
	if ident, ok := expr.(*ast.Ident); ok {
		n.declare(ident.Name)
	}
}

// declareFieldListNames declares all named bindings in a FieldList (params or
// results). A Field can carry multiple names (e.g. "a, b int").
func declareFieldListNames(n *normalizer, fl *ast.FieldList) {
	if fl == nil {
		return
	}

	for _, field := range fl.List {
		n.declareFieldNames(field)
	}
}

func (n *normalizer) declareFieldNames(field *ast.Field) {
	for _, name := range field.Names {
		n.declare(name.Name)
	}
}

// collectSnippetLocals declares every free lowercase-initial identifier in a
// templ-embedded snippet as a local binding, in source order. Snippets have no
// parameter list of their own, so the templ scope's locals surface as free
// identifiers; normalizing them mirrors how the Go pipeline normalizes
// function parameters and keeps renamed-variable clones detectable.
//
// Kept verbatim (API surface): called functions (call Fun), field selectors
// (SelectorExpr.Sel), composite-literal keys (KeyValueExpr.Key), and
// predeclared identifiers (builtins, basic types, nil/true/false/iota).
func (n *normalizer) collectSnippetLocals(node ast.Node) {
	if !n.enabled || node == nil {
		return
	}

	skip := make(map[*ast.Ident]bool)

	var order []*ast.Ident

	ast.Inspect(node, func(child ast.Node) bool {
		switch d := child.(type) {
		case *ast.CallExpr:
			if ident, ok := d.Fun.(*ast.Ident); ok {
				skip[ident] = true
			}
		case *ast.SelectorExpr:
			skip[d.Sel] = true
		case *ast.KeyValueExpr:
			if ident, ok := d.Key.(*ast.Ident); ok {
				skip[ident] = true
			}
		case *ast.Ident:
			order = append(order, d)
		}

		return true
	})

	for _, ident := range order {
		if skip[ident] {
			continue
		}

		if isSnippetLocalName(ident.Name) {
			n.declare(ident.Name)
		}
	}
}

// isSnippetLocalName reports whether a snippet identifier should be treated as
// a local binding: lowercase-initial, non-blank, and not a predeclared
// identifier.
func isSnippetLocalName(name string) bool {
	if name == "" || name == "_" {
		return false
	}

	if name[0] < 'a' || name[0] > 'z' {
		return false
	}

	return !predeclaredSnippetIdentifiers[name]
}

// predeclaredSnippetIdentifiers lists Go's universe-scope identifiers, which
// never refer to templ-local bindings.
var predeclaredSnippetIdentifiers = map[string]bool{ //nolint:gochecknoglobals // static universe-scope lookup
	"append": true, "bool": true, "byte": true, "cap": true, "clear": true,
	"close": true, "complex": true, "complex64": true, "complex128": true,
	"copy": true, "delete": true, "error": true, "false": true,
	"float32": true, "float64": true, "imag": true, "int": true,
	"int8": true, "int16": true, "int32": true, "int64": true, "iota": true,
	"len": true, "make": true, "max": true, "min": true, "new": true,
	"nil": true, "panic": true, "print": true, "println": true, "real": true,
	"recover": true, "rune": true, "string": true, "true": true,
	"uint": true, "uint8": true, "uint16": true, "uint32": true,
	"uint64": true, "uintptr": true, "any": true,
}
