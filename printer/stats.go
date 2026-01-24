package printer

import (
	"fmt"
	"io"
	"sort"

	"github.com/LarsArtmann/art-dupl/syntax"
)

// stats provides aggregated statistics about code duplication.
type stats struct {
	ReadFile
	w         io.Writer
	threshold int
	statsData *StatsData
}

// StatsData holds all aggregated statistics.
type StatsData struct {
	TotalFilesScanned    int
	TotalCloneGroups     int
	TotalClones          int
	TotalDuplicateLines  int
	TotalTokens          int
	AverageCloneSize     int
	ComplexityScore      float64
	ImpactScore         int
	FileDuplication      map[string]int // filename -> duplicate line count
	SizeDistribution     map[string]int // size range -> count
	DetectionMethods     string
}

// NewStats creates a new stats printer.
func NewStats(w io.Writer, fread ReadFile, threshold int) Printer {
	return &stats{
		w:         w,
		ReadFile:  fread,
		threshold: threshold,
		statsData: &StatsData{
			FileDuplication:  make(map[string]int),
			SizeDistribution: make(map[string]int),
		},
	}
}

// SetFilesCount sets the total number of files scanned.
func (p *stats) SetFilesCount(count int) {
	p.statsData.TotalFilesScanned = count
}

// PrintHeader prints the stats header.
func (p *stats) PrintHeader() error {
	return nil
}

// PrintClones collects statistics from the clone groups.
func (p *stats) PrintClones(dups [][]*syntax.Node, sortBy ...SortBy) error {
	// Count clone group
	p.statsData.TotalCloneGroups++

	// Count tokens in this clone group
	tokensInGroup := 0

	// Process each clone in the group
	for _, dup := range dups {
		if len(dup) == 0 {
			continue
		}

		// Get file and line information using shared processor
		nstart := dup[0]
		nend := dup[len(dup)-1]

		fileInfo, err := ProcessNodeRange(p.ReadFile, nstart, nend)
		if err != nil {
			return fmt.Errorf("failed to process node range for file %s: %w", nstart.Filename, err)
		}

		lineCount := fileInfo.LineEnd - fileInfo.LineStart + 1
		tokensInGroup += len(dup)

		// Update statistics
		p.statsData.TotalClones++
		p.statsData.TotalDuplicateLines += lineCount
		p.statsData.TotalTokens += len(dup)

		// Track file-level duplication
		p.statsData.FileDuplication[nstart.Filename] += lineCount

		// Track size distribution
		sizeRange := p.getSizeRange(lineCount)
		p.statsData.SizeDistribution[sizeRange]++
	}

	// Update impact score (tokens × instances)
	p.statsData.ImpactScore += tokensInGroup * len(dups)

	return nil
}

// PrintFooter prints the aggregated statistics.
func (p *stats) PrintFooter() error {
	// Calculate average clone size
	if p.statsData.TotalClones > 0 {
		p.statsData.AverageCloneSize = p.statsData.TotalDuplicateLines / p.statsData.TotalClones
	}

	// Calculate complexity score (clones per group)
	if p.statsData.TotalCloneGroups > 0 {
		p.statsData.ComplexityScore = float64(p.statsData.TotalClones) / float64(p.statsData.TotalCloneGroups)
	}

	// Print statistics
	p.printStats()

	return nil
}

// printStats prints the collected statistics.
func (p *stats) printStats() {
	fmt.Fprintf(p.w, "Code Duplication Statistics\n")
	fmt.Fprintf(p.w, "============================\n\n")

	fmt.Fprintf(p.w, "Configuration:\n")
	fmt.Fprintf(p.w, "  Threshold: %d tokens\n", p.threshold)
	fmt.Fprintf(p.w, "  Detection Methods: %s\n", p.statsData.DetectionMethods)
	fmt.Fprintf(p.w, "\n")

	fmt.Fprintf(p.w, "Overview:\n")
	fmt.Fprintf(p.w, "  Files Scanned: %d\n", p.statsData.TotalFilesScanned)
	fmt.Fprintf(p.w, "  Clone Groups: %d\n", p.statsData.TotalCloneGroups)
	fmt.Fprintf(p.w, "  Total Clones: %d\n", p.statsData.TotalClones)
	fmt.Fprintf(p.w, "\n")

	fmt.Fprintf(p.w, "Duplicate Code:\n")
	fmt.Fprintf(p.w, "  Total Duplicate Lines: %d\n", p.statsData.TotalDuplicateLines)
	fmt.Fprintf(p.w, "  Total Duplicate Tokens: %d\n", p.statsData.TotalTokens)
	fmt.Fprintf(p.w, "  Average Clone Size: %d lines\n", p.statsData.AverageCloneSize)
	fmt.Fprintf(p.w, "  Complexity Score: %.2f\n", p.statsData.ComplexityScore)
	fmt.Fprintf(p.w, "  Impact Score: %d\n", p.statsData.ImpactScore)
	fmt.Fprintf(p.w, "\n")

	// Print size distribution
	if len(p.statsData.SizeDistribution) > 0 {
		fmt.Fprintf(p.w, "Clone Size Distribution:\n")
		printSizeDistribution(p.w, p.statsData.SizeDistribution)
		fmt.Fprintf(p.w, "\n")
	}

	// Print top files with most duplicates
	if len(p.statsData.FileDuplication) > 0 {
		fmt.Fprintf(p.w, "Top Files by Duplicate Lines:\n")
		printTopFiles(p.w, p.statsData.FileDuplication, 10)
	}
}

// getSizeRange returns a human-readable size range for a line count.
func (p *stats) getSizeRange(lines int) string {
	switch {
	case lines <= 5:
		return "1-5 lines"
	case lines <= 10:
		return "6-10 lines"
	case lines <= 20:
		return "11-20 lines"
	case lines <= 50:
		return "21-50 lines"
	case lines <= 100:
		return "51-100 lines"
	default:
		return "100+ lines"
	}
}

// printSizeDistribution prints the size distribution.
func printSizeDistribution(w io.Writer, distribution map[string]int) {
	ranges := make([]string, 0, len(distribution))
	for r := range distribution {
		ranges = append(ranges, r)
	}
	sort.Strings(ranges)

	for _, r := range ranges {
		fmt.Fprintf(w, "  %s: %d clones\n", r, distribution[r])
	}
}

// printTopFiles prints the top N files with most duplicate lines.
func printTopFiles(w io.Writer, fileDuplication map[string]int, topN int) {
	// Convert to slice for sorting
	type fileStat struct {
		filename string
		lines    int
	}
	files := make([]fileStat, 0, len(fileDuplication))
	for filename, lines := range fileDuplication {
		files = append(files, fileStat{filename, lines})
	}

	// Sort by duplicate lines (descending)
	sort.Slice(files, func(i, j int) bool {
		return files[i].lines > files[j].lines
	})

	// Print top N
	limit := len(files)
	if limit > topN {
		limit = topN
	}

	for i := 0; i < limit; i++ {
		fmt.Fprintf(w, "  %d lines in %s\n", files[i].lines, files[i].filename)
	}

	if len(files) > topN {
		fmt.Fprintf(w, "  ... and %d more files\n", len(files)-topN)
	}
}

// GetStatsData returns the collected statistics data.
func (p *stats) GetStatsData() *StatsData {
	return p.statsData
}

// SetDetectionMethods sets the detection methods used.
func (p *stats) SetDetectionMethods(methods string) {
	p.statsData.DetectionMethods = methods
}
