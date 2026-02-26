package printer

import (
	"fmt"

	"github.com/LarsArtmann/art-dupl/config"
)

// Format is an alias to config.OutputFormat for backward compatibility.
// This consolidates the stats format with the main output format type.
type Format = config.OutputFormat

// Format constants - aliases to config.OutputFormat values for backward compatibility.
const (
	// FormatText produces human-readable text output (default).
	FormatText = config.OutputFormatText
	// FormatJSON produces machine-readable JSON output.
	FormatJSON = config.OutputFormatJSON
	// FormatCSV produces machine-readable CSV output.
	FormatCSV = config.OutputFormatCSV
)

// ParseFormat converts a string to a Format with validation.
// Returns an error if the value is not a valid format.
func ParseFormat(value string) (Format, error) {
	format := config.OutputFormat(value)
	if !format.IsValid() {
		return "", fmt.Errorf(
			"invalid output format %q: must be one of (text|json|csv|html|plumbing|simple-json)",
			value,
		)
	}

	return format, nil
}
