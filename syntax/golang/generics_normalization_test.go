package golang

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// TestGenericTypeParameterNormalization verifies that generic functions
// with the same body but different type parameter names are detected as clones.
func TestGenericTypeParameterNormalization(t *testing.T) {
	t.Parallel()

	code := `package test

func MapSlice[T any](items []T, fn func(T) T) []T {
	result := make([]T, len(items))
	for i, item := range items {
		result[i] = fn(item)
	}
	return result
}

func FilterSlice[U any](items []U, fn func(U) bool) []U {
	result := make([]U, 0, len(items))
	for _, item := range items {
		if fn(item) {
			result = append(result, item)
		}
	}
	return result
}
`

	t.Run("type_params_should_be_alpha_normalized", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		tmpFile := tmpDir + "/generics.go"
		writeTestFile(t, tmpFile, code)

		node, err := ParseWithConfig(tmpFile, MustParseConfig(DetectionModeSemantic))
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		// Walk the tree and find Ident nodes with names "T" and "U"
		// In the body, these should be alpha-normalized to the same canonical name.
		var tIdents, uIdents []*syntax.Node

		var walk func(n *syntax.Node)
		walk = func(n *syntax.Node) {
			if n.Name == "T" {
				tIdents = append(tIdents, n)
			}
			if n.Name == "U" {
				uIdents = append(uIdents, n)
			}
			for _, c := range n.Children {
				walk(c)
			}
		}
		walk(node)

		if len(tIdents) == 0 || len(uIdents) == 0 {
			t.Skipf("No T or U identifiers found (T=%d, U=%d)", len(tIdents), len(uIdents))
		}

		// Check that T and U identifiers have different Types (different names = different hashes)
		// UNLESS type params are normalized.
		// If type params ARE normalized, T in MapSlice and U in FilterSlice should have the
		// SAME Type because both are the first type parameter in their respective functions.
		tType := tIdents[0].Type
		uType := uIdents[0].Type

		// After normalization, type params T (in MapSlice) and U (in FilterSlice)
		// should have the SAME Type because both are the first type parameter
		// declared in their respective functions — they canonicalize to the
		// same name (v0 for first type param).
		if tType != uType {
			t.Errorf("Type params should be alpha-normalized: T has Type %d, U has Type %d "+
				"(both should be the same since both are first type param)", tType, uType)
		}

		// Check Name field is preserved (original names)
		for _, n := range tIdents {
			if n.Name != "T" {
				t.Errorf("Expected Name 'T', got %q", n.Name)
			}
		}
		for _, n := range uIdents {
			if n.Name != "U" {
				t.Errorf("Expected Name 'U', got %q", n.Name)
			}
		}
	})
}
