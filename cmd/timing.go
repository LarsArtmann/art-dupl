package cmd

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"sort"
	"time"

	"github.com/LarsArtmann/art-dupl/job"
)

// runTiming wraps one analysis run's stage collector with wall-clock phase
// recording and allocation counters. A nil *runTiming is inert: every method
// tolerates nil so call sites never need a timing-enabled special case.
type runTiming struct {
	stages    *job.StageTiming
	startedAt time.Time
	memBefore runtime.MemStats
}

// startRunTiming begins timing a run and returns the context the analysis
// pipeline should use so its stages record into the collector.
func startRunTiming(ctx context.Context) (*runTiming, context.Context) {
	var before runtime.MemStats
	runtime.ReadMemStats(&before)

	rt := &runTiming{stages: job.NewStageTiming(), startedAt: time.Now(), memBefore: before}

	return rt, job.WithStageTiming(ctx, rt.stages)
}

// recordPrintPhase records the print phase wall time.
func (rt *runTiming) recordPrintPhase(ctx context.Context, start time.Time) {
	if rt != nil {
		job.RecordStage(ctx, job.PhasePrint, time.Since(start))
	}
}

// finish records the total wall time and prints the stage report plus
// allocation deltas to w.
func (rt *runTiming) finish(ctx context.Context, w io.Writer) {
	if rt == nil {
		return
	}

	job.RecordStage(ctx, job.PhaseTotal, time.Since(rt.startedAt))

	var after runtime.MemStats
	runtime.ReadMemStats(&after)

	printStageTimingReport(w, rt.stages.Snapshot(), &rt.memBefore, &after)
}

// timingStageOrder is the fixed presentation order of stage rows; stages not
// present in the snapshot are skipped.
var timingStageOrder = []string{ //nolint:gochecknoglobals // static presentation table
	job.StageCrawl,
	job.StageParse,
	job.StageSerialize,
	job.StageTreeBuild,
	job.PhaseIngest,
	job.PhaseSearch,
	job.PhasePrint,
	job.PhaseTotal,
}

// timingStageNotes explains what each row measures so the numbers cannot be
// misread (active time vs wall time, overlap).
var timingStageNotes = map[string]string{ //nolint:gochecknoglobals // static presentation table
	job.StageCrawl:     "wall; includes back-pressure while feeding parsers",
	job.StageParse:     "active, summed across workers",
	job.StageSerialize: "active",
	job.StageTreeBuild: "active, excludes channel waits",
	job.PhaseIngest:    "wall; crawl+parse+serialize+build overlapped",
	job.PhaseSearch:    "wall; includes print back-pressure",
	job.PhasePrint:     "wall",
	job.PhaseTotal:     "wall",
}

// printStageTimingReport writes the per-stage table and allocation deltas.
// Active stages sum goroutine-local time and can exceed the total when
// workers run in parallel; phase rows are wall-clock and can overlap where
// the pipeline streams concurrently.
func printStageTimingReport(w io.Writer, stages map[string]time.Duration, before, after *runtime.MemStats) {
	_, _ = fmt.Fprintln(w, "⏱️  Stage timing:")
	_, _ = fmt.Fprintln(w, "     (parse/serialize/tree-build sum active time across workers; phases are wall and may overlap)")

	for _, stage := range timingStageOrder {
		d, ok := stages[stage]
		if !ok {
			continue
		}

		_, _ = fmt.Fprintf(w, "  %-12s %10s   %s\n", stage, d.Round(time.Millisecond), timingStageNotes[stage])
	}

	objects := after.Mallocs - before.Mallocs
	bytes := after.TotalAlloc - before.TotalAlloc
	gcCycles := after.NumGC - before.NumGC

	var pause time.Duration
	if after.NumGC >= before.NumGC && len(after.PauseNs) > 0 {
		// PauseTotalNs is cumulative across the process lifetime, so the
		// delta is valid even when NumGC wraps uint32.
		pause = time.Duration(after.PauseTotalNs - before.PauseTotalNs)
	}

	_, _ = fmt.Fprintf(
		w,
		"  allocations: %d objects, %s total, %d GC cycles, %s GC pause\n",
		objects, humanBytes(bytes), gcCycles, pause.Round(time.Microsecond),
	)
}

// humanBytes renders a byte count with an SI suffix for compact display.
func humanBytes(b uint64) string {
	const unit = 1024

	if b < unit {
		return fmt.Sprintf("%d B", b)
	}

	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit && exp < 5; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

// sortedStageNames returns snapshot keys in deterministic order for tests.
func sortedStageNames(stages map[string]time.Duration) []string {
	names := make([]string, 0, len(stages))
	for name := range stages {
		names = append(names, name)
	}
	sort.Strings(names)

	return names
}
