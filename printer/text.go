package printer

import (
	"fmt"
	"io"
	"sort"

	"github.com/LarsArtmann/art-dupl/errors"
	"github.com/LarsArtmann/art-dupl/syntax"
)

type text struct {
	cnt         int
	w           io.Writer
	totalSize   int
	cloneGroups [][]clone
	ReadFile
}

func NewText(w io.Writer, fread ReadFile) Printer {
	return &text{w: w, ReadFile: fread, cloneGroups: make([][]clone, 0)}
}

func (p *text) PrintHeader() error { return nil }

func (p *text) PrintClones(dups [][]*syntax.Node, sortBy ...string) error {
	p.cnt++

	// Extract sortBy parameter, default to "size"
	sortCriteria := "size"
	if len(sortBy) > 0 {
		sortCriteria = sortBy[0]
	}

	// Apply sorting to the clone groups before processing
	sortedDups := make([][]*syntax.Node, len(dups))
	copy(sortedDups, dups)

	switch sortCriteria {
	case "size":
		sortedDups = SortClonesBySize(sortedDups)
	case "occurrence":
		sortedDups = SortClonesByOccurrence(sortedDups)
	case "hash":
		sortedDups = SortClonesByHash(sortedDups)
	default:
		sortedDups = SortClonesBySize(sortedDups) // Default to size
	}

	if _, err := fmt.Fprintf(p.w, "found %d clones:\n", len(sortedDups)); err != nil {
		return err
	}
	clones, err := prepareClonesInfo(p.ReadFile, sortedDups)
	if err != nil {
		return err
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
			return err
		}
	}
	return nil
}

// PrintClonesSorted prints clones with specified sorting criteria
func (p *text) PrintClonesSorted(dups [][]*syntax.Node, sortBy string) error {
	p.cnt++
	if _, err := fmt.Fprintf(p.w, "found %d clones (sorted by %s):\n", len(dups), sortBy); err != nil {
		return err
	}
	clones, err := prepareClonesInfo(p.ReadFile, dups)
	if err != nil {
		return err
	}

	// Apply sorting based on criteria
	switch sortBy {
	case "size", "":
		sort.Slice(clones, func(i, j int) bool {
			return len(dups[i]) > len(dups[j]) // Sort by token count
		})
	case "occurrence":
		// For text output, occurrence is same as size (each dup in different file)
		sort.Slice(clones, func(i, j int) bool {
			return len(dups[i]) > len(dups[j])
		})
	case "hash":
		sort.Slice(clones, func(i, j int) bool {
			if len(dups[i]) == 0 && len(dups[j]) == 0 {
				return false
			}
			if len(dups[i]) == 0 {
				return true
			}
			if len(dups[j]) == 0 {
				return false
			}
			// Use Filename for deterministic sorting
			if dups[i][0].Filename == dups[j][0].Filename {
				return dups[i][0].Pos < dups[j][0].Pos
			}
			return dups[i][0].Filename < dups[j][0].Filename
		})
	default:
		// Default to size sorting
		sort.Slice(clones, func(i, j int) bool {
			return len(dups[i]) > len(dups[j])
		})
	}

	for _, cl := range clones {
		if _, err := fmt.Fprintf(p.w, "  %s:%d,%d\n", cl.filename, cl.lineStart, cl.lineEnd); err != nil {
			return err
		}
	}
	return nil
}

func (p *text) PrintFooter() error {
	_, err := fmt.Fprintf(p.w, "\nFound total %d clone groups.\n", p.cnt)
	return err
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
			return nil, err
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

// OutputText generates text output with sorting
func (p *text) OutputText(threshold int, sortBy string) error {
	// Sort all clone groups based on the specified criteria
	sortedCloneGroups := make([][]clone, len(p.cloneGroups))
	copy(sortedCloneGroups, p.cloneGroups)

	switch sortBy {
	case "size":
		// Sort by total token size of each clone group
		sort.Slice(sortedCloneGroups, func(i, j int) bool {
			// Calculate total size for group i
			sizeI := 0
			for _, cl := range sortedCloneGroups[i] {
				sizeI += cl.size
			}
			// Calculate total size for group j
			sizeJ := 0
			for _, cl := range sortedCloneGroups[j] {
				sizeJ += cl.size
			}
			return sizeI > sizeJ
		})
	case "occurrence":
		// Sort by number of files in each clone group
		sort.Slice(sortedCloneGroups, func(i, j int) bool {
			return len(sortedCloneGroups[i]) > len(sortedCloneGroups[j])
		})
	case "hash":
		// Sort by filename for deterministic output
		sort.Slice(sortedCloneGroups, func(i, j int) bool {
			if len(sortedCloneGroups[i]) == 0 && len(sortedCloneGroups[j]) == 0 {
				return false
			}
			if len(sortedCloneGroups[i]) == 0 {
				return true
			}
			if len(sortedCloneGroups[j]) == 0 {
				return false
			}
			// Compare by first filename in each group
			if sortedCloneGroups[i][0].filename == sortedCloneGroups[j][0].filename {
				return sortedCloneGroups[i][0].lineStart < sortedCloneGroups[j][0].lineStart
			}
			return sortedCloneGroups[i][0].filename < sortedCloneGroups[j][0].filename
		})
	default:
		// Default to size sorting
		sort.Slice(sortedCloneGroups, func(i, j int) bool {
			sizeI := 0
			for _, cl := range sortedCloneGroups[i] {
				sizeI += cl.size
			}
			sizeJ := 0
			for _, cl := range sortedCloneGroups[j] {
				sizeJ += cl.size
			}
			return sizeI > sizeJ
		})
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
					return err
				}
			} else {
				if _, err := fmt.Fprintf(p.w, "%s:%d,%d\n", cl.filename, cl.lineStart, cl.lineEnd); err != nil {
					return err
				}
			}
		}
	}

	return p.PrintFooter()
}
