# Code Deduplication Execution Plan

## 📊 Current State Analysis

### What I Discovered (Self-Reflection)

**What I Did Wrong Initially:**
1. ❌ Jumped directly to extracting code without understanding the architecture
2. ❌ Didn't trace code execution from entry point (main.go) to understand what's active
3. ❌ Assumed both cli.go and cmd/run.go were active implementations
4. ❌ Missed the presence of cli.go.old which indicated deprecation in progress
5. ❌ Didn't check if code was actually being used before attempting to refactor it

**What I Should Have Done:**
1. ✅ Trace execution from main.go to understand the active code paths
2. ✅ Check for deprecation patterns (old files, migration docs)
3. ✅ Verify if functions are actually called before refactoring
4. ✅ Understand the architectural context before making changes
5. ✅ Use tools like agent for comprehensive searches instead of manual grepping

### Key Findings

#### Architecture Discovery

**Active Entry Point:**
- `main.go` → uses `cmd` package (Cobra-based)
- `cmd/run.go` contains the active `runCmd()` function

**Dead Code:**
- `cli.go` (~605 lines) - completely unused
  - `Run()` function never called
  - `RunCobraCommand()` never called
  - Contains duplicate logic from cmd/run.go

**Active Shared Package:**
- `cli/` directory is ACTIVE and used
  - `cli.NewCLIConfig()`, `cli.ExitIfBothSet()`, `cli.DefaultThreshold`, etc.

### Duplicate Detection Results (threshold 70)

**Real Code Duplicates:**
1. `cli.go:321,348` ↔ `cmd/run.go:354,381` - `sortCloneGroupKeys` function
2. `cli.go:564,592` ↔ `cmd/run.go:457,485` - Generate all output formats loop
3. `cli.go:333,343` ↔ `cmd/run.go:366,376` → also in `cli/cli_sorting_test.go:145,155` - Size calculation
4. `detection/todos.go:66,93` ↔ `detection/todos.go:174,200` - FindTodos vs FindLegacy pattern

**Test File Duplicates (lower priority):**
- `domain/domain_types_test.go` - Multiple test setup/teardown duplicates
- `pkg/filter/filter_test.go:57,72` ↔ `pkg/filter/filter_test.go:74,89`

## 🎯 Execution Plan (Sorted by Impact vs Work)

### Phase 1: Quick Wins (High Impact, Low Work)

#### 1.1 Remove cli.go Dead Code ⚡ IMMEDIATE IMPACT
**Impact:** Eliminates 605 lines of dead code + removes 2 duplicate groups
**Work:** Low (remove file, test build)
**Risk:** Very Low (verified unused)

**Steps:**
1. ✅ Verify no imports depend on cli.go functions (DONE - confirmed unused)
2. Remove cli.go
3. Test: `go build`
4. Test: `go test ./...`
5. Commit: "refactor: Remove dead code from cli.go (605 lines)"

#### 1.2 Extract Common Pattern from detection/todos.go
**Impact:** Removes 1 duplicate group, improves maintainability
**Work:** Medium
**Risk:** Low (isolated change)

**Analysis:**
Both `FindTodos` and `FindLegacy` share this pattern:
```go
// Group nodes by filename
nodesByFile := make(map[string][]*syntax.Node)
for _, node := range data {
    nodesByFile[node.Filename] = append(nodesByFile[node.Filename], node)
}

// Process each file
for filename, nodes := range nodesByFile {
    items := detector.findItemsInFile(filename, nodes)
    for _, item := range items {
        match := syntax.Match{
            Hash:  fmt.Sprintf("TYPE-%s-%d", filename, item.Line),
            Frags: [][]*syntax.Node{{}},
        }
        resultChan <- match
    }
}
```

**Approach:**
Extract a generic `processByFile()` function that takes:
- A function that finds items in a file
- A function that creates matches from items
- A prefix for the hash (e.g., "TODO", "LEGACY")

**Steps:**
1. Create `processByFile()` helper in detection/todos.go
2. Refactor `FindTodos` to use it
3. Refactor `FindLegacy` to use it
4. Run tests
5. Commit: "refactor(detection): Extract common file processing pattern"

### Phase 2: Architecture Improvements (High Impact, Medium Work)

#### 2.1 Move Helper Functions from cmd/run.go to printer Package
**Impact:** Improves code organization, reduces duplication, makes functions reusable
**Work:** Medium
**Risk:** Low (well-tested functions)

**Functions to Move:**
- `buildCloneGroups()` - Builds clone groups from matches
- `computeUniqueCounts()` - Calculates unique file counts
- `sortCloneGroupKeys()` - Sorts clone group keys

**Current Location:**
- `cmd/run.go` (used only there)
- `cli.go` (dead code, had duplicates)

**New Location:**
- `printer/` package (logical home for clone grouping/sorting)

