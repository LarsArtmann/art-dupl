# Architecture Review - art-dupl

**Date:** 2026-02-11
**Review Type:** Comprehensive Type Safety, DDD, and Code Quality Review
**Standard:** Extremely High (Principal Engineer Level)

---

## Executive Summary

The art-dupl codebase demonstrates **excellent domain-driven design** with a strong type system, good separation of concerns, and comprehensive error handling. However, several files violate the 350-line limit, and there are opportunities for improved type safety, reduced duplication, and better architectural organization.

### Key Strengths

- ✅ Excellent domain type system with strong typing (domain/domain_types.go)
- ✅ Well-organized error handling with typed errors (errors/types.go)
- ✅ Good use of enums throughout the codebase
- ✅ Proper StringPool implementation for memory efficiency
- ✅ Memory-optimized types (uint16 for LineNumber, uint32 for BytePosition)

### Critical Issues

- ❌ **6 files exceed 350-line limit** (needs immediate attention)
- ❌ **Hash-based detection not implemented** (delegates to suffix tree)
- ❌ **Type safety violations** in conversion functions
- ❌ **Duplicate SQLC filtering logic** (applied twice)
- ❌ **Code duplications** across multiple files

---

## 1. Files Exceeding 350-Line Limit

| File                    | Lines | Issue                                                                                                                     | Priority   |
| ----------------------- | ----- | ------------------------------------------------------------------------------------------------------------------------- | ---------- |
| printer/stats.go        | 727   | Multi-responsibility: collection, health scoring, formatting, visualization, recommendations                              | **HIGH**   |
| pkg/artdupl/detector.go | 546   | Multi-responsibility: lifecycle, orchestration, pipeline, detection, conversion, validation, content extraction, progress | **HIGH**   |
| cmd/run.go              | 528   | Multi-responsibility: CLI handling, analysis orchestration, output formatting, file crawling, all-modes execution         | **HIGH**   |
| domain/domain_types.go  | 525   | Large aggregation of type definitions (well-structured but too large)                                                     | **MEDIUM** |
| domain/clone.go         | 495   | Multi-responsibility: enums, Clone/CloneGroup, Analysis, Repository, conversion functions                                 | **MEDIUM** |
| pkg/filter/filter.go    | 461   | Multi-responsibility: enums, metrics, Filter struct, detection logic, pattern matching                                    | **MEDIUM** |

### Detailed Recommendations

#### printer/stats.go (727 lines)

**Current structure:**

- Statistics collection and formatting
- Health score calculation
- Size distribution and file ranking
- Visualization helpers
- Recommendations generation
- Style management

**Suggested split:**

- `printer/stats_collector.go`: Statistics collection
- `printer/stats_health.go`: Health score calculation
- `printer/stats_formatter.go`: Format-specific output
- `printer/stats_visualization.go`: Visualization helpers
- `printer/stats_recommendations.go`: Recommendation logic
- `printer/stats_styles.go`: Style management

#### pkg/artdupl/detector.go (546 lines)

**Current structure:**

- Detector lifecycle (NewDetector, Close)
- Analysis orchestration (FindClones, FindClonesStream)
- Pipeline construction (buildAnalysisPipeline)
- Detection execution (runDetection, streamDetectionResults)
- Result conversion (convertToCloneGroup, convertFragmentToClone)
- File validation (validateFile)
- Content extraction (extractFragmentContent)
- Progress reporting (reportProgress)
- Configuration conversion (convertOptionsToConfig, hashConfig)

**Suggested split:**

- `pkg/artdupl/detector.go`: Core detector struct and public API
- `pkg/artdupl/detector_pipeline.go`: Pipeline construction and execution
- `pkg/artdupl/detector_conversion.go`: Result conversion and formatting
- `pkg/artdupl/detector_validation.go`: Input validation
- `pkg/artdupl/detector_utils.go`: Helper functions (hashConfig, reportProgress)

