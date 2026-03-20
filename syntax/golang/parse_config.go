package golang

import "fmt"

// ParseConfig holds configuration for the Go source code parser.
// It controls how AST nodes are transformed and matched.
type ParseConfig struct {
	// Mode determines whether semantic matching is enabled.
	// SemanticMode: Clones matched by structure AND identifier semantics (default).
	// StructuralMode: Clones matched by AST structure only.
	Mode DetectionMode
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
	if err := cfg.Validate(); err != nil {
		panic(err)
	}

	return cfg
}

// Validate returns nil if the config is valid, otherwise an error describing the problem.
func (cfg ParseConfig) Validate() error {
	if !cfg.Mode.IsValid() {
		return fmt.Errorf("invalid detection mode: %q", cfg.Mode)
	}

	return nil
}

// IsSemantic returns true if semantic matching is enabled.
func (cfg ParseConfig) IsSemantic() bool {
	return cfg.Mode.IsSemantic()
}

// SemanticHashEnabled controls whether semantic-aware hashing is enabled.
// This is a package-level setting used by the default parser.
// Deprecated: Use ParseConfig with explicit DetectionMode instead.
var SemanticHashEnabled bool

// SetDefaultParseConfig sets the default parse configuration for the package.
// This affects all subsequent Parse() calls that use DefaultParseConfig().
func SetDefaultParseConfig(cfg ParseConfig) {
	SemanticHashEnabled = cfg.Mode.IsSemantic()
}
