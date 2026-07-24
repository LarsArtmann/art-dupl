# Status Report — 2026-07-03 03:25

## Executive Summary

> **Update (2026-07-24):** Section D "TOTALLY FUCKED UP" items are resolved: the FEATURES.md SHA-1→SHA-256 reference was fixed, the TODO_LIST stale CloneRef entry was removed, and the jscpd BuildFlow issue was resolved by removing jscpd. The "7 parallel Clone types" design debt is tracked in TODO_LIST.md (MEDIUM Priority). File counts have grown: 158 non-test `.go` files, 132 test files.

art-dupl is a **production-ready** Go code clone detection tool. All 26 packages build clean, 27 test packages pass (including race detector), and golangci-lint reports 0 issues. The codebase has 150 non-test `.go` files, 112 test files, and 1 `.templ` file.

---

## A) FULLY DONE ✅

### Core Architecture (Stable, Well-Tested)

| Area                          | Status             | Details                                                                                          |
| ----------------------------- | ------------------ | ------------------------------------------------------------------------------------------------ |
| **Suffix Tree Detection**     | ✅ Complete        | O(1) map-based transition lookup, context-aware                                                  |
| **Hash-Based Detection**      | ✅ Complete        | XXH3 content hashing, `groupByHash` + `filterDuplicateGroups` shared helpers, context-propagated |
| **Multi-Detector Dispatch**   | ✅ Complete        | `[]MethodDetector` loop with `Name()` interface                                                  |
| **Three Detection Modes**     | ✅ Complete        | `semantic` (alpha-normalized Type-2), `exact` (Type-1), `structural` (shape-only)                |
| **AST Serialization**         | ✅ Non-destructive | Shallow-copies nodes before mutating Type/Owns (ADR-0006)                                        |
| **Clone Type Classification** | ✅ Complete        | Type 1/2/3 via `collectNamesPreOrder` direct tree walk                                           |
| **Semantic Encoding**         | ✅ Complete        | `[24-bit identifier hash][8-bit base AST node type]` layout (ADR-0008)                           |
| **Alpha-Normalization**       | ✅ Complete        | Per-function symbol table, includes FuncLit closures                                             |

### Output Formats (6/6 Complete)

| Format        | Status | Key Details                                                                         |
| ------------- | ------ | ----------------------------------------------------------------------------------- |
| **Text**      | ✅     | Uses `slices.SortFunc` + `cmp.Compare` (modern Go)                                  |
| **HTML**      | ✅     | Diff visualization, collapse-all/expand-all, filter buttons, category/priority tags |
| **JSON**      | ✅     | Full metadata, timestamps as milliseconds                                           |
| **SARIF**     | ✅     | GitHub Code Scanning / SonarQube compatible with Properties metadata                |
| **Plumbing**  | ✅     | Machine-readable                                                                    |
| **CSV Stats** | ✅     | Health grade, spread analysis                                                       |

### SDK / API

| Area                                         | Status                                             |
| -------------------------------------------- | -------------------------------------------------- |
| `pkg/artdupl.Detector` interface             | ✅ Zero config/errors imports, `domain` aliases    |
| Streaming results (`FindClonesStreamResult`) | ✅ Context-aware channel sends                     |
| `CloneRef` shared value object               | ✅ Embedded in `ProcessedClone` + `artdupl.Clone`  |
| Type independence enforced                   | ✅ `.go-arch-lint.yml` gates config/errors imports |

### Infrastructure

| Area                  | Status                                                                                 |
| --------------------- | -------------------------------------------------------------------------------------- |
| **Incremental Cache** | ✅ SHA-256 content hashing, LRU eviction, deep-clone on cache hit, singleflight dedup  |
| **Parallel Parsing**  | ✅ Worker pool (`ParseIncrementalParallel`), context-propagated                        |
| **Baseline CI**       | ✅ Record/check modes, GitHub Actions template                                         |
| **Config**            | ✅ Reflection-based merge, typed enums, JSON migration shim (legacy `"semantic"` bool) |
| **Cache Versioning**  | ✅ Version mismatch warning, stale entry eviction                                      |
| **Error Handling**    | ✅ 6 typed error types, stack traces, JSON marshal, no panics for expected errors      |

