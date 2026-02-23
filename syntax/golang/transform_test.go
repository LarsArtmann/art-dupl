package golang

import (
	"go/ast"
	"os"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

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
						Type: &ast.Ident{Name: "MyType"},
					},
				},
			},
			expected: "MyType",
		},
		{
			name: "pointer receiver",
			recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Type: &ast.StarExpr{
							X: &ast.Ident{Name: "MyType"},
						},
					},
				},
			},
			expected: "MyType",
		},
		{
			name: "receiver with names",
			recv: &ast.FieldList{
				List: []*ast.Field{
					{
						Names: []*ast.Ident{{Name: "m"}},
						Type:  &ast.Ident{Name: "MyType"},
					},
				},
			},
			expected: "MyType",
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
	original := SemanticHashEnabled
	SemanticHashEnabled = true
	defer func() { SemanticHashEnabled = original }()

	// Parse a file with methods on different types
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
	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

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
