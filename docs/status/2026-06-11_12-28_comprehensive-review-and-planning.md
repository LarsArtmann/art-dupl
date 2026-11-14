# Status Report — 2026-06-11 12:28

> **Resolution (2026-10-05):** ~~Open items unresolved at write time.~~ Audited 2026-10-05 docs-health pass: 12 items verified resolved, 4 open (DTO/TokenValue/enum/CloneRef shipped; branded NodeType deferred; micro-tails speculative).

> **Resolution (2026-10-05):** ~~Open items unresolved at write time.~~ Audited 2026-10-05 docs-health pass: 12 items verified resolved, 4 open (DTO/TokenValue/enum/CloneRef shipped; branded NodeType deferred; micro-tails speculative).

## Session Summary

Extended the AST-aware false-positive filter system and completed several TODO items. 6 commits, all tests green, clone count stable at 68.

---

## a) FULLY DONE ✅

<!-- ledger note: historical session record: rows cite landed commit hashes -->
<!-- check-rows:ledger -->
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

<!-- ledger note: analysis snapshot; 2-file interface limitation tracked as Top-25 item 6 (open) -->
<!-- check-rows:ledger -->
| Item                    | Status                                                                                                                                                                                                                                       | What Remains                                                                                                                |
| ----------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- |
| Interface impl detector | Works for 3+ files with FuncType root. **Does NOT catch 2-file interface implementations** (e.g., PrintHeader, PrintFooter which have only 2 implementations each). The threshold of 3 files is correct for precision but misses some cases. | Could lower to 2 files with an additional signal (e.g., both in same package, both implement same interface). Low priority. |
| Clone count reduction   | 76 → 68 (8 groups eliminated across sessions). Still 68 groups, mostly test file boilerplate.                                                                                                                                                | See "Top 25" below for further reduction paths.                                                                             |

## c) NOT STARTED ⬜

From TODO_LIST.md, these items remain untouched:

<!-- ledger note: one open row: transform.go refactor still pending -->
<!-- check-rows:ledger -->
| Priority  | Item                                                                          |
| --------- | ----------------------------------------------------------------------------- |
| ~~🔴 HIGH~~ done — `domain/processed_clone.go` | ~~ProcessedClone DTO to decouple Printer from syntax.Node (111 test call sites)~~ |
| ~~🔴 HIGH~~ superseded — ADR-0005 split-brain type unification | ~~Consolidate three parallel Clone types~~ |
| ~~🔴 HIGH~~ done — `suffixtree/suffixtree.go` `TokenValue` | ~~TokenValue type with validation~~ |
| ~~🟡 MEDIUM~~ done — `printer/stats/stats_formatter.go` | ~~CSV output format using encoding/csv~~ |
| ~~🟡 MEDIUM~~ done — `pkg/enum/enum.go`, `config/enum_helpers.go` | ~~Unify enum patterns with config's generic helpers~~ |
| ~~🟡 MEDIUM~~ done — string interning in `syntax/intern.go` | ~~SIMD memory layouts + string interning~~ |
| ~~🟡 MEDIUM~~ done — `cmd/stats.go` | ~~--output-file flag for stats subcommand~~ |
| 🟢 LOW    | Refactor transform.go (369L, 300L switch)                                     |
| ~~🟢 LOW~~ done at c0261563 | ~~Fix remaining LSP hints~~ |
| ~~🟢 LOW~~ done — `SDK_DESIGN.md` | ~~SDK documentation for pkg/artdupl/~~ |
| ~~🟢 LOW~~ done — `bdd/output_formats_and_filters_test.go` (`--include-generic` superseded by `--include-generated`) | ~~BDD tests for --only templ/go, --include-generic~~ |
| ~~🟢 LOW~~ done — `syntax/templ/fuzz_test.go` | ~~Fuzz tests for templ parser~~ |
| ~~🟢 LOW~~ done — `.github/workflows/release.yml` | ~~Validate GoReleaser config~~ |

## d) TOTALLY FUCKED UP 💥

<!-- ledger note: split-brain analysis snapshot — superseded by docs/adr/0005 -->
<!-- check-rows:ledger -->
| Issue                                                             | Severity | Details                                                                                                                                                                                                                                            |
| ----------------------------------------------------------------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `printer/actionability.go` still imports `syntax/golang` directly | Medium   | `baseTypeOf()` calls `golang.DecodeBaseType()`. The actionability detectors (all 8 of them) use `golang.FuncType`, `golang.FuncDecl`, `golang.BlockStmt`, etc. This is a layer violation — printer should work through `syntax.Node` abstractions. |
| `printer/clone_processor.go` imports `syntax/golang`              | Medium   | Uses `golang.DecodeBaseType()` and `golang.Ident` constant.                                                                                                                                                                                        |
| `PatternLabel` defined in `printer/` not `domain/`                | Low      | Semantically belongs with `CloneCategory`, `ClonePriority`, `CloneActionability`.                                                                                                                                                                  |
| `detection/issue_helpers.go` raw string enums                     | Low      | `TodoIssue.Type`, `LegacyIssue.Type`, `LegacyPattern.Severity` all raw strings when `domain` typed enums exist.                                                                                                                                    |
| `NodeType` inconsistency                                          | Low      | `ClassificationInput.NodeType` is `int32`, `CloneClassification.NodeType` is `string` — same concept, two representations.                                                                                                                         |

