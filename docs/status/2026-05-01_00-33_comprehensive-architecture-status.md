# Comprehensive Status Report — art-dupl Codebase

**Date:** 2026-05-01 00:33\
**Branch:** `fork` (only branch, 1188 total commits)\
**Working tree:** Clean\
**Reporter:** Crush (GLM-5.1)

---

## Quality Metrics

| Metric                         | Value                                                     | Status           |
| ------------------------------ | --------------------------------------------------------- | ---------------- |
| Linter issues                  | **0**                                                     | Clean            |
| Test packages                  | **23/23 pass** (0 FAIL)                                   | Clean            |
| BDD tests                      | **Pass**                                                  | Clean            |
| Total coverage                 | **73.2%**                                                 | Below 80% target |
| Total Go lines (excl vendor)   | **45,481**                                                | —                |
| Production nolint suppressions | **81**                                                    | Manageable       |
| Packages                       | **27**                                                    | —                |
| CI workflows                   | **5** (build, checks, performance, art-dupl, deploy-site) | —                |
| Nix build                      | **Fails** (dependency issue)                              | Broken           |
| `just check`                   | **0 issues**                                              | Clean            |
| `just test`                    | **All pass**                                              | Clean            |

### Coverage by Package (key packages)

| Package      | Coverage      |
| ------------ | ------------- |
| syntax/templ | 85.3%         |
| suffixtree   | High (cached) |
| Overall      | 73.2%         |

---

## a) FULLY DONE ✅

### Phase 0: Dead Code Elimination (Commits: `2aef77b`, `5342b85`, `daca611`, `0b88414`)

- **Deleted `cli/runtime.go` dead code:** `RuntimeConfig`, `ToConfig`, `DefaultRuntimeConfig` — kept only `DefaultThreshold` constant
- **Deleted `cli/runtime_test.go`** entirely
- **Deleted deprecated `CacheKey()`** from `cache/file_cache.go`, updated all tests to `Key()`
- **Deleted deprecated `LineNumber.Uint()`, `BytePosition.Uint()`** from `domain/types_file.go`
- **Deleted dead `SemanticHashEnabled`, `SetDefaultParseConfig`** from `syntax/golang/parse_config.go`
- **Deleted `hash/detector.go`** (shallow `HashDetector` wrapper), callers → `hash.NewFileDetector`
- **Deleted stale `examples/domain_types_usage.go`**
- **Deleted stale `domain/STRINGID_BENCHMARK_RESULTS.md`**
- **Net: ~835 lines removed**

### Phase 1: Config Builder Safety (Commit: `891fba3`)

- **Replaced `panic(err)`** in `cmd/config_builder.go:280` and `:297` with proper error returns
- `applyTimeoutFlag` and `applyDiffModeFlag` now return `error`, propagated by `applyFlagValues`

### Phase 3: Pipeline Unification (Commit: `e72f85d`)

- **Critical bug fix:** SDK's `buildAnalysisPipeline` built suffix tree but discarded it, then `runSuffixTreeDetection` rebuilt from scratch — **doubling memory and CPU**
- Introduced `pipelineResult` struct carrying both `data` and `tree`
- Both CLI and SDK now share detection dispatch through `detection.MultiDetector`
- Deleted redundant `buildSuffixTree`, `runSuffixTreeDetection`, `runHashDetection` methods

### Phase 5 (Partial): Architecture Enforcement (Commit: `dbfd4ce`, `a4db4a2`)

- Replaced ghost `.go-arch-lint.yml` with project-specific config matching actual package structure
- Consolidated `StatsPrinter`'s 8 individual setters into single `ApplyStatsConfig(StatsConfig)`
- Replaced hand-rolled `htmlEscape()` with stdlib `html.EscapeString`

### Domain Cleanup (Commit: `5601475`)

- **Deleted 6 unused domain types:** `TokenCount`, `FileCount`, `CloneCount`, `Threshold`, `BytePosition` + all their methods
- **Deleted 15 of 17 error variables** in `domain/analysis_errors.go`
- **Deleted 4 dead helper functions:** `unmarshalUintNonZero`, `unmarshalUint`, `marshalUint`, `unmarshalUintGeneric`
- domain/ now: **225 lines, 3 types, 2 error vars** — all production-used
- **Net: -185 lines**

