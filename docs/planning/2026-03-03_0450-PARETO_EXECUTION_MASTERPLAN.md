# Pareto Execution Masterplan: Art-Dupl Optimization

**Date:** 2026-03-03 04:50 CET  
**Status:** Phase 1 Complete, Planning Phases 2-5  
**Objective:** Apply Pareto Principle to maximize impact of remaining work

---

## 🎯 Pareto Analysis: What Delivers the Most Impact?

### The 1% that Delivers 51% of Results

**The Single Highest-Impact Task Remaining:**

| Task                                   | Impact | Why                                                                                                                  |
| -------------------------------------- | ------ | -------------------------------------------------------------------------------------------------------------------- |
| **Type Safety: TokenValue Type Alias** | 51%    | Prevents entire classes of bugs, improves maintainability, enables future optimizations, makes code self-documenting |

**Details:**

- Current: `Token.Val()` returns `int`, `Pos` is `int32`, mismatch causes bugs
- Fix: Create `TokenValue int32` type alias used consistently
- Effort: 2 hours
- Impact: Prevents overflow bugs, type mismatches, enables compiler checking
- Ripple Effect: Touches suffixtree/, syntax/, detection/ packages

---

### The 4% that Delivers 64% of Results

**Four High-Impact Tasks (1% + 3 additional):**

| #   | Task                       | Impact | Cumulative | Effort | Value/Effort |
| --- | -------------------------- | ------ | ---------- | ------ | ------------ |
| 1   | TokenValue type alias      | 51%    | 51%        | 2h     | 25.5%/h      |
| 2   | README performance section | 5%     | 56%        | 30min  | 10%/h        |
| 3   | CHANGELOG for release      | 4%     | 60%        | 20min  | 12%/h        |
| 4   | Real-world benchmark CI    | 4%     | 64%        | 1h     | 4%/h         |

**Why These Four:**

1. **TokenValue**: Core type safety improvement
2. **README**: Users immediately see performance improvement
3. **CHANGELOG**: Release documentation, marketing, user communication
4. **CI Benchmarks**: Prevents regression, ensures quality

---

### The 20% that Delivers 80% of Results

**20 Tasks Delivering 80% Cumulative Impact:**

#### Core Improvements (4 tasks → 64%)

1. ✅ ~~Semantic optimization~~ (DONE - delivered 546x improvement)
2. TokenValue type alias (2h)
3. README performance section (30min)
4. CHANGELOG for release (20min)

#### Documentation (4 tasks → +6% = 70%)

5. AGENTS.md type safety section (20min)
6. PERFORMANCE.md new file (1h)
7. API documentation update (30min)
8. Code comments for complex logic (30min)

#### Testing Infrastructure (4 tasks → +5% = 75%)

9. Property-based tests for suffix tree (2h)
10. Fuzzing for tree construction (1h)
11. CI benchmark regression detection (1h)
12. Memory leak detection tests (30min)

#### Quality & Maintenance (4 tasks → +3% = 78%)

13. TokenValue validation functions (30min)
14. Transition count statistics (30min)
15. Error message improvements (30min)
16. Linting configuration fix (20min)

#### User Experience (4 tasks → +2% = 80%)

17. Progress indicator improvements (30min)
18. Better error reporting for users (30min)
19. Configuration validation (30min)
20. Quick-start guide update (30min)

---

## 📋 Comprehensive 27-Task Plan (100-30min each)

### Phase 1: Type Safety Foundation (Critical Path)

| #   | Task                                     | Duration | Impact   | Dependencies |
| --- | ---------------------------------------- | -------- | -------- | ------------ |
| 1   | Design TokenValue type with validation   | 100min   | Critical | None         |
| 2   | Create domain/tokenvalue.go with types   | 60min    | Critical | Task 1       |
| 3   | Update Token interface to use TokenValue | 90min    | Critical | Task 2       |
| 4   | Refactor suffixtree to use TokenValue    | 90min    | Critical | Task 3       |
| 5   | Refactor syntax/ to use TokenValue       | 90min    | High     | Task 3       |
| 6   | Add TokenValue marshaling for JSON       | 60min    | Medium   | Task 2       |
| 7   | Update all tests for TokenValue          | 90min    | High     | Tasks 4-5    |

