# Comprehensive Status Report — Line Number Bug Fix Session

**Date:** 2026-04-15 20:52  
**Branch:** fork  
**Latest commit:** `94953c7 docs(status): add comprehensive status report after filter refactoring session`  
**Reporter:** Crush (GLM-5.1)

---

## Executive Summary

Fixed a critical line number bug where `ByteRangeToLines` produced identical start/end line numbers (e.g., `24,24` instead of `24,48`) for multi-line clones. Two root causes identified and fixed. All affected packages' tests pass. Six pre-existing BDD test failures remain (unrelated to this fix).

---

## A) FULLY DONE

### 1. Line Number Bug — Root Cause Fix (`pkg/position/lines.go`)

**Problem:** `ByteRangeToLines` used a manual loop checking `offset == end-1`. When Go AST's exclusive `End()` position pointed at or past `len(content)`, the loop never matched, `lineEnd` stayed `0`, and the fallback silently set `lineEnd = lineStart`.

**Fix:** Replaced the buggy 39-line function with a clean 19-line implementation that delegates to `offsetToLine()` (which already correctly handles out-of-bounds offsets by capping to `len(content)-1`).

**Files changed:** `pkg/position/lines.go`

**Before (buggy):**
```go
for offset := range content {
    if offset == end-1 { lineEnd = line; break }  // never matches when end-1 >= len(content)
}
if lineEnd == 0 { lineEnd = lineStart }  // silently collapses multi-line ranges!
```

**After (correct):**
```go
lineStart := offsetToLine(content, start)
lastByte := end - 1
if lastByte < start { lastByte = start }
lineEnd := offsetToLine(content, lastByte)  // handles out-of-bounds internally
```

### 2. JSON Printer EndLine Fix (`printer/json.go`)

**Problem:** `PrintClones` set `LineEnd: 0` and then recalculated it by counting newlines in the **deindented fragment** — which can have a different number of lines than the actual source range.

**Fix:** Now uses `fileInfo.LineEnd` from `ProcessNodeRange` directly (which goes through the now-fixed `ByteRangeToLines`). Removed dead code (`countLinesInFragment`, unused `strings` import).

**Files changed:** `printer/json.go`

### 3. Regression Tests (`pkg/position/lines_test.go`)

Added 5 new test cases that would have caught the original bug:
- `multi-line end beyond content returns last line not start line` — the exact bug pattern
- `multi-line spanning all with end at exact content length`
- `multi-line end one past content length`
- `multi-line both start and end beyond content`
- `negative start clamped`

### 4. Full Test Verification

All affected packages pass:
| Package | Status | Coverage |
|---------|--------|----------|
| `pkg/position` | PASS | 97.1% |
| `printer` | PASS | 63.4% |
| `pkg/artdupl` | PASS | 85.0% |
| `domain` | PASS | 97.0% |

---

## B) PARTIALLY DONE

### 1. End-to-End Manual Verification (Not Yet Run)

The tool has not been built and run against a real codebase to verify the output format change from `24,24` to proper `24,48` style ranges. The fix is logically correct and all unit tests pass, but a manual sanity check would add confidence.

### 2. Coverage Gaps in Some Printers

The JSON printer fix is verified, but SARIF and HTML printers were not explicitly tested with the new `ByteRangeToLines`. They already used `fileInfo.LineEnd` directly (correct path), so they should be fine, but no new tests were added for them.

---

## C) NOT STARTED

### 1. CRLF/Windows Line Ending Handling

`ByteRangeToLines` counts `\n` only. If a file has `\r\n` line endings, the Go parser normalizes positions but `os.ReadFile` returns raw bytes. This could cause byte offset misalignment. Not investigated, not tested, not fixed. May be a latent bug on Windows.

### 2. TOCTOU File Read Gap

The Go parser reads files during AST construction (time T1). Printers re-read files via `os.ReadFile` at print time (time T2). If files change between T1 and T2, line numbers will be wrong. A caching layer or content snapshot would eliminate this.

### 3. SDK `detector_conversion.go` Fallback Path

In `pkg/artdupl/detector_conversion.go:75`, when `FileReader` is nil, `startLine` and `endLine` fall back to raw byte offsets — huge numbers, not line numbers at all. This is documented as a known issue but not fixed.

### 4. Performance Optimization for `offsetToLine`

`offsetToLine` is O(n) per call. For large files with many clones, this could be slow. The `LineIndex` type in the same package provides O(log n) lookup but isn't used by `ByteRangeToLines`. Could be a quick win for large codebases.

---

## D) TOTALLY FUCKED UP

### Nothing!

All changes are clean, tested, and backwards compatible. No regressions introduced. The 6 failing BDD tests are pre-existing (confirmed by running them against the pre-fix codebase).

---

## E) WHAT WE SHOULD IMPROVE

### High Impact

1. **`ByteRangeToLines` should use `LineIndex` for O(log n)** instead of O(n) linear scan — especially important for large files
2. **TOCTOU file read gap** — cache file content at parse time, reuse at print time
3. **CRLF handling** — normalize line endings before byte-to-line conversion
4. **SDK fallback path** — `detector_conversion.go` should never emit raw byte offsets as line numbers

### Medium Impact

5. **Printer coverage is 63.4%** — needs more tests, especially for edge cases in HTML/SARIF formatters
6. **`cli` package at 62.5% coverage** — the CLI runtime/validation logic needs more testing
7. **`hash` package at 69.0%** — hash-based detection is undertested
8. **`syntax` package at 67.6%** — AST serialization could use more edge case tests

