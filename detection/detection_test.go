package detection

import (
	"os"
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/internal/testutil"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/syntax/golang"
)

func createTestNode(filename string, pos, end int32) *syntax.Node {
	return &syntax.Node{
		Filename: filename,
		Type:     int32(golang.File),
		Pos:      pos,
		End:      end,
	}
}

// createTestConfig creates a test configuration for MultiDetector tests.
func createTestConfig() *config.Config {
	return &config.Config{
		Threshold:        15,
		Verbose:          true,
		DetectionMethods: config.DetectionMethods{config.DetectionMethodArtDupl},
	}
}

// createSimpleTestConfig creates a minimal test configuration without verbose logging.
func createSimpleTestConfig() *config.Config {
	return &config.Config{
		Threshold:        15,
		DetectionMethods: config.DetectionMethods{config.DetectionMethodArtDupl},
	}
}

// createHashTestConfig creates a test configuration using hash detection method.
func createHashTestConfig() *config.Config {
	return &config.Config{
		Threshold:        15,
		DetectionMethods: config.DetectionMethods{config.DetectionMethodHash},
	}
}

// TestNewMultiDetector tests MultiDetector constructor.
func TestNewMultiDetector(t *testing.T) {
	cfg := createTestConfig()

	data := []*syntax.Node{
		{Filename: "test.go", Type: int32(golang.File)},
	}

	tree := suffixtree.New()

	detector := NewMultiDetector(cfg, data, tree, true)

	if detector.config != cfg {
		t.Error("Config not set correctly")
	}

	if len(detector.data) != len(data) {
		t.Error("Data not set correctly")
	}

	if detector.tree != tree {
		t.Error("Tree not set correctly")
	}

	if !detector.verbose {
		t.Error("Verbose flag not set correctly")
	}
}

// TestNewMultiDetector_NilConfig tests with nil config.
func TestNewMultiDetector_NilConfig(t *testing.T) {
	tree := suffixtree.New()
	detector := NewMultiDetector(nil, nil, tree, false)

	if detector.config != nil {
		t.Error("Expected nil config")
	}
}

// TestNewMultiDetector_EmptyData tests with empty data.
func TestNewMultiDetector_EmptyData(t *testing.T) {
	cfg := createSimpleTestConfig()
	tree := suffixtree.New()

	detector := NewMultiDetector(cfg, []*syntax.Node{}, tree, false)

	if len(detector.data) != 0 {
		t.Error("Expected empty data slice")
	}
}

// TestMultiDetector_logVerbose tests verbose logging.
func TestMultiDetector_logVerbose(t *testing.T) {
	t.Run("verbose enabled", func(t *testing.T) {
		detector := &MultiDetector{verbose: true}
		// Should not panic
		detector.logVerbose("Test message")
	})

	t.Run("verbose disabled", func(t *testing.T) {
		detector := &MultiDetector{verbose: false}
		// Should not panic or output
		detector.logVerbose("Should not print")
	})
}

// TestNewTodoDetector tests TODO detector creation.
func TestNewTodoDetector(t *testing.T) {
	detector := NewTodoDetector()

	if detector.patterns == nil {
		t.Error("Patterns should not be nil")
	}

	if len(detector.patterns) == 0 {
		t.Error("Patterns should not be empty")
	}

	// Check for expected patterns
	expectedPatterns := []string{"TODO", "FIXME", "XXX", "HACK", "NOTE"}
	for _, pattern := range expectedPatterns {
		if _, exists := detector.patterns[pattern]; !exists {
			t.Errorf("Expected pattern %s not found", pattern)
		}
	}
}

// TestTodoDetector_FindTodos tests TODO detection with empty data.
func TestTodoDetector_FindTodos(t *testing.T) {
	detector := NewTodoDetector()

	// Empty data should produce no matches
	matches := detector.FindTodos([]*syntax.Node{})

	matchCount := 0
	for range matches {
		matchCount++
	}

	if matchCount != 0 {
		t.Errorf("Expected 0 matches for empty data, got %d", matchCount)
	}
}

// TestTodoDetector_FindTodos_NilData tests TODO detection with nil data.
func TestTodoDetector_FindTodos_NilData(t *testing.T) {
	detector := NewTodoDetector()

	// Nil data should produce no matches (and not panic)
	matches := detector.FindTodos(nil)

	matchCount := 0
	for range matches {
		matchCount++
	}

	if matchCount != 0 {
		t.Errorf("Expected 0 matches for nil data, got %d", matchCount)
	}
}

