# Migration Guide: From Primitive Types to Domain Types

This guide helps you migrate from primitive types (int, string, float64) to domain types (domain.LineNumber, domain.Threshold, etc.) in the art-dupl codebase.

## Why Migrate?

### Benefits of Domain Types

1. **Type Safety** - Compile-time checks prevent accidental type mismatches
2. **Self-Documenting Code** - Domain types make intent explicit
3. **Validation at Construction** - Invalid states impossible by design
4. **Better IDE Support** - Autocomplete shows only valid types
5. **Easier Refactoring** - Find all usages of specific domain type

### Example: Type Safety

```go
// ❌ WITHOUT DOMAIN TYPES: Can accidentally use wrong value
func findClone(line int, size int) Clone {
    // line might be -1 (invalid)
    // size might be 0 (invalid)
    // No compile-time checking!
    return Clone{Line: line, Size: size}
}

// ✅ WITH DOMAIN TYPES: Compile-time guarantees
func findClone(line domain.LineNumber, size domain.TokenCount) Clone {
    // line is always valid (construction enforced)
    // size is always valid (construction enforced)
    // Compile-time type checking!
    return Clone{Line: line, Size: size}
}
```

## Migration Strategy

### Principle: Incremental Migration

Don't migrate everything at once! Use a layered approach:

1. **Layer 1**: Use domain types at package boundaries (APIs)
2. **Layer 2**: Use domain types in new code
3. **Layer 3**: Migrate critical paths (detection, config)
4. **Layer 4**: Migrate remaining code incrementally

### Principle: Backward Compatibility

Keep old APIs working while adding new type-safe APIs:

```go
// Old API (still works)
match := syntax.FindSyntaxUnits(data, match, 15)

// New type-safe API (available)
threshold := domain.NewThreshold(15)
match := syntax.FindSyntaxUnitsWithDomainThreshold(data, match, threshold)
```

## Common Migration Patterns

### Pattern 1: Function Parameters

#### Before
```go
func analyzeFile(threshold int, line int, tokens int) error {
    // No validation
    if threshold < 1 { return errors.New("invalid threshold") }
    if line < 1 { return errors.New("invalid line") }
    // ...
}
```

#### After
```go
func analyzeFile(threshold domain.Threshold, line domain.LineNumber, tokens domain.TokenCount) error {
    // Validation enforced at construction
    // No runtime checks needed!
    // ...
}
```

### Pattern 2: Struct Fields

#### Before
```go
type Clone struct {
    Line     int    `json:"line"`
    Size      int    `json:"size"`
    Hash      string `json:"hash"`
    File      string `json:"file"`
}
```

#### After
```go
type Clone struct {
    Line     domain.LineNumber `json:"line"`
    Size      domain.TokenCount `json:"size"`
    Hash      domain.HashString    `json:"hash"`
    File      domain.Filepath   `json:"file"`
}
```

### Pattern 3: Map Keys/Values

#### Before
```go
fileDuplication := make(map[string]int)
fileDuplication["file.go"] = 100
```

#### After
```go
fileDuplication := make(map[domain.Filepath]domain.TokenCount)
fileDuplication[domain.Filepath("/file.go")] = domain.TokenCount(100)
```

### Pattern 4: Slices and Arrays

#### Before
```go
thresholds := []int{10, 15, 20}
```

#### After
```go
thresholds := []domain.Threshold{
    domain.Threshold(10),
    domain.Threshold(15),
    domain.Threshold(20),
}
```

### Pattern 5: JSON Marshaling

#### Before
```go
data, err := json.Marshal(config)
if err != nil { ... }
```

#### After (type-safe)
```go
data, err := errors.SafeMarshalConfig(&config)
if err != nil { ... }
```

## Type Mapping Table

