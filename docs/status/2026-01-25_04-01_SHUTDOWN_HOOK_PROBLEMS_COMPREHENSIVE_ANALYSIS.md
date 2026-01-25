# Shutdown Hook Implementation Problems - Comprehensive Analysis

**Date**: 2026-01-25
**Time**: 04:01
**Status**: 🚨 CRITICAL - Multiple fundamental issues identified
**Severity**: HIGH - Signal handling and timeout features completely non-functional

---

## Executive Summary

The shutdown hook implementation has **multiple critical problems** that render both signal handling (Ctrl+C) and timeout features completely non-functional. While Fang's signal handling is properly configured in `main.go`, the context is never propagated through the analysis pipeline, making it impossible to cancel long-running operations.

---

## Problem Analysis

### 🚨 Problem 1: Fang Context Not Passed to Analysis Pipeline

**Location**: `cmd/art-dupl/main.go:66-69`

```go
options := []fang.Option{
    fang.WithVersion(cmd.GetVersion()),
    fang.WithColorSchemeFunc(fang.DefaultColorScheme),
    fang.WithErrorHandler(errorHandler),
    fang.WithNotifySignal(os.Interrupt), // Handle Ctrl+C gracefully
}

if err := fang.Execute(context.Background(), rootCmd, options...); err != nil {
    os.Exit(1)
}
```

**Issue**:

- Fang creates a `signal.NotifyContext` when `WithNotifySignal` is used
- This context is passed to Cobra via `root.ExecuteContext(ctx)`
- However, `run.go` functions never use this context
- The context is available via `cmd.Context()` but never accessed

**Impact**: Ctrl+C signal handling does NOT work at all for analysis operations

---

### 🚨 Problem 2: Timeout Context Created But Never Used

**Location**: `cmd/run.go:142-152`

```go
// Add timeout context if specified
ctx := context.Background()
if mergedConfig.Timeout > 0 {
    var cancel context.CancelFunc
    ctx, cancel = context.WithTimeout(ctx, time.Duration(mergedConfig.Timeout)*time.Second)
    defer cancel()
    fmt.Fprintf(os.Stderr, "⏱️  Execution timeout: %ds\n", mergedConfig.Timeout)
    // Note: Timeout context is configured but not yet passed through the analysis pipeline
    // This would require refactoring executeAnalysis to accept and propagate context
}
// For now, timeout config is stored but not fully implemented in analysis
duplChan, filesCount, err := executeAnalysis(mergedConfig, mergedConfig.Paths)
```

**Issue**:

- Timeout context is created locally with `WithTimeout`
- The context variable `ctx` is NEVER passed to `executeAnalysis`
- Comment explicitly admits this is not yet implemented
- Timeout configuration is stored in config but completely non-functional

**Impact**: Timeout feature is completely broken; operations will never timeout

---

### 🚨 Problem 3: No Context Propagation Through Analysis Pipeline

**Locations**: Multiple functions in the pipeline don't accept context parameters

#### Function Signature Issues

1. **`cmd/run.go:290`** - `executeAnalysis`:

   ```go
   func executeAnalysis(cfg *config.Config, paths []string) (chan syntax.Match, int, error)
   ```

   - Missing `ctx context.Context` parameter
   - Cannot be cancelled from outside

2. **`cmd/run.go:201`** - `buildSuffixTree`:

   ```go
   func buildSuffixTree(paths []string, verbose, filesFromStdin bool, filterParam *filter.Filter, includeVendor bool) (*suffixtree.STree, []*syntax.Node, int, error)
   ```

   - Missing `ctx context.Context` parameter
   - Cannot be cancelled during tree building

3. **`job/parse.go:9`** - `Parse`:

   ```go
   func Parse(fchan chan string) (chan []*syntax.Node, chan int)
   ```

   - Missing `ctx context.Context` parameter
   - File parsing cannot be cancelled

4. **`job/buildtree.go:8`** - `BuildTree`:
   ```go
   func BuildTree(schan chan []*syntax.Node) (t *suffixtree.STree, d *[]*syntax.Node, done chan bool)
   ```

   - Missing `ctx context.Context` parameter
   - Tree building cannot be cancelled

