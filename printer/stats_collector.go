package printer

import "time"

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
func (p *stats) SetFormat(format Format) {
	p.format = format
}

// GetStatsData returns the collected statistics data.
func (p *stats) GetStatsData() any {
	return p.statsData
}
