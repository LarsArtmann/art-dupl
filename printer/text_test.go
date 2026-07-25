package printer

import (
	"bytes"
	"errors"
	"io"
	"strings"
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

	node := testutil.CreateNodeWithPos(0, testFilename, 0, 5)
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

func TestTotalFragmentSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		clones []domain.ProcessedClone
		want   int
	}{
		{"nil", nil, 0},
		{"two fragments", []domain.ProcessedClone{
			{CloneRef: domain.CloneRef{Fragment: "hello"}},
			{CloneRef: domain.CloneRef{Fragment: "world!"}},
		}, 11},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := totalFragmentSize(tc.clones); got != tc.want {
				t.Errorf("totalFragmentSize() = %d, want %d", got, tc.want)
			}
		})
	}
}

func TestPreviewFirstLine(t *testing.T) {
	t.Parallel()

	longLine := strings.Repeat("a", maxPreviewRunes+20)

	tests := []struct {
		name     string
		clone    domain.ProcessedClone
		readFile ReadFile
		want     string
	}{
		{
			name:     "fragment single line",
			clone:    domain.ProcessedClone{CloneRef: domain.CloneRef{Fragment: "func foo() {"}},
			readFile: nil,
			want:     "  | func foo() {",
		},
		{
			name: "fragment multiline uses first non-empty",
			clone: domain.ProcessedClone{
				CloneRef: domain.CloneRef{Fragment: "\n\n\tfunc bar() int {\n\t\treturn 1\n\t}"},
			},
			readFile: nil,
			want:     "  | func bar() int {",
		},
		{
			name: "fragment empty falls back to ReadFile line",
			clone: domain.ProcessedClone{CloneRef: domain.CloneRef{
				Filename: "src.go", LineStart: 2, Fragment: "",
			}},
			readFile: mockReadFile("package main\nfunc baz() {}\nvar x = 1"),
			want:     "  | func baz() {}",
		},
		{
			name: "fragment empty and ReadFile error yields no preview",
			clone: domain.ProcessedClone{CloneRef: domain.CloneRef{
				Filename: "missing.go", LineStart: 1, Fragment: "",
			}},
			readFile: errorReadFile("not found"),
			want:     "",
		},
		{
			name: "fragment empty and nil ReadFile yields no preview",
			clone: domain.ProcessedClone{CloneRef: domain.CloneRef{
				Filename: "x.go", LineStart: 1, Fragment: "",
			}},
			readFile: nil,
			want:     "",
		},
		{
			name:     "long fragment line truncated to maxPreviewRunes",
			clone:    domain.ProcessedClone{CloneRef: domain.CloneRef{Fragment: longLine}},
			readFile: nil,
			want:     "  | " + strings.Repeat("a", maxPreviewRunes-1) + "…",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tp := &TextPrinter{ReadFile: tc.readFile}
			got := tp.previewFirstLine(tc.clone)

			if got != tc.want {
				t.Errorf("previewFirstLine() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestPrintCloneListIncludesPreview(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	tp := &TextPrinter{w: &buf, ReadFile: nil}

	clones := []domain.ProcessedClone{
		{CloneRef: domain.CloneRef{Filename: "a.go", LineStart: 10, LineEnd: 20, Fragment: "func hello() {"}},
		{CloneRef: domain.CloneRef{Filename: "b.go", LineStart: 5, LineEnd: 15, Fragment: ""}},
	}

	if err := tp.printCloneList(clones); err != nil {
		t.Fatalf("printCloneList error: %v", err)
	}

	output := buf.String()

	testutil.AssertStringContains(t, output, "a.go:10-20  | func hello() {", "preview should appear after location")
	testutil.AssertStringContains(t, output, "b.go:5-15\n", "no preview when fragment empty and no reader")
}

func TestProcessClones(t *testing.T) {
	t.Parallel()

	content := testPlumbCode

	fread := mockReadFile(content)

	node1 := testutil.CreateNodeWithPos(0, testFilename, 15, 40)
	node2 := testutil.CreateNodeWithPos(0, testFilename, 40, 50)
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

	testNode := testutil.CreateNodeWithPos(0, "missing.go", 0, 10)
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

	testNode := testutil.CreateNodeWithPos(0, testFilename, 15, 40)
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
		newTestProcessedClone("a.go", 1, 3, "code"),
		{CloneRef: domain.CloneRef{Filename: "b.go", LineStart: 1, LineEnd: 3, Fragment: "code"}},
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
				{CloneRef: domain.CloneRef{Filename: tc.filename, LineStart: 1, LineEnd: 3, Fragment: "code"}},
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

	testNode := testutil.CreateNodeWithPos(0, testFilename, 15, 40)
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

	node := testutil.CreateNodeWithPos(0, "missing.go", 0, 10)

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
		{CloneRef: domain.CloneRef{Filename: "a.go", LineStart: 1, LineEnd: 3, Fragment: ""}},
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

func TestTextPrinter_writeExplanation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		cls        domain.CloneClassification
		cloneCount int
		want       []string
		notWant    []string
	}{
		{
			name: "actionable clone",
			cls: domain.CloneClassification{
				CloneType:     domain.CloneType2,
				Actionability: domain.Actionable,
				Category:      domain.CategoryFunction,
				Tokens:        42,
				Lines:         8,
			},
			cloneCount: 3,
			want:       []string{"type-2", "actionable", "function", "42 tokens, 8 lines"},
			notWant:    []string{"non-actionable", "extractable"},
		},
		{
			name: "non-actionable with pattern",
			cls: domain.CloneClassification{
				CloneType:            domain.CloneType2,
				Actionability:        domain.NonActionable,
				NonActionablePattern: "guard-clause",
				Category:             domain.CategoryFunction,
				Tokens:               5,
				Lines:                2,
			},
			cloneCount: 4,
			want:       []string{"non-actionable (guard-clause)"},
			notWant:    []string{"extractable"},
		},
		{
			name: "non-actionable without pattern defaults to boilerplate",
			cls: domain.CloneClassification{
				CloneType:     domain.CloneType1,
				Actionability: domain.NonActionable,
				Category:      domain.CategoryBlock,
				Tokens:        3,
				Lines:         1,
			},
			cloneCount: 2,
			want:       []string{"non-actionable (boilerplate)"},
		},
		{
			name: "extractable shows savings",
			cls: domain.CloneClassification{
				CloneType:      domain.CloneType2,
				Actionability:  domain.Actionable,
				Category:       domain.CategoryFunction,
				Tokens:         10,
				Lines:          5,
				Extractability: domain.Extractability{CanExtract: true, EstimatedLinesSaved: 12},
			},
			cloneCount: 3,
			want:       []string{"extractable: ~12 lines saved across 3 sites"},
		},
		{
			name: "suggestion on actionable clone uses fix label",
			cls: domain.CloneClassification{
				CloneType:     domain.CloneType2,
				Actionability: domain.Actionable,
				Category:      domain.CategoryFunction,
				Suggestion:    "extract to a helper",
			},
			want: []string{"  fix: extract to a helper"},
		},
		{
			name: "suggestion on non-actionable clone uses why label",
			cls: domain.CloneClassification{
				CloneType:            domain.CloneType2,
				Actionability:        domain.NonActionable,
				NonActionablePattern: "signature-only",
				Category:             domain.CategoryFunction,
				Suggestion:           "required method signature",
			},
			want: []string{"  why: required method signature"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			tp := &TextPrinter{w: &buf, ReadFile: mockReadFile("")}

			if err := tp.writeExplanation(tt.cls, tt.cloneCount); err != nil {
				t.Fatalf("writeExplanation: %v", err)
			}

			got := buf.String()

			for _, s := range tt.want {
				if !strings.Contains(got, s) {
					t.Errorf("output missing %q\noutput:\n%s", s, got)
				}
			}

			for _, s := range tt.notWant {
				if strings.Contains(got, s) {
					t.Errorf("output unexpectedly contains %q\noutput:\n%s", s, got)
				}
			}
		})
	}
}