### Linter to Zero (Commit: `9681e34`)

- Fixed all 7 linter issues → **0 issues from `golangci-lint run`**
- Fixes: gci formatting, revive package comment, 5× wsl_v5 whitespace

### Previous Session Cleanup (Commits: `0197d2a` → `4e92403`)

- Dependency updates: Go 1.26.2, ginkgo 2.28.3, gomega 1.40.0, charmbracelet packages
- Site landing page redesign (1921 insertions, 945 deletions)
- Documentation formatting and updates
- CI fix: stable Go version instead of pinned 1.26rc2
- AGENTS.md formatting, gitattributes for vendor

### Documentation (Commit: `96eadf3`)

- AGENTS.md updated with all architecture decisions from this multi-session effort
- Outstanding issues documented with migration path

---

## b) PARTIALLY DONE 🔧

### Printer Interface Decoupling

**Status:** Analysis complete, implementation NOT started.\
**What exists:**

- Full analysis of the `PrintClones(dups [][]*syntax.Node)` interface and its 6 implementations
- Identified that all 6 printers independently call `ProcessNodeRange()` + `extractContent()`
- Identified 111 test call sites that would need updating
- Identified `printer/clone_classify.go` as the main coupling point (imports `syntax/golang`)

**What's missing:**

- `ProcessedClone` DTO type definition
- Printer interface change from `[][]*syntax.Node` to `[]ProcessedCloneGroup`
- Shared `prepareClonesInfo()` extraction from printer-specific code
- Test migration (111 call sites)
- `clone_classify.go` decoupling from `syntax/golang`

### Architecture Lint Integration

**Status:** Config written, NOT integrated into CI.\
**What exists:**

- `.go-arch-lint.yml` with project-specific rules
- Rules enforce: domain must not import syntax, suffixtree must have zero deps

**What's missing:**

- `go-arch-lint` not in CI pipeline
- No architecture enforcement tests
- 93 nolint suppressions not reviewed against arch rules (81 remain)

### Clone Type Consolidation

**Status:** Analysis complete, implementation NOT started.\
**Parallel types identified:**

| Type                            | Package              | Fields                                                                 | Used By            |
| ------------------------------- | -------------------- | ---------------------------------------------------------------------- | ------------------ |
| `printer.clone` (unexported)    | printer/common.go    | filename, lineStart, lineEnd, fragment, size, fileSize, classification | All 6 printers     |
| `pkg/artdupl.Clone`             | pkg/artdupl/types.go | Filename, StartLine, EndLine, StartPos, EndPos, Fragment, Size         | SDK                |
| `printer.JSONClone`             | printer/json.go      | Filename, LineStart, LineEnd, Fragment                                 | JSON output        |
| `printer.SimpleJSONClone`       | printer/json.go      | Filename, TokenCount, LineRangeMixin                                   | Simple JSON output |
| `printer.CloneGroup`            | printer/json.go      | Hash, Size, Files []JSONClone                                          | JSON output        |
| `pkg/artdupl.CloneGroup`        | pkg/artdupl/types.go | Hash, Clones, Size, LineCount, Method                                  | SDK                |
| `printer.CloneWithContentMixin` | printer/diff.go      | Filename, LineStart, LineEnd                                           | Diff output        |
| `printer.LineRangeMixin`        | printer/json.go      | StartLine, EndLine                                                     | JSON output        |

---

## c) NOT STARTED ⬜

