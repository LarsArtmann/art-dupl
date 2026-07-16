# TODO List

**Last Updated: 2026-07-16**

Actionable items planned for the next 2-4 weeks.

---

## ✅ Completed (2026-07-16) — P4 Tests, Code Quality & CLI Polish

### Code Quality Fixes

- [x] **gocyclo on `runCmd`** — Extracted `dispatchAnalysis()` to handle allFlag/dumpTokens/standard routing (complexity 16→<15)
- [x] **9 recvcheck warnings** — Added `//nolint:recvcheck` to all domain string-enum types (standard Go JSON convention: MarshalJSON value receiver, UnmarshalJSON pointer receiver)
- [x] **ProcessedCloneGroup exhaustruct warning** — Added `NewProcessedCloneGroup` constructor that computes `TokenCount` from clones; updated 3 production call sites
- [x] **Config validation for `--workers`, `--min-lines`, `--max-cache-entries`** — All reject negative values via `validateNonNegative`
- [x] **`printSearchStatus` quiet check** — Now returns early when `cfg.Quiet` is set (was still printing checkmark emoji)

### Unit Tests (15 new)

- [x] `printBuildingStatus` quiet behavior (4 tests: quiet/non-quiet × verbose/non-verbose)
- [x] `version` subcommand text/JSON/short output (3 tests)
- [x] `parseOutputFormat` for all 6 output formats (6 subtests)
- [x] `ExitCodeForError` with wrapped internal/validation/cancel errors (3 tests)

### Integration & BDD Tests (11 new)

- [x] Config validation → exit code mapping: bad threshold, bad sort, negative workers, negative min-lines, nil error (5 tests)
- [x] BDD: version subcommand text/JSON/short output (3 scenarios)
- [x] BDD: exit code errors for bad threshold and bad sort (2 scenarios)
- [x] BDD: exit code documentation visible in `--help` (1 scenario)

### CLI Features

- [x] **`version --short`/`-s` flag** — Prints just the version string
- [x] **Exit codes in `--help`** — Root command Long description now includes exit code table
- [x] **Version cmd uses `cmd.OutOrStdout()`** — Improved testability over `fmt.Println`

### Documentation

- [x] **ADR-0013** — Typed exit codes design decision
- [x] **ADR-0014** — SuppressionConfig struct extraction rationale
- [x] **FEATURES.md** — Added exit codes, quiet/no-color, version --short, line-count filtering
- [x] **CHANGELOG.md** — Updated with all P4 additions

### End-to-End Verification

- [x] `--quiet` fully suppresses stderr (0 bytes verified)
- [x] `--no-color` produces no ANSI escape codes
- [x] `version --short` prints version string
- [x] `version --json` produces valid JSON with all 7 fields
- [x] Exit codes visible in `--help` output

---

## ✅ Completed (2026-07-16) — P3 Code Quality, Tests & UX

### Bug Fixes

- [x] **`--min-lines` filtering bug** — fixed to check ALL clones in a group (minimum LineCount), not just `Clones[0]` (ADR-0011)
- [x] **`dumpTokensOutput` testability** — refactored to accept `io.Writer` (ADR-0012)

### Integration Tests

- [x] **Actionability integration tests** — error wrapping, cobra command, builder callback, table-driven test with non-testing receiver (6 tests via `EvaluateActionabilityWithLabel`)
- [x] **Exit code tests** — 9 subtests covering nil, context cancel, validation, config, internal, and generic errors

### CLI & UX

- [x] **Typed exit codes** — `ExitCodeForError`: 0=success, 1=general, 2=config/validation, 3=internal, 130=interrupted
- [x] **`--quiet`/`-q` flag** — suppresses non-essential status output (progress messages, profiling notices)
- [x] **`--no-color` flag** — explicitly disables colored output (complements `NO_COLOR` env var)
- [x] **Shell completion** — provided by Fang (bash/zsh/fish/powershell via `art-dupl completion <shell>`)

### Code Quality

- [x] **`SuppressionConfig` struct** — bundles `SuppressTestLow`, `TestThreshold`, `MinLines` into single value (eliminates 3-param function signatures)
- [x] **`parseOutputFormat()` extraction** — reduces `runCmd` gocyclo below threshold
- [x] **`runStandardAnalysis()` extraction** — separates analysis logic from flag parsing
- [x] **Lint config fixes** — added `exhaustruct` and `gochecknoglobals` to `_test.go` exclusions, `cobra.Command` to exhaustruct exclude list

### Documentation

- [x] **CHANGELOG.md** — added entries for exit codes, quiet/no-color flags, SuppressionConfig, bug fixes, lint config
- [x] **HOW_TO_USE.md** — added sections for `--min-lines`, `--dump-tokens`, `--quiet`, `--no-color`, exit codes
- [x] **ADR-0011** — `--min-lines` minimum across all clones
- [x] **ADR-0012** — `dumpTokensOutput` io.Writer injection
- [x] **`docs/ACTIONABILITY_PATTERNS.md`** — comprehensive table of all 15 detected patterns

