# art-dupl — Comprehensive Status Report

**Date:** 2026-05-20 04:26\
**Branch:** fork\
**Head:** a8f166c `refactor: modernize config merge, test helpers, and actionability constants`\
**Uncommitted Changes:** `cmd/flags.go` — moved `--only` flag from root-only to shared flags (fixes `art-dupl stats --only templ`)

---

## Executive Summary

The project is in **excellent health**. All 23 packages compile clean, 0 lint issues, all 253 BDD specs green, average test coverage ~85%. The just-completed session fixed a flag availability bug where `--only` was unavailable on the `stats` subcommand. The codebase has been through a major modernization pass (2026-04-30 to 2026-05-17) that resolved 5 critical architecture issues, introduced semantic-as-default, and eliminated significant technical debt.

**One session change:** `--only` flag moved to shared flags so `art-dupl stats --only templ` works.

---

## a) FULLY DONE ✅

### Session Fix (2026-05-20)

- **`--only` flag on stats subcommand** — Moved from root-only (`AddFlags`) to shared flags (`addSharedFlags`) in `cmd/flags.go`. The config builder (`applyChangedStringFlags`) already handled it. Stats now accepts `--only go` or `--only templ`.

### Architecture Modernization (2026-04-30 — 2026-05-17)

- **Semantic-as-default** — `DefaultConfig.Semantic` changed from `false` to `true`. `--structural` disables it. First-use noise reduced by 68%.
- **Reflection-based config merge** — `mergeConfig()` reduced from 170-line manual field list to 30-line reflection-based merge. New Config fields no longer require touching merge code.
- **Pipeline unification** — CLI and SDK share `detection.MultiDetector`. Fixed critical bug where SDK rebuilt suffix tree from scratch (double CPU/memory).
- **Config extraction** — `DetectionConfig` extracted from `config.Config`. `MethodDetector` interface for pluggable detectors.
- **Printer cleanup** — Deleted `printer/format.go` (moved `ParseFormat` to `config`), deleted `printer/sort_type.go` (moved `SortBy` to `config.SortCriteria`). Printer interface uses `config.SortCriteria`.

### File Splits & Size Reduction (2026-04-30 — 2026-05-03)

- `printer/html.go` 1484L → 4 files (365L, 523L, 315L, 302L)
- `detection/todos.go` 352L → 3 files
- `config/config.go` 344L → 3 files
- `cmd/run_analysis.go` 450L → 3 files

### Code Quality (2026-04-30 — 2026-05-17)

- **Domain cleanup** — Removed 6 unused types (TokenCount, FileCount, CloneCount, Threshold, BytePosition) and 15 dead error variables
- **Error modernization** — `errors.As` → `errors.AsType` (Go 1.24+). All `//nolint:err113` replaced with typed errors
- **Config builder safety** — Replaced `panic(err)` in `cmd/config_builder.go` with proper error returns
- **Config.Only typed as FileType** — Uses `FileType.Matches()` instead of hand-rolled matcher
- **Timestamps as time.Time** — `domain.Analysis.CreatedAt` from `string` to `time.Time`
- **Changed() tracking** — Eliminated FlagValues split brain; `--semantic`/`--structural` mutual exclusion handled properly

### CI & Infrastructure

- Lint: **0 issues** (golangci-lint)
- Build: **clean** (just build)
- Tests: **all 23 packages pass**
- BDD: **253 specs all green**
- Nix flake: **functional** with private dependency pattern for gogenfilter

---

## b) PARTIALLY DONE 🟡

### CSV Output

- **Stats CSV:** Works via `art-dupl stats --format csv`
- **General clone CSV:** Uses manual `strings.Builder` formatting, not `encoding/csv`. Won't handle edge cases (commas in paths, quotes).

### TODO/Legacy Detection Methods

- **Implemented** in `detection/todos.go` (split into 3 files)
- **Wired** through `MultiDetector` and accessible via `-m` flag
- **Not documented** as user-facing features in README or HOW_TO_USE
- **Status in FEATURES.md:** `DEFINED_ONLY` — may need user docs or deliberate `EXPERIMENTAL` labeling

### `docs/status/` Archive

- **334 files** accumulated. TODO_LIST says "archive old files, keep last 30 days" but it keeps growing. Not urgent but messy.

---

## c) NOT STARTED ⬜

### From TODO_LIST.md

