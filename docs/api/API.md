# art-dupl API Documentation

**Generated:** 2026-01-07  
**Package:** `github.com/LarsArtmann/art-dupl`

---

## Overview

art-dupl is a Go tool for finding code clones using suffix tree algorithms and hash-based detection. It analyzes abstract syntax trees (ASTs) to find structural code clones while ignoring literal values.

---

## Package Index

### 📦 cli
**Location:** `./cli`  
**Purpose:** Command-line interface configuration and helpers

**Key Types:**
- `CLIConfig`: CLI configuration structure
- `OutputFormat`: Supported output formats (Text, HTML, JSON, Plumbing)

**Key Functions:**
- `NewCLIConfig()`: Creates default CLI configuration
- `Run()`: Main CLI entry point

---

### 📦 config
**Location:** `./config`  
**Purpose:** Configuration management with validation

**Key Types:**
- `Config`: Main configuration structure
  - `Threshold`: Minimum token sequence size (default: 15)
  - `IncludeVendor`: Include vendor directory (default: false)
  - `FilesFromStdin`: Read paths from stdin (default: false)
  - `OutputFormat`: Output format (Text, HTML, JSON, Plumbing)
  - `Verbose`: Enable verbose logging (default: false)
  - `Paths`: Paths to analyze (default: ["."])
  - `IgnoreFiles`: File patterns to ignore (default: [])
  - `MaxChildrenSerial`: Max children serial for large slices (default: 10000)
  - `OutputFile`: Output file path (default: "")
  - `SortBy`: Sort criteria (size, occurrence, hash, total-tokens)
  - `DetectionMethods`: Detection methods (art-dupl, hash, todos, legacy)
  - `Profile`: Enable performance profiling (default: false)
  - `Timeout`: Execution timeout in seconds (default: 0 = no timeout)

- `OutputFormat`: Supported output formats with type safety
  - `OutputFormatText`: Text output (default)
  - `OutputFormatHTML`: HTML output
  - `OutputFormatJSON`: JSON output
  - `OutputFormatPlumbing`: Machine-readable plumbing output

- `SortCriteria`: Sort criteria options
  - `SortBySize`: Sort by token count (largest first)
  - `SortByOccurrence`: Sort by file count (most widespread first)
  - `SortByHash`: Sort by hash value (alphabetical)
  - `SortByTotalTokens`: Sort by total tokens across all files

- `DetectionMethod`: Detection method options
  - `DetectionMethodArtDupl`: Original suffix tree detection
  - `DetectionMethodHash`: SHA1 hash-based detection
  - `DetectionMethodTodos`: TODO detection
  - `DetectionMethodLegacy`: Legacy detection

**Key Functions:**
- `DefaultConfig()`: Returns default configuration
- `LoadConfig(path string)`: Loads configuration from JSON file
- `MergeConfigs(file, cli *Config)`: Merges file and CLI configurations
- `ValidateConfig(cfg *Config)`: Validates configuration constraints
- `AllOutputFormats()`: Returns list of all output formats
- `AllSortCriteria()`: Returns list of all sort criteria
- `AllDetectionMethods()`: Returns list of all detection methods

---

### 📦 detection
**Location:** `./detection`  
**Purpose:** Multi-detector coordination and clone detection

**Key Types:**
- `MultiDetector`: Multi-detection orchestrator
- `DetectionMethod`: Detection method implementation interface

**Key Functions:**
- `NewMultiDetector(cfg *config.Config, data *[]*syntax.Node, t *suffixtree.STree, verbose bool)`: Creates multi-detector
- `FindDuplOver(threshold int)`: Finds duplicates over threshold
- `FindAllDupl()`: Finds all duplicates

---

### 📦 hash
**Location:** `./hash`  
**Purpose:** SHA1 hash-based detection implementation

**Key Functions:**
- `Detect(data *[]*syntax.Node, threshold int)`: Hash-based duplicate detection
- `ComputeHash(nodes []*syntax.Node)`: Computes SHA1 hash for nodes

---

### 📦 job
**Location:** `./job`  
**Purpose:** Job orchestration and performance profiling

**Key Types:**
- `ProfileResult`: Performance profiling metrics
  - `AllocMB`: Memory allocated in MB
  - `TotalAllocMB`: Total memory allocated in MB
  - `SysMB`: System memory in MB
  - `NumGC`: Number of garbage collections
  - `PauseTotalMS`: Total GC pause time in ms
  - `Duration`: Total execution time
  - `NumGoroutine`: Number of goroutines