---

## ✅ Completed (2026-07-16) — Pareto Roadmap P0-P2 Execution

### P0: Documentation & Lint Hygiene (10 tasks)

- [x] Marked 5 feedback docs as IMPLEMENTED/ADDRESSED with resolution banners
- [x] Fixed CONTRIBUTING.md — replaced all `just` commands with `go`/`nix` equivalents
- [x] Fixed MIGRATION_QUICK_START.md — replaced `just build` with `go build`
- [x] Fixed HOW_TO_USE.md — GitHub Actions Go version 1.21→1.26, added GOEXPERIMENT
- [x] Fixed TESTING.md — added GOEXPERIMENT=jsonv2 to build commands
- [x] Fixed `assertionMethodNames` global var → `isAssertionMethod()` switch function
- [x] Fixed `isWrappingCall` per-call map allocation → `isWrappingCallName()` switch function
- [x] Added `meta.description` to nix apps in flake.nix
- [x] `cmd/filter_stats.go` — verified trailing newline already present

### P1: Detection Features & Fixes (8 tasks)

- [x] **`--test-threshold` flag** — verified already fully implemented (config, CLI flag, validation, filtering pipeline)
- [x] **KeyValueExpr field name encoding** — struct field names in composite literals now encoded via `encodeSemanticType`. `Point{X:1}` no longer matches `Size{W:1}`. 3 tests added.
- [x] **`--dump-tokens` debug flag** — outputs serialized token stream for debugging (skips detection). Shows filename, position, base type, semantic hash, name.
- [x] **`containsTRunCall` specificity fix** — now verifies receiver Ident matches common `*testing.T` variable names (t, tt, tc, test, etc.)
- [x] **Cobra detection fix** — `isCommandLiteral` now verifies receiver Ident is "cobra" or "fang", not just any SelectorExpr named "Command"
- [x] **Builder callback threshold** lowered from 3 to 2 calls for more FP suppression
- [x] **Race safety verified** — all tests pass with `-race` flag across syntax, printer, job, and pkg packages
- [x] **AGENTS.md updated** with callee encoding convention, KeyValueExpr encoding, and templ parser hierarchy gotcha

### P2: Polish & Completeness (20 tasks, key items)

- [x] **Error wrapping detection** — extended `isReturnOrWrappedReturn` to handle 2-stmt `log.Print(err); return err` pattern with `isLogOrPrintStmt` + `isLoggingMethod`
- [x] **`--min-lines` flag** — `Config.MinLines`, CLI flag, config builder wiring, filtering pipeline (all call sites updated)
- [x] **DOMAIN_LANGUAGE.md updated** — added entries for sendCtx, CloneNode, Test Threshold, Min Lines, Dump Tokens, Include Generated
- [x] **ADR-0009** — Default threshold change (1→5) with rationale
- [x] **ADR-0010** — encoding/json/v2 migration decision
- [x] **Threshold recommendation table** updated in HOW_TO_USE.md for default=5
- [x] **Benchmark** — `BenchmarkSemanticVsExactVsStructural` in syntax/golang/ showing relative performance across modes
- [x] **Stale docs trashed** — 11 obsolete SIMD/planning/research docs removed (SIMD_IMPLEMENTATION_COMPLETED, SIMD_OPTIMIZATION_ANALYSIS, SIMD_PERFORMANCE_BASELINE, SIMD_READY_ARCHITECTURE, STATICPOOL_PERFORMANCE_ANALYSIS, EXECUTION_PLAN, IMPROVEMENT_PLAN, MODERNIZATION_FINAL_REPORT, code-quality-improvements, enum-consolidation-plan, phase0-validation-safety-report)

---

## ✅ Completed (2026-07-16) — Semantic Precision + Templ Semantic Mode + 15-Project Validation

### Go Semantic Mode Fixes (3 root-cause bugs)

- [x] **Node.Fingerprint field** — `serial()` was overwriting `node.Type` with the fingerprint hash, breaking ALL actionability pattern matching for statement-level clones (every `BaseType == golang.IfStmt` check was silently broken). Added separate `Fingerprint int32` field; `Val()` routes correctly (commit `930b91a`).
- [x] **Literal value normalization** — Semantic mode now hashes BasicLit KIND (STRING, INT, FLOAT) instead of VALUE. Eliminates false negatives: Type-2 clones with different literal values now detected (commit `930b91a`).
- [x] **Generic type parameter normalization** — Type parameters (`T`, `U`) are now alpha-normalized in the symbol table. `func Map[T any]()` and `func Filter[U any]()` with same body now match (commit `61aeca8`).
- [x] **Lock+Defer Unlock actionability pattern** — `m.Lock(); defer m.Unlock()` and `m.RLock(); defer m.RUnlock()` 2-statement pattern now suppressed as idiomatic Go. Added `RUnlock` to cleanup methods, `isAcquireMethod()` helper (commit `930b91a`).
- [x] **Lint fix** — Replaced `acquireMethodNames` global var with `isAcquireMethod()` switch function, following existing `isCleanupMethod` pattern (commit `027feee`).