1. **`cmd/run_analysis.go` extraction** — 447-line god file, 7 internal imports, 6 responsibilities
2. **Nix build fix** — `nix build .` fails with dependency issue
3. **Coverage improvement** — 73.2% vs 80% target
4. **Dead error constructors removal** — `NewParseError`, `NewDetectionError`, `NewAnalysisError`, `NewCancelledError`, `NewTimeoutError`, `Wrapf` have 0 production callers (only test callers)
5. **Dead error types review** — `CancelledError`, `TimeoutError` error types have 0 production usage
6. **exhaustruct nolint cleanup** — 21 suppressions, most are `DuplError` constructions where File/Line are legitimately optional
7. **funlen complexity reduction** — 5 functions suppressed for length (cmd/run_all_modes.go, cmd/run_flags.go, cmd/stats.go, config/config_merge.go, syntax/golang/transform.go)
8. **gocyclo complexity reduction** — 2 functions (config/config_merge.go, printer/html_diff.go)
9. **Printer test migration** — 111 test call sites using `[][]*syntax.Node`
10. **SDK_DESIGN.md update** — documents architecture that has changed significantly
11. **FEATURES.md update** — last updated 2026-02-12, many changes since
12. **`justfile` → `flake.nix` migration** — justfile still primary, flake.nix exists but build fails

---

## d) TOTALLY FUCKED UP 💥

### Nothing is catastrophically broken.

However, there are two areas of concern:

1. **Nix build is broken** — `nix build .` fails with a dependency error. The binary builds fine with `go build` and `just build`, so this is a Nix-specific issue (likely the gogenfilter private dependency handling). The `flake.nix` exists but the vendorHash or the dummy/replace pattern needs updating.

2. **`domain/` package is a ghost system with training wheels** — Only 1 production consumer (`detection/todos.go`) uses 3 types (Filepath, LineNumber, CloneSeverity). The remaining domain types were deleted, but the package structure is questionable:
   - `domain.Filepath` is just `string` with a non-empty validation
   - `domain.LineNumber` is `uint16` with a non-zero validation
   - `domain.CloneSeverity` is `string` with 4 valid values
   - All three add overhead (JSON marshal/unmarshal, validation) for minimal type safety gain
   - These could live directly in `detection/` without the domain abstraction

---

## e) WHAT WE SHOULD IMPROVE

### Architecture

1. **Printer ↔ syntax.Node coupling is THE keystone issue.** The `PrintClones(dups [][]*syntax.Node)` interface forces all 6 printers to understand AST internals. Every printer independently calls `ProcessNodeRange()` → `extractContent()`. This is the single highest-leverage change remaining.

2. **`printer/` package is 10,652 lines across 44 files.** It's the largest package by 3.6× (next is `syntax/` at 4,723). It contains printing logic, sorting, classification, diffing, statistics, HTML templates, and JSON marshaling. Should be split into `printer/`, `printer/html/`, `printer/stats/`, `printer/diff/`.

3. **Three parallel Clone type hierarchies** represent the same data differently. The DTO from fix #1 would collapse them.

4. **`cmd/run_analysis.go` (447 lines)** is a god file importing 7 internal packages. It orchestrates parsing, tree building, detection, filtering, printing, and progress reporting. Should be decomposed.

5. **`domain/` package's existence is questionable.** Only `detection/todos.go` uses it. The types could live in `detection/` directly, eliminating a package boundary.

### Code Quality

6. **21 `exhaustruct` suppressions** — almost all are `DuplError{}` constructions where `File` and `Line` are legitimately optional. Fix: make `DuplError` a builder pattern or use functional options.

7. **10 `gochecknoglobals` suppressions** — mostly lookup tables and build-time variables. All legitimate, but could use `sync.Once` patterns for the tables.

8. **5 `funlen` suppressions** — these are real complexity issues:
   - `cmd/run_all_modes.go:16` — orchestrates multi-format output
   - `cmd/run_flags.go:40` — handles many CLI flags
   - `cmd/stats.go:63` — stats command flags
   - `config/config_merge.go:17` — merges ~20 config fields
   - `syntax/golang/transform.go:12` — AST transformation (inherent complexity)

### Testing

9. **Coverage at 73.2% vs 80% target.** The `just check-coverage` recipe exists but fails.

10. **111 test call sites use `[][]*syntax.Node` directly.** These make the Printer DTO refactor expensive.

### Documentation

11. **FEATURES.md** last updated 2026-02-12 (2.5 months ago). Does not reflect multi-method detection, BDD tests, stats subcommand improvements.

