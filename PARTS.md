# PARTS.md - Extractable Components Analysis

> Analysis of components that could be extracted as standalone reusable libraries/SDKs.
>
> **Last Updated:** March 3, 2026

## Executive Summary

art-dupl contains several high-quality, well-tested components that could be extracted as standalone libraries. This analysis evaluates each component's extraction potential, compares with existing alternatives, and provides recommendations.

**Top Extraction Candidates:**

| Priority | Component | Extraction Value | Recommendation |
|----------|-----------|------------------|----------------|
| 1 | Suffix Tree | High | Extract as generic library |
| 2 | String Intern Pool | Medium | Extract if API refined |
| 3 | Smart File Filter | Medium | Extract with broader tool support |
| 4 | Clone Detection SDK | High | Already designed, needs refinement |
| 5 | Domain Types | Low | Keep internal, patterns valuable |

---

## 1. Suffix Tree Library (`suffixtree/`)

### Description

A generic suffix tree implementation optimized for finding duplicate sequences.

**Key Features:**
- O(1) map-based transition lookup (optimized from O(n) linear search)
- Stream processing via `Update()` for incremental builds
- `FindDuplOver()` for threshold-based duplicate detection
- Clean `Token` interface for any tokenizable data
- Memory-efficient with 32-bit positions

### Current State

```go
// Core interface
type Token interface {
    Val() int
}

type STree struct {
    data     []Token
    root     *state
    // ...
}

// Key operations
func New() *STree
func (t *STree) Update(data ...Token)
func (t *STree) FindDuplOver(threshold int) <-chan Match
```

### Alternatives

| Library | Stars | Notes |
|---------|-------|-------|
| None found | - | No dedicated Go suffix tree library exists |
| Custom implementations | - | Typically inline, not reusable |

### Unique Value Proposition

1. **Generic Token Interface**: Works with any tokenizable data, not just strings
2. **O(1) Transitions**: Map-based lookup vs. linear search in naive implementations
3. **Streaming API**: Process large datasets without loading everything into memory
4. **Well-Tested**: Comprehensive tests including memory benchmarks

### Extraction Recommendation

**Extract as `github.com/LarsArtmann/go-suffixtree`**

**Proposed API:**

```go
package suffixtree

// Token represents any tokenizable element
type Token interface {
    Val() int
}

// Match represents a found duplicate sequence
type Match struct {
    Ps  []Pos  // Positions of duplicates
    Len int    // Length of duplicate sequence
}

// STree is a suffix tree for finding repeated sequences
type STree struct { /* ... */ }

// New creates an empty suffix tree
func New() *STree

// Update adds tokens to the tree (streaming support)
func (t *STree) Update(tokens ...Token)

// FindRepeated finds all repeated sequences >= threshold
func (t *STree) FindRepeated(threshold int) <-chan Match

// At returns the token at position p
func (t *STree) At(p Pos) Token
```

**Effort:** Low (API already clean, minimal dependencies)

---

## 2. String Intern Pool (`domain/stringpool.go`)

### Description

Thread-safe string interning pool for memory-efficient string storage.

**Key Features:**
- 32-bit StringID (4B vs 16B for string header)
- Thread-safe with RWMutex
- JSON marshaling/unmarshaling support
- Global pool singleton with test injection

### Current State

```go
type StringID uint32

type StringInternPool struct {
    mu      sync.RWMutex
    strings []string
    index   map[string]StringID
    nextID  StringID
}

func (p *StringInternPool) Intern(s string) StringID
func (p *StringInternPool) Lookup(id StringID) string
func GlobalPool() *StringInternPool
```

### Alternatives

| Library | Stars | Notes |
|---------|-------|-------|
| `intern` (various) | ~50-100 | Simple interning, no ID system |
| `go.stringinterner` | ~30 | Basic interning, no JSON support |

### Unique Value Proposition

1. **Compact IDs**: 32-bit IDs vs string headers (4B vs 16B per reference)
2. **JSON Serialization**: Automatic string serialization via MarshalJSON
3. **Thread-Safe Singleton**: Global pool with test injection
4. **Stats API**: Monitor pool usage

### Extraction Recommendation

**Consider extracting as `github.com/LarsArtmann/go-stringpool`**

**Improvements needed:**
- Remove JSON dependency (make it optional)
- Add `Clear()` for pool reset
- Consider generational pooling for long-running processes

**Effort:** Medium (API refinement needed)

---

## 3. Smart File Filter (`pkg/filter/`)

### Description

Intelligent filtering of auto-generated code files.

**Key Features:**
- Detects SQLC-generated files via `sqlc.yaml` detection
- Detects Templ-generated files via header comments
- Detects GoEnum-generated files
- Custom include/exclude pattern support
- Metrics tracking for filter statistics

### Current State

```go
type Filter struct {
    options         map[FilterOption]bool
    enabled         bool
    includePatterns []string
    excludePatterns []string
    metrics         *Metrics
}

func (f *Filter) ShouldFilter(filePath string) bool
func (f *Filter) GetStats() FilterStats
```

