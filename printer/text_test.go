package printer

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// errorReadFile creates a ReadFile that always returns the given error.
func errorReadFile(msg string) ReadFile {
	return func(filename string) ([]byte, error) {
		return nil, errors.New(msg)
	}
}

// mustPrintClones calls PrintClones with a ProcessedCloneGroup and fails the test if there's an error.
func mustPrintClones(t *testing.T, p Printer, dups [][]*syntax.Node) {
	t.Helper()

	fread := mockReadFile(testPlumbCode)

	group, err := NodesToGroup(fread, "test", dups)
	if err != nil {
		t.Fatalf("NodesToGroup() error: %v", err)
	}

	if err := p.PrintClones(group); err != nil {
		t.Fatalf("PrintClones() error: %v", err)
	}
}

// mustPrintHeader calls PrintHeader and fails the test if there's an error.
func mustPrintHeader(t *testing.T, p Printer) {
	t.Helper()

	testutil.AssertFatalNoError(t, p.PrintHeader(), "PrintHeader")
}

// mustPrintFooter calls PrintFooter and fails the test if there's an error.
func mustPrintFooter(t *testing.T, p Printer) {
	t.Helper()

	testutil.AssertFatalNoError(t, p.PrintFooter(), "PrintFooter")
}

func TestNewText(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(""))

	if p == nil {
		t.Fatal("NewText returned nil")
	}

	_, ok := p.(*TextPrinter)
	if !ok {
		t.Error("NewText should return *TextPrinter")
	}
}

func TestTextPrinter_PrintHeaderAndFooter(t *testing.T) {
	t.Parallel()

	for _, tc := range printerMethodChecks {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			p := NewText(&buf, mockReadFile(""))

			err := tc.call(p)
			testutil.AssertFatalNoError(t, err, tc.name+"()")

			if tc.name == "PrintHeader" && buf.Len() != 0 {
				t.Errorf("%s() wrote %d bytes, want 0", tc.name, buf.Len())
			}
		})
	}
}

func TestTextPrinter_PrintClones(t *testing.T) {
	t.Parallel()

	content := testPlumbCode

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(content))

	nodes := createMockNodes(t)
	group := processTestNodes(mockReadFile(content), "test", [][]*syntax.Node{nodes})

	err := p.PrintClones(group)
	if err != nil {
		t.Fatalf("PrintClones() error: %v", err)
	}

	output := buf.String()

	testutil.AssertStringContains(t, output, "found", "PrintClones() output missing 'found'")

	testutil.AssertStringContains(t, output, "clones", "PrintClones() output missing 'clones'")
}

func TestTextPrinter_PrintClones_FileDuplicate(t *testing.T) {
	t.Parallel()

	content := testPlumbCode

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(content))
	tp := p.(*TextPrinter)
	tp.SetHash("abcdef123456")
	tp.SetFileDuplicate(true)

	node := &syntax.Node{Filename: testFilename, Pos: 0, End: 5}
	group := processTestNodes(mockReadFile(content), "test", [][]*syntax.Node{{node}})

	err := p.PrintClones(group)
	if err != nil {
		t.Fatalf("PrintClones() error: %v", err)
	}

	output := buf.String()

	testutil.AssertStringContains(
		t,
		output,
		"FILE DUPLICATE",
		"PrintClones(file dupe) missing FILE DUPLICATE",
	)
}

func TestTextPrinter_PrintClones_ReadError(t *testing.T) {
	t.Parallel()

	fread := errorReadFile("file not found")

	nodes := createMockNodes(t)

	_, err := NodesToGroup(fread, "test", [][]*syntax.Node{nodes})
	if err == nil {
		t.Error("PrintClones() should return error on read failure")
	}
}

func TestTextPrinter_PrintFooter(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(""))

	err := p.PrintFooter()
	if err != nil {
		t.Fatalf("PrintFooter() error: %v", err)
	}

	output := buf.String()

	testutil.AssertStringContains(t, output, "Found total", "PrintFooter() missing summary")

	testutil.AssertStringContains(t, output, "clone groups", "PrintFooter() missing 'clone groups'")
}

func TestTextPrinter_PrintFooter_DiffHint(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(""))
	tp := p.(*TextPrinter)
	tp.diffHintFiles = []string{"a.go", "b.go"}

	err := p.PrintFooter()
	if err != nil {
		t.Fatalf("PrintFooter() error: %v", err)
	}

	output := buf.String()

	testutil.AssertStringContains(t, output, "diff a.go b.go", "PrintFooter() missing diff hint")
}