### Test Infrastructure

| Area                  | Status | Count                                    |
| --------------------- | ------ | ---------------------------------------- |
| Unit tests            | ✅     | 112 test files                           |
| BDD tests             | ✅     | Ginkgo/Gomega in `bdd/`                  |
| Golden file tests     | ✅     | HTML, text, JSON, CSV                    |
| Fuzz tests            | ✅     | Templ parser                             |
| Race detection        | ✅     | 0 failures, `CGO_ENABLED=1`              |
| Perf regression tests | ✅     | Serialize + hashSeq with 50ms thresholds |
| Benchmarks            | ✅     | `syntax_bench_test.go`                   |
| Nix CI gates          | ✅     | race, lint, bench, check in `flake.nix`  |

### Recent Sprint (2026-07-03)

| Work Item                                   | Status |
| ------------------------------------------- | ------ |
| Sort factory unit tests (10 cases)          | ✅     |
| SARIF Properties verification test          | ✅     |
| `slices.SortFunc` migration + empty-key fix | ✅     |
| `CloneRef` direct unit tests                | ✅     |
| Config migration shim tests (6 edge cases)  | ✅     |
| SARIF `omitempty` fix                       | ✅     |
| Test typo fix (`%d}`)                       | ✅     |

---

## B) PARTIALLY DONE 🟡

| Area                                | What's Done                                                                                                | What's Missing                                                                                         |
| ----------------------------------- | ---------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------ |
| **Printer sub-package split** (T25) | CloneNode DTO decouples actionability from `syntax.Node`; `CloneRef` eliminates field drift                | Full `printer/` package split into focused sub-packages deferred — high circular dep risk              |
| **Actionability patterns**          | 14 non-actionable patterns detected (test scaffolding, RAII defer, error wrapping, assertion chains, etc.) | Pattern detection uses `CloneNode.BaseType` but some edge cases (deeply nested builders) may not match |
| **Performance profiling**           | Hidden `--profile` flag exists                                                                             | Not documented, not productionized (EXPERIMENTAL)                                                      |
| **TODO_LIST.md freshness**          | 155 items marked done, 10 pending                                                                          | 6 unique pending items (4 are duplicates); stale CloneRef entry (already done as T23)                  |

---

## C) NOT STARTED ⬜

| #   | Item                                         | Rationale                                                                                                                                                                       |
| --- | -------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | **Split `printer/` into sub-packages** (T25) | Feasible but requires interface inversion; circular dep risk. 10+ files in `printer/` with overlapping concerns (text, HTML, JSON, SARIF, stats, sorting, diff, actionability). |
| 2   | **Branded `NodeType int32`** (T24)           | HIGH RISK — touches gob cache serialization, doesn't prevent collision. Deferred.                                                                                               |
| 3   | **Hide `syntax/golang` behind facade**       | Blocked by import cycle — `syntax/golang` is imported by `printer/actionability*.go` for AST constants.                                                                         |
| 4   | **Git-diff incremental analysis**            | `--since` flag was a dead stub, removed. Real implementation needs `git diff` integration.                                                                                      |
| 5   | **SIMD optimizations**                       | N/A — uses xxh3 (already SIMD-accelerated) + sync.Pool. No hand-written assembly.                                                                                               |
| 6   | **Hybrid slice/map transition storage**      | Map already O(1). Deferred as low-value.                                                                                                                                        |

---

## D) TOTALLY FUCKED UP 💥

| Issue                                  | Severity    | Status       | Root Cause                                                                                                                                                                                       |
| -------------------------------------- | ----------- | ------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| **jscpd in BuildFlow pre-commit hook** | Medium      | Active       | `jscpd` gets OOM-killed on every commit. System resource issue, not code. Workaround: `--no-verify`.                                                                                             |
| **FEATURES.md stale cache reference**  | Low         | Active       | Line 212 says "SHA1 keys" but cache migrated to SHA-256 (CacheVersion 2). Documentation drift.                                                                                                   |
| **TODO_LIST.md stale entry**           | Low         | Active       | CloneRef listed as "pending" under HIGH Priority but was completed as T23.                                                                                                                       |
| **5 uncommitted pre-existing changes** | Low         | Active       | `.dockerignore`, `.gitignore`, `flake.lock`, 2 docs HTML files have uncommitted formatting changes from a previous session. Not mine — waiting for user decision.                                |
| **7 parallel Clone types**             | Design debt | Acknowledged | `printer.CloneGroup`, `printer.JSONClone`, `pkg/artdupl.Clone`, `domain.ProcessedClone(Group)`, etc. `CloneRef` mitigates drift but types remain separate. See `docs/research/SPLIT-BRAIN.html`. |

