package templ

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

func TestParseBytes(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantNodes bool // whether we expect any nodes
	}{
		{
			name: "simple component",
			input: `package main

templ hello(name string) {
	<div>Hello, { name }!</div>
}`,
			wantNodes: true,
		},
		{
			name: "component with attributes",
			input: `package main

templ page(title string) {
	<html>
		<head><title>{ title }</title></head>
		<body class="container">
			<h1>{ title }</h1>
		</body>
	</html>
}`,
			wantNodes: true,
		},
		{
			name: "component with if statement",
			input: `package main

templ conditional(show bool) {
	if show {
		<p>Visible</p>
	}
}`,
			wantNodes: true,
		},
		{
			name: "component with for loop",
			input: `package main

templ list(items []string) {
	for _, item := range items {
		<li>{ item }</li>
	}
}`,
			wantNodes: true,
		},
		{
			name: "empty input",
			input: `package main

templ empty() {
}`,
			wantNodes: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, lineCount, err := ParseBytes("test.templ", []byte(tt.input))
			if err != nil {
				t.Fatalf("ParseBytes() error = %v", err)
			}
			if lineCount <= 0 {
				t.Errorf("ParseBytes() lineCount = %d, want > 0", lineCount)
			}
			if tt.wantNodes && node == nil {
				t.Errorf("ParseBytes() returned nil node, expected non-nil")
			}
		})
	}
}

func TestNodeTypeConstants(t *testing.T) {
	// Verify all node type constants are defined and non-zero
	types := []int32{
		ComponentDeclaration,
		Element,
		TagStart,
		TagEnd,
		SelfClosingTag,
		Attribute,
		ComponentIfStatement,
		ComponentForStatement,
	}

	for _, nt := range types {
		if nt == 0 {
			t.Error("Node type constant is zero, may not be properly defined")
		}
	}
}

func TestNodeTreeStructure(t *testing.T) {
	input := `package main

templ nested() {
	<div>
		<span>
			<p>Deep</p>
		</span>
	</div>
}`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}

	// Verify tree structure is traversable
	var countNodes func(*syntax.Node) int
	countNodes = func(n *syntax.Node) int {
		if n == nil {
			return 0
		}
		count := 1
		for _, child := range n.Children {
			count += countNodes(child)
		}
		return count
	}

	nodeCount := countNodes(node)
	if nodeCount < 1 {
		t.Errorf("Expected at least 1 node, got %d", nodeCount)
	}
}

func TestParseWithLineCountFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.templ")

	content := `package main

templ hello() {
	<div>Hello</div>
}
`
	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	node, lineCount, err := ParseWithLineCount(tmpFile)
	if err != nil {
		t.Fatalf("ParseWithLineCount() error = %v", err)
	}

	if node == nil {
		t.Error("ParseWithLineCount() returned nil node")
	}

	if lineCount != 6 {
		t.Errorf("ParseWithLineCount() lineCount = %d, want 6", lineCount)
	}
}
