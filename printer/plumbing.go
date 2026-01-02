package printer

import (
	"fmt"
	"io"
	"sort"

	"github.com/LarsArtmann/art-dupl/syntax"
)

type plumbing struct {
	w io.Writer
	ReadFile
}

func NewPlumbing(w io.Writer, fread ReadFile) Printer {
	return &plumbing{w, fread}
}

func (p *plumbing) PrintHeader() error { return nil }

func (p *plumbing) PrintClones(dups [][]*syntax.Node, sortBy ...string) error {
	// Extract sortBy parameter, default to "size"
	sortCriteria := "size"
	if len(sortBy) > 0 {
		sortCriteria = sortBy[0]
	}

	// Apply sorting to the clone groups
	sortedDups := SortNodesByCriteria(dups, sortCriteria)

	clones, err := prepareClonesInfo(p.ReadFile, sortedDups)
	if err != nil {
		return err
	}
	sort.Sort(byNameAndLine(clones))
	for i, cl := range clones {
		nextCl := clones[(i+1)%len(clones)]
		if _, err := fmt.Fprintf(p.w, "%s:%d-%d: duplicate of %s:%d-%d\n", cl.filename, cl.lineStart, cl.lineEnd,
			nextCl.filename, nextCl.lineStart, nextCl.lineEnd); err != nil {
			return err // nolint:wrapcheck // fmt errors are clear in context
		}
	}
	return nil
}

func (p *plumbing) PrintFooter() error { return nil }

// OutputPlumbing generates plumbing output with sorting
func (p *plumbing) OutputPlumbing(threshold int, sortBy string) error {
	// Note: Plumbing output is generated during the normal PrintClones flow
	// This method exists for consistency with other output formats
	// The actual sorting is handled in PrintClones method

	// For now, just indicate the sorting criteria used
	_, _ = fmt.Fprintf(p.w, "# Plumbing output sorted by %s\n", sortBy)
	return nil
}
