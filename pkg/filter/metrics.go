// Package filter provides smart filtering of auto-generated Go code.
package filter

import (
	"maps"
	"sync"
)

// Metrics tracks filter statistics for analysis and debugging.
type Metrics struct {
	mu sync.RWMutex

	// TotalFilesChecked is the total number of files the filter evaluated
	TotalFilesChecked int

	// FilteredByReason tracks how many files were filtered for each reason
	FilteredByReason map[FilterReason]int

	// FilteredFiles maps reason to list of files filtered for that reason
	FilteredFiles map[FilterReason][]string
}

// NewMetrics creates a new filter metrics tracker.
func NewMetrics() *Metrics {
	return &Metrics{
		FilteredByReason: make(map[FilterReason]int),
		FilteredFiles:    make(map[FilterReason][]string),
	}
}

// Record records that a file was filtered for a given reason.
// OPTIMIZATION: Handle nil metrics for better performance.
func (m *Metrics) Record(filePath string, reason FilterReason) {
	if m == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalFilesChecked++

	if reason != ReasonNotFiltered {
		m.FilteredByReason[reason]++
		m.FilteredFiles[reason] = append(m.FilteredFiles[reason], filePath)
	}
}

// RecordChecked records that a file was checked but not filtered.
// Convenience method for cleaner code.
func (m *Metrics) RecordChecked(filePath string) {
	m.Record(filePath, ReasonNotFiltered)
}

// RecordFiltered records that a file was filtered for a given reason.
// Convenience method for cleaner code.
func (m *Metrics) RecordFiltered(filePath string, reason FilterReason) {
	m.Record(filePath, reason)
}

// GetStats returns the current filter statistics.
func (m *Metrics) GetStats() FilterStats {
	if m == nil {
		return FilterStats{}
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	// Create copies of the maps to avoid concurrent access issues
	filteredByReason := maps.Clone(m.FilteredByReason)

	return FilterStats{
		TotalFilesChecked: m.TotalFilesChecked,
		FilteredByReason:  filteredByReason,
	}
}

// FilterStats represents a snapshot of filter statistics.
type FilterStats struct {
	TotalFilesChecked int
	FilteredByReason  map[FilterReason]int
}

// TotalFiltered returns the total number of filtered files.
func (fs FilterStats) TotalFiltered() int {
	total := 0
	for reason, count := range fs.FilteredByReason {
		if reason != ReasonNotFiltered {
			total += count
		}
	}
	return total
}