## e) WHAT WE SHOULD IMPROVE

### Architecture

~~1. **Layer violation: `printer/` → `syntax/golang/`** — 4 production files bypass the `syntax/` abstraction. The actionability system checks AST node types (`golang.FuncType`, `golang.Ident`, etc.) but should work through a `syntax.NodeType` abstraction or the classification should live closer to the data source.~~ won't implement — deferred: import cycle blocks facade (TODO_LIST "Architecturally constrained (DEFERRED)")

~~2. **Three parallel Clone types** — `printer.clone` (internal), `pkg/artdupl.Clone` (public SDK), `domain.ProcessedClone` (DTO). Each has overlapping fields. The ProcessedClone DTO migration (111 test call sites) would unify them.~~ done — ADR-0005; canonical domain.ProcessedCloneGroup

3. **NodeType representation** — int32 in some places, string in others. Should be a single typed representation throughout the pipeline.

### Code Quality

~~4. **`syntax/golang/transform.go`** — 369 lines with a 300-line switch statement. Could use a registry/table-driven approach.~~ won't implement — skip documented: unique per-case logic, table harder to read

~~5. **3 pre-existing godoclint warnings** — Multiple `doc.go` files in same package with different package comments.~~ done — fixed; 0 lint issues since

### Testing

~~6. **No fuzz tests for templ parser** — The templ parser handles complex CSS/HTML/JS edge cases and would benefit from fuzzing.~~ done — syntax/templ/fuzz_test.go FuzzParseBytes

~~7. **Missing BDD tests** — `--only templ`, `--only go`, `--include-generic` flags lack end-to-end coverage.~~ done — BDD --only + --include-generic coverage

## f) Top 25 Things We Should Get Done Next

Sorted by **impact × effort⁻¹** (highest value first):

