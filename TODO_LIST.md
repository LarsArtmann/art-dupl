# TODO List

**Last Updated: 2026-06-16**

Actionable items planned for the next 2-4 weeks.

---

## 🔴 HIGH Priority

### Architecture (Multi-session refactors — deferred)

- [ ] Introduce ProcessedClone DTO to decouple Printer from syntax.Node internals (clone_processor.go bridges partially; actionability.go still imports syntax.Node)
- [ ] Consolidate three parallel Clone types (printer.CloneGroup, pkg/artdupl.Clone, domain.ProcessedClone)
- [ ] Split `printer/` into sub-packages (stats, html, analyze) — 50 files is too many for one package
- [ ] Type-strengthen ProcessedClone: Filename string → domain.Filepath, LineStart/LineEnd int → domain.LineNumber (~25 consumer sites)

### Architecturally Constrained

- [ ] Hide `syntax/golang` behind facade — **BLOCKED** by import cycle (syntax/golang imports syntax for Node type)

---

## 🟡 MEDIUM Priority

### UX

- [ ] Add fuzz tests for templ parser edge cases
- [ ] Wire InternFilename into node creation in transform functions

### Code Quality

- [ ] Refactor `syntax/golang/transform.go` (369L, 300L switch statement)
- [ ] Refactor `printer/actionability.go` (600L — partially done, patterns extracted)
- [ ] Fix remaining LSP hints: unused params, unnecessary type args
- [ ] Implement hybrid slice/map transition storage for small transition counts (deferred — map already O(1))

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
- [x] Deprecate FindClonesStream (delegates to FindClonesStreamResult)
- [x] Add MarshalJSON/UnmarshalJSON for 3 domain enums
- [x] Add Parse constructors for 3 domain enums
- [x] Collapse CloneSeverity into ClonePriority
- [x] Migrate LegacyIssue.Severity to ClonePriority
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
- [x] Route TODO/legacy detections via domain.Finding + FindFindings pipeline
- [x] Add --suppress-test-low and --test-threshold flags
- [x] Add DescribeTable and builder/callback detection patterns
- [x] Add string interning (InternFilename)
- [x] Add fuzz tests for suffix tree
- [x] Move test constants to test files
- [x] Fix issue_helpers.go Frags filter

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
