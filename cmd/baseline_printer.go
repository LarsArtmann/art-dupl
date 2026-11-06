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
	stats         printer.SuppressionStats
}

func newBaselineFilterPrinter(inner printer.Printer, bf *baseline.File) *baselineFilterPrinter {
	return &baselineFilterPrinter{Printer: inner, baseline: bf}
}

// SetSuppressionStats keeps the counts the analysis pipeline reports and
// forwards them to the wrapped printer, so the summary line stays accurate
// when a check run suppresses groups for reasons other than the baseline
// (actionability, thresholds, accept directives).
func (p *baselineFilterPrinter) SetSuppressionStats(stats printer.SuppressionStats) {
	p.stats = stats

	if setter, ok := p.Printer.(printer.SuppressionStatsSetter); ok {
		setter.SetSuppressionStats(stats)
	}
}

// SetHash forwards the current group hash to the wrapped printer. The embedded
// Printer interface does not declare SetHash, so without this method the
// HashSetter assertion in printCloneGroups fails for the wrapper and every
// forwarded group reaches the inner printer with a stale or empty hash (which
// made `check --json` emit an empty hash for new groups).
func (p *baselineFilterPrinter) SetHash(hash string) {
	if hs, ok := p.Printer.(printer.HashSetter); ok {
		hs.SetHash(hash)
	}
}

// Split returns the baseline-vs-new accounting for the summary line: groups
// the baseline already accepts, groups that are new (the number the CI verdict
// hinges on), and groups suppressed elsewhere in the pipeline.
func (p *baselineFilterPrinter) Split() printer.BaselineSplit {
	// Shown counts every group handed to PrintClones, baseline-matched ones
	// included (they are silently dropped here), so the pipeline-suppressed
	// count is DetectedTotal - Shown. Subtracting the baseline count again
	// would double-count it.
	suppressed := max(p.stats.DetectedTotal-p.stats.Shown, 0)

	return printer.BaselineSplit{
		InBaseline: p.suppressed,
		New:        p.newCloneCount,
		Suppressed: suppressed,
	}
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
