package artdupl

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/detection"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/pkg/logger"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// testNodes returns a slice of test nodes for testing.
func testNodes() []*syntax.Node {
	return []*syntax.Node{
		{Type: 1, Filename: "file.go", Pos: 10, End: 20},
		{Type: 2, Filename: "file.go", Pos: 11, End: 21},
	}
}

// makeFrag creates a fragment with the given filename and positions.
func makeFrag(filename string, pos, end int32) []*syntax.Node {
	return []*syntax.Node{
		{Type: 1, Filename: filename, Pos: pos, End: end},
	}
}

// verifyCloneCount verifies the expected number of clones.
func verifyCloneCount(t *testing.T, got, expected int) {
	t.Helper()

	if got != expected {
		t.Errorf("clone count mismatch: want %d, got %d", expected, got)
	}
}

// newTestDetector creates a detector with default test configuration.
func newTestDetector() *detector {
	return &detector{
		opts: &Options{
			DetectionMethods: []DetectionMethod{MethodArtDupl},
		},
		config: newTestConfig(),
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
	frag1 := []*syntax.Node{
		{Type: 1, Filename: "file1.go", Pos: 10, End: 20},
		{Type: 2, Filename: "file1.go", Pos: 11, End: 21},
	}
	frag2 := []*syntax.Node{
		{Type: 1, Filename: "file2.go", Pos: 30, End: 40},
		{Type: 2, Filename: "file2.go", Pos: 31, End: 41},
	}

	frags := [][]*syntax.Node{frag1, frag2}

	group := d.convertToCloneGroup("test-hash", frags, MethodArtDupl)

	assertCloneGroupBasic(t, group, "test-hash", 2)

	if group.Method != MethodArtDupl {
		t.Errorf("Expected Method=MethodArtDupl, got %v", group.Method)
	}

	if group.Size != 4 {
		t.Errorf("Expected Size=4, got %d", group.Size)
	}

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
	frags := make([][]*syntax.Node, 5)
	for i := range 5 {
		frags[i] = []*syntax.Node{
			{Type: 1, Filename: "file.go", Pos: int32(i), End: int32(i + 1)},
		}
	}

	group := d.convertToCloneGroup("hash", frags, MethodArtDupl)

	verifyCloneCount(t, len(group.Clones), 2)
}

// TestConvertFragmentToClone tests the convertFragmentToClone function.
func TestConvertFragmentToClone(t *testing.T) {
	t.Run("with nodes", func(t *testing.T) {
		d := &detector{opts: &Options{IncludeFragments: false}}
		frag := []*syntax.Node{
			{Type: 1, Filename: "test.go", Pos: 10, End: 20},
			{Type: 2, Filename: "test.go", Pos: 11, End: 21},
			{Type: 3, Filename: "test.go", Pos: 12, End: 22},
		}

		clone := d.convertFragmentToClone(frag)

		if clone.Filename != "test.go" {
			t.Errorf("Expected Filename='test.go', got %s", clone.Filename)
		}

		if clone.StartLine != 10 {
			t.Errorf("Expected StartLine=10, got %d", clone.StartLine)
		}

		if clone.EndLine != 22 {
			t.Errorf("Expected EndLine=22, got %d", clone.EndLine)
		}

		if clone.Size != 3 {
			t.Errorf("Expected Size=3, got %d", clone.Size)
		}

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
			FileReader: func(filename string) ([]byte, error) {
				return []byte("line1\nline2\n"), nil
			},
		},
		logger: &logger.NoOpLogger{},
	}

	// Create fragment with out-of-range positions
	frag := []*syntax.Node{
		{Type: 1, Filename: "test.go", Pos: 100, End: 200}, // Way beyond file
	}

	content := d.extractFragmentContent(frag)
	if content != "" {
		t.Errorf("Expected empty string for invalid range, got %q", content)
	}
}

// TestExtractFragmentContent_WithFragments tests extractFragmentContent with IncludeFragments.
func TestExtractFragmentContent_WithFragments(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := []byte("line1\nline2\nline3\n")

	err := os.WriteFile(testFile, content, 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	d := &detector{
		opts: &Options{
			IncludeFragments: true,
			FileReader:       os.ReadFile,
		},
		logger: &logger.NoOpLogger{},
	}

	frag := []*syntax.Node{testutil.CreateNodeWithPos(1, testFile, 1, 2)}

	result := d.extractFragmentContent(frag)
	if result == "" {
		t.Error("Expected non-empty fragment content")
	}
}

// TestRunHashDetection tests hash-based detection via MultiDetector.
func TestRunHashDetection(t *testing.T) {
	d := newTestDetector()

	data := testNodes()

	tree := suffixtree.New()
	for _, node := range data {
		tree.Update(node)
	}

	tree.Update(&syntax.Node{Type: -1})

	md := detection.NewMultiDetector(config.DetectionConfig{Methods: d.config.DetectionMethods}, data, tree)
	matchesChan := md.FindDuplOver(1)

	if matchesChan == nil {
		t.Error("Expected non-nil matches channel")
	}

	for range matchesChan {
	}
}

// TestReportProgress tests the reportProgress function.
func TestReportProgress(t *testing.T) {
	var receivedProgress *Progress

	callback := func(p *Progress) error {
		receivedProgress = p

		return nil
	}

	d := &detector{
		opts: &Options{
			ProgressCallback: callback,
		},
	}

	d.reportProgress(75.5, "Analyzing", "current.go")

	if receivedProgress == nil {
		t.Fatal("Progress callback was not called")
	}

	if receivedProgress.Stage != "Analyzing" {
		t.Errorf("Expected Stage='Analyzing', got %s", receivedProgress.Stage)
	}

	if receivedProgress.Completed != 75 {
		t.Errorf("Expected Completed=75, got %d", receivedProgress.Completed)
	}

	if receivedProgress.CurrentFile != "current.go" {
		t.Errorf("Expected CurrentFile='current.go', got %s", receivedProgress.CurrentFile)
	}
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
	callback := func(p *Progress) error {
		return errors.New("callback error")
	}

	logger := &logger.NoOpLogger{}
	d := &detector{
		opts: &Options{
			ProgressCallback: callback,
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

	verifyCloneCount(t, len(groups), 2)
	verifyCloneCount(t, len(groups["hash1"]), 2)
	verifyCloneCount(t, len(groups["hash2"]), 1)
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
	data := []*syntax.Node{testutil.CreateNodeWithPos(1, "file.go", 10, 20)}

	tree := suffixtree.New()
	for _, node := range data {
		tree.Update(node)
	}

	tree.Update(&syntax.Node{Type: -1})

	pipeline := &pipelineResult{data: data, tree: tree}

	ctx := t.Context()

	err := d.streamDetectionResults(ctx, pipeline, resultChan)
	if err != nil {
		t.Logf("streamDetectionResults returned: %v", err)
	}
}

// TestStreamDetectionResults_Cancelled tests streamDetectionResults with cancelled context.
func TestStreamDetectionResults_Cancelled(t *testing.T) {
	d := newTestDetector()

	resultChan := make(chan *CloneGroup)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	data := []*syntax.Node{testutil.CreateNodeWithPos(1, "file.go", 10, 20)}

	tree := suffixtree.New()
	for _, node := range data {
		tree.Update(node)
	}

	tree.Update(&syntax.Node{Type: -1})

	pipeline := &pipelineResult{data: data, tree: tree}

	_ = d.streamDetectionResults(ctx, pipeline, resultChan)
}

// TestConvertFragmentToClone_WithFragments tests convertFragmentToClone with IncludeFragments.
func TestConvertFragmentToClone_WithFragments(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.go")

	content := []byte("line1\nline2\nline3\n")

	err := os.WriteFile(testFile, content, 0o644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	d := &detector{
		opts: &Options{
			IncludeFragments: true,
			FileReader:       os.ReadFile,
		},
	}

	frag := []*syntax.Node{testutil.CreateNodeWithPos(1, testFile, 1, 2)}

	clone := d.convertFragmentToClone(frag)

	if clone.Fragment == "" {
		t.Error("Expected fragment content when IncludeFragments=true")
	}
}

// TestValidateFile_IgnorePatternMatching tests file pattern matching in validateFile.
func TestValidateFile_IgnorePatternMatching(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.go")
	testutil.WriteTestFile(t, testFile, "package test")

	d := &detector{
		opts: &Options{
			MaxFileSize: 1024 * 1024,
			IgnoreFiles: []string{"test.go"}, // Exact match
		},
	}

	err := d.validateFile(testFile)
	if !errors.Is(err, ErrParsingFailed) {
		t.Errorf("Expected ErrParsingFailed for ignored file, got %v", err)
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

	verifyCloneCount(t, len(group.Clones), 1)

	if group.Clones[0].Filename != "single.go" {
		t.Errorf("Expected Filename='single.go', got %s", group.Clones[0].Filename)
	}
}

// TestRunSuffixTreeDetection tests suffix tree detection via MultiDetector.
func TestRunSuffixTreeDetection(t *testing.T) {
	d := newTestDetector()

	data := testNodes()

	tree := suffixtree.New()
	for _, node := range data {
		tree.Update(node)
	}

	tree.Update(&syntax.Node{Type: -1})

	md := detection.NewMultiDetector(config.DetectionConfig{Methods: d.config.DetectionMethods}, data, tree)
	matchesChan := md.FindDuplOver(1)

	if matchesChan == nil {
		t.Error("Expected non-nil matches channel")
	}

	for range matchesChan {
	}
}

// TestBuildSuffixTree tests building a suffix tree from node data.
func TestBuildSuffixTree(t *testing.T) {
	data := testNodes()

	tree := suffixtree.New()
	for _, node := range data {
		tree.Update(node)
	}

	if tree == nil {
		t.Error("Expected non-nil tree")
	}
}
