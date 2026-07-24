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

type Printer interface {
	PrintHeader() error
	PrintClones(group domain.ProcessedCloneGroup, sortBy ...config.SortCriteria) error
	PrintFooter() error
}

// StatsConfig holds configuration for statistics output.
// Used by StatsPrinter.ApplyStatsConfig to set all stats metadata in one call.
type StatsConfig struct {
	Format              config.OutputFormat
	FilesCount          int
	DetectionMethods    string
	SemanticDetection   bool
	Timestamp           string
	AnalysisDuration    time.Duration
	TotalEstimatedLines int
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
