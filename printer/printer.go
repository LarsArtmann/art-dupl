package printer

import (
	"time"

	"github.com/LarsArtmann/art-dupl/syntax"
)

type ReadFile func(filename string) ([]byte, error)

// HashSetter is an optional interface for printers that support hash metadata.
type HashSetter interface {
	SetHash(hash string)
}

type Printer interface {
	PrintHeader() error
	PrintClones(dups [][]*syntax.Node, sortBy ...SortBy) error // Add optional sortBy parameter
	PrintFooter() error
}

// StatsPrinter extends Printer interface with stats-specific setters.
type StatsPrinter interface {
	Printer
	SetFilesCount(count int)
	SetDetectionMethods(methods string)
	SetSemanticDetection(enabled bool)
	SetFormat(format Format)
	SetTimestamp(timestamp string)
	SetAnalysisDuration(duration time.Duration)
	SetTotalEstimatedLines(lines int)
	SetFilterStats(filesFiltered int, breakdown map[string]int)
	GetStatsData() any
}