### Templ Semantic Mode (implemented from scratch — was purely structural)

- [x] **Phase 1: Element + attribute name encoding** — HTML tag names (`<a>`, `<div>`, `<button>`) and attribute names (`href`, `class`, `hx-get`) now hashed into node Types via `syntax.EncodeSemanticType()`. `<a href>` no longer matches `<div class>` (commit `268e3bb`).
- [x] **Phase 2: Statement-level tokenization** — Each HTML element subtree becomes one composite fingerprint token. Component names encoded. Sentinel nodes between files fix suffix-tree maximal-repeat detection. Threshold now means "N duplicated HTML elements" (commit `931d472`).
- [x] **Callee name encoding FP fix** — `extractCalleeName()` splits at first `(` to get callee name; encoded via `EncodeSemanticType`. Applied to both `transformTemplElementExpression` and `transformCallTemplateExpression`. Eliminated 2 FPs in templ-components demo files (commit `23a3b03`).
- [x] **BDD test updates** — Adjusted 5 templ BDD tests for statement-level tokenization model (commit `bbfb1c5`).

### Validation (15-project, 100% precision)

- [x] **15-project validation** — Tested against 15 real projects (6,222 Go files, 320 templ files). **Precision: 100%** (0 false positives). 119 clone groups, all verified as true positives. Performance: 70ms–1.7s per project. See `docs/status/2026-07-16_semantic-validation-15-projects.md`.

### Tests Added (19+ test cases across 5 files)

- [x] `syntax/fingerprint_test.go` — Fingerprint field behavior, serialization idempotency, non-statement Val(), Clone() (4 tests)
- [x] `syntax/golang/generics_normalization_test.go` — Type parameter alpha-normalization (1 test)
- [x] `printer/semantic_precision_test.go` — 6 end-to-end: literal normalization, Lock/Defer, error definitions, business logic, validation chains
- [x] `printer/clone_type_literal_test.go` — Clone type classification with literal normalization (2 tests)
- [x] `printer/actionability_test.go` — Lock+Defer Unlock actionability pattern (1 test)
- [x] `syntax/templ/transform_components_test.go` — Callee name extraction + semantic encoding (4 tests + 7 edge cases)

---

## ✅ Completed (2026-07-13) — Public Presence Overhaul + Docs Health Audit

- [x] **Astro + Starlight documentation website** — 15 pages: landing page, 13 Starlight docs, brand theming. Deployed to Firebase (commit `2d3bc35`).
- [x] **README rewrite** — Sales-page style, accurate threshold (5), comparison table, all features covered (commit `2d3bc35`).
- [x] **CI/CD pipeline overhaul** — Two-job deploy pattern, security headers, `FIREBASE_TOKEN` → `GOOGLE_APPLICATION_CREDENTIALS` migration (commit `23f7203`).
- [x] **Website domain migration** — `art-dupl.web.app` → `art-dupl.lars.software` (commit `23f7203`).
- [x] **Docs health audit** — 22 findings fixed across 9 core docs (FEATURES, TODO_LIST, DOMAIN_LANGUAGE, ROADMAP, HOW_TO_USE, TESTING, README, CHANGELOG, AGENTS). Threshold 15→5, SHA1→SHA-256, `just`→`go`/`nix`, removed deleted types (commit `23f7203`).
- [x] **GOEXPERIMENT=jsonv2 added to CI** — Required for `encoding/json/v2` migration (commit `e007d62`).

---

## ✅ Completed (2026-07-11) — BuildFlow Failure Recovery

- [x] **Fixed 4 BuildFlow OOM failures** — golines, nix-build, nix-build-verify, nix-hash-fix all resolved (commit `43362f9`).
- [x] **Fixed stale vendorHash** — `flake.nix` vendorHash updated for nixpkgs nixos-unstable 2026-07 (commit `43362f9`).
- [x] **Fixed 88 lint issues (88→0)** — Disabled 3 anti-idiomatic linters (exhaustruct, gochecknoglobals, recvcheck); refactored 2 gocyclo hotspots to table-driven patterns; fixed gomoddirectives v1→v2 config (commit `43362f9`).
- [x] **Refactored `evaluateActionabilityDetailed`** — gocyclo 17→3 via table-driven pattern (commit `43362f9`).
- [x] **Refactored `applyPatternLabel`** — gocyclo 17→2 via map lookup (commit `43362f9`).

---

## ✅ Completed (2026-07-01) — Full TODO Sprint (13 tasks)

### Architecture

