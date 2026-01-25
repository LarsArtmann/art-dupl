package printer

import "fmt"

// Format represents the output format for stats command.
type Format string

const (
	// FormatText produces human-readable text output (default).
	FormatText Format = "text"
	// FormatJSON produces machine-readable JSON output.
	FormatJSON Format = "json"
	// FormatCSV produces machine-readable CSV output.
	FormatCSV Format = "csv"
)

// IsValid checks if the format is one of the supported formats.
func (f Format) IsValid() bool {
	switch f {
	case FormatText, FormatJSON, FormatCSV:
		return true
	default:
		return false
	}
}

// ParseFormat converts a string to a Format with validation.
// Returns an error if the value is not a valid format.
func ParseFormat(value string) (Format, error) {
	format := Format(value)
	if !format.IsValid() {
		return "", fmt.Errorf("invalid output format %q: must be one of (text|json|csv)", value)
	}
	return format, nil
}

// String returns the string representation of the format.
func (f Format) String() string {
	return string(f)
}
