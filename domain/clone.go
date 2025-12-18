package domain

import (
	"fmt"
	"strings"

	"github.com/LarsArtmann/art-dupl/syntax"
	"github.com/LarsArtmann/art-dupl/types"
)

// Clone represents a code clone with strong typing
type Clone struct {
	ID         string                    `json:"id"`
	Filename   string                    `json:"filename"`
	StartLine  uint                      `json:"startLine"`
	EndLine    uint                      `json:"endLine"`
	StartPos   uint                      `json:"startPos"`
	EndPos     uint                      `json:"endPos"`
	Fragment   string                    `json:"fragment"`
	Hash       string                    `json:"hash"`
	Confidence float64                   `json:"confidence"`
	Complexity uint                      `json:"complexity"`
	Status     types.FileProcessingState `json:"status"`
}

// IsValid validates clone data
func (c Clone) IsValid() error {
	if c.ID == "" {
		return fmt.Errorf("clone ID cannot be empty")
	}
	if c.Filename == "" {
		return fmt.Errorf("clone filename cannot be empty")
	}
	if c.StartLine == 0 {
		return fmt.Errorf("clone start line cannot be zero")
	}
	if c.EndLine < c.StartLine {
		return fmt.Errorf("clone end line must be >= start line")
	}
	if c.StartPos >= c.EndPos {
		return fmt.Errorf("clone end position must be > start position")
	}
	if !c.Status.IsValid() {
		return fmt.Errorf("invalid clone processing state: %s", c.Status)
	}
	if c.Confidence < 0 || c.Confidence > 1 {
		return fmt.Errorf("clone confidence must be between 0 and 1")
	}
	return nil
}

// CloneGroup represents a group of clones
type CloneGroup struct {
	ID       string                    `json:"id"`
	Clones   []Clone                   `json:"clones"`
	Hash     string                    `json:"hash"`
	Size     uint                      `json:"size"`
	Severity CloneSeverity             `json:"severity"`
	Status   types.FileProcessingState `json:"status"`
}

// IsValid validates clone group
func (cg CloneGroup) IsValid() error {
	if cg.ID == "" {
		return fmt.Errorf("clone group ID cannot be empty")
	}
	if len(cg.Clones) == 0 {
		return fmt.Errorf("clone group must have at least one clone")
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
			return fmt.Errorf("clone %d in group %s is invalid: %w", i, cg.ID, err)
		}
	}
	return nil
}

// CloneSeverity represents clone severity levels
type CloneSeverity string

const (
	CloneSeverityLow      CloneSeverity = "low"
	CloneSeverityMedium   CloneSeverity = "medium"
	CloneSeverityHigh     CloneSeverity = "high"
	CloneSeverityCritical CloneSeverity = "critical"
)

// IsValid checks if severity is valid
func (cs CloneSeverity) IsValid() bool {
	switch cs {
	case CloneSeverityLow, CloneSeverityMedium, CloneSeverityHigh, CloneSeverityCritical:
		return true
	default:
		return false
	}
}

// String implements fmt.Stringer
func (cs CloneSeverity) String() string {
	return string(cs)
}

