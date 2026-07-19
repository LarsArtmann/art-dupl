package config

import "github.com/LarsArtmann/art-dupl/domain"

// DetectionMode is an alias for domain.DetectionMode. It exists so callers
// that import config continue to write config.DetectionModeSemantic /
// config.IsValid etc. without an additional import.
//
//nolint:recvcheck // methods are inherited from the domain alias.
type DetectionMode = domain.DetectionMode

const (
	DetectionModeSemantic   = domain.DetectionModeSemantic
	DetectionModeExact      = domain.DetectionModeExact
	DetectionModeStructural = domain.DetectionModeStructural
)

// ErrInvalidDetectionMode aliases the domain-level sentinel.
var ErrInvalidDetectionMode = domain.ErrInvalidDetectionMode
