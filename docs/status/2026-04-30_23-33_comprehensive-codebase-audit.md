# Comprehensive Status Report — 2026-04-30

**Date:** 2026-04-30 23:33 CEST  
**Branch:** `fork`  
**Head:** `76313e5` (pushed to `origin/fork`)  
**Build:** Clean | **Tests:** 27/27 PASS | **Vet:** Clean | **Uncommitted:** 4 domain files (dead code removal, verified)

---

## A. FULLY DONE ✅

### Dead Code Elimination (Multiple Sessions)

| Item | Commit | Lines Removed |
|------|--------|---------------|
| 690 lines of dead domain types (Clone, CloneGroup, Analysis, Repository, SourceFile, CoverageAnalysis, etc.) | `0b88414` | -690 |
| Unused `DetectionOptions` type | `40c6c5b` | - |
| Unused `GetThresholdAsDomain`, `SetThresholdFromDomain` | `44e736a` | - |
| Unused `ParseSortBy` function | `f8e466e` | - |
| Unused `BatchHash`, `HashSeqWithConfig`, `HashConfig` | `b657700` | - |
| Exported test-only helpers from config | `daca611` | - |
| 6 dead `migration/` exclusion blocks from `.golangci.yml` | `5342b85` | - |
| Panics replaced with error returns in `cmd/config_builder.go` | `891fba3` | - |
| Dead wrappers and shallow redirect functions | `2aef77b` | - |
| **Uncommitted: Remove dead domain helpers** (BytePosition, TokenCount, FileCount, CloneCount, Threshold, unmarshalUint, unmarshalUintNonZero, marshalUint, unmarshalUintGeneric, 15 dead error sentinels) | pending | -201 |

**Total dead code removed across all sessions: ~1,100+ lines**

### Type Safety Improvements

| Change | Commit | Impact |
|--------|--------|--------|
| `config.Config.Only` `string` → `config.FileType` | `5342b85` | Eliminates hand-rolled `matchesOnlyFilter()` |
| `domain.Analysis.CreatedAt` `string` → `time.Time` | `5342b85` | Compile-time type safety |
| `domain.Analysis.CompletedAt` `*string` → `*time.Time` | `5342b85` | Compile-time type safety |
| `GetStatsData()` `any` → `*StatsData` | `dbfd4ce` | Eliminates type assertions at all 7 call sites |
| `buildJSONData()` `any` → `jsonStatsOutput` | (same session) | Return type documents intent |
| `errors.As` → `errors.AsType` (3 instances) | `5342b85` | Named error type matching |
| All `//nolint:err113` replaced with typed errors | `5342b85` | 4 files fixed |

### API Design: StatsPrinter Consolidation

| Before | After | Commit |
|--------|-------|--------|
| 8 individual setter methods on `StatsPrinter` interface | `ApplyStatsConfig(StatsConfig)` + `GetStatsData() *StatsData` | `dbfd4ce` |

### Stdlib Over Hand-Rolled Code

| Item | Before | After | Commit |
|------|--------|-------|--------|
| `config/filetype.go` `hasSuffix()` | 8-line hand-rolled suffix check | `strings.HasSuffix` | `dbfd4ce` |
| `printer/diff.go` `htmlEscape()` | 4x `strings.ReplaceAll` | `html.EscapeString` (also escapes `'`) | `a4db4a2` |

### Structural Splits

| File | Before | After | Commit |
|------|--------|-------|--------|
| `printer/html.go` | 1,484 lines monolith | 4 files: `html.go` (364L), `html_template.go` (523L), `html_diff.go` (315L), `html_summary.go` (301L) | `24d1902` |

### CI/Infrastructure

| Fix | Commit |
|-----|--------|
| CI go-version `1.26rc2` → `stable` | `8a6931c` |
| Nix flake private dep pattern (dummy/replace for gogenfilter) | `358c317` (prior session) |

### Documentation

| Item | Commit |
|------|--------|
| `CONTRIBUTING.md` created | `6d1bcfe` |
| `TODO_LIST.md` updated with completed items | `76313e5` |
| `AGENTS.md` updated with architecture decisions | `e13997b` |
| `.go-arch-lint.yml` replaced with project-specific config | `dbfd4ce` |

### Other Completed Work (Parallel Sessions)

| Item | Commit |
|------|--------|
| SDK detection pipeline unified via `MultiDetector` | `e72f85d` |
| Go toolchain and dependency updates | `0197d2a` |
| Landing page redesign and deployment config | `b601e44` |
| Site factual errors corrected (6 fixes) | `74943e6` |
| Consistent Go formatting across codebase | `51de4e7` |

---

## B. PARTIALLY DONE ⚠️

### Domain Package Slim-Down

**Status:** Uncommitted changes ready to commit.

The domain package has been audited and dead types identified. Uncommitted changes remove:
- `domain/types_metric.go` (entire file, 114 lines) — `TokenCount`, `FileCount`, `CloneCount`, `Threshold` types (all unused)
- `domain/types_file.go` — `BytePosition` type removed (22 lines)
- `domain/helpers.go` — `unmarshalUintNonZero`, `unmarshalUint`, `marshalUint`, `unmarshalUintGeneric` removed (44 lines)
- `domain/analysis_errors.go` — 15 dead error sentinels removed, 2 kept (`ErrInvalidCloneSeverity`, `ErrInvalidSeverity`)

