package printer

import (
	"bytes"
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
)

// benchSARIFGroup builds a clone group with the given number of occurrences
// in distinct files, each carrying a fragment so the snippet path is hot.
func benchSARIFGroup(hash string, occurrences int) domain.ProcessedCloneGroup {
	group := domain.ProcessedCloneGroup{Hash: hash}

	for i := range occurrences {
		group.Clones = append(group.Clones, domain.ProcessedClone{
			Filename:   "bench" + string(rune('a'+i)) + ".go",
			LineStart:  10 + i,
			LineEnd:    20 + i,
			Fragment:   "func bench() {\n\tresult := compute(input)\n\treturn result\n}",
			TokenCount: 42,
			Classification: domain.CloneClassification{
				CloneType: domain.CloneType1,
				Category:  domain.CategoryBlock,
			},
		})
	}

	return group
}

// BenchmarkSARIFPrintClonesMulti measures the full SARIF serialization of one
// multi-occurrence group (relatedLocations fan-out + primary snippet), the
// path the 2026-10-11 wire corrections touched. Baseline for future
// benchstat comparisons; the detection hot path is guarded separately by the
// suffixtree/syntax alloc budgets.
func BenchmarkSARIFPrintClonesMulti(b *testing.B) {
	group := benchSARIFGroup("benchhash0000000001", 3)

	b.ReportAllocs()

	for b.Loop() {
		var buf bytes.Buffer

		p := newTestSARIFPrinter(&buf, "bench")

		if err := p.PrintHeader(); err != nil {
			b.Fatalf("PrintHeader: %v", err)
		}

		p.SetHash(group.Hash)

		if err := p.PrintClones(group); err != nil {
			b.Fatalf("PrintClones: %v", err)
		}

		if err := p.PrintFooter(); err != nil {
			b.Fatalf("PrintFooter: %v", err)
		}
	}
}