<!-- ledger note: open rows 6/15/22 are the point (harvest candidates) -->
<!-- check-rows:ledger -->
| #  | Item                                                                              | Impact                                                          | Effort                       | Type                 |
| -- | --------------------------------------------------------------------------------- | --------------------------------------------------------------- | ---------------------------- | -------------------- |
| ~~1~~ done — `printer/actionability/actionability.go` `PatternLabel` | ~~Move PatternLabel to domain package~~ | ~~Medium~~ | ~~Low (30 min)~~ | ~~Architecture~~ |
| ~~2~~ done at 8889a7a3 | ~~Fix 3 godoclint warnings (remove duplicate doc.go files)~~ | ~~Low~~ | ~~Low (15 min)~~ | ~~Lint~~ |
| ~~3~~ resolved — `TodoIssue`/`LegacyIssue` types removed in 64057ba7 | ~~Type TodoIssue.Type and LegacyIssue.Type as domain enums~~ | ~~Low~~ | ~~Low (30 min)~~ | ~~Type safety~~ |
| ~~4~~ done — `domain/processed_clone.go` `NodeType int32` | ~~Add NodeType typed int32 to domain, unify ClassificationInput/CloneClassification~~ | ~~Medium~~ | ~~Medium (1 hr)~~ | ~~Type safety~~ |
| ~~5~~ done — `printer/actionability/` package | ~~Extract actionability node-type checks into syntax/ abstraction~~ | ~~High~~ | ~~Medium (2 hr)~~ | ~~Architecture~~ |
| 6 | Lower interface impl detector threshold to 2 files + package signal | Low | Low (30 min) | Feature |
| ~~7~~ done — `.github/workflows/release.yml` | ~~Validate GoReleaser release config~~ | ~~Low~~ | ~~Low (30 min)~~ | ~~Ops~~ |
| ~~8~~ done — `bdd/output_formats_and_filters_test.go` | ~~Add BDD test for `--only templ` and `--only go`~~ | ~~Medium~~ | ~~Low (1 hr)~~ | ~~Testing~~ |
| ~~9~~ superseded — flag renamed `--include-generated` (`cmd/flags.go`) | ~~Add BDD test for `--include-generic` end-to-end~~ | ~~Medium~~ | ~~Low (1 hr)~~ | ~~Testing~~ |
| ~~10~~ done — `SDK_DESIGN.md` | ~~Write SDK documentation for `pkg/artdupl/`~~ | ~~Medium~~ | ~~Medium (2 hr)~~ | ~~Docs~~ |
| ~~11~~ done — `cmd/stats.go` | ~~Add `--output-file` flag to stats subcommand~~ | ~~Medium~~ | ~~Low (1 hr)~~ | ~~Feature~~ |
| ~~12~~ done — `printer/stats/stats_formatter.go` | ~~Implement CSV output format using encoding/csv~~ | ~~Medium~~ | ~~Medium (2 hr)~~ | ~~Feature~~ |
| ~~13~~ done — `pkg/enum/enum.go` + `config/enum_helpers.go` | ~~Unify enum patterns with config's generic helpers~~ | ~~Low~~ | ~~Medium (2 hr)~~ | ~~Architecture~~ |
| ~~14~~ done — `syntax/templ/fuzz_test.go` | ~~Add fuzz tests for templ parser edge cases~~ | ~~Medium~~ | ~~Medium (2 hr)~~ | ~~Testing~~ |
| 15 | Refactor `syntax/golang/transform.go` (table-driven) | Medium | Medium (3 hr) | Code quality |
| ~~16~~ done — `suffixtree/suffixtree.go` `TokenValue` | ~~Implement TokenValue type with validation~~ | ~~High~~ | ~~Medium (3 hr)~~ | ~~Architecture~~ |
| ~~17~~ done — `domain/processed_clone.go` | ~~ProcessedClone DTO — decouple Printer from syntax.Node~~ | ~~Very High~~ | ~~Very High (1-2 days)~~ | ~~Architecture~~ |
| ~~18~~ superseded — ADR-0005 split-brain type unification | ~~Consolidate three parallel Clone types~~ | ~~High~~ | ~~High (1 day, blocked on #17)~~ | ~~Architecture~~ |
| ~~19~~ done — string interning in `syntax/intern.go` | ~~Optimize memory layouts for SIMD + string interning~~ | ~~Medium~~ | ~~High (1-2 days)~~ | ~~Performance~~ |
| ~~20~~ done — `syntax/golang/typeinfo.go` | ~~Investigate using `go/types` for precise interface satisfaction~~ | ~~Medium~~ | ~~Medium (2 hr)~~ | ~~Feature~~ |
| ~~21~~ done — `scripts/dogfood-goroot.sh` | ~~Dogfood on external projects (test real-world precision)~~ | ~~High~~ | ~~Low (1 hr)~~ | ~~Validation~~ |
| 22 | Add `--min-files` flag to filter by minimum file count | Low | Low (30 min) | Feature |
| ~~23~~ done — `cmd/run_crawl.go:115` | ~~Fix `cmd/run_crawl.go` scanner.Err() unchecked~~ | ~~Low~~ | ~~Low (15 min)~~ | ~~Bug~~ |
| ~~24~~ done — `printer/stats/stats_formatter.go` actionability block | ~~Add actionability metrics to stats output~~ | ~~Medium~~ | ~~Medium (2 hr)~~ | ~~Feature~~ |
| ~~25~~ done — `printer/actionability/extractability_bench_test.go` | ~~Benchmark actionability filters on large codebases~~ | ~~Medium~~ | ~~Medium (2 hr)~~ | ~~Performance~~ |
## g) Top #1 Question I Cannot Figure Out Myself

**How should the actionability system work for non-Go languages (templ)?**

Currently all 8 actionability detectors (`isInterfaceImplementation`, `isSignatureOnlyMatch`, `isPureDeferPattern`, etc.) use `golang.DecodeBaseType()` and `golang.*` constants. They only work on Go AST nodes. Templ files produce different node types.

The architecture question is:

- Should actionability detectors work on **language-agnostic** `syntax.Node` patterns (base types only)?
- Or should we have **per-language actionability registries** where Go has its detectors and templ has its own?
- The current `baseTypeOf()` function already decodes to language-specific constants (Go: FuncDecl=19, Templ: different values). A language-agnostic approach would need a `syntax.NodeCategory` enum that abstracts over language differences.

This is an architectural decision that affects the ProcessedClone DTO design. If we move forward with the DTO, we need to decide whether actionability is computed before or after the DTO conversion.

## Metrics

<!-- ledger note: session metrics snapshot, not action items -->
<!-- check-rows:ledger -->
| Metric                                 | Value                               |
| -------------------------------------- | ----------------------------------- |
| Clone groups (threshold 15, semantic)  | 68                                  |
| Test packages                          | 25 (all passing)                    |
| Lint issues                            | 3 (pre-existing godoclint)          |
| Packages with syntax/golang violations | 3 (printer, job, internal/testutil) |
| Commits this session                   | 6                                   |
| Net lines added                        | ~150                                |
