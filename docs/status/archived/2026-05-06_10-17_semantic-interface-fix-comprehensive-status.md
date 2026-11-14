# art-dupl — Comprehensive Status Report

> **Live remainder:** archived 2026-11-14 (cadence wave); every item in this report is struck or routed, and open work is owned by TODO_LIST.md and ROADMAP.md.

**Date:** 2026-05-06 10:17 CEST
**Branch:** fork
**Last Commit:** `88b05a5` feat(semantic): add interface context tracking for semantic mode
**Codebase:** 207 Go files | 16,297 LOC (prod) | 29,505 LOC (tests) | 58 deps

---

## a) FULLY DONE ✓

### Core Detection Engine

- [x] **Suffix tree algorithm** — Ukkonen's algorithm, O(1) map-based transitions, streaming
- [x] **Hash-based detection** — XXH3 streaming, ~20x faster than SHA-256
- [x] **Multi-method detection** — Run both methods simultaneously via goroutines, deduplicated results
- [x] **Semantic-aware mode** (`--semantic`) — FNV-1a 24-bit identifier hashing, receiver+method combined
- [x] **Structural-only mode** (`--structural`) — Default behavior, AST shape matching
- [x] **Interface context tracking** — NEW: `--semantic` now distinguishes interface method signatures from concrete implementations (eliminates false positive where `Service.GetUser(ctx, id)` matched its interface declaration)

### Output Formats (7 formats)

- [x] Text output — Human-readable with diff hints
- [x] HTML output — Dark theme, syntax highlighting, VSCode links, diff visualization
- [x] JSON output — Structured with version, timestamp, clone_groups, summary
- [x] Simple-JSON output — Simpler format with impact score
- [x] Plumbing output — Machine-readable `file:startLine-endLine` for CI/CD
- [x] SARIF output — SARIF 2.1.0 for GitHub Advanced Security / CodeQL
- [x] CSV output (stats only) — Spreadsheet-compatible

### Statistics Subcommand

- [x] Text stats with lipgloss colored output
- [x] JSON stats for CI/CD
- [x] CSV stats for spreadsheets
- [x] Health grade (A-F) based on duplication metrics
- [x] Clone metrics: total clones, groups, files affected, duplication %
- [x] Spread analysis: complexity scores, severity distributions

### CLI & Configuration

- [x] Professional CLI with Fang/Cobra framework
- [x] Shell completion (bash, zsh, fish, PowerShell)
- [x] JSON config file support (`-c dupl.json`)
- [x] Config merge: defaults → file → CLI flags (flags win)
- [x] Sorting: size, occurrence, hash, total-tokens
- [x] Threshold control (`-t` / `--threshold`)
- [x] Batch output (`--all --output-dir`)

### Smart Filtering

- [x] SQLC generated code detection via `sqlc.yaml` parent directory scan
- [x] Templ `.templ` files included by default (`--exclude-templ` to filter)
- [x] `*_templ.go` files filtered by default
- [x] Protobuf `.pb.go` filtering with `--filter-generated`
- [x] Mockgen and stringer generated file filtering
- [x] Custom include/exclude patterns (`--include-pattern`, `--exclude-pattern`)
- [x] Vendor directory excluded by default (`--vendor` to include)

### Architecture & Code Quality (Recent Sprint 2026-04-30 → 2026-05-06)

- [x] Config extraction: `DetectionConfig` separated from `config.Config`
- [x] `MethodDetector` interface for pluggable detectors
- [x] File splits: `run_analysis.go` (450→3), `config.go` (344→3), `todos.go` (352→3), `html.go` (1484→4)
- [x] Deleted dead code: `cli/` package, `printer/format.go`, `printer/sort_type.go`, 6 unused domain types, 15 dead error vars
- [x] Config safety: `panic(err)` → proper error returns in `config_builder.go`
- [x] SARIF output: uses real clone hashes instead of position strings

### Testing

