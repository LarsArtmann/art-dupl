package job

import (
	"context"
	"testing"
	"time"

	"github.com/LarsArtmann/art-dupl/cache"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
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

// helloWorldFile is the standard test file content used across incremental parser tests.
const helloWorldFile = `package main

func main() {
	println("hello")
}`

// writeHelloWorldFile creates a test file with the standard hello-world content.
func writeHelloWorldFile(t *testing.T, setup *testutil.TestFileSetup) {
	t.Helper()

	if err := setup.CreateTestFile("test.go", helloWorldFile); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
}

// singleFileChannel creates a buffered channel pre-populated with a single file path.
func singleFileChannel(setup *testutil.TestFileSetup) chan string {
	fchan := make(chan string, 1)
	fchan <- setup.GetFilePath("test.go")

	close(fchan)

	return fchan
}

func TestIncrementalParserBasic(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	writeHelloWorldFile(t, setup)

	parser := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0, 0)
	ctx := t.Context()

	fchan := singleFileChannel(setup)

	schan, statsChan := parser.ParseIncremental(ctx, fchan)

	waitForParsedNodes(t, schan, "ParseIncremental timed out")

	stats := <-statsChan
	testutil.AssertIncrementalStats(t, stats.FilesCount, stats.CacheMisses, 1, 1)
}

func TestIncrementalParserCacheHit(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	writeHelloWorldFile(t, setup)

	parser := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0, 0)
	ctx := t.Context()

	fchan := singleFileChannel(setup)

	schan, _ := parser.ParseIncremental(ctx, fchan)
	for range schan {
		// Drain channel
	}

	fchan2 := singleFileChannel(setup)

	schan2, statsChan := parser.ParseIncremental(ctx, fchan2)

	for seq := range schan2 {
		if len(seq) == 0 {
			t.Error("Expected some parsed nodes from cache")
		}
	}

	stats := <-statsChan
	testutil.AssertFieldValue(t, stats.CacheHits, 1, "CacheHits")
}

func TestIncrementalParserClearCache(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	writeHelloWorldFile(t, setup)

	parser := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0, 0)
	ctx := t.Context()

	fchan := singleFileChannel(setup)

	schan, _ := parser.ParseIncremental(ctx, fchan)
	for range schan {
		// Drain channel
	}

	parserWithClear := NewIncrementalParser(cacheDir, true, golang.DetectionModeSemantic, 0, 0, 0)

	fchan2 := singleFileChannel(setup)

	schan2, statsChan := parserWithClear.ParseIncremental(ctx, fchan2)

	for range schan2 {
		// Drain channel
	}

	stats := <-statsChan
	testutil.AssertFieldValue(t, stats.CacheMisses, 1, "CacheMisses")
}

func TestIncrementalParserContextCancellation(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	parser := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0, 0)
	cancel, fchan := cancelledContextAndChannel(t.Context())

	schan, _ := parser.ParseIncremental(t.Context(), fchan)

	for range schan {
	}

	_ = cancel
}

func TestIncrementalParserNonexistentFile(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	parser := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0, 0)
	ctx := t.Context()

	fchan := make(chan string, 1)
	fchan <- "nonexistent.go"

	close(fchan)

	schan, statsChan := parser.ParseIncremental(ctx, fchan)

	waitForChannelOrTimeout(t, schan, 5*time.Second, "Should handle nonexistent file gracefully")

	stats := <-statsChan
	testutil.AssertFieldValue(t, stats.FilesCount, 1, "FilesCount")
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

	parser := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0, 0)
	ctx := t.Context()

	fchan := make(chan string, 2)
	fchan <- setup.GetFilePath("file1.go")

	fchan <- setup.GetFilePath("file2.go")

	close(fchan)

	schan, statsChan := parser.ParseIncremental(ctx, fchan)

	count := 0

	for nodeSeq := range schan {
		if len(nodeSeq) == 0 {
			t.Error("Expected parsed nodes")
		}

		count++
	}

	if count != 2 {
		t.Errorf("Expected 2 nodeSequences, got %d", count)
	}

	stats := <-statsChan
	testutil.AssertIncrementalStats(t, stats.FilesCount, stats.CacheMisses, 2, 2)
}

