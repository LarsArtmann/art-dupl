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

	cnt         int
	w           io.Writer
	totalSize   int
	cloneGroups [][]clone
	currentHash string // Hash for the current clone group (for hash detection)
	isFileDupe  bool   // True if current group is an entire file duplicate
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
	return &TextPrinter{w: w, ReadFile: fread, cloneGroups: make([][]clone, 0)} //nolint:exhaustruct
}

func (p *TextPrinter) PrintHeader() error { return nil }

func (p *TextPrinter) PrintClones(dups [][]*syntax.Node, sortBy ...SortBy) error {
	p.cnt++

	// Extract sortBy parameter, default to SortBySize
	sortCriteria := SortBySize
	if len(sortBy) > 0 {
		sortCriteria = sortBy[0]
	}

	// Apply sorting to the clone groups before processing
	sortedDups := SortNodesByCriteria(dups, sortCriteria)

	// Detect if this is a file duplicate (hash detection mode)
	// File duplicates have: single node per fragment, node.Pos == 0
	isFileDupe := p.isFileDupe
	if !isFileDupe && len(sortedDups) > 0 {
		// Check if all fragments are single nodes starting at position 0
		allSingleNodes := true
		for _, frag := range sortedDups {
			if len(frag) != 1 || frag[0].Pos != 0 {
				allSingleNodes = false
			}
		}

		if allSingleNodes {
			isFileDupe = true
		}
	}

	// Print enhanced header for file duplicates
	if isFileDupe && p.currentHash != "" {
		hashPrefix := p.currentHash
		if len(hashPrefix) > 12 {
			hashPrefix = hashPrefix[:12]
		}
		if _, err := fmt.Fprintf(
			p.w,
			"📄 FILE DUPLICATE | 🔗 [%s...] | %d files\n\n",
			hashPrefix,
			len(sortedDups),
		); err != nil {
			return err
		}
	} else {
		if _, err := fmt.Fprintf(p.w, "found %d clones:\n", len(sortedDups)); err != nil {
			return err
		}
	}

	clones, err := prepareClonesInfo(p.ReadFile, sortedDups)
	if err != nil {
		return fmt.Errorf(
			"failed to prepare clones info for %d duplicates: %w",
			len(sortedDups),
			err,
		)
	}

	// Store clones with size for sorting
	groupCloneSize := 0

	for _, cl := range clones {
		cl.size = len(cl.fragment) // Size is the length of the fragment
		groupCloneSize += cl.size
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
		return err //nolint:wrapcheck // fmt errors are clear in context
	}

	clones, err := prepareClonesInfo(p.ReadFile, dups)
	if err != nil {
		return fmt.Errorf(
			"failed to prepare clones info for sorted output (%d duplicates): %w",
			len(dups),
			err,
		)
	}

	// Apply sorting based on criteria
	_ = SortNodesByCriteria(dups, sortBy) // Sorting applied, result not needed for this output

	sort.Slice(clones, func(i, j int) bool {
		return len(dups[i]) > len(dups[j])
	})

	return p.printCloneList(clones)
}

func (p *TextPrinter) PrintFooter() error {
	_, err := fmt.Fprintf(p.w, "\nFound total %d clone groups.\n", p.cnt)

	return err //nolint:wrapcheck // fmt errors are clear in context
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

		cl := clone{ //nolint:exhaustruct // fragment and size set separately below
			filename:  fileInfo.Filename,
			lineStart: fileInfo.LineStart,
			lineEnd:   fileInfo.LineEnd,
		}
		cl.fragment = extractContent(fileInfo, nstart, nend)
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
					return err //nolint:wrapcheck // fmt errors are clear in context
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
