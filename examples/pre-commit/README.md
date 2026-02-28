# Pre-commit Hook Example

This directory contains example pre-commit hooks for art-dupl.

## Quick Setup

1. Install pre-commit:

   ```bash
   pip install pre-commit
   # or
   brew install pre-commit
   ```

2. Copy the config to your project:

   ```bash
   cp .pre-commit-config.yaml /path/to/your/project/.pre-commit-config.yaml
   ```

3. Install the hooks:
   ```bash
   cd /path/to/your/project
   pre-commit install
   ```

## Configuration Options

### Basic (Recommended)

Runs art-dupl on all Go files before each commit:

```yaml
repos:
  - repo: local
    hooks:
      - id: art-dupl
        name: Detect code duplication
        entry: art-dupl --plumbing -t 30
        language: system
        types: [go]
        pass_filenames: false
```

### Strict Mode

Fails if any duplication is found (threshold 15):

```yaml
repos:
  - repo: local
    hooks:
      - id: art-dupl-strict
        name: Detect code duplication (strict)
        entry: art-dupl --plumbing -t 15
        language: system
        types: [go]
        pass_filenames: false
```

### With Semantic Detection

Uses semantic-aware detection to reduce false positives:

```yaml
repos:
  - repo: local
    hooks:
      - id: art-dupl-semantic
        name: Detect code duplication (semantic)
        entry: art-dupl --plumbing --semantic -t 20
        language: system
        types: [go]
        pass_filenames: false
```

### With Filtering

Filters generated code (SQLC, templ):

```yaml
repos:
  - repo: local
    hooks:
      - id: art-dupl-filtered
        name: Detect code duplication (filtered)
        entry: art-dupl --plumbing --filter-generated -t 25
        language: system
        types: [go]
        pass_filenames: false
```

## CI Integration

You can also run pre-commit in CI:

```yaml
# .github/workflows/pre-commit.yml
name: Pre-commit
on: [push, pull_request]
jobs:
  pre-commit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "stable"
      - uses: actions/setup-python@v5
        with:
          python-version: "3.x"
      - run: pip install pre-commit
      - run: go install github.com/LarsArtmann/art-dupl@latest
      - uses: pre-commit/action@v3.0.1
```

## Bypassing Hooks

To bypass pre-commit hooks temporarily:

```bash
git commit --no-verify
```

**Note:** Only use this for WIP commits, not for code intended to be merged.
