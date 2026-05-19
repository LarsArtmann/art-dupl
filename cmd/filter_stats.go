package cmd

import (
	"sync"

	"github.com/LarsArtmann/gogenfilter/v3"
)

// FilterStats holds aggregated filter statistics.
// Replaces the removed gogenfilter.FilterStats type — stats aggregation
// is now the caller's responsibility per gogenfilter's API redesign.
type FilterStats struct {
	mu         sync.Mutex
	total      int
	byReason   map[string]int
	reasons    []gogenfilter.FilterReason
}

// NewFilterStats creates a new FilterStats with the given filter reasons.
func NewFilterStats(reasons []gogenfilter.FilterReason) *FilterStats {
	return &FilterStats{
		byReason: make(map[string]int),
		reasons:  reasons,
	}
}

// Record records a filter result.
func (s *FilterStats) Record(result gogenfilter.FilterResult) {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if result.Filtered {
		s.total++
		s.byReason[string(result.Reason)]++
	}
}

// TotalFiltered returns the total number of filtered files.
func (s *FilterStats) TotalFiltered() int {
	if s == nil {
		return 0
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.total
}

// FilteredBy returns the number of files filtered by the given reason.
func (s *FilterStats) FilteredBy(reason string) int {
	if s == nil {
		return 0
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return s.byReason[reason]
}

// Breakdown returns a copy of the per-reason breakdown.
func (s *FilterStats) Breakdown() map[string]int {
	if s == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	result := make(map[string]int, len(s.byReason))
	for k, v := range s.byReason {
		result[k] = v
	}

	return result
}

// Reasons returns the filter reasons this tracker was configured with.
func (s *FilterStats) Reasons() []gogenfilter.FilterReason {
	if s == nil {
		return nil
	}

	return s.reasons
}
