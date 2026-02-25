# Comprehensive Execution Plan - All TODOs

**Date:** 2026-02-25 00:41  
**Status:** CRITICAL - Build/Test Failures Detected  
**Total Tasks:** 47 small tasks (max 12 min each)

---

## 🚨 CRITICAL - Fix Immediately (Blocking)

| #   | Task                                                            | Effort | Impact      | Why Critical            |
| --- | --------------------------------------------------------------- | ------ | ----------- | ----------------------- |
| C1  | Fix config test - TestSemanticField expects false but gets true | 5m     | 🔴 BLOCKING | Test failure blocks CI  |
| C2  | Fix BDD semantic detection tests (2 failing)                    | 10m    | 🔴 BLOCKING | Semantic feature broken |
| C3  | Fix templ parser API breaking changes (Range field removed)     | 12m    | 🔴 BLOCKING | Build failure           |
| C4  | Run full test suite after fixes                                 | 5m     | 🔴 BLOCKING | Verify fixes work       |

---

## ⚡ HIGH IMPACT / LOW EFFORT (Quick Wins)

| #   | Task                                                      | Effort | Impact    | Customer Value              |
| --- | --------------------------------------------------------- | ------ | --------- | --------------------------- |
| Q1  | Update README.md install command: make build → just build | 3m     | 🟢 HIGH   | Users use correct build cmd |
| Q2  | Extract unique() function from bdd_test.go to shared util | 8m     | 🟢 HIGH   | Code deduplication          |
| Q3  | Add t.Helper() to test helper functions                   | 5m     | 🟢 MEDIUM | Better test error messages  |
| Q4  | Add context.Context to git exec commands (fix noctx lint) | 10m    | 🟢 MEDIUM | Security/timeout handling   |
| Q5  | Fix errcheck issues in cache tests                        | 5m     | 🟢 MEDIUM | Error handling completeness |
| Q6  | Mark test helpers with t.Helper() in simd tests           | 3m     | 🟢 LOW    | Cleaner test output         |

---

## 🔧 HIGH IMPACT / MEDIUM EFFORT (Important)

| #   | Task                                                   | Effort | Impact    | Customer Value       |
| --- | ------------------------------------------------------ | ------ | --------- | -------------------- |
| H1  | Split git/change_detector.go (361 lines) into 2 files  | 12m    | 🟢 HIGH   | Maintainability      |
| H2  | Add missing test coverage for core pipeline            | 12m    | 🟢 HIGH   | Reliability          |
| H3  | Create GitHub issue templates (bug, feature, question) | 10m    | 🟢 MEDIUM | Community support    |
| H4  | Add package examples for domain types                  | 12m    | 🟢 MEDIUM | Developer experience |
| H5  | Add package examples for config package                | 10m    | 🟢 MEDIUM | Developer experience |
| H6  | Verify ignoreFiles config field is fully implemented   | 8m     | 🟢 MEDIUM | Feature completeness |
| H7  | Add comprehensive help examples to all commands        | 12m    | 🟢 MEDIUM | User experience      |

---

## 📊 MEDIUM IMPACT / MEDIUM EFFORT (Polish)

| #   | Task                                                           | Effort | Impact    | Customer Value           |
| --- | -------------------------------------------------------------- | ------ | --------- | ------------------------ |
| M1  | Replace remaining log.Fatal() calls with proper error handling | 12m    | 🟡 MEDIUM | Stability                |
| M2  | Add performance benchmarks for suffix tree                     | 12m    | 🟡 MEDIUM | Performance tracking     |
| M3  | Add performance benchmarks for hash detection                  | 10m    | 🟡 MEDIUM | Performance tracking     |
| M4  | Improve HTML template styling                                  | 12m    | 🟡 MEDIUM | Better reports           |
| M5  | Add integration tests for concurrent processing (--workers)    | 10m    | 🟡 MEDIUM | Verify parallelism works |
| M6  | Add edge case tests for config validation                      | 10m    | 🟡 MEDIUM | Robustness               |
| M7  | Add package documentation for printer package                  | 8m     | 🟡 LOW    | API documentation        |
| M8  | Add package documentation for detection package                | 8m     | 🟡 LOW    | API documentation        |

---

## 🏗️ ARCHITECTURE IMPROVEMENTS (Long-term Value)