func TestTextPrinter_PrintClonesSorted(t *testing.T) {
	t.Parallel()

	content := "package main\n\nfunc foo() {\n}\nfunc bar() {\n}\n"

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(content))

	nodes := createMockNodes(t)
	group := processTestNodes(mockReadFile(content), "test", [][]*syntax.Node{nodes})

	err := p.PrintClones(group, config.SortBySize)
	if err != nil {
		t.Fatalf("PrintClones(sorted) error: %v", err)
	}

	output := buf.String()

	testutil.AssertStringContains(
		t,
		output,
		"found",
		"PrintClones(sorted) missing clone output",
	)
}

func TestCalculateProcessedCloneSizes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		clones []domain.ProcessedClone
		want   int
	}{
		{"nil", nil, 0},
		{"two fragments", []domain.ProcessedClone{
			{Fragment: []byte("hello")},
			{Fragment: []byte("world!")},
		}, 11},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := calculateProcessedCloneSizes(tc.clones); got != tc.want {
				t.Errorf("calculateProcessedCloneSizes() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestProcessClones(t *testing.T) {
	t.Parallel()

	content := testPlumbCode

	fread := mockReadFile(content)

	node1 := &syntax.Node{Filename: testFilename, Pos: 15, End: 40}
	node2 := &syntax.Node{Filename: testFilename, Pos: 40, End: 50}
	dups := [][]*syntax.Node{{node1, node2}}

	clones, err := ProcessClones(fread, dups)
	if err != nil {
		t.Fatalf("ProcessClones() error: %v", err)
	}

	if len(clones) != 1 {
		t.Fatalf("ProcessClones() returned %d clones, want 1", len(clones))
	}

	if clones[0].Filename != testFilename {
		t.Errorf("filename = %q, want %q", clones[0].Filename, testFilename)
	}

	if clones[0].LineStart < 1 {
		t.Errorf("lineStart = %d, want >= 1", clones[0].LineStart)
	}
}

func TestProcessClones_EmptyDup(t *testing.T) {
	t.Parallel()

	fread := mockReadFile("")

	dups := [][]*syntax.Node{{}}

	_, err := ProcessClones(fread, dups)
	if err == nil {
		t.Error("ProcessClones(empty dup) should return error")
	}
}

func TestProcessClones_ReadError(t *testing.T) {
	t.Parallel()

	fread := errorReadFile("read error")

	testNode := &syntax.Node{Filename: "missing.go", Pos: 0, End: 10}
	dups := [][]*syntax.Node{{testNode}}

	_, err := ProcessClones(fread, dups)
	if err == nil {
		t.Error("ProcessClones() should return error on read failure")
	}
}

func TestTextPrinter_SetHash(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(""))
	tp := p.(*TextPrinter)

	tp.SetHash("testhash123")

	if tp.currentHash != "testhash123" {
		t.Errorf("currentHash = %q, want %q", tp.currentHash, "testhash123")
	}
}

func TestTextPrinter_SetFileDuplicate(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(""))
	tp := p.(*TextPrinter)

	tp.SetFileDuplicate(true)

	if !tp.isFileDupe {
		t.Error("isFileDupe should be true")
	}

	tp.SetFileDuplicate(false)

	if tp.isFileDupe {
		t.Error("isFileDupe should be false")
	}
}

func TestTextPrinter_OutputText(t *testing.T) {
	t.Parallel()

	content := testPlumbCode

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(content))
	tp := p.(*TextPrinter)

	testNode := &syntax.Node{Filename: testFilename, Pos: 15, End: 40}
	dups := [][]*syntax.Node{{testNode}}

	clones, err := ProcessClones(tp.ReadFile, dups)
	if err != nil {
		t.Fatalf("ProcessClones() error: %v", err)
	}

	tp.cloneGroups = [][]domain.ProcessedClone{clones}

	err = tp.OutputText(15, config.SortBySize)
	if err != nil {
		t.Fatalf("OutputText() error: %v", err)
	}

	output := buf.String()

	testutil.AssertStringContains(t, output, "Found total", "OutputText() missing footer")
}

func TestTextPrinter_OutputText_Empty(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(""))
	tp := p.(*TextPrinter)

	err := tp.OutputText(15, config.SortBySize)
	if err != nil {
		t.Fatalf("OutputText(empty) error: %v", err)
	}

	output := buf.String()

	testutil.AssertStringContains(
		t,
		output,
		"Found total 0 clone groups",
		"OutputText(empty) missing summary",
	)
}

