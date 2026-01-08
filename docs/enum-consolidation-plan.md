# Enum Utilities Consolidation Plan - Hybrid Approach

**Date**: 2026-01-08 03:00
**Status**: Phase 1 - Documentation and Clarity (Low Risk)
**Next**: Phase 2 - Unification Interface (Medium Risk)
**Goal**: Full Consolidation (Phase 3 - High Risk)

---

## Problem Statement

### Current State: Split Brain

We have **TWO enum utility packages** with **same function names, different APIs**:

#### config/unmarshal_helper.go
```go
// Generic constraint: T ~string
// Validation: Passed as parameter func(T) bool
// Return type: T (not pointer)
// Used by: config package enums (DetectionMethod, OutputFormat, SortCriteria)

func MarshalEnumJSON[T ~string](value T, isValid func(T) bool, typeName string) ([]byte, error)
func UnmarshalEnumJSON[T ~string](data []byte, enumType func(string) T, isValid func(T) bool, typeName string) (T, error)
```

**Usage Pattern**:
```go
func (dm DetectionMethod) MarshalJSON() ([]byte, error) {
    return MarshalEnumJSON(dm, DetectionMethod.IsValid, "detection method")
}
```

#### types/enum_utils.go
```go
// Generic constraint: T ValidatableEnum interface
// Validation: Inferred from interface method
// Return type: *T (pointer)
// Used by: types package enums (DetectionState, AnalysisMode, FileProcessingState)

func MarshalEnumJSON[T ValidatableEnum](enum T, typeName string) ([]byte, error)
func UnmarshalEnumJSON[T ValidatableEnum](data []byte, constructor func(string) T, typeName string) (*T, error)
```

**Usage Pattern**:
```go
func (ds DetectionState) MarshalJSON() ([]byte, error) {
    data, err := MarshalEnumJSON(ds, "detection state")
    if err != nil {
        return nil, fmt.Errorf("failed to marshal detection state: %w", err)
    }
    return data, nil
}
```

### Why This is Bad

1. **Same Names, Different Signatures**: Developer confusion when switching between packages
2. **No Single Source of Truth**: Two implementations to maintain
3. **Inconsistent Error Messages**: Different format and content
4. **Incompatible Return Types**: T vs *T makes code non-portable
5. **Different Validation Approaches**: Method value vs interface inference

---

## Hybrid Approach: Three-Phase Consolidation

### Phase 1: Documentation & Clarity (Low Risk, Do Now)

**Goal**: Make separation intentional and well-documented without breaking changes

**Actions**:

1. ✅ Add package-level documentation to both files explaining the split
2. ✅ Add deprecation notices for future consolidation
3. ✅ Create this migration plan document
4. ✅ Document differences in APIs

**Benefits**:
- Zero breaking changes
- Immediate clarity for developers
- Clear path forward
- Minimal effort (1-2 hours)

**Risks**:
- None (documentation only)

---

### Phase 2: Unification Interface (Medium Risk, Next Sprint)

**Goal**: Create shared interface that both packages can implement

**Actions**:

1. Create `pkg/enum/interface.go` with shared interface:
```go
// Package enum provides shared interfaces and utilities for enum marshaling.
package enum

// Marshaler defines the interface for types that can marshal themselves.
type Marshaler interface {
    MarshalJSON() ([]byte, error)
}

// Unmarshaler defines the interface for types that can unmarshal from JSON.
type Unmarshaler[T any] interface {
    UnmarshalJSON([]byte) error
}

// Validatable defines the interface for types that can validate themselves.
type Validatable interface {
    IsValid() bool
}

// Enum is the complete interface for enum-like types.
type Enum[T any] interface {
    Validatable
    Marshaler
    Unmarshaler[T]
    String() string
}
```

2. Update both packages to implement `Enum[T]` interface
3. Create adapter functions for backward compatibility
4. Add migration tests

**Benefits**:
- Shared contract without full consolidation
- Can verify both implementations work correctly
- Incremental migration possible
- Clear interface for documentation

