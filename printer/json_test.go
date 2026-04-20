package printer

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax"
)

func TestJSONPrinter_PrintHeader(t *testing.T) {
	var buf bytes.Buffer

	printer := NewJSON(&buf, mockReadFile(""))

	err := printer.PrintHeader()
	if err != nil {
		t.Fatalf("PrintHeader failed: %v", err)
	}

	jsonPrinter := printer.(*JSONPrinter)
	testutil.AssertFieldValue(t, jsonPrinter.iota, 0, "iota")

	if jsonPrinter.filesCount != 0 {
		t.Errorf("Expected filesCount to be 0, got %d", jsonPrinter.filesCount)
	}

	if jsonPrinter.totalClones != 0 {
		t.Errorf("Expected totalClones to be 0, got %d", jsonPrinter.totalClones)
	}
}

func TestJSONPrinter_PrintClones(t *testing.T) {
	// Create test file content
	testContent := `package main

import "fmt"

func foo() {
	fmt.Println("hello")
}

func bar() {
	fmt.Println("hello")
}
`

	var buf bytes.Buffer

	printer := NewJSON(&buf, mockReadFile(testContent))

	// Create mock duplicate nodes
	nodes := createMockNodes(t)
	dups := [][]*syntax.Node{nodes}

	err := printer.PrintClones(dups)
	if err != nil {
		t.Fatalf("PrintClones failed: %v", err)
	}

	jsonPrinter := printer.(*JSONPrinter)
	testutil.AssertFieldValue(t, jsonPrinter.iota, 1, "iota")

	testutil.AssertFieldValue(t, len(jsonPrinter.cloneGroups), 1, "clone groups")
}

func TestJSONPrinter_OutputJSON(t *testing.T) {
	testContent := `package main

func foo() {
	return "test"
}
`

	var buf bytes.Buffer

	printer := NewJSON(&buf, mockReadFile(testContent))

	// Add some mock data
	jsonPrinter := printer.(*JSONPrinter)
	jsonPrinter.cloneGroups = []CloneGroup{
		{
			Hash: "test-hash",
			Size: 100,
			Files: []JSONClone{
				{
					Filename:  "test.go",
					LineStart: 1,
					LineEnd:   5,
					Fragment:  "test fragment",
				},
			},
		},
	}
	jsonPrinter.totalClones = 1
	jsonPrinter.filesCount = 1

	err := jsonPrinter.OutputJSON(15, "size", "")
	if err != nil {
		t.Fatalf("OutputJSON failed: %v", err)
	}

	output := buf.String()
	assertContains(t, output, "\"version\":", "JSON output missing version field")
	assertContains(t, output, "\"clone_groups\":", "JSON output missing clone_groups field")
	assertContains(t, output, "test-hash", "JSON output missing test hash")
}

func TestJSONPrinter_EmptyOutput(t *testing.T) {
	var buf bytes.Buffer

	printer := NewJSON(&buf, mockReadFile(""))

	// Properly initialize the printer
	err := printer.PrintHeader()
	if err != nil {
		t.Fatalf("PrintHeader failed: %v", err)
	}

	jsonPrinter := printer.(*JSONPrinter)

	err = jsonPrinter.OutputJSON(15, "size", "")
	if err != nil {
		t.Fatalf("OutputJSON with empty data failed: %v", err)
	}

	// Parse JSON to validate structure
	var output JSONOutput

	err = json.Unmarshal(buf.Bytes(), &output)
	if err != nil {
		t.Fatalf("Failed to parse JSON output: %v", err)
	}

	if len(output.CloneGroups) != 0 {
		t.Errorf("Expected 0 clone groups, got %d", len(output.CloneGroups))
	}

	if output.Summary.TotalCloneGroups != 0 {
		t.Errorf("Expected 0 total clone groups, got %d", output.Summary.TotalCloneGroups)
	}

	if output.Summary.TotalClones != 0 {
		t.Errorf("Expected 0 total clones, got %d", output.Summary.TotalClones)
	}
}

func createMockNodes(t *testing.T) []*syntax.Node {
	t.Helper()
	// Create a simple mock node structure
	// In practice, these would be real AST nodes from parsed Go code
	nodes := testutil.CreateMockNodes(2, "test.go")

	// Adjust positions for specific test requirements
	nodes[0].Pos = 10
	nodes[0].End = 30
	nodes[1].Pos = 40
	nodes[1].End = 60

	return nodes
}

