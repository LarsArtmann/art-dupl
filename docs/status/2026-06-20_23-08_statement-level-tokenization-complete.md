# Status Report — 2026-06-20 23:08 CEST

**Branch:** `fork` (1 commit ahead of `origin/fork`)
**Head:** `807e373` — fix: calibrate BDD tests for statement-level tokenization and fix size sort
**Working tree:** Clean
**Binary:** Builds clean
**Tests:** 24/24 packages passing, 0 failures

---

## Executive Summary

Statement-level tokenization (T1) is **complete and verified**. The core algorithm was committed in `008ccf9`, and this session finished the follow-up work: calibrating 240+ BDD tests for new threshold semantics, fixing two real algorithm bugs discovered during calibration, and achieving a fully green test suite across all 24 packages.

The project is in a **strong architectural state** — 100% domain coverage, 92.7% detection coverage, zero goroutine leaks, decoupled SDK, and 4 ADRs documenting major decisions. The remaining work is primarily **type consolidation** (5 parallel Clone types) and **language expansion** (TypeScript/Python support).

---

## a) FULLY DONE ✅

### T1: Statement-Level Tokenization (This Session)

| Item                                              | Status  | Detail                                                           |
| ------------------------------------------------- | ------- | ---------------------------------------------------------------- |
| Core algorithm (`serial()`, `fingerprintSubtree`) | ✅ Done | FNV-1a composite hashing, committed in `008ccf9`                 |
| Structural wrapper false-positive fix             | ✅ Done | `dataContainsStatements()` guard in `FindSyntaxUnits()`          |
| Size sort bug fix                                 | ✅ Done | `GetCloneSize` now uses fragment length, not broken `Owns` field |
| BDD test calibration (240+ specs)                 | ✅ Done | Thresholds rescaled: node-level (8-15) → statement-level (1-3)   |
| Templ threshold preservation                      | ✅ Done | Templ uses legacy tokenization; thresholds kept at 3+            |
| Full test suite green                             | ✅ Done | 24/24 packages, 0 failures, stable across 5 runs                 |

### Previously Completed (Prior Sprints)

| Area                                | Status  | Highlight                                                                                     |
| ----------------------------------- | ------- | --------------------------------------------------------------------------------------------- |
| Semantic detection (3 modes)        | ✅ Done | Structural, semantic (default), hash-based                                                    |
| Alpha-normalization (Type 2 clones) | ✅ Done | Identifier name normalization                                                                 |
| SDK decoupling                      | ✅ Done | `pkg/artdupl` has ZERO imports from `config/` or `errors/`                                    |
| Goroutine leak elimination          | ✅ Done | 11 leaks fixed across SDK, job, cmd packages                                                  |
| Actionability classification        | ✅ Done | 10 non-actionable patterns detected                                                           |
| Baseline recording + CI check       | ✅ Done | `baseline` + `check` subcommands                                                              |
| Extractability scoring              | ✅ Done | JSON output includes `extractable` + `lines_saved`                                            |
| Output formats (7 total)            | ✅ Done | Text, HTML, JSON, Simple-JSON, Plumbing, SARIF, CSV                                           |
| Sorting (4 criteria)                | ✅ Done | Size, occurrence, hash, total-tokens                                                          |
| Context propagation                 | ✅ Done | All detection goroutines respect `context.Context`                                            |
| Dead code removal                   | ✅ Done | Ghost systems eliminated (Filepath/LineNumber branded types, 5 dead error constructors, etc.) |
| Error handling cleanup              | ✅ Done | SDK uses stdlib only, no `debug.Stack()` on hot paths                                         |

### Quality Metrics

| Metric               | Value                                                   |
| -------------------- | ------------------------------------------------------- |
| Production code      | ~20,200 lines                                           |
| Test code            | ~31,000 lines (1.53x ratio)                             |
| BDD specs            | ~240 `It` blocks across 20 test files                   |
| Domain coverage      | **100%**                                                |
| Detection coverage   | **92.7%**                                               |
| SDK coverage         | **84.3%**                                               |
| Syntax coverage      | **78.5%**                                               |
| Printer coverage     | **75.8%**                                               |
| Job coverage         | **71.6%**                                               |
| golangci-lint issues | 5 (3 exhaustruct in test helpers, 2 pre-existing gosec) |
| ADRs                 | 4                                                       |

