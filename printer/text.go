package printer

import (
	"fmt"
	"io"
	"sort"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
)

type TextPrinter struct {
	ReadFile

	cnt           int
	w             io.Writer
	totalSize     int
	cloneGroups   [][]domain.ProcessedClone
	currentHash   string
	isFileDupe    bool
	diffHintFiles []string
}

func (p *TextPrinter) SetHash(hash string) {
	p.currentHash = hash
}

func (p *TextPrinter) SetFileDuplicate(isDupe bool) {
	p.isFileDupe = isDupe
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

	groupCloneSize := calculateProcessedCloneSizes(clones)

	isFileDupe := p.isFileDupe

	if isFileDupe && p.currentHash != "" {
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
			return err
		}

		for _, cl := range clones {
			p.diffHintFiles = append(p.diffHintFiles, cl.Filename)
		}
	} else {
		if _, err := fmt.Fprintf(p.w, "found %d clones:\n", len(clones)); err != nil {
			return err
		}
	}

	p.cloneGroups = append(p.cloneGroups, clones)
	p.totalSize += groupCloneSize

	sort.Sort(byNameAndLineProcessed(clones))

	return p.printCloneList(clones)
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
	for _, cl := range clones {
		if _, err := fmt.Fprintf(
			p.w,
			"  %s:%d,%d\n",
			cl.Filename,
			cl.LineStart,
			cl.LineEnd,
		); err != nil {
			return err
		}
	}

	return nil
}

func (p *TextPrinter) OutputText(threshold int, sortBy config.SortCriteria) error {
	sortedCloneGroups := make([][]domain.ProcessedClone, len(p.cloneGroups))
	copy(sortedCloneGroups, p.cloneGroups)

	switch sortBy {
	case config.SortBySize:
		sort.Slice(sortedCloneGroups, func(i, j int) bool {
			return totalFragmentSize(sortedCloneGroups[i]) > totalFragmentSize(sortedCloneGroups[j])
		})
	case config.SortByOccurrence:
		sort.Slice(sortedCloneGroups, func(i, j int) bool {
			return len(sortedCloneGroups[i]) > len(sortedCloneGroups[j])
		})
	case config.SortByHash:
		sort.Slice(sortedCloneGroups, func(i, j int) bool {
			if len(sortedCloneGroups[i]) == 0 || len(sortedCloneGroups[j]) == 0 {
				return false
			}

			return sortedCloneGroups[i][0].Filename < sortedCloneGroups[j][0].Filename
		})
	case config.SortByTotalTokens:
		sort.Slice(sortedCloneGroups, func(i, j int) bool {
			return totalFragmentSize(sortedCloneGroups[i])*len(sortedCloneGroups[i]) >
				totalFragmentSize(sortedCloneGroups[j])*len(sortedCloneGroups[j])
		})
	default:
		sort.Slice(sortedCloneGroups, func(i, j int) bool {
			return totalFragmentSize(sortedCloneGroups[i]) > totalFragmentSize(sortedCloneGroups[j])
		})
	}

	err := p.PrintHeader()
	if err != nil {
		return err
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
					return err
				}
			} else {
				if _, err := fmt.Fprintf(
					p.w,
					"%s:%d,%d\n",
					cl.Filename,
					cl.LineStart,
					cl.LineEnd,
				); err != nil {
					return err
				}
			}
		}
	}

	return p.PrintFooter()
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

func calculateProcessedCloneSizes(clones []domain.ProcessedClone) int {
	total := 0
	for i := range clones {
		clones[i].Size = len(clones[i].Fragment)
		total += clones[i].Size
	}

	return total
}

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
