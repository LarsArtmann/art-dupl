# go-composable-business-types Usage Analysis for art-dupl

## Executive Summary

This document analyzes how the `github.com/larsartmann/go-composable-business-types/id` library can and should be integrated into **art-dupl** to improve type safety, reduce boilerplate, and standardize ID handling across the codebase.

**Current State:** art-dupl defines custom string-based ID types (`CloneGroupID`, `AnalysisID`) with manual validation and JSON serialization.

**Recommended Action:** **Use `go-composable-business-types/id` for external-facing IDs** while keeping internal string interning (`StringID`) for performance-critical internal structures.

---

## 1. Library Overview

The `go-composable-business-types/id` package provides:

- **Phantom-type branded IDs**: Compile-time prevention of ID mixing
- **Zero-overhead**: Direct wrapper around primitive types (no allocations)
- **Full serialization support**: JSON, SQL, Binary, Text, Gob
- **Type-safe operations**: Comparison, equality, zero-value checks
- **NanoId integration**: Optional cryptographically secure ID generation

### Installation

```bash
go get github.com/larsartmann/go-composable-business-types/id
```

---

## 2. Current ID Implementation in art-dupl

### 2.1 Existing ID Types

Located in `domain/types_id.go`:

```go
// CloneGroupID represents a unique identifier for a clone group.
type CloneGroupID string

// NewCloneGroupID creates a validated CloneGroupID from a string.
func NewCloneGroupID(id string) (CloneGroupID, error) {
    if id == "" {
        return "", errors.NewValidationError("clone group ID cannot be empty", nil)
    }
    return CloneGroupID(id), nil
}

// String returns the string representation.
func (id CloneGroupID) String() string { return string(id) }

// MarshalJSON implements json.Marshaler.
func (id CloneGroupID) MarshalJSON() ([]byte, error) {
    return marshalStringID(string(id), "clone group ID cannot be empty")
}

// UnmarshalJSON implements json.Unmarshaler.
func (id *CloneGroupID) UnmarshalJSON(data []byte) error {
    return unmarshalStringID(data, "CloneGroupID", "clone group ID cannot be empty", func(s string) {
        *id = CloneGroupID(s)
    })
}

// AnalysisID - similar 60+ lines of boilerplate
```

### 2.2 Analysis: Strengths and Weaknesses

| Aspect            | Current Implementation                             | id Package                                       |
| ----------------- | -------------------------------------------------- | ------------------------------------------------ |
| **Type Safety**   | ✅ Prevents mixing `CloneGroupID` and `AnalysisID` | ✅ Stronger phantom types prevent mixing ANY IDs |
| **Validation**    | ✅ Manual `New*` constructors                      | ✅ Built-in validation                           |
| **Serialization** | ✅ Custom JSON (50+ lines per type)                | ✅ Automatic (zero code)                         |
| **Boilerplate**   | ❌ ~60 lines per ID type                           | ✅ ~3 lines per ID type                          |
| **SQL Support**   | ❌ Not implemented                                 | ✅ Built-in Scanner/Valuer                       |
| **Comparison**    | ❌ Manual implementation needed                    | ✅ Built-in Compare method                       |
| **Zero Value**    | ❌ No standardized check                           | ✅ Built-in IsZero()                             |

### 2.3 StringID vs ID Package

**Important Distinction:**

- **`StringID` (uint32)**: Internal performance optimization for string interning
  - Used in `Clone.Filename`, `Clone.Fragment`, `Clone.Hash`
  - Reduces memory from 16B string header → 4B uint32
  - NOT a candidate for `id` package (not a domain entity ID)

- **`CloneGroupID`/`AnalysisID` (string)**: External domain entity identifiers
  - Used for API boundaries, JSON serialization, user-facing IDs
  - **Ideal candidates** for `id` package migration

---

## 3. Integration Strategy

### 3.1 Recommended Approach: Hybrid Model

```
┌─────────────────────────────────────────────────────────────┐
│                    External Interface                        │
│  (API, JSON, CLI, Database)                                  │
│                                                              │
│  CloneGroupID  →  id.ID[CloneGroupBrand, string]            │
│  AnalysisID    →  id.ID[AnalysisBrand, string]              │
│                                                              │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
┌─────────────────────────────────────────────────────────────┐
│                   Internal Processing                        │
│  (Suffix tree, AST analysis, Memory-optimized structures)    │
│                                                              │
│  StringID      →  Keep as uint32 (performance critical)     │
│  (Filename, Fragment, Hash interning)                        │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### 3.2 Migration Plan

#### Phase 1: Add Dependency

```bash
go get github.com/larsartmann/go-composable-business-types/id
```

#### Phase 2: Create Domain ID Types

Create `domain/id.go`:

```go
package domain

