# Status Report — 2026-06-11 12:28

## Session Summary

Extended the AST-aware false-positive filter system and completed several TODO items. 6 commits, all tests green, clone count stable at 68.

---

## a) FULLY DONE ✅

| Item                                              | Commit    | Impact                                                            |
| ------------------------------------------------- | --------- | ----------------------------------------------------------------- |
| FuncType-based interface impl detector            | `49ae11a` | Eliminates 6-way PrintClones clone group (non-actionable)         |
| HealthScore typed enum (`domain.HealthScore`)     | `b4ea5f8` | Type safety for A-F grades with JSON marshaling                   |
| Decouple `clone_classify.go` from `syntax/golang` | `60f402d` | Moved 56-line `nodeTypeNames` map to `syntax/golang/nodetypes.go` |
| Reduce `GetCategoryEmoji` complexity 16→2         | `2583add` | Fixed cyclop lint violation                                       |
| Complete exhaustive switch in `applyPatternLabel` | `49ae11a` | Added all missing PatternLabel cases                              |
| Test cases for `isInterfaceImplementation`        | `49ae11a` | 7 test cases covering edge cases                                  |
| Lint fixes (gci, wsl)                             | `d2952f9` | Clean lint (only 3 pre-existing godoclint remain)                 |
| TODO_LIST.md updated                              | `76d1db4` | Marked completed items, updated date                              |

## b) PARTIALLY DONE 🔶

| Item                    | Status                                                                                                                                                                                                                                       | What Remains                                                                                                                |
| ----------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| Interface impl detector | Works for 3+ files with FuncType root. **Does NOT catch 2-file interface implementations** (e.g., PrintHeader, PrintFooter which have only 2 implementations each). The threshold of 3 files is correct for precision but misses some cases. | Could lower to 2 files with an additional signal (e.g., both in same package, both implement same interface). Low priority. |
| Clone count reduction   | 76 → 68 (8 groups eliminated across sessions). Still 68 groups, mostly test file boilerplate.                                                                                                                                                | See "Top 25" below for further reduction paths.                                                                             |

## c) NOT STARTED ⬜

From TODO_LIST.md, these items remain untouched:

| Priority  | Item                                                                          |
| --------- | ----------------------------------------------------------------------------- |
| 🔴 HIGH   | ProcessedClone DTO to decouple Printer from syntax.Node (111 test call sites) |
| 🔴 HIGH   | Consolidate three parallel Clone types                                        |
| 🔴 HIGH   | TokenValue type with validation                                               |
| 🟡 MEDIUM | CSV output format using encoding/csv                                          |
| 🟡 MEDIUM | Unify enum patterns with config's generic helpers                             |
| 🟡 MEDIUM | SIMD memory layouts + string interning                                        |
| 🟡 MEDIUM | --output-file flag for stats subcommand                                       |
| 🟢 LOW    | Refactor transform.go (369L, 300L switch)                                     |
| 🟢 LOW    | Fix remaining LSP hints                                                       |
| 🟢 LOW    | SDK documentation for pkg/artdupl/                                            |
| 🟢 LOW    | BDD tests for --only templ/go, --include-generic                              |
| 🟢 LOW    | Fuzz tests for templ parser                                                   |
| 🟢 LOW    | Validate GoReleaser config                                                    |

## d) TOTALLY FUCKED UP 💥

| Issue                                                             | Severity | Details                                                                                                                                                                                                                                            |
| ----------------------------------------------------------------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `printer/actionability.go` still imports `syntax/golang` directly | Medium   | `baseTypeOf()` calls `golang.DecodeBaseType()`. The actionability detectors (all 8 of them) use `golang.FuncType`, `golang.FuncDecl`, `golang.BlockStmt`, etc. This is a layer violation — printer should work through `syntax.Node` abstractions. |
| `printer/clone_processor.go` imports `syntax/golang`              | Medium   | Uses `golang.DecodeBaseType()` and `golang.Ident` constant.                                                                                                                                                                                        |
| `PatternLabel` defined in `printer/` not `domain/`                | Low      | Semantically belongs with `CloneCategory`, `ClonePriority`, `CloneActionability`.                                                                                                                                                                  |
| `detection/issue_helpers.go` raw string enums                     | Low      | `TodoIssue.Type`, `LegacyIssue.Type`, `LegacyPattern.Severity` all raw strings when `domain` typed enums exist.                                                                                                                                    |
| `NodeType` inconsistency                                          | Low      | `ClassificationInput.NodeType` is `int32`, `CloneClassification.NodeType` is `string` — same concept, two representations.                                                                                                                         |

## e) WHAT WE SHOULD IMPROVE

### Architecture

1. **Layer violation: `printer/` → `syntax/golang/`** — 4 production files bypass the `syntax/` abstraction. The actionability system checks AST node types (`golang.FuncType`, `golang.Ident`, etc.) but should work through a `syntax.NodeType` abstraction or the classification should live closer to the data source.

2. **Three parallel Clone types** — `printer.clone` (internal), `pkg/artdupl.Clone` (public SDK), `domain.ProcessedClone` (DTO). Each has overlapping fields. The ProcessedClone DTO migration (111 test call sites) would unify them.

3. **NodeType representation** — int32 in some places, string in others. Should be a single typed representation throughout the pipeline.

