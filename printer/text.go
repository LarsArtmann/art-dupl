package printer

import (
	"fmt"
	"io"
	"sort"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	duplerrors "github.com/LarsArtmann/art-dupl/errors"
)

type TextPrinter struct {
	ReadFile

	cnt           int
	w             io.Writer
	totalSize     int
	cloneGroups   [][]domain.ProcessedClone
	currentHash   string
	isFileDupe    bool
	richText      bool
	diffHintFiles []string
}

func (p *TextPrinter) SetHash(hash string) {
	p.currentHash = hash
}

func (p *TextPrinter) SetFileDuplicate(isDupe bool) {
	p.isFileDupe = isDupe
}

func (p *TextPrinter) SetRichText(enabled bool) {
	p.richText = enabled
}

func NewText(w io.Writer, fread ReadFile) Printer {
	return &TextPrinter{w: w, ReadFile: fread, cloneGroups: make([][]domain.ProcessedClone, 0)}
}

func (p *TextPrinter) PrintHeader() error { return nil }

func (p *TextPrinter) PrintClones(
	group domain.ProcessedCloneGroup,
	sortBy ...config.SortCriteria,
) error {
	p.cnt++

	clones := group.Clones
	SortProcessedClonesByCriteria(clones, ExtractSortCriteria(sortBy...))

	groupCloneSize := totalFragmentSize(clones)

	isFileDupe := p.isFileDupe

	if err := p.writeGroupHeader(isFileDupe, clones); err != nil {
		return duplerrors.Wrap(err, duplerrors.InternalError, "write group header")
	}

	p.cloneGroups = append(p.cloneGroups, clones)
	p.totalSize += groupCloneSize

	sort.Sort(byNameAndLineProcessed(clones))

	return p.printCloneList(clones)
}

func (p *TextPrinter) writeGroupHeader(isFileDupe bool, clones []domain.ProcessedClone) error {
	if !isFileDupe || p.currentHash == "" {
		return p.writeCloneHeader(clones)
	}

	return p.writeFileDupeHeader(clones)
}

func (p *TextPrinter) writeFileDupeHeader(clones []domain.ProcessedClone) error {
	hashPrefix := p.currentHash
	if len(hashPrefix) > 12 {
		hashPrefix = hashPrefix[:12]
	}

	fileSizeStr := formatBytes(clones[0].FileSize)
	if _, err := fmt.Fprintf(
		p.w,
		"\U0001f4c4 FILE DUPLICATE | \U0001f517 [%s...] | %d files | %s\n\n",
		hashPrefix,
		len(clones),
		fileSizeStr,
	); err != nil {
		return fmt.Errorf(
			"write file duplicate header (hash: %s, files: %d): %w",
			hashPrefix,
			len(clones),
			err,
		)
	}

	for _, cl := range clones {
		p.diffHintFiles = append(p.diffHintFiles, cl.Filename)
	}

	return nil
}

func (p *TextPrinter) writeCloneHeader(clones []domain.ProcessedClone) error {
	if p.richText && len(clones) > 0 {
		return p.writeRichGroupHeader(len(clones), clones[0].Classification)
	}

	if _, err := fmt.Fprintf(p.w, "found %d clones:\n", len(clones)); err != nil {
		return fmt.Errorf("write clone count header (%d clones): %w", len(clones), err)
	}

	return nil
}

func (p *TextPrinter) PrintFooter() error {
	if _, err := fmt.Fprintf(p.w, "\nFound total %d clone groups.\n", p.cnt); err != nil {
		return err
	}

	if len(p.diffHintFiles) >= 2 {
		file1 := p.diffHintFiles[0]
		file2 := p.diffHintFiles[1]

		if _, err := fmt.Fprintf(p.w, "\n\u2192 diff %s %s\n", file1, file2); err != nil {
			return err
		}
	}

	return nil
}

func (p *TextPrinter) printCloneList(clones []domain.ProcessedClone) error {
	return writeCloneLines(p.w, clones, "  %s:%d-%d\n")
}

func (p *TextPrinter) writeRichGroupHeader(count int, cls domain.CloneClassification) error {
	actionabilityBadge := ""
	if cls.Actionability == domain.NonActionable {
		actionabilityBadge = " [non-actionable]"
	}

	if _, err := fmt.Fprintf(
		p.w,
		"found %d clones: [%s] [%s] %s%s (%d tokens, %d lines) suggestion: %s\n",
		count,
		cls.Priority,
		cls.CloneType,
		cls.Category,
		actionabilityBadge,
		cls.Tokens,
		cls.Lines,
		cls.Suggestion,
	); err != nil {
		return duplerrors.Wrapf(
			err,
			duplerrors.IOError,
			"write rich group header (count=%d, category=%q, priority=%q)",
			count, cls.Category, cls.Priority,
		)
	}

	return nil
}

func writeCloneLines(w io.Writer, clones []domain.ProcessedClone, format string) error {
	for _, cl := range clones {
		if _, err := fmt.Fprintf(w, format, cl.Filename, cl.LineStart, cl.LineEnd); err != nil {
			return err
		}
	}

	return nil
}

func (p *TextPrinter) OutputText(threshold int, sortBy config.SortCriteria) error {
	sortedCloneGroups := make([][]domain.ProcessedClone, len(p.cloneGroups))
	copy(sortedCloneGroups, p.cloneGroups)

	sortGroupsByCriteria(sortedCloneGroups, sortBy, GroupMetrics[[]domain.ProcessedClone]{
		Size:  totalFragmentSize,
		Count: func(g []domain.ProcessedClone) int { return len(g) },
		SortKey: func(g []domain.ProcessedClone) string {
			if len(g) == 0 {
				return ""
			}

			return g[0].Filename
		},
	})

	err := p.PrintHeader()
	if err != nil {
		return fmt.Errorf(
			"print header (threshold: %d, sortBy: %s): %w",
			threshold,
			sortBy.String(),
			err,
		)
	}

	for _, cloneGroup := range sortedCloneGroups {
		for _, cl := range cloneGroup {
			if len(cl.Fragment) > 0 {
				if _, err := fmt.Fprintf(
					p.w,
					"%s\n%s:%d-%d\n\n",
					cl.Fragment,
					cl.Filename,
					cl.LineStart,
					cl.LineEnd,
				); err != nil {
					return cloneWriteErr(cl, "fragment", err)
				}
			} else {
				if err := writeCloneLines(
					p.w,
					[]domain.ProcessedClone{cl},
					"%s:%d-%d\n",
				); err != nil {
					return cloneWriteErr(cl, "line", err)
				}
			}
		}
	}

	return p.PrintFooter()
}

// cloneWriteErr wraps a write error with a "write clone X filename:line-line" prefix.
func cloneWriteErr(cl domain.ProcessedClone, kind string, err error) error {
	return fmt.Errorf("write clone %s %s:%d-%d: %w", kind, cl.Filename, cl.LineStart, cl.LineEnd, err)
}

func formatBytes(bytes int) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

// totalFragmentSize returns the total byte length of all clone fragments.
// This is used for display ordering by fragment size and is non-mutating.
func totalFragmentSize(clones []domain.ProcessedClone) int {
	total := 0
	for _, cl := range clones {
		total += len(cl.Fragment)
	}

	return total
}

type byNameAndLineProcessed []domain.ProcessedClone

func (c byNameAndLineProcessed) Len() int      { return len(c) }
func (c byNameAndLineProcessed) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
func (c byNameAndLineProcessed) Less(i, j int) bool {
	return compareByNameThenPos(c[i].Filename, c[j].Filename, c[i].LineStart, c[j].LineStart)
}
