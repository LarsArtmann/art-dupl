# Troubleshooting Guide

## Common Issues

### Installation

#### Issue: `go install` fails with "module not found"

**Solution:**
```bash
# Ensure Go 1.26+ is installed
go version

# Use the correct module path
go install github.com/LarsArtmann/art-dupl/cmd/art-dupl@latest
```

#### Issue: Binary not found in PATH

**Solution:**
```bash
# Add Go bin to PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Or use full path
~/go/bin/art-dupl --help
```

### Runtime Issues

#### Issue: "no Go files found"

**Cause:** art-dupl only processes `.go` and `.templ` files by default.

**Solution:**
```bash
# Check file extensions
ls -la *.go

# Include vendor if needed
art-dupl --vendor ./...
```

#### Issue: High memory usage on large codebases

**Solution:**
```bash
# Use threshold to filter small clones
art-dupl -t 100 ./...

# Process specific directories
art-dupl ./pkg ./cmd
```

#### Issue: "too many open files"

**Solution:**
```bash
# Increase file descriptor limit
ulimit -n 4096

# Or use --files flag with find
find . -name '*.go' | art-dupl --files
```

### Configuration

#### Issue: Config file not loading

**Solution:**
```bash
# Check config file syntax
art-dupl --config dupl.json ./...

# Validate JSON
jq . dupl.json
```

#### Issue: SQLC files not being filtered

**Solution:**
```bash
# Enable filtering
art-dupl --filter-generated ./...

# Or explicitly exclude
art-dupl --exclude-pattern "*_gen.go" ./...
```

### Performance

#### Issue: Slow analysis

**Solutions:**
1. Use hash detection (faster): `art-dupl -m hash ./...`
2. Increase threshold: `art-dupl -t 50 ./...`
3. Exclude vendor: `art-dupl --vendor=false ./...`
4. Use incremental mode: `art-dupl --incremental --since HEAD~1 ./...`

#### Issue: Out of memory

**Solutions:**
1. Reduce concurrency: `GOMAXPROCS=2 art-dupl ./...`
2. Process smaller batches
3. Use --threshold flag with higher value

## Getting Help

1. Check version: `art-dupl version`
2. Run with verbose: `art-dupl -vv ./...`
3. Check docs: [HOW_TO_USE.md](../HOW_TO_USE.md)
4. Report issues: GitHub Issues