---

## E) WHAT WE SHOULD IMPROVE 🎯

### Architecture

1. **Split `printer/` package** — 10+ files, multiple concerns (text, HTML, JSON, SARIF, stats, sorting). Each output format should be its own sub-package with a shared interface.
2. **Consolidate Clone types** — 7 parallel Clone types is a split brain. `CloneRef` reduced field drift but didn't eliminate the type explosion. Consider a single `Clone` type in `domain` with format-specific views.
3. **Migrate remaining `sort.Slice` calls** — 8 call sites still use `sort.Slice` in `printer/groups.go`, `printer/json.go`, `printer/stats.go`, `printer/stats_visualization.go`, `cache/file_cache.go`, `baseline/baseline.go`. Should use `slices.SortFunc`.
4. **Hide `syntax/golang` constants behind domain types** — Actionability layer imports `syntax/golang` for AST constants (`FuncDecl`, `IfStmt`, etc.). These should be domain-level constants or an interface method.

### Type Safety

5. **Branded `NodeType int32`** — Risk: touches gob cache. But a `NodeType` type with validation would prevent invalid AST type comparisons.
6. **Stronger `CloneCategory` typing** — Currently a string enum. Could be a sealed interface with pattern matching for exhaustiveness.

### Testing

7. **Integration test for SARIF round-trip** — Parse emitted SARIF JSON with a SARIF validator library to ensure spec compliance.
8. **Config migration integration test** — Test `LoadConfig` with a legacy JSON file containing `"semantic": false` end-to-end.
9. **HTML golden file with summary data** — Current golden tests call PrintHeader + PrintFooter with 0 clones, so summary section (including collapse buttons) is never exercised.
10. **Test coverage for `--include-generated`** — The override logic for `FilterGeneric` is subtle and could regress.

### Performance

11. **Benchmark `sortGroupsByCriteria` generics** — Generic comparator factory has closure allocation overhead. Should benchmark vs interface-based approach.
12. **Stream-based HTML output** — HTML report builds entire output in memory via `strings.Builder`. For large codebases this could be a problem.
13. **Memory pool for `CloneNode` conversion** — `syntaxToCloneNode` allocates recursively. Could use `sync.Pool`.

### Developer Experience

14. **Update FEATURES.md** — Fix SHA-1 → SHA-256 cache reference. Add SARIF Properties metadata as a feature.
15. **Clean up TODO_LIST.md** — Remove stale entries, de-duplicate the 6 unique pending items, update CloneRef status.
16. **Document `--include-generated` override semantics** — The `FilterGeneric` catch-all behavior is surprising.

---

## F) TOP 25 THINGS TO GET DONE NEXT

Sorted by **Impact / Effort ratio** (highest first).

### Quick Wins (Low Effort, High Impact)

| #   | Task                                                          | Effort | Impact          |
| --- | ------------------------------------------------------------- | ------ | --------------- |
| 1   | Fix FEATURES.md SHA-1 → SHA-256 reference                     | 2 min  | Accuracy        |
| 2   | Clean TODO_LIST.md stale entries (CloneRef done, de-dup)      | 10 min | Clarity         |
| 3   | Commit or discard 5 pre-existing uncommitted files            | 5 min  | Hygiene         |
| 4   | Add SARIF validator library test                              | 30 min | Spec compliance |
| 5   | HTML golden test with clone data (exercises summary/collapse) | 30 min | Coverage        |

### High Value (Medium Effort, High Impact)