**Impact**: Long-running operations in the pipeline cannot be cancelled

---

### 🚨 Problem 4: Missing Context Cancellation Checks in Job Functions

**Location**: `job/parse.go` and `job/buildtree.go`

#### Parse Function (`job/parse.go:9-37`)

```go
func Parse(fchan chan string) (chan []*syntax.Node, chan int) {
    // parse AST
    achan := make(chan *syntax.Node)
    countChan := make(chan int, 1)
    go func() {
        fileCount := 0
        for file := range fchan {
            // NO context cancellation check here!
            fileCount++
            ast, err := golang.Parse(file)
            if err != nil {
                logger.Default.Error("failed to parse file", "file", file, "err", err)
                continue
            }
            achan <- ast
        }
        countChan <- fileCount
        close(achan)
    }()

    // serialize
    schan := make(chan []*syntax.Node)
    go func() {
        for ast := range achan {
            // NO context cancellation check here!
            seq := syntax.Serialize(ast)
            schan <- seq
        }
        close(schan)
    }()
    return schan, countChan
}
```

**Issues**:

- No `select` with `ctx.Done()` check in file parsing loop
- No `select` with `ctx.Done()` check in serialization loop
- Will continue processing all files even after cancellation

#### BuildTree Function (`job/buildtree.go:8-21`)

```go
func BuildTree(schan chan []*syntax.Node) (t *suffixtree.STree, d *[]*syntax.Node, done chan bool) {
    t = suffixtree.New()
    data := make([]*syntax.Node, 0, 100)
    done = make(chan bool)
    go func() {
        for seq := range schan {
            // NO context cancellation check here!
            data = append(data, seq...)
            for _, node := range seq {
                t.Update(node)
            }
        }
        done <- true
    }()
    return t, &data, done
}
```

**Issues**:

- No `select` with `ctx.Done()` check in tree building loop
- No `select` with `ctx.Done()` check in node update loop
- Will continue building tree even after cancellation

**Impact**: Operations will continue running until complete, even after Ctrl+C or timeout

---

### 🚨 Problem 5: Incomplete Context Support in Detection Layer

**Location**: Comparison between `pkg/artdupl/detector.go` and `cmd/run.go`

#### What Works Properly: `pkg/artdupl/detector.go`

The SDK detector properly uses context throughout:

```go
func (d *detector) buildAnalysisPipeline(ctx context.Context, files []string) ([]*syntax.Node, int, error) {
    // ... file processing ...
    for i, filename := range files {
        // Check for context cancellation
        select {
        case <-ctx.Done():
            return  // Properly handles cancellation
        default:
        }
        // ... process file ...
    }

    // Wait for tree building to complete
    select {
    case <-done:
        // Tree building complete
    case <-ctx.Done():
        return nil, 0, ctx.Err()  // Properly handles cancellation
    case <-time.After(d.opts.Timeout):
        return nil, 0, ErrAnalysisTimeout  // Properly handles timeout
    }
}

func (d *detector) runDetection(ctx context.Context, data []*syntax.Node) ([]*CloneGroup, error) {
    for match := range matchesChan {
        // Check for cancellation
        select {
        case <-ctx.Done():
            return nil, ctx.Err()  // Properly handles cancellation
        default:
        }
        // ... process match ...
    }
}
```

#### What's Broken: `cmd/run.go` Path

The CLI completely bypasses the SDK detector:

```go
func executeAnalysis(cfg *config.Config, paths []string) (chan syntax.Match, int, error) {
    // ... build suffix tree ...
    t, data, filesCount, err := buildSuffixTree(paths, cfg.Verbose, cfg.FilesFromStdin, filterParam, cfg.IncludeVendor)

    // ... find duplicates using multiDetector ...
    multiDetector := detection.NewMultiDetector(cfg, data, t, cfg.Verbose)
    duplChan := make(chan syntax.Match)

    go func() {
        defer close(duplChan)
        matches := multiDetector.FindDuplOver(cfg.Threshold)
        for match := range matches {
            // NO context cancellation check!
            duplChan <- match
        }
    }()

    return duplChan, filesCount, nil
}
```

**Issues**:

