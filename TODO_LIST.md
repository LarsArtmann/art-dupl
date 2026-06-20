# TODO List

**Last Updated: 2026-06-20**

Actionable items planned for the next 2-4 weeks.

---

## 🔴 HIGH Priority

### Correctness — Dead/Misleading Flags

- [ ] `--since <git-ref>` is a **dead flag**: accepted and stored in `config.Since` (`cmd/config_builder.go:149`) but **never read** by any analysis code. Git-aware incremental is not implemented — `job/incremental.go` only does SHA1 content-hash caching. Either implement git-diff file selection or remove the flag + docs (FEATURES.md marked it `PARTIALLY_FUNCTIONAL`).

### Architecture (Multi-session refactors — deferred with rationale)

- [ ] Decouple `pkg/artdupl` SDK from internal `config/`: only `DetectionMethod` was split off, but 5 SDK files still import `config` directly (`config.Config`, `config.DetectionConfig`, `config.DetectionMethods` slice, and re-export `config.ErrInvalidThreshold`/`ErrThresholdTooLarge`). SDK should define its own config surface.
- [ ] Remove `domain/` → internal `errors/` dependency: `domain/types_file.go:7` and `domain/helpers.go:7` import the internal `errors` package (`errors.NewValidationError`). Domain should be a leaf package depending only on stdlib + `pkg/enum`.
- [ ] Introduce ProcessedClone DTO to decouple Printer from syntax.Node internals (clone_processor.go bridges partially; actionability.go still imports syntax.Node — 34 references)
- [ ] Consolidate **five** parallel Clone/Group types (not three as previously documented): `printer.CloneGroup`, `pkg/artdupl.Clone`, `pkg/artdupl.CloneGroup`, `domain.ProcessedClone`, `domain.ProcessedCloneGroup`
- [ ] Split `printer/` into sub-packages (stats, html, analyze) — ~29 source files / ~3500+ lines is too many for one package
- [ ] Type-strengthen ProcessedClone: Filename string → domain.Filepath, LineStart/LineEnd int → domain.LineNumber (~25 consumer sites)

### Architecturally Constrained

- [ ] Hide `syntax/golang` behind facade — **BLOCKED** by import cycle (syntax/golang imports syntax for Node type)

---

## 🟡 MEDIUM Priority

### Code Quality

- [ ] Refactor `printer/actionability.go` (623L — patterns extracted, file still large; consider splitting pattern detection functions into sub-files by category: test-patterns, defer-patterns, data-patterns)
- [ ] Remove misleading filename `syntax/hash_simd.go` — contains no SIMD; uses `sync.Pool` + `xxh3.Hash`. Rename to `hash_seq.go` or similar.
- [ ] Remove empty package-anchor file `syntax/golang/golang.go` (7 lines, zero declarations) or document why it's retained.
- [ ] Implement hybrid slice/map transition storage for small transition counts (deferred — map already O(1))

### Assessed — No Action Needed

- [x] ~~Refactor `syntax/golang/transform.go` (370L, 49-case switch)~~ — Switch is inherent to Go AST type dispatch. Already has `//nolint:funlen,maintidx,nonamedreturns,gocognit`. Extracting cases to methods would reduce readability.
- [x] ~~Fix remaining LSP hints~~ — All golangci-lint issues resolved (0 issues). LSP/gopls shows stale diagnostics; always trust `golangci-lint run` over IDE.

---

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
