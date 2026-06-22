package job

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

// ProfileResult contains performance profiling metrics.
type ProfileResult struct {
	AllocMB      float64       // Memory allocated in MB
	TotalAllocMB float64       // Total memory allocated in MB
	SysMB        float64       // System memory in MB
	NumGC        uint32        // Number of garbage collections
	PauseTotalMS float64       // Total GC pause time in ms
	Duration     time.Duration // Total execution time
	NumGoroutine int           // Number of goroutines
	Timestamp    time.Time     // Start timestamp for duration calculation
}

// Profile captures performance metrics at a point in time.
func Profile() ProfileResult {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return ProfileResult{ //nolint:exhaustruct // Duration/Timestamp computed later by ProfileDiff
		TotalAllocMB: float64(m.TotalAlloc) / 1024 / 1024,
		SysMB:        float64(m.Sys) / 1024 / 1024,
		NumGC:        m.NumGC,
		PauseTotalMS: float64(m.PauseTotalNs) / 1000000,
		NumGoroutine: runtime.NumGoroutine(),
	}
}

// ProfileDiff calculates the difference between two profiles.
func ProfileDiff(start, end ProfileResult) ProfileResult {
	return ProfileResult{ //nolint:exhaustruct // Delta result; Timestamp meaningless for diffs
		TotalAllocMB: end.TotalAllocMB - start.TotalAllocMB,
		SysMB:        end.SysMB - start.SysMB,
		NumGC:        end.NumGC - start.NumGC,
		PauseTotalMS: end.PauseTotalMS - start.PauseTotalMS,
		Duration:     end.Duration - start.Duration,
		NumGoroutine: end.NumGoroutine - start.NumGoroutine,
	}
}

// StartProfile returns a profile with start time.
func StartProfile() ProfileResult {
	p := Profile()
	p.Timestamp = time.Now()
	p.Duration = 0

	return p
}

// EndProfile completes a profile and calculates duration.
func EndProfile(start ProfileResult) ProfileResult {
	end := Profile()
	end.Duration = time.Since(start.Timestamp)

	return end
}

// ProfileWithDuration creates a profile with a specific duration.
func ProfileWithDuration(duration time.Duration) ProfileResult {
	p := Profile()
	p.Duration = duration

	return p
}

// PrintProfileResult outputs profile metrics to stderr.
func PrintProfileResult(result ProfileResult) {
	runtime.GC() // Force GC for accurate measurement

	// Re-read after GC for final stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")
	fmt.Fprintln(os.Stderr, "                    PERFORMANCE PROFILING RESULTS")
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")
	fmt.Fprintln(os.Stderr)

	fmt.Fprintln(os.Stderr, "Execution Time:")
	fmt.Fprintf(os.Stderr, "  %s\n", result.Duration)
	fmt.Fprintln(os.Stderr)

	fmt.Fprintln(os.Stderr, "Memory Usage:")
	fmt.Fprintf(os.Stderr, "  Current Allocation: %8.2f MB\n", float64(m.Alloc)/1024/1024)
	fmt.Fprintf(os.Stderr, "  Total Allocated:    %8.2f MB\n", float64(m.TotalAlloc)/1024/1024)
	fmt.Fprintf(os.Stderr, "  System Memory:     %8.2f MB\n", float64(m.Sys)/1024/1024)
	fmt.Fprintln(os.Stderr)

	fmt.Fprintln(os.Stderr, "Garbage Collection:")
	fmt.Fprintf(os.Stderr, "  GC Cycles:          %8d\n", m.NumGC)
	fmt.Fprintf(os.Stderr, "  Total Pause Time:    %8.2f ms\n", float64(m.PauseTotalNs)/1000000)

	if m.NumGC > 0 {
		fmt.Fprintf(
			os.Stderr,
			"  Avg Pause/Cycle:    %8.2f ms\n",
			float64(m.PauseTotalNs)/float64(m.NumGC)/1000000,
		)
	}

	fmt.Fprintln(os.Stderr)

	fmt.Fprintln(os.Stderr, "Concurrency:")
	fmt.Fprintf(os.Stderr, "  Goroutines:         %8d\n", runtime.NumGoroutine())
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "═══════════════════════════════════════════════════════════")
	fmt.Fprintln(os.Stderr)
}