import (
    "github.com/larsartmann/go-composable-business-types/id"
)

// Brand types for compile-time ID separation
type CloneGroupBrand struct{}
type AnalysisBrand struct{}

// ID type aliases for convenience
type CloneGroupID = id.ID[CloneGroupBrand, string]
type AnalysisID = id.ID[AnalysisBrand, string]

// Constructor functions for validation
type ValidationError = id.ValidationError

func NewCloneGroupID(s string) (CloneGroupID, error) {
    if s == "" {
        return CloneGroupID{}, NewValidationError("clone group ID cannot be empty", nil)
    }
    return id.NewID[CloneGroupBrand](s), nil
}

func NewAnalysisID(s string) (AnalysisID, error) {
    if s == "" {
        return AnalysisID{}, NewValidationError("analysis ID cannot be empty", nil)
    }
    return id.NewID[AnalysisBrand](s), nil
}
```

#### Phase 3: Update Usage Sites

Replace manual string extraction:

```go
// BEFORE: Manual string conversion
id := CloneGroupID("group-123")
str := string(id)  // Direct cast

// AFTER: Type-safe Get()
id := id.NewID[CloneGroupBrand]("group-123")
str := id.Get()    // Explicit value extraction
```

#### Phase 4: Remove Old Implementation

Delete `domain/types_id.go` and `domain/helpers.go` (marshaling functions).

---

## 4. Complete Implementation Example

### 4.1 New `domain/id.go`

```go
package domain

import (
    "github.com/larsartmann/go-composable-business-types/id"
    "github.com/LarsArtmann/art-dupl/errors"
)

// Brand types (empty structs for compile-time differentiation)
type CloneGroupBrand struct{}
type AnalysisBrand struct{}

// Domain ID types (type aliases for convenience)
type CloneGroupID = id.ID[CloneGroupBrand, string]
type AnalysisID = id.ID[AnalysisBrand, string]

// NewCloneGroupID creates a validated clone group ID.
func NewCloneGroupID(s string) (CloneGroupID, error) {
    if s == "" {
        return CloneGroupID{}, errors.NewValidationError("clone group ID cannot be empty", nil)
    }
    return id.NewID[CloneGroupBrand](s), nil
}

// NewAnalysisID creates a validated analysis ID.
func NewAnalysisID(s string) (AnalysisID, error) {
    if s == "" {
        return AnalysisID{}, errors.NewValidationError("analysis ID cannot be empty", nil)
    }
    return id.NewID[AnalysisBrand](s), nil
}
```

### 4.2 Usage in Domain Types

```go
// domain/clone.go
// No changes needed - works seamlessly:
type CloneGroup struct {
    ID       CloneGroupID   `json:"id"`    // Still works
    Clones   []Clone        `json:"clones"`
    // ...
}

// domain/analysis.go
type Analysis struct {
    ID          AnalysisID    `json:"id"`   // Still works
    CloneGroups []CloneGroup  `json:"cloneGroups"`
    // ...
}
```

### 4.3 Serialization Example

```go
package domain_test

import (
    "encoding/json"
    "testing"

    "github.com/LarsArtmann/art-dupl/domain"
    "github.com/larsartmann/go-composable-business-types/id"
)

func TestCloneGroupIDSerialization(t *testing.T) {
    // Create ID
    gid, err := domain.NewCloneGroupID("group-123")
    if err != nil {
        t.Fatal(err)
    }

    // Serialize to JSON
    data, err := json.Marshal(gid)
    if err != nil {
        t.Fatal(err)
    }

    // Result: "group-123" (quoted string)
    if string(data) != `"group-123"` {
        t.Errorf("expected \"group-123\", got %s", string(data))
    }

    // Deserialize
    var restored domain.CloneGroupID
    err = json.Unmarshal(data, &restored)
    if err != nil {
        t.Fatal(err)
    }

    // Verify equality
    if !gid.Equal(restored) {
        t.Error("restored ID should equal original")
    }

    // Extract value
    if restored.Get() != "group-123" {
        t.Errorf("expected group-123, got %s", restored.Get())
    }
}

