package printer

import (
	"errors"
	"fmt"

	"github.com/LarsArtmann/art-dupl/config"
	duplerrors "github.com/LarsArtmann/art-dupl/errors"
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

// ErrInvalidOutputFormat indicates an unsupported output format was requested.
var ErrInvalidOutputFormat = errors.New("invalid output format")

// ParseFormat converts a string to a Format with validation.
// Returns an error if the value is not a valid format.
func ParseFormat(value string) (Format, error) {
	format := config.OutputFormat(value)
	if !format.IsValid() {
		return "", duplerrors.NewEnumValidationError(
			"OutputFormat",
			value,
			fmt.Errorf(
				"%w: %q must be one of (text|json|csv|html|plumbing|simple-json)",
				ErrInvalidOutputFormat,
				value,
			),
		)
	}

	return format, nil
}
