package cmd

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/job"
)

func TestCacheMetricsFrom(t *testing.T) {
	t.Parallel()

	if got := cacheMetricsFrom(nil); got != nil {
		t.Errorf("cacheMetricsFrom(nil) = %+v, want nil", got)
	}

	got := cacheMetricsFrom(&job.RunCacheStats{Hits: 3, Misses: 1, MemoryHits: 2, Entries: 4})
	if got == nil {
		t.Fatal("cacheMetricsFrom returned nil for non-nil stats")
	}

	if got.Hits != 3 || got.Misses != 1 || got.MemoryHits != 2 || got.Entries != 4 {
		t.Errorf("field mapping mismatch: %+v", got)
	}

	if got.HitRatePct != 75 {
		t.Errorf("HitRatePct = %v, want 75", got.HitRatePct)
	}

	empty := cacheMetricsFrom(&job.RunCacheStats{})
	if empty.HitRatePct != 0 {
		t.Errorf("HitRatePct with no lookups = %v, want 0 (no division by zero)", empty.HitRatePct)
	}
}