### Code Quality

4. **`syntax/golang/transform.go`** — 369 lines with a 300-line switch statement. Could use a registry/table-driven approach.

5. **3 pre-existing godoclint warnings** — Multiple `doc.go` files in same package with different package comments.

### Testing

6. **No fuzz tests for templ parser** — The templ parser handles complex CSS/HTML/JS edge cases and would benefit from fuzzing.

7. **Missing BDD tests** — `--only templ`, `--only go`, `--include-generic` flags lack end-to-end coverage.

## f) Top 25 Things We Should Get Done Next

Sorted by **impact × effort⁻¹** (highest value first):

| #   | Item                                                                              | Impact    | Effort                       | Type         |
| --- | --------------------------------------------------------------------------------- | --------- | ---------------------------- | ------------ |
| 1   | Move PatternLabel to domain package                                               | Medium    | Low (30 min)                 | Architecture |
| 2   | Fix 3 godoclint warnings (remove duplicate doc.go files)                          | Low       | Low (15 min)                 | Lint         |
| 3   | Type TodoIssue.Type and LegacyIssue.Type as domain enums                          | Low       | Low (30 min)                 | Type safety  |
| 4   | Add NodeType typed int32 to domain, unify ClassificationInput/CloneClassification | Medium    | Medium (1 hr)                | Type safety  |
| 5   | Extract actionability node-type checks into syntax/ abstraction                   | High      | Medium (2 hr)                | Architecture |
| 6   | Lower interface impl detector threshold to 2 files + package signal               | Low       | Low (30 min)                 | Feature      |
| 7   | Validate GoReleaser release config                                                | Low       | Low (30 min)                 | Ops          |
| 8   | Add BDD test for `--only templ` and `--only go`                                   | Medium    | Low (1 hr)                   | Testing      |
| 9   | Add BDD test for `--include-generic` end-to-end                                   | Medium    | Low (1 hr)                   | Testing      |
| 10  | Write SDK documentation for `pkg/artdupl/`                                        | Medium    | Medium (2 hr)                | Docs         |
| 11  | Add `--output-file` flag to stats subcommand                                      | Medium    | Low (1 hr)                   | Feature      |
| 12  | Implement CSV output format using encoding/csv                                    | Medium    | Medium (2 hr)                | Feature      |
| 13  | Unify enum patterns with config's generic helpers                                 | Low       | Medium (2 hr)                | Architecture |
| 14  | Add fuzz tests for templ parser edge cases                                        | Medium    | Medium (2 hr)                | Testing      |
| 15  | Refactor `syntax/golang/transform.go` (table-driven)                              | Medium    | Medium (3 hr)                | Code quality |
| 16  | Implement TokenValue type with validation                                         | High      | Medium (3 hr)                | Architecture |
| 17  | ProcessedClone DTO — decouple Printer from syntax.Node                            | Very High | Very High (1-2 days)         | Architecture |
| 18  | Consolidate three parallel Clone types                                            | High      | High (1 day, blocked on #17) | Architecture |
| 19  | Optimize memory layouts for SIMD + string interning                               | Medium    | High (1-2 days)              | Performance  |
| 20  | Investigate using `go/types` for precise interface satisfaction                   | Medium    | Medium (2 hr)                | Feature      |
| 21  | Dogfood on external projects (test real-world precision)                          | High      | Low (1 hr)                   | Validation   |
| 22  | Add `--min-files` flag to filter by minimum file count                            | Low       | Low (30 min)                 | Feature      |
| 23  | Fix `cmd/run_crawl.go` scanner.Err() unchecked                                    | Low       | Low (15 min)                 | Bug          |
| 24  | Add actionability metrics to stats output                                         | Medium    | Medium (2 hr)                | Feature      |
| 25  | Benchmark actionability filters on large codebases                                | Medium    | Medium (2 hr)                | Performance  |

## g) Top #1 Question I Cannot Figure Out Myself

**How should the actionability system work for non-Go languages (templ)?**

Currently all 8 actionability detectors (`isInterfaceImplementation`, `isSignatureOnlyMatch`, `isPureDeferPattern`, etc.) use `golang.DecodeBaseType()` and `golang.*` constants. They only work on Go AST nodes. Templ files produce different node types.

The architecture question is:

- Should actionability detectors work on **language-agnostic** `syntax.Node` patterns (base types only)?
- Or should we have **per-language actionability registries** where Go has its detectors and templ has its own?
- The current `baseTypeOf()` function already decodes to language-specific constants (Go: FuncDecl=19, Templ: different values). A language-agnostic approach would need a `syntax.NodeCategory` enum that abstracts over language differences.

This is an architectural decision that affects the ProcessedClone DTO design. If we move forward with the DTO, we need to decide whether actionability is computed before or after the DTO conversion.

## Metrics

| Metric                                 | Value                               |
| -------------------------------------- | ----------------------------------- |
| Clone groups (threshold 15, semantic)  | 68                                  |
| Test packages                          | 25 (all passing)                    |
| Lint issues                            | 3 (pre-existing godoclint)          |
| Packages with syntax/golang violations | 3 (printer, job, internal/testutil) |
| Commits this session                   | 6                                   |
| Net lines added                        | ~150                                |
