package printer

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/LarsArtmann/art-dupl/config"
)

// roundFloat rounds a float64 to the specified number of decimal places.
func roundFloat(v float64, decimals int) float64 {
	multiplier := math.Pow(10, float64(decimals))
	return math.Round(v*multiplier) / multiplier
}

// jsonStatsOutput represents the JSON output structure for statistics.
type jsonStatsOutput struct {
	Configuration struct {
		Threshold         int    `json:"threshold"`
		DetectionMethods  string `json:"detectionMethods"`
		DetectionMode     string `json:"detectionMode,omitempty"`
		DetectionModeDesc string `json:"detectionModeDescription,omitempty"`
		SemanticDetection bool   `json:"semanticDetection"`
	} `json:"configuration"`
	Overview struct {
		FilesScanned    int            `json:"filesScanned"`
		FilesFiltered   int            `json:"filesFiltered,omitempty"`
		FilterBreakdown map[string]int `json:"filterBreakdown,omitempty"`
		CloneGroups     int            `json:"cloneGroups"`
		TotalClones     int            `json:"totalClones"`
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
		HealthScore           string `json:"healthScore,omitempty"`
		HealthScoreThresholds string `json:"healthScoreThresholds,omitempty"`
		AnalysisTime          string `json:"analysisTime,omitempty"`
		Timestamp             string `json:"timestamp,omitempty"`
	} `json:"metrics"`
	Note              string         `json:"note"`
	SizeDistribution  map[string]int `json:"sizeDistribution"`
	TokenDistribution map[string]int `json:"tokenDistribution"`
	SeverityBreakdown map[string]int `json:"severityBreakdown"`
	CategoryBreakdown map[string]int `json:"categoryBreakdown,omitempty"`
	PriorityBreakdown map[string]int `json:"priorityBreakdown,omitempty"`
	Actionability     struct {
		Actionable    int `json:"actionable,omitempty"`
		NonActionable int `json:"nonActionable,omitempty"`
	} `json:"actionability,omitempty"`
	TopFiles          []jsonTopFile  `json:"topFiles"`
}

type jsonTopFile struct {
	Filename string `json:"filename"`
	Lines    int    `json:"duplicateLines"`
}

// printStats prints the collected statistics.
func (p *stats) printStats() {
	switch p.format {
	case config.OutputFormatJSON:
		p.printJSON()
	case config.OutputFormatCSV:
		p.printCSV()
	case config.OutputFormatText:
		p.printText()
	case config.OutputFormatHTML, config.OutputFormatPlumbing,
		config.OutputFormatSimpleJSON, config.OutputFormatSARIF:
		p.printText()
	}
}

// printCSV prints statistics in CSV format.
func (p *stats) printCSV() {
	// Write CSV header
	_, _ = fmt.Fprintf(p.w, "Metric,Value\n")

	// Configuration
	_, _ = fmt.Fprintf(p.w, "Threshold,%d\n", p.threshold)
	_, _ = fmt.Fprintf(p.w, "Detection Methods,%s\n", p.statsData.DetectionMethods)

	if p.statsData.DetectionMode != "" {
		_, _ = fmt.Fprintf(p.w, "Detection Mode,%s\n", p.statsData.DetectionMode)
	}

	if p.statsData.DetectionModeDesc != "" {
		_, _ = fmt.Fprintf(p.w, "Detection Mode Description,%s\n", p.statsData.DetectionModeDesc)
	}

	_, _ = fmt.Fprintf(p.w, "Semantic Detection,%t\n", p.statsData.SemanticDetection)
	_, _ = fmt.Fprintf(p.w, "Timestamp,%s\n", p.statsData.Timestamp)
	_, _ = fmt.Fprintf(p.w, "Analysis Time,%s\n", p.statsData.AnalysisDuration)
	_, _ = fmt.Fprintf(p.w, "\n")

	// Overview
	_, _ = fmt.Fprintf(p.w, "Files Scanned,%d\n", p.statsData.TotalFilesScanned)
	if p.statsData.FilesFiltered > 0 {
		_, _ = fmt.Fprintf(p.w, "Files Filtered,%d\n", p.statsData.FilesFiltered)
	}

	_, _ = fmt.Fprintf(p.w, "Clone Groups,%d\n", p.statsData.TotalCloneGroups)
	_, _ = fmt.Fprintf(p.w, "Total Clones,%d\n", p.statsData.TotalClones)
	_, _ = fmt.Fprintf(p.w, "\n")

	// Duplicate Code
	_, _ = fmt.Fprintf(p.w, "Total Duplicate Lines,%d\n", p.statsData.TotalDuplicateLines)
	_, _ = fmt.Fprintf(p.w, "Estimated Total Lines,%d\n", p.statsData.TotalEstimatedLines)
	_, _ = fmt.Fprintf(p.w, "Duplication Ratio,%.1f%%\n", p.statsData.DuplicationRatio)
	_, _ = fmt.Fprintf(p.w, "Total Duplicate Tokens,%d\n", p.statsData.TotalTokens)
	_, _ = fmt.Fprintf(p.w, "Average Clone Size,%d\n", p.statsData.AverageCloneSize)
	_, _ = fmt.Fprintf(p.w, "Complexity Score,%.2f\n", p.statsData.ComplexityScore)
	_, _ = fmt.Fprintf(p.w, "Impact Score,%d\n", p.statsData.ImpactScore)
	_, _ = fmt.Fprintf(p.w, "Health Score,%s\n", p.statsData.HealthScore)
	_, _ = fmt.Fprintf(
		p.w,
		"Health Score Thresholds,A: <5%% dup, B: <10%%, C: <15%%, D: <25%%, F: >=25%%\n",
	)
	_, _ = fmt.Fprintf(p.w, "\n")
	_, _ = fmt.Fprintf(p.w, "Note,Metrics count unique duplicate patterns not total occurrences\n")
}

