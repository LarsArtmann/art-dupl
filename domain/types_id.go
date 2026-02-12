package domain

import (
	"github.com/LarsArtmann/art-dupl/errors"
)

// CloneGroupID represents a unique identifier for a clone group.
type CloneGroupID string

// NewCloneGroupID creates a validated CloneGroupID from a string.
func NewCloneGroupID(id string) (CloneGroupID, error) {
	if id == "" {
		return "", errors.NewValidationError("clone group ID cannot be empty", nil)
	}
	return CloneGroupID(id), nil
}

// String returns the string representation of CloneGroupID.
func (id CloneGroupID) String() string {
	return string(id)
}

// MarshalJSON implements json.Marshaler for CloneGroupID.
func (id CloneGroupID) MarshalJSON() ([]byte, error) {
	return marshalStringID(string(id), "clone group ID cannot be empty")
}

// UnmarshalJSON implements json.Unmarshaler for CloneGroupID.
func (id *CloneGroupID) UnmarshalJSON(data []byte) error {
	return unmarshalStringID(data, "CloneGroupID", "clone group ID cannot be empty", func(s string) {
		*id = CloneGroupID(s)
	})
}

// AnalysisID represents a unique identifier for an analysis.
type AnalysisID string

// NewAnalysisID creates a validated AnalysisID from a string.
func NewAnalysisID(id string) (AnalysisID, error) {
	if id == "" {
		return "", errors.NewValidationError("analysis ID cannot be empty", nil)
	}
	return AnalysisID(id), nil
}

// String returns the string representation of AnalysisID.
func (id AnalysisID) String() string {
	return string(id)
}

// MarshalJSON implements json.Marshaler for AnalysisID.
func (id AnalysisID) MarshalJSON() ([]byte, error) {
	return marshalStringID(string(id), "analysis ID cannot be empty")
}

// UnmarshalJSON implements json.Unmarshaler for AnalysisID.
func (id *AnalysisID) UnmarshalJSON(data []byte) error {
	return unmarshalStringID(data, "AnalysisID", "analysis ID cannot be empty", func(s string) {
		*id = AnalysisID(s)
	})
}
