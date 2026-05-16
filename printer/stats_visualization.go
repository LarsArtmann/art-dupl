package printer

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

func calcPercentage(count, total int) float64 {
	if total == 0 {
		return 0.0
	}

	return float64(count) / float64(total) * 100
}

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

		// Create bar (max width 20 characters)
		barWidth := 0
		if maxCount > 0 {
			barWidth = int(float64(count) / float64(maxCount) * 20)
		}

		bar := strings.Repeat("█", barWidth)

		pct := calcPercentage(count, total)

		_, _ = fmt.Fprintf(
			w,
			"  %-15s: %4d clones [%s] %.1f%%\n",
			r, count, bar, pct,
		)
	}
}

// printTopFiles prints the top N files with most duplicate lines.
func printTopFiles(w io.Writer, fileDuplication map[string]int, topN int) {
	// Convert to slice for sorting
	files := make([]FileStatMixin, 0, len(fileDuplication))
	for filename, lines := range fileDuplication {
		files = append(files, FileStatMixin{filename, lines})
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

// severityOrder defines the display order for severity levels.
var severityOrder = []string{healthSmall, healthMedium, healthLarge, healthHuge}

// printSeverityDistribution prints the severity breakdown with visualization.
func printSeverityDistribution(w io.Writer, distribution map[string]int) {
	total := 0
	for _, count := range distribution {
		total += count
	}

	for _, severity := range severityOrder {
		count, exists := distribution[severity]
		if !exists {
			continue
		}

		pct := calcPercentage(count, total)

		_, _ = fmt.Fprintf(
			w,
			"  %-10s: %4d clones (%.1f%%)\n",
			severity, count, pct,
		)
	}
}
