package config

import (
	"fmt"

	"github.com/LarsArtmann/art-dupl/domain"
)

type DiffMode = domain.DiffMode

const (
	DiffModeDisabled   = domain.DiffModeDisabled
	DiffModeSideBySide = domain.DiffModeSideBySide
	DiffModeInline     = domain.DiffModeInline
)

var (
	AllDiffModes       = domain.AllDiffModes
	DefaultDiffMode    = domain.DefaultDiffMode
	ErrInvalidDiffMode = domain.ErrInvalidDiffMode
)

// ParseDiffMode parses a diff mode string with backwards-compatible aliases.
// "true"→side-by-side, "false"→disabled for legacy --diff flag compatibility.
func ParseDiffMode(s string) (DiffMode, error) {
	switch s {
	case "true", "side-by-side":
		return DiffModeSideBySide, nil
	case "inline":
		return DiffModeInline, nil
	case "false", "disabled", "":
		return DiffModeDisabled, nil
	default:
		return DiffModeDisabled, fmt.Errorf("%w: %s", ErrInvalidDiffMode, s)
	}
}
