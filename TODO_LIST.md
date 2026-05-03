# TODO List

**Last Updated: 2026-05-03**

Actionable items planned for the next 2-4 weeks.

---

## 🔴 HIGH Priority

- [ ] Implement TokenValue type with validation and refactor suffixtree/syntax to use it

## 🟡 MEDIUM Priority

- [ ] Optimize memory layouts for SIMD-friendly data structures and implement string interning
- [ ] Implement CSV output format properly using encoding/csv
- [ ] Unify enum patterns: domain enums should use config's generic helpers
- [ ] Introduce ProcessedClone DTO to decouple Printer from syntax.Node (111 test call sites)
- [ ] Consolidate three parallel Clone types (printer.clone, pkg/artdupl.Clone, printer.CloneGroup)

## 🟢 LOW Priority

- [ ] Refactor `syntax/golang/transform.go` (355L, 300L switch statement)
- [ ] Archive old docs/status/ files (304 files, keep last 30 days)
- [ ] Fix remaining LSP hints: unused params, unnecessary type args in tests
- [ ] Implement SIMD TODOs (6 items in syntax/hash_simd.go and internal/simd/)

## ✅ Recently Completed (2026-05-03)

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

## ✅ Previously Completed (2026-05-01)

- [x] Replace hand-rolled `hasSuffix` with `strings.HasSuffix` in `config/filetype.go`
- [x] Fix `GetStatsData() any` → `*StatsData` (concrete return type)
- [x] Fix `buildJSONData() any` → `jsonStatsOutput` (concrete return type)
- [x] Simplify `StatsPrinter` with `StatsConfig` option struct (8 setters → 1 struct)
- [x] Fix CI go-version from `1.26rc2` to `stable`
- [x] Replace hand-rolled `htmlEscape` with `html.EscapeString` stdlib
- [x] Fix `domain.Analysis.CreatedAt` string → `time.Time` (prior session)
- [x] Fix `config.Config.Only` string → `config.FileType` (prior session)
- [x] Split `printer/html.go` 1484L → 4 files (prior session)