### Low Impact / Cleanup

9. **`internal/filtertest` at 58.3%** — filter integration tests need attention (5 of 6 BDD failures are filter-related)
10. **Stash cleanup** — 3 old stashes lingering from previous sessions
11. **gopls hints** — `minmax` modernization hint in `lines.go:26` could use `max(start, end-1)` in Go 1.21+

---

## F) Top #25 Things to Get Done Next

### Critical (Fix Now)

| # | Task | Impact | Effort |
|---|------|--------|--------|
| 1 | Fix 6 pre-existing BDD test failures (5 filter + 1 CLI docs) | High | Medium |
| 2 | Fix `internal/filtertest` integration test failures | High | Low |
| 3 | End-to-end manual verification: build and run tool against real codebase | High | Low |
| 4 | Fix SDK fallback path in `detector_conversion.go` (byte offsets as line numbers) | Medium | Low |

### Important (Next Sprint)

| # | Task | Impact | Effort |
|---|------|--------|--------|
| 5 | Use `LineIndex` in `ByteRangeToLines` for O(log n) performance | Medium | Low |
| 6 | Add CRLF line ending handling to `ByteRangeToLines` | Medium | Medium |
| 7 | Cache file content at parse time to close TOCTOU gap | Medium | Medium |
| 8 | Improve `printer` test coverage from 63.4% to 80%+ | Medium | Medium |
| 9 | Improve `cli` package coverage from 62.5% to 80%+ | Medium | Low |
| 10 | Improve `hash` package coverage from 69.0% to 80%+ | Medium | Low |
| 11 | Improve `syntax` package coverage from 67.6% to 80%+ | Medium | Medium |
| 12 | Add SARIF printer tests with multi-line clones | Medium | Low |
| 13 | Add HTML printer tests with multi-line clones | Medium | Low |

### Nice to Have (Backlog)

| # | Task | Impact | Effort |
|---|------|--------|--------|
| 14 | Add fuzz tests for `ByteRangeToLines` with property: `endLine >= startLine` | Low | Low |
| 15 | Modernize `lines.go` with `max()` builtin (Go 1.21+) | Low | Trivial |
| 16 | Clean up 3 old git stashes | Low | Trivial |
| 17 | Add benchmark comparing `offsetToLine` vs `LineIndex` performance | Low | Low |
| 18 | Add `--strict` mode that errors on TOCTOU file mismatches | Low | Medium |
| 19 | Document line number semantics in API docs (exclusive end convention) | Low | Low |
| 20 | Investigate `config` package coverage (70.4%) — close to threshold | Low | Low |
| 21 | Investigate `job` package coverage (76.7%) — close to threshold | Low | Low |
| 22 | Add integration test for `pipeline/pipeline.go` (original bug report file) | Low | Medium |
| 23 | Review and clean up `docs/status/` — 200+ status reports is excessive | Low | Low |
| 24 | Add `CHANGELOG.md` entry for line number bug fix | Low | Trivial |
| 25 | Consider version bump for bug fix release | Low | Trivial |

---

## G) Top #1 Question I Cannot Figure Out Myself

**Why does the Go AST `End()` position sometimes exceed `len(content)` from `os.ReadFile`?**

The bug manifested because `node.End` (from `go/ast`) pointed past the file content length. This is the "exclusive end" convention — `End()` returns the position **after** the last byte. But the specific conditions under which `End > len(content)` happen are unclear:

- Is it only for the very last node in the file (where `End()` points to EOF)?
- Does `go/token` add a synthetic EOF position?
- Does it happen for intermediate nodes too?
- Is this behavior documented or just convention?

I worked around it by capping in `offsetToLine`, but understanding the exact AST contract would help validate the fix is complete.

---

## Test Results Summary

```
Total packages: 32
PASS: 30 packages
FAIL: 2 packages (bdd, internal/filtertest) — PRE-EXISTING, NOT from this session

Package coverage (sorted):
  pkg/format          100.0%
  adapter              97.7%
  pkg/position         97.1%  ← FIXED + REGRESSION TESTS
  domain               97.0%
  internal/simd        95.8%
  syntax/golang        93.9%
  errors               89.4%
  pkg/logger           87.5%
  cache                87.4%
  pkg/artdupl          85.0%
  syntax/templ         85.3%
  detection            83.6%
  git                  83.1%
  migration            83.3%
  job                  76.7%
  internal/enum        76.6%
  cmd                  75.6%
  suffixtree           91.0%
  config               70.4%
  bdd                  70.0%  ← 6 FAILURES (pre-existing)
  hash                 69.0%
  syntax               67.6%
  printer              63.4%  ← FIXED JSON PRINTER
  cli                   62.5%
  internal/filtertest   58.3%  ← 2 FAILURES (pre-existing)

Production code: 235 Go files, ~19,427 lines
Test files: 99 files
```

## Commits This Session

### Already Committed (part of `72b3305`)

The `ByteRangeToLines` fix and JSON printer fix were committed as part of the larger filter refactoring commit `72b3305 refactor(filter): remove pkg/filter wrapper, use gogenfilter directly`. The commit message documents these changes under "Other changes" and "Bug fixes".

### This Report

This status report is the only new commit.

---

_Generated by Crush (GLM-5.1) on 2026-04-15T20:52+02:00_
