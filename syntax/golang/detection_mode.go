package golang

import "github.com/LarsArtmann/art-dupl/domain"

// DetectionMode is an alias for domain.DetectionMode. Use the inherited
// methods (IsSemantic, HashesIdentifiers, NormalizesLocals,
// NormalizesLiterals) from the domain package.
//
//art-dupl:accept architectural alias: syntax/golang cannot import config (layering), so it re-exports domain types independently (ADR-0005)
type DetectionMode = domain.DetectionMode

const (
	DetectionModeExact      = domain.DetectionModeExact      //art-dupl:accept architectural alias (ADR-0005)
	DetectionModeSemantic   = domain.DetectionModeSemantic   //art-dupl:accept architectural alias (ADR-0005)
	DetectionModeStructural = domain.DetectionModeStructural //art-dupl:accept architectural alias (ADR-0005)
)
