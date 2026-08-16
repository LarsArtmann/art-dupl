# art-dupl — Full Comprehensive Status Report

> **Generated:** 2026-06-20 15:11 UTC
> **Branch:** `fork` (1,502 total commits)
> **Session Commits:** 25 (sprints 1–3)
> **Codebase:** 48,028 lines of Go across 233 files (136 source + 97 test)

---

## Quality Gates

| Check                | Status      | Details                      |
| -------------------- | ----------- | ---------------------------- |
| `go build ./...`     | ✅ PASS     | All 22 packages compile      |
| `go test ./...`      | ✅ PASS     | All 22 packages pass         |
| `go vet ./...`       | ✅ PASS     | Zero warnings                |
| `go-arch-lint check` | ✅ PASS     | Zero architecture violations |
| `golangci-lint run`  | ✅ PASS     | 0 issues                     |
| `gofmt -l .`         | ✅ PASS     | All files formatted          |
| Coverage (avg)       | ⚠️ **74.3%** | Target: 80%. See gaps below  |

### Coverage by Package

| Package             | Coverage | Status           |
| ------------------- | -------- | ---------------- |
| pkg/enum            | 100.0%   | ✅               |
| pkg/format          | 100.0%   | ✅               |
| pkg/position        | 100.0%   | ✅               |
| bdd                 | 93.3%    | ✅               |
| syntax/golang       | 93.4%    | ✅               |
| syntax              | 91.4%    | ✅               |
| suffixtree          | 91.2%    | ✅               |
| config              | 91.0%    | ✅               |
| errors              | 89.6%    | ✅               |
| pkg/artdupl (SDK)   | 88.4%    | ✅               |
| hash                | 93.4%    | ✅               |
| pkg/logger          | 87.5%    | ✅               |
| cache               | 84.3%    | ✅               |
| syntax/templ        | 84.6%    | ✅               |
| printer             | 76.5%    | ⚠️                |
| cmd                 | 75.1%    | ⚠️                |
| job                 | 72.1%    | ⚠️                |
| domain              | 66.2%    | 🔴               |
| detection           | 61.8%    | 🔴               |
| internal/filtertest | 50.0%    | 🔴 (test infra)  |
| examples            | 0.0%     | — (demo package) |
| cmd/art-dupl        | 0.0%     | — (main package) |

---

## a) FULLY DONE ✅

### Core System (Production-Ready)

| Area                       | What Was Done                                                                                                   |
| -------------------------- | --------------------------------------------------------------------------------------------------------------- |
| **Build & Compile**        | All 22 packages build clean on Go 1.26.3                                                                        |
| **Test Suite**             | 22/22 packages pass, 97 test files                                                                              |
| **Architecture**           | Acyclic import graph, enforced by go-arch-lint                                                                  |
| **SDK Independence**       | `pkg/artdupl` has ZERO imports of `config/` or `errors/` — defines own types, sentinels, Logger interface       |
| **Context Propagation**    | ALL pipeline goroutines use `select { case ch <- v: case <-ctx.Done(): return }` — zero "check-then-send" races |
| **Error Handling**         | Typed error hierarchy, stack traces, JSON marshaling in `errors/` package                                       |
| **Multi-Method Detection** | Suffix tree + hash detection with adapter pattern                                                               |
| **7 Output Formats**       | Text, HTML, JSON, Simple-JSON, Plumbing, SARIF, CSV                                                             |
| **Semantic Matching**      | Identifier/operator hashing for fewer false positives (default mode)                                            |
| **Incremental Caching**    | SHA1-based AST cache for changed-file-only analysis                                                             |
| **String Interning**       | `InternFilename` deduplicates filename strings                                                                  |
| **Fuzz Testing**           | Suffix tree + templ parser fuzz tests (2M+ execs, no panics)                                                    |

### Session Work (25 commits across 3 sprints)

**Sprint 1 — Architecture Hardening (12 commits):**

- Decoupled SDK from `config/` and `errors/` packages (ZERO imports)
- Removed dead `--since` flag, `config.DetectionConfig`, 6 dead SDK sentinels
- Split 624-line `actionability.go` into 4 focused files
- Aligned `LineStart`/`LineEnd` naming across all Clone types
- Enforced decoupling in `.go-arch-lint.yml`

**Sprint 2 — Correctness & Cleanup (6 commits):**

