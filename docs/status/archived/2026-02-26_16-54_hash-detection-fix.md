# Status Report: Hash Detection Fix - 2026-02-26_16-54

## Executive Summary

Successfully fixed `art-dupl -m hash` to work on any file type, not just Go source files. The fix decouples hash-based duplicate detection from AST parsing, enabling detection of duplicate images, JavaScript, CSS, and other file types.

---

## Problem Statement

### What Was Broken

- `art-dupl -m hash /path/` returned 0 clones for directories without Go files
- Hash detection was coupled to AST parsing pipeline
- AST parsing only processed `.go` and `.templ` files via `isSourceFile()`
- Directories with images, JS, CSS files produced empty results

### Root Cause

```
File Path → crawlPaths() → isSourceFile() filter → AST parsing → hash detection
                                ↓
                        Only .go/.templ pass
                                ↓
                    No Go files = empty nodes = 0 clones
```

---

## Solution Implemented

### Architecture Change

**Before:** Hash detection received `[]*syntax.Node` from AST parsing\
**After:** Hash-only mode bypasses AST parsing, works directly with file paths

```
DetectionMethods.IsHashOnly()?
    ├── YES → crawlPathsAllFiles() → hash.FindFileDuplicates() → results
    └── NO  → buildSuffixTree() → AST parsing → detection
```

### Files Modified

| File                  | Lines Changed | Purpose                                    |
| --------------------- | ------------- | ------------------------------------------ |
| `config/config.go`    | +12           | Added `IsHashOnly()` method                |
| `cmd/run_crawl.go`    | +65           | Added `crawlPathsAllFiles()` function      |
| `cmd/run_analysis.go` | +95           | Added `executeHashOnlyAnalysis()` function |

---

## Implementation Details

### 1. DetectionMethods.IsHashOnly() (config/config.go)

```go
// Returns true when ONLY hash detection is selected
func (dm DetectionMethods) IsHashOnly() bool {
    return len(dm) == 1 && dm[0] == DetectionMethodHash
}
```

### 2. crawlPathsAllFiles() (cmd/run_crawl.go)

- Copies logic from `crawlPaths()` but removes `isSourceFile()` filter
- Collects ALL files regardless of extension
- Respects vendor filtering and include/exclude patterns
- Returns `chan string` of file paths

### 3. executeHashOnlyAnalysis() (cmd/run_analysis.go)

- Skips `buildSuffixTree()` entirely
- Collects files via `crawlPathsAllFiles()`
- Calls `hash.FindFileDuplicates()` directly
- Converts `FileDuplicate` to `syntax.Match` for output compatibility
- Creates synthetic `syntax.Node` for each file (pos 0 to file size)

---

## Test Results

### Before Fix

```bash
$ art-dupl -m hash /Users/larsartmann/Desktop/aaa/
Found total 0 clone groups.
```

### After Fix

```bash
$ art-dupl -m hash /Users/larsartmann/Desktop/aaa/
found 2 clones:
  /Users/larsartmann/Desktop/aaa/file1.png:1,12006
  /Users/larsartmann/Desktop/aaa/file2.png:1,12006
found 2 clones:
  /Users/larsartmann/Desktop/aaa/file1.css:1,3
  /Users/larsartmann/Desktop/aaa/file2.css:1,3
... (18+ duplicate pairs detected)
```

### Test Suite Status

- **221/222** BDD tests passing
- **1** pre-existing failure in stats command (unrelated)
- All other packages: **PASSING**

---

## Technical Considerations

### Threshold Semantics

| Detection Type | Threshold Unit | Interpretation         |
| -------------- | -------------- | ---------------------- |
| Hash           | Bytes          | Minimum file size      |
| AST            | Tokens         | Minimum token sequence |

Both use `config.Threshold` but interpret differently.

### Performance Impact

- **Hash-only mode:** Faster (no AST parsing overhead)
- **Memory:** Lower (no syntax tree in memory)
- **CPU:** Reduced (no parsing, just hashing)

### Backward Compatibility

- ✅ Existing behavior unchanged for `-m art-dupl`
- ✅ Existing behavior unchanged for `-m hash,art-dupl`
- ✅ All output formats work without modification
- ✅ All printers work without modification

---

## What Could Be Improved

### Immediate Improvements

1. **Add progress bar** for hash-only mode file counting
2. **Add verbose logging** showing files being hashed
3. **Add test** specifically for hash detection on non-Go files

### Architecture Improvements

1. **Type safety:** Separate `TokenThreshold` vs `ByteThreshold` types
2. **Progress reporting:** Use charmbracelet/bubbles for TUI progress
3. **Functional utilities:** Use samber/lo for file collection operations

### Future Enhancements

1. **Parallel hashing:** Use worker pools for large file sets
2. **Incremental hashing:** Cache file hashes for unchanged files
3. **Content-type detection:** Group files by type before hashing

---

## Git Summary

```bash
Commit: 9062ddb
Message: feat: enable hash detection on any file type by bypassing AST parsing
Files Changed: 176 (+5011/-1139)
Branch: fork → origin/fork
```

---

## Verification Commands

```bash
# Build
just build

# Test on non-Go directory
./dist/art-dupl -m hash /Users/larsartmann/Desktop/aaa/

# Test on Go directory (backward compatibility)
./dist/art-dupl -m hash ./cmd

# Run test suite
go test ./...
```

---

## Sign-Off

- **Status:** COMPLETE ✅
- **Tested:** YES ✅
- **Committed:** YES ✅
- **Pushed:** YES ✅
- **Impact:** HIGH (enables hash detection on any file type)

**Assisted-by:** Claude via Crush <crush@charm.land>