**Key Functions:**
- `BuildTree(schan chan []*syntax.Node)`: Builds suffix tree from node channel
- `Profile()`: Captures performance metrics
- `StartProfile()`: Starts profiling session
- `EndProfile(start)`: Completes profiling and calculates duration
- `ProfileDiff(start, end)`: Calculates difference between profiles
- `PrintProfileResult(result)`: Outputs formatted profiling metrics

---

### 📦 printer
**Location:** `./printer`  
**Purpose:** Output formatting for all supported formats

**Key Types:**
- `Printer`: Output formatter interface
- `TextPrinter`: Human-readable text output
- `HTMLPrinter`: HTML report with syntax highlighting
- `JSONPrinter`: Structured JSON with metadata
- `PlumbingPrinter`: Machine-readable format for scripts
- `CloneGroup`: Clone group with hash and fragments

**Key Functions:**
- `CreatePrinter(format)`: Creates printer for format
- `PrintClones(printer, groups, threshold)`: Outputs clone groups
- `SortCloneGroups(groups, sortBy)`: Sorts groups by criteria
- `SortClonesBySize(dups)`: Sorts by token count
- `SortClonesByOccurrence(dups)`: Sorts by file count
- `SortClonesByHash(dups)`: Sorts by hash
- `SortClonesByTotalTokens(dups)`: Sorts by total tokens

---

### 📦 suffixtree
**Location:** `./suffixtree`  
**Purpose:** Core suffix tree implementation

**Key Types:**
- `STree`: Suffix tree structure
- `Node`: Suffix tree node

**Key Functions:**
- `New()`: Creates new suffix tree
- `Update(node)`: Adds node to suffix tree
- `FindDupl(threshold)`: Finds duplicates using suffix tree

---

### 📦 syntax
**Location:** `./syntax`  
**Purpose:** AST handling, serialization, and node processing

**Key Types:**
- `Node`: Abstract syntax tree node
- `Match`: Duplicate match with hash and fragments
- `Fragment`: Code fragment with positions

**Key Functions:**
- `Parse(files, verbose)`: Parses Go source files to ASTs
- `Serialize(node)`: Serializes AST node to string
- `GetUnitsIndexes(nodeSeq)`: Identifies complete syntax units
- `IsCyclic(nodeSeq)`: Checks for cyclic dependencies
- `GetUnits(nodeSeq, threshold)`: Gets complete syntax units

**Language Support:**
- `syntax/golang`: Go language implementation

---

### 📦 types
**Location:** `./types`  
**Purpose:** Common type definitions and utilities

**Key Types:**
- `StringEnum`: Generic string-based enum
- `ValidatableEnum`: Interface for enum validation

---

### 📦 errors
**Location:** `./errors`  
**Purpose:** Rich error handling with context and stack traces

**Key Types:**
- `ErrorType`: Error type categories
  - `ParseError`: Parsing errors
  - `ConfigError`: Configuration errors
  - `IOError`: Input/output errors
  - `ValidationError`: Validation errors
  - `InternalError`: Internal errors

- `DuplError`: Rich error structure
  - `Type`: Error type
  - `Message`: Error message
  - `File`: File where error occurred
  - `Line`: Line number where error occurred
  - `Cause`: Wrapped error
  - `Stack`: Stack trace

**Key Functions:**
- `NewParseError(file, line, msg, cause)`: Creates parse error
- `NewConfigError(msg, cause)`: Creates config error
- `NewIOError(file, msg, cause)`: Creates I/O error
- `NewValidationError(msg, cause)`: Creates validation error
- `NewInternalError(msg, cause)`: Creates internal error
- `Is(err, errorType)`: Checks if error matches type

---

### 📦 util
**Location:** `./util`  
**Purpose:** Utility functions

**Key Functions:**
- `Unique(strings)`: Returns unique strings
- `Contains(slice, item)`: Checks if item in slice

---

### 📦 testutils
**Location:** `./testutils`  
**Purpose:** Test helpers and utilities

**Key Functions:**
- `GenerateRandomSuffix()`: Generates random suffix for tests
- `UniqueTestHelper`: Unique test helper utilities

---

## Usage Examples

### Basic Analysis

