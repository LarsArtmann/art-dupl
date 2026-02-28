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
	nodeCount := countAllNodes(node)
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
	if err := os.WriteFile(tmpFile, []byte(content), 0o644); err != nil {
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

func TestParseCSSTemplate(t *testing.T) {
	input := `package main

css className() {
	color: red;
	background: blue;
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}

	// Should have at least one child (CSS declaration)
	if len(node.Children) == 0 {
		t.Error("Expected at least one child node for CSS template")
	}
}

func TestParseScriptTemplate(t *testing.T) {
	input := `package main

script onClick() {
	console.log("clicked");
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}

	// Should have at least one child (Script declaration)
	if len(node.Children) == 0 {
		t.Error("Expected at least one child node for Script template")
	}
}

func TestParseSwitchStatement(t *testing.T) {
	input := `package main

templ switchExample(val int) {
	switch val {
		case 1:
			<p>One</p>
		case 2:
			<p>Two</p>
		default:
			<p>Other</p>
	}
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}

	// Verify tree is traversable
	nodeCount := countAllNodes(node)
	if nodeCount < 3 {
		t.Errorf("Expected at least 3 nodes for switch statement, got %d", nodeCount)
	}
}

func TestParseComponentWithChildren(t *testing.T) {
	input := `package main

templ parent() {
	<div>
		@child()
	</div>
}

templ child() {
	<span>Child</span>
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}

	// Should have 2 component declarations
	if len(node.Children) < 2 {
		t.Errorf("Expected at least 2 component declarations, got %d", len(node.Children))
	}
}

func TestParseElseIf(t *testing.T) {
	input := `package main

templ elseifExample(a, b bool) {
	if a {
		<p>A is true</p>
	} else if b {
		<p>B is true</p>
	} else {
		<p>Neither</p>
	}
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}

	// Count nodes in tree
	nodeCount := countAllNodes(node)
	if nodeCount < 3 {
		t.Errorf("Expected at least 3 nodes for if/elseif/else, got %d", nodeCount)
	}
}

func TestParseExpressionAttributes(t *testing.T) {
	input := `package main

templ attrs(name string, active bool) {
	<div class={ name } disabled?={ active } data-id={ "test" }>Content</div>
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}

	// Verify structure
	nodeCount := countAllNodes(node)
	if nodeCount < 2 {
		t.Errorf("Expected at least 2 nodes with attributes, got %d", nodeCount)
	}
}

func TestParseDoctype(t *testing.T) {
	input := `package main

templ page() {
	<!DOCTYPE html>
	<html>
		<body>Content</body>
	</html>
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}
}

func TestParseScriptElement(t *testing.T) {
	input := `package main

templ withScript() {
	<div>
		<script type="text/javascript">
			console.log("hello");
		</script>
	</div>
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}
}

func TestParseRawElement(t *testing.T) {
	input := `package main

templ withRaw() {
	<div>
		<style>
			.test { color: red; }
		</style>
	</div>
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}
}

func TestParseChildrenExpression(t *testing.T) {
	input := `package main

templ wrapper() {
	<div class="wrapper">
		{ children... }
	</div>
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}
}

func TestParseGoCode(t *testing.T) {
	input := `package main

templ withGoCode() {
	<div>
		{ fmt.Sprintf("Hello %s", "World") }
	</div>
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}
}

func TestParseNestedElements(t *testing.T) {
	input := `package main

templ nested() {
	<div>
		<span>
			<p>
				<a href="#">Link</a>
			</p>
		</span>
	</div>
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}

	// Count all nested nodes
	nodeCount := countAllNodes(node)
	if nodeCount < 5 {
		t.Errorf("Expected at least 5 nodes for nested elements, got %d", nodeCount)
	}
}

func TestParseMultipleComponents(t *testing.T) {
	input := `package main

templ header() {
	<header>Header</header>
}

templ main() {
	<main>Content</main>
}

templ footer() {
	<footer>Footer</footer>
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}

	// Should have 3 component declarations
	if len(node.Children) != 3 {
		t.Errorf("Expected 3 component declarations, got %d", len(node.Children))
	}
}

func TestParseInvalidSyntax(t *testing.T) {
	// Invalid templ syntax should return an error
	input := `package main

templ invalid( {
	<div>Missing closing paren</div>
}
`

	_, _, err := ParseBytes("test.templ", []byte(input))
	if err == nil {
		t.Error("Expected error for invalid templ syntax")
	}
}

func TestParseNonexistentFile(t *testing.T) {
	_, err := Parse("nonexistent_file.templ")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestParseWithLineCountNonexistent(t *testing.T) {
	_, _, err := ParseWithLineCount("nonexistent_file.templ")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}

func TestParseBytesWithComplexAttributes(t *testing.T) {
	input := `package main

templ complex(id string, data map[string]any) {
	<div
		id={ id }
		class={ "container " + id }
		data-attrs={ data }
		if true {
			enabled?={ true }
		}
	>
		Content
	</div>
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}
}

func TestParseConditionalAttribute(t *testing.T) {
	input := `package main

templ conditionalAttr(show bool, name string) {
	<div if show {
		class={ name }
	}>Content</div>
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}
}

func TestParseBoolAttributes(t *testing.T) {
	input := `package main

templ boolAttrs(disabled bool, checked bool) {
	<input disabled?={ disabled } checked?={ checked } />
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}
}

func TestParseConstantBoolAttribute(t *testing.T) {
	input := `package main

templ constBoolAttr() {
	<input disabled checked />
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}
}

func TestParseCSSExpressionProperty(t *testing.T) {
	input := `package main

css dynamicStyle(color string) {
	color: color;
	background: #fff;
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}
}

func TestParseGoExpression(t *testing.T) {
	input := `package main

var globalVar = "test"

templ withGlobal() {
	<div>{ globalVar }</div>
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}
}

func TestParseTemplRenderCall(t *testing.T) {
	input := `package main

templ child() {
	<span>Child</span>
}

templ parent() {
	<div>
		@child()
	</div>
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}
}

func TestParseCallTemplateExpression(t *testing.T) {
	input := `package main

templ partial() {
	<span>Partial</span>
}

templ main() {
	!partial()
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}
}

func TestParseFallthrough(t *testing.T) {
	input := `package main

templ withFallthrough(val int) {
	switch val {
		case 1:
			<p>One</p>
			fallthrough
		case 2:
			<p>Two</p>
	}
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}
}

func TestParseComments(t *testing.T) {
	input := `package main

// This is a Go comment
templ withComments() {
	<!-- HTML comment -->
	<div>
		// Another Go comment
		<span>Content</span>
	</div>
}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}
}

func TestParseWhitespace(t *testing.T) {
	input := `package main

templ withWhitespace() {



	<div>Content</div>


}
`

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}
}

// countAllNodes recursively counts all nodes in a tree.
func countAllNodes(n *syntax.Node) int {
	if n == nil {
		return 0
	}

	count := 1
	for _, child := range n.Children {
		count += countAllNodes(child)
	}

	return count
}
