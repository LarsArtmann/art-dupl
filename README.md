# dupl

**dupl** is a Go tool for finding code clones using suffix tree algorithms on serialized ASTs. It identifies structural duplicates while ignoring literal values.

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
./dupl

# Higher threshold (larger clones only)
./dupl -t 100

# HTML report
./dupl -html > report.html

# JSON output (new in this fork)
./dupl -json -t 20
```

## Key Features

- **Structural clone detection** using suffix tree algorithms
- **JSON output** for CI/CD automation
- **Configuration files** for team consistency
- **Multiple output formats**: text, HTML, JSON, plumbing

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
-config string     Configuration file path
-files             Read file names from stdin
-html              HTML output with code fragments
-json              JSON output (new)
-plumbing          Machine-readable output
-t, -threshold     Minimum token size (default 15)
-vendor            Include vendor directory
-v, -verbose       Verbose logging
```

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