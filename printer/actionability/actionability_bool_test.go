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