// printText prints statistics in text format.
func (p *stats) printText() {
	p.printTextHeader()
	p.printTextConfiguration()
	p.printTextOverview()
	p.printTextDuplicateCode()
	p.printTextDistributions()
	p.printTextRecommendations()
}

func (p *stats) printTextHeader() {
	p.printHeader("Code Duplication Statistics")
	p.printLinef("============================")
}

func (p *stats) printTextConfiguration() {
	p.printSection("Configuration:")
	p.printMetric("Threshold", fmt.Sprintf("%d tokens", p.threshold))
	p.printMetric("Detection Methods", p.statsData.DetectionMethods)

	if p.statsData.DetectionMode != "" {
		p.printMetric("Detection Mode", p.statsData.DetectionMode)
	}

	if p.statsData.DetectionModeDesc != "" {
		p.printLinef("(%s)", p.statsData.DetectionModeDesc)
	}

	if p.statsData.Timestamp != "" {
		p.printMetric("Timestamp", p.statsData.Timestamp)
	}

	if p.statsData.AnalysisDuration != "" {
		p.printMetric("Analysis Time", p.statsData.AnalysisDuration)
	}

	_, _ = fmt.Fprintf(p.w, "\n")
}

func (p *stats) printTextOverview() {
	p.printSection("Overview:")
	p.printMetric("Files Scanned", strconv.Itoa(p.statsData.TotalFilesScanned))

	if p.statsData.FilesFiltered > 0 {
		p.printFilterBreakdown()
	}

	p.printMetric("Clone Groups", strconv.Itoa(p.statsData.TotalCloneGroups))
	p.printMetric("Total Clones", strconv.Itoa(p.statsData.TotalClones))
	_, _ = fmt.Fprintf(p.w, "\n")
}

func (p *stats) printFilterBreakdown() {
	filterPercent := float64(p.statsData.FilesFiltered) /
		float64(p.statsData.TotalFilesScanned+p.statsData.FilesFiltered) * 100
	p.printMetric(
		"Files Filtered",
		fmt.Sprintf("%d (%.0f%%)", p.statsData.FilesFiltered, filterPercent),
	)

	if len(p.statsData.FilterBreakdown) > 0 {
		_, _ = fmt.Fprintf(p.w, "\n%s\n", p.section.Render("Filtering Breakdown:"))
		for reason, count := range p.statsData.FilterBreakdown {
			_, _ = fmt.Fprintf(p.w, "  %s %s: %s\n",
				p.metric.Render("•"),
				p.base.Render(reason),
				p.base.Render(fmt.Sprintf("%d files", count)))
		}
	}
}

