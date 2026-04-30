# Comprehensive Status Report — art-dupl

**Date:** 2026-04-16 08:54  
**Branch:** `fork` (6 commits ahead of `origin/fork` — needs push)  
**Head Commit:** `1142c06` docs(status): add code deduplication session status report  
**Working Tree:** CLEAN  
**Build Status:** PASSING ✅  
**Total Production Go LOC:** ~19,153  
**Go Version:** 1.26.0 (arm64/darwin)

---

## A) FULLY DONE ✅

### Major Milestone: Filter Refactoring (Session 2026-04-15)

| Item                                                                | Commit    | Details                                  |
| ------------------------------------------------------------------- | --------- | ---------------------------------------- |
| Remove `pkg/filter/` wrapper (5 files)                              | `72b3305` | Direct gogenfilter imports everywhere    |
| Add `--include-protobuf`, `--include-mockgen`, `--include-stringer` | `72b3305` | CLI flags + config fields + filter setup |
| Fix nil-interface wrapping bug                                      | `72b3305` | `FindSQLCConfigs` typed nil gotcha       |
| Fix include pattern test                                            | `72b3305` | `vendor/*` → `**/vendor/*`               |
| Refactor `ByteRangeToLines`                                         | `72b3305` | Exclusive-end convention                 |
| Fix JSON printer `LineEnd`                                          | `72b3305` | Was hardcoded 0                          |

### Code Deduplication Session (Commits `2af5621..16f7b7f`)

| Item                                            | Commit               | Details                                                   |
| ----------------------------------------------- | -------------------- | --------------------------------------------------------- |
| Extract `withThreshold` helper                  | `2201f5d`            | Eliminated anonymous function duplication in run_analysis |
| Replace syntax.Node literals with testutil      | `0e203b2`            | Centralized node creation                                 |
| Add `newTestClone` helper                       | `015e4b7`            | Reduced diff_test duplication                             |
| Replace local assertion helpers with testutil   | `57e3ee8`            | Centralized assertions                                    |
| Replace duplicated test loggers with NoOpLogger | `470a5fa`            | Used pkg/logger consistently                              |
| Use `isEmptyOrLessThanEmpty` in sort            | `16f7b7f`            | Reduced code duplication                                  |
| Centralize config field assertions              | `2af5621`            | Test utility consolidation                                |
| Centralize assertion patterns                   | `0f418cb`            | Test utility consolidation                                |
| Extract boolean/pattern/timeout flag helpers    | `b14eba0`            | Reduced gocyclo in config_builder                         |
| Extract PanicRecovery                           | `cd315ba`, `84d2cd0` | Shared test helper                                        |
| Table-driven getTokenRange                      | `b2a829a`            | Replaced switch statement                                 |
| Add unit tests for errors, detector, printer    | `683a27f`            | Coverage improvement                                      |

### Ongoing (From Previous Months)

| Category                                               | Status  |
| ------------------------------------------------------ | ------- |
| Multi-method detection (suffix tree + hash)            | ✅ DONE |
| Professional CLI via Fang framework                    | ✅ DONE |
| All output formats (text, HTML, JSON, plumbing, SARIF) | ✅ DONE |
| Statistics subcommand (text, JSON, CSV)                | ✅ DONE |
| SQLC auto-detection & filtering                        | ✅ DONE |
| Templ filtering                                        | ✅ DONE |
| Protobuf/mockgen/stringer filtering                    | ✅ DONE |
| Sorting (size, occurrence, hash, total tokens)         | ✅ DONE |
| Semantic detection (ON by default)                     | ✅ DONE |
| Configuration files (JSON)                             | ✅ DONE |
| BDD test suite (Ginkgo/Gomega, 234/240 pass)           | ✅ DONE |
| Domain types with strong typing                        | ✅ DONE |
| SIMD optimizations                                     | ✅ DONE |
| `.golangci.yml` with 90+ linters                       | ✅ DONE |
| gogenfilter extraction to separate repo                | ✅ DONE |

---

## B) PARTIALLY DONE 🔶

| Item               | Progress    | Details                                                                                                                    |
| ------------------ | ----------- | -------------------------------------------------------------------------------------------------------------------------- |
| Lint compliance    | 98%         | Down from 21+ warnings to 4. Remaining: gci (1), golines (1), nlreturn (1), wsl_v5 (1)                                     |
| BDD test suite     | 97.5%       | 234/240 pass. Same 6 pre-existing failures (5 filter, 1 CLI docs)                                                          |
| Test coverage      | ~75% avg    | domain 97%, suffixtree 91%, simd 95.8%, position 100%, format 100%. Gaps: printer 63.3%, syntax 67.6%, hash 70%, cli 62.5% |
| Documentation      | 85%         | FEATURES.md 2 months stale, HOW_TO_USE.md 2 months stale, README needs semantic default update. AGENTS.md updated.         |
| Flag deduplication | NOT STARTED | 18 flags duplicated between cmd/flags.go and cmd/stats.go                                                                  |
| TODO_LIST.md       | STALE       | Last updated 2026-04-05, doesn't reflect any recent work                                                                   |

### Ghost Packages (Pending Decision)

| Package                   | Lines | Imported By     | Recommendation                                   | Status                |
| ------------------------- | ----- | --------------- | ------------------------------------------------ | --------------------- |
| `pkg/errors/`             | 24    | ZERO            | Delete (pure duplicate of `errors/`)             | Pending user decision |
| `testutils/`              | 173   | ZERO            | Delete (pure duplicate of `internal/testutil/`)  | Pending user decision |
| `internal/enum/`          | 684   | ZERO            | **Decide**: integrate or extract to shared lib?  | Pending user decision |
| `git/`                    | 1,064 | ZERO non-test   | **Decide**: wire into `--incremental` or delete? | Pending user decision |
| `migration/` + `adapter/` | 1,239 | Only each other | **Decide**: keep for future config migrations?   | Pending user decision |

---

## C) NOT STARTED ⬜

| #   | Item                                                                    | Priority | Impact | Effort |
| --- | ----------------------------------------------------------------------- | -------- | ------ | ------ |
| 1   | Decide fate of 4 "maybe" ghost packages                                 | HIGH     | HIGH   | LOW    |
| 2   | Delete 2 unambiguous ghost packages (`pkg/errors/`, `testutils/`)       | HIGH     | MEDIUM | LOW    |
| 3   | Fix 6 pre-existing BDD test failures                                    | HIGH     | HIGH   | MEDIUM |
| 4   | Extract shared `addCommonFlags()` for flag deduplication                | HIGH     | MEDIUM | MEDIUM |
| 5   | Update FEATURES.md with protobuf/mockgen/stringer + all recent features | HIGH     | HIGH   | LOW    |
| 6   | Update HOW_TO_USE.md with new flags and examples                        | HIGH     | MEDIUM | LOW    |
| 7   | Update README.md with semantic default behavior                         | MEDIUM   | MEDIUM | LOW    |
| 8   | Update TODO_LIST.md                                                     | MEDIUM   | LOW    | LOW    |
| 9   | Fix remaining 4 lint warnings                                           | LOW      | LOW    | LOW    |
| 10  | Add BDD tests for --include-protobuf/mockgen/stringer                   | MEDIUM   | HIGH   | MEDIUM |
| 11  | Improve printer test coverage (63% → 80%)                               | MEDIUM   | MEDIUM | MEDIUM |
| 12  | Improve syntax test coverage (67% → 80%)                                | MEDIUM   | MEDIUM | MEDIUM |
| 13  | Implement TokenValue type                                               | HIGH     | HIGH   | HIGH   |
| 14  | Fix SIMD race condition                                                 | MEDIUM   | MEDIUM | MEDIUM |
| 15  | Create `go.work` for multi-module development                           | LOW      | LOW    | LOW    |
| 16  | Fix justfile `build-all` (missing source path)                          | MEDIUM   | LOW    | LOW    |

---

## D) TOTALLY FUCKED UP 💀

**Nothing.** Clean sessions throughout. All changes tested before commit. No broken builds, no reverted commits.

### Historical Issues Still Present

| Issue                       | Severity | Age            | Details                                                                        |
| --------------------------- | -------- | -------------- | ------------------------------------------------------------------------------ |
| 6 BDD test failures         | MEDIUM   | Months         | Pre-existing, not caused by our changes. Filter pattern tests + CLI docs test. |
| SIMD race condition         | MEDIUM   | Months         | In `internal/simd/`, pre-existing                                              |
| `justfile build-all` broken | LOW      | Since creation | Missing `./cmd/art-dupl` source path on cross-compile targets                  |
| `justfile test` double-run  | LOW      | Since creation | Runs tests twice on success (silent then verbose)                              |