- [x] **T23 CloneRef value object** — `domain.CloneRef` type with `Filename`, `LineStart`, `LineEnd`, `Fragment` + `LineCount()` method. Embedded in `domain.ProcessedClone` and `pkg/artdupl.Clone`. Eliminates field-name drift across 7 parallel Clone types without collapsing DTO boundary.
- [x] **T28 Sort comparator factory** — Generic `GroupMetrics[T]` + `makeGroupComparator[T]` + `sortGroupsByCriteria[T]` in `sort_unified.go`. Unified 3 parallel 4-criteria sort switch implementations (`SortCloneGroups`, `OutputText`, `sort_unified`). Deleted 4 dead `[][]*syntax.Node` sort functions + `test_helper.go`.
- [x] **T18 Hash pipeline consolidation** — Extracted `groupByHash` + `filterDuplicateGroups` shared helpers. `FindFileDuplicates` and `FindDuplOver` now share the hash+group pipeline. Added `ctx` to `FindFileDuplicates`.
- [x] **T41 Context through file feeders** — `filepath.Walk` now respects context cancellation via early-return-error in `handleWalkEntry`. Cancellation errors suppressed in `crawlDirectoryWithOpts`.

### Features

- [x] **HTML collapsible clone groups** — Added "Collapse All" / "Expand All" toolbar buttons + `collapseAll()` JS function. Added ▼/▶ collapse indicator on clone headers via CSS `::before`.
- [x] **SARIF rule metadata enrichment** — Added `Properties` field (`precision`, `problem.severity`, `tags`) to SARIF rules for GitHub Code Scanning / SonarQube compatibility.

### Infrastructure

- [x] **T36 Perf regression CI** — Added `TestPerfRegressionSerialize` + `TestPerfRegressionHashSeq` with generous thresholds (50ms). Added `bench` check to `flake.nix`.
- [x] **Cache version-mismatch warning** — `Get()` now logs the deserialize error before removing stale entries. `loadMetadata()` warns on metadata version mismatch.
- [x] **JSON config migration shim** — `Config.UnmarshalJSON` converts legacy `"semantic": false` to `"detectionMode": "exact"` when `detectionMode` is absent.
- [x] **ADR-0008** — Documented the semantic encoding layout (`[24-bit identifier/operator hash][8-bit base AST node type]`).

### Assessments (No Action Needed)

- [x] ~~Apply sendCtx to remaining channel send sites~~ — All 13 remaining bare sends target buffered(1) write-once channels (`statsChan`, `done`). No deadlock risk; converting would be cosmetic.
- [x] ~~go.mod dependency audit~~ — All 12 direct deps justified (lipgloss, log, gogenfilter, templ, fang, ginkgo, gomega, go-diff, cobra, xxh3, x/sync). No banned or unnecessary deps.
- [x] ~~Validate CI templates~~ — Fixed `./...` → `.` path in GitHub Actions template. Pre-commit hook already properly configured (`types: [go]`, `pass_filenames: false`).

### Assessed — Deferred with Rationale

- [ ] **T25 Split printer/ into sub-packages** — Feasible but requires interface inversion: core `printer.go` references `StatsPrinter`, creating circular deps with any sub-package. Clean split requires moving `Printer`/`ReadFile`/`StatsPrinter` interfaces to a separate base package + extracting shared test helpers from `_test.go` files + handling `.(*stats)` type assertions to unexported types. Multi-session architectural design work.
- [ ] **T24 Branded NodeType int32** — A single `syntax.NodeType` type does NOT prevent the stated cross-package constant value collision (golang and templ constants would share the same `NodeType` type). The proper fix requires per-package `NodeType` types (`golang.NodeType`, `templ.NodeType`), which is even more invasive. Current 8-bit shared encoding space is intentional (see ADR-0008). HIGH RISK: touches gob cache format.

---

## 🔴 Previously Deferred (Architecturally Constrained)

- [ ] **Hide `syntax/golang` behind facade** — **BLOCKED** by import cycle (syntax/golang imports syntax for Node type)
- [ ] **Thread `context.Context` through stdin scanner** — stdin `bufio.Scanner` is inherently blocking; can't interrupt without closing stdin

### Critical Correctness (4/4 resolved)

- [x] Non-destructive serialization — `serial()` now shallow-copies each node before writing Type/Owns (commit `8498d01`)
- [x] Type-2 classification — `classifyCloneType` walks `node.Children` directly via `collectNamesPreOrder`, no longer calls `syntax.Serialize` (commit `69d3c1c`)
- [x] Incremental cache-miss aliasing — deep-clone-on-store + `singleflight.Group` deduplicates concurrent parses (commits `8498d01`, `94b5205`)
- [x] `suffixtree.Update` returns `error` instead of panicking — all 37 call sites updated (commit `8498d01`)

### Quick Wins (7/7)

- [x] `Summary.AnalysisTime` marshals as milliseconds (commit `69d3c1c`)
- [x] Removed `runtime.GC()` from `PrintProfileResult` (commit `69d3c1c`)
- [x] Deterministic sort of clone groups by hash key (commit `69d3c1c`)
- [x] `config.MaxChildrenSerial` wired into `serial()` via `SerializeWithMaxChildren` (commit `69d3c1c`)
- [x] `crypto/sha1` → `crypto/sha256` for cache keys; `CacheVersion` bumped 1→2 (commit `69d3c1c`)
- [x] Fixed broken `RunTableTest` (reflection-based Name extraction) (commit `69d3c1c`)
- [x] Removed dead `FileDetector.threshold` field (commit `69d3c1c`)

