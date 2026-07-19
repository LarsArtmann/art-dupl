package templ

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

func TestExtractCalleeName(t *testing.T) {
	tests := []struct {
		name string
		expr string
		want string
	}{
		{"simple call with args", `demoSection("foo", "bar")`, "demoSection"},
		{"package-qualified call", `display.DataTable(display.DataTableProps{...})`, "display.DataTable"},
		{"no parens", "display.Card", "display.Card"},
		{"empty string", "", ""},
		{"only opening paren", "(", "("},
		{"method call", `icons.Icon(icons.Star, "h-3.5")`, "icons.Icon"},
		{"dotted with no parens", "layout.Base", "layout.Base"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractCalleeName(tt.expr)
			if got != tt.want {
				t.Errorf("extractCalleeName(%q) = %q, want %q", tt.expr, got, tt.want)
			}
		})
	}
}

func TestTemplElementSemanticEncoding(t *testing.T) {
	src := `package main

templ ComponentA() {
	@demoSection("A", "a")
}
`

	node, _, err := ParseBytesWithMode("test.templ", []byte(src), true)
	if err != nil {
		t.Fatalf("ParseBytesWithMode() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytesWithMode() returned nil node")
	}

	renderNodes := findComponentRenderNodes(node)
	if len(renderNodes) == 0 {
		t.Fatal("expected at least one ComponentRender node with callee name")
	}

	for _, n := range renderNodes {
		if n.Name != "demoSection" {
			t.Errorf("expected callee name 'demoSection', got %q", n.Name)
		}

		if syntax.DecodeBaseType(n.Type) != int32(ComponentRender) {
			t.Errorf("expected base type ComponentRender (%d), got %d", ComponentRender, syntax.DecodeBaseType(n.Type))
		}
	}
}

func TestDifferentCalleesProduceDifferentTypes(t *testing.T) {
	templA := `package main

templ A() {
	@demoSection("title", "id")
}
`
	templB := `package main

templ A() {
	@display.DataTable()
}
`

	nodesA, nodesB := parseTwoComponentRenderers(t, templA, templB)

	if nodesA[0].Type == nodesB[0].Type {
		t.Error("expected different Types for @demoSection and @display.DataTable, got identical Types")
	}
}

func TestSameCalleeDifferentArgsProduceSameType(t *testing.T) {
	templA := `package main

templ A() {
	@demoSection("title_a", "id_a")
}
`
	templB := `package main

templ A() {
	@demoSection("title_b", "id_b")
}
`

	nodesA, nodesB := parseTwoComponentRenderers(t, templA, templB)

	if nodesA[0].Type != nodesB[0].Type {
		t.Error("expected same Types for @demoSection with different args (Type-2 detection), got different Types")
	}
}

func findComponentRenderNodes(root *syntax.Node) []*syntax.Node {
	var result []*syntax.Node

	var walk func(*syntax.Node)

	walk = func(n *syntax.Node) {
		if n == nil {
			return
		}

		if syntax.DecodeBaseType(n.Type) == int32(ComponentRender) && n.Name != "" {
			result = append(result, n)
		}

		for _, child := range n.Children {
			walk(child)
		}
	}
	walk(root)

	return result
}

func parseTwoComponentRenderers(t *testing.T, srcA, srcB string) (nodesA, nodesB []*syntax.Node) {
	t.Helper()

	nodeA, _, err := ParseBytesWithMode("a.templ", []byte(srcA), true)
	if err != nil {
		t.Fatalf("ParseBytesWithMode(a) error = %v", err)
	}

	nodeB, _, err := ParseBytesWithMode("b.templ", []byte(srcB), true)
	if err != nil {
		t.Fatalf("ParseBytesWithMode(b) error = %v", err)
	}

	nodesA = findComponentRenderNodes(nodeA)
	nodesB = findComponentRenderNodes(nodeB)

	if len(nodesA) == 0 || len(nodesB) == 0 {
		t.Fatal("expected ComponentRender nodes in both trees")
	}

	return nodesA, nodesB
}
