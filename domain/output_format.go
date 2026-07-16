package domain

import (
	"errors"
	"fmt"
	"strings"

	"github.com/LarsArtmann/art-dupl/pkg/enum"
)

// ErrInvalidOutputFormat is returned when an output format value is not recognized.
var ErrInvalidOutputFormat = errors.New("invalid output format")

// OutputFormat controls the presentation format of clone detection results.
//
//nolint:recvcheck // standard Go JSON convention: MarshalJSON value receiver, UnmarshalJSON pointer receiver
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

func (of OutputFormat) String() string { return string(of) }

func (of OutputFormat) IsValid() bool {
	switch of {
	case OutputFormatText, OutputFormatHTML, OutputFormatJSON, OutputFormatCSV,
		OutputFormatPlumbing, OutputFormatSimpleJSON, OutputFormatSARIF:
		return true
	default:
		return false
	}
}

func (of OutputFormat) MarshalJSON() ([]byte, error) {
	return enum.MarshalJSON(of, OutputFormat.IsValid, ErrInvalidOutputFormat)
}

func (of *OutputFormat) UnmarshalJSON(data []byte) error {
	return enum.UnmarshalJSONInto(of, data, OutputFormat.IsValid, ErrInvalidOutputFormat)
}

func AllOutputFormats() []OutputFormat {
	return []OutputFormat{
		OutputFormatText, OutputFormatHTML, OutputFormatJSON, OutputFormatCSV,
		OutputFormatPlumbing, OutputFormatSimpleJSON, OutputFormatSARIF,
	}
}

func DefaultOutputFormat() OutputFormat {
	return OutputFormatText
}

func ParseOutputFormat(value string) (OutputFormat, error) {
	of := OutputFormat(strings.ToLower(value))
	if of.IsValid() {
		return of, nil
	}

	return "", fmt.Errorf("%w: %q must be one of (text|html|json|csv|plumbing|simple-json|sarif)",
		ErrInvalidOutputFormat, value)
}