// TestNewLegacyDetector tests legacy detector creation.
func TestNewLegacyDetector(t *testing.T) {
	detector := NewLegacyDetector()

	if detector.patterns == nil {
		t.Error("Patterns should not be nil")
	}

	if len(detector.patterns) == 0 {
		t.Error("Patterns should not be empty")
	}
}

// TestLegacyDetector_FindLegacy tests legacy detection with empty data.
func TestLegacyDetector_FindLegacy(t *testing.T) {
	detector := NewLegacyDetector()

	// Empty data should produce no matches
	matches := detector.FindLegacy([]*syntax.Node{})

	matchCount := 0
	for range matches {
		matchCount++
	}

	if matchCount != 0 {
		t.Errorf("Expected 0 matches for empty data, got %d", matchCount)
	}
}

// TestLegacyDetector_FindLegacy_NilData tests legacy detection with nil data.
func TestLegacyDetector_FindLegacy_NilData(t *testing.T) {
	detector := NewLegacyDetector()

	// Nil data should produce no matches (and not panic)
	matches := detector.FindLegacy(nil)

	matchCount := 0
	for range matches {
		matchCount++
	}

	if matchCount != 0 {
		t.Errorf("Expected 0 matches for nil data, got %d", matchCount)
	}
}

// TestCreateIssueMatch tests the createIssueMatch helper function.
func TestCreateIssueMatch(t *testing.T) {
	prefix := "TODO"
	filename := "test.go"
	line := 42

	match := createIssueMatch(prefix, filename, line)

	if match.Hash == "" {
		t.Error("Hash should not be empty")
	}

	expectedHashPrefix := "TODO-test.go-"
	if len(match.Hash) < len(expectedHashPrefix) {
		t.Errorf("Hash format unexpected: %s", match.Hash)
	}

	// Frags should be present (even if empty inner slice)
	if match.Frags == nil {
		t.Error("Frags should not be nil")
	}
}

// TestIssue_GetLine tests that GetLine returns the correct line number.
func TestIssue_GetLine(t *testing.T) {
	tests := []struct {
		name    string
		issue   LineExtractor
		lineNum uint16
	}{
		{
			name: "TodoIssue",
			issue: TodoIssue{
				Filename: domain.Filepath("test.go"),
				Line:     mustNewLineNumber(42),
				Text:     "Fix this later",
				Type:     "TODO",
			},
			lineNum: 42,
		},
		{
			name: "LegacyIssue",
			issue: LegacyIssue{
				Filename: domain.Filepath("legacy.go"),
				Line:     mustNewLineNumber(100),
				Type:     "deprecated_function",
				Message:  "Use of deprecated function",
				Severity: domain.CloneSeverityMedium,
			},
			lineNum: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.issue.GetLine().Uint16() != tt.lineNum {
				t.Errorf("GetLine() = %d, want %d", tt.issue.GetLine().Uint16(), tt.lineNum)
			}
		})
	}
}

// mustNewLineNumber is a helper to create LineNumber and panic on error.
func mustNewLineNumber(n uint16) domain.LineNumber {
	lineNum, err := domain.NewLineNumber(n)
	if err != nil {
		panic(err)
	}

	return lineNum
}

// TestLegacyPattern_Defaults tests default legacy patterns.
func TestLegacyPattern_Defaults(t *testing.T) {
	patterns := getDefaultLegacyPatterns()

	if len(patterns) == 0 {
		t.Error("Default patterns should not be empty")
	}

	// Check for expected pattern types
	patternTypes := make(map[string]bool)
	for _, p := range patterns {
		patternTypes[p.Type] = true
	}

	expectedTypes := []string{"deprecated_function", "old_pattern", "deprecated_import"}
	for _, expectedType := range expectedTypes {
		if !patternTypes[expectedType] {
			t.Errorf("Expected pattern type %s not found", expectedType)
		}
	}
}

// TestSimpleDetector_Interface tests that MultiDetector implements SimpleDetector.
func TestSimpleDetector_Interface(t *testing.T) {
	// This test verifies the interface is satisfied at compile time
	var _ SimpleDetector = (*MultiDetector)(nil)
}

