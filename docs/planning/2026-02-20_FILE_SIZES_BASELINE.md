# File Sizes Baseline - 2026-02-20

**Status**: 52 files exceed 250 lines (HOW_TO_GOLANG.md limit)

## Strategy

Focus on **"easy wins"** first (files 250-290 lines):

- Lower risk
- Less time investment
- Build confidence
- Validate splitting approach

## Top 18 Easy Wins (250-290 lines)

| #  | File                                      | Lines | Type | Strategy                      |
| -- | ----------------------------------------- | ----- | ---- | ----------------------------- |
| 1  | suffixtree/suffixtree_bench_test.go       | 252   | Test | Extract benchmark functions   |
| 2  | syntax/templ/templ.go                     | 254   | Code | Extract parser functions      |
| 3  | errors/types.go                           | 257   | Code | Split error types by category |
| 4  | suffixtree/suffixtree.go                  | 257   | Code | Core algorithm - careful      |
| 5  | bdd/detection_methods_test.go             | 264   | Test | Split test scenarios          |
| 6  | cmd/stats.go                              | 268   | Code | Extract subcommand handlers   |
| 7  | migration/migration.go                    | 268   | Code | Split migration types         |
| 8  | config/config.go                          | 269   | Code | Extract config sections       |
| 9  | internal/testutil/bdd_helpers.go          | 269   | Code | Organize by helper type       |
| 10 | bdd/sorting_test.go                       | 271   | Test | Split test scenarios          |
| 11 | examples/domain_types_usage.go            | 274   | Code | Split by example type         |
| 12 | syntax/golang/transform.go                | 274   | Code | Extract transform functions   |
| 13 | printer/stats_formatter.go                | 276   | Code | Split formatters              |
| 14 | config/detectionmethod.go                 | 282   | Code | Split method types            |
| 15 | internal/filtertest/user_scenario_test.go | 282   | Test | Split scenarios               |
| 16 | bdd/all_format_generation_test.go         | 285   | Test | Split format tests            |
| 17 | bdd/semantic_detection_test.go            | 289   | Test | Split semantic tests          |
| 18 | examples/examples_test.go                 | 290   | Test | Split example tests           |

## Major Files (300+ lines) - Phase 3

| File                         | Lines | Priority |
| ---------------------------- | ----- | -------- |
| syntax/golang/parse_test.go  | 1,284 | P3       |
| domain/coverage_test.go      | 1,281 | P3       |
| pkg/artdupl/detector_test.go | 1,252 | P3       |
| cmd/cmd_test.go              | 1,121 | P3       |
| domain/domain_types_test.go  | 870   | P3       |

## Progress Tracking

- [ ] Phase 1: 18 easy wins (252-290 lines)
- [ ] Phase 2: 14 medium files (290-350 lines)
- [ ] Phase 3: 5 major files (350+ lines)
- [ ] Phase 4: 15 large files (350-500 lines)

**Total**: 52 files → Target: All < 250 lines
