# art-dupl — Full Comprehensive Status Report

> **Live remainder:** archived 2026-11-14 (cadence wave); every item in this report is struck or routed, and open work is owned by TODO_LIST.md and ROADMAP.md.

> **Resolution (2026-10-05):** ~~Open items unresolved at write time.~~ Audited 2026-10-05 docs-health pass: 17 items verified resolved, 8 open (coverage 92.7%, leaks fixed, subpackages shipped; facade/consolidation deferred).

> **Generated:** 2026-06-20 15:11 UTC
> **Branch:** `fork` (1,502 total commits)
> **Session Commits:** 25 (sprints 1–3)
> **Codebase:** 48,028 lines of Go across 233 files (136 source + 97 test)

---

## Quality Gates

<!-- check-rows:ledger -->

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

<!-- check-rows:ledger -->

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
| domain              | ~~66.2%~~ | ~~🔴~~ superseded — 2026-06 snapshot; coverage is tracked by the CI Coverage job now |
| detection           | ~~61.8%~~ | ~~🔴~~ superseded — 2026-06 snapshot; coverage is tracked by the CI Coverage job now |
| internal/filtertest | 50.0%    | 🔴 (test infra)  |
| examples            | 0.0%     | — (demo package) |
| cmd/art-dupl        | 0.0%     | — (main package) |

---

## a) FULLY DONE ✅

### Core System (Production-Ready)

<!-- check-rows:ledger -->

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

| Area                         | Current State               | What Remains                                                                                                                        |
| ---------------------------- | --------------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| ~~**Test Coverage**~~ done — resolved — coverage is CI-tracked (Coverage job); detection and domain are long past the 2026-06 61.8/66.2 numbers | ~~74.3% average~~ | ~~`detection` (61.8%) and `domain` (66.2%) below 80% target. Need integration tests for MultiDetector and enum MarshalJSON roundtrips~~ |
| ~~**Concurrency**~~ done — resolved — every pipeline goroutine is ctx-aware (AGENTS.md context-propagation invariant); the crawl leaks were fixed the same week | ~~14/17 goroutines fixed~~ |
| ~~**Printer Package**~~ done — resolved — printer split landed: printer/stats, printer/actionability, printer/finding | ~~~3500 lines across 29 files~~ | ~~Splitting into sub-packages (stats, html, analyze) deferred — needs core extraction first~~ |
| ~~**Clone Type Consolidation**~~ done — resolved — domain.CloneRef embedded in every clone-bearing type (2026-07-01 consolidation) | ~~Field names aligned~~ | ~~5 parallel Clone types exist with identical field names but separate type definitions. Consolidation needs DTO architecture design~~ |
| ~~**Fragment Type**~~ done — resolved — Fragment is string everywhere (ADR-0005) | ~~Works at boundaries~~ | ~~`[]byte` in domain vs `string` in SDK — flips at every boundary. Needs unified type decision~~ |
| ~~**depguard**~~ done — resolved — import boundaries enforced by go-arch-lint in CI (arch-lint gate) | ~~Fixed and configured~~ | ~~Allow-list covers all 16 direct deps but config wasn't verified to actually block unapproved imports in CI~~ |

---

## c) NOT STARTED

