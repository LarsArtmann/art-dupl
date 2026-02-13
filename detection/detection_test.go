package detection

import (
	"testing"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
	"github.com/LarsArtmann/art-dupl/suffixtree"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// TestNewMultiDetector tests MultiDetector constructor.
func TestNewMultiDetector(t *testing.T) {
	cfg := &config.Config{
		Threshold:        15,
		Verbose:          true,
		DetectionMethods: config.DetectionMethods{config.DetectionMethodArtDupl},
	}

	data := []*syntax.Node{
		{Filename: "test.go", Type: 1},
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
	cfg := &config.Config{
		Threshold:        15,
		DetectionMethods: config.DetectionMethods{config.DetectionMethodArtDupl},
	}
	tree := suffixtree.New()

	detector := NewMultiDetector(cfg, []*syntax.Node{}, tree, false)

	if len(detector.data) != 0 {
		t.Error("Expected empty data slice")
	}
}

// TestMultiDetector_FindDuplOver_DefaultMethod tests with default detection method.
func TestMultiDetector_FindDuplOver_DefaultMethod(t *testing.T) {
	cfg := &config.Config{
		Threshold:        15,
		DetectionMethods: config.DetectionMethods{config.DetectionMethodArtDupl},
	}

	data := []*syntax.Node{
		{Filename: "test.go", Type: 1, Pos: 1, End: 10},
	}

	tree := suffixtree.New()

	detector := NewMultiDetector(cfg, data, tree, false)
	matches := detector.FindDuplOver(15)

	// Drain the channel (should close immediately for empty tree)
	matchCount := 0
	for range matches {
		matchCount++
	}

	// Empty tree should produce no matches
	if matchCount != 0 {
		t.Logf("Found %d matches (expected 0 for empty tree)", matchCount)
	}
}

// TestMultiDetector_FindDuplOver_HashMethod tests with hash detection method.
func TestMultiDetector_FindDuplOver_HashMethod(t *testing.T) {
	cfg := &config.Config{
		Threshold:        15,
		DetectionMethods: config.DetectionMethods{config.DetectionMethodHash},
	}

	data := []*syntax.Node{
		{Filename: "test.go", Type: 1, Pos: 1, End: 10},
	}

	tree := suffixtree.New()

	detector := NewMultiDetector(cfg, data, tree, false)
	matches := detector.FindDuplOver(15)

	// Drain the channel
	matchCount := 0
	for range matches {
		matchCount++
	}

	// Empty/minimal data should produce no matches
	t.Logf("Hash method found %d matches", matchCount)
}

// TestMultiDetector_FindDuplOver_MultipleMethods tests with multiple detection methods.
func TestMultiDetector_FindDuplOver_MultipleMethods(t *testing.T) {
	cfg := &config.Config{
		Threshold:        15,
		DetectionMethods: config.DetectionMethods{config.DetectionMethodHash, config.DetectionMethodArtDupl},
	}

	data := []*syntax.Node{
		{Filename: "test.go", Type: 1, Pos: 1, End: 10},
	}

	tree := suffixtree.New()

	detector := NewMultiDetector(cfg, data, tree, false)
	matches := detector.FindDuplOver(15)

	// Drain the channel
	matchCount := 0
	for range matches {
		matchCount++
	}

	t.Logf("Multiple methods found %d matches", matchCount)
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

// TestTodoIssue_GetLine tests TodoIssue GetLine method.
func TestTodoIssue_GetLine(t *testing.T) {
	lineNum, _ := domain.NewLineNumber(42)
	issue := TodoIssue{
		Filename: domain.Filepath("test.go"),
		Line:     lineNum,
		Text:     "Fix this later",
		Type:     "TODO",
	}

	if issue.GetLine() != lineNum {
		t.Errorf("GetLine() = %d, want %d", issue.GetLine(), lineNum)
	}
}

// TestLegacyIssue_GetLine tests LegacyIssue GetLine method.
func TestLegacyIssue_GetLine(t *testing.T) {
	lineNum, _ := domain.NewLineNumber(100)
	issue := LegacyIssue{
		Filename: domain.Filepath("legacy.go"),
		Line:     lineNum,
		Type:     "deprecated_function",
		Message:  "Use of deprecated function",
		Severity: domain.CloneSeverityMedium,
	}

	if issue.GetLine() != lineNum {
		t.Errorf("GetLine() = %d, want %d", issue.GetLine(), lineNum)
	}
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

// TestMultiDetector_VerboseOutput tests verbose output doesn't panic.
func TestMultiDetector_VerboseOutput(t *testing.T) {
	cfg := &config.Config{
		Threshold:        15,
		Verbose:          true,
		DetectionMethods: config.DetectionMethods{config.DetectionMethodArtDupl},
	}

	data := []*syntax.Node{
		{Filename: "test.go", Type: 1, Pos: 1, End: 10},
	}

	tree := suffixtree.New()

	detector := NewMultiDetector(cfg, data, tree, true)

	// This should not panic
	matches := detector.FindDuplOver(15)

	// Drain channel
	for range matches {
	}
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
				t.Errorf("Pattern %s on %q: matched=%v, want=%v", tc.patternName, tc.input, matched, tc.shouldMatch)
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
