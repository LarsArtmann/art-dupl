# Feedback and Issues from auto-deduplicate

**Source Project**: auto-deduplicate
**Target Project**: art-dupl
**Date**: 2026-02-15
**Purpose**: Document gaps in art-dupl that auto-deduplicate needed to implement locally

---

## Executive Summary

auto-deduplicate is being refactored to delegate all duplicate detection to art-dupl, keeping only AI integration (Crush service) and UI/UX (diff visualization, reports) features. This document captures what auto-deduplicate implemented that could/should be provided by art-dupl.

---

## Gaps Identified

### 1. Progress Callback Interface for Long-Running Operations

**Status**: NEEDS ENHANCEMENT

**What auto-deduplicate needs**:

```go
type ProgressReporter interface {
    Update(current, total int, message string) error
    Complete(message string) error
    Reset(total int) error
}
```

**Current art-dupl provides**:

```go
// pkg/artdupl/types.go
ProgressCallback func(*Progress) error
```

**Gap**: The `Progress` struct in art-dupl has the fields but:

- No standardized way to integrate with external progress systems (e.g., progress bars, UI)
- No completion signal separate from updates
- No reset capability for multi-phase operations

**Recommendation**: Consider adding:

```go
type ProgressCallback interface {
    OnProgress(p *Progress) error
    OnComplete(summary *Summary) error
    OnError(err error) error
}
```

### 2. Content-Based Cache Invalidation

**Status**: NEEDS FEATURE

**What auto-deduplicate implemented**:

- `ProjectFingerprinter` - Computes content hash of all project files
- Cache keys that include fingerprint to auto-invalidate on code changes
- Staleness validation for cached results

**Current art-dupl provides**:

- `lib.RunIncremental()` with AST caching per-file
- No project-level fingerprinting
- No automatic cache invalidation based on content changes

**Gap**: The incremental cache in art-dupl is file-based, but for cross-project consistency, a content-based fingerprint would be valuable.

**Recommendation**: Add project-level fingerprinting to `pkg/artdupl`:

```go
type Fingerprinter interface {
    Compute(ctx context.Context, root string) (string, error)
    ComputeShort(ctx context.Context, root string) (string, error) // Fast partial fingerprint
}
```

### 3. Result Caching with TTL

**Status**: NEEDS FEATURE

**What auto-deduplicate needs**:

- Cache full detection results with configurable TTL
- Staleness checks (file existence, modification time validation)
- Cache statistics (hits, misses)

**Current art-dupl provides**:

- Per-file AST caching only
- No result-level caching
- No TTL support

**Recommendation**: Consider adding result caching to the SDK:

```go
type CacheOptions struct {
    Enabled    bool
    TTL        time.Duration
    MaxSize    int64
    CacheDir   string
}

type CachedResult struct {
    Result      *Result
    CachedAt    time.Time
    Fingerprint string
}
```

### 4. Non-Go File Support

**Status**: POTENTIAL ENHANCEMENT

**What auto-deduplicate implemented**:

- Generic file hash detection (any file type, not just Go)
- Configurable file filters (include/exclude patterns)
- .gitignore integration

**Current art-dupl provides**:

- Go AST-based detection only
- Limited to .go files

**Recommendation**: This may be intentionally scoped to Go. If so, document clearly. If non-Go support is desired, consider a plugin architecture for language handlers.

### 5. Streaming Results with Backpressure

**Status**: EXISTS BUT LIMITED

**What auto-deduplicate needs**:

- Streaming results for large projects
- Backpressure handling
- Cancellation propagation

**Current art-dupl provides**:

```go
FindClonesStream(ctx context.Context, files []string) (<-chan *CloneGroup, error)
```

**Gap**: The streaming exists but:

- Channel buffer size is hardcoded (10)
- No backpressure signaling
- No partial results on error

**Recommendation**: Add configurable streaming options:

```go
type StreamOptions struct {
    BufferSize    int
    OnBackpressure func() // Callback when consumer is slow
    YieldPartial  bool    // Return partial results on error
}
```

### 6. Detection Method Configuration via Options

**Status**: EXISTS - GOOD

**What auto-deduplicate uses**:

```go
DetectionMethods []DetectionMethod
```

**Current art-dupl provides**: Full support via `config.DetectionMethods`

**Status**: This is well-implemented. No action needed.

---

## Minor Issues

### 1. Error Types Not Exported

The `errors` package in art-dupl has typed errors, but they're not easily accessible from `pkg/artdupl`. Auto-deduplicate had to create its own error handling.

**Recommendation**: Export error types from `pkg/artdupl/errors.go`:

```go
var (
    ErrNoClonesFound = errors.New("no clones found")
    ErrAnalysisTimeout = errors.New("analysis timeout")
    // etc.
)
```

### 2. File Filtering Not Configurable at SDK Level

Auto-deduplicate needs to filter files by:

- .gitignore patterns
- Size limits
- Include/exclude glob patterns
- Generated file detection

**Recommendation**: Add `FileFilter` interface to SDK:

```go
type FileFilter interface {
    ShouldProcess(path string, info os.FileInfo) bool
}
```

---

## What auto-deduplicate Will DELETE

These components are redundant with art-dupl:

1. **`duplicate_service.go`** - Thin wrapper around `lib.Run()` with caching
   - Cache logic can be moved to art-dupl if desired
   - File collection is straightforward

2. **`file_hash_service.go`** - Binary file hash detection
   - art-dupl's `hash.FileDetector` provides this
   - Migration: Replace with `hash.NewFileDetector()`

3. **`file_hash_worker.go`** - Parallel worker implementation
   - art-dupl handles parallelization internally

---

## Integration Notes

### Preferred Integration Pattern

```go
import (
    "github.com/LarsArtmann/art-dupl/pkg/artdupl"
    "github.com/LarsArtmann/art-dupl/lib"
)

// For code duplicate detection
detector, _ := artdupl.NewDetector(&artdupl.Options{
    Threshold:        15,
    DetectionMethods: []artdupl.DetectionMethod{artdupl.MethodArtDupl},
    IncludeFragments: true,
})
result, _ := detector.FindClones(ctx, files)

// For incremental detection with caching
issues, stats, _ := lib.RunIncremental(ctx, files, threshold, cacheDir, false)
```

### Type Mapping

| auto-deduplicate           | art-dupl                                 |
| -------------------------- | ---------------------------------------- |
| `interfaces.Duplicate`     | `printer.Issue` or `artdupl.CloneGroup`  |
| `interfaces.FileDuplicate` | `artdupl.CloneGroup` (with `MethodHash`) |
| `interfaces.FileHash`      | `artdupl.Clone`                          |

---

## Action Items for art-dupl

1. [ ] Consider adding progress callback interface for external integration
2. [ ] Add project-level fingerprinting for cache invalidation
3. [ ] Document non-Go file support decision (intentional limitation?)
4. [ ] Export error types from SDK
5. [ ] Add configurable file filtering

---

## Questions

1. Should art-dupl support non-Go files, or is that intentionally out of scope?
2. Is there interest in adding result-level caching to the SDK?
3. Would progress callback interface improvements be welcome?

---

_Generated by auto-deduplicate refactoring effort_
