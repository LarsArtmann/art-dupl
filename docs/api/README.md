# API Documentation

This directory contains generated API documentation for the art-dupl project.

## Files

### API.md
Comprehensive Markdown documentation covering:
- Package index with descriptions
- Key types and structures
- Key functions with parameters
- Usage examples
- Integration points
- Error handling
- Testing guides
- Contributing guidelines

### api-go.txt
Raw Go API documentation generated using `go doc -all`.

## Generating Documentation

### Markdown Documentation
```bash
# Already generated in docs/api/API.md
# View directly or convert to HTML
cat docs/api/API.md
```

### Go Doc Format
```bash
# Generate raw Go doc
go doc -all github.com/LarsArtmann/art-dupl > docs/api/api-go.txt
```

### HTML Documentation (Requires godoc)
```bash
# Install godoc
go install golang.org/x/tools/cmd/godoc@latest

# Generate HTML documentation
godoc -html github.com/LarsArtmann/art-dupl > docs/api/api-go.html
```

## Documentation Coverage

### Packages Documented
- ✅ cli - Command-line interface
- ✅ config - Configuration management
- ✅ detection - Multi-detector coordination
- ✅ hash - SHA1 hash-based detection
- ✅ job - Job orchestration and profiling
- ✅ printer - Output formatting
- ✅ suffixtree - Suffix tree implementation
- ✅ syntax - AST handling and serialization
- ✅ types - Common type definitions
- ✅ errors - Rich error handling
- ✅ util - Utility functions
- ✅ testutils - Test helpers

### Topics Covered
- ✅ Package overview
- ✅ Type definitions
- ✅ Function signatures
- ✅ Usage examples
- ✅ Integration points
- ✅ Error handling patterns
- ✅ Performance profiling
- ✅ Testing guidelines
- ✅ Contributing instructions

## Quick Links

- [Main README](../README.md) - Project overview
- [HOW_TO_USE](../HOW_TO_USE.md) - Usage guide
- [FEATURES](../FEATURES.md) - Feature documentation
- [API.md](./API.md) - Comprehensive API docs

## License

See LICENSE file for details.
