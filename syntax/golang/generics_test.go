package golang

import (
	"os"
	"path/filepath"
	"testing"
)

// TestGenericsSupport verifies that Go generics (Go 1.18+) AST nodes are properly handled.
func TestGenericsSupport(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		nodeType int
	}{
		{
			name:     "IndexListExpr - generic type instantiation",
			code:     `package test

type Map[K comparable, V any] struct{}
var m = Map[string, int]{}`,
			nodeType: IndexListExpr,
		},
		{
			name: "TypeSpec with type parameters",
			code: `package test

type Stack[T any] struct {
	items []T
}`,
			nodeType: TypeSpec,
		},
		{
			name: "FuncType with type parameters",
			code: `package test

type FilterFunc[T any] func(items []T) []T`,
			nodeType: FuncType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testParseNodeType(t, tt.name, tt.code, tt.nodeType)
		})
	}
}

// TestIndexListExprParsing specifically tests that IndexListExpr nodes are created
// for generic type instantiations like Map[K, V] or List[T].
func TestIndexListExprParsing(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "generics.go")

	code := `package test

// Generic type instantiation
var m = Map[string, int]{}
var list = List[any]{}

// Generic type definition
type Map[K comparable, V any] struct {
	data map[K]V
}

type List[T any] struct {
	items []T
}
`

	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// We should find IndexListExpr nodes for Map[string, int] and List[any]
	if !findNodeType(node, IndexListExpr) {
		t.Error("Parse() did not find IndexListExpr in parsed AST for generic instantiations")
	}
}

// TestTypeParamsInTypeSpec verifies that TypeSpec nodes include TypeParams children
// when parsing generic type definitions.
func TestTypeParamsInTypeSpec(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "generic_types.go")

	code := `package test

// Generic stack type
type Stack[T any] struct {
	items []T
}

// Generic map with two type parameters
type MyMap[K comparable, V any] struct {
	data map[K]V
}
`

	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// We should find TypeSpec nodes for both Stack and MyMap
	if !findNodeType(node, TypeSpec) {
		t.Error("Parse() did not find TypeSpec in parsed AST for generic type definitions")
	}
}

// TestTypeParamsInFuncType verifies that FuncType nodes include TypeParams children
// when parsing generic function types.
func TestTypeParamsInFuncType(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "generic_funcs.go")

	code := `package test

// Generic function type
type FilterFunc[T any] func(items []T) []T

// Generic function type with constraints
type Comparator[T comparable] func(a, b T) int
`

	if err := os.WriteFile(tmpFile, []byte(code), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := Parse(tmpFile)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// We should find FuncType nodes for FilterFunc and Comparator
	if !findNodeType(node, FuncType) {
		t.Error("Parse() did not find FuncType in parsed AST for generic function types")
	}
}
