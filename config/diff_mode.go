package config

import "fmt"

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
	return marshalStringType(dm, isValidMethod[DiffMode](), "diff mode")
}

// UnmarshalJSON implements json.Unmarshaler.
func (dm *DiffMode) UnmarshalJSON(data []byte) error {
	return unmarshalStringTypeToPointer(
		data,
		isValidMethod[DiffMode](),
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
