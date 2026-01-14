package config

import (
	"encoding/json"
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

// String implements fmt.Stringer.
func (dm DetectionMethod) String() string {
	return string(dm)
}

// IsValid validates detection method.
func (dm DetectionMethod) IsValid() bool {
	switch dm {
	case DetectionMethodHash, DetectionMethodArtDupl, DetectionMethodTodos, DetectionMethodLegacy:
		return true
	default:
		return false
	}
}

// MarshalJSON implements json.Marshaler.
func (dm DetectionMethod) MarshalJSON() ([]byte, error) {
	if !dm.IsValid() {
		return nil, fmt.Errorf("invalid detection method: %s", dm)
	}
	return json.Marshal(string(dm))
}

// UnmarshalJSON implements json.Unmarshaler.
func (dm *DetectionMethod) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	parsed := DetectionMethod(str)
	if !parsed.IsValid() {
		*dm = DetectionMethodArtDupl // default
		return fmt.Errorf("invalid detection method: %s", str)
	}

	*dm = parsed
	return nil
}

// OutputFormat represents the output format type.
type OutputFormat string

const (
	OutputFormatText      OutputFormat = "text"
	OutputFormatHTML      OutputFormat = "html"
	OutputFormatJSON      OutputFormat = "json"
	OutputFormatPlumbing OutputFormat = "plumbing"
	OutputFormatSimpleJSON OutputFormat = "simple-json"
)

// String implements fmt.Stringer.
func (of OutputFormat) String() string {
	return string(of)
}

// IsValid validates output format.
func (of OutputFormat) IsValid() bool {
	switch of {
	case OutputFormatText, OutputFormatHTML, OutputFormatJSON, OutputFormatPlumbing, OutputFormatSimpleJSON:
		return true
	default:
		return false
	}
}

// MarshalJSON implements json.Marshaler.
func (of OutputFormat) MarshalJSON() ([]byte, error) {
	if !of.IsValid() {
		return nil, fmt.Errorf("invalid output format: %s", of)
	}
	return json.Marshal(string(of))
}

// UnmarshalJSON implements json.Unmarshaler.
func (of *OutputFormat) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	parsed := OutputFormat(str)
	if !parsed.IsValid() {
		*of = OutputFormatText // default
		return fmt.Errorf("invalid output format: %s", str)
	}

	*of = parsed
	return nil
}

// SortCriteria represents the sort criteria type.
type SortCriteria string

const (
	SortBySize        SortCriteria = "size"
	SortByOccurrence  SortCriteria = "occurrence"
	SortByHash        SortCriteria = "hash"
	SortByTotalTokens SortCriteria = "total-tokens"
)

// String implements fmt.Stringer.
func (sc SortCriteria) String() string {
	return string(sc)
}

// IsValid validates sort criteria.
func (sc SortCriteria) IsValid() bool {
	switch sc {
	case SortBySize, SortByOccurrence, SortByHash, SortByTotalTokens:
		return true
	default:
		return false
	}
}

// MarshalJSON implements json.Marshaler.
func (sc SortCriteria) MarshalJSON() ([]byte, error) {
	if !sc.IsValid() {
		return nil, fmt.Errorf("invalid sort criteria: %s", sc)
	}
	return json.Marshal(string(sc))
}

// UnmarshalJSON implements json.Unmarshaler.
func (sc *SortCriteria) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}

	parsed := SortCriteria(str)
	if !parsed.IsValid() {
		*sc = SortBySize // default
		return fmt.Errorf("invalid sort criteria: %s", str)
	}

	*sc = parsed
	return nil
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