#### cmd/run.go (528 lines)

**Current structure:**

- Flag parsing and config setup
- Analysis execution
- Output handling (printDupls, createPrinter)
- File crawling (crawlPaths, filesFeedWithOptions)
- All modes execution (runAllModes)

**Suggested split:**

- `cmd/run_flags.go`: Flag parsing and config setup
- `cmd/run_analysis.go`: Analysis execution
- `cmd/run_output.go`: Output handling
- `cmd/run_crawl.go`: File crawling
- `cmd/run_all_modes.go`: All modes execution

#### domain/domain_types.go (525 lines)

**Current structure:**

- 13 domain type definitions (CloneGroupID, AnalysisID, Filepath, LineNumber, BytePosition, TokenCount, Confidence, ComplexityScore, Hash, FileCount, CloneCount, ProcessingTime, Threshold)
- Helper functions for marshaling/unmarshaling

**Suggested split:**

- `domain/types_id.go`: ID types (CloneGroupID, AnalysisID)
- `domain/types_file.go`: File-related types (Filepath, LineNumber, BytePosition)
- `domain/types_metric.go`: Metric types (TokenCount, FileCount, CloneCount, Threshold)
- `domain/types_metadata.go`: Metadata types (Hash, ComplexityScore, Confidence, ProcessingTime)
- `domain/helpers.go`: Marshaling/unmarshaling helper functions

#### domain/clone.go (495 lines)

**Current structure:**

- Enum types (FileProcessingState, DetectionState, AnalysisMode, CloneSeverity)
- Clone and CloneGroup domain objects
- Analysis and AnalysisStats
- Repository and SourceFile
- DetectionOptions
- NodeToClone, CalculateSeverity, calculateComplexity
- Validation helpers

**Suggested split:**

- `domain/types_enums.go`: Enum types
- `domain/clone.go`: Clone and CloneGroup domain objects
- `domain/analysis.go`: Analysis and AnalysisStats
- `domain/repository.go`: Repository and SourceFile
- `domain/options.go`: DetectionOptions
- `domain/conversion.go`: NodeToClone, CalculateSeverity, calculateComplexity
- `domain/validation.go`: Validation helpers

#### pkg/filter/filter.go (461 lines)

**Current structure:**

- Enum types (FilterOption, FilterReason)
- Metrics and FilterStats
- Filter struct and main filtering logic
- Auto-generated detection (SQLC, Templ)
- Pattern matching utilities

**Suggested split:**

- `filter/types.go`: Enum types and FilterStats
- `filter/metrics.go`: Metrics type with thread-safe tracking
- `filter/filter.go`: Filter struct and main filtering logic
- `filter/detection.go`: Auto-generated detection
- `filter/pattern.go`: Pattern matching utilities

---

## 2. Type Safety Issues

### 2.1 Primitive Types vs Domain Types

#### Issue: SDK Conversion Functions Use Primitives

**Location:** `pkg/artdupl/detector.go:398-423`

**Problem:**

```go
func (d *detector) convertFragmentToClone(frag []*syntax.Node) *Clone {
    // ...
    clone := &Clone{
        Filename:  firstNode.Filename,  // string instead of domain.Filepath
        StartLine: int(firstNode.Pos),    // int instead of domain.LineNumber
        EndLine:   int(lastNode.End),     // int instead of domain.LineNumber
        StartPos:  int(firstNode.Pos),    // int instead of domain.BytePosition
        EndPos:    int(lastNode.End),     // int instead of domain.BytePosition
        Size:      len(frag),              // int instead of domain.TokenCount
    }
}
```

**Impact:**

- Loses type safety guarantees provided by domain types
- Allows invalid values (e.g., negative line numbers)
- Violates DDD principle of using domain types throughout

**Recommendation:**
Change Clone struct to use domain types:

```go
type Clone struct {
    Filename  domain.Filepath    `json:"filename"`
    StartLine domain.LineNumber  `json:"startLine"`
    EndLine   domain.LineNumber  `json:"endLine"`
    // etc.
}
```

