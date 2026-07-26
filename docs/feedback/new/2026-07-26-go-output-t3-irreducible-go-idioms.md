# Feedback: Threshold 3 reports irreducible Go idioms as actionable clones

**Date:** 2026-07-26  
**Project:** `go-output`, a 19-module Go workspace for CLI output formatting and progress visualization  
**Command:** `art-dupl --semantic --sort total-tokens -t 3 --html`  
**Goal:** Review every finding and drive harmful duplication to zero without introducing detector-driven abstractions.

> Verdict: The report found **3 groups / 6 occurrences / 18 tokens**, but none represented harmful duplication. All three were minimum Go idioms or intentionally module-local API structure. The codebase has zero findings at `-t 4`; at `-t 3`, literal zero would require hiding lock scope, wrapping `strings.Builder`, coupling independent modules, or gaming source shape. The report should classify and suppress or down-rank these patterns instead of recommending “Review and extract common logic.”

## Results

| Group                          | Locations                                                  | Report category   | Actual category                  | Decision |
| ------------------------------ | ---------------------------------------------------------- | ----------------- | -------------------------------- | -------- |
| Read lock plus timestamp       | `nom/activity_snapshot.go:84`, `nom/state_accessors.go:97` | `unknown`, low    | Go synchronization idiom         | Accept   |
| String-builder opener          | `markup/xml.go:114`, `plantuml/plantuml.go:50`             | `expression`, low | Standard-library usage           | Accept   |
| Functional-option declarations | `markdown/cqrs.go:10`, `tree/cqrs.go:12`                   | `unknown`, low    | Module-local public API contract | Accept   |

No source files were modified. Each possible extraction was evaluated against readability, concurrency transparency, type safety, and module boundaries.

## Finding 1: Read-lock scope plus timestamp

The report matched:

```go
ns.mu.RLock()
defer ns.mu.RUnlock()

now := time.Now()
```

The surrounding functions are unrelated:

- `SnapshotActivities` copies mutable activity state into immutable snapshots.
- `EstimatedTotalRemaining` computes an aggregate duration for unfinished work.

The only shared concept is that both need a consistent view of subscriber state at one point in time.

### Why extraction would be worse

The apparent extraction would require a callback helper such as:

```go
func withReadLockAt[T any](ns *NOMSubscriber, read func(time.Time) T) T
```

That abstraction would:

1. Hide the exact mutex scope from readers.
2. Make lock-order and deadlock audits less direct.
3. Introduce a generic callback for only two unrelated computations.
4. Add closure and type machinery around three canonical statements.
5. Save no domain logic or maintenance burden.

In concurrency-sensitive Go code, visible lock acquisition and `defer`-based release are desirable. The lock scope is part of the function's correctness proof, not accidental boilerplate.

### Tool feedback

This should be recognized as a built-in Go idiom when the sequence consists of:

```go
receiver.mu.RLock()
defer receiver.mu.RUnlock()
```

optionally followed by a timestamp read. Suggested classification:

- Pattern: `sync-rwmutex-read`
- Priority: informational or suppressed at thresholds below the default
- Suggestion: “Idiomatic read-lock scope; extraction usually obscures synchronization.”

This extends the mutex-pattern feedback already documented for threshold 2. The notable addition here is that appending `now := time.Now()` makes the sequence long enough to survive at threshold 3.

## Finding 2: `strings.Builder` initialization

The report matched the opening of two renderers:

```go
var b strings.Builder
b.WriteString("...")
b.WriteString("...")
```

The complete functions have different responsibilities:

- `markup.MarshalXMLFromTable` serializes table data using XML grammar and XML escaping.
- `plantuml.PlantUMLDiagram.Render` emits PlantUML graph syntax with PlantUML-specific escaping and styling.

Only the standard mechanism for constructing strings is shared. `strings.Builder` is already the abstraction.

### Why extraction would be worse

Potential fixes include a helper that accepts initial strings, a builder constructor that preloads lines, or a callback-based rendering helper. Each would merely replace direct standard-library calls with an indirect wrapper:

```go
b := newBuilder("header one", "header two")
```

That helper has no domain meaning. It would couple unrelated output grammars and still require both renderers to contain all meaningful logic independently. It also risks encouraging a generic rendering abstraction where escaping and grammar must remain format-specific.

Changing one renderer to use `fmt.Fprint`, `io.WriteString`, or a seeded string solely to alter its token sequence would achieve literal zero only by gaming the detector.

### Tool feedback

A clone whose normalized content is only:

- declaration of `strings.Builder`, followed by
- two or three `WriteString` calls with different literals

should be classified as standard-library builder scaffolding rather than actionable duplication.

Suggested classification:

- Pattern: `strings-builder-opener`
- Priority: informational or suppressed at aggressive thresholds
- Suggestion: “Standard string-construction idiom; inspect surrounding grammar before extracting.”

