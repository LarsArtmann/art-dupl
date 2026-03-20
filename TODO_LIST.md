# TODO List

**Generated from 60 items**

## 🔴 HIGH Priority

- [ ] Fix gosec security violations (G115 integer overflow, G301/G304/G306 file permissions) (source: 2026-02-27_06-22_golangci-lint-fix-comprehensive.md)
- [ ] Fix cyclomatic complexity (cyclop) and cognitive complexity (gocognit) in critical functions (source: 2026-02-27_18-30-REMAINING_LINT_EXECUTION_PLAN.md)
- [ ] Add SARIF output format for security tool integration (source: 2026-02-28_05-20_COMPREHENSIVE_STATUS_REPORT.md)

## 🟡 MEDIUM Priority

- [ ] Fix BDD test suite (37-54 tests failing with exit status 1) (source: bdd/ test suite, 2026-01-21_03-58_BUILDFLOW_EXECUTION_STATUS.md)
- [ ] Fix compilation errors and import cycles between config/domain packages (source: domain/threshold_test.go, config/config.go, domain/domain.go)
- [ ] Fix JSON character corruption and UTF-8 marshaling issues (source: 2025-12-18_00-23_CODE_DUPLICATION_ELIMINATION_CRITICAL_STATUS.md)
- [ ] Fix threshold flag functionality and config merge precedence bugs (source: 2026-02-14_02-04_incremental-detection-config-merge-fix.md)
- [ ] Fix nil pointer dereference in suffix tree (SA5011) and add bounds checks (source: detection/working_test.go, printer/sorter.go:48)
- [ ] Split cli.go (605-847 lines) into executor.go, analyzer.go, formatter.go, validator.go (source: 2026-01-14_02-13_UTILITY-ENUM-CONSOLIDATION-COMPLETE-AND-CLI-ISSUE.md)
- [ ] Implement TokenValue type with validation and refactor suffixtree/syntax to use it (source: 2026-03-03_0450-PARETO_EXECUTION_MASTERPLAN.md)
- [ ] Fix TYPE SAFETY ISSUE in FindSyntaxUnits - use domain types instead of int for positions (source: syntax/syntax.go:76, 111-120)
- [ ] Fix type safety violations in SDK conversion functions and adapter layer (source: docs/ARCHITECTURE_REVIEW.md)
- [ ] Fix err113 dynamic error creation and wrapcheck error wrapping violations (source: 2026-03-01_00-36_COMPREHENSIVE_LINT_FIX_STATUS.md)
- [ ] Fix revive linter warnings (50-103 issues) and tagliatelle JSON naming (29-50 issues) (source: 2026-02-28_05-34_COMPREHENSIVE STATUS REPORT.md)
- [ ] Fix staticcheck issues (20 warnings) and errcheck unchecked errors (source: 2026-01-03_12-20_EXECUTION_REFLECTION.md)
- [ ] Fix receiver naming consistency (recvcheck) and add t.Helper() to test helpers (source: 2026-02-27_18-30-REMAINING_LINT_EXECUTION_PLAN.md)
- [ ] Extract test binary builder utility from BDD tests to bddutil/build.go (source: 2026-01-07_14-45_comprehensive-codebase-status-update.md)
- [ ] Fix JSON output format inconsistencies (detection_method vs detection_methods) (source: 2026-02-12_19-06_COMPREHENSIVE_STATUS_UPDATE.md)
- [ ] Fix double-counting in TotalDuplicateLines and add unique duplicate lines metric (source: 2026-02-12_17-51_stats-metrics-misleading-analysis.md)
- [ ] Create generic SortStrategy[T] interface and consolidate sorting logic (source: 2026-01-07_14-45_comprehensive-codebase-status-update.md)
- [ ] Update README with new default semantic behavior and install commands (source: 2026-02-25_01-59_semantic-default-implementation.md)
- [ ] Optimize memory layouts for SIMD-friendly data structures and implement string interning (source: MEMORY_LAYOUT_OPTIMIZATION_PLAN.md)

## 🟢 LOW Priority

- [ ] Implement CSV output format properly using encoding/csv (source: printer/stats.go)

## ⚪ Unknown Priority

