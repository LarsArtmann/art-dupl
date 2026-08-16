# Feedback: `-t 2` on a well-factored Go codebase produces 87 single-statement `t.Parallel()` clones — threshold cliff between t2 and t3 is absolute

**Date:** 2026-08-07
**Project:** `licenseforge`, an enterprise-grade license management CLI tool (Go 1.26.4, Clean Architecture with `samber/do/v2` DI, 96 Go files discovered, 71 production / 28 test files, ~10 packages across domain/application/infrastructure/interface layers)
**art-dupl version:** `0.6.1-c170f4d`
**Command:** `art-dupl --type-aware --sort total-tokens -t 2`
**Goal:** Drive harmful duplication to zero with the user's mandate: _"GET IT DOWN TO ZERO! ... DO NOT STOP UNTIL THE ENTIRE LIST IS FINISHED and VERIFIED!"_

> **Verdict:** The codebase is genuinely clean. At `-t 3` and above: **0 clone groups**. At `-t 2`: **1 clone group with 87 occurrences**, every single one the identical one-line `t.Parallel()` call in test files. This is the most basic Go testing idiom — not harmful duplication by any definition. No source files were modified; no refactoring was needed. The session confirms that art-dupl's higher thresholds (≥3) produce excellent signal-to-noise on Go code, but `-t 2` generates a wall of single-statement test-idiom noise that buries the "zero clones" result in false positives. The `--ignore-tests` flag eliminates the noise entirely (0 clone groups at `-t 2`), proving that test files are the sole source of `-t 2` noise on this codebase.

> **Status:** No changes made. All tests pass (`GOEXPERIMENT=jsonv2 go test ./...` — all packages OK). Codebase had recent active dedup work (commits `760de50` and `c0bde9d` extract shared helpers and centralize file operations). The clean bill of health at `-t 3` is trustworthy.

---

## Results

| Threshold | Clone groups | Total occurrences | Content                                                              | Actual category                                     | Decision        |
| --------- | ------------ | ----------------- | -------------------------------------------------------------------- | --------------------------------------------------- | --------------- |
| `-t 10`   | 0            | 0                 | —                                                                    | —                                                   | —               |
| `-t 5`    | 0            | 0                 | —                                                                    | —                                                   | —               |
| `-t 3`    | 0            | 0                 | —                                                                    | —                                                   | —               |
| `-t 2`    | 1            | 87                | `t.Parallel()`                                                       | Go test idiom — single-statement concurrency marker | **Accept**      |
| `-t 1`    | 2            | 89                | 87× `t.Parallel()` + 2× `GenerateRequest = requests.GenerateRequest` | Test idiom + Go type alias re-export                | **Accept both** |

**No source files were modified.** The codebase required zero refactoring.

### Per-file breakdown of the `t.Parallel()` group at `-t 2`

| File                                                  | Occurrences |
| ----------------------------------------------------- | ----------- |
| `pkg/domain/license_type_test.go`                     | 15          |
| `pkg/validation/finding_adapter_test.go`              | 13          |
| `pkg/domain/confidence_test.go`                       | 13          |
| `pkg/domain/filepath_test.go`                         | 8           |
| `pkg/domain/email_test.go`                            | 7           |
| `pkg/domain/author_test.go`                           | 7           |
| `pkg/domain/year_test.go`                             | 6           |
| `pkg/autodetect/detector_test.go`                     | 6           |
| `pkg/workflow/workflow_test.go`                       | 5           |
| `pkg/generator/generator_test.go`                     | 3           |
| `pkg/validation/application/package_registry_test.go` | 2           |
| `pkg/autodetect/validation_test.go`                   | 2           |
| **Total**                                             | **87**      |

All 87 occurrences are in `*_test.go` files. Zero occurrences in production code (`t.Parallel()` does not exist outside test files).

---

## Finding 1: `t.Parallel()` — 87-occurrence single-statement test idiom

### What the report matched

Every single clone at `-t 2` is the identical one-line statement:

```go
t.Parallel()
```

This is the standard Go testing convention for marking a test function as eligible for parallel execution. The `testing` package documentation recommends calling `t.Parallel()` at the top of every independent test function. It is a one-method-call concurrency marker with no domain logic, no shared state, and no maintenance burden. It appears once per test function because Go test functions are independent top-level functions.

### Why this is never actionable

`t.Parallel()` is the most repetitive line in idiomatic Go test code. Extracting it is:

1. **Impossible** — it must be called individually within each `func(t *testing.T)` closure; there is no shared receiver or wrapper.
2. **Meaningless** — it carries no business logic, no data, and no intent beyond "run this test concurrently."
3. **Harmful to remove** — the Go standard library explicitly recommends parallel test execution for faster, race-detecting test suites.

Every Go project with `t.Parallel()` discipline will produce this exact clone group at `-t 2`. The group size scales linearly with the number of test functions. A project with 200 test functions would report 200 clones in one group — all `t.Parallel()`.

### Tool feedback

