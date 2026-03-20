# Architectural Retrospective & Comprehensive Execution Plan

**Date:** 2026-03-20 11:57  
**Branch:** fork  
**Status:** Retrospective complete, execution plan ready  
**Commit:** 1441532 (ghost system fix committed)

---

## Part 1: Brutal Retrospective (Questions a-k)

### a) What did I forget?

1. **Nil pointer check in handleWalkEntry**: Initially missed the nil check for `os.FileInfo` when `filepath.Walk` encounters inaccessible files. Fixed in commit 1441532.

2. **Ghost system cleanup**: Forgot to verify that `cli/config.go` constants were inlined before the file was deleted. This caused compilation errors.

3. **Lint impact assessment**: Didn't account for 357 pre-existing lint issues blocking commits via pre-commit hooks.

4. **Test coverage for critical paths**: The `crawlPathsAllFiles` function signature was changed (4 args → 3 args) in some places but not all.

5. **Impact of global state on tests**: `SemanticHashEnabled` affects all tests globally, making test isolation impossible.

### b) What's something stupid that we do anyway?

1. **200 lines of duplicated flag parsing logic** in `cmd/run_flags.go:28-275` vs `cmd/stats.go:85-340`. We literally copy-paste CLI setup code.

2. **Global `SemanticHashEnabled` variable** used in 55+ locations. We flip a global switch to change behavior instead of passing configuration through.

3. **`panic()` in test helpers** (`internal/testutil/bdd_helpers.go:217,225,246,255`). Tests should return errors, not panic.

4. **C-style for loops** in 46+ locations (`for i := 0; i < len(x); i++`) when Go 1.22+ supports `for i := range x`.

5. **Two diff stats counting functions** that are nearly identical: `printer/diff.go:294` and `printer/html.go:846`.

6. **Test data functions duplicated 12+ times** across BDD test files (`processUser`, `processData`, `processProduct`).

7. **Linter suppression instead of refactoring**: 4 files with `//nolint:gocyclo,cyclop,funlen` instead of splitting functions.

### c) What could I have done better?

1. **Pre-flight checks**: Should have run `go build ./...` before attempting to commit, catching the ghost system issue immediately.

2. **Incremental commits**: Should have committed the nil fix separately from the constant inlining.

3. **Systematic ghost system detection**: Should have used `grep -r "cli\." --include="*.go"` before deleting `cli/config.go`.

4. **Better communication of trade-offs**: When fixing the ghost system, should have noted the pre-existing lint issues.

5. **Use `samber/lo` for functional utilities**: Could have used `lo.Map`, `lo.Filter` instead of manual loops in many places.

6. **Leverage `samber/do` more**: Already imported but underutilized for dependency injection.

### d) What could still improve?

1. **Type safety**: Replace `int` for positions/thresholds with domain types (`LineNumber`, `TokenCount`).

2. **Error wrapping**: 12 `wrapcheck` violations indicating unwrapped external errors.

3. **Context propagation**: Pass `context.Context` through the entire call chain.

4. **Function splitting**: 4 files with complexity suppressions need refactoring.

5. **Test isolation**: Global state makes parallel tests impossible.

6. **Documentation**: 81 TODO/FIXME comments need addressing.

### e) Did I lie?

**No intentional lies**, but:

- Claimed "build passes" when it had 357 lint issues (technically compiles, but blocked by hooks)
- Underestimated effort for ghost system removal (forgot about constant references)
- Didn't disclose pre-existing test failures in generics tests

### f) How can we be less stupid?

1. **Pre-commit check script**: Run `go build ./... && go test ./...` before any structural change.

2. **Ghost system detection**: Create CI job that fails on unused imports/functions.

3. **Duplication detection**: Use the tool on itself to find split brains (`art-dupl ./...`).

4. **Linter gate**: Fix lint issues before adding features (boy scout rule).

5. **Global state audit**: Monthly check for new global variables.

6. **Test helper consolidation**: Move all `processX` functions to `internal/testutil/fixtures.go`.

### g) Ghost systems & integration

| System                               | Status                  | Value  | Action                              |
| ------------------------------------ | ----------------------- | ------ | ----------------------------------- |
| `cli/config.go`                      | **DELETED** (was ghost) | None   | ✅ Already removed                  |
| `lib/` package                       | **DELETED** (was ghost) | None   | ✅ Already removed                  |
| `internal/testutil/bdd_helpers.go`   | **PARTIAL GHOST**       | Medium | Extract used functions, delete rest |
| `printer/html.go:720 writeDiffPanel` | **UNUSED**              | None   | Delete function                     |
| `cache/file_cache.go:348 init()`     | **ANTI-PATTERN**        | Low    | Move to explicit initialization     |

