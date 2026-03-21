package domain

import (
	"fmt"
	"strings"
)

// CloneSeverity represents clone severity levels for categorizing duplicate code.
//
//nolint:recvcheck // UnmarshalJSON requires pointer receiver, others use value receiver
type CloneSeverity string

// Clone severity constants define the importance level of detected clones.
const (
	// CloneSeverityLow indicates minor duplication with low impact.
	CloneSeverityLow CloneSeverity = "low"
	// CloneSeverityMedium indicates moderate duplication that should be reviewed.
	CloneSeverityMedium CloneSeverity = "medium"
	// CloneSeverityHigh indicates significant duplication requiring attention.
	CloneSeverityHigh CloneSeverity = "high"
	// CloneSeverityCritical indicates severe duplication that must be addressed.
	CloneSeverityCritical CloneSeverity = "critical"
)

// IsValid returns true if the severity is one of the defined constants.
func (cs CloneSeverity) IsValid() bool {
	switch cs {
	case CloneSeverityLow, CloneSeverityMedium, CloneSeverityHigh, CloneSeverityCritical:
		return true
	default:
		return false
	}
}

// String returns the string representation of the severity.
func (cs CloneSeverity) String() string { return string(cs) }

// MarshalJSON implements json.Marshaler for CloneSeverity.
func (cs CloneSeverity) MarshalJSON() ([]byte, error) {
	if !cs.IsValid() {
		return nil, fmt.Errorf(
			"%w: %s (valid options: low, medium, high, critical)",
			ErrInvalidCloneSeverity,
			cs,
		)
	}

	return []byte(`"` + string(cs) + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler for CloneSeverity.
func (cs *CloneSeverity) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)


	severity := CloneSeverity(str)
	if !severity.IsValid() {
		return fmt.Errorf(
			"%w: %s (valid options: low, medium, high, critical)",
			ErrInvalidSeverity,
			str,
		)
	}

	*cs = severity

	return nil
}
