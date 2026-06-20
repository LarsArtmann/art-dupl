# TODO List

**Last Updated: 2026-06-20**

Actionable items planned for the next 2-4 weeks.

---

## 🔴 HIGH Priority

### Architecture (Multi-session refactors — deferred with rationale)

- [ ] Introduce ProcessedClone DTO to decouple Printer from syntax.Node internals (clone_processor.go bridges partially; actionability.go still imports syntax.Node — 34 references)
- [ ] Consolidate **five** parallel Clone/Group types: `printer.CloneGroup`, `pkg/artdupl.Clone`, `pkg/artdupl.CloneGroup`, `domain.ProcessedClone`, `domain.ProcessedCloneGroup`. Field names are now aligned (`LineStart`/`LineEnd` canonical across all types), but the types themselves are still separate.
- [ ] Split `printer/` into sub-packages (stats, html, analyze) — ~29 source files / ~3500+ lines is too many for one package
- [ ] Type-strengthen ProcessedClone: Filename string → domain.Filepath, LineStart/LineEnd int → domain.LineNumber (~25 consumer sites)

### Architecturally Constrained

- [ ] Hide `syntax/golang` behind facade — **BLOCKED** by import cycle (syntax/golang imports syntax for Node type)

---

## 🟡 MEDIUM Priority

### Code Quality

- [ ] Implement hybrid slice/map transition storage for small transition counts (deferred — map already O(1))

### Assessed — No Action Needed

- [x] ~~Refactor `syntax/golang/transform.go` (370L, 49-case switch)~~ — Switch is inherent to Go AST type dispatch. Already has `//nolint:funlen,maintidx,nonamedreturns,gocognit`. Extracting cases to methods would reduce readability.
- [x] ~~Fix remaining LSP hints~~ — All golangci-lint issues resolved (0 issues). LSP/gopls shows stale diagnostics; always trust `golangci-lint run` over IDE.

---

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
