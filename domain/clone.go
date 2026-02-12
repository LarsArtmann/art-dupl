// Package domain provides core domain types for clone detection.
//
// TODO: ARCHITECTURE - This file is 495 lines and exceeds 350 line limit.
// Should be split into focused files for better maintainability:
// - domain/types_enums.go: Enum types (FileProcessingState, DetectionState, AnalysisMode, CloneSeverity)
// - domain/clone.go: Clone and CloneGroup domain objects with validation
// - domain/analysis.go: Analysis and AnalysisStats types
// - domain/repository.go: Repository and SourceFile types
// - domain/options.go: DetectionOptions and configuration types
// - domain/conversion.go: NodeToClone, CalculateSeverity, calculateComplexity
// - domain/validation.go: Validation helpers (validationRule, validateRules, isValid*)
package domain

import (
	"crypto/sha256"
	stderrors "errors"
	"fmt"
	"strings"

	"github.com/LarsArtmann/art-dupl/pkg/position"
	"github.com/LarsArtmann/art-dupl/syntax"
)

// FileProcessingState represents the processing state of a file.
type FileProcessingState string

const (
	// FileProcessingStatePending indicates file is pending processing.
	FileProcessingStatePending FileProcessingState = "pending"
	// FileProcessingStateProcessing indicates file is currently being processed.
	FileProcessingStateProcessing FileProcessingState = "processing"
	// FileProcessingStateCompleted indicates file processing is complete.
	FileProcessingStateCompleted FileProcessingState = "completed"
	// FileProcessingStateFailed indicates file processing failed.
	FileProcessingStateFailed FileProcessingState = "failed"
)

// String implements fmt.Stringer for FileProcessingState.
func (fps FileProcessingState) String() string {
	return string(fps)
}

// IsValid checks if FileProcessingState is valid.
func (fps FileProcessingState) IsValid() bool {
	switch fps {
	case FileProcessingStatePending, FileProcessingStateProcessing, FileProcessingStateCompleted, FileProcessingStateFailed:
		return true
	default:
		return false
	}
}

// DetectionState represents the state of clone detection.
type DetectionState string

const (
	// DetectionStateIdle indicates no detection in progress.
	DetectionStateIdle DetectionState = "idle"
	// DetectionStateRunning indicates detection is running.
	DetectionStateRunning DetectionState = "running"
	// DetectionStateCompleted indicates detection completed.
	DetectionStateCompleted DetectionState = "completed"
	// DetectionStateFailed indicates detection failed.
	DetectionStateFailed DetectionState = "failed"
)

// String implements fmt.Stringer for DetectionState.
func (ds DetectionState) String() string {
	return string(ds)
}

// IsValid checks if DetectionState is valid.
func (ds DetectionState) IsValid() bool {
	switch ds {
	case DetectionStateIdle, DetectionStateRunning, DetectionStateCompleted, DetectionStateFailed:
		return true
	default:
		return false
	}
}

// AnalysisMode represents the mode of code analysis.
type AnalysisMode string

const (
	// AnalysisModeFull performs full analysis.
	AnalysisModeFull AnalysisMode = "full"
	// AnalysisModeQuick performs quick analysis.
	AnalysisModeQuick AnalysisMode = "quick"
	// AnalysisModeDeep performs deep analysis.
	AnalysisModeDeep AnalysisMode = "deep"
)

// String implements fmt.Stringer for AnalysisMode.
func (am AnalysisMode) String() string {
	return string(am)
}

// IsValid checks if AnalysisMode is valid.
func (am AnalysisMode) IsValid() bool {
	switch am {
	case AnalysisModeFull, AnalysisModeQuick, AnalysisModeDeep:
		return true
	default:
		return false
	}
}

// Clone represents a code clone with strong typing.
//
// Memory Layout Optimized: 8B fields grouped together, 16B fields at end
// to minimize padding and improve cache locality.
//
// Note: Clone intentionally omits an ID field. Identity is determined by the
// combination of Filename, StartLine, EndLine, and Hash. This eliminates
// unnecessary storage overhead and avoids cargo-cult "entities need IDs" patterns.
type Clone struct {
	// 8B fields (grouped for cache efficiency)
	StartLine  LineNumber      `json:"startLine"`
	EndLine    LineNumber      `json:"endLine"`
	StartPos   BytePosition    `json:"startPos"`
	EndPos     BytePosition    `json:"endPos"`
	Complexity ComplexityScore `json:"complexity"`
	// StringIDs (4B each, interned for memory efficiency)
	Filename StringID            `json:"filename"`
	Fragment StringID            `json:"fragment"`
	Hash     StringID            `json:"hash"`
	Status   FileProcessingState `json:"status"`
}

