# Comprehensive Project Status Report — art-dupl

**Generated:** 2026-05-14 09:48 UTC
**Branch:** fork (up to date with origin/fork)
**Go Version:** 1.26.2
**Module:** github.com/LarsArtmann/art-dupl

---

## a) FULLY DONE

### Code Deduplication Sprint (This Session — 2026-05-14)

| Clone Group | Before | After | Fix Applied |
|-------------|--------|-------|-------------|
| **Total clone groups** | 130 | **125** | −5 groups (8 raw clones eliminated) |
| `internal/simd/simd_test.go` | 2 duplicate nil-check loops | 1 helper `assertAllHashSliceResultsNil()` | Extracted helper for HashSlice validation |
| `printer/diff_test.go` + `html_test.go` | 2 identical Base.Filename assertions | 1 helper `assertBaseFilename()` | Extracted CloneGroupDiff assertion helper |
| `internal/testutil/bdd_helpers.go` | 2 nearly-identical methods | 1 unified method | Merged `CreateNamedDuplicateFilesAndRun` + `CreateAndRunDupl` |
| `internal/filtertest/integration_filter_test.go` | 2 inline filter checks | `AssertFileShouldBeFiltered()` helper | Used existing helper from `assertions.go` |
| `bdd/plumbing_output_test.go` | Called removed `CreateAndRunDupl` | Updated to `CreateNamedDuplicateFilesAndRun` | Fixed caller after helper merge |

### Previous Major Work (2026-05-03 → 2026-05-13)

- **Detection pipeline wiring** — `TodoDetector` and `LegacyDetector` wired through `MultiDetector` registry
- **Config extraction** — `DetectionConfig` split from `config.Config` for detection layer isolation
- **File splits** — `detection/todos.go` (352L→3 files), `config/config.go` (344L→3 files), `cmd/run_analysis.go` (450L→3 files)
- **Domain type cleanup** — Removed 6 unused types (TokenCount, FileCount, CloneCount, etc.) and 15 dead error variables
- **Threshold consolidation** — Error sentinels unified in `config/enum_helpers.go`
- **Printer format migration** — `ParseFormat` moved from `printer/format.go` to `config.ParseOutputFormat`
- **Sort criteria migration** — `SortBy` constants moved from `printer/sort_type.go` to `config.SortCriteria`
- **HTML split** — `printer/html.go` (1484L→4 files: html.go, html_template.go, html_diff.go, html_summary.go)
- **Semantic detection** — FNV-1a 24-bit identifier hashing, `--semantic`/`--structural` flags working
- **Stats subcommand** — Text/JSON/CSV with health grades, clone metrics, severity distributions
- **SQLC/Templ filtering** — Auto-detects sqlc.yaml, includes templ by default, --exclude-templ flag
- **SARIF output** — Full SARIF 2.1.0 for GitHub Advanced Security / CodeQL integration
- **Multi-method detection** — `-m "hash,art-dupl"` runs both simultaneously via goroutines

### Test Infrastructure

- **All 22 packages pass** with `go test ./...`
- **BDD suite:** 250/250 specs pass (Ginkgo/Gomega)
- **Coverage** see section below — median ~90%, several at 100%
- **Internal test utilities** extracted: `testutil.AssertCount`, `AssertFilesShouldBeFiltered`, BDD helpers

---

## b) PARTIALLY DONE

| Item | Status | What's Missing |
|------|--------|----------------|
| **Code deduplication** | 130→125 clone groups | Remaining groups are architectural (PrintClones 6-file loop, sort switches) or idiomatic Go patterns |
| **CSV output** | Manual formatting works | Not using `encoding/csv` package |
| **SIMD optimizations** | `VectorSize()`, `Available()`, AVX-512 prep | 6 TODOs in `syntax/hash_simd.go` and `internal/simd/` — no actual SIMD paths active |
| **SDK/API** | `pkg/artdupl/` has `Detector` interface | Streaming support partially implemented, documentation incomplete |
| **Nix flake** | Builds successfully | Private `gogenfilter` dependency requires two-phase dummy/replace pattern — fragile when gogenfilter rev changes |
| **Semantic default** | Config default `Semantic: false` | Flag description says "already the default" which is inconsistent |
| **Domain types** | Cleaned 6 unused types | `Filepath`, `LineNumber`, `CloneSeverity` still in use but underutilized |
| **Printer DTO refactor** | Documented in AGENTS.md | `Printer.PrintClones(dups [][]*syntax.Node)` forces all 6 implementations to couple to AST internals — 111 test call sites to migrate |

---

## c) NOT STARTED