func TestTextPrinter_OutputText_SortByOccurrence(t *testing.T) {
	t.Parallel()

	content := "package main\n\nfunc foo() {}\nfunc bar() {}\n"

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(content))
	tp := p.(*TextPrinter)

	clones := []domain.ProcessedClone{
		{Filename: "a.go", LineStart: 1, LineEnd: 3, Fragment: []byte("code")},
		{Filename: "b.go", LineStart: 1, LineEnd: 3, Fragment: []byte("code")},
	}
	tp.cloneGroups = [][]domain.ProcessedClone{clones}

	err := tp.OutputText(15, config.SortByOccurrence)
	if err != nil {
		t.Fatalf("OutputText(config.SortByOccurrence) error: %v", err)
	}
}

func TestTextPrinter_OutputText_SortVariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		sortBy   config.SortCriteria
		filename string
	}{
		{"SortByHash", config.SortByHash, "b.go"},
		{"SortByTotalTokens", config.SortByTotalTokens, "a.go"},
		{"UnknownSort", config.SortCriteria("unknown_sort_value"), "a.go"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			p := NewText(&buf, mockReadFile(""))
			tp := p.(*TextPrinter)

			clones := []domain.ProcessedClone{
				{Filename: tc.filename, LineStart: 1, LineEnd: 3, Fragment: []byte("code")},
			}
			tp.cloneGroups = [][]domain.ProcessedClone{clones}

			err := tp.OutputText(15, tc.sortBy)
			testutil.AssertFatalNoError(t, err, tc.name+"()")
		})
	}
}

func TestTextPrinter_OutputText_PrintFooterError(t *testing.T) {
	t.Parallel()

	content := testPlumbCode

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(content))
	tp := p.(*TextPrinter)

	testNode := &syntax.Node{Filename: testFilename, Pos: 15, End: 40}
	dups := [][]*syntax.Node{{testNode}}

	clones, err := ProcessClones(tp.ReadFile, dups)
	if err != nil {
		t.Fatalf("ProcessClones() error: %v", err)
	}

	tp.cloneGroups = [][]domain.ProcessedClone{clones}

	fw := &firstWriteFailsWriter{w: &buf}
	tp.w = fw

	err = tp.OutputText(15, config.SortBySize)
	if err == nil {
		t.Error("expected error when PrintFooter write fails")
	}
}

func TestTextPrinter_PrintClonesSorted_ReadError(t *testing.T) {
	t.Parallel()

	fread := errorReadFile("simulated read error")

	node := &syntax.Node{Filename: "missing.go", Pos: 0, End: 10}

	_, err := NodesToGroup(fread, "test", [][]*syntax.Node{{node}})
	if err == nil {
		t.Error("expected error on read failure in PrintClones(sorted)")
	}
}

func TestTextPrinter_PrintClonesSorted_Variants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		content    string
		sortBy     config.SortCriteria
		wantSubstr string
	}{
		{
			"SortByOccurrence",
			"package main\n\nfunc foo() {}\nfunc bar() {}\n",
			config.SortByOccurrence,
			"found",
		},
		{"SortByHash", "package main\n\nfunc foo() {}\n", config.SortByHash, "found"},
		{"UnknownSort", "package main\n\nfunc foo() {}\n", config.SortCriteria("unknown"), "found"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			p := NewText(&buf, mockReadFile(tc.content))

			nodes := createMockNodes(t)
			group := processTestNodes(mockReadFile(tc.content), "test", [][]*syntax.Node{nodes})

			err := p.PrintClones(group, tc.sortBy)
			if err != nil {
				t.Fatalf("PrintClones(%s) error: %v", tc.sortBy, err)
			}

			if tc.wantSubstr != "" {
				testutil.AssertStringContains(
					t,
					buf.String(),
					tc.wantSubstr,
					tc.name+" missing "+tc.wantSubstr,
				)
			}
		})
	}
}

func TestTextPrinter_OutputText_EmptyFragmentWrite(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(""))
	tp := p.(*TextPrinter)

	clones := []domain.ProcessedClone{
		{Filename: "a.go", LineStart: 1, LineEnd: 3, Fragment: nil},
	}
	tp.cloneGroups = [][]domain.ProcessedClone{clones}

	err := tp.OutputText(15, config.SortBySize)
	if err != nil {
		t.Fatalf("OutputText(empty fragment) error: %v", err)
	}
}

// firstWriteFailsWriter fails on first write, succeeds thereafter.
type firstWriteFailsWriter struct {
	w     io.Writer
	calls int
}

func (fw *firstWriteFailsWriter) Write(p []byte) (int, error) {
	fw.calls++
	if fw.calls == 1 {
		return 0, errors.New("simulated write error")
	}

	return fw.w.Write(p)
}
