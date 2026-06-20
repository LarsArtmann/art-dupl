package cmd

import (
	"fmt"

	"github.com/LarsArtmann/art-dupl/baseline"
	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/printer"
)

// baselineFilterPrinter wraps a Printer and suppresses clone groups whose hash
// is already accepted in the baseline. Only NEW clones (hashes not in the
// baseline) are forwarded to the inner printer. Used by the `check` command.
type baselineFilterPrinter struct {
	printer.Printer

	baseline      *baseline.File
	newCloneCount int
	suppressed    int
}

func newBaselineFilterPrinter(inner printer.Printer, bf *baseline.File) *baselineFilterPrinter {
	return &baselineFilterPrinter{Printer: inner, baseline: bf}
}

func (p *baselineFilterPrinter) PrintClones(
	group domain.ProcessedCloneGroup,
	sortBy ...config.SortCriteria,
) error {
	if p.baseline.Has(group.Hash) {
		p.suppressed++

		return nil
	}

	p.newCloneCount++

	err := p.Printer.PrintClones(group, sortBy...)
	if err != nil {
		return fmt.Errorf("print clone group %s: %w", group.Hash, err)
	}

	return nil
}

// baselineRecorderPrinter collects clone-group hashes into a baseline.File
// without producing output. Used by the `baseline` command to snapshot the
// currently-accepted clones.
type baselineRecorderPrinter struct {
	bf *baseline.File
}

func newBaselineRecorderPrinter(bf *baseline.File) *baselineRecorderPrinter {
	return &baselineRecorderPrinter{bf: bf}
}

func (p *baselineRecorderPrinter) PrintHeader() error { return nil }

func (p *baselineRecorderPrinter) PrintClones(
	group domain.ProcessedCloneGroup,
	_ ...config.SortCriteria,
) error {
	files := make([]string, 0, len(group.Clones))
	for _, c := range group.Clones {
		files = append(files, c.Filename)
	}

	p.bf.Add(group.Hash, files, group.TokenCount)

	return nil
}

func (p *baselineRecorderPrinter) PrintFooter() error { return nil }
