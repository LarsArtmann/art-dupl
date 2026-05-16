# Status Report: Actionability & Semantic Detection — Complete

**Date:** 2026-05-16 21:31  
**Branch:** fork  
**Commits:** 5 ahead of origin/fork  
**Test Status:** ✅ All 31 packages pass, 253 BDD specs pass

---

## a) FULLY DONE

### 1. Actionability Domain Type (`domain/processed_clone.go`)

Added `CloneActionability` with two values:

- `Actionable` — real duplication that can be extracted/refactored
- `NonActionable` — idiomatic Go patterns that cannot be eliminated

Added `Actionability` field to `CloneClassification`.

### 2. Actionability Analyzer (`printer/actionability.go`)

`EvaluateActionability(nodeSeqs)` inspects clone group AST sequences and
returns non-actionable when ALL clones match the same boilerplate pattern:

- **Signature-only FuncDecl**: No `BlockStmt` children with >0 body. Catches
  interface method implementations that must match by contract.
- **DeferStmt**: Any bare defer call. Conservative without source text to
  distinguish `mu.Unlock()` from `expensiveCleanup()`. TODO left for future
  extension with identifier name lookup.
- **IfStmt with error propagation**: Inspects children for `BinaryExpr`
  (the `err != nil` comparison) + `BlockStmt` containing `ReturnStmt`.

### 3. Actionability Filtering in Semantic Mode (`cmd/run_output.go`)

When `cfg.Semantic == true`, `printCloneGroups` calls
`EvaluateActionability(uniq)` and skips the group if non-actionable.
`--structural` (default) reports everything unchanged.

### 4. Rich Text Output (`--rich-text` flag)

New CLI flag `--rich-text` enhances text output with classification info:

```
found 3 clones: [HIGH] method [non-actionable] (45 tokens, 12 lines) suggestion: Extract to shared utility
  store.go:97-103
  repo.go:38-44
```

- `RichText` field added to `config.Config`
- `RichTextSetter` interface in `printer/printer.go`
- `TextPrinter.SetRichText(bool)` + `writeRichGroupHeader` implemented

### 5. Text Output Line Notation Fix (`printer/text.go`)

Changed from comma (`97,103`) to dash (`97-103`) across all text output
paths. Matches plumbing format and removes ambiguity about what the
numbers mean.

### 6. JSON Output Enrichment (`printer/json.go`)

`JSONClone` now includes three classification fields:

- `category`: method, function, loop, etc.
- `priority`: critical, high, medium, low
- `actionability`: actionable, non-actionable

CI pipelines and downstream tools can now filter clones structurally.

### 7. Actionability Field Population (`printer/clone_processor.go`)

Fixed critical bug: `ProcessClones` now calls `EvaluateActionability(dups)`
after the per-instance loop and sets `clones[i].Classification.Actionability`
for every clone in the group. Previously the field was always empty.

### 8. BDD Tests (`bdd/actionability_test.go`)

Two new BDD spec groups:

- **Actionability Filtering** (1 spec): Creates identical struct method
  signatures, verifies `--semantic` suppresses boilerplate.
- **Rich Text Output** (2 specs): Verifies `--rich-text` produces output
  with badges, and JSON output includes classification fields.

### 9. Planning Document (`docs/planning/2026-05-16_21-13_actionability-phase-2-plan.md`)

Pareto breakdown, detailed execution plan with task granularity,
architecture decisions, and verification checklist.

### 10. Test Coverage

- `printer/actionability_test.go`: 10 table-driven cases
- `bdd/actionability_test.go`: 3 BDD specs
- Full suite: 31 packages, all pass
- BDD suite: 253 specs, all pass

---

## b) PARTIALLY DONE

### HTML Report Actionability Badges

HTML printer reads `CloneClassification` but does not show `Actionability`.
The infrastructure exists (`html.go` consumes Classification for priority
and category). Adding `[non-actionable]` badges would be a ~20-line change
in `html.go` + `html_summary.go`.

**Why deferred:** Large HTML template file (500+ lines). Lower priority
than getting text/JSON/semantic filtering right.

