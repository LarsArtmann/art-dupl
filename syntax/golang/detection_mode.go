package golang

// DetectionMode controls how identifier names participate in clone matching.
//
// Three modes form a spectrum from strictest (fewest matches) to loosest
// (most matches):
//
//   - Exact: identifier names are hashed verbatim. Two clones match only if
//     they use the same variable/function/method names. This is copy-paste
//     detection (Type 1 clones).
//   - Semantic: local identifiers are alpha-normalized (canonicalized to v0,
//     v1, …) before hashing. Two clones match when they have the same
//     structure even if every variable was renamed. This detects Type 2
//     (parameterized) clones — the most common real-world duplication.
//   - Structural: identifier names are ignored entirely. Two clones match on
//     AST shape alone. Loosest matching; produces the most candidates.
type DetectionMode string

const (
	// DetectionModeExact hashes identifier names verbatim, so only
	// copy-paste-identical code (same names) matches.
	DetectionModeExact DetectionMode = "exact"

	// DetectionModeSemantic alpha-normalizes local identifiers before hashing,
	// so structurally identical code with renamed variables matches. This is
	// the default and the mode that detects Type 2 clones.
	DetectionModeSemantic DetectionMode = "semantic"

	// DetectionModeStructural ignores identifier names, matching on AST shape
	// alone.
	DetectionModeStructural DetectionMode = "structural"
)

func (dm DetectionMode) String() string { return string(dm) }

// IsValid returns true if the DetectionMode is one of the defined constants.
func (dm DetectionMode) IsValid() bool {
	switch dm {
	case DetectionModeExact, DetectionModeSemantic, DetectionModeStructural:
		return true
	default:
		return false
	}
}

// hashesIdentifiers returns true when identifier names are folded into the
// token hash (Exact and Semantic). Structural mode returns false — it matches
// on shape alone.
func (dm DetectionMode) hashesIdentifiers() bool {
	return dm == DetectionModeExact || dm == DetectionModeSemantic
}

// normalizesLocals returns true when local identifiers are alpha-normalized
// before hashing (Semantic only). Exact mode hashes original names; Structural
// mode ignores names entirely.
func (dm DetectionMode) normalizesLocals() bool {
	return dm == DetectionModeSemantic
}

// normalizesLiterals returns true when literal VALUES (strings, numbers) are
// normalized to their KIND (STRING, INT, FLOAT) before hashing. Semantic mode
// normalizes literals so that Type-2 clones with different literal values are
// detected. Exact mode hashes literal values verbatim (Type-1 only).
func (dm DetectionMode) normalizesLiterals() bool {
	return dm == DetectionModeSemantic
}

// IsSemantic returns true if the new alpha-normalizing semantic mode is active.
// Deprecated for internal gating: prefer hashesIdentifiers / normalizesLocals,
// which express the actual question being asked.
func (dm DetectionMode) IsSemantic() bool {
	return dm == DetectionModeSemantic
}