| Priority | Item                                                      | Notes                                                         |
| -------- | --------------------------------------------------------- | ------------------------------------------------------------- |
| HIGH     | Implement TokenValue type with validation                 | Refactor suffixtree/syntax to use it                          |
| MEDIUM   | Optimize memory layouts for SIMD-friendly data structures | + string interning                                            |
| MEDIUM   | Implement CSV output using `encoding/csv`                 | General clone CSV, not stats                                  |
| MEDIUM   | Unify enum patterns                                       | Domain enums → config's generic helpers                       |
| MEDIUM   | Introduce ProcessedClone DTO                              | Decouple Printer from syntax.Node (111 test call sites!)      |
| MEDIUM   | Consolidate three parallel Clone types                    | `printer.clone`, `pkg/artdupl.Clone`, `domain.ProcessedClone` |
| LOW      | Refactor `syntax/golang/transform.go`                     | 355L, 300L switch statement                                   |
| LOW      | Archive old docs/status/ files                            | 334 files                                                     |
| LOW      | Fix remaining LSP hints                                   | Unused params, unnecessary type args                          |
| LOW      | Implement SIMD TODOs                                      | 6 items — though xxh3 already has SIMD built-in               |

### Not on TODO_LIST but Notable

- Multi-language support (blocked by `printer/clone_classify.go` importing `syntax/golang` directly)
- Release automation (GoReleaser config exists but had multiple fixup commits — may need validation)
- SDK documentation (pkg/artdupl exists but no dedicated docs)

---

## d) TOTALLY FUCKED UP 💥

### Nothing is critically broken right now.

All tests pass, lint is clean, build is clean. But here's what's fragile or regrettable:

### `internal/simd/` — Dead Code

- `Available()` always returns `false`. `Hash()` returns `nil`. `HashSlice()` returns `nil`.
- Only 2 real TODOs left in the codebase are both in this package.
- xxh3 already provides SIMD-optimized hashing natively — this abstraction adds zero value.
- `hashSeqSIMD()` in `syntax/hash_simd.go` is just dead indirection calling `hashSeqFallback()`.
- **Verdict:** Delete the entire package. Replace any call sites with direct xxh3 usage.

### `docs/status/` — 334 Files

- This is a mess. Most are session-generated. Should be archived aggressively.

### Printer ↔ syntax.Node Coupling

- `Printer.PrintClones(dups [][]*syntax.Node)` forces all 6 implementations to depend on AST internals.
- Each printer independently calls `ProcessNodeRange()` and `extractContent()`.
- Fix requires ProcessedClone DTO — touches **111 test call sites**. Deferred multiple times but the debt grows.

---

## e) WHAT WE SHOULD IMPROVE 🔧

### Immediate Quality Gains

1. **Delete `internal/simd/`** — Pure dead code. Zero risk, removes 163 lines and 2 stale TODOs.
2. **Domain test coverage (67.2%)** — Lowest production package. New `ProcessedClone`/`ProcessedCloneGroup` types need tests.
3. **`cmd/` coverage (72.2%)** — Test the new `--only` on stats. Test flag combinations.
4. **CSV proper implementation** — Replace manual formatting with `encoding/csv`. Edge cases will bite eventually.
5. **Archive `docs/status/`** — 334 files is absurd. Keep last 10, archive the rest.

### Architecture Improvements

6. **ProcessedClone DTO** — Decouple Printer from syntax.Node. This is the #1 architecture debt. 111 test call sites makes it scary but every session that passes makes it worse.
7. **Consolidate Clone types** — `printer.clone`, `pkg/artdupl.Clone`, `domain.ProcessedClone` are 3 parallel representations. Unify.
8. **`config/detection_method.go` (426L)** — God file holding DetectionMethod, OutputFormat, SortBy, DiffMode, FileType all in one. Split by concern.
9. **`syntax/golang/transform.go` (355L)** — 300L switch statement. Could be table-driven.
10. **Printer package size (44 files)** — Natural split: formatters (text, html, json, plumbing, sarif) vs processing (diff, classify, build).

### Process Improvements

11. **ADR discipline** — Only 1 ADR exists (0001-map-based-transition). Major decisions like semantic-as-default, reflection merge, ProcessedClone DTO should have ADRs.
12. **BDD coverage for edge cases** — Stats subcommand with `--only`, config file + CLI flag interaction, incremental mode edge cases.
13. **Error handling consistency** — `cmd/` uses structured `duplerrors.Wrap*`, most other packages use bare `fmt.Errorf`. Pick one and standardize.
14. **`//nolint:` audit** — 71 directives. Some may be stale after refactoring.

---

## f) Top #25 Things We Should Get Done Next

