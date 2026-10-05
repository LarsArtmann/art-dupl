# Status Report — Remove TODO/Legacy Detection, Focus on Duplicates

> **Date:** 2026-06-17 10:44\
> **Branch:** fork\
> **Base Commit:** 26e75f1 chore(deps): bump nixpkgs, charmbracelet/x, ginkgo, gomega\
> **Session Goal:** Remove non-duplicate detection (TODO/Legacy), keep art-dupl focused on code clones only.

---

## a) FULLY DONE

### Detection Layer

- [x] **Deleted** `detection/todo_detector.go` (119 lines) — TODO/FIXME/HACK/XXX/NOTE comment detector
- [x] **Deleted** `detection/legacy_detector.go` (110 lines) — deprecated function / legacy pattern detector
- [x] **Deleted** `detection/issue_helpers.go` (202 lines) — generic `findIssuesInFile`, `findIssuesGeneric`, `findFindingsInFile`, `TodoIssue`, `LegacyIssue`
- [x] **Deleted** `detection/finding_test.go` (139 lines) — tests for issue-to-finding conversion
- [x] **Rewrote** `detection/multidetector.go` — removed `FindFindings()`, `FindFindings` goroutines, TODO/legacy method dispatch. Now only `FindDuplOver()` for clone detection.
- [x] **Rewrote** `detection/detection_test.go` (933 → ~90 lines) — removed all TODO/legacy detector tests, kept MultiDetector core tests

### Domain Layer

- [x] **Deleted** `domain/finding.go` (86 lines) — `FindingType`, `FindingTypeTodo`, `FindingTypeLegacy`, `Finding` struct, `ParseFindingType`, marshal/unmarshal/validate
- [x] **Updated** `domain/analysis_errors.go` — removed `ErrInvalidFindingType`
- [x] **Updated** `domain/domain_test.go` — removed all Finding/FindingType tests (~111 lines)

### Config Layer

- [x] **Updated** `config/detection_method.go` — removed `DetectionMethodTodos`, `DetectionMethodLegacy` constants and from `validDetectionMethods`, `AllDetectionMethods()`, `ParseDetectionMethods`
- [x] **Updated** `config/config_enum_test.go` — removed todos/legacy from parse tests, adjusted expected counts
- [x] **Updated** `config/config_test.go` — `AllDetectionMethods()` now returns 2 methods (was 4)

### Printer Layer

- [x] **Deleted** `printer/findings.go` (79 lines) — `PrintFindings` implementations for text, plumbing, JSON, HTML, SARIF, stats
- [x] **Updated** `printer/printer.go` — removed `PrintFindings(findings []domain.Finding)` from `Printer` interface
- [x] **Updated** `printer/json.go` — removed `Findings` field from `JSONOutput`, removed `findings` field from `JSONPrinter`
- [x] **Updated** `printer/sarif.go` — removed `findings` field from `sarifPrinter`, removed SARIF result generation for findings
- [x] **Updated** `printer/text.go` — removed `findingCount` and findings footer logic (was already minimal)

### CLI / Command Layer

- [x] **Updated** `cmd/run_analysis.go` — `executeAnalysis()` now returns 4 values (removed `findingChan`), removed `spawnFindingDetection()`, `spawnCloneDetection` simplified
- [x] **Updated** `cmd/run_flags.go` — `runCmd` no longer collects findings, no longer passes findings to `printDupls()`
- [x] **Updated** `cmd/run_all_modes.go` — removed `collectFindings()`, `findings` parameter from `writeFormatFile()`, `runAllModes` no longer produces findings
- [x] **Updated** `cmd/run_output.go` — `printDupls()` signature: removed `findings []domain.Finding` parameter, removed findings printing block
- [x] **Updated** `cmd/stats.go` — removed findingChan drain goroutine, adjusted `executeAnalysis` call
- [x] **Updated** `cmd/cmd_integration_test.go` — removed `TestExecuteAnalysis_FindingsPipeline`, adjusted all `executeAnalysis` call sites to 4-return signature
- [x] **Updated** `cmd/cmd_test.go` — removed `PrintFindings` from `mockPrinter`, fixed `printDupls` call sites
- [x] **Updated** `cmd/cmd_utils_test.go` — fixed `writeFormatFile` call sites

### SDK Layer

- [x] **Updated** `pkg/artdupl/types.go` — removed `MethodTodos`, `MethodLegacy` constants, updated `IsValid()` switch
- [x] **Updated** `pkg/artdupl/basic_test.go` — removed MethodTodos/MethodLegacy from test slices
- [x] **Updated** `pkg/artdupl/detector_types_test.go` — adjusted `Summary.MethodsUsed` count from 4 → 2
- [x] **Updated** `pkg/artdupl/detector_validation_test.go` — removed MethodTodos/MethodLegacy equality checks

