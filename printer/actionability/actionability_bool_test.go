package actionability

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func TestIsBoolAccumulatorInitializer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		seqs     [][]*domain.CloneNode
		expected bool
	}{
		{
			name: "two bool initializer pairs across files",
			seqs: [][]*domain.CloneNode{
				{
					boolAssign("hasFloatFormat", "false"),
					boolAssign("hasSeparatorLoop", "false"),
				},
				{
					boolAssign("hasAll", "false"),
					boolAssign("hasLinter", "false"),
				},
			},
			expected: true,
		},
		{
			name: "two bool initializer pairs with true values",
			seqs: [][]*domain.CloneNode{
				{
					boolAssign("enabled", "true"),
					boolAssign("ready", "true"),
				},
				{
					boolAssign("found", "true"),
					boolAssign("valid", "true"),
				},
			},
			expected: true,
		},
		{
			name: "single bool initializer pair is not enough",
			seqs: [][]*domain.CloneNode{
				{boolAssign("hasX", "false")},
				{boolAssign("hasY", "false")},
			},
			expected: false,
		},
		{
			name: "mixed with non-bool assignment",
			seqs: [][]*domain.CloneNode{
				{
					boolAssign("hasX", "false"),
					boolAssign("hasY", "false"),
				},
				{
					boolAssign("count", "0"),
					boolAssign("hasY", "false"),
				},
			},
			expected: false,
		},
		{
			name: "assignment between boolean variables is not a literal init",
			seqs: [][]*domain.CloneNode{
				{
					boolAssign("hasX", "false"),
					boolAssign("hasY", "false"),
				},
				{
					boolAssign("hasA", "otherBool"),
					boolAssign("hasB", "anotherBool"),
				},
			},
			expected: false,
		},
		{
			name: "multi-variable bool initialization across two statements",
			seqs: [][]*domain.CloneNode{
				{
					multiBoolAssign([]string{"x", "y"}, []string{"false", "true"}),
					multiBoolAssign([]string{"u", "v"}, []string{"true", "false"}),
				},
				{
					multiBoolAssign([]string{"a", "b"}, []string{"false", "true"}),
					multiBoolAssign([]string{"c", "d"}, []string{"true", "false"}),
				},
			},
			expected: true,
		},
		{
			name: "var-declaration bool init pairs (DeclStmt form)",
			seqs: [][]*domain.CloneNode{
				{
					boolVarDecl("hasFloatFormat", "bool", "false"),
					boolVarDecl("hasSeparatorLoop", "bool", "false"),
				},
				{
					boolVarDecl("hasAll", "bool", "false"),
					boolVarDecl("hasLinter", "bool", "false"),
				},
			},
			expected: true,
		},
		{
			name: "var-declaration without type annotation",
			seqs: [][]*domain.CloneNode{
				{
					boolVarDeclNoType("enabled", "true"),
					boolVarDeclNoType("ready", "true"),
				},
				{
					boolVarDeclNoType("found", "true"),
					boolVarDeclNoType("valid", "true"),
				},
			},
			expected: true,
		},
		{
			name: "package-level ValueSpec bool init pairs",
			seqs: [][]*domain.CloneNode{
				{
					boolValueSpec("FeatureA", "false"),
					boolValueSpec("FeatureB", "false"),
				},
				{
					boolValueSpec("OptionX", "false"),
					boolValueSpec("OptionY", "false"),
				},
			},
			expected: true,
		},
		{
			name: "var-declaration with non-bool value is not matched",
			seqs: [][]*domain.CloneNode{
				{
					boolVarDecl("count", "int", "0"),
					boolVarDecl("limit", "int", "0"),
				},
				{
					boolVarDecl("max", "int", "0"),
					boolVarDecl("min", "int", "0"),
				},
			},
			expected: false,
		},
		{
			name: "grouped var-declaration (multi-spec) is not matched",
			seqs: [][]*domain.CloneNode{
				{
					groupedVarDecl(
						[]string{"a", "b"},
						[]string{"false", "true"},
					),
				},
				{
					groupedVarDecl(
						[]string{"c", "d"},
						[]string{"false", "true"},
					),
				},
			},
			expected: false,
		},
		{
			name:     "empty sequences are vacuously true",
			seqs:     [][]*domain.CloneNode{},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isBoolAccumulatorInitializer(tt.seqs)
			if result != tt.expected {
				t.Errorf("isBoolAccumulatorInitializer() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsBoolVariableInitialization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		node     *domain.CloneNode
		expected bool
	}{
		{
			name:     "simple false init",
			node:     boolAssign("hasX", "false"),
			expected: true,
		},
		{
			name:     "simple true init",
			node:     boolAssign("hasX", "true"),
			expected: true,
		},
		{
			name:     "numeric literal is not a bool init",
			node:     boolAssign("count", "0"),
			expected: false,
		},
		{
			name:     "variable-to-variable assignment is not a literal init",
			node:     boolAssign("x", "y"),
			expected: false,
		},
		{
			name: "CallExpr children are not a literal init",
			node: &domain.CloneNode{
				BaseType: golang.AssignStmt,
				Children: []*domain.CloneNode{{BaseType: golang.CallExpr}},
			},
			expected: false,
		},
		{
			name:     "single child is not a literal init",
			node:     &domain.CloneNode{BaseType: golang.AssignStmt, Children: []*domain.CloneNode{ident("x")}},
			expected: false,
		},
		{
			name:     "var-declaration with bool literal",
			node:     boolVarDecl("hasX", "bool", "false"),
			expected: true,
		},
		{
			name:     "var-declaration without type annotation",
			node:     boolVarDeclNoType("enabled", "true"),
			expected: true,
		},
		{
			name:     "ValueSpec (package-level) with bool literal",
			node:     boolValueSpec("FeatureA", "false"),
			expected: true,
		},
		{
			name:     "var-declaration with non-bool literal value",
			node:     boolVarDecl("count", "int", "0"),
			expected: false,
		},
		{
			name:     "DeclStmt wrapping TypeSpec is not matched",
			node:     &domain.CloneNode{BaseType: golang.DeclStmt, Children: []*domain.CloneNode{{BaseType: golang.GenDecl, Children: []*domain.CloneNode{{BaseType: golang.TypeSpec}}}}},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := isBoolVariableInitialization(tt.node)
			if result != tt.expected {
				t.Errorf("isBoolVariableInitialization() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsBoolLiteralName(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		expected bool
	}{
		{"false", "false", true},
		{"true", "true", true},
		{"False", "False", false},
		{"TRUE", "TRUE", false},
		{"x", "x", false},
		{"", "", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := isBoolLiteralName(tc.input)
			if result != tc.expected {
				t.Errorf("isBoolLiteralName(%q) = %v, want %v", tc.input, result, tc.expected)
			}
		})
	}
}

func TestBoolAccumulatorInitializer_RealGoSource(t *testing.T) {
	t.Parallel()

	src := `package fixture

func detectA(fn *ast.FuncDecl) bool {
	hasFloatFormat := false
	hasSeparatorLoop := false
	return hasFloatFormat || hasSeparatorLoop
}

func detectB(tokens []string) bool {
	hasAll := false
	hasLinter := false
	return hasAll || hasLinter
}
`

	path := writeTempGo(t, "fixture.go", src)

	root, err := golang.Parse(path)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	assignStmts := findNodesByBaseType(root, golang.AssignStmt)
	if len(assignStmts) < 4 {
		t.Fatalf("expected at least 4 AssignStmt nodes, got %d", len(assignStmts))
	}

	seqs := [][]*domain.CloneNode{
		{cloneNodeFromSyntax(assignStmts[0]), cloneNodeFromSyntax(assignStmts[1])},
		{cloneNodeFromSyntax(assignStmts[2]), cloneNodeFromSyntax(assignStmts[3])},
	}

	label, action := EvaluateActionabilityWithLabel(seqs)
	if action != domain.NonActionable {
		t.Errorf("action = %q, want %q", action, domain.NonActionable)
	}

	if label != PatternBoolAccumulatorInitializer {
		t.Errorf("label = %q, want %q", label, PatternBoolAccumulatorInitializer)
	}
}

func TestBoolAccumulatorInitializer_RealGoSource_VarDecl(t *testing.T) {
	t.Parallel()

	src := `package fixture

func detectA() {
	var hasFloatFormat bool = false
	var hasSeparatorLoop bool = false
	_ = hasFloatFormat || hasSeparatorLoop
}

func detectB() {
	var hasAll bool = false
	var hasLinter bool = false
	_ = hasAll || hasLinter
}
`

	path := writeTempGo(t, "fixture.go", src)

	root, err := golang.Parse(path)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	declStmts := findNodesByBaseType(root, golang.DeclStmt)
	if len(declStmts) < 4 {
		t.Fatalf("expected at least 4 DeclStmt nodes, got %d", len(declStmts))
	}

	seqs := [][]*domain.CloneNode{
		{cloneNodeFromSyntax(declStmts[0]), cloneNodeFromSyntax(declStmts[1])},
		{cloneNodeFromSyntax(declStmts[2]), cloneNodeFromSyntax(declStmts[3])},
	}

	label, action := EvaluateActionabilityWithLabel(seqs)
	if action != domain.NonActionable {
		t.Errorf("action = %q, want %q", action, domain.NonActionable)
	}

	if label != PatternBoolAccumulatorInitializer {
		t.Errorf("label = %q, want %q", label, PatternBoolAccumulatorInitializer)
	}
}

func boolAssign(lhs, rhs string) *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.AssignStmt,
		Children: []*domain.CloneNode{ident(rhs), ident(lhs)},
	}
}

func multiBoolAssign(lhs, rhs []string) *domain.CloneNode {
	children := make([]*domain.CloneNode, 0, len(lhs)+len(rhs))

	for _, r := range rhs {
		children = append(children, ident(r))
	}

	for _, l := range lhs {
		children = append(children, ident(l))
	}

	return &domain.CloneNode{
		BaseType: golang.AssignStmt,
		Children: children,
	}
}

func ident(name string) *domain.CloneNode {
	return &domain.CloneNode{BaseType: golang.Ident, Name: name}
}

// boolVarDecl builds a DeclStmt -> GenDecl -> ValueSpec tree matching
// `var name boolType = value`.
func boolVarDecl(name, boolType, value string) *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.DeclStmt,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.GenDecl,
				Children: []*domain.CloneNode{
					{
						BaseType: golang.ValueSpec,
						Children: []*domain.CloneNode{
							ident(name),
							ident(boolType),
							ident(value),
						},
					},
				},
			},
		},
	}
}

