# TODO List

**Last Updated: 2026-06-15**

Actionable items planned for the next 2-4 weeks.

---

## 🔴 HIGH Priority

### Architecture

- [ ] Activate `MethodDetector` interface for polymorphic dispatch (currently hard-coded in MultiDetector)
- [ ] Introduce ProcessedClone DTO to decouple Printer from syntax.Node internals
- [ ] Consolidate four parallel Clone types (printer.clone, pkg/artdupl.Clone, printer.CloneGroup, domain.ProcessedClone)
- [ ] Split `printer/` into sub-packages (stats, html, analyze) — 50 files is too many for one package

### Correctness

- [ ] Fix SDK `FindClonesStream` error handling — pipeline errors silently swallowed (logged, not returned)
- [ ] Deep-copy `Options` in `NewDetector` — shared `*Options` allows post-construction mutation panic
- [ ] Fix `legacy_detector.go:36` fragile string matching (stringified AST → false positives)
- [ ] Wire `ErrNoDuplicatesFound` sentinel — advertised in SDK but never returned by `FindClones`
- [ ] Fix `Clone.IsValid()` — skips length check when `StartPos == 0` (byte offset zero)

### Type Safety

- [ ] Collapse `CloneSeverity`/`ClonePriority` into one type (same 4 values, two types)
- [ ] Add JSON validation to CloneCategory, ClonePriority, CloneActionability
- [ ] Wire `Actionability` field in `CloneClassification` — currently always zero value (invalid)
- [ ] Break SDK type aliases (`DetectionMethod = config.DetectionMethod`, `Logger = logger.Logger`)

---

## 🟡 MEDIUM Priority

### Safety

- [ ] Add `context.Context` to MultiDetector goroutines (goroutine leak on consumer abandon)
- [ ] Document detector thread-safety contract (stateful `d.started` field races under concurrent use)
- [ ] Remove dead `Patterns`/`Imports` fields in `LegacyPattern` struct
- [ ] Fix `issue_helpers.go:82` — `Frags: [][]*syntax.Node{{}}` always passes length filter

### UX

- [ ] Fix HealthScore legend vs formula mismatch (legend says <5%=A, formula actually scores differently)
- [ ] Add `ClonePriority.Rank()` to domain (deduplicate ordinal logic in printer/stats.go and printer/html.go)
- [ ] Add `--suppress-test-low` flag for blanket suppression of test-only low-priority clones
- [ ] Separate test/production threshold support

### Architecture

- [ ] Hide `syntax/golang` and `syntax/templ` behind `syntax` facade (4 packages import sub-packages directly)
- [ ] Unify enum patterns: domain enums should use config's generic helpers
- [ ] Implement string interning for duplicate identifier names
- [ ] Fix `.go-arch-lint.yml` sdk/pkg-utils glob overlap

---

## 🟢 LOW Priority

### Code Quality

- [ ] Refactor `syntax/golang/transform.go` (369L, 300L switch statement)
- [ ] Refactor `printer/actionability.go` (558L — extract `everySequenceMatch` helper, move test constants)
- [ ] Fix remaining LSP hints: unused params, unnecessary type args in tests
- [ ] Extract `validateLocation` helper to deduplicate todo_detector/legacy_detector boilerplate
- [ ] Fix `todo_detector.go:43` — silent parse error (should log at Warn)

### Features

- [ ] Add fuzz tests for templ parser edge cases
- [ ] Add property-based/fuzz tests for suffix tree invariants
- [ ] Add Ginkgo `DescribeTable` lambda detection pattern
- [ ] Add builder/callback pattern detection (`makeFix`/builder)
- [ ] Implement hybrid slice/map transition storage for small transition counts

### Documentation

- [ ] Create ADR for actionability pattern detection system
- [ ] Document detection method help text (CLI `-m` help omits `todos` and `legacy`)

---

## ✅ Recently Completed (2026-06-15)

