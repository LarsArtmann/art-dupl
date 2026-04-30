# Architecture Review — art-dupl

**Date:** 2026-04-30
**Review Type:** Deep Architecture Review — Module Depth, Coupling, Locality
**Method:** Skill-guided (improve-codebase-architecture), deletion-test-driven

---

## Scorecard

| Package             | Depth      | Coupling         | Blast Radius     | Test Surface | Grade  |
| ------------------- | ---------- | ---------------- | ---------------- | ------------ | ------ |
| `suffixtree`        | ⬛⬛⬛⬛⬛ | ⬜ Minimal       | ⬜ Isolated      | ⬛⬛⬛⬛⬛   | **A**  |
| `syntax/golang`     | ⬛⬛⬛⬛   | ⬜ Minimal       | 🟡 Moderate      | ⬛⬛⬛⬛     | **A-** |
| `syntax/templ`      | ⬛⬛⬛     | ⬜ Minimal       | ⬜ Low           | ⬛⬛⬛⬛     | **A**  |
| `errors`            | ⬛⬛⬛⬛   | ⬜ Leaf          | ⬜ None          | ⬛⬛⬛⬛⬛   | **A**  |
| `domain` (core)     | ⬛⬛⬛⬛⬛ | ⬜ Minimal       | 🟡 Moderate      | ⬛⬛⬛⬛⬛   | **A**  |
| `cache`             | ⬛⬛⬛⬛   | 🟡 Low           | ⬜ Low           | ⬛⬛⬛⬛     | **A-** |
| `job`               | ⬛⬛⬛     | 🟢 Low           | 🟡 Moderate      | ⬛⬛⬛⬛     | **B+** |
| `config`            | ⬛⬛⬛     | 🟢 Low           | 🟡 Moderate      | ⬛⬛⬛⬛     | **B+** |
| `syntax`            | ⬛⬛⬛⬛   | 🟡 Moderate      | 🔴 **HIGH**      | ⬛⬛⬛⬛     | **B+** |
| `printer`           | ⬛⬛⬛⬛   | 🟡 Moderate      | 🟡 Moderate      | ⬛⬛⬛⬛     | **B+** |
| `hash`              | 🟡 Mixed   | 🟡 Moderate      | ⬜ Low           | ⬛⬛⬛       | **B**  |
| `pkg/artdupl`       | ⬛⬛⬛⬛   | 🔴 HIGH          | 🟡 Moderate      | ⬛⬛⬛⬛     | **B-** |
| `domain/conversion` | ⬛⬛⬛⬛   | 🔴 HIGH          | 🟡 Moderate      | ⬛⬛⬛       | **C+** |
| `cmd` (aggregate)   | ⬛⬛       | 🔴 **VERY HIGH** | 🔴 **VERY HIGH** | ⬛⬛⬛       | **C**  |

---

## Glossary

These terms are used exactly as defined in the skill vocabulary:

- **Module** — anything with an interface and an implementation (function, class, package, slice)
- **Interface** — everything a caller must know to use the module: types, invariants, error modes, ordering, config
- **Depth** — leverage at the interface: a lot of behaviour behind a small interface
- **Seam** — where an interface lives; a place behaviour can be altered without editing in place
- **Adapter** — a concrete thing satisfying an interface at a seam
- **Leverage** — what callers get from depth
- **Locality** — what maintainers get from depth: change, bugs, knowledge concentrated in one place

---

## Dependency Flow

```
                    cmd/art-dupl/main.go
                         │
                         ▼
                       cmd/
              ┌────┬────┼────────────┬──────────┐
              ▼    ▼    ▼            ▼          ▼
           config  printer  detection  pkg/artdupl  errors
              │       │       │            │
              ▼       ▼       ▼            ▼
           domain  syntax  suffixtree    job
                    │       │            │
                    │       ▼            ▼
                    │    hash         syntax
                    │
              syntax/golang
              syntax/templ

 errors/  ←──  leaf (zero internal deps)
 suffixtree/ ←── leaf (zero internal deps)
 hash/    ←── leaf (zero internal deps, except syntax via FileDetector)
 domain/  ←── near-leaf (only errors)
 syntax/  ←── keystone (15+ files depend on Node)
```

No circular dependencies detected. The DAG is clean, but the **fan-in on `syntax.Node`** is the dominant structural risk.

---

## Deepening Opportunities

### 1. Duplicated Pipeline — `cmd/run_analysis.go` and `pkg/artdupl/detector_pipeline.go`

**Files:** `cmd/run_analysis.go`, `pkg/artdupl/detector_pipeline.go`, `pkg/artdupl/detector_conversion.go`

