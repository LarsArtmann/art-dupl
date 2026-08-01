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

### Global state and `t.Parallel()` — do not mix

`t.Parallel()` is for tests that touch **only their own** state. A test that
mutates **process-global** state must be serial, because the race detector flags
the unsynchronized read/write of the shared global even when the mutation is
"correct."

The main offender is `os.Stdout`/`os.Stderr`:

- `testutil.CaptureStdoutStderr` (and `CaptureCombinedOutput`, `cmd.executeTestCommand`)
  swap the global `os.Stdout`/`os.Stderr` pointers. A mutex serializes captures
  against **each other**, but NOT against sibling tests that read those globals
  directly (e.g. `fmt.Printf` → `os.Stdout`, `fmt.Fprintln(os.Stderr, …)`).
- So while a capture test is parallel, any parallel sibling that reads
  `os.Stdout`/`os.Stderr` races on the global pointer.

Rule: **a test that captures or reassigns `os.Stdout`/`os.Stderr` must NOT call
`t.Parallel()`** (or must guarantee no sibling test in the package reads those
globals directly). Prefer keeping such tests serial — the parallelism win on a
fast capture test is negligible and the race is silent and intermittent.

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
export GOEXPERIMENT=jsonv2    # Required for encoding/json/v2
go test ./...                               # All tests
go test -race ./...                         # With race detector
go test $(go list ./... | grep -v /bdd)     # Unit tests only (excludes BDD)
go test ./cmd/... ./bdd/...                 # Integration tests only
golangci-lint run --timeout 5m ./...        # Lint
nix flake check                             # Full CI (format + lint + test)
```

## Self-Test Gate

The `self-test` Nix check enforces the **zero-duplication invariant**: art-dupl must detect zero clones in its own source at threshold 1.

```bash
nix build .#checks.x86_64-linux.self-test
```

This builds art-dupl, runs `art-dupl -t 1 --plumbing .` on the full source tree, and fails if any output is produced. The invariant means: **art-dupl must never ship with duplication that it itself can detect**. If this check fails, either fix the duplication or lower the detection sensitivity (but prefer fixing the duplication).
