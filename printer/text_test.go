package printer

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/syntax"
)

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

func TestTextPrinter_PrintHeader(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(""))

	err := p.PrintHeader()
	if err != nil {
		t.Fatalf("PrintHeader() error: %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("PrintHeader() wrote %d bytes, want 0", buf.Len())
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

	if !strings.Contains(output, "found") {
		t.Errorf("PrintClones() output missing 'found': %q", output)
	}

	if !strings.Contains(output, "clones") {
		t.Errorf("PrintClones() output missing 'clones': %q", output)
	}
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

	if !strings.Contains(output, "FILE DUPLICATE") {
		t.Errorf("PrintClones(file dupe) missing FILE DUPLICATE: %q", output)
	}
}

func TestTextPrinter_PrintClones_ReadError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	fread := func(filename string) ([]byte, error) {
		return nil, errors.New("file not found")
	}

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

	if !strings.Contains(output, "Found total") {
		t.Errorf("PrintFooter() missing summary: %q", output)
	}

	if !strings.Contains(output, "clone groups") {
		t.Errorf("PrintFooter() missing 'clone groups': %q", output)
	}
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

	if !strings.Contains(output, "diff a.go b.go") {
		t.Errorf("PrintFooter() missing diff hint: %q", output)
	}
}

func TestTextPrinter_PrintClonesSorted(t *testing.T) {
	t.Parallel()

	content := "package main\n\nfunc foo() {\n}\nfunc bar() {\n}\n"

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(content))

	nodes := createMockNodes(t)
	dups := [][]*syntax.Node{nodes}

	err := p.(*TextPrinter).PrintClonesSorted(dups, SortBySize)
	if err != nil {
		t.Fatalf("PrintClonesSorted() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "sorted by size") {
		t.Errorf("PrintClonesSorted() missing sort info: %q", output)
	}
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

	clones := []clone{
		{fragment: []byte("hello")},
		{fragment: []byte("world!")},
	}

	total := calculateCloneSizes(clones)

	if total != 11 {
		t.Errorf("calculateCloneSizes() = %d, want 11", total)
	}

	if clones[0].size != 5 {
		t.Errorf("clones[0].size = %d, want 5", clones[0].size)
	}

	if clones[1].size != 6 {
		t.Errorf("clones[1].size = %d, want 6", clones[1].size)
	}
}

func TestCalculateCloneSizes_Empty(t *testing.T) {
	t.Parallel()

	total := calculateCloneSizes(nil)
	if total != 0 {
		t.Errorf("calculateCloneSizes(nil) = %d, want 0", total)
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

	fread := func(filename string) ([]byte, error) {
		return nil, errors.New("read error")
	}

	node := &syntax.Node{Filename: "missing.go", Pos: 0, End: 10}
	dups := [][]*syntax.Node{{node}}

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

	node := &syntax.Node{Filename: "test.go", Pos: 15, End: 40}
	dups := [][]*syntax.Node{{node}}

	clones, err := prepareClonesInfo(tp.ReadFile, dups)
	if err != nil {
		t.Fatalf("prepareClonesInfo() error: %v", err)
	}

	tp.cloneGroups = [][]clone{clones}

	err = tp.OutputText(15, SortBySize)
	if err != nil {
		t.Fatalf("OutputText() error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Found total") {
		t.Errorf("OutputText() missing footer: %q", output)
	}
}

func TestTextPrinter_OutputText_Empty(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(""))
	tp := p.(*TextPrinter)

	err := tp.OutputText(15, SortBySize)
	if err != nil {
		t.Fatalf("OutputText(empty) error: %v", err)
	}

	output := buf.String()

	if !strings.Contains(output, "Found total 0 clone groups") {
		t.Errorf("OutputText(empty) missing summary: %q", output)
	}
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

	err := tp.OutputText(15, SortByOccurrence)
	if err != nil {
		t.Fatalf("OutputText(SortByOccurrence) error: %v", err)
	}
}

func TestTextPrinter_OutputText_SortByHash(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(""))
	tp := p.(*TextPrinter)

	clones := []clone{
		{filename: "b.go", lineStart: 1, lineEnd: 3, fragment: []byte("code")},
	}
	tp.cloneGroups = [][]clone{clones}

	err := tp.OutputText(15, SortByHash)
	if err != nil {
		t.Fatalf("OutputText(SortByHash) error: %v", err)
	}
}