| Item | Priority | Rationale |
|------|----------|-----------|
| **TokenValue type** | HIGH | Would strengthen type safety between suffixtree/syntax; no work started |
| **String interning** | MEDIUM | Memory optimization for repeated identifiers; requires profiling data |
| **Multi-language support** | MEDIUM | Currently Go + Templ only; architecture partly extensible via `syntax/` package |
| **Plugin system** | LOW | Architecture could support pluggable detectors via `MethodDetector` interface but no dynamic loading |
| **Performance benchmarks** | LOW | `just bench` exists but no baseline tracked; no CI performance regression testing |
| **Mutation testing** | LOW | No Go mutation testing tools integrated |
| **Property-based testing** | LOW | Only one fuzz test exists; could expand to generators |

---

## d) TOTALLY FUCKED UP!

### None — project builds and tests pass.

**However, critical risks exist:**

| Risk | Severity | Detail |
|------|----------|--------|
| **Nix flake gogenfilter pin** | HIGH | If gogenfilter rev changes, `vendorHash` must be manually updated. Build will fail with cryptic error until hash is fixed. The dummy/replace pattern is necessary but brittle. |
| **domain/ package 0% coverage** | MEDIUM | `domain/` contains production types with `IsValid()` methods but no unit tests. Coverage tool reports 0%. |
| **Printer ↔ syntax.Node coupling** | MEDIUM | 6 printer files + 111 test call sites all depend on `syntax.Node` internals. Refactoring this is a breaking change. |
| **Three parallel Clone types** | MEDIUM | `printer.clone` (unexported), `pkg/artdupl.Clone` (SDK), `printer.CloneGroup` (output). Data duplication across boundaries. |
| **304 stale docs/status/ files** | LOW | docs/status/ has 304 files from months of status reports. Should archive or git-ignore. |
| **LSP unused-write hints** | LOW | 15+ `gopls unusedwrite` diagnostics in `pkg/artdupl/detector_types_test.go` — tests assign to struct fields that are never read. These are valid hints but not compilation errors. |

---

## e) WHAT WE SHOULD IMPROVE!

### Immediate (this week)