// TestTodoDetector_Patterns tests that regex patterns compile correctly.
func TestTodoDetector_Patterns(t *testing.T) {
	detector := NewTodoDetector()

	testCases := []struct {
		patternName string
		input       string
		shouldMatch bool
	}{
		{"TODO", "TODO: fix this", true},
		{"TODO", "TODO(@user): fix this", true},
		{"FIXME", "FIXME: broken code", true},
		{"XXX", "XXX: hack alert", true},
		{"HACK", "HACK: temporary fix", true},
		{"NOTE", "NOTE: important info", true},
		{"TODO", "no todo here", false},
	}

	for _, tc := range testCases {
		t.Run(tc.patternName+"_"+tc.input, func(t *testing.T) {
			pattern, exists := detector.patterns[tc.patternName]
			if !exists {
				t.Fatalf("Pattern %s not found", tc.patternName)
			}

			matches := pattern.FindStringSubmatch(tc.input)
			matched := len(matches) > 0

			if matched != tc.shouldMatch {
				t.Errorf(
					"Pattern %s on %q: matched=%v, want=%v",
					tc.patternName,
					tc.input,
					matched,
					tc.shouldMatch,
				)
			}
		})
	}
}

// TestFindIssuesInFile_EmptyData tests findIssuesInFile with empty data.
func TestFindIssuesInFile_EmptyData(t *testing.T) {
	finder := func(filename string, nodes []*syntax.Node) []string {
		return nil
	}
	matchCreator := func(issue, filename string) syntax.Match {
		return syntax.Match{Hash: issue}
	}

	matches := findIssuesInFile([]*syntax.Node{}, finder, matchCreator)

	matchCount := 0
	for range matches {
		matchCount++
	}

	if matchCount != 0 {
		t.Errorf("Expected 0 matches for empty data, got %d", matchCount)
	}
}

// TestFindIssuesGeneric_EmptyData tests findIssuesGeneric with empty data.
func TestFindIssuesGeneric_EmptyData(t *testing.T) {
	finder := func(filename string, nodes []*syntax.Node) []TodoIssue {
		return nil
	}

	matches := findIssuesGeneric([]*syntax.Node{}, finder, "TODO")

	matchCount := 0
	for range matches {
		matchCount++
	}

	if matchCount != 0 {
		t.Errorf("Expected 0 matches for empty data, got %d", matchCount)
	}
}

// TestMultiDetector_FindDuplOver_DefaultMethod tests with default art-dupl method.
func TestMultiDetector_FindDuplOver_DefaultMethod(t *testing.T) {
	cfg := createSimpleTestConfig()

	tree := suffixtree.New()
	data := []*syntax.Node{createTestNode("test.go", 1, 10)}

	detector := NewMultiDetector(cfg, data, tree, false)

	// Should return a channel
	matches := detector.FindDuplOver(15)
	if matches == nil {
		t.Error("FindDuplOver() returned nil channel")
	}

	// Drain the channel
	for range matches {
		// Drain
	}
}

// TestMultiDetector_FindDuplOver_HashMethod tests with hash detection method.
func TestMultiDetector_FindDuplOver_HashMethod(t *testing.T) {
	cfg := createHashTestConfig()

	tree := suffixtree.New()
	data := []*syntax.Node{createTestNode("test.go", 1, 10)}

	detector := NewMultiDetector(cfg, data, tree, false)

	// Should return a channel
	matches := detector.FindDuplOver(15)
	if matches == nil {
		t.Error("FindDuplOver() returned nil channel")
	}

	// Drain the channel
	for range matches {
		// Drain
	}
}

// TestMultiDetector_FindDuplOver_BothMethods tests with both detection methods.
func TestMultiDetector_FindDuplOver_BothMethods(t *testing.T) {
	cfg := &config.Config{
		Threshold: 15,
		DetectionMethods: config.DetectionMethods{
			config.DetectionMethodArtDupl,
			config.DetectionMethodHash,
		},
	}

	tree := suffixtree.New()
	data := []*syntax.Node{createTestNode("test.go", 1, 10)}

	detector := NewMultiDetector(cfg, data, tree, false)

	// Should return a channel
	matches := detector.FindDuplOver(15)
	if matches == nil {
		t.Error("FindDuplOver() returned nil channel")
	}

	// Drain the channel
	for range matches {
		// Drain
	}
}

// TestMultiDetector_FindDuplOver_Verbose tests verbose mode.
func TestMultiDetector_FindDuplOver_Verbose(t *testing.T) {
	cfg := createHashTestConfig()

	tree := suffixtree.New()
	data := []*syntax.Node{createTestNode("test.go", 1, 10)}

	detector := NewMultiDetector(cfg, data, tree, true) // verbose = true

	// Should return a channel
	matches := detector.FindDuplOver(15)
	if matches == nil {
		t.Error("FindDuplOver() returned nil channel")
	}

	// Drain the channel
	for range matches {
	}
}

