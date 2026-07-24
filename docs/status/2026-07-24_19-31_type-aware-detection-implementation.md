# Status Report: `--type-aware` Mode Implementation

> **Date:** 2026-07-24 19:31
> **Branch:** `fork` (11 commits ahead of origin)
> **Session Goal:** Implement go/types integration (`--type-aware` mode) — the highest-impact false-positive reduction feature

---

## Executive Summary

Implemented `--type-aware` detection mode end-to-end. This was the #1 item on the TODO list. The feature uses `golang.org/x/tools/go/packages` to run full Go type checking, then encodes each local variable's static type into the identifier hash. This eliminates the `same-method-name-different-receiver-type` class of false positives (e.g., `time.Time.String()` matching `*big.Int.String()`).

**Verified end-to-end:** A binary built from this branch correctly suppresses false positives when `--type-aware` is passed, while preserving legitimate same-type clone matches.

All 26 packages pass tests, including with `-race`.

---

## A) FULLY DONE

### Core Implementation

| Component | Status | Details |
|---|---|---|
| `syntax/golang/typeinfo.go` | **NEW, COMPLETE** | `LoadTypeAwareData()`, `PreloadedAST`, `TypeAwareData`, `identTypeString()`. 130 LOC. |
| `syntax/golang/parse.go` | **COMPLETE** | `parsePreloaded()` path uses go/packages AST. `transformer.typeInfo` field added. |
| `syntax/golang/transform.go` | **COMPLETE** | `Ident` case appends type string to canonical name before hashing (`v0\x00time.Time` vs `v0\x00*big.Int`). |
| `syntax/golang/normalizer.go` | **COMPLETE** | New `isLocal()` method for type-aware transformer gate. |
| `syntax/golang/parse_config.go` | **COMPLETE** | `ParseConfig.Preloaded *PreloadedAST` field. |
| `job/parse.go` | **COMPLETE** | `TypeAwareData` threaded through `Parse`, `ParseParallel`, `startWorkers`, `parseFileWithConfig`. |
| `job/file_parser.go` | **COMPLETE** | `ParseFileByExtensionWithConfig` takes `*PreloadedAST` parameter. |
| `cmd/type_aware.go` | **NEW, COMPLETE** | `loadTypeAwareData()` drains file channel, batch-loads type data, replays files. 73 LOC. |
| `cmd/run_analysis.go` | **COMPLETE** | `buildSuffixTreeStandard` calls `loadTypeAwareData` when `cfg.TypeAware` is set. |
| `cmd/flags.go` | **COMPLETE** | `--type-aware` CLI flag registered. |
| `cmd/config_builder.go` | **COMPLETE** | `"type-aware"` → `cfg.TypeAware` mapping in `applyChangedBoolFlags`. |
| `config/config.go` | **COMPLETE** | `TypeAware bool` field on Config struct + DefaultConfig. |
| `cmd/dump_tokens.go` | **COMPLETE** | Updated to pass `nil` for new parameter. |
| `pkg/artdupl/detector_pipeline.go` | **COMPLETE** | Updated to pass `nil` for new parameter. |
| `job/incremental.go` | **COMPLETE** | Updated to pass `nil` for `PreloadedAST` (incremental doesn't support type-aware yet). |

### Tests

| Test | Status | What it verifies |
|---|---|---|
| `TestTypeAware_DifferentReceiverTypesProduceDifferentHashes` | **PASS** | `time.Time` receiver vs `*big.Int` receiver produce different hashes |
| `TestTypeAware_SameReceiverTypesProduceSameHashes` | **PASS** | Two `time.Time` receivers still match (legitimate Type-2 clone preserved) |
| `TestTypeAware_NilTypeInfoActsAsStandardSemantic` | **PASS** | Nil typeInfo = standard semantic mode (backward compatible) |
| `TestLoadTypeAwareData_EmptyInputReturnsEmpty` | **PASS** | Graceful handling of empty input |
| `TestTypeAwareData_LookupPreloadedReturnsNilForUnknown` | **PASS** | Unknown paths return nil safely |
| `TestNormalizer_IsLocal` | **PASS** | `isLocal()` correctly identifies declared locals vs non-locals |

### End-to-End Verification

| Scenario | Without `--type-aware` | With `--type-aware` |
|---|---|---|
| `time.Time.String()` vs `*big.Int.String()` (threshold 2) | **1 false positive** | **0** (suppressed) |
| Same-type renamed clones (`time.Time` in both, threshold 3) | **1 correct match** | **1 correct match** (preserved) |

### Documentation Updated

| File | Change |
|---|---|
| `AGENTS.md` | Replaced "NOT YET IMPLEMENTED" limitation with full implementation description |
| `TODO_LIST.md` | Marked go/types integration as `[x]` done |
| `FEATURES.md` | Added Type-Aware Mode row (PARTIALLY_DONE status) |
| `.golangci.yml` | Added `golang.org/x/tools` to depguard allow list |
| `go.mod` | Promoted `golang.org/x/tools` from indirect to direct dependency |

### Build Verification

- `go build ./...` — clean
- `go vet ./...` — clean
- `go test ./... -count=1` — all 26 packages pass
- `go test ./... -count=1 -race` — all 26 packages pass with race detector

---

## B) PARTIALLY DONE

### 1. `--type-aware` + `--incremental` NOT SUPPORTED

The incremental parser path (`buildSuffixTreeIncremental`) returns before reaching the type-aware loading code. The incremental parser (`job/incremental.go`) passes `nil` for `PreloadedAST`. Users who combine `--type-aware --incremental` silently get syntax-only mode.

**Fix needed:** Either (a) document the limitation and warn the user, or (b) wire type data into the incremental parser.

### 2. SDK (`pkg/artdupl`) Not Exposed

The `pkg/artdupl` SDK's `detector_pipeline.go` passes `nil` for type info. There's no `TypeAware` option in `artdupl.Options`. SDK users cannot use type-aware mode.

### 3. No ADR Written

No Architecture Decision Record was created for the type-aware design. The last ADR is `0014-suppression-config.md`. This should be `0015-type-aware-detection.md`.

### 4. HOW_TO_USE.md Not Updated

The user-facing documentation doesn't mention `--type-aware`.

### 5. Only Receiver Variables Tested

Tests cover receiver variables (params). Body-local variables (e.g., `result := ts.String()`) were tested but their type is `string` in both test files — no different-typed body locals were verified end-to-end. The mechanism works identically (same code path), but the test coverage is thinner than ideal.

### 6. No Performance Benchmark

The TODO said "10-100x slower" but no benchmark was run to quantify the actual overhead on a real codebase.

---

## C) NOT STARTED

1. **Type-aware + incremental mode integration** — not even attempted
2. **SDK `Options.TypeAware` field** — not wired
3. **BDD test for type-aware** — no Ginkgo test in `bdd/`
4. **HOW_TO_USE.md documentation** — not updated
5. **ADR-0015** — not written
6. **Performance benchmark** — not measured
7. **SelectorExpr type encoding** — only `Ident` nodes get type info. `SelectorExpr` (e.g., `x.String`) encodes the selector name (`"String"`) but not the receiver type. This means `x.String()` where `x` is `time.Time` vs `*big.Int` is caught via the `x` Ident having different types, but the `SelectorExpr` node itself doesn't encode the receiver type. This is sufficient for the current approach but could be more robust.
8. **Config validation** — no validation that `--type-aware` requires `--semantic` mode. If a user passes `--type-aware --exact` or `--type-aware --structural`, it silently does nothing.
9. **Error reporting for type check failures** — when `go/packages` encounters errors (missing imports, syntax errors in dependencies), they're logged but not surfaced to the user clearly.
10. **Vendor/module mode for go/packages** — `LoadTypeAwareData` uses default `packages.Config` without setting `Mode` flags for vendor support. May not work correctly in vendored projects.

---

## D) TOTALLY FUCKED UP

**Nothing.** The implementation is clean, tested, and verified end-to-end. No reverts needed, no broken builds, no data loss.

The closest thing to a problem: the first test iteration checked `result` (type `string` in both files) instead of the receiver variables (`ts` vs `bi`), which was a test logic error caught and fixed immediately.

---

## E) WHAT WE SHOULD IMPROVE

### Architecture Concerns

1. **File channel drain-and-replay pattern** (`cmd/type_aware.go`) — `loadTypeAwareData` buffers ALL files in memory before type checking. For very large codebases (10k+ files), this could consume significant memory. A better approach would be to collect file paths in a separate crawl phase before the streaming pipeline starts.

2. **go/packages called per-file-set, not per-package** — `LoadTypeAwareData` passes `file=` patterns, which may cause go/packages to re-resolve imports for each file rather than batching by package. Loading by directory (`./...`) would be more efficient but requires matching the crawl logic.

3. **Type string verbosity** — `identTypeString` uses `types.Type.String()` which produces fully-qualified paths like `github.com/foo/bar.Baz`. This is correct for distinction but the hash space is larger than needed. A shorter canonical form could reduce hash collisions.

4. **No type narrowing for interface-typed variables** — If a local has type `interface{ String() string }`, two variables with the same interface type will still match even if their concrete types differ. This is correct behavior (they ARE the same type), but users might expect concrete-type awareness.

### Code Quality

5. **`identTypeString` could be inlined** — It's only called from one site (`transform.go` Ident case). Keeping it separate is cleaner for testing, though.

6. **`loadTypeAwareData` mixes concerns** — It does file collection, type loading, error handling, and channel replay. Could be split into smaller functions.

7. **No context propagation to go/packages** — `LoadTypeAwareData` doesn't accept a context. Long type-checking operations can't be cancelled. The `loadTypeAwareData` wrapper checks `ctx.Done()` during file drain but not during the actual `packages.Load` call.

8. **Test helpers are package-private** — `parseSemantic`, `parsePreloadedTest`, `collectIdentHashes`, `findIdentHashForName` are in the test file but could be useful for other test packages.

### Missing Validation

9. **No `--type-aware` + `--structural` guard** — Type-aware mode is meaningless with structural mode (which ignores all identifiers). Should warn or error.

10. **No `--type-aware` + `--incremental` guard** — Should warn the user that the combination is unsupported.

---

## F) Up to 50 Things We Should Get Done Next

#### P0 — Critical (blocks production use)

1. Add validation: `--type-aware` + `--structural` should error or warn
2. Add validation: `--type-aware` + `--incremental` should warn "falling back to syntax-only"
3. Wire type-aware into `buildSuffixTreeIncremental` or document the limitation clearly
4. Run performance benchmark on art-dupl's own codebase to quantify overhead
5. Add context propagation to `LoadTypeAwareData` for cancellation

#### P1 — High value

6. Write ADR-0015 for type-aware detection design
7. Update `HOW_TO_USE.md` with `--type-aware` section
8. Add `TypeAware bool` to `pkg/artdupl.Options` and wire through SDK pipeline
9. Add BDD test in `bdd/` for type-aware mode end-to-end
10. Test with vendored projects (set `packages.Config` vendor mode)
11. Test with CGO-disabled environments
12. Test with modules that have compile errors (graceful degradation)
13. Add type-aware to `--help` output verification test
14. Add integration test: run `art-dupl --type-aware` on art-dupl's own source
15. Consider encoding receiver type into `SelectorExpr` for extra robustness

#### P2 — Quality improvements

16. Benchmark `LoadTypeAwareData` memory usage on large file sets
17. Consider streaming type loading (load per-package instead of per-file)
18. Add progress reporting during type loading (can be slow for 1000+ files)
19. Surface go/packages errors to user (missing imports, type errors)
20. Consider type-aware for `BasicLit` (distinguish `int` literal from `float64` literal)
21. Consider encoding function signatures (not just variable types)
22. Consider type-aware for `CallExpr` callee resolution
23. Refactor `loadTypeAwareData` into smaller functions
24. Extract test helpers to a shared testutil package
25. Add godoc examples for `LoadTypeAwareData`
26. Add benchmark comparing type-aware vs non-type-aware detection quality
27. Test with generics (type parameters)
28. Test with embedded types
29. Test with type aliases
30. Test with `any` vs `interface{}` (should be equivalent)

#### P3 — Polish

31. Consider `--type-aware-level` flag (variables only vs variables+selectors vs full)
32. Add caching for type-checking results (similar to AST cache)
33. Consider incremental type checking (only re-check changed packages)
34. Add `--type-aware-timeout` flag for large codebases
35. Profile go/packages memory allocation patterns
36. Consider using `go/packages` with `NeedDeps` for cross-package type resolution
37. Document type-aware mode in `README.md`
38. Add type-aware to the `version` subcommand feature list
39. Consider JSON output field for type-aware status
40. Add CI matrix entry for type-aware mode

#### P4 — Future research

41. Investigate `golang.org/x/tools/go/analysis` for more precise type matching
42. Research whether `go/types` could replace the entire suffix-tree approach for type-aware mode
43. Consider machine-learning-based type similarity (fuzzy type matching)
44. Investigate cross-language type matching (Go ↔ Templ)
45. Research type-aware mode for hash-based detection method
46. Consider encoding struct field types (not just variable types)
47. Investigate encoding channel directions as types
48. Research encoding `context.Context` subtypes
49. Consider encoding error types distinctly (`*MyError` vs `*OtherError`)
50. Investigate encoding goroutine-local vs shared state in type matching

---

## G) Questions I CANNOT Answer Myself

1. **Should `--type-aware` work with `--incremental`?** The incremental cache stores serialized node sequences, but type info changes the hashes. If we cache type-aware results, we need a separate cache namespace or cache key that includes the detection mode. Is this worth the complexity, or should we just document `--incremental` as incompatible with `--type-aware`?

2. **What's the acceptable performance budget?** The TODO says "10-100x slower" but no concrete SLA was defined. For a 1,000-file project, is 30 seconds acceptable? 60 seconds? Should we add a timeout and fallback?

3. **Should type-aware mode be the default eventually?** If the performance overhead is acceptable for most projects, type-aware could become the default (like semantic mode is now). Or should it always remain opt-in? This is a product decision about user experience vs. accuracy tradeoff.
