package domain

import (
	"fmt"
	"path/filepath"
	"testing"
)

// BenchmarkStringPool_RealisticDuplLoad simulates actual dupl usage patterns
func BenchmarkStringPool_RealisticDuplLoad(b *testing.B) {
	pool := NewStringInternPool(5000)
	
	// Simulate scanning a large Go project
	files := generateRealisticFileList(1000) // 1000 .go files
	fragments := generateRealisticCodeFragments()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	// Simulate the actual dupl workflow
	for i := 0; i < b.N; i++ {
		// Scan phase: Intern filenames (writes, but limited)
		if i < len(files) {
			_ = pool.Intern(files[i])
		}
		
		// Analysis phase: Repeated lookups of same strings (reads, dominant)
		for j := 0; j < 10; j++ { // Each file accessed multiple times
			fileIdx := (i + j) % len(files)
			id := pool.Intern(files[fileIdx]) // Fast path: already interned
			_ = pool.Lookup(id)
		}
		
		// Fragment storage: Intern code fragments (writes, but limited)
		fragIdx := i % len(fragments)
		_ = pool.Intern(fragments[fragIdx])
	}
}

// BenchmarkStringPool_GlobalPoolRealistic uses the actual global pool
func BenchmarkStringPool_GlobalPoolRealistic(b *testing.B) {
	// Pre-populate with common patterns
	commonFiles := []string{
		"main.go", "go.mod", "go.sum", "README.md",
		"LICENSE", ".gitignore", "Dockerfile",
		"cmd/server/main.go", "cmd/client/main.go",
		"pkg/config/config.go", "pkg/utils/helpers.go",
	}
	
	for _, f := range commonFiles {
		_ = GlobalPool().Intern(f)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// Simulate multiple goroutines analyzing files
			file := fmt.Sprintf("pkg/service%d/service.go", i%100)
			
			// 90% of the time we're looking up (reading) existing strings
			if i%10 < 9 {
				id := GlobalPool().Intern(file) // Will be fast path after first time
				_ = GlobalPool().Lookup(id)
			} else {
				// 10% of the time we encounter new files (writing)
				newFile := fmt.Sprintf("internal/package%d/new_file.go", i%50)
				_ = GlobalPool().Intern(newFile)
			}
			i++
		}
	})
}

// Test the actual NodeToClone usage pattern
func BenchmarkStringPool_NodeToClonePattern(b *testing.B) {
	files := generateRealisticFileList(500)
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		// This is what actually happens in NodeToClone
		clone := Clone{}
		clone.SetFilename(files[i%len(files)])
		clone.SetFragment("func test() { return nil }")
		clone.SetHash("a1b2c3d4e5f6")
		
		// These are the most common operations
		_ = clone.FilenameString()
		_ = clone.FragmentString()
		_ = clone.HashString()
	}
}

// Generate realistic file names for a Go project
func generateRealisticFileList(count int) []string {
	files := make([]string, count)
	
	for i := 0; i < count; i++ {
		// Simulate realistic project structure
		packages := []string{
			"cmd", "pkg", "internal", "api", "web", "scripts",
		}
		pkg := packages[i%len(packages)]
		
		subpackages := []string{
			"server", "client", "config", "utils", "model", "service",
			"repository", "handler", "middleware", "test",
		}
		subpkg := subpackages[(i/len(packages))%len(subpackages)]
		
		fileTypes := []string{
			"main.go", "config.go", "handler.go", "service.go",
			"repository.go", "model.go", "test.go", "helpers.go",
		}
		fileType := fileTypes[i%len(fileTypes)]
		
		files[i] = filepath.Join(pkg, subpkg, fileType)
	}
	
	// Add lots of duplicate patterns (realistic!)
	dupCount := count / 4
	for i := 0; i < dupCount; i++ {
		files[i] = files[dupCount+i%100] // Reuse filenames
	}
	
	return files
}

// Generate realistic code fragments
func generateRealisticCodeFragments() []string {
	return []string{
		`func main() { fmt.Println("Hello, World!") }`,
		`func (s *Service) Process(data interface{}) error { return nil }`,
		`type Config struct { Debug bool; Port int }`,
		`if err != nil { return fmt.Errorf("failed: %w", err) }`,
		`for i := 0; i < len(items); i++ { process(items[i]) }`,
		`go func() { defer wg.Done(); worker() }()`,
		`select { case <-ctx.Done(): return ctx.Err(); default: }`,
		`type Repository interface { Save(entity interface{}) error }`,
		`const ( StatusActive = "active"; StatusInactive = "inactive" )`,
		`var logger = log.New(os.Stdout, "[app] ", log.LstdFlags)`,
	}
}

// BenchmarkStringPool_ContentionProfile is for manual profiling
func BenchmarkStringPool_ContentionProfile(b *testing.B) {
	// Run this with: go test -bench=BenchmarkStringPool_ContentionProfile -cpuprofile=cpu.prof -mutexprofile=mutex.prof
	
	pool := NewStringInternPool(5000)
	
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			// 70% reads, 30% writes like realistic workload
			if i%10 < 7 {
				_ = pool.Lookup(1) // Read operation
			} else {
				_ = pool.Intern(fmt.Sprintf("file%d.go", i%1000))
			}
			i++
		}
	})
}

// Analyze results from profiler
// go tool pprof -http=:6060 mutex.prof
// Look for: sync.(*RWMutex).Lock and sync.(*RWMutex).RLock
