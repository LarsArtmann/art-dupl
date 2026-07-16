package golang

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

func TestKeyValueExpr_FieldNameEncoding_Semantic(t *testing.T) {
	t.Parallel()

	src1 := `package test
type Point struct { X int }
func main() {
	x := 1
	_ = Point{X: x}
}`

	src2 := `package test
type Size struct { W int }
func main() {
	w := 1
	_ = Size{W: w}
}`

	tokens1 := parseAndSerializeT(t, src1, DetectionModeSemantic)
	tokens2 := parseAndSerializeT(t, src2, DetectionModeSemantic)

	if tokensEqualKV(tokens1, tokens2) {
		t.Errorf(
			"Point{X: x} and Size{W: w} should NOT produce identical tokens in semantic mode (field names are API surface)",
		)
	}
}

func TestKeyValueExpr_SameFieldName_Semantic(t *testing.T) {
	t.Parallel()

	src1 := `package test
type Point struct { X int }
func main() {
	a := 1
	_ = Point{X: a}
}`

	src2 := `package test
type Point struct { X int }
func main() {
	b := 1
	_ = Point{X: b}
}`

	tokens1 := parseAndSerializeT(t, src1, DetectionModeSemantic)
	tokens2 := parseAndSerializeT(t, src2, DetectionModeSemantic)

	if !tokensEqualKV(tokens1, tokens2) {
		t.Errorf(
			"Point{X: a} and Point{X: b} should produce identical tokens in semantic mode (only variable name differs)",
		)
	}
}

func TestKeyValueExpr_FieldNameEncoding_Structural(t *testing.T) {
	t.Parallel()

	src1 := `package test
type Point struct { X int }
func main() {
	x := 1
	_ = Point{X: x}
}`

	src2 := `package test
type Size struct { W int }
func main() {
	w := 1
	_ = Size{W: w}
}`

	tokens1 := parseAndSerializeT(t, src1, DetectionModeStructural)
	tokens2 := parseAndSerializeT(t, src2, DetectionModeStructural)

	if !tokensEqualKV(tokens1, tokens2) {
		t.Errorf("Structural mode should ignore field names")
	}
}

func parseAndSerializeT(t *testing.T, src string, mode DetectionMode) []*syntax.Node {
	t.Helper()

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.go")
	if err := os.WriteFile(tmpFile, []byte(src), 0o644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	node, err := ParseWithConfig(tmpFile, ParseConfig{Mode: mode})
	if err != nil {
		t.Fatalf("ParseWithConfig failed: %v", err)
	}
	if node == nil {
		t.Fatal("ParseWithConfig returned nil node")
	}

	return syntax.Serialize(node)
}

func tokensEqualKV(a, b []*syntax.Node) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Type != b[i].Type {
			return false
		}
	}
	return true
}
