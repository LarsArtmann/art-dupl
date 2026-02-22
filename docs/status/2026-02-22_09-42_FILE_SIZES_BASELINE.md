# File Sizes Baseline - Pareto Optimization

**Date**: 2026-02-22 09:42 CET
**Status**: Starting Pareto Optimization Phase 1
**Goal**: All files < 250 lines (HOW_TO_GOLANG.md compliance)

## Summary

- **Total files over 250 lines**: 51
- **Build status**: ✅ PASS
- **Test status**: ✅ PASS (all packages)

## Files by Category

### Easy Wins (250-280 lines) - 10 files

| File | Lines | Priority |
|------|-------|----------|
| suffixtree/suffixtree_bench_test.go | 252 | 1 |
| syntax/templ/templ.go | 254 | 2 |
| errors/types.go | 257 | 3 |
| suffixtree/suffixtree.go | 257 | 4 |
| bdd/detection_methods_test.go | 264 | 5 |
| cmd/stats.go | 268 | 6 |
| migration/migration.go | 268 | 7 |
| config/config.go | 269 | 8 |
| internal/testutil/bdd_helpers.go | 269 | 9 |
| bdd/sorting_test.go | 271 | 10 |

### Medium Files (280-400 lines) - 16 files

| File | Lines |
|------|-------|
| printer/stats_formatter.go | 276 |
| syntax/golang/transform.go | 274 |
| examples/domain_types_usage.go | 274 |
| config/detectionmethod.go | 282 |
| internal/filtertest/user_scenario_test.go | 282 |
| bdd/all_format_generation_test.go | 285 |
| bdd/semantic_detection_test.go | 289 |
| examples/examples_test.go | 290 |
| cache/file_cache.go | 312 |
| detection/todos.go | 314 |
| cmd/stats_integration_test.go | 316 |
| printer/html.go | 336 |
| bdd/error_handling_test.go | 337 |
| suffixtree/suffixtree_test.go | 342 |
| syntax/golang/identifier_hash_test.go | 358 |
| git/change_detector.go | 361 |

### Larger Files (400-600 lines) - 13 files

| File | Lines |
|------|-------|
| bdd/templ_clone_detection_test.go | 378 |
| bdd/plumbing_and_paths_test.go | 386 |
| internal/enum/marshal_test.go | 390 |
| bdd/default_filtering_test.go | 399 |
| bdd/incremental_detection_test.go | 402 |
| bdd/cli_commands_test.go | 411 |
| bdd/configuration_file_test.go | 423 |
| adapter/printer_adapter_test.go | 432 |
| internal/simd/simd_test.go | 443 |
| bdd/filter_features_test.go | 453 |
| bdd/stats_subcommand_test.go | 457 |
| bdd/stats_command_test.go | 477 |
| cache/file_cache_test.go | 497 |
| config/config_test.go | 497 |

### Major Files (500+ lines) - 12 files

| File | Lines |
|------|-------|
| bdd/bdd_test.go | 509 |
| bdd/plumbing_output_test.go | 568 |
| git/change_detector_test.go | 616 |
| printer/stats_test.go | 800 |
| detection/detection_test.go | 802 |
| pkg/filter/filter_test.go | 822 |
| domain/domain_types_test.go | 870 |
| cmd/cmd_test.go | 1121 |
| pkg/artdupl/detector_test.go | 1252 |
| domain/coverage_test.go | 1281 |
| syntax/golang/parse_test.go | 1284 |

## Execution Strategy

1. **Phase 1**: Easy wins (10 files, ~2 lines to remove each)
2. **Phase 2**: Medium files (16 files, 30-150 lines to remove)
3. **Phase 3**: Larger files (13 files, 150-350 lines to remove)
4. **Phase 4**: Major files (12 files, 250-1034 lines to remove)

## Progress Tracking

| Phase | Files | Status |
|-------|-------|--------|
| 1 | 10 | 🔄 In Progress |
| 2 | 16 | ⏳ Pending |
| 3 | 13 | ⏳ Pending |
| 4 | 12 | ⏳ Pending |

## Rules

1. **One file at a time** - Process, verify, commit
2. **Run tests after each split** - Must pass before proceeding
3. **Use goimports/gofumpt** - Fix imports automatically
4. **Commit after each successful split** - Small atomic commits
