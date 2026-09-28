# COMPREHENSIVE EXECUTION PLAN & STATUS REPORT

**Date:** 2026-03-24 11:34 CET
**Project:** art-dupl - Code Duplication Detection Tool
**Branch:** fork (clean, up to date with origin/fork)

---

## EXECUTIVE SUMMARY

| Metric   | Status        | Value                |
| -------- | ------------- | -------------------- |
| Tests    | ✅ PASSING    | 29/29 packages       |
| Coverage | ⚠️ MIXED       | 8 packages below 80% |
| Lint     | 🔴 NEEDS WORK | 116 issues           |
| Build    | ✅ WORKING    | Binary compiles      |
| Git      | ✅ CLEAN      | Nothing to commit    |

---

## WORK STATUS

### A) FULLY COMPLETED (✅)

- Golden file testing support (charmbracelet/x/exp/golden)
- Clone classification system
- Strong ID type refactoring
- Domain type safety (IsValid methods)
- Stats subcommand (JSON, CSV, Text)
- Semantic detection (FNV-1a)
- Multi-method detection (suffix tree + hash)
- BDD test suite (226 specs)
- HTML output with diff view
- Templ/SQLC filtering
- Incremental detection

### B) PARTIALLY COMPLETED (🔄)

- Golden file test coverage - 5 files added, more needed
- Test coverage - 8 packages below 80% threshold
- Lint cleanup - 116 issues remain

### C) NOT STARTED (📋)

- Gosec security fixes (3 issues)
- Large file refactoring (printer/html.go: 1377 lines)
- Package comments (10+ missing)
- Pointer receiver consistency (19 recvcheck)

### D) TOTALLY FUCKED UP (💀)

- **Nothing is broken.** Core functionality works.

---

## LINT ISSUES BREAKDOWN (116 Total)

| Category                 | Count | Impact    | Effort |
| ------------------------ | ----- | --------- | ------ |
| **gosec** (security)     | 3     | 🔴 HIGH   | 15min  |
| revive (comments/naming) | 50    | 🟡 MEDIUM | 60min  |
| recvcheck (pointers)     | 19    | 🟡 MEDIUM | 30min  |
| nonamedreturns           | 9     | 🟢 LOW    | 15min  |
| prealloc                 | 9     | 🟢 LOW    | 10min  |
| unparam                  | 8     | 🟡 MEDIUM | 10min  |
| unconvert                | 1     | 🟢 LOW    | 1min   |
| thelper                  | 4     | 🟡 MEDIUM | 5min   |
| funlen/gocyclo/gocognit  | 4     | 🟡 MEDIUM | 120min |
| Other                    | 9     | 🟢 LOW    | 10min  |

---

## TEST COVERAGE BREAKDOWN

| Package       | Coverage | Target | Gap    |
| ------------- | -------- | ------ | ------ |
| printer       | 64.1%    | 80%    | -15.9% |
| cli           | 62.5%    | 80%    | -17.5% |
| syntax        | 67.6%    | 80%    | -12.4% |
| bdd           | 70.0%    | 80%    | -10.0% |
| hash          | 73.8%    | 80%    | -6.2%  |
| job           | 76.6%    | 80%    | -3.4%  |
| config        | 75.3%    | 80%    | -4.7%  |
| internal/enum | 77.9%    | 80%    | -2.1%  |

---

## COMPREHENSIVE EXECUTION PLAN

### PHASE 1: QUICK WINS (Security & Easy Fixes)

**Estimated: 30 minutes | Impact: 🔴 HIGH**

| #   | Task                                                                    | Time  | Impact  | Files        |
| --- | ----------------------------------------------------------------------- | ----- | ------- | ------------ |
| 1.1 | Fix gosec G115 overflow in `internal/simd/simd.go:161`                  | 5min  | 🔴 HIGH | simd.go      |
| 1.2 | Fix gosec G115 overflow in `syntax/hash_simd.go:90`                     | 5min  | 🔴 HIGH | hash_simd.go |
| 1.3 | Add nolint directive for gosec G204 in `internal/testutil/binary.go:43` | 5min  | 🟡 MED  | binary.go    |
| 1.4 | Fix unconvert in `adapter/printer_adapter.go:86`                        | 1min  | 🟢 LOW  | adapter.go   |
| 1.5 | Add missing package comments                                            | 10min | 🟡 MED  | 10 files     |
| 1.6 | Remove dot imports from test files                                      | 5min  | 🟡 MED  | 3 files      |

### PHASE 2: TEST COVERAGE IMPROVEMENTS

**Estimated: 90 minutes | Impact: 🟡 MEDIUM**

| #   | Task                                    | Time  | Impact | Files    |
| --- | --------------------------------------- | ----- | ------ | -------- |
| 2.1 | Increase printer coverage (64.1% → 80%) | 20min | 🟡 MED | printer/ |
| 2.2 | Increase cli coverage (62.5% → 80%)     | 15min | 🟡 MED | cli/     |
| 2.3 | Increase syntax coverage (67.6% → 80%)  | 15min | 🟡 MED | syntax/  |
| 2.4 | Increase bdd coverage (70.0% → 80%)     | 15min | 🟡 MED | bdd/     |
| 2.5 | Increase hash coverage (73.8% → 80%)    | 10min | 🟢 LOW | hash/    |
| 2.6 | Increase config coverage (75.3% → 80%)  | 10min | 🟢 LOW | config/  |
| 2.7 | Fix job coverage (76.6% → 80%)          | 5min  | 🟢 LOW | job/     |

### PHASE 3: LINT CLEANUP

**Estimated: 120 minutes | Impact: 🟡 MEDIUM**

