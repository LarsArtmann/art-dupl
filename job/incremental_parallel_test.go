package job

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"sort"
	"testing"
	"time"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

// nodeSignatures returns a sorted slice of "Type:Name" signatures for every
// node in every sequence, allowing order-independent comparison of parse
// output between sequential and parallel runs.
func nodeSignatures(seqs [][]*syntax.Node) []string {
	var sigs []string

	for _, seq := range seqs {
		for _, node := range seq {
			collectNodeSigs(node, &sigs)
		}
	}

	sort.Strings(sigs)

	return sigs
}

func collectNodeSigs(n *syntax.Node, sigs *[]string) {
	if n == nil {
		return
	}

	*sigs = append(*sigs, fmt.Sprintf("%d:%s", n.Type, n.Name))

	for _, c := range n.Children {
		collectNodeSigs(c, sigs)
	}
}

// drainIncremental reads all node sequences from schan and returns them.
func drainIncremental(t *testing.T, schan chan []*syntax.Node) [][]*syntax.Node {
	t.Helper()

	var all [][]*syntax.Node

	for seq := range schan {
		all = append(all, seq)
	}

	return all
}

// filenamesFromSeqs returns the set of distinct filenames across all node trees.
func filenamesFromSeqs(seqs [][]*syntax.Node) map[string]int {
	seen := make(map[string]int)

	for _, seq := range seqs {
		for _, node := range seq {
			if node != nil && node.Filename != "" {
				seen[node.Filename]++
			}
		}
	}

	return seen
}

func TestIncrementalParallelBasic(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	writeHelloWorldFile(t, setup)

	parser := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0)
	ctx := t.Context()

	fchan := singleFileChannel(setup)

	workers := runtime.GOMAXPROCS(0)
	schan, statsChan := parser.ParseIncrementalParallel(ctx, fchan, workers)

	seqs := drainIncremental(t, schan)

	if len(seqs) != 1 {
		t.Fatalf("Expected 1 node sequence, got %d", len(seqs))
	}

	if len(seqs[0]) == 0 {
		t.Error("Expected non-empty node sequence")
	}

	stats := <-statsChan
	testutil.AssertIncrementalStats(t, stats.FilesCount, stats.CacheMisses, 1, 1)
}

func TestIncrementalParallelMultipleFiles(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	files := map[string]string{
		"alpha.go": `package main

func alpha() {
	println("alpha")
}`,
		"beta.go": `package main

func beta() {
	println("beta")
}`,
		"gamma.go": `package main

func gamma() {
	println("gamma")
}`,
	}

	err := setup.CreateTestFiles(files)
	if err != nil {
		t.Fatalf("Failed to create test files: %v", err)
	}

	parser := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0)
	ctx := t.Context()

	fchan := make(chan string, len(files))
	fchan <- setup.GetFilePath("alpha.go")
	fchan <- setup.GetFilePath("beta.go")
	fchan <- setup.GetFilePath("gamma.go")
	close(fchan)

	schan, statsChan := parser.ParseIncrementalParallel(ctx, fchan, 4)

	seqs := drainIncremental(t, schan)

	if len(seqs) != len(files) {
		t.Fatalf("Expected %d node sequences, got %d", len(files), len(seqs))
	}

	stats := <-statsChan
	testutil.AssertIncrementalStats(t, stats.FilesCount, stats.CacheMisses, len(files), len(files))
}

// TestIncrementalParallelIdenticalContent is the CRITICAL concurrency test.
// It creates many files with byte-identical content, forcing all workers to
// collide on the same content hash. This exercises:
//   - the singleflight coalescing of concurrent cache misses
//   - the shared singleflight result being deep-cloned per caller
//   - each caller stamping its OWN distinct filename on an independent copy
//
// Run with -race to catch any shared-mutation data race.
func TestIncrementalParallelIdenticalContent(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	const numFiles = 16

	filenames := make([]string, numFiles)
	for i := range numFiles {
		filenames[i] = fmt.Sprintf("dup_%02d.go", i)
	}

	content := `package main

func duplicated() {
	println("identical body")
	x := 42
	_ = x
}`

	err := setup.CreateDuplicateFiles(filenames, content)
	if err != nil {
		t.Fatalf("Failed to create duplicate files: %v", err)
	}

	parser := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0)
	ctx := t.Context()

	fchan := make(chan string, numFiles)
	for _, name := range filenames {
		fchan <- setup.GetFilePath(name)
	}
	close(fchan)

	// Workers > files to maximize concurrent collision on the shared hash.
	workers := numFiles

	schan, statsChan := parser.ParseIncrementalParallel(ctx, fchan, workers)

	seqs := drainIncremental(t, schan)

	if len(seqs) != numFiles {
		t.Fatalf("Expected %d node sequences, got %d", numFiles, len(seqs))
	}

	// Every sequence must carry a DISTINCT filename — proving each worker got
	// an independent clone rather than sharing mutated Filename pointers.
	seen := filenamesFromSeqs(seqs)

	if len(seen) != numFiles {
		t.Errorf("Expected %d distinct filenames, got %d (shared/aliased nodes?)", numFiles, len(seen))
	}

	// Each expected filename must be present. Additionally, every filename must
	// appear the SAME number of times (= nodes per serialized tree): if two
	// files shared a single cloned tree, one filename would be over-counted and
	// another missing.
	var firstCount int

	for _, name := range filenames {
		full := filepath.Join(setup.TmpDir, name)
		count, ok := seen[full]
		if !ok {
			t.Errorf("expected filename %q missing from output", full)
			continue
		}

		if firstCount == 0 {
			firstCount = count
		} else if count != firstCount {
			t.Errorf("filename %q appeared %d times, want %d (uniform node count) — possible tree sharing",
				full, count, firstCount)
		}
	}

	stats := <-statsChan
	testutil.AssertFieldValue(t, stats.FilesCount, numFiles, "FilesCount")
}

