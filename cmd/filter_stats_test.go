package cmd

import (
	"reflect"
	"testing"

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

	if got[string(FilterSourceDefenseInDepth)] != 0 {
		t.Errorf("Record() should not attribute to %s, got %v", FilterSourceDefenseInDepth, got)
	}
}

func TestRecordDoesNotCountUnfiltered(t *testing.T) {
	s := NewFilterStats(nil)

	s.Record(gogenfilter.FilterResult{Filtered: false, Reason: gogenfilter.ReasonNotFiltered})
	s.RecordWithSource(gogenfilter.FilterResult{Filtered: false}, FilterSourceDefenseInDepth)

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
	}, FilterSourceDefenseInDepth)
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

// TestSourceBreakdown is the dedicated regression test for the per-source
// tracking path (FilterSourceGogenfilter vs FilterSourceDefenseInDepth),
// which previously had no direct coverage.
func TestSourceBreakdown(t *testing.T) {
	s := NewFilterStats(nil)

	s.RecordWithSource(gogenfilter.FilterResult{
		Filtered: true, Reason: gogenfilter.ReasonTempl,
	}, FilterSourceGogenfilter)
	s.RecordWithSource(gogenfilter.FilterResult{
		Filtered: true, Reason: gogenfilter.ReasonSQLC,
	}, FilterSourceGogenfilter)
	s.RecordWithSource(gogenfilter.FilterResult{
		Filtered: true, Reason: gogenfilter.ReasonTempl,
	}, FilterSourceDefenseInDepth)
	s.RecordWithSource(gogenfilter.FilterResult{
		Filtered: true, Reason: gogenfilter.ReasonProtobuf,
	}, FilterSourceDefenseInDepth)
	s.RecordWithSource(gogenfilter.FilterResult{
		Filtered: true, Reason: gogenfilter.ReasonProtobuf,
	}, FilterSourceDefenseInDepth)

	want := map[string]int{
		string(FilterSourceGogenfilter):     2,
		string(FilterSourceDefenseInDepth): 3,
	}
	if got := s.SourceBreakdown(); !reflect.DeepEqual(got, want) {
		t.Errorf("SourceBreakdown() = %v, want %v", got, want)
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
	s.RecordWithSource(gogenfilter.FilterResult{Filtered: true}, FilterSourceDefenseInDepth)
}
