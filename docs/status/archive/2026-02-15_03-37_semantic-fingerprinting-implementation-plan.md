# Semantic Fingerprinting Implementation Plan

**Created**: 2026-02-15 03:37  
**Total Tasks**: 40  
**Est. Total Time**: ~6 hours (40 tasks × avg 9 min)

---

## Priority Matrix

Sorted by: **Impact → Customer Value → Effort (ascending)**

| #                                      | Task                                               | Phase | Impact | Effort | Time | Dependencies |
| -------------------------------------- | -------------------------------------------------- | ----- | ------ | ------ | ---- | ------------ |
| **CORE IMPLEMENTATION (Must Have)**    |
| 1                                      | Analyze transform.go for identifier capture points | 1     | HIGH   | LOW    | 8m   | -            |
| 2                                      | Design identifier hash function (avoid collisions) | 1     | HIGH   | LOW    | 10m  | 1            |
| 3                                      | Create identifier hash utility function            | 1     | HIGH   | LOW    | 8m   | 2            |
| 4                                      | Modify SelectorExpr case to include selector name  | 1     | HIGH   | LOW    | 8m   | 3            |
| 5                                      | Modify CallExpr case to include function name      | 1     | HIGH   | LOW    | 8m   | 3            |
| 6                                      | Modify Ident case to include identifier name       | 1     | HIGH   | LOW    | 8m   | 3            |
| **POST-MATCH FILTERING (Should Have)** |
| 7                                      | Add semantic filtering types to semantic_filter.go | 2     | HIGH   | MED    | 10m  | -            |
| 8                                      | Implement extractMethodNames helper                | 2     | HIGH   | LOW    | 10m  | 7            |
| 9                                      | Implement extractIdentifiers helper                | 2     | HIGH   | LOW    | 8m   | 7            |
| 10                                     | Implement intersection helper for string slices    | 2     | MED    | LOW    | 5m   | -            |
| 11                                     | Implement IsValidDuplicate with generic filtering  | 2     | HIGH   | MED    | 12m  | 8,9,10       |
| **CONFIGURATION (Must Have)**          |
| 12                                     | Add SemanticThreshold field to Config              | 3     | HIGH   | LOW    | 5m   | -            |
| 13                                     | Add SemanticEnabled convenience field              | 3     | MED    | LOW    | 3m   | 12           |
| 14                                     | Update DefaultConfig() with default (0.0)          | 3     | HIGH   | LOW    | 3m   | 12           |
| 15                                     | Add SemanticThreshold JSON validation              | 3     | HIGH   | LOW    | 5m   | 12           |
| 16                                     | Update ValidateConfig() for range check            | 3     | HIGH   | LOW    | 5m   | 15           |
| **CLI INTEGRATION (Must Have)**        |
| 17                                     | Add --semantic-threshold CLI flag                  | 3     | HIGH   | LOW    | 8m   | 12           |
| 18                                     | Wire semantic config to detector pipeline          | 3     | HIGH   | MED    | 10m  | 11,16        |
| 19                                     | Integrate SemanticFilter into FindSyntaxUnits()    | 3     | HIGH   | MED    | 12m  | 11,18        |
| **LITERAL HANDLING (Optional)**        |
| 20                                     | Modify BasicLit case for literal value hash        | 4     | MED    | MED    | 10m  | 3            |
| **UNIT TESTS (Must Have)**             |
| 21                                     | Write unit test for identifier hash function       | 5     | HIGH   | LOW    | 8m   | 3            |
| 22                                     | Write unit test for extractMethodNames             | 5     | HIGH   | LOW    | 8m   | 8            |
| 23                                     | Write unit test for IsValidDuplicate               | 5     | HIGH   | MED    | 10m  | 11           |
| **INTEGRATION TESTS (Must Have)**      |
| 24                                     | Create test fixture: Ginkgo different methods      | 5     | HIGH   | LOW    | 8m   | -            |
| 25                                     | Create test fixture: Renamed variables             | 5     | MED    | LOW    | 8m   | -            |
| 26                                     | Write integration test: semantic disabled          | 5     | HIGH   | LOW    | 8m   | 24           |
| 27                                     | Write integration test: semantic at 0.8            | 5     | HIGH   | MED    | 10m  | 24,26        |
| 28                                     | Write integration test: semantic at 1.0            | 5     | MED    | LOW    | 8m   | 24,26        |
| **PERFORMANCE (Should Have)**          |
| 29                                     | Add benchmark for identifier hashing               | 6     | MED    | LOW    | 8m   | 6            |
| 30                                     | Add benchmark for semantic filtering               | 6     | MED    | LOW    | 8m   | 11           |
| 31                                     | Profile memory impact                              | 6     | MED    | MED    | 12m  | 29,30        |
| 32                                     | Profile CPU impact                                 | 6     | MED    | MED    | 12m  | 29,30        |
| **DOCUMENTATION (Should Have)**        |
| 33                                     | Update README with semantic filtering              | 7     | MED    | LOW    | 10m  | -            |
| 34                                     | Update HOW_TO_USE with examples                    | 7     | MED    | LOW    | 10m  | -            |
| 35                                     | Add section to FEATURES.md                         | 7     | MED    | LOW    | 8m   | -            |
| 36                                     | Update status report with progress                 | 7     | LOW    | LOW    | 5m   | -            |
| **VALIDATION (Must Have)**             |
| 37                                     | Run full test suite for regressions                | 8     | HIGH   | LOW    | 10m  | ALL          |
| 38                                     | Test on real-world codebase (BuildFlow)            | 8     | HIGH   | MED    | 12m  | 37           |
| 39                                     | Tune generic method list based on results          | 8     | MED    | MED    | 10m  | 38           |
| 40                                     | Final review and cleanup                           | 8     | HIGH   | LOW    | 12m  | ALL          |