// TestIncrementalParallelMatchesSequential verifies that parallel parsing
// produces the same set of serialized nodes as sequential parsing. Only the
// ORDER may differ; the multiset of node signatures must be identical.
func TestIncrementalParallelMatchesSequential(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)

	files := map[string]string{
		"a.go": `package main

func aFunc() {
	println("a")
	for i := 0; i < 3; i++ {
		println(i)
	}
}`,
		"b.go": `package main

func bFunc(x int) int {
	if x > 0 {
		return x
	}
	return -x
}`,
		"c.go": `package main

type Thing struct {
	Name string
	Val  int
}

func (t Thing) String() string {
	return t.Name
}`,
	}

	err := setup.CreateTestFiles(files)
	if err != nil {
		t.Fatalf("Failed to create test files: %v", err)
	}

	makeChan := func() chan string {
		ch := make(chan string, len(files))
		ch <- setup.GetFilePath("a.go")
		ch <- setup.GetFilePath("b.go")
		ch <- setup.GetFilePath("c.go")
		close(ch)

		return ch
	}

	// Sequential run (fresh cache)
	seqParser := NewIncrementalParser(setup.TmpDir+"/cache_seq", false, golang.DetectionModeSemantic, 0, 0)
	seqSchan, _ := seqParser.ParseIncremental(t.Context(), makeChan())
	seqSigs := nodeSignatures(drainIncremental(t, seqSchan))

	// Parallel run (fresh, separate cache)
	parParser := NewIncrementalParser(setup.TmpDir+"/cache_par", false, golang.DetectionModeSemantic, 0, 0)
	parSchan, _ := parParser.ParseIncrementalParallel(t.Context(), makeChan(), 4)
	parSigs := nodeSignatures(drainIncremental(t, parSchan))

	if len(seqSigs) != len(parSigs) {
		t.Fatalf("Signature count mismatch: sequential=%d parallel=%d", len(seqSigs), len(parSigs))
	}

	for i := range seqSigs {
		if seqSigs[i] != parSigs[i] {
			t.Errorf("Signature mismatch at index %d:\n  sequential=%q\n  parallel  =%q",
				i, seqSigs[i], parSigs[i])

			break
		}
	}
}

// TestIncrementalParallelCacheHit verifies that a second parallel run reuses
// cached results from the first run.
func TestIncrementalParallelCacheHit(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	files := map[string]string{
		"one.go": `package main

func one() { println("one") }`,
		"two.go": `package main

func two() { println("two") }`,
	}

	err := setup.CreateTestFiles(files)
	if err != nil {
		t.Fatalf("Failed to create test files: %v", err)
	}

	parser := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0)
	ctx := t.Context()

	// First run — populates cache (misses)
	run := func() (chan []*syntax.Node, chan IncrementalStats) {
		fchan := make(chan string, len(files))
		fchan <- setup.GetFilePath("one.go")
		fchan <- setup.GetFilePath("two.go")
		close(fchan)

		return parser.ParseIncrementalParallel(ctx, fchan, 4)
	}

	schanFirst, _ := run()
	drainIncremental(t, schanFirst)

	// Second run — should hit cache
	schan2, statsChan := run()
	drainIncremental(t, schan2)

	stats := <-statsChan
	testutil.AssertFieldValue(t, stats.CacheHits, len(files), "CacheHits")
	testutil.AssertFieldValue(t, stats.CacheMisses, 0, "CacheMisses")
}

