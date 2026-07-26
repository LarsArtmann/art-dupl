package stats

import (
	"time"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/printer"
)

// ApplyStatsConfig applies all statistics configuration in one call.
func (p *stats) ApplyStatsConfig(config printer.StatsConfig) {
	p.SetFormat(config.Format)
	p.SetFilesCount(config.FilesCount)
	p.SetDetectionMethods(config.DetectionMethods)
	p.SetSemanticDetection(config.SemanticDetection)
	p.SetTimestamp(config.Timestamp)
	p.SetAnalysisDuration(config.AnalysisDuration)
	p.SetTotalEstimatedLines(config.TotalEstimatedLines)

	if config.FilesFiltered > 0 || len(config.FilterBreakdown) > 0 {
		p.SetFilterStats(config.FilesFiltered, config.FilterBreakdown)
	}

	if len(config.FilterSourceBreakdown) > 0 {
		p.SetFilterSourceStats(config.FilterSourceBreakdown)
	}
}

// SetFilesCount sets the total number of files scanned.
func (p *stats) SetFilesCount(count int) {
	p.statsData.TotalFilesScanned = count
}

// SetFilterStats sets the filter statistics from the filter package.
// This allows the stats printer to report how many files were filtered and why.
func (p *stats) SetFilterStats(filesFiltered int, breakdown map[string]int) {
	p.statsData.FilesFiltered = filesFiltered
	p.statsData.FilterBreakdown = breakdown
}

// SetFilterSourceStats sets the per-source breakdown of filtered files,
// distinguishing gogenfilter catches from defense-in-depth content checks.
func (p *stats) SetFilterSourceStats(sourceBreakdown map[string]int) {
	p.statsData.FilterSourceBreakdown = sourceBreakdown
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

// SetDetectionMethods sets the detection methods used.
func (p *stats) SetDetectionMethods(methods string) {
	p.statsData.DetectionMethods = methods
}

// SetSemanticDetection sets whether semantic detection was enabled
// and updates the detection mode description accordingly.
func (p *stats) SetSemanticDetection(enabled bool) {
	p.statsData.SemanticDetection = enabled

	mode, desc := semanticModeConfig(enabled)
	p.statsData.DetectionMode = mode
	p.statsData.DetectionModeDesc = desc
}

func semanticModeConfig(enabled bool) (string, string) {
	if enabled {
		return "semantic", "Matches clones by structure AND identifier names (fewer false positives)"
	}

	return "structural", "Matches clones by AST structure only (may include code with different identifiers)"
}

// SetFormat sets the output format.
func (p *stats) SetFormat(format config.OutputFormat) {
	p.format = format
}

// GetStatsView returns the collected statistics data.
func (p *stats) GetStatsView() *printer.StatsView {
	return p.statsData
}
