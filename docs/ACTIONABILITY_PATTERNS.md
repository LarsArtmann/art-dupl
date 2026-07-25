# Actionability Patterns Reference

art-dupl classifies clone groups as **Actionable** or **NonActionable**. A group is
NonActionable only when EVERY clone in the group matches the same boilerplate pattern.
These patterns represent Go idioms that cannot be eliminated without breaking semantics.

## Detected Patterns

| Pattern            | Label                      | Description                                                             | Example                                 |
| ------------------ | -------------------------- | ----------------------------------------------------------------------- | --------------------------------------- |
| Signature-only     | `signature-only`           | FuncDecl without a body (interface stub, forwarding method)             | `func (s *Svc) Name() string`           |
| Interface impl     | `interface-implementation` | 3+ FuncType fragments from different files (satisfies common interface) | Multiple files implementing `io.Reader` |
| RAII defer         | `raii-defer`               | DeferStmt wrapping cleanup (Unlock, Close, etc.)                        | `defer m.Unlock()`                      |
| Error propagation  | `error-propagation`        | `if err != nil { return err }`                                          | Pure error forwarding                   |
| Assign+error-check | `assign-error-check`       | 2-stmt: `err := f(); if err != nil { return }`                          | Most common Go boilerplate              |
| Single call        | `single-call-expression`   | Lone CallExpr or ExprStmt(CallExpr) (different data, same API)          | `t.Parallel()`, `errors.New("foo")`    |
| Single simple stmt| `single-simple-statement`  | Lone terminal statement (return, assignment, var declaration, etc.)     | `return nil`, `x := 0`, `var buf []byte`|
| Guard clause      | `guard-clause`             | IfStmt with return-only body and no else (boolean/value guard)          | `if !enabled { return }`                |
| Error wrapping     | `error-wrapping`           | `if err != nil { return fmt.Errorf(...) }`                              | Error wrapping idiom                    |
| Assertion chain    | `assertion-chain`          | 3+ test assertion calls (Expect/Assert/Require)                         | `Expect(x).To(Equal(y))`                |
| Cobra boilerplate  | `cobra-boilerplate`        | `cobra.Command{}` or `fang.Command{}` struct literals                   | CLI framework setup                     |
| Test data pair     | `testdata-pair`            | All clones from `testdata/` directories                                 | Golden/input file pairs                 |
| Table-driven test  | `table-driven-test`        | RangeStmt with `t.Run()` in `_test.go`                                  | Standard Go test pattern                |
| Test scaffolding   | `test-scaffolding`         | TempDir + WriteFile + assertions in `_test.go`                          | Test setup/teardown                     |
| Data-dominated     | `data-dominated`           | 60%+ BasicLit/KeyValueExpr nodes                                        | Config fixtures, struct init            |
| Describe table     | `describe-table`           | Ginkgo `DescribeTable`/`Entry` patterns                                 | Ginkgo parametrized tests               |
| Builder callback   | `builder-callback`         | 3+ chained calls on 2+ different receivers                              | Builder/fluent API pattern              |

## How It Works

1. `EvaluateActionabilityWithLabel` runs 17 pattern checks in priority order.
2. The first matching pattern wins (returns its `PatternLabel`).
3. If no pattern matches, the group is **Actionable**.
4. Only `semantic` detection mode runs actionability checks. `exact` and `structural` skip them.

## Pattern Priority Order

Patterns are checked in this order (first match wins):

1. Signature-only
2. Interface implementation
3. RAII defer
4. Error propagation
5. Guard clause
6. Assign+error-check
7. Single call expression
8. Single simple statement
9. Error wrapping
10. Assertion chain
11. Cobra boilerplate
12. Test data pair
13. Table-driven test
14. Test scaffolding
15. Data-dominated
16. Describe table
17. Builder callback

## Key Design Decisions

- **ALL clones must match**: A group is NonActionable only when EVERY clone matches the same pattern. If any clone differs, the group is Actionable.
- **PatternLabel feeds into classification**: The detected pattern adjusts the clone's category and priority (e.g., testdata pair -> CategoryTestFixture, PriorityLow).
- **No mutation**: Pattern detection operates on `domain.CloneNode` trees (immutable copies), not `syntax.Node` (which carries mutable serialization state).