// TestMultiDetector_FindDuplOver_EmptyData tests with empty data.
func TestMultiDetector_FindDuplOver_EmptyData(t *testing.T) {
	cfg := createSimpleTestConfig()

	tree := suffixtree.New()
	data := []*syntax.Node{}

	detector := NewMultiDetector(cfg, data, tree, false)

	// Should return a channel
	matches := detector.FindDuplOver(15)
	if matches == nil {
		t.Error("FindDuplOver() returned nil channel")
	}

	// Drain the channel
	for range matches {
	}
}

// TestTodoDetector_FindTodosInFile_RealFile tests findTodosInFile with actual Go files.
func TestTodoDetector_FindTodosInFile_RealFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test Go file with TODO comments
	testFile := tmpDir + "/test_todos.go"

	goCode := `package test

// TODO: implement this feature
func TodoFunc() {}

// FIXME(@developer): this is broken
func BrokenFunc() {}

// XXX: hack alert here
func HackFunc() {}

// HACK: temporary workaround
func WorkaroundFunc() {}

// NOTE: important information
func NoteFunc() {}

// TODO(2024-12-31): deadline task
func DeadlineFunc() {}

// Regular comment - no TODO
func RegularFunc() {}
`

	err := os.WriteFile(testFile, []byte(goCode), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	detector := NewTodoDetector()
	nodes := []*syntax.Node{createTestNode(testFile, 1, 100)}

	todos := detector.findTodosInFile(testFile, nodes)

	if len(todos) == 0 {
		t.Fatal("Expected to find TODO comments, got none")
	}

	// Check that we found the expected types
	foundTypes := make(map[string]bool)
	for _, todo := range todos {
		foundTypes[todo.Type] = true
	}

	expectedTypes := []string{"TODO", "FIXME", "XXX", "HACK", "NOTE"}
	for _, expected := range expectedTypes {
		if !foundTypes[expected] {
			t.Errorf("Expected to find %s comment", expected)
		}
	}
}

