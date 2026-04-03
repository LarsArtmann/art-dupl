package printer

import (
	"bytes"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
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

	err := printer.PrintFooter()
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

	err := printer.PrintFooter()
	if err != nil {
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