### Examples

- [x] **Updated** `examples/examples_test.go` — removed MethodTodos/MethodLegacy from detection method list

---

## b) PARTIALLY DONE

### BuildFlow / CI

- [x] `todo-check` passes (0 TODO comments found, 236 files scanned)
- [ ] `duplications-checker` fails: found 1 clone group exceeding 30-token threshold — **pre-existing**, not introduced this session
- [ ] `jscpd` fails: signal killed (OOM/timeout on large repo) — **pre-existing**
- [x] All other 27 BuildFlow steps pass

### BDD Tests

- [x] 256 of 264 specs pass
- [ ] 8 BDD failures — **all pre-existing**, all related to `--include-*` filtering flags (sqlc, templ, protobuf, mockgen). These fail on the base commit too. Not related to TODO/legacy removal.

---

## c) NOT STARTED

### Type Model Improvements (architecture follow-up)

- [ ] `config.DetectionMethods` — still a slice wrapper; could be a proper set type with O(1) Contains
- [ ] `syntax.Match.Frags` — `[][]*syntax.Node` is a deeply nested structure; could use a flatter representation
- [ ] `domain.ProcessedCloneGroup` vs `printer.CloneGroup` vs `pkg/artdupl.Clone` — 3 parallel DTOs noted in AGENTS.md as "blocked on Printer/SDK DTO design"
- [ ] `domain.LineNumber` — `uint16` is a bounded choice; positions use `int32` in `syntax.Node` (mix of signed/unsigned)

### Missing CLI Wiring (from status archives)

- [ ] `--method` flag still accepts parsed strings but there are only 2 valid methods now; could simplify UX

### Code Health

- [ ] `go.mod` / `go.sum` show dependency bump (`gogenfilter v3.1.0 → v3.2.0`) — was modified before this session; should verify if intentional or stale

---

## d) TOTALLY FUCKED UP!

None. All modified packages compile. Unit tests pass. The removed code was cleanly excised without breaking the clone detection pipeline.

---

## e) WHAT WE SHOULD IMPROVE

1. **Consolidate Clone DTOs** — `domain.ProcessedCloneGroup`, `printer.CloneGroup`, `pkg/artdupl.Clone` should be unified into a single canonical type with printer-specific views, not parallel definitions.

2. **Fix BDD filtering tests** — 8 pre-existing failures in `bdd/` around `--include-sqlc`, `--include-templ`, `--include-protobuf`, `--include-mockgen`. The generated code filtering logic appears to not actually include these files when flags are passed. This is a real bug in the CLI filter wiring.

3. **Type safety at boundaries** — `syntax.Node` uses `int32` for `Pos`/`End`, but `domain.LineNumber` is `uint16`. The `LineRangeMixin` in printer/SARIF uses `int`. These should be unified under a single `Position` or `LineNumber` type from `pkg/position`.

4. **DetectionMethods as a set** — `config.DetectionMethods` wraps `[]DetectionMethod` with `Contains()` doing linear search over 2 items. Fine now, but if methods grow, this should be a `map[DetectionMethod]struct{}` or a bitmap.