**What remains:**
- `domain/helpers.go` still has `marshalStringID`, `unmarshalWithValidation`, `unmarshalStringID` — used by `Filepath`, `LineNumber`, `CloneSeverity`
- `domain/domain.go` is just a package doc comment (9 lines)
- Domain package has **zero test files** — no test coverage

### Threshold Validation Consolidation

**Status:** Identified, not started.

Three independent threshold validations exist:
1. `config/config.go:266` — `validateThreshold()` with `errors.NewValidationError`
2. `pkg/artdupl/types.go:166` — `ValidateOptions()` with `ErrInvalidThreshold` + `ErrThresholdTooLarge`
3. `domain/types_metric.go` — `NewThreshold()` with `ErrThresholdInvalid` (being deleted)

---

## C. NOT STARTED 📋

### High Priority

| # | Item | Effort | Impact |
|---|------|--------|--------|
| 1 | TokenValue type for suffixtree/syntax | Medium | High |
| 2 | Implement CSV output with `encoding/csv` | Low | Medium |
| 3 | Consolidate 3 threshold validation points → 1 | Low | Medium |
| 4 | Unify enum patterns (domain switch vs config map) | Medium | Medium |

### Medium Priority

| # | Item | Effort | Impact |
|---|------|--------|--------|
| 5 | Type `domain.Options.OutputFormat` as `config.OutputFormat` | Low | Low |
| 6 | Refactor `syntax/golang/transform.go` (355L, 300L switch) | Medium | Medium |
| 7 | Split `detection/todos.go` (TodoDetector + LegacyDetector) | Low | Low |
| 8 | Split `config/config.go` (341L) — config + validation mixed | Low | Low |
| 9 | Split `cmd/run_analysis.go` (447L) — multiple responsibilities | Medium | Low |
| 10 | Update README with semantic detection behavior | Low | Medium |

### Low Priority

| # | Item | Effort | Impact |
|---|------|--------|--------|
| 11 | Archive old docs/status/ files (312 files) | Low | Low |
| 12 | Fix remaining nolint directives (funlen in 4 cmd/ files) | Medium | Low |
| 13 | SIMD TODOs (6 items) | High | Unknown |
| 14 | Remove `fmt` import from `domain/helpers.go` (only used by deleted functions) | Trivial | Trivial |

---

## D. TOTALLY FUCKED UP 💥

### Gopls Phantom Duplicate Errors

**Severity:** Annoying, not blocking  
**Status:** Known bug, NOT a real compilation error

gopls reports 15+ `DuplicateDecl` / `DuplicateMethod` errors between `printer/html.go` and `printer/html_diff.go` / `printer/html_summary.go`. These are **phantom errors** — `go build` compiles cleanly. This is a known gopls caching bug that appeared after the `html.go` split in commit `24d1902`.

**Attempted fixes:** None successful yet. The LSP cache appears to hold stale pre-split file contents.

### Vendor Directory Inconsistency

**Severity:** CI-blocking for linter (if using vendor mode)  
**Status:** Not fixed

The `vendor/modules.txt` is out of sync with `go.mod`. Several packages are at different versions:
- `gogenfilter`: go.mod has `v0.2.1-0.20260430200039`, vendor has `v0.2.1-0.20260430195342`
- `ginkgo/v2`: go.mod has `v2.28.3`, vendor has `v2.28.1`
- `gomega`: go.mod has `v1.40.0`, vendor has `v1.39.1`

This causes `golangci-lint` to fail when loading packages in vendor mode. Not blocking `go build` or `go test` (they use module mode).

**Fix:** Run `go mod vendor` to sync, but this needs to be coordinated with the Nix flake which has special vendor handling.

### Domain Package Has Zero Test Coverage

**Severity:** Quality concern  
**Status:** Not addressed

`go test ./domain/...` reports `[no test files]`. The domain package defines `Filepath`, `LineNumber`, `CloneSeverity` with validation, JSON marshaling — all untested.

---

## E. WHAT WE SHOULD IMPROVE

### Architecture

1. **Domain package is hollowed out** — After removing dead types, domain/ has 226 lines total (Filepath, LineNumber, CloneSeverity, helpers). It should either gain a clear purpose or be merged into another package.

2. **Three parallel type systems** — `config` enums use generic `isValidStringType[T]()` with map lookups, while `domain` enums use switch-based `IsValid()`. Unify the pattern.

3. **Printer package is still the largest** — 14 non-test files, 9 of which are stats-related. Consider a `printer/stats/` sub-package.

4. **`examples/domain_types_usage.go` was referenced in planning but never existed** — the example file uses the SDK, not domain types directly.

### Process

5. **No CI running** — CI workflow exists but uses stale vendor. Need to run `go mod vendor` and push, or switch CI to module mode.

6. **312 status report files in `docs/status/`** — Should archive older than 30 days. Most are session-specific and stale.

