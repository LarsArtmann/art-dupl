// Package printer provides output formatting and statistics for duplicate detection.
//
// This package implements multiple output formats and statistical analysis
// of detected code duplicates.
//
// Output Formats:
// - TextPrinter: Human-readable text output
// - JSONPrinter: Structured JSON output
// - HTMLPrinter: HTML report with syntax highlighting
// - PlumbingPrinter: Machine-readable output for scripting
// - StatsPrinter: Comprehensive statistics and health analysis
//
// Design:
// - Printer interface: Common API for all formats
// - Format-specific implementations in separate files
// - Statistics split across focused files:
//   - stats.go: Core type and interface methods
//   - stats_collector.go: SetX methods for data collection
//   - stats_health.go: Health score calculation
//   - stats_formatter.go: Output formatters (CSV, JSON, Text)
//   - stats_visualization.go: Visualization helpers
//   - stats_recommendations.go: Recommendation generation
//   - stats_styles.go: Style management
//   - stats_data.go: Data structures
//
// Configuration:
// - SortBy criteria: Size, Occurrence, Hash, TotalTokens
// - OutputFormat selection: Text, HTML, JSON, Plumbing, Simple JSON
// - Threshold settings: Minimum size for duplicates
// - Verbosity: Detailed output for debugging
//
// Performance:
// - Streaming output for large projects
// - Efficient memory usage for statistics
// - Lazy evaluation where possible
package printer

import (
	"fmt"
	"io"

	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/charmbracelet/lipgloss"
)

// stats provides aggregated statistics about code duplication.
type stats struct {
	ReadFile

	w         io.Writer
	threshold int
	format    Format
	statsData *StatsData
	base      lipgloss.Style
	header    lipgloss.Style
	section   lipgloss.Style
	metric    lipgloss.Style
	success   lipgloss.Style
	warning   lipgloss.Style
	error     lipgloss.Style
}

// NewStats creates a new stats printer.
//
// Note: Accepts `threshold int` for backward compatibility.
// For type-safe version, use domain.Threshold at call site.
func NewStats(w io.Writer, fread ReadFile, threshold int) Printer {
	// Initialize styles
	styles := initStyles()

	return &stats{
		w:         w,
		ReadFile:  fread,
		threshold: threshold,
		statsData: &StatsData{
			FileDuplication:   make(map[string]int),
			SizeDistribution:  make(map[string]int),
			TokenDistribution: make(map[string]int),
		},
		base:    styles.base,
		header:  styles.header,
		section: styles.section,
		metric:  styles.metric,
		success: styles.success,
		warning: styles.warning,
		error:   styles.error,
	}
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

	// Track unique duplicate lines (pattern counted once, not per instance)
	uniqueLineCount := 0

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
		// Use unique line count from first clone (all clones in group have same pattern)
		if uniqueLineCount == 0 {
			uniqueLineCount = lineCount
		}
		p.statsData.TotalTokens += len(dup)

		// Track file-level duplication
		p.statsData.FileDuplication[nstart.Filename] += lineCount

		// Track size distribution
		sizeRange := p.getSizeRange(lineCount)
		p.statsData.SizeDistribution[sizeRange]++

		// Track token distribution
		tokenRange := p.getTokenRange(len(dup))
		p.statsData.TokenDistribution[tokenRange]++
	}

	// Add unique line count once per clone group (not once per clone instance)
	p.statsData.TotalDuplicateLines += uniqueLineCount

	// Update impact score (tokens × instances)
	p.statsData.ImpactScore += tokensInGroup * len(dups)

	return nil
}

// PrintFooter prints the aggregated statistics.
func (p *stats) PrintFooter() error {
	// Calculate average clone size (unique pattern size per group)
	if p.statsData.TotalCloneGroups > 0 {
		p.statsData.AverageCloneSize = p.statsData.TotalDuplicateLines / p.statsData.TotalCloneGroups
	}

	// Calculate complexity score (clones per group)
	if p.statsData.TotalCloneGroups > 0 {
		p.statsData.ComplexityScore = float64(p.statsData.TotalClones) / float64(p.statsData.TotalCloneGroups)
	}

	// Calculate duplication ratio (percentage)
	if p.statsData.TotalEstimatedLines > 0 {
		p.statsData.DuplicationRatio = float64(p.statsData.TotalDuplicateLines) / float64(p.statsData.TotalEstimatedLines) * 100
	}

	// Calculate health score based on duplication ratio
	p.statsData.HealthScore = p.calculateHealthScore()

	// Print statistics
	p.printStats()

	return nil
}