| Primitive Type | Domain Type | Constructor | Validation Rules |
|---------------|--------------|-------------|------------------|
| `int` (lines) | `domain.LineNumber` | `domain.NewLineNumber(value)` | > 0 |
| `int` (tokens) | `domain.TokenCount` | `domain.NewTokenCount(value)` | > 0 |
| `int` (threshold) | `domain.Threshold` | `domain.NewThreshold(value)` | > 0 and <= 1000 |
| `int` (bytes) | `domain.BytePosition` | `domain.NewBytePosition(value)` | >= 0 |
| `uint` (generic) | `domain.Uint` | `domain.NewUint(value)` | >= 0 |
| `string` (file) | `domain.Filepath` | `domain.NewFilepath(value)` | Non-empty, valid path |
| `string` (fragment) | `domain.FragmentString` | `domain.NewFragmentString(value)` | Non-empty |
| `string` (hash) | `domain.HashString` | `domain.NewHashString(value)` | Non-empty |
| `string` (group ID) | `domain.CloneGroupID` | `domain.NewCloneGroupID(value)` | Non-empty |
| `string` (analysis ID) | `domain.AnalysisID` | `domain.NewAnalysisID(value)` | Non-empty |
| `string` (string ID) | `domain.StringID` | `domain.NewStringID(value)` | Non-empty |
| `float64` (confidence) | `domain.Confidence` | `domain.NewConfidence(value)` | 0.0 to 1.0 |
| `float64` (complexity) | `domain.ComplexityScore` | `domain.NewComplexityScore(value)` | >= 0.0 |

## Package-Specific Migration Guides

### 1. config Package

**Before**:
```go
type Config struct {
    Threshold int `json:"threshold"`
    Timeout   int `json:"timeout"`
}
```

**After** (incremental):
```go
type Config struct {
    Threshold int `json:"threshold"` // Keep int for JSON compatibility
    Timeout   int `json:"timeout"`  // Keep int for JSON compatibility
}

// Add typed access helpers
func (c *Config) GetThresholdAsDomain() domain.Threshold {
    threshold, _ := domain.NewThreshold(uint(c.Threshold))
    return threshold
}

func (c *Config) SetThresholdFromDomain(t domain.Threshold) error {
    // Validation already enforced
    c.Threshold = int(t.Uint())
    return nil
}
```

**Usage**:
```go
cfg := config.DefaultConfig()
domainThreshold := cfg.GetThresholdAsDomain()
```

### 2. syntax Package

**Before**:
```go
func FindSyntaxUnits(data []*Node, m Match, threshold int) Match {
    // ...
}
```

**After** (backward compatible):
```go
// Old API (still works)
func FindSyntaxUnits(data []*Node, m Match, threshold int) Match

// New type-safe API
func FindSyntaxUnitsWithDomainThreshold(data []*Node, m Match, threshold domain.Threshold) Match {
    return FindSyntaxUnits(data, m, int(threshold.Uint()))
}
```

### 3. printer Package

**Before**:
```go
type StatsData struct {
    TotalFilesScanned   int
    TotalDuplicateLines int
    ComplexityScore     float64
    // ...
}
```

**After** (incremental):
```go
import "github.com/LarsArtmann/art-dupl/domain"

type StatsData struct {
    // Keep primitives for JSON compatibility
    // Domain types available for future migration
    TotalFilesScanned   int
    TotalDuplicateLines int
    ComplexityScore     float64
    // ...
}
```

### 4. errors Package

**Before**:
```go
func SafeMarshal(v any, context string) ([]byte, error) {
    data, err := json.Marshal(v)
    if err != nil { return nil, HandleMarshalingError(...) }
    return data, nil
}
```

**After** (add typed helpers):
```go
// Old generic API (still works)
func SafeMarshal(v any, context string) ([]byte, error)

// New type-safe APIs
func SafeMarshalConfig(cfg *config.Config) ([]byte, error)
func SafeMarshalClone(c *domain.Clone) ([]byte, error)
func SafeMarshalCloneGroup(g *domain.CloneGroup) ([]byte, error)
func SafeMarshalAnalysis(a *domain.Analysis) ([]byte, error)
```

## Testing Your Migration

### 1. Compile-Time Checks

After migration, code should compile without errors related to type mismatches:

