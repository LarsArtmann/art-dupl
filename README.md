# art-dupl

**art-dupl** is a Go tool for finding code clones using suffix tree algorithms on serialized ASTs. It identifies structural duplicates while ignoring literal values.

## Installation

```bash
go install github.com/LarsArtmann/art-dupl@latest
```

Or build from source:

```bash
git clone https://github.com/LarsArtmann/art-dupl.git && cd art-dupl && make build
```

## Quick Start

```bash
# Basic usage
./art-dupl

# Higher threshold (larger clones only)
./art-dupl -t 100

# HTML report
./art-dupl -html > report.html

# JSON output (new in this fork)
./art-dupl -json -t 20

# Check version
./art-dupl --version

# Enable shell completions (bash)
source <(./art-dupl completion bash)
# Enable shell completions (zsh)
source <(./art-dupl completion zsh)

# Generate man page
./art-dupl man > art-dupl.1
```

## Key Features

- **Structural clone detection** using suffix tree algorithms
- **JSON output** for CI/CD automation
- **Configuration files** for team consistency
- **Multiple output formats**: text, HTML, JSON, plumbing
- **Statistics subcommand** (`art-dupl stats`) for project overview
- **Professional CLI** with auto-completion and version info
- **Enhanced help** with styling and examples

## Configuration

Create `dupl.json`:

```json
{
  "threshold": 30,
  "outputFormat": "json",
  "paths": ["./src", "./lib"]
}
```

Use with:

```bash
./dupl -config dupl.json
```

## CLI Flags

```
-config string           Configuration file path
-files                   Read file names from stdin
-html                    HTML output with code fragments
-json                    JSON output (new)
-plumbing                Machine-readable output
-t, -threshold           Minimum token size (default 15)
-vendor                  Include vendor directory
-v, -verbose             Verbose logging
-filter-generated        Smart filtering of generated code
-include-sqlc            Include sqlc.dev generated files
-include-templ           Include templ.guide generated files
-include-pattern value   File patterns to always include
-exclude-pattern value   File patterns to exclude
-profile                 Enable performance profiling
-timeout duration        Maximum execution time (default 30m)
-detection-methods       Detection methods: hash, art-dupl (default: art-dupl)
```

### Subcommands

**stats** - Show aggregated duplication statistics

```bash
art-dupl stats [flags] [paths...]
```

Supports all root command flags plus:

```
-t, -threshold          Minimum token size (default 15)
-m, -detection-methods  Detection methods to use
--format                Output format: text, json (default: text)
```

**Format validation**: The `--format` flag only accepts "text" or "json". Providing an invalid format will result in an error.

## Examples

### CI/CD Integration

```bash
# Fail build if too many duplicates
TOTAL_CLONES=$(dupl -json . | jq '.summary.total_clones')
if [ "$TOTAL_CLONES" -gt 100 ]; then
  echo "Too many code duplicates: $TOTAL_CLONES"
  exit 1
fi
```

### Analysis

```bash
# Find test file duplicates
find . -name '*_test.go' | dupl -files

# Analyze specific paths
./dupl ./src ./lib -t 50
```

### Statistics

The `stats` subcommand provides aggregated duplication statistics for quick project overviews:

```bash
# Show statistics for current directory
./art-dupl stats

# Show statistics for specific paths
./art-dupl stats ./src ./lib

# Show statistics in JSON format (for post-processing)
./art-dupl stats --format json ./src | jq '.overview.totalClones'

# Show statistics with custom threshold
./art-dupl stats -t 50 ./src

# Compare projects using stats output
./art-dupl stats project1/ > stats1.txt
./art-dupl stats project2/ > stats2.txt
```

**Statistics include:**

- Files scanned and clone groups found
- Total duplicate lines and tokens
- Average clone size and complexity score
- Impact score (tokens × instances)
- Clone size distribution (1-5, 6-10, 11-20, 21-50, 51-100, 100+ lines)
- Top files with most duplication

Example output:

```
Code Duplication Statistics
============================

Configuration:
  Threshold: 15 tokens
  Detection Methods: art-dupl

Overview:
  Files Scanned: 114
  Clone Groups: 156
  Total Clones: 423

Duplicate Code:
  Total Duplicate Lines: 1247
  Total Duplicate Tokens: 893
  Average Clone Size: 3 lines
  Complexity Score: 2.71
  Impact Score: 18234

Clone Size Distribution:
  1-5 lines: 89 clones
  6-10 lines: 45 clones
  11-20 lines: 23 clones
  21-50 lines: 12 clones
  100+ lines: 2 clones

Top Files by Duplicate Lines:
  234 lines in src/handlers/user.go
  189 lines in src/models/data.go
  ...
```

## Output Formats

- **Text**: Simple clone listing with file paths and line numbers
- **HTML**: Detailed report with syntax-highlighted code fragments
- **JSON**: Structured data with metadata and summary statistics
- **Plumbing**: Machine-readable format for scripts

## Architecture

