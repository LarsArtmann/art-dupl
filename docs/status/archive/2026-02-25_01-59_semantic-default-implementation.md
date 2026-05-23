# Status Report: Semantic Detection Default Implementation

**Date**: 2026-02-25 01:59 CET
**Status**: 🟡 IN PROGRESS - Critical bug identified, fix pending
**Author**: Crush (AI Assistant)
**Related**: Issue to make `--semantic` the default behavior

---

## Executive Summary

Implementing the change to make `--semantic` detection the DEFAULT behavior for art-dupl. The majority of the implementation is complete, but **a critical bug was discovered** in `cmd/run_flags.go` that prevents the feature from working correctly.

**Blocking Issue**: Missing post-merge structural flag handling causes BDD tests to fail.

---

## What Was Requested

> `--semantic` MUST be the default for "art-dupl --html -t 50"

The user wanted semantic-aware duplicate detection to be ON by default, reducing false positives from structurally similar but semantically different code.

---

## Implementation Status

### ✅ Completed Tasks

| Task                                      | File                             | Status  |
| ----------------------------------------- | -------------------------------- | ------- |
| Change default `Semantic: false` → `true` | `config/config.go:161`           | ✅ Done |
| Update config struct documentation        | `config/config.go:118-127`       | ✅ Done |
| Update merge config comment               | `config/config_merge.go:133`     | ✅ Done |
| Update flag description (root)            | `cmd/flags.go:42-47`             | ✅ Done |
| Update flag description (stats)           | `cmd/stats.go:66-68`             | ✅ Done |
| Update conflict error message             | `cmd/run_flags.go:57-60`         | ✅ Done |
| Update conflict error message             | `cmd/stats.go:95-98`             | ✅ Done |
| Update structural warning message         | `cmd/run_flags.go:62-65`         | ✅ Done |
| Update structural warning message         | `cmd/stats.go:100-103`           | ✅ Done |
| Add post-merge structural handling        | `cmd/stats.go:183-186`           | ✅ Done |
| Update AGENTS.md documentation            | `AGENTS.md:347-355`              | ✅ Done |
| Update BDD test descriptions              | `bdd/semantic_detection_test.go` | ✅ Done |

### ❌ Critical Bug Discovered

**File**: `cmd/run_flags.go`
**Issue**: Missing post-merge structural flag handling

The code currently does:

```go
mergedConfig := config.MergeConfigs(fileConfig, appConfig)

// Wire semantic detection to golang package global
golang.SemanticHashEnabled = mergedConfig.Semantic
```

**MISSING** (that exists in stats.go but not here):

```go
// Handle --structural flag to explicitly disable semantic detection (opt-out from default)
if structural {
    mergedConfig.Semantic = false
}
```

### 🧪 Test Results

**BDD Tests**: 2 FAILURES out of 222 tests

| Test                                                                             | File:Line                            | Reason                         |
| -------------------------------------------------------------------------------- | ------------------------------------ | ------------------------------ |
| `should distinguish between different handler tests with semantic detection`     | `bdd/semantic_detection_test.go:285` | Semantic detection not working |
| `should NOT flag methods on different types as duplicates by default (semantic)` | `bdd/semantic_detection_test.go:344` | Semantic detection not working |

**Root Cause**: The missing post-merge handling means `--structural` opt-out doesn't work, AND potentially the default `true` isn't being properly applied.

---

## Files Modified

### Core Configuration

1. **`config/config.go`**
   - Line 127: `Semantic: true` (was `false`)
   - Lines 118-127: Updated documentation

2. **`config/config_merge.go`**
   - Line 133: Updated comment

### CLI Commands

3. **`cmd/flags.go`**
   - Lines 42-47: Updated flag descriptions

4. **`cmd/stats.go`**
   - Lines 66-68: Updated flag descriptions
   - Lines 95-98: Updated error message
   - Lines 100-103: Updated warning message
   - Lines 175: Removed pre-merge structural handling
   - Lines 183-186: Added post-merge structural handling

5. **`cmd/run_flags.go`**
   - Lines 57-65: Updated error and warning messages
   - Line 155-160: Removed pre-merge structural handling
   - **BUG**: Missing post-merge structural handling!

### Documentation

6. **`AGENTS.md`**
   - Lines 347-355: Updated semantic detection documentation

### Tests

7. **`bdd/semantic_detection_test.go`**
   - Updated test descriptions and expectations
   - Changed from testing `--semantic` vs default to testing `--structural` vs default

---

## Architecture Overview

### How Semantic Detection Works

