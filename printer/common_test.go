package printer

import (
	"bytes"
	"testing"

	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// newTestProcessedClone creates a domain.ProcessedClone for testing.
func newTestProcessedClone(filename string, lineStart, lineEnd int, fragment string) domain.ProcessedClone {
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

// assertBufferContains checks that a buffer's contents contain the expected substring.
func assertBufferContains(t *testing.T, buf *bytes.Buffer, expected, msg string) {
	t.Helper()

	testutil.AssertStringContains(t, buf.String(), expected, msg)
}

// processTestNodes is a convenience function that converts nodes to a ProcessedCloneGroup
// for use in tests. Uses the provided fread function.
func processTestNodes(fread ReadFile, hash string, dups [][]*syntax.Node) domain.ProcessedCloneGroup {
	group, err := NodesToGroup(fread, hash, dups)
	if err != nil {
		return domain.ProcessedCloneGroup{Hash: hash, Clones: []domain.ProcessedClone{}}
	}

	return group
}
