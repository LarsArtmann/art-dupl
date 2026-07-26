# Migration Complete: duplicates → art-dupl

## Summary

All functionality from the `duplicates` project has been successfully merged into `art-dupl`. The migration enhances `art-dupl` with valuable features from the simpler `duplicates` tool while maintaining its advanced capabilities.

## What's New in art-dupl

### 1. Efficient LineIndex ✨

Binary search-based line number lookups (O(log n)) for better performance on large files.

**Use**: Automatically integrated into position package
**Tests**: ✅ All tests passing

### 2. Simple JSON Format 📄

Legacy JSON format matching duplicates project exactly, for backward compatibility.

**CLI Flag**: `--simple-json`
**Example**:

```bash
art-dupl --simple-json > report.json
```

### 3. Impact Scoring 📊

Simple scoring metric: `tokens × instances` - complements existing complexity_score.

**Use**: Available in both JSON formats
**Benefits**: Prioritize by total duplicated code volume

## Quick Migration Guide

### If You're Using duplicates

Replace these commands:

```bash
# Old (duplicates):
duplicates -threshold 20 -json report.json -html report.html -text report.txt

# New (art-dupl):
art-dupl --all --output-dir ./reports --threshold 20
```

### Flag Mapping

| duplicates Flag      | art-dupl Flag                 | Notes                             |
| -------------------- | ----------------------------- | --------------------------------- |
| `-threshold N`       | `--threshold N` or `-t N`     | Same functionality                |
| `-json`              | `--simple-json`               | Legacy format                     |
| `--json`             | `--json`                      | Enhanced format with metadata     |
| `-html`              | `--html`                      | Same functionality                |
| `-text`              | default (no flag)             | Text is default                   |
| `-plumbing`          | `--plumbing`                  | Same functionality                |
| `-v`                 | `--verbose` or `-v`           | Same functionality                |
| `-exclude "pattern"` | `--exclude-pattern "pattern"` | More powerful patterns            |
| `-json report.json`  | `--simple-json > report.json` | Use redirection                   |
| `-html report.html`  | `--html > report.html`        | Use redirection or `--output-dir` |

### New Features to Try

```bash
# Generate all formats at once
art-dupl --all --output-dir ./reports

# Sort by occurrence (most files first)
art-dupl --sort occurrence

# Filter out auto-generated code
art-dupl --filter-generated

# Use config file for team consistency
art-dupl --config dupl.json

# Performance profiling
art-dupl --profile
```

## Output Formats

### Simple JSON (Legacy)

Format from duplicates project - simple and straightforward.

```bash
art-dupl --simple-json
```

Output:

```json
[
	{
		"hash": "abc123...",
		"score": 150,
		"instances": [
			{
				"filename": "file.go",
				"start_line": 45,
				"end_line": 78,
				"token_count": 50
			}
		]
	}
]
```

### Enhanced JSON

Rich format with metadata and statistics.

```bash
art-dupl --json
```

Output:

```json
{
  "version": "1.0",
  "timestamp": "2026-01-13T...",
  "threshold": 15,
  "files_analyzed": 100,
  "clone_groups": [...],
  "summary": {
    "total_clone_groups": 10,
    "total_clones": 20,
    "complexity_score": 1.8,
    "impact_score": 1500
  }
}
```

## Scoring Metrics

### Impact Score (from duplicates)

```
score = token_count × instance_count
```

Measures total duplicated code volume. Use to prioritize refactoring by how much code is duplicated.

### Complexity Score (from art-dupl)

```
complexity_score = total_clones / clone_groups
```

Measures duplication density. Use to assess overall codebase complexity.

Both are included in enhanced JSON format!

## Installation

```bash
# Install art-dupl
go install github.com/LarsArtmann/art-dupl@latest

# Or build from source
cd /path/to/art-dupl
export GOEXPERIMENT=jsonv2  # Required for encoding/json/v2
go build ./cmd/art-dupl

# Run
./dist/art-dupl --help
```

## Documentation

- **Migration Report**: See `MIGRATION_REPORT_duplicates.md` for complete technical details
- **Full Documentation**: See `README.md` for all features
- **API Docs**: See `docs/api/` for programmatic usage

## Status

✅ **Migration Complete**

- All valuable features migrated
- Tests passing
- Documentation updated
- Zero breaking changes

art-dupl is now a superset of duplicates with enhanced functionality.

---

**Date**: January 13, 2026
**Status**: Production Ready ✅
