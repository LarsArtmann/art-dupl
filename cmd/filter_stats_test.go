package cmd

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/gogenfilter/v3"
)

func TestNewFilterStats(t *testing.T) {
	reasons := []gogenfilter.FilterReason{gogenfilter.ReasonTempl, gogenfilter.ReasonSQLC}

	s := NewFilterStats(reasons)

	if s == nil {
		t.Fatal("NewFilterStats returned nil")
	}

	if got := s.TotalFiltered(); got != 0 {
		t.Errorf("TotalFiltered() = %d, want 0 on fresh stats", got)
	}

	if got := s.FilteredBy(string(gogenfilter.ReasonTempl)); got != 0 {
		t.Errorf("FilteredBy() = %d, want 0 on fresh stats", got)
	}

	if got := s.Reasons(); !reflect.DeepEqual(got, reasons) {
		t.Errorf("Reasons() = %v, want %v", got, reasons)
	}
}

func TestRecordDefaultsToGogenfilterSource(t *testing.T) {
	s := NewFilterStats(nil)

	s.Record(gogenfilter.FilterResult{
		Filtered: true,
		Reason:   gogenfilter.ReasonTempl,
		Path:     "a_templ.go",
	})

	got := s.SourceBreakdown()
	if got[string(FilterSourceGogenfilter)] != 1 {
		t.Errorf("Record() should attribute to %s, got %v", FilterSourceGogenfilter, got)
	}
}

func TestRecordDoesNotCountUnfiltered(t *testing.T) {
	s := NewFilterStats(nil)

	s.Record(gogenfilter.FilterResult{Filtered: false, Reason: gogenfilter.ReasonNotFiltered})
	s.RecordWithSource(gogenfilter.FilterResult{Filtered: false}, FilterSource("test-source"))

	if got := s.TotalFiltered(); got != 0 {
		t.Errorf("TotalFiltered() = %d, want 0 when no results are filtered", got)
	}

	if got := s.Breakdown(); len(got) != 0 {
		t.Errorf("Breakdown() = %v, want empty map", got)
	}
}

func TestRecordWithSourceAndBreakdown(t *testing.T) {
	s := NewFilterStats(nil)

	s.RecordWithSource(gogenfilter.FilterResult{
		Filtered: true, Reason: gogenfilter.ReasonTempl, Path: "a_templ.go",
	}, FilterSourceGogenfilter)
	s.RecordWithSource(gogenfilter.FilterResult{
		Filtered: true, Reason: gogenfilter.ReasonTempl, Path: "b_templ.go",
	}, FilterSourceGogenfilter)
	s.RecordWithSource(gogenfilter.FilterResult{
		Filtered: true, Reason: gogenfilter.ReasonSQLC, Path: "c_sqlc.go",
	}, FilterSource("test-source"))
	// Unfiltered results must not move any counter.
	s.RecordWithSource(gogenfilter.FilterResult{
		Filtered: false, Reason: gogenfilter.ReasonNotFiltered, Path: "regular.go",
	}, FilterSourceGogenfilter)

	if got, want := s.TotalFiltered(), 3; got != want {
		t.Errorf("TotalFiltered() = %d, want %d", got, want)
	}

	if got, want := s.FilteredBy(string(gogenfilter.ReasonTempl)), 2; got != want {
		t.Errorf("FilteredBy(templ) = %d, want %d", got, want)
	}

	if got, want := s.FilteredBy(string(gogenfilter.ReasonSQLC)), 1; got != want {
		t.Errorf("FilteredBy(sqlc) = %d, want %d", got, want)
	}

	wantBreakdown := map[string]int{
		string(gogenfilter.ReasonTempl): 2,
		string(gogenfilter.ReasonSQLC):  1,
	}
	if got := s.Breakdown(); !reflect.DeepEqual(got, wantBreakdown) {
		t.Errorf("Breakdown() = %v, want %v", got, wantBreakdown)
	}
}

func TestBreakdownAndSourceBreakdownAreDefensiveCopies(t *testing.T) {
	s := NewFilterStats(nil)

	s.RecordWithSource(gogenfilter.FilterResult{
		Filtered: true, Reason: gogenfilter.ReasonTempl,
	}, FilterSourceGogenfilter)

	original := s.Breakdown()
	original["mutated"] = 999

	again := s.Breakdown()
	if _, leaked := again["mutated"]; leaked {
		t.Error("Breakdown() did not return a defensive copy; internal map was mutated")
	}

	originalSource := s.SourceBreakdown()
	originalSource["mutated"] = 999

	againSource := s.SourceBreakdown()
	if _, leaked := againSource["mutated"]; leaked {
		t.Error("SourceBreakdown() did not return a defensive copy; internal map was mutated")
	}
}