```bash
go build ./...
```

### 2. Unit Tests

Update or add tests for domain type validation:

```go
func TestNewThreshold(t *testing.T) {
    // Valid values
    _, err := domain.NewThreshold(15)
    assert.NoError(t, err)

    // Invalid values
    _, err = domain.NewThreshold(0)
    assert.Error(t, err)

    _, err = domain.NewThreshold(1001)
    assert.Error(t, err)
}
```

### 3. Integration Tests

Run full integration tests to ensure behavior is unchanged:

```bash
go test -run TestIntegration ./...
```

## Common Pitfalls to Avoid

### 1. Don't Mix Primitive and Domain Types

```go
// ❌ WRONG: Mixing types
type Clone struct {
    Line      domain.LineNumber // Domain type
    Size      int               // Primitive type - inconsistent!
}

// ✅ CORRECT: All domain types
type Clone struct {
    Line      domain.LineNumber
    Size      domain.TokenCount
}
```

### 2. Don't Bypass Validation

```go
// ❌ WRONG: Bypassing validation
threshold := domain.Threshold(-1)  // Constructor catches this!

// ✅ CORRECT: Let constructor validate
threshold, err := domain.NewThreshold(-1)
if err != nil { /* handle error */ }
```

### 3. Don't Use Type Assertions

```go
// ❌ WRONG: Type assertion (unsafe!)
threshold := someInt.(domain.Threshold)

// ✅ CORRECT: Constructor (safe!)
threshold := domain.Threshold(uint(someInt))
```

### 4. Don't Forget Uint() Conversion

```go
// ❌ WRONG: Direct comparison
if domainThreshold < someInt { ... }  // Won't compile!

// ✅ CORRECT: Convert to uint first
if domainThreshold.Uint() < uint(someInt) { ... }
```

## Rollback Strategy

If migration causes issues, you can roll back:

### Option 1: Revert Commits
```bash
git log --oneline  # Find migration commits
git revert <commit-hash>    # Revert specific commits
```

### Option 2: Keep Old API
Don't remove old APIs - keep both versions:
```go
// Old API (safe, known to work)
func FindSyntaxUnits(..., threshold int) Match

// New API (type-safe)
func FindSyntaxUnitsWithDomainThreshold(..., threshold domain.Threshold) Match
```

### Option 3: Gradual Migration
Start with non-critical paths:
1. Migrate examples first
2. Migrate tests next
3. Migrate CLI flags last

## FAQ

### Q: Should I migrate all code at once?
**A**: No! Migrate incrementally. Start with new code, then migrate critical paths.

### Q: What about existing code that uses primitive types?
**A**: Keep it working! Don't break existing APIs. Add new type-safe APIs alongside old ones.

### Q: Can I use domain types in my own code?
**A**: Yes! Import the domain package and use types:
```go
import "github.com/LarsArtmann/art-dupl/domain"

func myFunction() {
    line := domain.LineNumber(10)
    // ...
}
```

### Q: What if I need to convert domain type back to primitive?
**A**: Use the Uint(), String() methods:
```go
threshold := domain.Threshold(15)
asInt := int(threshold.Uint())
asUint := threshold.Uint()
```

### Q: Can I add new domain types?
**A**: Yes! Follow the pattern:
1. Define type in domain/domain_types.go
2. Add constructor: NewTypeName(value) (TypeName, error)
3. Add validation logic in constructor
4. Add Uint(), String() methods
5. Add json.Marshaler/json.Unmarshaler if needed

## Resources

- [Domain Types Documentation](domain/domain.go) - Package-level documentation
- [Domain Type Reference](domain/domain_types.go) - All domain types and methods
- [Usage Examples](examples/domain_types_usage.go) - Complete code examples
- [Module Documentation](go.mod) - Module-level documentation and design decisions

## Getting Help

If you encounter issues during migration:

1. Check this guide's patterns
2. See examples/domain_types_usage.go for working code
3. Check domain package documentation
4. Ask questions in GitHub discussions

Happy migrating! 🚀