func TestZeroValueHandling(t *testing.T) {
    var zero domain.CloneGroupID

    // Zero value detection
    if !zero.IsZero() {
        t.Error("zero value should be zero")
    }

    // JSON serialization of zero → null
    data, _ := json.Marshal(zero)
    if string(data) != "null" {
        t.Errorf("expected null, got %s", string(data))
    }
}
```

---

## 5. Benefits Summary

### 5.1 Immediate Benefits

| Metric                   | Before                   | After      | Improvement       |
| ------------------------ | ------------------------ | ---------- | ----------------- |
| **Lines per ID type**    | ~60 lines                | ~3 lines   | **95% reduction** |
| **Files to maintain**    | types_id.go + helpers.go | id.go only | **50% reduction** |
| **Serialization code**   | Manual (error-prone)     | Automatic  | Zero bugs         |
| **SQL support**          | Not implemented          | Built-in   | Feature gain      |
| **Binary serialization** | Not implemented          | Built-in   | Feature gain      |

### 5.2 Long-term Benefits

1. **Standardization**: Aligns with ecosystem standards
2. **Maintainability**: Less code to maintain and test
3. **Extensibility**: Easy to add new ID types (1 line vs 60 lines)
4. **Compatibility**: Works with future database integrations
5. **Type Safety**: Phantom types prevent accidental mixing even across packages

---

## 6. Anti-Patterns to Avoid

### ❌ Don't Replace StringID with id Package

```go
// BAD: Using id for internal string interning
type Clone struct {
    Filename id.ID[FilenameBrand, StringID]  // Unnecessary wrapper
}

// GOOD: Keep StringID for performance
type Clone struct {
    Filename StringID  // uint32 - fast, compact
}
```

### ❌ Don't Force Library on SDK Consumers

```go
// BAD: SDK forces id package usage
func Analyze(paths []string) ([]id.ID[CloneGroupBrand, string], error)

// GOOD: SDK returns domain types (which use id internally)
func Analyze(paths []string) ([]CloneGroup, error)
```

### ❌ Don't Mix ID Types for Same Entity

```go
// BAD: Multiple brands for same entity
type CloneGroupBrand1 struct{}
type CloneGroupBrand2 struct{}

// GOOD: Single brand per entity
type CloneGroupBrand struct{}
```

---

## 7. Decision Matrix

| Component              | Current Type            | Recommended                      | Rationale                      |
| ---------------------- | ----------------------- | -------------------------------- | ------------------------------ |
| `CloneGroup.ID`        | `CloneGroupID` (string) | `id.ID[CloneGroupBrand, string]` | External API, needs validation |
| `Analysis.ID`          | `AnalysisID` (string)   | `id.ID[AnalysisBrand, string]`   | External API, needs validation |
| `Clone.Filename`       | `StringID` (uint32)     | **Keep as-is**                   | Performance-critical internal  |
| `Clone.Fragment`       | `StringID` (uint32)     | **Keep as-is**                   | Performance-critical internal  |
| `Clone.Hash`           | `StringID` (uint32)     | **Keep as-is**                   | Performance-critical internal  |
| Future: `RepositoryID` | N/A                     | `id.ID[RepositoryBrand, string]` | New ID type (easy addition)    |
| Future: `ProjectID`    | N/A                     | `id.ID[ProjectBrand, string]`    | New ID type (easy addition)    |

---

## 8. Migration Checklist

- [ ] Add `go get github.com/larsartmann/go-composable-business-types/id`
- [ ] Create `domain/id.go` with brand types and constructors
- [ ] Update imports in files using `CloneGroupID`/`AnalysisID`
- [ ] Replace `string(id)` with `id.Get()` where needed
- [ ] Delete `domain/types_id.go`
- [ ] Delete `marshalStringID`/`unmarshalStringID` helpers
- [ ] Update tests to use new constructors
- [ ] Run full test suite: `just test`
- [ ] Verify JSON serialization still works
- [ ] Update documentation references

---

## 9. Conclusion

The `go-composable-business-types/id` package is a **perfect fit** for art-dupl's external-facing ID types while **preserving** the performance-optimized `StringID` for internal structures.

**Recommended Action:** Proceed with migration for `CloneGroupID` and `AnalysisID`. This will:

1. **Reduce code** by ~120 lines (2 ID types × 60 lines)
2. **Add features** (SQL, Binary, Text serialization)
3. **Improve safety** (phantom types prevent cross-entity mixing)
4. **Maintain performance** (StringID remains unchanged)

**Not Recommended:** Replacing `StringID` (uint32 interning) - this is correctly optimized for internal memory layout.

---

## 10. References

- [go-composable-business-types/id README](/Users/larsartmann/projects/go-composable-business-types/id/README.md)
- [art-dupl domain types](/Users/larsartmann/projects/art-dupl/domain/types_id.go)
- [art-dupl stringpool](/Users/larsartmann/projects/art-dupl/domain/stringpool.go)
- [Library usage guide](/Users/larsartmann/projects/go-composable-business-types/docs/planning/go-composable-business-types-usage.md)