// IsValid validates clone data.
// Note: Domain types handle their own validation (e.g., non-empty strings, non-zero line numbers).
// This method only validates cross-field relationships and enum types.
func (c Clone) IsValid() error {
	// Cross-field validation: end must be >= start
	if c.EndLine < c.StartLine {
		return stderrors.New("clone end line must be >= start line")
	}
	// Only validate positions if both are set (non-zero)
	if c.StartPos > 0 || c.EndPos > 0 {
		if c.StartPos >= c.EndPos {
			return stderrors.New("clone end position must be > start position")
		}
	}

	// Enum type validation
	if !c.Status.IsValid() {
		return fmt.Errorf("invalid clone processing state: %s", c.Status)
	}

	return nil
}

// FilenameString returns the filename as a string by looking up the StringID.
func (c Clone) FilenameString() string {
	return GlobalPool().Lookup(c.Filename)
}

// FragmentString returns the fragment as a string by looking up the StringID.
func (c Clone) FragmentString() string {
	return GlobalPool().Lookup(c.Fragment)
}

// HashString returns the hash as a string by looking up the StringID.
func (c Clone) HashString() string {
	return GlobalPool().Lookup(c.Hash)
}

// SetFilename interns a filename string and sets the FilenameID.
func (c *Clone) SetFilename(filename string) {
	c.Filename = GlobalPool().Intern(filename)
}

// SetFragment interns a fragment string and sets the FragmentID.
func (c *Clone) SetFragment(fragment string) {
	c.Fragment = GlobalPool().Intern(fragment)
}

// SetHash interns a hash string and sets the HashID.
func (c *Clone) SetHash(hash string) {
	c.Hash = GlobalPool().Intern(hash)
}

// CloneGroup represents a group of clones.
type CloneGroup struct {
	ID       string              `json:"id"`
	Clones   []Clone             `json:"clones"`
	Hash     string              `json:"hash"`
	Size     uint                `json:"size"`
	Severity CloneSeverity       `json:"severity"`
	Status   FileProcessingState `json:"status"`
}

// IsValid validates clone group.
func (cg CloneGroup) IsValid() error {
	if len(cg.Clones) == 0 {
		return stderrors.New("clone group must have at least one clone")
	}
	if !cg.Severity.IsValid() {
		return fmt.Errorf("invalid clone severity: %s", cg.Severity)
	}
	if !cg.Status.IsValid() {
		return fmt.Errorf("invalid clone group status: %s", cg.Status)
	}

	// Validate all clones in group
	for i, clone := range cg.Clones {
		if err := clone.IsValid(); err != nil {
			return fmt.Errorf("clone %d in group is invalid: %w", i, err)
		}
	}
	return nil
}

// CloneSeverity represents clone severity levels.
type CloneSeverity string

const (
	CloneSeverityLow      CloneSeverity = "low"
	CloneSeverityMedium   CloneSeverity = "medium"
	CloneSeverityHigh     CloneSeverity = "high"
	CloneSeverityCritical CloneSeverity = "critical"
)

// IsValid checks if severity is valid.
func (cs CloneSeverity) IsValid() bool {
	switch cs {
	case CloneSeverityLow, CloneSeverityMedium, CloneSeverityHigh, CloneSeverityCritical:
		return true
	default:
		return false
	}
}

// String implements fmt.Stringer.
func (cs CloneSeverity) String() string {
	return string(cs)
}

