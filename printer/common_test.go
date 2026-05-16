package printer

import (
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// newTestProcessedClone creates a domain.ProcessedClone for testing.
func newTestProcessedClone(
	filename string,
	lineStart, lineEnd int,
	fragment string,
) domain.ProcessedClone {
	return domain.ProcessedClone{
		Filename:  filename,
		LineStart: lineStart,
		LineEnd:   lineEnd,
		Fragment:  []byte(fragment),
	}
}

// processedCloneFixture creates a clone fixture for testing.
func processedCloneFixture(filename string, start, end int, fragment string) domain.ProcessedClone {
	return domain.ProcessedClone{
		Filename:  filename,
		LineStart: start,
		LineEnd:   end,
		Fragment:  []byte(fragment),
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
