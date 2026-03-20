package golang

// DetectionMode represents the mode of code detection/analysis.
type DetectionMode string

const (
	// DetectionModeSemantic enables semantic-aware matching based on identifier names.
	// Clones are matched by both structure AND identifier semantics.
	// Example: `a.String()` will NOT match `b.Error()` (different method names).
	DetectionModeSemantic DetectionMode = "semantic"

	// DetectionModeStructural enables structure-only matching.
	// Clones are matched by AST structure only, ignoring identifier names.
	// Example: `a.String()` WILL match `b.Error()` (same call structure).
	DetectionModeStructural DetectionMode = "structural"
)

func (dm DetectionMode) String() string { return string(dm) }

// IsValid returns true if the DetectionMode is a valid value.
func (dm DetectionMode) IsValid() bool {
	switch dm {
	case DetectionModeSemantic, DetectionModeStructural:
		return true
	default:
		return false
	}
}

// IsSemantic returns true if semantic matching is enabled.
func (dm DetectionMode) IsSemantic() bool {
	return dm == DetectionModeSemantic
}
