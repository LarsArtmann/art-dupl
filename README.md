# art-dupl

[![Go Report Card](https://goreportcard.com/badge/github.com/LarsArtmann/art-dupl)](https://goreportcard.com/report/github.com/LarsArtmann/art-dupl) [![codecov](https://codecov.io/gh/LarsArtmann/art-dupl/graph/badge.svg?token=art-dupl)](https://codecov.io/gh/LarsArtmann/art-dupl)

**art-dupl** is a Go tool for finding code clones using suffix tree algorithms on serialized ASTs. It identifies structural duplicates while ignoring literal values.

## Installation

```bash
go install github.com/LarsArtmann/art-dupl@latest
```

Or build from source:

```bash
git clone https://github.com/LarsArtmann/art-dupl.git && cd art-dupl && just build
```

## Quick Start

```bash
# Basic usage
./art-dupl

# Higher threshold (larger clones only)
./art-dupl -t 100

# HTML report with dark theme
./art-dupl -html > report.html

# JSON output for CI/CD
./art-dupl -json -t 20

# Semantic-aware detection (opt-in with --semantic flag)
# Matches by identifier names to reduce false positives
./art-dupl --semantic ./src

# Parallel parsing for faster analysis (auto-detect CPU cores)
./art-dupl --workers 0 ./src

# Use 8 workers for file parsing
./art-dupl --workers 8 ./src

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
- **Semantic-aware detection** (opt-in with `--semantic` flag) for content-aware duplicate matching
  - Distinguishes methods by receiver type (e.g., `CrushMode.IsValid` vs `SafetyMode.IsValid`)
  - Matches by identifier names, not just AST structure
  - Default: structural-only matching for backward compatibility
- **Multi-language support**: Go files and `.templ` templates (templ.guide)
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
./art-dupl -config dupl.json
```

## CLI Flags

```
-a, --all                    Generate all output formats for all detection methods
    --cache-dir              Cache directory for AST caching (default: .cache/art-dupl)
    --clear-cache            Clear cache before running
-c, --config                 Path to configuration file (JSON format)
-m, --detection-methods      Detection methods: hash, art-dupl, or hash,art-dupl (default: art-dupl)
    --exclude-pattern        Additional file patterns to exclude
-f, --files                  Read file names from stdin, one per line
    --filter-generated       Enable filtering of sqlc.dev and templ.guide generated code
-h, --help                   Help for art-dupl
    --html                   Output results as HTML with syntax-highlighted code fragments
    --include-pattern        File patterns to always include (takes precedence over filter)
    --include-sqlc           Include sqlc.dev generated files (override auto-detection)
    --include-templ          Include templ.guide generated files (override default filtering)
    --incremental            Enable incremental analysis (only analyze changed files)
-j, --json                   Output structured JSON format with metadata and statistics
-o, --output-dir             Output directory for generated files (used with --all)
-p, --plumbing               Output machine-readable plumbing format for script integration
    --semantic               Enable semantic-aware detection (match by identifier names)
    --since                  Git reference for incremental mode (e.g., HEAD~1, main)
-s, --sort                   Sort clone groups: size, occurrence, hash, total-tokens (default: size)
    --structural             Use structural-only matching [deprecated: this is now default]
-t, --threshold              Minimum token sequence size (default: 15)
    --vendor                 Include vendor directory in analysis
-v, --verbose                Enable verbose logging (repeat for more verbosity)
    --version                Version for art-dupl
    --workers                Number of concurrent workers (0 = auto-detect CPU cores)
```

### Supported Languages

| Language | Extension | Support Level |
| -------- | --------- | ------------- |
| Go       | `.go`     | Full analysis |
| Templ    | `.templ`  | Full analysis |

**Note:** `.templ` files (from [templ.guide](https://templ.guide)) are fully analyzed for code clones.
By default, templ files are filtered as generated code. Use `--include-templ` to analyze them.

### Subcommands

**stats** - Show aggregated duplication statistics

```bash
art-dupl stats [flags] [paths...]
```

Supports all root command flags plus:

```
-t, --threshold          Minimum token size (default: 15)
-m, --detection-methods  Detection methods to use
    --format             Output format: text, json, csv (default: text)
    --semantic           Enable semantic-aware detection
    --structural         Use structural-only matching [deprecated]
```

**Format validation**: The `--format` flag accepts "text", "json", or "csv".

## Examples

### CI/CD Integration

```bash
# Fail build if too many duplicates
TOTAL_CLONES=$(art-dupl -json . | jq '.summary.total_clones')
if [ "$TOTAL_CLONES" -gt 100 ]; then
  echo "Too many code duplicates: $TOTAL_CLONES"
  exit 1
fi
```

### Analysis

```bash
# Find test file duplicates
find . -name '*_test.go' | art-dupl -files

# Analyze specific paths
./art-dupl ./src ./lib -t 50
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

## Migration Guide

See [MIGRATION_GUIDE.md](docs/MIGRATION_GUIDE.md) for comprehensive guide on migrating from primitive types to domain types.

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
  - Transformations for Templ templates (`syntax/templ/`)
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