#### Issue: StatsData Uses Primitives

**Location:** `printer/stats_data.go`

**Problem:**

```go
type StatsData struct {
    TotalFiles    int                    // Should be domain.FileCount
    TotalClones   int                    // Should be domain.CloneCount
    TotalTokens   int                    // Should be domain.TokenCount
    HealthScore   string                 // Should be domain.HealthGrade (new type)
    // ...
}
```

**Impact:**

- Type safety lost in statistics reporting
- No validation on values (e.g., negative counts)
- HealthScore as string instead of typed enum

**Recommendation:**

1. Create `domain.HealthGrade` enum (A, B, C, D, F)
2. Use domain types throughout StatsData
3. Add validation at construction time

### 2.2 Deprecated Methods (Technical Debt)

#### Issue: Uint() Methods for Type Conversion

**Location:** `domain/domain_types.go:199-203, 241-244, 350-353`

**Problem:**

```go
// Uint returns the underlying uint value (for backward compatibility).
// Deprecated: Use Uint16() instead for type safety.
func (ln LineNumber) Uint() uint {
    return uint(ln)
}
```

**Impact:**

- Technical debt that weakens type safety
- Allows callers to bypass strongly-typed accessors
- No migration plan for removal

**Recommendation:**

1. Audit usage of `.Uint()` across codebase
2. Replace all calls with `.Uint16()` or `.Uint32()`
3. Mark with deprecation comment and timeline for removal
4. Remove after migration is complete

---

## 3. Duplications and Split Brains

### 3.1 Duplicate Logic: SQLC Filtering

**Location:** `cmd/run.go:319-347`

**Problem:** SQLC filtering logic appears twice, adding `FilterSQLC` to filterOptions both times.

**First instance (lines 319-338):**

```go
if len(sqlcOutputDirs) > 0 && !cfg.IncludeSQLC {
    filterOptions = append(filterOptions, filter.FilterSQLC)
    // ...
}
```

**Second instance (lines 340-347):**

```go
if !cfg.IncludeSQLC {
    filterOptions = append(filterOptions, filter.FilterSQLC)
    // ...
}
```

**Impact:**

- Redundant filter option added
- Confusing logic flow
- Potential performance impact (duplicate checks)

**Recommendation:**

```go
// Combined logic
if !cfg.IncludeSQLC {
    filterOptions = append(filterOptions, filter.FilterSQLC)
    if cfg.Verbose && len(sqlcOutputDirs) > 0 {
        fmt.Fprintf(os.Stderr, "🔍 Auto-detected sqlc.yaml, filtering sqlc generated code\n")
        for _, dir := range sqlcOutputDirs {
            fmt.Fprintf(os.Stderr, "   - %s\n", dir)
        }
    }
}
```

### 3.2 Duplicate Logic: Detection Methods to String Conversion

**Location:** `cmd/run.go:179-187` and `473-479`

**Problem:** Same logic for converting detection methods to comma-separated string appears twice.

**Recommendation:**
Extract to shared function:

```go
func detectionMethodsToString(methods config.DetectionMethods) string {
    if len(methods) == 0 {
        return ""
    }
    result := make([]string, len(methods))
    for i, dm := range methods {
        result[i] = dm.String()
    }
    return strings.Join(result, ",")
}
```

### 3.3 Duplicate Logic: Match Collection

**Location:** `pkg/artdupl/detector.go:252-275` and `297-324`

**Problem:** `runDetection` and `streamDetectionResults` both collect matches into groups using identical logic.

**Recommendation:**
Extract to shared function:

```go
func collectMatchesIntoGroups(matchesChan <-chan syntax.Match) map[string][][]*syntax.Node {
    groups := make(map[string][][]*syntax.Node)
    for match := range matchesChan {
        if len(match.Frags) > 0 {
            groups[match.Hash] = append(groups[match.Hash], match.Frags...)
        }
    }
    return groups
}
```