### Phase 2: Documentation & Communication

| #   | Task                                 | Duration | Impact | Dependencies |
| --- | ------------------------------------ | -------- | ------ | ------------ |
| 8   | Write README performance section     | 60min    | High   | None         |
| 9   | Create CHANGELOG for vNext release   | 40min    | High   | None         |
| 10  | Create PERFORMANCE.md guide          | 100min   | Medium | None         |
| 11  | Update AGENTS.md type safety section | 40min    | Medium | Task 1       |
| 12  | Document TokenValue usage patterns   | 60min    | Medium | Task 2       |
| 13  | Update API documentation             | 60min    | Medium | None         |

### Phase 3: Testing & Quality

| #   | Task                                 | Duration | Impact | Dependencies |
| --- | ------------------------------------ | -------- | ------ | ------------ |
| 14  | Design property-based tests          | 90min    | High   | None         |
| 15  | Implement suffix tree properties     | 100min   | High   | Task 14      |
| 16  | Add fuzzing for tree construction    | 90min    | Medium | None         |
| 17  | Create CI benchmark regression check | 90min    | High   | None         |
| 18  | Add memory profiling tests           | 60min    | Medium | None         |
| 19  | Create transition count benchmarks   | 50min    | Low    | None         |

### Phase 4: User Experience

| #   | Task                         | Duration | Impact | Dependencies |
| --- | ---------------------------- | -------- | ------ | ------------ |
| 20  | Improve progress indicators  | 60min    | Low    | None         |
| 21  | Enhance error messages       | 60min    | Medium | None         |
| 22  | Add configuration validation | 60min    | Medium | None         |
| 23  | Update quick-start guide     | 40min    | Low    | None         |

### Phase 5: Memory Optimization (Evaluation)

| #   | Task                                   | Duration | Impact | Dependencies |
| --- | -------------------------------------- | -------- | ------ | ------------ |
| 24  | Profile memory usage on large codebase | 60min    | High   | None         |
| 25  | Design hybrid slice/map approach       | 100min   | Medium | Task 24      |
| 26  | Implement hybrid approach prototype    | 100min   | Medium | Task 25      |
| 27  | Benchmark hybrid vs current            | 60min    | Medium | Task 26      |

---

## 🔬 Detailed 150-Task Breakdown (Max 15min Each)

### Phase 1: Type Safety Foundation (42 tasks)

#### Task 1: Design TokenValue (8 subtasks)

| #   | Subtask                            | Duration | Details                       |
| --- | ---------------------------------- | -------- | ----------------------------- |
| 1.1 | Research Go type alias patterns    | 10min    | Check domain/ for examples    |
| 1.2 | Define TokenValue type constraints | 10min    | int32, bounds, validation     |
| 1.3 | Design validation interface        | 10min    | IsValid() error pattern       |
| 1.4 | Document design decisions          | 10min    | ADR-style documentation       |
| 1.5 | Review existing Token usage        | 15min    | Grep all Token.Val() calls    |
| 1.6 | Identify migration points          | 10min    | Map all files needing changes |
| 1.7 | Create migration plan              | 10min    | Order of changes              |
| 1.8 | Design review checklist            | 5min     | Validation criteria           |

#### Task 2: Create domain/tokenvalue.go (6 subtasks)

| #   | Subtask                | Duration | Details                     |
| --- | ---------------------- | -------- | --------------------------- |
| 2.1 | Create file header     | 5min     | Package, imports, copyright |
| 2.2 | Define TokenValue type | 5min     | type TokenValue int32       |
| 2.3 | Add Min/Max constants  | 5min     | Validation bounds           |
| 2.4 | Implement IsValid()    | 10min    | Range checking              |
| 2.5 | Add String() method    | 10min    | For debugging               |
| 2.6 | Write unit tests       | 15min    | Table-driven tests          |