func TestJSONPrinter_SetHash(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewJSON(&buf, mockReadFile(""))
	jp := p.(*JSONPrinter)

	jp.SetHash("abc123")

	if jp.currentHash != "abc123" {
		t.Errorf("currentHash = %q, want %q", jp.currentHash, "abc123")
	}
}

func TestJSONPrinter_SetFilesCount(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewJSON(&buf, mockReadFile(""))
	jp := p.(*JSONPrinter)

	jp.SetFilesCount(42)

	if jp.filesCount != 42 {
		t.Errorf("filesCount = %d, want 42", jp.filesCount)
	}
}

func TestJSONPrinter_PrintFooter(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewJSON(&buf, mockReadFile(""))

	err := p.PrintFooter()
	if err != nil {
		t.Fatalf("PrintFooter() error: %v", err)
	}
}

func TestJSONPrinter_PrintClones_WithHash(t *testing.T) {
	t.Parallel()

	testContent := "package main\n\nfunc foo() {\n\tfmt.Println(\"hello\")\n}\n"

	var buf bytes.Buffer

	p := NewJSON(&buf, mockReadFile(testContent))
	jp := p.(*JSONPrinter)
	jp.SetHash("deadbeef")

	nodes := createMockNodes(t)
	dups := [][]*syntax.Node{nodes}

	err := p.PrintClones(dups)
	if err != nil {
		t.Fatalf("PrintClones() error: %v", err)
	}

	if len(jp.cloneGroups) != 1 {
		t.Fatalf("clone groups = %d, want 1", len(jp.cloneGroups))
	}

	if jp.cloneGroups[0].Hash != "deadbeef" {
		t.Errorf("hash = %q, want %q", jp.cloneGroups[0].Hash, "deadbeef")
	}
}

func TestJSONPrinter_OutputSimpleJSON(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewJSON(&buf, mockReadFile(""))
	jp := p.(*JSONPrinter)

	jp.cloneGroups = []CloneGroup{
		{
			Hash: "abc",
			Size: 10,
			Files: []JSONClone{
				{Filename: "a.go", LineStart: 1, LineEnd: 5, Fragment: "code"},
				{Filename: "b.go", LineStart: 2, LineEnd: 6, Fragment: "code"},
			},
		},
	}
	jp.totalClones = 2

	err := jp.OutputSimpleJSON()
	if err != nil {
		t.Fatalf("OutputSimpleJSON() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, `"hash"`) {
		t.Errorf("OutputSimpleJSON() missing hash field: %q", output)
	}

	if !strings.Contains(output, `"abc"`) {
		t.Errorf("OutputSimpleJSON() missing hash value: %q", output)
	}

	if !strings.Contains(output, `"score"`) {
		t.Errorf("OutputSimpleJSON() missing score field: %q", output)
	}

	if !strings.Contains(output, `"instances"`) {
		t.Errorf("OutputSimpleJSON() missing instances field: %q", output)
	}
}

func TestJSONPrinter_OutputSimpleJSON_Empty(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewJSON(&buf, mockReadFile(""))
	jp := p.(*JSONPrinter)

	err := jp.OutputSimpleJSON()
	if err != nil {
		t.Fatalf("OutputSimpleJSON(empty) error: %v", err)
	}

	output := buf.String()

	if output != "[]" && output != "[\n]" {
		t.Errorf("OutputSimpleJSON(empty) = %q, want empty array", output)
	}
}

func TestJSONPrinter_PrintClones_MultipleGroups(t *testing.T) {
	t.Parallel()

	testContent := "package main\n\nfunc foo() {\n\tfmt.Println(\"hello\")\n}\n"

	var buf bytes.Buffer

	p := NewJSON(&buf, mockReadFile(testContent))
	jp := p.(*JSONPrinter)

	nodes := createMockNodes(t)

	err := p.PrintClones([][]*syntax.Node{nodes})
	if err != nil {
		t.Fatalf("PrintClones(1) error: %v", err)
	}

	err = p.PrintClones([][]*syntax.Node{nodes})
	if err != nil {
		t.Fatalf("PrintClones(2) error: %v", err)
	}

	if len(jp.cloneGroups) != 2 {
		t.Errorf("clone groups = %d, want 2", len(jp.cloneGroups))
	}

	if jp.totalClones != 2 {
		t.Errorf("totalClones = %d, want 2", jp.totalClones)
	}
}

func mockReadFile(content string) ReadFile {
	return func(filename string) ([]byte, error) {
		return []byte(content), nil
	}
}