### 3.4 Duplicate Logic: Filter Check Pattern

**Location:** `cmd/run.go:249-252, 273-275, 287-289`

**Problem:** Same filter check pattern appears in multiple places.

**Recommendation:**
Extract to shared function:

```go
func shouldIncludeFile(filter *filter.Filter, path string) bool {
    return filter == nil || !filter.ShouldFilter(path)
}
```

### 3.5 Duplicate Data: SQLC Pattern Lists

**Location:** `pkg/filter/filter.go:303-313` and `390-401`

**Problem:** Same SQLC file pattern list defined in two places.

**Recommendation:**
Extract to package-level constant:

```go
const sqlcFilePatterns = []string{
    "models.go",
    "querier.go",
    "query.sql.go",
    "batch.go",
}
```

### 3.6 Split Brain: Stats Data

**Location:** `printer/stats_data.go` vs `printer/stats.go`

**Problem:**

- `StatsData` in `printer/stats_data.go` has its own type system using primitives
- Separated from domain types in `domain/domain_types.go`
- Creates two separate type systems for similar concepts

**Impact:**

- Type inconsistency across the codebase
- Loss of type safety in statistics
- No unified validation strategy

**Recommendation:**

1. Create domain types for statistics metrics
2. Migrate `StatsData` to use domain types
3. Consolidate type systems

---

## 4. Boolean-to-Enum Conversion Opportunities

### 4.1 Analysis Mode

**Current:** Boolean flags in `config.Config`:

```go
type Config struct {
    Verbose           bool
    IncludeVendor     bool
    FilesFromStdin    bool
    FilterGenerated   bool
    IncludeSQLC       bool
    IncludeTempl      bool
    Profile           bool
}
```

**Observation:** These are mostly fine as boolean flags. However, consider:

**Potential Enhancement:** For `FilterGenerated`, which combines multiple filters:

```go
type FilterMode string

const (
    FilterModeNone     FilterMode = "none"
    FilterModeAuto     FilterMode = "auto"
    FilterModeCustom   FilterMode = "custom"
    FilterModeStrict   FilterMode = "strict"
)
```

**Rationale:** Provides clearer semantics about filtering strategy rather than simple on/off.

### 4.2 Output Format

**Current:** Already uses enums correctly:

```go
type OutputFormat string

const (
    OutputFormatText       OutputFormat = "text"
    OutputFormatHTML        OutputFormat = "html"
    OutputFormatJSON       OutputFormat = "json"
    OutputFormatSimpleJSON OutputFormat = "simplejson"
    OutputFormatPlumbing   OutputFormat = "plumbing"
)
```

**Assessment:** ✅ Well-implemented. No changes needed.

---

## 5. Error Handling Patterns

### 5.1 Strengths

**Typed Errors:** `errors/types.go` provides excellent typed error handling:

```go
type ErrorType string

const (
    ValidationError ErrorType = "validation"
    ConfigError    ErrorType = "config"
    AnalysisError  ErrorType = "analysis"
    DetectionError ErrorType = "detection"
)
```

**Wrapper Functions:** Comprehensive error wrapping:

```go
func Wrap(err error, errType ErrorType, context string) DuplError
func WrapConfig(err error, context string) DuplError
func WrapValidation(err error, context string) DuplError
```

**Assessment:** ✅ Excellent implementation. No changes needed.

### 5.2 Areas for Improvement

#### Issue: Unused Parameter in Marshal Function

**Location:** `domain/domain_types.go:41`

**Problem:**

```go
func marshalStringID(s, typeName, validationMsg string) ([]byte, error) {
    if s == "" {
        return nil, errors.NewValidationError(validationMsg, nil)
    }
    return json.Marshal(s)  // typeName is never used
}
```

**Recommendation:**
Either use `typeName` for better error messages or remove the parameter:

```go
func marshalStringID(s, validationMsg string) ([]byte, error) {
    if s == "" {
        return nil, errors.NewValidationError(validationMsg, nil)
    }
    return json.Marshal(s)
}
```

---

## 6. uint Type Usage and Optimizations

### 6.1 Excellent Optimizations

**LineNumber (uint16):**

```go
type LineNumber uint16
// Optimized: uint16 provides 0-65,535 range (sufficient for any source file)
```

✅ Correct - No source file has >65,535 lines

**BytePosition (uint32):**

```go
type BytePosition uint32
// Optimized: uint32 provides 0-4GB range (sufficient for file positions)
```

✅ Correct - No source file exceeds 4GB

**ComplexityScore (uint16):**

```go
type ComplexityScore uint16
// Optimized: uint16 provides 0-65,535 range (sufficient for code complexity)
```

✅ Correct - No realistic code complexity exceeds 65,535

### 6.2 Areas for Improvement

#### Issue: TokenCount Uses uint

**Location:** `domain/domain_types.go:258-261`

**Current:**

```go
type TokenCount uint
```

**Consideration:** uint is platform-dependent (32-bit or 64-bit). For most cases this is fine, but consider:

**Recommendation:**
Current implementation is acceptable. `uint` provides sufficient range for token counts. No change needed.

---

## 7. Code Quality Observations

### 7.1 Unused Code

#### Unused Function: isGeneratedByFilename

**Location:** `pkg/filter/filter.go:329-332`

**Status:** Never called, marked as "legacy, kept for compatibility"

**Recommendation:** Either remove entirely or document intended use case clearly.

#### Unused Parameter: ctx in runSuffixTreeDetection

**Location:** `pkg/artdupl/detector.go:330`

**Problem:** Parameter exists but never used for cancellation checking.

**Recommendation:** Either implement proper cancellation or remove parameter.

### 7.2 Inefficiencies

#### Issue: Channel Conversion in runAllModes

**Location:** `cmd/run.go:465-508`

**Problem:**

```go
// Line 465: Convert channel to slice
matches := collectMatches(duplChan)

// Lines 501-508: Convert slice back to channel
matchChan := make(chan syntax.Match)
go func() {
    defer close(matchChan)
    for _, match := range matches {
        matchChan <- match
    }
}()
```

**Impact:** Defeats the purpose of streaming, adds memory overhead.

**Recommendation:**
Option 1: Keep streaming to each output file sequentially

```go
for _, format := range formats {
    // Create new printer
    p := createPrinter(format, cfg.Threshold)(file, os.ReadFile)
    // Reuse duplChan for streaming
    if err := printDupls(p, duplChan, sortByEnum, cfg.Threshold, detectionMethodStr); err != nil {
        // Handle error
    }
}
```

Option 2: Store results once and write multiple times without reconversion

```go
// Store in printer-ready format
cloneGroups := printer.BuildCloneGroups(duplChan)
for _, format := range formats {
    // Write cloneGroups directly without channel conversion
}
```

#### Issue: Manual Map Copying

**Location:** `pkg/filter/filter.go:97-100`

**Problem:**

```go
filteredByReason := make(map[FilterReason]int)
for k, v := range m.FilteredByReason {
    filteredByReason[k] = v
}
```

**Recommendation:** Use `maps.Copy` (Go 1.21+):

```go
filteredByReason := maps.Clone(m.FilteredByReason)
```

### 7.3 Algorithm Improvements

#### Issue: Complexity Calculation is Too Basic

**Location:** `domain/clone.go:477-494`

**Problem:**

```go
func calculateComplexity(node *syntax.Node) uint {
    complexity := uint(1)
    for _, child := range node.Children {
        complexity += calculateComplexity(child)
    }
    switch node.Type {
    case 0:  // Magic number without documentation
        complexity += 2
    default:
        complexity += 1
    }
    return complexity
}
```

**Issues:**

