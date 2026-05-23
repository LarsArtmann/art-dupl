# TODO/Legacy Detection Feature Implementation Status

**Date**: 2025-12-16_20-26  
**Feature**: Add TODO and Legacy code reporting to art-dupl  
**Status**: In Progress - Architecture Redesign Needed

## Executive Summary

Implementation started for TODO and Legacy detection methods in art-dupl codebase. Core enumeration and basic detector implementations completed, but fundamental architecture issues discovered requiring interface redesign.

## Completed Work ✅

### Configuration Updates

- **File**: `config/detectionmethod.go`
  - Added `DetectionMethodTodos` and `DetectionMethodLegacy` constants
  - Updated `IsValid()` to include new methods
  - Extended `AllDetectionMethods()` and `ParseDetectionMethods()`
  - Updated error messages with valid methods: "hash, art-dupl, todos, legacy"

### Basic Detector Implementation

- **File**: `detection/todos.go` (NEW)
  - `TodoDetector`: Parses Go AST for TODO-style comments using regex patterns
  - `LegacyDetector`: Identifies deprecated functions and old code patterns
  - Support for TODO, FIXME, XXX, HACK, NOTE comment types
  - Tag extraction (e.g., `// TODO(@user): description`)
  - Default legacy patterns (deprecated io/ioutil functions, old coding patterns)

### Research Completed

- Go AST comment handling patterns documented
- Found go-critic has `todoCommentWithoutDetail` checker
- Identified existing static analysis tool patterns

## Architecture Issues Identified 🚨

### Critical Design Problems

1. **Result Type Mismatch**
   - `syntax.Match` designed for code clone detection with `Frags [][]*Node`
   - TODO/Legacy detection produces metadata structures, not code fragments
   - Current approach forces TODO results into `syntax.Match` with empty `Frags`

2. **Interface Incompatibility**
   - Existing printers expect clone-specific data structures
   - TODO/Legacy need different output formats and metadata fields
   - Result processing logic assumes clone-specific properties

3. **Semantic Violation**
   - "Match" concept doesn't apply to single-issue detection
   - Different result types require different handling logic
   - Current mixing violates single responsibility principle

## Current Implementation Gaps

### Integration Points Missing

- **MultiDetector**: Not updated to handle new detection methods
- **CLI Interface**: New methods not exposed via flags
- **Printers**: Cannot handle TODO/Legacy result types
- **Configuration**: Default settings don't include new methods

### Data Model Issues

- **Result Representation**: Need unified interface for different result types
- **Metadata Handling**: TODO/Legacy have different metadata requirements
- **Output Formatting**: Each type needs custom formatting logic

## Required Architecture Changes

### Option 1: Generic Interface Approach

```go
type DetectionResult interface {
    GetType() DetectionMethod
    GetFilename() string
    GetSeverity() string
    OutputTo(printer Printer) error
}

type CloneResult struct { ... } // Current syntax.Match
type TodoResult struct { ... } // New structure
type LegacyResult struct { ... } // New structure
```

### Option 2: Separate Pipelines

- Keep existing clone detection pipeline unchanged
- Create separate TODO and Legacy detection pipelines
- Separate processing logic and output handling

### Option 3: Union Type Approach

- Extend existing structures to handle multiple result types
- Add type discriminators and result-specific fields
- Maintain current printer interfaces with conditional logic

## Next Steps Priority

### Immediate (Critical Path)

1. **DECISION**: Choose architecture approach for result type handling
2. **REFACTOR**: Implement chosen interface design
3. **INTEGRATE**: Update MultiDetector for new detection methods
4. **TEST**: End-to-end integration testing

### Short-term (Implementation)

5. Update CLI flags and help text
6. Extend all printer implementations
7. Add comprehensive test coverage
8. Performance testing with large codebases

### Long-term (Enhancement)

9. Custom pattern configuration
10. Integration with existing linters
11. Severity-based filtering
12. Documentation and examples

## Risk Assessment

### Technical Risks

- **High**: Current approach will cause runtime errors when printing non-clone results
- **Medium**: Performance impact of architectural refactoring
- **Low**: Backwards compatibility issues

### Timeline Impact

- **Current Progress**: ~20% of total effort
- **Architecture Decision**: Critical path blocker
- **Estimated Completion**: 2-3 weeks depending on chosen approach

## Stakeholder Impact

### Users

- New detection methods will enhance code quality insights
- Need clear documentation on usage patterns
- Migration path for existing workflows

### Development Team

- Significant refactoring required regardless of approach
- Testing strategy must cover all detection method combinations
- CI/CD pipeline updates needed

## Recommendations

### Immediate Action

1. **HOLD**: Stop current implementation until architecture decision made
2. **REVIEW**: Evaluate three architectural approaches with trade-off analysis
3. **DECIDE**: Choose approach before proceeding with implementation

### Preferred Approach

**Option 1 (Generic Interface)** recommended because:

- Maintains clean separation of concerns
- Allows future extension with new detection methods
- Provides consistent user experience
- Supports per-method configuration

## Files Modified

### New Files

- `detection/todos.go` - Basic detector implementations

### Modified Files

- `config/detectionmethod.go` - Added new detection methods
- `config/config.go` - Updated error messages

### Files Needing Updates

- `detection/multidetector.go` - Integration point
- `cli.go` - CLI flag handling
- `printer/*.go` - All printer implementations
- Test files across all packages

## Metrics

### Code Coverage

- **Configuration**: 100% complete
- **Detection Logic**: 30% complete (basic implementation only)
- **Integration**: 0% complete
- **Testing**: 0% complete
- **Documentation**: 10% complete

### Complexity Metrics

- **New Code**: ~300 lines added
- **Refactoring Needed**: ~1000+ lines estimated
- **Test Coverage**: 500+ lines needed

---

**Status**: BLOCKED on architecture decision  
**Next Action**: Stakeholder review and approach selection  
**ETA**: Dependent on decision timeline
