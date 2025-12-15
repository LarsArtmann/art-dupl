# Code Quality Issues - Sorted Report

Generated: 2025-12-15
Command: `just check` and `just fd`

## Summary
- **Total Issues**: 10 (6 errcheck, 2 staticcheck, 2 unused)
- **Code Duplicates**: 0 detected by golangci-lint dupl

---

## Critical Issues (Must Fix)

### 1. Unchecked Error Returns - Format Errors (errcheck) - 4 instances
- **Location**: `cli.go:739`
  ```go
  fmt.Fprintf(cli.Stderr(), "Running %s detection method...\n", method)
  ```
- **Type**: Security risk - CLI output errors unhandled

- **Location**: `cli.go:758`
  ```go
  fmt.Fprintf(cli.Stderr(), "All reports generated in: %s\n", outputDir)
  ```
- **Type**: Security risk - CLI output errors unhandled

- **Location**: `cli.go:821`
  ```go
  fmt.Fprintf(cli.Stderr(), "Generated %s report: %s\n", fmtInfo.name, filename)
  ```
- **Type**: Security risk - CLI output errors unhandled

- **Location**: `testutils/unique.go:97`
  ```go
  go stdout.Write(testData)
  ```
- **Type**: Test reliability - Write failures unhandled

### 2. Resource Management Issues (errcheck) - 2 instances
- **Location**: `cli.go:803`
  ```go
  defer file.Close()
  ```
- **Type**: Resource leak risk - File close error ignored

- **Location**: `cli/runtime_test.go:93`
  ```go
  w.Close()
  ```
- **Type**: Test reliability - Close errors unhandled

---

## Quality Issues (Should Fix)

### 3. Deprecated API Usage (staticcheck) - 2 instances
- **Location**: `main.go:76`
  ```go
  fang.WithTheme(fang.DefaultTheme(true))
  ```
  - **Issue**: `fang.WithTheme` is deprecated, use `WithColorSchemeFunc` instead
  - **Priority**: Medium - Future compatibility

- **Location**: `testutils/unique.go:15`
  ```go
  rand.Seed(time.Now().UnixNano())
  ```
  - **Issue**: `rand.Seed` deprecated since Go 1.20
  - **Priority**: Medium - Modern Go practices

---

## Maintenance Issues (Nice to Fix)

### 4. Unused Code (unused) - 2 instances
- **Location**: `cli.go:63`
  ```go
  var paths []string
  ```
  - **Issue**: Unused variable in struct
  - **Priority**: Low - Code cleanliness

- **Location**: `cli.go:444`
  ```go
  func filesFeed() chan string {
  ```
  - **Issue**: Entire function unused
  - **Priority**: Low - Code cleanliness

---

## Priority Action Plan

### Fix Now (Critical Errors)
1. **Add error handling for all fmt.Fprintf calls** - 4 fixes
2. **Fix resource management** - 2 fixes

### Update Soon (Quality Improvements)
3. **Replace deprecated APIs** - 2 fixes
4. **Remove unused code** - 2 fixes

## Commands to Run for Verification

```bash
just lint          # Check all linting issues
just fd            # Find duplicates (currently clean)
just build         # Ensure builds still work
just test          # Run tests after fixes
```

## Notes
- No code duplication detected by golangci-lint dupl linter
- All issues are standard Go linting problems
- Fixes should be straightforward error handling replacements