- **suffixtree/** - Core suffix tree implementation
- **syntax/** - AST handling and serialization
- **job/** - File parsing orchestration
- **printer/** - Output formatting

## Testing

```bash
make test  # Run all tests
make check # Run linting
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT

## Architecture Overview

art-dupl is organized into focused packages following clean architecture principles:

### Core Packages

- **domain/** - Domain model and value objects
- **syntax/** - Unified AST representation
- **suffixtree/** - Suffix tree data structure
- **detection/** - Multi-method detection coordination
- **config/** - Configuration and validation

### Supporting Packages

- **errors/** - Rich error types with context
- **printer/** - Output formatting and statistics
- **types/** - Functional programming primitives
- **cmd/** - CLI application
- **pkg/artdupl/** - SDK for programmatic use

### Type Safety Approach

Three-layer type safety:
1. Domain Types (strong safety)
2. Helper Functions (safe access)
3. Backward Compatible API

See go.mod for detailed module documentation.

## License

MIT
## Architecture Overview

art-dupl is organized into focused packages following clean architecture principles:

### Core Packages

- **domain/** - Domain model and value objects
  - Value objects: `LineNumber`, `Threshold`, `TokenCount`, `BytePosition`
  - Entities: `Clone`, `CloneGroup`, `Analysis`
  - Enums: `CloneSeverity`, `DetectionState`, `AnalysisMode`
  - All types are immutable and validated at construction
  - String interning for memory efficiency (`StringInternPool`)

- **syntax/** - Unified AST representation
  - Language-agnostic `Node` type representing any language construct
  - Transformations for Go AST (`syntax/golang/`)
  - Functions for finding complete syntax units
  - Hash computation for duplicate detection

- **suffixtree/** - Suffix tree data structure
  - Efficient duplicate search using compressed trie
  - SIMD-optimized transition search for >8 transitions
  - `O(n)` construction and search for typical code
  - Memory-optimized for large codebases

- **detection/** - Multi-method detection coordination
  - `MultiDetector` coordinates multiple detection algorithms
  - Methods: syntax-level, hash-based, TODO comments, legacy patterns
  - Combines and deduplicates results
  - Verbose logging for debugging

- **config/** - Configuration and validation
  - Type-safe enums: `DetectionMethod`, `OutputFormat`, `SortCriteria`
  - `Config` struct with typed access helpers
  - JSON/YAML loading with validation
  - Configuration merging (file + CLI flags)

### Supporting Packages

- **errors/** - Rich error types with context
  - `DuplError` with type, message, file, line, cause, stack
  - Error wrapping: `WrapConfig()`, `WrapValidation()`, `WrapAnalysis()`
  - Typed marshaling: `SafeMarshalConfig()`, `SafeMarshalClone()`, etc.
  - Consistent error handling across codebase

- **printer/** - Output formatting and statistics
  - Multiple formats: Text, JSON, HTML, Plumbing
  - `StatsData` with count, size, complexity, quality metrics
  - Sorting by size, occurrence, hash, tokens
  - Filtering and threshold handling

- **types/** - Functional programming primitives
  - `Result[T]` for type-safe error handling
  - `Option[T]` for optional values
  - Helper functions: `Ok()`, `Err()`, `Some()`, `None()`

### CLI and SDK

- **cmd/** - CLI application
  - `run.go` - Main execution logic (513 lines)
  - Flag parsing and validation
  - File crawling and filtering
  - Output formatting

- **pkg/artdupl/** - SDK for programmatic use
  - `Detector` interface: `FindClones()`, `FindClonesStream()`
  - `Options` for configuration
  - `Result` with summary and metadata
  - Streaming support for large projects

### Type Safety Approach

art-dupl uses a layered approach to type safety:

1. **Domain Types** (Strong Safety)
   ```go
   // Enforced at construction time
   threshold, err := domain.NewThreshold(15)
   if err != nil { ... }

   // Compile-time guarantees
   line := domain.LineNumber(10) // Can't accidentally use wrong value
   ```

2. **Helper Functions** (Safe Access)
   ```go
   // Typed access without breaking changes
   cfg := config.DefaultConfig()
   domainThreshold := cfg.GetThresholdAsDomain()

   // Typed marshaling
   data, err := errors.SafeMarshalClone(&clone, "marshaling")
   ```

3. **Backward Compatible** (Incremental Migration)
   ```go
   // Old API still works
   match := syntax.FindSyntaxUnits(data, match, 15)

   // New type-safe API available
   domainThreshold := domain.NewThreshold(15)
   match := syntax.FindSyntaxUnitsWithDomainThreshold(data, match, domainThreshold)
   ```

### Performance Optimizations

- **Memory Layout**: Node struct is 40B (37.5% reduction from 64B)
- **SIMD**: Vectorized transition search for >8 transitions
- **String Interning**: Duplicate strings use same memory (StringInternPool)
- **Streaming**: Large projects use channels for non-blocking results
- **Thresholds**: `maxChildrenSerial = 10,000` prevents goroutine stack overflow

### Data Flow

```
Source Files
    ↓
Parsing (go/parser)
    ↓
Syntax Transform (syntax/golang/)
    ↓
Unified AST (syntax.Node[])
    ↓
Suffix Tree Build (suffixtree.STree)
    ↓
Duplicate Search (FindDuplOver())
    ↓
Syntax Unit Matching (FindSyntaxUnits())
    ↓
Clone Groups (domain.CloneGroup[])
    ↓
Output Formatting (printer/*)
    ↓
Text/HTML/JSON/Plumbing
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT
