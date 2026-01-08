package types

// TEMPORARY: This package contains internal enum marshaling utilities for types package enums.
//
// ⚠️ DEPRECATION NOTICE: This functionality will be moved to pkg/enum package.
// See docs/enum-consolidation-plan.md for migration roadmap.
//
// This file implements JSON marshaling/unmarshaling for types package enums
// (DetectionState, AnalysisMode, FileProcessingState). It uses an interface-based
// approach where validation is inferred from the ValidatableEnum interface.
//
// Note: There is a similar implementation in config/unmarshal_helper.go with a
// different API. This split-brain situation is temporary and will be resolved.
// Please do not add new types to this utility - see consolidation plan.
//
// Current Usage:
// - types.DetectionState.MarshalJSON()
// - types.AnalysisMode.MarshalJSON()
// - types.FileProcessingState.MarshalJSON()
//
// Migration Target: pkg/enum package (see Phase 3 in consolidation plan)

import (
	"fmt"
)

// ValidatableEnum interface for enums that can validate themselves.
type ValidatableEnum interface {
	IsValid() bool
}

// UnmarshalEnumJSON provides generic unmarshaling for enum types.
func UnmarshalEnumJSON[T ValidatableEnum](data []byte, constructor func(string) T, typeName string) (*T, error) {
	str := string(data)
	if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
		str = str[1 : len(str)-1]
	}

	enum := constructor(str)
	if !enum.IsValid() {
		return nil, fmt.Errorf("enum validation failed for %s: invalid value %q (data: %q)", typeName, str, string(data))
	}
	return &enum, nil
}

// MarshalEnumJSON provides generic marshaling for enum types.
func MarshalEnumJSON[T ValidatableEnum](enum T, typeName string) ([]byte, error) {
	if !enum.IsValid() {
		return nil, fmt.Errorf("failed to marshal %s: invalid value %v (value must pass validation)", typeName, enum)
	}
	return []byte(`"` + fmt.Sprintf("%v", enum) + `"`), nil
}

// StringEnum is a generic type for string-based enums.
type StringEnum string

func (se StringEnum) String() string {
	return string(se)
}