| #   | Task                                                      | Effort | Impact    | Customer Value  |
| --- | --------------------------------------------------------- | ------ | --------- | --------------- |
| A1  | Refactor: Remove golang.SemanticHashEnabled global        | 12m    | 🟢 HIGH   | Testability     |
| A2  | Refactor: Extract common flags between root and stats     | 12m    | 🟢 HIGH   | Maintainability |
| A3  | Add dependency injection for ChangeDetector               | 12m    | 🟢 MEDIUM | Testability     |
| A4  | Refactor: Extract file reading interface for testability  | 12m    | 🟢 MEDIUM | Testability     |
| A5  | Add exhaustive switch case handling (fix exhaustive lint) | 8m     | 🟡 MEDIUM | Code safety     |
| A6  | Reduce cyclomatic complexity in ParseParallel             | 12m    | 🟡 MEDIUM | Maintainability |
| A7  | Reduce cognitive complexity in runCmd                     | 12m    | 🟡 MEDIUM | Maintainability |

---

## 📁 LARGE FILE SPLITTING (26 files >350 lines)

### Non-Test Files (Priority)

| #   | Task                                     | Effort | Impact  | Lines |
| --- | ---------------------------------------- | ------ | ------- | ----- |
| L1  | Split git/change_detector.go (361 lines) | 12m    | 🟢 HIGH | 361   |

### Test Files (Lower Priority)

| #   | Task                                                         | Effort   | Impact    | Lines   |
| --- | ------------------------------------------------------------ | -------- | --------- | ------- |
| L2  | Split domain/coverage_test.go (1278 lines) into 3 files      | 12m      | 🟡 MEDIUM | 1278    |
| L3  | Split pkg/artdupl/detector_test.go (1252 lines) into 3 files | 12m      | 🟡 MEDIUM | 1252    |
| L4  | Split cmd/cmd_test.go (1070 lines) into 3 files              | 12m      | 🟡 MEDIUM | 1070    |
| L5  | Split remaining 22 test files                                | 12m each | 🟢 LOW    | 350-900 |

---

## 🎯 ADVANCED FEATURES (Future)

| #   | Task                                     | Effort       | Impact    | Customer Value    |
| --- | ---------------------------------------- | ------------ | --------- | ----------------- |
| F1  | Design plugin architecture               | 12m research | 🟡 MEDIUM | Extensibility     |
| F2  | Research web interface feasibility       | 12m research | 🟡 LOW    | Accessibility     |
| F3  | Add telemetry/metrics for feature usage  | 12m          | 🟡 LOW    | Product insights  |
| F4  | Create performance regression test suite | 12m          | 🟢 MEDIUM | Quality assurance |

---

## 📋 TASK PRIORITY MATRIX

### Execution Order (By Value/Effort Ratio)

| Phase               | Tasks | Time | Goal                |
| ------------------- | ----- | ---- | ------------------- |
| **1. CRITICAL**     | C1-C4 | 32m  | Unblock build/tests |
| **2. QUICK WINS**   | Q1-Q6 | 34m  | High impact, fast   |
| **3. HIGH VALUE**   | H1-H7 | 76m  | Core improvements   |
| **4. ARCHITECTURE** | A1-A7 | 80m  | Long-term health    |
| **5. MEDIUM**       | M1-M8 | 86m  | Polish              |
| **6. FILES**        | L1-L5 | 72m  | Code organization   |
| **7. FUTURE**       | F1-F4 | 48m  | Research            |

**Total Estimated Time:** ~7 hours (428 minutes)

---

## 🎲 IMPACT/EFFORT MATRIX

### High Impact / Low Effort (Do First)

- C1-C4 (Critical fixes)
- Q1-Q6 (Quick wins)
- H1 (Split change_detector.go)

### High Impact / High Effort (Schedule)

- H2-H7 (Important features)
- A1-A4 (Architecture improvements)
- L2-L5 (File splitting)

### Low Impact / Low Effort (Fill gaps)

- Q6, M7-M8 (Documentation)
- M6 (Edge cases)

### Low Impact / High Effort (Defer)

- F1-F4 (Advanced features)
- L5 (All test file splits)

---

## ✅ COMPLETION CRITERIA

### Must Have (Critical Path)

- [ ] All tests passing
- [ ] Build successful
- [ ] No linting errors
- [ ] README updated

### Should Have (High Value)

- [ ] Large files split
- [ ] GitHub templates created
- [ ] Package examples added
- [ ] Test coverage >80%

### Nice to Have (Polish)

- [ ] All linting warnings fixed
- [ ] Architecture refactored
- [ ] Advanced features researched

---

## 🚀 IMMEDIATE NEXT ACTIONS

1. **Fix C1:** TestSemanticField failure
2. **Fix C2:** BDD semantic detection tests
3. **Fix C3:** Templ parser API changes
4. **Verify:** Full test suite passes
5. **Execute Q1:** README update
6. **Execute Q2:** Extract unique() function

---

_Generated: 2026-02-25 00:41_  
_Total Tasks: 47_  
_Estimated Time: 7 hours_  
_Priority: CRITICAL → QUICK WINS → HIGH VALUE → ARCHITECTURE_
