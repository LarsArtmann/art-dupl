package printer

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func TestJSONPrinter_PrintHeader(t *testing.T) {
	var buf bytes.Buffer
	printer := NewJSON(&buf, mockReadFile(""))

	err := printer.PrintHeader()
	if err != nil {
		t.Fatalf("PrintHeader failed: %v", err)
	}

	jsonPrinter := printer.(*JSONPrinter)
	if jsonPrinter.iota != 0 {
		t.Errorf("Expected iota to be 0, got %d", jsonPrinter.iota)
	}
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
	if jsonPrinter.iota != 1 {
		t.Errorf("Expected iota to be 1, got %d", jsonPrinter.iota)
	}
	if len(jsonPrinter.cloneGroups) != 1 {
		t.Errorf("Expected 1 clone group, got %d", len(jsonPrinter.cloneGroups))
	}
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

	err := jsonPrinter.OutputJSON(15, "size")
	if err != nil {
		t.Fatalf("OutputJSON failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "\"version\":") {
		t.Error("JSON output missing version field")
	}
	if !strings.Contains(output, "\"clone_groups\":") {
		t.Error("JSON output missing clone_groups field")
	}
	if !strings.Contains(output, "test-hash") {
		t.Error("JSON output missing test hash")
	}
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
	err = jsonPrinter.OutputJSON(15, "size")
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
	nodes := make([]*syntax.Node, 2)

	nodes[0] = &syntax.Node{
		Type:     golang.FuncDecl,
		Filename: "test.go",
		Pos:      10,
		End:      30,
	}

	nodes[1] = &syntax.Node{
		Type:     golang.FuncDecl,
		Filename: "test.go",
		Pos:      40,
		End:      60,
	}

	return nodes
}

func mockReadFile(content string) ReadFile {
	return func(filename string) ([]byte, error) {
		return []byte(content), nil
	}
}
