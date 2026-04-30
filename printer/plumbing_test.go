package printer

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

func TestNewPlumbing(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewPlumbing(&buf, mockReadFile(""))

	if p == nil {
		t.Fatal("NewPlumbing returned nil")
	}

	_, ok := p.(*plumbing)
	if !ok {
		t.Error("NewPlumbing should return *plumbing")
	}
}

func TestPlumbing_PrintHeader(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewPlumbing(&buf, mockReadFile(""))

	err := p.PrintHeader()
	if err != nil {
		t.Fatalf("PrintHeader() error: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("PrintHeader() wrote %d bytes, want 0", buf.Len())
	}
}

func TestPlumbing_PrintFooter(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewPlumbing(&buf, mockReadFile(""))

	err := p.PrintFooter()
	if err != nil {
		t.Fatalf("PrintFooter() error: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("PrintFooter() wrote %d bytes, want 0", buf.Len())
	}
}

func TestPlumbing_PrintClones(t *testing.T) {
	t.Parallel()

	content := testPlumbCode

	var buf bytes.Buffer

	p := NewPlumbing(&buf, mockReadFile(content))

	nodes := createMockNodes(t)
	dups := [][]*syntax.Node{nodes}

	err := p.PrintClones(dups)
	if err != nil {
		t.Fatalf("PrintClones() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "test.go") {
		t.Errorf("PrintClones() output missing filename: %q", output)
	}

	if !strings.Contains(output, ":") {
		t.Errorf("PrintClones() output missing colon separator: %q", output)
	}
}

func TestPlumbing_PrintClones_SortedBySize(t *testing.T) {
	t.Parallel()

	content := "package main\n\nfunc foo() {\n}\nfunc bar() {\n}\n"

	var buf bytes.Buffer

	p := NewPlumbing(&buf, mockReadFile(content))

	nodes := createMockNodes(t)
	dups := [][]*syntax.Node{nodes}

	err := p.PrintClones(dups, SortBySize)
	if err != nil {
		t.Fatalf("PrintClones(SortBySize) error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "test.go") {
		t.Errorf("PrintClones(sorted) output missing filename: %q", output)
	}
}

func TestPlumbing_PrintClones_ReadError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	fread := func(filename string) ([]byte, error) {
		return nil, errors.New("file not found")
	}

	p := NewPlumbing(&buf, fread)

	nodes := createMockNodes(t)
	dups := [][]*syntax.Node{nodes}

	err := p.PrintClones(dups)
	if err == nil {
		t.Error("PrintClones() should return error on read failure")
	}
}

func TestPlumbing_PrintClones_EmptyDups(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewPlumbing(&buf, mockReadFile(""))

	err := p.PrintClones([][]*syntax.Node{{}})
	if err == nil {
		t.Error("PrintClones(empty dup) should return error")
	}
}

func TestPlumbing_PrintClones_MultipleGroups(t *testing.T) {
	t.Parallel()

	content := "package main\n\nfunc foo() {\n}\nfunc bar() {\n}\nfunc baz() {\n}\n"

	var buf bytes.Buffer

	p := NewPlumbing(&buf, mockReadFile(content))

	group1 := testNodesAt("a.go", 0, 20)
	group2 := testNodesAt("b.go", 25, 50)
	dups := [][]*syntax.Node{group1, group2}

	err := p.PrintClones(dups)
	if err != nil {
		t.Fatalf("PrintClones(multiple) error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "a.go") {
		t.Errorf("output missing a.go: %q", output)
	}

	if !strings.Contains(output, "b.go") {
		t.Errorf("output missing b.go: %q", output)
	}
}

func TestPlumbing_OutputPlumbing(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewPlumbing(&buf, mockReadFile(""))

	pl := p.(*plumbing)
	err := pl.OutputPlumbing(15, SortBySize)
	if err != nil {
		t.Fatalf("OutputPlumbing() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "size") {
		t.Errorf("OutputPlumbing() output missing sort criteria: %q", output)
	}

	if !strings.Contains(output, "# Plumbing output sorted by") {
		t.Errorf("OutputPlumbing() output missing header: %q", output)
	}
}

func testNodesAt(filename string, pos, end int32) []*syntax.Node {
	return []*syntax.Node{
		{Filename: filename, Pos: pos, End: end},
	}
}
