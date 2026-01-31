package errors

//
// DOMAIN TYPES STATUS:
// ✅ Added SafeMarshalConfig for config.Config
// ✅ Added SafeMarshalClone for domain.Clone
// ✅ Added SafeMarshalCloneGroup for domain.CloneGroup
// ✅ Added SafeMarshalAnalysis for domain.Analysis
//
// TYPE SAFETY ENHANCEMENT: Typed marshaling functions provide:
// - Compile-time type safety (can't pass wrong type)
// - Better IDE autocomplete (specific functions)
// - Reduced reflection overhead (less interface{})
// - Self-documenting code (clear intent)
//
// Also: Consider using generics for type-safe marshaling:
//   func SafeMarshalTyped[T any](v T, context string) ([]byte, error)

import (
	"encoding/json"
	"fmt"

	"github.com/LarsArtmann/art-dupl/config"
	"github.com/LarsArtmann/art-dupl/domain"
)

// MarshalError is a specialized error for JSON marshaling failures.
type MarshalError struct {
	Operation string
	Context   string
	Cause     error
}

func (e *MarshalError) Error() string {
	return fmt.Sprintf("JSON %s error for %s: %v", e.Operation, e.Context, e.Cause)
}

func (e *MarshalError) Unwrap() error {
	return e.Cause
}

// HandleMarshalingError provides unified JSON marshaling error handling.
func HandleMarshalingError(operation, context string, err error) error {
	if err == nil {
		return nil
	}

	switch err.Error() {
	case "json: unsupported value":
		return &MarshalError{Operation: operation, Context: context, Cause: fmt.Errorf("unsupported value type: %w", err)}
	case "json: unsupported type":
		return &MarshalError{Operation: operation, Context: context, Cause: fmt.Errorf("unsupported type: %w", err)}
	case "json: invalid UTF-8":
		return &MarshalError{Operation: operation, Context: context, Cause: fmt.Errorf("invalid UTF-8 encoding: %w", err)}
	default:
		return &MarshalError{Operation: operation, Context: context, Cause: err}
	}
}

// SafeMarshal provides safe marshaling with consistent error handling.
func SafeMarshal(v any, context string) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, HandleMarshalingError("marshal", context, err)
	}
	return data, nil
}

// SafeMarshalIndent provides safe indented marshaling with consistent error handling.
func SafeMarshalIndent(v any, prefix, indent, context string) ([]byte, error) {
	data, err := json.MarshalIndent(v, prefix, indent)
	if err != nil {
		return nil, HandleMarshalingError("marshal indent", context, err)
	}
	return data, nil
}

// SafeUnmarshal provides safe unmarshaling with consistent error handling.
func SafeUnmarshal(data []byte, v any, context string) error {
	err := json.Unmarshal(data, v)
	if err != nil {
		return HandleMarshalingError("unmarshal", context, err)
	}
	return nil
}

// SafeMarshalConfig provides type-safe marshaling for config.Config.
//
// Usage:
//	cfg := config.DefaultConfig()
//	data, err := SafeMarshalConfig(cfg, "config loading")
//	if err != nil { ... }
func SafeMarshalConfig(cfg *config.Config) ([]byte, error) {
	if cfg == nil {
		return nil, NewValidationError("config cannot be nil", nil)
	}
	data, err := json.Marshal(cfg)
	if err != nil {
		return HandleMarshalingError("marshal", "config.Config", err)
	}
	return data, nil
}

// SafeMarshalClone provides type-safe marshaling for domain.Clone.
//
// Usage:
//	clone := domain.Clone{...}
//	data, err := SafeMarshalClone(&clone, "clone marshaling")
//	if err != nil { ... }
func SafeMarshalClone(c *domain.Clone) ([]byte, error) {
	if c == nil {
		return nil, NewValidationError("clone cannot be nil", nil)
	}
	data, err := json.Marshal(c)
	if err != nil {
		return HandleMarshalingError("marshal", "domain.Clone", err)
	}
	return data, nil
}

// SafeMarshalCloneGroup provides type-safe marshaling for domain.CloneGroup.
//
// Usage:
//	group := domain.CloneGroup{...}
//	data, err := SafeMarshalCloneGroup(&group, "clone group marshaling")
//	if err != nil { ... }
func SafeMarshalCloneGroup(g *domain.CloneGroup) ([]byte, error) {
	if g == nil {
		return nil, NewValidationError("clone group cannot be nil", nil)
	}
	data, err := json.Marshal(g)
	if err != nil {
		return HandleMarshalingError("marshal", "domain.CloneGroup", err)
	}
	return data, nil
}
// SafeMarshalAnalysis provides type-safe marshaling for domain.Analysis.
//
// Usage:
//	analysis := domain.Analysis{...}
//	data, err := SafeMarshalAnalysis(&analysis, "analysis marshaling")
//	if err != nil { ... }
func SafeMarshalAnalysis(a *domain.Analysis) ([]byte, error) {
	if a == nil {
		return nil, NewValidationError("analysis cannot be nil", nil)
	}
	data, err := json.Marshal(a)
	if err != nil {
		return HandleMarshalingError("marshal", "domain.Analysis", err)
	}
	return data, nil
}