- Recursive implementation could be expensive for deep trees
- Magic number `0` without documentation
- Very basic metric, not aligned with industry standards

**Recommendation:**
Consider established complexity metrics:

- **McCabe Cyclomatic Complexity:** Counts decision points
- **Nesting Depth:** Maximum nesting level
- **Cognitive Complexity:** Accounts for nesting and logic flow

Example:

```go
func calculateCognitiveComplexity(node *syntax.Node) uint {
    // Use established algorithm from cognitive complexity research
    // Account for:
    // - Binary decisions (+1)
    // - Switch cases (+1 for each case)
    // - Logical operators (&&, ||) (+1 for each)
    // - Nesting depth multiplier
}
```

#### Issue: Hash-Based Detection Not Implemented

**Location:** `pkg/artdupl/detector.go:354-357`

**Problem:**

```go
func (d *detector) runHashDetection(ctx context.Context, data []*syntax.Node, threshold int) <-chan syntax.Match {
    return d.runSuffixTreeDetection(ctx, data, threshold)  // INCORRECT
}
```

**Impact:** Hash detection option doesn't work as expected.

**Recommendation:**
Implement actual rolling hash detection:

```go
func (d *detector) runHashDetection(ctx context.Context, data []*syntax.Node, threshold int) <-chan syntax.Match {
    resultChan := make(chan syntax.Match)
    go func() {
        defer close(resultChan)

        // Use rolling hash algorithm from pkg/hash
        hasher := hash.NewRollingHash(threshold)
        for _, node := range data {
            if hash := hasher.Process(node); hash != nil {
                resultChan <- hash
            }
        }
    }()
    return resultChan
}
```

---

## 8. Single Responsibility Principle Violations

### 8.1 printer/stats.go

**Concerns:** 727 lines handling too many responsibilities:

- Data collection
- Health calculation
- Formatting (text, JSON, CSV)
- Visualization (ASCII bars, tables)
- Recommendations
- Style management

**Impact:** Difficult to test, maintain, and understand.

**Recommendation:** See Section 1.1 for suggested file split.

### 8.2 cmd/run.go

**Concerns:** 528 lines mixing:

- Flag parsing
- Config management
- File crawling
- Analysis execution
- Output formatting

**Impact:** Violates separation of concerns, hard to test individual components.

**Recommendation:** See Section 1.2 for suggested file split.

---

## 9. Testing and Validation

### 9.1 Observations

**Test Coverage:** BDD tests exist in `bdd/` directory using Ginkgo/Gomega.

**Test Structure:** Good behavior-driven development approach.

**Areas for Improvement:**

1. Add unit tests for helper functions in large files
2. Test conversion functions (NodeToClone, convertFragmentToClone)
3. Test validation functions more thoroughly
4. Add performance regression tests for algorithms

---

## 10. Security Considerations

### 10.1 Observations

**File Reading:**

```go
// cmd/run.go:285
content, err := os.ReadFile(filePath) //nolint:gosec //G304 filePath is from controlled source
```

✅ Properly reviewed and annotated

**Path Validation:** File crawling validates paths before processing.

**Output Directory Creation:**

```go
// cmd/run.go:453
if err := os.MkdirAll(outputDir, 0o750); err != nil {
    return fmt.Errorf("failed to create output directory %q: %w", outputDir, err)
}
```

✅ Proper permissions (0o750 = rwxr-x---)

**Assessment:** Security practices are well-implemented.

---

## 11. Performance Considerations

### 11.1 Strengths

**Memory Efficiency:**

- StringPool for string interning ✅
- Optimized uint sizes (uint16, uint32) ✅
- Streaming APIs (FindClonesStream) ✅

**Concurrency:**

- Goroutine-based pipeline processing ✅
- Thread-safe Metrics with RWMutex ✅
- Proper channel usage ✅

### 11.2 Areas for Improvement

**Algorithm Complexity:**

- Complexity calculation needs improvement (Section 7.3)
- Hash detection not implemented (Section 7.3)

