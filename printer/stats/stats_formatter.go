package stats

import (
	"encoding/csv"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"fmt"
	"math"
	"strconv"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/printer"
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
		FilesScanned          int            `json:"filesScanned"`
		FilesFiltered         int            `json:"filesFiltered,omitempty"`
		FilterBreakdown       map[string]int `json:"filterBreakdown,omitempty"`
		FilterSourceBreakdown map[string]int `json:"filterSourceBreakdown,omitempty"`
		CloneGroups           int            `json:"cloneGroups"`
		TotalClones           int            `json:"totalClones"`
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
	} `json:"actionability"`
	TestVsProduction struct {
		Production int `json:"production,omitempty"`
		Test       int `json:"test,omitempty"`
	} `json:"testVsProduction"`
	TopClones []printer.TopCloneGroup `json:"topClones,omitempty"`
	TopFiles  []jsonTopFile   `json:"topFiles"`
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

// printCSV prints statistics in CSV format using encoding/csv for proper escaping.
func (p *stats) printCSV() {
	w := csv.NewWriter(p.w)
	write := func(record ...string) {
		_ = w.Write(record)
	}

	// Configuration
	write("Metric", "Value")
	write("Threshold", strconv.Itoa(p.threshold))
	write("Detection Methods", p.statsData.DetectionMethods)

	if p.statsData.DetectionMode != "" {
		write("Detection Mode", p.statsData.DetectionMode)
	}

	if p.statsData.DetectionModeDesc != "" {
		write("Detection Mode Description", p.statsData.DetectionModeDesc)
	}

	write("Semantic Detection", strconv.FormatBool(p.statsData.SemanticDetection))
	write("Timestamp", p.statsData.Timestamp)
	write("Analysis Time", p.statsData.AnalysisDuration)

	// Blank separator
	write()

	// Overview
	write("Files Scanned", strconv.Itoa(p.statsData.TotalFilesScanned))

	if p.statsData.FilesFiltered > 0 {
		write("Files Filtered", strconv.Itoa(p.statsData.FilesFiltered))
	}

	write("Clone Groups", strconv.Itoa(p.statsData.TotalCloneGroups))
	write("Total Clones", strconv.Itoa(p.statsData.TotalClones))

	// Blank separator
	write()

	// Duplicate Code
	write("Total Duplicate Lines", strconv.Itoa(p.statsData.TotalDuplicateLines))
	write("Estimated Total Lines", strconv.Itoa(p.statsData.TotalEstimatedLines))
	write("Duplication Ratio", fmt.Sprintf("%.1f%%", p.statsData.DuplicationRatio))
	write("Total Duplicate Tokens", strconv.Itoa(p.statsData.TotalTokens))
	write("Average Clone Size", strconv.Itoa(p.statsData.AverageCloneSize))
	write("Complexity Score", fmt.Sprintf("%.2f", p.statsData.ComplexityScore))
	write("Impact Score", strconv.Itoa(p.statsData.ImpactScore))
	write("Health Score", string(p.statsData.HealthScore))
	write("Health Score Thresholds", "Weighted score: A<5, B<10, C<15, D<25, F>=25")

	// Blank separator
	write()

	write("Note", "Metrics count unique duplicate patterns not total occurrences")

	w.Flush()
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

	if len(p.statsData.FilterSourceBreakdown) > 0 {
		_, _ = fmt.Fprintf(p.w, "\n%s\n", p.section.Render("Filter Source:"))
		for source, count := range p.statsData.FilterSourceBreakdown {
			_, _ = fmt.Fprintf(p.w, "  %s %s: %s\n",
				p.metric.Render("•"),
				p.base.Render(source),
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
		styled := p.healthScoreStyle(p.statsData.HealthScore).Render(string(p.statsData.HealthScore))
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

	if p.statsData.TestCloneGroups > 0 || p.statsData.ProductionCloneGroups > 0 {
		p.printSection("Test vs Production:")
		p.printMetric("Production Groups", strconv.Itoa(p.statsData.ProductionCloneGroups))
		p.printMetric("Test Groups", strconv.Itoa(p.statsData.TestCloneGroups))
		_, _ = fmt.Fprintf(p.w, "\n")
	}

	if len(p.statsData.TopClones) > 0 {
		p.printSection("Top Clones to Fix:")
		p.printTopClones()
		_, _ = fmt.Fprintf(p.w, "\n")
	}

	if len(p.statsData.FileDuplication) > 0 {
		p.printSection("Top Files by Duplicate Lines:")
		printTopFiles(p.w, p.statsData.FileDuplication, 10)
	}
}

func (p *stats) printTopClones() {
	for i, clone := range p.statsData.TopClones {
		badge := fmt.Sprintf("[%s] %s", clone.Priority, clone.Category)
		_, _ = fmt.Fprintf(
			p.w,
			"  %d. %s | %d lines in %d files\n",
			i+1,
			badge,
			clone.Lines,
			clone.Files,
		)
		_, _ = fmt.Fprintf(
			p.w,
			"     %s:%d  →  %s\n",
			clone.FirstFile,
			clone.FirstLineStart,
			clone.Suggestion,
		)
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

	data, err := json.Marshal(jsonData, jsontext.WithIndentPrefix(""), jsontext.WithIndent("  "))
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

	p.fillJSONConfig(&jsonData)
	p.fillJSONOverview(&jsonData)
	p.fillJSONDuplicateCode(&jsonData)
	p.fillJSONMetrics(&jsonData)
	p.fillJSONBreakdowns(&jsonData)
	p.fillJSONTopFiles(&jsonData)

	jsonData.Note = "Metrics count unique duplicate patterns, not total occurrences. A clone group with 3 instances counts once for line calculations."

	return jsonData
}

func (p *stats) fillJSONConfig(jsonData *jsonStatsOutput) {
	jsonData.Configuration.Threshold = p.threshold
	jsonData.Configuration.DetectionMethods = p.statsData.DetectionMethods
	jsonData.Configuration.SemanticDetection = p.statsData.SemanticDetection
	jsonData.Configuration.DetectionMode = p.statsData.DetectionMode
	jsonData.Configuration.DetectionModeDesc = p.statsData.DetectionModeDesc
}

func (p *stats) fillJSONOverview(jsonData *jsonStatsOutput) {
	jsonData.Overview.FilesScanned = p.statsData.TotalFilesScanned
	jsonData.Overview.CloneGroups = p.statsData.TotalCloneGroups
	jsonData.Overview.TotalClones = p.statsData.TotalClones

	if p.statsData.FilesFiltered > 0 {
		jsonData.Overview.FilesFiltered = p.statsData.FilesFiltered
		if len(p.statsData.FilterBreakdown) > 0 {
			jsonData.Overview.FilterBreakdown = p.statsData.FilterBreakdown
		}

		if len(p.statsData.FilterSourceBreakdown) > 0 {
			jsonData.Overview.FilterSourceBreakdown = p.statsData.FilterSourceBreakdown
		}
	}
}

func (p *stats) fillJSONDuplicateCode(jsonData *jsonStatsOutput) {
	jsonData.DuplicateCode.TotalLines = p.statsData.TotalDuplicateLines
	jsonData.DuplicateCode.TotalTokens = p.statsData.TotalTokens
	jsonData.DuplicateCode.AverageCloneSize = p.statsData.AverageCloneSize
	jsonData.DuplicateCode.ComplexityScore = roundFloat(p.statsData.ComplexityScore, 2)
	jsonData.DuplicateCode.ImpactScore = p.statsData.ImpactScore

	if p.statsData.TotalEstimatedLines > 0 {
		jsonData.DuplicateCode.EstimatedLines = p.statsData.TotalEstimatedLines
		jsonData.DuplicateCode.DuplicationRatio = roundFloat(p.statsData.DuplicationRatio, 2)
	}
}

func (p *stats) fillJSONMetrics(jsonData *jsonStatsOutput) {
	if p.statsData.HealthScore != "" {
		jsonData.Metrics.HealthScore = string(p.statsData.HealthScore)
		jsonData.Metrics.HealthScoreThresholds = "Weighted score: A<5, B<10, C<15, D<25, F>=25"
	}

	if p.statsData.AnalysisDuration != "" {
		jsonData.Metrics.AnalysisTime = p.statsData.AnalysisDuration
	}

	if p.statsData.Timestamp != "" {
		jsonData.Metrics.Timestamp = p.statsData.Timestamp
	}
}

func (p *stats) fillJSONBreakdowns(jsonData *jsonStatsOutput) {
	jsonData.SizeDistribution = p.statsData.SizeDistribution
	jsonData.TokenDistribution = p.statsData.TokenDistribution
	jsonData.SeverityBreakdown = p.statsData.SeverityBreakdown

	if len(p.statsData.CategoryBreakdown) > 0 {
		jsonData.CategoryBreakdown = p.statsData.CategoryBreakdown
	}

	if len(p.statsData.PriorityBreakdown) > 0 {
		jsonData.PriorityBreakdown = p.statsData.PriorityBreakdown
	}

	if p.hasActionabilityData() {
		jsonData.Actionability.Actionable = p.statsData.ActionableGroups
		jsonData.Actionability.NonActionable = p.statsData.NonActionableGroups
	}

	if p.hasTestProdData() {
		jsonData.TestVsProduction.Production = p.statsData.ProductionCloneGroups
		jsonData.TestVsProduction.Test = p.statsData.TestCloneGroups
	}

	if len(p.statsData.TopClones) > 0 {
		jsonData.TopClones = p.statsData.TopClones
	}
}

func (p *stats) hasActionabilityData() bool {
	return p.statsData.ActionableGroups > 0 || p.statsData.NonActionableGroups > 0
}

func (p *stats) hasTestProdData() bool {
	return p.statsData.TestCloneGroups > 0 || p.statsData.ProductionCloneGroups > 0
}

func (p *stats) fillJSONTopFiles(jsonData *jsonStatsOutput) {
	if len(p.statsData.FileDuplication) > 0 {
		files := sortTopFiles(p.statsData.FileDuplication, 10)

		jsonData.TopFiles = make([]jsonTopFile, 0, len(files))
		for _, f := range files {
			jsonData.TopFiles = append(jsonData.TopFiles, jsonTopFile{
				Filename: f.filename,
				Lines:    f.lines,
			})
		}
	}
}
