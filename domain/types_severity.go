package domain

import (
	"fmt"
	"strings"
)

// CloneSeverity represents clone severity levels.
type CloneSeverity string

const (
	CloneSeverityLow      CloneSeverity = "low"
	CloneSeverityMedium   CloneSeverity = "medium"
	CloneSeverityHigh     CloneSeverity = "high"
	CloneSeverityCritical CloneSeverity = "critical"
)

func (cs CloneSeverity) IsValid() bool {
	switch cs {
	case CloneSeverityLow, CloneSeverityMedium, CloneSeverityHigh, CloneSeverityCritical:
		return true
	default:
		return false
	}
}

func (cs CloneSeverity) String() string { return string(cs) }

func (cs CloneSeverity) MarshalJSON() ([]byte, error) {
	if !cs.IsValid() {
		return nil, fmt.Errorf(
			"invalid clone severity: %s",
			cs,
		)
	}

	return []byte(`"` + string(cs) + `"`), nil
}

func (cs *CloneSeverity) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)

	severity := CloneSeverity(str)
	if !severity.IsValid() {
		return fmt.Errorf(
			"invalid clone severity: %s",
			str,
		)
	}

	*cs = severity

	return nil
}
