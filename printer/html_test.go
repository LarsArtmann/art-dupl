package printer

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// assertOutputContainsAll checks that the output string contains all wanted substrings.
func assertOutputContainsAll(t *testing.T, output, testName string, wantSubstr []string) {
	t.Helper()

	for _, want := range wantSubstr {
		testutil.AssertStringContains(t, output, want, testName+" missing "+want)
	}
}

// makeNodePairForFiles creates two syntax.Node slices for two files with the same structure.
func makeNodePairForFiles(file1, file2 string) ([]*syntax.Node, []*syntax.Node) {
	return []*syntax.Node{
			{Type: golang.FuncDecl, Filename: file1, Pos: 14, End: 44},
			{Type: golang.ExprStmt, Filename: file1, Pos: 15, End: 35},
		},
		[]*syntax.Node{
			{Type: golang.FuncDecl, Filename: file2, Pos: 14, End: 44},
			{Type: golang.ExprStmt, Filename: file2, Pos: 15, End: 35},
		}
}

var errReadFail = errors.New("read failed")

const (
	testGoCode      = "package main\n\nfunc foo() {\n\tfmt.Println(\"hello\")\n}\n"
	testGoMultiCode = "package main\n\nfunc foo() {\n\tfmt.Println(\"hello\")\n}\n\nfunc bar() {\n\tfmt.Println(\"hello\")\n}\n"
	testPlumbCode   = "package main\n\nfunc foo() {\n\tprintln(\"hello\")\n}\n"
	testFilename    = "test.go"
)

type errorWriter struct{}

func (errorWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

func htmlPrinterWithContent(content string) (Printer, *bytes.Buffer) {
	var buf bytes.Buffer

	p := NewHTML(&buf, mockReadFile(content), 15)

	return p, &buf
}

func htmlPrinterWithMetadata(content string, meta ReportMetadata) (Printer, *bytes.Buffer) {
	var buf bytes.Buffer

	p := NewHTMLWithOptions(&buf, mockReadFile(content), config.DiffModeSideBySide, meta, 15)

	return p, &buf
}

func nodesForHTML(filename string) []*syntax.Node {
	return []*syntax.Node{
		{Type: golang.FuncDecl, Filename: filename, Pos: 14, End: 44},
		{Type: golang.ExprStmt, Filename: filename, Pos: 15, End: 35},
	}
}

func TestNewHTML_Variants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name              string
		create            func() Printer
		wantThreshold     int
		wantDiffMode      config.DiffMode
		wantSemantic      bool
		checkSemanticOnly bool
	}{
		{
			"default",
			func() Printer {
				var buf bytes.Buffer

				return NewHTML(&buf, mockReadFile(""))
			},
			15, config.DiffModeDisabled, false, false,
		},
		{
			"custom threshold",
			func() Printer {
				var buf bytes.Buffer

				return NewHTML(&buf, mockReadFile(""), 42)
			},
			42, config.DiffModeDisabled, false, false,
		},
		{
			"with options",
			func() Printer {
				var buf bytes.Buffer

				meta := ReportMetadata{
					Semantic:         true,
					DetectionMethods: []string{"suffix-tree"},
					SortBy:           "size",
				}

				return NewHTMLWithOptions(
					&buf,
					mockReadFile(""),
					config.DiffModeSideBySide,
					meta,
					30,
				)
			},
			30, config.DiffModeSideBySide, true, true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			p := tc.create()
			hp := p.(*htmlprinter)

			if hp.threshold != tc.wantThreshold {
				t.Errorf("threshold = %d, want %d", hp.threshold, tc.wantThreshold)
			}

			if !tc.checkSemanticOnly && hp.diffMode != tc.wantDiffMode {
				t.Errorf("diffMode = %v, want %v", hp.diffMode, tc.wantDiffMode)
			}

			if tc.checkSemanticOnly && !hp.metadata.Semantic {
				t.Error("metadata.Semantic = false, want true")
			}
		})
	}
}

