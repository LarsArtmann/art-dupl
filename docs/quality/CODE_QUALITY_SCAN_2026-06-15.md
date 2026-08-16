# Code Quality Scan — 2026-06-15

## Summary

| Check       | Status   | Notes                                                      |
| ----------- | -------- | ---------------------------------------------------------- |
| Build       | ✅ PASS  | `go build ./...` clean                                     |
| Lint        | ✅ PASS  | `golangci-lint run` — 0 issues (fixed 1 exhaustive switch) |
| Tests       | ✅ PASS  | All 24 packages pass                                       |
| go vet      | ✅ PASS  | No issues                                                  |
| go.sum tidy | ✅ FIXED | `go mod tidy` added 47 missing transitive hashes           |
| Duplication | ✅ CLEAN | Zero duplication in production code at 8-line threshold    |

## Fixed This Session

| # | File             | Line | Issue                                  | Fix                 |
| - | ---------------- | ---- | -------------------------------------- | ------------------- |
| 1 | printer/stats.go | 159  | Missing `domain.PriorityLow` in switch | Added explicit case |
| 2 | go.sum           | —    | Missing transitive dependency hashes   | Ran `go mod tidy`   |

## Issues Found (Sorted by Impact)

### 🔴 High — Files Over 350 Lines (Architectural)

These files exceed the 350-line guideline. Splitting improves testability and locality.

| #  | File                             | Lines | Recommendation                            |
| -- | -------------------------------- | ----- | ----------------------------------------- |
| 1  | printer/actionability.go         | 558   | Extract pattern matching logic            |
| 2  | printer/html_template.go         | 523   | Split HTML generation from templating     |
| 3  | printer/stats_formatter.go       | 484   | Split formatting by output type           |
| 4  | printer/diff.go                  | 389   | Extract diff rendering from diff logic    |
| 5  | syntax/golang/transform.go       | 369   | Split AST transform by node category      |
| 6  | cache/file_cache.go              | 364   | Split cache operations from serialization |
| 7  | syntax/syntax.go                 | 359   | Split node processing from tree building  |
| 8  | printer/html.go                  | 354   | Split HTML rendering phases               |
| 9  | internal/testutil/bdd_helpers.go | 352   | Split BDD helpers by concern              |
| 10 | cmd/run_analysis.go              | 351   | Split analysis orchestration steps        |

### 🟡 Medium — Test Code Duplication

13 clone groups in test code (threshold=20). Mostly repetitive test setup patterns in `printer/actionability_patterns_test.go` and `printer/actionability_test.go`.

### 🟢 Low — Observations

- No panics in production code (only in `Must*` pattern and test helpers — expected)
- No `interface{}` in production code (test data only)
- No real TODOs in production code (test TODOs are test fixtures for the TODO detector)
- No unused exports detected
- Architecture lint config (`.go-arch-lint.yml`) present and well-defined