**Memory Usage:**

- Channel conversion inefficiency (Section 7.2)
- Consider memory profiling for large codebases

---

## 12. Recommendations Summary

### Immediate Actions (High Priority)

1. **Split files exceeding 350 lines:**
   - printer/stats.go (727 lines) → 6 focused files
   - pkg/artdupl/detector.go (546 lines) → 5 focused files
   - cmd/run.go (528 lines) → 5 focused files

2. **Fix hash-based detection:**
   - Implement actual rolling hash algorithm
   - Remove incorrect delegation to suffix tree

3. **Remove duplicate SQLC filtering:**
   - Consolidate into single location
   - Remove redundant filter option addition

4. **Fix type safety violations:**
   - Change SDK Clone to use domain types
   - Create domain.HealthGrade enum
   - Migrate StatsData to domain types

### Short-Term Actions (Medium Priority)

5. **Extract shared logic:**
   - Detection methods to string conversion
   - Match collection into groups
   - Filter check pattern

6. **Remove unused code:**
   - Delete or document isGeneratedByFilename
   - Use or remove unused ctx parameter

7. **Improve algorithms:**
   - Implement better complexity calculation
   - Use maps.Copy for map copying

8. **Address deprecated methods:**
   - Audit Uint() usage
   - Plan migration to Uint16()/Uint32()
   - Remove after migration

### Long-Term Actions (Low Priority)

9. **Refactor remaining large files:**
   - domain/domain_types.go (525 lines) → 5 focused files
   - domain/clone.go (495 lines) → 7 focused files
   - pkg/filter/filter.go (461 lines) → 5 focused files

10. **Consolidate type systems:**
    - Unify StatsData with domain types
    - Remove split brain between printer and domain

11. **Enhance testing:**
    - Add unit tests for helper functions
    - Test conversion functions
    - Add performance regression tests

---

## 13. Conclusion

The art-dupl codebase demonstrates **strong engineering fundamentals** with excellent domain-driven design, comprehensive error handling, and good performance optimizations. The main areas for improvement are:

1. **File organization** - Several files violate the 350-line limit and need splitting
2. **Type safety** - Some conversion functions lose type safety by using primitives
3. **Code duplication** - Several instances of duplicate logic across files
4. **Feature completeness** - Hash-based detection is not implemented

**Overall Assessment:** B+ (Good with clear improvement path)

**Key Strengths:**

- Excellent domain type system
- Strong error handling with typed errors
- Memory-efficient implementation
- Good use of enums throughout

**Key Weaknesses:**

- Large files violating SRP
- Type safety gaps in conversion layer
- Duplicate logic across files
- Incomplete feature implementation (hash detection)

**Recommended Approach:**

1. Address immediate issues first (file splitting, hash detection)
2. Improve type safety systematically
3. Eliminate duplications through extraction
4. Continue iterative refactoring for long-term health

---

## Appendix A: File Size Distribution

```
< 100 lines:     8 files
100-199 lines:   12 files
200-299 lines:   6 files
300-399 lines:   3 files
400-499 lines:   2 files
500-599 lines:   2 files
600-699 lines:   0 files
700-799 lines:   1 file (stats.go)
800+ lines:      0 files
```

## Appendix B: Type System Coverage

**Domain Types (Excellent Coverage):**

- ✅ IDs: CloneGroupID, AnalysisID
- ✅ File: Filepath, LineNumber, BytePosition
- ✅ Metrics: TokenCount, FileCount, CloneCount, Threshold
- ✅ Metadata: Hash, ComplexityScore, Confidence, ProcessingTime
- ✅ Enums: FileProcessingState, DetectionState, AnalysisMode, CloneSeverity

**Missing Domain Types:**

- ❌ HealthGrade (currently string in StatsData)
- ⚠️ Clone severity thresholds (currently hardcoded in CalculateSeverity)

**Recommendation:** Complete type system coverage by adding missing domain types.
