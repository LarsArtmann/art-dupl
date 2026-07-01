package printer

import (
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
)

var printerMethodChecks = []struct {
	name string
	call func(p Printer) error
}{
	{"PrintHeader", func(p Printer) error { return p.PrintHeader() }},
	{"PrintFooter", func(p Printer) error { return p.PrintFooter() }},
}

// newTestProcessedClone creates a domain.ProcessedClone for testing.
func newTestProcessedClone(
	filename string,
	lineStart, lineEnd int,
	fragment string,
) domain.ProcessedClone {
	return domain.ProcessedClone{
		CloneRef: domain.CloneRef{
			Filename:  filename,
			LineStart: lineStart,
			LineEnd:   lineEnd,
			Fragment:  fragment,
		},
	}
}

// processTestNodes is a convenience function that converts nodes to a ProcessedCloneGroup
// for use in tests. Uses the provided fread function.
func processTestNodes(
	fread ReadFile,
	hash string,
	dups [][]*syntax.Node,
) domain.ProcessedCloneGroup {
	group, err := NodesToGroup(fread, hash, dups)
	if err != nil {
		return domain.ProcessedCloneGroup{Hash: hash, Clones: []domain.ProcessedClone{}}
	}

	return group
}

// printTestClones calls PrintClones with the standard test node processing pipeline.
func printTestClones(p Printer, fread ReadFile, dups [][]*syntax.Node) error {
	return p.PrintClones(processTestNodes(fread, "test", dups))
}