---

## b) PARTIALLY DONE 🟡

| Area                         | What's Done                                                              | What Remains                                                                                                                                             |
| ---------------------------- | ------------------------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Clone type consolidation** | Field names aligned (`LineStart`/`LineEnd` canonical across all 5 types) | 5 separate types still exist: `printer.CloneGroup`, `pkg/artdupl.Clone`, `pkg/artdupl.CloneGroup`, `domain.ProcessedClone`, `domain.ProcessedCloneGroup` |
| **Printer package split**    | Actionability split into 4 files; view models renamed to `*View`         | ~29 source files / ~3500+ lines still in one package                                                                                                     |
| **Context propagation**      | All detection goroutines + suffix tree respect ctx                       | `cmd/run_crawl.go` file feeders (stdin scanner, `filepath.Walk`) still blocking                                                                          |
| **Templ semantic mode**      | Templ detection works (structural mode)                                  | No statement-level tokenization for templ — `syntax/templ/` is purely structural                                                                         |
| **InternFilename**           | Wired into all 4 transformer construction sites                          | Type-safe RWMutex+map (could use further profiling)                                                                                                      |
| **Performance benchmarks**   | Suffix tree benchmarks added (T20)                                       | No regression test suite; no performance baseline file                                                                                                   |

---

## c) NOT STARTED ⬜

| Item                                | Priority | Notes                                                      |
| ----------------------------------- | -------- | ---------------------------------------------------------- |
| TypeScript/JavaScript support       | Roadmap  | Would need new AST transformer in `syntax/`                |
| Python support                      | Roadmap  | Would need new AST transformer in `syntax/`                |
| Watch mode (incremental)            | Roadmap  | Continuous monitoring + incremental detection              |
| Hybrid slice/map transition storage | Medium   | Map already O(1); deferred optimization                    |
| Performance regression suite        | Roadmap  | No automated perf regression detection                     |
| More ADRs                           | Roadmap  | Only 4 exist; T1 statement-level tokenization deserves one |

---

## d) TOTALLY FUCKED UP 💥 (Honest Assessment)

Nothing is catastrophically broken. But here's what caused the most pain:

### 1. `GetCloneSize` Was Silently Broken in Semantic Mode (Pre-Existing Bug)

**The bug:** `GetCloneSize()` used `group[0][0].Owns` to determine clone size for sorting. But statement-level tokenization sets `Owns=0` for all statement nodes. This meant **every clone group had size 0**, making size-based sort completely non-deterministic. This bug existed in the **default mode** (semantic) and was only exposed when the sorting BDD test started failing intermittently.