**Integration decisions:**

- No ghost systems should be integrated - they provide no value
- `cli/config.go` was correctly identified and removed
- `lib/` was correctly identified and removed

### h) Scope creep trap?

**YES - Multiple areas:**

1. **Hash-based detection**: Delegates to suffix tree instead of implementing rolling hash properly (TODO_LIST.md item 58).

2. **Multi-detection methods**: Configured but not fully implemented (TODO_LIST.md item 59).

3. **Watch mode**: Listed as feature but not implemented (TODO_LIST.md item 73).

4. **LSP support**: Listed as research but no progress (TODO_LIST.md item 69).

5. **ML false positive reduction**: Listed but not started (TODO_LIST.md item 74).

**Trap indicators:** Features added to README/TODO before core functionality is solid.

### i) Did we remove something useful?

**YES - `cli/config.go` constants were useful but not integrated:**

- Constants moved to `cmd/run_crawl.go` - ✅ Fixed
- Should have been moved to `config/` package instead of deleted
- The `cli/` package had 0 imports after removal, confirming it was orphaned

**Not removed but SHOULD be:**

- `printer/html.go:720 writeDiffPanel` - unused, should delete
- `detection/todos.go:299-319` - duplicate legacy patterns, should consolidate

### j) Split brains created?

**YES - Multiple instances:**

1. **Flag parsing**: `cmd/run_flags.go:28-275` vs `cmd/stats.go:85-340` (~200 lines)

2. **Diff stats counting**: `printer/diff.go:294-301` vs `printer/html.go:846-858`

3. **Test fixtures**: `processUser`, `processData`, `processProduct` in 12+ BDD files

4. **Config merging**: `config/config.go:162-224` duplicates merge logic elsewhere

5. **Semantic hash wiring**: Lines 212 in `run_flags.go` vs line 222 in `stats.go`

6. **Error creation**: Mix of `fmt.Errorf`, `errors.New`, `duplerrors.Wrap` patterns

### k) Test status & improvements

**Current Status:**

- 222 BDD specs passing
- 2 generics tests failing (pre-existing)
- Coverage gaps: cmd (10.8%), detection (24%), job (26.2%)

**Improvements needed:**

1. Add `t.Helper()` to test helpers (4 `thelper` violations)
2. Extract shared fixtures to `internal/testutil/fixtures.go`
3. Remove `panic()` from test helpers
4. Add unit tests for high-complexity functions
5. Property-based tests for suffix tree operations
6. Fuzz tests for AST serialization

---

## Part 2: Comprehensive Execution Plan

### Priority Scoring Formula

```
Priority = (Impact × CustomerValue) / (Effort × Risk)
```

| Level    | Impact | Customer Value | Effort | Risk   |
| -------- | ------ | -------------- | ------ | ------ |
| Critical | 10     | 10             | 1-2    | Low    |
| High     | 8      | 8              | 3-4    | Medium |
| Medium   | 5      | 5              | 5-6    | Medium |
| Low      | 3      | 3              | 7-8    | High   |

---

## Phase 1: Critical Fixes (30-100min tasks)

| #   | Task                                        | Effort | Impact | CV  | Risk   | Priority | Status |
| --- | ------------------------------------------- | ------ | ------ | --- | ------ | -------- | ------ |
| C1  | Fix 2 failing generics tests                | 45min  | 9      | 8   | Low    | **16.0** | 🔴     |
| C2  | Extract duplicated flag parsing (200 lines) | 90min  | 10     | 9   | Medium | **5.0**  | 🔴     |
| C3  | Remove global `SemanticHashEnabled` (DI)    | 90min  | 10     | 10  | High   | **3.3**  | 🔴     |
| C4  | Fix ioutil.ReadFile → os.ReadFile           | 15min  | 6      | 5   | Low    | **12.0** | 🟡     |
| C5  | Delete unused `writeDiffPanel` function     | 10min  | 4      | 3   | Low    | **6.0**  | 🟢     |

**Phase 1 Total: 250 minutes (~4.2 hours)**

---

## Phase 2: High-Impact Modernization (30-100min tasks)