---

## Phase Overview

| Phase | Name                 | Tasks | Time | Status  |
| ----- | -------------------- | ----- | ---- | ------- |
| 1     | Core Implementation  | 1-6   | 50m  | PENDING |
| 2     | Post-Match Filtering | 7-11  | 45m  | PENDING |
| 3     | Configuration & CLI  | 12-19 | 56m  | PENDING |
| 4     | Literal Handling     | 20    | 10m  | PENDING |
| 5     | Testing              | 21-28 | 68m  | PENDING |
| 6     | Performance          | 29-32 | 40m  | PENDING |
| 7     | Documentation        | 33-36 | 33m  | PENDING |
| 8     | Validation           | 37-40 | 44m  | PENDING |

---

## Task Details

### Phase 1: Core Implementation (50 min)

**Goal**: Hash identifiers into token types to differentiate structurally similar code.

#### Task 1: Analyze transform.go (8 min)

- **File**: `syntax/golang/transform.go`
- **Action**: Identify all AST nodes that contain identifiers/literals
- **Output**: List of case statements to modify
- **Est. lines**: ~5 lines of notes

#### Task 2: Design hash function (10 min)

- **Criteria**:
  - Fast (no crypto overhead)
  - Stable (same input = same output)
  - Low collision rate
  - Fits in int32 (current Node.Type)
- **Decision**: Use FNV-1a or xxh3 truncated

#### Task 3: Create hash utility (8 min)

- **File**: `syntax/golang/identifier_hash.go` (new)
- **Function**: `func hashIdentifier(name string) int32`
- **Est. lines**: ~15 lines

#### Task 4: Modify SelectorExpr (8 min)

- **File**: `syntax/golang/transform.go:212-214`
- **Change**: Fold `n.Sel.Name` hash into `o.Type`
- **Est. lines**: ~3 lines changed

#### Task 5: Modify CallExpr (8 min)

- **File**: `syntax/golang/transform.go:50-55`
- **Change**: Fold function name hash into `o.Type`
- **Est. lines**: ~5 lines changed

#### Task 6: Modify Ident (8 min)

- **File**: `syntax/golang/transform.go:158-159`
- **Change**: Fold identifier name hash into `o.Type`
- **Est. lines**: ~3 lines changed

---

### Phase 2: Post-Match Filtering (45 min)

**Goal**: Filter structural matches that lack semantic similarity.

#### Task 7: Create semantic_filter.go types (10 min)

- **File**: `syntax/semantic_filter.go` (new)
- **Types**:
  ```go
  type SemanticFilter struct {
      GenericMethods map[string]bool
      Threshold      float64
  }
  ```
