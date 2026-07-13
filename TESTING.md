# Testing Guide for art-dupl

## Testing Strategy

The project uses two complementary testing approaches:

- **Unit tests**: Native Go `testing` package — for algorithms, domain types, config validation
- **BDD tests**: Ginkgo/Gomega in `bdd/` — for user-facing workflows, CLI integration, feature scenarios

Both are first-class citizens. Do not migrate one to the other.

## Unit Tests (Native Go)

### Structure

- Test files: `*_test.go` next to source files
- Table-driven tests for multiple scenarios
- `t.Parallel()` for independent tests
- `t.Helper()` in assertion helpers

```go
func TestFunction(t *testing.T) {
    t.Parallel()

    tests := []struct {
        name     string
        input    InputType
        expected ExpectedType
    }{
        {name: "scenario 1", input: input1, expected: expected1},
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            result := FunctionUnderTest(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## BDD Tests (Ginkgo/Gomega)

Location: `bdd/` directory. Helpers in `internal/testutil/bdd.go`.

Key helpers: `NewBDDTestSetupForGinkgo()`, `CreateTestFiles()`, `CreateDuplicateFiles()`, `RunArtDupl()`, `RunArtDuplOnDir()`, `RunArtDuplWithStdin()`, `Cleanup()`.

Covers: CLI commands, config files, filtering, sorting, stats subcommand, plumbing output, SDK workflows.

```bash
go test -v ./bdd          # Run BDD tests
go test -v ./bdd -run "..." # Specific test
```

## Benchmarks

```go
func BenchmarkFunction(b *testing.B) {
    b.ReportAllocs()
    setupData := prepareTestData()
    b.ResetTimer()
    for b.Loop() {
        ImportantFunction(setupData)
    }
}
```

```bash
go test -bench=. ./...          # Run all benchmarks
go test -bench=. -benchmem ./... # With allocation reporting
```

## Fuzzing

```go
func FuzzSerialize(f *testing.F) {
    f.Add("package main\nfunc main() {}")
    f.Fuzz(func(t *testing.T, input string) {
        node := Parse(input)
        tokens := Serialize(node)
        if len(tokens) == 0 && len(input) > 0 {
            t.Errorf("empty tokens for non-empty input")
        }
    })
}
```

```bash
go test -fuzz=FuzzParseBytes -fuzztime=10s ./syntax/templ/  # Run fuzz tests
go test -fuzz=FuzzParseBytes -fuzztime=60s ./syntax/templ/  # Longer duration
```

## Coverage

Target: 80% minimum.

```bash
go test -coverprofile=coverage.out ./...   # Tests with coverage
go tool cover -html=coverage.out           # HTML coverage report → browser
go tool cover -func=coverage.out           # Summary in terminal
```

## Running Tests

```bash
go test ./...                               # All tests
go test -race ./...                         # With race detector
go test $(go list ./... | grep -v /bdd)     # Unit tests only (excludes BDD)
go test ./cmd/... ./bdd/...                 # Integration tests only
golangci-lint run --timeout 5m ./...        # Lint
nix flake check                             # Full CI (format + lint + test)
```