### Alternatives

| Library | Stars | Notes |
|---------|-------|-------|
| `.gitignore` parsers | ~100+ | Pattern-based only |
| `golang.org/x/tools/go/packages` | N/A | Has `IgnoreFile` but limited |

### Unique Value Proposition

1. **Tool-Specific Detection**: Knows about SQLC, Templ, GoEnum patterns
2. **Smart Detection**: Looks for config files (`sqlc.yaml`) not just file names
3. **Metrics**: Track what was filtered and why
4. **Extensible**: Easy to add new tool patterns

### Extraction Recommendation

**Extract as `github.com/LarsArtmann/go-genfilter`**

**Improvements needed:**
- Add more tool patterns (protobuf, wire, mockgen, etc.)
- Make detection logic pluggable
- Add file content inspection for better accuracy

**Effort:** Medium (needs broader tool support)

---

## 4. Clone Detection SDK (`pkg/artdupl/`)

### Description

High-level SDK for programmatic code clone detection.

**Key Features:**
- Clean `Detector` interface
- Streaming and batch APIs
- Progress reporting
- Multiple detection methods
- Context support for cancellation

### Current State

Already designed in `SDK_DESIGN.md`:

```go
type Detector interface {
    FindClones(ctx context.Context, files []string) (*Result, error)
    FindClonesStream(ctx context.Context, files []string) (<-chan *CloneGroup, error)
}

type Options struct {
    Threshold        int
    DetectionMethods []DetectionMethod
    IncludeVendor    bool
    ProgressCallback func(*Progress) error
    FileReader       FileReaderFunc
}
```

### Alternatives

| Library | Stars | Notes |
|---------|-------|-------|
| `mibk/dupl` | 357 | CLI tool only, no SDK |
| PMD CPD | 4k+ | Java-based, Java-centric |
| SonarQube | 8k+ | Platform, not library |

### Unique Value Proposition

1. **Library-First**: Designed for programmatic use, not just CLI
2. **Multiple Methods**: Suffix tree + hash-based detection
3. **Streaming**: Handle large codebases without memory issues
4. **Go-Native**: Built for Go code, understands Go AST

### Extraction Recommendation

**Already the right approach - refine `pkg/artdupl/`**

**Improvements needed:**
- Complete SDK implementation per `SDK_DESIGN.md`
- Add plugin system for custom detection methods
- Add caching/incremental analysis support
- Better error handling with typed errors

**Effort:** Low (design exists, implementation needed)

---

## 5. Domain Types (`domain/`)

### Description

Strongly-typed domain model with validation.

**Key Features:**
- Value objects: `LineNumber`, `BytePosition`, `TokenCount`, `Threshold`
- Entities: `Clone`, `CloneGroup`, `Analysis`
- Enums: `FileProcessingState`, `DetectionState`, `CloneSeverity`
- Validation at construction time

### Current State

```go
type LineNumber int
type BytePosition int64
type TokenCount int
type Threshold int

func NewLineNumber(n int) (LineNumber, error)
func (ln LineNumber) Validate() error

type Clone struct {
    FileID    StringID
    StartLine LineNumber
    EndLine   LineNumber
    Size      TokenCount
}
```

### Alternatives

| Library | Stars | Notes |
|---------|-------|-------|
| `go-types` (various) | ~50-100 | Generic type wrappers |
| Custom per-project | - | Most projects use primitives |

### Unique Value Proposition

1. **Domain-Specific**: Types match the problem domain exactly
2. **Compile-Time Safety**: Make impossible states unrepresentable
3. **Self-Validating**: Types validate at construction

### Extraction Recommendation

**Do NOT extract - keep as internal patterns**

Domain types are highly specific to this project. The patterns are valuable but should be documented, not extracted.

**Alternative:** Create a blog post or example repository showing the pattern.

---

## 6. Syntax/AST Package (`syntax/`)

### Description

Unified AST representation for code analysis.

**Key Features:**
- Memory-optimized Node struct (40B vs 64B)
- Language-agnostic design
- Serialization for suffix tree processing
- `FindSyntaxUnits()` for complete syntax unit extraction

### Current State

```go
type Node struct {
    Type     int32
    Pos      int32
    End      int32
    Owns     int32
    Children []*Node
    Filename string
}

func Serialize(n *Node) []*Node
func FindSyntaxUnits(data []*Node, m Match, threshold int) Match
```

### Alternatives

| Library | Stars | Notes |
|---------|-------|-------|
| `go/ast` | Stdlib | Go-specific |
| `tree-sitter` | 18k+ | Multi-language but heavy |
| `sitter` (Go bindings) | ~500 | Tree-sitter Go bindings |

### Unique Value Proposition

1. **Lightweight**: No external dependencies, minimal memory
2. **Suffix Tree Ready**: Designed to work with suffix tree algorithm
3. **Go-Native**: Optimized for Go AST analysis

### Extraction Recommendation

**Do NOT extract independently**

