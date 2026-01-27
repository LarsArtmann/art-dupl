package domain

import (
	"crypto/sha256"
	"errors"
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
	Confidence Confidence      `json:"confidence"`
	Complexity ComplexityScore `json:"complexity"`
	// 16B string headers (at end to minimize padding)
	Filename Filepath            `json:"filename"`
	Fragment string              `json:"fragment"`
	Hash     Hash                `json:"hash"`
	Status   FileProcessingState `json:"status"`
}

// IsValid validates clone data.
// Note: Domain types handle their own validation (e.g., non-empty strings, non-zero line numbers).
// This method only validates cross-field relationships and enum types.
func (c Clone) IsValid() error {
	// Cross-field validation: end must be >= start
	if c.EndLine < c.StartLine {
		return errors.New("clone end line must be >= start line")
	}
	// Only validate positions if both are set (non-zero)
	if c.StartPos > 0 || c.EndPos > 0 {
		if c.StartPos >= c.EndPos {
			return errors.New("clone end position must be > start position")
		}
	}

	// Enum type validation
	if !c.Status.IsValid() {
		return fmt.Errorf("invalid clone processing state: %s", c.Status)
	}

	return nil
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
		return errors.New("clone group must have at least one clone")
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
		return errors.New("analysis threshold cannot be zero")
	}
	if a.CreatedAt == "" {
		return errors.New("analysis created at cannot be empty")
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
type validationRule struct {
	valid bool
	msg   string
}

// validateRules checks all validation rules and returns the first error encountered.
func validateRules(rules []validationRule) error {
	for _, rule := range rules {
		if !rule.valid {
			return errors.New(rule.msg)
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
		return errors.New("repository path cannot be empty")
	}
	if r.Name == "" {
		return errors.New("repository name cannot be empty")
	}
	if r.Language == "" {
		return errors.New("repository language cannot be empty")
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
		return errors.New("threshold must be > 0")
	}
	if !do.Mode.IsValid() {
		return fmt.Errorf("invalid analysis mode: %s", do.Mode)
	}
	if len(do.Paths) == 0 {
		return errors.New("at least one path must be specified")
	}
	return nil
}

// NodeToClone converts syntax nodes to domain Clone.
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
	hash, _ := NewHash(hashStr)

	// Create domain types from primitive values
	fp, _ := NewFilepath(filename)
	startLn, _ := NewLineNumber(uint(lineStart)) //nolint:gosec //G115 lineStart >= 1 guaranteed by initialize default
	endLn, _ := NewLineNumber(uint(lineEnd))     //nolint:gosec //G115 lineEnd >= 1 guaranteed by initialize default
	startPos := NewBytePosition(uint(node.Pos))  //nolint:gosec //G115 node.Pos validated >= 0 in fragment extraction
	endPos := NewBytePosition(uint(node.End))    //nolint:gosec //G115 node.End validated >= 0 in fragment extraction
	conf, _ := NewConfidence(1.0)                // Calculate actual confidence
	complexity := NewComplexityScore(calculateComplexity(node))

	return Clone{
		Filename:   fp,
		StartLine:  startLn,
		EndLine:    endLn,
		StartPos:   startPos,
		EndPos:     endPos,
		Fragment:   fragment,
		Hash:       hash,
		Confidence: conf,
		Complexity: complexity,
		Status:     FileProcessingStateCompleted,
	}
}

// CalculateSeverity determines clone severity based on size and complexity.
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
