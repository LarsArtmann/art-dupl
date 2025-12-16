package config

import (
	"fmt"
	"slices"
	"strings"
)

// DetectionMethod represents the supported detection methods with type safety
type DetectionMethod string

const (
	// DetectionMethodHash uses hash-based comparison for code clones
	DetectionMethodHash DetectionMethod = "hash"

	// DetectionMethodArtDupl uses suffix tree-based detection (current method)
	DetectionMethodArtDupl DetectionMethod = "art-dupl"
	
	// DetectionMethodTodos finds TODO comments in code
	DetectionMethodTodos DetectionMethod = "todos"
	
	// DetectionMethodLegacy finds legacy code patterns
	DetectionMethodLegacy DetectionMethod = "legacy"
)

// String implements fmt.Stringer for DetectionMethod
func (dm DetectionMethod) String() string {
	return string(dm)
}

// IsValid checks if the detection method is supported
func (dm DetectionMethod) IsValid() bool {
	switch dm {
	case DetectionMethodHash, DetectionMethodArtDupl, DetectionMethodTodos, DetectionMethodLegacy:
		return true
	default:
		return false
	}
}

// MarshalJSON implements json.Marshaler for DetectionMethod
func (dm DetectionMethod) MarshalJSON() ([]byte, error) {
	// Explicitly type the isValid function
	isValid := func(d DetectionMethod) bool { return d.IsValid() }
	return MarshalEnumJSON(dm, isValid, "detection method")
}

// UnmarshalJSON implements json.Unmarshaler for DetectionMethod
func (dm *DetectionMethod) UnmarshalJSON(data []byte) error {
	return UnmarshalJSONForEnum(dm, data, "detection method")
}

// AllDetectionMethods returns list of all supported detection methods
func AllDetectionMethods() []DetectionMethod {
	return []DetectionMethod{
		DetectionMethodHash,
		DetectionMethodArtDupl,
	}
}

// DetectionMethods represents a collection of detection methods
type DetectionMethods []DetectionMethod

// String implements fmt.Stringer for DetectionMethods
func (dms DetectionMethods) String() string {
	if len(dms) == 0 {
		return ""
	}
	methods := make([]string, len(dms))
	for i, dm := range dms {
		methods[i] = dm.String()
	}
	return strings.Join(methods, ",")
}

// ParseDetectionMethods parses a comma-separated string of detection methods
func ParseDetectionMethods(s string) (DetectionMethods, error) {
	if s == "" {
		return DetectionMethods{DetectionMethodArtDupl}, nil // default
	}

	parts := strings.Split(s, ",")
	methods := make(DetectionMethods, 0, len(parts))

	for _, part := range parts {
		method := DetectionMethod(strings.TrimSpace(part))
		if !method.IsValid() {
			return nil, fmt.Errorf("invalid detection method: %s", part)
		}
		// Avoid duplicates
		found := slices.Contains(methods, method)
		if !found {
			methods = append(methods, method)
		}
	}

	if len(methods) == 0 {
		return DetectionMethods{DetectionMethodArtDupl}, nil
	}

	return methods, nil
}

// Contains checks if the methods contain a specific method
func (dms DetectionMethods) Contains(method DetectionMethod) bool {
	return slices.Contains(dms, method)
}

// IsDefault checks if only art-dupl method is selected
func (dms DetectionMethods) IsDefault() bool {
	return len(dms) == 1 && dms[0] == DetectionMethodArtDupl
}