| #   | Task                               | Time  | Impact | Files                 |
| --- | ---------------------------------- | ----- | ------ | --------------------- |
| 3.1 | Fix unparam issues (8)             | 10min | 🟡 MED | various               |
| 3.2 | Add thelper to test helpers (4)    | 5min  | 🟡 MED | various               |
| 3.3 | Remove named returns (9)           | 15min | 🟢 LOW | various               |
| 3.4 | Add prealloc hints (9)             | 10min | 🟢 LOW | various               |
| 3.5 | Fix recvcheck (19 pointer issues)  | 30min | 🟡 MED | domain/, config/      |
| 3.6 | Fix function complexity (4 issues) | 30min | 🟡 MED | printer/html.go, cmd/ |
| 3.7 | Fix noctx issues (3)               | 10min | 🟡 MED | bdd/, git/            |
| 3.8 | Fix nolintlint directive           | 5min  | 🟡 MED | syntax/golang/        |
| 3.9 | Fix goconst issue                  | 5min  | 🟢 LOW | job/                  |

### PHASE 4: LARGE FILE REFACTORING

**Estimated: 180 minutes | Impact: 🟡 MEDIUM**

| #   | Task                                           | Time  | Impact | Files    |
| --- | ---------------------------------------------- | ----- | ------ | -------- |
| 4.1 | Split printer/html.go (1377 lines)             | 90min | 🟡 MED | printer/ |
| 4.2 | Split printer/stats_test.go (950 lines)        | 45min | 🟢 LOW | printer/ |
| 4.3 | Split domain/domain_types_test.go (1027 lines) | 45min | 🟢 LOW | domain/  |

### PHASE 5: GOLDEN FILE TEST EXPANSION

**Estimated: 60 minutes | Impact: 🟢 LOW**

| #   | Task                             | Time  | Impact | Files    |
| --- | -------------------------------- | ----- | ------ | -------- |
| 5.1 | Add text output golden tests     | 15min | 🟢 LOW | printer/ |
| 5.2 | Add JSON output golden tests     | 15min | 🟢 LOW | printer/ |
| 5.3 | Add CSV output golden tests      | 15min | 🟢 LOW | printer/ |
| 5.4 | Add plumbing output golden tests | 15min | 🟢 LOW | printer/ |

---

## PRIORITIZED TASK TABLE (Sorted by Impact/Effort)

| Priority | Task                             | Time   | Impact   | Phase |
| -------- | -------------------------------- | ------ | -------- | ----- |
| 🔴 P0    | Fix gosec G115 simd.go:161       | 5min   | SECURITY | 1.1   |
| 🔴 P0    | Fix gosec G115 hash_simd.go:90   | 5min   | SECURITY | 1.2   |
| 🔴 P0    | Fix gosec G204 binary.go:43      | 5min   | SECURITY | 1.3   |
| 🟡 P1    | Increase printer coverage to 80% | 20min  | QUALITY  | 2.1   |
| 🟡 P1    | Increase cli coverage to 80%     | 15min  | QUALITY  | 2.2   |
| 🟡 P1    | Add package comments             | 10min  | QUALITY  | 1.5   |
| 🟡 P1    | Remove dot imports               | 5min   | QUALITY  | 1.6   |
| 🟡 P2    | Fix recvcheck (19 issues)        | 30min  | QUALITY  | 3.5   |
| 🟡 P2    | Fix function complexity          | 30min  | QUALITY  | 3.6   |
| 🟡 P2    | Increase syntax coverage         | 15min  | QUALITY  | 2.3   |
| 🟡 P2    | Increase bdd coverage            | 15min  | QUALITY  | 2.4   |
| 🟢 P3    | Fix unparam (8 issues)           | 10min  | QUALITY  | 3.1   |
| 🟢 P3    | Fix noctx issues                 | 10min  | QUALITY  | 3.7   |
| 🟢 P3    | Remove named returns             | 15min  | STYLE    | 3.3   |
| 🟢 P3    | Add prealloc hints               | 10min  | STYLE    | 3.4   |
| 🟢 P3    | Split large files                | 180min | MAINT    | 4     |

---

## TOTAL TIME ESTIMATE

| Phase     | Tasks  | Time         |
| --------- | ------ | ------------ |
| Phase 1   | 6      | 30min        |
| Phase 2   | 7      | 90min        |
| Phase 3   | 9      | 120min       |
| Phase 4   | 3      | 180min       |
| Phase 5   | 4      | 60min        |
| **TOTAL** | **29** | **~8 hours** |

---

## WHAT WE SHOULD IMPROVE

1. **Security** - Fix 3 gosec issues immediately
2. **Test Coverage** - Bring 8 packages above 80%
3. **Large Files** - Split printer/html.go (1377 lines)
4. **Documentation** - Add missing package comments
5. **Code Style** - Address revive issues

---

## TOP #1 QUESTION

**Question:** Should we fix all lint issues before adding new features, or continue with incremental cleanup?

**Context:**

- 116 lint issues but only 3 are security-related
- Core functionality works perfectly
- Tests pass with good coverage

**Recommendation:** Fix security issues immediately, then incremental cleanup.

---

## RECOMMENDED NEXT STEPS (Next Session)

1. Fix 3 gosec security issues (15 min)
2. Add package comments to key packages (10 min)
3. Increase printer coverage by 10% (10 min)
4. Remove dot imports from test files (5 min)

**After these: ~40 min invested, ~3 security issues fixed, ~10% less lint issues**

---

**Report Generated:** 2026-03-24 11:34 CET
**Next Review:** After Phase 1 completion