- [ ] Resolve CLI argument routing and flag redefinition conflicts between Cobra/Ginkgo frameworks (source: 2025-12-14_08-56_critical-cli-crisis-fang-migration.md)
- [ ] Remove dual CLI system (delete old Run() function from cli.go:28-134) (source: 2026-01-07_14-43_dual-cli-systems-investigation-complete.md)
- [ ] Split pkg/artdupl/detector.go (546 lines) into detector.go, detection_methods.go, pipeline.go, result_builder.go, validators.go (source: pkg/artdupl/detector.go:530)
- [ ] Split cmd/run.go (528 lines) into 5 focused files (source: cmd/run.go:513)
- [ ] Split printer/stats.go (727 lines) into 6 focused files (source: printer/stats.go)
- [ ] Split domain/clone.go (495 lines) by entity (Clone, CloneGroup, Analysis, Repository) (source: domain/clone.go:376)
- [ ] Split domain/domain_types.go (525 lines) into 5 focused files (source: domain/domain_types.go)
- [ ] Split domain/coverage_test.go (1277-1348 lines) into type-specific test files (source: 2026-02-28_05-37_COMPREHENSIVE_STATUS_REPORT.md)
- [ ] Split pkg/artdupl/detector_test.go (1252 lines) and cmd/cmd_test.go (1156 lines) (source: 2026-02-28_05-37_COMPREHENSIVE_STATUS_REPORT.md)
- [ ] Replace primitives with domain types in job/, printer/, detection/, suffixtree/ packages (source: 2026-01-13_22-26_ARCHITECTURE-ANALYSIS-COMPLETE.md)
- [ ] Remove CloneID from Clone struct and delete CloneID type from domain_types.go (source: MEMORY_LAYOUT_OPTIMIZATION_PLAN.md)
- [ ] Implement proper global variable elimination with dependency injection (source: 2025-12-15_11-06_WELL-NAMED_COMPREHENSIVE_PROJECT_STATUS.md)
- [ ] Replace cli global variable and cliConfig global with DI context (source: docs/planning/2025-12-15)
- [ ] Convert SemanticHashEnabled global to dependency injection (source: 2026-02-25_04-01_EXECUTION_PROGRESS_REPORT.md)
- [ ] Create unified config merging helper to eliminate duplicate merge logic (source: config/config.go:162-224)
- [ ] Implement semantic detection with identifier hashing and wire to CLI flags (source: 2026-02-15_08-29_semantic-detection-implementation.md)
- [ ] Complete --profile and --timeout flag implementations (source: 2026-02-12_21-55_SESSION-COMPLETE.md)
- [ ] Implement concurrent file processing with worker pools and --workers flag (source: 2026-02-20_04-01_PARETO_EXECUTION_PLAN.md)
- [ ] Complete ignore file support with .duplignore pattern matching (source: 2025-12-15_11-06_WELL-NAMED_COMPREHENSIVE_PROJECT_STATUS.md)
- [ ] Implement hash-based detection (currently delegates to suffix tree) (source: docs/ARCHITECTURE_REVIEW.md)
- [ ] Implement multi-detection method support in runDetection and streamDetectionResults (source: pkg/artdupl/detector.go:221,270)
- [ ] Increase test coverage: cmd (10.8%→50%), detection (24%→50%), job (26.2%→80%), internal/utils (37.5%→80%) (source: 2026-02-14_03-35_COMPREHENSIVE STATUS REPORT.md)
- [ ] Increase pkg/artdupl coverage (54%→80%), syntax (66%→80%), printer (67%→80%) (source: 2026-03-01_00-24_comprehensive-status-report.md)
- [ ] Create internal/testutil package with shared BDD helpers and benchmark fixtures (source: 2026-01-22_03-04_test_utility_extraction.md)
- [ ] Add fuzzing tests for Suffix Tree Construction, AST Serialization, and Clone Detection (source: 2026-01-14_04-02_NATIVE-GO-TESTING-TOOLS-ANALYSIS-AND-MODERNIZATION-PLAN.md)
- [ ] Create GitHub Actions workflow, CI/CD integration templates, and pre-commit hooks (source: 2026-02-28_05-37_COMPREHENSIVE STATUS REPORT.md)
- [ ] Add package examples and godoc documentation for domain, config, syntax, printer packages (source: 2026-02-20_06-55_PARETO_EXECUTION_PLAN_v2.md)
- [ ] Create Architecture Decision Records (ADRs) for major design decisions (source: 2026-02-24_12-45_EXECUTIVE_STATUS_REPORT.md)
- [ ] Design plugin architecture for extensible detection methods (source: 2026-02-15_03-37_semantic-fingerprinting-implementation-plan.md)
- [ ] Add TypeScript/JavaScript and Python language support (source: 2026-02-28_05-37_COMPREHENSIVE STATUS REPORT.md)
- [ ] Research and implement Language Server Protocol (LSP) support (source: 2026-02-28_05-20_COMPREHENSIVE_STATUS_REPORT.md)
- [ ] Implement SIMD optimization for hash operations when ARM64 SIMD becomes available (source: internal/simd/simd.go:35)
- [ ] Implement hybrid slice/map approach for state.trans memory optimization (source: docs/IMPROVEMENT_PLAN.md)
- [ ] Create performance baseline benchmarks and regression test suite (source: 2026-02-28_05-20_COMPREHENSIVE_STATUS_REPORT.md)
- [ ] Create web dashboard for real-time stats visualization (source: 2026-01-25_13-53_Enhanced-Stats-Features-Complete.md)
- [ ] Implement machine learning for false positive reduction (source: 2026-02-28_05-21_COMPREHENSIVE_STATUS_REPORT.md)
- [ ] Implement watch mode for continuous monitoring and incremental detection (source: 2026-02-28_05-20_COMPREHENSIVE_STATUS_REPORT.md)
