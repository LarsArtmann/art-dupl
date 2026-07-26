package stats

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

// rangeStart extracts the starting number from a size range string.
// Handles formats like "1-5 lines", "6-10 lines", "100+ lines".
func rangeStart(r string) int {
	var n int

	_, _ = fmt.Sscanf(r, "%d", &n)

	return n
}

// printSizeDistribution prints the size distribution with ASCII bar visualization.
func printSizeDistribution(w io.Writer, distribution map[string]int) {
	ranges := make([]string, 0, len(distribution))
	//art-dupl:accept idiomatic range loop; false-positive match with test code (different domain)
	for r := range distribution {
		ranges = append(ranges, r)
	}

	sort.Slice(ranges, func(i, j int) bool {
		return rangeStart(ranges[i]) < rangeStart(ranges[j])
	})

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

// categoryOrder defines a stable display order for clone categories.
var categoryOrder = []string{ //nolint:gochecknoglobals // stable display order
	"function", "method", "handler", "test",
	"struct", "interface",
	"loop", "conditional",
	"assignment", "expression",
	"unknown",
}

// printCategoryDistribution prints the category breakdown with percentages.
func printCategoryDistribution(w io.Writer, distribution map[string]int) {
	printOrderedDistribution(w, distribution, categoryOrder, "  %-12s: %4d groups (%.1f%%)\n")
}

// priorityOrder defines a stable display order for priority levels.
var priorityOrder = []string{"critical", "high", "medium", "low"} //nolint:gochecknoglobals // stable display order

// printPriorityDistribution prints the priority breakdown with percentages.
func printPriorityDistribution(w io.Writer, distribution map[string]int) {
	printOrderedDistribution(w, distribution, priorityOrder, "  %-10s: %4d groups (%.1f%%)\n")
}

// severityOrder defines the display order for severity levels.
var severityOrder = []string{ //nolint:gochecknoglobals // stable display order
	healthSmall,
	healthMedium,
	healthLarge,
	healthHuge,
}

// printSeverityDistribution prints the severity breakdown with visualization.
func printSeverityDistribution(w io.Writer, distribution map[string]int) {
	printOrderedDistribution(w, distribution, severityOrder, "  %-10s: %4d clones (%.1f%%)\n")
}

// printOrderedDistribution prints a breakdown of distribution in the given order
// using the provided format string. Categories missing from the distribution are
// skipped. The format receives (label, count, percentage).
func printOrderedDistribution(w io.Writer, distribution map[string]int, order []string, format string) {
	total := 0
	for _, count := range distribution {
		total += count
	}

	for _, label := range order {
		count, exists := distribution[label]
		if !exists {
			continue
		}

		pct := calcPercentage(count, total)

		_, _ = fmt.Fprintf(w, format, label, count, pct)
	}
}