This pattern should be recognized and suppressed at all thresholds, not just `-t 2`. Suggested classification:

- **Pattern:** `go-test-parallel-marker`
- **Detection criteria:** Single statement `t.Parallel()` in a file matching `*_test.go`, where the enclosing function signature is `func(t *testing.T)` or `func(t *testing.TB)`.
- **Priority:** Suppressed entirely (never reported as a clone).
- **Rationale:** "Go test concurrency marker; structural repetition required by the testing framework. Not extractable."

This is directly analogous to how art-dupl already auto-excludes generated code (sqlc, templ, protobuf, mockgen, stringer). Test-file framework idioms deserve the same treatment.

---

## Finding 2 (at `-t 1`): `GenerateRequest = requests.GenerateRequest` — Go type alias re-export

### What the report matched

At `-t 1`, a second clone group appears:

```go
// pkg/generator/generator.go:34
type GenerateRequest = requests.GenerateRequest

// pkg/workflow/orchestrator.go:32
type GenerateRequest = requests.GenerateRequest
```

Both are Go type alias declarations that re-export `requests.GenerateRequest` under a local name. The comments in both files explain the identical reason: _"Moved to requests package to avoid import cycle."_

### Why this is not actionable

Type alias re-exports are a Go language feature for breaking import cycles. The `requests` package holds the canonical type definition; both `generator` and `workflow` consume it via alias to avoid circular imports. The repetition is structural — each consuming package must declare its own alias. Extracting a "shared alias" would require either:

1. A new package that both import (adding a dependency layer for one line).
2. Removing the aliases and importing `requests.GenerateRequest` directly everywhere (which may re-introduce the cycle).

Neither improves the code. The alias is the minimum ceremony Go requires.

### Tool feedback

At `-t 1` this is expected noise. The actionability filter correctly down-ranks it (it does not appear at `-t 2` default actionability). No change needed — documenting for completeness.

---

## The threshold cliff: t2 → t3

The most striking finding is the absolute cliff between `-t 2` and `-t 3`:

```
-t 10:  0 clone groups
-t 5:   0 clone groups
-t 3:   0 clone groups
-t 2:   1 clone group, 87 occurrences (all t.Parallel())
-t 1:   2 clone groups, 89 occurrences (t.Parallel() + type alias)
```

There is **nothing between** "87 clones" and "0 clones." The entire `-t 2` report is a single test idiom. This means:

- For this codebase, `-t 2` provides **zero actionable signal** while producing **maximum noise** (87 lines of output).
- A user running `-t 2` for the first time would see "found 87 clones" and have to manually verify that all 87 are the same trivial one-liner — a significant time cost with no payoff.
- The default threshold (5) is well-calibrated: it produces a clean report that correctly reflects the codebase's health.

### Recommended threshold guidance

For Go projects, the useful threshold range is narrow:

| Threshold | Value on Go codebases                                                                                         |
| --------- | ------------------------------------------------------------------------------------------------------------- |
| `-t 5+`   | Clean signal. Only real multi-statement clones. Recommended default.                                          |
| `-t 3-4`  | High signal. Catches 3-4 statement duplication. Occasionally surfaces idioms (lock scopes, builder patterns). |
| `-t 2`    | **Noise wall.** Single-statement test idioms dominate. Near-zero actionable signal on Go.                     |
| `-t 1`    | Every repeated statement. Useful only for deep forensic dedup with heavy manual filtering.                    |

The skill documentation (`-t 5` default) is well-calibrated. The user's request for `-t 2` is an aggressive edge case where the tool's output should ideally warn or annotate that the vast majority of findings are single-statement test idioms.

---

## `--ignore-tests` flag validation

The `--ignore-tests` flag works perfectly and completely eliminates the noise:

```
$ art-dupl --type-aware --sort total-tokens -t 2 --ignore-tests
    68 files discovered
 Found total 0 clone groups.
```

File count drops from 96 → 68 (28 test files excluded), and the result is a clean zero. This confirms that **100% of the `-t 2` noise originates from test files** and that the production code has no duplication at any threshold ≥ 2.

### `--exclude-pattern` did not work as expected

Attempting `--exclude-pattern '_test\.go$'` did **not** exclude test files — the full 87-clone report was still produced with all 96 files. The flag's help text says "Additional file patterns to exclude," but regex-style patterns (`_test\.go$`) appear not to be matched. The dedicated `--ignore-tests` flag is the correct mechanism for Go test exclusion.

This is a minor UX issue: a user reaching for `--exclude-pattern` to filter test files will be confused when it has no effect. Either the pattern matching should be documented more clearly (glob vs regex), or `--exclude-pattern` should warn when a pattern matches no files.

---

## What went well

1. **Type-aware mode correctly suppressed false positives.** The `GenerateRequest = requests.GenerateRequest` type alias clones are correctly type-aware — the detector recognizes they reference the same underlying type and down-ranks them at the default actionability setting.

2. **The default threshold (5) is excellent.** A user running art-dupl with no arguments would see "0 clone groups" — a correct, trustworthy result that reflects the codebase's clean state.