// MarshalJSON implements json.Marshaler.
func (cs CloneSeverity) MarshalJSON() ([]byte, error) {
	if !cs.IsValid() {
		return nil, fmt.Errorf("invalid clone severity: %s", cs)
	}
	return []byte(`"` + string(cs) + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (cs *CloneSeverity) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	severity := CloneSeverity(str)
	if !severity.IsValid() {
		return fmt.Errorf("invalid clone severity: %s", str)
	}
	*cs = severity
	return nil
}

// Analysis represents main analysis domain object.
type Analysis struct {
	ID          string            `json:"id"`
	State       DetectionState    `json:"state"`
	Mode        AnalysisMode      `json:"mode"`
	Threshold   uint              `json:"threshold"`
	CloneGroups []CloneGroup      `json:"cloneGroups"`
	Stats       AnalysisStats     `json:"stats"`
	Metadata    map[string]string `json:"metadata"`
	CreatedAt   string            `json:"createdAt"`
	CompletedAt *string           `json:"completedAt,omitempty"`
}

// IsValid validates analysis.
func (a Analysis) IsValid() error {
	if !a.State.IsValid() {
		return fmt.Errorf("invalid analysis state: %s", a.State)
	}
	if !a.Mode.IsValid() {
		return fmt.Errorf("invalid analysis mode: %s", a.Mode)
	}
	if a.Threshold == 0 {
		return stderrors.New("analysis threshold cannot be zero")
	}
	if a.CreatedAt == "" {
		return stderrors.New("analysis created at cannot be empty")
	}

	// Validate all clone groups
	for i, group := range a.CloneGroups {
		if err := group.IsValid(); err != nil {
			return fmt.Errorf("clone group %d in analysis is invalid: %w", i, err)
		}
	}
	return nil
}

// validationRule represents a single validation rule.
// TODO: REFACTOR - Lines 285-299 are internal validation helpers that are used across multiple types.
// Consider moving to domain/validation.go or making them more generic.
// Current pattern: validationRule struct + validateRules function + isValid* functions for each type.
// Could consolidate into a more generic validation framework.
type validationRule struct {
	valid bool
	msg   string
}

// validateRules checks all validation rules and returns the first error encountered.
func validateRules(rules []validationRule) error {
	for _, rule := range rules {
		if !rule.valid {
			return stderrors.New(rule.msg)
		}
	}
	return nil
}

// AnalysisStats represents analysis statistics.
type AnalysisStats struct {
	FilesAnalyzed    uint    `json:"filesAnalyzed"`
	TotalClones      uint    `json:"totalClones"`
	TotalTokenSize   uint    `json:"totalTokenSize"`
	ComplexityScore  float64 `json:"complexityScore"`
	DuplicationRatio float64 `json:"duplicationRatio"`
	ProcessingTime   uint    `json:"processingTime"`
}

// isValidAnalysisStats validates analysis stats.
func isValidAnalysisStats(as AnalysisStats) error {
	rules := []validationRule{
		{as.FilesAnalyzed > 0, "files analyzed cannot be zero"},
		{as.ProcessingTime > 0, "processing time cannot be zero"},
		{as.ComplexityScore >= 0, "complexity score cannot be negative"},
		{as.DuplicationRatio >= 0, "duplication ratio cannot be negative"},
	}
	return validateRules(rules)
}

// IsValid validates analysis stats.
func (as AnalysisStats) IsValid() error {
	return isValidAnalysisStats(as)
}

// Repository represents source code repository.
type Repository struct {
	Path        string       `json:"path"`
	Name        string       `json:"name"`
	Files       []SourceFile `json:"files"`
	Language    string       `json:"language"`
	Framework   string       `json:"framework"`
	Size        uint64       `json:"size"`
	LastIndexed string       `json:"lastIndexed"`
}

// IsValid validates repository.
func (r Repository) IsValid() error {
	if r.Path == "" {
		return stderrors.New("repository path cannot be empty")
	}
	if r.Name == "" {
		return stderrors.New("repository name cannot be empty")
	}
	if r.Language == "" {
		return stderrors.New("repository language cannot be empty")
	}
	return nil
}

// SourceFile represents a source code file.
type SourceFile struct {
	Path         string   `json:"path"`
	Name         string   `json:"name"`
	Size         uint64   `json:"size"`
	Language     string   `json:"language"`
	Content      string   `json:"content,omitempty"`
	Hash         string   `json:"hash"`
	Dependencies []string `json:"dependencies,omitempty"`
}

// isValidSourceFile validates source file.
func isValidSourceFile(sf SourceFile) error {
	rules := []validationRule{
		{sf.Path != "", "source file path cannot be empty"},
		{sf.Name != "", "source file name cannot be empty"},
		{sf.Size > 0, "source file size cannot be zero"},
		{sf.Hash != "", "source file hash cannot be empty"},
	}
	return validateRules(rules)
}

// IsValid validates source file.
func (sf SourceFile) IsValid() error {
	return isValidSourceFile(sf)
}

// DetectionOptions represents configuration for detection.
type DetectionOptions struct {
	Threshold     uint         `json:"threshold"`
	Mode          AnalysisMode `json:"mode"`
	IncludeVendor bool         `json:"includeVendor"`
	Verbose       bool         `json:"verbose"`
	Paths         []string     `json:"paths"`
	OutputFormat  string       `json:"outputFormat"`
}

// IsValid validates detection options.
func (do DetectionOptions) IsValid() error {
	if do.Threshold == 0 {
		return stderrors.New("threshold must be > 0")
	}
	if !do.Mode.IsValid() {
		return fmt.Errorf("invalid analysis mode: %s", do.Mode)
	}
	if len(do.Paths) == 0 {
		return stderrors.New("at least one path must be specified")
	}
	return nil
}

// NodeToClone converts syntax nodes to domain Clone.
//
// TODO: TYPE SAFETY ISSUE - This function accepts primitive types (string, []byte)
// but creates domain types internally. Consider:
// - Accept domain.Filepath instead of string for filename
// - Return (Clone, error) instead of Clone to handle validation failures
// - Use domain types throughout the conversion pipeline
//
// Also: This function does multiple things (line calculation, fragment extraction,
// hash generation, complexity calculation). Consider breaking into smaller functions
// for better testability.
func NodeToClone(node *syntax.Node, filename string, fileContent []byte) Clone {
	// Calculate line numbers from file content
	lineStart, lineEnd := 1, 1 // defaults
	if fileContent != nil {
		lineStart, lineEnd = position.ByteRangeToLines(fileContent, int(node.Pos), int(node.End))
	}

	// Extract fragment from file content
	var fragment string
	if fileContent != nil {
		start := node.Pos
		end := node.End
		if start >= 0 && int(end) <= len(fileContent) && start < end {
			fragment = string(fileContent[start:end])
		}
	}

	// Generate hash from fragment
	hashStr := ""
	if fragment != "" {
		hashStr = fmt.Sprintf("%x", sha256.Sum256([]byte(fragment)))
	}

	// Create domain types from primitive values
	startLn, _ := NewLineNumber(uint16(lineStart)) //nolint:gosec //G115 lineStart >= 1 guaranteed by initialize default
	endLn, _ := NewLineNumber(uint16(lineEnd))     //nolint:gosec //G115 lineEnd >= 1 guaranteed by initialize default
	startPos := NewBytePosition(uint32(node.Pos))  //nolint:gosec //G115 node.Pos validated >= 0 in fragment extraction
	endPos := NewBytePosition(uint32(node.End))    //nolint:gosec //G115 node.End validated >= 0 in fragment extraction
	complexity := NewComplexityScore(uint16(calculateComplexity(node)))

	// Create clone with interned strings
	clone := Clone{
		StartLine:  startLn,
		EndLine:    endLn,
		StartPos:   startPos,
		EndPos:     endPos,
		Complexity: complexity,
		Status:     FileProcessingStateCompleted,
	}

	// Intern strings for memory efficiency
	clone.SetFilename(filename)
	clone.SetFragment(fragment)
	clone.SetHash(hashStr)

	return clone
}

// CalculateSeverity determines clone severity based on size and complexity.
// TODO: MOVE TO SEPARATE FILE - Lines 462-495 contain conversion/calculations functions
// (CalculateSeverity, calculateComplexity) that are not domain types.
// Move to domain/conversion.go or domain/severity.go for better separation of concerns.
func CalculateSeverity(size, complexity uint) CloneSeverity {
	if complexity > 50 || size > 200 {
		return CloneSeverityCritical
	}
	if complexity > 20 || size > 100 {
		return CloneSeverityHigh
	}
	if complexity > 10 || size > 50 {
		return CloneSeverityMedium
	}
	return CloneSeverityLow
}

// calculateComplexity calculates a basic complexity metric for a node.
// TODO: IMPROVE ALGORITHM - Current implementation (lines 477-494) is very basic:
// - Only counts nodes and adds small increments based on type
// - Recursive implementation could be expensive for deep trees
// - Node.Type == 0 magic number without documentation
// Consider using established complexity metrics like:
// - McCabe cyclomatic complexity
// - Nesting depth
// - Cognitive complexity
func calculateComplexity(node *syntax.Node) uint {
	// Simple complexity based on node count and depth
	complexity := uint(1) // Base complexity

	// Add complexity for each child
	for _, child := range node.Children {
		complexity += calculateComplexity(child)
	}

	// Add complexity based on node value (type)
	switch node.Type {
	case 0: // Assume 0 might be a control structure
		complexity += 2
	default:
		complexity += 1
	}

	return complexity
}