- [x] Fix exhaustive switch: missing `PriorityLow` case in printer/stats.go
- [x] Fix assertion matcher typo: `(HaveOccurred` → `HaveOccurred` in actionability.go
- [x] Fix dead code in `isReturnOrWrappedReturn` (unreachable ReturnStmt branch)
- [x] Fix emoji collision: CategoryHandler and CategoryTestFixture both used 🎯
- [x] Rename `CloneClassification.NodeType` → `NodeTypeName` (avoid int32/string name collision)
- [x] Rename `cnt` → `count` in syntax/syntax.go (4 functions)
- [x] Fix `.go-arch-lint.yml`: Add `domain` to detection deps, map `internal/utils` to pkg-utils
- [x] Update FEATURES.md: Fix stale claims, add 11 missing features
- [x] Update AGENTS.md: Add cache/ and errors/ to architecture, fix Clone count
- [x] Run `go mod tidy` to fix missing transitive hashes in go.sum

## ✅ Previously Completed (2026-06-11)

- [x] Decouple printer/clone_classify.go from syntax/golang direct import
- [x] Create domain.HealthScore typed enum with validation and JSON marshaling
- [x] Add FuncType-based interface implementation detector (non-actionable pattern)
- [x] Complete exhaustive switch in applyPatternLabel for all PatternLabel cases
- [x] Reduce GetCategoryEmoji cyclomatic complexity 16→2 via map lookup
- [x] Add test cases for isInterfaceImplementation detector
- [x] Implement CSV output format properly using encoding/csv
- [x] Add --output-file flag to stats subcommand
- [x] Write SDK documentation for pkg/artdupl/
- [x] Add BDD tests for --only and --include-generic with stats
- [x] Validate GoReleaser release config (remove unnecessary -r flag)
- [x] Eliminate all 3 godoclint warnings to achieve zero lint issues

## ✅ Previously Completed (2026-05-23)

- [x] Delete internal/simd/ dead code package (163L, 2 stale TODOs)
- [x] Delete hashSeqSIMD() dead indirection in syntax/hash_simd.go
- [x] Fix nix lint check sandbox: set GOLANGCI_LINT_CACHE in flake.nix
- [x] Fix all 21 lint issues to 0 (errcheck, goconst, exhaustruct, err113, gci, golines, gocyclo)
- [x] Reduce buildJSONData cyclomatic complexity 16→<10 via 6 extracted helpers
- [x] Create .gitleaks.toml to suppress false positives
- [x] Archive old docs/status/ files (349 → 67, 283 archived)
- [x] Fix os.Exit(1) test killer — statError removed, validatePaths added
- [x] Fix BDD prepareSubcommandArgs flag-value parsing bug (9 hidden failures)
- [x] Wire --include-generic filter (catch-all for generated code)
- [x] Stats output improvements (category, priority, actionability, test/prod, top clones)
- [x] Templ position audit — fixed 5 bugs (ConstantAttribute, BoolConstantAttribute, ChildrenExpression, CaseExpression, File root End)
- [x] In-process BDD test migration (333 lines net reduction)
- [x] Makefile deleted, CI workflows consolidated (3 → 1)
- [x] SDK hardcoded version → runtime/debug.ReadBuildInfo()
- [x] AGENTS.md accuracy audit (5 ghost dirs, 3 ghost types removed)
- [x] Add ADR for semantic-as-default and reflection-based config merge
- [x] Split printer/stats_test.go (975L → 5 files)

## ✅ Previously Completed (2026-05-03)

- [x] Wire TODO/Legacy detectors through MultiDetector registry
- [x] Consolidate threshold validation: error sentinels moved to config/enum_helpers.go
- [x] Split detection/todos.go (352L → issue_helpers.go, todo_detector.go, legacy_detector.go)
- [x] Split config/config.go (344L → config.go, config_io.go, config_validate.go)
- [x] Split cmd/run_analysis.go (450L → run_analysis.go, run_hash.go, run_printer.go)
- [x] Fix README.md semantic defaults, templ defaults, add missing features
- [x] Delete printer/format.go — move ParseFormat to config.ParseOutputFormat
- [x] Delete printer/sort_type.go — migrate all SortBy constants to config.SortCriteria
- [x] Extract config.DetectionConfig from config.Config (Methods + Verbose)
- [x] Define detection.MethodDetector interface for pluggable detectors
- [x] Refactor MultiDetector to use DetectionConfig instead of \*config.Config
- [x] Fix SortByTotalTokens bug in printer/sorter.go and printer/text.go
- [x] Fix --semantic/--structural flag descriptions
- [x] Fix TestFindProjectRoot false positives
- [x] Delete dead cli/ package, move DefaultThreshold to config/
- [x] Add --simple-json CLI flag