---

## E) WHAT WE SHOULD IMPROVE 🔧

### Critical

1. **Resolve ghost package decisions** — We've been debating for 2 sessions now. The 4 "maybe" packages (3,184 lines) need a yes/no decision. Each day they sit is tech debt accruing interest.

2. **Fix the 6 BDD failures** — They've been failing for months. Every CI run shows red. It erodes confidence and masks real regressions. The filter tests likely just need updating after `--filter-generated` semantics evolved.

3. **Update documentation** — FEATURES.md is 2 months stale. Users can't discover the new protobuf/mockgen/stringer filters. This is direct customer value lost.

### Important

4. **Flag deduplication** — 18 duplicated flag definitions. Every new flag requires 2 edits. We WILL forget eventually (we almost did with protobuf/mockgen/stringer).

5. **Printer coverage gap** (63.3%) — The package handling all output formats is the least tested. Output bugs go directly to users.

6. **Syntax coverage gap** (67.6%) — Core AST processing. Bugs here affect all detection.

### Nice-to-Have

7. **Remaining 4 lint warnings** — Trivial to fix, just haven't gotten to them.
8. **Justfile fixes** — `build-all` broken, `test` double-runs. Developer experience.
9. **go.work** — Multi-module setup with gogenfilter.
10. **TODO_LIST.md update** — Currently stale and misleading.

---

## F) TOP 25 THINGS WE SHOULD GET DONE NEXT

### Tier 1: Decisions + Quick Wins (1-10)

| #   | Task                                                                    | Impact | Effort | Est. Time | Customer Value    |
| --- | ----------------------------------------------------------------------- | ------ | ------ | --------- | ----------------- |
| 1   | **DECIDE**: Keep or delete `internal/enum/` (684 lines)                 | HIGH   | ZERO   | 2min      | Unblocks cleanup  |
| 2   | **DECIDE**: Wire `git/` into `--incremental` or delete (1,064 lines)    | HIGH   | ZERO   | 2min      | Unblocks cleanup  |
| 3   | **DECIDE**: Keep `migration/` + `adapter/` for future use (1,239 lines) | MEDIUM | ZERO   | 2min      | Unblocks cleanup  |
| 4   | Delete `pkg/errors/` (24 lines, zero imports, pure duplicate)           | HIGH   | LOW    | 5min      | Codebase clarity  |
| 5   | Delete `testutils/` (173 lines, zero imports, pure duplicate)           | HIGH   | LOW    | 5min      | Codebase clarity  |
| 6   | Fix 4 remaining lint warnings (gci, golines, nlreturn, wsl_v5)          | LOW    | LOW    | 10min     | CI green          |
| 7   | Update FEATURES.md with all new features                                | HIGH   | LOW    | 20min     | User discovery    |
| 8   | Update HOW_TO_USE.md with new filter flags                              | HIGH   | LOW    | 15min     | User education    |
| 9   | Update README.md with semantic default                                  | MEDIUM | LOW    | 10min     | First impression  |
| 10  | Update TODO_LIST.md to reflect reality                                  | MEDIUM | LOW    | 10min     | Planning accuracy |

### Tier 2: Structural Improvements (11-18)

| #   | Task                                                         | Impact | Effort | Est. Time | Customer Value       |
| --- | ------------------------------------------------------------ | ------ | ------ | --------- | -------------------- |
| 11  | Extract `addCommonFlags()` — deduplicate 18 flag definitions | HIGH   | MEDIUM | 45min     | Maintainability      |
| 12  | Fix 6 pre-existing BDD test failures                         | HIGH   | HIGH   | 90min     | Test reliability     |
| 13  | Add BDD tests for --include-protobuf/mockgen/stringer        | HIGH   | MEDIUM | 45min     | Feature coverage     |
| 14  | Improve printer test coverage (63% → 80%)                    | MEDIUM | MEDIUM | 60min     | Reliability          |
| 15  | Improve syntax test coverage (67% → 80%)                     | MEDIUM | MEDIUM | 60min     | Reliability          |
| 16  | Fix justfile `build-all` and `test` double-run               | MEDIUM | LOW    | 15min     | Developer experience |
| 17  | Create `go.work` for multi-module setup                      | LOW    | LOW    | 5min      | Developer experience |
| 18  | Fix SIMD race condition                                      | MEDIUM | MEDIUM | 60min     | Stability            |

