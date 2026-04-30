package printer

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

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

// assertContains checks that s contains substr.
func assertContains(t *testing.T, s, substr, msg string) {
	t.Helper()

	if !strings.Contains(s, substr) {
		t.Error(msg)
	}
}

func nodesForHTML(filename string) []*syntax.Node {
	return []*syntax.Node{
		{Type: golang.FuncDecl, Filename: filename, Pos: 14, End: 44},
		{Type: golang.ExprStmt, Filename: filename, Pos: 15, End: 35},
	}
}

func TestNewHTML(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	p := NewHTML(&buf, mockReadFile(""))

	hp := p.(*htmlprinter)
	if hp.threshold != 15 {
		t.Errorf("default threshold = %d, want 15", hp.threshold)
	}

	if hp.diffMode != config.DiffModeDisabled {
		t.Errorf("diffMode = %v, want disabled", hp.diffMode)
	}
}

func TestNewHTML_CustomThreshold(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	p := NewHTML(&buf, mockReadFile(""), 42)

	hp := p.(*htmlprinter)
	if hp.threshold != 42 {
		t.Errorf("threshold = %d, want 42", hp.threshold)
	}
}

func TestNewHTMLWithOptions(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	meta := ReportMetadata{
		Semantic:         true,
		DetectionMethods: []string{"suffix-tree"},
		SortBy:           "size",
	}
	p := NewHTMLWithOptions(&buf, mockReadFile(""), config.DiffModeSideBySide, meta, 30)

	hp := p.(*htmlprinter)
	if hp.threshold != 30 {
		t.Errorf("threshold = %d, want 30", hp.threshold)
	}

	if hp.diffMode != config.DiffModeSideBySide {
		t.Errorf("diffMode = %v, want side-by-side", hp.diffMode)
	}

	if !hp.metadata.Semantic {
		t.Error("metadata.Semantic = false, want true")
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
	if !strings.Contains(output, "Code Duplication Report") {
		t.Error("Expected report title in header")
	}

	if !strings.Contains(output, "Threshold (tokens)") {
		t.Error("Expected threshold card in header")
	}

	if !strings.Contains(output, "15") {
		t.Error("Expected threshold value 15 in header")
	}
}

func TestHTMLPrintHeader_WithMetadata(t *testing.T) {
	t.Parallel()

	meta := ReportMetadata{
		Semantic:         true,
		DetectionMethods: []string{"suffix-tree", "hash"},
		SortBy:           "size",
		FilterGenerated:  true,
		IncludeSQLC:      true,
		IncludeTempl:     true,
	}
	p, buf := htmlPrinterWithMetadata("", meta)

	err := p.PrintHeader()
	if err != nil {
		t.Fatalf("PrintHeader with metadata failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Semantic") {
		t.Error("Expected Semantic badge in output")
	}

	if !strings.Contains(output, "suffix-tree, hash") {
		t.Error("Expected detection methods in output")
	}

	if !strings.Contains(output, "size") {
		t.Error("Expected SortBy badge in output")
	}

	if !strings.Contains(output, "Filter Generated") {
		t.Error("Expected FilterGenerated badge in output")
	}

	if !strings.Contains(output, "Include SQLC") {
		t.Error("Expected IncludeSQLC badge in output")
	}

	if !strings.Contains(output, "Include Templ") {
		t.Error("Expected IncludeTempl badge in output")
	}
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
	if !strings.Contains(output, "Structural") {
		t.Error("Expected Structural badge when Semantic=false")
	}
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
	if !strings.Contains(output, "Structural") {
		t.Error("Expected Structural badge in output for default metadata")
	}
}

func TestHTMLPrintClones(t *testing.T) {
	t.Parallel()

	content := testGoCode
	p, buf := htmlPrinterWithContent(content)

	if err := p.PrintHeader(); err != nil {
		t.Fatalf("PrintHeader failed: %v", err)
	}

	nodes := nodesForHTML(testFilename)
	dups := [][]*syntax.Node{nodes}

	err := p.PrintClones(dups)
	if err != nil {
		t.Fatalf("PrintClones failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "clone-group") {
		t.Error("Expected clone-group div in output")
	}

	if !strings.Contains(output, "Clone Group #1") {
		t.Error("Expected 'Clone Group #1' header")
	}

	if !strings.Contains(output, "occurrences") {
		t.Error("Expected occurrences badge")
	}
}

func TestHTMLPrintClones_MultipleGroups(t *testing.T) {
	t.Parallel()

	content := testGoMultiCode
	p, buf := htmlPrinterWithContent(content)

	if err := p.PrintHeader(); err != nil {
		t.Fatalf("PrintHeader failed: %v", err)
	}

	nodes1 := nodesForHTML(testFilename)
	nodes2 := nodesForHTML(testFilename)

	if err := p.PrintClones([][]*syntax.Node{nodes1}); err != nil {
		t.Fatalf("PrintClones #1 failed: %v", err)
	}

	if err := p.PrintClones([][]*syntax.Node{nodes2}); err != nil {
		t.Fatalf("PrintClones #2 failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Clone Group #1") {
		t.Error("Expected 'Clone Group #1'")
	}

	if !strings.Contains(output, "Clone Group #2") {
		t.Error("Expected 'Clone Group #2'")
	}
}

func TestHTMLPrintClones_DiffMode(t *testing.T) {
	t.Parallel()

	content := testGoMultiCode
	p, buf := htmlPrinterWithMetadata(content, ReportMetadata{})

	if err := p.PrintHeader(); err != nil {
		t.Fatalf("PrintHeader failed: %v", err)
	}

	nodes1 := []*syntax.Node{
		{Type: golang.FuncDecl, Filename: "a.go", Pos: 14, End: 44},
		{Type: golang.ExprStmt, Filename: "a.go", Pos: 15, End: 35},
	}
	nodes2 := []*syntax.Node{
		{Type: golang.FuncDecl, Filename: "b.go", Pos: 14, End: 44},
		{Type: golang.ExprStmt, Filename: "b.go", Pos: 15, End: 35},
	}

	err := p.PrintClones([][]*syntax.Node{nodes1, nodes2})
	if err != nil {
		t.Fatalf("PrintClones with diff mode failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "diff-mode") {
		t.Error("Expected diff-mode section in output")
	}

	if !strings.Contains(output, "BASE REFERENCE") {
		t.Error("Expected BASE REFERENCE in diff output")
	}
}

func TestHTMLPrintClones_EmptyDups(t *testing.T) {
	t.Parallel()

	p, _ := htmlPrinterWithContent("")

	if err := p.PrintHeader(); err != nil {
		t.Fatalf("PrintHeader failed: %v", err)
	}

	err := p.PrintClones([][]*syntax.Node{{}})
	if err == nil {
		t.Error("Expected error for empty node slice")
	}
}

func TestHTMLPrintFooter(t *testing.T) {
	t.Parallel()

	p, buf := htmlPrinterWithContent("")

	if err := p.PrintHeader(); err != nil {
		t.Fatalf("PrintHeader failed: %v", err)
	}

	err := p.PrintFooter()
	if err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "</html>") {
		t.Error("Expected closing </html> in footer")
	}

	if !strings.Contains(output, "Generated by art-dupl") {
		t.Error("Expected 'Generated by art-dupl' in footer")
	}
}

func TestHTMLPrintFooter_WithSummary(t *testing.T) {
	t.Parallel()

	content := testGoCode
	p, buf := htmlPrinterWithContent(content)

	if err := p.PrintHeader(); err != nil {
		t.Fatalf("PrintHeader failed: %v", err)
	}

	nodes := nodesForHTML(testFilename)
	if err := p.PrintClones([][]*syntax.Node{nodes}); err != nil {
		t.Fatalf("PrintClones failed: %v", err)
	}

	if err := p.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Summary") {
		t.Error("Expected Summary section when clones were printed")
	}

	if !strings.Contains(output, "Total Clones") {
		t.Error("Expected 'Total Clones' in summary")
	}
}

func TestHTMLPrintFooter_NoSummaryWhenEmpty(t *testing.T) {
	t.Parallel()

	p, buf := htmlPrinterWithContent("")

	if err := p.PrintHeader(); err != nil {
		t.Fatalf("PrintHeader failed: %v", err)
	}

	if err := p.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	output := buf.String()
	// buildSummarySection returns "" when totalClones==0, so no summary data div
	if strings.Contains(output, "Total Clones") {
		t.Error("Expected no summary data when no clones printed")
	}

	// Footer still contains the closing tags
	if !strings.Contains(output, "</html>") {
		t.Error("Expected closing html tag")
	}
}

func TestHTMLOutputHTML(t *testing.T) {
	t.Parallel()

	content := testGoCode
	p, buf := htmlPrinterWithContent(content)

	hp := p.(*htmlprinter)

	nodes1 := nodesForHTML(testFilename)
	nodes2 := []*syntax.Node{
		{Type: golang.FuncDecl, Filename: "other.go", Pos: 39, End: 68},
		{Type: golang.ExprStmt, Filename: "other.go", Pos: 40, End: 60},
	}

	hp.dupls = [][][]*syntax.Node{
		{nodes1, nodes2},
	}

	err := hp.OutputHTML(15, SortBySize)
	if err != nil {
		t.Fatalf("OutputHTML failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Clone Group #") {
		t.Error("Expected clone groups in OutputHTML output")
	}
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
		{"html escape", "<script>alert('x')</script>", `<div class="suggestion">💡 &lt;script&gt;alert(&#39;x&#39;)&lt;/script&gt;</div>`},
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
		t.Errorf("expected all zeros, got added=%d removed=%d modified=%d", added, removed, modified)
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
		t.Errorf("expected 11 categories, got %d", len(cats))
	}

	if cats[0] != CategoryFunction {
		t.Errorf("first category = %v, want CategoryFunction", cats[0])
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
		t.Errorf("expected 4 priorities, got %d", len(pris))
	}

	if pris[0] != PriorityCritical {
		t.Errorf("first priority = %v, want PriorityCritical", pris[0])
	}

	if pris[len(pris)-1] != PriorityLow {
		t.Errorf("last priority = %v, want PriorityLow", pris[len(pris)-1])
	}
}

func TestBuildSummarySection(t *testing.T) {
	t.Parallel()

	p, _ := htmlPrinterWithContent("")
	hp := p.(*htmlprinter)

	hp.stats = classificationStats{
		categoryCounts: map[CloneCategory]int{
			CategoryFunction: 5,
			CategoryLoop:     2,
		},
		priorityCounts: map[ClonePriority]int{
			PriorityCritical: 1,
			PriorityMedium:   3,
		},
		testCount:   2,
		prodCount:   5,
		totalClones: 7,
		totalTokens: 150,
	}

	summary := hp.buildSummarySection()
	if !strings.Contains(summary, "Total Clones") {
		t.Error("Expected 'Total Clones' in summary")
	}

	if !strings.Contains(summary, "150") {
		t.Error("Expected total tokens in summary")
	}

	if !strings.Contains(summary, "Production") {
		t.Error("Expected 'Production' in summary")
	}

	if !strings.Contains(summary, "Test Code") {
		t.Error("Expected 'Test Code' in summary")
	}

	if !strings.Contains(summary, "function") {
		t.Error("Expected function category in summary")
	}

	if !strings.Contains(summary, "critical") {
		t.Error("Expected critical priority in summary")
	}
}

func TestBuildSummarySection_Empty(t *testing.T) {
	t.Parallel()

	p, _ := htmlPrinterWithContent("")
	hp := p.(*htmlprinter)

	summary := hp.buildSummarySection()
	if summary != "" {
		t.Errorf("expected empty summary for no clones, got %q", summary)
	}
}

func TestBuildClones(t *testing.T) {
	t.Parallel()

	content := testGoCode
	p, _ := htmlPrinterWithContent(content)
	hp := p.(*htmlprinter)

	nodes := nodesForHTML(testFilename)
	clones, err := hp.buildClones([][]*syntax.Node{nodes})
	if err != nil {
		t.Fatalf("buildClones failed: %v", err)
	}

	if len(clones) != 1 {
		t.Fatalf("expected 1 clone, got %d", len(clones))
	}

	if clones[0].filename != testFilename {
		t.Errorf("filename = %q, want 'test.go'", clones[0].filename)
	}
}

func TestBuildClones_EmptyDup(t *testing.T) {
	t.Parallel()

	p, _ := htmlPrinterWithContent("")
	hp := p.(*htmlprinter)

	_, err := hp.buildClones([][]*syntax.Node{{}})
	if err == nil {
		t.Error("expected error for empty node slice")
	}
}

func TestBuildClones_ReadError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	p := NewHTML(&buf, func(_ string) ([]byte, error) {
		return nil, errReadFail
	}, 15)
	hp := p.(*htmlprinter)

	nodes := nodesForHTML("missing.go")
	_, err := hp.buildClones([][]*syntax.Node{nodes})
	if err == nil {
		t.Error("expected error for read failure")
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
			"filter generated",
			ReportMetadata{FilterGenerated: true},
			"Filter Generated",
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

			if !strings.Contains(buf.String(), tt.contains) {
				t.Errorf("expected %q in output, got %q", tt.contains, buf.String())
			}
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
	if !strings.Contains(buf.String(), "Structural") {
		t.Error("expected Structural badge for default metadata")
	}
}

func TestHTMLErrorInWriter(t *testing.T) {
	t.Parallel()

	p := NewHTML(&errorWriter{}, mockReadFile(""), 15)

	err := p.PrintHeader()
	if err == nil {
		t.Error("expected error from failing writer")
	}
}

func TestHTMLErrorInPrintClones(t *testing.T) {
	t.Parallel()

	content := testGoCode
	p := NewHTML(&errorWriter{}, mockReadFile(content), 15)

	nodes := nodesForHTML(testFilename)
	err := p.PrintClones([][]*syntax.Node{nodes})
	if err == nil {
		t.Error("expected error from failing writer during PrintClones")
	}
}

func TestHTMLErrorInPrintFooter(t *testing.T) {
	t.Parallel()

	p := NewHTML(&errorWriter{}, mockReadFile(""), 15)

	err := p.PrintFooter()
	if err == nil {
		t.Error("expected error from failing writer during PrintFooter")
	}
}

func TestHTMLComputeCloneGroupDiff_Empty(t *testing.T) {
	t.Parallel()

	result := ComputeCloneGroupDiff(nil)
	if result.Base != nil {
		t.Error("expected nil Base for empty clones")
	}
}

func TestHTMLComputeCloneGroupDiff_SingleClone(t *testing.T) {
	t.Parallel()

	clones := []clone{
		{filename: "a.go", lineStart: 1, lineEnd: 5, fragment: []byte("line1\nline2\n")},
	}
	result := ComputeCloneGroupDiff(clones)
	if result.Base == nil {
		t.Fatal("expected non-nil Base")
	}

	if result.Base.Filename != "a.go" {
		t.Errorf("Base.Filename = %q, want 'a.go'", result.Base.Filename)
	}

	if len(result.Others) != 0 {
		t.Errorf("Others length = %d, want 0", len(result.Others))
	}
}

func TestHTMLComputeCloneGroupDiff_MultipleClones(t *testing.T) {
	t.Parallel()

	clones := []clone{
		{filename: "a.go", lineStart: 1, lineEnd: 5, fragment: []byte("line1\nline2\n")},
		{filename: "b.go", lineStart: 10, lineEnd: 14, fragment: []byte("line1\nmodified\n")},
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
		t.Error("expected no diff for identical content")
	}
}

func TestLineDiff_Different(t *testing.T) {
	t.Parallel()

	base := []byte("hello\nworld\n")
	compared := []byte("hello\nchanged\n")

	result := LineDiff(base, compared)
	if !result.HasDiff {
		t.Error("expected diff for different content")
	}

	found := false
	for _, line := range result.Base {
		if line.Type == DiffLineModified {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected at least one modified line")
	}
}

func TestLineDiff_DifferentLength(t *testing.T) {
	t.Parallel()

	base := []byte("line1\nline2\nline3\n")
	compared := []byte("line1\nline2\n")

	result := LineDiff(base, compared)
	if !result.HasDiff {
		t.Error("expected diff for different-length content")
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
		t.Error("expected diff for large files with different content")
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
	if !strings.Contains(result, "word-removed") {
		t.Error("expected word-removed class in diff")
	}

	if !strings.Contains(result, "word-added") {
		t.Error("expected word-added class in diff")
	}
}

func TestWordDiff_Identical(t *testing.T) {
	t.Parallel()

	result := WordDiff("hello world", "hello world")
	if strings.Contains(result, "word-removed") || strings.Contains(result, "word-added") {
		t.Error("expected no word-level diff for identical content")
	}
}

func TestHTMLEscape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name, input, want string
	}{
		{"ampersand", "a&b", "a&amp;b"},
		{"less than", "a<b", "a&lt;b"},
		{"greater than", "a>b", "a&gt;b"},
		{"quote", `a"b`, "a&quot;b"},
		{"combined", `<div class="x">&</div>`, "&lt;div class=&quot;x&quot;&gt;&amp;&lt;/div&gt;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := htmlEscape(tt.input); got != tt.want {
				t.Errorf("htmlEscape(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestHTMLWriteCloneOccurrences(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	hp := &htmlprinter{w: &buf, iota: 1}

	clones := []clone{
		{filename: "a.go", lineStart: 10, lineEnd: 20, fragment: []byte("code here")},
		{filename: "b.go", lineStart: 30, lineEnd: 40, fragment: []byte("more code")},
	}

	err := hp.writeCloneOccurrences(clones)
	if err != nil {
		t.Fatalf("writeCloneOccurrences failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "a.go") {
		t.Error("expected a.go in occurrence output")
	}

	if !strings.Contains(output, "b.go") {
		t.Error("expected b.go in occurrence output")
	}

	if !strings.Contains(output, "vscode://") {
		t.Error("Expected VSCode link in output")
	}
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
		t.Errorf("unexpected footer: %q", buf.String())
	}
}

func TestHTMLWriteDiffView_SingleClone(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	hp := &htmlprinter{w: &buf, iota: 1, diffMode: config.DiffModeSideBySide}

	clones := []clone{
		{filename: "a.go", lineStart: 1, lineEnd: 5, fragment: []byte("code\n")},
	}

	err := hp.writeDiffView(clones)
	if err != nil {
		t.Fatalf("writeDiffView with single clone failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "occurrence") {
		t.Error("single clone should fall through to writeCloneOccurrences")
	}
}

func TestHTMLWriteDiffView_MultipleClones(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	hp := &htmlprinter{w: &buf, iota: 1, diffMode: config.DiffModeSideBySide}

	clones := []clone{
		{filename: "a.go", lineStart: 1, lineEnd: 5, fragment: []byte("line1\nline2\n")},
		{filename: "b.go", lineStart: 10, lineEnd: 14, fragment: []byte("line1\nmodified\n")},
	}

	err := hp.writeDiffView(clones)
	if err != nil {
		t.Fatalf("writeDiffView with multiple clones failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "BASE REFERENCE") {
		t.Error("expected BASE REFERENCE header")
	}

	if !strings.Contains(output, "diff-legend") {
		t.Error("expected diff legend")
	}
}

func TestHTMLWriteDiffView_ThreeClonesWithSelector(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	hp := &htmlprinter{w: &buf, iota: 1, diffMode: config.DiffModeSideBySide}

	clones := []clone{
		{filename: "a.go", lineStart: 1, lineEnd: 5, fragment: []byte("line1\nline2\n")},
		{filename: "b.go", lineStart: 10, lineEnd: 14, fragment: []byte("line1\nchanged\n")},
		{filename: "c.go", lineStart: 20, lineEnd: 24, fragment: []byte("line1\nother\n")},
	}

	err := hp.writeDiffView(clones)
	if err != nil {
		t.Fatalf("writeDiffView with three clones failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "diff-select") {
		t.Error("expected diff selector dropdown for >2 clones")
	}

	if !strings.Contains(output, "Compare with...") {
		t.Error("expected selector placeholder text")
	}
}

func TestHTMLWriteDiffView_AggregateStats(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	hp := &htmlprinter{w: &buf, iota: 1, diffMode: config.DiffModeSideBySide}

	clones := []clone{
		{filename: "a.go", lineStart: 1, lineEnd: 5, fragment: []byte("line1\nline2\nline3\n")},
		{filename: "b.go", lineStart: 10, lineEnd: 14, fragment: []byte("line1\nmodified\nadded\n")},
	}

	err := hp.writeDiffView(clones)
	if err != nil {
		t.Fatalf("writeDiffView failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "diff-aggregate-stats") {
		t.Error("expected aggregate stats section")
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
	if !strings.Contains(output, "diff-toggle-5-side") {
		t.Error("expected side toggle button with group id 5")
	}

	if !strings.Contains(output, "diff-toggle-5-inline") {
		t.Error("expected inline toggle button with group id 5")
	}
}

func TestHTMLWriteDiffSelector(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	hp := &htmlprinter{w: &buf, iota: 3}

	groupDiff := CloneGroupDiff{
		Base: &CloneWithContent{
			CloneWithContentMixin: CloneWithContentMixin{Filename: "base.go", LineStart: 1},
			Content:               []byte("base\n"),
		},
		Others: []CloneDiff{
			{
				CloneWithContent: CloneWithContent{
					CloneWithContentMixin: CloneWithContentMixin{Filename: "other1.go", LineStart: 10},
					Content:               []byte("other1\n"),
				},
			},
			{
				CloneWithContent: CloneWithContent{
					CloneWithContentMixin: CloneWithContentMixin{Filename: "other2.go", LineStart: 20},
					Content:               []byte("other2\n"),
				},
			},
		},
	}

	err := hp.writeDiffSelector(groupDiff)
	if err != nil {
		t.Fatalf("writeDiffSelector failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "diff-select-3") {
		t.Error("expected selector with group id 3")
	}

	if !strings.Contains(output, "other1.go") {
		t.Error("expected other1.go in selector options")
	}

	if !strings.Contains(output, "other2.go") {
		t.Error("expected other2.go in selector options")
	}
}

func TestHTMLWriteDiffComparison(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	hp := &htmlprinter{w: &buf, iota: 1}

	base := &CloneWithContent{
		CloneWithContentMixin: CloneWithContentMixin{Filename: "base.go", LineStart: 1},
		Content:               []byte("line1\nline2\n"),
	}

	diff := LineDiff([]byte("line1\nline2\n"), []byte("line1\nmodified\n"))
	other := CloneDiff{
		CloneWithContent: CloneWithContent{
			CloneWithContentMixin: CloneWithContentMixin{Filename: "other.go", LineStart: 10},
			Content:               []byte("line1\nmodified\n"),
		},
		Diff: diff,
	}

	err := hp.writeDiffComparison(base, other, 0, 1)
	if err != nil {
		t.Fatalf("writeDiffComparison failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "diff-compare-1-0") {
		t.Error("expected diff comparison with id 1-0")
	}

	if !strings.Contains(output, "other.go") {
		t.Error("expected other.go in comparison header")
	}
}

func TestHTMLWriteDiffComparison_MultipleComparisons(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	hp := &htmlprinter{w: &buf, iota: 2}

	base := &CloneWithContent{
		CloneWithContentMixin: CloneWithContentMixin{Filename: "base.go", LineStart: 1},
		Content:               []byte("code\n"),
	}

	diff := LineDiff([]byte("code\n"), []byte("different\n"))
	other := CloneDiff{
		CloneWithContent: CloneWithContent{
			CloneWithContentMixin: CloneWithContentMixin{Filename: "other.go", LineStart: 5},
			Content:               []byte("different\n"),
		},
		Diff: diff,
	}

	err := hp.writeDiffComparison(base, other, 1, 2)
	if err != nil {
		t.Fatalf("writeDiffComparison failed: %v", err)
	}

	output := buf.String()
	// Second comparison (index>0) should NOT have active class
	if strings.Contains(output, "active") {
		t.Error("second comparison should not be active")
	}
}

func TestHTMLRenderDiffLines(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	hp := &htmlprinter{w: &buf}

	lines := []DiffLine{
		{Content: "equal line", Type: DiffLineEqual, LineNumber: 1},
		{Content: "added line", Type: DiffLineAdded, LineNumber: 2},
		{Content: "removed line", Type: DiffLineRemoved, LineNumber: 3},
		{Content: "modified line", Type: DiffLineModified, LineNumber: 4},
	}

	opposite := []DiffLine{
		{Content: "equal line", Type: DiffLineEqual, LineNumber: 1},
		{Content: "base added", Type: DiffLineAdded, LineNumber: 2},
		{Content: "base removed", Type: DiffLineRemoved, LineNumber: 3},
		{Content: "original modified", Type: DiffLineModified, LineNumber: 4},
	}

	err := hp.renderDiffLines(lines, opposite, true)
	if err != nil {
		t.Fatalf("renderDiffLines failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "equal") {
		t.Error("expected equal class")
	}

	if !strings.Contains(output, "added") {
		t.Error("expected added class")
	}

	if !strings.Contains(output, "removed") {
		t.Error("expected removed class")
	}

	if !strings.Contains(output, "modified") {
		t.Error("expected modified class")
	}
}

func TestHTMLRenderDiffLines_WordDiff(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	hp := &htmlprinter{w: &buf}

	lines := []DiffLine{
		{Content: "hello world", Type: DiffLineModified, LineNumber: 1},
	}
	opposite := []DiffLine{
		{Content: "hello earth", Type: DiffLineModified, LineNumber: 1},
	}

	err := hp.renderDiffLines(lines, opposite, true)
	if err != nil {
		t.Fatalf("renderDiffLines failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "word-removed") || !strings.Contains(output, "word-added") {
		t.Error("expected word-level diff highlighting for modified lines")
	}
}
