package templ

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

func assertParseBytesSuccess(t *testing.T, input string) *syntax.Node {
	t.Helper()

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}

	return node
}

func runParseTests(t *testing.T, tests []struct {
	name  string
	input string
},
) {
	t.Helper()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertParseBytesSuccess(t, tt.input)
		})
	}
}

func TestParseBytes(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantNodes bool
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

	testParseAndVerifyNodeCount(t, input, 1)
}

func TestParseWithLineCountFile(t *testing.T) {
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

	testParseTemplInputMin(t, input, 1)
}

func TestParseScriptTemplate(t *testing.T) {
	input := `package main

script onClick() {
	console.log("clicked");
}
`

	testParseTemplInputMin(t, input, 1)
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

	testParseAndVerifyNodeCount(t, input, 3)
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

	testParseTemplInputMin(t, input, 2)
}

func testParseTemplInputMin(t *testing.T, input string, expectedMin int) {
	t.Helper()

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}

	if len(node.Children) < expectedMin {
		t.Errorf(
			"Expected at least %d component declarations, got %d",
			expectedMin,
			len(node.Children),
		)
	}
}

func testParseAndVerifyNodeCount(t *testing.T, input string, expectedMin int) {
	t.Helper()

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}

	nodeCount := countAllNodes(node)
	if nodeCount < expectedMin {
		t.Errorf("Expected at least %d nodes, got %d", expectedMin, nodeCount)
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

	testParseAndVerifyNodeCount(t, input, 3)
}

func TestParseExpressionAttributes(t *testing.T) {
	input := `package main

templ attrs(name string, active bool) {
	<div class={ name } disabled?={ active } data-id={ "test" }>Content</div>
}
`

	testParseAndVerifyNodeCount(t, input, 2)
}

