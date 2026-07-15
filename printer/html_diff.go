package printer

import (
	"context"

	"github.com/LarsArtmann/art-dupl/domain"
)

func (p *htmlprinter) writeDiffView(clones []domain.ProcessedClone) error {
	groupDiff := ComputeCloneGroupDiff(clones)

	if len(groupDiff.Others) == 0 {
		return p.writeCloneOccurrences(clones)
	}

	diffData := toDiffView(p.iota, groupDiff)

	return diffViewContent(diffData).Render(context.Background(), p.w)
}

func (p *htmlprinter) writeDiffViewToggle() error {
	return diffViewToggle(p.iota).Render(context.Background(), p.w)
}

func (p *htmlprinter) writeDiffSelector(groupDiff CloneGroupDiff) error {
	diffData := DiffView{
		GroupNum: p.iota,
		Base:     groupDiff.Base,
		Others:   groupDiff.Others,
	}

	return diffSelectorTempl(diffData).Render(context.Background(), p.w)
}

func (p *htmlprinter) writeDiffComparison(
	_ *CloneWithContent,
	other CloneDiff,
	index, total int,
) error {
	return diffComparison(p.iota, other, index, total).Render(context.Background(), p.w)
}

func (p *htmlprinter) renderDiffLines(lines, oppositeLines []DiffLine, isBasePanel bool) error {
	return renderDiffLinesTempl(lines, oppositeLines, isBasePanel).Render(context.Background(), p.w)
}

// countDiffStats counts added, removed, and modified lines in a DiffResult.
// Alias for countDiffLineStats; preserves the templ-facing name used by
// printer/report.templ.
func countDiffStats(diff DiffResult) (int, int, int) {
	return countDiffLineStats(diff)
}