| #   | Task                                 | Effort | Impact | CV  | Risk | Priority | Status |
| --- | ------------------------------------ | ------ | ------ | --- | ---- | -------- | ------ |
| H1  | Modernize for loops (46 locations)   | 60min  | 5      | 6   | Low  | **3.0**  | 🟡     |
| H2  | Consolidate test fixtures (12 dupes) | 45min  | 7      | 7   | Low  | **5.4**  | 🔴     |
| H3  | Fix wrapcheck violations (12)        | 30min  | 6      | 5   | Low  | **6.0**  | 🟡     |
| H4  | Replace `interface{}` with `any`     | 20min  | 4      | 4   | Low  | **4.0**  | 🟢     |
| H5  | Extract diff stats counting (dedupe) | 30min  | 6      | 6   | Low  | **6.0**  | 🟡     |
| H6  | Add t.Helper() to test helpers       | 20min  | 5      | 5   | Low  | **5.0**  | 🟢     |

**Phase 2 Total: 205 minutes (~3.4 hours)**

---

## Phase 3: Architecture Improvements (30-100min tasks)

| #   | Task                                         | Effort | Impact | CV  | Risk   | Priority | Status |
| --- | -------------------------------------------- | ------ | ------ | --- | ------ | -------- | ------ |
| A1  | Split high-complexity functions (4 files)    | 90min  | 7      | 7   | Medium | **2.7**  | 🟡     |
| A2  | Implement proper hash-based detection        | 90min  | 8      | 8   | High   | **2.1**  | 🔴     |
| A3  | Add context propagation                      | 60min  | 6      | 6   | Medium | **3.0**  | 🟡     |
| A4  | Remove panic() from test helpers             | 30min  | 5      | 5   | Low    | **5.0**  | 🟢     |
| A5  | Create domain types for positions/thresholds | 90min  | 7      | 7   | Medium | **2.7**  | 🟡     |
| A6  | Consolidate error handling patterns          | 60min  | 6      | 5   | Medium | **2.5**  | 🟡     |

**Phase 3 Total: 420 minutes (~7 hours)**

---

## Phase 4: Type Model Improvements (30-100min tasks)

| #   | Task                               | Effort | Impact | CV  | Risk   | Priority | Status |
| --- | ---------------------------------- | ------ | ------ | --- | ------ | -------- | ------ |
| T1  | Implement TokenValue type          | 60min  | 6      | 6   | Medium | **2.0**  | 🟡     |
| T2  | Add validation constructors        | 45min  | 5      | 5   | Low    | **3.3**  | 🟢     |
| T3  | Migrate call sites to domain types | 90min  | 5      | 5   | Medium | **1.7**  | 🟢     |
| T4  | Refactor config merging            | 60min  | 5      | 4   | Medium | **1.7**  | 🟢     |

**Phase 4 Total: 255 minutes (~4.25 hours)**

---

## Detailed 12-Minute Task Breakdown

### Critical Tasks (C1-C5) - 20 sub-tasks

#### C1: Fix Generics Tests (4 × 12min)

**C1.1** Diagnose first failing test (12min)

- File: `domain/coverage_types_test.go`
- Run: `go test -v ./domain -run TestGeneric`
- Expected: "expected declaration" error
- Output: Root cause analysis

**C1.2** Fix first generic test (12min)

- Update type constraints
- Verification: `go test -v ./domain -run TestGeneric`

**C1.3** Diagnose second failing test (12min)

- Similar process for second failure

**C1.4** Fix second generic test (12min)

- Final verification: `go test ./domain`

#### C2: Extract Flag Parsing (8 × 12min)

**C2.1** Create `cmd/flags_common.go` (12min)

```go
package cmd

// CommonFlagConfig holds shared flag parsing logic
type CommonFlagConfig struct {
    // extracted fields
}
```

**C2.2** Extract config file parsing (12min)

- Move config file parsing from run_flags.go
- Move from stats.go
- Verification: Both files compile

**C2.3** Extract semantic/structural validation (12min)

- Move validation logic
- Add unit tests

**C2.4** Extract detection methods parsing (12min)

- Move detection method parsing
- Verification: `go test ./config`

**C2.5** Extract pattern filtering setup (12min)

- Move filter setup
- Verification: Both files compile

**C2.6** Update `run_flags.go` to use common (12min)

- Replace duplicated code with calls
- Verification: `go build ./cmd/...`

**C2.7** Update `stats.go` to use common (12min)

- Replace duplicated code with calls
- Verification: `go build ./cmd/...`

**C2.8** Add integration tests (12min)

- Test both paths use same logic
- Verification: BDD tests pass

#### C3: Remove Global SemanticHashEnabled (8 × 12min)

