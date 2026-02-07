package domain

import (
	"fmt"
	"sync"
	"testing"
)

// BenchmarkStringPool_SingleGoroutine measures baseline performance.
func BenchmarkStringPool_SingleGoroutine(b *testing.B) {
	pool := NewStringInternPool(1000)

	b.ResetTimer()
	b.ReportAllocs()

	for i := range b.N {
		id := pool.Intern(fmt.Sprintf("file%d.go", i%100))
		_ = pool.Lookup(id)
	}
}

// BenchmarkStringPool_ConcurrentReads measures read contention.
func BenchmarkStringPool_ConcurrentReads(b *testing.B) {
	pool := NewStringInternPool(100)

	// Pre-populate the pool
	ids := make([]StringID, 100)
	for i := range 100 {
		ids[i] = pool.Intern(fmt.Sprintf("file%d.go", i))
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			id := ids[i%100]
			_ = pool.Lookup(id)
			i++
		}
	})
}

// BenchmarkStringPool_ConcurrentWrites measures write contention.
func BenchmarkStringPool_ConcurrentWrites(b *testing.B) {
	pool := NewStringInternPool(10000)

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_ = pool.Intern(fmt.Sprintf("goroutine%d_file%d.go", b.N, i))
			i++
		}
	})
}

// BenchmarkStringPool_ConcurrentMixed measures realistic mixed workload.
func BenchmarkStringPool_ConcurrentMixed(b *testing.B) {
	pool := NewStringInternPool(1000)

	// Pre-populate some strings
	commonIDs := make([]StringID, 100)
	for i := range 100 {
		commonIDs[i] = pool.Intern(fmt.Sprintf("common%d.go", i))
	}

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// 70% reads, 30% writes (realistic pattern)
			if i%10 < 7 {
				_ = pool.Lookup(commonIDs[i%100]) // Read
			} else {
				_ = pool.Intern(fmt.Sprintf("unique%d_%d.go", b.N, i)) // Write
			}
			i++
		}
	})
}

// BenchmarkStringPool_GlobalPool measures actual global pool usage.
func BenchmarkStringPool_GlobalPool(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			id := GlobalPool().Intern(fmt.Sprintf("test%d.go", i%50))
			_ = GlobalPool().Lookup(id)
			i++
		}
	})
}

// BenchmarkStringPool_HighContention worst-case scenario.
func BenchmarkStringPool_HighContention(b *testing.B) {
	pool := NewStringInternPool(10)

	// Only 10 unique strings, all goroutines hammer the same locks
	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			id := pool.Intern("same_file.go")
			_ = pool.Lookup(id)
		}
	})
}

// BenchmarkStringPool_Stats measures stats collection overhead.
func BenchmarkStringPool_Stats(b *testing.B) {
	pool := NewStringInternPool(1000)

	// Pre-populate
	for i := range 1000 {
		_ = pool.Intern(fmt.Sprintf("file%d.go", i))
	}

	b.ResetTimer()
	b.ReportAllocs()

	for range b.N {
		_ = pool.Stats()
	}
}

// Measure lock contention with a custom stress test.
func TestStringPool_LockContention(t *testing.T) {
	pool := NewStringInternPool(10000)

	var wg sync.WaitGroup
	const goroutines = 100
	const operations = 10000

	// Track blocking time (requires runtime.MutexProfile)
	// This test is for manual profiling: go test -run TestStringPool_LockContention -cpuprofile=cpu.prof

	for g := range goroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := range operations {
				// Mix of reads and writes
				if i%2 == 0 {
					_ = pool.Intern(fmt.Sprintf("goroutine%d_file%d.go", id, i))
				} else {
					_ = pool.Stats()
				}
			}
		}(g)
	}

	wg.Wait()
}