3. **HTML output (when used) accurately fragments and locates clones.** The terminal output's `file:line-line | snippet` format is clean and immediately scannable.

4. **Recent dedup work is visible in git history.** Commits `760de50` ("extract OrNil and readFileString helpers to remove duplication") and `c0bde9d` ("centralize file ops and author emptiness checks") show that prior dedup passes successfully eliminated real clones. The `-t 3` clean result validates that those refactors worked.

5. **The `--ignore-tests` flag exists and works.** Users who want production-only analysis have a clean escape hatch.

---

## Recommended improvements

### 1. Auto-suppress `t.Parallel()` as a Go test idiom

Single-statement `t.Parallel()` in `*_test.go` files should never be reported as a clone at any threshold. This is the single highest-impact improvement for Go projects. It would eliminate the entire 87-clone noise wall on this codebase and similar noise on every Go project that follows parallel-test conventions.

Detection criteria:

- File matches `*_test.go`
- Statement is exactly `t.Parallel()` (or `tb.Parallel()` with `testing.TB`)
- Enclosing function is a Go test function (signature `func(t *testing.T)`)

Classification: **Suppressed** — not reported at any threshold, similar to generated code categories.

### 2. Add a "single-statement clone" summary annotation

When a clone group consists entirely of single-statement occurrences (each occurrence is exactly one statement), the report should annotate it prominently:

```
found 87 clones in 1 group (single-statement idiom — review recommended before acting):
```

or in HTML:

```
⚠ Single-statement clone group: 87 occurrences of 1 statement. Likely a language/framework idiom.
```

This helps users immediately distinguish "87 real 10-line clones" from "87 copies of `t.Parallel()`" without reading every line.

### 3. Document `--exclude-pattern` pattern syntax

The `--exclude-pattern` flag should document whether it accepts glob patterns (`*_test.go`), regex (`_test\.go$`), or both. Currently a user guessing regex syntax will silently get no filtering. A warning when a pattern matches zero files would also help.

### 4. Consider a "test idiom" category alongside generated code

art-dupl already has `--include-generated` categories (sqlc, templ, protobuf, mockgen, stringer). A parallel `--include-test-idioms` flag could govern whether framework-mandated test repetition (`t.Parallel()`, `t.Helper()`, `t.Cleanup()`, `t.Setenv()`) is reported. These idioms are:

- Single-statement
- Framework-required or framework-recommended
- Not extractable
- Present in every Go test suite

Defaulting to suppressed (with opt-in via `--include-test-idioms`) would align with the generated-code precedent.

---

## Context: why this codebase is clean

The licenseforge project has several structural factors that minimize duplication:

1. **Clean Architecture with strict layer boundaries** — domain types are zero-dependency, application layer depends only on domain, infrastructure depends on domain + application. This prevents cross-layer copy-paste.

2. **`samber/do/v2` dependency injection** — services are injected via `do:""` tags, eliminating manual constructor duplication.

3. **`samber/mo` Result types** — error handling uses `mo.Result[T]` consistently, preventing repeated error-wrapping boilerplate.

4. **`types/` sub-module** — shared license types (enum, file variants, validation) live in a zero-dependency sub-module consumed by both `pkg/domain/` and external tools. This centralizes the type definitions that would otherwise be duplicated.

5. **`pkg/fileops` centralized file operations** — atomic writes with rollback are provided once, not reimplemented per package.

6. **`pkg/testing/` shared test utilities** — fixtures, mocks, and helpers are extracted into dedicated packages, preventing test-setup duplication across packages.

7. **Recent active dedup work** — the two most recent commits (`760de50`, `c0bde9d`) are explicit dedup refactors. The `-t 3` clean result confirms they were effective.

This is a codebase that has already been through dedup passes and follows patterns that resist duplication. The clean art-dupl report at `-t 3` is a trustworthy validation, not a missed-something false negative.

---

## Summary

| Question                                      | Answer                                                                                                                                                      |
| --------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Were any clones actionable?                   | **No.** All 87 at `-t 2` are `t.Parallel()`. Zero at `-t 3+`.                                                                                               |
| Were any files modified?                      | **No.** No refactoring was needed.                                                                                                                          |
| Do tests pass?                                | **Yes.** All packages OK.                                                                                                                                   |
| Is the codebase genuinely clean?              | **Yes.** Validated by `-t 3` (0 groups), `--ignore-tests -t 2` (0 groups), and git history showing active dedup work.                                       |
| Biggest improvement opportunity for art-dupl? | **Auto-suppress `t.Parallel()`** and other single-statement Go test idioms. This would eliminate 100% of the `-t 2` noise on this and similar Go codebases. |

---

## Resolution (2026-08-10)

**PARTIALLY ADDRESSED.** `test-framework-call` pattern shipped for `t.Parallel()` noise. Threshold cliff mitigation → ROADMAP. `--exclude-pattern` UX improvement → TODO_LIST.
