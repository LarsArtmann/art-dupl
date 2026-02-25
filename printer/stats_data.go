package printer

// stats_data.go - Stats data structures
//
// This file contains the StatsData type definition for aggregated
// statistics about code duplication.
//
// Design:
// - StatsData is a simple struct for JSON marshaling
// - No methods here, just data definition
// - Formatting and calculation logic are in separate files

// StatsData holds all aggregated statistics about code duplication analysis.
//
// Fields:
// - Count metrics: TotalFilesScanned, TotalCloneGroups, TotalClones
// - Size metrics: TotalTokens, TotalDuplicateLines, AverageCloneSize
// - Complexity metrics: ComplexityScore, ImpactScore
// - Quality metrics: DuplicationRatio, HealthScore
// - Time metrics: AnalysisDuration, Timestamp
// - Aggregation metrics: FileDuplication, SizeDistribution
// - Filter metrics: FilesFiltered, FilterBreakdown (NEW)
// - Metadata: DetectionMethods
//
// Domain Types Status:
// - Uses primitive types (int, float64, string) for JSON compatibility
// - Could use domain types (FileCount, TokenCount, etc.) in future
// - See TODO in stats.go for migration path
//
// JSON Marshaling:
// - All fields are JSON tagged for easy marshaling
// - Use printer.JSONPrinter for formatted JSON output.
type StatsData struct {
	// Count metrics
	TotalFilesScanned int `json:"total_files_scanned"`
	TotalCloneGroups  int `json:"total_clone_groups"`
	TotalClones       int `json:"total_clones"`

	// Size metrics
	TotalDuplicateLines int `json:"total_duplicate_lines"`
	TotalTokens         int `json:"total_tokens"`
	TotalEstimatedLines int `json:"total_estimated_lines"` // Estimated total lines for duplication percentage
	AverageCloneSize    int `json:"average_clone_size"`

	// Complexity and impact metrics
	ComplexityScore  float64 `json:"complexity_score"`
	ImpactScore      int     `json:"impact_score"`
	DuplicationRatio float64 `json:"duplication_ratio"` // Percentage of duplicated code

	// Quality metrics
	HealthScore string `json:"health_score"` // A-F grade based on metrics

	// Time metrics
	AnalysisDuration string `json:"analysis_duration"` // Time taken for analysis
	Timestamp        string `json:"timestamp"`         // ISO 8601 timestamp

	// Aggregation metrics
	FileDuplication   map[string]int `json:"file_duplication"`   // filename -> duplicate line count
	SizeDistribution  map[string]int `json:"size_distribution"`  // size range -> count (lines)
	TokenDistribution map[string]int `json:"token_distribution"` // token range -> count

	// Filter metrics (NEW)
	FilesFiltered   int            `json:"files_filtered,omitempty"`   // Total files filtered out
	FilterBreakdown map[string]int `json:"filter_breakdown,omitempty"` // Reason -> count (e.g., "templ" -> 12)

	// Metadata
	DetectionMethods string `json:"detection_methods"` // Comma-separated detection methods used
}
