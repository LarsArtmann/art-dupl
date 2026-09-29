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

// tokensEqualKV compares token streams the way the suffix tree consumes
// them: statement tokens are distinguished by their composite Fingerprint
// (Val()), non-statement tokens by their (possibly semantically encoded)
// Type. Comparing Type alone would blind the tests to distinctions that
// live inside statement fingerprints — exactly where composite-literal
// keys differ.
func tokensEqualKV(a, b []*syntax.Node) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].Statement != b[i].Statement {
			return false
		}

		if a[i].Val() != b[i].Val() {
			return false
		}
	}

	return true
}

func TestKeyValueExpr_SelectorKeyEncoding_Semantic(t *testing.T) {
	t.Parallel()

	// Go 1.27: struct-literal keys may be field selectors for
	// embedded/promoted fields (T{A.B: 1}). Different selector paths are
	// different fields — they must NOT collapse into identical tokens.
	src1 := `package test
type InnerA struct { B int }
type OuterA struct { InnerA }
func main() {
	x := 1
	_ = OuterA{A: InnerA{}, InnerA: InnerA{B: x}}
}`

	src2 := `package test
type InnerX struct { Y int }
type OuterX struct { InnerX }
func main() {
	x := 1
	_ = OuterX{X: InnerX{}, InnerX: InnerX{Y: x}}
}`

	// Same selector path, different local variable name: must match
	// (Type-2 renamed clone).
	src3 := `package test
type InnerA struct { B int }
type OuterA struct { InnerA }
func main() {
	a := 1
	_ = OuterA{A: InnerA{}, InnerA: InnerA{B: a}}
}`

	tokensOuterInner := parseAndSerializeT(t, src1, DetectionModeSemantic)
	tokensOtherNames := parseAndSerializeT(t, src2, DetectionModeSemantic)
	tokensSamePath := parseAndSerializeT(t, src3, DetectionModeSemantic)

	if tokensEqualKV(tokensOuterInner, tokensOtherNames) {
		t.Errorf(
			"Outer{Inner.B: x} and Outer{Inner.Y: x} must NOT produce identical tokens (selector keys are API surface)",
		)
	}

	if !tokensEqualKV(tokensOuterInner, tokensSamePath) {
		t.Errorf(
			"Outer{Inner.B: x} and Outer{Inner.B: a} must produce identical tokens (only the local variable differs)",
		)
	}
}

func TestKeyValueExpr_SelectorKey_MultiLevel(t *testing.T) {
	t.Parallel()

	// T{A.B.C: 1} — the AST nests SelectorExpr{X: SelectorExpr{X: Ident(A), Sel: B}, Sel: C}.
	src1 := `package test
type L3 struct { C int }
type L2 struct { L3 }
type L1 struct { L2 }
func main() {
	x := 1
	_ = L1{L2: L2{L3: L3{C: x}}, L2_L3: L2{L3: L3{C: x}}}
}`

	src2 := `package test
type L3 struct { Z int }
type L2 struct { L3 }
type L1 struct { L2 }
func main() {
	x := 1
	_ = L1{L2: L2{L3: L3{Z: x}}, L2_L3: L2{L3: L3{Z: x}}}
}`

	tokens1 := parseAndSerializeT(t, src1, DetectionModeSemantic)
	tokens2 := parseAndSerializeT(t, src2, DetectionModeSemantic)

	if tokensEqualKV(tokens1, tokens2) {
		t.Errorf("L1{L2.L3.C: x} and L1{L2.L3.Z: x} must NOT produce identical tokens")
	}
}