7. **TODO_LIST.md tracks items but no due dates or ownership** — Fine for solo project, but should triage more aggressively.

### Code Quality

8. **Domain package needs tests** — Every exported function should have at least one test.

9. **`printer/stats.go:60` has a stale comment** referencing `domain.Threshold` which is being deleted.

10. **`config/config_merge.go` has `//nolint:funlen,gocognit,gocyclo,cyclop`** — Major complexity debt in a single function.

---

## F. TOP 25 THINGS TO DO NEXT

Prioritized by **impact × effort** (highest first):

| # | Item | Impact | Effort | Category |
|---|------|--------|--------|----------|
| 1 | Commit pending domain dead code removal (-201 lines) | High | Trivial | Dead code |
| 2 | Run `go mod vendor` to sync vendor with go.mod | High | Trivial | Infra |
| 3 | Consolidate threshold validation (3 → 1) | Medium | Low | Type safety |
| 4 | Add tests for domain package (Filepath, LineNumber, CloneSeverity) | Medium | Low | Quality |
| 5 | Implement CSV output with `encoding/csv` | Medium | Low | Feature |
| 6 | Fix stale comment in `printer/stats.go:60` | Low | Trivial | Cleanup |
| 7 | Unify enum patterns (domain switch → config generic pattern) | Medium | Medium | Architecture |
| 8 | Restart gopls or clear cache to fix phantom errors | Low | Trivial | DX |
| 9 | Archive docs/status/ files older than 30 days | Low | Trivial | Housekeeping |
| 10 | Refactor `config/config_merge.go` to reduce cyclomatic complexity | Medium | Medium | Quality |
| 11 | Split `printer/stats*` files into `printer/stats/` sub-package | Medium | Medium | Architecture |
| 12 | Update README with semantic detection as default behavior | Medium | Low | Docs |
| 13 | Implement TokenValue type for suffixtree/syntax | High | Medium | Type safety |
| 14 | Split `cmd/run_analysis.go` (447L) into focused files | Low | Low | Structural |
| 15 | Split `detection/todos.go` into TodoDetector + LegacyDetector | Low | Low | Structural |
| 16 | Split `config/config.go` (341L) into config + validation | Low | Low | Structural |
| 17 | Type `domain.Options.OutputFormat` as `config.OutputFormat` | Low | Low | Type safety |
| 18 | Extract detection mode strings to typed enum (DetectionMode) | Medium | Low | Type safety |
| 19 | Add CONTRIBUTING.md section on running tests | Low | Trivial | Docs |
| 20 | Fix nolint:funlen directives in cmd/ (4 files) | Low | Medium | Quality |
| 21 | Investigate and document Nix flake → CI integration | Medium | Medium | Infra |
| 22 | Add `fmt` import removal from `domain/helpers.go` (unused after deletion) | Trivial | Trivial | Cleanup |
| 23 | Benchmark stats printer performance with large projects | Unknown | Medium | Performance |
| 24 | Implement SIMD optimizations (6 TODOs) | Unknown | High | Performance |
| 25 | Investigate `domain/` package purpose — merge or expand | Medium | Medium | Architecture |

---

## G. TOP #1 QUESTION I CANNOT FIGURE OUT MYSELF

**Should the `domain/` package exist at all?**

After the dead code removal (pending commit), `domain/` contains only:
- `Filepath` (string wrapper with JSON marshaling)
- `LineNumber` (uint16 wrapper with JSON marshaling)
- `CloneSeverity` (string enum with 4 values)
- 2 error sentinels (`ErrInvalidCloneSeverity`, `ErrInvalidSeverity`)
- Generic JSON helper functions (`marshalStringID`, `unmarshalWithValidation`, `unmarshalStringID`)

The package is **226 lines total** with **zero runtime consumers** — the actual pipeline uses `syntax.Node`, `artdupl.Clone`, and raw primitives. No production code path flows through `domain.Filepath` or `domain.LineNumber` except in `detection/todos.go` (for creating log entries).

**The options as I see them:**
1. **Keep as-is** — Domain types serve as a reference model for what the codebase *should* use eventually
2. **Merge into `config/`** — These are configuration-adjacent types; `config/` already has typed enums
3. **Merge into `syntax/`** — `Filepath` and `LineNumber` are syntax-adjacent; but this creates circular deps
4. **Delete entirely** — `CloneSeverity` is the only runtime-used type; move it to `printer/` where it's consumed

This is an **architecture direction decision** that affects whether to invest in the domain layer or simplify it away. It determines the trajectory of items #7, #13, and #25 above.

---

## Session Stats

| Metric | Value |
|--------|-------|
| Total commits on fork (last 30) | 30 |
| Tests passing | 27/27 |
| Go files (non-test, non-vendor) | ~210 |
| Largest Go file | `printer/html_template.go` (523L) |
| Domain package | 226 lines, 0 test coverage |
| Status reports in docs/status/ | 312 |
| CI go-version | `stable` (was `1.26rc2`) |
| Vendor sync status | Out of date |
| Build | Clean |
| Vet | Clean |
