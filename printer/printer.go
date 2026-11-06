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

// VersionFooterSetter enables printers to stamp the effective binary
// version next to the report summary, so every pasted report names its
// build (issue #8: dev binaries carried no provenance).
type VersionFooterSetter interface {
	SetVersionFooter(version string)
}

// SuppressionStats carries counts of detected vs suppressed clone groups,
// enabling the summary to distinguish "what the detector found" from "what
// is actually actionable." This prevents the "drive to zero" failure mode
// where agents treat every detected group as harmful.
type SuppressionStats struct {
	DetectedTotal        int
	SuppressedActionable int
	SuppressedOther      int
	Shown                int
}

// SuppressionStatsSetter enables printers to receive suppression counts for
// summary display. Printers that implement this interface can show a richer
// footer separating "Detected" from "Actionable" clone groups.
type SuppressionStatsSetter interface {
	SetSuppressionStats(stats SuppressionStats)
}

// BaselineSplit is the `check` command's group accounting: groups that matched
// an entry in the baseline, groups that are new relative to it, and groups
// suppressed for other reasons (actionability, thresholds, accept directives).
// Reporting the split makes the CI verdict self-auditing: a reader can tell
// "green because nothing new was introduced" apart from "green because the
// baseline absorbed everything" without diffing the report against the
// baseline JSON (issue #12).
type BaselineSplit struct {
	InBaseline int
	New        int
	Suppressed int
}

// Total returns the number of clone groups the check run detected: the sum of
// baseline-matched, new, and otherwise suppressed groups.
func (b BaselineSplit) Total() int {
	return b.InBaseline + b.New + b.Suppressed
}

// BaselineSplitSetter enables printers to report the baseline split on the
// summary line. Only the `check` command sets it; every other run keeps its
// historical footer.
type BaselineSplitSetter interface {
	SetBaselineSplit(split BaselineSplit)
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
	Cache                 *CacheMetrics
}

// StatsPrinter extends Printer interface with stats configuration.
type StatsPrinter interface {
	Printer
	ApplyStatsConfig(config StatsConfig)
	SetFilterStats(filesFiltered int, breakdown map[string]int)
	SetFilterSourceStats(sourceBreakdown map[string]int)
	GetStatsView() *StatsView
}