**Problem:** Two independent implementations of the same analysis pipeline: parse files → build suffix tree → run detection → collect matches → convert to clones. `cmd/` builds the pipeline with 7+ internal imports; `pkg/artdupl/` duplicates it with 6+ imports. Neither calls the other. Changing the detection pipeline requires editing both files in lockstep.

Specific duplications:

- `buildAnalysisPipeline` (pipeline L15–87) duplicates `buildSuffixTree`/`buildSuffixTreeStandard` from cmd: same `job.Parse` → `job.BuildTree` → `tree.Update(&syntax.Node{Type: -1})` flow
- `runDetection` (pipeline L90–128) duplicates `executeAnalysis`'s detection dispatch: same `FindDuplOver` pattern
- `runSuffixTreeDetection` (pipeline L173–194) duplicates cmd's use of `suffixtree` + `syntax.FindSyntaxUnits`
- `collectMatchesIntoGroups` (pipeline L218–237) duplicates `printer.BuildCloneGroups` (printer/groups.go L20–27)

**Solution:** Extract the pipeline into `detection/` (which already coordinates multi-method detection). Make `detection/` own the full pipeline: `RunAnalysis(ctx, config, files) → []CloneGroup`. Both `cmd/` and `pkg/artdupl/` become thin callers over this single implementation.

**Benefits:** **Locality** — pipeline changes in one place, not two. **Leverage** — `cmd/` and `pkg/artdupl/` both get simpler interfaces. **Test surface** — one pipeline to test, not two.

---

### 2. Decorative Domain — `domain/` types unused by the pipeline

**Files:** `domain/clone.go`, `domain/conversion.go`, `domain/types_*.go`, `pkg/artdupl/types.go`, `pkg/artdupl/detector_conversion.go`

**Problem:** Three parallel type systems for the same concept:

| Concept    | `domain/`                                    | Pipeline                      | SDK (`artdupl`)               |
| ---------- | -------------------------------------------- | ----------------------------- | ----------------------------- |
| Clone      | `domain.Clone` (LineNumber, StringID)        | `[][]*syntax.Node`            | `artdupl.Clone` (int, string) |
| CloneGroup | `domain.CloneGroup` (CloneGroupID, Severity) | `map[string][][]*syntax.Node` | `artdupl.CloneGroup`          |
| Filepath   | `domain.Filepath` (validated)                | `string`                      | `string`                      |

The `domain/` package defines types the pipeline never constructs. `domain/conversion.go` bridges `syntax.Node → domain.Clone`, but nobody calls `NodeToClone` from the pipeline — `pkg/artdupl/detector_conversion.go` has its own `convertFragmentToClone` that bypasses domain entirely. The deletion test reveals: deleting `domain/` would **not** change any pipeline behavior.

**Solution:** Make the detection→printer seam use `domain.Clone`/`domain.CloneGroup`. Move `domain/conversion.go` to an adapter (it breaks domain purity by importing `syntax`). Delete `artdupl.Clone`/`artdupl.CloneGroup` in favor of `domain.Clone`/`domain.CloneGroup`. The printer interface changes from `[][]*syntax.Node` to `[]CloneGroup`.

**Benefits:** **Locality** — one type system for "a code clone." **Leverage** — validation, severity calculation, and hashing are free for every consumer. **Test surface** — test `Clone.IsValid()` once, trust it everywhere.

---

### 3. Keystone Type — `syntax.Node` coupled to everything

**Files:** `syntax/syntax.go`, `printer/printer.go`, `printer/groups.go`, `cmd/run_output.go`, `cmd/run_analysis.go`, `detection/multidetector.go`

**Problem:** `syntax.Node` is the de facto shared vocabulary — used by 15+ files across 8 packages. The `Printer` interface signature `PrintClones(dups [][]*syntax.Node)` forces every printer implementation to understand AST node internals. `cmd/run_output.go` constructs `map[string][][]*syntax.Node` and passes raw node pointers to printers. This makes it impossible to test printers without AST nodes, and impossible to add output formats without depending on `syntax`.

Specific couplings:

- `printer/printer.go:18` — `PrintClones(dups [][]*syntax.Node, sortBy ...SortBy) error`
- `printer/groups.go:11` — `GetCloneSize(group [][]*syntax.Node)` accesses `group[0][0].Owns`
- `printer/groups.go:20` — `BuildCloneGroups(duplChan <-chan syntax.Match)` returns `map[string][][]*syntax.Node`
- `cmd/run_analysis.go:257` — creates `chan syntax.Match`
- `cmd/run_analysis.go:316–318` — directly constructs `syntax.Match{Hash: ..., Frags: ...}`
- `detection/multidetector.go` — returns `<-chan syntax.Match`

