# Actionability Patterns Reference

art-dupl classifies clone groups as **Actionable** or **NonActionable**. A group is
NonActionable only when EVERY clone in the group matches the same boilerplate pattern.
These patterns represent Go idioms that cannot be eliminated without breaking semantics.

## Detected Patterns

| Pattern               | Label                          | Description                                                                                                                                                                                                                                                          | Example                                     |
| --------------------- | ------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------- |
| Signature-only        | `signature-only`               | FuncDecl without a body (interface stub, forwarding method)                                                                                                                                                                                                          | `func (s *Svc) Name() string`               |
| Interface impl        | `interface-implementation`     | 3+ FuncType fragments from different files (satisfies common interface)                                                                                                                                                                                              | Multiple files implementing `io.Reader`     |
| Interface method      | `interface-method`             | Statement-level clone inside a function that implements an interface contract (≤4 statements). Type-aware path propagates `InterfaceMethod` flag from FuncDecl to body statements; static fallback matches stdlib method names on FuncDecl-rooted clones (edge case) | `func (t Time) String() string { ... }`     |
| RAII defer            | `raii-defer`                   | DeferStmt wrapping cleanup (Unlock, Close, etc.)                                                                                                                                                                                                                     | `defer m.Unlock()`                          |
| Error propagation     | `error-propagation`            | `if err != nil { return err }`                                                                                                                                                                                                                                       | Pure error forwarding                       |
| Assign+error-check    | `assign-error-check`           | 2-stmt: `err := f(); if err != nil { return }`                                                                                                                                                                                                                       | Most common Go boilerplate                  |
| Single call           | `single-call-expression`       | Lone CallExpr or ExprStmt(CallExpr) (different data, same API)                                                                                                                                                                                                       | `t.Parallel()`, `errors.New("foo")`         |
| Single simple stmt    | `single-simple-statement`      | Lone terminal statement (return, assignment, var declaration, etc.)                                                                                                                                                                                                  | `return nil`, `x := 0`, `var buf []byte`    |
| Bool accumulator init | `bool-accumulator-initializer` | Two or more consecutive `name := true/false` or `var name bool = true/false` (independent boolean flags)                                                                                                                                                             | `hasX := false` / `hasY := false`           |
| Single declaration    | `single-declaration`           | Lone package-level ValueSpec/TypeSpec-alias (re-export, const alias, iota starter)                                                                                                                                                                                   | `type Mode = domain.Mode`, `BadNode = iota` |
| Test helper delegate  | `test-helper-delegate`         | 2-stmt body: `t.Helper()` + single delegate call (irreducible Go test boilerplate)                                                                                                                                                                                   | `t.Helper()` + `failIfNilf(t, got, ...)`    |
| Guard clause          | `guard-clause`                 | IfStmt with return-only body and no else (boolean/value guard)                                                                                                                                                                                                       | `if !enabled { return }`                    |
| Error wrapping        | `error-wrapping`               | `if err != nil { return fmt.Errorf(...) }`                                                                                                                                                                                                                           | Error wrapping idiom                        |
| Assertion chain       | `assertion-chain`              | 3+ test assertion calls (Expect/Assert/Require)                                                                                                                                                                                                                      | `Expect(x).To(Equal(y))`                    |
| Cobra boilerplate     | `cobra-boilerplate`            | `cobra.Command{}` or `fang.Command{}` struct literals                                                                                                                                                                                                                | CLI framework setup                         |
| Test data pair        | `testdata-pair`                | All clones from `testdata/` directories                                                                                                                                                                                                                              | Golden/input file pairs                     |
| Table-driven test     | `table-driven-test`            | RangeStmt with `t.Run()` in `_test.go`                                                                                                                                                                                                                               | Standard Go test pattern                    |
| Test scaffolding      | `test-scaffolding`             | TempDir + WriteFile + assertions in `_test.go`                                                                                                                                                                                                                       | Test setup/teardown                         |
| Data-dominated        | `data-dominated`               | 60%+ BasicLit/KeyValueExpr nodes                                                                                                                                                                                                                                     | Config fixtures, struct init                |
| Describe table        | `describe-table`               | Ginkgo `DescribeTable`/`Entry` patterns                                                                                                                                                                                                                              | Ginkgo parametrized tests                   |
| Builder callback      | `builder-callback`             | 3+ chained calls on 2+ different receivers                                                                                                                                                                                                                           | Builder/fluent API pattern                  |
| Bool-guard            | `bool-guard`                   | 2-stmt: `X, ok := helper(); if !ok { return }` (assign + guard on same variable)                                                                                                                                                                                     | `v, ok := m[key]; if !ok { return }`        |
| Templ rendering       | `templ-rendering-idiom`        | `if len(x) == 0 { ... } else { for ... }` (templ/HTML empty-state convention; `.go` files, not `.templ` source)                                                                                                                                                      | Generated `_templ.go` empty-state rendering |
| Defer call            | `defer-call`                   | Bare `defer cleanupFunc()` (Ident callee only, NOT method calls like `defer svc.processOrder()`)                                                                                                                                                                    | `defer unsubscribe()`                       |
| Test framework call   | `test-framework-call`          | Test framework method calls (`t.Parallel()`, `b.Helper()`, `t.Cleanup()`, etc.)                                                                                                                                                                                     | `t.Parallel()`                              |
| State flag mutation   | `state-flag-mutation`          | Single field assignment in methods (`x.flag = true`)                                                                                                                                                                                                                | `w.flushed = true`                          |
| Empty default         | `empty-default`                | `if x == "" { x = default }` idiom                                                                                                                                                                                                                                  | `if name == "" { name = "unknown" }`      |

