package templ

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// elementTokenTypes parses both sources and returns the token Type of the
// first Element node in each. Element tokens are statement composites, so
// the first statement child of the component body is the element.
func elementTokenTypes(t *testing.T, mode golang.DetectionMode, srcA, srcB string) (int32, int32) {
	t.Helper()

	nodeA, _, err := ParseBytesWithMode("a.templ", []byte(srcA), mode)
	if err != nil {
		t.Fatalf("ParseBytesWithMode(a, %s) error = %v", mode, err)
	}

	nodeB, _, err := ParseBytesWithMode("b.templ", []byte(srcB), mode)
	if err != nil {
		t.Fatalf("ParseBytesWithMode(b, %s) error = %v", mode, err)
	}

	firstA := firstStatementToken(nodeA)

	firstB := firstStatementToken(nodeB)

	if firstA == nil || firstB == nil {
		t.Fatalf("expected statement tokens in both trees (mode %s)", mode)
	}

	return firstA.Type, firstB.Type
}

func firstStatementToken(root *syntax.Node) *syntax.Node {
	for _, child := range root.Children {
		if child.Statement {
			return child
		}

		found := firstStatementToken(child)
		if found != nil {
			return found
		}
	}

	return nil
}

func TestExactModeDistinguishesTagNames(t *testing.T) {
	srcA := `package main

templ A() {
	<div>content</div>
}
`
	srcB := `package main

templ B() {
	<section>content</section>
}
`

	typeA, typeB := elementTokenTypes(t, golang.DetectionModeExact, srcA, srcB)
	if typeA == typeB {
		t.Errorf("exact mode must hash tag names verbatim: <div> and <section> produced identical token %d", typeA)
	}
}

func TestStructuralModeIgnoresTagNames(t *testing.T) {
	srcA := `package main

templ A() {
	<div>content</div>
}
`
	srcB := `package main

templ B() {
	<section>content</section>
}
`

	typeA, typeB := elementTokenTypes(t, golang.DetectionModeStructural, srcA, srcB)
	if typeA != typeB {
		t.Errorf("structural mode must ignore tag names: got %d vs %d", typeA, typeB)
	}
}

func TestSemanticModeDistinguishesTagNames(t *testing.T) {
	srcA := `package main

templ A() {
	<div>content</div>
}
`
	srcB := `package main

templ B() {
	<section>content</section>
}
`

	typeA, typeB := elementTokenTypes(t, golang.DetectionModeSemantic, srcA, srcB)
	if typeA == typeB {
		t.Errorf("semantic mode must hash tag names: <div> and <section> produced identical token %d", typeA)
	}
}

func TestExactModeSameTagMatchesVerbatim(t *testing.T) {
	srcA := `package main

templ A() {
	<div>content</div>
}
`

	typeA, typeB := elementTokenTypes(t, golang.DetectionModeExact, srcA, srcA)
	if typeA != typeB {
		t.Errorf("identical sources must produce identical tokens in exact mode: %d vs %d", typeA, typeB)
	}
}
