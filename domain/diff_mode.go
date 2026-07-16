package domain

import (
	"errors"

	"github.com/LarsArtmann/art-dupl/pkg/enum"
)

// ErrInvalidDiffMode is returned when a diff mode value is not recognized.
var ErrInvalidDiffMode = errors.New("invalid diff mode")

// DiffMode controls diff visualization in HTML output.
//
//nolint:recvcheck // standard Go JSON convention: MarshalJSON value receiver, UnmarshalJSON pointer receiver
type DiffMode string

const (
	DiffModeDisabled   DiffMode = "disabled"
	DiffModeSideBySide DiffMode = "side-by-side"
	DiffModeInline     DiffMode = "inline"
)

func (dm DiffMode) String() string { return string(dm) }

func (dm DiffMode) IsValid() bool {
	switch dm {
	case DiffModeDisabled, DiffModeSideBySide, DiffModeInline:
		return true
	default:
		return false
	}
}

func (dm DiffMode) IsEnabled() bool {
	return dm != DiffModeDisabled && dm.IsValid()
}

func (dm DiffMode) MarshalJSON() ([]byte, error) {
	return enum.MarshalJSON(dm, DiffMode.IsValid, ErrInvalidDiffMode)
}

func (dm *DiffMode) UnmarshalJSON(data []byte) error {
	return enum.UnmarshalJSONInto(dm, data, DiffMode.IsValid, ErrInvalidDiffMode)
}

func AllDiffModes() []DiffMode {
	return []DiffMode{DiffModeDisabled, DiffModeSideBySide, DiffModeInline}
}

func DefaultDiffMode() DiffMode {
	return DiffModeDisabled
}
