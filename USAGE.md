# dupl Usage Reference

Complete reference documentation for the dupl code duplication detection tool.

## Table of Contents

1. [Command Line Interface](#command-line-interface)
2. [Configuration Files](#configuration-files)
3. [Output Formats](#output-formats)
4. [Command Options](#command-options)
5. [Exit Codes](#exit-codes)
6. [Environment Variables](#environment-variables)
7. [Examples](#examples)

## Command Line Interface

### Basic Syntax

```bash
dupl [flags] [paths...]
```

- **flags**: Command-line options (see [Command Options](#command-options))
- **paths**: Files or directories to analyze (optional, defaults to current directory)

### Path Handling

| Input Type | Behavior | Example |
|------------|----------|---------|
| No path | Recursively scan `.` (current directory) for `.go` files | `dupl` |
| File path | Analyze specific file regardless of extension | `dupl main.go` |
| Directory | Recursively search directory for `.go` files | `dupl ./src` |
| Multiple paths | Analyze all specified paths | `dupl ./src ./lib` |

### File Selection Rules

- Only `.go` files are analyzed by default
- Vendor directory excluded unless `-vendor` flag used
- Files specified directly are analyzed regardless of extension
- Hidden files (starting with `.`) are skipped

## Configuration Files

### File Format

Configuration files use JSON format with the following structure:

```json
{
  "threshold": 30,
  "includeVendor": false,
  "filesFromStdin": false,
  "outputFormat": "text",
  "verbose": false,
  "paths": ["./cmd", "./internal"],
  "ignoreFiles": ["*_test.go", "*_gen.go"],
  "maxChildrenSerial": 10000,
  "outputFile": "report.json"
}
```

### Configuration Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `threshold` | integer | 15 | Minimum token sequence size to consider as clone |
| `includeVendor` | boolean | false | Include vendor directory in analysis |
| `filesFromStdin` | boolean | false | Read file paths from stdin (one per line) |
| `outputFormat` | string | "text" | Output format: "text", "html", "json", "plumbing" |
| `verbose` | boolean | false | Enable verbose logging |
| `paths` | array[string] | ["."] | Paths to analyze |
| `ignoreFiles` | array[string] | [] | File patterns to ignore |
| `maxChildrenSerial` | integer | 10000 | Maximum children for large composite literals |
| `outputFile` | string | "" | Output file path (empty = stdout) |

### Using Configuration Files

```bash
# Load configuration from file
dupl -config dupl.json

# Combine with CLI flags (CLI overrides config)
dupl -config dupl.json -t 50 -html

# Example config file path
dupl -config ./configs/team-dupl.json ./src
```

### Configuration Precedence

1. Default values
2. Configuration file values
3. Command-line flag values (highest priority)

## Output Formats

### Text Format (Default)

Human-readable clone listings with file paths and line numbers.

```
internal/handlers/auth.go:45-73
internal/handlers/user.go:89-117

internal/middleware/auth.go:123-145
internal/middleware/cors.go:67-89
internal/middleware/logging.go:234-256
```

**Use cases**: Quick terminal review, debugging, manual inspection

### HTML Format

Interactive HTML report with syntax-highlighted code fragments.

```bash
dupl -html -t 30 > report.html
```

**Features**:
- Syntax highlighting for Go code
- Side-by-side clone comparison
- Navigation between clone groups
- Statistics summary

**Use cases**: Code reviews, detailed analysis, team presentations

### JSON Format

Structured data with metadata and statistics for automation.

```bash
dupl -json -t 20 > report.json
```

**Output Structure**:
```json
{
  "summary": {
    "total_clones": 15,
    "files_analyzed": 42,
    "threshold": 20
  },
  "clones": [
    {
      "hash": "abc123def456...",
      "instances": [
        {
          "file": "internal/handlers/auth.go",
          "start": 45,
          "end": 73,
          "lines": 29
        },
        {
          "file": "internal/handlers/user.go",
          "start": 89,
          "end": 117,
          "lines": 29
        }
      ]
    }
  ]
}
```

**Use cases**: CI/CD integration, automated reporting, programmatic analysis

### Plumbing Format

Machine-readable format for script integration.

```bash
dupl -plumbing -t 15
```

**Output format**:
```
filename:start:line:end:line
filename:start:line:end:line
```

**Use cases**: Shell scripts, text processing, simple automation

## Command Options

### Core Options

| Option | Short | Type | Default | Description |
|--------|-------|------|---------|-------------|
| `-config` | | string | "" | Path to configuration file (JSON format) |
| `-t` | `-threshold` | int | 15 | Minimum token sequence size to consider as clone |
| `-vendor` | | flag | false | Include vendor directory in analysis |

### Output Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `-html` | flag | false | Output HTML report with syntax-highlighted code |
| `-json` | flag | false | Output structured JSON format with metadata |
| `-plumbing` | flag | false | Output machine-readable plumbing format |

### Input Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `-files` | flag | false | Read file names from stdin, one per line |
| `-v` | `-verbose` | flag | false | Enable verbose logging to show processing progress |

### Output Format Conflicts

The following flags are mutually exclusive (can only use one at a time):
- `-html`
- `-json` 
- `-plumbing`

Attempting to use multiple output formats will result in an error.

## Exit Codes

| Exit Code | Meaning | Common Causes |
|-----------|---------|----------------|
| 0 | Success | Analysis completed successfully |
| 1 | Error | Invalid arguments, file I/O errors, configuration errors |
| 2 | Usage | Invalid command-line usage (shown when `-h` or invalid flags) |

### Error Handling

dupl follows standard Unix conventions:
- Exit code 0 for success
- Non-zero exit codes for errors
- Error messages written to stderr
- Normal output written to stdout (unless redirected)

## Environment Variables

dupl does not currently use any environment variables for configuration. All settings are provided through command-line flags or configuration files.

## Examples

### Basic Usage

```bash
# Analyze current directory with default settings
dupl

# Analyze specific directories with higher threshold
dupl -t 50 ./src ./lib

# Verbose output for large codebases
dupl -v -t 30 ./internal
```

### Output Format Examples

```bash
# HTML report
dupl -html -t 30 > report.html

# JSON for CI/CD
dupl -json -t 20 | jq '.summary.total_clones'

# Plumbing format for scripts
dupl -plumbing -t 15 | while read line; do
  echo "Found clone: $line"
done
```

### Configuration Examples

```bash
# Team configuration
dupl -config team-config.json ./src

# Override config for specific analysis
dupl -config base.json -t 100 -html > large-clones.html

# CI/CD configuration
dupl -config ci-config.json -json > ci-report.json
```

### File Selection Examples

```bash
# Analyze test files only
find . -name '*_test.go' | dupl -files -t 20

# Exclude generated code
dupl -t 30 $(find . -name '*.go' ! -name '*_gen.go')

# Multiple specific files
dupl file1.go file2.go file3.go -t 15
```

### Advanced Workflows

```bash
# Progressive analysis
echo "=== All clones ===" && dupl -t 15
echo "=== Medium clones ===" && dupl -t 50  
echo "=== Large clones ===" && dupl -t 100

# Package-by-package analysis
for pkg in cmd internal pkg; do
  echo "=== Analyzing $pkg ==="
  dupl -t 30 ./$pkg
done

# Monthly comparison
MONTH=$(date +%Y-%m)
dupl -json -t 30 > "reports/dupl-$MONTH.json"
```

### Integration Examples

#### Git Hook Integration

```bash
#!/bin/sh
# pre-commit hook

# Skip if config file doesn't exist
if [ ! -f .duplrc ]; then
  exit 0
fi

# Run analysis
dupl -config .duplrc -json > .dupl_check.json
CLONES=$(jq '.summary.total_clones' .dupl_check.json)

# Fail if too many clones
if [ "$CLONES" -gt $(jq '.threshold' .duplrc) ]; then
  echo "Too many code clones: $CLONES"
  echo "Run 'dupl -config .duplrc -html' to review"
  exit 1
fi
```

#### Makefile Integration

```makefile
.PHONY: dupl-check dupl-report

dupl-check:
	@echo "Running duplication check..."
	@dupl -t 30 -json > .dupl_output.json
	@CLONES=$$(jq '.summary.total_clones' .dupl_output.json); \
	if [ $$CLONES -gt 50 ]; then \
		echo "❌ Too many clones: $$CLONES"; \
		exit 1; \
	else \
		echo "✅ Clone count acceptable: $$CLONES"; \
	fi

dupl-report:
	@mkdir -p reports
	@dupl -html -t 30 > reports/dupl_$(shell date +%Y%m%d).html
	@echo "HTML report generated: reports/dupl_$(shell date +%Y%m%d).html"
```

## Performance Considerations

### Memory Usage

- Large composite literals are serialized to prevent stack overflow
- Use `maxChildrenSerial` to control memory usage for large files
- Stream processing keeps memory usage bounded during analysis

### Processing Speed

- Processing time scales with number of files and clone threshold
- Higher thresholds reduce processing time and memory usage
- Verbose mode adds minimal overhead

### Optimization Tips

```bash
# For very large codebases
dupl -t 50 -v  # Higher threshold, verbose to monitor progress

# Limit analysis scope
dupl -t 30 ./critical/packages  # Focus on important code

# Use configuration for consistent settings
dupl -config production.json
```

## Troubleshooting

### Common Issues

#### "No such file or directory"
- Check file paths exist
- Verify working directory
- Use absolute paths if needed

#### "Permission denied"
- Check file read permissions
- Ensure directory traversal permissions

#### "config file not found"
- Verify configuration file path
- Check file exists and is readable

#### Performance Issues
- Increase threshold to reduce processing
- Limit analysis scope with specific paths
- Monitor with verbose mode

### Debug Options

```bash
# Show processing progress
dupl -v

# Check configuration
cat dupl.json | jq '.'

# Test file selection
find . -name '*.go' | head -10
```

This reference provides complete documentation for using dupl in various scenarios. For practical examples and workflows, see [HOW_TO_USE.md](HOW_TO_USE.md).