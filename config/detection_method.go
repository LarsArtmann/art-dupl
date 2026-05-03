package config

import (
	"fmt"
	"strings"
)

// DetectionMethod represents the detection method type.
type DetectionMethod string

const (
	// DetectionMethodHash uses hash-based comparison.
	DetectionMethodHash DetectionMethod = "hash"
	// DetectionMethodArtDupl uses suffix tree detection.
	DetectionMethodArtDupl DetectionMethod = "art-dupl"
	// DetectionMethodTodos finds TODO comments.
	DetectionMethodTodos DetectionMethod = "todos"
	// DetectionMethodLegacy finds legacy code patterns.
	DetectionMethodLegacy DetectionMethod = "legacy"
)

//nolint:gochecknoglobals // Lookup table for valid detection methods, initialized once at package load
var validDetectionMethods = map[DetectionMethod]bool{
	DetectionMethodHash:    true,
	DetectionMethodArtDupl: true,
	DetectionMethodTodos:   true,
	DetectionMethodLegacy:  true,
}

// String implements fmt.Stringer.
func (dm DetectionMethod) String() string {
	return string(dm)
}

// IsValid validates detection method.
func (dm DetectionMethod) IsValid() bool {
	return isValidStringType(dm, validDetectionMethods)
}

// MarshalJSON implements json.Marshaler.
func (dm DetectionMethod) MarshalJSON() ([]byte, error) {
	return marshalStringType(
		dm,
		isValidMethod[DetectionMethod](),
		"detection method",
	)
}

// UnmarshalJSON implements json.Unmarshaler.
func (dm *DetectionMethod) UnmarshalJSON(data []byte) error {
	return unmarshalStringTypeToPointer(data,
		isValidMethod[DetectionMethod](),
		DetectionMethodArtDupl,
		"detection method",
		dm,
	)
}

// ParseDetectionMethods parses comma-separated detection methods.
func ParseDetectionMethods(methodsStr string) ([]DetectionMethod, error) {
	if methodsStr == "" {
		return []DetectionMethod{DetectionMethodArtDupl}, nil
	}

	methods := strings.Split(methodsStr, ",")

	var result []DetectionMethod

	for _, method := range methods {
		method = strings.TrimSpace(method)
		if method == "" {
			continue
		}

		dm := DetectionMethod(method)
		if !dm.IsValid() {
			return nil, fmt.Errorf("%w: %s", ErrInvalidDetectionMethod, method)
		}

		result = append(result, dm)
	}

	// Remove duplicates while preserving order
	seen := make(map[DetectionMethod]bool)

	var unique []DetectionMethod

	for _, m := range result {
		if !seen[m] {
			seen[m] = true
			unique = append(unique, m)
		}
	}

	return unique, nil
}

// ValidateDetectionMethods validates a list of detection methods.
func ValidateDetectionMethods(methods []DetectionMethod) error {
	for _, method := range methods {
		if !method.IsValid() {
			return fmt.Errorf("%w: %s", ErrInvalidDetectionMethod, method)
		}
	}

	return nil
}

// AllDetectionMethods returns all supported detection methods.
func AllDetectionMethods() []DetectionMethod {
	return []DetectionMethod{
		DetectionMethodHash,
		DetectionMethodArtDupl,
		DetectionMethodTodos,
		DetectionMethodLegacy,
	}
}

// DefaultDetectionMethod returns the default detection method.
func DefaultDetectionMethod() DetectionMethod {
	return DetectionMethodArtDupl
}