func TestHTMLPrintHeader(t *testing.T) {
	t.Parallel()

	p, buf := htmlPrinterWithContent("")

	err := p.PrintHeader()
	if err != nil {
		t.Fatalf("PrintHeader failed: %v", err)
	}

	output := buf.String()
	testutil.AssertStringContains(
		t,
		output,
		"Code Duplication Report",
		"Expected report title in header",
	)

	testutil.AssertStringContains(
		t,
		output,
		"Threshold (tokens)",
		"Expected threshold card in header",
	)

	testutil.AssertStringContains(t, output, "15", "Expected threshold value 15 in header")
}

func TestHTMLPrintHeader_WithMetadata(t *testing.T) {
	t.Parallel()

	meta := ReportMetadata{
		Semantic:         true,
		DetectionMethods: []string{"suffix-tree", "hash"},
		SortBy:           "size",
		IncludeSQLC:      true,
		IncludeTempl:     true,
	}
	p, buf := htmlPrinterWithMetadata("", meta)

	err := p.PrintHeader()
	if err != nil {
		t.Fatalf("PrintHeader with metadata failed: %v", err)
	}

	output := buf.String()
	testutil.AssertStringContains(t, output, "Semantic", "Expected Semantic badge in output")

	testutil.AssertStringContains(
		t,
		output,
		"suffix-tree, hash",
		"Expected detection methods in output",
	)

	testutil.AssertStringContains(t, output, "size", "Expected SortBy badge in output")

	testutil.AssertStringContains(t, output, "Include SQLC", "Expected IncludeSQLC badge in output")

	testutil.AssertStringContains(
		t,
		output,
		"Include Templ",
		"Expected IncludeTempl badge in output",
	)
}

func TestHTMLPrintHeader_MetadataStructural(t *testing.T) {
	t.Parallel()

	meta := ReportMetadata{Semantic: false}
	p, buf := htmlPrinterWithMetadata("", meta)

	err := p.PrintHeader()
	if err != nil {
		t.Fatalf("PrintHeader failed: %v", err)
	}

	output := buf.String()
	testutil.AssertStringContains(
		t,
		output,
		"Structural",
		"Expected Structural badge when Semantic=false",
	)
}

func TestHTMLPrintHeader_EmptyMetadata(t *testing.T) {
	t.Parallel()

	p, buf := htmlPrinterWithMetadata("", ReportMetadata{})

	err := p.PrintHeader()
	if err != nil {
		t.Fatalf("PrintHeader failed: %v", err)
	}

	output := buf.String()
	// Empty metadata still writes Structural badge (Semantic=false default)
	testutil.AssertStringContains(
		t,
		output,
		"Structural",
		"Expected Structural badge in output for default metadata",
	)
}

func TestHTMLPrintClones(t *testing.T) {
	t.Parallel()

	content := testGoCode
	p, buf := htmlPrinterWithContent(content)

	mustPrintHeader(t, p)

	nodes := nodesForHTML(testFilename)
	dups := [][]*syntax.Node{nodes}

	mustPrintClones(t, p, dups)

	output := buf.String()
	testutil.AssertStringContains(t, output, "clone-group", "Expected clone-group div in output")

	testutil.AssertStringContains(t, output, "Clone Group #1", "Expected 'Clone Group #1' header")

	testutil.AssertStringContains(t, output, "occurrences", "Expected occurrences badge")
}

func TestHTMLPrintClones_MultipleGroups(t *testing.T) {
	t.Parallel()

	content := testGoMultiCode
	p, buf := htmlPrinterWithContent(content)

	mustPrintHeader(t, p)

	nodes1 := nodesForHTML(testFilename)
	secondNodes := nodesForHTML(testFilename)

	mustPrintClones(t, p, [][]*syntax.Node{nodes1})
	mustPrintClones(t, p, [][]*syntax.Node{secondNodes})

	output := buf.String()
	testutil.AssertStringContains(t, output, "Clone Group #1", "Expected 'Clone Group #1'")

	testutil.AssertStringContains(t, output, "Clone Group #2", "Expected 'Clone Group #2'")
}