**Why it was hard to find:** The test passed in isolation (Go's sort is not stable for equal elements, but with only 2 groups the order happened to work). It failed in the full suite due to map iteration order seeding differences. Required empirical binary comparison + in-process debug to isolate.

**Impact:** Users running `--sort size` in semantic mode (the default) got **random ordering** of clone groups. This was a silent correctness bug.

### 2. The Threshold Semantics Change Broke 67 Tests

Changing from node-level thresholds (8-15 AST nodes per statement) to statement-level (1 statement = 1 token) was a **breaking change** to the threshold parameter. Every BDD test with `--threshold 3` or `--threshold 5` suddenly found either nothing or everything. This required touching **20 test files** with 100+ individual threshold changes.

**Lesson:** When changing the semantic meaning of a user-facing parameter, the migration is never "just find and replace." Each test fixture needed individual calibration based on how many statements it actually contains.

### 3. Structural Wrapper False Positives

The `getUnitsIndexes()` hybrid mode fallback was supposed to handle legacy/synthetic nodes, but statement-level tokenized streams have structural wrappers (FuncDecl, File) that lack `Statement=true`. These fell into the legacy path and got reported as clones. This caused semantic mode to find **FuncDecl structural similarities** as false positive clones.

**The fix:** `dataContainsStatements()` — if the global stream has ANY statement nodes, matches with zero statement indexes are skipped entirely.

### 4. Pre-Existing Lint Debt in T1 Code

The `fingerprintSubtree` function has 2 gosec G115 warnings (int32→uint32 overflow). These are **intentional** (FNV hash deliberately wraps) but lack `//nolint:gosec` annotations. They were committed in `008ccf9` and should be annotated.

---

## e) WHAT WE SHOULD IMPROVE 🎯

### Architecture

1. **Consolidate the 5 Clone types** — This is the biggest architectural debt. Every clone passes through 3-5 type conversions. The field names are aligned but the types are separate. This adds cognitive load and conversion bugs.

2. **Split `printer/` package** — 29 source files / 3500+ lines is too many for one package. Natural splits: `printer/stats/`, `printer/html/`, `printer/json/`, `printer/text/`.

3. **Add statement-level tokenization to templ** — Templ is purely structural. Adding `Statement=true` marking to templ nodes would enable semantic mode for templ files too.

4. **Thread context through file feeders** — The last blocking I/O paths (stdin scanner, `filepath.Walk`) don't respect context cancellation. Full chain refactor needed.

5. **Write ADR-0005 for statement-level tokenization** — The T1 design (FNV-1a fingerprinting, semantic encoding layout) deserves documentation.

### Testing

6. **Fix exhaustruct warnings in test helpers** — `internal/testutil/node.go` has 3 instances of `syntax.Node` construction missing the `Statement` field. These are test helpers that should set all fields.

7. **Add performance regression tests** — We have benchmarks but no automated regression detection. A CI job that fails on >10% perf regression would catch algorithmic degradation.

8. **Increase job package coverage** — At 71.6%, this is the lowest coverage among core packages. The incremental detection path needs more tests.

9. **Add integration test for templ semantic mode** — When templ gets statement marking, we need tests ready.

### Code Quality

10. **Annotate gosec G115 warnings** — Add `//nolint:gosec // G115: FNV hash intentionally wraps` to `fingerprintSubtree` calls.

11. **Rename `Owns` field** — The name is misleading (it means "child count" in legacy mode but is always 0 in statement mode). Consider `LegacyChildCount` or deprecate entirely.

12. **Document the semantic encoding layout** — The `[24-bit identifier/operator hash][8-bit base AST node type]` layout is critical knowledge that's only in AGENTS.md. It should be in an ADR.

### Developer Experience

13. **Add `--threshold` help text explaining statement semantics** — Users coming from node-level thresholds will be confused that threshold 1 now finds matches.

14. **Consider a `--min-statements` alias** — The threshold parameter's meaning changed; a more explicit name could help.

15. **Update HOW_TO_USE.md** — Document the new threshold semantics and the three detection modes clearly.

---

## f) Top 25 Things to Get Done Next

### 🔴 Critical (Do First)

| #   | Task                                                                       | Impact                   | Effort |
| --- | -------------------------------------------------------------------------- | ------------------------ | ------ |
| 1   | **Push commit to origin**                                                  | Unblocks CI              | 1 min  |
| 2   | **Annotate gosec G115 in `fingerprintSubtree`**                            | Clean lint               | 5 min  |
| 3   | **Fix exhaustruct in `internal/testutil/node.go`** (add `Statement` field) | Clean lint               | 10 min |
| 4   | **Write ADR-0005: Statement-Level Tokenization**                           | Document critical design | 30 min |
| 5   | **Update HOW_TO_USE.md with new threshold semantics**                      | User-facing docs         | 20 min |

### 🟡 High Value

| #   | Task                                                        | Impact                     | Effort   |
| --- | ----------------------------------------------------------- | -------------------------- | -------- |
| 6   | **Consolidate Clone types** (5 → 2: internal DTO + SDK DTO) | Eliminates conversion bugs | 2-3 days |
| 7   | **Split `printer/` into sub-packages**                      | Maintainability            | 1 day    |
| 8   | **Add templ statement-level tokenization**                  | Semantic mode for templ    | 1 day    |
| 9   | **Thread context through `cmd/run_crawl.go`**               | Last cancellation gap      | 4 hours  |
| 10  | **Increase job package coverage** (71.6% → 85%+)            | Test confidence            | 4 hours  |
| 11  | **Add performance regression CI job**                       | Catch perf degradation     | 3 hours  |
| 12  | **Rename/deprecate `Owns` field**                           | Clarity                    | 2 hours  |

### 🟢 Medium Value

| #   | Task                                                                  | Impact                | Effort  |
| --- | --------------------------------------------------------------------- | --------------------- | ------- |
| 13  | **Add `--min-statements` alias for `--threshold`**                    | UX clarity            | 1 hour  |
| 14  | **Create performance baseline JSON**                                  | Regression tracking   | 2 hours |
| 15  | **Add TypeScript AST transformer (research spike)**                   | Language expansion    | 1 week  |
| 16  | **Implement watch mode**                                              | Continuous monitoring | 3 days  |
| 17  | **Add `--diff` support for all output formats** (currently HTML only) | Feature completeness  | 1 day   |
| 18  | **Cache invalidation strategy for incremental mode**                  | Performance           | 1 day   |
| 19  | **Add `--exclude-pattern` glob support**                              | Filtering flexibility | 3 hours |
| 20  | **Document SDK usage patterns with examples**                         | Adoption              | 4 hours |

### 🔵 Polish

| #   | Task                                                 | Impact        | Effort  |
| --- | ---------------------------------------------------- | ------------- | ------- |
| 21  | **Add shell completion generation** (Cobra built-in) | CLI UX        | 1 hour  |
| 22  | **Add `art-dupl init` for config file scaffolding**  | Onboarding    | 2 hours |
| 23  | **Add `--format table` interactive output**          | Terminal UX   | 4 hours |
| 24  | **Internationalize error messages**                  | Accessibility | 1 day   |
| 25  | **Add Homebrew formula**                             | Distribution  | 2 hours |

---

## g) Top #1 Question I Cannot Figure Out Myself 🤔

**Should the 5 Clone types be consolidated into 1, 2, or 3 types?**

The current state:

- `domain.ProcessedClone` — canonical internal DTO (richest, has actionability/classification)
- `domain.ProcessedCloneGroup` — canonical internal group DTO
- `printer.CloneGroup` — JSON serialization DTO (has `Files []JSONClone`)
- `pkg/artdupl.Clone` — SDK public API DTO (simpler, fewer fields)
- `pkg/artdupl.CloneGroup` — SDK public API group DTO

**The tension:**

- **1 type** (everything is `domain.ProcessedClone`): Simplest, but couples the SDK to internal domain concepts (actionability patterns, classification details). The SDK was deliberately decoupled from `config/` and `errors/` — should it also be decoupled from `domain/`?
- **2 types** (internal `domain.ProcessedClone` + external `pkg/artdupl.Clone`): Clean boundary, but requires conversion at every printer/SDK call site. The printer currently imports `domain.ProcessedClone` directly.
- **3 types** (domain + printer DTO + SDK DTO): Maximum decoupling, but maximum conversion overhead and risk of field drift.

**Why I can't decide:** The printer sits in a weird position — it's internal (not public API) but needs to produce both internal (text/HTML) and external (JSON/Simple-JSON) output. Should the printer consume `domain.ProcessedClone` directly, or should there be a `printer.CloneView` intermediate? The answer depends on whether we consider the JSON output format to be a **public API contract** (in which case the DTO should be stable and separate) or an **internal detail** (in which case domain DTO is fine).

This is a **business/architecture decision** that requires knowing the project's API stability guarantees and target audience (library consumers vs CLI users).

---

## Verification Commands

```bash
go test -count=1 ./...          # 24/24 packages, 0 failures
go build ./...                  # Clean compilation
golangci-lint run ./...         # 5 issues (3 exhaustruct test, 2 gosec pre-existing)
go build -o /tmp/art-dupl ./cmd/art-dupl/ && /tmp/art-dupl . -t 1  # Dogfood works
```

---

_Generated 2026-06-20 23:08 CEST_
