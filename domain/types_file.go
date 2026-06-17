package domain

import (
	"encoding/json"
	"fmt"

	"github.com/LarsArtmann/art-dupl/errors"
)

// Filepath represents a filesystem path.
type Filepath string

// NewFilepath creates a validated Filepath from a string.
func NewFilepath(path string) (Filepath, error) {
	if path == "" {
		return "", errors.NewValidationError("filepath cannot be empty", nil)
	}

	return Filepath(path), nil
}

// String returns the string representation of Filepath.
func (fp Filepath) String() string {
	return string(fp)
}

// MarshalJSON implements json.Marshaler for Filepath.
func (fp Filepath) MarshalJSON() ([]byte, error) {
	return marshalStringID(string(fp), "filepath cannot be empty")
}

// UnmarshalJSON implements json.Unmarshaler for Filepath.
func (fp *Filepath) UnmarshalJSON(data []byte) error {
	return unmarshalStringID(data, "Filepath", "filepath cannot be empty", func(s string) {
		*fp = Filepath(s)
	})
}

// LineNumber represents a line number in a source file.
// Line numbers start at 1 (not 0) in most editors.
// uint16 provides 0-65,535 range (sufficient for any source file).
type LineNumber uint16

// NewLineNumber creates a validated LineNumber from a uint16.
// Returns error if the line number is 0 (invalid).
func NewLineNumber(n uint16) (LineNumber, error) {
	if n == 0 {
		return 0, errors.NewValidationError("line number cannot be 0", nil)
	}

	return LineNumber(n), nil
}

// Uint16 returns the underlying uint16 value.
func (ln LineNumber) Uint16() uint16 {
	return uint16(ln)
}

// MarshalJSON implements json.Marshaler for LineNumber.
func (ln LineNumber) MarshalJSON() ([]byte, error) {
	if ln == 0 {
		return nil, errors.NewValidationError("line number cannot be 0", nil)
	}

	return json.Marshal(uint16(ln)) //nolint:wrapcheck // Standard JSON marshaling
}

// UnmarshalJSON implements json.Unmarshaler for LineNumber.
func (ln *LineNumber) UnmarshalJSON(data []byte) error {
	var n uint16

	err := json.Unmarshal(data, &n)
	if err != nil {
		return fmt.Errorf("unmarshal LineNumber failed (n=%v): %w", n, err)
	}

	if n == 0 {
		return errors.NewValidationError("line number cannot be 0", nil)
	}

	*ln = LineNumber(n)

	return nil
}
