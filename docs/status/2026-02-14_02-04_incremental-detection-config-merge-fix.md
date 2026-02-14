# Status Report: Incremental Detection Config Merge Bug Fix

**Date:** 2026-02-14 02:04
**Status:** In Progress - 4 BDD tests remaining
**Progress:** 212/216 tests passing (was 208/216)

## Summary

Fixed critical bug in `config/config_merge.go` where 4 incremental mode fields were not being merged from CLI flags to the final configuration. Cache directory is now correctly created and populated.

## Root Cause

The `mergeConfig()` function was missing merge cases for:
- `Incremental` (bool)
- `Since` (string)
- `CacheDir` (string)
- `ClearCache` (bool)

CLI flags were read correctly but dropped during the merge step.

## Fix Applied

Added 4 merge cases to `config/config_merge.go` after line 109:

```go
// Incremental (bool)
if !skipZeroValues || cfg.Incremental {
    result.Incremental = cfg.Incremental
}

// Since (string)
if !skipZeroValues || cfg.Since != "" {
    result.Since = cfg.Since
}

// CacheDir (string)
if !skipZeroValues || cfg.CacheDir != "" {
    result.CacheDir = cfg.CacheDir
}

// ClearCache (bool)
if !skipZeroValues || cfg.ClearCache {
    result.ClearCache = cfg.ClearCache
}
```

## Verification

- Cache directory IS created: `/tmp/test-cache/files/` contains `.gob` files
- Debug output now shows: `incremental=true, cacheDir="/tmp/test-cache"`

## Remaining Work

4 BDD tests still failing in `bdd/incremental_detection_test.go`:
1. Line 91: "should detect duplicates correctly on first run"
2. Line 269: "should clear cache before running"
3. Line 285: "should work with plumbing output format"
4. Line 313: "should work independently of cache state"

These appear to be output format assertion issues - tests expect substring "found" but output is "found 2 clones:".

## Files Modified

| File | Change |
|------|--------|
| `config/config_merge.go:109-128` | Added 4 missing merge cases |
| `docs/status/2026-02-14_01-57_*` | Previous status report |

## Next Steps

1. Investigate remaining 4 test failures
2. Adjust test assertions or output format as needed
3. Commit fix once all tests pass
4. Optional: Remove debug logging from `cmd/run_analysis.go` and `job/incremental.go`

## Key Code Locations

- **Flag definitions**: `cmd/flags.go:30-34`
- **Flag reading**: `cmd/run_flags.go:44-48`
- **Config merge (FIXED)**: `config/config_merge.go:109-128`
- **Incremental mode check**: `cmd/run_analysis.go:35`
- **IncrementalParser**: `job/incremental.go`
- **FileCache**: `cache/file_cache.go`
- **BDD tests**: `bdd/incremental_detection_test.go`

## Lesson Learned

When adding new config fields, ALWAYS add corresponding merge cases to `mergeConfig()`:

```go
// Bool fields
if !skipZeroValues || cfg.FieldName {
    result.FieldName = cfg.FieldName
}

// String fields
if !skipZeroValues || cfg.FieldName != "" {
    result.FieldName = cfg.FieldName
}
```
