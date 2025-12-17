# dupl SDK Design Document

## Current State Analysis

### ✅ Strengths
- Core algorithms (suffixtree, syntax, detection) are well-decoupled from CLI
- Clean interfaces exist: `Printer`, `Token`, `Config`
- MIT license allows flexible usage
- Configuration system is robust with validation
- File processing pipeline is modular and reusable

### ❌ Limitations  
- `lib.Run()` is too simplistic - only returns `[]printer.Issue`
- No unified high-level SDK interface
- Advanced features (multi-detection, hash detection) not exposed in lib
- No error handling customization
- No progress reporting or cancellation support
- No way to access raw matches or intermediate results

## Proposed SDK Design

### 1. Primary SDK Interface

```go
package artdupl

// Detector is the main interface for code duplication detection
type Detector interface {
    // FindClones performs duplication analysis
    FindClones(ctx context.Context, files []string) (*Result, error)
    
    // FindClonesStream provides streaming results for large projects
    FindClonesStream(ctx context.Context, files []string) (<-chan *CloneGroup, error)
}

// Result contains all detected duplicates with metadata
type Result struct {
    CloneGroups  []*CloneGroup `json:"clone_groups"`
    Summary      *Summary      `json:"summary"`
    Metadata     *Metadata     `json:"metadata"`
}

// CloneGroup represents a group of identical code fragments
type CloneGroup struct {
    Hash       string    `json:"hash"`
    Clones     []*Clone  `json:"clones"`
    Size       int       `json:"size"`
    LineCount  int       `json:"line_count"`
    Method     DetectionMethod `json:"detection_method"`
}

// Clone represents a single occurrence of duplicated code
type Clone struct {
    Filename   string `json:"filename"`
    StartLine  int    `json:"start_line"`
    EndLine    int    `json:"end_line"`
    Fragment   string `json:"fragment,omitempty"`
    Size       int    `json:"size"`
}

// Summary provides statistics about the analysis
type Summary struct {
    TotalFiles    int `json:"total_files"`
    TotalClones    int `json:"total_clones"`
    TotalGroups    int `json:"total_groups"`
    AnalysisTime   time.Duration `json:"analysis_time_ms"`
}
```

### 2. Configuration Options

```go
// Configures the detector behavior
type Options struct {
    // Detection settings
    Threshold         int               `json:"threshold"`
    DetectionMethods  []DetectionMethod `json:"detection_methods"`
    
    // File processing
    IncludeVendor    bool              `json:"include_vendor"`
    IgnoreFiles      []string          `json:"ignore_files"`
    MaxFileSize      int64             `json:"max_file_size"`
    
    // Performance
    MaxWorkers       int               `json:"max_workers"`
    Timeout          time.Duration     `json:"timeout"`
    
    // Output customization
    IncludeFragments bool              `json:"include_fragments"`
    
    // Callbacks for progress
    ProgressCallback func(progress *Progress) error
    
    // Custom file reader (for testing/virtual files)
    FileReader      FileReaderFunc
}

// Progress reports analysis progress
type Progress struct {
    Stage       string  `json:"stage"`
    Completed   int     `json:"completed"`
    Total       int     `json:"total"`
    Percentage  float64 `json:"percentage"`
    Message     string  `json:"message"`
}
```

### 3. Implementation Strategy

#### Phase 1: Core SDK Interface
- Create `pkg/artdupl/` package with clean API
- Implement Detector interface using existing components
- Expose all detection methods (art-dupl, hash)
- Add proper error handling and context support

#### Phase 2: Advanced Features  
- Streaming API for large projects
- Progress reporting and cancellation
- Custom file readers (in-memory, virtual files)
- Configurable output formats

#### Phase 3: Integration Features
- Plugin system for custom detection methods
- Export/import functionality
- Caching and incremental analysis
- Language extensibility

## Usage Examples

### Basic Usage
```go
import "github.com/LarsArtmann/art-dupl/pkg/artdupl"

detector := artdupl.NewDetector(&artdupl.Options{
    Threshold: 15,
    DetectionMethods: []artdupl.DetectionMethod{artdupl.MethodArtDupl},
})

result, err := detector.FindClones(context.Background(), []string{"./src"})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Found %d clone groups\n", len(result.CloneGroups))
```

### Advanced Usage with Streaming
```go
detector := artdupl.NewDetector(&artdupl.Options{
    Threshold: 20,
    IncludeFragments: true,
    ProgressCallback: func(p *artdupl.Progress) error {
        fmt.Printf("Progress: %.1f%% - %s\n", p.Percentage, p.Message)
        return nil
    },
})

cloneChan, err := detector.FindClonesStream(ctx, []string{"./src"})
if err != nil {
    log.Fatal(err)
}

for group := range cloneChan {
    fmt.Printf("Found clone group: %s with %d clones\n", group.Hash, len(group.Clones))
}
```

### Integration with CI/CD
```go
detector := artdupl.NewDetector(&artdupl.Options{
    Threshold: 30,
    DetectionMethods: []artdupl.DetectionMethod{artdupl.MethodHash},
})

result, err := detector.FindClones(context.Background(), []string{"./src"})
if err != nil {
    return err
}

// Fail build if too many duplicates
if result.Summary.TotalClones > 100 {
    return fmt.Errorf("too many code duplicates: %d", result.Summary.TotalClones)
}

// Export JSON for reporting
data, _ := json.Marshal(result)
os.WriteFile("duplicates.json", data, 0644)
```

## Migration Path

1. **Create SDK package** alongside existing CLI code
2. **Gradually migrate lib.Run** to use new SDK implementation
3. **Maintain backward compatibility** during transition
4. **Mark old lib as deprecated** with migration guide
5. **Document best practices** for different use cases

This design provides a clean, powerful API while leveraging the excellent existing architecture.