func (p *stats) printTextDuplicateCode() {
	p.printSection("Duplicate Code:")
	p.printMetric("Total Duplicate Lines", strconv.Itoa(p.statsData.TotalDuplicateLines))

	if p.statsData.TotalEstimatedLines > 0 {
		p.printMetric("Estimated Total Lines", strconv.Itoa(p.statsData.TotalEstimatedLines))
		p.printMetric("Duplication Ratio", fmt.Sprintf("%.1f%%", p.statsData.DuplicationRatio))
	}

	p.printMetric("Total Duplicate Tokens", strconv.Itoa(p.statsData.TotalTokens))
	p.printMetric("Average Clone Size", fmt.Sprintf("%d lines", p.statsData.AverageCloneSize))
	p.printMetric("Complexity Score", fmt.Sprintf("%.2f", p.statsData.ComplexityScore))
	p.printMetric("Impact Score", strconv.Itoa(p.statsData.ImpactScore))

	if p.statsData.HealthScore != "" {
		styled := p.healthScoreStyle(p.statsData.HealthScore).Render(p.statsData.HealthScore)
		p.printMetric("Health Score", styled)
		p.printLinef("  (A: <5%% dup, B: <10%%, C: <15%%, D: <25%%, F: >=25%%)")
	}

	_, _ = fmt.Fprintf(p.w, "\n")
}

func (p *stats) printTextDistributions() {
	if len(p.statsData.SizeDistribution) > 0 {
		p.printSection("Clone Size Distribution:")
		printSizeDistribution(p.w, p.statsData.SizeDistribution)
		_, _ = fmt.Fprintf(p.w, "\n")
	}

	if len(p.statsData.SeverityBreakdown) > 0 {
		p.printSection("Clone Severity:")
		printSeverityDistribution(p.w, p.statsData.SeverityBreakdown)
		_, _ = fmt.Fprintf(p.w, "\n")
	}

	if len(p.statsData.CategoryBreakdown) > 0 {
		p.printSection("Clone Categories:")
		printCategoryDistribution(p.w, p.statsData.CategoryBreakdown)
		_, _ = fmt.Fprintf(p.w, "\n")
	}

	if len(p.statsData.PriorityBreakdown) > 0 {
		p.printSection("Clone Priority:")
		printPriorityDistribution(p.w, p.statsData.PriorityBreakdown)
		_, _ = fmt.Fprintf(p.w, "\n")
	}

	if p.statsData.ActionableGroups > 0 || p.statsData.NonActionableGroups > 0 {
		p.printSection("Actionability:")
		p.printMetric("Actionable Groups", strconv.Itoa(p.statsData.ActionableGroups))
		if p.statsData.NonActionableGroups > 0 {
			p.printMetric("Non-Actionable Groups", strconv.Itoa(p.statsData.NonActionableGroups))
		}
		_, _ = fmt.Fprintf(p.w, "\n")
	}

	if len(p.statsData.FileDuplication) > 0 {
		p.printSection("Top Files by Duplicate Lines:")
		printTopFiles(p.w, p.statsData.FileDuplication, 10)
	}
}

func (p *stats) printTextRecommendations() {
	_, _ = fmt.Fprintf(p.w, "\n%s\n", p.header.Render("Recommendations:"))
	p.printRecommendations()

	p.printSection("Note:")
	p.printLinef("Metrics count unique duplicate patterns, not total occurrences.")
	p.printLinef("A clone group with 3 instances counts once for line calculations.")
}

// FileStatMixin provides common fields for file statistics.
type FileStatMixin struct {
	filename string
	lines    int
}

// topFileStat holds file duplication statistics for sorting.
type topFileStat struct {
	FileStatMixin
}

