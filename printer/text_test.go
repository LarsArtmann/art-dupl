package printer

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// errorReadFile creates a ReadFile that always returns the given error.
func errorReadFile(msg string) ReadFile {
	return func(filename string) ([]byte, error) {
		return nil, errors.New(msg)
	}
}

// mustPrintClones calls PrintClones and fails the test if there's an error.
func mustPrintClones(t *testing.T, p Printer, dups [][]*syntax.Node) {
	t.Helper()

	if err := p.PrintClones(dups); err != nil {
		t.Fatalf("PrintClones() error: %v", err)
	}
}

// mustPrintHeader calls PrintHeader and fails the test if there's an error.
func mustPrintHeader(t *testing.T, p Printer) {
	t.Helper()

	if err := p.PrintHeader(); err != nil {
		t.Fatalf("PrintHeader failed: %v", err)
	}
}

// mustPrintFooter calls PrintFooter and fails the test if there's an error.
func mustPrintFooter(t *testing.T, p Printer) {
	t.Helper()

	if err := p.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}
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

	tests := []struct {
		name string
		call func(p Printer) error
	}{
		{"PrintHeader", func(p Printer) error { return p.PrintHeader() }},
		{"PrintFooter", func(p Printer) error { return p.PrintFooter() }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			p := NewText(&buf, mockReadFile(""))

			err := tc.call(p)
	if err != nil {
		t.Fatalf("%s() error: %v", tc.name, err)
	}

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
	dups := [][]*syntax.Node{nodes}

	err := p.PrintClones(dups)
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

	node := &syntax.Node{Filename: "test.go", Pos: 0, End: 5}
	dups := [][]*syntax.Node{{node}}

	err := p.PrintClones(dups)
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

	var buf bytes.Buffer

	fread := errorReadFile("file not found")

	p := NewText(&buf, fread)

	nodes := createMockNodes(t)
	dups := [][]*syntax.Node{nodes}

	err := p.PrintClones(dups)
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
	dups := [][]*syntax.Node{nodes}

	err := p.(*TextPrinter).PrintClonesSorted(dups, config.SortBySize)
	if err != nil {
		t.Fatalf("PrintClonesSorted() error: %v", err)
	}

	output := buf.String()

	testutil.AssertStringContains(
		t,
		output,
		"sorted by size",
		"PrintClonesSorted() missing sort info",
	)
}

func TestDetectFileDuplicate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		flag bool
		dups [][]*syntax.Node
		want bool
	}{
		{
			name: "flag set",
			flag: true,
			dups: nil,
			want: true,
		},
		{
			name: "empty dups",
			flag: false,
			dups: nil,
			want: false,
		},
		{
			name: "pos zero single node",
			flag: false,
			dups: [][]*syntax.Node{{{Pos: 0}}},
			want: true,
		},
		{
			name: "pos nonzero",
			flag: false,
			dups: [][]*syntax.Node{{{Pos: 5}}},
			want: false,
		},
		{
			name: "multiple nodes in frag",
			flag: false,
			dups: [][]*syntax.Node{{{Pos: 0}, {Pos: 5}}},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := detectFileDuplicate(tc.flag, tc.dups)
			if got != tc.want {
				t.Errorf("detectFileDuplicate() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestCalculateCloneSizes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		clones []clone
		want   int
	}{
		{"nil", nil, 0},
		{"two fragments", []clone{
			{fragment: []byte("hello")},
			{fragment: []byte("world!")},
		}, 11},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := calculateCloneSizes(tc.clones); got != tc.want {
				t.Errorf("calculateCloneSizes() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestPrepareClonesInfo(t *testing.T) {
	t.Parallel()

	content := testPlumbCode

	fread := mockReadFile(content)

	node1 := &syntax.Node{Filename: "test.go", Pos: 15, End: 40}
	node2 := &syntax.Node{Filename: "test.go", Pos: 40, End: 50}
	dups := [][]*syntax.Node{{node1, node2}}

	clones, err := prepareClonesInfo(fread, dups)
	if err != nil {
		t.Fatalf("prepareClonesInfo() error: %v", err)
	}

	if len(clones) != 1 {
		t.Fatalf("prepareClonesInfo() returned %d clones, want 1", len(clones))
	}

	if clones[0].filename != "test.go" {
		t.Errorf("filename = %q, want %q", clones[0].filename, "test.go")
	}

	if clones[0].lineStart < 1 {
		t.Errorf("lineStart = %d, want >= 1", clones[0].lineStart)
	}
}

func TestPrepareClonesInfo_EmptyDup(t *testing.T) {
	t.Parallel()

	fread := mockReadFile("")

	dups := [][]*syntax.Node{{}}

	_, err := prepareClonesInfo(fread, dups)
	if err == nil {
		t.Error("prepareClonesInfo(empty dup) should return error")
	}
}

func TestPrepareClonesInfo_ReadError(t *testing.T) {
	t.Parallel()

	fread := errorReadFile("read error")

	testNode := &syntax.Node{Filename: "missing.go", Pos: 0, End: 10}
	dups := [][]*syntax.Node{{testNode}}

	_, err := prepareClonesInfo(fread, dups)
	if err == nil {
		t.Error("prepareClonesInfo() should return error on read failure")
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

	testNode := &syntax.Node{Filename: "test.go", Pos: 15, End: 40}
	dups := [][]*syntax.Node{{testNode}}

	clones, err := prepareClonesInfo(tp.ReadFile, dups)
	if err != nil {
		t.Fatalf("prepareClonesInfo() error: %v", err)
	}

	tp.cloneGroups = [][]clone{clones}

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

	clones := []clone{
		{filename: "a.go", lineStart: 1, lineEnd: 3, fragment: []byte("code")},
		{filename: "b.go", lineStart: 1, lineEnd: 3, fragment: []byte("code")},
	}
	tp.cloneGroups = [][]clone{clones}

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

			clones := []clone{
				{filename: tc.filename, lineStart: 1, lineEnd: 3, fragment: []byte("code")},
			}
			tp.cloneGroups = [][]clone{clones}

			err := tp.OutputText(15, tc.sortBy)
	if err != nil {
		t.Fatalf("%s() error: %v", tc.name, err)
	}
		})
	}
}

func TestTextPrinter_OutputText_PrintFooterError(t *testing.T) {
	t.Parallel()

	content := testPlumbCode

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(content))
	tp := p.(*TextPrinter)

	testNode := &syntax.Node{Filename: "test.go", Pos: 15, End: 40}
	dups := [][]*syntax.Node{{testNode}}

	clones, err := prepareClonesInfo(tp.ReadFile, dups)
	if err != nil {
		t.Fatalf("prepareClonesInfo() error: %v", err)
	}

	tp.cloneGroups = [][]clone{clones}

	fw := &firstWriteFailsWriter{w: &buf}
	tp.w = fw

	err = tp.OutputText(15, config.SortBySize)
	if err == nil {
		t.Error("expected error when PrintFooter write fails")
	}
}

func TestTextPrinter_PrintClonesSorted_ReadError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	fread := errorReadFile("simulated read error")

	p := NewText(&buf, fread)

	testNode := &syntax.Node{Filename: "missing.go", Pos: 0, End: 10}
	dups := [][]*syntax.Node{{testNode}}

	err := p.(*TextPrinter).PrintClonesSorted(dups, config.SortBySize)
	if err == nil {
		t.Error("expected error on read failure in PrintClonesSorted")
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
			"sorted by occurrence",
		},
		{"SortByHash", "package main\n\nfunc foo() {}\n", config.SortByHash, "sorted by hash"},
		{"UnknownSort", "package main\n\nfunc foo() {}\n", config.SortCriteria("unknown"), ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			p := NewText(&buf, mockReadFile(tc.content))

			nodes := createMockNodes(t)
			dups := [][]*syntax.Node{nodes}

			err := p.(*TextPrinter).PrintClonesSorted(dups, tc.sortBy)
	if err != nil {
		t.Fatalf("%s() error: %v", tc.name, err)
	}

			if tc.wantSubstr != "" {
				testutil.AssertStringContains(
					t,
					buf.String(),
					tc.wantSubstr,
					tc.name+" missing sort info",
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

	clones := []clone{
		{filename: "a.go", lineStart: 1, lineEnd: 3, fragment: nil},
	}
	tp.cloneGroups = [][]clone{clones}

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
