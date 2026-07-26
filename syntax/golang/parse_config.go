package golang

import (
	"fmt"

	"github.com/LarsArtmann/art-dupl/domain"
)

// ErrInvalidDetectionMode aliases the domain-level sentinel so that
// errors.Is matches consistently across packages.
//
// art-dupl:accept architectural alias: syntax/golang cannot import config, re-exports domain sentinel (ADR-0005)
var ErrInvalidDetectionMode = domain.ErrInvalidDetectionMode

// ParseConfig holds configuration for the Go source code parser.
// It controls how AST nodes are transformed and matched.
type ParseConfig struct {
	// Mode determines whether semantic matching is enabled.
	// SemanticMode: Clones matched by structure AND identifier semantics (default).
	// StructuralMode: Clones matched by AST structure only.
	Mode DetectionMode

	// Preloaded, when non-nil, supplies a pre-parsed AST and type-checking
	// results. The transformer uses this AST instead of re-parsing and encodes
	// variable types into identifier hashes to reduce false positives where
	// different types share a method name (e.g. time.Time.String vs *big.Int.String).
	Preloaded *PreloadedAST
}

// DefaultParseConfig returns a ParseConfig with sensible defaults.
// By default, semantic matching is enabled to reduce false positives.
func DefaultParseConfig() ParseConfig {
	return ParseConfig{
		Mode: DetectionModeSemantic,
	}
}

// MustParseConfig returns a ParseConfig with the given mode, panicking if invalid.
// Use this for compile-time constants where you know the mode is valid.
func MustParseConfig(mode DetectionMode) ParseConfig {
	cfg := ParseConfig{Mode: mode}

	err := cfg.Validate()
	if err != nil {
		panic(err)
	}

	return cfg
}

// Validate returns nil if the config is valid, otherwise an error describing the problem.
func (cfg ParseConfig) Validate() error {
	if !cfg.Mode.IsValid() {
		return fmt.Errorf("%w: %q", ErrInvalidDetectionMode, cfg.Mode)
	}

	return nil
}

// IsSemantic returns true if semantic matching is enabled.
func (cfg ParseConfig) IsSemantic() bool {
	return cfg.Mode.IsSemantic()
}
