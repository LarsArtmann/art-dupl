# Testing Guide for art-dupl

**Last Updated:** January 14, 2026
**Go Version:** 1.25.5+

## Overview

This guide covers the testing practices used in the art-dupl project, focusing on native Go testing tools and best practices.

## Philosophy

- **Use native Go testing tools** - No external frameworks (Ginkgo, Gomega, testify)
- **Test behavior, not implementation** - Focus on what code does, not how
- **Table-driven tests** - For multiple scenarios with different inputs
- **Parallel execution** - Enable `t.Parallel()` for independent tests
- **Clear test names** - Describe what is being tested
- **Fast tests** - Keep unit tests quick, integration tests focused

## Native Go Testing Tools

### Testing Package Types

**Type T - Unit Tests**
```go
func TestFunction(t *testing.T) {
    t.Run("case1", func(t *testing.T) {
        // Subtest logic
    })
    t.Parallel() // Enable parallel execution
    t.Cleanup(func() {
        // Cleanup resources
    })
}
```

**Type B - Benchmarks**
```go
func BenchmarkFunction(b *testing.B) {
    b.ReportAllocs() // Track memory allocations
    b.ResetTimer()   // Reset timer for setup
    for b.Loop() {
        // Code to benchmark
    }
}
```

**Type F - Fuzzing**
```go
func FuzzFunction(f *testing.F) {
    f.Add(seedInput) // Add seed corpus
    f.Fuzz(func(t *testing.T, input []byte) {
        // Fuzz test logic
    })
}
```

## Test Organization

### File Naming
- Test files: `*_test.go`
- Next to source files: `package.go` and `package_test.go`

### Test Structure
```go
func TestFeature(t *testing.T) {
    t.Parallel() // If independent

    tests := []struct {
        name     string
        input    InputType
        expected ExpectedType
    }{
        {
            name:     "scenario 1",
            input:    input1,
            expected: expected1,
        },
        // ... more cases
    }

    for _, tt := range tests {
        tt := tt // Capture range variable
        t.Run(tt.name, func(t *testing.T) {
            result := FunctionUnderTest(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

## Parallel Testing

### When to Use `t.Parallel()`
- Tests that don't share state
- Tests that don't use global variables
- Tests that don't write to same files
- Tests that are independent of each other

### When NOT to Use `t.Parallel()`
- Tests using shared resources (database, files)
- Tests with dependencies on other tests
- Tests that modify global state
- Integration tests with shared infrastructure

## Benchmarks

### Best Practices
```go
func BenchmarkImportantFunction(b *testing.B) {
    b.ReportAllocs() // Always track memory

    // Setup before timer
    setupData := prepareTestData()
    b.ResetTimer()

    for b.Loop() {
        // Benchmark code
        ImportantFunction(setupData)
    }
}
```

### Running Benchmarks
```bash
just bench                           # Run all benchmarks
go test -bench=. -benchmem ./...    # With memory tracking
go test -bench=BenchmarkName ./...   # Specific benchmark
```

## Fuzzing

### Purpose
- Find edge cases and bugs
- Test with random inputs
- Improve code robustness
- Prevent regression

### Writing Fuzz Tests
```go
func FuzzSerialize(f *testing.F) {
    // Add seed corpus with typical inputs
    f.Add("package main\nfunc main() {}")
    f.Add("var x int = 5")
    f.Add("func test() {\n    return 42\n}")

    f.Fuzz(func(t *testing.T, input string) {
        node := Parse(input)
        tokens := Serialize(node)

        // Verify invariants
        if len(tokens) == 0 && len(input) > 0 {
            t.Errorf("Serialize returned empty tokens for non-empty input")
        }
    })
}
```

### Running Fuzz Tests
```bash
go test -fuzz=. -fuzztime=30s ./...      # Run all fuzz tests
go test -fuzz=FuzzName -fuzztime=60s ./...  # Specific fuzz test
```

## Property-Based Testing

### Purpose
- Test invariants and properties
- Verify code behavior across input space
- Complement example-based tests

### Using `testing/quick`
```go
func TestSortingProperty(t *testing.T) {
    f := func(x []int) bool {
        sorted := Sort(x)
        // Property: sorted list is in non-decreasing order
        for i := 1; i < len(sorted); i++ {
            if sorted[i] < sorted[i-1] {
                return false
            }
        }
        return true
    }

    if err := quick.Check(f, nil); err != nil {
        t.Fatal(err)
    }
}
```

## Test Commands (justfile)

```bash
just test                    # Run all tests with coverage
just test-race              # Run tests with race detector
just coverage               # Generate HTML coverage report
just bench                  # Run benchmarks with memory stats
just test-fuzz              # Run fuzz tests
just test-coverage          # Show coverage percentage
```

## Coverage

### Target: 80% minimum

### Checking Coverage
```bash
just test-coverage          # Show coverage percentage
just coverage               # Generate HTML report
```

### Coverage Quality Gates
```bash
go test -coverprofile=cover.out ./...
COVERAGE=$(go tool cover -func=cover.out | grep total | awk '{print $3}' | tr -d '%')
if [ $(echo "$COVERAGE < 80" | bc) -eq 1 ]; then
    echo "Coverage is $COVERAGE%, below 80% threshold"
    exit 1
