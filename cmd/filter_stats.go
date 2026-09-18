package cmd

import (
	"fmt"
	"io"
	"maps"
	"sync"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/gogenfilter/v3"
)

// FilterSource identifies which filtering mechanism caught a generated file.
type FilterSource string

const (
	// FilterSourceGogenfilter: caught by gogenfilter's standard filename-gated
	// or content-based checks.
	FilterSourceGogenfilter FilterSource = "gogenfilter"
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

	// excludePatterns/excludeMatched track user-supplied --exclude-pattern
	// values for zero-match detection. Patterns that never match are likely
	// misconfigurations (e.g. a regex where a glob is expected) and are
	// reported by WarnUnmatchedExcludePatterns.
	excludePatterns []string
	excludeMatched  map[string]bool

	// includePatterns/includeMatched is the symmetric tracking for
	// user-supplied --include-pattern values, reported by
	// WarnUnmatchedIncludePatterns: an include that matches nothing silently
	// leaves the targeted files filtered out.
	includePatterns []string
	includeMatched  map[string]bool
}

// NewFilterStats creates a new FilterStats with the given filter reasons.
func NewFilterStats(reasons []gogenfilter.FilterReason) *FilterStats {
	return &FilterStats{
		byReason: make(map[string]int),
		bySource: make(map[string]int),
		reasons:  reasons,
	}
}

// newTrackedFilterStats builds the filter-stats recorder and registers the
// user's --exclude-pattern values for zero-match detection. Returns nil when
// no filter is configured (all methods stay nil-receiver-safe).
func newTrackedFilterStats(filterParam *gogenfilter.Filter, cfg *config.Config) *FilterStats {
	if filterParam == nil {
		return nil
	}

	filterStats := NewFilterStats(filterParam.FilterReasons())
	filterStats.TrackExcludePatterns(cfg.ExcludePatterns)
	filterStats.TrackIncludePatterns(cfg.IncludePatterns)

	return filterStats
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

// RecordWithSource records a filter result with its source, allowing stats
// output to distinguish how each file was caught.
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
// showing how many files were caught by each filtering mechanism.
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

// TrackExcludePatterns registers user-supplied --exclude-pattern values for
// zero-match detection. Only explicit user patterns belong here; patterns
// derived from other flags (--ignore-tests) are intentional and would only
// add warning noise.
func (s *FilterStats) TrackExcludePatterns(patterns []string) { //art-dupl:accept deliberate exclude/include symmetry — see TrackIncludePatterns
	if s == nil || len(patterns) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.excludePatterns = append(s.excludePatterns, patterns...)

	if s.excludeMatched == nil {
		s.excludeMatched = make(map[string]bool)
	}
}

// TrackIncludePatterns registers user-supplied --include-pattern values for
// zero-match detection, mirroring TrackExcludePatterns.
func (s *FilterStats) TrackIncludePatterns(patterns []string) { //art-dupl:accept deliberate exclude/include symmetry
	if s == nil || len(patterns) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.includePatterns = append(s.includePatterns, patterns...)

	if s.includeMatched == nil {
		s.includeMatched = make(map[string]bool)
	}
}

// recordPatternCandidate counts which tracked user-supplied exclude/include
// patterns match a candidate path seen during the crawl. Uses
// gogenfilter.MatchPattern so the counting semantics are identical to the
// actual filtering (gogenfilter applies MatchPattern to both lists).
func (s *FilterStats) recordPatternCandidate(path string) {
	if s == nil || (len(s.excludePatterns) == 0 && len(s.includePatterns) == 0) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, pattern := range s.excludePatterns {
		if !s.excludeMatched[pattern] && gogenfilter.MatchPattern(path, pattern) {
			s.excludeMatched[pattern] = true
		}
	}

	for _, pattern := range s.includePatterns {
		if !s.includeMatched[pattern] && gogenfilter.MatchPattern(path, pattern) {
			s.includeMatched[pattern] = true
		}
	}
}

// WarnUnmatchedExcludePatterns prints a warning for every tracked
// --exclude-pattern that matched no candidate file. Called once after the
// crawl completes.
func (s *FilterStats) WarnUnmatchedExcludePatterns(stderr io.Writer) { //art-dupl:accept deliberate exclude/include symmetry
	unmatched := withLock(s, nil, func() []string {
		var result []string

		for _, pattern := range s.excludePatterns {
			if !s.excludeMatched[pattern] {
				result = append(result, pattern)
			}
		}

		return result
	})

	for _, pattern := range unmatched {
		fmt.Fprintf(
			stderr,
			"warning: --exclude-pattern %q matched no files — patterns are globs (e.g. \"**/gen/*.go\", \"*_mock.go\"), not regular expressions\n",
			pattern,
		)
	}
}

// WarnUnmatchedIncludePatterns prints a warning for every tracked
// --include-pattern that matched no candidate file. Called once after the
// crawl completes. An include pattern that matches nothing silently leaves
// the targeted files filtered out — the misconfiguration trap mirrored from
// --exclude-pattern in the opposite direction.
func (s *FilterStats) WarnUnmatchedIncludePatterns(stderr io.Writer) { //art-dupl:accept deliberate exclude/include symmetry
	unmatched := withLock(s, nil, func() []string {
		var result []string

		for _, pattern := range s.includePatterns {
			if !s.includeMatched[pattern] {
				result = append(result, pattern)
			}
		}

		return result
	})

	for _, pattern := range unmatched {
		fmt.Fprintf(
			stderr,
			"warning: --include-pattern %q matched no files — patterns are globs (e.g. \"**/vendor/**\", \"internal/**/*.go\"), not regular expressions\n",
			pattern,
		)
	}
}
