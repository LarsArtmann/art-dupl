package printer

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// printSizeDistribution prints the size distribution with ASCII bar visualization.
func printSizeDistribution(w io.Writer, distribution map[string]int) {
	ranges := make([]string, 0, len(distribution))
	for r := range distribution {
		ranges = append(ranges, r)
	}
	sort.Strings(ranges)

	// Find max count for scaling bars
	maxCount := 0
	total := 0
	for _, count := range distribution {
		if count > maxCount {
			maxCount = count
		}
		total += count
	}

	// Print distribution with bars
	for _, r := range ranges {
		count := distribution[r]
		percentage := 0.0
		if total > 0 {
			percentage = float64(count) / float64(total) * 100
		}

		// Create bar (max width 20 characters)
		barWidth := 0
		if maxCount > 0 {
			barWidth = int(float64(count) / float64(maxCount) * 20)
		}
		bar := strings.Repeat("█", barWidth)

		_, _ = fmt.Fprintf(w, "  %-15s: %4d clones [%s] %.1f%%\n", r, count, bar, percentage)
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
	limit := min(len(files), topN)

	for i := range limit {
		_, _ = fmt.Fprintf(w, "  %d lines in %s\n", files[i].lines, files[i].filename)
	}

	if len(files) > topN {
		_, _ = fmt.Fprintf(w, "  ... and %d more files\n", len(files)-topN)
	}
}
