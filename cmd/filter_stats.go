package cmd

import (
	"maps"
	"sync"

	"github.com/LarsArtmann/gogenfilter/v3"
)

// FilterSource identifies which filtering mechanism caught a generated file.
type FilterSource string

const (
	// FilterSourceGogenfilter: caught by gogenfilter's standard filename-gated
	// or content-based checks.
	FilterSourceGogenfilter FilterSource = "gogenfilter"

	// FilterSourceDefenseInDepth: caught by the content-based defense-in-depth
	// check (filterExcludedGenerated) that catches generated files lacking the
	// expected filename suffix (_templ.go, _sqlc.go, *.pb.go).
	FilterSourceDefenseInDepth FilterSource = "defense-in-depth"
)

// FilterStats holds aggregated filter statistics.
// Replaces the removed gogenfilter.FilterStats type — stats aggregation
// is now the caller's responsibility per gogenfilter's API redesign.
type FilterStats struct {
	mu       sync.Mutex
	total    int
	byReason map[string]int
	bySource map[string]int
	reasons  []gogenfilter.FilterReason
}

// NewFilterStats creates a new FilterStats with the given filter reasons.
func NewFilterStats(reasons []gogenfilter.FilterReason) *FilterStats {
	return &FilterStats{
		byReason: make(map[string]int),
		bySource: make(map[string]int),
		reasons:  reasons,
	}
}

// withLock runs fn while holding s.mu and returns its result. When s is nil
// it returns zeroValue, so every read method is nil-receiver-safe. Map fields
// are read inside the closure (not as a bare argument) because evaluating
// s.byReason/s.bySource directly would panic on a nil receiver before the nil
// check runs.
func withLock[T any](s *FilterStats, zeroValue T, fn func() T) T {
	if s == nil {
		return zeroValue
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return fn()
}

// Record records a filter result from gogenfilter.
func (s *FilterStats) Record(result gogenfilter.FilterResult) {
	s.RecordWithSource(result, FilterSourceGogenfilter)
}

// RecordWithSource records a filter result with its source (gogenfilter or
// defense-in-depth), allowing stats output to distinguish how each file was caught.
func (s *FilterStats) RecordWithSource(result gogenfilter.FilterResult, source FilterSource) {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if result.Filtered {
		s.total++
		s.byReason[string(result.Reason)]++
		s.bySource[string(source)]++
	}
}

// TotalFiltered returns the total number of filtered files.
func (s *FilterStats) TotalFiltered() int {
	return withLock(s, 0, func() int { return s.total })
}

// FilteredBy returns the number of files filtered by the given reason.
func (s *FilterStats) FilteredBy(reason string) int {
	return withLock(s, 0, func() int { return s.byReason[reason] })
}

// Breakdown returns a defensive copy of the per-reason breakdown.
func (s *FilterStats) Breakdown() map[string]int {
	return withLock(s, nil, func() map[string]int {
		result := make(map[string]int, len(s.byReason))
		maps.Copy(result, s.byReason)
		return result
	})
}

// SourceBreakdown returns a defensive copy of the per-source breakdown,
// distinguishing files caught by gogenfilter vs the defense-in-depth content check.
func (s *FilterStats) SourceBreakdown() map[string]int {
	return withLock(s, nil, func() map[string]int {
		result := make(map[string]int, len(s.bySource))
		maps.Copy(result, s.bySource)
		return result
	})
}

// Reasons returns the filter reasons this tracker was configured with.
func (s *FilterStats) Reasons() []gogenfilter.FilterReason {
	if s == nil {
		return nil
	}

	return s.reasons
}