// TestIncrementalParallelContextCancellation verifies the parallel path
// terminates promptly when the context is canceled, with no deadlock or
// goroutine leak.
func TestIncrementalParallelContextCancellation(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	writeHelloWorldFile(t, setup)

	parser := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0)

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	// Feed files slowly from an unbuffered channel so cancellation happens
	// mid-stream rather than after all work is done.
	fchan := make(chan string)

	go func() {
		fchan <- setup.GetFilePath("test.go")

		// Block briefly so the cancel below lands while the pipeline is live.
		time.Sleep(50 * time.Millisecond)
		close(fchan)
	}()

	schan, _ := parser.ParseIncrementalParallel(ctx, fchan, 4)

	// Drain whatever was produced, then cancel.
	done := make(chan struct{})

	go func() {
		defer close(done)
		for range schan {
		}
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Pipeline shut down cleanly.
	case <-time.After(5 * time.Second):
		t.Fatal("parallel incremental parser deadlocked on context cancellation")
	}
}

// TestIncrementalParallelWorkerZero verifies that workers<=0 falls back to
// GOMAXPROCS (never panics or hangs).
func TestIncrementalParallelWorkerZero(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	writeHelloWorldFile(t, setup)

	parser := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0)

	schan, statsChan := parser.ParseIncrementalParallel(t.Context(), singleFileChannel(setup), 0)

	seqs := drainIncremental(t, schan)

	if len(seqs) != 1 {
		t.Fatalf("Expected 1 sequence with workers=0, got %d", len(seqs))
	}

	stats := <-statsChan
	testutil.AssertIncrementalStats(t, stats.FilesCount, stats.CacheMisses, 1, 1)
}

// TestIncrementalParallelMutationIsolation verifies that mutating the nodes
// returned from one parallel parse does not affect a second parse of the same
// file (cache independence + deep-clone correctness under concurrency).
func TestIncrementalParallelMutationIsolation(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	writeHelloWorldFile(t, setup)

	parser := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0)
	ctx := t.Context()

	// First run — cache miss, populates cache
	schan1, _ := parser.ParseIncrementalParallel(ctx, singleFileChannel(setup), 4)
	seqs1 := drainIncremental(t, schan1)

	// Mutate the returned nodes aggressively
	for _, seq := range seqs1 {
		for _, node := range seq {
			mutateSubtree(node)
		}
	}

	// Second run — cache hit; must return UNMUTATED nodes
	schan2, _ := parser.ParseIncrementalParallel(ctx, singleFileChannel(setup), 4)
	seqs2 := drainIncremental(t, schan2)

	for _, seq := range seqs2 {
		for _, node := range seq {
			if node.Type == 999999 {
				t.Error("cache-hit nodes were mutated by the previous caller — clone isolation broken")
			}
		}
	}
}

// mutateSubtree corrupts a node subtree to detect aliasing.
func mutateSubtree(n *syntax.Node) {
	if n == nil {
		return
	}

	n.Type = 999999
	n.Name = "CORRUPTED"

	for _, c := range n.Children {
		mutateSubtree(c)
	}
}

// TestIncrementalParallelConcurrentStress runs many identical files through a
// small worker pool repeatedly, maximizing the chance of catching a race with
// -race -count. This is a broader stress than the single identical-content test.
func TestIncrementalParallelConcurrentStress(t *testing.T) {
	setup := testutil.NewTestFileSetup(t)
	cacheDir := setup.TmpDir + "/cache"

	const (
		numFiles = 24
		workers  = 8
	)

	filenames := make([]string, numFiles)
	for i := range numFiles {
		filenames[i] = fmt.Sprintf("s_%02d.go", i)
	}

	// Two content variants so we exercise both cache-miss and singleflight paths.
	contentA := `package main

func fa() { println("a"); x := 1; _ = x }`
	contentB := `package main

func fb() { println("b"); y := 2; _ = y }`

	files := make(map[string]string, numFiles)
	for i := range numFiles {
		if i%2 == 0 {
			files[filenames[i]] = contentA
		} else {
			files[filenames[i]] = contentB
		}
	}

	err := setup.CreateTestFiles(files)
	if err != nil {
		t.Fatalf("Failed to create test files: %v", err)
	}

	parser := NewIncrementalParser(cacheDir, false, golang.DetectionModeSemantic, 0, 0)
	ctx := t.Context()

	fchan := make(chan string, numFiles)
	// Interleave A and B to maximize hash collisions across workers.
	for i := range numFiles {
		fchan <- setup.GetFilePath(filenames[i])
	}
	close(fchan)

	schan, statsChan := parser.ParseIncrementalParallel(ctx, fchan, workers)

	seqs := drainIncremental(t, schan)

	if len(seqs) != numFiles {
		t.Fatalf("Expected %d sequences, got %d", numFiles, len(seqs))
	}

	// Every file must be represented by a distinct filename.
	seen := filenamesFromSeqs(seqs)
	if len(seen) != numFiles {
		t.Errorf("Expected %d distinct filenames, got %d", numFiles, len(seen))
	}

	stats := <-statsChan
	testutil.AssertFieldValue(t, stats.FilesCount, numFiles, "FilesCount")
}
