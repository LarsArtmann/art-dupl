package config

import (
	"fmt"
)

// FileType represents a file type filter for analysis.
type FileType string

const (
	// FileTypeGo represents Go source files (.go).
	FileTypeGo FileType = "go"
	// FileTypeTempl represents Templ template files (.templ).
	FileTypeTempl FileType = "templ"
	// FileTypeAll represents all file types (no filtering).
	FileTypeAll FileType = ""
)

//nolint:gochecknoglobals // Lookup table for valid file types, initialized once at package load
var validFileTypes = map[FileType]bool{
	FileTypeGo:    true,
	FileTypeTempl: true,
	FileTypeAll:   true,
}

// String implements fmt.Stringer.
func (ft FileType) String() string {
	return string(ft)
}

// IsValid validates file type.
func (ft FileType) IsValid() bool {
	return isValidStringType(ft, validFileTypes)
}

// MarshalJSON implements json.Marshaler.
// Returns null for empty string to allow omitempty to work.
func (ft FileType) MarshalJSON() ([]byte, error) {
	if ft == FileTypeAll {
		return []byte("null"), nil
	}
	return marshalStringType(
		ft,
		func(s FileType) bool { return s.IsValid() },
		"file type",
	)
}

// UnmarshalJSON implements json.Unmarshaler.
func (ft *FileType) UnmarshalJSON(data []byte) error {
	return unmarshalStringTypeToPointer(data,
		func(s FileType) bool { return s.IsValid() },
		FileTypeAll,
		"file type",
		ft,
	)
}

// Matches returns true if the given file path matches this file type filter.
// Empty string (FileTypeAll) matches all files.
func (ft FileType) Matches(path string) bool {
	switch ft {
	case FileTypeGo:
		return hasSuffix(path, ".go")
	case FileTypeTempl:
		return hasSuffix(path, ".templ")
	case FileTypeAll:
		return true
	default:
		return true
	}
}

// hasSuffix is a helper function to check if a string ends with a suffix.
// Defined here to avoid importing strings package just for this.
func hasSuffix(s, suffix string) bool {
	if len(suffix) > len(s) {
		return false
	}
	return s[len(s)-len(suffix):] == suffix
}

// ParseFileType parses a string into a FileType, validating the value.
// Returns FileTypeAll and an error if the value is invalid.
func ParseFileType(s string) (FileType, error) {
	if s == "" {
		return FileTypeAll, nil
	}
	ft := FileType(s)
	if !ft.IsValid() {
		return FileTypeAll, fmt.Errorf("invalid file type: %q (valid: go, templ)", s)
	}
	return ft, nil
}
