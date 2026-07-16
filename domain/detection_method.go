package domain

import (
	"errors"

	"github.com/LarsArtmann/art-dupl/pkg/enum"
)

// ErrInvalidDetectionMethod is returned when a detection method is not recognized.
var ErrInvalidDetectionMethod = errors.New("invalid detection method")

// DetectionMethod represents the algorithm used for duplicate detection.
// This is the canonical definition — config, sdk, and detection packages
// reference this type via aliases to prevent drift.
//
//nolint:recvcheck // standard Go JSON convention: MarshalJSON value receiver, UnmarshalJSON pointer receiver
type DetectionMethod string

const (
	// MethodArtDupl uses suffix tree algorithm on AST tokens.
	MethodArtDupl DetectionMethod = "art-dupl"

	// MethodHash uses rolling hash on file content.
	MethodHash DetectionMethod = "hash"
)

// String returns the string representation of the detection method.
func (m DetectionMethod) String() string { return string(m) }

// IsValid returns true if the detection method is one of the defined constants.
func (m DetectionMethod) IsValid() bool {
	switch m {
	case MethodArtDupl, MethodHash:
		return true
	default:
		return false
	}
}

// MarshalJSON implements json.Marshaler.
func (m DetectionMethod) MarshalJSON() ([]byte, error) {
	return enum.MarshalJSON(m, DetectionMethod.IsValid, ErrInvalidDetectionMethod)
}

// UnmarshalJSON implements json.Unmarshaler.
func (m *DetectionMethod) UnmarshalJSON(data []byte) error {
	return enum.UnmarshalJSONInto(
		m,
		data,
		DetectionMethod.IsValid,
		ErrInvalidDetectionMethod,
	)
}

// AllDetectionMethods returns all supported detection methods.
func AllDetectionMethods() []DetectionMethod {
	return []DetectionMethod{MethodArtDupl, MethodHash}
}

// DefaultDetectionMethod returns the default detection method.
func DefaultDetectionMethod() DetectionMethod {
	return MethodArtDupl
}
