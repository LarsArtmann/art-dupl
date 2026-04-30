# TODO List

**Last Updated: 2026-05-01**

Actionable items planned for the next 2-4 weeks.

---

## 🔴 HIGH Priority

- [ ] Implement TokenValue type with validation and refactor suffixtree/syntax to use it

## 🟡 MEDIUM Priority

- [ ] Update README with new default semantic behavior and install commands
- [ ] Optimize memory layouts for SIMD-friendly data structures and implement string interning
- [ ] Implement CSV output format properly using encoding/csv
- [ ] Type `domain.Options.OutputFormat` as config.OutputFormat (currently string)
- [ ] Consolidate threshold validation: 3 error sentinels (pkg/artdupl, domain, config) → 1
- [ ] Unify enum patterns: domain enums should use config's generic helpers

## 🟢 LOW Priority

- [ ] Refactor `syntax/golang/transform.go` (355L, 300L switch statement)
- [ ] Split `detection/todos.go` (TodoDetector + LegacyDetector into separate files)
- [ ] Split `config/config.go` (341L, config + validation mixed)
- [ ] Split `cmd/run_analysis.go` (447L, multiple responsibilities)
- [ ] Archive old docs/status/ files (304 files, keep last 30 days)
- [ ] Fix remaining LSP hints: unused params, unnecessary type args in tests
- [ ] Implement SIMD TODOs (6 items in syntax/hash_simd.go and internal/simd/)

## ✅ Recently Completed (2026-05-01)

- [x] Replace hand-rolled `hasSuffix` with `strings.HasSuffix` in `config/filetype.go`
- [x] Fix `GetStatsData() any` → `*StatsData` (concrete return type)
- [x] Fix `buildJSONData() any` → `jsonStatsOutput` (concrete return type)
- [x] Simplify `StatsPrinter` with `StatsConfig` option struct (8 setters → 1 struct)
- [x] Fix CI go-version from `1.26rc2` to `stable`
- [x] Replace hand-rolled `htmlEscape` with `html.EscapeString` stdlib
- [x] Fix `domain.Analysis.CreatedAt` string → `time.Time` (prior session)
- [x] Fix `config.Config.Only` string → `config.FileType` (prior session)
- [x] Split `printer/html.go` 1484L → 4 files (prior session)