1. **Printer DTO decoupling** — Introducing `ProcessedClone` type would eliminate the `syntax.Node` coupling and unify clone representations across the 3 parallel types. This is the single highest-impact structural improvement.
2. **domain/ test coverage** — Add unit tests for `Clone.IsValid()`, `CloneGroup.IsValid()`, `StringPool`. Currently invisible to coverage.
3. **Archive docs/status/** — Move 304 stale status files to `docs/status/archive/YYYY-MM/` pattern. Keep only last 30 days.
4. **Fix LSP unused-write hints** — 15+ in `pkg/artdupl/detector_types_test.go`; likely test setup code that assigns but never asserts.

### Short-term (2-4 weeks)

5. **TokenValue type** — Formalize what "a token" means with validation rules. Strengthens the core data model.
6. **CSV output via encoding/csv** — Replace manual string formatting with proper CSV encoding.
7. **Unify enum patterns** — `domain/` enums should use config's generic `ParseEnum`/`MarshalJSON` helpers.
8. **Consolidate three Clone types** — Depends on #1 (ProcessedClone DTO).
9. **Error context enhancement** — Some errors still use `%w` wrapping without typed context; `errors/` package has good patterns but not used everywhere.
10. **Implement 6 SIMD TODOs** — `syntax/hash_simd.go` and `internal/simd/` have placeholder methods.

### Medium-term (1-2 months)

11. **Performance benchmarking** — Establish baselines for large repo scanning. Track memory and CPU.
12. **Cli/ command help text consolidation** — Cobra command help is scattered across initialize methods; could use a registry pattern.
13. **Add more output format integration tests** — HTML, JSON, SARIF tested but not exhaustively for edge cases.
14. **Multi-detection method performance** — Running hash+art-dupl is not benchmarked; potential for optimizing deduplication channel logic.

---

## f) Top #25 Things to Get Done Next

| # | Task | Priority | Effort | Impact |
|---|------|----------|--------|--------|
| 1 | Introduce `ProcessedClone` DTO, decouple Printer from syntax.Node | HIGH | 2-3d | CRITICAL |
| 2 | Add unit tests for domain/ types (Clone, CloneGroup, StringPool) | HIGH | 4h | HIGH |
| 3 | Archive 304 stale docs/status/ files | LOW | 1h | LOW |
| 4 | Fix LSP unused-write diagnostics in detector_types_test.go | LOW | 2h | LOW |
| 5 | Implement TokenValue type with validation | HIGH | 1-2d | HIGH |
| 6 | Implement CSV output via encoding/csv | MEDIUM | 4h | MEDIUM |
| 7 | Unify enum patterns across domain and config | MEDIUM | 1d | MEDIUM |
| 8 | Consolidate three parallel Clone types | MEDIUM | 1-2d | MEDIUM |
| 9 | Implement 6 SIMD TODOs in hash_simd.go + internal/simd | LOW | 1-2d | LOW |
| 10 | Refactor syntax/golang/transform.go (355L main function) | LOW | 1-2d | LOW |
| 11 | Unify semantic default: align config default with flag description | LOW | 30m | LOW |
| 12 | Add performance benchmarks for large repo scanning | MEDIUM | 1d | MEDIUM |
| 13 | Wire TodoDetector + LegacyDetector to CLI (currently registry-only) | LOW | 4h | LOW |
| 14 | Enhanced error context in job/ and detection/ packages | MEDIUM | 1d | MEDIUM |
| 15 | Property-based tests for core algorithms (suffix tree, hash) | LOW | 2d | LOW |
| 16 | Mutation testing integration | LOW | 2d | LOW |
| 17 | Support additional languages (TypeScript, Python) via plugin | LOW | 1-2w | HIGH |
| 18 | Memory layout optimization for SIMD-friendly structures | MEDIUM | 1-2d | MEDIUM |
| 19 | String interning for repeated identifiers | LOW | 1-2d | MEDIUM |
| 20 | Add `art-dupl check` subcommand for self-analysis | LOW | 1d | MEDIUM |
| 21 | Improve nix flake gogenfilter dependency robustness | MEDIUM | 4h | MEDIUM |
| 22 | Unify BDD test setup patterns — shared contexts for common scenarios | LOW | 1d | LOW |
| 23 | Add integration tests for all output formats with golden files | MEDIUM | 1-2d | MEDIUM |
| 24 | Profiling and memory leak detection for long-running scans | MEDIUM | 1d | MEDIUM |
| 25 | Auto-update vendorHash in nix flake via CI | LOW | 4h | LOW |

---

## g) Top #1 Question I Cannot Figure Out

### Q: Should `ProcessedClone` DTO be in `domain/` or `printer/`?

**Context:** The Printer DTO refactor (Task #1) is the most critical architectural improvement. It would eliminate the `syntax.Node` coupling across 6 printer implementations + 111 test call sites, and enable consolidation of the three parallel Clone types.

**Dilemma:**
- If placed in `domain/`: Keeps domain types centralized. But `domain/` currently has zero deps on `syntax/` which is a required invariant per `.go-arch-lint.yml`. `ProcessedClone` might need `syntax.Node` for initial construction, even if it stores extracted primitives.
- If placed in `printer/`: Solves the immediate printer coupling problem. But `pkg/artdupl/` (SDK) would then depend on `printer/` or we'd need yet another DTO for the SDK. Three Clone types becomes two, not one.
- If placed in a new `core/` or `models/` package: Clean separation but adds a new package to the architecture.

**What I've tried:** Re-reading `AGENTS.md` notes says "Consolidation depends on Printer DTO change above." It doesn't specify where.

**What would unblock me:** A decision on:
1. Which package owns `ProcessedClone` (domain/, printer/, or new)
2. Whether `.go-arch-lint.yml` rule "domain must not import syntax" is flexible for DTO construction-time only
3. Whether `pkg/artdupl/` should use `ProcessedClone` directly or keep its own `Clone` type

---

## Metrics Summary

| Metric | Value |
|--------|-------|
| Clone groups detected (self-analysis) | **125** (down from 130) |
| Total Go lines | 45,035 |
| Non-test Go lines | 16,303 |
| Test Go lines | ~28,732 |
| Packages | 22 testable |
| BDD specs | 250/250 pass |
| Build status | **PASS** |
| Tests | **ALL PASS** |

### Coverage by Package

| Package | Coverage |
|---------|----------|
| pkg/format | **100.0%** |
| pkg/position | **100.0%** |
| hash | **96.6%** |
| internal/simd | **95.8%** |
| config | **94.8%** |
| internal/utils | **93.2%** |
| pkg/artdupl | **92.2%** |
| syntax/golang | **94.5%** |
| suffixtree | **91.0%** |
| syntax | **91.6%** |
| errors | **89.4%** |
| cache | **87.3%** |
| pkg/logger | **87.5%** |
| syntax/templ | **85.3%** |
| printer | **84.8%** |
| detection | **78.3%** |
| job | **76.7%** |
| cmd | **75.5%** |
| bdd | **70.0%** |
| internal/filtertest | **50.0%** |
| internal/configtest | [no statements] |
| examples | **0.0%** |
| domain | **0.0%** (NO TESTS) |
| cmd/art-dupl | **0.0%** (main only) |
| internal/testutil | [test helpers] |
| internal/testhelpers | [no tests] |

---

*End of report. Next expected action: Review Pareto priorities and decide on ProcessedClone DTO location.*
