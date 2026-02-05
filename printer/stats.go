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
// - Statistics aggregation in stats.go/stats_data.go
//
// Key Components:
// - Formatting: Syntax highlighting, indentation, alignment
// - Statistics: Duplicate ratio, health score, complexity metrics
// - Sorting: Multiple sort criteria (size, occurrence, hash, tokens)
// - Filtering: Include/exclude patterns, vendor filtering
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
//
// StatsData fields:
// - Count metrics: TotalFilesScanned, TotalCloneGroups, TotalClones
// - Size metrics: TotalTokens, TotalDuplicateLines, AverageCloneSize
// - Complexity metrics: ComplexityScore, ImpactScore
// - Quality metrics: DuplicationRatio, HealthScore
// - Time metrics: AnalysisDuration, Timestamp
// - Aggregation metrics: FileDuplication, SizeDistribution
// - Metadata: DetectionMethods
//
// Domain Types Status:
// ✅ Added domain package import
// ✅ threshold uses int for backward compatibility
// ✅ StatsData could use domain types in future
//
// For type-safe threshold access:
//
//	cfg := config.DefaultConfig()
//	domainThreshold := cfg.GetThresholdAsDomain()
package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/charmbracelet/lipgloss"
)

// stats provides aggregated statistics about code duplication.
//
// TODO: TYPE SAFETY ISSUE - Uses primitive types instead of domain types:
// - threshold uses int instead of domain.Threshold
// - StatsData fields use int/float64 instead of domain types (TokenCount, ComplexityScore, etc.)
//
// Also: This file is 639 lines - getting large. Consider splitting:
// - stats_data.go for StatsData type and methods
// - stats_format.go for formatting logic
// - stats_calc.go for calculation logic
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
			FileDuplication:  make(map[string]int),
			SizeDistribution: make(map[string]int),
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

// styleConfig holds all lipgloss styles.
type styleConfig struct {
	base    lipgloss.Style
	header  lipgloss.Style
	section lipgloss.Style
	metric  lipgloss.Style
	success lipgloss.Style
	warning lipgloss.Style
	error   lipgloss.Style
}

// initStyles initializes lipgloss styles, respecting NO_COLOR environment variable.
func initStyles() styleConfig {
	// Check for NO_COLOR environment variable
	noColor := os.Getenv("NO_COLOR") != ""

	// Create styles
	headerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500")).Bold(true)
	sectionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00E676")).Bold(true)
	metricStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#738ADB"))
	successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00C853"))
	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#E53E3"))

	// Disable all colors if NO_COLOR is set
	if noColor {
		return styleConfig{
			base:    lipgloss.NewStyle(),
			header:  lipgloss.NewStyle(),
			section: lipgloss.NewStyle(),
			metric:  lipgloss.NewStyle(),
			success: lipgloss.NewStyle(),
			warning: lipgloss.NewStyle(),
			error:   lipgloss.NewStyle(),
		}
	}

	return styleConfig{
		base:    lipgloss.NewStyle().Bold(true),
		header:  headerStyle,
		section: sectionStyle,
		metric:  metricStyle,
		success: successStyle,
		warning: warningStyle,
		error:   errorStyle,
	}
}

// SetFilesCount sets the total number of files scanned.
func (p *stats) SetFilesCount(count int) {
	p.statsData.TotalFilesScanned = count
}

// SetAnalysisDuration sets the analysis duration.
func (p *stats) SetAnalysisDuration(duration time.Duration) {
	p.statsData.AnalysisDuration = duration.String()
}

// SetTotalEstimatedLines sets the estimated total lines of code.
func (p *stats) SetTotalEstimatedLines(lines int) {
	p.statsData.TotalEstimatedLines = lines
}