- CLI uses `detection.MultiDetector` directly instead of `artdupl.Detector` SDK
- MultiDetector doesn't support context (different code path)
- No context checks in match processing loop

**Impact**: Well-implemented context support in SDK is completely bypassed by CLI

---

## Root Cause Analysis

The shutdown hook problems stem from a fundamental architectural disconnect:

1. **Signal handling is set up correctly at the entry point** (Fang in main.go)
2. **Context is never extracted or used** in the command execution
3. **Analysis pipeline functions don't accept context** parameters
4. **Long-running operations don't check for cancellation**
5. **SDK has proper context support** but CLI bypasses it

This creates a scenario where:

- User presses Ctrl+C → Fang creates context cancellation → Nothing in the code checks it → Operations continue
- User sets timeout → Context is created but never passed → Operations continue forever

---

## Impact Assessment

### User Experience Impact

**High Severity Issues**:

1. ❌ Ctrl+C does NOT stop the analysis - operations run to completion
2. ❌ Timeout flag (`--timeout`) is completely non-functional
3. ❌ User cannot cancel long-running operations on large codebases
4. ❌ Analysis may appear "frozen" when actually just slow
5. ❌ No way to interrupt operations consuming high CPU/memory

### System Impact

**Resource Management Issues**:

1. ❌ No way to free resources on user cancellation
2. ❌ May waste CPU/memory on unwanted operations
3. ❌ No graceful shutdown - abrupt termination only
4. ❌ Cannot enforce time limits on CI/CD systems

### Code Quality Impact

**Maintainability Issues**:

1. ❌ Dead code (timeout context created but never used)
2. ❌ Confusing comments admitting incomplete implementation
3. ❌ Inconsistent patterns between SDK and CLI
4. ❌ Missing fundamental Go concurrency best practices

---

## Recommended Fix Approach

### Phase 1: Context Propagation (Foundation)

1. **Update function signatures** to accept context:
   - `executeAnalysis(ctx context.Context, cfg *config.Config, paths []string)`
   - `buildSuffixTree(ctx context.Context, ...)`
   - `Parse(ctx context.Context, fchan chan string)`
   - `BuildTree(ctx context.Context, schan chan []*syntax.Node)`

2. **Pass context through the call chain**:

   ```go
   // In run.go
   ctx := cmd.Context() // Get context from Cobra (Fang's signal context)
   if mergedConfig.Timeout > 0 {
       var cancel context.CancelFunc
       ctx, cancel = context.WithTimeout(ctx, time.Duration(mergedConfig.Timeout)*time.Second)
       defer cancel()
   }
   duplChan, filesCount, err := executeAnalysis(ctx, mergedConfig, mergedConfig.Paths)
   ```

3. **Update all function calls** to pass context parameter

### Phase 2: Cancellation Checks (Implementation)

1. **Add context checks in job/parse.go**:

   ```go
   for file := range fchan {
       select {
       case <-ctx.Done():
           countChan <- fileCount
           close(achan)
           return
       default:
       }
       // ... process file ...
   }
   ```

2. **Add context checks in job/buildtree.go**:

   ```go
   for seq := range schan {
       select {
       case <-ctx.Done():
           done <- true
           return
       default:
       }
       // ... build tree ...
   }
   ```

3. **Add context checks in cmd/run.go**:
   ```go
   for match := range matches {
       select {
       case <-ctx.Done():
           close(duplChan)
           return
       default:
       }
       duplChan <- match
   }
   ```

### Phase 3: Testing (Verification)

1. **Add integration tests for signal handling**:
   - Send SIGINT and verify cancellation
   - Check resources are freed

2. **Add integration tests for timeout**:
   - Set short timeout and verify timeout occurs
   - Verify graceful shutdown

3. **Add BDD scenarios**:

   ```gherkin
   Scenario: User cancels analysis with Ctrl+C
     Given I start analysis of large codebase
     When I press Ctrl+C during analysis
     Then analysis should cancel within 1 second
     And resources should be freed
     And process should exit cleanly

   Scenario: Analysis times out
     Given I set timeout to 5 seconds
     And I start analysis of large codebase
     When analysis exceeds timeout
     Then analysis should cancel
     And timeout message should be displayed
   ```

