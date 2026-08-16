package cmd

import (
	"bytes"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/LarsArtmann/art-dupl/job"
)

// TestTimingFlagPrintsStageReport verifies the --timing flag produces a
// stderr report covering every pipeline stage and phase plus allocation
// counters.
func TestTimingFlagPrintsStageReport(t *testing.T) {
	tmpDir := t.TempDir()

	createDuplicateTestFiles(t, tmpDir)

	cmd := NewRootCommand()
	AddFlags(cmd)
	cmd.SetArgs([]string{flagKeyThreshold, "10", "--timing", "--quiet", tmpDir})

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("runCmd() error = %v", err)
	}

	out := buf.String()

	for _, want := range []string{
		"Stage timing:",
		"crawl",
		"parse",
		"serialize",
		"tree-build",
		"ingest",
		"search",
		"print",
		"total",
		"allocations:",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("timing report missing %q\nreport:\n%s", want, out)
		}
	}
}

// TestTimingDisabledByDefault verifies no timing output appears without the
// flag, so default runs stay clean.
func TestTimingDisabledByDefault(t *testing.T) {
	tmpDir := t.TempDir()

	createDuplicateTestFiles(t, tmpDir)

	cmd := NewRootCommand()
	AddFlags(cmd)
	cmd.SetArgs([]string{flagKeyThreshold, "10", "--quiet", tmpDir})

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)

	if err := cmd.Execute(); err != nil {
		t.Fatalf("runCmd() error = %v", err)
	}

	if strings.Contains(buf.String(), "Stage timing") {
		t.Errorf("timing report printed without --timing:\n%s", buf.String())
	}
}

func TestHumanBytes(t *testing.T) {
	cases := []struct {
		in   uint64
		want string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{2048, "2.0 KB"},
		{1536 * 1024, "1.5 MB"},
		{3 << 30, "3.0 GB"},
	}

	for _, tc := range cases {
		if got := humanBytes(tc.in); got != tc.want {
			t.Errorf("humanBytes(%d) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestPrintStageTimingReportRowOrder(t *testing.T) {
	stages := map[string]time.Duration{
		job.StageCrawl:     10 * time.Millisecond,
		job.StageParse:     100 * time.Millisecond,
		job.StageSerialize: 30 * time.Millisecond,
		job.StageTreeBuild: 20 * time.Millisecond,
		job.PhaseIngest:    500 * time.Millisecond,
		job.PhaseSearch:    50 * time.Millisecond,
		job.PhasePrint:     5 * time.Millisecond,
		job.PhaseTotal:     time.Second,
	}

	var (
		before runtime.MemStats
		after  runtime.MemStats
	)

	buf := &bytes.Buffer{}
	printStageTimingReport(buf, stages, &before, &after)

	out := buf.String()

	rows := []string{
		job.StageCrawl,
		job.StageParse,
		job.StageSerialize,
		job.StageTreeBuild,
		job.PhaseIngest,
		job.PhaseSearch,
		job.PhasePrint,
		job.PhaseTotal,
	}

	lineOf := make(map[string]int, len(rows))

	for i, line := range strings.Split(out, "\n") {
		for _, stage := range rows {
			if strings.HasPrefix(line, "  "+stage+" ") {
				lineOf[stage] = i
			}
		}
	}

	for i, stage := range rows {
		if _, ok := lineOf[stage]; !ok {
			t.Fatalf("stage row %q missing from report:\n%s", stage, out)
		}

		if i > 0 && lineOf[stage] <= lineOf[rows[i-1]] {
			t.Errorf("stage %q printed before %q:\n%s", rows[i-1], stage, out)
		}
	}
}