// sortTopFiles returns the top N files sorted by duplicate lines.
func sortTopFiles(fileDuplication map[string]int, limit int) []topFileStat {
	files := make([]topFileStat, 0, len(fileDuplication))
	for filename, lines := range fileDuplication {
		files = append(files, topFileStat{FileStatMixin{filename, lines}})
	}

	// Sort by lines descending
	for i := range len(files) - 1 {
		for j := i + 1; j < len(files); j++ {
			if files[i].lines < files[j].lines {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	return files[:min(len(files), limit)]
}

// printJSON prints statistics in JSON format.
func (p *stats) printJSON() {
	jsonData := p.buildJSONData()

	data, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		_, _ = fmt.Fprintf(p.w, "Error encoding JSON: %v\n", err)
	} else {
		if _, err := p.w.Write(data); err != nil {
			_, _ = fmt.Fprintf(p.w, "Error writing JSON: %v\n", err)
		}
	}
}

// buildJSONData constructs the JSON output structure.
func (p *stats) buildJSONData() jsonStatsOutput {
	var jsonData jsonStatsOutput

	// Fill configuration
	jsonData.Configuration.Threshold = p.threshold
	jsonData.Configuration.DetectionMethods = p.statsData.DetectionMethods
	jsonData.Configuration.SemanticDetection = p.statsData.SemanticDetection
	jsonData.Configuration.DetectionMode = p.statsData.DetectionMode
	jsonData.Configuration.DetectionModeDesc = p.statsData.DetectionModeDesc

	// Fill overview
	jsonData.Overview.FilesScanned = p.statsData.TotalFilesScanned
	jsonData.Overview.CloneGroups = p.statsData.TotalCloneGroups
	jsonData.Overview.TotalClones = p.statsData.TotalClones

	// Fill filter info if available
	if p.statsData.FilesFiltered > 0 {
		jsonData.Overview.FilesFiltered = p.statsData.FilesFiltered
		if len(p.statsData.FilterBreakdown) > 0 {
			jsonData.Overview.FilterBreakdown = p.statsData.FilterBreakdown
		}
	}

	// Fill duplicate code metrics
	jsonData.DuplicateCode.TotalLines = p.statsData.TotalDuplicateLines
	jsonData.DuplicateCode.TotalTokens = p.statsData.TotalTokens
	jsonData.DuplicateCode.AverageCloneSize = p.statsData.AverageCloneSize
	jsonData.DuplicateCode.ComplexityScore = roundFloat(p.statsData.ComplexityScore, 2)

	jsonData.DuplicateCode.ImpactScore = p.statsData.ImpactScore
	if p.statsData.TotalEstimatedLines > 0 {
		jsonData.DuplicateCode.EstimatedLines = p.statsData.TotalEstimatedLines
		jsonData.DuplicateCode.DuplicationRatio = roundFloat(p.statsData.DuplicationRatio, 2)
	}

	// Fill metrics
	if p.statsData.HealthScore != "" {
		jsonData.Metrics.HealthScore = p.statsData.HealthScore
		jsonData.Metrics.HealthScoreThresholds = "A: <5% dup, B: <10%, C: <15%, D: <25%, F: >=25%"
	}

	if p.statsData.AnalysisDuration != "" {
		jsonData.Metrics.AnalysisTime = p.statsData.AnalysisDuration
	}

	if p.statsData.Timestamp != "" {
		jsonData.Metrics.Timestamp = p.statsData.Timestamp
	}

	// Fill size distribution
	jsonData.SizeDistribution = p.statsData.SizeDistribution

	// Fill token distribution
	jsonData.TokenDistribution = p.statsData.TokenDistribution

	// Fill severity breakdown
	jsonData.SeverityBreakdown = p.statsData.SeverityBreakdown

	// Fill category breakdown
	if len(p.statsData.CategoryBreakdown) > 0 {
		jsonData.CategoryBreakdown = p.statsData.CategoryBreakdown
	}

	// Fill priority breakdown
	if len(p.statsData.PriorityBreakdown) > 0 {
		jsonData.PriorityBreakdown = p.statsData.PriorityBreakdown
	}

	// Fill actionability
	if p.statsData.ActionableGroups > 0 || p.statsData.NonActionableGroups > 0 {
		jsonData.Actionability.Actionable = p.statsData.ActionableGroups
		jsonData.Actionability.NonActionable = p.statsData.NonActionableGroups
	}

	// Add methodology note
	jsonData.Note = "Metrics count unique duplicate patterns, not total occurrences. A clone group with 3 instances counts once for line calculations."

	// Fill top files
	if len(p.statsData.FileDuplication) > 0 {
		files := sortTopFiles(p.statsData.FileDuplication, 10)

		jsonData.TopFiles = make([]jsonTopFile, len(files))
		for i := range files {
			jsonData.TopFiles[i] = jsonTopFile{
				Filename: files[i].filename,
				Lines:    files[i].lines,
			}
		}
	}

	return jsonData
}
