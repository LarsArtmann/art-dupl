# TODO List

**Last Updated: 2026-06-16**

Actionable items planned for the next 2-4 weeks.

---

## 🔴 HIGH Priority

### Architecture (Multi-session refactors — deferred)

- [ ] Activate `MethodDetector` interface for polymorphic dispatch (interface updated with ctx, needs adapter implementations)
- [ ] Introduce ProcessedClone DTO to decouple Printer from syntax.Node internals
- [ ] Consolidate four parallel Clone types (printer.clone, pkg/artdupl.Clone, printer.CloneGroup, domain.ProcessedClone)
- [ ] Split `printer/` into sub-packages (stats, html, analyze) — 50 files is too many for one package

### Correctness

- [ ] Fix SDK `FindClonesStream` error handling — `FindClonesStreamResult` added, old method still swallows errors for backward compat

### Type Safety

- [x] ~~Collapse `CloneSeverity`/`ClonePriority` into one type~~ ✅ Done
- [x] ~~Add JSON validation to CloneCategory, ClonePriority, CloneActionability~~ ✅ Done
- [x] ~~Wire `Actionability` field in `CloneClassification`~~ ✅ Done
- [x] ~~Break SDK type aliases~~ ✅ Done

---

## 🟡 MEDIUM Priority

### Safety

- [x] ~~Add `context.Context` to MultiDetector goroutines~~ ✅ Done
- [x] ~~Document detector thread-safety contract~~ ✅ Done
- [x] ~~Remove dead `Patterns`/`Imports` fields in `LegacyPattern` struct~~ ✅ Done
- [x] ~~Fix `issue_helpers.go:82`~~ ✅ Done

### UX

- [x] ~~Fix HealthScore legend vs formula mismatch~~ ✅ Done
- [x] ~~Add `ClonePriority.Rank()` to domain~~ ✅ Done
- [ ] Add `--suppress-test-low` flag for blanket suppression of test-only low-priority clones
- [ ] Separate test/production threshold support

### Architecture

- [ ] Hide `syntax/golang` and `syntax/templ` behind `syntax` facade (4 packages import sub-packages directly)
- [ ] Unify enum patterns: domain enums should use config's generic helpers
- [ ] Implement string interning for duplicate identifier names
- [x] ~~Fix `.go-arch-lint.yml` sdk/pkg-utils glob overlap~~ ✅ Done

---

## 🟢 LOW Priority

### Code Quality

- [ ] Refactor `syntax/golang/transform.go` (369L, 300L switch statement)
- [ ] Refactor `printer/actionability.go` (552L — partially done)
- [ ] Fix remaining LSP hints: unused params, unnecessary type args in tests
- [x] ~~Extract `validateLocation` helper~~ ✅ Done
- [x] ~~Fix `todo_detector.go:43` silent parse error~~ ✅ Done

### Features

- [ ] Add fuzz tests for templ parser edge cases
- [ ] Add property-based/fuzz tests for suffix tree invariants
- [ ] Add Ginkgo `DescribeTable` lambda detection pattern
- [ ] Add builder/callback pattern detection (`makeFix`/builder)
- [ ] Implement hybrid slice/map transition storage for small transition counts

### Documentation

- [x] ~~Create ADR for actionability pattern detection system~~ ✅ Done (ADR-0004)
- [x] ~~Document detection method help text~~ ✅ Done

---

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
- [x] Update CLI -m help text to include todos and legacy
- [x] Add ClonePriority.Rank() to domain
- [x] Replace priorityScore/priorityHigher with Rank()

### Tier 2: Architecture
- [x] Add StreamResult type for streaming error propagation
- [x] Add FindClonesStreamResult to Detector interface
- [x] Add MarshalJSON/UnmarshalJSON for 3 domain enums
- [x] Add Parse constructors for 3 domain enums
- [x] Collapse CloneSeverity into ClonePriority
- [x] Migrate LegacyIssue.Severity to ClonePriority
- [x] Fix HealthScore legend

### Tier 3: Features & Code Quality
- [x] Add context.Context to all detectors (goroutine leak prevention)
- [x] Break SDK type aliases (DetectionMethod, Logger)
- [x] Wire Actionability defaults in ClassifyClone
- [x] Extract everySequenceMatch helper
- [x] Extract validateLocation helper
- [x] Create ADR-0004

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
