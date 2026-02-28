package job

import (
	"context"
	"testing"
	"time"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

func TestIncrementalParserBasic(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	content := `package main

func main() {
	println("hello")
}`

	err := setup.CreateTestFile("test.go", content)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewIncrementalParser(cacheDir, false)
	ctx := context.Background()

	fchan := make(chan string, 1)
	fchan <- setup.GetFilePath("test.go")
	close(fchan)

	schan, statsChan := parser.ParseIncremental(ctx, fchan)

	select {
	case seq := <-schan:
		if len(seq) == 0 {
			t.Error("Expected some parsed nodes")
		}
	case <-time.After(5 * time.Second):
		t.Error("ParseIncremental timed out")
	}

	stats := <-statsChan
	if stats.FilesCount != 1 {
		t.Errorf("Expected FilesCount=1, got %d", stats.FilesCount)
	}

	if stats.CacheMisses != 1 {
		t.Errorf("Expected CacheMisses=1 (first run), got %d", stats.CacheMisses)
	}
}

func TestIncrementalParserCacheHit(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	content := `package main

func main() {
	println("hello")
}`

	err := setup.CreateTestFile("test.go", content)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewIncrementalParser(cacheDir, false)
	ctx := context.Background()

	fchan := make(chan string, 1)
	fchan <- setup.GetFilePath("test.go")
	close(fchan)

	schan, _ := parser.ParseIncremental(ctx, fchan)
	for range schan {
	}

	fchan2 := make(chan string, 1)
	fchan2 <- setup.GetFilePath("test.go")
	close(fchan2)

	schan2, statsChan := parser.ParseIncremental(ctx, fchan2)

	for seq := range schan2 {
		if len(seq) == 0 {
			t.Error("Expected some parsed nodes from cache")
		}
	}

	stats := <-statsChan
	if stats.CacheHits != 1 {
		t.Errorf("Expected CacheHits=1 (second run), got %d", stats.CacheHits)
	}
}

func TestIncrementalParserClearCache(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	content := `package main

func main() {
	println("hello")
}`

	err := setup.CreateTestFile("test.go", content)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewIncrementalParser(cacheDir, false)
	ctx := context.Background()

	fchan := make(chan string, 1)
	fchan <- setup.GetFilePath("test.go")
	close(fchan)

	schan, _ := parser.ParseIncremental(ctx, fchan)
	for range schan {
	}

	parserWithClear := NewIncrementalParser(cacheDir, true)

	fchan2 := make(chan string, 1)
	fchan2 <- setup.GetFilePath("test.go")
	close(fchan2)

	schan2, statsChan := parserWithClear.ParseIncremental(ctx, fchan2)

	for range schan2 {
	}

	stats := <-statsChan
	if stats.CacheMisses != 1 {
		t.Errorf("Expected CacheMisses=1 (after clear), got %d", stats.CacheMisses)
	}
}

func TestIncrementalParserContextCancellation(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	parser := NewIncrementalParser(cacheDir, false)
	ctx, cancel := context.WithCancel(context.Background())

	fchan := make(chan string, 1)

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
		close(fchan)
	}()

	schan, _ := parser.ParseIncremental(ctx, fchan)

	for range schan {
	}
}

func TestIncrementalParserNonexistentFile(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	parser := NewIncrementalParser(cacheDir, false)
	ctx := context.Background()

	fchan := make(chan string, 1)
	fchan <- "nonexistent.go"
	close(fchan)

	schan, statsChan := parser.ParseIncremental(ctx, fchan)

	select {
	case <-schan:
	case <-time.After(5 * time.Second):
		t.Error("Should handle nonexistent file gracefully")
	}

	stats := <-statsChan
	if stats.FilesCount != 1 {
		t.Errorf("Expected FilesCount=1, got %d", stats.FilesCount)
	}
}

func TestIncrementalParserMultipleFiles(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	files := map[string]string{
		"file1.go": `package main

func function1() {
	println("test1")
}`,
		"file2.go": `package main

func function2() {
	println("test2")
}`,
	}

	err := setup.CreateTestFiles(files)
	if err != nil {
		t.Fatalf("Failed to create test files: %v", err)
	}

	parser := NewIncrementalParser(cacheDir, false)
	ctx := context.Background()

	fchan := make(chan string, 2)
	fchan <- setup.GetFilePath("file1.go")
	fchan <- setup.GetFilePath("file2.go")
	close(fchan)

	schan, statsChan := parser.ParseIncremental(ctx, fchan)

	count := 0
	for seq := range schan {
		if len(seq) == 0 {
			t.Error("Expected parsed nodes")
		}
		count++
	}

	if count != 2 {
		t.Errorf("Expected 2 sequences, got %d", count)
	}

	stats := <-statsChan
	if stats.FilesCount != 2 {
		t.Errorf("Expected FilesCount=2, got %d", stats.FilesCount)
	}

	if stats.CacheMisses != 2 {
		t.Errorf("Expected CacheMisses=2, got %d", stats.CacheMisses)
	}
}

func TestIncrementalParserGetCacheStats(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	content := `package main

func main() {
	println("hello")
}`

	err := setup.CreateTestFile("test.go", content)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewIncrementalParser(cacheDir, false)
	ctx := context.Background()

	fchan := make(chan string, 1)
	fchan <- setup.GetFilePath("test.go")
	close(fchan)

	schan, _ := parser.ParseIncremental(ctx, fchan)
	for range schan {
	}

	stats := parser.GetCacheStats()
	if stats.Entries != 1 {
		t.Errorf("Expected 1 cache entry, got %d", stats.Entries)
	}
}