#### Task 3: Update Token Interface (8 subtasks)

| #   | Subtask                     | Duration | Details                  |
| --- | --------------------------- | -------- | ------------------------ |
| 3.1 | Locate Token interface      | 5min     | Find definition          |
| 3.2 | Update interface definition | 5min     | Change return type       |
| 3.3 | Find all implementations    | 10min    | Grep for Val() methods   |
| 3.4 | Update syntax/golang token  | 10min    | Change int to TokenValue |
| 3.5 | Update syntax/templ token   | 10min    | Change int to TokenValue |
| 3.6 | Update test tokens          | 10min    | suffixtree test tokens   |
| 3.7 | Update benchmark tokens     | 10min    | Benchmark test tokens    |
| 3.8 | Verify no int usages remain | 10min    | Final grep check         |

#### Task 4: Refactor suffixtree (10 subtasks)

| #    | Subtask                        | Duration | Details          |
| ---- | ------------------------------ | -------- | ---------------- |
| 4.1  | Update state.trans map key     | 10min    | int → TokenValue |
| 4.2  | Update findTran signature      | 10min    | int → TokenValue |
| 4.3  | Update findTran implementation | 10min    | Use typed key    |
| 4.4  | Update addTran signature       | 10min    | Use TokenValue   |
| 4.5  | Update testAndSplit            | 10min    | Type conversions |
| 4.6  | Update canonize                | 10min    | Type conversions |
| 4.7  | Update dupl.go iteration       | 10min    | Sorted keys      |
| 4.8  | Update all tests               | 15min    | Test token types |
| 4.9  | Run suffixtree tests           | 10min    | Verify pass      |
| 4.10 | Fix any type errors            | 10min    | Compiler fixes   |

#### Task 5: Refactor syntax/ (6 subtasks)

| #   | Subtask                    | Duration | Details            |
| --- | -------------------------- | -------- | ------------------ |
| 5.1 | Update golang/parse.go     | 15min    | Token return types |
| 5.2 | Update golang/transform.go | 15min    | Token return types |
| 5.3 | Update templ/ parser       | 15min    | Token return types |
| 5.4 | Update syntax.go           | 10min    | Token interface    |
| 5.5 | Run syntax tests           | 10min    | Verify pass        |
| 5.6 | Fix any type errors        | 10min    | Compiler fixes     |

#### Task 6: JSON Marshaling (4 subtasks)

| #   | Subtask                 | Duration | Details                 |
| --- | ----------------------- | -------- | ----------------------- |
| 6.1 | Implement MarshalJSON   | 15min    | int32 → JSON number     |
| 6.2 | Implement UnmarshalJSON | 15min    | JSON → int32 validation |
| 6.3 | Add marshaling tests    | 15min    | Round-trip tests        |
| 6.4 | Test with real config   | 15min    | Integration test        |

### Phase 2: Documentation (24 tasks)

#### Task 8: README Performance (6 subtasks)

| #   | Subtask                   | Duration | Details                    |
| --- | ------------------------- | -------- | -------------------------- |
| 8.1 | Analyze current README    | 10min    | Find performance mentions  |
| 8.2 | Write performance section | 15min    | 546x improvement highlight |
| 8.3 | Add benchmark table       | 10min    | Before/after comparison    |
| 8.4 | Document --semantic flag  | 10min    | Now faster than default    |
| 8.5 | Add usage example         | 10min    | Quick start with flags     |
| 8.6 | Review and polish         | 5min     | Final edits                |

#### Task 9: CHANGELOG (4 subtasks)

| #   | Subtask                        | Duration | Details                 |
| --- | ------------------------------ | -------- | ----------------------- |
| 9.1 | Check current CHANGELOG format | 5min     | Understand structure    |
| 9.2 | Draft vNext section            | 10min    | List all changes        |
| 9.3 | Write performance highlights   | 10min    | 546x, 5.9x improvements |
| 9.4 | Add thanks/contributors        | 5min     | Acknowledge work        |

#### Task 10: PERFORMANCE.md (10 subtasks)