The current HTML suggestion, “Review and extract common logic,” is actively misleading because no common logic is present.

## Finding 3: Module-local functional-option declarations

The report matched this public API structure in the `markdown` and `tree` modules:

```go
type Option func(*Config)

type Config struct {
    output.ColorConfig
}

func WithColorMode(mode output.ColorMode) Option {
    return func(c *Config) { c.ColorMode = mode }
}
```

The genuinely shared model has already been extracted into the dependency-light root module:

```go
output.ColorConfig
output.DefaultColorConfig()
```

What remains is the minimum module-local functional-option shell required by Go.

### Why extraction would be worse

`markdown.Option` and `tree.Option` configure different renderer APIs. Their function parameters target different module-local `Config` types. Sharing the option type would require one of these compromises:

1. Move renderer-specific configuration into root, making root own sub-module API concerns.
2. Introduce a generic root option type, increasing API complexity for one field.
3. Alias both configurations to one root type, making independent modules evolve together.
4. Make options interchangeable across renderers, weakening type safety and semantic ownership.
5. Merge modules, violating the workspace's deliberate dependency boundaries.

The modules currently happen to expose only `ColorMode`, but they are independent extension points. Their configurations may diverge without forcing unrelated API changes. The local `Option func(*Config)` types preserve that freedom.

### Tool feedback

The detector should distinguish duplicated functional-option declarations from duplicated behavior, especially across Go module or package boundaries.

Suggested classification:

- Pattern: `functional-option-contract`
- Priority: informational
- Actionability test:
  - local named `Option` function type;
  - parameter points to a local `Config` type;
  - `WithX` returns a closure mutating that local config;
  - occurrences are in different packages.

The report could state: “Package-local functional-option contracts are often intentionally repeated because Go lacks parameterized type aliases over local configuration ownership.”

## HTML report quality

The HTML output correctly showed all source locations and concise fragments, but its metadata overstated actionability:

- All six occurrences were counted as “Production.” That is factually true but implies they are production defects rather than reviewed idioms.
- Two groups were categorized as `unknown` despite recognizable Go patterns.
- Every group received the same suggestion: “Review and extract common logic.”
- The summary offered no distinction between clone findings and actionable clone findings.

At an aggressive threshold, the report needs stronger wording discipline. A low-priority finding can still cause users or automated agents to manufacture abstractions to satisfy a literal-zero instruction.

## Recommended improvements

### 1. Add idiom-aware actionability patterns

Recognize and annotate:

- `sync.Mutex` and `sync.RWMutex` lock/defer-unlock scopes;
- `strings.Builder` declaration plus initial writes;
- package-local functional-option contracts.

These should be down-ranked or suppressed at low thresholds unless additional duplicated domain logic follows them.

### 2. Separate “detected” from “actionable” totals

The summary should report both:

```text
Detected clone groups: 3
Actionable clone groups: 0
Recognized idioms/contracts: 3
```

This preserves detector honesty without encouraging harmful changes.

### 3. Tailor suggestions to classification

Replace the unconditional suggestion with pattern-aware guidance. For example:

```text
Idiomatic synchronization pattern. Keep lock scope visible unless duplicated domain logic also exists.
```

```text
Standard strings.Builder setup. Similar construction mechanics alone are not a shared abstraction.
```

```text
Package-local functional-option API. Confirm configs share lifecycle and ownership before unifying.
```

### 4. Warn when threshold 3 enters idiom territory

The default threshold is already higher, but an explicit report notice would help:

```text
Threshold 3 is aggressive for Go. Expect minimum idioms and API-contract scaffolding; review actionability before extracting.
```

### 5. Preserve zero as “zero harmful duplication”

The tool and companion workflow should explicitly distinguish:

- **Literal detector zero**, which may reward source-shape manipulation.
- **Zero actionable duplication**, where every remaining group is classified and defensibly accepted.

For this session:

| Threshold | Groups | Assessment                               |
| --------- | -----: | ---------------------------------------- |
| 4         |      0 | Clean report                             |
| 3         |      3 | All intentional minimum idioms/contracts |

## Acceptance criteria for a fix

Run the same command against `go-output`:

```bash
art-dupl --semantic --sort total-tokens -t 3 --html
```

A successful improvement should produce one of these outcomes:

1. Zero actionable groups, with all three findings suppressed as recognized idioms; or
2. Three informational groups, accurately classified, with no extraction recommendation and an actionable count of zero.

The tool should continue reporting longer sequences where a lock or builder opener is followed by genuinely duplicated domain logic. The goal is not blanket suppression of `RLock` or `strings.Builder`, but correct actionability at the minimum sequence length.

## Conclusion

`art-dupl` accurately found structural similarity, but at threshold 3 it could not distinguish syntax reuse from shared behavior. In this codebase, forcing literal zero would reduce clarity, concurrency auditability, type safety, or module independence. The correct result is **zero harmful duplication with three accepted idioms**, and the report should be able to express that directly.