- **Est. lines**: ~20 lines

#### Task 8: Implement extractMethodNames (10 min)

- **Function**: `func extractMethodNames(nodes []*Node) []string`
- **Logic**: Walk nodes, extract CallExpr function names
- **Est. lines**: ~25 lines

#### Task 9: Implement extractIdentifiers (8 min)

- **Function**: `func extractIdentifiers(nodes []*Node) []string`
- **Logic**: Walk nodes, extract Ident names
- **Est. lines**: ~20 lines

#### Task 10: Implement intersection (5 min)

- **Function**: `func intersection(a, b []string) []string`
- **Est. lines**: ~10 lines

#### Task 11: Implement IsValidDuplicate (12 min)

- **Method**: `func (sf *SemanticFilter) IsValidDuplicate(a, b []*Node) bool`
- **Logic**:
  1. Extract method names from both
  2. Find shared methods
  3. Filter out generic methods
  4. Return true if any shared non-generic methods
- **Est. lines**: ~30 lines

---

### Phase 3: Configuration & CLI (56 min)

**Goal**: Make semantic filtering configurable.

#### Task 12: Add SemanticThreshold field (5 min)

- **File**: `config/config.go`
- **Field**: `SemanticThreshold float64`
- **Est. lines**: 1 line

#### Task 13: Add SemanticEnabled field (3 min)

- **Field**: `SemanticEnabled bool` (convenience, derived from threshold > 0)
- **Est. lines**: 1 line

#### Task 14: Update DefaultConfig (3 min)

- **File**: `config/config.go:119-145`
- **Add**: `SemanticThreshold: 0.0`
- **Est. lines**: 1 line

#### Task 15: Add JSON validation (5 min)

- **File**: `config/config.go`
- **Logic**: Validate 0.0 <= threshold <= 1.0
- **Est. lines**: ~5 lines

#### Task 16: Update ValidateConfig (5 min)

- **File**: `config/config.go:220-253`
- **Add**: Threshold range check
- **Est. lines**: ~5 lines

#### Task 17: Add CLI flag (8 min)

- **File**: `cmd/art-dupl/root.go`
- **Flag**: `--semantic-threshold, -s float`
- **Est. lines**: ~5 lines

#### Task 18: Wire to pipeline (10 min)

- **File**: `pkg/artdupl/detector_pipeline.go`
- **Logic**: Pass config to detection, create filter if enabled
- **Est. lines**: ~15 lines

#### Task 19: Integrate into FindSyntaxUnits (12 min)

- **File**: `syntax/syntax.go:119-166`
- **Logic**: Filter matches through SemanticFilter before returning
- **Est. lines**: ~20 lines

---

### Phase 4: Literal Handling (10 min)

**Goal**: Optionally include literals in semantic hash.

#### Task 20: Modify BasicLit (10 min)

- **File**: `syntax/golang/transform.go:33-34`
- **Change**: Hash literal value into type (configurable)
- **Note**: Only if `--include-literals` flag set
- **Est. lines**: ~8 lines

---

### Phase 5: Testing (68 min)

**Goal**: Comprehensive test coverage for new functionality.

#### Task 21: Hash function unit test (8 min)

- **File**: `syntax/golang/identifier_hash_test.go`
- **Cases**: Empty, short, long, special chars, collision check
- **Est. lines**: ~30 lines

#### Task 22: extractMethodNames test (8 min)

- **File**: `syntax/semantic_filter_test.go`
- **Cases**: No calls, single call, nested calls, selectors
- **Est. lines**: ~40 lines

#### Task 23: IsValidDuplicate test (10 min)

- **File**: `syntax/semantic_filter_test.go`
- **Cases**:
  - Same methods → true
  - Different methods → false
  - Only generic methods → false
  - Mixed → true
- **Est. lines**: ~60 lines

#### Task 24: Ginkgo fixture (8 min)

- **File**: `internal/testutil/fixtures/ginkgo_tests/`
- **Content**: Two test blocks with different methods
- **Est. lines**: ~30 lines

#### Task 25: Renamed variable fixture (8 min)

- **File**: `internal/testutil/fixtures/renamed_vars/`
- **Content**: Copy-pasted code with variable renames
- **Est. lines**: ~30 lines