**Solution:** Introduce a **printer DTO** (e.g., `printer.CloneData` or use `domain.CloneGroup` from #2) that carries filename, line range, fragment content, and hash. The conversion from `syntax.Node → DTO` happens once in `cmd/` or `detection/`, not in every printer.

**Benefits:** **Leverage** — new output formats need zero knowledge of AST internals. **Locality** — AST representation changes don't cascade to printer implementations. **Test surface** — printers testable with plain DTOs, no AST fixtures needed.

---

### 4. `cmd/run_analysis.go` — God file with 6 responsibilities

**Files:** `cmd/run_analysis.go` (7 internal imports, ~450 lines, 6 distinct concerns)

**Problem:** This single file does: status printing, filter construction, suffix tree building, analysis dispatch, hash-only analysis, and printer factory. It directly constructs `syntax.Node{Type: -1}` (sentinel nodes) and `syntax.Match{}` (detection results) — it knows too much about the AST layer. The file is the #1 coupling hotspot in the entire codebase.

Specific responsibilities:

1. **Status printing** — `printSearchStatus`, `printBuildingStatus`, `verboseFprintf`, `printFileCollectionStatus`
2. **Filter construction** — `setupFilter`
3. **Pipeline orchestration / suffix tree building** — `buildSuffixTree`, `buildSuffixTreeIncremental`, `buildSuffixTreeStandard`
4. **Analysis execution / detection dispatch** — `executeAnalysis`
5. **Hash-only analysis** — `collectFilesFromChannel`, `convertFileDuplicatesToMatches`, `createFragmentsFromFileHashes`, `executeHashOnlyAnalysis`
6. **Printer factory** — `withThreshold`, `createPrinter`

**Solution:** After extracting the pipeline to `detection/` (#1), this file becomes a thin CLI adapter: parse flags → call `detection.RunAnalysis()` → pass results to printer. The 6 responsibilities collapse to 1: "wire CLI flags to the analysis module."

**Benefits:** **Locality** — each concern moves to the package that owns it. **Leverage** — `cmd/` becomes a ~100-line adapter. **Test surface** — CLI tests focus on flag wiring, not detection logic.

---

### 5. Domain purity breach — `domain/conversion.go` imports `syntax`

**Files:** `domain/conversion.go` (imports `syntax`, `pkg/format`, `pkg/position`)

**Problem:** The domain layer directly depends on the `syntax` AST package. `NodeToClone` accepts `*syntax.Node` and traverses `node.Children`, `node.Pos`, `node.End`, `node.Type`. This makes the domain package impossible to use without the entire AST infrastructure. It's the only file in `domain/` that breaks the "domain knows nothing about infrastructure" rule.

Specific imports:

- L5: `pkg/format`
- L6: `pkg/position`
- L7: `syntax`

**Solution:** Move `NodeToClone` to a dedicated bridge package (e.g., `adapter/clone_converter.go` or `syntax/convert.go`). The domain package defines `Clone`/`CloneGroup`; the adapter knows both `syntax.Node` and `domain.Clone` and converts between them.

**Benefits:** **Locality** — domain stays pure; AST changes don't ripple through domain types. **Leverage** — domain types reusable without AST dependency. **Test surface** — domain tests need no AST fixtures; adapter tests are the only ones that need both.

---

### 6. Printer leaks to `syntax/golang` — `printer/clone_classify.go`

**Files:** `printer/clone_classify.go` (imports `syntax/golang`)

**Problem:** The printer layer imports the Go-language-specific parser for node type constants (`golang.FuncDecl`, `golang.StructType`, etc.). This makes the entire `printer` package Go-specific — you can't print results from `syntax/templ/` detection through the classification system. The node type constants live in `syntax/golang/`, not in `syntax/`, so the printer is reaching across a language seam.

Specific coupling:

- L7: `"github.com/LarsArtmann/art-dupl/syntax/golang"`
- `nodeTypeNames` map (L64–113) references `golang.FuncDecl`, `golang.StructType`, etc.
- `nodeTypeToCategory` function (L134–155) references Go-specific AST node types

**Solution:** Define a language-agnostic `NodeType` enum in `syntax/` (e.g., `NodeTypeFuncDecl`, `NodeTypeStructType`) that both `syntax/golang/` and `syntax/templ/` map to. The printer classifies using `syntax.NodeType` values, not `golang.*` constants.

**Benefits:** **Leverage** — printer works with any language parser. **Locality** — adding a new language doesn't require touching `printer/`. **Test surface** — classification testable with `syntax.NodeType` constants, no Go AST knowledge needed.

---

### 7. Dead modules — `cli/runtime.go` and `hash/detector.go`

**Files:** `cli/runtime.go`, `hash/detector.go`

**Problem (deletion test):**

- `cli/runtime.go`: `RuntimeConfig.ToConfig()` is never called by `cmd/`. The CLI builds config via `cmd/config_builder.go`. Deleting `cli/runtime.go` removes 0 complexity from the pipeline — it's a pass-through that nobody passes through.
- `hash/detector.go`: `HashDetector` embeds `FileDetector` and delegates `FindDuplOver` with zero added behavior (L57–58: pure delegation). Deleting it and using `FileDetector` directly removes 0 complexity.

**Solution:** Delete both. Replace `hash.NewHashDetector` calls with `hash.NewFileDetector` (trivial find-replace). Remove `cli/runtime.go` and its test.

**Benefits:** **Leverage** — fewer modules to understand. No functional loss.

---

### 8. Printer interface bypass — `cmd/run_output.go` type-asserts `*printer.JSONPrinter`

**Files:** `cmd/run_output.go:116`

**Problem:** `handleJSONOutput` does `p.(*printer.JSONPrinter)` — a concrete type assertion that breaks the `Printer` interface abstraction. Any alternative JSON printer implementation would silently be ignored. This is a seam with only one adapter, but the type assertion makes it impossible to add a second.

Specific violations:

- L116: `jsonPrinter, ok := p.(*printer.JSONPrinter)` — breaks interface abstraction
- L95–97: `printer.HashSetter` interface assertion — optional interface not part of `Printer`

**Solution:** Either: (a) add `OutputJSON()` to the `Printer` interface if every printer needs it, or (b) move JSON-specific output into `printer.JSONPrinter.PrintClones` itself so `cmd/` never needs to type-assert.

**Benefits:** **Leverage** — the `Printer` interface becomes a real seam with 2+ adapters. **Locality** — JSON output logic stays in the JSON printer, not scattered across `cmd/`.

---

## Additional Observations

### Type Alias Coupling

Several packages create unnecessary import edges through type aliases:

- `printer.Format = config.OutputFormat` — printer depends on config just for a type alias
- `artdupl.DetectionMethod = config.DetectionMethod` — SDK re-exports config's enum
- `artdupl.Logger = logger.Logger` — SDK re-exports logger interface

These should be their own types with conversion functions, not aliases that create import edges.

### `cmd/config_builder.go` Uses `panic` for Control Flow

Lines 308 and 325 use `panic(err)` for invalid timeout/diff-mode flags instead of returning errors. This is a control flow hazard in library code.

### SIMD Path Unimplemented

`syntax/hash_simd.go` — `hashSeqSIMD` falls back to `hashSeqFallback` every time. The `simd.Available()` check and dispatch add overhead for no current benefit.

### `StringID` Global Pool is a Hidden Singleton

`GlobalPool()` is used implicitly in `Clone` setters/getters. Tests must call `SetGlobalPoolForTesting()` or risk shared-state pollution. This is a global singleton pattern that can cause test isolation issues.

### Channel Conversion Inefficiency in `runAllModes`

`cmd/run_all_modes.go` converts channel → slice → channel for each output format, defeating the purpose of streaming and adding memory overhead.

---

## Priority Order

| #   | Candidate                   | Impact    | Effort                   | ROI          |
| --- | --------------------------- | --------- | ------------------------ | ------------ |
| 1   | Duplicated pipeline         | 🔴 High   | Medium                   | **Highest**  |
| 2   | Decorative domain           | 🔴 High   | Large                    | High         |
| 3   | Keystone `syntax.Node`      | 🟡 Medium | Medium                   | High         |
| 4   | God file `cmd/run_analysis` | 🔴 High   | Medium (unblocked by #1) | High         |
| 5   | Domain purity breach        | 🟡 Medium | Small                    | High         |
| 6   | Printer → golang coupling   | 🟡 Medium | Small                    | Medium       |
| 7   | Dead modules                | 🟢 Low    | Trivial                  | **Easy win** |
| 8   | Printer interface bypass    | 🟡 Medium | Small                    | Medium       |

**#7 (dead modules) is the easy win.** #1 + #4 are the highest impact — they eliminate the core duplication and god-file problem together. #2 + #3 are the deep structural fix but require the most work. #5 and #6 are quick hygiene.

---

## Existing ADRs

- **ADR-0001**: Map-Based Transition Lookup in Suffix Tree — well-documented decision with benchmarks and alternatives. No conflicts with any candidate above.

## Previous Architecture Review

- **docs/ARCHITECTURE_REVIEW.md** (2026-02-11): Most file-splitting recommendations have been acted on. The remaining structural issues (dual domain model, pipeline duplication, printer coupling) persist and are covered by the candidates above.

---

_Review conducted using the improve-codebase-architecture skill with deletion-test-driven analysis._
