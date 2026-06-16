package domain

import "fmt"

// FindingType represents the kind of code-quality finding.
type FindingType string

const (
	FindingTypeTodo   FindingType = "todo"
	FindingTypeLegacy FindingType = "legacy"
)

// IsValid returns true if the finding type is one of the defined constants.
func (f FindingType) IsValid() bool {
	switch f {
	case FindingTypeTodo, FindingTypeLegacy:
		return true
	default:
		return false
	}
}

// String returns the string representation of the finding type.
func (f FindingType) String() string { return string(f) }

// Finding represents a single-line code-quality finding — a TODO comment,
// a legacy pattern call, or similar issue that is NOT a code clone.
//
// Findings are distinct from ProcessedClone: a clone is a duplication of
// code fragments across multiple locations; a finding is a single-location
// annotation that signals technical debt or code smell.
type Finding struct {
	Filename Filepath      `json:"filename"`
	Line     LineNumber    `json:"line"`
	Type     FindingType   `json:"type"`
	Message  string        `json:"message"`
	Priority ClonePriority `json:"priority"`
	Tags     []string      `json:"tags,omitempty"`
}

// Validate checks that the finding's invariants hold.
func (f Finding) Validate() error {
	if f.Filename == "" {
		return ErrEmptyFilename
	}

	if f.Line == 0 {
		return ErrLineEndBeforeStart
	}

	if !f.Type.IsValid() {
		return fmt.Errorf("%w: %q", ErrInvalidFindingType, f.Type)
	}

	if !f.Priority.IsValid() {
		return fmt.Errorf("%w: %q", ErrInvalidClonePriority, f.Priority)
	}

	return nil
}