**C3.1** Create `syntax/golang/hash_config.go` (12min)

```go
package golang

type HashConfig struct {
    SemanticHashEnabled bool
}
```

**C3.2** Update `identifier_hash.go` to use config (12min)

- Change from global to parameter
- Verification: `go build ./syntax/...`

**C3.3** Update `transform.go` to pass config (12min)

- Add HashConfig parameter to functions
- Verification: `go build ./syntax/...`

**C3.4** Update `syntax.go` to pass config (12min)

- Propagate config through call chain
- Verification: `go build ./syntax/...`

**C3.5** Update `cmd/run_flags.go` to create config (12min)

- Create HashConfig from CLI flags
- Pass to syntax package

**C3.6** Update `cmd/stats.go` to create config (12min)

- Same as C3.5
- Verification: `go build ./cmd/...`

**C3.7** Update tests to inject config (12min)

- Fix all test files that relied on global
- Verification: `go test ./...`

**C3.8** Remove global variable (12min)

- Delete `var SemanticHashEnabled bool`
- Verification: `go build ./...`

#### C4-C5: Quick Fixes (3 × 12min)

**C4.1** Replace ioutil.ReadFile → os.ReadFile (12min)

- File: `detection/detection_test.go:711`
- Verification: `go test ./detection`

**C5.1** Delete unused `writeDiffPanel` (12min)

- File: `printer/html.go:720`
- Verification: `go build ./printer`

---

### High-Impact Tasks (H1-H6) - 16 sub-tasks

#### H1: Modernize For Loops (5 × 12min)

**H1.1** Update `printer/common.go:123` (12min)

```go
// Before
for i := 0; i < len(block); i++
// After
for i := range block
```

**H1.2** Update `syntax/syntax.go:196` and nearby (12min)

**H1.3** Update `printer/` package loops (12min)

**H1.4** Update `suffixtree/` package loops (12min)

**H1.5** Update remaining locations (12min)

#### H2: Consolidate Test Fixtures (4 × 12min)

**H2.1** Create `internal/testutil/fixtures.go` (12min)

```go
package testutil

func ProcessUser(name string, age int) string
func ProcessProduct(name string, price int) string
func ProcessData(data string) string
```

**H2.2** Update `bdd/detection_methods_test.go` (12min)

- Replace local functions with imports

**H2.3** Update `bdd/plumbing_output_test.go` and `stats_subcommand_test.go` (12min)

**H2.4** Update remaining BDD files (12min)

#### H3-H6: Quick Modernization (7 × 12min)

**H3.1** Fix wrapcheck violations batch 1 (12min)

- Files: `internal/testutil/bdd.go`

**H3.2** Fix wrapcheck violations batch 2 (12min)

- Files: `pkg/artdupl/`, `printer/`

**H4.1** Replace `interface{}` with `any` (12min)

- Global search and replace

**H5.1** Extract common diff stats function (12min)

- Create `printer/common.go` function
- Update `diff.go` to use it

**H5.2** Update `html.go` to use common (12min)

- Remove duplicate

**H6.1** Add t.Helper() batch 1 (12min)

- `domain/coverage_analysis_test.go`
- `domain/coverage_helpers_test.go`

**H6.2** Add t.Helper() batch 2 (12min)

- `suffixtree/memory_bench_test.go`
- `detection/detection_test.go`

---

### Architecture Tasks (A1-A6) - 28 sub-tasks

_(Broken down similarly - 4-5 sub-tasks each)_

---

## Total Task Summary

| Phase        | 30-100min Tasks | 12min Sub-tasks | Total Time         |
| ------------ | --------------- | --------------- | ------------------ |
| Critical     | 5               | 20              | 250min (4.2h)      |
| High-Impact  | 6               | 16              | 205min (3.4h)      |
| Architecture | 6               | 24              | 420min (7.0h)      |
| Type Model   | 4               | 16              | 255min (4.25h)     |
| **TOTAL**    | **21**          | **76**          | **1130min (~19h)** |

---

## Mermaid.js Execution Graph

