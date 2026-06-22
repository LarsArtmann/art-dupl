package config

import (
	"fmt"
	"strings"

	"github.com/LarsArtmann/art-dupl/domain"
)

// DetectionMethod is an alias for domain.DetectionMethod to prevent drift
// across config, SDK, and detection packages.
type DetectionMethod = domain.DetectionMethod

// Constant aliases for backward compatibility with existing config code.
const (
	// DetectionMethodHash uses hash-based comparison.
	DetectionMethodHash = domain.MethodHash

	// DetectionMethodArtDupl uses suffix tree detection.
	DetectionMethodArtDupl = domain.MethodArtDupl
)

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
			return nil, fmt.Errorf(
				"%w: %s (input: %q)",
				domain.ErrInvalidDetectionMethod,
				method,
				methodsStr,
			)
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
			return fmt.Errorf("%w: %s", domain.ErrInvalidDetectionMethod, method)
		}
	}

	return nil
}

// AllDetectionMethods returns all supported detection methods.
func AllDetectionMethods() []DetectionMethod {
	return domain.AllDetectionMethods()
}

// DefaultDetectionMethod returns the default detection method.
func DefaultDetectionMethod() DetectionMethod {
	return domain.DefaultDetectionMethod()
}

// detectionMethodStrings converts a slice of DetectionMethod to string slice.
func detectionMethodStrings(methods []DetectionMethod) []string {
	result := make([]string, len(methods))
	for i, m := range methods {
		result[i] = string(m)
	}

	return result
}
