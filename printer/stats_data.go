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

// TopCloneGroup represents a high-impact clone group for the "top clones to fix" preview.
type TopCloneGroup struct {
	Priority       string `json:"priority"`
	Category       string `json:"category"`
	Lines          int    `json:"lines"`
	Files          int    `json:"files"`
	Suggestion     string `json:"suggestion"`
	FirstFile      string `json:"firstFile"`
	FirstLineStart int    `json:"firstLineStart"`
}

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
	FileDuplication   map[string]int `json:"file_duplication"`             // filename -> duplicate line count
	SizeDistribution  map[string]int `json:"size_distribution"`            // size range -> count (lines)
	TokenDistribution map[string]int `json:"token_distribution"`           // token range -> count
	SeverityBreakdown map[string]int `json:"severity_breakdown"`           // severity -> count (small/medium/large/huge)
	CategoryBreakdown map[string]int `json:"category_breakdown,omitempty"` // category -> count (function/test/struct/etc)
	PriorityBreakdown map[string]int `json:"priority_breakdown,omitempty"` // priority -> count (critical/high/medium/low)

	// Actionability metrics
	ActionableGroups    int `json:"actionable_groups,omitempty"`     // Groups that can be refactored
	NonActionableGroups int `json:"non_actionable_groups,omitempty"` // Groups that are idiomatic boilerplate

	// Test vs production separation
	TestCloneGroups       int `json:"test_clone_groups,omitempty"`       // Groups found in test files
	ProductionCloneGroups int `json:"production_clone_groups,omitempty"` // Groups found in production files

	// Top impactful clones to fix
	TopClones []TopCloneGroup `json:"top_clones,omitempty"` // Highest priority actionable clones

	// Filter metrics (NEW)
	FilesFiltered   int            `json:"files_filtered,omitempty"`   // Total files filtered out
	FilterBreakdown map[string]int `json:"filter_breakdown,omitempty"` // Reason -> count (e.g., "templ" -> 12)

	// Detection mode
	DetectionMode     string `json:"detection_mode,omitempty"`             // "semantic" or "structural"
	DetectionModeDesc string `json:"detection_mode_description,omitempty"` // Human-readable description

	// Metadata
	DetectionMethods  string `json:"detection_methods"`  // Comma-separated detection methods used
	SemanticDetection bool   `json:"semantic_detection"` // Whether semantic-aware detection was enabled
}