func TestParseValidTemplates(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "doctype",
			input: `package main

templ page() {
	<!DOCTYPE html>
	<html>
		<body>Content</body>
	</html>
}
`,
		},
		{
			name: "script_element",
			input: `package main

templ withScript() {
	<div>
		<script type="text/javascript">
			console.log("hello");
		</script>
	</div>
}
`,
		},
		{
			name: "raw_element",
			input: `package main

templ withRaw() {
	<div>
		<style>
			.test { color: red; }
		</style>
	</div>
}
`,
		},
		{
			name: "children_expression",
			input: `package main

templ wrapper() {
	<div class="wrapper">
		{ children... }
	</div>
}
`,
		},
		{
			name: "go_code",
			input: `package main

templ withGoCode() {
	<div>
		{ fmt.Sprintf("Hello %s", "World") }
	</div>
}
`,
		},
		{
			name: "complex_attributes",
			input: `package main

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
`,
		},
		{
			name: "conditional_attribute",
			input: `package main

templ conditionalAttr(show bool, name string) {
	<div if show {
		class={ name }
	}>Content</div>
}
`,
		},
		{
			name: "bool_attributes",
			input: `package main

templ boolAttrs(disabled bool, checked bool) {
	<input disabled?={ disabled } checked?={ checked } />
}
`,
		},
		{
			name: "constant_bool_attribute",
			input: `package main

templ constBoolAttr() {
	<input disabled checked />
}
`,
		},
		{
			name: "css_expression_property",
			input: `package main

css dynamicStyle(color string) {
	color: color;
	background: #fff;
}
`,
		},
		{
			name: "go_expression",
			input: `package main

var globalVar = "test"

templ withGlobal() {
	<div>{ globalVar }</div>
}
`,
		},
		{
			name: "templ_render_call",
			input: `package main

templ child() {
	<span>Child</span>
}

templ parent() {
	<div>
		@child()
	</div>
}
`,
		},
		{
			name: "call_template_expression",
			input: `package main

templ partial() {
	<span>Partial</span>
}

templ main() {
	!partial()
}
`,
		},
		{
			name: "fallthrough",
			input: `package main

templ withFallthrough(val int) {
	switch val {
		case 1:
			<p>One</p>
			fallthrough
		case 2:
			<p>Two</p>
	}
}
`,
		},
		{
			name: "comments",
			input: `package main

// This is a Go comment
templ withComments() {
	<!-- HTML comment -->
	<div>
		// Another Go comment
		<span>Content</span>
	</div>
}
`,
		},
		{
			name: "whitespace",
			input: `package main

templ withWhitespace() {



	<div>Content</div>


}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testParseValid(t, tt.input)
		})
	}
}

func testParseValid(t *testing.T, input string) {
	t.Helper()

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

	testParseAndVerifyNodeCount(t, input, 5)
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

	testParseTemplInputExact(t, input, 3)
}

func testParseTemplInputExact(t *testing.T, input string, expected int) {
	t.Helper()

	node, _, err := ParseBytes("test.templ", []byte(input))
	if err != nil {
		t.Fatalf("ParseBytes() error = %v", err)
	}

	if node == nil {
		t.Fatal("ParseBytes() returned nil node")
	}

	if len(node.Children) != expected {
		t.Errorf("Expected %d component declarations, got %d", expected, len(node.Children))
	}
}

func TestParseInvalidSyntax(t *testing.T) {
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

// TestParseChildrenCount tests parsing of various templ constructs with expected children counts.
func TestParseChildrenCount(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedChildren int
	}{
		{
			name: "call_template_expression",
			input: `package main

templ greeting(name string) {
	<p>Hello, { name }!</p>
}

templ page() {
	<div>
		!greeting("World")
	</div>
}
`,
			expectedChildren: 2,
		},
		{
			name: "switch_with_default",
			input: `package main

templ withDefault(value int) {
	switch value {
		case 1:
			<p>One</p>
		case 2:
			<p>Two</p>
		default:
			<p>Other</p>
	}
}
`,
			expectedChildren: 1,
		},
		{
			name: "complex_script_template",
			input: `package main

script complexHandler(event string) {
	console.log("Event:", event);
	const data = { event: event, timestamp: Date.now() };
	fetch('/api/log', { method: 'POST', body: JSON.stringify(data) });
}

templ button() {
	<button onClick={ complexHandler("click") }>Click me</button>
}
`,
			expectedChildren: 2,
		},
		{
			name: "mixed_content",
			input: `package main

import "fmt"

css sharedStyles() {
	background-color: #ffffff;
	color: { "blue" };
}

script analytics(event string) {
	console.log(event);
}

templ Layout(title string) {
	<!DOCTYPE html>
	<html>
		<head>
			<title>{ title }</title>
		</head>
		<body>
			{ children... }
		</body>
	</html>
}

templ Button(text string) {
	<button onClick={ analytics("click") }>
		{ text }
	</button>
}
`,
			expectedChildren: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := assertParseBytesSuccess(t, tt.input)
			if len(node.Children) != tt.expectedChildren {
				t.Errorf("Expected %d children, got %d", tt.expectedChildren, len(node.Children))
			}
		})
	}
}

// TestParseGoCode tests parsing of Go code blocks inside templ components.
func TestParseGoCode(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "variable_declaration",
			input: `package main

templ withVar() {
	{{ x := 42 }}
	<div>{ fmt.Sprintf("%d", x) }</div>
}
`,
		},
		{
			name: "multiple_statements",
			input: `package main

templ withMultipleVars() {
	{{ name := "test"; count := 5 }}
	<div>{ name } - { fmt.Sprintf("%d", count) }</div>
}
`,
		},
	}

	runParseTests(t, tests)
}

// TestParseComplexAttributes tests various attribute types.
func TestParseComplexAttributes(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "spread_attributes",
			input: `package main

templ withSpread(attrs templ.Attributes) {
	<div { attrs... }>Content</div>
}
`,
		},
		{
			name: "conditional_attribute",
			input: `package main

templ withConditional(show bool) {
	<div if show { class="visible" }>Content</div>
}
`,
		},
		{
			name: "multiple_attributes",
			input: `package main

templ withMultiple(id string, active bool) {
	<div id={ id } class="item" disabled?={ active } data-test="value">Content</div>
}
`,
		},
	}

	runParseTests(t, tests)
}

// TestParseEdgeCases tests edge cases and unusual but valid syntax.
func TestParseEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name: "empty_file_except_package",
			input: `package main
`,
		},
		{
			name: "only_imports",
			input: `package main

import "fmt"
import "strings"
`,
		},
		{
			name: "self_closing_tags",
			input: `package main

templ withSelfClosing() {
	<input type="text" />
	<br />
	<hr />
}
`,
		},
		{
			name: "deeply_nested",
			input: `package main

templ deeplyNested() {
	<div>
		<div>
			<div>
				<div>
					<div>
						<span>Deep</span>
					</div>
				</div>
			</div>
		</div>
	</div>
}
`,
		},
		{
			name: "templ_in_expression",
			input: `package main

templ component() {
	<div class={ templ.SafeCSS("color: red;") }>Styled</div>
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, _, err := ParseBytes("test.templ", []byte(tt.input))
			if err != nil {
				// Empty file and imports-only are valid
				t.Fatalf("ParseBytes() error = %v", err)
			}

			if node == nil {
				t.Fatal("ParseBytes() returned nil node")
			}
		})
	}
}
