package golang

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

// IsValid returns nil if the config is valid, otherwise an error describing the problem.
func (cfg ParseConfig) IsValid() error {
	if !cfg.Mode.IsValid() {
		return ErrInvalidDetectionMode{Mode: cfg.Mode}
	}
	return nil
}

// ErrInvalidDetectionMode is returned when an invalid DetectionMode is provided.
type ErrInvalidDetectionMode struct {
	Mode DetectionMode
}

func (e ErrInvalidDetectionMode) Error() string {
	return "invalid detection mode: " + string(e.Mode)
}