- [x] 22 packages, all passing, 0 failures
- [x] BDD test suite with Ginkgo/Gomega (21 test files)
- [x] Coverage: avg ~85% across packages (config 94.8%, syntax/golang 94.5%, hash 96.6%)
- [x] Fuzz testing infrastructure
- [x] Benchmark infrastructure with race detection

### Language Support

- [x] Go (`.go`) — Full AST analysis, 45 node types
- [x] Templ (`.templ`) — Pure Go parser, 28 node types, ON by default

---

## b) PARTIALLY DONE ⚠️

### TODO/FIXME/HACK Detection

- **Status:** Detector implemented (`detection/todo_detector.go`, `detection/legacy_detector.go`) but **not wired to CLI flags**
- `TodoDetector` and `LegacyDetector` exist in the `MultiDetector` registry but have no `-m todos` or `-m legacy` CLI flag exposure
- `MethodDetector` interface is defined but the CLI plumbing is incomplete

### CSV Clone Output

- **Status:** CSV works for stats subcommand. General clone CSV output uses manual string formatting, not `encoding/csv` — potential escaping issues with code containing commas/quotes

### SIMD Optimizations

- **Status:** Framework exists in `internal/simd/`, but 6 TODO items remain:
  - ARM64 SIMD detection not enabled (`internal/simd/simd.go:26`)
  - Actual vector size not returned (`internal/simd/simd.go:108`)
  - Placeholder implementations on non-x86 platforms

---

## c) NOT STARTED ○

1. ~~**TokenValue type** — Replace raw `int32` with validated `TokenValue` type across suffixtree/syntax (HIGH priority in TODO_LIST.md)~~ done — TokenValue type (suffixtree/suffixtree.go:34)
2. ~~**ProcessedClone DTO** — Decouple Printer from `syntax.Node` (111 test call sites to update)~~ done at `7e5a5cb`
3. ~~**Clone type consolidation** — Three parallel types (`printer.clone`, `pkg/artdupl.Clone`, `printer.CloneGroup`) need unification~~ done — CloneRef embed + format DTOs (ADR-0005)
4. ~~**`encoding/csv` for clone output** — Replace manual CSV formatting~~ **Won't implement — dropped — clone CSV keeps manual formatting; stats CSV uses encoding/csv (printer/stats).**
5. ~~**Enum unification** — Domain enums should use config's generic helpers~~ done — domain enums use pkg/enum helpers (domain/detection_mode.go)
6. ~~**String interning** — Memory optimization for large codebases~~ **Won't implement — dropped — memory-compact tree chosen instead (ADR-0022).**
7. ~~**SIMD-friendly memory layouts** — Align data structures for vectorized operations~~ **Superseded — internal/simd removed; ADR-0022 layout optimizations instead.**
8. ~~**transform.go refactor** — 366L, cognitive complexity 37 (gocognit threshold 35), 300L switch statement~~ done — resolved — lint clean today (106 linters, 0 findings)
9. ~~**Language support beyond Go/Templ** — No Rust, TypeScript, Python, etc.~~ **Won't implement — dropped — Go + templ only (AGENTS.md).**
10. ~~**nix flake migration** — Justfile still primary; flake.nix exists but not primary build tool~~ done — nix flake is the primary build (justfile removed)

---

## d) TOTALLY FUCKED UP 💥

### Lint Issues (2 pre-existing)

| Issue               | File                            | Detail                                                                                                              |
| ------------------- | ------------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| ~~`gocognit: 37 > 35`~~ done — resolved — lint clean today | ~~`syntax/golang/transform.go:13`~~ | ~~The `trans()` function's cognitive complexity is 2 over threshold. The 366-line switch statement is the root cause.~~ |
| ~~`unused`~~ done — resolved — lint clean today | ~~`bdd/stats_command_test.go:84`~~ | ~~`assertGeneratedFilesFilteredWithFlag` function declared but never called. Dead test helper.~~ |