12. **SDK_DESIGN.md** documents an architecture that has changed significantly (pipeline unification, domain cleanup).

13. **MIGRATION_TO_NIX_FLAKES_PROPOSAL.md** exists but nix build is broken.

### Infrastructure

14. **Nix build broken** — needs investigation.

15. **CI uses `actions/setup-go@v5`** — could use `nix develop` for full parity (per migration proposal).

---

## f) TOP 25 THINGS TO DO NEXT (Ranked by Impact × Effort⁻¹)

| #  | Task                                                                                                                                                         | Impact    | Effort           | Category       |
| -- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ | --------- | ---------------- | -------------- |
| 1  | **Fix nix build** — investigate dependency error, update vendorHash                                                                                          | High      | Low              | Infrastructure |
| 2  | **Delete dead error constructors** — `NewParseError`, `NewDetectionError`, `NewAnalysisError`, `NewCancelledError`, `NewTimeoutError`, `Wrapf` + their tests | Low-Med   | Very Low         | Cleanup        |
| 3  | **Fix `exhaustruct` on DuplError** — make File/Line optional via builder or functional options, remove 10+ nolint                                            | Med       | Low              | Code Quality   |
| 4  | **Extract `cmd/run_analysis.go`** — split into run_parser.go, run_tree.go, run_detection.go                                                                  | Med       | Low-Med          | Architecture   |
| 5  | **Reduce `config/config_merge.go` complexity** — extract per-field merge helpers                                                                             | Low-Med   | Low              | Code Quality   |
| 6  | **Split `printer/` package** — html/ stats/ diff/ subpackages                                                                                                | High      | Med-High         | Architecture   |
| 7  | **Introduce `ProcessedClone` DTO** — define the type, implement converter from `[][]*syntax.Node`                                                            | High      | Med              | Architecture   |
| 8  | **Migrate Printer interface** to `PrintClones([]ProcessedCloneGroup)`                                                                                        | Very High | High (111 tests) | Architecture   |
| 9  | **Consolidate Clone types** — collapse 3 Clone/CloneGroup hierarchies                                                                                        | High      | Med              | Architecture   |
| 10 | **Move `clone_classify.go`** out of printer/ — make classification language-agnostic                                                                         | Med       | Med              | Architecture   |
| 11 | **Fix coverage to ≥80%** — identify uncovered paths, add tests                                                                                               | Med       | Med              | Testing        |
| 12 | **Update FEATURES.md** — reflect current capabilities                                                                                                        | Low       | Very Low         | Documentation  |
| 13 | **Update SDK_DESIGN.md** — reflect pipeline unification                                                                                                      | Low       | Very Low         | Documentation  |
| 14 | **Reduce `html_diff.go` complexity** — extract case handlers                                                                                                 | Low-Med   | Low              | Code Quality   |
| 15 | **Consolidate `LineRangeMixin` + `CloneWithContentMixin`** — both provide filename+lineStart+lineEnd                                                         | Low       | Low              | Cleanup        |
| 16 | **Add `go-arch-lint` to CI** — enforce package boundaries automatically                                                                                      | Med       | Low              | Infrastructure |
| 17 | **Remove or justify `domain/` package** — only 1 consumer, questionable value                                                                                | Low       | Low-Med          | Architecture   |
| 18 | **Write architecture enforcement tests** — verify domain doesn't import syntax, etc.                                                                         | Med       | Low              | Testing        |
| 19 | **Fix `funlen` in `cmd/run_flags.go`** — extract flag groups into separate functions                                                                         | Low       | Low              | Code Quality   |
| 20 | **Fix `funlen` in `cmd/run_all_modes.go`** — extract format-specific output                                                                                  | Low       | Low              | Code Quality   |
| 21 | **Delete `domain/analysis_errors.go` dead errors** — only 2 of 17 remain, move to consumers                                                                  | Low       | Very Low         | Cleanup        |
| 22 | **Add `//go:build` tags** for SIMD files — separate portable vs platform-specific code                                                                       | Low       | Low              | Code Quality   |
| 23 | **Integrate `nix develop` with CI** — reproducible builds in CI                                                                                              | Med       | Med              | Infrastructure |
| 24 | **Add `nix flake check` to CI** — automated nix validation                                                                                                   | Med       | Low              | Infrastructure |
| 25 | **Create CONTRIBUTING.md** — document the development workflow, commit conventions                                                                           | Low       | Low              | Documentation  |

