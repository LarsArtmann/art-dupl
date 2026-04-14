package printer

import (
	"fmt"
	"io"
	"sort"

	"github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// TextPrinter implements text-based output for duplicate detection.
type TextPrinter struct {
	ReadFile

	cnt           int
	w             io.Writer
	totalSize     int
	cloneGroups   [][]clone
	currentHash   string   // Hash for the current clone group (for hash detection)
	isFileDupe    bool     // True if current group is an entire file duplicate
	diffHintFiles []string // Files to show diff hint for
}

// SetHash sets the hash for the current clone group.
func (p *TextPrinter) SetHash(hash string) {
	p.currentHash = hash
}

// SetFileDuplicate marks the current group as a file duplicate.
func (p *TextPrinter) SetFileDuplicate(isDupe bool) {
	p.isFileDupe = isDupe
}

func NewText(w io.Writer, fread ReadFile) Printer {
	return &TextPrinter{w: w, ReadFile: fread, cloneGroups: make([][]clone, 0)}
}

func (p *TextPrinter) PrintHeader() error { return nil }

// detectFileDuplicate checks if all fragments are entire file duplicates.
func detectFileDuplicate(isFileDupeFlag bool, dups [][]*syntax.Node) bool {
	if isFileDupeFlag {
		return true
	}

	if len(dups) == 0 {
		return false
	}

	for _, frag := range dups {
		if len(frag) != 1 || frag[0].Pos != 0 {
			return false
		}
	}

	return true
}

// calculateCloneSizes calculates fragment sizes and total size for clones.
func calculateCloneSizes(clones []clone) int {
	totalSize := 0

	for i := range clones {
		clones[i].size = len(clones[i].fragment)
		totalSize += clones[i].size
	}

	return totalSize
}

func (p *TextPrinter) PrintClones(dups [][]*syntax.Node, sortBy ...SortBy) error {
	p.cnt++

	sortCriteria := SortBySize
	if len(sortBy) > 0 {
		sortCriteria = sortBy[0]
	}

	sortedDups := SortNodesByCriteria(dups, sortCriteria)

	clones, err := prepareClonesInfo(p.ReadFile, sortedDups)
	if err != nil {
		return fmt.Errorf(
			"failed to prepare clones info for %d duplicates: %w",
			len(sortedDups),
			err,
		)
	}

	groupCloneSize := calculateCloneSizes(clones)

	isFileDupe := detectFileDuplicate(p.isFileDupe, sortedDups)

	if isFileDupe && p.currentHash != "" {
		hashPrefix := p.currentHash
		if len(hashPrefix) > 12 {
			hashPrefix = hashPrefix[:12]
		}

		fileSizeStr := formatBytes(clones[0].fileSize)
		if _, err := fmt.Fprintf(
			p.w,
			"📄 FILE DUPLICATE | 🔗 [%s...] | %d files | %s\n\n",
			hashPrefix,
			len(sortedDups),
			fileSizeStr,
		); err != nil {
			return err
		}

		for _, cl := range clones {
			p.diffHintFiles = append(p.diffHintFiles, cl.filename)
		}
	} else {
		if _, err := fmt.Fprintf(p.w, "found %d clones:\n", len(sortedDups)); err != nil {
			return err
		}
	}

	p.cloneGroups = append(p.cloneGroups, clones)
	p.totalSize += groupCloneSize

	sort.Sort(byNameAndLine(clones))

	return p.printCloneList(clones)
}

// PrintClonesSorted prints clones with specified sorting criteria.
func (p *TextPrinter) PrintClonesSorted(dups [][]*syntax.Node, sortBy SortBy) error {
	p.cnt++
	if _, err := fmt.Fprintf(
		p.w,
		"found %d clones (sorted by %s):\n",
		len(dups),
		sortBy,
	); err != nil {
		return err
	}

	sortedDups := SortNodesByCriteria(dups, sortBy)

	clones, err := prepareClonesInfo(p.ReadFile, sortedDups)
	if err != nil {
		return fmt.Errorf(
			"failed to prepare clones info for sorted output (%d duplicates): %w",
			len(dups),
			err,
		)
	}

	return p.printCloneList(clones)
}

func (p *TextPrinter) PrintFooter() error {
	if _, err := fmt.Fprintf(p.w, "\nFound total %d clone groups.\n", p.cnt); err != nil {
		return err
	}

	// Add diff hint if we have file duplicates
	if len(p.diffHintFiles) >= 2 {
		// Use first two files for diff hint
		file1 := p.diffHintFiles[0]

		file2 := p.diffHintFiles[1]
		if _, err := fmt.Fprintf(p.w, "\n→ diff %s %s\n", file1, file2); err != nil {
			return err
		}
	}

	return nil
}

func prepareClonesInfo(fread ReadFile, dups [][]*syntax.Node) ([]clone, error) {
	clones := make([]clone, len(dups))
	for i, dup := range dups {
		cnt := len(dup)
		if cnt == 0 {
			return nil, errors.NewInternalError("zero length duplicate found", nil)
		}

		nstart := dup[0]
		nend := dup[cnt-1]

		// Use unified file processor to get file info
		fileInfo, err := ProcessNodeRange(fread, nstart, nend)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to read file %s for clone info: %w",
				nstart.Filename,
				err,
			)
		}

		cl := clone{
			filename:  fileInfo.Filename,
			lineStart: fileInfo.LineStart,
			lineEnd:   fileInfo.LineEnd,
		}
		cl.fragment = extractContent(fileInfo, nstart, nend)
		cl.fileSize = len(fileInfo.Content)
		clones[i] = cl
	}

	return clones, nil
}

// OutputText generates text output with sorting.
func (p *TextPrinter) OutputText(threshold int, sortBy SortBy) error {
	// Sort all clone groups based on the specified criteria
	sortedCloneGroups := make([][]clone, len(p.cloneGroups))
	copy(sortedCloneGroups, p.cloneGroups)

	switch sortBy {
	case SortBySize:
		// Sort by total token size of each clone group
		sortCloneGroupsBySize(sortedCloneGroups)
	case SortByOccurrence:
		// Sort by number of files in each clone group
		sort.Slice(sortedCloneGroups, func(i, j int) bool {
			return len(sortedCloneGroups[i]) > len(sortedCloneGroups[j])
		})
	case SortByHash:
		// Sort by filename for deterministic output
		sortClonesByFilename(sortedCloneGroups)
	case SortByTotalTokens:
		// Sort by total tokens across all files in each clone group
		sortCloneGroupsBySize(sortedCloneGroups)
	default:
		// Default to size sorting
		sortCloneGroupsBySize(sortedCloneGroups)
	}

	// Print header
	err := p.PrintHeader()
	if err != nil {
		return err
	}

	// Print sorted clone groups
	for _, cloneGroup := range sortedCloneGroups {
		for _, cl := range cloneGroup {
			if len(cl.fragment) > 0 {
				if _, err := fmt.Fprintf(
					p.w,
					"%s\n%s:%d-%d\n\n",
					cl.fragment,
					cl.filename,
					cl.lineStart,
					cl.lineEnd,
				); err != nil {
					return err
				}
			} else {
				err := writeCloneLines(p.w, []clone{cl}, "%s:%d,%d")
				if err != nil {
					return err
				}
			}
		}
	}

	return p.PrintFooter()
}

func (p *TextPrinter) printCloneList(clones []clone) error {
	return writeCloneLines(p.w, clones, "  %s:%d,%d")
}

// formatBytes converts bytes to human-readable format (KB, MB, etc.).
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
