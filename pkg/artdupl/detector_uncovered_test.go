package artdupl

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LarsArtmann/art-dupl/detection"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/pkg/logger"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// testNodes returns a slice of test nodes for testing.
func testNodes() []*syntax.Node {
	return makeNodePair("file.go", 10, 20)
}

// makeNodePair creates a pair of test nodes with sequential positions.
func makeNodePair(filename string, pos, end int32) []*syntax.Node {
	return testutil.CreateNodeSlice([]struct {
		Type     int32
		Filename string
		Pos      int32
		End      int32
	}{
		{Type: 1, Filename: filename, Pos: pos, End: end},
		{Type: 2, Filename: filename, Pos: pos + 1, End: end + 1},
	})
}

// captureProgress returns a progress callback that stores the received Progress.
func captureProgress(received **Progress) func(*Progress) error {
	return func(p *Progress) error {
		*received = p

		return nil
	}
}

// makeFrag creates a fragment with the given filename and positions.
func makeFrag(filename string, pos, end int32) []*syntax.Node {
	return testutil.CreateNodeSlice([]struct {
		Type     int32
		Filename string
		Pos      int32
		End      int32
	}{
		{Type: 1, Filename: filename, Pos: pos, End: end},
	})
}

// buildTreeFromFrag creates a suffix tree populated from the given fragment nodes.
func buildTreeFromFrag(frag []*syntax.Node) *suffixtree.STree {
	tree := suffixtree.New()
	for _, node := range frag {
		tree.Update(node)
	}

	tree.Update(&syntax.Node{Type: -1})

	return tree
}

// drainMatchesChannel builds a test pipeline and drains the MultiDetector matches channel.
func drainMatchesChannel(t *testing.T) {
	t.Helper()

	d := newTestDetector()
	data := testNodes()
	tree := buildTestSuffixTree(data)

	md := detection.NewMultiDetector(
		detection.Config{Methods: d.cfg.DetectionMethods},
		data,
		tree,
	)
	matchesChan := md.FindDuplOver(context.Background(), 1)

	if matchesChan == nil {
		t.Error("Expected non-nil matches channel")
	}

	for range matchesChan {
	}
}

// buildTestSuffixTree builds a suffix tree from nodes with a sentinel.
func buildTestSuffixTree(data []*syntax.Node) *suffixtree.STree {
	tree := suffixtree.New()
	for _, node := range data {
		tree.Update(node)
	}

	tree.Update(&syntax.Node{Type: -1})

	return tree
}

// newTestDetector creates a detector with default test configuration.
func newTestDetector() *detector {
	return &detector{
		opts: &Options{
			DetectionMethods: []DetectionMethod{MethodArtDupl},
		},
		cfg: newTestConfig(),
	}
}

// newFragmentDetector creates a detector configured for fragment extraction.
func newFragmentDetector() *detector {
	return &detector{
		opts: &Options{
			IncludeFragments: true,
			FileReader:       os.ReadFile,
		},
		logger: &logger.NoOpLogger{},
	}
}

// TestConvertToCloneGroup tests the convertToCloneGroup function.
func TestConvertToCloneGroup(t *testing.T) {
	d := &detector{
		opts: &Options{
			MaxClonesPerGroup: 10,
		},
	}

	// Create test fragments
	frag1 := makeNodePair("file1.go", 10, 20)
	frag2 := makeNodePair("file2.go", 30, 40)

	frags := [][]*syntax.Node{frag1, frag2}

	group := d.convertToCloneGroup("test-hash", frags, MethodArtDupl)

	assertCloneGroupBasic(t, group, "test-hash", 2)

	testutil.AssertFieldValue(t, group.Method, MethodArtDupl, "Method")

	testutil.AssertFieldValue(t, group.Size, 4, "Size")

	if group.LineCount == 0 {
		t.Error("Expected LineCount > 0")
	}
}

// TestConvertToCloneGroup_MaxClonesLimit tests the max clones per group limit.
func TestConvertToCloneGroup_MaxClonesLimit(t *testing.T) {
	d := &detector{
		opts: &Options{
			MaxClonesPerGroup: 2,
		},
	}

	// Create 5 fragments
	frags := make([][]*syntax.Node, 0, 5)
	for i := range 5 {
		frags = append(frags, []*syntax.Node{
			{Type: 1, Filename: "file.go", Pos: int32(i), End: int32(i + 1)},
		})
	}

	group := d.convertToCloneGroup("hash", frags, MethodArtDupl)

	assertCloneCount(t, len(group.Clones), 2)
}