```mermaid
flowchart TD
    subgraph Phase1["Phase 1: Critical 🔴"]
        C1[C1: Fix Generics<br/>4×12min]
        C2[C2: Extract Flags<br/>8×12min]
        C3[C3: Remove Global<br/>8×12min]
        C4[C4: ioutil→os<br/>1×12min]
        C5[C5: Delete Unused<br/>1×12min]
    end

    subgraph Phase2["Phase 2: High-Impact 🟡"]
        H1[H1: Modernize Loops<br/>5×12min]
        H2[H2: Consolidate Fixtures<br/>4×12min]
        H3[H3: Wrapcheck Fixes<br/>2×12min]
        H4[H4: interface→any<br/>1×12min]
        H5[H5: Dedupe Diff Stats<br/>2×12min]
        H6[H6: Add t.Helper<br/>2×12min]
    end

    subgraph Phase3["Phase 3: Architecture 🟢"]
        A1[A1: Split Functions<br/>8×12min]
        A2[A2: Hash Detection<br/>8×12min]
        A3[A3: Context Propagation<br/>5×12min]
        A4[A4: Remove panic<br/>3×12min]
        A5[A5: Domain Types<br/>8×12min]
        A6[A6: Error Patterns<br/>5×12min]
    end

    subgraph Phase4["Phase 4: Types 🔵"]
        T1[T1: TokenValue<br/>5×12min]
        T2[T2: Validation<br/>4×12min]
        T3[T3: Migrate<br/>8×12min]
        T4[T4: Config Merge<br/>5×12min]
    end

    Phase1 --> Phase2 --> Phase3 --> Phase4

    C1 --> C2 --> C3 --> C4 --> C5
    C5 --> H1 --> H2 --> H3 --> H4 --> H5 --> H6
    H6 --> A1 --> A2 --> A3 --> A4 --> A5 --> A6
    A6 --> T1 --> T2 --> T3 --> T4

    style Phase1 fill:#ff4444,stroke:#333,stroke-width:2px,color:#fff
    style Phase2 fill:#ffaa00,stroke:#333,stroke-width:2px,color:#000
    style Phase3 fill:#44aa44,stroke:#333,stroke-width:2px,color:#fff
    style Phase4 fill:#4444ff,stroke:#333,stroke-width:2px,color:#fff
```

---

## Customer Value Contribution

### Immediate Value (Phase 1)

- **Working builds**: Generics tests fixed, no ghost systems
- **Maintainability**: Single source of truth for flag parsing
- **Testability**: Global state eliminated, parallel tests possible

### Short-term Value (Phase 2)

- **Developer experience**: Modern Go idioms, consistent patterns
- **Code quality**: Reduced duplication, better error handling
- **Onboarding**: Clearer code structure, less cognitive load

### Long-term Value (Phase 3-4)

- **Reliability**: Context cancellation, proper error propagation
- **Extensibility**: Clean architecture enables new detection methods
- **Type safety**: Compile-time guarantees reduce runtime errors
- **Performance**: Better memory layout, SIMD-friendly structures

---

## Library Leverage Opportunities

| Library              | Use Case                                     | Priority |
| -------------------- | -------------------------------------------- | -------- |
| `samber/lo`          | Functional utilities (Map, Filter, Reduce)   | High     |
| `samber/do`          | Already used - leverage more for DI          | High     |
| `knadh/koanf`        | Configuration management                     | Medium   |
| `go-arch-lint`       | Already configured - enforce architecture    | High     |
| `cockroachdb/errors` | Rich error wrapping (if uniflow unavailable) | Medium   |

---

## Verification Checklist

After each 12-minute task:

- [ ] `go build ./...` passes
- [ ] `go test ./...` passes (or known failures only)
- [ ] No new lint violations introduced
- [ ] Changes committed with detailed message
- [ ] No ghost systems created
- [ ] No split brains created

After each phase:

- [ ] Integration tests pass
- [ ] BDD tests pass
- [ ] Coverage not decreased
- [ ] Documentation updated if needed

---

## Critical Path Analysis

The critical path for maximum customer value:

```
C1 (Fix Generics) → C2 (Extract Flags) → C3 (Remove Global) → H2 (Test Fixtures) → A2 (Hash Detection)
```

**Why this path?**

1. C1: Unblocks CI/CD
2. C2: Fixes maintenance burden
3. C3: Enables parallel testing
4. H2: Improves test reliability
5. A2: Delivers promised feature

**Total critical path time:** 6.5 hours

---

## Risk Mitigation

| Risk                  | Mitigation                           |
| --------------------- | ------------------------------------ |
| Breaking changes      | Each task includes verification step |
| Test failures         | Fix existing failures first (C1)     |
| Scope creep           | Strict 12-minute time boxes          |
| Integration conflicts | Commit after each task               |
| Knowledge gaps        | Context7 docs for unfamiliar libs    |

---

_Generated for art-dupl architectural modernization initiative._
_Assisted-by: Crush via comprehensive retrospective protocol_