// boolVarDeclNoType builds a DeclStmt -> GenDecl -> ValueSpec tree matching
// `var name = value` (no explicit type annotation).
func boolVarDeclNoType(name, value string) *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.DeclStmt,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.GenDecl,
				Children: []*domain.CloneNode{
					{
						BaseType: golang.ValueSpec,
						Children: []*domain.CloneNode{
							ident(name),
							ident(value),
						},
					},
				},
			},
		},
	}
}

// boolValueSpec builds a bare ValueSpec (package-level declaration form)
// matching `var name = value` at package scope.
func boolValueSpec(name, value string) *domain.CloneNode {
	return &domain.CloneNode{
		BaseType: golang.ValueSpec,
		Children: []*domain.CloneNode{
			ident(name),
			ident(value),
		},
	}
}

// groupedVarDecl builds a DeclStmt -> GenDecl with multiple ValueSpec children,
// matching `var ( a = false; b = true )`. This should NOT be matched by the
// bool-accumulator pattern because the GenDecl has multiple specs.
func groupedVarDecl(names, values []string) *domain.CloneNode {
	specs := make([]*domain.CloneNode, 0, len(names))
	for i, name := range names {
		value := values[i]
		specs = append(specs, &domain.CloneNode{
			BaseType: golang.ValueSpec,
			Children: []*domain.CloneNode{ident(name), ident(value)},
		})
	}

	return &domain.CloneNode{
		BaseType: golang.DeclStmt,
		Children: []*domain.CloneNode{
			{
				BaseType: golang.GenDecl,
				Children: specs,
			},
		},
	}
}
