package printer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
)

func TestHTMLDiffRendering(t *testing.T) {
	var buf bytes.Buffer

	printer := NewHTMLWithOptions(
		&buf,
		mockReadFile(""),
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

	output := buf.String()

	// Verify CSS classes for word-level diff are present in template
	if !strings.Contains(output, `.word-added`) {
		t.Error("Expected word-added CSS class definition")
	}

	if !strings.Contains(output, `.word-removed`) {
		t.Error("Expected word-removed CSS class definition")
	}

	// Verify JavaScript for diff mode toggle is present in template
	if !strings.Contains(output, `toggleDiffView`) {
		t.Error("Expected toggleDiffView JavaScript function")
	}

	// Verify aggregate stats CSS is present
	if !strings.Contains(output, `.diff-aggregate-stats`) {
		t.Error("Expected diff-aggregate-stats CSS class definition")
	}
}

func TestHTMLInlineViewMode(t *testing.T) {
	var buf bytes.Buffer

	printer := NewHTMLWithOptions(
		&buf,
		mockReadFile(""),
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

	output := buf.String()

	// Verify inline view CSS is present
	if !strings.Contains(output, `.diff-content.inline`) {
		t.Error("Expected diff-content.inline CSS class")
	}

	// Verify JavaScript for localStorage handling
	if !strings.Contains(output, `localStorage.getItem('artdupl-diff-mode')`) {
		t.Error("Expected localStorage handling for view mode preference")
	}
}

func TestHTMLDiffJavaScriptFunctions(t *testing.T) {
	var buf bytes.Buffer

	printer := NewHTMLWithOptions(
		&buf,
		mockReadFile(""),
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

	output := buf.String()

	// Verify all JavaScript functions are present
	testCases := []string{
		"function toggleDiffView",
		"function showDiff",
		"function copyCode",
		"document.addEventListener('DOMContentLoaded'",
		"localStorage.setItem('artdupl-diff-mode'",
	}

	for _, tc := range testCases {
		if !strings.Contains(output, tc) {
			t.Errorf("Expected JavaScript function/pattern '%s' in output", tc)
		}
	}
}

func TestToWhitespace(t *testing.T) {
	testCases := []struct {
		in     string
		expect string
	}{
		{"\t   ", "\t   "},
		{"\tčřď", "\t   "},
		{"  \ta", "  \t "},
	}

	for _, tc := range testCases {
		actual := toWhitespace([]byte(tc.in))
		if tc.expect != string(actual) {
			t.Errorf("got '%s', want '%s'", actual, tc.expect)
		}
	}
}

func TestDeindent(t *testing.T) {
	testCases := []struct {
		in     string
		expect string
	}{
		{"\t$\n\t\t$\n\t$", "$\n\t$\n$"},
		{"\t$\r\n\t\t$\r\n\t$", "$\r\n\t$\r\n$"},
		{"\t$\n\t\t$\n", "$\n\t$\n"},
		{"\t$\n\n\t\t$", "$\n\n\t$"},
	}
	for _, tc := range testCases {
		actual := deindent([]byte(tc.in))
		if tc.expect != string(actual) {
			t.Errorf("got '%s', want '%s'", actual, tc.expect)
		}
	}
}
