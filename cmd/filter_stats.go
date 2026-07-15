package cmd

import (
	"maps"
	"sync"

	"github.com/LarsArtmann/gogenfilter/v3"
)

// FilterStats holds aggregated filter statistics.
// Replaces the removed gogenfilter.FilterStats type — stats aggregation
// is now the caller's responsibility per gogenfilter's API redesign.
type FilterStats struct {
	mu       sync.Mutex
	total    int
	byReason map[string]int
	reasons  []gogenfilter.FilterReason
}

// NewFilterStats creates a new FilterStats with the given filter reasons.
func NewFilterStats(reasons []gogenfilter.FilterReason) *FilterStats {
	return &FilterStats{
		byReason: make(map[string]int),
		reasons:  reasons,
	}
}

// withReadLock runs fn while holding s.mu and returns its result.
// Returns zeroValue when s is nil, so callers can defer to that case.
func (s *FilterStats) withReadLock(zeroValue int, fn func() int) int {
	if s == nil {
		return zeroValue
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return fn()
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
	return s.withReadLock(0, func() int { return s.total })
}

// FilteredBy returns the number of files filtered by the given reason.
func (s *FilterStats) FilteredBy(reason string) int {
	return s.withReadLock(0, func() int { return s.byReason[reason] })
}

// Breakdown returns a copy of the per-reason breakdown.
func (s *FilterStats) Breakdown() map[string]int {
	if s == nil {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	result := make(map[string]int, len(s.byReason))
	maps.Copy(result, s.byReason)

	return result
}

// Reasons returns the filter reasons this tracker was configured with.
func (s *FilterStats) Reasons() []gogenfilter.FilterReason {
	if s == nil {
		return nil
	}

	return s.reasons
}