**Risks**:
- Medium (interface changes require testing)
- No code reduction yet (still two implementations)

---

### Phase 3: Full Consolidation (High Risk, Later)

**Goal**: Single implementation in shared package

**Actions**:

1. Design unified API combining best of both approaches:
```go
// pkg/enum/marshal.go
package enum

import (
    "fmt"
    "github.com/LarsArtmann/art-dupl/errors"
)

// MarshalJSON marshals any validatable enum type.
func MarshalJSON[T Validatable](value T, typeName string) ([]byte, error) {
    if !value.IsValid() {
        validationErr := fmt.Errorf("invalid %s value: %v", typeName, value)
        return nil, errors.NewValidationError("enum validation failed", validationErr)
    }
    return fmt.Appendf(nil, `"%s"`, value), nil
}

// UnmarshalJSON unmarshals JSON into any validatable enum type.
func UnmarshalJSON[T Validatable](data []byte, constructor func(string) T, typeName string) (*T, error) {
    str := string(data)
    if len(str) >= 2 && str[0] == '"' && str[len(str)-1] == '"' {
        str = str[1 : len(str)-1]
    }

    enum := constructor(str)
    if !enum.IsValid() {
        validationErr := fmt.Errorf("invalid %s value: %q (data: %q)", typeName, str, string(data))
        return nil, errors.NewValidationError("enum validation failed", validationErr)
    }
    return &enum, nil
}

// Validatable is the constraint for enum types that can validate themselves.
type Validatable interface {
    IsValid() bool
}
```

2. Migrate config package enums:
```go
// Before (config/detectionmethod.go)
func (dm DetectionMethod) MarshalJSON() ([]byte, error) {
    return MarshalEnumJSON(dm, DetectionMethod.IsValid, "detection method")
}

// After
func (dm DetectionMethod) MarshalJSON() ([]byte, error) {
    return enum.MarshalJSON(dm, "detection method")
}
```

3. Migrate types package enums:
```go
// Before (types/enums.go)
func (ds DetectionState) MarshalJSON() ([]byte, error) {
    data, err := MarshalEnumJSON(ds, "detection state")
    if err != nil {
        return nil, fmt.Errorf("failed to marshal detection state: %w", err)
    }
    return data, nil
}

// After
func (ds DetectionState) MarshalJSON() ([]byte, error) {
    return enum.MarshalJSON(ds, "detection state")
}
```

4. Remove old implementations:
   - Delete `config/unmarshal_helper.go`
   - Delete `types/enum_utils.go`
   - Update imports

**Benefits**:
- Single source of truth
- Consistent error handling
- Eliminates split brain completely
- Easier to maintain
- Better type safety

**Risks**:
- High (breaking change for both packages)
- Requires comprehensive testing
- Circular dependency concerns
- Large amount of code to change

**Estimated Effort**: 1-2 days

---

## Migration Strategy

### Rollout Plan

**Week 1**: Phase 1 (Documentation)
- Add package docs
- Create migration plan
- Review with team

**Week 2-3**: Phase 2 (Interface)
- Create shared interface
- Update both packages
- Add migration tests
- Monitor for issues

**Week 4-5**: Phase 3 (Consolidation)
- Create shared implementation
- Migrate one package (config)
- Test thoroughly
- Migrate other package (types)
- Full test suite
- Remove old code

### Rollback Plan

If Phase 3 causes issues:
- Revert to Phase 2 (interface-based approach)
- Keep both implementations for now
- Re-evaluate consolidation strategy

---

## Decision Matrix

| Approach | Risk | Effort | Benefits | When to Use |
|-----------|-------|---------|-----------|--------------|
| **Phase 1: Docs Only** | None | 1-2 hours | Clarity, minimal changes | **NOW** - Immediate |
| **Phase 2: Interface** | Medium | 2-3 days | Shared contract, incremental | **NEXT SPRINT** |
| **Phase 3: Consolidate** | High | 1-2 weeks | Single source, complete fix | **QUARTER** |
| **Option A (Full Now)** | High | 1-2 days | Best long-term, single effort | Only if no constraints |
| **Option B (Keep Forever)** | None | 0 hours | Safe, but technical debt | Not recommended |
| **Option C (Coupling)** | Medium | 2-4 hours | Quick, but bad design | Not recommended |