### Robustness (5/5)

- [x] DetectionMode enum — replaced 2 config bools with single `Config.DetectionMode` enum (commit `df1a754`, ADR-0007)
- [x] FuncLit alpha-normalization — closures now canonicalized in semantic mode (commit `fae336b`)
- [x] Deleted dead code: `STree.String()`, `cache.GetStats()`, profiler dead funcs, BDDError type (commit `fae336b`)
- [x] Extracted `sendCtx[T]` generic helper for context-aware channel sends (commit `fae336b`)
- [x] Cache eviction via `cache.Prune(maxEntries)` + `Config.MaxCacheEntries` (commit `bed1dcf`)

### Infrastructure

- [x] Parallel incremental parsing with worker pool (`ParseIncrementalParallel`) (commit `94b5205`)
- [x] GitHub Actions workflow template + pre-commit hook template (commit `c0a5f22`)
- [x] Benchmark suite for `syntax.Serialize` (commit `c0a5f22`)
- [x] ADR-0006 (non-destructive serial) + ADR-0007 (DetectionMode enum) (commit `c0a5f22`)
- [x] CI templates written to `templates/` (commit `c0a5f22`)

### Lint Cleanup (post-sprint, 2026-07-01 02:50)

- [x] Fixed exhaustive switch in `cmd/util.go` (T8 fallout — missing `DetectionModeSemantic` case)
- [x] Fixed wrapcheck in `pkg/artdupl/types.go` (T2 fallout — `json.Marshal/Unmarshal`)
- [x] Fixed 27 test errcheck via `mustUpdate` helper (T7 fallout)
- [x] Reconciled stale AGENTS.md (destructive serial OPEN→RESOLVED)

---

## ✅ Completed (2026-06-28) — Generated-Code Inclusion Flag Unification

- [x] Unify the six per-generator `--include-*` flags under a single `--include-generated <category>` flag (commit `35c79de`). Categories: `sqlc`, `templ`, `protobuf`, `mockgen`, `stringer`, `generic`, `all` (repeatable / comma-separated). Legacy flags survive as hidden deprecated aliases.
- [x] Scope the statement-level tokenization guard to the match's filename so templ matches survive in mixed Go/templ corpora (commit `3d1d841`, `dataContainsStatements` → `fileContainsStatements`).
- [x] Catch & correct a `go.sum` trim regression before it reached history (`go mod tidy` proved the 47 removed checksums were required); only the stale `flake.nix` `vendorHash` bump was committed.

---

## 🔴 HIGH Priority

### Type Safety (From 2026-06-23 Data Model Review)

- [ ] Introduce branded `NodeType int32` in syntax/ to prevent cross-package int32 collision between golang and templ node type constants — **HIGH RISK**: touches gob serialization cache format and semantic encoding layout (`[24-bit hash][8-bit base type]`). Needs feature flag + cache-version migration.
- [x] ~~Introduce shared `CloneRef` value object in domain to unify the 7 parallel Clone types (ProcessedClone, SDK Clone, JSONClone, etc.) via embedding without collapsing DTO boundary~~ — Done: T23 completed (2026-07-01). `domain.CloneRef` embedded in `ProcessedClone` and `pkg/artdupl.Clone`.
- [x] ~~Relocate `SortCriteria` and `OutputFormat` enums from config to domain~~ — Done: both live in `domain/` (`domain/sort_criteria.go`, `domain/output_format.go`) with `config/` aliases (2026-07-01)
- [x] ~~Unify `FileReaderFunc` type~~ — Done: canonical type in domain, aliased in printer and SDK (2026-06-23)
- [x] ~~Add `Config.Validate()` method~~ — Done: single entry point delegating to existing ValidateConfig (2026-06-23)
- [x] ~~Fix stringly-typed JSON DTO fields~~ — Done: JSONClone.Category/Priority/Actionability/CloneType now use domain enums directly (2026-06-23)

### Architecture (Multi-session refactors — deferred with rationale)

- [x] ~~Introduce ProcessedClone DTO to decouple Printer from syntax.Node internals~~ — Done: `domain.CloneNode` recursive tree type bridges `syntax.Node` → actionability evaluation; actionability files no longer import `syntax` (2026-07-01)
- [x] ~~Consolidate **five** parallel Clone/Group types~~ — Field names aligned (`LineStart`/`LineEnd`/`StartPos`/`EndPos` canonical), Fragment unified to `string`, `printer.CloneGroup.Files`→`Clones`. Types remain separate for Printer/SDK DTO independence (see ADR-0005, `docs/research/SPLIT-BRAIN.html`).
- [ ] Split `printer/` into sub-packages (stats, html, analyze) — ~29 source files / ~3500+ lines. Previously blocked by ProcessedClone DTO coupling; now unblocked.
- [x] ~~Unify `Fragment` type (`[]byte` in domain vs `string` in SDK — flips at every boundary)~~ — Done: `domain.ProcessedClone.Fragment` is now `string` everywhere.
- [x] ~~Rename `…Data` view models to `…View` in printer/~~ — Verified: all view models already use `…View` suffix.

