package printer

import (
	"bytes"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
)

// newTestHTMLPrinter creates a configured HTML printer for testing.
func newTestHTMLPrinter(t *testing.T) (Printer, *bytes.Buffer) {
	t.Helper()

	var buf bytes.Buffer

	printer := NewHTMLWithOptions(
		&buf,
		mockReadFile(""),
		config.DiffModeSideBySide,
		ReportMetadata{},
		15,
	)

	if err := printer.PrintHeader(); err != nil {
		t.Fatalf("PrintHeader failed: %v", err)
	}

	if err := printer.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter failed: %v", err)
	}

	return printer, &buf
}

// assertContains checks that s contains substr.
func assertContains(t *testing.T, s, substr, msg string) {
	t.Helper()

	if !strings.Contains(s, substr) {
		t.Error(msg)
	}
}

// assertStringEqual checks that actual equals expected.
func assertStringEqual(t *testing.T, actual, expected, msg string) {
	t.Helper()

	if actual != expected {
		t.Errorf("got '%s', want '%s'", actual, expected)
	}
}

func TestHTMLDiffRendering(t *testing.T) {
	_, buf := newTestHTMLPrinter(t)
	output := buf.String()

	// Verify CSS classes for word-level diff are present in template
	assertContains(t, output, `.word-added`, "Expected word-added CSS class definition")
	assertContains(t, output, `.word-removed`, "Expected word-removed CSS class definition")

	// Verify JavaScript for diff mode toggle is present in template
	assertContains(t, output, `toggleDiffView`, "Expected toggleDiffView JavaScript function")

	// Verify aggregate stats CSS is present
	assertContains(t, output, `.diff-aggregate-stats`, "Expected diff-aggregate-stats CSS class definition")
}

func TestHTMLInlineViewMode(t *testing.T) {
	_, buf := newTestHTMLPrinter(t)
	output := buf.String()

	// Verify inline view CSS is present
	assertContains(t, output, `.diff-content.inline`, "Expected diff-content.inline CSS class")

	// Verify JavaScript for localStorage handling
	assertContains(t, output, `localStorage.getItem('artdupl-diff-mode')`, "Expected localStorage handling for view mode preference")
}

func TestHTMLDiffJavaScriptFunctions(t *testing.T) {
	_, buf := newTestHTMLPrinter(t)
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
		assertStringEqual(t, string(actual), tc.expect, "toWhitespace() mismatch")
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
		assertStringEqual(t, string(actual), tc.expect, "deindent() mismatch")
	}
}