1. **Configuration Layer**:
   - `config.Config.Semantic` (bool) - default is now `true`
   - CLI flags: `--semantic` (explicit enable), `--structural` (opt-out)

2. **Merge Layer**:
   - `config.MergeConfigs()` combines file config + CLI config
   - Default from `DefaultConfig()` is `Semantic: true`
   - File config can override
   - CLI flags override everything

3. **Runtime Layer**:
   - After merge, `golang.SemanticHashEnabled` global is set
   - This global is checked during AST transformation
   - When enabled, identifier names are hashed into AST node types

4. **Detection Layer**:
   - `syntax/golang/identifier_hash.go` contains `SemanticHashEnabled`
   - `encodeSemanticType()` combines base type + identifier hash
   - Different identifiers → different hashes → no match

### The Bug Explained

```
┌─────────────────────────────────────────────────────────────┐
│                     CURRENT FLOW (BUGGY)                    │
├─────────────────────────────────────────────────────────────┤
│  1. Parse flags: structural=false, semantic=false (default) │
│  2. Build appConfig: Semantic NOT SET (zero value)          │
│  3. Merge configs: DefaultConfig.Semantic=true wins         │
│  4. ❌ MISSING: Check structural flag after merge           │
│  5. Set golang.SemanticHashEnabled = mergedConfig.Semantic  │
│                                                             │
│  PROBLEM: If user passes --structural, it's ignored!        │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                     CORRECT FLOW                            │
├─────────────────────────────────────────────────────────────┤
│  1. Parse flags: structural=true/false, semantic=true/false │
│  2. Build appConfig: Semantic NOT SET (zero value)          │
│  3. Merge configs: DefaultConfig.Semantic=true wins         │
│  4. ✅ AFTER MERGE: if structural { Semantic = false }      │
│  5. Set golang.SemanticHashEnabled = mergedConfig.Semantic  │
└─────────────────────────────────────────────────────────────┘
```

---

## Next Steps

### Immediate (Critical)

1. **Fix `cmd/run_flags.go`**: Add post-merge structural handling

   ```go
   // Handle --structural flag to explicitly disable semantic detection (opt-out from default)
   if structural {
       mergedConfig.Semantic = false
   }
   ```

2. **Rebuild**: `just build`

3. **Run BDD tests**: `go test -v ./bdd/...`

4. **Verify manually**:

   ```bash
   # Should NOT detect (semantic on by default)
   ./dist/art-dupl ./bdd/testdata --threshold 15

   # Should detect (structural opt-out)
   ./dist/art-dupl ./bdd/testdata --threshold 15 --structural
   ```

### Short-term

5. Run full test suite: `just test`
6. Run linting: `just check`
7. Run CI checks: `just ci`

### Documentation

8. Update `README.md` with new default behavior
9. Update `HOW_TO_USE.md` with examples
10. Check for other documentation files mentioning semantic

---

## Questions / Unknowns

### Primary Question

**Why might semantic detection still not work even after the fix?**

The semantic hashing logic in `syntax/golang/identifier_hash.go` needs to be verified:

1. Is `SemanticHashEnabled` checked at the right time?
2. Are all relevant AST node types being encoded with semantic hashes?
3. Is the hash computation correct for all identifier types?

### Investigation Needed

If tests still fail after the fix:

1. Add debug logging to show `SemanticHashEnabled` value at runtime
2. Add debug logging to show which identifiers are being hashed
3. Compare AST node types with/without semantic enabled

---

## Related Files Reference

| Category       | Files                                                                     |
| -------------- | ------------------------------------------------------------------------- |
| Configuration  | `config/config.go`, `config/config_merge.go`                              |
| CLI            | `cmd/flags.go`, `cmd/run_flags.go`, `cmd/stats.go`                        |
| Semantic Logic | `syntax/golang/identifier_hash.go`, `syntax/golang/transform.go`          |
| Tests          | `bdd/semantic_detection_test.go`, `syntax/golang/identifier_hash_test.go` |
| Documentation  | `AGENTS.md`, `README.md`, `HOW_TO_USE.md`                                 |

---

## Metrics

| Metric         | Value           |
| -------------- | --------------- |
| Files Modified | 7               |
| Lines Changed  | ~50             |
| Tests Passing  | 220/222 (99.1%) |
| Tests Failing  | 2/222 (0.9%)    |
| Critical Bugs  | 1               |
| Time Spent     | ~30 minutes     |

---

## Conclusion

The implementation is 90% complete with one critical bug blocking completion. The fix is straightforward (add 4 lines to `cmd/run_flags.go`), after which verification and documentation updates can proceed.

**Estimated time to completion**: 10-15 minutes after fix is applied.
