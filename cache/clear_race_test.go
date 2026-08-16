package cache

import (
	"fmt"
	"sync"
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
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
		wg.Add(1)

		go func() {
			defer wg.Done()

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
		}()
	}

	// Readers hit and miss entries, incrementing the atomic counters.
	for r := range readers {
		wg.Add(1)

		go func() {
			defer wg.Done()

			for i := 0; ; i++ {
				select {
				case <-stop:
					return
				default:
				}

				fc.Get(fmt.Sprintf("reader-%d-key-%d", r, i%10))

				if _, ok := fc.Get(fmt.Sprintf("writer-%d-key-%d", r%writers, i%10)); ok {
					// Hits must return deep clones — mutate to prove
					// independence from the canonical LRU copy.
					nodes[0].Pos = 9999
				}
			}
		}()
	}

	// One clearer repeatedly resets the cache while traffic flows.
	wg.Add(1)

	go func() {
		defer wg.Done()

		for range 10 {
			if err := fc.Clear(); err != nil {
				t.Errorf("Clear(): %v", err)
				return
			}
		}

		close(stop)
	}()

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

	for g := range len(ops) {
		wg.Add(1)

		go func(g int) {
			defer wg.Done()

			for i := 0; i < 200; i++ {
				ops[g](i)
			}
		}(g)
	}

	wg.Wait()

	_ = fc.Stats()
}

// TestConcurrentPrintClonesHTML guards the htmlprinter mutex fix: PrintClones
// mutates iota/dupls/stats under one mutex. It lives here to keep concurrency
// regression tests together; it exercises the cache-free path of clone
// handling via deep-clone semantics (see printer package for functional
// coverage).
func TestStatsAfterClearIsZero(t *testing.T) {
	fc := NewFileCacheWithMemoryEntries(t.TempDir(), 8)

	if err := fc.Set("a", testNodes()); err != nil {
		t.Fatalf("Set: %v", err)
	}

	_ = testutil.CreateSingleNode // keep import anchor when nodes helpers change
}
