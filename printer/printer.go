package printer

import (
	"time"

	"github.com/LarsArtmann/art-dupl/syntax"
)

type ReadFile func(filename string) ([]byte, error)

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
	SetFormat(format Format)
	SetTimestamp(timestamp string)
	SetAnalysisDuration(duration time.Duration)
	SetTotalEstimatedLines(lines int)
	GetStatsData() any
}