---

## Current Status

### Phase 1 Progress
- [x] Analyzed both enum utility packages
- [ ] Add package-level documentation to config/unmarshal_helper.go
- [ ] Add package-level documentation to types/enum_utils.go
- [ ] Add deprecation notices
- [ ] Create migration plan document (this file)

### Phase 2 Planning
- [ ] Design shared interface
- [ ] Define Enum[T] contract
- [ ] Plan migration tests
- [ ] Estimate effort for Phase 3

### Phase 3 Planning
- [ ] Define unified API
- [ ] Plan migration order (config first, then types)
- [ ] Identify circular dependency risks
- [ ] Create comprehensive test plan

---

## Key Differences Between Implementations

### config/unmarshal_helper.go (Method Value Approach)

**Pros**:
- Explicit about validation (passed as parameter)
- Works with types that don't have IsValid() method
- More flexible (validation can be external)

**Cons**:
- Requires passing validation function explicitly
- More verbose calling convention
- Can accidentally pass wrong validation function

**Signature**:
```go
func MarshalEnumJSON[T ~string](value T, isValid func(T) bool, typeName string) ([]byte, error)
```

### types/enum_utils.go (Interface Approach)

**Pros**:
- Cleaner calling convention (validation inferred)
- More idiomatic Go (interface-based)
- Less verbose
- Validates that type has IsValid() method at compile time

**Cons**:
- Requires all types to implement IsValid() method
- Less flexible (validation must be built-in)
- Returns *T (pointer) which is inconsistent

**Signature**:
```go
func MarshalEnumJSON[T ValidatableEnum](enum T, typeName string) ([]byte, error)
```

---

## Recommendations

### For Phase 1 (Immediate)
- Add clear documentation to both packages
- Explain this is TEMPORARY and will be consolidated
- Add deprecation notices
- No code changes (minimize risk)

### For Phase 2 (Next Sprint)
- Create `pkg/enum` with shared interface
- Update both packages to implement interface
- Add adapter functions for backward compatibility
- Write migration tests

### For Phase 3 (Consolidation)
- Use interface approach (types/enum_utils.go) as base
- Keep return type as *T (more Go-idiomatic)
- Add richer error messages from config version
- Create comprehensive test suite
- Migrate incrementally, test thoroughly

### Long-Term
- Consider code generation for enum marshaling
- Use go:generate to create MarshalJSON/UnmarshalJSON automatically
- Reduce boilerplate further

---

## Success Criteria

### Phase 1 Success
- [ ] Developers understand the split brain issue
- [ ] Documentation clearly explains the situation
- [ ] Migration path is documented
- [ ] No breaking changes

### Phase 2 Success
- [ ] Shared interface exists and is clear
- [ ] Both packages implement the interface
- [ ] Tests verify compatibility
- [ ] Code compiles and passes all tests

### Phase 3 Success
- [ ] Single implementation in pkg/enum
- [ ] No duplicate code
- [ ] All tests pass
- [ ] Error handling is consistent
- [ ] Performance is maintained or improved
- [ ] Documentation is updated

---

## Open Questions

1. **Timeline**: When should we start Phase 2? (Recommendation: Next sprint after Phase 1 complete)
2. **Risk Tolerance**: Can we accept Phase 3 high risk? (Recommendation: Yes, with comprehensive testing)
3. **Team Review**: Has team reviewed this plan? (Recommendation: Get buy-in before Phase 2)
4. **Dependencies**: Are there external packages depending on these? (Recommendation: Check for external usage)

---

**💘 Generated with Crush**

**Assisted-by**: GLM-4.7 via Crush <crush@charm.land>
