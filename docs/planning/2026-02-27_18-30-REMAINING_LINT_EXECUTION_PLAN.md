# Comprehensive Lint Fix Execution Plan

**Date:** 2026-02-27 18:30\
**Status:** P0 (Security) and P1 (Error Handling) COMPLETE\
**Remaining Issues:** ~340

---

## Current State Analysis

### Completed (P0 + P1)

- ✅ **gosec:** 5 issues (G115, G204) - Security vulnerabilities fixed
- ✅ **forcetypeassert:** 2 issues - Safe type assertions added
- ✅ **err113:** 16 issues - Dynamic errors converted to wrapped static errors

### Remaining Issues by Category

```
Priority 2 - Complexity (6 issues)          [P2]
  - gocyclo: 2          (Cyclomatic complexity)
  - gocognit: 2         (Cognitive complexity)
  - funlen: 1           (Function length)
  - nestif: 2           (Nested if statements)

Priority 3 - Quality (76 issues)            [P3]
  - revive: 50          (General code quality)
  - unparam: 9          (Unused parameters)
  - prealloc: 11        (Slice preallocation)
  - noinlineerr: 3      (Inline error handling)
  - unconvert: 1        (Unnecessary conversions)
  - maintidx: 1         (Maintainability index)
  - nolintlint: 2       (Bad nolint directives)

Priority 4 - Testing (26 issues)            [P4]
  - thelper: 4          (Test helper functions)
  - tparallel: 1         (Parallel test execution)
  - usetesting: 2        (Use testing package)
  - recvcheck: 19        (Receiver naming)

Priority 5 - Documentation/Style (94 issues) [P5]
  - godoclint: 20       (Documentation)
  - godox: 6            (TODO markers)
  - nonamedreturns: 7    (Named return values)
  - goprintffuncname: 5  (Printf naming)

Priority 6 - Structural (200+ issues)        [P6]
  - mnd: 50             (Magic numbers)
  - tagliatelle: 50      (Struct field tags)
  - varnamelen: 50       (Variable name length)
  - exhaustruct: 50+     (Struct initialization)
  - gochecknoglobals: 1   (Global variables)
  - gosmopolitan: 2      (i18n issues)
```

---

## Multi-Step Execution Plan

### Phase P2: Complexity Reduction (6 issues, ~60 min)

**Goal:** Reduce cyclomatic and cognitive complexity

| Step | Task                | File(s)                | Time   | Verification                               |
| ---- | ------------------- | ---------------------- | ------ | ------------------------------------------ |
| P2.1 | Fix gocyclo issues  | syntax/golang/parse.go | 15 min | `golangci-lint run syntax/golang/parse.go` |
| P2.2 | Fix gocognit issues | detection/todos.go     | 15 min | `golangci-lint run detection/todos.go`     |
| P2.3 | Fix funlen issue    | cmd/run_analysis.go    | 10 min | `golangci-lint run cmd/run_analysis.go`    |
| P2.4 | Fix nestif issues   | cmd/run_analysis.go    | 20 min | `golangci-lint run cmd/run_analysis.go`    |

**Pattern:** Extract helper functions, use early returns, reduce nesting

---

### Phase P3: Code Quality (76 issues, ~120 min)

**Goal:** Improve maintainability and performance

| Step  | Task                        | Issues | Time   | Pattern                                |
| ----- | --------------------------- | ------ | ------ | -------------------------------------- |
| P3.1  | Fix unparam (unused params) | 9      | 20 min | Remove unused params or add `_` prefix |
| P3.2  | Fix prealloc                | 11     | 15 min | Add `make([]T, 0, capacity)`           |
| P3.3  | Fix noinlineerr             | 3      | 10 min | Extract error to variable              |
| P3.4  | Fix unconvert               | 1      | 5 min  | Remove unnecessary conversion          |
| P3.5  | Fix maintidx                | 1      | 10 min | Simplify complex function              |
| P3.6  | Fix nolintlint              | 2      | 10 min | Add explanations to nolint             |
| P3.7  | Fix revive batch 1          | 12     | 20 min | Add comments, fix naming               |
| P3.8  | Fix revive batch 2          | 12     | 20 min | Add comments, fix naming               |
| P3.9  | Fix revive batch 3          | 12     | 20 min | Add comments, fix naming               |
| P3.10 | Fix revive batch 4          | 14     | 20 min | Add comments, fix naming               |

**Pattern:**

- Use `t.Helper()` in test helpers
- Add package comments: `// Package X provides...`
- Export comments: `// FuncName does...`

---

### Phase P4: Testing Improvements (26 issues, ~45 min)

**Goal:** Consistent test patterns

| Step | Task                            | Issues | Time   |
| ---- | ------------------------------- | ------ | ------ |
| P4.1 | Add t.Helper() to test helpers  | 4      | 10 min |
| P4.2 | Enable tparallel                | 1      | 5 min  |
| P4.3 | Use t.TempDir()                 | 2      | 10 min |
| P4.4 | Fix receiver naming (recvcheck) | 19     | 25 min |

**Pattern:**

```go
func helper(t *testing.T) {
    t.Helper()
    // ...
}
```

---

### Phase P5: Documentation & Style (38 issues, ~60 min)

**Goal:** Documentation coverage and style consistency

| Step | Task                       | Issues | Time   |
| ---- | -------------------------- | ------ | ------ |
| P5.1 | Add package documentation  | 10     | 15 min |
| P5.2 | Add function documentation | 10     | 15 min |
| P5.3 | Resolve godox TODOs        | 6      | 10 min |
| P5.4 | Fix nonamedreturns         | 7      | 10 min |
| P5.5 | Fix goprintffuncname       | 5      | 10 min |

**Pattern:**

```go
// Package domain provides core business logic types.
package domain

// FuncName performs operation X.
func FuncName() {}
```

---

### Phase P6: Structural (200+ issues, ~180 min)

**Goal:** Address architectural patterns

| Step  | Task                    | Issues | Strategy                 |
| ----- | ----------------------- | ------ | ------------------------ |
| P6.1  | Fix mnd batch 1         | 15     | Extract to const         |
| P6.2  | Fix mnd batch 2         | 15     | Extract to const         |
| P6.3  | Fix mnd batch 3         | 10     | Extract to const         |
| P6.4  | Fix tagliatelle batch 1 | 15     | Fix struct tags          |
| P6.5  | Fix tagliatelle batch 2 | 15     | Fix struct tags          |
| P6.6  | Fix tagliatelle batch 3 | 10     | Fix struct tags          |
| P6.7  | Fix varnamelen batch 1  | 15     | Rename variables         |
| P6.8  | Fix varnamelen batch 2  | 15     | Rename variables         |
| P6.9  | Fix varnamelen batch 3  | 10     | Rename variables         |
| P6.10 | Fix exhaustruct batch 1 | 15     | Add nolint for 3rd party |
| P6.11 | Fix exhaustruct batch 2 | 15     | Add nolint for 3rd party |
| P6.12 | Fix exhaustruct batch 3 | 10     | Add nolint for 3rd party |

**Key Decision:** For `exhaustruct` with third-party types (cobra.Command, etc.), use `//nolint:exhaustruct` since we can't modify those structs.

---

## Architecture Improvements Considered

### 1. Error Handling Pattern (Implemented in P1)

**Before:**

```go
return fmt.Errorf("invalid %s: %s", typeName, value)
```

**After:**

```go
var ErrInvalidType = errors.New("invalid type")
return fmt.Errorf("%w: %s=%s", ErrInvalidType, typeName, value)
```

**Benefits:**

- Errors can be checked with `errors.Is()`
- Better error wrapping with context
- Consistent across codebase

### 2. Type Model Consolidation

**Observation:** Multiple similar validation patterns exist across domain types.

**Potential Improvement:**

```go
// Validatable interface for all domain types
type Validatable interface {
    IsValid() error
}

// Generic validation helper
func Validate[T Validatable](v T) error {
    return v.IsValid()
}
```

**Deferred:** Not required for lint fixes, consider for future refactoring.

### 3. Library Usage

**Current:** Custom enum marshaling in `internal/enum/`
**Consider:** `github.com/samber/mo` for Option types (already in go.mod)

**Decision:** Keep current implementation - it's working and tested.

---

## Execution Rules

1. **Commit after each file:** `git add <file> && git commit -m "lint: fix X in file.go"`
2. **Verify before commit:** `go build ./... && golangci-lint run <file>`
3. **Small batches:** Max 3 files per commit
4. **Document patterns:** Update this plan if new patterns emerge
5. **Push regularly:** Every 30 minutes or after each phase

---

## Success Metrics

| Phase | Issues | Time    | Cumulative |
| ----- | ------ | ------- | ---------- |
| P0+P1 | 23     | 45 min  | 23 (6%)    |
| P2    | 6      | 60 min  | 29 (8%)    |
| P3    | 76     | 120 min | 105 (31%)  |
| P4    | 26     | 45 min  | 131 (39%)  |
| P5    | 38     | 60 min  | 169 (50%)  |
| P6    | 170    | 180 min | 339 (100%) |

**Target:** Zero lint issues, clean build, all tests passing.

---

## Questions to Consider

1. **Should we disable certain linters?**
   - `exhaustruct` for test files might be too strict
   - `varnamelen` for short-lived variables may be acceptable

2. **Priority adjustment?**
   - P6 issues (mnd, tagliatelle, varnamelen) are style-heavy
   - Could defer some to post-MVP if time-constrained

3. **Test file exemptions?**
   - Many linters already disabled for `*_test.go` in `.golangci.yml`
   - Verify exclusions are working

---

**Ready to execute Phase P2.**

**Assisted-by:** Claude via Crush <crush@charm.land>
