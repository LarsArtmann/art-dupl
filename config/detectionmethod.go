package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Static errors for config validation.
// Static errors for config validation.
var (
	ErrInvalidDetectionMethod = errors.New("invalid detection method")
	ErrInvalidType            = errors.New("invalid type")
)

// isValidStringType validates a string type against a set of valid values.
func isValidStringType[T ~string](val T, validValues map[T]bool) bool {
	return validValues[val]
}

// unmarshalStringType unmarshals JSON into a string type with validation.
func unmarshalStringType[T ~string](
	data []byte,
	isValid func(T) bool,
	defaultVal T,
	typeName string,
) (T, error) {
	var str string

	err := json.Unmarshal(data, &str)
	if err != nil {
		return defaultVal, fmt.Errorf(
			"unmarshaling %s from data %q (defaultValue=%s): %w",
			typeName,
			string(data),
			defaultVal,
			err,
		)
	}

	typed := T(str)
	if !isValid(typed) {
		return defaultVal, fmt.Errorf(
			"%w %s (value=%q, defaultVal=%v): %s",
			ErrInvalidType,
			typeName,
			str,
			defaultVal,
			str,
		)
	}

	return typed, nil
}

// unmarshalStringTypeToPointer unmarshals JSON into a pointer to a string type with validation.
func unmarshalStringTypeToPointer[T ~string](
	data []byte,
	isValid func(T) bool,
	defaultVal T,
	typeName string,
	target *T,
) error {
	val, err := unmarshalStringType[T](data, isValid, defaultVal, typeName)
	if err != nil {
		return fmt.Errorf(
			"unmarshal to %s failed (target=%v, defaultVal=%v): %w",
			typeName,
			target,
			defaultVal,
			err,
		)
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

	data, err := json.Marshal(string(val))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal %s (val=%v): %w", typeName, val, err)
	}

	return data, nil
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

//nolint:gochecknoglobals // Lookup table for valid detection methods, initialized once at package load
var validDetectionMethods = map[DetectionMethod]bool{
	DetectionMethodHash:    true,
	DetectionMethodArtDupl: true,
	// DetectionMethodTodos and DetectionMethodLegacy are defined but not yet implemented
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
		func(s DetectionMethod) bool { return s.IsValid() },
		"detection method",
	)
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
	OutputFormatCSV        OutputFormat = "csv"
	OutputFormatPlumbing   OutputFormat = "plumbing"
	OutputFormatSimpleJSON OutputFormat = "simple-json"
	OutputFormatSARIF      OutputFormat = "sarif"
)

//nolint:gochecknoglobals // Lookup table for valid output formats, initialized once at package load
var validOutputFormats = map[OutputFormat]bool{
	OutputFormatText:       true,
	OutputFormatHTML:       true,
	OutputFormatJSON:       true,
	OutputFormatCSV:        true,
	OutputFormatPlumbing:   true,
	OutputFormatSimpleJSON: true,
	OutputFormatSARIF:      true,
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

//nolint:gochecknoglobals // Lookup table for valid sort criteria
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
		OutputFormatCSV,
		OutputFormatPlumbing,
		OutputFormatSimpleJSON,
		OutputFormatSARIF,
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

// DiffMode represents the diff visualization mode.
type DiffMode string

const (
	// DiffModeDisabled disables diff visualization.
	DiffModeDisabled DiffMode = "disabled"
	// DiffModeSideBySide enables side-by-side diff visualization.
	DiffModeSideBySide DiffMode = "side-by-side"
	// DiffModeInline enables inline diff visualization.
	DiffModeInline DiffMode = "inline"
)

//nolint:gochecknoglobals // Lookup table for valid diff modes
var validDiffModes = map[DiffMode]bool{
	DiffModeDisabled:   true,
	DiffModeSideBySide: true,
	DiffModeInline:     true,
}

// String implements fmt.Stringer.
func (dm DiffMode) String() string {
	return string(dm)
}

// IsValid validates diff mode.
func (dm DiffMode) IsValid() bool {
	return isValidStringType(dm, validDiffModes)
}

// IsEnabled returns true if diff mode is enabled.
func (dm DiffMode) IsEnabled() bool {
	return dm != DiffModeDisabled && dm.IsValid()
}

// MarshalJSON implements json.Marshaler.
func (dm DiffMode) MarshalJSON() ([]byte, error) {
	return marshalStringType(dm, func(s DiffMode) bool { return s.IsValid() }, "diff mode")
}

// UnmarshalJSON implements json.Unmarshaler.
func (dm *DiffMode) UnmarshalJSON(data []byte) error {
	return unmarshalStringTypeToPointer(data,
		func(s DiffMode) bool { return s.IsValid() },
		DiffModeDisabled,
		"diff mode",
		dm,
	)
}

// ParseDiffMode parses a diff mode string.
func ParseDiffMode(s string) (DiffMode, error) {
	switch s {
	case "true", "side-by-side":
		return DiffModeSideBySide, nil
	case "inline":
		return DiffModeInline, nil
	case "false", "disabled", "":
		return DiffModeDisabled, nil
	default:
		return DiffModeDisabled, fmt.Errorf("%w: %s", ErrInvalidType, s)
	}
}

// AllDiffModes returns all supported diff modes.
func AllDiffModes() []DiffMode {
	return []DiffMode{
		DiffModeDisabled,
		DiffModeSideBySide,
		DiffModeInline,
	}
}

// DefaultDiffMode returns the default diff mode.
func DefaultDiffMode() DiffMode {
	return DiffModeDisabled
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