| #     | Subtask                      | Duration | Details               |
| ----- | ---------------------------- | -------- | --------------------- |
| 10.1  | Create file structure        | 10min    | Headers, sections     |
| 10.2  | Write benchmark methodology  | 10min    | How we measure        |
| 10.3  | Document current performance | 10min    | All benchmark results |
| 10.4  | Add comparison tables        | 10min    | Before/after          |
| 10.5  | Document memory usage        | 10min    | Trade-offs            |
| 10.6  | Add tuning guide             | 10min    | --threshold tips      |
| 10.7  | Document --semantic flag     | 10min    | When to use           |
| 10.8  | Add troubleshooting          | 10min    | Performance issues    |
| 10.9  | Review and edit              | 10min    | Polish                |
| 10.10 | Link from README             | 10min    | Cross-reference       |

#### Task 11: AGENTS.md Update (4 subtasks)

| #    | Subtask                  | Duration | Details                |
| ---- | ------------------------ | -------- | ---------------------- |
| 11.1 | Review current AGENTS.md | 5min     | Find relevant sections |
| 11.2 | Add type safety section  | 10min    | TokenValue guidance    |
| 11.3 | Update performance notes | 10min    | Latest optimizations   |
| 11.4 | Cross-reference docs     | 5min     | Link to PERFORMANCE.md |

### Phase 3: Testing (36 tasks)

#### Task 14: Property-Based Test Design (8 subtasks)

| #    | Subtask                              | Duration | Details                  |
| ---- | ------------------------------------ | -------- | ------------------------ |
| 14.1 | Research property testing            | 10min    | rapid vs stdlib fuzz     |
| 14.2 | List suffix tree properties          | 10min    | What to verify           |
| 14.3 | Design property: find all duplicates | 10min    | Specification            |
| 14.4 | Design property: no false positives  | 10min    | Specification            |
| 14.5 | Design property: deterministic       | 10min    | Same input = same output |
| 14.6 | Design property: threshold respect   | 10min    | Min length honored       |
| 14.7 | Create test generators               | 15min    | Random token sequences   |
| 14.8 | Document test plan                   | 10min    | Write design doc         |

#### Task 15: Implement Properties (12 subtasks)

| #     | Subtask                   | Duration | Details             |
| ----- | ------------------------- | -------- | ------------------- |
| 15.1  | Create prop_test.go       | 5min     | File setup          |
| 15.2  | Import rapid/testing      | 5min     | Dependencies        |
| 15.3  | Implement token generator | 10min    | Random tokens       |
| 15.4  | Implement tree builder    | 10min    | Build from tokens   |
| 15.5  | Test: find all duplicates | 15min    | Property 1          |
| 15.6  | Test: no false positives  | 15min    | Property 2          |
| 15.7  | Test: deterministic       | 10min    | Property 3          |
| 15.8  | Test: threshold respect   | 10min    | Property 4          |
| 15.9  | Run property tests        | 10min    | Check for failures  |
| 15.10 | Debug any failures        | 15min    | Fix issues          |
| 15.11 | Add edge cases            | 10min    | Boundary conditions |
| 15.12 | Final verification        | 5min     | All pass            |

#### Task 16: Fuzzing (8 subtasks)

| #    | Subtask                            | Duration | Details              |
| ---- | ---------------------------------- | -------- | -------------------- |
| 16.1 | Review existing fuzz tests         | 5min     | What's already there |
| 16.2 | Design fuzz targets                | 10min    | What to fuzz         |
| 16.3 | Create seed corpus                 | 10min    | Initial inputs       |
| 16.4 | Implement tree construction fuzz   | 15min    | Fuzz target          |
| 16.5 | Implement duplicate detection fuzz | 15min    | Fuzz target          |
| 16.6 | Run fuzzing briefly                | 10min    | Check setup          |
| 16.7 | Add to CI                          | 10min    | Workflow update      |
| 16.8 | Document fuzzing                   | 5min     | How to run           |

#### Task 17: CI Benchmark Regression (8 subtasks)

