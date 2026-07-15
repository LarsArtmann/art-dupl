package golang

import (
	"go/ast"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

const testTypeName = "MyType"

func TestExtractReceiverTypeName(t *testing.T) {
	tests := []struct {
		name     string
		recv     *ast.FieldList
		expected string
	}{
		{
			name:     "nil receiver",
			recv:     nil,
			expected: "",
		},
		{
			name:     "empty receiver list",
			recv:     &ast.FieldList{List: []*ast.Field{}},
			expected: "",
		},
		{
			name: "value receiver",
			recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.Ident{Name: testTypeName},
					},
				},
			},
			expected: testTypeName,
		},
		{
			name: "pointer receiver",
			recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.StarExpr{
							X: &ast.Ident{Name: testTypeName},
						},
					},
				},
			},
			expected: testTypeName,
		},
		{
			name: "receiver with names",
			recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{{Name: "m"}},
						Type:  &ast.Ident{Name: testTypeName},
					},
				},
			},
			expected: testTypeName,
		},
		{
			name: "nil type in field",
			recv: &ast.FieldList{
				List: []*ast.Field{
					{Type: nil},
				},
			},
			expected: "",
		},
		{
			name: "complex type (not Ident or StarExpr)",
			recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.SelectorExpr{
							X:   &ast.Ident{Name: "pkg"},
							Sel: &ast.Ident{Name: "Type"},
						},
					},
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractReceiverTypeName(tt.recv)
			if result != tt.expected {
				t.Errorf("extractReceiverTypeName() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSemanticFuncDeclDetection(t *testing.T) {
	// Parse a file with methods on different types (semantic mode is default)
	code := `package test

type CrushMode string
type SafetyMode string

func (cm CrushMode) IsValid() bool { return true }
func (sm SafetyMode) IsValid() bool { return true }

func ParseCrushMode(s string) CrushMode { return CrushMode(s) }
func ParseSafetyMode(s string) SafetyMode { return SafetyMode(s) }
`

	tmpDir := t.TempDir()

	tmpFile := tmpDir + "/test.go"
	writeTestFile(t, tmpFile, code)

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Collect all FuncDecl types
	funcDeclTypes := make(map[int32]string)

	var collectFuncDecls func(n *syntax.Node)

	collectFuncDecls = func(n *syntax.Node) {
		if DecodeBaseType(n.Type) == FuncDecl {
			// Store the semantic hash
			funcDeclTypes[n.Type] = ""
		}

		for _, child := range n.Children {
			collectFuncDecls(child)
		}
	}
	collectFuncDecls(node)

	// We should have 4 unique FuncDecl types (all semantically different)
	// - IsValid on CrushMode
	// - IsValid on SafetyMode
	// - ParseCrushMode (function)
	// - ParseSafetyMode (function)
	uniqueTypes := len(funcDeclTypes)
	if uniqueTypes != 4 {
		t.Errorf("Expected 4 unique FuncDecl types, got %d", uniqueTypes)
		t.Logf("Types: %v", funcDeclTypes)
	}
}

func TestBasicLitValueHashing(t *testing.T) {
	t.Parallel()

	code := `package test

func a() int { return 42 }
func b() int { return 999 }
func c() int { return 42 }
`

	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/test.go"
	writeTestFile(t, tmpFile, code)

	t.Run("semantic_mode_normalizes_literal_values_to_kind", func(t *testing.T) {
		t.Parallel()

		node, err := ParseWithConfig(tmpFile, MustParseConfig(DetectionModeSemantic))
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		litTypes := collectBasicLitTypes(node)
		// "42", "999", "42" → 1 distinct type in semantic mode because
		// literal VALUES are normalized to KIND (all INT → same hash).
		// This enables Type-2 clone detection where only literal values differ.
		if len(litTypes) != 1 {
			t.Errorf("Expected 1 BasicLit type in semantic mode (INT kind normalized), got %d: %v",
				len(litTypes), litTypes)
		}
	})

	t.Run("structural_mode_all_literals_same_type", func(t *testing.T) {
		t.Parallel()

		node, err := ParseWithConfig(tmpFile, MustParseConfig(DetectionModeStructural))
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		litTypes := collectBasicLitTypes(node)
		if len(litTypes) != 1 {
			t.Errorf("Expected exactly 1 BasicLit type in structural mode, got %d: %v", len(litTypes), litTypes)
		}
	})
}

func collectBasicLitTypes(node *syntax.Node) map[int32]struct{} {
	result := make(map[int32]struct{})

	var walk func(n *syntax.Node)

	walk = func(n *syntax.Node) {
		if DecodeBaseType(n.Type) == BasicLit {
			result[n.Type] = struct{}{}
		}

		for _, child := range n.Children {
			walk(child)
		}
	}

	walk(node)

	return result
}
