# Actionability Patterns Reference

art-dupl classifies clone groups as **Actionable** or **NonActionable**. A group is
NonActionable only when EVERY clone in the group matches the same boilerplate pattern.
These patterns represent Go idioms that cannot be eliminated without breaking semantics.

## Detected Patterns

| Pattern              | Label                      | Description                                                                                                    | Example                                     |
| -------------------- | -------------------------- | -------------------------------------------------------------------------------------------------------------- | ------------------------------------------- |
| Signature-only       | `signature-only`           | FuncDecl without a body (interface stub, forwarding method)                                                    | `func (s *Svc) Name() string`               |
| Interface impl       | `interface-implementation` | 3+ FuncType fragments from different files (satisfies common interface)                                        | Multiple files implementing `io.Reader`     |
| Interface method     | `interface-method`         | FuncDecl body matching common stdlib interface method name (String, Read, Close, etc.) with ≤4 body statements | `func (t Time) String() string { ... }`     |
| RAII defer           | `raii-defer`               | DeferStmt wrapping cleanup (Unlock, Close, etc.)                                                               | `defer m.Unlock()`                          |
| Error propagation    | `error-propagation`        | `if err != nil { return err }`                                                                                 | Pure error forwarding                       |
| Assign+error-check   | `assign-error-check`       | 2-stmt: `err := f(); if err != nil { return }`                                                                 | Most common Go boilerplate                  |
| Single call          | `single-call-expression`   | Lone CallExpr or ExprStmt(CallExpr) (different data, same API)                                                 | `t.Parallel()`, `errors.New("foo")`         |
| Single simple stmt   | `single-simple-statement`  | Lone terminal statement (return, assignment, var declaration, etc.)                                            | `return nil`, `x := 0`, `var buf []byte`    |
| Single declaration   | `single-declaration`       | Lone package-level ValueSpec/TypeSpec-alias (re-export, const alias, iota starter)                             | `type Mode = domain.Mode`, `BadNode = iota` |
| Test helper delegate | `test-helper-delegate`     | 2-stmt body: `t.Helper()` + single delegate call (irreducible Go test boilerplate)                             | `t.Helper()` + `failIfNilf(t, got, ...)`    |
| Guard clause         | `guard-clause`             | IfStmt with return-only body and no else (boolean/value guard)                                                 | `if !enabled { return }`                    |
| Error wrapping       | `error-wrapping`           | `if err != nil { return fmt.Errorf(...) }`                                                                     | Error wrapping idiom                        |
| Assertion chain      | `assertion-chain`          | 3+ test assertion calls (Expect/Assert/Require)                                                                | `Expect(x).To(Equal(y))`                    |
| Cobra boilerplate    | `cobra-boilerplate`        | `cobra.Command{}` or `fang.Command{}` struct literals                                                          | CLI framework setup                         |
| Test data pair       | `testdata-pair`            | All clones from `testdata/` directories                                                                        | Golden/input file pairs                     |
| Table-driven test    | `table-driven-test`        | RangeStmt with `t.Run()` in `_test.go`                                                                         | Standard Go test pattern                    |
| Test scaffolding     | `test-scaffolding`         | TempDir + WriteFile + assertions in `_test.go`                                                                 | Test setup/teardown                         |
| Data-dominated       | `data-dominated`           | 60%+ BasicLit/KeyValueExpr nodes                                                                               | Config fixtures, struct init                |
| Describe table       | `describe-table`           | Ginkgo `DescribeTable`/`Entry` patterns                                                                        | Ginkgo parametrized tests                   |
| Builder callback     | `builder-callback`         | 3+ chained calls on 2+ different receivers                                                                     | Builder/fluent API pattern                  |

## How It Works

1. `EvaluateActionabilityWithLabel` runs 20 pattern checks in priority order.
2. The first matching pattern wins (returns its `PatternLabel`).
3. If no pattern matches, the group is **Actionable**.
4. Only `semantic` detection mode runs actionability checks. `exact` and `structural` skip them.
5. Use `--list-patterns` to print all labels, `--disable-pattern <label>` to selectively re-enable a pattern, or `--no-actionability` to disable all filtering.

## Pattern Priority Order

Patterns are checked in this order (first match wins):

1. Signature-only
2. Interface implementation
3. Interface method
4. RAII defer
5. Error propagation
6. Guard clause
7. Assign+error-check
8. Single call expression
9. Single simple statement
10. Single declaration
11. Test helper delegate
12. Error wrapping
13. Assertion chain
14. Cobra boilerplate
15. Test data pair
16. Table-driven test
17. Test scaffolding
18. Data-dominated
19. Describe table
20. Builder callback
21. Bool-guard (`X, ok := helper(); if !ok { return }`)
22. Templ-rendering-idiom (`if len(x) == 0 { text } else { for ... }`)

## Property-Based Classification Engine (Second Pass)

After the 22 pattern-table checks run, a **property-based extractability engine** runs as a second-pass fallback. It evaluates four computable properties that define "harmful duplication" from first principles, catching false positives that the pattern table misses.

See `docs/adr/0017-property-based-classification.md` for the full design.

### The Four Properties

| #   | Property                        | What it checks                                                  | Example false positives killed                  |
| --- | ------------------------------- | --------------------------------------------------------------- | ----------------------------------------------- |
| 1   | **Control-flow extractability** | Clone contains `return`/`break`/`continue` forced by caller sig | HTTP handler void-return error guards           |
| 2   | **ROI positive**                | >60% of tokens inside a single CallExpr = helper invocation     | `queryError` wrappers, `defer cancel()`         |
| 3   | **Parameterizable**             | Clones differ only in string-literal domain values              | bool-to-string funcs, format specifier variants |
| 4   | **Mechanical extractability**   | Always true (any code can be wrapped in a function)             | —                                               |

### Confidence Tiers

The engine produces a confidence score (0.0-1.0) that maps to three tiers:

| Tier             | Confidence | Meaning                                                   |
| ---------------- | ---------- | --------------------------------------------------------- |
| `actionable`     | >= 0.8     | Genuinely harmful duplication — extract it                |
| `low-confidence` | 0.5-0.8    | Ambiguous — review manually, may need `//art-dupl:accept` |
| `non-actionable` | < 0.5      | Idiomatic boilerplate — suppressed from default output    |

The `low-confidence` tier is shown in `--explain` output and included as `confidence` in JSON.

### Architecture

The property engine runs **after** the pattern table (not as a pre-filter). Patterns are specific and well-tested; the property engine catches what they miss. This prevents the engine from stealing labels from existing patterns.

## Key Design Decisions

- **ALL clones must match**: A group is NonActionable only when EVERY clone matches the same pattern. If any clone differs, the group is Actionable.
- **PatternLabel feeds into classification**: The detected pattern adjusts the clone's category and priority (e.g., testdata pair -> CategoryTestFixture, PriorityLow).
- **No mutation**: Pattern detection operates on `domain.CloneNode` trees (immutable copies), not `syntax.Node` (which carries mutable serialization state).
- **Interface method threshold**: The `interface-method` pattern uses a static name list (`commonInterfaceMethodNames`) covering 25+ stdlib interface method names and a body size limit of ≤4 statements. A deeper type-aware variant using `go/types` is tracked in ROADMAP.
