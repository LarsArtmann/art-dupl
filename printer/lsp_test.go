package printer

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	gofinding "github.com/larsartmann/go-finding"
)

// lspTestGroup builds a group at a severity-ladder rung: tokens vs threshold
// (4x error, 2x warning, else info/note).
func lspTestGroup(hash string, tokens int) domain.ProcessedCloneGroup {
	return domain.ProcessedCloneGroup{
		Hash: hash,
		Clones: []domain.ProcessedClone{
			{
				Filename:   "a.go",
				LineStart:  10,
				LineEnd:    19,
				Fragment:   "func a() {\n\tprintln(1)\n}",
				StartPos:   100,
				EndPos:     220,
				TokenCount: tokens,
				Classification: domain.CloneClassification{
					CloneType: domain.CloneType1,
				},
			},
			{
				Filename:   "b.go",
				LineStart:  30,
				LineEnd:    39,
				Fragment:   "func b() {\n\tprintln(2)\n}",
				StartPos:   300,
				EndPos:     420,
				TokenCount: tokens,
				Classification: domain.CloneClassification{
					CloneType: domain.CloneType2,
				},
			},
		},
	}
}

// TestLSPPrinterDocumentsAndRoundTrip pins the --lsp wire shape: one
// LSPDocument per file (sorted), one diagnostic per clone occurrence with
// the severity ladder converted to LSP severities, the go-finding GroupID in
// the diagnostic data, and FromLSP(fileURI, diag) restoring a Finding that
// keeps the GroupID (G6).
func TestLSPPrinterDocumentsAndRoundTrip(t *testing.T) {
	t.Parallel()

	const threshold = 5

	groups := []domain.ProcessedCloneGroup{
		lspTestGroup("err00000000000001", 40), // error
		lspTestGroup("warn0000000000001", 12), // warning
		lspTestGroup("note0000000000001", 6),  // note
	}

	var buf bytes.Buffer

	p := NewLSP(&buf, mockSARIFReadFile, threshold)
	if err := p.PrintHeader(); err != nil {
		t.Fatalf("PrintHeader: %v", err)
	}

	for _, group := range groups {
		if err := p.PrintClones(group); err != nil {
			t.Fatalf("PrintClones(%s): %v", group.Hash, err)
		}
	}

	if err := p.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter: %v", err)
	}

	var documents []LSPDocument
	if err := json.Unmarshal(buf.Bytes(), &documents); err != nil {
		t.Fatalf("unmarshal LSP output: %v\noutput:\n%s", err, buf.String())
	}

	if len(documents) != 2 {
		t.Fatalf("documents = %d, want 2 (a.go, b.go)", len(documents))
	}

	if documents[0].URI != "a.go" || documents[1].URI != "b.go" {
		t.Fatalf("documents not file-sorted: %q, %q", documents[0].URI, documents[1].URI)
	}

	wantSeverity := map[string]gofinding.LSPSeverity{
		"err00000000000001":  gofinding.LSPSeverityError,
		"warn00000000000001": gofinding.LSPSeverityWarning,
		"note00000000000001": gofinding.LSPSeverityInfo,
	}

	seen := map[string]bool{}

	for _, doc := range documents {
		if len(doc.Diagnostics) != len(groups) {
			t.Fatalf("%s: %d diagnostics, want %d (one per ladder-rung group)",
				doc.URI, len(doc.Diagnostics), len(groups))
		}

		for _, diag := range doc.Diagnostics {
			if diag.Code != "art-dupl/duplicate-code" {
				t.Errorf("%s: diagnostic code = %q", doc.URI, diag.Code)
			}

			if diag.Source != "art-dupl" {
				t.Errorf("%s: diagnostic source = %q", doc.URI, diag.Source)
			}

			if diag.Data == nil || diag.Data.GroupID == "" {
				t.Errorf("%s: diagnostic carries no data.GroupID", doc.URI)

				continue
			}

			if want, ok := wantSeverity[string(diag.Data.GroupID)]; ok && diag.Severity != want {
				t.Errorf("%s: group %s severity = %d, want %d",
					doc.URI, diag.Data.GroupID, diag.Severity, want)
			}

			// G6: FromLSP(fileURI, diag) restores the GroupID.
			restored := gofinding.FromLSP(gofinding.FilePath(doc.URI), diag)
			if restored.GroupID != diag.Data.GroupID {
				t.Errorf("FromLSP lost the GroupID: %q -> %q", diag.Data.GroupID, restored.GroupID)
			}

			seen[string(diag.Data.GroupID)] = true
		}
	}

	if len(seen) != len(groups) {
		t.Errorf("distinct GroupIDs across diagnostics = %d, want %d", len(seen), len(groups))
	}
}

// TestLSPPrinterDeduplicatesAndSkipsEmptyHash mirrors the SARIF printer's
// pipeline semantics: a repeated group hash emits nothing the second time
// and an empty-hash group is dropped entirely.
func TestLSPPrinterDeduplicatesAndSkipsEmptyHash(t *testing.T) {
	t.Parallel()

	group := lspTestGroup("dedup000000000001", 40)
	empty := lspTestGroup("", 40)

	var buf bytes.Buffer

	p := NewLSP(&buf, mockSARIFReadFile, 5).(*lspPrinter)
	_ = p.PrintClones(group)
	_ = p.PrintClones(group) // duplicate hash
	_ = p.PrintClones(empty) // no hash

	if got := len(p.byFile["a.go"]); got != 1 {
		t.Errorf("a.go diagnostics = %d, want 1 (hash dedup + empty-hash skip)", got)
	}
}

// TestLSPPrinterEmptyRunEmitsEmptyArray pins the no-clones wire shape: an
// empty JSON array, never nil — machine consumers parse one stable shape.
func TestLSPPrinterEmptyRunEmitsEmptyArray(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer

	p := NewLSP(&buf, mockSARIFReadFile, 5)
	if err := p.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter: %v", err)
	}

	if buf.String() != "[]" {
		t.Errorf("empty-run output = %q, want %q", buf.String(), "[]")
	}
}
