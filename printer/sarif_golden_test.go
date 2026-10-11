package printer

import (
	"bytes"
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

// TestSARIFOutputGolden locks the exact byte shape of the SARIF wire format
// after the 2026-10-11 M5/M6/G7 corrections: relatedLocations link every
// sibling occurrence (uri + line range + clone-of message, NO snippet —
// upstream parity), the primary region carries the fragment snippet, and the
// category property is emitted only when non-empty. This golden is the
// freeze consumers can diff against when the wire next moves.
//
// Regenerate after intentional wire changes via:
//
//	go test -update -run TestSARIFOutputGolden ./printer/
func TestSARIFOutputGolden(t *testing.T) {
	var buf bytes.Buffer

	printer := newTestSARIFPrinter(&buf, "1.0.0")

	multiClone := domain.ProcessedCloneGroup{
		Hash: "abc123def4567890",
		Clones: []domain.ProcessedClone{
			newTestProcessedClone("alpha.go", 10, 15, "result := x + y\nreturn result"),
			newTestProcessedClone("beta.go", 20, 25, "sum := a + b\nreturn sum"),
			newTestProcessedClone("gamma.go", 30, 35, "total := p + q\nreturn total"),
		},
	}
	for i := range multiClone.Clones {
		multiClone.Clones[i].Classification = domain.CloneClassification{
			CloneType: domain.CloneType1,
			Category:  domain.CategoryBlock,
		}
	}

	if err := printer.PrintHeader(); err != nil {
		t.Fatalf("PrintHeader() error: %v", err)
	}

	printer.SetHash(multiClone.Hash)

	if err := printer.PrintClones(multiClone); err != nil {
		t.Fatalf("PrintClones(multiClone) error: %v", err)
	}

	single := multiClone
	single.Hash = "single0000000001"
	single.Clones = []domain.ProcessedClone{
		newTestProcessedClone("solo.go", 5, 9, "println(\"only\")"),
	}
	single.Clones[0].Classification = domain.CloneClassification{
		CloneType: domain.CloneType1,
		Category:  domain.CategoryBlock,
	}

	printer.SetHash(single.Hash)

	if err := printer.PrintClones(single); err != nil {
		t.Fatalf("PrintClones(single) error: %v", err)
	}

	if err := printer.PrintFooter(); err != nil {
		t.Fatalf("PrintFooter() error: %v", err)
	}

	testutil.RequireGolden(t, buf.Bytes())
}