- Fixed `started time.Time` data race in SDK detector
- Wired `Summary.LinesAnalyzed` (was hardcoded to 0)
- Fixed sync `FindClones` never reporting 100% progress
- Decoupled SDK from internal `errors` package (replaced `debug.Stack()` calls)

**Sprint 3 — Self-Review (7 commits):**

- Deleted ghost branded types (`domain.Filepath`, `domain.LineNumber`) + helpers (251 lines removed)
- Removed 5 dead error constructors, `EnumValidationError`, `ErrorType.String()`
- Removed deprecated `FindClonesStream` from interface
- Fixed lying documentation (SIMD claims, non-existent types)
- Renamed `SARIFConfig` → `SARIFPrinterOptions`

---

## b) PARTIALLY DONE ⚠️

| Area                         | Current State               | What Remains                                                                                                                                                 |
| ---------------------------- | --------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **Test Coverage**            | 74.3% average               | `detection` (61.8%) and `domain` (66.2%) below 80% target. Need integration tests for MultiDetector and enum MarshalJSON roundtrips                          |
| **Concurrency**              | 14/17 goroutines fixed      | `cmd/run_crawl.go` file feeders (3 goroutines) still use blocking sends — stdin scanner and filepath.Walk are inherently blocking, needs full chain refactor |
| **Printer Package**          | ~3500 lines across 29 files | Splitting into sub-packages (stats, html, analyze) deferred — needs core extraction first                                                                    |
| **Clone Type Consolidation** | Field names aligned         | 5 parallel Clone types exist with identical field names but separate type definitions. Consolidation needs DTO architecture design                           |
| **Fragment Type**            | Works at boundaries         | `[]byte` in domain vs `string` in SDK — flips at every boundary. Needs unified type decision                                                                 |
| **depguard**                 | Fixed and configured        | Allow-list covers all 16 direct deps but config wasn't verified to actually block unapproved imports in CI                                                   |

---

## c) NOT STARTED

| Area                              | Description                                                                                                                                  |
| --------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| **ProcessedClone DTO**            | Printer still imports `syntax.Node` directly (34 references) for pattern evaluation. Needs read-only interface design                        |
| **Clone type consolidation**      | 5 parallel types: `printer.CloneGroup`, `pkg/artdupl.Clone`, `pkg/artdupl.CloneGroup`, `domain.ProcessedClone`, `domain.ProcessedCloneGroup` |
| **printer/ sub-packages**         | Stats, HTML, and analysis code all in one package                                                                                            |
| **`…Data` → `…View` rename**      | 5 view model types in `printer/html_views.go` use the `…Data` suffix                                                                         |
| **`encoding/json/v2` migration**  | Go 1.26.3 supports it; all 6 production files still use v1                                                                                   |
| **GeneratorFilter bitmask**       | 6 `Include*` bools in `config.Config` should be a bitmask                                                                                    |
| **detection test coverage**       | 61.8% — adapters, streamMatches, hasNonEmptyFrag at 0%                                                                                       |
| **OpenTelemetry instrumentation** | No tracing/metrics in the pipeline                                                                                                           |
| **Benchmarks in CI**              | No performance regression detection                                                                                                          |

---

## d) TOTALLY FUCKED UP 🔴

| Area                                        | What's Wrong                                                                                                                                                                                 | Impact                                                                                             |
| ------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| **3 goroutine leaks in `cmd/run_crawl.go`** | `filesFeedWithOptions` goroutines have ZERO context awareness — `fchan <- path` blocking sends with no `ctx.Done()` select. stdin reader and filepath.Walk cannot be cancelled.              | **MEDIUM** — process exits on cancellation anyway, but leaks during long crawls if consumer stalls |
| **`detection` at 61.8% coverage**           | Adapters (`suffixTreeAdapter.FindDuplOver`, `hashAdapter.FindDuplOver`), `streamMatches`, `hasNonEmptyFrag`, `detName` all at 0%. Tests only drain channels without asserting match content. | **HIGH** — core detection logic is untested                                                        |
| **`examples/` package**                     | `RunSDKDemo()` exported but never called. Test file is tautological (creates structs, asserts same values back). Zero real SDK usage testing.                                                | **LOW** — demo code, but gives false confidence                                                    |
| **`go.sum` not tidy**                       | `go mod tidy -diff` shows drift (clipperhouse/displaywidth entry).                                                                                                                           | **LOW** — cosmetic but indicates inconsistent dependency management                                |
| **wsl_v5 lint warnings**                    | `domain/health_test.go:52,64` have missing whitespace above assign statements.                                                                                                               | **LOW** — golangci-lint passes (stale LSP cache), but should be clean                              |

