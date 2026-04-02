# TODO List

**Last Updated: 2026-03-28**

## 🔴 HIGH Priority

- [x] Fix gosec security violations (G115 integer overflow, G301/G304/G306 file permissions) - **COMPLETED**
- [x] Add SARIF output format for security tool integration - **COMPLETED**

## 🟡 MEDIUM Priority

- [x] Fix cyclomatic complexity (cyclop) and cognitive complexity (gocognit) in critical functions - **COMPLETED**
- [ ] Implement TokenValue type with validation and refactor suffixtree/syntax to use it
- [ ] Update README with new default semantic behavior and install commands
- [x] Fix JSON output format inconsistencies (detection_method vs detection_methods) - **COMPLETED (already consistent)**
- [x] Fix double-counting in TotalDuplicateLines and add unique duplicate lines metric - **COMPLETED (already fixed)**
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
- [x] **Fix funlen lint issue in cmd/run_analysis.go (split buildSuffixTree into smaller functions)**
- [x] **Add SARIF output format with full test coverage**
- [x] **Verify gosec annotations are properly in place**

---

## Summary of Recent Changes (2026-03-28)

### 1. Fixed Lint Issue (cmd/run_analysis.go)

- Split `buildSuffixTree` function (92 lines) into three focused functions:
  - `buildSuffixTree` - Entry point that delegates to appropriate implementation
  - `buildSuffixTreeIncremental` - Handles incremental parsing with cache
  - `buildSuffixTreeStandard` - Handles standard parsing without cache

### 2. Added SARIF Output Format

- Added `OutputFormatSARIF` constant in `config/detectionmethod.go`
- Created `printer/sarif.go` with full SARIF 2.1.0 implementation:
  - SARIFOutput, SARIFRun, SARIFTool, SARIFResult structures
  - Proper level mapping (error/warning/note based on clone size)
  - Duplicate hash filtering to avoid redundant results
  - Invocation timing information
- Created `printer/sarif_test.go` with 8 comprehensive tests
- Updated `cmd/flags.go` to add `--sarif` flag
- Updated `cmd/run_flags.go` to handle sarif flag
- Updated `cmd/run_analysis.go` to create SARIF printer
- Updated test in `config/config_test.go` to expect 7 formats

### 3. Verified Gosec Annotations

- All G115 integer overflow annotations are properly in place
- All G304 file permission annotations are properly in place
- No new security issues introduced

### Build & Test Status

- ✅ All builds pass
- ✅ All unit tests pass (config, printer, cmd packages)
- ✅ SARIF printer tests: 8/8 passing
- ✅ Lint: 0 issues