**Steps:**
1. Create `printer/groups.go`
2. Move functions from cmd/run.go
3. Update imports in cmd/run.go
4. Run tests
5. Commit: "refactor(printer): Extract clone grouping functions to printer package"

#### 2.2 Extract Output Format Generation Loop
**Impact:** Removes duplication, makes logic clearer
**Work:** Medium
**Risk:** Low

**Current Situation:**
The loop for generating all output formats exists in:
- `cmd/run.go:457,485` (active)
- `cli.go:564,592` (dead code, will be removed)

After removing cli.go, this won't be duplicated anymore. However, we can still improve it:

**Improvement:**
Extract the format generation loop into a function:
```go
func generateAllFormats(cfg *config.Config, matches []syntax.Match, ...) error
```

**Steps:**
1. Extract the format generation loop in cmd/run.go
2. Test with `--all` flag
3. Commit: "refactor(cmd): Extract output format generation logic"

### Phase 3: Test Improvements (Medium Impact, Low Work)

#### 3.1 Fix Test File Duplicates
**Impact:** Better test maintainability
**Work:** Low to Medium
**Risk:** Low

**Files:**
- `domain/domain_types_test.go` - Extract test helpers
- `pkg/filter/filter_test.go` - Extract common test setup

**Approach:**
Create helper functions for common test setup/teardown patterns.

**Steps:**
1. Identify duplicate test patterns
2. Extract test helpers
3. Run tests
4. Commit: "refactor(test): Extract common test helpers"

### Phase 4: Type Model Improvements (Long-term)

#### 4.1 Strengthen Type Models for Better Architecture
**Impact:** Prevents duplication through better design
**Work:** Higher (requires architectural changes)
**Risk:** Medium

**Analysis Opportunities:**

1. **Detection Methods** - Already enum-based, good
2. **Output Format** - Already enum-based, good
3. **Match/Clone** - Could be strengthened with:
   - Builder pattern for creating matches
   - Validation at construction time
   - Stronger types for hash values

**Current State:**
```go
type Match struct {
    Frags [][]*syntax.Node
    Hash  string
}
```

**Potential Improvement:**
```go
type Hash string // Stronger type with validation

func NewHash(prefix, filename string, line int) Hash {
    return Hash(fmt.Sprintf("%s-%s-%d", prefix, filename, line))
}

type Match struct {
    Frags [][]*syntax.Node
    Hash  Hash
}
```

### Phase 5: Library Research (Medium Impact, Low Work)

#### 5.1 Research Well-Established Libraries
**Impact:** Potential to simplify code, reduce boilerplate
**Work:** Low
**Risk:** Low

**Areas to Research:**
1. **Testing** - Already using native Go testing ✅
2. **CLI** - Using Cobra + fang ✅ (good choice)
3. **Sorting** - Using standard sort package ✅
4. **Generics** - Could use generics for common patterns (Go 1.18+)

**Potential Libraries to Consider:**
- `samber/lo` - Functional programming utilities (filter, map, reduce, etc.)
  - Could simplify the Unique function and other transformations
  - Well-tested, popular in Go community
  - Reduces boilerplate

**Decision:**
Given the project already has minimal dependencies and the code is simple, adding a library might not be worth it unless there's significant boilerplate reduction.

## 📋 Execution Checklist

- [ ] 1.1 Remove cli.go
- [ ] 1.2 Extract common pattern from detection/todos.go
- [ ] 2.1 Move helper functions to printer package
- [ ] 2.2 Extract output format generation (if still beneficial after 1.1)
- [ ] 3.1 Fix test file duplicates
- [ ] 4.1 Consider type model improvements (defer to later)
- [ ] 5.1 Research libraries (already done, decision: not needed)

## 🎯 Success Metrics

**Before:**
- Total duplicate groups: 16
- Dead code lines: ~605 (cli.go)
- Active duplicate functions: 4+ instances

**After (Target):**
- Total duplicate groups: < 10 (test duplicates may remain)
- Dead code lines: 0
- Active duplicate functions: 0-1 instances

## 🚨 Risks & Mitigations

1. **Removing cli.go breaks something**
   - Mitigation: Comprehensive testing after removal
   - Rollback: Can restore from git if needed

2. **Refactoring breaks existing functionality**
   - Mitigation: Run full test suite after each change
   - Rollback: Commit frequently with small changes

3. **Type model changes break compatibility**
   - Mitigation: Don't do type changes in this session (defer to later)
   - Focus only on deduplication

## 📊 Estimated Impact

**Code Reduction:**
- cli.go: -605 lines
- Duplicated functions: ~100 lines
- **Total: ~700 lines removed**

**Maintainability:**
- Reduced code duplication
- Better organized packages
- Clearer separation of concerns

**Test Coverage:**
- All tests should pass
- No functional changes