func TestIncrementalParserGetCacheStats(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	writeHelloWorldFile(t, setup)

	parser := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0, 0)
	ctx := t.Context()

	fchan := singleFileChannel(setup)

	schan, _ := parser.ParseIncremental(ctx, fchan)
	for range schan {
	}

	stats := parser.GetCacheStats()
	testutil.AssertFieldValue(t, stats.Size, 1, "cache entry")
}

// TestIncrementalParserCacheKeyIsolationMode verifies that switching detection
// modes (semantic vs exact) does NOT reuse cached ASTs from a prior mode.
// Before the fix, the cache key was SHA-256(content) only, so an exact-mode run
// would incorrectly reuse a semantic-mode cached AST.
func TestIncrementalParserCacheKeyIsolationMode(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	writeHelloWorldFile(t, setup)

	ctx := t.Context()

	// First parse with semantic mode → cache miss.
	semanticParser := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0, 0)
	fchan1 := singleFileChannel(setup)

	schan1, statsChan1 := semanticParser.ParseIncremental(ctx, fchan1)
	for range schan1 {
	}

	stats1 := <-statsChan1
	testutil.AssertFieldValue(t, stats1.CacheMisses, 1, "CacheMisses on first semantic parse")

	// Same file with exact mode → should ALSO be a cache miss (different key).
	exactParser := NewIncrementalParser(cacheDir, false, golang.DetectionModeExact, 0, 0, 0)
	fchan2 := singleFileChannel(setup)

	schan2, statsChan2 := exactParser.ParseIncremental(ctx, fchan2)
	for range schan2 {
	}

	stats2 := <-statsChan2
	testutil.AssertFieldValue(t, stats2.CacheMisses, 1, "CacheMisses on exact parse (key isolation)")
	testutil.AssertFieldValue(t, stats2.CacheHits, 0, "CacheHits should be 0 for different mode")

	// But repeating exact mode → should be a cache hit.
	fchan3 := singleFileChannel(setup)

	schan3, statsChan3 := exactParser.ParseIncremental(ctx, fchan3)
	for range schan3 {
	}

	stats3 := <-statsChan3
	testutil.AssertFieldValue(t, stats3.CacheHits, 1, "CacheHits on repeat exact parse")
}

// TestIncrementalParserCacheKeyIsolationMaxChildren verifies that different
// maxChildren values produce separate cache entries.
func TestIncrementalParserCacheKeyIsolationMaxChildren(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	writeHelloWorldFile(t, setup)

	ctx := t.Context()

	// Parse with maxChildren=0 → cache miss.
	parser1 := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0, 0)
	fchan1 := singleFileChannel(setup)

	schan1, statsChan1 := parser1.ParseIncremental(ctx, fchan1)
	for range schan1 {
	}

	stats1 := <-statsChan1
	testutil.AssertFieldValue(t, stats1.CacheMisses, 1, "CacheMisses on first parse")

	// Same file with maxChildren=5 → should be cache miss (different key).
	parser2 := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 5, 0, 0)
	fchan2 := singleFileChannel(setup)

	schan2, statsChan2 := parser2.ParseIncremental(ctx, fchan2)
	for range schan2 {
	}

	stats2 := <-statsChan2
	testutil.AssertFieldValue(t, stats2.CacheMisses, 1, "CacheMisses on different maxChildren")
	testutil.AssertFieldValue(t, stats2.CacheHits, 0, "CacheHits should be 0 for different maxChildren")
}