fi
```

## Test Helpers

### Common Patterns
```go
// Helper for test setup
func setupTestFile(t *testing.T, content string) string {
    t.Helper()
    tmpFile, err := os.CreateTemp("", "test-*.go")
    if err != nil {
        t.Fatalf("Failed to create temp file: %v", err)
    }
    defer tmpFile.Close()

    if _, err := tmpFile.WriteString(content); err != nil {
        t.Fatalf("Failed to write to temp file: %v", err)
    }

    return tmpFile.Name()
}

// Helper for comparing slices
func compareSlices[T comparable](t *testing.T, got, want []T) {
    t.Helper()
    if len(got) != len(want) {
        t.Errorf("length mismatch: got %d, want %d", len(got), len(want))
        return
    }
    for i := range got {
        if got[i] != want[i] {
            t.Errorf("index %d: got %v, want %v", i, got[i], want[i])
        }
    }
}
```

## Common Pitfalls

### 1. Not Using `t.Helper()`
```go
// Bad
func assertEqual(t *testing.T, got, want any) {
    if got != want {
        t.Errorf("not equal") // Line number points to helper
    }
}

// Good
func assertEqual(t *testing.T, got, want any) {
    t.Helper()
    if got != want {
        t.Errorf("not equal") // Line number points to caller
    }
}
```

### 2. Race Conditions in Parallel Tests
```go
// Bad - shared variable
var counter int
for i := 0; i < 10; i++ {
    t.Run(fmt.Sprintf("test%d", i), func(t *testing.T) {
        t.Parallel()
        counter++ // RACE!
    })
}

// Good - local variable
for i := 0; i < 10; i++ {
    i := i // Capture loop variable
    t.Run(fmt.Sprintf("test%d", i), func(t *testing.T) {
        t.Parallel()
        counter := i // Local copy
    })
}
```

### 3. Not Resetting Timer in Benchmarks
```go
// Bad
func BenchmarkSlow(b *testing.B) {
    setup := slowSetup()
    for b.Loop() {
        // Setup time included in benchmark
    }
}

// Good
func BenchmarkFast(b *testing.B) {
    setup := slowSetup()
    b.ResetTimer() // Reset after setup
    for b.Loop() {
        // Only benchmark this
    }
}
```

## Integration Testing

Integration tests should:
- Use `t.Parallel()` carefully (avoid shared state)
- Have clear setup/teardown with `t.Cleanup()`
- Use realistic data and scenarios
- Be marked with integration in test names

```go
func TestIntegrationWorkflow(t *testing.T) {
    tmpDir := setupIntegrationEnvironment(t)
    defer os.RemoveAll(tmpDir)

    // Test complete workflow
    result := RunAnalysis(tmpDir)
    if result.Error != nil {
        t.Fatalf("Analysis failed: %v", result.Error)
    }

    // Verify end-to-end behavior
    if len(result.Clones) == 0 {
        t.Error("Expected to find clones")
    }
}
```

## Test-Driven Development (TDD)

### Workflow
1. Write a failing test
2. Run test and confirm it fails
3. Write minimal code to make test pass
4. Run test and confirm it passes
5. Refactor if needed

### Example
```go
// Step 1: Write failing test
func TestCalculateComplexity(t *testing.T) {
    result := CalculateComplexity("func test() { return 1 + 1 }")
    if result != 2 {
        t.Errorf("got %d, want 2", result)
    }
}

// Step 2: Run test - fails (function doesn't exist)

// Step 3: Implement
func CalculateComplexity(code string) int {
    return 2 // Simple implementation
}

// Step 4: Run test - passes

// Step 5: Refactor to real implementation
```

## Continuous Integration

### CI Test Requirements
- All tests must pass
- Coverage must be ≥ 80%
- No race conditions (`-race`)
- No linter warnings

### CI Test Matrix
- Multiple Go versions
- Multiple OS platforms
- With and without race detector

## Resources

### Official Documentation
- [Go Testing Package](https://pkg.go.dev/testing)
- [Go Fuzzing Tutorial](https://go.dev/doc/tutorial/fuzz)
- [Property-Based Testing](https://pkg.go.dev/testing/quick)
- [Table-Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)

### Best Practices
- Write tests before or with code (TDD)
- Keep tests simple and focused
- Use descriptive test names
- Test edge cases and error conditions
- Mock external dependencies
- Keep test data minimal and clear

## Migration from External Frameworks

### From Ginkgo/Gomega to Native
```go
// Old (Ginkgo)
var _ = Describe("Feature", func() {
    It("should do something", func() {
        Expect(result).To(Equal(expected))
    })
})

// New (Native)
func TestFeature(t *testing.T) {
    t.Run("should do something", func(t *testing.T) {
        if result != expected {
            t.Errorf("got %v, want %v", result, expected)
        }
    })
}
```

### From Testify to Native
```go
// Old (testify)
assert.Equal(t, expected, result)
assert.NoError(t, err)

// New (Native)
if result != expected {
    t.Errorf("got %v, want %v", result, expected)
}
if err != nil {
    t.Errorf("unexpected error: %v", err)
}
```

## Summary

- Use native Go testing tools exclusively
- Write table-driven tests for multiple scenarios
- Enable parallel execution for independent tests
- Track memory allocations in benchmarks
- Add fuzz tests for critical algorithms
- Use property-based tests for invariants
- Maintain 80%+ code coverage
- Keep tests fast, focused, and clear