// TestConvertFragmentToClone tests the convertFragmentToClone function.
func TestConvertFragmentToClone(t *testing.T) {
	t.Run("with nodes", func(t *testing.T) {
		d := &detector{opts: &Options{IncludeFragments: false}}
		frag := []*syntax.Node{
			{Type: 1, Filename: testFilename, Pos: 10, End: 20},
			{Type: 2, Filename: testFilename, Pos: 11, End: 21},
			{Type: 3, Filename: testFilename, Pos: 12, End: 22},
		}

		clone := d.convertFragmentToClone(frag)

		testutil.AssertFieldValue(t, clone.Filename, testFilename, "Filename")

		testutil.AssertFieldValue(t, clone.LineStart, 10, "LineStart")

		testutil.AssertFieldValue(t, clone.LineEnd, 22, "LineEnd")

		testutil.AssertFieldValue(t, clone.Size, 3, "Size")

		if clone.Fragment != "" {
			t.Error("Expected no fragment content when IncludeFragments=false")
		}
	})

	t.Run("empty fragment", func(t *testing.T) {
		d := &detector{opts: &Options{}}
		clone := d.convertFragmentToClone([]*syntax.Node{})

		// Empty fragments now return nil (not valid clones)
		if clone != nil {
			t.Error("Expected nil clone for empty fragment")
		}
	})
}

// TestExtractFragmentContent_NotFound tests extractFragmentContent when file not found.
func TestExtractFragmentContent_NotFound(t *testing.T) {
	d := &detector{
		opts: &Options{
			IncludeFragments: true,
			FileReader: func(filename string) ([]byte, error) {
				return nil, errors.New("file not found")
			},
		},
		logger: &logger.NoOpLogger{},
	}

	frag := makeFrag("/nonexistent/file.go", 1, 2)

	content := d.extractFragmentContent(frag)
	if content != "" {
		t.Errorf("Expected empty string for missing file, got %q", content)
	}
}

// TestExtractFragmentContent_Empty tests extractFragmentContent with empty fragment.
func TestExtractFragmentContent_Empty(t *testing.T) {
	d := &detector{opts: &Options{IncludeFragments: true}}

	content := d.extractFragmentContent([]*syntax.Node{})
	if content != "" {
		t.Errorf("Expected empty string, got %s", content)
	}
}

// TestExtractFragmentContent_InvalidRange tests extractFragmentContent with invalid line range.
func TestExtractFragmentContent_InvalidRange(t *testing.T) {
	d := &detector{
		opts: &Options{
			IncludeFragments: true,
			FileReader: func(fn string) ([]byte, error) {
				return []byte("line1\nline2\n"), nil
			},
		},
		logger: &logger.NoOpLogger{},
	}

	// Create fragment with out-of-range positions
	frag := []*syntax.Node{
		{Type: 1, Filename: testFilename, Pos: 100, End: 200}, // Way beyond file
	}

	content := d.extractFragmentContent(frag)
	if content != "" {
		t.Errorf("Expected empty string for invalid range, got %q", content)
	}
}

// TestExtractFragmentContent_WithFragments tests extractFragmentContent with IncludeFragments.
func TestExtractFragmentContent_WithFragments(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, testFilename)

	content := []byte("line1\nline2\nline3\n")

	err := os.WriteFile(testFile, content, 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	d := newFragmentDetector()

	frag := []*syntax.Node{testutil.CreateNodeWithPos(1, testFile, 1, 2)}

	result := d.extractFragmentContent(frag)
	if result == "" {
		t.Error("Expected non-empty fragment content")
	}
}

func TestRunDetectionMethods(t *testing.T) {
	t.Parallel()

	for _, name := range []string{"hash", "suffix-tree"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			drainMatchesChannel(t)
		})
	}
}

// TestReportProgress tests the reportProgress function.
func TestReportProgress(t *testing.T) {
	var receivedProgress *Progress

	d := &detector{
		opts: &Options{
			ProgressCallback: captureProgress(&receivedProgress),
		},
	}

	d.reportProgress(75.5, "Analyzing", "current.go")

	if receivedProgress == nil {
		t.Fatal("Progress onProgress was not called")
	}

	testutil.AssertFieldValue(t, receivedProgress.Stage, "Analyzing", "Stage")

	testutil.AssertFieldValue(t, receivedProgress.Completed, 75, "Completed")

	testutil.AssertFieldValue(t, receivedProgress.CurrentFile, "current.go", "CurrentFile")
}

// TestReportProgress_NoCallback tests reportProgress with no callback.
func TestReportProgress_NoCallback(t *testing.T) {
	d := &detector{
		opts: &Options{
			ProgressCallback: nil,
		},
	}

	// Should not panic
	d.reportProgress(50, "Stage", "file.go")
}

// TestReportProgress_CallbackError tests reportProgress when callback returns error.
func TestReportProgress_CallbackError(t *testing.T) {
	progressHandler := func(p *Progress) error {
		return errors.New("progressHandler error")
	}

	logger := &logger.NoOpLogger{}
	d := &detector{
		opts: &Options{
			ProgressCallback: progressHandler,
			Logger:           logger,
		},
		logger: logger,
	}

	// Should not panic, just log the error
	d.reportProgress(50, "Stage", "file.go")
}