## How It Works

1. `EvaluateActionabilityWithLabel` runs 29 pattern checks in priority order.
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
8. Bool-guard (`X, ok := helper(); if !ok { return }`)
9. Single call expression
10. Single simple statement
11. Bool accumulator init (`name := true/false` pairs)
12. Interface assertion (`var _ I = (*T)(nil)`)
13. Single declaration (true aliases only: `type X = pkg.Y`)
14. Type alias block (2+ consecutive `type X = pkg.Y` re-export shims)
15. Test helper delegate (`t.Helper()` + delegate call)
16. Error wrapping
17. Assertion chain
18. Cobra boilerplate
19. Test data pair
20. Table-driven test
21. Test scaffolding
22. Data-dominated
23. Describe table
24. Builder callback
25. Templ-rendering-idiom (`if len(x) == 0 { text } else { for ... }`)
26. Defer-call (bare `defer cleanupFunc()` — Ident callee only)
27. Test-framework-call (`t.Parallel()`, `b.Helper()`, `t.Cleanup()`)
28. State-flag-mutation (`x.flag = true` single field assignment)
29. Empty-default (`if x == "" { x = default }`)

## Property-Based Classification Engine (Second Pass)

After the 29 pattern-table checks run, a **property-based extractability engine** runs as a second-pass fallback. It evaluates four computable properties that define "harmful duplication" from first principles, catching false positives that the pattern table misses.

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
- **Interface method detection**: The `interface-method` pattern has two paths. **Path 1** (FuncDecl root, edge case) checks the `InterfaceMethod` flag or the static `commonInterfaceMethodNames` list on FuncDecl clone roots — this only fires in files without statements. **Path 2** (statement root, normal Go files) checks the `InterfaceMethod` flag propagated from the enclosing FuncDecl to body statement nodes by the transformer (same save/restore pattern as `EnclosingReturnArity`). This propagation is necessary because the structural filter in `FindSyntaxUnits` rejects non-Statement clone roots, making FuncDecl-rooted clones unreachable in real Go files. When `--type-aware` is active, `golang.IsInterfaceMethod()` uses `go/types` to check if the method satisfies a same-package interface. The static name list covers 25+ stdlib interface method names for the edge-case path. Body size limit is ≤4 statements. Cross-package interface scanning remains future work (see ROADMAP).