### Tier 3: Future Work (19-25)

| #   | Task                                                      | Impact | Effort | Est. Time | Customer Value  |
| --- | --------------------------------------------------------- | ------ | ------ | --------- | --------------- |
| 19  | Implement TokenValue type with validation                 | HIGH   | HIGH   | 4h        | Type safety     |
| 20  | Consolidate clone representations (SB-1 split brain)      | HIGH   | HIGH   | 4h        | Architecture    |
| 21  | Unify detection pipelines (SB-2 split brain)              | HIGH   | HIGH   | 4h        | Architecture    |
| 22  | Extract shared printer clone-info prep (SB-3 split brain) | MEDIUM | MEDIUM | 2h        | Maintainability |
| 23  | Implement CSV output via encoding/csv                     | LOW    | MEDIUM | 2h        | Robustness      |
| 24  | Optimize memory layouts for SIMD                          | LOW    | HIGH   | 4h        | Performance     |
| 25  | Add Go Example\* functions for godoc                      | LOW    | MEDIUM | 3h        | Documentation   |

---

## G) TOP #1 QUESTION

**What should we do with the 4 "maybe" ghost packages?**

I need your decision on each:

1. **`internal/enum/`** (684 lines) — Generic enum marshaling utilities. Zero imports. Options:
   - **(a)** Delete — no one uses it
   - **(b)** Integrate into types that need enum handling (config, domain)
   - **(c)** Extract to a shared lib for your other Go projects

2. **`git/`** (1,064 lines) — Complete git change detection system. Zero production imports. Options:
   - **(a)** Delete — it's in git history if we ever need it
   - **(b)** Wire into `--incremental` flag (the flag exists but doesn't use this code)

3. **`migration/` + `adapter/`** (1,239 lines) — Config migration framework + printer adapter. Only depend on each other. Options:
   - **(a)** Delete — no current need
   - **(b)** Keep for future config format migrations
   - **(c)** Keep adapter/ only (it's a useful abstraction), delete migration/

I cannot proceed with cleanup without your call on these.

---

## Test Results Summary

### Full Test Suite: ALL PASS (except 6 pre-existing BDD failures)

| Package             | Coverage  | Status                  |
| ------------------- | --------- | ----------------------- |
| adapter             | 97.7%     | ✅                      |
| cache               | 87.4%     | ✅                      |
| cli                 | 62.5%     | ✅                      |
| cmd                 | 75.9%     | ✅                      |
| config              | 70.4%     | ✅                      |
| detection           | 83.6%     | ✅                      |
| domain              | 97.0%     | ✅                      |
| errors              | 89.4%     | ✅                      |
| examples            | 0.0%      | ✅                      |
| git                 | 83.1%     | ✅ (but unused)         |
| hash                | 70.0%     | ✅                      |
| internal/configtest | —         | ✅                      |
| internal/enum       | 76.6%     | ✅ (but unused)         |
| internal/filtertest | 41.2%     | ✅                      |
| internal/simd       | 95.8%     | ✅ (race fails)         |
| internal/utils      | 93.3%     | ✅                      |
| job                 | 76.7%     | ✅                      |
| migration           | 83.3%     | ✅ (but unused)         |
| pkg/artdupl         | 85.2%     | ✅                      |
| pkg/format          | 100.0%    | ✅                      |
| pkg/logger          | 87.5%     | ✅                      |
| pkg/position        | 100.0%    | ✅                      |
| printer             | 63.3%     | ✅                      |
| suffixtree          | 91.0%     | ✅                      |
| syntax              | 67.6%     | ✅                      |
| syntax/golang       | 93.9%     | ✅                      |
| syntax/templ        | 85.3%     | ✅                      |
| **bdd**             | **70.0%** | **⚠️ 234 pass, 6 fail** |

### Lint: 4 warnings remaining

- `printer/stats.go:43` — gci formatting
- `internal/filtertest/assertions.go:46` — golines line length
- `detection/detection_test.go:285` — nlreturn
- `printer/sorter.go:157` — wsl_v5

### Build: PASS ✅ | `go vet`: PASS ✅ | Working Tree: CLEAN ✅

### Git: 6 commits ahead of origin — NEEDS PUSH
