# TODO List

**Last Updated: 2026-05-23**

Actionable items planned for the next 2-4 weeks.

---

## 🔴 HIGH Priority

- [ ] Introduce ProcessedClone DTO to decouple Printer from syntax.Node (111 test call sites)
- [ ] Consolidate three parallel Clone types (printer.clone, pkg/artdupl.Clone, printer.CloneGroup)
- [ ] Implement TokenValue type with validation and refactor suffixtree/syntax to use it

## 🟡 MEDIUM Priority

- [ ] Implement CSV output format properly using encoding/csv
- [ ] Unify enum patterns: domain enums should use config's generic helpers
- [ ] Optimize memory layouts for SIMD-friendly data structures and implement string interning
- [ ] Decouple printer/clone_classify.go from syntax/golang direct import
- [ ] Add --output-file flag to stats subcommand
- [ ] Split printer/stats_test.go (975L → 3 files)

## 🟢 LOW Priority

- [ ] Refactor `syntax/golang/transform.go` (369L, 300L switch statement)
- [ ] Fix remaining LSP hints: unused params, unnecessary type args in tests
- [ ] Create domain.HealthScore typed enum (currently just a string 'A'-'F')
- [ ] Write SDK documentation for pkg/artdupl/
- [ ] Add BDD test for art-dupl stats --only templ and --only go
- [ ] Add BDD test for --include-generic end-to-end
- [ ] Add fuzz tests for templ parser edge cases
- [ ] Add ADR for semantic-as-default and reflection-based config merge
- [ ] Validate GoReleaser release config

## ✅ Recently Completed (2026-05-23)

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
- [x] Refactor MultiDetector to use DetectionConfig instead of *config.Config
- [x] Fix SortByTotalTokens bug in printer/sorter.go and printer/text.go
- [x] Fix --semantic/--structural flag descriptions
- [x] Fix TestFindProjectRoot false positives
- [x] Delete dead cli/ package, move DefaultThreshold to config/
- [x] Add --simple-json CLI flag
