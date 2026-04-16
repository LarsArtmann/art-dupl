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

	"charm.land/lipgloss/v2"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// stats provides aggregated statistics about code duplication.
type stats struct {
	ReadFile
	StyleMixin

	w         io.Writer
	threshold int
	format    Format
	statsData *StatsData
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
			SeverityBreakdown: make(map[string]int),
		},
		StyleMixin: StyleMixin{
			base:    styles.base,
			header:  styles.header,
			section: styles.section,
			metric:  styles.metric,
			success: styles.success,
			warning: styles.warning,
			error:   styles.error,
		},
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

	// Track severity based on tokens in the clone group (once per group)
	severity := p.getSeverity(tokensInGroup)
	p.statsData.SeverityBreakdown[severity]++

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
		p.statsData.ComplexityScore = float64(
			p.statsData.TotalClones,
		) / float64(
			p.statsData.TotalCloneGroups,
		)
	}

	// Calculate duplication ratio (percentage)
	if p.statsData.TotalEstimatedLines > 0 {
		p.statsData.DuplicationRatio = float64(
			p.statsData.TotalDuplicateLines,
		) / float64(
			p.statsData.TotalEstimatedLines,
		) * 100
	}

	// Calculate health score based on duplication ratio
	p.statsData.HealthScore = p.calculateHealthScore()

	// Print statistics
	p.printStats()

	return nil
}

// printStyledLine prints a styled line with indentation and optional bullet.
func (p *stats) printStyledLinef(prefix, format string, args ...any) {
	_, _ = fmt.Fprintf(p.w, "%s %s\n", prefix, p.base.Render(fmt.Sprintf(format, args...)))
}

// printLine prints a base-styled line with indentation.
func (p *stats) printLinef(format string, args ...any) {
	p.printStyledLinef("  ", format, args...)
}

// printSection prints a section header.
func (p *stats) printSection(title string) {
	p.printWithStyle(p.section, title)
}

// printWithStyle prints text with the specified style.
func (p *stats) printWithStyle(style lipgloss.Style, text string) {
	_, _ = fmt.Fprintf(p.w, "%s\n", style.Render(text))
}

// printMetric prints a metric label and value.
// Note: We use Inherit() to combine styles and render the complete line at once.
// This avoids lipgloss v2 adding reset codes between styled segments, which would
// break substring matching in tests.
func (p *stats) printMetric(label, value string) {
	// Combine styles: base (bold) inherits from metric (color)
	styled := p.base.Inherit(p.metric)

	// Render the complete line with both styles applied
	_, _ = fmt.Fprintf(p.w, "%s\n", styled.Render(fmt.Sprintf("  %s: %s", label, value)))
}

// printSuccess prints a success-styled message.
func (p *stats) printSuccessf(format string, args ...any) {
	_, _ = fmt.Fprintf(p.w, "%s\n", p.success.Render(fmt.Sprintf(format, args...)))
}

// printWarning prints a warning-styled message.
func (p *stats) printWarningf(format string, args ...any) {
	_, _ = fmt.Fprintf(p.w, "%s\n", p.warning.Render(fmt.Sprintf(format, args...)))
}

// printError prints an error-styled message.
func (p *stats) printErrorf(format string, args ...any) {
	_, _ = fmt.Fprintf(p.w, "%s\n", p.error.Render(fmt.Sprintf(format, args...)))
}

// printHeader prints a header-styled title.
func (p *stats) printHeader(title string) {
	p.printWithStyle(p.header, title)
}

// printBullet prints a bullet point with base styling.
func (p *stats) printBulletf(format string, args ...any) {
	p.printStyledLinef("  •", format, args...)
}
