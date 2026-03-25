# TODO List

**Last Updated: 2026-03-25**

## 🔴 HIGH Priority

- [ ] Fix gosec security violations (G115 integer overflow, G301/G304/G306 file permissions)
- [ ] Add SARIF output format for security tool integration

## 🟡 MEDIUM Priority

- [ ] Fix cyclomatic complexity (cyclop) and cognitive complexity (gocognit) in critical functions
- [ ] Implement TokenValue type with validation and refactor suffixtree/syntax to use it
- [ ] Update README with new default semantic behavior and install commands
- [ ] Fix JSON output format inconsistencies (detection_method vs detection_methods)
- [ ] Fix double-counting in TotalDuplicateLines and add unique duplicate lines metric
- [ ] Optimize memory layouts for SIMD-friendly data structures and implement string interning

## 🟢 LOW Priority

- [ ] Implement CSV output format properly using encoding/csv
- [ ] Split large files for maintainability:
  - pkg/artdupl/detector.go (546 lines)
  - cmd/run.go (528 lines)
  - printer/stats.go (727 lines)
  - domain/clone.go (495 lines)
  - domain/domain_types.go (525 lines)
- [ ] Create Architecture Decision Records (ADRs) for major design decisions
- [ ] Add package examples and godoc documentation

## ⚪ Future Considerations

- [ ] Create GitHub Actions workflow templates and pre-commit hooks
- [ ] Create performance baseline benchmarks and regression test suite
- [ ] Add TypeScript/JavaScript and Python language support
- [ ] Implement watch mode for continuous monitoring and incremental detection

## ✅ Recently Completed

- [x] Fix migration/ package tests (0% → 83.1% coverage)
- [x] Implement semantic detection with identifier hashing (default ON)
- [x] Implement hash-based detection method
- [x] Fix lint issues (60+ → 0)
- [x] Add multi-detection method support
- [x] Create internal/testutil package with shared BDD helpers