This is tightly coupled to the suffix tree algorithm. Extract only as part of a larger code analysis SDK.

---

## 7. Printer System (`printer/`, `adapter/`)

### Description

Multi-format output system for clone reports.

**Key Features:**
- Text, HTML, JSON, Plumbing, Stats formats
- Sortable output (size, occurrence, hash)
- Adapter pattern for extensibility
- Stats visualization with health indicators

### Current State

```go
type Printer interface {
    PrintHeader() error
    PrintClones(dups [][]*syntax.Node, sortBy ...SortBy) error
    PrintFooter() error
}

type StatsPrinter interface {
    Printer
    SetFilesCount(count int)
    SetDetectionMethods(methods string)
    // ...
}
```

### Alternatives

| Library | Stars | Notes |
|---------|-------|-------|
| `tabwriter` | Stdlib | Basic table output |
| `termui` | 7k+ | Terminal UI, overkill |
| `printer` (various) | ~50 | Format-specific |

### Unique Value Proposition

1. **Multi-Format**: One interface, multiple output formats
2. **Clone-Specific**: Understands clone data structure
3. **Sortable**: Built-in sorting by multiple criteria

### Extraction Recommendation

**Do NOT extract - too domain-specific**

The printer system is tightly coupled to clone representation. The adapter pattern is the reusable part.

---

## 8. Error Types (`errors/`)

### Description

Typed error handling with context.

**Key Features:**
- Categorized errors (validation, config, analysis)
- Error wrapping with context
- JSON marshalable errors
- Internal vs user-facing errors

### Current State

```go
type Error struct {
    Category string
    Code     string
    Message  string
    Context  map[string]any
    Cause    error
}

func NewValidationError(msg string, cause error) *Error
func NewConfigError(msg string, cause error) *Error
func NewAnalysisError(msg string, cause error) *Error
```

### Alternatives

| Library | Stars | Notes |
|---------|-------|-------|
| `cockroachdb/errors` | 1.5k+ | Rich error handling |
| `pkg/errors` | 8k+ | Deprecated but popular |
| `larsartmann/uniflow` | - | Railway-oriented errors |

### Unique Value Proposition

1. **Categorized**: Structured error categories for handling
2. **JSON-Ready**: Serializable for API responses
3. **Context-Rich**: Additional context fields

### Extraction Recommendation

**Do NOT extract - use `cockroachdb/errors` or `uniflow`**

The project should adopt `cockroachdb/errors` or `larsartmann/uniflow` per HOW_TO_GOLANG.md guidelines instead of custom error types.

---

## Extraction Priority Matrix

```
                    High Value
                        │
    ┌───────────────────┼───────────────────┐
    │                   │                   │
    │  Suffix Tree ●    │  Clone SDK ●      │
    │                   │                   │
    │                   │                   │
Low ┼───────────────────┼───────────────────┤ High
    │                   │                   │ Effort
    │  String Pool ●    │  Smart Filter ●   │
    │                   │                   │
    │                   │                   │
    │  Domain Types ○   │  Printer ○        │
    │                   │                   │
    └───────────────────┼───────────────────┘
                        │
                    Low Value
```

**Legend:** ● = Extract, ○ = Keep Internal

---

## Recommended Extraction Roadmap

### Phase 1: High-Value, Low-Effort

1. **Suffix Tree Library**
   - Extract `suffixtree/` as standalone package
   - Clean API, minimal dependencies
   - High value for other projects needing sequence analysis

2. **Clone Detection SDK**
   - Complete `pkg/artdupl/` implementation per SDK_DESIGN.md
   - Add comprehensive documentation
   - This IS the product for library consumers

### Phase 2: Medium-Effort Additions

3. **String Intern Pool**
   - Extract if broader use cases identified
   - Consider merging with existing interning libraries

4. **Smart File Filter**
   - Extract with expanded tool support
   - Add plugin system for custom detectors

### Phase 3: Documentation

5. **Domain Types Pattern**
   - Document the pattern in examples
   - Create template for other projects

6. **Printer Adapter Pattern**
   - Document the adapter pattern usage
   - Show how to add new formats

---

## Integration with HOW_TO_GOLANG.md

Per the library policy, extracted libraries should:

1. **Use idiomatic Go**: `(T, error)` returns, not Result types
2. **Use `samber/do/v2`**: For dependency injection if needed
3. **Use `log/slog` + `charmbracelet/log`**: For logging
4. **Use `koanf`**: For configuration if needed
5. **Use `cockroachdb/errors`**: For error handling
6. **Follow composition over inheritance**: No deep hierarchies

---

## Conclusion

**Extract Now:**
- Suffix Tree Library (high value, clean API)
- Complete Clone Detection SDK (core product)

**Consider Later:**
- String Intern Pool (needs API refinement)
- Smart File Filter (needs broader tool support)

**Keep Internal:**
- Domain Types (patterns, not library)
- Syntax Package (coupled to algorithm)
- Printer System (domain-specific)
- Error Types (use standard libraries)
