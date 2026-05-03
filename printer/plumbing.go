package printer

import (
	"fmt"
	"io"
	"sort"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/syntax"
)

type plumbing struct {
	ReadFile

	w io.Writer
}

func NewPlumbing(w io.Writer, fread ReadFile) Printer {
	return &plumbing{ReadFile: fread, w: w}
}

func (p *plumbing) PrintHeader() error { return nil }

func (p *plumbing) PrintClones(dups [][]*syntax.Node, sortBy ...config.SortCriteria) error {
	// Apply sorting to the clone groups
	sortedDups := SortNodesByCriteria(dups, ExtractSortCriteria(sortBy...))

	clones, err := prepareClonesInfo(p.ReadFile, sortedDups)
	if err != nil {
		return err
	}

	sort.Sort(byNameAndLine(clones))

	return writeCloneLines(p.w, clones, "%s:%d-%d")
}

func (p *plumbing) PrintFooter() error { return nil }

// OutputPlumbing generates plumbing output with sorting.
func (p *plumbing) OutputPlumbing(threshold int, sortBy config.SortCriteria) error {
	// Note: Plumbing output is generated during the normal PrintClones flow
	// This method exists for consistency with other output formats
	// The actual sorting is handled in PrintClones method

	// For now, just indicate the sorting criteria used
	_, _ = fmt.Fprintf(p.w, "# Plumbing output sorted by %s\n", sortBy.String())

	return nil
}
