package domain

import (
	"fmt"
	"strings"
)

// HealthScore represents an A-F grade for code health based on duplication metrics.
type HealthScore string

const (
	HealthScoreA HealthScore = "A"
	HealthScoreB HealthScore = "B"
	HealthScoreC HealthScore = "C"
	HealthScoreD HealthScore = "D"
	HealthScoreF HealthScore = "F"
)

// IsValid reports whether the health score is one of the defined constants.
func (h HealthScore) IsValid() bool {
	switch h {
	case HealthScoreA, HealthScoreB, HealthScoreC, HealthScoreD, HealthScoreF:
		return true
	default:
		return false
	}
}

// String returns the string representation of the health score.
func (h HealthScore) String() string { return string(h) }

// MarshalJSON implements json.Marshaler for HealthScore.
func (h HealthScore) MarshalJSON() ([]byte, error) {
	if !h.IsValid() {
		return nil, fmt.Errorf(
			"%w: %s (valid options: A, B, C, D, F)",
			ErrInvalidHealthScore,
			h,
		)
	}

	return []byte(`"` + string(h) + `"`), nil
}

// UnmarshalJSON implements json.Unmarshaler for HealthScore.
func (h *HealthScore) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)

	score := HealthScore(str)
	if !score.IsValid() {
		return fmt.Errorf(
			"%w: %s (valid options: A, B, C, D, F)",
			ErrInvalidHealthScore,
			str,
		)
	}

	*h = score

	return nil
}
