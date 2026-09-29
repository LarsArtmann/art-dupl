package printer

import (
	"bytes"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

func TestHTMLOutputGolden(t *testing.T) {
	var buf bytes.Buffer

	printer := NewHTMLWithOptions(
		&buf,
		mockReadFile(testFileContent),
		config.DiffModeSideBySide,
		ReportMetadata{},
		15,
	)

	err := printer.PrintHeader()
	if err != nil {
		t.Fatalf("PrintHeader failed: %v", err)
	}

	err = printer.PrintFooter()
	if err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	testutil.RequireGolden(t, buf.Bytes())
}

func TestHTMLOutputGoldenNoDiff(t *testing.T) {
	var buf bytes.Buffer

	printer := NewHTMLWithOptions(
		&buf,
		mockReadFile(testFileContent),
		config.DiffModeDisabled,
		ReportMetadata{},
		15,
	)

	err := printer.PrintHeader()
	if err != nil {
		t.Fatalf("PrintHeader failed: %v", err)
	}

	err = printer.PrintFooter()
	if err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	testutil.RequireGolden(t, buf.Bytes())
}

// TestHTMLOutputGoldenWithCloneGroup renders one REAL clone group (two
// occurrences, diff mode on) between header and footer. Header+footer-only
// goldens never exercised the templ clone-group/diff templates, so HTML
// changes there shipped golden-blind (status 2026-08-16_17-37 f21).
func TestHTMLOutputGoldenWithCloneGroup(t *testing.T) {
	var buf bytes.Buffer

	printer := NewHTMLWithOptions(
		&buf,
		mockReadFile(testFileContent),
		config.DiffModeSideBySide,
		ReportMetadata{},
		15,
	)

	if err := printer.PrintHeader(); err != nil {
		t.Fatalf("PrintHeader failed: %v", err)
	}

	group := domain.ProcessedCloneGroup{
		Hash:       "abc123def4567890",
		TokenCount: 42,
		Clones: []domain.ProcessedClone{
			{
				//nolint:modernize,embedlit // AGENTS.md mandates nested CloneRef initialization
				CloneRef: domain.CloneRef{
					Filename:  "foo.go",
					LineStart: 5,
					LineEnd:   7,
					Fragment:  "fmt.Println(\"hello\")",
				},
				StartPos:   50,
				EndPos:     120,
				TokenCount: 21,
				Classification: domain.CloneClassification{
					CloneType: domain.CloneType1,
					Category:  domain.CategoryCall,
					Priority:  domain.PriorityMedium,
				},
			},
			{
				//nolint:modernize,embedlit // AGENTS.md mandates nested CloneRef initialization
				CloneRef: domain.CloneRef{
					Filename:  "bar.go",
					LineStart: 10,
					LineEnd:   12,
					Fragment:  "fmt.Println(\"hello\")",
				},
				StartPos:   95,
				EndPos:     165,
				TokenCount: 21,
				Classification: domain.CloneClassification{
					CloneType: domain.CloneType1,
					Category:  domain.CategoryCall,
					Priority:  domain.PriorityMedium,
				},
			},
		},
	}

	if err := printer.PrintClones(group); err != nil {
		t.Fatalf("PrintClones failed: %v", err)
	}

	if err := printer.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	testutil.RequireGolden(t, buf.Bytes())
}

// testFileContent is a small Go file used for golden file testing.
const testFileContent = `package main

import "fmt"

func foo() {
	fmt.Println("hello")
}

func bar() {
	fmt.Println("hello")
}
`