// TestIncrementalParserCacheKeyFormat verifies that cacheKey() produces the
// expected "mode:maxChildren:typeAwareTag" params format and that each
// parameter axis independently affects the key.
func TestIncrementalParserCacheKeyFormat(t *testing.T) {
	t.Parallel()

	content := []byte("package main\n\nfunc f() { println(\"hi\") }\n")

	t.Run("params format is mode:maxChildren:typeAwareTag", func(t *testing.T) {
		t.Parallel()

		ip := NewIncrementalParser(t.TempDir(), false, golang.DetectionModeSemantic, 10, 0, 0)
		ip.typeAwareTag = "ta"

		// The key should be KeyWithParams(content, "semantic:10:ta").
		expectedParams := "semantic:10:ta"
		expectedKey := cache.KeyWithParams(content, expectedParams)

		gotKey := ip.cacheKey(content)
		if gotKey != expectedKey {
			t.Errorf("cacheKey mismatch:\n  got:    %s\n  expect: %s", gotKey, expectedKey)
		}
	})

	t.Run("different modes produce different keys", func(t *testing.T) {
		t.Parallel()

		semantic := NewIncrementalParser(t.TempDir(), false, golang.DetectionModeSemantic, 0, 0, 0)
		exact := NewIncrementalParser(t.TempDir(), false, golang.DetectionModeExact, 0, 0, 0)
		structural := NewIncrementalParser(t.TempDir(), false, golang.DetectionModeStructural, 0, 0, 0)

		keyS := semantic.cacheKey(content)
		keyE := exact.cacheKey(content)
		keySt := structural.cacheKey(content)

		if keyS == keyE {
			t.Error("expected different keys for semantic vs exact mode")
		}

		if keyS == keySt {
			t.Error("expected different keys for semantic vs structural mode")
		}

		if keyE == keySt {
			t.Error("expected different keys for exact vs structural mode")
		}
	})

	t.Run("different maxChildren produce different keys", func(t *testing.T) {
		t.Parallel()

		p0 := NewIncrementalParser(t.TempDir(), false, golang.DetectionModeSemantic, 0, 0, 0)
		p5 := NewIncrementalParser(t.TempDir(), false, golang.DetectionModeSemantic, 5, 0, 0)
		p100 := NewIncrementalParser(t.TempDir(), false, golang.DetectionModeSemantic, 100, 0, 0)

		key0 := p0.cacheKey(content)
		key5 := p5.cacheKey(content)
		key100 := p100.cacheKey(content)

		if key0 == key5 {
			t.Error("expected different keys for maxChildren 0 vs 5")
		}

		if key5 == key100 {
			t.Error("expected different keys for maxChildren 5 vs 100")
		}
	})

	t.Run("different typeAwareTags produce different keys", func(t *testing.T) {
		t.Parallel()

		none := NewIncrementalParser(t.TempDir(), false, golang.DetectionModeSemantic, 0, 0, 0)
		none.typeAwareTag = ""

		ta := NewIncrementalParser(t.TempDir(), false, golang.DetectionModeSemantic, 0, 0, 0)
		ta.typeAwareTag = "ta"

		sg := NewIncrementalParser(t.TempDir(), false, golang.DetectionModeSemantic, 0, 0, 0)
		sg.typeAwareTag = "sg"

		keyNone := none.cacheKey(content)
		keyTA := ta.cacheKey(content)
		keySG := sg.cacheKey(content)

		if keyNone == keyTA {
			t.Error("expected different keys for typeAwareTag '' vs 'ta'")
		}

		if keyNone == keySG {
			t.Error("expected different keys for typeAwareTag '' vs 'sg'")
		}

		if keyTA == keySG {
			t.Error("expected different keys for typeAwareTag 'ta' vs 'sg'")
		}
	})

	t.Run("same params produce same key", func(t *testing.T) {
		t.Parallel()

		p1 := NewIncrementalParser(t.TempDir(), false, golang.DetectionModeExact, 7, 0, 0)
		p1.typeAwareTag = "sg"

		p2 := NewIncrementalParser(t.TempDir(), false, golang.DetectionModeExact, 7, 0, 0)
		p2.typeAwareTag = "sg"

		if p1.cacheKey(content) != p2.cacheKey(content) {
			t.Error("expected identical keys for identical params")
		}
	})
}