### LSP Diagnostics (pre-existing, non-blocking)

| Type             | Count | Location                                                          |
| ---------------- | ----- | ----------------------------------------------------------------- |
| ~~`unusedwrite`~~ done — resolved — lint clean today | ~~6~~ | ~~`detection_test.go`, `detector_test.go`, `detector_types_test.go`~~ |
| ~~`unusedfunc`~~ done — resolved — lint clean today | ~~1~~ | ~~`bdd/stats_command_test.go`~~ |
| ~~`infertypeargs`~~ done — resolved — lint clean today | ~~1~~ | ~~`config/enum_helpers.go:78`~~ |
| ~~`gci` formatting~~ done — resolved — lint clean today | ~~1~~ | ~~`config/config.go:163`~~ |

### Test Coverage Gaps

| Package                      | Coverage | Risk                                                                     |
| ---------------------------- | -------- | ------------------------------------------------------------------------ |
| ~~`cmd/art-dupl` (main binary)~~ done — verified — main entry still tested via cmd/ package | ~~0.0%~~ | ~~No direct test of the main entry point (tested via cmd/ package instead)~~ |
| ~~`domain`~~ **Superseded — domain has extensive tests now (62.2%).** | ~~0.0%~~ | ~~Pure types, validated at construction sites~~ |
| ~~`examples`~~ done — verified — demo code | ~~0.0%~~ | ~~Demo code, not production~~ |
| ~~`internal/testhelpers`~~ done — verified — test infrastructure | ~~0.0%~~ | ~~Test infrastructure~~ |
| ~~`internal/filtertest`~~ **Superseded — internal/filtertest retained as a test utility.** | ~~50.0%~~ | ~~Missing edge case coverage~~ |

### Architecture Debt

- **318 status files** in `docs/status/` — massive accumulation, needs archival (keep last 30 days)
- **Printer ↔ syntax.Node coupling** — 6 printer implementations all depend on AST internals
- **58 dependencies** — Runtime deps are minimal (fang, cobra, ginkgo, gomega) but transitive tree is large

---

## e) WHAT WE SHOULD IMPROVE! 🎯

### Code Quality

1. ~~**Fix gocognit in transform.go** — Extract handler functions per AST node type (visitor pattern or table-driven dispatch). This is the #1 quality issue.~~ done — resolved — lint clean today
2. ~~**Delete dead test helper** `assertGeneratedFilesFilteredWithFlag` in `bdd/stats_command_test.go:84`~~ done — resolved — lint clean today
3. ~~**Fix unused writes** in test files (6 instances across detection/pkg tests)~~ done — resolved — lint clean today
4. ~~**Fix gci formatting** in `config/config.go:163`~~ done — resolved — lint clean today

### Architecture

5. ~~**ProcessedClone DTO** — This is the highest-value architectural improvement. It unlocks: printer simplification, clone type consolidation, and non-Go language support.~~ done at `7e5a5cb`
6. ~~**Wire TODO/Legacy detectors** — Code exists, just needs CLI flag plumbing (~20 lines)~~ **Superseded — Todo/Legacy detectors removed from detection/.**
7. ~~**Simplify status file accumulation** — Archive or auto-prune `docs/status/` beyond 30 days~~ done — docs/status/archived/ regime + archival README

### Performance

8. ~~**TokenValue type** — Replace raw `int32` with validated type for safety without sacrificing perf~~ done — TokenValue type (suffixtree/suffixtree.go:34)
9. ~~**SIMD on ARM64** — Enable the disabled SIMD paths for Apple Silicon~~ **Superseded — internal/simd removed.**
10. ~~**String interning** — Reduce memory for large codebases with repeated identifiers~~ **Won't implement — dropped — memory-compact tree chosen instead (ADR-0022).**

### Testing