| Area                              | Description                                                                                                                                  |
| --------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| ~~**ProcessedClone DTO**~~ done — resolved — serializable domain.CloneNode DTO bridges printer pattern evaluation (ADR-0006) | ~~Printer still imports `syntax.Node` directly (34 references) for pattern evaluation. Needs read-only interface design~~ |
| ~~**Clone type consolidation**~~ done — CloneRef consolidation shipped 2026-07-01 | ~~5 parallel types: `printer.CloneGroup`, `pkg/artdupl.Clone`, `pkg/artdupl.CloneGroup`, `domain.ProcessedClone`, `domain.ProcessedCloneGroup`~~ |
| ~~**printer/ sub-packages**~~ done — printer/stats + printer/actionability + printer/finding sub-packages exist | ~~Stats, HTML, and analysis code all in one package~~ |
| ~~**`…Data` → `…View` rename**~~ done — CloneGroupView views with stable anchors (printer/html_views.go) | ~~5 view model types in `printer/html_views.go` use the `…Data` suffix~~ |
| ~~**`encoding/json/v2` migration**~~ **NOT-DO — decided against — ADR-0024 pins the stable v1 API; Go 1.27 dropped the v2 struct-tag grammar and the jsonv2gate bans production v2 imports.** | ~~Go 1.26.3 supports it; all 6 production files still use v1~~ |
| ~~**GeneratorFilter bitmask**~~ **Superseded — superseded — landed as the gogenfilter v3 integration: generatorIncludes keyed by gogenfilter.FilterReason (cmd/filter_stats.go).** | ~~6 `Include*` bools in `config.Config` should be a bitmask~~ |
| ~~**detection test coverage**~~ done — detection adapter tests with real assertions; coverage CI-tracked | ~~61.8% — adapters, streamMatches, hasNonEmptyFrag at 0%~~ |
| ~~**OpenTelemetry instrumentation**~~ **Won't implement — no tracing demand surfaced since; per-run cache stats and progress output cover observability.** | ~~No tracing/metrics in the pipeline~~ |
| ~~**Benchmarks in CI**~~ done — CI Performance Tests workflow + flake alloc-gate (scripts/check-alloc-regression.sh) | ~~No performance regression detection~~ |

---

## d) TOTALLY FUCKED UP 🔴

| Area                                        | What's Wrong                                                                                                                                                                                 | Impact                                                                                             |
| ------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- |
| ~~**3 goroutine leaks in `cmd/run_crawl.go`**~~ done — resolved — ctx threaded through the crawl feeders the same week (cmd/run_crawl.go select guards) | ~~`filesFeedWithOptions` goroutines have ZERO context awareness — `fchan <- path` blocking sends with no `ctx.Done()` select. stdin reader and filepath.Walk cannot be cancelled.~~ | ~~**MEDIUM** — process exits on cancellation anyway, but leaks during long crawls if consumer stalls~~ |
| ~~**`detection` at 61.8% coverage**~~ done — resolved — adapter tests with real assertions shipped; coverage is CI-tracked | ~~Adapters (`suffixTreeAdapter.FindDuplOver`, `hashAdapter.FindDuplOver`), `streamMatches`, `hasNonEmptyFrag`, `detName` all at 0%. Tests only drain channels without asserting match content.~~ | ~~**HIGH** — core detection logic is untested~~ |
| ~~**`examples/` package**~~ **Won't implement — examples/ stays as the SDK demo; crawls exclude it by default.** | ~~`RunSDKDemo()` exported but never called. Test file is tautological (creates structs, asserts same values back). Zero real SDK usage testing.~~ | ~~**LOW** — demo code, but gives false confidence~~ |
| ~~**`go.sum` not tidy**~~ done — resolved — dependency management rides the BuildFlow sweeps; go.sum is tidy | ~~`go mod tidy -diff` shows drift (clipperhouse/displaywidth entry).~~ | ~~**LOW** — cosmetic but indicates inconsistent dependency management~~ |
| ~~**wsl_v5 lint warnings**~~ done — resolved — lint sits at 0 findings across 106 linters (2026-09-29 baseline) | ~~`domain/health_test.go:52,64` have missing whitespace above assign statements.~~ | ~~**LOW** — golangci-lint passes (stale LSP cache), but should be clean~~ |

---

## e) WHAT WE SHOULD IMPROVE

### Architecture

~~1. **Clone type consolidation is the #1 architecture debt.** Five parallel types with identical fields is a maintenance liability. The `LineRange` value object (enforcing `end >= start` at construction) should be the first step — it eliminates 3 copies of the same runtime validation.~~ resolved by alternative — CloneRef (Filename/LineStart/LineEnd + LineCount()) consolidated instead (0.5.0)

~~2. **Printer→syntax.Node coupling (34 references).** All are field reads, no method calls — a read-only `NodeReader` interface would decouple without changing behavior.~~ resolved by alternative — actionability decoupled via serializable CloneNode DTO (ADR-0006)

3. **`config.Config` is a 43-field god object.** The `Include*` family (6 bools) is a bitmask masquerading as independent fields. Reflection-based merge prevents extracting sub-structs without breaking field-by-field override. → routed — long-term refactor idea tracked on ROADMAP (config.Config god object row, docs-health harvest 2026-11-14); the Include-family bitmask never happened

