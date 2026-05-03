package config

import (
	"errors"
	"fmt"
	"strings"
)

// ErrInvalidOutputFormat indicates an unsupported output format was requested.
var ErrInvalidOutputFormat = errors.New("invalid output format")

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
	return marshalStringType(of, isValidMethod[OutputFormat](), "output format")
}

// UnmarshalJSON implements json.Unmarshaler.
func (of *OutputFormat) UnmarshalJSON(data []byte) error {
	return unmarshalStringTypeToPointer(data,
		isValidMethod[OutputFormat](),
		OutputFormatText,
		"output format",
		of,
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

// DefaultOutputFormat returns the default output format.
func DefaultOutputFormat() OutputFormat {
	return OutputFormatText
}

// ParseOutputFormat converts a string to OutputFormat with validation.
func ParseOutputFormat(value string) (OutputFormat, error) {
	of := OutputFormat(strings.ToLower(value))
	if of.IsValid() {
		return of, nil
	}

	return "", fmt.Errorf("%w: %q must be one of (text|html|json|csv|plumbing|simple-json|sarif)",
		ErrInvalidOutputFormat, value)
}
