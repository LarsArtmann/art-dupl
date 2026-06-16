package domain

import (
	"fmt"

	"github.com/LarsArtmann/art-dupl/pkg/enum"
)

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

// ParseFindingType parses a string into a FindingType, returning an error if invalid.
func ParseFindingType(s string) (FindingType, error) {
	return enum.Parse[FindingType](s, FindingType.IsValid, ErrInvalidFindingType)
}

// MarshalJSON implements json.Marshaler for FindingType.
func (f FindingType) MarshalJSON() ([]byte, error) {
	return enum.MarshalJSON(f, FindingType.IsValid, ErrInvalidFindingType) //nolint:wrapcheck
}

// UnmarshalJSON implements json.Unmarshaler for FindingType.
func (f *FindingType) UnmarshalJSON(data []byte) error {
	parsed, err := enum.UnmarshalJSON(data, FindingType.IsValid, ErrInvalidFindingType)
	if err != nil {
		return err //nolint:wrapcheck // domain sentinel passed through
	}

	*f = parsed

	return nil
}

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
		return ErrInvalidLineNumber
	}

	if !f.Type.IsValid() {
		return fmt.Errorf("%w: %q", ErrInvalidFindingType, f.Type)
	}

	if !f.Priority.IsValid() {
		return fmt.Errorf("%w: %q", ErrInvalidClonePriority, f.Priority)
	}

	return nil
}