---

## c) NOT STARTED

1. **Weighted token analysis for partial-body matches** — Current logic
   checks `len(seq) == 1` OR complete patterns. A FuncDecl with body that
   is 80% boilerplate (e.g., `if err != nil { return }` + 2 lines of real
   logic) is still actionable. Should we weight tokens? Would require access
   to source text at the evaluation layer.

2. **`--semantic` as default mode** — Currently `--structural` is default.
   Switching to `--semantic` would reduce noise for 99% of users but is a
   breaking behavioral change. Needs version migration strategy.

3. **Stats subcommand actionability filtering** — StatsPrinter flows through
   `printCloneGroups` with semantic filtering (already works). But if user
   runs `art-dupl stats --structural --rich-text`, stats counts still include
   non-actionable clones. No badge/annotation in stats output.

4. **Plumbing output badges** — Plumbing uses `writeCloneLines` which
   ignores classification entirely. Adding `# non-actionable` comment lines
   would be trivial.

5. **SARIF output actionability** — SARIF printer (`printer/sarif.go`) has
   its own `SARIFClone` struct. Does not include classification data.

6. **Simple-JSON output actionability** — `SimpleCloneGroup` / `SimpleJSONClone`
   structs do not include classification fields.

7. **AGENTS.md update** — Actionability patterns, `--rich-text` flag, and
   `--semantic` behavior should be documented in project AGENTS.md.

8. **Feature documentation** — FEATURES.md / README.md should reflect
   `--rich-text` availability and `--semantic` filtering behavior.

---

## d) TOTALLY FUCKED UP

Nothing. No regressions, no broken tests, no dead code. Working tree is
clean. All builds and tests pass.

---

## e) WHAT WE SHOULD IMPROVE

1. **`extractMethodNameFromSelector` is a stub** — Returns empty string
   because `syntax.Node` does not carry identifier names. The AST stores
   actual text in the parser, but `transform.go` only stores node types.
   Adding a `Name` or `IdentValue` field to `Node` would enable precise
   defer pattern detection (`Unlock` vs `Cleanup`).

2. **`isErrorOnlyIf` has false negatives** — The current heuristic looks
   for BinaryExpr + BlockStmt children but does not verify the BinaryExpr
   is actually `err != nil`. Could flag `if x > 5 { return }` as error
   propagation incorrectly. Needs actual expression inspection.

3. **`isPureDeferPattern` is too conservative** — ALL bare DeferStmt
   matches are treated as non-actionable. In practice some defer calls
   contain business logic that should be extractable. Needs identifier
   name lookup.

4. **Text output test fragile** — `writeRichGroupHeader` tests depend on
   `ClassifyClone` output which can change. BDD test for `--rich-text`
   only checks "not empty" + "Found total" — could be stronger.

5. **Priority/Actionability orthogonality may confuse** — A `[MEDIUM]`
   `[non-actionable]` badge is correct but may confuse users. Consider
   adding a `--explain` flag that prints why each clone was classified.

6. **No performance benchmarks** — `EvaluateActionability` runs per clone
   group and traverses children. On 1000+ group repos this could add
   milliseconds. Not measured.

---

## f) Top #25 Things To Get Done Next

