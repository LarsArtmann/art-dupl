# TODO List

**Last Updated: 2026-06-22**

Actionable items planned for the next 2-4 weeks.

---

## 🔴 HIGH Priority

### Architecture (Multi-session refactors — deferred with rationale)

- [ ] Introduce ProcessedClone DTO to decouple Printer from syntax.Node internals (clone_processor.go bridges partially; actionability.go still imports syntax.Node — 34 references)
- [x] ~~Consolidate **five** parallel Clone/Group types~~ — Field names aligned (`LineStart`/`LineEnd`/`StartPos`/`EndPos` canonical), Fragment unified to `string`, `printer.CloneGroup.Files`→`Clones`. Types remain separate for Printer/SDK DTO independence (see ADR-0005, `docs/research/SPLIT-BRAIN.html`).
- [ ] Split `printer/` into sub-packages (stats, html, analyze) — ~29 source files / ~3500+ lines is too many for one package
- [x] ~~Unify `Fragment` type (`[]byte` in domain vs `string` in SDK — flips at every boundary)~~ — Done: `domain.ProcessedClone.Fragment` is now `string` everywhere.
- [x] ~~Rename `…Data` view models to `…View` in printer/~~ — Verified: all view models already use `…View` suffix.

### Split-Brain Resolution (2026-06-22)

- [x] Unify `DetectionMethod` across `config`/`pkg/artdupl`/`detection` — now type aliases to `domain.DetectionMethod` (see ADR-0005)
- [x] Align error sentinels (`ErrInvalidThreshold`/`ErrThresholdTooLarge`/`ErrInvalidDetectionMethod`) — re-exported from `domain` for cross-package `errors.Is()` compatibility
- [x] Unify `Logger` interface — explicit in `pkg/logger` with compile-time assertions; SDK aliases it
- [x] Fix `Timeout` type mismatch — `config.Config.Timeout` is now `time.Duration` (not `int` seconds)
- [x] Remove dead `DetectionMethods.Strings()` and `methodsToStrings()` — typed slices pass directly after alias unification

### Architecturally Constrained

- [ ] Hide `syntax/golang` behind facade — **BLOCKED** by import cycle (syntax/golang imports syntax for Node type)
- [ ] Thread `context.Context` through `cmd/run_crawl.go` file feeders — stdin scanner and `filepath.Walk` are inherently blocking; needs full chain refactor (`filesFeedWithOptions` → `buildParams` → all callers)

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
