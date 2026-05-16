package printer

import (
	"bytes"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
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

func TestPlumbing_PrintHeaderAndFooter(t *testing.T) {
	t.Parallel()

	for _, tc := range printerMethodChecks {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			p := NewPlumbing(&buf, mockReadFile(""))

			err := tc.call(p)
			testutil.AssertFatalNoError(t, err, tc.name+"()")

			if buf.Len() != 0 {
				t.Errorf("%s() wrote %d bytes, want 0", tc.name, buf.Len())
			}
		})
	}
}

func TestPlumbing_PrintClones(t *testing.T) {
	t.Parallel()

	content := testPlumbCode

	var buf bytes.Buffer

	p := NewPlumbing(&buf, mockReadFile(content))

	nodes := createMockNodes(t)
	dups := [][]*syntax.Node{nodes}

	err := p.PrintClones(processTestNodes(mockReadFile(content), "test", dups))
	if err != nil {
		t.Fatalf("PrintClones() error: %v", err)
	}

	output := buf.String()

	testutil.AssertStringContains(t, output, "test.go", "PrintClones() output missing filename")

	testutil.AssertStringContains(t, output, ":", "PrintClones() output missing colon separator")
}

func TestPlumbing_PrintClones_SortedBySize(t *testing.T) {
	t.Parallel()

	content := "package main\n\nfunc foo() {\n}\nfunc bar() {\n}\n"

	var buf bytes.Buffer

	p := NewPlumbing(&buf, mockReadFile(content))

	nodes := createMockNodes(t)
	dups := [][]*syntax.Node{nodes}

	err := p.PrintClones(processTestNodes(mockReadFile(content), "test", dups), config.SortBySize)
	if err != nil {
		t.Fatalf("PrintClones(config.SortBySize) error: %v", err)
	}

	output := buf.String()

	testutil.AssertStringContains(
		t,
		output,
		"test.go",
		"PrintClones(sorted) output missing filename",
	)
}

func TestPlumbing_PrintClones_ReadError(t *testing.T) {
	t.Parallel()

	fread := errorReadFile("file not found")

	nodes := createMockNodes(t)
	dups := [][]*syntax.Node{nodes}

	_, err := NodesToGroup(fread, "test", dups)
	if err == nil {
		t.Error("PrintClones() should return error on read failure")
	}
}

func TestPlumbing_PrintClones_EmptyDups(t *testing.T) {
	t.Parallel()

	_, err := NodesToGroup(mockReadFile(""), "test", [][]*syntax.Node{{}})
	if err == nil {
		t.Error("PrintClones(empty dup) should return error")
	}
}

func TestPlumbing_PrintClones_MultipleGroups(t *testing.T) {
	t.Parallel()

	content := "package main\n\nfunc foo() {\n}\nfunc bar() {\n}\nfunc baz() {\n}\n"

	var buf bytes.Buffer

	p := NewPlumbing(&buf, mockReadFile(content))

	group1 := testutil.CreateSingleNode("a.go", 0, 20)
	group2 := testutil.CreateSingleNode("b.go", 25, 50)
	dups := [][]*syntax.Node{group1, group2}

	err := p.PrintClones(processTestNodes(mockReadFile(content), "test", dups))
	if err != nil {
		t.Fatalf("PrintClones(multiple) error: %v", err)
	}

	output := buf.String()

	testutil.AssertStringContains(t, output, "a.go", "output missing a.go")

	testutil.AssertStringContains(t, output, "b.go", "output missing b.go")
}

func TestPlumbing_OutputPlumbing(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewPlumbing(&buf, mockReadFile(""))

	pl := p.(*plumbing)

	err := pl.OutputPlumbing(15, config.SortBySize)
	if err != nil {
		t.Fatalf("OutputPlumbing() error: %v", err)
	}

	output := buf.String()

	testutil.AssertStringContains(
		t,
		output,
		"size",
		"OutputPlumbing() output missing sort criteria",
	)

	testutil.AssertStringContains(
		t,
		output,
		"# Plumbing output sorted by",
		"OutputPlumbing() output missing header",
	)
}
