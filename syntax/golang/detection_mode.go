package golang

import "github.com/LarsArtmann/art-dupl/domain"

// DetectionMode is an alias for domain.DetectionMode. Use the inherited
// methods (IsSemantic, HashesIdentifiers, NormalizesLocals,
// NormalizesLiterals) from the domain package.
//
//art-dupl:accept architectural alias: syntax/golang cannot import config (layering), so it re-exports domain types independently (ADR-0005)
type DetectionMode = domain.DetectionMode

//art-dupl:accept architectural alias: constants re-exported for package-local use (ADR-0005)
const (
	DetectionModeExact      = domain.DetectionModeExact
	DetectionModeSemantic   = domain.DetectionModeSemantic
	DetectionModeStructural = domain.DetectionModeStructural
)
