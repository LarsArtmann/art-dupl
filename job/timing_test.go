package job

import (
	"context"
	"testing"
	"time"
)

func TestStageTimingRecordAndSnapshot(t *testing.T) {
	st := NewStageTiming()

	st.Record(StageParse, 100*time.Millisecond)
	st.Record(StageParse, 50*time.Millisecond)
	st.Record(StageTreeBuild, 10*time.Millisecond)

	snap := st.Snapshot()

	if got := snap[StageParse]; got != 150*time.Millisecond {
		t.Errorf("parse = %v, want 150ms", got)
	}

	if got := snap[StageTreeBuild]; got != 10*time.Millisecond {
		t.Errorf("tree-build = %v, want 10ms", got)
	}
}

func TestStageTimingSnapshotIsCopy(t *testing.T) {
	st := NewStageTiming()
	st.Record(StageSerialize, time.Second)

	snap := st.Snapshot()
	snap[StageSerialize] = 0

	if got := st.Snapshot()[StageSerialize]; got != time.Second {
		t.Errorf("mutating the snapshot leaked into the collector: %v", got)
	}
}

func TestRecordStageWithoutTimingIsNoop(t *testing.T) {
	ctx := context.Background()

	if st := StageTimingFrom(ctx); st != nil {
		t.Fatalf("StageTimingFrom(plain ctx) = %v, want nil", st)
	}

	RecordStage(ctx, StageParse, time.Second)
}

func TestWithStageTimingCarriesCollector(t *testing.T) {
	st := NewStageTiming()
	ctx := WithStageTiming(context.Background(), st)

	if got := StageTimingFrom(ctx); got != st {
		t.Fatalf("StageTimingFrom(ctx) = %v, want the injected collector", got)
	}

	RecordStage(ctx, PhaseSearch, 5*time.Millisecond)

	if got := st.Snapshot()[PhaseSearch]; got != 5*time.Millisecond {
		t.Errorf("search = %v, want 5ms", got)
	}
}