| #  | Item                                                                                          | Impact | Effort            | Priority       |
| -- | --------------------------------------------------------------------------------------------- | ------ | ----------------- | -------------- |
| 1  | Delete `internal/simd/` package (dead code)                                                   | Low    | Tiny              | Quick win      |
| 2  | Delete `hashSeqSIMD()` dead indirection in syntax/hash_simd.go                                | Low    | Tiny              | Quick win      |
| 3  | Archive `docs/status/` — keep last 10, move rest to `docs/status/archive/`                    | Medium | Small             | Cleanup        |
| 4  | Add tests for `domain/` — target 80%+ coverage (currently 67.2%)                              | Medium | Medium            | Quality        |
| 5  | Add BDD test for `art-dupl stats --only templ` and `--only go`                                | Medium | Small             | Coverage       |
| 6  | Add BDD test for `art-dupl stats --format csv`                                                | Low    | Small             | Coverage       |
| 7  | Implement CSV output using `encoding/csv` (general clone CSV)                                 | Medium | Small             | Correctness    |
| 8  | Consolidate three parallel Clone types into unified hierarchy                                 | High   | Large             | Architecture   |
| 9  | Introduce ProcessedClone DTO — decouple Printer from syntax.Node                              | High   | Large (111 sites) | Architecture   |
| 10 | Split `config/detection_method.go` (426L) — separate FileType, OutputFormat, SortBy, DiffMode | Medium | Medium            | Code quality   |
| 11 | Refactor `syntax/golang/transform.go` — table-driven 300L switch                              | Medium | Medium            | Code quality   |
| 12 | Split printer/ into sub-packages (formatters vs processing)                                   | Medium | Large             | Architecture   |
| 13 | Standardize error handling — structured errors everywhere or bare fmt.Errorf                  | Medium | Medium            | Consistency    |
| 14 | Implement TokenValue type with validation                                                     | Medium | Medium            | Type safety    |
| 15 | Unify enum patterns — domain enums → config's generic helpers                                 | Low    | Medium            | Consistency    |
| 16 | Add ADR for semantic-as-default decision                                                      | Low    | Tiny              | Documentation  |
| 17 | Add ADR for reflection-based config merge                                                     | Low    | Tiny              | Documentation  |
| 18 | Write SDK documentation for `pkg/artdupl/`                                                    | Medium | Medium            | Documentation  |
| 19 | Audit 71 `//nolint:` directives — remove stale ones                                           | Low    | Small             | Cleanup        |
| 20 | Document TODO/Legacy detection methods as experimental or remove DEFINED_ONLY status          | Low    | Small             | Documentation  |
| 21 | Fix LSP hints: unused params, unnecessary type args in tests                                  | Low    | Small             | Polish         |
| 22 | Implement string interning for AST processing memory optimization                             | Medium | Large             | Performance    |
| 23 | Optimize memory layouts for SIMD-friendly data structures                                     | Medium | Large             | Performance    |
| 24 | Validate GoReleaser release config — had 5+ fixup commits                                     | Medium | Small             | Infrastructure |
| 25 | Remove `lib/` legacy package (listed as "being phased out" in AGENTS.md)                      | Low    | Small             | Cleanup        |

---

## g) Top #1 Question I Cannot Figure Out Myself

**Should `internal/simd/` be deleted entirely, or is there a genuine future plan for ARM64/architecture-specific SIMD beyond what xxh3 already provides?**

The package has been a stub since creation. xxh3 already ships with runtime SIMD detection for amd64 and arm64. The only caller (`syntax/hash_simd.go`) immediately falls through to the xxh3 fallback. Two TODOs reference ARM64 SIMD but the code path is never reached. If there's no plan for custom SIMD operations beyond hashing, this entire package (163 lines + 2 TODOs) is dead weight.

---

## Scorecard

| Metric              | Value                            | Trend                        |
| ------------------- | -------------------------------- | ---------------------------- |
| Build               | ✅ Clean                         | Stable                       |
| Lint                | ✅ 0 issues                      | Stable                       |
| Tests               | ✅ 253 specs green               | Stable                       |
| Avg Coverage        | ~85%                             | Stable                       |
| Packages below 80%  | 3 (domain 67%, cmd 72%, job 77%) | Slightly worse (cmd was 75%) |
| Open TODOs          | 2 (both in dead simd/)           | Stable                       |
| Direct Dependencies | 8                                | Lean                         |
| Dead Code           | `internal/simd/` (163L)          | Unchanged                    |
| Files > 300L        | ~6 prod files                    | Improved (was more)          |
| docs/status/ files  | 334                              | Growing                      |
| ADRs                | 1                                | Stale                        |
