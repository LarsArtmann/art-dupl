# Feedback: Aggressive threshold (t=2) on an interface-heavy Go monorepo

> ✅ **ADDRESSED** — Default threshold is now 5 (commit `930b91a`). This feedback confirmed that t=5 produces 0 clone groups on this project, which is now the default. Semantic mode correctly handles interface-implementation patterns via the `interface-implementation` actionability pattern.

**Date:** 2026-07-06
**Project:** go-auto-upgrade (Go, ~3500 LOC, 15 packages, 3 migrators sharing a Migrator interface)
**Command:** `art-dupl --semantic --sort total-tokens -t 2`
**Result:** 14 clone groups at t=2. 0 clone groups at t=5 (the skill's recommended default).

## Session Summary

Ran art-dupl at the extremely aggressive threshold 2 on a Go project built around a shared `Migrator` interface with three implementations (jsonv1tov2, lo2stdlib, stdlibwrappers). The goal was to find and eliminate every meaningful clone.

**What was found:** 14 clone groups totaling 114 tokens.
**What was actually deduplicated:** 3 groups (3 new helpers extracted). The remaining 11 groups are interface contracts, type references, mutex patterns, and sentinel error declarations — not real duplication.

## What Worked Well

### Semantic detection correctly identifies structural clones across packages

The tool found the same guard pattern (`if len(rangeStmt.Body.List) != 1` + type-assertion) in 5 locations across 3 files in `stdlibwrappers/`, even though the variable names and surrounding code differed. Extracting a single `rangeBodySingleIf` helper eliminated the entire group.

### Helper-consumption re-detection is honest

After extracting `pendingCall`, art-dupl reported the two call sites as a new 2-clone group. This is **correct behavior** — the call sites are now structurally identical because they consume the same helper. It's not a bug; it's the tool accurately reflecting that two functions now share an identical 4-line preamble. The duplication moved from "type assertion logic" to "helper call pattern," which is an acceptable trade-off.

## Actionable Deduplications Found

### 1. Extracted `rangeBodySingleIf` (5 sites eliminated)

```go
// Before — duplicated in loop_detection.go, detection_equal_join.go, rewriter.go
if len(rangeStmt.Body.List) != 1 {
    return false
}
ifStmt, ok := rangeStmt.Body.List[0].(*dst.IfStmt)
if !ok {
    return false
}

// After — single helper in tables.go
func rangeBodySingleIf(rangeStmt *dst.RangeStmt) (*dst.IfStmt, bool) {
    if len(rangeStmt.Body.List) != 1 {
        return nil, false
    }
    ifStmt, ok := rangeStmt.Body.List[0].(*dst.IfStmt)
    return ifStmt, ok
}
```

### 2. Extracted `assignIndexIdent` (2 sites eliminated)

The `if !ok { return nil }; name := indexExprIdent(x); if name == "" { return nil }` opener was duplicated in `extractMinMaxArgs` and `extractJoinArgs`.

### 3. Extracted `pendingCall` (2 sites unified)

The `rewrite.node.(*dst.CallExpr)` type-assertion guard was duplicated in `doAutoSimple` and `doAutoKeysValues`.

## False Positives at Threshold 2

The remaining 11 groups are all Go idioms or contractual patterns:

### Shared domain type references (Groups #1, #2, #3 — 82 occurrences)

```go
// 39 sites report this as a "clone" — it's just using the shared type
func readFile(path domain.PathString) ([]byte, error) {
func writeToFile(path domain.PathString, content []byte) error {
func (m *Migrator) Migrate(_ context.Context, path domain.PathString, ...
```

Every function that accepts a `domain.PathString` parameter triggers a clone match. This is the **intended consequence** of using a shared branded type — the type is _supposed_ to appear everywhere. Extracting would mean renaming the type or breaking the public API.

### Interface method signatures (Group #7 — 3 implementations)

```go
// All three migrators MUST have identical signatures — it's the interface contract
func (m *Migrator) Name() domain.MigratorName { return Name }
func (m *Migrator) Description() string { return "..." }
func (m *Migrator) MinGoVersion() domain.GoVersion { return ... }
```

The `Migrator` interface defines these methods. Every implementation must match. Three implementations = three identical signatures. This is not duplication — it's polymorphism.

### RWMutex lock/unlock patterns (Groups #5, #10 — 5 sites)

```go
// registry.go — appears in TryRegister, Get, All, Unregister, Clear
global.mu.Lock()
defer global.mu.Unlock()
```

Idiomatic Go sync pattern. The `defer` must be inline because of Go's defer semantics. Cannot be extracted without callbacks or closures that would obscure the code.

### Sentinel error declarations (Groups #6, #14 — 11 sites)

```go
// pkg/errors/errors.go — each is a DISTINCT error value
ErrPathNotExist = errorfamily.NewRejection("path.not_found", "...")
ErrRequiresFileArg = errorfamily.NewRejection("args.missing_paths", "...")
ErrMigratorNotFound = errorfamily.NewRejection("registry.migrator_not_found", "...")
```

Each sentinel carries a unique code and message. They look structurally identical because they all use the same builder — but they're distinct values by design.

### Helper consumption patterns (Groups #8, #12, #13 — 6 sites)

After extracting `ParseMigratorInput`, `firstAssignStmtOfN`, `firstDefineAssignStmtOfN`, and `pendingCall`, the call sites of these helpers are now identical. art-dupl reports them as clones. But this is the **correct outcome of using a helper** — the call sites are intentionally identical because they consume the same abstraction.

## Metrics

| Metric                  | Value |
| ----------------------- | ----- |
| Clone groups (t=2)      | 14    |
| Clone groups (t=5)      | 0     |
| Real duplications found | 3     |
| False positives         | 11    |
| False positive rate     | 79%   |
| Production clones       | 95    |
| Test clones             | 19    |
| Actionable (production) | 3     |
| Actionable (test)       | 0     |

## Suggestions

### 1. Interface-method-aware suppression

The single largest false-positive source is interface method signatures. When a function signature matches an interface method declaration in the same module (or an imported package), it should be down-ranked or annotated. The tool could:

- Detect `func (recv T) MethodName(...) ...` where `MethodName` matches a method in an interface type in the same package or an imported `pkg.Migrator`-style interface
- Tag these as `interface-impl` category instead of `unknown`
- Down-rank to `info` priority

This would have eliminated Groups #1, #2, #3, and #7 (55 of the 114 tokens) in this session.

### 2. Shared-type-reference suppression

When a clone group consists entirely of single-line type references (e.g., `domain.PathString` as a parameter type, `domain.GoVersion` as a return type), these are uses of a shared branded type — not duplication. A heuristic: if the clone spans exactly one token (a type identifier) and appears in function signatures across multiple files, suppress or annotate as `type-ref`.

### 3. Threshold guidance in the CLI output

At t=2, 79% of results were false positives. At t=5, there were zero clones. The skill documentation recommends t=5 as the default, but users who run `art-dupl -t 2` (or lower) get no warning that they're entering a noise-heavy zone.

A one-line note in the report summary would help:

```
⚠️  Threshold 2 is below the recommended minimum (5).
    Expect a high false-positive rate from interface signatures
    and type references.
```

### 4. Mutex-pattern recognition

`mu.Lock(); defer mu.Unlock()` and `mu.RLock(); defer mu.RUnlock()` are among the most common Go idioms. Recognizing these as a built-in pattern category (`sync-mutex`) and down-ranking them would eliminate 2 clone groups (8 occurrences) in this project. This pattern is structurally invariant — there's no way to "deduplicate" it without sacrificing clarity.

## Conclusion

At t=5, art-dupl is excellent: zero noise, the codebase is clean. At t=2, the signal-to-noise ratio drops to 21% — but the 3 real findings were genuinely valuable (5-site elimination, 2-site elimination, 2-site unification).

The tool's semantic detection is strong. The main improvement opportunity is **interface-awareness**: recognizing that interface method implementations and shared branded-type references are structurally identical by contract, not by accident. This would dramatically improve the experience at low thresholds.
