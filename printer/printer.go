package printer

import (
	"time"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/syntax"
)

type ReadFile func(filename string) ([]byte, error)

// HashSetter is an optional interface for printers that support hash metadata.
type HashSetter interface {
	SetHash(hash string)
}

type Printer interface {
	PrintHeader() error
	PrintClones(dups [][]*syntax.Node, sortBy ...config.SortCriteria) error
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
	FilesFiltered       int
	FilterBreakdown     map[string]int
}

// StatsPrinter extends Printer interface with stats configuration.
type StatsPrinter interface {
	Printer
	ApplyStatsConfig(config StatsConfig)
	GetStatsData() *StatsData
}