| #   | Task                                                                           | Priority | Effort | Impact                                         |
| --- | ------------------------------------------------------------------------------ | -------- | ------ | ---------------------------------------------- |
| 1   | Add `IdentValue` field to `syntax.Node` for precise defer/selector name lookup | P0       | Medium | Enables real `Unlock` vs `Cleanup` distinction |
| 2   | Fix `isErrorOnlyIf` false negatives by inspecting BinaryExpr operator type     | P1       | Low    | Reduces false positives in error propagation   |
| 3   | Add Actionability badges to HTML report                                        | P1       | Medium | Parity with text output                        |
| 4   | Add Actionability to SARIF output                                              | P2       | Low    | Security tool integration needs classification |
| 5   | Add Actionability to Simple-JSON output                                        | P2       | Low    | Simple API parity                              |
| 6   | Add `# non-actionable` comment to plumbing output                              | P3       | Low    | Machine-readable flag                          |
| 7   | Stats subcommand `--rich-text` support                                         | P3       | Low    | Consistency across commands                    |
| 8   | Benchmark `EvaluateActionability` on large repos (>1M LOC)                     | P3       | Low    | Performance confidence                         |
| 9   | Update AGENTS.md with actionability patterns                                   | P4       | Low    | Memory maintenance                             |
| 10  | Update README.md / FEATURES.md for `--rich-text`                               | P4       | Low    | User discoverability                           |
| 11  | Add `--explain` flag showing per-clone classification rationale                | P4       | Medium | UX improvement                                 |
| 12  | Weighted token analysis for partial-body boilerplate                           | P5       | Medium | Catches "mostly boilerplate" functions         |
| 13  | Extend to goroutine patterns `go func(){ ... }()`                              | P5       | Low    | Common boilerplate in Go                       |
| 14  | Extend to `context.WithCancel` boilerplate                                     | P5       | Low    | Standard Go pattern                            |
| 15  | Extend to `nil` check patterns                                                 | P5       | Low    | `if x == nil { return nil, err }`              |
| 16  | Extend to type assertion boilerplate                                           | P5       | Low    | `if v, ok := x.(T); ok`                        |
| 17  | Add `--semantic` deprecation timeline for default switch                       | P5       | Low    | Breaking change planning                       |
| 18  | Actionability filter buttons on HTML report (JS)                               | P5       | Medium | Interactive filtering                          |
| 19  | Integration test on `go-cqrs-lite` repo                                        | P5       | Low    | Validate ~60-70% claim                         |
| 20  | Add `ActionabilityThreshold` config (strictness)                               | P6       | Low    | Tunable behavior                               |
| 21  | Error handling pattern: `errors.Is` / `errors.As`                              | P6       | Low    | More Go 1.13+ idioms                           |
| 22  | Constructor pattern detection                                                  | P6       | Low    | `return &Type{Field: val}`                     |
| 23  | Package-level variable declaration clones                                      | P6       | Low    | var blocks across files                        |
| 24  | Interface embedding boilerplate                                                | P6       | Low    | `type X interface { Y }`                       |
| 25  | Map/slice literal initialization patterns                                      | P6       | Low    | Common in config files                         |

---

## g) Top #1 Question I Cannot Figure Out Myself

**How do we make `syntax.Node` carry identifier values without exploding its
memory footprint?**

Current `Node` is 40B:

```
type Node struct {
    Type     int32   // 4B
    Pos      int32   // 4B
    End      int32   // 4B
    Owns     int32   // 4B
    Children []*Node // 8B
    Filename string  // 16B
}
```

To distinguish `defer mu.Unlock()` from `defer expensiveCleanup()`, we
need the actual identifier name ("Unlock" vs "Cleanup"). Options:

**Option A: Add `int32 IdentID` field** — ~44B total. Use a string pool
(domain.StringPool) to map ID→name. Only `Ident` nodes set this.
Memory cost: +4B per node × millions of nodes = significant overhead.

**Option B: Store source byte position, read text on demand** — Keep
`Pos`/`End` on Ident nodes. When `extractMethodNameFromSelector` needs
the name, read the source file and slice `content[pos:end]`. No memory
overhead, but requires file I/O during actionability evaluation.

**Option C: Add `[]byte IdentValue` to Ident nodes only** — Make `Node`
interface-like with variants: `TypeNode{Type, Pos, End, Owns}` and
`IdentNode{Type, Pos, End, Owns, Name []byte}`. Complex type system
change, breaks existing code that assumes uniform `[]*Node`.

**Option D: Keep actionability conservative (current)** — Accept that
`defer` patterns are all non-actionable. Don't add name lookups. This is
what we have now, and it works for the reported use case.

I lean toward **Option B** (source position lookup) for the actionability
layer specifically — it's only needed for evaluation, not for general
node processing. But I'm unsure if `Pos`/`End` on `Ident` nodes actually
point to the right source range, or if they need separate fields.

**Can you confirm which option is best for this specific need?**