// TestCollectMatchesIntoGroups tests the collectMatchesIntoGroups function.
func TestCollectMatchesIntoGroups(t *testing.T) {
	matchesChan := make(chan syntax.Match, 3)

	// Send test matches
	matchesChan <- testutil.CreateMatch("hash1", "file1.go")

	matchesChan <- testutil.CreateMatch("hash1", "file2.go")

	matchesChan <- testutil.CreateMatch("hash2", "file3.go")

	close(matchesChan)

	groups, err := collectMatchesIntoGroups(t.Context(), matchesChan)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	assertCloneCount(t, len(groups), 2)
	assertCloneCount(t, len(groups["hash1"]), 2)
	assertCloneCount(t, len(groups["hash2"]), 1)
}

// TestCollectMatchesIntoGroups_Cancelled tests collectMatchesIntoGroups with cancelled context.
func TestCollectMatchesIntoGroups_Cancelled(t *testing.T) {
	matchesChan := make(chan syntax.Match, 1)
	ctx, cancel := context.WithCancel(t.Context())

	// Send match and close channel
	matchesChan <- syntax.Match{Hash: "hash1", Frags: [][]*syntax.Node{{{}}}}

	close(matchesChan)

	// Cancel after processing starts
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := collectMatchesIntoGroups(ctx, matchesChan)
	// The function may or may not return an error depending on timing
	// We just want to make sure it doesn't hang or panic
	_ = err
}

// TestStreamDetectionResults tests the streamDetectionResults function.
func TestStreamDetectionResults(t *testing.T) {
	d := newTestDetector()

	resultChan := make(chan *CloneGroup, 1)
	testFrag := []*syntax.Node{testutil.CreateNodeWithPos(1, "file.go", 10, 20)}

	tree := buildTreeFromFrag(testFrag)

	pipeline := &pipelineResult{data: testFrag, tree: tree}

	ctx := t.Context()

	err := d.streamDetectionResults(ctx, pipeline, resultChan)
	if err != nil {
		t.Errorf("streamDetectionResults with valid pipeline should not error, got: %v", err)
	}
}

// TestStreamDetectionResults_Cancelled tests streamDetectionResults with cancelled context.
func TestStreamDetectionResults_Cancelled(t *testing.T) {
	d := newTestDetector()

	resultChan := make(chan *CloneGroup)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	fragNodes := []*syntax.Node{testutil.CreateNodeWithPos(1, "file.go", 10, 20)}

	tree := buildTreeFromFrag(fragNodes)

	pipeline := &pipelineResult{data: fragNodes, tree: tree}

	err := d.streamDetectionResults(ctx, pipeline, resultChan)
	if err != nil {
		t.Errorf("streamDetectionResults with cancelled context should return nil (graceful), got: %v", err)
	}
}

// TestConvertFragmentToClone_WithFragments tests convertFragmentToClone with IncludeFragments.
func TestConvertFragmentToClone_WithFragments(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, testFilename)

	content := []byte("line1\nline2\nline3\n")

	err := os.WriteFile(testFile, content, 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	d := newFragmentDetector()

	frag := []*syntax.Node{testutil.CreateNodeWithPos(1, testFile, 1, 2)}

	clone := d.convertFragmentToClone(frag)

	if clone.Fragment == "" {
		t.Error("Expected fragment content when IncludeFragments=true")
	}
}

// TestValidateFile_IgnorePatternMatching tests file pattern matching in validateFile.
func TestValidateFile_IgnorePatternMatching(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, testFilename)
	testutil.WriteTestFile(t, testFile, "package test")

	d := &detector{
		opts: &Options{
			MaxFileSize: 1024 * 1024,
			IgnoreFiles: []string{testFilename}, // Exact match
		},
	}

	err := d.validateFile(testFile)
	if !errors.Is(err, ErrFileIgnored) {
		t.Errorf("Expected ErrFileIgnored for ignored file, got %v", err)
	}
}

// TestConvertToCloneGroup_SingleFragment tests convertToCloneGroup with single fragment.
func TestConvertToCloneGroup_SingleFragment(t *testing.T) {
	d := &detector{
		opts: &Options{MaxClonesPerGroup: 0}, // No limit
	}

	frag := [][]*syntax.Node{
		{
			{Type: 1, Filename: "single.go", Pos: 1, End: 5},
		},
	}

	group := d.convertToCloneGroup("single", frag, MethodArtDupl)

	assertCloneCount(t, len(group.Clones), 1)

	if group.Clones[0].Filename != "single.go" {
		t.Errorf("Expected Filename='single.go', got %s", group.Clones[0].Filename)
	}
}

func TestBuildSuffixTree(t *testing.T) {
	tree := buildTestSuffixTree(testNodes())

	if tree == nil {
		t.Error("Expected non-nil tree")
	}
}
