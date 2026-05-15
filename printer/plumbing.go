package printer

import (
	"fmt"
	"io"
	"sort"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
)

type plumbing struct {
	ReadFile

	w io.Writer
}

func NewPlumbing(w io.Writer, fread ReadFile) Printer {
	return &plumbing{ReadFile: fread, w: w}
}

func (p *plumbing) PrintHeader() error { return nil }

func (p *plumbing) PrintClones(group domain.ProcessedCloneGroup, sortBy ...config.SortCriteria) error {
	clones := group.Clones
	SortProcessedClonesByCriteria(clones, ExtractSortCriteria(sortBy...))
	sort.Sort(byNameAndLineProcessed(clones))

	for _, cl := range clones {
		if _, err := fmt.Fprintf(p.w, "%s:%d-%d\n", cl.Filename, cl.LineStart, cl.LineEnd); err != nil {
			return err
		}
	}

	return nil
}

func (p *plumbing) PrintFooter() error { return nil }

func (p *plumbing) OutputPlumbing(threshold int, sortBy config.SortCriteria) error {
	_, _ = fmt.Fprintf(p.w, "# Plumbing output sorted by %s\n", sortBy.String())

	return nil
}