### Testing

~~4. **Detection package needs real assertions.** Current tests drain channels without checking what comes out. Need table-driven tests with known clone inputs and expected match counts.~~ done — detection/coverage_test.go adapter tests with real assertions

~~5. **Domain enum coverage gap.** `CloneCategory` and `CloneActionability` MarshalJSON/UnmarshalJSON at 0%. Should follow the same pattern as the HealthScore tests just added.~~ done — domain/enums_test.go MarshalJSON tests incl. invalid cases

### Concurrency

~~6. **`cmd/run_crawl.go` needs ctx threaded through.** The `filesFeedWithOptions` → `buildParams` → caller chain needs `context.Context` as first param. The stdin scanner should use a context-aware reader.~~ done — ctx.Done() selects in cmd/run_crawl.go feeders

### Type Safety

~~7. **`Fragment []byte` vs `string` split brain.** Pick one representation at the domain layer. `[]byte` is more efficient for diffing; `string` is simpler for JSON. A branded `SourceFragment` type with conversion methods would localize the flip.~~ resolved by alternative — single CloneRef.Fragment string embedded everywhere

8. ~~**`CloneHash string` unbranded in 5 places.** A `type CloneHash string` with `Partial()` method would localize the `hash[:8]` hack in SARIF.~~ superseded — clone identity is the XXH3 group hash surfaced through the printer GroupID contract (ADR-0028); a branded CloneHash type was never adopted

### Tooling

~~9. **`encoding/json/v2`** is available in Go 1.26.3 but unused. Migration would get better performance and `nil` handling.~~ resolved by alternative — v2 abandoned (Go 1.27 dropped v2 tags); v1 API behind internal/jsonutil (ADR-0024)

~~10. **No benchmarks in CI.** The suffix tree algorithm should have performance regression tests.~~ done — alloc benchmark gates + pinned protocol (0.7.1)

---

## f) Top #25 Things to Get Done Next

Sorted by **impact / effort ratio** (highest first).