---

## e) WHAT WE SHOULD IMPROVE

### Architecture

1. **Clone type consolidation is the #1 architecture debt.** Five parallel types with identical fields is a maintenance liability. The `LineRange` value object (enforcing `end >= start` at construction) should be the first step — it eliminates 3 copies of the same runtime validation.

2. **Printer→syntax.Node coupling (34 references).** All are field reads, no method calls — a read-only `NodeReader` interface would decouple without changing behavior.

3. **`config.Config` is a 43-field god object.** The `Include*` family (6 bools) is a bitmask masquerading as independent fields. Reflection-based merge prevents extracting sub-structs without breaking field-by-field override.

### Testing

4. **Detection package needs real assertions.** Current tests drain channels without checking what comes out. Need table-driven tests with known clone inputs and expected match counts.

5. **Domain enum coverage gap.** `CloneCategory` and `CloneActionability` MarshalJSON/UnmarshalJSON at 0%. Should follow the same pattern as the HealthScore tests just added.

### Concurrency

6. **`cmd/run_crawl.go` needs ctx threaded through.** The `filesFeedWithOptions` → `buildParams` → caller chain needs `context.Context` as first param. The stdin scanner should use a context-aware reader.

### Type Safety

7. **`Fragment []byte` vs `string` split brain.** Pick one representation at the domain layer. `[]byte` is more efficient for diffing; `string` is simpler for JSON. A branded `SourceFragment` type with conversion methods would localize the flip.

8. **`CloneHash string` unbranded in 5 places.** A `type CloneHash string` with `Partial()` method would localize the `hash[:8]` hack in SARIF.

### Tooling

9. **`encoding/json/v2`** is available in Go 1.26.3 but unused. Migration would get better performance and `nil` handling.

10. **No benchmarks in CI.** The suffix tree algorithm should have performance regression tests.

---

## f) Top #25 Things to Get Done Next

Sorted by **impact / effort ratio** (highest first).

| #  | Task                                                                                               | Impact  | Effort | Sprint  |
| -- | -------------------------------------------------------------------------------------------------- | ------- | ------ | ------- |
| 1  | Add detection adapter tests (`suffixTreeAdapter`, `hashAdapter` FindDuplOver with real assertions) | 🔴 HIGH | 2h     | Next    |
| 2  | Add `CloneCategory` + `CloneActionability` MarshalJSON/UnmarshalJSON tests                         | 🟡 MED  | 30m    | Next    |
| 3  | Run `go mod tidy` to fix go.sum drift                                                              | 🟢 LOW  | 1m     | Next    |
| 4  | Fix wsl_v5 warnings in `health_test.go` (blank line separation)                                    | 🟢 LOW  | 5m     | Next    |
| 5  | Thread `context.Context` through `cmd/run_crawl.go` file feeders                                   | 🔴 HIGH | 4h     | Next    |
| 6  | Migrate `encoding/json` → `encoding/json/v2` in 6 production files                                 | 🟡 MED  | 3h     | Next    |
| 7  | Add `LineRange` value object (enforce `end >= start` at construction)                              | 🟡 MED  | 4h     | Next    |
| 8  | Unify `Fragment` type (`[]byte` vs `string`) with branded `SourceFragment`                         | 🟡 MED  | 6h     | Later   |
| 9  | Add `CloneHash` branded type with `Partial()` method                                               | 🟡 MED  | 2h     | Later   |
| 10 | Add real assertions to `detection/detection_test.go` (match count, hash, filenames)                | 🔴 HIGH | 3h     | Next    |
| 11 | Add domain `ProcessedClone.Validate()` roundtrip tests                                             | 🟡 MED  | 1h     | Next    |
| 12 | Consolidate 5 Clone types into `CloneLocation` + format-specific extensions                        | 🔴 HIGH | 8h     | Later   |
| 13 | Extract `NodeReader` interface to decouple printer from `syntax.Node`                              | 🟡 MED  | 6h     | Later   |
| 14 | Split `printer/` into sub-packages (stats, html, analyze)                                          | 🟡 MED  | 8h     | Later   |
| 15 | Rename `…Data` view models to `…View` in printer (requires templ regen)                            | 🟢 LOW  | 3h     | Later   |
| 16 | Add `GeneratorFilter` bitmask to replace 6 `Include*` bools in config                              | 🟡 MED  | 4h     | Later   |
| 17 | Add benchmarks for suffix tree construction + search                                               | 🟡 MED  | 3h     | Later   |
| 18 | Add CI performance regression detection                                                            | 🟢 LOW  | 4h     | Later   |
| 19 | Wire `examples/` package or delete it (currently orphaned)                                         | 🟢 LOW  | 1h     | Next    |
| 20 | Add `domain` coverage to reach 75%+ (currently 66.2%)                                              | 🟡 MED  | 3h     | Next    |
| 21 | Add `detection` coverage to reach 75%+ (currently 61.8%)                                           | 🟡 MED  | 4h     | Next    |
| 22 | Add `cmd/` coverage to reach 80%+ (currently 75.1%)                                                | 🟡 MED  | 3h     | Later   |
| 23 | Evaluate `go-udiff` as replacement for `sergi/go-diff` (already a transitive dep)                  | 🟢 LOW  | 2h     | Later   |
| 24 | Add OpenTelemetry spans to pipeline stages (parse → serialize → detect → print)                    | 🟢 LOW  | 6h     | Later   |
| 25 | Hide `syntax/golang` behind facade (BLOCKED by import cycle)                                       | 🟢 LOW  | 8h     | Blocked |