### Correctness Fixes (From 2026-06-23 Full Code Review)

- [x] ~~Fix suffixtree error swallow in canonize~~ — Done: silent `_` discard replaced with explicit panic-with-context (2026-06-23)
- [x] ~~Fix cache atomic read race~~ — Done: Stats()/GetStats() now use atomic.LoadInt64 (2026-06-23)
- [x] ~~Remove dead DuplError.Line field~~ — Done: field never set, always printed `:0` (2026-06-23)
- [x] ~~Remove dead SortNodesByCriteria~~ — Done: zero callers (2026-06-23)
- [x] ~~Remove dead SARIF rules~~ — Done: art-dupl/todo and art-dupl/legacy never produced in results (2026-06-23)
- [x] ~~Cache deep-copy on incremental cache-hit path~~ — Done: deep-clone-on-store + `singleflight.Group` (commits `8498d01`, `94b5205`)
- [x] ~~Add `Name()` method to MethodDetector interface~~ — Done: `Name() string` on interface, eliminates type switch (2026-07-01)
- [x] ~~Fix `debug.Stack()` called unconditionally on every error~~ — Done: `DuplError` no longer captures `debug.Stack()` (removed in prior sprint)

### Split-Brain Resolution (2026-06-22)

- [x] Unify `DetectionMethod` across `config`/`pkg/artdupl`/`detection` — now type aliases to `domain.DetectionMethod` (see ADR-0005)
- [x] Align error sentinels (`ErrInvalidThreshold`/`ErrThresholdTooLarge`/`ErrInvalidDetectionMethod`) — re-exported from `domain` for cross-package `errors.Is()` compatibility
- [x] Unify `Logger` interface — explicit in `pkg/logger` with compile-time assertions; SDK aliases it
- [x] Fix `Timeout` type mismatch — `config.Config.Timeout` is now `time.Duration` (not `int` seconds)
- [x] Remove dead `DetectionMethods.Strings()` and `methodsToStrings()` — typed slices pass directly after alias unification

### Architecturally Constrained

- [ ] Hide `syntax/golang` behind facade — **BLOCKED** by import cycle (syntax/golang imports syntax for Node type)
- [ ] Thread `context.Context` through `cmd/run_crawl.go` stdin scanner — stdin `bufio.Scanner` is inherently blocking; can't interrupt without closing stdin. (`filepath.Walk` ctx cancellation was completed in T41.)

---

## 🟡 MEDIUM Priority

### Code Quality

- [ ] Implement hybrid slice/map transition storage for small transition counts (deferred — map already O(1))

### Assessed — No Action Needed

- [x] ~~Refactor `syntax/golang/transform.go` (370L, 49-case switch)~~ — Switch is inherent to Go AST type dispatch. Already has `//nolint:funlen,maintidx,nonamedreturns,gocognit`. Extracting cases to methods would reduce readability.
- [x] ~~Fix remaining LSP hints~~ — All golangci-lint issues resolved (0 issues). LSP/gopls shows stale diagnostics; always trust `golangci-lint run` over IDE.

---

## ✅ Completed (2026-06-20) — Brutal Self-Review Sprint #3

### Concurrency Fixes (Goroutine Leak Elimination)

- [x] Fix 3 goroutine leaks in SDK `FindClonesStreamResult` — blocking sends on `resultChan` without `ctx.Done()` select
- [x] Fix 4 goroutine leaks in `job/parse.go` pipeline — "check-then-send" race pattern in `Parse`, `startWorkers`, `collectResults` (added ctx param), `serializeAST`
- [x] Fix goroutine leak in `job/buildtree.go` — `done` channel was unbuffered, `done <- true` blocked if caller selected `ctx.Done()`
- [x] Fix goroutine leak in `job/incremental.go` — `schan <- nodes` blocking send without select
- [x] Fix goroutine leak in `cmd/run_all_modes.go` — `matchChan <- match` without select
- [x] Fix goroutine leak in `cmd/run_hash.go` — `duplChan <- match` check-then-send race
- [x] Fix goroutine leak in `pkg/artdupl/detector_pipeline.go` — `fileChan <- filename` check-then-send race

### Stale Reference Cleanup

- [x] Remove stale `--since` flag from BDD test harness (`internal/testutil/bdd_runners.go`)
- [x] Fix misleading test names referencing removed `ParseFileByExtension` wrapper

## ✅ Completed (2026-06-20) — Brutal Self-Review Sprint #2

### Dead Code Removal

- [x] Remove dead `ParseFileByExtension` wrapper (only test callers, delegated to `WithConfig` variant)

### Error Handling Fixes

- [x] Fix 3 swallowed errors in `cache/file_cache.go` (MkdirAll, Remove, saveMetadata) — now log warnings to stderr
- [x] Add `context.Context` to `hash.FileDetector.FindDuplOver` — was the only MethodDetector without cancellation support

### Linting/Config Fixes