// SetTimestamp sets the analysis timestamp.
func (p *stats) SetTimestamp(timestamp string) {
	p.statsData.Timestamp = timestamp
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

// calculateHealthScore calculates an A-F grade based on duplication, complexity, and impact metrics.
func (p *stats) calculateHealthScore() string {
	if p.statsData.DuplicationRatio == 0 && p.statsData.ComplexityScore == 0 && p.statsData.ImpactScore == 0 {
		return "A"
	}

	// Normalize metrics to 0-100 scale (higher is worse)
	var duplicationScore float64
	if p.statsData.TotalEstimatedLines > 0 {
		duplicationScore = p.statsData.DuplicationRatio
	}

	// Complexity score: higher = worse. Normalize to 0-100 assuming max complexity of 10.0
	complexityScore := (p.statsData.ComplexityScore / 10.0) * 100
	if complexityScore > 100 {
		complexityScore = 100
	}

	// Impact score: higher = worse. Normalize to 0-100 assuming max impact of 10000
	impactScore := (float64(p.statsData.ImpactScore) / 10000.0) * 100
	if impactScore > 100 {
		impactScore = 100
	}

	// Weighted average: duplication 60%, complexity 25%, impact 15%
	totalScore := duplicationScore*0.6 + complexityScore*0.25 + impactScore*0.15

	// Convert to A-F grade based on total score
	switch {
	case totalScore < 3:
		return "A"
	case totalScore < 6:
		return "B"
	case totalScore < 10:
		return "C"
	case totalScore < 15:
		return "D"
	default:
		return "F"
	}
}

// printStats prints the collected statistics.
func (p *stats) printStats() {
	switch p.format {
	case FormatJSON:
		p.printJSON()
	case FormatCSV:
		p.printCSV()
	case FormatText:
		p.printText()
	}
}

// printCSV prints statistics in CSV format.
func (p *stats) printCSV() {
	// Write CSV header
	fmt.Fprintf(p.w, "Metric,Value\n")

	// Configuration
	fmt.Fprintf(p.w, "Threshold,%d\n", p.threshold)
	fmt.Fprintf(p.w, "Detection Methods,%s\n", p.statsData.DetectionMethods)
	fmt.Fprintf(p.w, "Timestamp,%s\n", p.statsData.Timestamp)
	fmt.Fprintf(p.w, "Analysis Time,%s\n", p.statsData.AnalysisDuration)
	fmt.Fprintf(p.w, "\n")

	// Overview
	fmt.Fprintf(p.w, "Files Scanned,%d\n", p.statsData.TotalFilesScanned)
	fmt.Fprintf(p.w, "Clone Groups,%d\n", p.statsData.TotalCloneGroups)
	fmt.Fprintf(p.w, "Total Clones,%d\n", p.statsData.TotalClones)
	fmt.Fprintf(p.w, "\n")

	// Duplicate Code
	fmt.Fprintf(p.w, "Total Duplicate Lines,%d\n", p.statsData.TotalDuplicateLines)
	fmt.Fprintf(p.w, "Estimated Total Lines,%d\n", p.statsData.TotalEstimatedLines)
	fmt.Fprintf(p.w, "Duplication Ratio,%.1f%%\n", p.statsData.DuplicationRatio)
	fmt.Fprintf(p.w, "Total Duplicate Tokens,%d\n", p.statsData.TotalTokens)
	fmt.Fprintf(p.w, "Average Clone Size,%d\n", p.statsData.AverageCloneSize)
	fmt.Fprintf(p.w, "Complexity Score,%.2f\n", p.statsData.ComplexityScore)
	fmt.Fprintf(p.w, "Impact Score,%d\n", p.statsData.ImpactScore)
	fmt.Fprintf(p.w, "Health Score,%s\n", p.statsData.HealthScore)
}

// printText prints statistics in text format.
func (p *stats) printText() {
	// Print header
	fmt.Fprintf(p.w, "%s\n", p.header.Render("Code Duplication Statistics"))
	fmt.Fprintf(p.w, "%s\n\n", p.base.Render("============================"))

	// Print configuration
	fmt.Fprintf(p.w, "%s\n", p.section.Render("Configuration:"))
	fmt.Fprintf(p.w, "  %s %s\n", p.metric.Render("Threshold:"), p.base.Render(fmt.Sprintf("%d tokens", p.threshold)))
	fmt.Fprintf(p.w, "  %s %s\n", p.metric.Render("Detection Methods:"), p.base.Render(p.statsData.DetectionMethods))
	if p.statsData.Timestamp != "" {
		fmt.Fprintf(p.w, "  %s %s\n", p.metric.Render("Timestamp:"), p.base.Render(p.statsData.Timestamp))
	}
	if p.statsData.AnalysisDuration != "" {
		fmt.Fprintf(p.w, "  %s %s\n", p.metric.Render("Analysis Time:"), p.base.Render(p.statsData.AnalysisDuration))
	}
	fmt.Fprintf(p.w, "\n")

	// Print overview
	fmt.Fprintf(p.w, "%s\n", p.section.Render("Overview:"))
	fmt.Fprintf(p.w, "  %s %s\n", p.metric.Render("Files Scanned:"), p.base.Render(fmt.Sprintf("%d", p.statsData.TotalFilesScanned)))
	fmt.Fprintf(p.w, "  %s %s\n", p.metric.Render("Clone Groups:"), p.base.Render(fmt.Sprintf("%d", p.statsData.TotalCloneGroups)))
	fmt.Fprintf(p.w, "  %s %s\n", p.metric.Render("Total Clones:"), p.base.Render(fmt.Sprintf("%d", p.statsData.TotalClones)))
	fmt.Fprintf(p.w, "\n")

	// Print duplicate code metrics
	fmt.Fprintf(p.w, "%s\n", p.section.Render("Duplicate Code:"))
	fmt.Fprintf(p.w, "  %s %s\n", p.metric.Render("Total Duplicate Lines:"), p.base.Render(fmt.Sprintf("%d", p.statsData.TotalDuplicateLines)))
	if p.statsData.TotalEstimatedLines > 0 {
		fmt.Fprintf(p.w, "  %s %s\n", p.metric.Render("Estimated Total Lines:"), p.base.Render(fmt.Sprintf("%d", p.statsData.TotalEstimatedLines)))
		fmt.Fprintf(p.w, "  %s %s\n", p.metric.Render("Duplication Ratio:"), p.healthScoreStyle(p.statsData.HealthScore).Render(fmt.Sprintf("%.1f%%", p.statsData.DuplicationRatio)))
	}
	fmt.Fprintf(p.w, "  %s %s\n", p.metric.Render("Total Duplicate Tokens:"), p.base.Render(fmt.Sprintf("%d", p.statsData.TotalTokens)))
	fmt.Fprintf(p.w, "  %s %s\n", p.metric.Render("Average Clone Size:"), p.base.Render(fmt.Sprintf("%d lines", p.statsData.AverageCloneSize)))
	fmt.Fprintf(p.w, "  %s %s\n", p.metric.Render("Complexity Score:"), p.base.Render(fmt.Sprintf("%.2f", p.statsData.ComplexityScore)))
	fmt.Fprintf(p.w, "  %s %s\n", p.metric.Render("Impact Score:"), p.base.Render(fmt.Sprintf("%d", p.statsData.ImpactScore)))
	if p.statsData.HealthScore != "" {
		fmt.Fprintf(p.w, "  %s %s\n", p.metric.Render("Health Score:"), p.healthScoreStyle(p.statsData.HealthScore).Render(p.statsData.HealthScore))
	}
	fmt.Fprintf(p.w, "\n")

	// Print size distribution
	if len(p.statsData.SizeDistribution) > 0 {
		fmt.Fprintf(p.w, "%s\n", p.section.Render("Clone Size Distribution:"))
		printSizeDistribution(p.w, p.statsData.SizeDistribution)
		fmt.Fprintf(p.w, "\n")
	}

	// Print top files with most duplicates
	if len(p.statsData.FileDuplication) > 0 {
		fmt.Fprintf(p.w, "%s\n", p.section.Render("Top Files by Duplicate Lines:"))
		printTopFiles(p.w, p.statsData.FileDuplication, 10)
	}

	// Print actionable recommendations
	fmt.Fprintf(p.w, "\n%s\n", p.header.Render("Recommendations:"))
	p.printRecommendations()
}

// printRecommendations prints actionable recommendations based on the health score.
func (p *stats) printRecommendations() {
	switch p.statsData.HealthScore {
	case "A":
		fmt.Fprintf(p.w, "%s\n", p.success.Render("✓ Excellent code health! Duplication is minimal."))
		fmt.Fprintf(p.w, "  %s\n", p.base.Render("Keep up the good work. Maintain current practices."))
	case "B":
		fmt.Fprintf(p.w, "%s\n", p.success.Render("✓ Good code health with minor duplication."))
		fmt.Fprintf(p.w, "  %s\n", p.base.Render("Consider extracting small duplicate patterns into shared functions."))
		fmt.Fprintf(p.w, "  %s\n", p.base.Render("Review the 'Top Files' section to identify problem areas."))
	case "C":
		fmt.Fprintf(p.w, "%s\n", p.warning.Render("! Moderate code duplication detected."))
		fmt.Fprintf(p.w, "  %s\n", p.base.Render("Prioritize refactoring duplicate code blocks:"))
		fmt.Fprintf(p.w, "  %s\n", p.base.Render("  1. Focus on large clones (50+ lines) first"))
		fmt.Fprintf(p.w, "  %s\n", p.base.Render("  2. Create shared utility functions or base classes"))
		fmt.Fprintf(p.w, "  %s\n", p.base.Render("  3. Consider domain-driven design patterns"))
	case "D":
		fmt.Fprintf(p.w, "%s\n", p.warning.Render("⚠ High code duplication - action needed."))
		fmt.Fprintf(p.w, "  %s\n", p.base.Render("Immediate actions recommended:"))
		fmt.Fprintf(p.w, "  %s\n", p.base.Render("  1. Extract all medium/large duplicate blocks (>20 lines)"))
		fmt.Fprintf(p.w, "  %s\n", p.base.Render("  2. Implement shared libraries or services"))
		fmt.Fprintf(p.w, "  %s\n", p.base.Render("  3. Establish code review guidelines to prevent new duplication"))
		fmt.Fprintf(p.w, "  %s\n", p.base.Render("  4. Consider architectural changes (e.g., introduce new abstractions)"))
	case "F":
		fmt.Fprintf(p.w, "%s\n", p.error.Render("✗ Critical code duplication - immediate action required!"))
		fmt.Fprintf(p.w, "  %s\n", p.base.Render("Urgent steps to take:"))
		fmt.Fprintf(p.w, "  %s\n", p.base.Render("  1. Prioritize ALL duplicate code extraction immediately"))
		fmt.Fprintf(p.w, "  %s\n", p.base.Render("  2. Halt new feature development until duplication is reduced"))
		fmt.Fprintf(p.w, "  %s\n", p.base.Render("  3. Create comprehensive refactoring plan"))
		fmt.Fprintf(p.w, "  %s\n", p.base.Render("  4. Invest in architectural review and design patterns"))
		fmt.Fprintf(p.w, "  %s\n", p.base.Render("  5. Consider team training on DRY principles"))
	default:
		fmt.Fprintf(p.w, "%s\n", p.base.Render("No recommendations available."))
	}

	// Additional recommendations based on specific metrics
	fmt.Fprintf(p.w, "\n%s\n", p.section.Render("Next Steps:"))

	if p.statsData.TotalCloneGroups > 10 {
		fmt.Fprintf(p.w, "  • %s\n", p.base.Render(fmt.Sprintf("You have %d clone groups - focus on the largest ones first", p.statsData.TotalCloneGroups)))
	}

	if p.statsData.AverageCloneSize > 50 {
		fmt.Fprintf(p.w, "  • %s\n", p.base.Render(fmt.Sprintf("Average clone size is %d lines - prioritize extracting large blocks", p.statsData.AverageCloneSize)))
	}

	if p.statsData.ComplexityScore > 3.0 {
		fmt.Fprintf(p.w, "  • %s\n", p.base.Render(fmt.Sprintf("Complexity score of %.2f suggests multiple clones per group - consider patterns", p.statsData.ComplexityScore)))
	}

	fmt.Fprintf(p.w, "  • %s\n", p.base.Render("Run with --threshold 50 to focus on large duplications only"))
	fmt.Fprintf(p.w, "  • %s\n", p.base.Render("Use --format json for machine-readable output"))
	fmt.Fprintf(p.w, "  • %s\n", p.base.Render("Integrate into CI/CD pipeline for continuous monitoring"))
}

// healthScoreStyle returns the appropriate style for a health score grade.
func (p *stats) healthScoreStyle(grade string) lipgloss.Style {
	switch grade {
	case "A":
		return p.success
	case "B":
		return p.success
	case "C":
		return p.warning
	case "D":
		return p.warning
	case "F":
		return p.error
	default:
		return p.base
	}
}

// printJSON prints statistics in JSON format.
func (p *stats) printJSON() {
	// Create a struct for JSON output
	jsonData := struct {
		Configuration struct {
			Threshold        int    `json:"threshold"`
			DetectionMethods string `json:"detectionMethods"`
		} `json:"configuration"`
		Overview struct {
			FilesScanned int `json:"filesScanned"`
			CloneGroups  int `json:"cloneGroups"`
			TotalClones  int `json:"totalClones"`
		} `json:"overview"`
		DuplicateCode struct {
			TotalLines       int     `json:"totalDuplicateLines"`
			EstimatedLines   int     `json:"estimatedTotalLines,omitempty"`
			TotalTokens      int     `json:"totalDuplicateTokens"`
			AverageCloneSize int     `json:"averageCloneSize"`
			ComplexityScore  float64 `json:"complexityScore"`
			ImpactScore      int     `json:"impactScore"`
			DuplicationRatio float64 `json:"duplicationRatio,omitempty"`
		} `json:"duplicateCode"`
		Metrics struct {
			HealthScore  string `json:"healthScore,omitempty"`
			AnalysisTime string `json:"analysisTime,omitempty"`
			Timestamp    string `json:"timestamp,omitempty"`
		} `json:"metrics"`
		SizeDistribution map[string]int `json:"sizeDistribution"`
		TopFiles         []struct {
			Filename string `json:"filename"`
			Lines    int    `json:"duplicateLines"`
		} `json:"topFiles"`
	}{}

	// Fill configuration
	jsonData.Configuration.Threshold = p.threshold
	jsonData.Configuration.DetectionMethods = p.statsData.DetectionMethods

	// Fill overview
	jsonData.Overview.FilesScanned = p.statsData.TotalFilesScanned
	jsonData.Overview.CloneGroups = p.statsData.TotalCloneGroups
	jsonData.Overview.TotalClones = p.statsData.TotalClones

	// Fill duplicate code metrics
	jsonData.DuplicateCode.TotalLines = p.statsData.TotalDuplicateLines
	jsonData.DuplicateCode.TotalTokens = p.statsData.TotalTokens
	jsonData.DuplicateCode.AverageCloneSize = p.statsData.AverageCloneSize
	jsonData.DuplicateCode.ComplexityScore = p.statsData.ComplexityScore
	jsonData.DuplicateCode.ImpactScore = p.statsData.ImpactScore
	if p.statsData.TotalEstimatedLines > 0 {
		jsonData.DuplicateCode.EstimatedLines = p.statsData.TotalEstimatedLines
		jsonData.DuplicateCode.DuplicationRatio = p.statsData.DuplicationRatio
	}

	// Fill metrics
	if p.statsData.HealthScore != "" {
		jsonData.Metrics.HealthScore = p.statsData.HealthScore
	}
	if p.statsData.AnalysisDuration != "" {
		jsonData.Metrics.AnalysisTime = p.statsData.AnalysisDuration
	}
	if p.statsData.Timestamp != "" {
		jsonData.Metrics.Timestamp = p.statsData.Timestamp
	}

	// Fill size distribution
	jsonData.SizeDistribution = p.statsData.SizeDistribution

	// Fill top files
	if len(p.statsData.FileDuplication) > 0 {
		// Convert map to slice and sort by duplicate lines (descending)
		type fileStat struct {
			filename string
			lines    int
		}
		files := make([]fileStat, 0, len(p.statsData.FileDuplication))
		for filename, lines := range p.statsData.FileDuplication {
			files = append(files, fileStat{filename, lines})
		}

		// Sort by lines descending
		for i := 0; i < len(files)-1; i++ {
			for j := i + 1; j < len(files); j++ {
				if files[i].lines < files[j].lines {
					files[i], files[j] = files[j], files[i]
				}
			}
		}

		// Take top 10
		limit := min(len(files), 10)
		jsonData.TopFiles = make([]struct {
			Filename string `json:"filename"`
			Lines    int    `json:"duplicateLines"`
		}, limit)
		for i := range limit {
			jsonData.TopFiles[i].Filename = files[i].filename
			jsonData.TopFiles[i].Lines = files[i].lines
		}
	}

	data, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		fmt.Fprintf(p.w, "Error encoding JSON: %v\n", err)
	} else {
		if _, err := p.w.Write(data); err != nil {
			fmt.Fprintf(p.w, "Error writing JSON: %v\n", err)
		}
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

		fmt.Fprintf(w, "  %-15s: %4d clones [%s] %.1f%%\n", r, count, bar, percentage)
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
		fmt.Fprintf(w, "  %d lines in %s\n", files[i].lines, files[i].filename)
	}

	if len(files) > topN {
		fmt.Fprintf(w, "  ... and %d more files\n", len(files)-topN)
	}
}

// GetStatsData returns the collected statistics data.
func (p *stats) GetStatsData() interface{} {
	return p.statsData
}

// SetDetectionMethods sets the detection methods used.
func (p *stats) SetDetectionMethods(methods string) {
	p.statsData.DetectionMethods = methods
}

// SetFormat sets the output format.
func (p *stats) SetFormat(format Format) {
	p.format = format
}