```go
package main

import (
    "github.com/LarsArtmann/art-dupl/config"
    "github.com/LarsArtmann/art-dupl/detection"
    "github.com/LarsArtmann/art-dupl/job"
    "github.com/LarsArtmann/art-dupl/syntax"
    "github.com/LarsArtmann/art-dupl/suffixtree"
)

func main() {
    // Default configuration
    cfg := config.DefaultConfig()
    cfg.Threshold = 20
    cfg.Paths = []string{"./src"}
    
    // Parse files
    t, data, filesCount, err := buildSuffixTree(cfg.Paths, cfg.Verbose, cfg.FilesFromStdin)
    if err != nil {
        panic(err)
    }
    
    // Detect duplicates
    multiDetector := detection.NewMultiDetector(cfg, data, t, cfg.Verbose)
    matches := multiDetector.FindDuplOver(cfg.Threshold)
    
    // Output results
    for match := range matches {
        println(match.Hash, len(match.Frags))
    }
}
```

### Performance Profiling

```go
// Enable profiling
cfg := config.DefaultConfig()
cfg.Profile = true

// Analysis runs with profiling
start := job.StartProfile()
// ... analysis ...
end := job.EndProfile(start)
job.PrintProfileResult(end)
```

### Configuration Loading

```go
// Load from file
fileConfig, err := config.LoadConfig("dupl.json")
if err != nil {
    panic(err)
}

// Merge with CLI config
cliConfig := &config.Config{
    Threshold: 25,
}
merged := config.MergeConfigs(fileConfig, cliConfig)

// Validate
if err := config.ValidateConfig(merged); err != nil {
    panic(err)
}
```

---

## Error Handling

All errors are wrapped with rich context:

```go
import "github.com/LarsArtmann/art-dupl/errors"

// Create errors with context
err := errors.NewParseError("file.go", 10, "failed to parse", ioErr)

// Check error type
if errors.Is(err, errors.ParseError) {
    // Handle parse error
}

// Access wrapped error
if err != nil {
    fmt.Printf("Cause: %v\n", errors.Unwrap(err))
}
```

---

## Integration Points

### As a Library

```go
import (
    "github.com/LarsArtmann/art-dupl/syntax"
    "github.com/LarsArtmann/art-dupl/suffixtree"
    "github.com/LarsArtmann/art-dupl/detection"
)

// 1. Parse files to AST
nodeSeq, err := syntax.Parse(files, verbose)

// 2. Build suffix tree
stree := suffixtree.New()
for _, node := range nodeSeq {
    stree.Update(node)
}

// 3. Detect duplicates
detector := detection.NewMultiDetector(cfg, &nodeSeq, stree, verbose)
matches := detector.FindDuplOver(threshold)
```

### As a CLI Tool

```bash
# Basic usage
art-dupl ./src

# Custom threshold
art-dupl -t 50 ./src

# Multiple output formats
art-dupl --html --json --plumbing ./src

# Sorting options
art-dupl --sort occurrence ./src
art-dupl --sort hash ./src
art-dupl --sort total-tokens ./src

# Advanced features
art-dupl --profile --timeout 30m ./src
art-dupl --detection-methods "hash,art-dupl" ./src
```

---

## Testing

### Run Tests

```bash
# All tests
go test ./...

# Specific package
go test ./config

# Verbose output
go test -v ./syntax

# Coverage
go test -cover ./...
```

---

## Contributing

### Adding New Detection Methods

1. Implement detection interface in `detection` package
2. Add detection method constant to `config`
3. Register in `AllDetectionMethods()`
4. Add tests in `detection/` package

### Adding New Output Formats

1. Implement `Printer` interface in `printer` package
2. Add output format constant to `config`
3. Register in `AllOutputFormats()`
4. Add tests in `printer/` package

### Adding New Languages

1. Create language package in `syntax/` directory
2. Implement `Parse()` function
3. Implement `Serialize()` function
4. Add language detection logic
5. Add tests for language

---

## Performance

### Profiling

Enable performance profiling to analyze:

- Memory usage (allocation, system memory)
- Garbage collection (cycles, pause time)
- Execution time
- Concurrency (goroutines)

**Usage:**
```bash
art-dupl --profile ./src
```

**Example Output:**
```
═════════════════════════════════════════════════════════
                    PERFORMANCE PROFILING RESULTS
═════════════════════════════════════════════════════════

Execution Time:
    2m 5s

Memory Usage:
  Current Allocation:    12.45 MB
  Total Allocated:       145.67 MB
  System Memory:        78.23 MB

Garbage Collection:
  GC Cycles:                23
  Total Pause Time:       45.23 ms
  Avg Pause/Cycle:        1.97 ms

Concurrency:
  Goroutines:                4

═════════════════════════════════════════════════════════
```

---

## License

See LICENSE file for details.

---

**Last Updated:** 2026-01-07  
**Version:** 1.0.0  
**Generated by:** AI Assistant via Crush
