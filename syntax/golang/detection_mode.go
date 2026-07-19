package golang

import "github.com/LarsArtmann/art-dupl/domain"

// DetectionMode is an alias for domain.DetectionMode. Use the inherited
// methods (IsSemantic, HashesIdentifiers, NormalizesLocals,
// NormalizesLiterals) from the domain package.
//

type DetectionMode = domain.DetectionMode

const (
	DetectionModeExact      = domain.DetectionModeExact
	DetectionModeSemantic   = domain.DetectionModeSemantic
	DetectionModeStructural = domain.DetectionModeStructural
)
