package job

import (
	"context"
	"testing"
	"time"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax"
)

func waitForParsedNodes(
	t *testing.T,
	schan chan []*syntax.Node,
	timeoutMsg string,
	emptyMsg ...string,
) {
	t.Helper()

	emptyError := "Expected some parsed nodes"
	if len(emptyMsg) > 0 {
		emptyError = emptyMsg[0]
	}

	select {
	case seq := <-schan:
		if len(seq) == 0 {
			t.Error(emptyError)
		}
	case <-time.After(5 * time.Second):
		t.Error(timeoutMsg)
	}
}

// cancelledContextAndChannel creates a cancelled context and a buffered channel.
// The channel will be closed after a short delay. This is useful for testing
// that parser functions handle cancellation gracefully.
func cancelledContextAndChannel(ctx context.Context) (context.CancelFunc, chan string) {
	_, cancel := context.WithCancel(ctx)

	fchan := make(chan string, 1)

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
		close(fchan)
	}()

	return cancel, fchan
}

func TestIncrementalParserBasic(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	//nolint:goconst // Intentional duplicate for independent test cases
	content := `package main

func main() {
	println("hello")
}`

	err := setup.CreateTestFile("test.go", content)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	parser := NewIncrementalParser(cacheDir, false, true)
	ctx := t.Context()

	fchan := make(chan string, 1)
	fchan <- setup.GetFilePath("test.go")

	close(fchan)

	schan, statsChan := parser.ParseIncremental(ctx, fchan)

	waitForParsedNodes(t, schan, "ParseIncremental timed out")

	stats := <-statsChan
	testutil.AssertIncrementalStats(t, stats.FilesCount, stats.CacheMisses, 1, 1)
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

	parser := NewIncrementalParser(cacheDir, false, true)
	ctx := t.Context()

	fchan := make(chan string, 1)
	fchan <- setup.GetFilePath("test.go")

	close(fchan)

	schan, _ := parser.ParseIncremental(ctx, fchan)
	for range schan {
		// Drain channel
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

	parser := NewIncrementalParser(cacheDir, false, true)
	ctx := t.Context()

	fchan := make(chan string, 1)
	fchan <- setup.GetFilePath("test.go")

	close(fchan)

	schan, _ := parser.ParseIncremental(ctx, fchan)
	for range schan {
		// Drain channel
	}

	parserWithClear := NewIncrementalParser(cacheDir, true, true)

	fchan2 := make(chan string, 1)
	fchan2 <- setup.GetFilePath("test.go")

	close(fchan2)

	schan2, statsChan := parserWithClear.ParseIncremental(ctx, fchan2)

	for range schan2 {
		// Drain channel
	}

	stats := <-statsChan
	if stats.CacheMisses != 1 {
		t.Errorf("Expected CacheMisses=1 (after clear), got %d", stats.CacheMisses)
	}
}

func TestIncrementalParserContextCancellation(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	parser := NewIncrementalParser(cacheDir, false, true)
	cancel, fchan := cancelledContextAndChannel(t.Context())

	schan, _ := parser.ParseIncremental(t.Context(), fchan)

	for range schan {
	}
	_ = cancel
}

func TestIncrementalParserNonexistentFile(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	parser := NewIncrementalParser(cacheDir, false, true)
	ctx := t.Context()

	fchan := make(chan string, 1)
	fchan <- "nonexistent.go"

	close(fchan)

	schan, statsChan := parser.ParseIncremental(ctx, fchan)

	waitForChannelOrTimeout(t, schan, 5*time.Second, "Should handle nonexistent file gracefully")

	stats := <-statsChan
	if stats.FilesCount != 1 {
		t.Errorf("Expected FilesCount=1, got %d", stats.FilesCount)
	}
}

func waitForChannelOrTimeout[T any](
	t *testing.T,
	ch <-chan T,
	timeout time.Duration,
	timeoutMsg string,
) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(timeout):
		t.Error(timeoutMsg)
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

	parser := NewIncrementalParser(cacheDir, false, true)
	ctx := t.Context()

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
	testutil.AssertIncrementalStats(t, stats.FilesCount, stats.CacheMisses, 2, 2)
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

	parser := NewIncrementalParser(cacheDir, false, true)
	ctx := t.Context()

	fchan := make(chan string, 1)
	fchan <- setup.GetFilePath("test.go")

	close(fchan)

	schan, _ := parser.ParseIncremental(ctx, fchan)
	for range schan {
	}

	stats := parser.GetCacheStats()
	if stats.Size != 1 {
		t.Errorf("Expected 1 cache entry, got %d", stats.Size)
	}
}
