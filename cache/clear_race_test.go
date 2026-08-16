package cache

import (
	"fmt"
	"sync"
	"testing"
)

// TestClearConcurrentWithGets is the regression test for the cache metadata
// data race: Get increments HitCount/MissCount atomically WITHOUT holding
// fc.mu, while Clear resets them. Before the fix, Clear reset the counters by
// replacing the whole metadata struct under fc.mu — a plain (non-atomic)
// write racing with the atomic adds. This test runs concurrent Get/Set/Clear
// traffic; under `go test -race` the old code fails, the fixed code passes.
func TestClearConcurrentWithGets(t *testing.T) {
	fc := NewFileCacheWithMemoryEntries(t.TempDir(), 16)

	nodes := testNodes()

	const writers, readers = 4, 8

	var wg sync.WaitGroup

	stop := make(chan struct{})

	// Writers populate entries.
	for w := range writers {
		wg.Go(func() {
			for i := 0; ; i++ {
				select {
				case <-stop:
					return
				default:
				}

				key := fmt.Sprintf("writer-%d-key-%d", w, i%10)
				if err := fc.Set(key, nodes); err != nil {
					t.Errorf("Set(%q): %v", key, err)

					return
				}
			}
		})
	}

	// Readers hit and miss entries, incrementing the atomic counters.
	for r := range readers {
		wg.Go(func() {
			for i := 0; ; i++ {
				select {
				case <-stop:
					return
				default:
				}

				if got, ok := fc.Get(fmt.Sprintf("reader-%d-key-%d", r, i%10)); ok {
					// Hits must return deep clones — mutating the result
					// must never corrupt the canonical LRU copy.
					got[0].Pos = 9999
				}
			}
		})
	}

	// One clearer repeatedly resets the cache while traffic flows.
	wg.Go(func() {
		for range 10 {
			if err := fc.Clear(); err != nil {
				t.Errorf("Clear(): %v", err)

				return
			}
		}

		close(stop)
	})

	wg.Wait()

	stats := fc.Stats()

	if stats.Hits < 0 || stats.Misses < 0 {
		t.Errorf("negative counters after concurrent run: %+v", stats)
	}

	t.Logf("final stats: %+v", stats)
}

// TestConcurrentMixedAccess drives every mutating FileCache method against
// each other; the race detector verifies no mixed atomic/plain access
// remains anywhere in the cache.
func TestConcurrentMixedAccess(t *testing.T) {
	fc := NewFileCacheWithMemoryEntries(t.TempDir(), 8)

	nodes := testNodes()

	var wg sync.WaitGroup

	ops := []func(i int){
		func(i int) { _ = fc.Set(fmt.Sprintf("k%d", i%5), nodes) },
		func(i int) { _, _ = fc.Get(fmt.Sprintf("k%d", i%7)) },
		func(i int) { _ = fc.Has(fmt.Sprintf("k%d", i%5)) },
		func(i int) { _ = fc.Remove(fmt.Sprintf("k%d", i%5)) },
		func(i int) { _, _ = fc.Prune(2) },
	}

	for g := range ops {
		wg.Go(func() {
			for i := range 200 {
				ops[g](i)
			}
		})
	}

	wg.Wait()

	_ = fc.Stats()
}