| #    | Subtask                   | Duration | Details             |
| ---- | ------------------------- | -------- | ------------------- |
| 17.1 | Check existing CI config  | 10min    | .github/workflows/  |
| 17.2 | Design benchmark workflow | 10min    | When to run         |
| 17.3 | Create benchmark script   | 15min    | Compare to baseline |
| 17.4 | Set regression threshold  | 10min    | 10% tolerance       |
| 17.5 | Create workflow file      | 15min    | GitHub Actions      |
| 17.6 | Test workflow locally     | 10min    | act or dry-run      |
| 17.7 | Add status badge          | 5min     | README.md           |
| 17.8 | Document process          | 5min     | How it works        |

### Phase 4: User Experience (18 tasks)

#### Task 20: Progress Indicators (6 subtasks)

| #    | Subtask                  | Duration | Details            |
| ---- | ------------------------ | -------- | ------------------ |
| 20.1 | Review current progress  | 10min    | Find progress code |
| 20.2 | Design improvements      | 10min    | What to enhance    |
| 20.3 | Add file count display   | 10min    | X of Y files       |
| 20.4 | Add ETA calculation      | 15min    | Time remaining     |
| 20.5 | Test with large codebase | 10min    | Verify accuracy    |
| 20.6 | Polish output            | 5min     | Formatting         |

#### Task 21: Error Messages (6 subtasks)

| #    | Subtask                   | Duration | Details               |
| ---- | ------------------------- | -------- | --------------------- |
| 21.1 | Review current errors     | 10min    | Find error messages   |
| 21.2 | Identify unclear messages | 10min    | User confusion points |
| 21.3 | Rewrite top 5 messages    | 15min    | Clearer language      |
| 21.4 | Add error context         | 10min    | File:line info        |
| 21.5 | Add suggestions           | 10min    | "Did you mean..."     |
| 21.6 | Test error paths          | 5min     | Verify improvements   |

#### Task 22: Config Validation (6 subtasks)

| #    | Subtask                    | Duration | Details                 |
| ---- | -------------------------- | -------- | ----------------------- |
| 22.1 | List config options        | 10min    | All flags               |
| 22.2 | Define validation rules    | 10min    | Min/max values          |
| 22.3 | Implement validation       | 15min    | Validate function       |
| 22.4 | Add helpful error messages | 10min    | Clear validation errors |
| 22.5 | Test edge cases            | 10min    | Boundary values         |
| 22.6 | Document validation        | 5min     | Update docs             |

### Phase 5: Memory Optimization (30 tasks)

#### Task 24: Memory Profiling (8 subtasks)

| #    | Subtask                 | Duration | Details           |
| ---- | ----------------------- | -------- | ----------------- |
| 24.1 | Create profiling script | 10min    | pprof integration |
| 24.2 | Profile small codebase  | 10min    | Baseline          |
| 24.3 | Profile medium codebase | 10min    | Mid-size          |
| 24.4 | Profile large codebase  | 15min    | Stress test       |
| 24.5 | Analyze heap profile    | 10min    | Find allocations  |
| 24.6 | Document findings       | 10min    | Write report      |
| 24.7 | Identify hotspots       | 10min    | Top allocators    |
| 24.8 | Recommendations         | 5min     | Action items      |

#### Task 25: Hybrid Design (12 subtasks)

| #     | Subtask                  | Duration | Details         |
| ----- | ------------------------ | -------- | --------------- |
| 25.1  | Research hybrid patterns | 10min    | Prior art       |
| 25.2  | Define threshold         | 10min    | 8 transitions?  |
| 25.3  | Design state struct      | 15min    | Dual storage    |
| 25.4  | Design addTran logic     | 15min    | When to switch  |
| 25.5  | Design findTran logic    | 10min    | Which to use    |
| 25.6  | Document design          | 10min    | ADR-style       |
| 25.7  | Create prototype         | 15min    | Implementation  |
| 25.8  | Unit test prototype      | 15min    | Test both paths |
| 25.9  | Benchmark prototype      | 10min    | vs current      |
| 25.10 | Memory benchmark         | 10min    | Check savings   |
| 25.11 | Evaluate trade-offs      | 10min    | Worth it?       |
| 25.12 | Decision document        | 5min     | Go/no-go        |