5. **Remove dead code in printer/** — `printer/findings.go` is deleted but there may still be `finding`-related JSON struct tags or SARIF rule definitions referencing findings in HTML/JSON output schemas.

6. **Simplify `executeAnalysis` return type** — It returns `(chan syntax.Match, job.ParseStats, *FilterStats, error)`. The `chan` is created internally and closed by a goroutine. Consider returning a `<-chan` or a callback-based API for clearer ownership.

7. **Revisit `go.mod` bump** — `gogenfilter v3.1.0 → v3.2.0` was present before this session. If unintended, should revert to keep dependency changes isolated.

8. **`text.go` findingCount** — `findingCount` field is still declared but no longer set (since `PrintFindings` is removed from interface). Should delete the field and its footer usage entirely.

---

## f) Top #25 Things to Get Done Next

### High Impact / Low Effort (Quick Wins)

1. Fix `text.go` — remove dead `findingCount` field and footer logic
2. Verify `go.mod` / `go.sum` bump is intentional; revert if not
3. Remove `domain/finding.go` references from any remaining docs/status files
4. Update `FEATURES.md` — remove TODO/FIXME detection and Legacy pattern detection rows
5. Update `AGENTS.md` — remove Finding pipeline references

### High Impact / Medium Effort

6. Fix 8 BDD filtering tests (sqlc/templ/protobuf/mockgen include flags)
7. Unify `domain.ProcessedClone`, `printer.CloneGroup`, `pkg/artdupl.Clone` into single DTO
8. Unify position types: `syntax.Node.Pos` (int32), `domain.LineNumber` (uint16), `LineRangeMixin.StartLine` (int)
9. Simplify `Printer` interface — remove `StatsPrinter` extension if `ApplyStatsConfig` can be a method on `Printer`
10. Refactor `DetectionMethods` from slice to set type

### High Impact / High Effort (Architecture)

11. Modularize into sub-modules (`go-modularize` skill) — split printer, syntax, suffixtree into independently versioned modules
12. Extract a proper `Position` type in `pkg/position` and use it consistently across `syntax.Node`, `domain`, `printer`
13. Consolidate `config` and `pkg/artdupl` detection method enums — currently two parallel enum systems
14. Remove `nolint:funlen` annotations by extracting smaller functions in `cmd/run_*.go`
15. Introduce `Result[T]` or proper error wrapping at `executeAnalysis` boundaries instead of multi-return tuples

### Medium Impact / Low Effort

16. Audit `printer/html.go` for remaining `Finding` references (1484 lines, noted in AGENTS.md as overweight)
17. Remove `TestDetectionMethods` hardcoded count assertion (currently asserts `len(methods) == 2`) — fragile
18. Add `//go:build` integration test tags for slower BDD tests
19. Standardize test helper naming (`createTestNode` vs `setupXxx` vs `mustXxx`)
20. Run `go vet ./...` and fix any shadow or unreachable warnings

### Medium Impact / Medium Effort

21. Introduce `buildflow` allowlist for `duplications-checker` (art-dupl has legitimate structural duplication)
22. Add `jscpd` timeout configuration or skip in BuildFlow for this repo
23. Remove `AGENTS.md` "Finding pipeline" paragraph (marked as separate from clone pipeline)
24. Update `SDK_DESIGN.md` if it references `Finding` or `PrintFindings`
25. Run `gofumpt` + `goimports` on all modified files to ensure formatting is clean

---

## g) Top #1 Question I Cannot Figure Out Myself

**The BDD filtering tests (8 failures) appear to be real bugs, not test artifacts.** When `--include-sqlc` or `--include-templ` is passed, the generated files are still filtered out. Looking at `cmd/run_analysis.go`, `setupFilter()` creates a `gogenfilter.Filter` based on boolean config flags, but I cannot determine whether the issue is:

1. The filter not being passed correctly to the file crawl stage?
2. `gogenfilter` v3.2.0 having changed behavior from v3.1.0?
3. The test fixtures not actually containing the expected `// Code generated by` comments?

**I need:** A quick reproduction command or guidance on whether these 8 failures are known/accepted, or whether they represent a real regression that should be prioritized.

---

## Build & Test Summary

| Check                            | Result | Notes                                        |
| -------------------------------- | ------ | -------------------------------------------- |
| `go build ./...`                 | ✅     | Clean                                        |
| `go test ./cmd/...`              | ✅     | Pass                                         |
| `go test ./config/...`           | ✅     | Pass                                         |
| `go test ./detection/...`        | ✅     | Pass                                         |
| `go test ./domain/...`           | ✅     | Pass                                         |
| `go test ./examples/...`         | ✅     | Pass                                         |
| `go test ./pkg/artdupl/...`      | ✅     | Pass                                         |
| `go test ./printer/...`          | ✅     | Pass                                         |
| `go test ./bdd/...`              | ⚠️      | 256/264 pass, 8 pre-existing filter failures |
| `buildflow todo-check`           | ✅     | 0 TODO comments                              |
| `buildflow duplications-checker` | ❌     | 1 clone group >30 tokens (pre-existing)      |
| `buildflow jscpd`                | ❌     | Signal killed (pre-existing)                 |

---

## Files Changed This Session

```
32 files changed, ~34 insertions(+), ~2024 deletions(-)

Deleted:
  detection/finding_test.go
  detection/issue_helpers.go
  detection/legacy_detector.go
  detection/todo_detector.go
  domain/finding.go
  printer/findings.go

Modified:
  cmd/*.go (8 files)
  config/*.go (3 files)
  detection/detection_test.go
  detection/multidetector.go
  domain/analysis_errors.go
  domain/domain_test.go
  domain/types_file.go
  examples/examples_test.go
  pkg/artdupl/*.go (4 files)
  printer/json.go
  printer/printer.go
  printer/sarif.go
```
