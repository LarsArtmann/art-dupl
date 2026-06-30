package config

import (
	"errors"

	"github.com/LarsArtmann/art-dupl/pkg/enum"
)

// ErrInvalidDetectionMode is returned when parsing an invalid detection mode.
var ErrInvalidDetectionMode = errors.New("invalid detection mode")

// DetectionMode controls how identifier names participate in clone matching.
type DetectionMode string

const (
	DetectionModeSemantic   DetectionMode = "semantic"
	DetectionModeExact      DetectionMode = "exact"
	DetectionModeStructural DetectionMode = "structural"
)

func (d DetectionMode) String() string { return string(d) }

func (d DetectionMode) IsValid() bool {
	return d == DetectionModeSemantic || d == DetectionModeExact || d == DetectionModeStructural
}

// IsSemantic reports whether this mode includes identifier/operator hashes
// in the AST token encoding (i.e. Semantic or Exact, not Structural).
func (d DetectionMode) IsSemantic() bool {
	return d == DetectionModeSemantic || d == DetectionModeExact
}

// MarshalJSON returns the string representation.
func (d DetectionMode) MarshalJSON() ([]byte, error) {
	return enum.MarshalJSON(d, DetectionMode.IsValid, ErrInvalidDetectionMode)
}

// UnmarshalJSON parses the string representation.
func (d *DetectionMode) UnmarshalJSON(data []byte) error {
	return enum.UnmarshalJSONInto(d, data, DetectionMode.IsValid, ErrInvalidDetectionMode)
}