---

## g) TOP #1 QUESTION I CANNOT FIGURE OUT MYSELF

**What is the strategic intent for the `domain/` package?**

It currently has 225 lines, 3 types (Filepath, LineNumber, CloneSeverity), and only 1 production consumer (`detection/todos.go`). The types provide:

- Non-empty/non-zero validation on construction
- Custom JSON marshaling
- Type safety (you can't pass a `Filepath` where a `LineNumber` is expected)

But the entire rest of the codebase uses raw `int`, `string` for these same concepts. There are two paths:

**Path A: Keep and expand domain/ types** — Use them in `syntax.Node` (replacing `int` positions), in `pkg/artdupl.Clone` (replacing `int` fields), in printer DTOs. This would require the Printer DTO change first (item #8 above) and would cascade through the codebase.

**Path B: Absorb into detection/** — Move the 3 types into `detection/` (their only consumer), eliminate the `domain/` package entirely. Accept that the project uses idiomatic Go primitives (`int`, `string`) for positions and paths, with validation at boundaries.

This decision affects the Printer DTO design (#7-8), the Clone type consolidation (#9), and whether `domain/` is worth keeping as a package. I can't make this call without understanding the project's longer-term vision for type safety vs. idiomatic Go.

---

## Appendix: Codebase Composition

| Package     | Lines      | % of Total |
| ----------- | ---------- | ---------- |
| printer/    | 10,652     | 23.4%      |
| syntax/     | 4,723      | 10.4%      |
| pkg/        | 4,029      | 8.9%       |
| cmd/        | 3,813      | 8.4%       |
| internal/   | 3,833      | 8.4%       |
| config/     | 2,936      | 6.5%       |
| job/        | 1,525      | 3.4%       |
| detection/  | 1,384      | 3.0%       |
| suffixtree/ | 1,170      | 2.6%       |
| hash/       | 889        | 2.0%       |
| cache/      | 901        | 2.0%       |
| errors/     | 952        | 2.1%       |
| domain/     | 225        | 0.5%       |
| cli/        | 225        | 0.5%       |
| **Total**   | **45,481** | **100%**   |

### Coupling Map (production imports)

```
syntax/golang ← job/file_parser.go, internal/testutil, printer/clone_classify.go
syntax        ← printer (9 files), detection, hash, pkg/artdupl, cmd
domain        ← detection/todos.go (only prod consumer)
detection     ← cmd, pkg/artdupl
suffixtree    ← syntax, pkg/artdupl
config        ← cmd, cli, pkg/artdupl, detection, printer
```

### Nolint Breakdown (production code, 81 total)

| Suppression      | Count | Mostly Legitimate?               |
| ---------------- | ----- | -------------------------------- |
| exhaustruct      | 21    | No — `DuplError` design issue    |
| gochecknoglobals | 10    | Yes — lookup tables + build vars |
| wrapcheck        | 9     | Yes — stdlib passthrough         |
| gosec            | 7     | Yes — bounded values             |
| nonamedreturns   | 5     | Yes — counting functions         |
| funlen           | 5     | No — real complexity             |
| mnd              | 4     | Yes — file permissions           |
| funcorder        | 4     | Partially                        |
| forbidigo        | 3     | Yes — version/demo output        |
| godox            | 2     | Yes — test patterns              |
| gocyclo/cyclop   | 2     | No — real complexity             |
| gocognit         | 2     | No — real complexity             |
| nilnil           | 1     | Yes — intentional nil+nil        |
| nestif           | 1     | Yes — HTML generation            |
| maintidx         | 1     | No — AST transform               |
| gochecknoinits   | 1     | Yes — gob registration           |