func TestFilterStatsNilReceiverSafe(t *testing.T) {
	var s *FilterStats

	if got := s.TotalFiltered(); got != 0 {
		t.Errorf("nil TotalFiltered() = %d, want 0", got)
	}

	if got := s.FilteredBy(string(gogenfilter.ReasonTempl)); got != 0 {
		t.Errorf("nil FilteredBy() = %d, want 0", got)
	}

	if got := s.Breakdown(); got != nil {
		t.Errorf("nil Breakdown() = %v, want nil", got)
	}

	if got := s.SourceBreakdown(); got != nil {
		t.Errorf("nil SourceBreakdown() = %v, want nil", got)
	}

	if got := s.Reasons(); got != nil {
		t.Errorf("nil Reasons() = %v, want nil", got)
	}

	// Recording on a nil receiver must be a no-op, not a panic.
	s.Record(gogenfilter.FilterResult{Filtered: true, Reason: gogenfilter.ReasonTempl})
	s.RecordWithSource(gogenfilter.FilterResult{Filtered: true}, FilterSource("test-source"))
}

func TestNewTrackedFilterStatsNilFilter(t *testing.T) {
	if got := newTrackedFilterStats(nil, &config.Config{ExcludePatterns: []string{"*_gen.go"}}); got != nil {
		t.Error("newTrackedFilterStats(nil, cfg) should return nil")
	}
}

func TestNewTrackedFilterStatsTracksUserPatterns(t *testing.T) {
	filter, err := gogenfilter.NewFilter()
	if err != nil {
		t.Fatalf("gogenfilter.NewFilter() failed: %v", err)
	}

	s := newTrackedFilterStats(filter, &config.Config{ExcludePatterns: []string{"nomatch*.go"}})
	if s == nil {
		t.Fatal("newTrackedFilterStats returned nil for a live filter")
	}

	var buf bytes.Buffer
	s.WarnUnmatchedExcludePatterns(&buf)

	if got := buf.String(); !strings.Contains(got, "nomatch*.go") {
		t.Errorf("warning output %q should contain the unmatched pattern", got)
	}
}

func TestExcludePatternMethodsNilReceiverSafe(t *testing.T) {
	var s *FilterStats

	// All three new methods must be no-ops on a nil receiver, not panics.
	s.TrackExcludePatterns([]string{"*_gen.go"})
	s.recordPatternCandidate("a_gen.go")

	var buf bytes.Buffer
	s.WarnUnmatchedExcludePatterns(&buf)

	if got := buf.String(); got != "" {
		t.Errorf("nil WarnUnmatchedExcludePatterns wrote %q, want no output", got)
	}
}

func TestTrackExcludePatternsEmptyIsNoOp(t *testing.T) {
	s := NewFilterStats(nil)

	s.TrackExcludePatterns(nil)
	s.recordPatternCandidate("a_gen.go")

	var buf bytes.Buffer
	s.WarnUnmatchedExcludePatterns(&buf)

	if got := buf.String(); got != "" {
		t.Errorf("warning output %q, want none when no patterns are tracked", got)
	}
}

func TestWarnUnmatchedExcludePatternsOnlyWarnsUnmatched(t *testing.T) {
	s := NewFilterStats(nil)
	s.TrackExcludePatterns([]string{"*_gen.go", "regex\\.go"})

	// Glob semantics: basename-only patterns match at any depth.
	s.recordPatternCandidate("pkg/mock/a_gen.go")
	// A candidate that matches nothing (patterns above stay unmatched for it).
	s.recordPatternCandidate("main.go")

	var buf bytes.Buffer
	s.WarnUnmatchedExcludePatterns(&buf)

	got := buf.String()
	if strings.Contains(got, "*_gen.go") {
		t.Errorf("warning output should not contain the matched pattern:\n%s", got)
	}

	if !strings.Contains(got, `regex\\.go`) {
		t.Errorf("warning output should contain the unmatched pattern:\n%s", got)
	}

	warnings := strings.Count(got, "warning:")
	if want := 1; warnings != want {
		t.Errorf("warning count = %d, want %d:\n%s", warnings, want, got)
	}
}
