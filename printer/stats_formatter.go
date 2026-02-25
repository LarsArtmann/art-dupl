package printer

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// jsonStatsOutput represents the JSON output structure for statistics.
type jsonStatsOutput struct {
	Configuration struct {
		Threshold        int    `json:"threshold"`
		DetectionMethods string `json:"detectionMethods"`
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
		HealthScore  string `json:"healthScore,omitempty"`
		AnalysisTime string `json:"analysisTime,omitempty"`
		Timestamp    string `json:"timestamp,omitempty"`
	} `json:"metrics"`
	Note              string         `json:"note"`
	SizeDistribution  map[string]int `json:"sizeDistribution"`
	TokenDistribution map[string]int `json:"tokenDistribution"`
	TopFiles          []jsonTopFile  `json:"topFiles"`
}

type jsonTopFile struct {
	Filename string `json:"filename"`
	Lines    int    `json:"duplicateLines"`
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
	_, _ = fmt.Fprintf(p.w, "Metric,Value\n")

	// Configuration
	_, _ = fmt.Fprintf(p.w, "Threshold,%d\n", p.threshold)
	_, _ = fmt.Fprintf(p.w, "Detection Methods,%s\n", p.statsData.DetectionMethods)
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
	_, _ = fmt.Fprintf(p.w, "\n")
	_, _ = fmt.Fprintf(p.w, "Note,Metrics count unique duplicate patterns not total occurrences\n")
}

// printText prints statistics in text format.
func (p *stats) printText() {
	// Print header
	p.printHeader("Code Duplication Statistics")
	p.printLine("============================")

	// Print configuration
	p.printSection("Configuration:")
	p.printMetric("Threshold", fmt.Sprintf("%d tokens", p.threshold))
	p.printMetric("Detection Methods", p.statsData.DetectionMethods)
	if p.statsData.Timestamp != "" {
		p.printMetric("Timestamp", p.statsData.Timestamp)
	}
	if p.statsData.AnalysisDuration != "" {
		p.printMetric("Analysis Time", p.statsData.AnalysisDuration)
	}
	_, _ = fmt.Fprintf(p.w, "\n")

	// Print overview
	p.printSection("Overview:")
	p.printMetric("Files Scanned", strconv.Itoa(p.statsData.TotalFilesScanned))

	// Print filter information if files were filtered
	if p.statsData.FilesFiltered > 0 {
		filterPercent := float64(p.statsData.FilesFiltered) / float64(p.statsData.TotalFilesScanned+p.statsData.FilesFiltered) * 100
		filterText := fmt.Sprintf("%d (%.0f%%)", p.statsData.FilesFiltered, filterPercent)
		p.printMetric("Files Filtered", filterText)

		// Print filter breakdown
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

	p.printMetric("Clone Groups", strconv.Itoa(p.statsData.TotalCloneGroups))
	p.printMetric("Total Clones", strconv.Itoa(p.statsData.TotalClones))
	_, _ = fmt.Fprintf(p.w, "\n")

	// Print duplicate code metrics
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
		styledHealthScore := p.healthScoreStyle(p.statsData.HealthScore).Render(p.statsData.HealthScore)
		p.printMetric("Health Score", styledHealthScore)
	}
	_, _ = fmt.Fprintf(p.w, "\n")

	// Print size distribution
	if len(p.statsData.SizeDistribution) > 0 {
		p.printSection("Clone Size Distribution:")
		printSizeDistribution(p.w, p.statsData.SizeDistribution)
		_, _ = fmt.Fprintf(p.w, "\n")
	}

	// Print top files with most duplicates
	if len(p.statsData.FileDuplication) > 0 {
		p.printSection("Top Files by Duplicate Lines:")
		printTopFiles(p.w, p.statsData.FileDuplication, 10)
	}

	// Print actionable recommendations
	_, _ = fmt.Fprintf(p.w, "\n%s\n", p.header.Render("Recommendations:"))
	p.printRecommendations()

	// Print methodology note
	p.printSection("Note:")
	p.printLine("Metrics count unique duplicate patterns, not total occurrences.")
	p.printLine("A clone group with 3 instances counts once for line calculations.")
}

// topFileStat holds file duplication statistics for sorting.
type topFileStat struct {
	filename string
	lines    int
}

// sortTopFiles returns the top N files sorted by duplicate lines.
func sortTopFiles(fileDuplication map[string]int, limit int) []topFileStat {
	files := make([]topFileStat, 0, len(fileDuplication))
	for filename, lines := range fileDuplication {
		files = append(files, topFileStat{filename, lines})
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
func (p *stats) buildJSONData() any {
	var jsonData jsonStatsOutput

	// Fill configuration
	jsonData.Configuration.Threshold = p.threshold
	jsonData.Configuration.DetectionMethods = p.statsData.DetectionMethods

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

	// Fill token distribution
	jsonData.TokenDistribution = p.statsData.TokenDistribution

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
