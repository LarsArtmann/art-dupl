# TODO List

**Last Updated: 2026-04-05**

Actionable items planned for the next 2-4 weeks.

---

## 🔴 HIGH Priority

- [ ] Implement TokenValue type with validation and refactor suffixtree/syntax to use it

## 🟡 MEDIUM Priority

- [ ] Update README with new default semantic behavior and install commands
- [ ] Optimize memory layouts for SIMD-friendly data structures and implement string interning
- [ ] Implement CSV output format properly using encoding/csv

## 🟢 LOW Priority

- [ ] Split large files for maintainability:
  - pkg/artdupl/detector.go (546 lines)
  - cmd/run.go (528 lines)
  - printer/stats.go (727 lines)
  - domain/clone.go (495 lines)
  - domain/domain_types.go (525 lines)
- [ ] Add package examples and godoc documentation