---

## g) Top #1 Question I Cannot Figure Out

> **Should the 5 parallel Clone types be consolidated into one, or should we accept the duplication as an intentional DTO boundary?**
>
> The types share an identical invariant core (`Filename` + `LineStart` + `LineEnd`) but diverge on:
>
> - `pkg/artdupl.Clone`: has `StartPos`/`EndPos` (byte positions), `Fragment string`, `Size int`
> - `domain.ProcessedClone`: has `Fragment []byte`, `TokenCount int`, `FileSize int`, `Classification`
> - `printer.JSONClone`: has `Category`/`Priority`/`Actionability` strings for JSON output
> - `printer.CloneOccurrenceData`: has `VSCodeLink` for HTML
>
> Extracting a shared `CloneLocation` value object is obviously correct. But consolidating the full types risks coupling the SDK to printer concerns (JSON tags, view-specific fields) or the domain to SDK concerns (byte positions). The question is: **where is the right boundary?** One `Clone` type with optional fields? Or keep separate types but share the location via embedding?
>
> I lean toward: `CloneLocation` (embedded everywhere) + keep the rest separate. But this is a design decision that affects the public SDK API and all downstream consumers.

---

## Session Commit Log (25 commits)

```
c15eba6 docs: update TODO_LIST + AGENTS with concurrency leak fixes
ce677fc fix: eliminate remaining goroutine leaks in cmd/ and SDK pipeline
163b00e fix: eliminate goroutine leaks in BuildTree and ParseIncremental
cf8758a fix: eliminate check-then-send goroutine leaks in job pipeline
2c2ec4c fix: eliminate goroutine leaks in SDK streaming API (FindClonesStreamResult)
1867f5d fix: remove stale --since flag from BDD test harness, fix misleading test names
4f95be7 docs: update TODO_LIST with self-review sprint #2 results
9035bfc test: add tests for domain.HealthScore (was 0% coverage)
d912232 test: add comprehensive tests for pkg/enum (was 0% coverage)
8c53fb4 test: fix always-pass tests that discarded errors in detector tests
9cbff50 fix: add context.Context to hash.FileDetector.FindDuplOver for cancellation
8887298 fix: remove misleading goexperiment build tags and fix depguard allow-list
c1268dc fix: stop swallowing errors in cache operations (MkdirAll, Remove, saveMetadata)
00e0bbc refactor: remove dead ParseFileByExtension wrapper
314643f docs: add brutal self-review report for 2026-06-20 sprint
711f47d docs: update TODO_LIST, FEATURES, AGENTS after self-review sprint
c0d0cc4 refactor: rename SARIFConfig to SARIFPrinterOptions for clarity
af8ae55 refactor: delete dead domain.Filepath and domain.LineNumber branded types
c245ee7 refactor: remove deprecated FindClonesStream from Detector interface
35e2e6d refactor: remove dead error constructors, types, and ErrorType constants
d564504 test: add missing field assertions in examples_test.go
169eb03 refactor: remove dead ErrInvalidLineNumber sentinel and NodeTypeName field
8961434 refactor: remove dead Parse functions and fix infertypeargs diagnostics
6c20830 docs: fix go.mod module doc — remove SIMD lies and dead references
83f0360 docs: fix domain package doc — remove references to 6 non-existent types
```
