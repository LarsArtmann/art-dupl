# How to Use art-dupl - A Practical Guide

This guide provides practical examples and workflows for using **art-dupl** effectively in real-world scenarios.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Basic Workflows](#basic-workflows)
3. [CI/CD Integration](#cicd-integration)
4. [Advanced Scenarios](#advanced-scenarios)
5. [Output Interpretation](#output-interpretation)
6. [Best Practices](#best-practices)

## Quick Start

### Installation

```bash
# Install directly
go install github.com/LarsArtmann/art-dupl@latest

# Or build from source
git clone https://github.com/LarsArtmann/art-dupl.git && cd art-dupl
make build
```

### First Analysis

```bash
# Analyze current directory
./art-dupl

# See what's happening
./art-dupl -v
```

## Basic Workflows

### 1. Finding Large Duplicates

For meaningful refactoring opportunities, focus on larger code blocks:

```bash
# Find only substantial duplicates (50+ tokens)
./art-dupl -t 50

# Even larger for architectural issues
./art-dupl -t 100
```

### 2. Creating Reports

#### HTML Report for Review
```bash
# Generate comprehensive HTML report
./art-dupl -html -t 30 > dupl_report.html

# Open in browser
open dupl_report.html  # macOS
xdg-open dupl_report.html  # Linux
```

#### JSON for Automation
```bash
# JSON output with statistics
./art-dupl -json -t 20 > dupl_report.json

# Pretty-print JSON
./art-dupl -json -t 20 | jq '.' > dupl_report_pretty.json
```

### 3. Analyzing Specific Code

#### Target Directories
```bash
# Analyze only source code (exclude tests)
./art-dupl ./cmd ./internal -t 30

# Analyze specific package
./art-dupl ./internal/handlers -t 25
```

#### Target File Types
```bash
# Analyze test files only
find . -name '*_test.go' | ./art-dupl -files -t 20

# Analyze generated files separately
find . -name '*_gen.go' | ./art-dupl -files -t 50
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Code Duplication Check
on: [push, pull_request]

jobs:
  art-dupl-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: '1.21'
      
      - name: Install art-dupl
        run: go install github.com/LarsArtmann/art-dupl@latest
      
      - name: Run art-dupl analysis
        run: |
          art-dupl -json -t 30 > dupl_report.json
          TOTAL_CLONES=$(jq '.summary.total_clones' dupl_report.json)
          echo "Found $TOTAL_CLONES code clones"
          
          # Fail if too many duplicates
          if [ "$TOTAL_CLONES" -gt 50 ]; then
            echo "Too many code duplicates: $TOTAL_CLONES"
            exit 1
          fi
```

### Pre-commit Hook

```bash
#!/bin/sh
# .git/hooks/pre-commit

echo "Running duplication check..."
art-dupl -json -t 25 > /tmp/dupl_check.json
CLONES=$(jq '.summary.total_clones' /tmp/dupl_check.json)

if [ "$CLONES" -gt 10 ]; then
  echo "⚠️  Found $CLONES code clones. Consider refactoring before committing."
  echo "Run 'art-dupl -html -t 25 > report.html' to see details."
  exit 1
fi
```

## Advanced Scenarios

### 1. Using Configuration Files

Create `dupl.json` for team consistency:

```json
{
  "threshold": 30,
  "outputFormat": "json",
  "paths": ["./cmd", "./internal"],
  "includeVendor": false,
  "ignoreFiles": ["*_test.go", "*_mock.go", "*_gen.go"],
  "verbose": true,
  "outputFile": "reports/art-dupl.json"
}
```

```bash
# Use configuration
./art-dupl -config dupl.json

# Override specific settings
./art-dupl -config dupl.json -t 50 -html
```

### 2. Progressive Refactoring Workflow

```bash
# Step 1: Get baseline
./art-dupl -json -t 15 > baseline.json
echo "Baseline: $(jq '.summary.total_clones' baseline.json) clones"

# Step 2: Focus on largest duplicates first
./art-dupl -t 100 -html > large_clones.html

# Step 3: After refactoring, measure improvement
./art-dupl -json -t 15 > after_refactor.json
echo "After: $(jq '.summary.total_clones' after_refactor.json) clones"

# Step 4: Compare
echo "Improvement: $(jq '.summary.total_clones' baseline.json) → $(jq '.summary.total_clones' after_refactor.json')"
```

### 3. Monitoring Technical Debt

```bash
#!/bin/bash
# art-dupl_monitor.sh - Monthly technical debt report

DATE=$(date +%Y-%m)
REPORT_DIR="reports/$DATE"
mkdir -p "$REPORT_DIR"

# Generate different views
./art-dupl -json -t 15 > "$REPORT_DIR/all_clones.json"
./art-dupl -json -t 50 > "$REPORT_DIR/large_clones.json"
./art-dupl -html -t 30 > "$REPORT_DIR/detailed_report.html"

# Extract metrics
ALL_CLONES=$(jq '.summary.total_clones' "$REPORT_DIR/all_clones.json")
LARGE_CLONES=$(jq '.summary.total_clones' "$REPORT_DIR/large_clones.json")
FILES_ANALYZED=$(jq '.summary.files_analyzed' "$REPORT_DIR/all_clones.json")

echo "=== $DATE Duplication Report ==="
echo "Files analyzed: $FILES_ANALYZED"
echo "All clones (15+ tokens): $ALL_CLONES"
echo "Large clones (50+ tokens): $LARGE_CLONES"
echo "Duplication ratio: $(echo "scale=2; $ALL_CLONES / $FILES_ANALYZED" | bc) per file"
```

## Output Interpretation

### Understanding Text Output

```
internal/handlers/auth.go:45-73
internal/handlers/user.go:89-117
```

- `file.go:startLine-endLine` - Location of duplicate code
- Two or more lines grouped together = one clone set
- Line numbers are inclusive

### Understanding JSON Output

```json
{
  "version": "1.0",
  "timestamp": "2025-12-14T09:19:06.351297Z",
  "threshold": 15,
  "files_analyzed": 2,
  "clone_groups": [
    {
      "hash": "5e8f50b6f5a834485490605819523fd92711f92ba855f603bc2375925bc4753a",
      "size": 4,
      "files": [
        {
          "filename": "./cli.go",
          "line_start": 80,
          "line_end": 88,
          "fragment": "if err != nil {\n\tif _, err := fmt.Fprintf(..."
        }
      ]
    }
  ],
  "summary": {
    "total_clone_groups": 8,
    "total_clones": 19,
    "complexity_score": 2.11
  }
}
```

- `version` - Output format version
- `timestamp` - When analysis was performed
- `threshold` - Threshold used for analysis
- `files_analyzed` - Number of files processed
- `clone_groups` - Array of clone groups
- `hash` - Unique identifier for this clone group
- `size` - Number of files in this clone group
- `files` - Array of files with this clone pattern
- `filename` - Path to file containing clone
- `line_start` - Starting line number of clone
- `line_end` - Ending line number of clone
- `fragment` - Code fragment (truncated for display)
- `summary` - Analysis summary statistics
- `total_clone_groups` - Total number of clone groups found
- `total_clones` - Total number of individual clones
- `complexity_score` - Duplication complexity metric

### Understanding HTML Report

The HTML report provides:
- Syntax-highlighted code fragments
- Side-by-side comparison
- Easy navigation between clone sets
- Statistics summary

## Best Practices

### 1. Setting Thresholds

| Project Size | Recommended Threshold | Rationale |
|--------------|----------------------|-----------|
| Small (<1000 lines) | 10-15 | Catch even small duplicates |
| Medium (1K-10K lines) | 20-30 | Balance between noise and signal |
| Large (>10K lines) | 30-50 | Focus on meaningful duplicates |
| Codebases with boilerplate | 40+ | Avoid flagging template code |

### 2. Common Ignore Patterns

```json
{
  "ignoreFiles": [
    "*_test.go",     // Test files often repeat setup code
    "*_mock.go",     // Generated mocks
    "*_gen.go",      // Generated code
    "vendor/*",      // Dependencies
    "*.pb.go",       // Protocol buffers
    "*.mock.go",     // Mock implementations
    "testdata/*"     // Test data
  ]
}
```

### 3. Refactoring Strategy

1. **Start Large**: Begin with high threshold (50+) to find major issues
2. **Analyze Impact**: Check if clones are in critical paths
3. **Extract Functions**: Pull common code to shared functions
4. **Create Utilities**: Build helper functions for common patterns
5. **Measure Improvement**: Run art-dupl again to verify reduction

### 4. Team Guidelines

- **Reviews**: Include art-dupl reports in code reviews
- **Thresholds**: Set team-wide thresholds via config files
- **Automation**: Integrate into CI/CD with clear failure criteria
- **Documentation**: Document accepted patterns that may generate clones

## Troubleshooting

### Common Issues

#### Too Many False Positives
```bash
# Increase threshold
./art-dupl -t 40

# Or ignore certain patterns
./art-dupl -config dupl.json  # with ignoreFiles patterns
```

#### Performance Issues
```bash
# Use verbose to see progress
./art-dupl -v

# Limit scope
./art-dupl ./specific/package -t 30
```

#### Memory Issues with Large Files
```bash
# Adjust maxChildrenSerial in config
{
  "maxChildrenSerial": 5000  // Lower for large files
}
```

### Getting Help

```bash
# See all options
./art-dupl -h

# Check configuration
cat dupl.json | jq '.'
```

## Integration Examples

### with jq for Analysis

```bash
# Top 10 files with most clones
./art-dupl -json | jq -r '.clones[] | .instances[] | "\(.file)"' | sort | uniq -c | sort -nr | head -10

# Average clone size
./art-dupl -json | jq '[.clones[] | .instances[0].lines] | add / length'

# Clones by package
./art-dupl -json | jq -r '.clones[] | .instances[0].file | split("/")[0:2] | join("/")' | sort | uniq -c
```

### with grep for Context

```bash
# Find clones containing specific patterns
./art-dupl -t 20 | grep -B1 -A1 "func.*Error"

# Check if clones involve specific packages
./art-dupl -t 20 | grep -B1 -A1 "database/sql"
```

This guide should help you effectively integrate art-dupl into your development workflow. Remember that the goal is maintainable code, not zero duplication—some duplication may be acceptable for clarity or performance reasons.