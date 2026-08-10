package printer

import (
	"time"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
)

// ReadFile is an alias for domain.FileReaderFunc, the canonical file-reader
// function type shared across the codebase.
type ReadFile = domain.FileReaderFunc

type HashSetter interface {
	SetHash(hash string)
}

// RichTextSetter enables enhanced text output with classification badges.
type RichTextSetter interface {
	SetRichText(enabled bool)
}

// ExplainSetter enables explanation output that describes why each clone
// group was reported (clone type, actionability, category, extractability).
type ExplainSetter interface {
	SetExplain(enabled bool)
}

// SuppressionStats carries counts of detected vs suppressed clone groups,
// enabling the summary to distinguish "what the detector found" from "what
// is actually actionable." This prevents the "drive to zero" failure mode
// where agents treat every detected group as harmful.
type SuppressionStats struct {
	DetectedTotal        int
	SuppressedActionable int
	SuppressedOther      int
	SuppressedGenerics   int
	Shown                int
}

// SuppressionStatsSetter enables printers to receive suppression counts for
// summary display. Printers that implement this interface can show a richer
// footer separating "Detected" from "Actionable" clone groups.
type SuppressionStatsSetter interface {
	SetSuppressionStats(stats SuppressionStats)
}

type Printer interface {
	PrintHeader() error
	PrintClones(group domain.ProcessedCloneGroup, sortBy ...config.SortCriteria) error
	PrintFooter() error
}

// StatsConfig holds configuration for statistics output.
// Used by StatsPrinter.ApplyStatsConfig to set all stats metadata in one call.
type StatsConfig struct {
	Format                config.OutputFormat
	FilesCount            int
	DetectionMethods      string
	SemanticDetection     bool
	Timestamp             string
	AnalysisDuration      time.Duration
	TotalEstimatedLines   int
	FilesFiltered         int
	FilterBreakdown       map[string]int
	FilterSourceBreakdown map[string]int
}

// StatsPrinter extends Printer interface with stats configuration.
type StatsPrinter interface {
	Printer
	ApplyStatsConfig(config StatsConfig)
	SetFilterStats(filesFiltered int, breakdown map[string]int)
	SetFilterSourceStats(sourceBreakdown map[string]int)
	GetStatsView() *StatsView
}
