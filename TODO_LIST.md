# TODO List

**Last Updated: 2026-04-30**

Actionable items planned for the next 2-4 weeks.

---

## 🔴 HIGH Priority

- [ ] Implement TokenValue type with validation and refactor suffixtree/syntax to use it
- [ ] Set up CI pipeline (GitHub Actions) — no automated quality gates exist

## 🟡 MEDIUM Priority

- [ ] Update README with new default semantic behavior and install commands
- [ ] Optimize memory layouts for SIMD-friendly data structures and implement string interning
- [ ] Implement CSV output format properly using encoding/csv
- [ ] Type `domain.Options.OutputFormat` as config.OutputFormat (currently string)
- [ ] Type `domain.Repository.Path` as domain.Filepath (currently string)
- [ ] Type `GetStatsData()` return type (currently any)
- [ ] Unify enum patterns: domain enums should use config's generic helpers

## 🟢 LOW Priority

- [ ] Refactor `syntax/golang/transform.go` (355L, 300L switch statement)
- [ ] Split `detection/todos.go` (TodoDetector + LegacyDetector into separate files)
- [ ] Split `config/config.go` (391L, config + validation mixed)
- [ ] Split `cmd/run_analysis.go` (447L, multiple responsibilities)
- [ ] Add explicit conversion layer domain.Clone → artdupl.Clone
- [ ] Archive old docs/status/ files (304 files, keep last 30 days)
- [ ] Fix remaining LSP hints: unused params, unnecessary type args in tests
- [ ] Implement SIMD TODOs (6 items in syntax/hash_simd.go and internal/simd/)