11. ~~**Add interface false-positive BDD test** — The interface/impl semantic fix has no dedicated BDD test~~ **Superseded — no dedicated interface BDD spec recorded; semantic behavior pinned by bdd/semantic_detection_test.go suite.**
12. ~~**Coverage for `internal/filtertest`** — Currently at 50%, should be 80%+~~ **Superseded — internal/filtertest retained as a test utility.**
13. ~~**Edge case tests for SARIF output** — Verify edge cases (empty results, single file, etc.)~~ done — SARIF golden + round-trip tests (printer/testdata/TestSARIFOutputGolden.golden)

---

## f) Top #25 Things We Should Get Done Next

| #  | Priority | Item                                                                      | Est. Effort | Impact              |
| -- | -------- | ------------------------------------------------------------------------- | ----------- | ------------------- |
| ~~1~~ **Superseded — semantic suite pins behavior; no dedicated interface-impl spec.** | ~~🔴 HIGH~~ | ~~Add BDD test for interface semantic false-positive fix~~ | ~~30min~~ | ~~Correctness~~ |
| ~~2~~ done — resolved — lint clean today | ~~🔴 HIGH~~ | ~~Delete dead `assertGeneratedFilesFilteredWithFlag` test helper~~ | ~~5min~~ | ~~Clean code~~ |
| ~~3~~ done — resolved — lint clean today | ~~🔴 HIGH~~ | ~~Fix gci formatting in `config/config.go:163`~~ | ~~2min~~ | ~~Clean code~~ |
| ~~4~~ done — resolved — lint clean today | ~~🔴 HIGH~~ | ~~Fix 6 unused writes in test files~~ | ~~15min~~ | ~~Clean code~~ |
| ~~5~~ done — docs/status/archived/ regime | ~~🔴 HIGH~~ | ~~Archive `docs/status/` — keep last 30 days, compress rest~~ | ~~30min~~ | ~~Repo hygiene~~ |
| ~~6~~ **Superseded — Todo/Legacy detectors removed.** | ~~🟡 MED~~ | ~~Wire TODO/Legacy detectors to CLI flags (`-m todos`, `-m legacy`)~~ | ~~1hr~~ | ~~Feature complete~~ |
| ~~7~~ **Won't implement — dropped — clone CSV keeps manual formatting.** | ~~🟡 MED~~ | ~~Implement `encoding/csv` for clone CSV output~~ | ~~2hr~~ | ~~Correctness~~ |
| ~~8~~ done — TokenValue type | ~~🟡 MED~~ | ~~Implement TokenValue type with validation~~ | ~~3hr~~ | ~~Type safety~~ |
| ~~9~~ done — resolved — lint clean today | ~~🟡 MED~~ | ~~Extract AST node handlers from `transform.go` to reduce gocognit~~ | ~~4hr~~ | ~~Maintainability~~ |
| ~~10~~ done at `7e5a5cb` | ~~🟡 MED~~ | ~~Introduce ProcessedClone DTO to decouple Printer from syntax.Node~~ | ~~8hr~~ | ~~Architecture~~ |
| ~~11~~ done — CloneRef embed + format DTOs (ADR-0005) | ~~🟡 MED~~ | ~~Consolidate three Clone types into unified domain model~~ | ~~6hr~~ | ~~Architecture~~ |
| ~~12~~ done — file-kind-scoped dispatch in printer/actionability/clone_classify.go (AGENTS.md) | ~~🟡 MED~~ | ~~Move `printer/clone_classify.go` language-specific logic behind interface~~ | ~~3hr~~ | ~~Multi-language prep~~ |
| ~~13~~ **Superseded — internal/filtertest retained as a test utility.** | ~~🟡 MED~~ | ~~Add coverage for `internal/filtertest` (50% → 80%)~~ | ~~2hr~~ | ~~Quality~~ |
| ~~14~~ done — FEATURES.md maintained under the docs-health count gate | ~~🟡 MED~~ | ~~Update FEATURES.md to reflect semantic interface tracking~~ | ~~15min~~ | ~~Documentation~~ |
| ~~15~~ done — TODO_LIST.md maintained | ~~🟡 MED~~ | ~~Update TODO_LIST.md to mark completed items~~ | ~~15min~~ | ~~Documentation~~ |
| ~~16~~ done — AGENTS.md documents semantic-mode interface handling | ~~🟡 MED~~ | ~~Update AGENTS.md with interface context tracking decision~~ | ~~10min~~ | ~~Memory~~ |
| ~~17~~ done — domain enums use pkg/enum helpers | ~~🟢 LOW~~ | ~~Unify domain enums to use config's generic helpers~~ | ~~2hr~~ | ~~Consistency~~ |
| ~~18~~ **Won't implement — dropped — memory-compact tree chosen instead (ADR-0022).** | ~~🟢 LOW~~ | ~~Implement string interning for repeated identifiers~~ | ~~3hr~~ | ~~Performance~~ |
| ~~19~~ **Superseded — internal/simd removed.** | ~~🟢 LOW~~ | ~~Enable ARM64 SIMD detection in `internal/simd/`~~ | ~~4hr~~ | ~~Performance~~ |
| ~~20~~ **Superseded — parser fuzzing landed instead (syntax/templ/fuzz_test.go).** | ~~🟢 LOW~~ | ~~Add fuzz tests for semantic encoding (FNV collision detection)~~ | ~~2hr~~ | ~~Robustness~~ |
| ~~21~~ done — BenchmarkSemanticVsExactVsStructural (syntax/golang/benchmark_test.go) | ~~🟢 LOW~~ | ~~Benchmark: semantic vs structural mode performance comparison~~ | ~~1hr~~ | ~~Documentation~~ |
| ~~22~~ **Won't implement — dropped — no --dry-run flag.** | ~~🟢 LOW~~ | ~~Add `--dry-run` flag that shows what would be analyzed without running~~ | ~~2hr~~ | ~~UX~~ |
| ~~23~~ done — nix flake is the primary build (justfile removed) | ~~🟢 LOW~~ | ~~Migrate justfile recipes to nix flake~~ | ~~4hr~~ | ~~Build system~~ |
| ~~24~~ **Won't implement — dropped — no .git-blame-ignore-revs.** | ~~🟢 LOW~~ | ~~Add `.git-blame-ignore-revs` for mass-refactor commits~~ | ~~10min~~ | ~~DX~~ |
| ~~25~~ **Won't implement — dropped — Go + templ only (AGENTS.md).** | ~~🟢 LOW~~ | ~~Investigate supporting Rust/TypeScript via tree-sitter~~ | ~~Research~~ | ~~Multi-language~~ |

---

## g) Top #1 Question I Can NOT Figure Out Myself

**Should `--semantic` become the default detection mode?**

Currently:

- Default = structural-only (`--structural` explicit, but that's the default behavior)
- `--semantic` is opt-in

The new interface context tracking makes semantic mode strictly better for reducing false positives (interface↔impl, enum patterns, mock↔interface all eliminated). However:

- **Performance**: Semantic mode adds FNV hashing overhead per identifier. Unknown if this is measurable at scale (100K+ files).
- **Breaking change**: Users relying on structural matching would get different results.
- **Completeness**: Are there other false-positive patterns that semantic mode doesn't handle yet?

**I need your decision**: Should we flip the default to semantic, keep structural as default, or add a config option for project-level defaults?

---

## Current Working Tree Changes

```
modified:   syntax/golang/transform.go    (interface context tracking for Field nodes + whitespace fix)
deleted:    testdata/interface_semantic/   (cleaned up test data)
```

## Test Results

```
22/22 packages PASS | 0 FAIL
Coverage: config 94.8% | syntax/golang 94.5% | hash 96.6% | suffixtree 91.0% | printer 84.8%
Lint: 2 pre-existing issues (gocognit, unused func)
```

---

_Auto-generated by Crush on 2026-05-06_