func TestHTMLPrintClones_DiffMode(t *testing.T) {
	t.Parallel()

	content := testGoMultiCode
	p, buf := htmlPrinterWithMetadata(content, ReportMetadata{})

	mustPrintHeader(t, p)

	sourceNodes, secondNodes := makeNodePairForFiles("a.go", "b.go")

	fread := mockReadFile(content)

	err := p.PrintClones(
		processTestNodes(fread, "test", [][]*syntax.Node{sourceNodes, secondNodes}),
	)
	if err != nil {
		t.Fatalf("PrintClones with diff mode failed: %v", err)
	}

	output := buf.String()
	testutil.AssertStringContains(t, output, "diff-mode", "Expected diff-mode section in output")

	testutil.AssertStringContains(
		t,
		output,
		"BASE REFERENCE",
		"Expected BASE REFERENCE in diff output",
	)
}

func TestHTMLPrintClones_EmptyDups(t *testing.T) {
	p, _ := htmlPrinterWithContent("")
	mustPrintHeader(t, p)

	_, err := NodesToGroup(mockReadFile(""), "test", [][]*syntax.Node{{}})
	if err == nil {
		t.Error("Expected error for empty node slice")
	}
}

func TestHTMLPrintFooter(t *testing.T) {
	t.Parallel()

	p, buf := htmlPrinterWithContent("")

	mustPrintHeader(t, p)
	mustPrintFooter(t, p)

	output := buf.String()
	testutil.AssertStringContains(t, output, "</html>", "Expected closing </html> in footer")

	testutil.AssertStringContains(
		t,
		output,
		"Generated by art-dupl",
		"Expected 'Generated by art-dupl' in footer",
	)
}

func TestHTMLPrintFooter_WithSummary(t *testing.T) {
	t.Parallel()

	content := testGoCode
	p, buf := htmlPrinterWithContent(content)

	mustPrintHeader(t, p)

	nodes := nodesForHTML(testFilename)
	mustPrintClones(t, p, [][]*syntax.Node{nodes})

	mustPrintFooter(t, p)

	output := buf.String()
	testutil.AssertStringContains(
		t,
		output,
		"Summary",
		"Expected Summary section when clones were printed",
	)

	testutil.AssertStringContains(t, output, "Total Clones", "Expected 'Total Clones' in summary")
}

func TestHTMLPrintFooter_NoSummaryWhenEmpty(t *testing.T) {
	t.Parallel()

	p, buf := htmlPrinterWithContent("")

	testutil.AssertFatalNoError(t, p.PrintHeader(), "PrintHeader")

	testutil.AssertFatalNoError(t, p.PrintFooter(), "PrintFooter")

	output := buf.String()
	// buildSummarySection returns "" when totalClones==0, so no summary data div
	if strings.Contains(output, "Total Clones") {
		t.Error("Expected no summary data when no clones printed")
	}

	// Footer still contains the closing tags
	testutil.AssertStringContains(t, output, "</html>", "Expected closing html tag")
}

func TestHTMLOutputHTML(t *testing.T) {
	t.Parallel()

	content := testGoCode
	p, buf := htmlPrinterWithContent(content)

	hp := p.(*htmlprinter)

	hp.dupls = [][]domain.ProcessedClone{
		{
			newTestProcessedClone(testFilename, 1, 5, testGoCode),
			newTestProcessedClone("other.go", 3, 5, testGoCode),
		},
	}

	err := hp.OutputHTML(15, config.SortBySize)
	if err != nil {
		t.Fatalf("OutputHTML failed: %v", err)
	}

	output := buf.String()
	testutil.AssertStringContains(
		t,
		output,
		"Clone Group #",
		"Expected clone groups in OutputHTML output",
	)
}

