package job

import (
	"context"
	"maps"
	"sync"
	"time"
)

// Pipeline stage names recorded via stage timing. Stages marked "active" sum
// the time each goroutine spends doing that stage's work, so with parallel
// workers a stage total can exceed its wall-clock share — each record is
// goroutine-local active time, not elapsed time.
const (
	// StageCrawl covers file discovery, filtering, and feeding the file
	// channel. Recorded as the wall time of the feed goroutine, which
	// includes back-pressure waits while downstream stages consume.
	StageCrawl = "crawl"
	// StageParse covers AST construction for one file (parse only,
	// serialization is measured separately).
	StageParse = "parse"
	// StageSerialize covers syntax.SerializeWithMaxChildren per file.
	StageSerialize = "serialize"
	// StageTreeBuild covers suffix tree updates per file batch, excluding
	// time blocked waiting on the parser channel.
	StageTreeBuild = "tree-build"
)

// Phase names recorded by the cmd layer. Phases are wall-clock spans and may
// overlap each other where the pipeline streams concurrently.
const (
	// PhaseIngest spans crawl + parse + serialize + tree build (all
	// overlapped) from analysis start until the tree is complete.
	PhaseIngest = "ingest"
	// PhaseSearch spans suffix tree search from spawn until the match
	// channel closes; includes consumer back-pressure while printing.
	PhaseSearch = "search"
	// PhasePrint spans actionability evaluation and output writing.
	PhasePrint = "print"
	// PhaseTotal spans the whole analysis.
	PhaseTotal = "total"
)

// StageTiming accumulates per-stage durations for one analysis run. It is
// carried through the analysis context (WithStageTiming) so pipeline
// goroutines can record without signature changes, and is safe for
// concurrent use.
type StageTiming struct {
	mu     sync.Mutex
	stages map[string]time.Duration
}

// NewStageTiming returns an empty stage timing collector.
func NewStageTiming() *StageTiming {
	return &StageTiming{stages: make(map[string]time.Duration)}
}

type stageTimingContextKey struct{}

// WithStageTiming returns a context whose pipeline stages record into st.
// A nil st disables recording, keeping instrumentation zero-cost.
func WithStageTiming(ctx context.Context, st *StageTiming) context.Context {
	return context.WithValue(ctx, stageTimingContextKey{}, st)
}

// StageTimingFrom returns the collector in ctx, or nil when timing is off.
func StageTimingFrom(ctx context.Context) *StageTiming {
	st, _ := ctx.Value(stageTimingContextKey{}).(*StageTiming)

	return st
}

// RecordStage adds d to the named stage when ctx carries a StageTiming.
func RecordStage(ctx context.Context, stage string, d time.Duration) {
	if st := StageTimingFrom(ctx); st != nil {
		st.Record(stage, d)
	}
}

// Record adds d to the named stage.
func (s *StageTiming) Record(stage string, d time.Duration) {
	s.mu.Lock()
	s.stages[stage] += d
	s.mu.Unlock()
}

// Snapshot returns a copy of the accumulated per-stage durations.
func (s *StageTiming) Snapshot() map[string]time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make(map[string]time.Duration, len(s.stages))
	maps.Copy(out, s.stages)

	return out
}