| #  | Task                                                                              | Impact                                                                                             | Effort  | Sprint  |
| -- | --------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------- | ------- | ------- |
| ~~1~~ done — detection adapter tests with real assertions (detection/coverage_test.go) | ~~Add detection adapter tests (`suffixTreeAdapter`, `hashAdapter` FindDuplOver with real assertions)~~ | ~~🔴 HIGH~~ | ~~2h~~ |
| ~~2~~ done — domain enum MarshalJSON tests incl. invalid cases | ~~Add `CloneCategory` + `CloneActionability` MarshalJSON/UnmarshalJSON tests~~ | ~~🟡 MED~~ | ~~30m~~ |
| ~~3~~ done — go.sum managed by the BuildFlow dependency sweeps | ~~Run `go mod tidy` to fix go.sum drift~~ | ~~🟢 LOW~~ | ~~1m~~ |
| ~~4~~ done — lint at 0 findings since 2026-09-29 | ~~Fix wsl_v5 warnings in `health_test.go` (blank line separation)~~ | ~~🟢 LOW~~ | ~~5m~~ |
| ~~5~~ done — feeders take ctx and select on ctx.Done() (cmd/run_crawl.go) | ~~Thread `context.Context` through `cmd/run_crawl.go` file feeders~~ | ~~🔴 HIGH~~ | ~~4h~~ | ~~Next~~ |
| ~~6~~ **NOT-DO — decided against — v1-only invariant (ADR-0024); jsonv2gate bans production v2 imports.** | ~~Migrate `encoding/json` → `encoding/json/v2` in 6 production files~~ | ~~🟡 MED~~ | ~~3h~~ |
| ~~7~~ done — resolved — CloneRef + the ErrLineEndBeforeStart sentinel carry the line-range invariant | ~~Add `LineRange` value object (enforce `end >= start` at construction)~~ | ~~🟡 MED~~ | ~~4h~~ |
| ~~8~~ done — resolved — Fragment is string everywhere (ADR-0005) | ~~Unify `Fragment` type (`[]byte` vs `string`) with branded `SourceFragment`~~ | ~~🟡 MED~~ | ~~6h~~ |
| ~~9~~ **Superseded — superseded — clone identity is the XXH3 group hash surfaced via the printer GroupID contract (ADR-0028); a branded CloneHash type was never adopted.** | ~~Add `CloneHash` branded type with `Partial()` method~~ | ~~🟡 MED~~ | ~~2h~~ | ~~Later~~ |
| ~~10~~ done — detection adapter assertions shipped | ~~Add real assertions to `detection/detection_test.go` (match count, hash, filenames)~~ | ~~🔴 HIGH~~ | ~~3h~~ |
| ~~11~~ done — domain validation sentinels + cross-package alias tests | ~~Add domain `ProcessedClone.Validate()` roundtrip tests~~ | ~~🟡 MED~~ | ~~1h~~ |
| ~~12~~ done — CloneRef consolidation (2026-07-01) | ~~Consolidate 5 Clone types into `CloneLocation` + format-specific extensions~~ | ~~🔴 HIGH~~ | ~~8h~~ |
| ~~13~~ **Superseded — superseded — decoupled via the serializable CloneNode DTO (ADR-0006) instead of a NodeReader interface.** | ~~Extract `NodeReader` interface to decouple printer from `syntax.Node`~~ | ~~🟡 MED~~ | ~~6h~~ |
| ~~14~~ done — printer/stats + printer/actionability + printer/finding | ~~Split `printer/` into sub-packages (stats, html, analyze)~~ | ~~🟡 MED~~ | ~~8h~~ |
| ~~15~~ done — CloneGroupView naming (printer/html_views.go) | ~~Rename `…Data` view models to `…View` in printer (requires templ regen)~~ | ~~🟢 LOW~~ | ~~3h~~ |
| ~~16~~ **Superseded — superseded — generator filtering landed as gogenfilter integration: generatorIncludes keyed by gogenfilter.FilterReason.** | ~~Add `GeneratorFilter` bitmask to replace 6 `Include*` bools in config~~ | ~~🟡 MED~~ | ~~4h~~ | ~~Later~~ |
| ~~17~~ done — suffixtree alloc benchmarks pinned (docs/benchmarks/ baselines + alloc_budget_test.go) | ~~Add benchmarks for suffix tree construction + search~~ | ~~🟡 MED~~ | ~~3h~~ |
| ~~18~~ done — CI Performance Tests workflow + flake alloc-gate | ~~Add CI performance regression detection~~ | ~~🟢 LOW~~ | ~~4h~~ |
| ~~19~~ **Won't implement — examples/ stays as the SDK demo; crawls exclude it by default.** | ~~Wire `examples/` package or delete it (currently orphaned)~~ | ~~🟢 LOW~~ | ~~1h~~ | ~~Next~~ |
| ~~20~~ done — coverage is CI-tracked; domain well past 75% | ~~Add `domain` coverage to reach 75%+ (currently 66.2%)~~ | ~~🟡 MED~~ | ~~3h~~ | ~~Next~~ |
| ~~21~~ done — coverage is CI-tracked; detection well past 75% | ~~Add `detection` coverage to reach 75%+ (currently 61.8%)~~ | ~~🟡 MED~~ | ~~4h~~ | ~~Next~~ |
| ~~22~~ done — coverage is CI-tracked (92.7% average by the 2026-10-05 audit) | ~~Add `cmd/` coverage to reach 80%+ (currently 75.1%)~~ | ~~🟡 MED~~ | ~~3h~~ | ~~Later~~ |
| ~~23~~ done — resolved — go-udiff is already in the module graph (indirect); sergi/go-diff stays direct; no swap was ever needed | ~~Evaluate `go-udiff` as replacement for `sergi/go-diff` (already a transitive dep)~~ | ~~🟢 LOW~~ | ~~2h~~ | ~~Later~~ |
| ~~24~~ **Won't implement — no tracing demand surfaced since.** | ~~Add OpenTelemetry spans to pipeline stages (parse → serialize → detect → print)~~ | ~~🟢 LOW~~ | ~~6h~~ | ~~Later~~ |
| ~~25~~ **Won't implement — the import cycle persists; boundaries are enforced by go-arch-lint instead.** | ~~Hide `syntax/golang` behind facade (BLOCKED by import cycle)~~ | ~~🟢 LOW~~ | ~~8h~~ | ~~Blocked~~ |

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
