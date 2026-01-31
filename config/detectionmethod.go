package config

import (
	"encoding/json"
	"fmt"
	"strings"
)

// isValidStringType validates a string type against a set of valid values.
func isValidStringType[T ~string](val T, validValues map[T]bool) bool {
	return validValues[val]
}

// unmarshalStringType unmarshals JSON into a string type with validation.
func unmarshalStringType[T ~string](data []byte, isValid func(T) bool, defaultVal T, typeName string) (T, error) {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return defaultVal, err
	}

	typed := T(str)
	if !isValid(typed) {
		return defaultVal, fmt.Errorf("invalid %s: %s", typeName, str)
	}

	return typed, nil
}

// unmarshalStringTypeToPointer unmarshals JSON into a pointer to a string type with validation.
func unmarshalStringTypeToPointer[T ~string](data []byte, isValid func(T) bool, defaultVal T, typeName string, target *T) error {
	val, err := unmarshalStringType[T](data, isValid, defaultVal, typeName)
	if err != nil {
		return err
	}
	*target = val
	return nil
}

// marshalStringType marshals a string type to JSON with validation.
// For invalid values (typically zero values with omitempty), returns null to allow omission.
func marshalStringType[T ~string](val T, isValid func(T) bool, typeName string) ([]byte, error) {
	if !isValid(val) {
		// Return null for invalid values to allow omitempty to work
		// This handles zero values that should be omitted from JSON output
		return []byte("null"), nil
	}
	return json.Marshal(string(val))
}

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
	return marshalStringType(dm, func(s DetectionMethod) bool { return s.IsValid() }, "detection method")
}

// UnmarshalJSON implements json.Unmarshaler.
func (dm *DetectionMethod) UnmarshalJSON(data []byte) error {
	return unmarshalStringTypeToPointer(data,
		func(s DetectionMethod) bool { return s.IsValid() },
		DetectionMethodArtDupl,
		"detection method",
		dm,
	)
}

// OutputFormat represents the output format type.
type OutputFormat string

const (
	OutputFormatText       OutputFormat = "text"
	OutputFormatHTML       OutputFormat = "html"
	OutputFormatJSON       OutputFormat = "json"
	OutputFormatPlumbing   OutputFormat = "plumbing"
	OutputFormatSimpleJSON OutputFormat = "simple-json"
)

var validOutputFormats = map[OutputFormat]bool{
	OutputFormatText:       true,
	OutputFormatHTML:       true,
	OutputFormatJSON:       true,
	OutputFormatPlumbing:   true,
	OutputFormatSimpleJSON: true,
}

// String implements fmt.Stringer.
func (of OutputFormat) String() string {
	return string(of)
}

// IsValid validates output format.
func (of OutputFormat) IsValid() bool {
	return isValidStringType(of, validOutputFormats)
}

// MarshalJSON implements json.Marshaler.
func (of OutputFormat) MarshalJSON() ([]byte, error) {
	return marshalStringType(of, func(s OutputFormat) bool { return s.IsValid() }, "output format")
}

// UnmarshalJSON implements json.Unmarshaler.
func (of *OutputFormat) UnmarshalJSON(data []byte) error {
	return unmarshalStringTypeToPointer(data,
		func(s OutputFormat) bool { return s.IsValid() },
		OutputFormatText,
		"output format",
		of,
	)
}

// SortCriteria represents the sort criteria type.
type SortCriteria string

const (
	SortBySize        SortCriteria = "size"
	SortByOccurrence  SortCriteria = "occurrence"
	SortByHash        SortCriteria = "hash"
	SortByTotalTokens SortCriteria = "total-tokens"
)

var validSortCriteria = map[SortCriteria]bool{
	SortBySize:        true,
	SortByOccurrence:  true,
	SortByHash:        true,
	SortByTotalTokens: true,
}

// String implements fmt.Stringer.
func (sc SortCriteria) String() string {
	return string(sc)
}

// IsValid validates sort criteria.
func (sc SortCriteria) IsValid() bool {
	return isValidStringType(sc, validSortCriteria)
}

// MarshalJSON implements json.Marshaler.
func (sc SortCriteria) MarshalJSON() ([]byte, error) {
	return marshalStringType(sc, func(s SortCriteria) bool { return s.IsValid() }, "sort criteria")
}

// UnmarshalJSON implements json.Unmarshaler.
func (sc *SortCriteria) UnmarshalJSON(data []byte) error {
	return unmarshalStringTypeToPointer(data,
		func(s SortCriteria) bool { return s.IsValid() },
		SortBySize,
		"sort criteria",
		sc,
	)
}

// AllOutputFormats returns all supported output formats.
func AllOutputFormats() []OutputFormat {
	return []OutputFormat{
		OutputFormatText,
		OutputFormatHTML,
		OutputFormatJSON,
		OutputFormatPlumbing,
		OutputFormatSimpleJSON,
	}
}

// AllSortCriteria returns all supported sort criteria.
func AllSortCriteria() []SortCriteria {
	return []SortCriteria{
		SortBySize,
		SortByOccurrence,
		SortByHash,
		SortByTotalTokens,
	}
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
			return nil, fmt.Errorf("invalid detection method: %s", method)
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
			return fmt.Errorf("invalid detection method: %s", method)
		}
	}
	return nil
}

// DefaultDetectionMethod returns the default detection method.
func DefaultDetectionMethod() DetectionMethod {
	return DetectionMethodArtDupl
}

// DefaultOutputFormat returns the default output format.
func DefaultOutputFormat() OutputFormat {
	return OutputFormatText
}

// DefaultSortCriteria returns the default sort criteria.
func DefaultSortCriteria() SortCriteria {
	return SortBySize
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