// MarshalJSON implements json.Marshaler
func (cs CloneSeverity) MarshalJSON() ([]byte, error) {
	if !cs.IsValid() {
		return nil, fmt.Errorf("invalid clone severity: %s", cs)
	}
	return []byte(`"` + string(cs) + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler
func (cs *CloneSeverity) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	severity := CloneSeverity(str)
	if !severity.IsValid() {
		return fmt.Errorf("invalid clone severity: %s", str)
	}
	*cs = severity
	return nil
}

// Analysis represents main analysis domain object
type Analysis struct {
	ID          string               `json:"id"`
	State       types.DetectionState `json:"state"`
	Mode        types.AnalysisMode   `json:"mode"`
	Threshold   uint                 `json:"threshold"`
	CloneGroups []CloneGroup         `json:"cloneGroups"`
	Stats       AnalysisStats        `json:"stats"`
	Metadata    map[string]string    `json:"metadata"`
	CreatedAt   string               `json:"createdAt"`
	CompletedAt *string              `json:"completedAt,omitempty"`
}

// IsValid validates analysis
func (a Analysis) IsValid() error {
	if a.ID == "" {
		return fmt.Errorf("analysis ID cannot be empty")
	}
	if !a.State.IsValid() {
		return fmt.Errorf("invalid analysis state: %s", a.State)
	}
	if !a.Mode.IsValid() {
		return fmt.Errorf("invalid analysis mode: %s", a.Mode)
	}
	if a.Threshold == 0 {
		return fmt.Errorf("analysis threshold cannot be zero")
	}
	if a.CreatedAt == "" {
		return fmt.Errorf("analysis created at cannot be empty")
	}

	// Validate all clone groups
	for i, group := range a.CloneGroups {
		if err := group.IsValid(); err != nil {
			return fmt.Errorf("clone group %d in analysis %s is invalid: %w", i, a.ID, err)
		}
	}
	return nil
}

// AnalysisStats represents analysis statistics
type AnalysisStats struct {
	FilesAnalyzed    uint    `json:"filesAnalyzed"`
	TotalClones      uint    `json:"totalClones"`
	TotalTokenSize   uint    `json:"totalTokenSize"`
	ComplexityScore  float64 `json:"complexityScore"`
	DuplicationRatio float64 `json:"duplicationRatio"`
	ProcessingTime   uint    `json:"processingTime"`
}

// IsValid validates analysis stats
func (as AnalysisStats) IsValid() error {
	if as.FilesAnalyzed == 0 {
		return fmt.Errorf("files analyzed cannot be zero")
	}
	if as.ProcessingTime == 0 {
		return fmt.Errorf("processing time cannot be zero")
	}
	if as.ComplexityScore < 0 {
		return fmt.Errorf("complexity score cannot be negative")
	}
	if as.DuplicationRatio < 0 {
		return fmt.Errorf("duplication ratio cannot be negative")
	}
	return nil
}

// Repository represents source code repository
type Repository struct {
	Path        string       `json:"path"`
	Name        string       `json:"name"`
	Files       []SourceFile `json:"files"`
	Language    string       `json:"language"`
	Framework   string       `json:"framework"`
	Size        uint64       `json:"size"`
	LastIndexed string       `json:"lastIndexed"`
}

// IsValid validates repository
func (r Repository) IsValid() error {
	if r.Path == "" {
		return fmt.Errorf("repository path cannot be empty")
	}
	if r.Name == "" {
		return fmt.Errorf("repository name cannot be empty")
	}
	if r.Language == "" {
		return fmt.Errorf("repository language cannot be empty")
	}
	return nil
}

// SourceFile represents a source code file
type SourceFile struct {
	Path         string   `json:"path"`
	Name         string   `json:"name"`
	Size         uint64   `json:"size"`
	Language     string   `json:"language"`
	Content      string   `json:"content,omitempty"`
	Hash         string   `json:"hash"`
	Dependencies []string `json:"dependencies,omitempty"`
}

// IsValid validates source file
func (sf SourceFile) IsValid() error {
	if sf.Path == "" {
		return fmt.Errorf("source file path cannot be empty")
	}
	if sf.Name == "" {
		return fmt.Errorf("source file name cannot be empty")
	}
	if sf.Size == 0 {
		return fmt.Errorf("source file size cannot be zero")
	}
	if sf.Hash == "" {
		return fmt.Errorf("source file hash cannot be empty")
	}
	return nil
}

// DetectionOptions represents configuration for detection
type DetectionOptions struct {
	Threshold     uint               `json:"threshold"`
	Mode          types.AnalysisMode `json:"mode"`
	IncludeVendor bool               `json:"includeVendor"`
	Verbose       bool               `json:"verbose"`
	Paths         []string           `json:"paths"`
	OutputFormat  string             `json:"outputFormat"`
}

// IsValid validates detection options
func (do DetectionOptions) IsValid() error {
	if do.Threshold == 0 {
		return fmt.Errorf("threshold must be > 0")
	}
	if !do.Mode.IsValid() {
		return fmt.Errorf("invalid analysis mode: %s", do.Mode)
	}
	if len(do.Paths) == 0 {
		return fmt.Errorf("at least one path must be specified")
	}
	return nil
}

// NodeToClone converts syntax nodes to domain Clone
func NodeToClone(node *syntax.Node, filename string) Clone {
	// Generate unique ID for clone
	cloneID := fmt.Sprintf("%s-%d-%d", filename, node.Pos, node.End)

	return Clone{
		ID:         cloneID,
		Filename:   filename,
		StartLine:  uint(node.LineStart),
		EndLine:    uint(node.LineEnd),
		StartPos:   uint(node.Pos),
		EndPos:     uint(node.End),
		Fragment:   strings.Join(node.Fragments, ""),
		Hash:       string(node.Hash),
		Confidence: 1.0, // TODO: Calculate actual confidence
		Complexity: uint(node.Complexity),
		Status:     types.FileProcessingStateCompleted,
	}
}

// CalculateSeverity determines clone severity based on size and complexity
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