- [x] Remove 5 misleading `goexperiment.*` build tags from `.golangci.yml` (arenas, goroutineleakprofile, jsonv2, runtimesecret, simd) — none used by any code
- [x] Fix non-functional depguard allow-list — was set to only `$gostd` + `$module`, now lists all 16 approved external dependencies

### Test Quality Improvements

- [x] Fix 2 always-pass tests in `detector_uncovered_test.go` that discarded errors
- [x] Add `pkg/enum` tests (was 0% coverage, now 100%) — critical shared dependency for all domain enum JSON marshaling
- [x] Add `domain.HealthScore` tests (was 0% coverage) — covers IsValid, String, MarshalJSON, UnmarshalJSON

## ✅ Completed (2026-06-20) — Brutal Self-Review Sprint

### Dead Code Removal (Ghost Systems Eliminated)

- [x] Delete dead `domain.Filepath`/`LineNumber` branded types — full smart-constructors + JSON marshaling existed but zero consumers outside domain package. Deleted `types_file.go`, `helpers.go`, and their tests.
- [x] Remove dead `ParseClonePriority`/`ParseCloneCategory`/`ParseCloneActionability` — exported, never called.
- [x] Remove dead `ErrInvalidLineNumber` sentinel — never referenced.
- [x] Remove dead `CloneClassification.NodeTypeName` field — set but never read.
- [x] Remove 5 dead error constructors (`NewParseError`, `NewDetectionError`, `NewAnalysisError`, `NewTimeoutError`, `NewCancelledError`) + `EnumValidationError` type — only used in own test files.
- [x] Remove dead `ErrorType.String()` method — never called.
- [x] Remove 4 dead `ErrorType` constants (`ParseError`, `DetectionError`, `TimeoutError`, `CancelledError`).

### Lying Documentation Fixed

- [x] Fix `domain/domain.go` package doc — removed references to 6 non-existent types (`BytePosition`, `Threshold`, `TokenCount`, etc.).
- [x] Fix `go.mod` module doc — removed "SIMD-optimized" lies, `CloneSeverity` references, non-existent `SafeMarshalConfig`/`SafeMarshalClone`/`domain.NewThreshold` functions.

### API Cleanup

- [x] Remove deprecated `FindClonesStream` from `Detector` interface — was marked deprecated but still required by interface, forcing all implementations to provide it.
- [x] Rename `SARIFConfig` → `SARIFPrinterOptions` — collided with `SARIFConfiguration` (SARIF spec type in same file).
- [x] Fix 12 `unusedwrite` diagnostics in `examples_test.go` — added missing field assertions.
- [x] Fix 3 `infertypeargs` diagnostics — removed with dead Parse functions.

### Assessed — No Action Needed

- [x] Duplicate error sentinels (config vs artdupl) — ACCEPTED: expected decoupling pattern (SDK defines own sentinels).
- [x] Duplicate `noOpLogger` implementations — ACCEPTED: expected decoupling pattern.
- [x] Raw `"art-dupl"` strings in tests — ACCEPTED: tests should use literal expected values.
- [x] `CloneCategory` type alias in printer — ACCEPTED: convenience alias, not a split brain.
- [x] `config.Config` GeneratorFilter extraction — DEFERRED: reflection-based merge treats embedded structs as single units, would break field-by-field override semantics.
- [x] `Fragment` `[]byte` vs `string` — DEFERRED: works correctly at boundaries, needs design decision.

## ✅ Completed (2026-06-20) — Correctness & Cleanup Sprint

### Correctness (Critical bugs found in self-review)

- [x] Fix `started time.Time` data race — field was written by `FindClones` and `FindClonesStreamResult` (both public methods), read by `buildResult`. Concurrent calls on one detector would race. Removed field; `startTime` is now a local variable.
- [x] Fix `Summary.LinesAnalyzed` hardcoded to 0 — the data existed in `job.ParseStats.LinesCount` but was never wired through to `buildResult`. Now populated from pipeline stats.
- [x] Fix sync `FindClones` never reporting 100% progress — only streaming path emitted completion event.
- [x] Decouple SDK from internal `errors` package — `detector.go` was the last file importing `github.com/LarsArtmann/art-dupl/errors`. That package calls `debug.Stack()` on every error. Replaced all `errors.Wrap*` with `fmt.Errorf` using `%w`. SDK now uses only stdlib for error handling.

### Code Quality (Dead code + naming)

- [x] Remove dead `config.DetectionConfig` — zero consumers after SDK+detection decoupling.
- [x] Remove 6 dead SDK sentinel errors never returned by production code.
- [x] Unexport `SimpleJSONClone`/`SimpleCloneGroup`/`SimpleJSONOutput` — used only within `printer/json.go`.
- [x] Rename `validateInputsOrError` → `validateInputsWithContext` (misleading "OrError" suffix).
- [x] Rename `hashConfig` → `configDebugString` (it's not a hash, it's a debug string).
- [x] Align `LineRangeMixin` fields: `StartLine`→`LineStart`, `EndLine`→`LineEnd` — was a split brain within printer (JSONClone used LineStart/LineEnd while the embedded mixin used StartLine/EndLine).

