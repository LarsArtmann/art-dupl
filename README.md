# art-dupl

[![Go Report Card](https://goreportcard.com/badge/github.com/LarsArtmann/art-dupl)](https://goreportcard.com/report/github.com/LarsArtmann/art-dupl) [![codecov](https://codecov.io/gh/LarsArtmann/art-dupl/graph/badge.svg?token=art-dupl)](https://codecov.io/gh/LarsArtmann/art-dupl)

**art-dupl** is a professional code duplication detection tool for Go projects. It finds repeated code patterns in your source files, helping you identify opportunities for refactoring and reduce code duplication.

The tool analyzes Go source code at the AST level, ignoring literal values (like strings and numbers) to focus on structural patterns. It supports semantic-aware detection that considers identifier names to reduce false positives.

## Installation

```bash
go install github.com/LarsArtmann/art-dupl/cmd/art-dupl@latest
```

Or build from source:

```bash
git clone https://github.com/LarsArtmann/art-dupl.git
cd art-dupl
just build    # Creates ./dist/art-dupl
./dist/art-dupl
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

# Structural-only detection (default, matches by AST structure only)
# Use --semantic to enable identifier-aware matching
./art-dupl ./src

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
- **Semantic-aware detection** (optional) reduces false positives by matching identifier names
  - Distinguishes methods by receiver type (e.g., `CrushMode.IsValid` vs `SafetyMode.IsValid`)
  - Enable with `--semantic`; default is structural-only matching
- **Smart filtering**: Auto-detects and filters SQLC and Templ generated code with configurable patterns
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
    --diff                      Enable diff visualization for HTML output (values: side-by-side, inline, or true for side-by-side)
    --exclude-pattern           Additional file patterns to exclude
    -f --files                  Read file names from stdin, one per line
    --filter-generated          Enable filtering of sqlc.dev and templ.guide generated code
    -h --help                   Help for art-dupl
    --html                      Output results as HTML with syntax-highlighted code fragments
    --include-node-modules      Include node_modules directory in hash-based detection (excluded by default)
    --include-pattern           File patterns to always include (takes precedence over filter)
    --include-sqlc              Include sqlc.dev generated files (override auto-detection)
    --exclude-templ            Exclude .templ source files from analysis
    --incremental               Enable incremental analysis with AST caching
-j, --json                   Output structured JSON format with metadata and statistics
-o, --output-dir             Output directory for generated files (used with --all)
-p, --plumbing               Output machine-readable plumbing format for script integration
    --semantic               Enable semantic-aware detection (match by identifier names; off by default)
    --since                  Git reference for incremental mode (e.g., HEAD~1, main)
-s, --sort                   Sort clone groups: size, occurrence, hash, total-tokens (default: size)
    --structural            Use structural-only matching (this is already the default)
-t, --threshold              Minimum token sequence size (default: 15)
    --vendor                 Include vendor directory in analysis
-v, --verbose                Enable verbose logging (repeat for more verbosity)
    --version                Version for art-dupl
    --workers                Number of concurrent workers (0 = auto-detect CPU cores)
    --sarif                  Output SARIF format (GitHub Advanced Security, CodeQL)
    --simple-json            Output simplified JSON format with impact scores
    --only string            Only analyze specific file type: 'go' or 'templ' (default: both)
    --diff string            Enable diff visualization for HTML output (side-by-side, inline)
```

**Sort options** (`-s`/`--sort`): `size` (largest first), `occurrence` (most files first), `hash` (alphabetical), `total-tokens` (highest total token count = size × occurrences). Default: `size`.

### Supported Languages

| Language | Extension | Support Level |
| -------- | --------- | ------------- |
| Go       | `.go`     | Full analysis |
| Templ    | `.templ`  | Full analysis |

**Note:** `.templ` files (from [templ.guide](https://templ.guide)) are fully analyzed for code clones.
By default, templ files are **included** in analysis. Use `--exclude-templ` to exclude them.

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
    --structural         Use structural-only matching (may increase false positives)
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
- **HTML**: Detailed report with syntax-highlighted code fragments (supports `--diff` visualization)
- **JSON**: Structured data with metadata and summary statistics
- **Simple JSON**: Lightweight JSON output with impact scores (`--simple-json`)
- **SARIF**: Static Analysis Results Interchange Format for GitHub Advanced Security (`--sarif`)
- **Plumbing**: Machine-readable format for scripts

## Architecture

- **suffixtree/** - Core suffix tree implementation
- **syntax/** - AST handling and serialization
- **job/** - File parsing orchestration
- **printer/** - Output formatting

## Testing

```bash
just test  # Run all tests
just check # Run linting
```

Or with make:

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