### Phase 4: Documentation (Knowledge)

1. **Update README** with signal handling documentation
2. **Update CLI help** to mention Ctrl+C support
3. **Document timeout behavior** and recommended values
4. **Add examples** of graceful shutdown scenarios

---

## Implementation Priority

| Priority | Issue                                | Complexity | Impact | Estimate   |
| -------- | ------------------------------------ | ---------- | ------ | ---------- |
| P0       | Context propagation                  | Medium     | HIGH   | 2-3 hours  |
| P0       | Cancellation checks in job functions | Low        | HIGH   | 1 hour     |
| P0       | Fix timeout context passing          | Low        | HIGH   | 30 minutes |
| P1       | Integration tests                    | Medium     | MEDIUM | 2-3 hours  |
| P1       | BDD scenarios                        | Low        | MEDIUM | 1-2 hours  |
| P2       | Documentation updates                | Low        | LOW    | 1 hour     |
| P2       | Consider using SDK detector path     | High       | MEDIUM | 4-6 hours  |

**Total Estimate**: 12-16 hours for complete fix

---

## Testing Strategy

### Manual Testing Steps

1. **Test Ctrl+C cancellation**:

   ```bash
   # Start analysis of large codebase
   ./art-dupl ./large-project
   # Press Ctrl+C during tree building
   # Should exit within 1-2 seconds
   ```

2. **Test timeout functionality**:

   ```bash
   # Set short timeout
   ./art-dupl --timeout 5s ./large-project
   # Should timeout and exit after 5 seconds
   ```

3. **Test normal completion**:
   ```bash
   # Ensure normal operations still work
   ./art-dupl ./small-project
   ```

### Automated Testing Requirements

1. **Unit tests** for each function with context support
2. **Integration tests** for full pipeline cancellation
3. **Race detector testing** (`go test -race`)
4. **Performance tests** to ensure no performance regression

---

## Related Code Patterns in Codebase

### Proper Context Usage Examples

The codebase already has good examples of context handling:

1. **`pkg/artdupl/detector.go:136-139`** - Proper context cancellation check
2. **`pkg/artdupl/detector.go:158-160`** - Proper context cancellation in loop
3. **`pkg/artdupl/detector.go:182-189`** - Proper timeout and cancellation handling
4. **`pkg/artdupl/detector.go:237-241`** - Proper context cancellation in detection

These patterns should be replicated throughout the CLI code.

---

## Next Steps

### Immediate Actions

1. ✅ **Document all identified problems** (this report)
2. ⏳ **Create TODO items** for each fix phase
3. ⏳ **Implement Phase 1** (Context propagation)
4. ⏳ **Implement Phase 2** (Cancellation checks)
5. ⏳ **Test all fixes** thoroughly
6. ⏳ **Update documentation**

### Follow-up Considerations

1. **Consider consolidating CLI path to use SDK detector** for consistency
2. **Add context cancellation to other long-running operations** if found
3. **Consider adding graceful shutdown hooks** for cleanup operations
4. **Add metrics/telemetry** for cancellation events (optional)

---

## References

- **Fang Documentation**: https://github.com/charmbracelet/fang
- **Go Context Package**: https://pkg.go.dev/context
- **Go Signal Handling**: https://pkg.go.dev/os/signal#NotifyContext
- **Existing SDK Implementation**: `pkg/artdupl/detector.go`

---

## Conclusion

The shutdown hook implementation has fundamental architectural problems that completely disable both signal handling and timeout functionality. The fix requires:

1. ✅ Adding context parameters throughout the analysis pipeline
2. ✅ Implementing context cancellation checks in all long-running loops
3. ✅ Properly passing the Fang context (with signal handling) through the command chain
4. ✅ Integrating timeout context into the analysis pipeline
5. ✅ Adding comprehensive tests for cancellation scenarios

The good news is that:

- The SDK (`pkg/artdupl/detector.go`) already has excellent context support
- The patterns are already present in the codebase
- The fix is well-understood and straightforward
- Estimated effort is reasonable (12-16 hours)

**This should be prioritized as P0 given the severity of impact on user experience and system reliability.**

---

_Report generated on 2026-01-25 at 04:01 CET_