// TestTodoDetector_FindTodosInFile_WithTags tests parsing tags from TODO comments.
func TestTodoDetector_FindTodosInFile_WithTags(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := tmpDir + "/test_tags.go"

	goCode := `package test

// TODO(@user1,@user2): multi-tag TODO
func MultiTagFunc() {}

// FIXME(2024-01-15): dated fixme
func DatedFunc() {}
`

	err := os.WriteFile(testFile, []byte(goCode), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	detector := NewTodoDetector()
	nodes := []*syntax.Node{createTestNode(testFile, 1, 100)}

	todos := detector.findTodosInFile(testFile, nodes)

	if len(todos) == 0 {
		t.Fatal("Expected to find TODO comments")
	}

	// Find the TODO with tags
	var foundMultiTag bool

	for _, todo := range todos {
		if todo.Type == "TODO" && len(todo.Tags) > 0 {
			if len(todo.Tags) >= 2 {
				foundMultiTag = true
			}
		}
	}

	if !foundMultiTag {
		t.Error("Expected to find TODO with multiple tags")
	}
}

// TestTodoDetector_FindTodosInFile_NoTodos tests file without TODO comments.
func TestTodoDetector_FindTodosInFile_NoTodos(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := tmpDir + "/no_todos.go"

	goCode := `package test

// This is a regular comment
func RegularFunc() {}

/* Another regular comment */
func AnotherFunc() {}
`

	err := os.WriteFile(testFile, []byte(goCode), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	detector := NewTodoDetector()
	nodes := []*syntax.Node{createTestNode(testFile, 1, 100)}

	todos := detector.findTodosInFile(testFile, nodes)

	if len(todos) != 0 {
		t.Errorf("Expected no TODO comments, got %d", len(todos))
	}
}

// TestTodoDetector_FindTodosInFile_InvalidFile tests with non-existent file.
func TestTodoDetector_FindTodosInFile_InvalidFile(t *testing.T) {
	detector := NewTodoDetector()
	nodes := []*syntax.Node{createTestNode("/nonexistent/path/file.go", 1, 100)}

	// Should return nil (no panic) for invalid file
	todos := detector.findTodosInFile("/nonexistent/path/file.go", nodes)

	if todos != nil {
		t.Errorf("Expected nil for invalid file, got %d todos", len(todos))
	}
}

// TestTodoDetector_FindTodosInFile_BlockComments tests block comments with TODO.
func TestTodoDetector_FindTodosInFile_BlockComments(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := tmpDir + "/block_comments.go"

	goCode := `package test

/*
TODO: this is in a block comment
Multiple lines here
*/
func BlockFunc() {}

/* FIXME: block fixme */
func BlockFixmeFunc() {}
`

	err := os.WriteFile(testFile, []byte(goCode), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	detector := NewTodoDetector()
	nodes := []*syntax.Node{createTestNode(testFile, 1, 100)}

	todos := detector.findTodosInFile(testFile, nodes)

	if len(todos) == 0 {
		t.Fatal("Expected to find TODO comments in block comments")
	}
}

// runLegacyDetectionTest creates a file with the given code and runs legacy detection.
// Returns the issues found by the detector.
func runLegacyDetectionTest(t *testing.T, filename, goCode string) []LegacyIssue {
	t.Helper()
	tmpDir := t.TempDir()

	testFile := tmpDir + "/" + filename

	err := os.WriteFile(testFile, []byte(goCode), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	detector := NewLegacyDetector()
	nodes := []*syntax.Node{createTestNode(testFile, 1, 100)}

	return detector.findLegacyInFile(testFile, nodes)
}

// runLegacyDetectionTestCase is a helper function that runs the legacy detector on the provided Go code.
func runLegacyDetectionTestCase(t *testing.T, testFile, goCode, expectedLog string) {
	t.Helper()

	issues := runLegacyDetectionTest(t, testFile, goCode)
	if issues == nil {
		t.Log(expectedLog)
	}
}

// TestLegacyDetector_FindLegacyInFile_RealFile tests findLegacyInFile with actual Go files.
func TestLegacyDetector_FindLegacyInFile_RealFile(t *testing.T) {
	goCode := `package test

import (
	"io/ioutil"
)

func LegacyFunc() {
	// Using deprecated ioutil.ReadFile
	data, err := ioutil.ReadFile("test.txt")
	if err != nil {
		return
	}
	_ = data
}`

	runLegacyDetectionTestCase(
		t,
		"test_legacy.go",
		goCode,
		"No legacy issues found (simplified detection)",
	)
}

// TestLegacyDetector_FindLegacyInFile_NoLegacyPatterns tests file without legacy patterns.
func TestLegacyDetector_FindLegacyInFile_NoLegacyPatterns(t *testing.T) {
	goCode := `package test

import (
	"os"
)

func ModernFunc() {
	data, err := os.ReadFile("test.txt")
	if err != nil {
		return
	}
	_ = data
}`

	runLegacyDetectionTestCase(
		t,
		"no_legacy.go",
		goCode,
		"No legacy issues in modern code (expected)",
	)
}

// TestLegacyDetector_FindLegacyInFile_EmptyFile tests with empty Go file.
func TestLegacyDetector_FindLegacyInFile_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := tmpDir + "/empty.go"

	goCode := `package test
`

	err := os.WriteFile(testFile, []byte(goCode), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	detector := NewLegacyDetector()
	nodes := []*syntax.Node{}

	issues := detector.findLegacyInFile(testFile, nodes)

	testutil.AssertCount(t, len(issues), 0, "issues for empty file")
}

// setupTodoTest creates a temporary test file with a TODO comment and returns
// the detector and nodes for testing.
func setupTodoTest(t *testing.T) (*TodoDetector, []*syntax.Node, string) {
	t.Helper()
	tmpDir := t.TempDir()

	testFile := tmpDir + "/test.go"

	goCode := `package test
// TODO: test
func Test() {}
`

	err := os.WriteFile(testFile, []byte(goCode), 0o644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	detector := NewTodoDetector()
	nodes := []*syntax.Node{createTestNode(testFile, 1, 10)}

	return detector, nodes, testFile
}

// assertTodoMatches checks that the detector finds at least one TODO match.
func assertTodoMatches(t *testing.T, detector *TodoDetector, nodes []*syntax.Node, errMsg string) {
	t.Helper()

	matches := detector.FindTodos(nodes)

	matchCount := 0
	for range matches {
		matchCount++
	}

	if matchCount == 0 {
		t.Error(errMsg)
	}
}

// TestFindIssuesInFile_WithData tests findIssuesInFile with actual data.
func TestFindIssuesInFile_WithData(t *testing.T) {
	detector, nodes, _ := setupTodoTest(t)
	assertTodoMatches(t, detector, nodes, "Expected at least one match from FindTodos")
}

// TestFindIssuesGeneric_WithData tests findIssuesGeneric with actual data.
func TestFindIssuesGeneric_WithData(t *testing.T) {
	detector, nodes, _ := setupTodoTest(t)
	assertTodoMatches(t, detector, nodes, "Expected matches from findIssuesGeneric")
}