func TestSuggestionHTML(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		suggestion string
		want       string
	}{
		{"empty", "", ""},
		{"with text", "Extract helper", `<div class="suggestion">💡 Extract helper</div>`},
		{
			"html escape",
			"<script>alert('x')</script>",
			`<div class="suggestion">💡 &lt;script&gt;alert(&#39;x&#39;)&lt;/script&gt;</div>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := suggestionHTML(tt.suggestion)
			if got != tt.want {
				t.Errorf("suggestionHTML(%q) = %q, want %q", tt.suggestion, got, tt.want)
			}
		})
	}
}

func TestHTMLCountDiffStats(t *testing.T) {
	t.Parallel()

	diff := DiffResult{
		Compared: []DiffLine{
			{Content: "a", Type: DiffLineAdded},
			{Content: "b", Type: DiffLineEqual},
			{Content: "c", Type: DiffLineRemoved},
			{Content: "d", Type: DiffLineModified},
			{Content: "e", Type: DiffLineAdded},
		},
	}

	added, removed, modified := countDiffStats(diff)
	if added != 2 {
		t.Errorf("added = %d, want 2", added)
	}

	if removed != 1 {
		t.Errorf("removed = %d, want 1", removed)
	}

	if modified != 1 {
		t.Errorf("modified = %d, want 1", modified)
	}
}

func TestHTMLCountDiffStats_Empty(t *testing.T) {
	t.Parallel()

	diff := DiffResult{}
	added, removed, modified := countDiffStats(diff)

	if added != 0 || removed != 0 || modified != 0 {
		t.Errorf(
			"want all zeros, got added=%d removed=%d modified=%d",
			added,
			removed,
			modified,
		)
	}
}

func TestCountDiffLineStats(t *testing.T) {
	t.Parallel()

	diff := DiffResult{
		Base: []DiffLine{
			{Content: "a", Type: DiffLineRemoved},
			{Content: "b", Type: DiffLineEqual},
		},
		Compared: []DiffLine{
			{Content: "c", Type: DiffLineAdded},
			{Content: "d", Type: DiffLineModified},
		},
	}

	added, removed, modified := countDiffLineStats(diff)
	if added != 1 {
		t.Errorf("added = %d, want 1", added)
	}

	if removed != 1 {
		t.Errorf("removed = %d, want 1", removed)
	}

	if modified != 1 {
		t.Errorf("modified = %d, want 1", modified)
	}
}

func TestOrderedCategories(t *testing.T) {
	t.Parallel()

	cats := orderedCategories()
	if len(cats) != 11 {
		t.Errorf("want 11 categories, got %d", len(cats))
	}

	if cats[0] != domain.CategoryFunction {
		t.Errorf("first category = %v, want domain.CategoryFunction", cats[0])
	}

	seen := make(map[CloneCategory]bool)
	for _, c := range cats {
		if seen[c] {
			t.Errorf("duplicate category: %v", c)
		}

		seen[c] = true
	}
}

func TestOrderedPriorities(t *testing.T) {
	t.Parallel()

	pris := orderedPriorities()
	if len(pris) != 4 {
		t.Errorf("want 4 priorities, got %d", len(pris))
	}

	if pris[0] != domain.PriorityCritical {
		t.Errorf("first priority = %v, want domain.PriorityCritical", pris[0])
	}

	if pris[len(pris)-1] != domain.PriorityLow {
		t.Errorf("last priority = %v, want domain.PriorityLow", pris[len(pris)-1])
	}
}

func TestBuildSummarySection(t *testing.T) {
	t.Parallel()

	p, _ := htmlPrinterWithContent("")
	hp := p.(*htmlprinter)

	hp.stats = classificationStats{
		categoryCounts: map[CloneCategory]int{
			domain.CategoryFunction: 5,
			domain.CategoryLoop:     2,
		},
		priorityCounts: map[ClonePriority]int{
			domain.PriorityCritical: 1,
			domain.PriorityMedium:   3,
		},
		testCount:   2,
		prodCount:   5,
		totalClones: 7,
		totalTokens: 150,
	}

	summary := hp.buildSummarySection()
	testutil.AssertStringContains(t, summary, "Total Clones", "Expected 'Total Clones' in summary")

	testutil.AssertStringContains(t, summary, "150", "Expected total tokens in summary")

	testutil.AssertStringContains(t, summary, "Production", "Expected 'Production' in summary")

	testutil.AssertStringContains(t, summary, "Test Code", "Expected 'Test Code' in summary")

	testutil.AssertStringContains(t, summary, "function", "Expected function category in summary")

	testutil.AssertStringContains(t, summary, "critical", "Expected critical priority in summary")
}

func TestBuildSummarySection_Empty(t *testing.T) {
	t.Parallel()

	p, _ := htmlPrinterWithContent("")
	hp := p.(*htmlprinter)

	summary := hp.buildSummarySection()
	if summary != "" {
		t.Errorf("want empty summary for no clones, got %q", summary)
	}
}

func TestBuildClones(t *testing.T) {
	t.Parallel()

	content := testGoCode
	fread := mockReadFile(content)

	nodes := nodesForHTML(testFilename)

	group, err := NodesToGroup(fread, "test", [][]*syntax.Node{nodes})
	if err != nil {
		t.Fatalf("NodesToGroup failed: %v", err)
	}

	testutil.AssertFatalLen(t, group.Clones, 1, "clones")

	if group.Clones[0].Filename != testFilename {
		t.Errorf("filename = %q, want 'test.go'", group.Clones[0].Filename)
	}
}

func TestBuildClones_EmptyDup(t *testing.T) {
	t.Parallel()

	fread := mockReadFile("")

	_, err := NodesToGroup(fread, "test", [][]*syntax.Node{{}})
	if err == nil {
		t.Error("want error for empty node slice")
	}
}

func TestBuildClones_ReadError(t *testing.T) {
	t.Parallel()

	fread := func(_ string) ([]byte, error) {
		return nil, errReadFail
	}

	nodes := nodesForHTML("missing.go")

	_, err := NodesToGroup(fread, "test", [][]*syntax.Node{nodes})
	if err == nil {
		t.Error("want error for read failure")
	}
}

func TestWriteMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		meta     ReportMetadata
		contains string
	}{
		{
			"semantic enabled",
			ReportMetadata{Semantic: true},
			"Semantic",
		},
		{
			"semantic disabled",
			ReportMetadata{Semantic: false},
			"Structural",
		},
		{
			"detection methods",
			ReportMetadata{DetectionMethods: []string{"hash"}},
			"hash",
		},
		{
			"sort by",
			ReportMetadata{SortBy: "occurrence"},
			"occurrence",
		},
		{
			"include sqlc",
			ReportMetadata{IncludeSQLC: true},
			"Include SQLC",
		},
		{
			"include templ",
			ReportMetadata{IncludeTempl: true},
			"Include Templ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			hp := &htmlprinter{
				w:        &buf,
				metadata: tt.meta,
			}

			err := hp.writeMetadata()
			if err != nil {
				t.Fatalf("writeMetadata failed: %v", err)
			}

			testutil.AssertStringContains(
				t,
				buf.String(),
				tt.contains,
				"want "+tt.contains+" in output",
			)
		})
	}
}

func TestWriteMetadata_Empty(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	hp := &htmlprinter{
		w:        &buf,
		metadata: ReportMetadata{},
	}

	err := hp.writeMetadata()
	if err != nil {
		t.Fatalf("writeMetadata failed: %v", err)
	}

	// Even empty metadata writes the Structural badge (Semantic=false default)
	testutil.AssertStringContains(
		t,
		buf.String(),
		"Structural",
		"want Structural badge for default metadata",
	)
}

func TestHTMLErrorInWriter(t *testing.T) {
	t.Parallel()

	p := NewHTML(&errorWriter{}, mockReadFile(""), 15)

	err := p.PrintHeader()
	if err == nil {
		t.Error("want error from failing writer")
	}
}

func TestHTMLErrorInPrintClones(t *testing.T) {
	t.Parallel()

	content := testGoCode
	p := NewHTML(&errorWriter{}, mockReadFile(content), 15)

	nodes := nodesForHTML(testFilename)

	fread := mockReadFile(content)

	err := p.PrintClones(processTestNodes(fread, "test", [][]*syntax.Node{nodes}))
	if err == nil {
		t.Error("want error from failing writer during PrintClones")
	}
}

func TestHTMLErrorInPrintFooter(t *testing.T) {
	t.Parallel()

	p := NewHTML(&errorWriter{}, mockReadFile(""), 15)

	err := p.PrintFooter()
	if err == nil {
		t.Error("want error from failing writer during PrintFooter")
	}
}

func TestHTMLComputeCloneGroupDiff_Empty(t *testing.T) {
	t.Parallel()

	result := ComputeCloneGroupDiff(nil)
	if result.Base != nil {
		t.Error("want nil Base for empty clones")
	}
}

func TestHTMLComputeCloneGroupDiff_SingleClone(t *testing.T) {
	t.Parallel()

	clones := []domain.ProcessedClone{newTestProcessedClone("a.go", 1, 5, "line1\nline2\n")}

	result := ComputeCloneGroupDiff(clones)
	assertBaseFilename(t, &result, "a.go")

	if len(result.Others) != 0 {
		t.Errorf("Others length = %d, want 0", len(result.Others))
	}
}

func TestHTMLComputeCloneGroupDiff_MultipleClones(t *testing.T) {
	t.Parallel()

	clones := []domain.ProcessedClone{
		newTestProcessedClone("a.go", 1, 5, "line1\nline2\n"),
		newTestProcessedClone("b.go", 10, 14, "line1\nmodified\n"),
	}

	result := ComputeCloneGroupDiff(clones)
	if len(result.Others) != 1 {
		t.Fatalf("Others length = %d, want 1", len(result.Others))
	}

	if result.Others[0].Filename != "b.go" {
		t.Errorf("Others[0].Filename = %q, want 'b.go'", result.Others[0].Filename)
	}
}

func TestLineDiff_Equal(t *testing.T) {
	t.Parallel()

	base := []byte("hello\nworld\n")
	compared := []byte("hello\nworld\n")

	result := LineDiff(base, compared)
	if result.HasDiff {
		t.Error("want no diff for identical content")
	}
}

func TestLineDiff_Different(t *testing.T) {
	t.Parallel()

	base := []byte("hello\nworld\n")
	compared := []byte("hello\nchanged\n")

	result := LineDiff(base, compared)
	if !result.HasDiff {
		t.Error("want diff for different content")
	}

	found := false

	for _, line := range result.Base {
		if line.Type == DiffLineModified {
			found = true

			break
		}
	}

	if !found {
		t.Error("want at least one modified line")
	}
}

func TestLineDiff_DifferentLength(t *testing.T) {
	t.Parallel()

	base := []byte("line1\nline2\nline3\n")
	compared := []byte("line1\nline2\n")

	result := LineDiff(base, compared)
	if !result.HasDiff {
		t.Error("want diff for different-length content")
	}
}

func TestLineDiff_LargeFile(t *testing.T) {
	t.Parallel()

	var baseLines, comparedLines []byte

	baseLines = make([]byte, 0, 1500)
	comparedLines = make([]byte, 0, 800)

	for i := range 150 {
		baseLines = append(baseLines, []byte("base line\n")...)
		if i < 80 {
			comparedLines = append(comparedLines, []byte("compared line\n")...)
		}
	}

	result := LineDiff(baseLines, comparedLines)
	if !result.HasDiff {
		t.Error("want diff for large files with different content")
	}
}

func TestSplitLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input []byte
		want  int
	}{
		{"empty", []byte{}, 1},
		{"single line no newline", []byte("hello"), 1},
		{"single line with newline", []byte("hello\n"), 1},
		{"two lines", []byte("hello\nworld\n"), 2},
		{"trailing no newline", []byte("hello\nworld"), 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lines := splitLines(tt.input)
			if len(lines) != tt.want {
				t.Errorf("splitLines returned %d lines, want %d", len(lines), tt.want)
			}
		})
	}
}

func TestWordDiff(t *testing.T) {
	t.Parallel()

	result := WordDiff("hello world", "hello earth")
	testutil.AssertStringContains(t, result, "word-removed", "want word-removed class in diff")

	testutil.AssertStringContains(t, result, "word-added", "want word-added class in diff")
}

func TestWordDiff_Identical(t *testing.T) {
	t.Parallel()

	result := WordDiff("hello world", "hello world")
	if strings.Contains(result, "word-removed") || strings.Contains(result, "word-added") {
		t.Error("want no word-level diff for identical content")
	}
}

// newDiffHTMLPrinter creates an htmlprinter with side-by-side diff mode for testing.
func newDiffHTMLPrinter() (*htmlprinter, *bytes.Buffer) {
	var buf bytes.Buffer

	return &htmlprinter{w: &buf, iota: 1, diffMode: config.DiffModeSideBySide}, &buf
}

func newCloneWithContent(filename string, lineStart int, content string) *CloneWithContent {
	return &CloneWithContent{
		CloneWithContentMixin: CloneWithContentMixin{Filename: filename, LineStart: lineStart},
		Content:               []byte(content),
	}
}

func newCloneDiff(filename string, lineStart int, content string, diff DiffResult) CloneDiff {
	return CloneDiff{
		CloneWithContent: *newCloneWithContent(filename, lineStart, content),
		Diff:             diff,
	}
}

func TestHTMLWriteCloneOccurrences(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	hp := &htmlprinter{w: &buf, iota: 1}

	clones := []domain.ProcessedClone{
		newTestProcessedClone("a.go", 10, 20, "code here"),
		newTestProcessedClone("b.go", 30, 40, "more code"),
	}

	err := hp.writeCloneOccurrences(clones)
	if err != nil {
		t.Fatalf("writeCloneOccurrences failed: %v", err)
	}

	output := buf.String()
	testutil.AssertStringContains(t, output, "a.go", "want a.go in occurrence output")

	testutil.AssertStringContains(t, output, "b.go", "want b.go in occurrence output")

	testutil.AssertStringContains(t, output, "vscode://", "Expected VSCode link in output")
}

func TestHTMLWriteCloneGroupFooter(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	hp := &htmlprinter{w: &buf}

	err := hp.writeCloneGroupFooter()
	if err != nil {
		t.Fatalf("writeCloneGroupFooter failed: %v", err)
	}

	if buf.String() != "</div></div>\n" {
		t.Errorf("unwant footer: %q", buf.String())
	}
}

func TestHTMLWriteDiffView_Variants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		clones     []domain.ProcessedClone
		wantSubstr []string
	}{
		{
			"single clone falls through",
			[]domain.ProcessedClone{newTestProcessedClone("a.go", 1, 5, "code\n")},
			[]string{"occurrence"},
		},
		{
			"multiple clones",
			[]domain.ProcessedClone{
				newTestProcessedClone("a.go", 1, 5, "line1\nline2\n"),
				newTestProcessedClone("b.go", 10, 14, "line1\nmodified\n"),
			},
			[]string{"BASE REFERENCE", "diff-legend"},
		},
		{
			"three clones with selector",
			[]domain.ProcessedClone{
				newTestProcessedClone("a.go", 1, 5, "line1\nline2\n"),
				newTestProcessedClone("b.go", 10, 14, "line1\nchanged\n"),
				newTestProcessedClone("c.go", 20, 24, "line1\nother\n"),
			},
			[]string{"diff-select", "Compare with..."},
		},
		{
			"aggregate stats",
			[]domain.ProcessedClone{
				newTestProcessedClone("a.go", 1, 5, "line1\nline2\nline3\n"),
				newTestProcessedClone("b.go", 10, 14, "line1\nmodified\nadded\n"),
			},
			[]string{"diff-aggregate-stats"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			hp, buf := newDiffHTMLPrinter()

			err := hp.writeDiffView(tc.clones)
			if err != nil {
				t.Fatalf("writeDiffView failed: %v", err)
			}

			output := buf.String()
			assertOutputContainsAll(t, output, tc.name, tc.wantSubstr)
		})
	}
}

func TestHTMLWriteDiffViewToggle(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	hp := &htmlprinter{w: &buf, iota: 5}

	err := hp.writeDiffViewToggle()
	if err != nil {
		t.Fatalf("writeDiffViewToggle failed: %v", err)
	}

	output := buf.String()
	testutil.AssertStringContains(
		t,
		output,
		"diff-toggle-5-side",
		"want side toggle button with group id 5",
	)

	testutil.AssertStringContains(
		t,
		output,
		"diff-toggle-5-inline",
		"want inline toggle button with group id 5",
	)
}

func TestHTMLWriteDiffSelector(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	hp := &htmlprinter{w: &buf, iota: 3}

	groupDiff := CloneGroupDiff{
		Base: newCloneWithContent("base.go", 1, "base\n"),
		Others: []CloneDiff{
			newCloneDiff("other1.go", 10, "other1\n", DiffResult{}),
			newCloneDiff("other2.go", 20, "other2\n", DiffResult{}),
		},
	}

	err := hp.writeDiffSelector(groupDiff)
	if err != nil {
		t.Fatalf("writeDiffSelector failed: %v", err)
	}

	output := buf.String()
	testutil.AssertStringContains(t, output, "diff-select-3", "want selector with group id 3")

	testutil.AssertStringContains(t, output, "other1.go", "want other1.go in selector options")

	testutil.AssertStringContains(t, output, "other2.go", "want other2.go in selector options")
}

func TestHTMLWriteDiffComparison_Variants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		iota        int
		compIdx     int
		groupID     int
		wantSubstr  string
		wantNoMatch string
	}{
		{"first_comparison", 1, 0, 1, "diff-compare-1-0", ""},
		{"second_comparison_inactive", 2, 1, 2, "", "active"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			hp := &htmlprinter{w: &buf, iota: tc.iota}

			base := newCloneWithContent("base.go", 1, "code\n")

			diff := LineDiff([]byte("code\n"), []byte("different\n"))
			other := newCloneDiff("other.go", 5, "different\n", diff)

			err := hp.writeDiffComparison(base, other, tc.compIdx, tc.groupID)
			if err != nil {
				t.Fatalf("writeDiffComparison failed: %v", err)
			}

			output := buf.String()
			if tc.wantSubstr != "" {
				testutil.AssertStringContains(
					t,
					output,
					tc.wantSubstr,
					tc.name+" missing "+tc.wantSubstr,
				)
			}

			if tc.wantNoMatch != "" && strings.Contains(output, tc.wantNoMatch) {
				t.Errorf("Expected no match for %q in output", tc.wantNoMatch)
			}
		})
	}
}

func TestHTMLRenderDiffLines_Variants(t *testing.T) {
	t.Parallel()

	diffLines := func(contents []string, types []DiffLineType, startLineNum int) []DiffLine {
		lines := make([]DiffLine, 0, len(contents))
		for i, content := range contents {
			lines = append(lines, DiffLine{Content: content, Type: types[i], LineNumber: startLineNum + i})
		}

		return lines
	}

	tests := []struct {
		name       string
		lines      []DiffLine
		opposite   []DiffLine
		wantSubstr []string
	}{
		{
			name: "all types",
			lines: diffLines(
				[]string{"equal line", "added line", "removed line", "modified line"},
				[]DiffLineType{DiffLineEqual, DiffLineAdded, DiffLineRemoved, DiffLineModified},
				1,
			),
			opposite: diffLines(
				[]string{"equal line", "base added", "base removed", "original modified"},
				[]DiffLineType{DiffLineEqual, DiffLineAdded, DiffLineRemoved, DiffLineModified},
				1,
			),
			wantSubstr: []string{"equal", "added", "removed", "modified"},
		},
		{
			name:       "word diff",
			lines:      diffLines([]string{"hello world"}, []DiffLineType{DiffLineModified}, 1),
			opposite:   diffLines([]string{"hello earth"}, []DiffLineType{DiffLineModified}, 1),
			wantSubstr: []string{"word-removed", "word-added"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer

			hp := &htmlprinter{w: &buf}

			err := hp.renderDiffLines(tc.lines, tc.opposite, true)
			if err != nil {
				t.Fatalf("renderDiffLines failed: %v", err)
			}

			output := buf.String()
			assertOutputContainsAll(t, output, tc.name, tc.wantSubstr)
		})
	}
}
