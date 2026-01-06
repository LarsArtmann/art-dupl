package printer

import (
	"fmt"
	"io"
	"sort"

	"github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/syntax"
)

type text struct {
	ReadFile

	cnt         int
	w           io.Writer
	totalSize   int
	cloneGroups [][]clone
}

//nolint:ireturn // Printer interface is appropriate return type for factory function
func NewText(w io.Writer, fread ReadFile) Printer {
	return &text{w: w, ReadFile: fread, cloneGroups: make([][]clone, 0)}
}

func (p *text) PrintHeader() error { return nil }

func (p *text) PrintClones(dups [][]*syntax.Node, sortBy ...string) error {
	p.cnt++

	// Extract sortBy parameter, default to sortBySize
	sortCriteria := sortBySize
	if len(sortBy) > 0 {
		sortCriteria = sortBy[0]
	}

	// Apply sorting to the clone groups before processing
	sortedDups := SortNodesByCriteria(dups, sortCriteria)

	if _, err := fmt.Fprintf(p.w, "found %d clones:\n", len(sortedDups)); err != nil {
		return err //nolint:wrapcheck // fmt errors are clear in context
	}
	clones, err := prepareClonesInfo(p.ReadFile, sortedDups)
	if err != nil {
		return fmt.Errorf("failed to prepare clones info for %d duplicates: %w", len(sortedDups), err)
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
	for _, cl := range clones {
		if _, err := fmt.Fprintf(p.w, "  %s:%d,%d\n", cl.filename, cl.lineStart, cl.lineEnd); err != nil {
			return err //nolint:wrapcheck // fmt errors are clear in context
		}
	}
	return nil
}

// PrintClonesSorted prints clones with specified sorting criteria.
func (p *text) PrintClonesSorted(dups [][]*syntax.Node, sortBy string) error {
	p.cnt++
	if _, err := fmt.Fprintf(p.w, "found %d clones (sorted by %s):\n", len(dups), sortBy); err != nil {
		return err //nolint:wrapcheck // fmt errors are clear in context
	}
	clones, err := prepareClonesInfo(p.ReadFile, dups)
	if err != nil {
		return fmt.Errorf("failed to prepare clones info for sorted output (%d duplicates): %w", len(dups), err)
	}

	// Apply sorting based on criteria
	_ = SortNodesByCriteria(dups, sortBy) // Sorting applied, result not needed for this output
	sort.Slice(clones, func(i, j int) bool {
		return len(dups[i]) > len(dups[j])
	})

	for _, cl := range clones {
		if _, err := fmt.Fprintf(p.w, "  %s:%d,%d\n", cl.filename, cl.lineStart, cl.lineEnd); err != nil {
			return err //nolint:wrapcheck // fmt errors are clear in context
		}
	}
	return nil
}

func (p *text) PrintFooter() error {
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

		file, err := fread(nstart.Filename)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s for clone info: %w", nstart.Filename, err)
		}

		cl := clone{filename: nstart.Filename}
		cl.lineStart, cl.lineEnd = blockLines(file, nstart.Pos, nend.End)
		clones[i] = cl
	}
	return clones, nil
}

func blockLines(file []byte, from, to int) (int, int) {
	line := 1
	lineStart, lineEnd := 0, 0
	for offset, b := range file {
		if b == '\n' {
			line++
		}
		if offset == from {
			lineStart = line
		}
		if offset == to-1 {
			lineEnd = line
			break
		}
	}
	return lineStart, lineEnd
}

// OutputText generates text output with sorting.
func (p *text) OutputText(threshold int, sortBy string) error { //nolint:cyclop // Text output with multiple sorting strategies
	// Sort all clone groups based on the specified criteria
	sortedCloneGroups := make([][]clone, len(p.cloneGroups))
	copy(sortedCloneGroups, p.cloneGroups)

	switch sortBy {
	case sortBySize:
		// Sort by total token size of each clone group
		sortCloneGroupsBySize(sortedCloneGroups)
	case "occurrence":
		// Sort by number of files in each clone group
		sort.Slice(sortedCloneGroups, func(i, j int) bool {
			return len(sortedCloneGroups[i]) > len(sortedCloneGroups[j])
		})
	case "hash":
		// Sort by filename for deterministic output
		sortClonesByFilename(sortedCloneGroups)
	case "total-tokens":
		// Sort by total tokens across all files in each clone group
		sortCloneGroupsBySize(sortedCloneGroups)
	default:
		// Default to size sorting
		sortCloneGroupsBySize(sortedCloneGroups)
	}

	// Print header
	if err := p.PrintHeader(); err != nil {
		return err
	}

	// Print sorted clone groups
	for _, cloneGroup := range sortedCloneGroups {
		for _, cl := range cloneGroup {
			if len(cl.fragment) > 0 {
				if _, err := fmt.Fprintf(p.w, "%s\n%s:%d-%d\n\n", cl.fragment, cl.filename, cl.lineStart, cl.lineEnd); err != nil {
					return err //nolint:wrapcheck // fmt errors are clear in context
				}
			} else {
				if _, err := fmt.Fprintf(p.w, "%s:%d,%d\n", cl.filename, cl.lineStart, cl.lineEnd); err != nil {
					return err //nolint:wrapcheck // fmt errors are clear in context
				}
			}
		}
	}

	return p.PrintFooter()
}