#### Task 26: Integration test disabled (8 min)

- **File**: `bdd/semantic_filter_test.go`
- **Case**: Default (threshold 0.0) should match old behavior
- **Est. lines**: ~30 lines

#### Task 27: Integration test 0.8 (10 min)

- **Case**: Threshold 0.8 should filter Ginkgo false positives
- **Est. lines**: ~40 lines

#### Task 28: Integration test 1.0 (8 min)

- **Case**: Threshold 1.0 should only match exact semantic duplicates
- **Est. lines**: ~30 lines

---

### Phase 6: Performance (40 min)

**Goal**: Verify performance impact is acceptable.

#### Task 29: Hash benchmark (8 min)

- **File**: `syntax/golang/identifier_hash_bench_test.go`
- **Cases**: Single identifier, 100, 1000, 10000
- **Est. lines**: ~30 lines

#### Task 30: Filter benchmark (8 min)

- **File**: `syntax/semantic_filter_bench_test.go`
- **Cases**: Small, medium, large node sequences
- **Est. lines**: ~30 lines

#### Task 31: Memory profile (12 min)

- **Action**: Run with `-memprofile`, analyze with pprof
- **Goal**: <5% memory increase
- **Output**: Profile results in docs

#### Task 32: CPU profile (12 min)

- **Action**: Run with `-cpuprofile`, analyze with pprof
- **Goal**: <10% CPU increase
- **Output**: Profile results in docs

---

### Phase 7: Documentation (33 min)

**Goal**: Document new feature for users.

#### Task 33: Update README (10 min)

- **Section**: Add "Semantic Filtering" under features
- **Content**: What it is, how to use, examples
- **Est. lines**: ~30 lines

#### Task 34: Update HOW_TO_USE (10 min)

- **Section**: Add semantic threshold examples
- **Content**: Common use cases with command examples
- **Est. lines**: ~40 lines

#### Task 35: Update FEATURES.md (8 min)

- **Section**: Add to feature list
- **Content**: Bullet points with benefits
- **Est. lines**: ~15 lines

#### Task 36: Update status report (5 min)

- **File**: `docs/status/2026-02-15_03-35_...`
- **Content**: Mark completed, add results
- **Est. lines**: ~10 lines

---

### Phase 8: Validation (44 min)

**Goal**: Ensure quality and real-world effectiveness.

#### Task 37: Run full test suite (10 min)

- **Command**: `just ci` or `just test`
- **Goal**: 100% pass rate, no regressions
- **Action**: Fix any failures before proceeding

#### Task 38: Test on BuildFlow (12 min)

- **Action**: Run art-dupl on BuildFlow codebase
- **Compare**: Results with/without semantic filtering
- **Goal**: False positive reduction >80%

#### Task 39: Tune generic methods (10 min)

- **Action**: Adjust default generic method list based on results
- **Goal**: Balance precision vs recall
- **Est. changes**: ~5-10 methods

#### Task 40: Final review (12 min)

- **Checklist**:
  - [ ] All tests pass
  - [ ] Documentation complete
  - [ ] Performance acceptable
  - [ ] Real-world testing done
  - [ ] Code reviewed for cleanup
  - [ ] Commit ready

---

## Success Criteria

| Metric                   | Target       | How to Measure                    |
| ------------------------ | ------------ | --------------------------------- |
| False positive reduction | >80%         | Run on BuildFlow, compare results |
| True positive retention  | >90%         | Manual review of filtered results |
| Performance overhead     | <10%         | Benchmark before/after            |
| Test coverage            | >80%         | `go test -cover`                  |
| User satisfaction        | Configurable | Threshold 0.0-1.0                 |

---

## Risk Mitigation

| Risk                   | Probability | Impact | Mitigation                        |
| ---------------------- | ----------- | ------ | --------------------------------- |
| Hash collisions        | Low         | Medium | Use xxh3, test edge cases         |
| Performance regression | Medium      | High   | Profile, benchmark, optimize      |
| Missed true duplicates | Medium      | High   | Tune threshold, test on real code |
| Complex config         | Low         | Medium | Good defaults, clear docs         |

---

## Next Action

**Start with Task 1**: Analyze transform.go for identifier capture points

```bash
# After completing all tasks
just ci  # Run full validation
```
