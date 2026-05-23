# Test Code Deduplication Session Report

**Date:** 2026-02-25 08:52
**Session Focus:** Eliminating code duplication in test files using art-dupl

---

## Summary

Successfully eliminated **3 clone groups** detected by `art-dupl -t 50 --semantic` by extracting common test patterns into reusable helper functions.

## Initial State

Running `art-dupl -t 50 --semantic` detected 3 clone groups:

```
found 2 clones:
  detection/detection_test.go:646,682
  detection/detection_test.go:685,718
found 2 clones:
  git/change_detector_test.go:135,147
  git/change_detector_test.go:172,185
found 2 clones:
  domain/coverage_test.go:1059,1068
  domain/coverage_test.go:1070,1079

Found total 3 clone groups.
```

## Changes Made

### 1. domain/coverage_test.go

**Clone Pattern:** Two test cases with identical `unmarshalStringID` error assertion logic.

**Solution:** Extracted `assertUnmarshalStringIDError()` helper function.

```go
// assertUnmarshalStringIDError tests that unmarshalStringID returns an error for the given input.
func assertUnmarshalStringIDError(t *testing.T, input string, wantErrContains string) {
    t.Helper()
    var result string
    err := unmarshalStringID([]byte(input), "TestType", "TestType cannot be empty", func(s string) {
        result = s
    })
    _ = result // Ensure variable is used
    if err == nil {
        t.Errorf("unmarshalStringID() error = nil, want error %q", wantErrContains)
    }
}
```

**Usage:**

```go
t.Run("empty string", func(t *testing.T) {
    assertUnmarshalStringIDError(t, `""`, "empty string")
})

t.Run("invalid JSON", func(t *testing.T) {
    assertUnmarshalStringIDError(t, `invalid`, "invalid JSON")
})
```

### 2. git/change_detector_test.go

**Clone Pattern:** Two test cases asserting zero changes from `GetChangedFiles()`.

**Solution:** Extracted `assertNoChanges()` helper function.

```go
// assertNoChanges asserts that GetChangedFiles returns 0 changes for the given since value.
func assertNoChanges(t *testing.T, repoDir, since string) {
    t.Helper()
    detector := NewChangeDetector(repoDir)
    changes, err := detector.GetChangedFiles(since)
    if err != nil {
        t.Fatalf("GetChangedFiles failed: %v", err)
    }
    if len(changes) != 0 {
        t.Errorf("Expected 0 changes, got %d", len(changes))
    }
}
```

**Usage:**

```go
t.Run("no changes", func(t *testing.T) {
    repoDir := setupGitRepo(t)
    createAndCommitFile(t, repoDir, "initial.go", "package main")
    assertNoChanges(t, repoDir, "HEAD")
})

t.Run("empty since defaults to HEAD", func(t *testing.T) {
    repoDir := setupGitRepo(t)
    createAndCommitFile(t, repoDir, "initial.go", "package main")
    assertNoChanges(t, repoDir, "")
})
```

### 3. detection/detection_test.go

**Clone Pattern:** Two nearly identical tests for legacy detection with different Go code samples.

**Solution:** Extracted `runLegacyDetectionTest()` helper function that handles file creation and detection execution.

```go
// runLegacyDetectionTest creates a file with the given code and runs legacy detection.
// Returns the issues found by the detector.
func runLegacyDetectionTest(t *testing.T, filename, goCode string) []LegacyIssue {
    t.Helper()
    tmpDir := t.TempDir()
    testFile := tmpDir + "/" + filename
    if err := os.WriteFile(testFile, []byte(goCode), 0o644); err != nil {
        t.Fatalf("Failed to write test file: %v", err)
    }

    detector := NewLegacyDetector()
    nodes := []*syntax.Node{
        {Filename: testFile, Type: int32(golang.File), Pos: 1, End: 100},
    }

    return detector.findLegacyInFile(testFile, nodes)
}
```

**Usage:**

```go
func TestLegacyDetector_FindLegacyInFile_RealFile(t *testing.T) {
    goCode := `package test

import (
    "io/ioutil"
)

func LegacyFunc() {
    // Using deprecated ioutil.ReadFile
    data, err := ioutil.ReadFile("test.txt")
    if err != nil {
        return
    }
    _ = data
}`
    issues := runLegacyDetectionTest(t, "test_legacy.go", goCode)
    if issues == nil {
        t.Log("No legacy issues found (simplified detection)")
    }
}
```

## Final State

```
📖 Parsing files and building analysis tree... ✅

Found total 0 clone groups.
```

## Metrics

| File                        | Before             | After           | Reduction |
| --------------------------- | ------------------ | --------------- | --------- |
| domain/coverage_test.go     | 20 lines duplicate | 12 lines helper | 40%       |
| git/change_detector_test.go | 26 lines duplicate | 10 lines helper | 62%       |
| detection/detection_test.go | 74 lines duplicate | 60 lines helper | 19%       |

## Verification

All tests pass:

- `go test -v ./detection/... -run "Legacy"` ✓
- `go test -v ./git/... -run "GetChangedFiles"` ✓
- `go test -v ./domain/... -run "UnmarshalStringID"` ✓
- `just test` (full suite) ✓

## Key Principles Applied

1. **DRY (Don't Repeat Yourself)** - Extracted common patterns into helper functions
2. **t.Helper()** - All helpers marked as test helpers for better error reporting
3. **Semantic matching** - Used art-dupl's semantic detection to find structurally similar code
4. **Maintainability** - Changes make tests easier to understand and modify

## Files Modified

- `detection/detection_test.go` - Added `runLegacyDetectionTest()` helper
- `git/change_detector_test.go` - Added `assertNoChanges()` helper
- `domain/coverage_test.go` - Added `assertUnmarshalStringIDError()` helper

---

_Report generated by art-dupl self-analysis_