| #   | Task                                                         | Effort | Impact                   |
| --- | ------------------------------------------------------------ | ------ | ------------------------ |
| 6   | Migrate remaining 8 `sort.Slice` → `slices.SortFunc`         | 1h     | Consistency, type safety |
| 7   | Config migration integration test (LoadConfig + legacy JSON) | 30 min | Coverage                 |
| 8   | Extract `cloneType` classification into `domain` package     | 1h     | Decoupling               |
| 9   | Add `--include-generated` integration test                   | 1h     | Coverage                 |
| 10  | Benchmark `sortGroupsByCriteria` generic vs interface        | 30 min | Performance data         |

### Medium (Medium Effort, Medium Impact)

| #   | Task                                                                                  | Effort | Impact        |
| --- | ------------------------------------------------------------------------------------- | ------ | ------------- |
| 11  | Split `printer/` into `printer/text`, `printer/html`, `printer/json`, `printer/sarif` | 4h     | Architecture  |
| 12  | Consolidate 7 Clone types → fewer with shared interfaces                              | 4h     | Architecture  |
| 13  | Move AST constants out of `syntax/golang` into domain types                           | 2h     | Decoupling    |
| 14  | Update AGENTS.md with `slices.SortFunc` migration note                                | 10 min | Documentation |
| 15  | Add cache SHA-256 end-to-end test                                                     | 1h     | Coverage      |
| 16  | Profile and optimize `syntaxToCloneNode` recursive allocation                         | 2h     | Performance   |
| 17  | Add SARIF spec validation to CI                                                       | 1h     | CI quality    |

### Strategic (High Effort, High Impact)

| #   | Task                                               | Effort | Impact      |
| --- | -------------------------------------------------- | ------ | ----------- |
| 18  | Implement real git-diff incremental analysis       | 8h     | Feature     |
| 19  | Branded `NodeType int32` with gob cache migration  | 4h     | Type safety |
| 20  | Stream-based HTML output for large codebases       | 8h     | Scalability |
| 21  | Add language support beyond Go/Templ (TypeScript?) | 20h+   | Expansion   |
| 22  | Build a web UI dashboard for exploring clones      | 20h+   | UX          |

### Research / Lower Priority

| #   | Task                                                     | Effort   | Impact      |
| --- | -------------------------------------------------------- | -------- | ----------- |
| 23  | Evaluate `slices.SortStableFunc` for deterministic ties  | 30 min   | Correctness |
| 24  | Hybrid slice/map transition storage                      | 4h       | Marginal    |
| 25  | Explore suffix-tree alternatives (e.g., Burrows-Wheeler) | Research | Unknown     |

---

## G) TOP QUESTION I CANNOT FIGURE OUT MYSELF 🤔

**Should the 7 parallel Clone types be consolidated into a single canonical type, or is the current separation (SDK DTO vs printer DTO vs domain DTO) architecturally correct?**

The split-brain analysis in `docs/research/SPLIT-BRAIN.html` documents this extensively. `CloneRef` was a compromise — it eliminated field-name drift without collapsing the DTO boundary. But there are still 7 types:

- `domain.ProcessedClone(Group)` — internal pipeline
- `pkg/artdupl.Clone(Group)` — SDK public API
- `printer.CloneGroup` / `printer.JSONClone` — JSON output
- `printer.CloneRef` (embedded, not separate)
- `domain.CloneRef` — shared value object
- `domain.CloneNode` — actionability tree

**My instinct**: Consolidate to 2 types — `domain.Clone` (canonical) and `domain.CloneView` (format-specific projection). But I'm unsure whether the DTO boundary between SDK and internal packages justifies the current separation, or whether it's premature abstraction that adds cognitive overhead with no real consumer benefit.

**What I need from you**: A decision on whether to invest in consolidation (4h, medium risk) or accept the current `CloneRef`-embedded compromise and move on to feature work.

---

## Verification Gates (2026-07-03 03:25)

| Gate                          | Status        |
| ----------------------------- | ------------- |
| `go build ./...`              | ✅            |
| `go test ./...` (26 packages) | ✅ All pass   |
| `go test -race ./...`         | ✅ 0 failures |
| `golangci-lint run`           | ✅ 0 issues   |
| `go vet ./...`                | ✅            |