#### Task 26: Hybrid Implementation (10 subtasks)

| #     | Subtask                | Duration | Details               |
| ----- | ---------------------- | -------- | --------------------- |
| 26.1  | Create feature branch  | 5min     | git checkout -b       |
| 26.2  | Implement dual storage | 15min    | transSmall + transMap |
| 26.3  | Implement addTran      | 15min    | Threshold logic       |
| 26.4  | Implement findTran     | 15min    | Dual lookup           |
| 26.5  | Update iteration       | 10min    | Both storage types    |
| 26.6  | Port all tests         | 15min    | Update test cases     |
| 26.7  | Run test suite         | 10min    | Verify pass           |
| 26.8  | Fix any failures       | 10min    | Debug                 |
| 26.9  | Code review            | 10min    | Self-review           |
| 26.10 | Polish                 | 5min     | Final touches         |

---

## 📊 Summary Tables

### By Phase

| Phase              | Tasks   | Total Time | Cumulative Impact |
| ------------------ | ------- | ---------- | ----------------- |
| 1: Type Safety     | 42      | ~12h       | 51%               |
| 2: Documentation   | 24      | ~6h        | +14% = 65%        |
| 3: Testing         | 36      | ~9h        | +10% = 75%        |
| 4: User Experience | 18      | ~4h        | +3% = 78%         |
| 5: Memory Opt      | 30      | ~10h       | +2% = 80%         |
| **Total**          | **150** | **~41h**   | **80%**           |

### By Effort (Pareto Order)

| Rank | Task             | Effort | Impact | Value/Effort |
| ---- | ---------------- | ------ | ------ | ------------ |
| 1    | TokenValue type  | 6.7h   | 51%    | 7.6%/h       |
| 2    | Property tests   | 3.3h   | 10%    | 3.0%/h       |
| 3    | README update    | 1.0h   | 5%     | 5.0%/h       |
| 4    | CHANGELOG        | 0.7h   | 4%     | 5.7%/h       |
| 5    | CI benchmarks    | 1.5h   | 4%     | 2.7%/h       |
| 6    | PERFORMANCE.md   | 1.7h   | 3%     | 1.8%/h       |
| 7    | AGENTS.md        | 0.7h   | 2%     | 2.9%/h       |
| 8    | Fuzzing          | 1.5h   | 2%     | 1.3%/h       |
| 9    | Error messages   | 1.0h   | 1%     | 1.0%/h       |
| 10   | Memory profiling | 1.0h   | 1%     | 1.0%/h       |

---

## 🎯 Execution Strategy

### Start With (1% → 51%)

1. Design TokenValue type (task 1.1-1.8)
2. Create domain/tokenvalue.go (task 2.1-2.6)

### Then Do (4% → 64%)

3. Update Token interface (task 3.1-3.8)
4. Refactor suffixtree (task 4.1-4.10)
5. Write README performance (task 8.1-8.6)
6. Create CHANGELOG (task 9.1-9.4)

### Then Do (20% → 80%)

7-27. All remaining tasks in priority order

---

## ✅ Success Criteria

- [ ] TokenValue type used throughout codebase
- [ ] All tests pass with new types
- [ ] Documentation reflects all changes
- [ ] Benchmarks show no regression
- [ ] CI detects performance regressions
- [ ] Memory profiling completed
- [ ] Hybrid approach evaluated (decision made)

---

## 🚀 Next Actions

1. **Immediate (Today)**: Begin Task 1.1 - Design TokenValue
2. **This Week**: Complete Phase 1 (Type Safety)
3. **Next Week**: Complete Phase 2 (Documentation)
4. **Following Week**: Complete Phases 3-5

---

**Plan Created:** 2026-03-03 04:50 CET  
**Total Tasks:** 150  
**Estimated Duration:** 41 hours  
**Expected Impact:** 80% of remaining value