func TestTextPrinter_OutputText_SortByTotalTokens(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(""))
	tp := p.(*TextPrinter)

	clones := []clone{
		{filename: "a.go", lineStart: 1, lineEnd: 3, fragment: []byte("code")},
	}
	tp.cloneGroups = [][]clone{clones}

	err := tp.OutputText(15, SortByTotalTokens)
	if err != nil {
		t.Fatalf("OutputText(SortByTotalTokens) error: %v", err)
	}
}

func TestTextPrinter_OutputText_UnknownSort(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(""))
	tp := p.(*TextPrinter)

	clones := []clone{
		{filename: "a.go", lineStart: 1, lineEnd: 3, fragment: []byte("code")},
	}
	tp.cloneGroups = [][]clone{clones}

	err := tp.OutputText(15, SortBy("unknown_sort_value"))
	if err != nil {
		t.Fatalf("OutputText(unknown) error: %v", err)
	}
}

func TestTextPrinter_OutputText_PrintFooterError(t *testing.T) {
	t.Parallel()

	content := testPlumbCode

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(content))
	tp := p.(*TextPrinter)

	node := &syntax.Node{Filename: "test.go", Pos: 15, End: 40}
	dups := [][]*syntax.Node{{node}}

	clones, err := prepareClonesInfo(tp.ReadFile, dups)
	if err != nil {
		t.Fatalf("prepareClonesInfo() error: %v", err)
	}

	tp.cloneGroups = [][]clone{clones}

	fw := &firstWriteFailsWriter{w: &buf}
	tp.w = fw

	err = tp.OutputText(15, SortBySize)
	if err == nil {
		t.Error("expected error when PrintFooter write fails")
	}
}

func TestTextPrinter_PrintClonesSorted_ReadError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	fread := func(filename string) ([]byte, error) {
		return nil, errors.New("simulated read error")
	}

	p := NewText(&buf, fread)

	node := &syntax.Node{Filename: "missing.go", Pos: 0, End: 10}
	dups := [][]*syntax.Node{{node}}

	err := p.(*TextPrinter).PrintClonesSorted(dups, SortBySize)
	if err == nil {
		t.Error("expected error on read failure in PrintClonesSorted")
	}
}

func TestTextPrinter_PrintClonesSorted_SortByOccurrence(t *testing.T) {
	t.Parallel()

	content := "package main\n\nfunc foo() {}\nfunc bar() {}\n"

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(content))

	nodes := createMockNodes(t)
	dups := [][]*syntax.Node{nodes}

	err := p.(*TextPrinter).PrintClonesSorted(dups, SortByOccurrence)
	if err != nil {
		t.Fatalf("PrintClonesSorted(SortByOccurrence) error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "sorted by occurrence") {
		t.Errorf("PrintClonesSorted missing sort info: %q", output)
	}
}

func TestTextPrinter_PrintClonesSorted_SortByHash(t *testing.T) {
	t.Parallel()

	content := "package main\n\nfunc foo() {}\n"

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(content))

	nodes := createMockNodes(t)
	dups := [][]*syntax.Node{nodes}

	err := p.(*TextPrinter).PrintClonesSorted(dups, SortByHash)
	if err != nil {
		t.Fatalf("PrintClonesSorted(SortByHash) error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "sorted by hash") {
		t.Errorf("PrintClonesSorted missing sort info: %q", output)
	}
}

func TestTextPrinter_PrintClonesSorted_UnknownSort(t *testing.T) {
	t.Parallel()

	content := "package main\n\nfunc foo() {}\n"

	var buf bytes.Buffer

	p := NewText(&buf, mockReadFile(content))

	nodes := createMockNodes(t)
	dups := [][]*syntax.Node{nodes}

	err := p.(*TextPrinter).PrintClonesSorted(dups, SortBy("unknown"))
	if err != nil {
		t.Fatalf("PrintClonesSorted(unknown) error: %v", err)
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

	err := tp.OutputText(15, SortBySize)
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
