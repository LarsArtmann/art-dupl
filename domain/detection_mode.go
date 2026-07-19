package domain

import (
	"errors"

	"github.com/LarsArtmann/art-dupl/pkg/enum"
)

// ErrInvalidDetectionMode is returned when a detection mode is not recognized.
var ErrInvalidDetectionMode = errors.New("invalid detection mode")

// DetectionMode controls how identifier names participate in clone matching.
// This is the canonical definition — config and syntax/golang reference this
// type via aliases to prevent drift.
//
// Three modes form a spectrum from strictest (fewest matches) to loosest
// (most matches):
//
//   - Exact: identifier names are hashed verbatim. Two clones match only if
//     they use the same variable/function/method names. This is copy-paste
//     detection (Type 1 clones).
//   - Semantic: local identifiers are alpha-normalized (canonicalized to v0,
//     v1, …) before hashing. Two clones match when they have the same
//     structure even if every variable was renamed. This detects Type 2
//     (parameterized) clones — the most common real-world duplication.
//   - Structural: identifier names are ignored entirely. Two clones match on
//     AST shape alone. Loosest matching; produces the most candidates.
//
//nolint:recvcheck // standard Go JSON convention: MarshalJSON value receiver, UnmarshalJSON pointer receiver
type DetectionMode string

const (
	// DetectionModeExact hashes identifier names verbatim, so only
	// copy-paste-identical code (same names) matches.
	DetectionModeExact DetectionMode = "exact"

	// DetectionModeSemantic alpha-normalizes local identifiers before hashing,
	// so structurally identical code with renamed variables matches. This is
	// the default and the mode that detects Type 2 clones.
	DetectionModeSemantic DetectionMode = "semantic"

	// DetectionModeStructural ignores identifier names, matching on AST shape
	// alone.
	DetectionModeStructural DetectionMode = "structural"
)

// String returns the string representation of the detection mode.
func (dm DetectionMode) String() string { return string(dm) }

// IsValid returns true if the DetectionMode is one of the defined constants.
func (dm DetectionMode) IsValid() bool {
	switch dm {
	case DetectionModeExact, DetectionModeSemantic, DetectionModeStructural:
		return true
	default:
		return false
	}
}

// MarshalJSON implements json.Marshaler.
func (dm DetectionMode) MarshalJSON() ([]byte, error) {
	return enum.MarshalJSON(dm, DetectionMode.IsValid, ErrInvalidDetectionMode)
}

// UnmarshalJSON implements json.Unmarshaler.
func (dm *DetectionMode) UnmarshalJSON(data []byte) error {
	return enum.UnmarshalJSONInto(dm, data, DetectionMode.IsValid, ErrInvalidDetectionMode)
}

// AllDetectionModes returns all supported detection modes.
func AllDetectionModes() []DetectionMode {
	return []DetectionMode{DetectionModeExact, DetectionModeSemantic, DetectionModeStructural}
}

// DefaultDetectionMode returns the default detection mode.
func DefaultDetectionMode() DetectionMode {
	return DetectionModeSemantic
}

// IsSemantic returns true when identifier names participate in matching.
// In Semantic and Exact modes names are hashed, so they participate; only
// Structural mode ignores them entirely.
//
// Deprecated for internal gating: prefer HashesIdentifiers / NormalizesLocals,
// which express the actual question being asked.
func (dm DetectionMode) IsSemantic() bool {
	return dm == DetectionModeExact || dm == DetectionModeSemantic
}

// HashesIdentifiers reports whether identifier names are folded into the
// token hash. Exact and Semantic include them; Structural does not.
func (dm DetectionMode) HashesIdentifiers() bool {
	return dm == DetectionModeExact || dm == DetectionModeSemantic
}

// NormalizesLocals reports whether local identifiers are alpha-normalized
// before hashing. Only Semantic mode does this; Exact hashes original names
// and Structural ignores names entirely.
func (dm DetectionMode) NormalizesLocals() bool {
	return dm == DetectionModeSemantic
}

// NormalizesLiterals reports whether literal VALUES (strings, numbers) are
// normalized to their KIND (STRING, INT, FLOAT) before hashing. Semantic mode
// normalizes literals so that Type-2 clones with different literal values are
// detected. Exact mode hashes literal values verbatim (Type-1 only).
func (dm DetectionMode) NormalizesLiterals() bool {
	return dm == DetectionModeSemantic
}