## ✅ Completed (2026-06-20) — Architecture Hardening Sprint

### Correctness

- [x] Remove `--since <git-ref>` dead flag — was accepted and stored but never read by any analysis code. Flag, config field, tests, and docs all removed.

### Architecture

- [x] Decouple `pkg/artdupl` SDK from internal `config/` — **ZERO config imports** in pkg/artdupl/. SDK defines own `detectorConfig`, error sentinels, and types. `detection` package also decoupled (owns `detection.Config` with `[]string` methods).
- [x] Remove `domain/` → internal `errors/` dependency — replaced 6 `errors.NewValidationError(msg, nil)` calls with stdlib `errors.New(msg)`. Domain is now a true leaf package.
- [x] Align Clone field names — `pkg/artdupl.Clone`: `StartLine`→`LineStart`, `EndLine`→`LineEnd` to match canonical naming in `domain.ProcessedClone` and `printer.JSONClone`.

### Code Quality

- [x] Split `printer/actionability.go` (624L) into 4 category files: `actionability.go` (public API + structural patterns), `actionability_control_flow.go` (defer + error), `actionability_test_patterns.go` (test data + scaffolding), `actionability_data.go` (data + builder).
- [x] Rename `syntax/hash_simd.go` → `hash_seq.go` — file contains no SIMD code (uses `sync.Pool` + `xxh3.Hash`).
- [x] Remove empty `syntax/golang/golang.go` anchor file — merged comment into `doc.go`.
- [x] Enforce decoupling in `.go-arch-lint.yml`: removed `config` from `sdk` and `detection` mayDependOn, removed `errors` from `domain` mayDependOn.

## ✅ Completed (2026-06-16) — Code Quality Sprint

### Code Quality

- [x] Wire InternFilename into all 4 transformer construction sites (golang parse, templ parse, NewSyntheticFileNode, incremental cache-hit path)
- [x] Fix forcetypeassert in intern.go (replaced sync.Map with RWMutex+map for type safety)
- [x] Add fuzz tests for templ parser (FuzzParseBytes — 2M+ execs, no panics)
- [x] Extract spawnCloneDetection helper from executeAnalysis
- [x] All lint issues resolved (0 golangci-lint issues)

## ✅ Completed (2026-06-16) — Full TODO Sprint

### Tier 1: Quick Fixes

- [x] Fix Clone.IsValid() StartPos==0 bypass
- [x] Wire ErrNoDuplicatesFound sentinel in FindClones
- [x] Deep-copy Options slices in NewDetector
- [x] Remove dead Patterns/Imports fields from LegacyPattern
- [x] Remove false positive: os/exec.CommandContext
- [x] Replace fragile fmt.Sprintf AST stringification with nodeContainsFunctionCall
- [x] Log parse errors at Warn instead of silently returning nil
- [x] Fix .go-arch-lint.yml sdk/pkg-utils glob overlap
- [x] Add ClonePriority.Rank() to domain
- [x] Replace priorityScore/priorityHigher with Rank()

### Tier 2: Architecture

- [x] Add StreamResult type for streaming error propagation
- [x] Add FindClonesStreamResult to Detector interface
- [x] Deprecate FindClonesStream (delegates to FindClonesStreamResult)
- [x] Add MarshalJSON/UnmarshalJSON for 3 domain enums
- [x] Add Parse constructors for 3 domain enums
- [x] Collapse CloneSeverity into ClonePriority
- [x] Fix HealthScore legend
- [x] Extract shared pkg/enum package, unify domain enums
- [x] Add context.Context to suffixtree FindDuplOver
- [x] Break SDK type aliases (DetectionMethod, Logger)

### Tier 3: Features & Code Quality

- [x] Add context.Context to all detectors (goroutine leak prevention)
- [x] Wire Actionability defaults in ClassifyClone
- [x] Extract everySequenceMatch helper
- [x] Extract validateLocation helper
- [x] Create ADR-0004
- [x] Activate MethodDetector interface with adapter implementations
- [x] Add --suppress-test-low and --test-threshold flags
- [x] Add DescribeTable and builder/callback detection patterns
- [x] Add string interning (InternFilename)
- [x] Add fuzz tests for suffix tree
- [x] Move test constants to test files

## ✅ Previously Completed (2026-06-15)

- [x] Fix exhaustive switch: missing `PriorityLow` case in printer/stats.go
- [x] Fix assertion matcher typo: `(HaveOccurred` → `HaveOccurred` in actionability.go
- [x] Fix dead code in `isReturnOrWrappedReturn`
- [x] Fix emoji collision: CategoryHandler and CategoryTestFixture
- [x] Rename `CloneClassification.NodeType` → `NodeTypeName`
- [x] Rename `cnt` → `count` in syntax/syntax.go
- [x] Fix `.go-arch-lint.yml`: Add `domain` to detection deps
- [x] Update FEATURES.md: Fix stale claims, add 11 missing features
- [x] Update AGENTS.md
- [x] Run `go mod tidy`
