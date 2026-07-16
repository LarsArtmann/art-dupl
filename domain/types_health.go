package domain

import (
	"github.com/LarsArtmann/art-dupl/pkg/enum"
)

// HealthScore represents an A-F grade for code health based on duplication metrics.
//
//nolint:recvcheck // standard Go JSON convention: MarshalJSON value receiver, UnmarshalJSON pointer receiver
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
	return enum.MarshalJSON(h, HealthScore.IsValid, ErrInvalidHealthScore)
}

// UnmarshalJSON implements json.Unmarshaler for HealthScore.
func (h *HealthScore) UnmarshalJSON(data []byte) error {
	return enum.UnmarshalJSONInto(
		h,
		data,
		HealthScore.IsValid,
		ErrInvalidHealthScore,
	)
}
